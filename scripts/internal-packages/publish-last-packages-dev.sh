#!/bin/bash

# Publishes internal packages using the versions stored in .last_version

PACKAGES=("internal-packages/messaging" "internal-packages/logging")
BASE_DIR=$(pwd)
cd "$BASE_DIR/infrastructure/go-registry" || { echo "Failed to cd into go-registry"; exit 1; }

for PACKAGE_PATH in "${PACKAGES[@]}"; do
    VERSION_FILE="$BASE_DIR/$PACKAGE_PATH/.last_version"

    if [ ! -f "$VERSION_FILE" ]; then
        echo "Error: .last_version file not found in $PACKAGE_PATH. Run the version generation script first."
        continue
    fi

    PACKAGE_VERSION=$(<"$VERSION_FILE")
    echo "Publishing package '$PACKAGE_PATH' with version '$PACKAGE_VERSION'..."
    go run . upload "$PACKAGE_VERSION" "$BASE_DIR/$PACKAGE_PATH"
done
