#!/bin/bash
# Industrial AI Platform - Quick Start Script
# 使用 Docker Compose 启动所有服务

set -e

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║       Industrial AI Platform - Quick Start               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 设置环境变量
echo "📋 配置环境变量..."
export JWT_SECRET="industrial-ai-platform-dev-secret"

# 检查是否有百炼 API Key
if [ -z "$LLM_API_KEY" ]; then
    echo "⚠️  LLM_API_KEY 未设置，AI Agent 将使用 Mock 模式"
    echo "   设置方法: export LLM_API_KEY='sk-sp-your-key'"
else
    echo "✅ LLM_API_KEY 已设置"
fi

# 启动服务
echo ""
echo "🚀 启动服务..."
docker-compose up -d postgres backend frontend

# 等待服务就绪
echo ""
echo "⏳ 等待服务启动..."
sleep 10

# 检查健康状态
echo ""
echo "🔍 检查服务状态..."
curl -s http://localhost:8080/health | head -1 || echo "❌ Backend 未就绪"

echo ""
echo "══════════════════════════════════════════════════════════"
echo "✅ 服务已启动！"
echo ""
echo "访问地址:"
echo "  📊 前端:      http://localhost:3000"
echo "  🔌 后端 API:  http://localhost:8080"
echo "  📖 API 文档:  http://localhost:8080/docs/"
echo "  💓 健康检查:  http://localhost:8080/health"
echo ""
echo "启动边缘模拟器 (可选):"
echo "  docker-compose --profile simulator up -d edge-simulator"
echo ""
echo "停止服务:"
echo "  docker-compose down"
echo "══════════════════════════════════════════════════════════"