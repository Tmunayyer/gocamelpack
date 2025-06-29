#!/bin/bash

# Build script that injects version information
set -e

# Get version from git tag or use default
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT_SHA=${COMMIT_SHA:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}
BUILD_DATE=${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}

# Build flags for version injection
LDFLAGS="-X github.com/Tmunayyer/gocamelpack/cmd.Version=${VERSION} \
         -X github.com/Tmunayyer/gocamelpack/cmd.CommitSHA=${COMMIT_SHA} \
         -X github.com/Tmunayyer/gocamelpack/cmd.BuildDate=${BUILD_DATE}"

echo "Building gocamelpack with:"
echo "  Version: ${VERSION}"
echo "  Commit: ${COMMIT_SHA}"
echo "  Build Date: ${BUILD_DATE}"

go build -ldflags "${LDFLAGS}" -o gocamelpack .