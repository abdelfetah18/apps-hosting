#!/bin/bash

# Exit immediately if any command fails
set -e

# Helpers
check_minikube_status() {
  local status
  status=$(minikube status 2>/dev/null)

  if [[ $? -ne 0 ]]; then
    echo "Minikube is not installed or not accessible in PATH."
    return 2
  fi

  if echo "$status" | grep -q "host: Running" &&
     echo "$status" | grep -q "kubelet: Running" &&
     echo "$status" | grep -q "apiserver: Running"; then
    return 0
  else
    echo "Minikube is not fully running."
    echo "$status"
    return 1
  fi
}

# 1. Run minikube
if ! check_minikube_status; then
    # Start minikube
    minikube start --embed-certs --insecure-registry="192.168.49.2:5000,docker-registry:5000"

# 2. Install ingress & registry
    # Enable ingress
    minikube addons enable ingress

    # Enable Registry
    minikube addons enable registry
fi

# 3. Deploy Artifcat Registry (GoLang)
#   - deploy minio
helm repo add minio https://charts.min.io/ --force-update
helm install object-storage -f infrastructure/go-registry/deployment/minio-values.yaml minio/minio

#   - deploy artifact
if ! kubectl create secret generic "go-registry-secret" --from-env-file="./infrastructure/go-registry/.server.env"; then
  echo "go-registry-secret already exists"
fi

docker build -t 192.168.49.2:5000/go-artifcat:dev "./infrastructure/go-registry"
docker push 192.168.49.2:5000/go-artifcat:dev
kubectl apply -f ./infrastructure/go-registry/deployment/deployment.yaml

# 4. Deploy tls config (ssl certificates)
echo "Installing ssl..."
if ! kubectl create secret tls tls-apps-hosting.com --key "config/tls/apps-hosting.com-key.pem" --cert "config/tls/apps-hosting.com.pem"; then
    echo "Failed to create tls secret"
fi
if ! kubectl create secret tls tls-wildcard.apps-hosting.com --key "config/tls/_wildcard.apps-hosting.com-key.pem" --cert "config/tls/_wildcard.apps-hosting.com.pem"; then
    echo "Failed to create tls-wildcard secret"
fi

# 5. Deply grafana UI only
helm repo add grafana https://grafana.github.io/helm-charts --force-update
helm install my-grafana grafana/grafana
helm install tempo grafana/tempo
kubectl apply -f ./infrastructure/grafana/ingress.yaml

# Deploy docker registry
helm repo add twuni https://twuni.github.io/docker-registry.helm --force-update
helm install docker-registry -f infrastructure/docker-registry/values.yaml twuni/docker-registry
