#!/bin/bash

# Industrial AI Platform 自动化部署脚本
# 用途: 一键部署应用 (构建/推送/部署/验证)

set -e

echo "=== Industrial AI Platform 自动化部署 ==="
echo ""

# ============================================
# 配置参数
# ============================================

# 应用配置
APP_NAME="industrial-ai"
VERSION="${VERSION:-$(git rev-parse --short HEAD)}"
BUILD_TIME=$(date +%Y%m%d-%H%M%S)

# Docker 配置
DOCKER_REGISTRY="${DOCKER_REGISTRY:-registry.example.com}"
DOCKER_IMAGE_BACKEND="$DOCKER_REGISTRY/$APP_NAME-backend:$VERSION"
DOCKER_IMAGE_FRONTEND="$DOCKER_REGISTRY/$APP_NAME-frontend:$VERSION"

# Kubernetes 配置
KUBE_NAMESPACE="${KUBE_NAMESPACE:-industrial-ai}"
KUBE_CONTEXT="${KUBE_CONTEXT:-production}"

# 环境配置
ENVIRONMENT="${ENVIRONMENT:-production}"

echo "部署参数:"
echo "- 应用名称: $APP_NAME"
echo "- 版本: $VERSION"
echo "- 构建时间: $BUILD_TIME"
echo "- 环境: $ENVIRONMENT"
echo "- Registry: $DOCKER_REGISTRY"
echo ""

# ============================================
# 1. 环境检查
# ============================================

echo "1. 检查部署环境..."

# 检查必要工具
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo "   ✗ 缺少工具: $1"
        exit 1
    else
        echo "   ✓ 工具已安装: $1"
    fi
}

check_tool "docker"
check_tool "kubectl"
check_tool "git"

# 检查 Docker 登录状态
if ! docker info | grep -q "Username"; then
    echo "   ⚠️ Docker 未登录，请先登录"
    echo "   命令: docker login $DOCKER_REGISTRY"
    exit 1
fi

# 检查 Kubernetes 连接
if ! kubectl cluster-info &> /dev/null; then
    echo "   ✗ Kubernetes 连接失败"
    exit 1
fi

echo "   ✓ 环境检查通过"

# ============================================
# 2. 拉取最新代码
# ============================================

echo ""
echo "2. 拉取最新代码..."

git fetch origin
git checkout main
git pull origin main

echo "   ✓ 代码已更新"

# ============================================
# 3. 构建后端应用
# ============================================

echo ""
echo "3. 构建后端应用..."

cd backend

# 构建 Go 应用
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" \
    -o industrial-ai-backend \
    ./cmd/server

if [ $? -eq 0 ]; then
    echo "   ✓ 后端构建成功"
else
    echo "   ✗ 后端构建失败"
    exit 1
fi

cd ..

# ============================================
# 4. 构建前端应用
# ============================================

echo ""
echo "4. 构建前端应用..."

cd frontend

# 构建 React 应用
npm install --production
npm run build

if [ $? -eq 0 ]; then
    echo "   ✓ 前端构建成功"
else
    echo "   ✗ 前端构建失败"
    exit 1
fi

cd ..

# ============================================
# 5. 构建 Docker 镜像
# ============================================

echo ""
echo "5. 构建 Docker 镜像..."

# 构建后端镜像
docker build -t $DOCKER_IMAGE_BACKEND \
    --build-arg VERSION=$VERSION \
    --build-arg BUILD_TIME=$BUILD_TIME \
    -f backend/Dockerfile \
    backend/

echo "   ✓ 后端镜像构建完成: $DOCKER_IMAGE_BACKEND"

# 构建前端镜像
docker build -t $DOCKER_IMAGE_FRONTEND \
    --build-arg VERSION=$VERSION \
    -f frontend/Dockerfile \
    frontend/

echo "   ✓ 前端镜像构建完成: $DOCKER_IMAGE_FRONTEND"

# ============================================
# 6. 推送 Docker 镇像
# ============================================

echo ""
echo "6. 推送 Docker 镜像到 Registry..."

docker push $DOCKER_IMAGE_BACKEND
echo "   ✓ 后端镜像推送完成"

docker push $DOCKER_IMAGE_FRONTEND
echo "   ✓ 前端镜像推送完成"

# ============================================
# 7. 部署到 Kubernetes
# ============================================

echo ""
echo "7. 部署到 Kubernetes..."

# 切换到目标 Context
kubectl config use-context $KUBE_CONTEXT

# 更新 Deployment 镜像
kubectl set image deployment/backend \
    backend=$DOCKER_IMAGE_BACKEND \
    -n $KUBE_NAMESPACE

kubectl set image deployment/frontend \
    frontend=$DOCKER_IMAGE_FRONTEND \
    -n $KUBE_NAMESPACE

echo "   ✓ Deployment 镜像已更新"

# 等待 Rolling Update 完成
echo ""
echo "8. 等待 Rolling Update 完成..."

kubectl rollout status deployment/backend -n $KUBE_NAMESPACE --timeout=300s
kubectl rollout status deployment/frontend -n $KUBE_NAMESPACE --timeout=300s

echo "   ✓ Rolling Update 完成"

# ============================================
# 9. 验证部署
# ============================================

echo ""
echo "9. 验证部署..."

# 检查 Pod 状态
POD_STATUS=$(kubectl get pods -n $KUBE_NAMESPACE -l app=$APP_NAME -o jsonpath='{.items[*].status.phase}')

if echo "$POD_STATUS" | grep -q "Running"; then
    echo "   ✓ Pods 运行正常"
else
    echo "   ✗ Pods 状态异常: $POD_STATUS"
    exit 1
fi

# 检查健康状态
HEALTH_STATUS=$(kubectl exec -n $KUBE_NAMESPACE deployment/backend -- curl -s http://localhost:8080/health/ready)

if echo "$HEALTH_STATUS" | grep -q "ready"; then
    echo "   ✓ 应用健康检查通过"
else
    echo "   ✗ 应用健康检查失败"
    exit 1
fi

# ============================================
# 10. 清理旧版本
# ============================================

echo ""
echo "10. 清理旧版本..."

# 清理本地旧镜像
docker image prune -f

echo "   ✓ 清理完成"

# ============================================
# 部署总结
# ============================================

echo ""
echo "=== 部署完成 ==="
echo ""

echo "部署总结:"
echo "- 版本: $VERSION"
echo "- 时间: $BUILD_TIME"
echo "- 环境: $ENVIRONMENT"
echo "- 镜像: $DOCKER_IMAGE_BACKEND"
echo "- 镜像: $DOCKER_IMAGE_FRONTEND"
echo "- Namespace: $KUBE_NAMESPACE"

echo ""
echo "验证命令:"
echo "- 查看 Pods: kubectl get pods -n $KUBE_NAMESPACE"
echo "- 查看日志: kubectl logs -f deployment/backend -n $KUBE_NAMESPACE"
echo "- 查看状态: kubectl rollout status deployment/backend -n $KUBE_NAMESPACE"

echo ""
echo "回滚命令 (如需要):"
echo "- 回滚后端: kubectl rollout undo deployment/backend -n $KUBE_NAMESPACE"
echo "- 回滚前端: kubectl rollout undo deployment/frontend -n $KUBE_NAMESPACE"

echo ""
echo "✅ Industrial AI Platform 部署成功！"