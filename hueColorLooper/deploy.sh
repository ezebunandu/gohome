#!/bin/bash

# Exit on any error
set -e

# Check if HUE_ID environment variable is set
if [[ -z "${HUE_ID}" ]]; then
    echo "Error: HUE_ID environment variable is not set"
    exit 1
fi

# Check if HUE_IP_ADDRESS environment variable is set
if [[ -z "${HUE_IP_ADDRESS}" ]]; then
    echo "Error: HUE_IP_ADDRESS environment variable is not set"
    exit 1
fi

# Create base64 encoded HUE_ID
export HUE_ID_BASE64=$(echo -n $HUE_ID | base64)

# Create base64 encoded HUE_IP_ADDRESS
export HUE_IP_ADDRESS_BASE64=$(echo -n $HUE_IP_ADDRESS | base64)

# Apply secret and other manifests
echo "Applying secret and manifests..."
sed -i -e "s#{HUE_ID}#${HUE_ID_BASE64}#g" manifests/secrets.yaml
sed -i -e "s#{HUE_IP_ADDRESS}#${HUE_IP_ADDRESS_BASE64}#g" manifests/secrets.yaml

# Apply Kubernetes manifests using kustomize
echo "Applying Kubernetes manifests..."
kubectl apply -k manifests/

# Clean up
echo "Cleaning up..."
rm manifests/base/config.yml

echo "Deployment complete!" 