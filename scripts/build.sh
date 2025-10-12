#!/bin/bash

set -e

VERSION=${VERSION:-"dev"}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Building library service..."
echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Build time: $BUILD_TIME"

# Build API server
echo ""
echo "Building API server..."
CGO_ENABLED=0 go build \
    -ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME" \
    -o bin/library-api \
    ./cmd/api

# Build worker
echo "Building worker..."
CGO_ENABLED=0 go build \
    -ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME" \
    -o bin/library-worker \
    ./cmd/worker

# Build migration tool
echo "Building migration tool..."
CGO_ENABLED=0 go build \
    -ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME" \
    -o bin/library-migrate \
    ./cmd/migrate

echo ""
echo "Build complete! Binaries are in ./bin/"
ls -lh bin/
