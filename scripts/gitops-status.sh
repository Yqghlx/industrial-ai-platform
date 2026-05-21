#!/bin/bash
# GitOps Status Check Script
# Industrial AI Platform - ArgoCD Deployment Status Monitor
# Version: 1.0.0

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
ARGOCD_SERVER="${ARGOCD_SERVER:-argocd.industrial-ai.example.com}"
ARGOCD_NAMESPACE="${ARGOCD_NAMESPACE:-argocd}"
APP_NAMESPACE="${APP_NAMESPACE:-industrial-ai}"
APP_NAME="${APP_NAME:-industrial-ai-platform}"

# Functions
log_info() {
    echo "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo "${RED}[✗]${NC} $1"
}

print_header() {
    echo ""
    echo "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo "${CYAN}$1${NC}"
    echo "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Check ArgoCD connectivity
check_argocd_connection() {
    if ! command -v argocd &> /dev/null; then
        log_error "ArgoCD CLI not installed"
        return 1
    fi
    
    if argocd version &> /dev/null; then
        log_success "ArgoCD CLI connected to $ARGOCD_SERVER"
        return 0
    else
        log_error "Cannot connect to ArgoCD server"
        return 1
    fi
}

# Get application status
get_app_status() {
    local status=$(argocd app get "$APP_NAME" -o json 2>/dev/null)
    
    if [[ -z "$status" ]]; then
        log_error "Cannot get application status"
        return 1
    fi
    
    echo "$status"
}

# Parse and display sync status
show_sync_status() {
    local app_json="$1"
    
    local sync_status=$(echo "$app_json" | jq -r '.status.sync.status')
    local revision=$(echo "$app_json" | jq -r '.status.sync.revision')
    local sync_time=$(echo "$app_json" | jq -r '.status.sync.startedAt')
    
    echo "┌─────────────────────────────────────────────────────┐"
    echo "│ Sync Status                                          │"
    echo "├─────────────────────────────────────────────────────┤"
    
    case "$sync_status" in
        Synced)
            echo "│ Status:      ${GREEN}Synced${NC}                                  │"
            ;;
        OutOfSync)
            echo "│ Status:      ${YELLOW}OutOfSync${NC}                               │"
            ;;
        Unknown)
            echo "│ Status:      ${RED}Unknown${NC}                                  │"
            ;;
        *)
            echo "│ Status:      ${RED}$sync_status${NC}                              │"
            ;;
    esac
    
    echo "│ Revision:    $revision              │"
    echo "│ Last Sync:   $sync_time      │"
    echo "└─────────────────────────────────────────────────────┘"
}

# Parse and display health status
show_health_status() {
    local app_json="$1"
    
    local health_status=$(echo "$app_json" | jq -r '.status.health.status')
    
    echo "┌─────────────────────────────────────────────────────┐"
    echo "│ Health Status                                         │"
    echo "├─────────────────────────────────────────────────────┤"
    
    case "$health_status" in
        Healthy)
            echo "│ Status:      ${GREEN}Healthy${NC}                                 │"
            ;;
        Degraded)
            echo "│ Status:      ${RED}Degraded${NC}                                │"
            ;;
        Progressing)
            echo "│ Status:      ${YELLOW}Progressing${NC}                            │"
            ;;
        Suspended)
            echo "│ Status:      ${BLUE}Suspended${NC}                               │"
            ;;
        Missing)
            echo "│ Status:      ${RED}Missing${NC}                                 │"
            ;;
        *)
            echo "│ Status:      ${RED}$health_status${NC}                           │"
            ;;
    esac
    
    echo "└─────────────────────────────────────────────────────┘"
    
    # Show resource health details
    local resources=$(echo "$app_json" | jq -r '.status.resources[]')
    
    if [[ -n "$resources" ]]; then
        echo ""
        echo "Resource Health Details:"
        echo "$app_json" | jq -r '.status.resources[] | "  \(.kind)/\(.name): \(.health.status)"' | while read line; do
            local status=$(echo "$line" | awk -F': ' '{print $2}')
            local resource=$(echo "$line" | awk -F': ' '{print $1}')
            
            case "$status" in
                Healthy)
                    echo "  ${GREEN}✓${NC} $resource: $status"
                    ;;
                Degraded|Missing)
                    echo "  ${RED}✗${NC} $resource: $status"
                    ;;
                Progressing)
                    echo "  ${YELLOW}◐${NC} $resource: $status"
                    ;;
                *)
                    echo "  ${BLUE}?${NC} $resource: $status"
                    ;;
            esac
        done
    fi
}

