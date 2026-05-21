#!/bin/bash
# Kubernetes Configuration Validation Script
# Industrial AI Platform - DEPLOY-003
# Validates all K8s configuration files

set -e

PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$(dirname "$0")/../../.." && pwd)}"
K8S_DIR="$PROJECT_ROOT/infra/k8s"
VALIDATION_RESULTS=()

echo "========================================"
echo "K8s Configuration Validation"
echo "Industrial AI Platform"
echo "========================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check file exists
check_file() {
    local file="$1"
    local description="$2"
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $description: $file"
        VALIDATION_RESULTS+=("PASS:$description")
        return 0
    else
        echo -e "${RED}✗${NC} $description: $file (NOT FOUND)"
        VALIDATION_RESULTS+=("FAIL:$description")
        return 1
    fi
}

# Function to validate YAML syntax
validate_yaml() {
    local file="$1"
    local description="$2"
    
    if command -v yq &> /dev/null; then
        if yq eval '.' "$file" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} $description: YAML valid"
            return 0
        else
            echo -e "${RED}✗${NC} $description: YAML syntax error"
            return 1
        fi
    elif command -v python3 &> /dev/null; then
        # Use yaml.safe_load_all for multi-document YAML
        if python3 -c "
import yaml
import sys
try:
    with open('$file') as f:
        for doc in yaml.safe_load_all(f):
            pass
    print('valid')
except Exception as e:
    print(f'error: {e}')
    sys.exit(1)
" 2>/dev/null; then
            echo -e "${GREEN}✓${NC} $description: YAML valid"
            return 0
        else
            echo -e "${RED}✗${NC} $description: YAML syntax error"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠${NC} $description: YAML validator not available (yq/python3)"
        return 0
    fi
}

# Function to check required fields in Deployment
check_deployment() {
    local file="$1"
    
    echo ""
    echo "Checking Deployment Configuration: $file"
    
    # Check replicas
    if grep -q "replicas:" "$file"; then
        local replicas=$(grep "replicas:" "$file" | head -1 | awk '{print $2}')
        if [ "$replicas" -ge 2 ]; then
            echo -e "${GREEN}✓${NC} Replicas: $replicas (>= 2, HA satisfied)"
        else
            echo -e "${YELLOW}⚠${NC} Replicas: $replicas (< 2, consider increasing for HA)"
        fi
    fi
    
    # Check resource limits
    if grep -q "limits:" "$file" && grep -q "requests:" "$file"; then
        echo -e "${GREEN}✓${NC} Resource limits and requests configured"
    else
        echo -e "${YELLOW}⚠${NC} Resource limits/requests may not be configured"
    fi
    
    # Check health probes
    if grep -q "livenessProbe:" "$file"; then
        echo -e "${GREEN}✓${NC} Liveness probe configured"
    else
        echo -e "${RED}✗${NC} Liveness probe NOT configured"
    fi
    
    if grep -q "readinessProbe:" "$file"; then
        echo -e "${GREEN}✓${NC} Readiness probe configured"
    else
        echo -e "${RED}✗${NC} Readiness probe NOT configured"
    fi
    
    # Check startup probe
    if grep -q "startupProbe:" "$file"; then
        echo -e "${GREEN}✓${NC} Startup probe configured"
    else
        echo -e "${YELLOW}⚠${NC} Startup probe not configured (optional but recommended)"
    fi
    
    # Check security context
    if grep -q "securityContext:" "$file" || grep -q "runAsNonRoot:" "$file" || grep -q "runAsUser:" "$file"; then
        echo -e "${GREEN}✓${NC} Security context configured"
    else
        echo -e "${YELLOW}⚠${NC} Security context not explicitly configured"
    fi
}

# Function to check HPA configuration
check_hpa() {
    local file="$1"
    
    echo ""
    echo "Checking HPA Configuration: $file"
    
    # Check min/max replicas
    if grep -q "minReplicas:" "$file" && grep -q "maxReplicas:" "$file"; then
        echo -e "${GREEN}✓${NC} Min/max replicas configured"
    else
        echo -e "${RED}✗${NC} Min/max replicas NOT configured"
    fi
    
    # Check metrics
    if grep -q "metrics:" "$file"; then
        echo -e "${GREEN}✓${NC} Scaling metrics configured"
    else
        echo -e "${RED}✗${NC} Scaling metrics NOT configured"
    fi
    
    # Check behavior
    if grep -q "behavior:" "$file"; then
        echo -e "${GREEN}✓${NC} Scaling behavior configured"
    else
        echo -e "${YELLOW}⚠${NC} Scaling behavior not configured (optional)"
    fi
}

