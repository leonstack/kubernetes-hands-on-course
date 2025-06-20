#!/bin/bash

# 演示如何使用 YAML 文件创建和管理 ReplicaSet 及 Service
# 作者: Grissom
# 版本: 1.0.0
# 日期: 2025-06-20

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 脚本信息
show_info() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Kubernetes ReplicaSet with YAML Demo${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${BLUE}功能说明：${NC}"
    echo "  • 演示 ReplicaSet YAML 文件的创建和验证"
    echo "  • 创建和管理 ReplicaSet 资源"
    echo "  • 测试 ReplicaSet 的自动恢复机制"
    echo "  • 演示 ReplicaSet 的扩缩容操作"
    echo "  • 创建和管理 Service 资源"
    echo "  • 网络连接测试和负载均衡验证"
    echo "  • 故障排除和调试技巧"
    echo "  • 最佳实践演示"
    echo ""
}

# 帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --create-all       创建所有资源（ReplicaSet + Service）"
    echo "  --create-rs        仅创建 ReplicaSet"
    echo "  --create-service   仅创建 Service"
    echo "  --test-recovery    测试自动恢复机制"
    echo "  --scale            演示扩缩容操作"
    echo "  --test-lb          测试负载均衡"
    echo "  --monitor          监控资源状态"
    echo "  --logs             查看 Pod 日志"
    echo "  --debug            调试模式"
    echo "  --cleanup          清理所有资源"
    echo "  --validate         验证 YAML 文件"
    echo "  --best-practices   最佳实践演示"
    echo "  --help             显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --create-all      # 创建所有资源"
    echo "  $0 --test-recovery   # 测试自动恢复"
    echo "  $0 --scale           # 演示扩缩容"
    echo "  $0 --cleanup         # 清理资源"
}

