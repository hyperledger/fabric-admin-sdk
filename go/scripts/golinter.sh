#!/usr/bin/env bash
#

set -e

IMAGE_VERSION="v1.49.0"
go mod vendor
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:${IMAGE_VERSION} golangci-lint --timeout 10m run -v
