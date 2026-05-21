#!/bin/bash
# GitOps Rollback Script
# Industrial AI Platform - ArgoCD Rollback Automation
# Version: 1.0.0

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ARGOCD_SERVER="${ARGOCD_SERVER:-argocd.industrial-ai.example.com}"
ARGOCD_NAMESPACE="${ARGOCD_NAMESPACE:-argocd}"
APP_NAMESPACE="${APP_NAMESPACE:-industrial-ai}"
APP_NAME="${APP_NAME:-industrial-ai-platform}"

# Rollback strategies
STRATEGY_IMMEDIATE="immediate"
STRATEGY_STAGED="staged"
STRATEGY_EMERGENCY="emergency"

# Functions
log_info() {
    echo "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

# Check ArgoCD CLI availability
check_argocd() {
    if ! command -v argocd &> /dev/null; then
        log_error "ArgoCD CLI not found. Please install: brew install argocd"
        exit 1
    fi
    
    if ! argocd context &> /dev/null; then
        log_error "Not logged in to ArgoCD. Run: argocd login $ARGOCD_SERVER"
        exit 1
    fi
}

# Get application history
get_history() {
    log_info "Fetching application history..."
    argocd app history "$APP_NAME" --namespace "$ARGOCD_NAMESPACE"
}

# Get current revision
get_current_revision() {
    argocd app get "$APP_NAME" --namespace "$ARGOCD_NAMESPACE" -o json | jq -r '.status.sync.revision'
}

# List available revisions
list_revisions() {
    log_info "Available revisions:"
    argocd app history "$APP_NAME" --namespace "$ARGOCD_NAMESPACE" | tail -n +2 | head -10
}

# Health check before rollback
pre_rollback_health_check() {
    log_info "Running pre-rollback health check..."
    
    # Check backend health
    if curl -sf "https://backend.industrial-ai.example.com/health" > /dev/null 2>&1; then
        log_success "Backend is healthy"
    else
        log_warning "Backend health check failed - proceeding with rollback"
    fi
    
    # Check frontend health
    if curl -sf "https://frontend.industrial-ai.example.com/" > /dev/null 2>&1; then
        log_success "Frontend is healthy"
    else
        log_warning "Frontend health check failed - proceeding with rollback"
    fi
}

# Create backup before rollback
create_backup() {
    log_info "Creating backup before rollback..."
    
    BACKUP_DIR="/tmp/rollback-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup current K8s resources
    kubectl get all -n "$APP_NAMESPACE" -o yaml > "$BACKUP_DIR/k8s-resources.yaml"
    kubectl get configmaps -n "$APP_NAMESPACE" -o yaml > "$BACKUP_DIR/configmaps.yaml"
    kubectl get secrets -n "$APP_NAMESPACE" -o yaml > "$BACKUP_DIR/secrets.yaml"
    
    # Backup current ArgoCD application state
    argocd app get "$APP_NAME" -o yaml > "$BACKUP_DIR/argocd-app-state.yaml"
    
    log_success "Backup created in $BACKUP_DIR"
    echo "$BACKUP_DIR"
}

# Execute rollback
execute_rollback() {
    local revision="$1"
    local force="$2"
    
    log_info "Rolling back to revision $revision..."
    
    if [[ "$force" == "true" ]]; then
        argocd app rollback "$APP_NAME" --revision "$revision" --force
    else
        argocd app rollback "$APP_NAME" --revision "$revision"
    fi
    
    log_success "Rollback command executed"
}

# Post-rollback health check
post_rollback_health_check() {
    log_info "Running post-rollback health check..."
    
    # Wait for pods to stabilize
    sleep 30
    
    # Check deployment status
    local ready_pods=$(kubectl get pods -n "$APP_NAMESPACE" -l app=backend --field-selector=status.phase=Running -o json | jq -r '.items | length')
    
    if [[ "$ready_pods" -ge 2 ]]; then
        log_success "Backend pods are running ($ready_pods)"
    else
        log_error "Backend pods are not ready"
        return 1
    fi
    
    # Check frontend
    ready_pods=$(kubectl get pods -n "$APP_NAMESPACE" -l app=frontend --field-selector=status.phase=Running -o json | jq -r '.items | length')
    
    if [[ "$ready_pods" -ge 2 ]]; then
        log_success "Frontend pods are running ($ready_pods)"
    else
        log_error "Frontend pods are not ready"
        return 1
    fi
    
    # Check service endpoints
    if curl -sf "https://backend.industrial-ai.example.com/health" > /dev/null 2>&1; then
        log_success "Backend health endpoint is accessible"
    else
        log_warning "Backend health endpoint not accessible yet"
    fi
    
    return 0
}

# Send notification
send_notification() {
    local status="$1"
    local revision="$2"
    local channel="${SLACK_CHANNEL:-#devops-alerts}"
    
    log_info "Sending notification..."
    
    local message="{
        \"text\": \"Rollback executed for Industrial AI Platform\",
        \"attachments\": [{
            \"color\": \"$([[ \"$status\" == \"success\" ]] && echo 'good' || echo 'danger')\",
            \"fields\": [
                {\"title\": \"Application\", \"value\": \"$APP_NAME\", \"short\": true},
                {\"title\": \"Revision\", \"value\": \"$revision\", \"short\": true},
                {\"title\": \"Status\", \"value\": \"$status\", \"short\": true},
                {\"title\": \"Timestamp\", \"value\": \"$(date -Iseconds)\", \"short\": true}
            ]
        }]
    }"
    
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        curl -s -X POST -H 'Content-type: application/json' \
            --data "$message" \
            "$SLACK_WEBHOOK_URL"
        log_success "Notification sent to Slack"
    else
        log_warning "SLACK_WEBHOOK_URL not set - skipping notification"
    fi
}

# Immediate rollback (fast, skip validation)
rollback_immediate() {
    local revision="$1"
    
    log_warning "Executing IMMEDIATE rollback (fast recovery)"
    
    execute_rollback "$revision" "false"
    
    # Brief health check
    sleep 10
    kubectl get pods -n "$APP_NAMESPACE" --no-headers | grep -c Running || 0
    
    send_notification "executed" "$revision"
    
    log_success "Immediate rollback completed"
}

# Staged rollback (safer, with validation)
rollback_staged() {
    local revision="$1"
    
    log_info "Executing STAGED rollback (with validation)"
    
    # Step 1: Create backup
    local backup_dir=$(create_backup)
    
    # Step 2: Freeze new deployments
    log_info "Freezing deployments..."
    argocd app set "$APP_NAME" --sync-policy none
    
    # Step 3: Execute rollback
    execute_rollback "$revision" "false"
    
    # Step 4: Wait for stabilization
    log_info "Waiting for pods to stabilize..."
    sleep 60
    
    # Step 5: Run validation tests
    log_info "Running validation tests..."
    scripts/test-api.sh || log_warning "Some tests failed"
    
    # Step 6: Health check
    if post_rollback_health_check; then
        send_notification "success" "$revision"
        log_success "Staged rollback completed successfully"
    else
        send_notification "failed" "$revision"
        log_error "Staged rollback health check failed"
        return 1
    fi
    
    # Step 7: Restore auto-sync
    log_info "Restoring auto-sync policy..."
    argocd app set "$APP_NAME" --sync-policy automated
}

# Emergency rollback (bypass all checks)
rollback_emergency() {
    local revision="$1"
    
    log_error "Executing EMERGENCY rollback (bypassing all checks)"
    
    # Force rollback immediately
    execute_rollback "$revision" "true"
    
    # Minimal health check (10 seconds)
    sleep 10
    kubectl get pods -n "$APP_NAMESPACE" --no-headers 2>&1 | head -5
    
    send_notification "emergency_executed" "$revision"
    
    log_warning "Emergency rollback executed - manual verification required"
}

# Main rollback function
rollback() {
    local revision="$1"
    local strategy="${2:-$STRATEGY_STAGED}"
    
    log_info "Starting rollback to revision $revision using $strategy strategy"
    
    # Validate revision
    if ! argocd app history "$APP_NAME" | grep -q "$revision"; then
        log_error "Revision $revision not found in history"
        list_revisions
        exit 1
    fi
    
    case "$strategy" in
        "$STRATEGY_IMMEDIATE")
            rollback_immediate "$revision"
            ;;
        "$STRATEGY_STAGED")
            rollback_staged "$revision"
            ;;
        "$STRATEGY_EMERGENCY")
            rollback_emergency "$revision"
            ;;
        *)
            log_error "Unknown strategy: $strategy"
            echo "Available strategies: immediate, staged, emergency"
            exit 1
            ;;
    esac
}

