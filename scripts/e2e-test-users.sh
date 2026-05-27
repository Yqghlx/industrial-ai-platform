#!/bin/bash
# E2E 测试用户初始化脚本
# 创建 operator 和 viewer 测试用户

set -e

# 测试用户配置
OPERATOR_PASSWORD="${E2E_OPERATOR_PASSWORD:-Operator@123}"
VIEWER_PASSWORD="${VIEWER_PASSWORD:-Viewer@123}"

# API 地址
API_URL="${E2E_API_URL:-http://localhost:8080}"

echo "=== E2E 测试用户初始化 ==="

# 获取 admin token（使用实际环境变量中的密码）
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin@TPby8q1dmPk}"
echo "获取 admin token..."
ADMIN_TOKEN=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"${ADMIN_PASSWORD}\"}" | jq -r '.token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
  echo "ERROR: 无法获取 admin token，请检查 ADMIN_PASSWORD 环境变量"
  echo "当前 ADMIN_PASSWORD: ${ADMIN_PASSWORD}"
  exit 1
fi

echo "Admin token 获取成功"

# 创建 operator 用户
echo "创建 operator 用户..."
curl -s -X POST "${API_URL}/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d "{
    \"username\": \"operator\",
    \"password\": \"${OPERATOR_PASSWORD}\",
    \"email\": \"operator@industrial.ai\",
    \"role\": \"operator\"
  }" > /dev/null 2>&1 || echo "operator 用户可能已存在"

# 创建 viewer 用户  
echo "创建 viewer 用户..."
curl -s -X POST "${API_URL}/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d "{
    \"username\": \"viewer\",
    \"password\": \"${VIEWER_PASSWORD}\",
    \"email\": \"viewer@industrial.ai\",
    \"role\": \"viewer\"
  }" > /dev/null 2>&1 || echo "viewer 用户可能已存在"

echo "=== 测试用户初始化完成 ==="
echo "用户列表:"
curl -s "${API_URL}/api/v1/admin/users" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.[].username'