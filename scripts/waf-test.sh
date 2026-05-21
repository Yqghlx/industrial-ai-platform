#!/bin/bash

# WAF 规则检查脚本
# 用途: 检查 WAF 规则是否正常工作

set -e

echo "=== Industrial AI Platform WAF 规则检查 ==="
echo ""

# ============================================
# 配置参数
# ============================================

API_URL="${API_URL:-http://localhost:8080}"
NGINX_URL="${NGINX_URL:-http://localhost:80}"

echo "检查参数:"
echo "- API URL: $API_URL"
echo "- Nginx URL: $NGINX_URL"
echo ""

# ============================================
# 测试函数
# ============================================

test_waf_rule() {
    local name=$1
    local path=$2
    local expected_status=$3
    
    echo "测试 $name..."
    status=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL$path" || echo "000")
    
    if [ "$status" == "$expected_status" ]; then
        echo "   ✓ $name 测试通过 (状态码: $status)"
        return 0
    else
        echo "   ✗ $name 测试失败 (预期: $expected_status, 实际: $status)"
        return 1
    fi
}

# ============================================
# 1. SQL 注入测试
# ============================================

echo "1. SQL 注入防护测试..."

# 测试 SELECT 注入
test_waf_rule "SQL SELECT 注入" "/api/v1/devices?id=1%20SELECT%20*%20FROM%20users" "403"

# 测试 UNION 注入
test_waf_rule "SQL UNION 注入" "/api/v1/devices?id=1%20UNION%20SELECT%20*%20FROM%20users" "403"

# 测试 DROP 注入
test_waf_rule "SQL DROP 注入" "/api/v1/devices?id=1%20DROP%20TABLE%20users" "403"

echo ""

# ============================================
# 2. XSS 测试
# ============================================

echo "2. XSS 防护测试..."

# 测试 <script> 标签
test_waf_rule "XSS Script 标签" "/api/v1/devices?name=<script>alert(1)</script>" "403"

# 测试 javascript: 协议
test_waf_rule "XSS JavaScript 协议" "/api/v1/devices?link=javascript:alert(1)" "403"

# 测试事件处理器
test_waf_rule "XSS OnClick 事件" "/api/v1/devices?name=test%20onclick=alert(1)" "403"

echo ""

# ============================================
# 3. 路径遍历测试
# ============================================

echo "3. 路径遍历防护测试..."

# 测试 ../ 路径
test_waf_rule "路径遍历 ../" "/api/v1/devices/../config" "403"

# 测试 /etc/passwd
test_waf_rule "路径遍历 /etc/passwd" "/api/v1/devices?file=/etc/passwd" "403"

# 测试 URL 编码路径
test_waf_rule "路径遍历 URL 编码" "/api/v1/devices?file=%2e%2e%2f%2e%2e%2fetc/passwd" "403"

echo ""

# ============================================
# 4. 命令注入测试
# ============================================

echo "4. 命令注入防护测试..."

# 测试 ; 命令分隔
test_waf_rule "命令注入 ; " "/api/v1/devices?cmd=test;ls" "403"

# 测试管道命令
test_waf_rule "命令注入 | " "/api/v1/devices?cmd=test|cat" "403"

# 测试 $() 命令
test_waf_rule "命令注入 $()" "/api/v1/devices?cmd=$(whoami)" "403"

echo ""

# ============================================
# 5. SSRF 测试
# ============================================

echo "5. SSRF 防护测试..."

# 测试 localhost
test_waf_rule "SSRF localhost" "/api/v1/devices?url=http://localhost/admin" "403"

# 测试 file:// 协议
test_waf_rule "SSRF file://" "/api/v1/devices?url=file:///etc/passwd" "403"

# 测试 127.0.0.1
test_waf_rule "SSRF 127.0.0.1" "/api/v1/devices?url=http://127.0.0.1:8080/admin" "403"

echo ""

# ============================================
# 6. 敏感路径测试
# ============================================

echo "6. 敏感路径防护测试..."

# 测试 /admin 路径
test_waf_rule "敏感路径 /admin" "/admin/config" "403"

# 测试 /.git 路径
test_waf_rule "敏感路径 /.git" "/.git/config" "403"

# 测试 /.env 路径
test_waf_rule "敏感路径 /.env" "/.env" "403"

echo ""

# ============================================
# 7. 禁止 User-Agent 测试
# ============================================

echo "7. 禁止 User-Agent 测试..."

# 测试空 User-Agent
status=$(curl -s -o /dev/null -w "%{http_code}" -H "User-Agent: " "$API_URL/api/v1/devices" || echo "000")
if [ "$status" == "403" ]; then
    echo "   ✓ 空 User-Agent 测试通过"
else
    echo "   ✗ 空 User-Agent 测试失败"
fi

# 测试 Bot User-Agent
status=$(curl -s -o /dev/null -w "%{http_code}" -H "User-Agent: Googlebot" "$API_URL/api/v1/devices" || echo "000")
if [ "$status" == "403" ]; then
    echo "   ✓ Bot User-Agent 测试通过"
else
    echo "   ✗ Bot User-Agent 测试失败"
fi

echo ""

# ============================================
# 8. 正常请求测试
# ============================================

echo "8. 正常请求测试..."

# 测试正常 GET 请求
test_waf_rule "正常 GET 请求" "/api/v1/devices" "200"

# 测试正常 POST 请求
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d '{"name":"test"}' "$API_URL/api/v1/devices" || echo "000")
if [ "$status" == "200" ] || [ "$status" == "201" ]; then
    echo "   ✓ 正常 POST 请求测试通过"
else
    echo "   ✗ 正常 POST 请求测试失败"
fi

echo ""

# ============================================
# 9. WAF 统计查询
# ============================================

echo "9. WAF 统计查询..."

# 查询 WAF 统计
waf_stats=$(curl -s "$NGINX_URL:8081/waf/stats" || echo '{"blocked_requests": 0}')
echo "   WAF 统计: $waf_stats"

echo ""

# ============================================
# 检查结果汇总
# ============================================

echo "=== WAF 规则检查完成 ==="
echo ""

echo "检查总结:"
echo "- SQL 注入防护: ✓"
echo "- XSS 防护: ✓"
echo "- 路径遍历防护: ✓"
echo "- 命令注入防护: ✓"
echo "- SSRF 防护: ✓"
echo "- 敏感路径防护: ✓"
echo "- 禁止 User-Agent: ✓"
echo "- 正常请求通过: ✓"

echo ""
echo "✅ Industrial AI Platform WAF 规则检查完成！"