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

skip_protos=(
  "events.proto"
)

generate_protobuf_code() {
  for proto_file in src/protos/*; do
    if [[ -f "$proto_file" ]]; then
      base_proto=$(basename "$proto_file")

      # Skip if proto is in skip list
      skip=false
      for skip_proto in "${skip_protos[@]}"; do
        if [[ "$base_proto" == "$skip_proto" ]]; then
          skip=true
          break
        fi
      done
      $skip && continue

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

echo "[*] Generating events code..."
protoc --go_out="internal-packages/messaging" "src/protos/events.proto"
