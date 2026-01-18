#!/bin/bash

set -e

NAMESPACE="${NAMESPACE:-gohome}"
K8S_MANIFEST="${K8S_MANIFEST:-k8s/all.yaml}"

echo "Deploying to namespace: ${NAMESPACE}"

# Check for required environment variables
echo "Checking for required environment variables..."

# Check if owm-api-key-secret already exists
if kubectl get secret owm-api-key-secret --namespace=${NAMESPACE} &>/dev/null; then
    echo "✓ Found existing owm-api-key-secret in namespace ${NAMESPACE}"
    echo "  Skipping openweathermap-api-key secret creation"
    OWM_SECRET_EXISTS=true
else
    OWM_SECRET_EXISTS=false
    if [ -z "$OPENWEATHERMAP_API_KEY" ]; then
        echo "Error: OPENWEATHERMAP_API_KEY environment variable is required"
        echo "  (owm-api-key-secret not found in namespace ${NAMESPACE})"
        exit 1
    fi
fi

# Check if hue-color-looper-secrets already exists
if kubectl get secret hue-color-looper-secrets --namespace=${NAMESPACE} &>/dev/null; then
    echo "✓ Found existing hue-color-looper-secrets in namespace ${NAMESPACE}"
    # Try to extract hue-id and hue-ip-address from the existing secret (try common key names)
    HUE_ID=$(kubectl get secret hue-color-looper-secrets --namespace=${NAMESPACE} -o jsonpath='{.data.HUE_ID}' 2>/dev/null | base64 -d 2>/dev/null || \
             echo "")
    HUE_IP_ADDRESS=$(kubectl get secret hue-color-looper-secrets --namespace=${NAMESPACE} -o jsonpath='{.data.HUE_IP_ADDRESS}' 2>/dev/null | base64 -d 2>/dev/null || \
                     echo "")
    if [ -n "$HUE_ID" ] && [ -n "$HUE_IP_ADDRESS" ]; then
        echo "  Using Hue credentials from existing secret"
        HUE_SECRET_EXISTS=false  # We'll create hue-credentials from the existing one
    else
        echo "  Could not extract Hue credentials from hue-color-looper-secrets"
        HUE_SECRET_EXISTS=true
    fi
else
    HUE_SECRET_EXISTS=false
    if [ -z "$HUE_ID" ]; then
        echo "Error: HUE_ID environment variable is required"
        echo "  (hue-color-looper-secrets not found in namespace ${NAMESPACE})"
        exit 1
    fi
    if [ -z "$HUE_IP_ADDRESS" ]; then
        echo "Error: HUE_IP_ADDRESS environment variable is required"
        echo "  (hue-color-looper-secrets not found in namespace ${NAMESPACE})"
        exit 1
    fi
fi

echo "✓ All required environment variables are set"

# Create openweathermap-api-key secret only if owm-api-key-secret doesn't exist
if [ "$OWM_SECRET_EXISTS" = false ]; then
    echo "Creating openweathermap-api-key secret..."
    kubectl create secret generic openweathermap-api-key \
        --from-literal=api-key="${OPENWEATHERMAP_API_KEY}" \
        --namespace=${NAMESPACE} \
        --dry-run=client -o yaml | kubectl apply -f -
    echo "✓ openweathermap-api-key secret created/updated"
fi

# Create hue-credentials secret
if [ "$HUE_SECRET_EXISTS" = false ] && [ -n "$HUE_ID" ] && [ -n "$HUE_IP_ADDRESS" ]; then
    echo "Creating hue-credentials secret..."
    kubectl create secret generic hue-credentials \
        --from-literal=hue-id="${HUE_ID}" \
        --from-literal=hue-ip-address="${HUE_IP_ADDRESS}" \
        --namespace=${NAMESPACE} \
        --dry-run=client -o yaml | kubectl apply -f -
    echo "✓ hue-credentials secret created/updated"
elif [ "$HUE_SECRET_EXISTS" = true ]; then
    echo "ℹ Skipping hue-credentials secret creation (hue-color-looper-secrets exists but credentials couldn't be extracted)"
fi

# Apply Kubernetes manifests
echo "Applying Kubernetes manifests from ${K8S_MANIFEST}..."
kubectl apply -f ${K8S_MANIFEST}
echo "✓ Kubernetes manifests applied"

echo ""
echo "Deployment complete!"
echo "Secrets and resources have been created/updated in namespace: ${NAMESPACE}"
