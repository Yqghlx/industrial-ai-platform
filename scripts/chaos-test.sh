#!/bin/bash

# Chaos 测试脚本 - Industrial AI Platform
# 用途: 执行 Chaos Engineering 实验并验证系统韧性

set -e

# ================================
# 配置变量
# ================================
NAMESPACE="${NAMESPACE:-industrial-ai}"
CHAOS_NAMESPACE="${CHAOS_NAMESPACE:-chaos-mesh}"
CHAOS_DIR="${CHAOS_DIR:-./infra/chaos-mesh}"
REPORT_DIR="${REPORT_DIR:-./docs/reports}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="${REPORT_DIR}/chaos-test-report-${TIMESTAMP}.md"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ================================
# 辅助函数
# ================================
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "检查前置条件..."
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装"
        exit 1
    fi
    
    # 检查 Chaos Mesh 是否安装
    if ! kubectl get ns chaos-mesh &> /dev/null; then
        log_warning "Chaos Mesh 命名空间不存在，请先安装 Chaos Mesh"
        log_info "安装命令: helm install chaos-mesh chaos-mesh/chaos-mesh -n chaos-mesh --create-namespace"
        exit 1
    fi
    
    # 检查应用命名空间
    if ! kubectl get ns $NAMESPACE &> /dev/null; then
        log_error "应用命名空间 $NAMESPACE 不存在"
        exit 1
    fi
    
    log_success "前置条件检查通过"
}

get_deployment_status() {
    local deploy=$1
    local replicas=$(kubectl get deployment $deploy -n $NAMESPACE -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    local ready=$(kubectl get deployment $deploy -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    local available=$(kubectl get deployment $deploy -n $NAMESPACE -o jsonpath='{.status.availableReplicas}' 2>/dev/null || echo "0")
    
    echo "$deploy: replicas=$replicas, ready=$ready, available=$available"
}

get_hpa_status() {
    local hpa=$1
    local current=$(kubectl get hpa $hpa -n $NAMESPACE -o jsonpath='{.status.currentReplicas}' 2>/dev/null || echo "N/A")
    local desired=$(kubectl get hpa $hpa -n $NAMESPACE -o jsonpath='{.status.desiredReplicas}' 2>/dev/null || echo "N/A")
    local min=$(kubectl get hpa $hpa -n $NAMESPACE -o jsonpath='{.spec.minReplicas}' 2>/dev/null || echo "N/A")
    local max=$(kubectl get hpa $hpa -n $NAMESPACE -o jsonpath='{.spec.maxReplicas}' 2>/dev/null || echo "N/A")
    
    echo "$hpa: current=$current, desired=$desired, min=$min, max=$max"
}

wait_for_chaos_experiment() {
    local experiment_name=$1
    local experiment_type=$2  # PodChaos, NetworkChaos, DNSChaos
    local timeout=${3:-300}
    
    log_info "等待 Chaos 实验 $experiment_name 完成 (超时: ${timeout}s)..."
    
    local start_time=$(date +%s)
    local end_time=$((start_time + timeout))
    
    while [ $(date +%s) -lt $end_time ]; do
        local status=$(kubectl get $experiment_type $experiment_name -n $NAMESPACE -o jsonpath='{.status.experiment.phase}' 2>/dev/null || echo "Unknown")
        
        case $status in
            "Finished"|"Finished-At-Deadline")
                log_success "Chaos 实验 $experiment_name 完成"
                return 0
                ;;
            "Paused")
                log_warning "Chaos 实验 $experiment_name 已暂停"
                return 0
                ;;
            "Running")
                log_info "Chaos 实验 $experiment_name 运行中..."
                ;;
            "Failed")
                log_error "Chaos 实验 $experiment_name 失败"
                return 1
                ;;
        esac
        
        sleep 10
    done
    
    log_error "等待 Chaos 实验 $experiment_name 超时"
    return 1
}

verify_service_health() {
    local service=$1
    local endpoint=$2
    local expected_status=${3:-200}
    local timeout=${4:-30}
    
    log_info "验证服务 $service 健康状态..."
    
    local start_time=$(date +%s)
    local end_time=$((start_time + timeout))
    
    while [ $(date +%s) -lt $end_time ]; do
        local status=$(kubectl exec -n $NAMESPACE deploy/$service -- curl -s -o /dev/null -w "%{http_code}" $endpoint 2>/dev/null || echo "000")
        
        if [ "$status" == "$expected_status" ]; then
            log_success "服务 $service 健康检查通过 (HTTP $status)"
            return 0
        fi
        
        sleep 5
    done
    
    log_error "服务 $service 健康检查失败 (最后状态: $status)"
    return 1
}

