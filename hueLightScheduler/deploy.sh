#!/bin/bash

# Exit on any error
set -e

# Script must be run from the hueLightScheduler directory
if [[ ! -f "config.yml" ]]; then
    echo "Error: config.yml not found. Please run this script from the hueLightScheduler directory"
    exit 1
fi

# Check if HUE_ID environment variable is set
if [[ -z "${HUE_ID}" ]]; then
    echo "Error: HUE_ID environment variable is not set"
    exit 1
fi

# Create manifests/base directory if it doesn't exist
mkdir -p manifests/base

# Copy config.yml to manifests/base
echo "Copying config.yml to manifests/base..."
cp config.yml manifests/base/

# Create base64 encoded HUE_ID
export HUE_ID_BASE64=$(echo -n $HUE_ID | base64)

# Apply secret and other manifests
echo "Applying secret and manifests..."
sed -i -e "s#{HUE_ID}#${HUE_ID_BASE64}#g" manifests/secrets.yaml

# Apply Kubernetes manifests using kustomize
echo "Applying Kubernetes manifests..."
kubectl apply -k manifests/

# Clean up
echo "Cleaning up..."
rm manifests/base/config.yml

echo "Deployment complete!" 