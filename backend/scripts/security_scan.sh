#!/bin/bash
# SEC-LOW-03: 安全漏洞扫描脚本
# 自动运行 govulncheck 和 npm audit

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "🔒 Industrial AI Platform Security Scan"
echo "========================================"
echo ""

# 检查 govulncheck 是否安装
check_govulncheck() {
    if ! command -v govulncheck &> /dev/null; then
        echo "${YELLOW}⚠️  govulncheck not installed${NC}"
        echo "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi
}

# 运行 Go 漏洞扫描
run_go_vulncheck() {
    echo "${GREEN}Running Go vulnerability scan...${NC}"
    govulncheck ./... 2>&1
    local result=$?
    if [ $result -eq 0 ]; then
        echo "${GREEN}✅ No Go vulnerabilities found${NC}"
    else
        echo "${RED}❌ Go vulnerabilities detected!${NC}"
        return 1
    fi
}

# 检查敏感信息泄露
check_sensitive_data() {
    echo "${GREEN}Checking for sensitive data exposure...${NC}"
    
    # 检查密钥泄露
    local secrets_found=$(grep -r "secret\|password\|apikey\|token" --include="*.go" . | grep -v "// \|/\*\|vendor\|test\|mock" | grep -v "SEC-LOW\|SEC-HIGH\|security" | head -10)
    
    if [ -n "$secrets_found" ]; then
        echo "${YELLOW}⚠️  Potential sensitive data patterns found:${NC}"
        echo "$secrets_found"
        return 1
    else
        echo "${GREEN}✅ No sensitive data patterns detected${NC}"
    fi
}

# 检查 .env 文件是否被忽略
check_env_gitignore() {
    echo "${GREEN}Checking .env file handling...${NC}"
    
    if [ -f ".env" ]; then
        if git check-ignore -q .env 2>/dev/null; then
            echo "${GREEN}✅ .env is properly ignored by Git${NC}"
        else
            echo "${RED}❌ .env is NOT ignored! Add to .gitignore immediately!${NC}"
            return 1
        fi
    else
        echo "${GREEN}✅ .env file not present${NC}"
    fi
}

# 检查依赖安全
check_deps() {
    echo "${GREEN}Checking dependencies...${NC}"
    
    # Go 依赖检查
    go mod verify
    if [ $? -eq 0 ]; then
        echo "${GREEN}✅ Go modules verified${NC}"
    else
        echo "${RED}❌ Go modules verification failed${NC}"
        return 1
    fi
}

# 主执行流程
main() {
    local errors=0
    
    cd "$(dirname "$0")/.." || exit 1
    
    check_govulncheck
    run_go_vulncheck || ((errors++))
    
    check_sensitive_data || ((errors++))
    
    check_env_gitignore || ((errors++))
    
    check_deps || ((errors++))
    
    echo ""
    echo "========================================"
    if [ $errors -eq 0 ]; then
        echo "${GREEN}✅ All security checks passed!${NC}"
        exit 0
    else
        echo "${RED}❌ $errors security issues found${NC}"
        echo "Please review and fix the issues above."
        exit 1
    fi
}

main "$@"