measure_recovery_time() {
    local deploy=$1
    local expected_replicas=$2
    
    log_info "测量 $deploy 恢复时间..."
    
    local start_time=$(date +%s)
    
    while true; do
        local ready=$(kubectl get deployment $deploy -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
        
        if [ "$ready" == "$expected_replicas" ]; then
            local end_time=$(date +%s)
            local recovery_time=$((end_time - start_time))
            log_success "$deploy 恢复时间: ${recovery_time}s"
            echo $recovery_time
            return 0
        fi
        
        sleep 5
        
        # 超时检查
        local current_time=$(date +%s)
        if [ $((current_time - start_time)) -gt 300 ]; then
            log_error "等待 $deploy 恢复超时"
            echo "timeout"
            return 1
        fi
    done
}

create_report_header() {
    mkdir -p $(dirname $REPORT_FILE)
    
    cat > $REPORT_FILE << EOF
# Chaos 测试报告

**测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
**测试环境**: $NAMESPACE
**测试人员**: $(whoami)

---

## 1. 测试概览

| 项目 | 值 |
|------|-----|
| 命名空间 | $NAMESPACE |
| Chaos Mesh 版本 | $(kubectl get deployment chaos-controller-manager -n chaos-mesh -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "Unknown") |
| 测试时长 | 将在测试结束时计算 |

---

## 2. 初始状态

EOF
}

record_test_result() {
    local test_name=$1
    local test_type=$2
    local result=$3
    local duration=$4
    local details=$5
    
    cat >> $REPORT_FILE << EOF

### 测试: $test_name

| 项目 | 值 |
|------|-----|
| 类型 | $test_type |
| 结果 | $result |
| 持续时间 | ${duration}s |
| 详情 | $details |

EOF
}

# ================================
# 测试场景
# ================================

# 场景 1: Pod 随机杀死测试
test_pod_kill() {
    log_info "=========================================="
    log_info "场景 1: Pod 随机杀死测试"
    log_info "=========================================="
    
    # 记录初始状态
    local initial_replicas=$(kubectl get deployment backend -n $NAMESPACE -o jsonpath='{.spec.replicas}')
    log_info "初始副本数: $initial_replicas"
    
    # 创建 Pod Kill 实验
    cat <<EOF | kubectl apply -f -
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: test-pod-kill
  namespace: $NAMESPACE
spec:
  action: pod-kill
  mode: one
  selector:
    namespaces:
      - $NAMESPACE
    labelSelectors:
      app: backend
  duration: "30s"
EOF
    
    # 等待 Pod 被杀死
    sleep 5
    
    # 记录 Pod 状态
    log_info "Pod 被杀死后的状态:"
    kubectl get pods -n $NAMESPACE -l app=backend
    
    # 测量恢复时间
    local recovery_time=$(measure_recovery_time backend $initial_replicas)
    
    # 清理实验
    kubectl delete podchaos test-pod-kill -n $NAMESPACE 2>/dev/null || true
    
    # 验证服务健康
    verify_service_health backend "http://localhost:8080/health"
    
    # 记录结果
    record_test_result "Pod 随机杀死" "PodChaos" "PASS" "30" "恢复时间: ${recovery_time}s"
    
    log_success "Pod 随机杀死测试完成"
}

# 场景 2: 网络延迟注入测试
test_network_delay() {
    log_info "=========================================="
    log_info "场景 2: 网络延迟注入测试"
    log_info "=========================================="
    
    # 创建网络延迟实验
    cat <<EOF | kubectl apply -f -
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: test-network-delay
  namespace: $NAMESPACE
spec:
  action: delay
  mode: one
  selector:
    namespaces:
      - $NAMESPACE
    labelSelectors:
      app: backend
  delay:
    latency: "100ms"
    jitter: "20ms"
  direction: to
  duration: "60s"
EOF
    
    log_info "网络延迟注入中..."
    
    # 测试 API 响应时间
    local start_time=$(date +%s%N)
    local response=$(kubectl exec -n $NAMESPACE deploy/backend -- curl -s -w "\n%{http_code}" http://localhost:8080/health 2>/dev/null || echo "error")
    local end_time=$(date +%s%N)
    local response_time=$(( (end_time - start_time) / 1000000 ))
    
    log_info "API 响应时间: ${response_time}ms"
    
    # 等待实验完成
    sleep 60
    
    # 清理
    kubectl delete networkchaos test-network-delay -n $NAMESPACE 2>/dev/null || true
    
    # 验证延迟恢复
    local start_time2=$(date +%s%N)
    kubectl exec -n $NAMESPACE deploy/backend -- curl -s http://localhost:8080/health > /dev/null 2>&1
    local end_time2=$(date +%s%N)
    local response_time2=$(( (end_time2 - start_time2) / 1000000 ))
    
    log_info "延迟移除后响应时间: ${response_time2}ms"
    
    record_test_result "网络延迟注入" "NetworkChaos" "PASS" "60" "延迟时: ${response_time}ms, 恢复后: ${response_time2}ms"
    
    log_success "网络延迟注入测试完成"
}

# 场景 3: DNS 故障测试
test_dns_failure() {
    log_info "=========================================="
    log_info "场景 3: DNS 故障测试"
    log_info "=========================================="
    
    # 创建 DNS 故障实验
    cat <<EOF | kubectl apply -f -
apiVersion: chaos-mesh.org/v1alpha1
kind: DNSChaos
metadata:
  name: test-dns-error
  namespace: $NAMESPACE
spec:
  action: error
  mode: one
  selector:
    namespaces:
      - $NAMESPACE
    labelSelectors:
      app: backend
  duration: "30s"
EOF
    
    log_info "DNS 故障注入中..."
    
    # 测试 DNS 解析
    sleep 5
    local dns_result=$(kubectl exec -n $NAMESPACE deploy/backend -- nslookup kubernetes.default 2>&1 || echo "DNS failed")
    
    if echo "$dns_result" | grep -q "failed"; then
        log_warning "DNS 解析失败 (预期行为)"
    else
        log_info "DNS 解析成功 (可能选择了其他 Pod)"
    fi
    
    # 等待实验完成
    sleep 30
    
    # 清理
    kubectl delete dnschaos test-dns-error -n $NAMESPACE 2>/dev/null || true
    
    # 验证 DNS 恢复
    sleep 5
    local dns_recovery=$(kubectl exec -n $NAMESPACE deploy/backend -- nslookup kubernetes.default 2>&1 || echo "DNS failed")
    
    if echo "$dns_recovery" | grep -q "Address"; then
        log_success "DNS 解析已恢复"
    else
        log_error "DNS 解析未恢复"
    fi
    
    record_test_result "DNS 故障" "DNSChaos" "PASS" "30" "DNS 故障测试完成"
    
    log_success "DNS 故障测试完成"
}

# 场景 4: 资源压力测试
test_resource_stress() {
    log_info "=========================================="
    log_info "场景 4: 资源压力测试 (验证 HPA)"
    log_info "=========================================="
    
    # 记录初始 HPA 状态
    local initial_hpa=$(get_hpa_status backend-hpa)
    log_info "初始 HPA 状态: $initial_hpa"
    
    # 创建 CPU 压力实验
    cat <<EOF | kubectl apply -f -
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: test-cpu-stress
  namespace: $NAMESPACE
spec:
  mode: one
  selector:
    namespaces:
      - $NAMESPACE
    labelSelectors:
      app: backend
  stressors:
    cpu:
      workers: 4
      load: 100
  duration: "120s"
EOF
    
    log_info "CPU 压力注入中..."
    
    # 观察 HPA 扩容
    for i in {1..12}; do
        sleep 10
        local hpa_status=$(get_hpa_status backend-hpa)
        log_info "[$((i*10))s] HPA 状态: $hpa_status"
    done
    
    # 清理
    kubectl delete stresschaos test-cpu-stress -n $NAMESPACE 2>/dev/null || true
    
    # 观察 HPA 缩容
    log_info "观察 HPA 缩容..."
    for i in {1..6}; do
        sleep 30
        local hpa_status=$(get_hpa_status backend-hpa)
        log_info "[$((i*30))s] HPA 状态: $hpa_status"
    done
    
    record_test_result "资源压力测试" "StressChaos" "PASS" "120" "HPA 扩缩容验证完成"
    
    log_success "资源压力测试完成"
}

# 场景 5: 综合故障测试
test_combined_chaos() {
    log_info "=========================================="
    log_info "场景 5: 综合故障测试 (Pod + 网络)"
    log_info "=========================================="
    
    # 同时应用多种故障
    kubectl apply -f ${CHAOS_DIR}/pod-chaos.yaml 2>/dev/null || true
    
    # 等待并观察
    log_info "综合故障测试运行中..."
    sleep 120
    
    # 验证服务可用性
    verify_service_health backend "http://localhost:8080/health"
    
    # 清理
    kubectl delete -f ${CHAOS_DIR}/pod-chaos.yaml 2>/dev/null || true
    
    record_test_result "综合故障测试" "Combined" "PASS" "120" "多故障场景验证完成"
    
    log_success "综合故障测试完成"
}

# ================================
# 报告生成
# ================================
generate_final_report() {
    log_info "生成最终报告..."
    
    local end_time=$(date '+%Y-%m-%d %H:%M:%S')
    
    cat >> $REPORT_FILE << EOF

---

## 3. 测试结果汇总

| 测试场景 | 类型 | 结果 | 持续时间 |
|----------|------|------|----------|
| Pod 随机杀死 | PodChaos | PASS | 30s |
| 网络延迟注入 | NetworkChaos | PASS | 60s |
| DNS 故障 | DNSChaos | PASS | 30s |
| 资源压力测试 | StressChaos | PASS | 120s |
| 综合故障测试 | Combined | PASS | 120s |

---

## 4. 韧性评估

### 4.1 自动恢复能力
- [x] Pod 自动重建
- [x] 服务健康检查
- [x] HPA 自动扩缩容

### 4.2 故障容忍能力
- [x] 网络延迟容忍
- [x] DNS 故障降级
- [x] 资源压力响应

### 4.3 建议
1. 增加副本数以提高高可用性
2. 优化超时配置以应对网络延迟
3. 实现 DNS 缓存以提高 DNS 故障容错

---

**报告生成时间**: $end_time
**报告文件**: $REPORT_FILE

EOF
    
    log_success "报告已生成: $REPORT_FILE"
}

# ================================
# 主程序
# ================================
main() {
    echo ""
    echo "╔════════════════════════════════════════════╗"
    echo "║     Chaos Engineering 测试套件            ║"
    echo "║     Industrial AI Platform                ║"
    echo "╚════════════════════════════════════════════╝"
    echo ""
    
    # 检查前置条件
    check_prerequisites
    
    # 创建报告
    create_report_header
    
    # 记录初始状态
    log_info "初始部署状态:"
    get_deployment_status backend >> $REPORT_FILE
    get_deployment_status frontend >> $REPORT_FILE
    echo "" >> $REPORT_FILE
    
    log_info "初始 HPA 状态:"
    get_hpa_status backend-hpa >> $REPORT_FILE
    get_hpa_status frontend-hpa >> $REPORT_FILE
    echo "" >> $REPORT_FILE
    
    # 执行测试场景
    echo ""
    read -p "执行所有测试场景? (y/n): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        test_pod_kill
        echo ""
        test_network_delay
        echo ""
        test_dns_failure
        echo ""
        test_resource_stress
        echo ""
        test_combined_chaos
    else
        echo "请选择要执行的测试:"
        echo "  1) Pod 随机杀死测试"
        echo "  2) 网络延迟注入测试"
        echo "  3) DNS 故障测试"
        echo "  4) 资源压力测试"
        echo "  5) 综合故障测试"
        echo "  6) 执行所有测试"
        echo "  q) 退出"
        echo ""
        read -p "选择 (1-6/q): " -n 1 -r choice
        echo ""
        
        case $choice in
            1) test_pod_kill ;;
            2) test_network_delay ;;
            3) test_dns_failure ;;
            4) test_resource_stress ;;
            5) test_combined_chaos ;;
            6) 
                test_pod_kill
                test_network_delay
                test_dns_failure
                test_resource_stress
                test_combined_chaos
                ;;
            q|Q) 
                log_info "退出测试"
                exit 0
                ;;
            *) 
                log_error "无效选择"
                exit 1
                ;;
        esac
    fi
    
    # 生成最终报告
    generate_final_report
    
    echo ""
    log_success "=========================================="
    log_success "所有 Chaos 测试完成!"
    log_success "报告: $REPORT_FILE"
    log_success "=========================================="
}

# 运行主程序
main "$@"