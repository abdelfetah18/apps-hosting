#!/bin/bash

services=(
  "gateway_service"
  "user_service"
  "app_service"
  "build_service"
  "deploy_service"
  "log_service"
  "project_service"
)

generate_protobuf_code() {
  for proto_file in src/protos/*; do
    if [[ -f "$proto_file" ]]; then
      echo ""
      for service in "${services[@]}"; do
        echo "[*] Generating client and server code $proto_file for $service..."
        protoc --go_out="src/${service}" --go-grpc_out="src/${service}" "$proto_file"  
      done
    fi
  done

  return 0
}

generate_protobuf_code