# Git revert rollback (alternative approach)
git_revert() {
    local commit="$1"
    
    log_info "Executing Git revert for commit $commit"
    
    # Check git status
    if ! git diff-index --quiet HEAD --; then
        log_error "Git working directory has uncommitted changes"
        exit 1
    fi
    
    # Create revert commit
    git revert "$commit" --no-edit
    
    # Push to remote
    git push origin main
    
    log_success "Git revert pushed - ArgoCD will auto-sync"
    
    send_notification "git_revert" "$commit"
}

# Show usage
show_usage() {
    echo "Industrial AI Platform GitOps Rollback Script"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  history                     Show application history"
    echo "  revisions                   List available revisions"
    echo "  rollback <rev> [strategy]   Rollback to specific revision"
    echo "  git-revert <commit>         Rollback via git revert"
    echo "  backup                      Create backup of current state"
    echo "  health                      Run health check"
    echo ""
    echo "Strategies:"
    echo "  immediate    Fast recovery, skip validation (default for critical)"
    echo "  staged       Safe rollback with backup and validation (default)"
    echo "  emergency    Bypass all checks, fastest recovery"
    echo ""
    echo "Examples:"
    echo "  $0 rollback 5 staged"
    echo "  $0 rollback 3 emergency"
    echo "  $0 git-revert abc123"
    echo "  $0 history"
    echo ""
    echo "Environment Variables:"
    echo "  ARGOCD_SERVER          ArgoCD server URL"
    echo "  ARGOCD_NAMESPACE       ArgoCD namespace (default: argocd)"
    echo "  APP_NAMESPACE          Application namespace (default: industrial-ai)"
    echo "  APP_NAME               Application name (default: industrial-ai-platform)"
    echo "  SLACK_WEBHOOK_URL      Slack webhook for notifications"
    echo "  SLACK_CHANNEL          Slack channel for alerts"
}

# Main
case "${1:-}" in
    history)
        check_argocd
        get_history
        ;;
    revisions)
        check_argocd
        list_revisions
        ;;
    rollback)
        check_argocd
        
        if [[ -z "${2:-}" ]]; then
            log_error "Revision required"
            show_usage
            exit 1
        fi
        
        rollback "${2}" "${3:-staged}"
        ;;
    git-revert)
        if [[ -z "${2:-}" ]]; then
            log_error "Commit hash required"
            show_usage
            exit 1
        fi
        
        git_revert "${2}"
        ;;
    backup)
        create_backup
        ;;
    health)
        post_rollback_health_check
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        log_error "Unknown command: ${1:-}"
        show_usage
        exit 1
        ;;
esac