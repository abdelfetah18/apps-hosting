#!/bin/bash


BASE_DIR=$(pwd)

cd $BASE_DIR/infrastructure/go-registry

go run . upload v1.0.0 $BASE_DIR/internal-packages/messaging
go run . upload v1.0.0 $BASE_DIR/internal-packages/logging