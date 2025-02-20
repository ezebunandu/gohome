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
HUE_ID_BASE64=$(echo -n "${HUE_ID}" | base64)

# Read the config file
CONFIG_CONTENT=$(cat manifests/base/config.yml)

# Apply both secret and configmap together
echo "Applying secret and configmap..."
(
    sed "s/\${HUE_ID_BASE64}/${HUE_ID_BASE64}/" manifests/secrets.yaml
    echo "---"
    sed "s|\${CONFIG_CONTENT}|${CONFIG_CONTENT}|" manifests/configmap.yaml
) | kubectl apply -f -

# Apply the kubernetes manifests
echo "Applying Kubernetes manifests..."
kubectl apply -k manifests/

# Clean up
echo "Cleaning up..."
rm manifests/base/config.yml

echo "Deployment complete!" 