# Show deployment resources
show_k8s_resources() {
    print_header "Kubernetes Resources"
    
    echo "Pods:"
    kubectl get pods -n "$APP_NAMESPACE" --no-headers | while read line; do
        local name=$(echo "$line" | awk '{print $1}')
        local ready=$(echo "$line" | awk '{print $2}')
        local status=$(echo "$line" | awk '{print $3}')
        
        if [[ "$status" == "Running" ]]; then
            echo "  ${GREEN}✓${NC} $name ($ready) Running"
        elif [[ "$status" == "Pending" ]]; then
            echo "  ${YELLOW}◐${NC} $name ($ready) Pending"
        else
            echo "  ${RED}✗${NC} $name ($ready) $status"
        fi
    done
    
    echo ""
    echo "Services:"
    kubectl get services -n "$APP_NAMESPACE" --no-headers | while read line; do
        local name=$(echo "$line" | awk '{print $1}')
        local type=$(echo "$line" | awk '{print $2}')
        local ports=$(echo "$line" | awk '{print $5}')
        echo "  ${GREEN}✓${NC} $name ($type) $ports"
    done
    
    echo ""
    echo "Deployments:"
    kubectl get deployments -n "$APP_NAMESPACE" --no-headers | while read line; do
        local name=$(echo "$line" | awk '{print $1}')
        local ready=$(echo "$line" | awk '{print $2}')
        local available=$(echo "$line" | awk '{print $4}')
        echo "  ${GREEN}✓${NC} $name ($ready/$available available)"
    done
}

# Show recent sync operations
show_operations() {
    print_header "Recent Operations"
    
    argocd app history "$APP_NAME" | tail -10
}

# Show diff if OutOfSync
show_diff() {
    local app_json="$1"
    local sync_status=$(echo "$app_json" | jq -r '.status.sync.status')
    
    if [[ "$sync_status" == "OutOfSync" ]]; then
        print_header "Configuration Diff (OutOfSync)"
        argocd app diff "$APP_NAME" 2>/dev/null || log_info "No differences to show"
    fi
}