# Function to check Ingress configuration
check_ingress() {
    local file="$1"
    
    echo ""
    echo "Checking Ingress Configuration: $file"
    
    # Check TLS
    if grep -q "tls:" "$file"; then
        echo -e "${GREEN}✓${NC} TLS configured"
    else
        echo -e "${RED}✗${NC} TLS NOT configured (security risk)"
    fi
    
    # Check HTTPS redirect
    if grep -q "ssl-redirect:" "$file" || grep -q "force-ssl-redirect:" "$file"; then
        echo -e "${GREEN}✓${NC} HTTPS redirect configured"
    else
        echo -e "${YELLOW}⚠${NC} HTTPS redirect may not be configured"
    fi
    
    # Check HSTS
    if grep -q "Strict-Transport-Security" "$file"; then
        echo -e "${GREEN}✓${NC} HSTS header configured"
    else
        echo -e "${YELLOW}⚠${NC} HSTS header not configured"
    fi
    
    # Check WebSocket support
    if grep -q "proxy-read-timeout" "$file" || grep -q "websocket" "$file"; then
        echo -e "${GREEN}✓${NC} WebSocket support configured"
    else
        echo -e "${YELLOW}⚠${NC} WebSocket support may need configuration"
    fi
}

# Function to check Secret configuration
check_secret() {
    local file="$1"
    
    echo ""
    echo "Checking Secret Configuration: $file"
    
    # Check type
    if grep -q "type: Opaque" "$file" || grep -q "type: kubernetes.io/tls" "$file"; then
        echo -e "${GREEN}✓${NC} Secret type properly set"
    else
        echo -e "${YELLOW}⚠${NC} Secret type not specified"
    fi
    
    # Check for placeholder values
    if grep -q "CHANGE_ME" "$file" || grep -q "<base64-encoded" "$file"; then
        echo -e "${YELLOW}⚠${NC} Placeholder values detected - replace before deployment"
    else
        echo -e "${GREEN}✓${NC} No placeholder values detected"
    fi
    
    # Check RBAC for secrets
    if grep -q "Role:" "$file" || grep -q "RoleBinding:" "$file"; then
        echo -e "${GREEN}✓${NC} RBAC configuration present"
    else
        echo -e "${YELLOW}⚠${NC} RBAC may need separate configuration"
    fi
}

# ==========================================
# Main Validation Process
# ==========================================

echo ""
echo "=== File Existence Checks ==="

# Core deployment files
check_file "$K8S_DIR/deployment.yaml" "Core Deployment"
check_file "$K8S_DIR/deployment-health.yaml" "Health-enhanced Deployment"
check_file "$K8S_DIR/hpa.yaml" "Horizontal Pod Autoscaler"
check_file "$K8S_DIR/ingress-tls.yaml" "Ingress with TLS"
check_file "$K8S_DIR/secrets.yaml" "Secrets Configuration"
check_file "$K8S_DIR/prometheus-adapter.yaml" "Prometheus Adapter"

# Monitoring files
check_file "$K8S_DIR/monitoring/prometheus-grafana.yaml" "Prometheus/Grafana Deployment"

echo ""
echo "=== YAML Syntax Validation ==="

# Validate YAML syntax for all files
for file in "$K8S_DIR"/*.yaml "$K8S_DIR"/monitoring/*.yaml; do
    if [ -f "$file" ]; then
        validate_yaml "$file" "$(basename $file)"
    fi
done

echo ""
echo "=== Configuration Content Checks ==="

# Check Deployment
if [ -f "$K8S_DIR/deployment.yaml" ]; then
    check_deployment "$K8S_DIR/deployment.yaml"
fi

if [ -f "$K8S_DIR/deployment-health.yaml" ]; then
    check_deployment "$K8S_DIR/deployment-health.yaml"
fi

# Check HPA
if [ -f "$K8S_DIR/hpa.yaml" ]; then
    check_hpa "$K8S_DIR/hpa.yaml"
fi

# Check Ingress
if [ -f "$K8S_DIR/ingress-tls.yaml" ]; then
    check_ingress "$K8S_DIR/ingress-tls.yaml"
fi

# Check Secrets
if [ -f "$K8S_DIR/secrets.yaml" ]; then
    check_secret "$K8S_DIR/secrets.yaml"
fi

echo ""
echo "========================================"
echo "Validation Summary"
echo "========================================"

# Count results
PASSES=0
FAILS=0
WARNINGS=0

for result in "${VALIDATION_RESULTS[@]}"; do
    if [[ $result == PASS:* ]]; then
        PASSES=$((PASSES + 1))
    elif [[ $result == FAIL:* ]]; then
        FAILS=$((FAILS + 1))
    fi
done

echo ""
echo -e "Passed: ${GREEN}$PASSES${NC}"
echo -e "Failed: ${RED}$FAILS${NC}"
echo ""

if [ $FAILS -gt 0 ]; then
    echo -e "${RED}Some validations failed. Please review the issues above.${NC}"
    exit 1
else
    echo -e "${GREEN}All essential validations passed!${NC}"
    exit 0
fi