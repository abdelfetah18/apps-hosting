#!/bin/bash

# Exit immediately if any command fails
set -e

# Point shell to minikube's docker-daemon
eval "$(minikube -p minikube docker-env)"

services=(
  "golang:gateway_service"
  "golang:user_service"
  "golang:app_service"
  "golang:build_service"
  "golang:deploy_service"
  "golang:log_service"
  "golang:project_service"
  "react:frontend_service"
)
registry_url="$(minikube ip):5000/"
private_goproxy=go-registry.apps-hosting.com

# Helpers
build_service() {
  local service="$1"

  if [[ -z "$service" ]]; then
    echo "Error: build_service: service name is required."
    return 1
  fi

  echo "[*] Building $service..."
  local image_name="$registry_url${service//_/-}:dev"

  if ! docker build -t "$image_name" "src/$service" --build-arg goproxy_url=$private_goproxy; then
    echo "Failed to build $service."
    return 1
  fi

  return 0
}

push_service_image() {
  local service="$1"

  if [[ -z "$service" ]]; then
    echo "Error: push_service_image: service name is required."
    return 1
  fi

  echo "[*] Pushing Service Image $service..."
  local image_name="$registry_url${service//_/-}:dev"
  if ! docker push $image_name; then
    echo "Failed to Push Service Image $service"
    return 1
  fi

  return 0
}

create_env_secret() {
  local service="$1"
  env_path="src/$service/.env"
  if ! kubectl create secret generic "${service//_/-}-secret" --from-env-file=$env_path; then
    echo "${service//_/-}-secret already exists"
  fi
}

# 1. Test

# 2. Build & Push Services
echo "[*] Building all services..."

# Applying Global Config
echo "[*] Adding global env vars..."
if ! kubectl create secret generic "global-secret" --from-env-file="config/.global.env"; then
  echo "global-secret already exists"
fi

for service in "${services[@]}"; do
    service_type=${service%%:*}
    service_name=${service##*:}
    
    if ! build_service "$service_name"; then
      continue
    fi
    if ! push_service_image "$service_name"; then
      continue
    fi
    create_env_secret "$service_name"
done

# 3. Deploy nats
install_nats() {
  if ! helm repo add nats https://nats-io.github.io/k8s/helm/charts/; then
    echo "Failed to add nats repo"
    return 1
  fi

  if ! helm install nats nats/nats --set config.jetstream.enabled=true; then
    echo "Failed to install nats"
    return 1
  fi

  return 0
}

install_nats

# Wait for NATS to be ready
echo "[*] Waiting for NATS to be ready..."
kubectl wait --for=condition=ready pod/nats-0 -n default --timeout=120s

# 4. Deploy Services
echo "[*] Deploying all services using helm"
helm install apps-hosting ./helm-chart
