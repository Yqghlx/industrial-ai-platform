#!/bin/bash
# Industrial AI Platform - API Test Script
# 测试主要 API 端点

set -e

API_BASE="http://localhost:8080/api/v1"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║       Industrial AI Platform - API Test                   ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# 1. 健康检查
echo "1️⃣ 健康检查"
curl -s "$API_BASE/../health" | jq . 2>/dev/null || echo "   ❌ 失败"
echo ""

# 2. 注册用户
echo "2️⃣ 用户注册"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123","email":"test@example.com"}')
echo "   $REGISTER_RESPONSE"
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token' 2>/dev/null)
echo ""

# 3. 登录
echo "3️⃣ 用户登录 (admin/admin123)"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token' 2>/dev/null)
echo "   Token: ${ADMIN_TOKEN:0:20}..."
echo ""

# 4. 发送遥测数据
echo "4️⃣ 发送遥测数据"
curl -s -X POST "$API_BASE/devices/telemetry" \
    -H "Content-Type: application/json" \
    -d '{"device_id":"CNC-001","timestamp":"2026-05-12T21:30:00Z","temperature":75.5,"pressure":100,"vibration":1.2,"humidity":50,"power":5.5,"status":"normal"}'
echo ""
echo ""

# 5. 获取设备列表
echo "5️⃣ 获取设备列表"
curl -s "$API_BASE/devices" \
    -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.data[:3]' 2>/dev/null || echo "   ❌ 失败"
echo ""

# 6. AI Agent 查询
echo "6️⃣ AI Agent 查询"
AGENT_RESPONSE=$(curl -s -X POST "$API_BASE/agent/query" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d '{"query":"分析CNC-001设备的运行状态"}')
echo "   Agent: $(echo "$AGENT_RESPONSE" | jq -r '.agent' 2>/dev/null)"
echo "   Response: $(echo "$AGENT_RESPONSE" | jq -r '.response[:100]' 2>/dev/null)..."
echo ""

# 7. ROI 统计
echo "7️⃣ ROI 统计"
curl -s "$API_BASE/roi/stats" \
    -H "Authorization: Bearer $ADMIN_TOKEN" | jq . 2>/dev/null || echo "   ❌ 失败"
echo ""

echo "══════════════════════════════════════════════════════════"
echo "✅ API 测试完成"
echo "══════════════════════════════════════════════════════════"