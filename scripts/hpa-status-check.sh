#!/bin/bash

# HPA 状态检查脚本
# 用途: 检查 HPA 状态并生成报告

set -e

echo "=== HPA 状态检查 ==="
echo ""

# 环境变量
NAMESPACE="${NAMESPACE:-industrial-ai}"

echo "1. 检查所有 HPA 状态..."
kubectl get hpa -n $NAMESPACE

echo ""
echo "2. 检查 HPA 详细状态..."

# Backend HPA
echo "Backend HPA:"
kubectl describe hpa backend-hpa -n $NAMESPACE | grep -E "Metrics|Replicas|Conditions|Events" || echo "   未找到 Backend HPA"

echo ""
echo "3. 检查 HPA 指标来源..."

# 检查 Metrics Server
echo "Metrics Server:"
kubectl get deployment metrics-server -n kube-system || echo "   ⚠️ Metrics Server 未部署"

# 检查 Prometheus Adapter
echo "Prometheus Adapter:"
kubectl get deployment prometheus-adapter -n monitoring || echo "   ⚠️ Prometheus Adapter 未部署"

echo ""
echo "4. 检查 Pod 资源使用..."

# Backend Pods 资源
echo "Backend Pods:"
kubectl top pods -l app=industrial-ai,component=backend -n $NAMESPACE || echo "   ⚠️ 无法获取 Pod 资源"

# Frontend Pods 资源
echo "Frontend Pods:"
kubectl top pods -l app=industrial-ai,component=frontend -n $NAMESPACE || echo "   ⚠️ 无法获取 Pod 资源"

echo ""
echo "5. 检查 Deployment 副本数..."

kubectl get deployments -n $NAMESPACE -o custom-columns="NAME:.metadata.name,REPLICAS:.spec.replicas,READY:.status.readyReplicas"

echo ""
echo "6. 检查扩缩容历史..."

# 获取最近扩缩容事件
echo "最近扩缩容事件:"
kubectl get events -n $NAMESPACE --field-selector reason=ScalingReplicaSet | tail -10 || echo "   无扩缩容事件"

echo ""
echo "7. 检查 HPA 配置..."

for HPA in $(kubectl get hpa -n $NAMESPACE -o jsonpath='{.items[*].metadata.name}'); do
    echo "$HPA 配置:"
    
    MIN_REPLICAS=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.spec.minReplicas}')
    MAX_REPLICAS=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.spec.maxReplicas}')
    CURRENT_REPLICAS=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.status.currentReplicas}')
    DESIRED_REPLICAS=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.status.desiredReplicas}')
    
    echo "   最小副本: $MIN_REPLICAS"
    echo "   最大副本: $MAX_REPLICAS"
    echo "   当前副本: $CURRENT_REPLICAS"
    echo "   目标副本: $DESIRED_REPLICAS"
    
    # 检查指标
    echo "   指标配置:"
    kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.spec.metrics}' | jq '.' 2>/dev/null || echo "   无法解析指标"
done

echo ""
echo "=== HPA 健康报告 ==="
echo ""

# 生成健康报告
HEALTHY=true

for HPA in $(kubectl get hpa -n $NAMESPACE -o jsonpath='{.items[*].metadata.name}'); do
    CURRENT=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.status.currentReplicas}')
    MAX=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.spec.maxReplicas}')
    
    if [ "$CURRENT" == "$MAX" ]; then
        echo "[WARNING] $HPA 已达到最大副本数 ($MAX)"
        HEALTHY=false
    fi
    
    # 检查 ScalingActive 条件
    CONDITION=$(kubectl get hpa $HPA -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="ScalingActive")].status}')
    if [ "$CONDITION" != "True" ]; then
        echo "[CRITICAL] $HPA 扩缩容不活跃"
        HEALTHY=false
    fi
done

if $HEALTHY; then
    echo "[OK] 所有 HPA 状态正常"
else
    echo "[ACTION] 需要检查 HPA 配置"
fi

echo ""
echo "HPA 操作命令:"
echo "- 查看 HPA: kubectl get hpa -n $NAMESPACE"
echo "- 详细状态: kubectl describe hpa backend-hpa -n $NAMESPACE"
echo "- 手动扩容: kubectl scale deployment backend --replicas=5 -n $NAMESPACE"
echo "- HPA YAML: kubectl get hpa backend-hpa -n $NAMESPACE -o yaml"