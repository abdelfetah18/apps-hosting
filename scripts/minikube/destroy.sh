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

# 1. Check if minikube is running
if ! check_minikube_status; then
    echo "Error: Minikube is not running."
fi

# 3. Delete Artifcat Registry (GoLang)
echo "[*] Deleting Artifcat Registry (GoLang)..."
helm uninstall object-storage
kubectl delete secret go-registry-secret
kubectl delete -f ./infrastructure/go-registry/deployment/deployment.yaml

# 4. Deleting tls config (ssl certificates)
echo "[*] Deleting tls secrets..."
kubectl delete secret tls-apps-hosting.com tls-wildcard.apps-hosting.com

# 5. Stop minikube
minikube stop

echo "[*] Done."