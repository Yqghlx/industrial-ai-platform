#!/bin/bash
# SEC-003: Secret Rotation Script
# This script generates new secrets and updates Kubernetes secrets
# It should be run as a CronJob in Kubernetes

set -euo pipefail

NAMESPACE="${NAMESPACE:-industrial-ai}"
SECRET_NAME="${SECRET_NAME:-industrial-ai-secrets}"
ROTATION_DAYS="${ROTATION_DAYS:-90}"

log() {
    echo "[secret-rotation] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# Generate a new JWT secret (256-bit random)
generate_jwt_secret() {
    openssl rand -base64 32
}

# Generate a new encryption key (AES-256)
generate_encryption_key() {
    openssl rand -base64 32
}

# Generate a new Redis password
generate_redis_password() {
    openssl rand -base64 24
}

# Generate a new database password
generate_db_password() {
    openssl rand -base64 24
}

# Update Kubernetes secret
update_k8s_secret() {
    local key_name="$1"
    local value="$2"
    
    # Base64 encode the value for Kubernetes
    encoded_value=$(echo -n "$value" | base64)
    
    # Check if secret exists
    if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" >/dev/null 2>&1; then
        # Update existing secret
        kubectl patch secret "$SECRET_NAME" -n "$NAMESPACE" \
            --type=json \
            -p='[{"op": "replace", "path": "/data/'"$key_name"'", "value": "'"$encoded_value"'"}]'
        log "Updated $key_name in secret $SECRET_NAME"
    else
        log "ERROR: Secret $SECRET_NAME not found in namespace $NAMESPACE"
        return 1
    fi
}

# Trigger pod restart to pick up new secrets
restart_pods() {
    local deployment="$1"
    
    log "Restarting deployment $deployment to pick up new secrets..."
    kubectl rollout restart deployment "$deployment" -n "$NAMESPACE"
    
    # Wait for rollout to complete
    kubectl rollout status deployment "$deployment" -n "$NAMESPACE" --timeout=300s
    log "Deployment $deployment restarted successfully"
}

# Main rotation logic
main() {
    log "Starting secret rotation process..."
    
    # Check required environment variables
    if [ -z "${KUBECTL_AVAILABLE:-}" ]; then
        # Check if kubectl is available
        if ! command -v kubectl >/dev/null 2>&1; then
            log "ERROR: kubectl not found. This script must run in a Kubernetes environment."
            exit 1
        fi
    fi
    
    # Track which secrets were rotated
    rotated_secrets=""
    
    # Rotate JWT secret
    if [ "${ROTATE_JWT:-true}" = "true" ]; then
        log "Rotating JWT secret..."
        new_jwt=$(generate_jwt_secret)
        update_k8s_secret "jwt-secret" "$new_jwt"
        rotated_secrets="$rotated_secrets jwt-secret"
    fi
    
    # Rotate encryption key
    if [ "${ROTATE_ENCRYPTION:-true}" = "true" ]; then
        log "Rotating encryption key..."
        new_encryption=$(generate_encryption_key)
        update_k8s_secret "encryption-key" "$new_encryption"
        rotated_secrets="$rotated_secrets encryption-key"
    fi
    
    # Rotate Redis password
    if [ "${ROTATE_REDIS:-true}" = "true" ]; then
        log "Rotating Redis password..."
        new_redis=$(generate_redis_password)
        update_k8s_secret "redis-password" "$new_redis"
        rotated_secrets="$rotated_secrets redis-password"
        
        # Need to update Redis deployment too
        if [ "${RESTART_REDIS:-true}" = "true" ]; then
            restart_pods "redis" || true
        fi
    fi
    
    # Rotate database password (optional, more complex)
    if [ "${ROTATE_DATABASE:-false}" = "true" ]; then
        log "Rotating database password..."
        new_db_password=$(generate_db_password)
        update_k8s_secret "database-url" "postgres://appuser:$new_db_password@postgres:5432/industrial_ai?sslmode=require"
        rotated_secrets="$rotated_secrets database-url"
        
        # Note: Database password rotation requires coordination with PostgreSQL
        # This only updates the K8s secret, actual DB password change needs manual steps
        log "WARNING: Database URL secret updated, but database password must be changed manually"
    fi
    
    # Restart backend to pick up new secrets
    if [ "${ROTATE_JWT:-true}" = "true" ] || [ "${ROTATE_ENCRYPTION:-true}" = "true" ]; then
        restart_pods "backend"
    fi
    
    log "Secret rotation completed successfully!"
    log "Rotated secrets: $rotated_secrets"
    
    # Output summary for monitoring
    echo "ROTATION_SUMMARY: rotated=$rotated_secrets timestamp=$(date '+%Y-%m-%dT%H:%M:%SZ')"
}

# Run main function
main "$@"