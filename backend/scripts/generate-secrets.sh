#!/bin/bash
# SEC-MED-06: Secure Secret Generation Script
# This script generates secrets for various services WITHOUT printing them to stdout.
# Secrets are written to files or environment configuration only.

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SECRETS_DIR="${PROJECT_ROOT}/.secrets"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print usage
print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --jwt          Generate JWT secret"
    echo "  --admin        Generate admin password"
    echo "  --api-key      Generate API key"
    echo "  --all          Generate all secrets"
    echo "  --output       Output directory for secrets (default: .secrets)"
    echo "  --env-file     Output to .env file format"
    echo "  --help         Show this help message"
    echo ""
    echo "SECURITY: Generated secrets are NOT printed to console."
    echo "          They are written to files in the .secrets directory."
}

# Generate a secure random secret
generate_secret() {
    local length=${1:-32}
    # Use openssl for secure random generation
    openssl rand -hex "$length" 2>/dev/null || \
    # Fallback to /dev/urandom
    dd if=/dev/urandom bs=1 count="$length" 2>/dev/null | xxd -p | tr -d '\n'
}

# Write secret to file (securely)
write_secret() {
    local name="$1"
    local value="$2"
    local file="${SECRETS_DIR}/${name}.txt"
    
    # Create secrets directory with restricted permissions
    mkdir -p "$SECRETS_DIR"
    chmod 700 "$SECRETS_DIR"
    
    # Write secret to file
    echo "$value" > "$file"
    chmod 600 "$file"
    
    echo -e "${GREEN}✓${NC} Generated ${name} (saved to ${file})"
}

# Write to .env file format
write_to_env_file() {
    local env_file="${PROJECT_ROOT}/.env.generated"
    local jwt_secret="$1"
    local admin_password="$2"
    local api_key="$3"
    
    cat > "$env_file" << EOF
# Generated Secrets - SEC-MED-06
# DO NOT print these secrets to console or commit to version control
# Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")

JWT_SECRET=${jwt_secret}
ADMIN_PASSWORD=${admin_password}
DEVICE_API_KEY=${api_key}
EOF
    
    chmod 600 "$env_file"
    echo -e "${GREEN}✓${NC} Generated .env.generated file with all secrets"
    echo "  Location: ${env_file}"
}

# Main script logic
main() {
    local generate_jwt=false
    local generate_admin=false
    local generate_api_key=false
    local generate_all=false
    local use_env_file=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --jwt)
                generate_jwt=true
                shift
                ;;
            --admin)
                generate_admin=true
                shift
                ;;
            --api-key)
                generate_api_key=true
                shift
                ;;
            --all)
                generate_all=true
                shift
                ;;
            --env-file)
                use_env_file=true
                shift
                ;;
            --output)
                SECRETS_DIR="$2"
                shift 2
                ;;
            --help|-h)
                print_usage
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                print_usage
                exit 1
                ;;
        esac
    done
    
    # If no options specified, generate all by default
    if [[ "$generate_jwt" == "false" && "$generate_admin" == "false" && "$generate_api_key" == "false" ]]; then
        generate_all=true
    fi
    
    echo -e "${YELLOW}Generating secrets...${NC}"
    echo "Security: Secrets will NOT be printed to console"
    echo ""
    
    local jwt_secret=""
    local admin_password=""
    local api_key=""
    
    if [[ "$generate_all" == "true" || "$generate_jwt" == "true" ]]; then
        jwt_secret=$(generate_secret 32)
        if [[ "$use_env_file" == "false" ]]; then
            write_secret "JWT_SECRET" "$jwt_secret"
        fi
    fi
    
    if [[ "$generate_all" == "true" || "$generate_admin" == "true" ]]; then
        admin_password=$(generate_secret 16)
        if [[ "$use_env_file" == "false" ]]; then
            write_secret "ADMIN_PASSWORD" "$admin_password"
        fi
    fi
    
    if [[ "$generate_all" == "true" || "$generate_api_key" == "true" ]]; then
        api_key=$(generate_secret 24)
        if [[ "$use_env_file" == "false" ]]; then
            write_secret "DEVICE_API_KEY" "$api_key"
        fi
    fi
    
    # Write to .env file if requested
    if [[ "$use_env_file" == "true" ]]; then
        write_to_env_file "$jwt_secret" "$admin_password" "$api_key"
    fi
    
    echo ""
    echo -e "${GREEN}=== Security Reminder ===${NC}"
    echo "1. NEVER print secrets to stdout or logs"
    echo "2. NEVER commit .secrets/ directory to version control"
    echo "3. Add '.secrets/' to your .gitignore file"
    echo "4. Rotate secrets regularly (recommend: 30 days)"
    echo "5. Use environment variables or secret management systems in production"
    echo ""
    
    # Security: Return success without exposing secrets
    exit 0
}

# Run main function
main "$@"