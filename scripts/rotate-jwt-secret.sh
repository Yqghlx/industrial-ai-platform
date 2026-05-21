#!/bin/bash

# JWT Secret 轮换脚本
# 用途: 更换 JWT Secret 并重启服务

set -e

echo "=== JWT Secret 轮换 ==="
echo ""

# 环境检查
if [ -z "$KUBECONFIG" ] && [ ! -f ~/.kube/config ]; then
    echo "错误: Kubernetes 未配置"
    echo "请设置 KUBECONFIG 或配置 ~/.kube/config"
    exit 1
fi

NAMESPACE="${NAMESPACE:-industrial-ai}"
SECRET_NAME="${SECRET_NAME:-industrial-ai-secrets}"
DEPLOYMENT="${DEPLOYMENT:-backend}"

echo "Namespace: $NAMESPACE"
echo "Secret: $SECRET_NAME"
echo "Deployment: $DEPLOYMENT"
echo ""

echo "⚠️  警告: 轮换 JWT Secret 将使所有现有 Token 失效"
echo "所有用户将需要重新登录！"
echo ""
read -p "确认继续？(yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "取消轮换"
    exit 0
fi

echo ""
echo "1. 生成新 JWT Secret..."
NEW_JWT_SECRET=$(openssl rand -base64 32)
echo "   ✓ 新 Secret 已生成"

echo ""
echo "2. 备份当前 Secret..."
kubectl get secret $SECRET_NAME -n $NAMESPACE -o yaml > /tmp/old-secret-backup.yaml
echo "   ✓ 备份保存到 /tmp/old-secret-backup.yaml"

echo ""
echo "3. 更新 Kubernetes Secret..."
kubectl create secret generic $SECRET_NAME \
    --from-literal=jwt-secret="$NEW_JWT_SECRET" \
    --namespace=$NAMESPACE \
    --dry-run=client -o yaml | kubectl apply -f -
echo "   ✓ Secret 已更新"

echo ""
echo "4. 验证 Secret 更新..."
kubectl get secret $SECRET_NAME -n $NAMESPACE
echo "   ✓ Secret 验证完成"

echo ""
echo "5. 重启后端服务..."
kubectl rollout restart deployment/$DEPLOYMENT -n $NAMESPACE
echo "   ✓ 部署重启命令已执行"

echo ""
echo "6. 等待部署就绪..."
kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=300s
echo "   ✓ 部署已就绪"

echo ""
echo "=== 轮换完成 ==="
echo ""
echo "新 JWT Secret: $NEW_JWT_SECRET"
echo ""
echo "重要提醒:"
echo "- 所有用户需要重新登录"
echo "- 新 Secret 已生效"
echo "- 旧 Token 已失效"
echo "- 备份文件: /tmp/old-secret-backup.yaml"
echo ""

# 输出轮换时间
echo "轮换时间: $(date)"
echo "下次轮换建议: $(date -v +90d '+%Y-%m-%d' 2>/dev/null || date -d '+90 days' '+%Y-%m-%d' 2>/dev/null || echo '90 days later')"