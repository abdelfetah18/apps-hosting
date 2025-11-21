#!/bin/bash

# Publishes the package using the latest commit for versioning

if [ $# -ne 2 ]; then
    echo "Usage: $0 <path_package> <version>"
    exit 1
fi

PACKAGE_PATH="$1"
BASE_VERSION="$2"

BASE_DIR=$(pwd)
cd "$BASE_DIR/infrastructure/go-registry" || { echo "Failed to cd into go-registry"; exit 1; }

# Get latest commit hash and date from the current repo
COMMIT_HASH=$(git rev-parse HEAD)
COMMIT_DATE=$(git show -s --format=%cd --date=format:%Y%m%d%H%M%S "$COMMIT_HASH")

generate_pseudo_version() {
    local base_version="$1"
    local hash="$2"
    local ts="$3"
    local short_hash="${hash:0:12}"
    echo "${base_version}-${ts}-${short_hash}"
}

PACKAGE_VERSION=$(generate_pseudo_version "$BASE_VERSION" "$COMMIT_HASH" "$COMMIT_DATE")

echo "Uploading package '$PACKAGE_PATH' with version '$PACKAGE_VERSION'..."
go run . upload "$PACKAGE_VERSION" "$BASE_DIR/$PACKAGE_PATH"

# Store the generated version in a file inside the package path
VERSION_FILE="$BASE_DIR/$PACKAGE_PATH/.last_version"
echo "$PACKAGE_VERSION" > "$VERSION_FILE"
echo "Stored version in $VERSION_FILE"
