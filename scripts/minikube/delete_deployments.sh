#!/bin/bash

# Exit immediately if any command fails
set -e

# Helpers
delete_env_secret() {
  local service="$1"
  kubectl delete secret "${service//_/-}-secret"
}

services=(
  "gateway_service"
  "user_service"
  "app_service"
  "build_service"
  "deploy_service"
  "log_service"
  "project_service"
  "frontend_service"
)

echo "[*] Deleting Deployments..."

echo "[*] Uninstall Apps Hosting Helm Charts..."
helm uninstall apps-hosting

echo "[*] Uninstall Nats Helm Charts..."
helm uninstall nats

echo "[*] Delete global-secret..."
kubectl delete secret "global-secret"

echo "[*] Delete services secrets..."
for service in "${services[@]}"; do
    delete_env_secret "$service"
done