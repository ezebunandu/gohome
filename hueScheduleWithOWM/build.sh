#!/bin/bash

set -e

REGISTRY="registry.home-k3s.lab/gohome"
VERSION="${1:-latest}"

echo "Building and pushing containers to ${REGISTRY} with version ${VERSION}"

# Build scheduler
echo "Building scheduler..."
docker build -f cmd/scheduler/Dockerfile -t ${REGISTRY}/scheduler:${VERSION} .
docker push ${REGISTRY}/scheduler:${VERSION}
echo "✓ Scheduler built and pushed"

# Build sunrise
echo "Building sunrise..."
docker build -f cmd/sunrise/Dockerfile -t ${REGISTRY}/sunrise:${VERSION} .
docker push ${REGISTRY}/sunrise:${VERSION}
echo "✓ Sunrise built and pushed"

# Build sunset
echo "Building sunset..."
docker build -f cmd/sunset/Dockerfile -t ${REGISTRY}/sunset:${VERSION} .
docker push ${REGISTRY}/sunset:${VERSION}
echo "✓ Sunset built and pushed"

echo "All containers built and pushed successfully!"