# 检查依赖
check_dependencies() {
    log_step "检查依赖工具..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在 PATH 中"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 验证 YAML 文件
validate_yaml() {
    log_step "验证 YAML 文件语法..."
    
    local files=(
        "kube-manifests/01-kube-base-definition.yml"
        "kube-manifests/02-replicaset-definition.yml"
        "kube-manifests/03-replicaset-nodeport-servie.yml"
    )
    
    for file in "${files[@]}"; do
        if [[ -f "$file" ]]; then
            log_info "验证文件: $file"
            if kubectl apply --dry-run=client -f "$file" &> /dev/null; then
                log_success "✓ $file 语法正确"
            else
                log_error "✗ $file 语法错误"
                kubectl apply --dry-run=client -f "$file"
                return 1
            fi
        else
            log_warning "文件不存在: $file"
        fi
    done
    
    log_success "YAML 文件验证完成"
}

# 创建 ReplicaSet
create_replicaset() {
    log_step "创建 ReplicaSet..."
    
    if kubectl get replicaset myapp2-rs &> /dev/null; then
        log_warning "ReplicaSet 'myapp2-rs' 已存在"
        return 0
    fi
    
    log_info "应用 ReplicaSet 配置文件..."
    kubectl apply -f kube-manifests/02-replicaset-definition.yml
    
    log_info "等待 ReplicaSet 就绪..."
    kubectl wait --for=condition=Ready pod -l app=myapp2,version=v2.0 --timeout=120s
    
    log_success "ReplicaSet 创建成功"
    kubectl get replicaset myapp2-rs
    kubectl get pods -l app=myapp2 -o wide
}

# 创建 Service
create_service() {
    log_step "创建 Service..."
    
    if kubectl get service replicaset-nodeport-service &> /dev/null; then
        log_warning "Service 'replicaset-nodeport-service' 已存在"
        return 0
    fi
    
    log_info "应用 Service 配置文件..."
    kubectl apply -f kube-manifests/03-replicaset-nodeport-servie.yml
    
    log_success "Service 创建成功"
    kubectl get service replicaset-nodeport-service
    
    # 显示访问信息
    show_access_info
}

# 显示访问信息
show_access_info() {
    log_step "获取访问信息..."
    
    local node_ip
    local node_port
    
    # 获取节点 IP
    node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' 2>/dev/null || \
              kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    
    # 获取 NodePort
    node_port=$(kubectl get service replicaset-nodeport-service -o jsonpath='{.spec.ports[0].nodePort}')
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  应用访问信息${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${GREEN}外部访问地址:${NC} http://${node_ip}:${node_port}"
    echo -e "${GREEN}集群内访问:${NC} http://replicaset-nodeport-service:80"
    echo -e "${GREEN}Service IP:${NC} $(kubectl get service replicaset-nodeport-service -o jsonpath='{.spec.clusterIP}')"
    echo -e "${GREEN}Pod 数量:${NC} $(kubectl get pods -l app=myapp2 --no-headers | wc -l)"
    echo ""
}

# 测试自动恢复机制
test_recovery() {
    log_step "测试 ReplicaSet 自动恢复机制..."
    
    # 检查 ReplicaSet 是否存在
    if ! kubectl get replicaset myapp2-rs &> /dev/null; then
        log_error "ReplicaSet 'myapp2-rs' 不存在，请先创建"
        return 1
    fi
    
    # 获取当前 Pod 列表
    log_info "当前 Pod 状态:"
    kubectl get pods -l app=myapp2 -o wide
    
    # 选择一个 Pod 进行删除
    local pod_name
    pod_name=$(kubectl get pods -l app=myapp2 -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$pod_name" ]]; then
        log_error "没有找到可用的 Pod"
        return 1
    fi
    
    log_info "删除 Pod: $pod_name（模拟故障）"
    kubectl delete pod "$pod_name"
    
    log_info "观察 ReplicaSet 自动恢复过程..."
    sleep 2
    
    # 监控恢复过程
    local count=0
    while [[ $count -lt 30 ]]; do
        local ready_pods
        ready_pods=$(kubectl get pods -l app=myapp2 --field-selector=status.phase=Running --no-headers | wc -l)
        
        echo -e "${BLUE}[监控]${NC} 就绪 Pod 数量: $ready_pods/3"
        
        if [[ $ready_pods -eq 3 ]]; then
            log_success "✓ ReplicaSet 自动恢复完成"
            break
        fi
        
        sleep 2
        ((count++))
    done
    
    log_info "恢复后的 Pod 状态:"
    kubectl get pods -l app=myapp2 -o wide
    
    log_info "ReplicaSet 事件:"
    kubectl describe replicaset myapp2-rs | tail -10
}

# 演示扩缩容操作
demonstrate_scaling() {
    log_step "演示 ReplicaSet 扩缩容操作..."
    
    # 检查 ReplicaSet 是否存在
    if ! kubectl get replicaset myapp2-rs &> /dev/null; then
        log_error "ReplicaSet 'myapp2-rs' 不存在，请先创建"
        return 1
    fi
    
    log_info "当前 ReplicaSet 状态:"
    kubectl get replicaset myapp2-rs
    kubectl get pods -l app=myapp2
    
    # 扩容到 5 个副本
    log_info "扩容到 5 个副本..."
    kubectl scale replicaset myapp2-rs --replicas=5
    
    log_info "等待扩容完成..."
    kubectl wait --for=condition=Ready pod -l app=myapp2 --timeout=60s
    
    log_success "扩容完成"
    kubectl get replicaset myapp2-rs
    kubectl get pods -l app=myapp2 -o wide
    
    sleep 3
    
    # 缩容到 2 个副本
    log_info "缩容到 2 个副本..."
    kubectl scale replicaset myapp2-rs --replicas=2
    
    sleep 5
    
    log_success "缩容完成"
    kubectl get replicaset myapp2-rs
    kubectl get pods -l app=myapp2 -o wide
    
    sleep 3
    
    # 恢复到原始副本数
    log_info "恢复到 3 个副本..."
    kubectl scale replicaset myapp2-rs --replicas=3
    
    kubectl wait --for=condition=Ready pod -l app=myapp2 --timeout=60s
    
    log_success "副本数恢复完成"
    kubectl get replicaset myapp2-rs
    kubectl get pods -l app=myapp2 -o wide
}

# 测试负载均衡
test_load_balancing() {
    log_step "测试负载均衡..."
    
    # 检查 Service 是否存在
    if ! kubectl get service replicaset-nodeport-service &> /dev/null; then
        log_error "Service 'replicaset-nodeport-service' 不存在，请先创建"
        return 1
    fi
    
    # 检查 Pod 是否就绪
    local ready_pods
    ready_pods=$(kubectl get pods -l app=myapp2 --field-selector=status.phase=Running --no-headers | wc -l)
    
    if [[ $ready_pods -eq 0 ]]; then
        log_error "没有就绪的 Pod"
        return 1
    fi
    
    log_info "当前就绪的 Pod 数量: $ready_pods"
    kubectl get pods -l app=myapp2 -o wide
    
    # 测试集群内部访问
    log_info "测试集群内部负载均衡..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  负载均衡测试结果${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    for i in {1..10}; do
        local response
        response=$(kubectl run test-pod-$i --image=busybox --rm -i --restart=Never --quiet -- \
            wget -qO- --timeout=5 "http://replicaset-nodeport-service:80" 2>/dev/null | \
            grep -i "hostname\|server" | head -1 || echo "请求 $i: 连接失败")
        
        if [[ -n "$response" ]]; then
            echo -e "${GREEN}请求 $i:${NC} $response"
        else
            echo -e "${RED}请求 $i:${NC} 无响应"
        fi
        
        sleep 1
    done
    
    echo ""
    log_info "负载均衡测试完成"
    
    # 显示 Endpoints
    log_info "Service Endpoints:"
    kubectl get endpoints replicaset-nodeport-service
}

# 监控资源状态
monitor_resources() {
    log_step "监控资源状态..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  ReplicaSet 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get replicasets -l app=myapp2 -o wide
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get pods -l app=myapp2 -o wide
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Service 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get services -l app=myapp2
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Endpoints 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get endpoints replicaset-nodeport-service 2>/dev/null || log_warning "Service 不存在"
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  资源使用情况${NC}"
    echo -e "${CYAN}========================================${NC}"
    if command -v kubectl &> /dev/null && kubectl top pods -l app=myapp2 &> /dev/null; then
        kubectl top pods -l app=myapp2
    else
        log_warning "metrics-server 未安装或不可用"
    fi
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  最近事件${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get events --sort-by=.metadata.creationTimestamp --field-selector involvedObject.name=myapp2-rs | tail -10
}

# 查看日志
view_logs() {
    log_step "查看 Pod 日志..."
    
    local pods
    pods=$(kubectl get pods -l app=myapp2 -o jsonpath='{.items[*].metadata.name}')
    
    if [[ -z "$pods" ]]; then
        log_error "没有找到相关的 Pod"
        return 1
    fi
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 日志 (最近 20 行)${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    for pod in $pods; do
        echo -e "${GREEN}--- Pod: $pod ---${NC}"
        kubectl logs "$pod" --tail=20 || log_warning "无法获取 Pod $pod 的日志"
        echo ""
    done
    
    log_info "实时日志查看命令: kubectl logs -l app=myapp2 -f --tail=50"
}

# 调试模式
debug_mode() {
    log_step "进入调试模式..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  调试信息收集${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    # 集群信息
    log_info "集群信息:"
    kubectl cluster-info
    
    echo ""
    # 节点信息
    log_info "节点信息:"
    kubectl get nodes -o wide
    
    echo ""
    # 命名空间资源
    log_info "当前命名空间资源:"
    kubectl get all -l app=myapp2
    
    echo ""
    # ReplicaSet 详细状态
    if kubectl get replicaset myapp2-rs &> /dev/null; then
        log_info "ReplicaSet 详细状态:"
        kubectl describe replicaset myapp2-rs
    fi
    
    echo ""
    # Service 详细状态
    if kubectl get service replicaset-nodeport-service &> /dev/null; then
        log_info "Service 详细状态:"
        kubectl describe service replicaset-nodeport-service
    fi
    
    echo ""
    # Pod 详细状态
    log_info "Pod 详细状态:"
    local pods
    pods=$(kubectl get pods -l app=myapp2 -o jsonpath='{.items[*].metadata.name}')
    for pod in $pods; do
        echo -e "${GREEN}--- Pod: $pod ---${NC}"
        kubectl describe pod "$pod" | head -30
        echo ""
    done
    
    echo ""
    # 事件信息
    log_info "相关事件:"
    kubectl get events --sort-by=.metadata.creationTimestamp | grep -E "myapp2|replicaset" | tail -20
}

# 最佳实践演示
show_best_practices() {
    log_step "最佳实践演示..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Kubernetes ReplicaSet 最佳实践${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    echo -e "${GREEN}1. ReplicaSet 配置最佳实践:${NC}"
    echo "   • 使用精确的标签选择器"
    echo "   • 设置资源请求和限制"
    echo "   • 配置健康检查探针"
    echo "   • 使用有意义的标签和注解"
    
    echo ""
    echo -e "${GREEN}2. 标签选择器最佳实践:${NC}"
    echo "   • 使用多个标签提高选择精度"
    echo "   • 避免使用过于宽泛的标签"
    echo "   • Pod 模板标签必须匹配选择器"
    echo "   • 标签选择器创建后不可修改"
    
    echo ""
    echo -e "${GREEN}3. 资源管理最佳实践:${NC}"
    echo "   • 设置合理的资源请求和限制"
    echo "   • 使用 HPA 进行自动扩缩容"
    echo "   • 监控资源使用情况"
    echo "   • 定期检查和优化配置"
    
    echo ""
    echo -e "${GREEN}4. 安全最佳实践:${NC}"
    echo "   • 使用非 root 用户运行容器"
    echo "   • 设置安全上下文"
    echo "   • 使用只读根文件系统"
    echo "   • 定期更新镜像版本"
    
    echo ""
    echo -e "${GREEN}5. 运维最佳实践:${NC}"
    echo "   • 使用声明式配置管理"
    echo "   • 版本控制 YAML 文件"
    echo "   • 实施监控和日志收集"
    echo "   • 定期备份配置"
    
    echo ""
    echo -e "${GREEN}6. 网络最佳实践:${NC}"
    echo "   • 使用 Service 提供稳定的网络入口"
    echo "   • 配置适当的会话亲和性"
    echo "   • 考虑外部流量策略"
    echo "   • 生产环境使用 LoadBalancer 或 Ingress"
    
    # 显示改进的配置示例
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  生产级 ReplicaSet 配置示例${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    cat << 'EOF'
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: production-app-rs
  labels:
    app: production-app
    version: v1.0
    tier: frontend
    environment: production
  annotations:
    description: "Production web application ReplicaSet"
    maintainer: "devops-team@company.com"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: production-app
      version: v1.0
  template:
    metadata:
      labels:
        app: production-app
        version: v1.0
        tier: frontend
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 2000
      containers:
      - name: app-container
        image: myapp:1.0.0
        imagePullPolicy: IfNotPresent
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        ports:
        - containerPort: 8080
          name: http
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
EOF
}

# 清理资源
cleanup_resources() {
    log_step "清理资源..."
    
    # 删除 Service
    if kubectl get service replicaset-nodeport-service &> /dev/null; then
        log_info "删除 Service..."
        kubectl delete service replicaset-nodeport-service
        log_success "Service 已删除"
    fi
    
    # 删除 ReplicaSet（会同时删除管理的 Pod）
    if kubectl get replicaset myapp2-rs &> /dev/null; then
        log_info "删除 ReplicaSet..."
        kubectl delete replicaset myapp2-rs
        log_success "ReplicaSet 已删除"
    fi
    
    # 清理测试 Pod（如果存在）
    kubectl delete pod --selector=run=test-pod --ignore-not-found=true &> /dev/null
    
    # 等待资源完全删除
    log_info "等待资源完全删除..."
    sleep 5
    
    # 验证清理结果
    local remaining_resources
    remaining_resources=$(kubectl get rs,svc,pods -l app=myapp2 --no-headers 2>/dev/null | wc -l)
    
    if [[ $remaining_resources -eq 0 ]]; then
        log_success "所有资源已成功清理"
    else
        log_warning "仍有 $remaining_resources 个相关资源存在"
        kubectl get rs,svc,pods -l app=myapp2
    fi
}

# 交互式菜单
interactive_menu() {
    while true; do
        echo ""
        echo -e "${CYAN}========================================${NC}"
        echo -e "${CYAN}  ReplicaSet YAML Demo 菜单${NC}"
        echo -e "${CYAN}========================================${NC}"
        echo "1. 验证 YAML 文件"
        echo "2. 创建 ReplicaSet"
        echo "3. 创建 Service"
        echo "4. 创建所有资源"
        echo "5. 测试自动恢复机制"
        echo "6. 演示扩缩容操作"
        echo "7. 测试负载均衡"
        echo "8. 监控资源状态"
        echo "9. 查看日志"
        echo "10. 调试模式"
        echo "11. 最佳实践演示"
        echo "12. 清理资源"
        echo "0. 退出"
        echo ""
        read -p "请选择操作 (0-12): " choice
        
        case $choice in
            1) validate_yaml ;;
            2) create_replicaset ;;
            3) create_service ;;
            4) create_replicaset && create_service ;;
            5) test_recovery ;;
            6) demonstrate_scaling ;;
            7) test_load_balancing ;;
            8) monitor_resources ;;
            9) view_logs ;;
            10) debug_mode ;;
            11) show_best_practices ;;
            12) cleanup_resources ;;
            0) log_info "退出程序"; exit 0 ;;
            *) log_error "无效选择，请重试" ;;
        esac
        
        echo ""
        read -p "按 Enter 键继续..."
    done
}

# 主函数
main() {
    show_info
    check_dependencies
    
    # 解析命令行参数
    case "${1:-}" in
        --create-all)
            validate_yaml
            create_replicaset
            create_service
            ;;
        --create-rs)
            validate_yaml
            create_replicaset
            ;;
        --create-service)
            validate_yaml
            create_service
            ;;
        --test-recovery)
            test_recovery
            ;;
        --scale)
            demonstrate_scaling
            ;;
        --test-lb)
            test_load_balancing
            ;;
        --monitor)
            monitor_resources
            ;;
        --logs)
            view_logs
            ;;
        --debug)
            debug_mode
            ;;
        --cleanup)
            cleanup_resources
            ;;
        --validate)
            validate_yaml
            ;;
        --best-practices)
            show_best_practices
            ;;
        --help)
            show_help
            ;;
        "")
            interactive_menu
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 错误处理
trap 'log_error "脚本执行出错，退出码: $?"' ERR

# 执行主函数
main "$@"