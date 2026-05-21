#!/bin/bash

# =============================================================================
# 生产配置验证脚本
# Industrial AI Platform - Production Configuration Verification
# =============================================================================
# 用途: 验证生产环境配置的安全性和正确性
# 用法: ./scripts/verify-production-config.sh [--env-file .env] [--k8s]
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 计数器
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

# 默认配置
ENV_FILE=".env"
K8S_MODE=false
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        --k8s)
            K8S_MODE=true
            shift
            ;;
        -h|--help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --env-file FILE  指定环境变量文件 (默认: .env)"
            echo "  --k8s            启用 Kubernetes 配置验证"
            echo "  -h, --help       显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            exit 1
            ;;
    esac
done

# =============================================================================
# 辅助函数
# =============================================================================

print_header() {
    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
    echo ""
}

print_check() {
    echo -n "  [$1] $2 ... "
}

pass() {
    echo -e "${GREEN}✓ PASS${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}"
    if [[ -n "$1" ]]; then
        echo -e "      ${RED}原因: $1${NC}"
    fi
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

warn() {
    echo -e "${YELLOW}⚠ WARN${NC}"
    if [[ -n "$1" ]]; then
        echo -e "      ${YELLOW}提示: $1${NC}"
    fi
    WARN_COUNT=$((WARN_COUNT + 1))
}

info() {
    echo -e "${BLUE}INFO${NC}: $1"
}

# =============================================================================
# 1. JWT_SECRET 配置检查
# =============================================================================
check_jwt_secret() {
    print_header "1. JWT_SECRET 配置检查"
    
    # 检查环境变量
    if [[ -f "$ENV_FILE" ]]; then
        JWT_SECRET=$(grep -E "^JWT_SECRET=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'")
    fi
    
    # 如果环境变量文件中没有，检查当前环境
    if [[ -z "$JWT_SECRET" ]]; then
        JWT_SECRET="${JWT_SECRET:-$(printenv JWT_SECRET 2>/dev/null || true)}"
    fi
    
    print_check "JWT" "JWT_SECRET 是否配置"
    if [[ -z "$JWT_SECRET" ]]; then
        fail "JWT_SECRET 未配置"
        return 0
    fi
    pass
    
    print_check "JWT" "JWT_SECRET 长度是否 >= 32 字符"
    if [[ -n "$JWT_SECRET" ]]; then
        JWT_LENGTH=${#JWT_SECRET}
        if [[ $JWT_LENGTH -lt 32 ]]; then
            fail "当前长度: $JWT_LENGTH 字符 (要求 >= 32)"
        else
            pass "长度: $JWT_LENGTH 字符"
        fi
    else
        warn "跳过 (JWT_SECRET 未配置)"
    fi
    
    print_check "JWT" "JWT_SECRET 是否为弱密钥"
    if [[ -n "$JWT_SECRET" ]]; then
        WEAK_PATTERNS=("your-jwt-secret" "change-me" "secret" "password" "jwt-secret-key" "test" "default" "example")
        IS_WEAK=false
        for pattern in "${WEAK_PATTERNS[@]}"; do
            if [[ "${JWT_SECRET,,}" == *"${pattern,,}"* ]]; then
                IS_WEAK=true
                break
            fi
        done
        
        if [[ "$IS_WEAK" == true ]]; then
            fail "检测到弱密钥模式，请使用强随机密钥"
        else
            pass
        fi
    else
        warn "跳过 (JWT_SECRET 未配置)"
    fi
    
    # K8s Secrets 检查
    if [[ "$K8S_MODE" == true ]]; then
        print_check "K8S" "Kubernetes jwt-secret 是否配置真实值"
        if command -v kubectl &> /dev/null; then
            K8S_JWT=$(kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath='{.data.jwt-secret}' 2>/dev/null || true)
            if [[ -n "$K8S_JWT" ]]; then
                DECODED_JWT=$(echo "$K8S_JWT" | base64 -d 2>/dev/null || true)
                if [[ -n "$DECODED_JWT" && ${#DECODED_JWT} -ge 32 ]]; then
                    pass "K8s Secret 配置正确"
                else
                    warn "K8s Secret 值可能不符合要求"
                fi
            else
                warn "K8s Secret 'jwt-secret' 未找到"
            fi
        else
            warn "kubectl 未安装，跳过 K8s 检查"
        fi
    fi
}

# =============================================================================
# 2. DATABASE_URL 配置检查
# =============================================================================
check_database_url() {
    print_header "2. DATABASE_URL 配置检查"
    
    # 检查环境变量
    if [[ -f "$ENV_FILE" ]]; then
        DATABASE_URL=$(grep -E "^DATABASE_URL=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'")
    fi
    
    if [[ -z "$DATABASE_URL" ]]; then
        DATABASE_URL="${DATABASE_URL:-$(printenv DATABASE_URL 2>/dev/null || true)}"
    fi
    
    print_check "DB" "DATABASE_URL 是否配置"
    if [[ -z "$DATABASE_URL" ]]; then
        fail "DATABASE_URL 未配置"
        return 0
    fi
    pass
    
    print_check "DB" "是否使用 SSL 连接 (sslmode=require)"
    if [[ -z "$DATABASE_URL" ]]; then
        warn "跳过 (DATABASE_URL 未配置)"
    elif [[ "$DATABASE_URL" == *"sslmode=require"* ]]; then
        pass "已启用 SSL 连接"
    elif [[ "$DATABASE_URL" == *"sslmode=disable"* ]]; then
        fail "SSL 已禁用，生产环境必须启用"
    else
        warn "未明确配置 SSL，建议添加 sslmode=require"
    fi
    
    print_check "DB" "是否使用硬编码密码"
    if [[ -z "$DATABASE_URL" ]]; then
        warn "跳过 (DATABASE_URL 未配置)"
    elif [[ "$DATABASE_URL" == *"password"* || "$DATABASE_URL" == *"123456"* || "$DATABASE_URL" == *"changeme"* ]]; then
        warn "检测到可能的硬编码密码，建议使用 Secrets 管理"
    else
        pass
    fi
    
    # K8s Secrets 检查
    if [[ "$K8S_MODE" == true ]]; then
        print_check "K8S" "Kubernetes database-url 是否配置"
        if command -v kubectl &> /dev/null; then
            K8S_DB=$(kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath='{.data.database-url}' 2>/dev/null || true)
            if [[ -n "$K8S_DB" ]]; then
                DECODED_DB=$(echo "$K8S_DB" | base64 -d 2>/dev/null || true)
                if [[ "$DECODED_DB" == *"sslmode=require"* ]]; then
                    pass "K8s Secret 配置正确且启用 SSL"
                else
                    warn "K8s Secret 未启用 SSL"
                fi
            else
                warn "K8s Secret 'database-url' 未找到"
            fi
        fi
    fi
}

# =============================================================================
# 3. CORS_ORIGINS 配置检查
# =============================================================================
check_cors_origins() {
    print_header "3. CORS_ORIGINS 配置检查"
    
    # 检查环境变量
    if [[ -f "$ENV_FILE" ]]; then
        CORS_ORIGINS=$(grep -E "^CORS_ORIGINS=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'")
    fi
    
    if [[ -z "$CORS_ORIGINS" ]]; then
        CORS_ORIGINS="${CORS_ORIGINS:-$(printenv CORS_ORIGINS 2>/dev/null || true)}"
    fi
    
    print_check "CORS" "CORS_ORIGINS 是否配置"
    if [[ -z "$CORS_ORIGINS" ]]; then
        warn "CORS_ORIGINS 未配置，将使用默认值"
    else
        pass
    fi
    
    print_check "CORS" "是否使用通配符 (*)"
    if [[ -z "$CORS_ORIGINS" ]]; then
        warn "跳过 (CORS_ORIGINS 未配置)"
    elif [[ "$CORS_ORIGINS" == "*" ]]; then
        fail "生产环境禁止使用 CORS 通配符 (*)"
    elif [[ "$CORS_ORIGINS" == *", *,"* ]]; then
        fail "CORS 配置包含通配符 (*)，生产环境不允许"
    else
        pass "未使用通配符"
    fi
    
    print_check "CORS" "是否使用 HTTPS"
    if [[ -n "$CORS_ORIGINS" ]]; then
        if [[ "$CORS_ORIGINS" == *"https://"* ]]; then
            pass "配置了 HTTPS 源"
        elif [[ "$CORS_ORIGINS" == *"http://"* ]]; then
            warn "检测到 HTTP 源，生产环境建议使用 HTTPS"
        fi
    fi
    
    print_check "CORS" "是否包含 localhost"
    if [[ -n "$CORS_ORIGINS" && "$CORS_ORIGINS" == *"localhost"* ]]; then
        warn "检测到 localhost，生产环境应移除"
    else
        pass
    fi
}

# =============================================================================
# 4. K8s Secrets 配置检查
# =============================================================================
check_k8s_secrets() {
    if [[ "$K8S_MODE" != true ]]; then
        return 0
    fi
    
    print_header "4. Kubernetes Secrets 配置检查"
    
    if ! command -v kubectl &> /dev/null; then
        warn "kubectl 未安装，跳过 K8s 检查"
        return 0
    fi
    
    # 检查 namespace 是否存在
    print_check "K8S" "Namespace 'industrial-ai' 是否存在"
    if kubectl get namespace industrial-ai &> /dev/null; then
        pass
    else
        warn "Namespace 不存在，需创建"
    fi
    
    # 检查 Secrets
    print_check "K8S" "Secret 'industrial-ai-secrets' 是否存在"
    if kubectl get secret industrial-ai-secrets -n industrial-ai &> /dev/null; then
        pass
        
        # 检查各个 Secret 键
        local SECRETS_TO_CHECK=("jwt-secret" "database-url" "redis-password" "admin-password")
        for secret_key in "${SECRETS_TO_CHECK[@]}"; do
            print_check "K8S" "Secret 键 '$secret_key' 是否存在"
            if kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath="{.data.$secret_key}" &> /dev/null | grep -q .; then
                pass
            else
                fail "Secret 键 '$secret_key' 未配置"
            fi
        done
    else
        fail "Secret 'industrial-ai-secrets' 不存在"
    fi
    
    # 检查占位符值
    print_check "K8S" "检查是否使用占位符值"
    local PLACEHOLDER_PATTERNS=("your-" "change-me" "changeme" "placeholder" "example" "test-")
    local HAS_PLACEHOLDER=false
    
    for secret_key in "jwt-secret" "database-url" "redis-password" "admin-password"; do
        VALUE=$(kubectl get secret industrial-ai-secrets -n industrial-ai -o jsonpath="{.data.$secret_key}" 2>/dev/null | base64 -d 2>/dev/null || true)
        for pattern in "${PLACEHOLDER_PATTERNS[@]}"; do
            if [[ "${VALUE,,}" == *"${pattern,,}"* ]]; then
                HAS_PLACEHOLDER=true
                break 2
            fi
        done
    done
    
    if [[ "$HAS_PLACEHOLDER" == true ]]; then
        fail "检测到占位符值，请替换为真实密钥"
    else
        pass
    fi
}

# =============================================================================
# 5. TLS 证书配置检查
# =============================================================================
check_tls_config() {
    print_header "5. TLS 证书配置检查"
    
    # 检查证书文件
    local CERT_PATHS=(
        "/etc/ssl/certs/industrial-ai"
        "/etc/tls"
        "$PROJECT_ROOT/docker/certs"
        "$PROJECT_ROOT/certs"
        "$PROJECT_ROOT/ssl"
    )
    
    local CERT_FOUND=false
    for cert_dir in "${CERT_PATHS[@]}"; do
        if [[ -d "$cert_dir" ]]; then
            if [[ -f "$cert_dir/tls.crt" ]] || [[ -f "$cert_dir/server.crt" ]]; then
                print_check "TLS" "证书目录 $cert_dir"
                pass
                CERT_FOUND=true
                
                # 检查证书有效期
                if [[ -f "$cert_dir/tls.crt" ]]; then
                    CERT_FILE="$cert_dir/tls.crt"
                elif [[ -f "$cert_dir/server.crt" ]]; then
                    CERT_FILE="$cert_dir/server.crt"
                fi
                
                if [[ -n "$CERT_FILE" && -f "$CERT_FILE" ]]; then
                    print_check "TLS" "证书有效期检查"
                    EXPIRY=$(openssl x509 -in "$CERT_FILE" -noout -enddate 2>/dev/null | cut -d= -f2 || true)
                    if [[ -n "$EXPIRY" ]]; then
                        EXPIRY_EPOCH=$(date -j -f "%b %d %T %Y %Z" "$EXPIRY" "+%s" 2>/dev/null || date -d "$EXPIRY" "+%s" 2>/dev/null || true)
                        NOW_EPOCH=$(date "+%s")
                        DAYS_LEFT=$(( (EXPIRY_EPOCH - NOW_EPOCH) / 86400 ))
                        if [[ $DAYS_LEFT -lt 0 ]]; then
                            fail "证书已过期"
                        elif [[ $DAYS_LEFT -lt 30 ]]; then
                            warn "证书将在 $DAYS_LEFT 天后过期"
                        else
                            pass "证书有效期: $DAYS_LEFT 天"
                        fi
                    fi
                fi
            fi
        fi
    done
    
    if [[ "$CERT_FOUND" == false ]]; then
        print_check "TLS" "本地证书文件"
        warn "未找到本地证书文件，请确认生产环境 TLS 配置"
    fi
    
    # K8s TLS 检查
    if [[ "$K8S_MODE" == true ]] && command -v kubectl &> /dev/null; then
        print_check "K8S" "TLS Secret 检查"
        if kubectl get secret -n industrial-ai -l "app.kubernetes.io/component=tls" &> /dev/null; then
            pass "找到 TLS Secret"
        else
            # 检查常见的 TLS Secret 名称
            if kubectl get secret tls-secret -n industrial-ai &> /dev/null; then
                pass "找到 tls-secret"
            elif kubectl get secret industrial-ai-tls -n industrial-ai &> /dev/null; then
                pass "找到 industrial-ai-tls"
            else
                warn "未找到 TLS Secret，请确认已配置"
            fi
        fi
    fi
}

# =============================================================================
# 6. 容器非 root 用户检查
# =============================================================================
check_container_security() {
    print_header "6. 容器安全配置检查"
    
    # 检查 Dockerfile
    local DOCKERFILES=(
        "$PROJECT_ROOT/backend/Dockerfile"
        "$PROJECT_ROOT/frontend/Dockerfile"
    )
    
    for dockerfile in "${DOCKERFILES[@]}"; do
        if [[ -f "$dockerfile" ]]; then
            COMPONENT=$(basename $(dirname "$dockerfile"))
            print_check "CONTAINER" "$COMPONENT Dockerfile 安全配置"
            
            # 检查是否有非 root 用户配置
            if grep -q "USER" "$dockerfile"; then
                USER_LINE=$(grep -E "^USER" "$dockerfile" | tail -1)
                if [[ "$USER_LINE" != *"root"* ]]; then
                    pass "配置了非 root 用户: $USER_LINE"
                else
                    fail "使用 root 用户运行容器"
                fi
            else
                warn "未显式配置 USER，可能以 root 运行"
            fi
        fi
    done
    
    # 检查 docker-compose.yml
    local COMPOSE_FILES=(
        "$PROJECT_ROOT/docker-compose.yml"
        "$PROJECT_ROOT/docker-compose.prod-ssl.yml"
    )
    
    for compose_file in "${COMPOSE_FILES[@]}"; do
        if [[ -f "$compose_file" ]]; then
            print_check "CONTAINER" "$(basename $compose_file) 安全配置"
            
            if grep -q "user:" "$compose_file" 2>/dev/null; then
                pass "配置了用户映射"
            else
                warn "未配置 user 映射，建议显式指定"
            fi
        fi
    done
    
    # K8s 安全上下文检查
    if [[ "$K8S_MODE" == true ]] && command -v kubectl &> /dev/null; then
        print_check "K8S" "Pod 安全上下文检查"
        
        local K8S_DEPLOY_DIR="$PROJECT_ROOT/infra/k8s"
        if [[ -d "$K8S_DEPLOY_DIR" ]]; then
            local SECURITY_FOUND=false
            
            for file in "$K8S_DEPLOY_DIR"/*.yaml; do
                if [[ -f "$file" ]]; then
                    if grep -q "runAsNonRoot:" "$file" || grep -q "runAsUser:" "$file"; then
                        SECURITY_FOUND=true
                        break
                    fi
                fi
            done
            
            if [[ "$SECURITY_FOUND" == true ]]; then
                pass "K8s 配置包含安全上下文"
            else
                warn "K8s 配置未配置安全上下文"
            fi
        fi
    fi
}

# =============================================================================
# 7. 额外安全检查
# =============================================================================
check_additional_security() {
    print_header "7. 额外安全检查"
    
    # 检查 REDIS_PASSWORD
    print_check "SEC" "Redis 密码配置"
    if [[ -f "$ENV_FILE" ]]; then
        REDIS_PASSWORD=$(grep -E "^REDIS_PASSWORD=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'" || true)
    fi
    
    if [[ -z "$REDIS_PASSWORD" ]]; then
        REDIS_PASSWORD="${REDIS_PASSWORD:-$(printenv REDIS_PASSWORD 2>/dev/null || true)}"
    fi
    
    if [[ -n "$REDIS_PASSWORD" ]]; then
        pass
    else
        warn "REDIS_PASSWORD 未配置，建议设置"
    fi
    
    # 检查 ADMIN_PASSWORD
    print_check "SEC" "管理员密码配置"
    if [[ -f "$ENV_FILE" ]]; then
        ADMIN_PASSWORD=$(grep -E "^ADMIN_PASSWORD=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'" || true)
    fi
    
    if [[ -z "$ADMIN_PASSWORD" ]]; then
        ADMIN_PASSWORD="${ADMIN_PASSWORD:-$(printenv ADMIN_PASSWORD 2>/dev/null || true)}"
    fi
    
    if [[ -n "$ADMIN_PASSWORD" ]]; then
        # 检查弱密码
        if [[ "$ADMIN_PASSWORD" == "admin" || "$ADMIN_PASSWORD" == "password" || "$ADMIN_PASSWORD" == "changeme" ]]; then
            fail "使用弱管理员密码"
        else
            pass
        fi
    else
        warn "ADMIN_PASSWORD 未配置"
    fi
    
    # 检查 GIN_MODE
    print_check "SEC" "GIN_MODE 配置"
    if [[ -f "$ENV_FILE" ]]; then
        GIN_MODE=$(grep -E "^GIN_MODE=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2- | tr -d '"' | tr -d "'" || true)
    fi
    
    if [[ -z "$GIN_MODE" ]]; then
        GIN_MODE="${GIN_MODE:-$(printenv GIN_MODE 2>/dev/null || true)}"
    fi
    
    if [[ "$GIN_MODE" == "release" ]]; then
        pass "生产模式已启用"
    elif [[ "$GIN_MODE" == "debug" ]]; then
        warn "GIN_MODE=debug，生产环境应使用 release"
    else
        info "GIN_MODE 未设置，默认为 debug"
    fi
    
    # 检查 .env 是否在 .gitignore
    print_check "SEC" ".env 文件保护"
    if [[ -f "$PROJECT_ROOT/.gitignore" ]]; then
        if grep -q "^.env$" "$PROJECT_ROOT/.gitignore" || grep -q "^\.env$" "$PROJECT_ROOT/.gitignore"; then
            pass ".env 已在 .gitignore 中"
        else
            fail ".env 未在 .gitignore 中，存在泄露风险"
        fi
    else
        warn ".gitignore 不存在"
    fi
}

# =============================================================================
# 输出验证报告
# =============================================================================
print_report() {
    print_header "验证报告"
    
    local TOTAL=$((PASS_COUNT + FAIL_COUNT + WARN_COUNT))
    
    echo ""
    echo -e "  ${GREEN}通过: $PASS_COUNT${NC}"
    echo -e "  ${RED}失败: $FAIL_COUNT${NC}"
    echo -e "  ${YELLOW}警告: $WARN_COUNT${NC}"
    echo -e "  总计: $TOTAL 项检查"
    echo ""
    
    if [[ $FAIL_COUNT -gt 0 ]]; then
        echo -e "${RED}════════════════════════════════════════════════════════════════${NC}"
        echo -e "${RED}  ❌ 验证失败: 存在 $FAIL_COUNT 个必须修复的问题${NC}"
        echo -e "${RED}════════════════════════════════════════════════════════════════${NC}"
        exit 1
    elif [[ $WARN_COUNT -gt 0 ]]; then
        echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
        echo -e "${YELLOW}  ⚠️  验证通过，但存在 $WARN_COUNT 个警告建议处理${NC}"
        echo -e "${YELLOW}════════════════════════════════════════════════════════════════${NC}"
        exit 0
    else
        echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}  ✅ 所有检查通过，配置符合生产安全要求${NC}"
        echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
        exit 0
    fi
}

# =============================================================================
# 主函数
# =============================================================================
main() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║     Industrial AI Platform - 生产配置验证                    ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "  配置文件: $ENV_FILE"
    echo "  K8s 模式: $K8S_MODE"
    echo "  项目目录: $PROJECT_ROOT"
    
    cd "$PROJECT_ROOT"
    
    # 执行检查
    check_jwt_secret
    check_database_url
    check_cors_origins
    check_k8s_secrets
    check_tls_config
    check_container_security
    check_additional_security
    
    # 输出报告
    print_report
}

main "$@"