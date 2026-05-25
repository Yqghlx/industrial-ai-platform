#!/bin/bash
# Industrial AI Platform - Local Development Script
# 本地开发模式启动脚本（不使用 Docker）

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║       Industrial AI Platform - Local Dev                  ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# 设置环境变量（从 .env 文件读取，如不存在则使用安全默认值）
echo "📋 配置环境变量..."

# SECURITY: Read from .env file instead of hardcoding
if [ -f "$PROJECT_ROOT/.env" ]; then
    echo "   从 .env 文件加载配置..."
    export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)
else
    echo "⚠️  未找到 .env 文件，请先创建:"
    echo "   cp .env.example .env"
    echo "   并填写必要的配置项"
    exit 1
fi

# 检查必要的配置
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL 未设置"
    exit 1
fi

if [ -z "$JWT_SECRET" ]; then
    echo "❌ JWT_SECRET 未设置"
    exit 1
fi

if [ -z "$LLM_API_KEY" ]; then
    echo "⚠️  未设置 LLM_API_KEY，AI Agent 将使用 Mock 模式"
    echo "   设置方法: export LLM_API_KEY='sk-sp-your-key'"
    echo "   export LLM_BASE_URL='https://coding.dashscope.aliyuncs.com/v1'"
    echo "   export LLM_MODEL='glm-5'"
fi

echo ""
echo "请先确保数据库已启动:"
echo "  docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=industrial_ai -p 5432:5432 timescale/timescaledb:latest-pg15"
echo ""
read -p "数据库已启动？按 Enter 继续..."

# 启动后端
echo ""
echo "🚀 启动后端..."
cd "$PROJECT_ROOT/backend"
go run main.go &
BACKEND_PID=$!
echo "   Backend PID: $BACKEND_PID"

# 等待后端启动
sleep 3

# 启动前端
echo ""
echo "🚀 启动前端..."
cd "$PROJECT_ROOT/frontend"

# 检查 node_modules
if [ ! -d "node_modules" ]; then
    echo "   安装前端依赖..."
    npm install
fi

npm run dev &
FRONTEND_PID=$!
echo "   Frontend PID: $FRONTEND_PID"

echo ""
echo "══════════════════════════════════════════════════════════"
echo "✅ 开发服务已启动！"
echo ""
echo "访问地址:"
echo "  📊 前端:      http://localhost:5173"
echo "  🔌 后端 API:  http://localhost:8080"
echo ""
echo "停止服务: Ctrl+C 或 kill $BACKEND_PID $FRONTEND_PID"
echo "══════════════════════════════════════════════════════════"

# 等待用户中断
trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; echo '已停止'; exit 0" SIGINT SIGTERM
wait