# Check application endpoints
check_endpoints() {
    print_header "Endpoint Health Checks"
    
    # Backend health
    echo "Backend Health Check:"
    if curl -sf "https://backend.industrial-ai.example.com/health" 2>/dev/null; then
        log_success "Backend health endpoint OK"
        local response=$(curl -s "https://backend.industrial-ai.example.com/health")
        echo "  Response: $response"
    else
        log_error "Backend health endpoint failed"
    fi
    
    echo ""
    
    # Frontend health
    echo "Frontend Health Check:"
    if curl -sf "https://frontend.industrial-ai.example.com/" -o /dev/null 2>/dev/null; then
        log_success "Frontend accessible"
    else
        log_error "Frontend not accessible"
    fi
    
    echo ""
    
    # API endpoints
    echo "API Endpoint Tests:"
    
    local endpoints=(
        "/api/v1/health"
        "/api/v1/devices"
        "/api/v1/telemetry/latest"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if curl -sf "https://backend.industrial-ai.example.com$endpoint" -o /dev/null 2>/dev/null; then
            log_success "$endpoint OK"
        else
            log_error "$endpoint failed"
        fi
    done
}

# Show resource usage
show_resource_usage() {
    print_header "Resource Usage"
    
    echo "Pod Resource Usage:"
    kubectl top pods -n "$APP_NAMESPACE" 2>/dev/null || log_warning "Metrics server not available"
    
    echo ""
    echo "Node Resource Usage:"
    kubectl top nodes 2>/dev/null || log_warning "Metrics server not available"
}

# Show HPA status
show_hpa_status() {
    print_header "HPA (Horizontal Pod Autoscaler) Status"
    
    kubectl get hpa -n "$APP_NAMESPACE" --no-headers | while read line; do
        local name=$(echo "$line" | awk '{print $1}')
        local reference=$(echo "$line" | awk '{print $2}')
        local targets=$(echo "$line" | awk '{print $3}')
        local current=$(echo "$line" | awk '{print $6}')
        local replicas=$(echo "$line" | awk '{print $7}')
        
        echo "┌─────────────────────────────────────────────────────┐"
        echo "│ $name                                                │"
        echo "├─────────────────────────────────────────────────────┤"
        echo "│ Reference:   $reference          │"
        echo "│ Targets:     $targets              │"
        echo "│ Current:     $current               │"
        echo "│ Replicas:    $replicas              │"
        echo "└─────────────────────────────────────────────────────┘"
    done
}

# Generate status summary
generate_summary() {
    local app_json="$1"
    
    local sync_status=$(echo "$app_json" | jq -r '.status.sync.status')
    local health_status=$(echo "$app_json" | jq -r '.status.health.status')
    
    print_header "Status Summary"
    
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                    Industrial AI Platform                    │"
    echo "├─────────────────────────────────────────────────────────────┤"
    
    # Overall status
    if [[ "$sync_status" == "Synced" && "$health_status" == "Healthy" ]]; then
        echo "│ Overall Status: ${GREEN}OPERATIONAL${NC}                           │"
    elif [[ "$sync_status" == "OutOfSync" ]]; then
        echo "│ Overall Status: ${YELLOW}NEEDS SYNC${NC}                             │"
    elif [[ "$health_status" == "Degraded" ]]; then
        echo "│ Overall Status: ${RED}DEGRADED${NC}                               │"
    else
        echo "│ Overall Status: ${YELLOW}CHECK REQUIRED${NC}                        │"
    fi
    
    echo "├─────────────────────────────────────────────────────────────┤"
    echo "│ Sync:      $sync_status                          │"
    echo "│ Health:    $health_status                          │"
    echo "│ ArgoCD:    $ARGOCD_SERVER              │"
    echo "│ Namespace: $APP_NAMESPACE                  │"
    echo "└─────────────────────────────────────────────────────────────┘"
}

# Quick status check
quick_status() {
    local app_json=$(get_app_status)
    
    local sync=$(echo "$app_json" | jq -r '.status.sync.status')
    local health=$(echo "$app_json" | jq -r '.status.health.status')
    
    if [[ "$sync" == "Synced" && "$health" == "Healthy" ]]; then
        echo "${GREEN}✓ All systems operational${NC}"
        return 0
    else
        echo "${YELLOW}! Status: Sync=$sync, Health=$health${NC}"
        return 1
    fi
}

# Watch status (continuous monitoring)
watch_status() {
    local interval="${1:-10}"
    
    log_info "Watching status (interval: $interval seconds, press Ctrl+C to stop)"
    
    while true; do
        clear
        echo "Industrial AI Platform Status Monitor (Updated: $(date))"
        echo ""
        quick_status
        sleep "$interval"
    done
}

# Export status as JSON
export_status() {
    local app_json=$(get_app_status)
    
    local output="${1:-status.json}"
    
    echo "$app_json" > "$output"
    log_success "Status exported to $output"
}

# Show usage
show_usage() {
    echo "Industrial AI Platform GitOps Status Script"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  status        Show full status report"
    echo "  quick         Quick status check (sync + health)"
    echo "  sync          Show sync status details"
    echo "  health        Show health status details"
    echo "  resources     Show Kubernetes resources"
    echo "  operations    Show recent operations history"
    echo "  diff          Show configuration differences"
    echo "  endpoints     Check endpoint health"
    echo "  usage         Show resource usage"
    echo "  hpa           Show HPA status"
    echo "  watch [n]     Continuous monitoring (interval n seconds)"
    echo "  export [file] Export status to JSON file"
    echo "  summary       Generate status summary"
    echo ""
    echo "Examples:"
    echo "  $0 status             Full status report"
    echo "  $0 quick              Quick check"
    echo "  $0 watch 5            Watch status every 5 seconds"
    echo "  $0 export status.json Export status to file"
    echo ""
    echo "Environment Variables:"
    echo "  ARGOCD_SERVER          ArgoCD server URL"
    echo "  ARGOCD_NAMESPACE       ArgoCD namespace"
    echo "  APP_NAMESPACE          Application namespace"
    echo "  APP_NAME               Application name"
}

# Main
case "${1:-}" in
    status)
        check_argocd_connection
        local app_json=$(get_app_status)
        generate_summary "$app_json"
        show_sync_status "$app_json"
        show_health_status "$app_json"
        show_k8s_resources
        ;;
    quick)
        check_argocd_connection || exit 1
        quick_status
        ;;
    sync)
        check_argocd_connection
        local app_json=$(get_app_status)
        show_sync_status "$app_json"
        ;;
    health)
        check_argocd_connection
        local app_json=$(get_app_status)
        show_health_status "$app_json"
        ;;
    resources)
        show_k8s_resources
        ;;
    operations)
        check_argocd_connection
        show_operations
        ;;
    diff)
        check_argocd_connection
        local app_json=$(get_app_status)
        show_diff "$app_json"
        ;;
    endpoints)
        check_endpoints
        ;;
    usage)
        show_resource_usage
        ;;
    hpa)
        show_hpa_status
        ;;
    watch)
        watch_status "${2:-10}"
        ;;
    export)
        check_argocd_connection
        export_status "${2:-status.json}"
        ;;
    summary)
        check_argocd_connection
        local app_json=$(get_app_status)
        generate_summary "$app_json"
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        if [[ -n "${1:-}" ]]; then
            log_error "Unknown command: $1"
        fi
        show_usage
        exit 1
        ;;
esac