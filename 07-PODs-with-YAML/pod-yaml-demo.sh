#!/bin/bash

# 演示如何使用 YAML 文件创建和管理 Pod 及 Service
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
    echo -e "${CYAN}  Kubernetes Pod with YAML Demo${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${BLUE}功能说明：${NC}"
    echo "  • 演示 YAML 文件的创建和验证"
    echo "  • 创建和管理 Pod 资源"
    echo "  • 创建和管理 Service 资源"
    echo "  • 网络连接测试和故障排除"
    echo "  • 资源监控和日志查看"
    echo "  • 最佳实践演示"
    echo ""
}

# 帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --create-all     创建所有资源（Pod + Service）"
    echo "  --create-pod     仅创建 Pod"
    echo "  --create-service 仅创建 Service"
    echo "  --test           测试网络连接"
    echo "  --monitor        监控资源状态"
    echo "  --logs           查看 Pod 日志"
    echo "  --debug          调试模式"
    echo "  --cleanup        清理所有资源"
    echo "  --validate       验证 YAML 文件"
    echo "  --best-practices 最佳实践演示"
    echo "  --help           显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --create-all    # 创建所有资源"
    echo "  $0 --test          # 测试连接"
    echo "  $0 --cleanup       # 清理资源"
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
        "kube-manifests/02-pod-definition.yml"
        "kube-manifests/03-pod-nodeport-service.yml"
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

# 创建 Pod
create_pod() {
    log_step "创建 Pod..."
    
    if kubectl get pod myapp-pod &> /dev/null; then
        log_warning "Pod 'myapp-pod' 已存在"
        return 0
    fi
    
    log_info "应用 Pod 配置文件..."
    kubectl apply -f kube-manifests/02-pod-definition.yml
    
    log_info "等待 Pod 启动..."
    kubectl wait --for=condition=Ready pod/myapp-pod --timeout=120s
    
    log_success "Pod 创建成功"
    kubectl get pod myapp-pod -o wide
}

# 创建 Service
create_service() {
    log_step "创建 Service..."
    
    if kubectl get service myapp-pod-nodeport-service &> /dev/null; then
        log_warning "Service 'myapp-pod-nodeport-service' 已存在"
        return 0
    fi
    
    log_info "应用 Service 配置文件..."
    kubectl apply -f kube-manifests/03-pod-nodeport-service.yml
    
    log_success "Service 创建成功"
    kubectl get service myapp-pod-nodeport-service
    
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
    node_port=$(kubectl get service myapp-pod-nodeport-service -o jsonpath='{.spec.ports[0].nodePort}')
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  应用访问信息${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${GREEN}外部访问地址:${NC} http://${node_ip}:${node_port}"
    echo -e "${GREEN}集群内访问:${NC} http://myapp-pod-nodeport-service:80"
    echo -e "${GREEN}Pod IP:${NC} $(kubectl get pod myapp-pod -o jsonpath='{.status.podIP}')"
    echo -e "${GREEN}Service IP:${NC} $(kubectl get service myapp-pod-nodeport-service -o jsonpath='{.spec.clusterIP}')"
    echo ""
}

# 测试网络连接
test_connectivity() {
    log_step "测试网络连接..."
    
    # 检查 Pod 是否运行
    if ! kubectl get pod myapp-pod &> /dev/null; then
        log_error "Pod 'myapp-pod' 不存在，请先创建"
        return 1
    fi
    
    # 检查 Service 是否存在
    if ! kubectl get service myapp-pod-nodeport-service &> /dev/null; then
        log_error "Service 'myapp-pod-nodeport-service' 不存在，请先创建"
        return 1
    fi
    
    # 测试 Pod 直接访问
    log_info "测试 Pod 直接访问..."
    local pod_ip
    pod_ip=$(kubectl get pod myapp-pod -o jsonpath='{.status.podIP}')
    
    if kubectl run test-pod --image=busybox --rm -i --restart=Never -- \
        wget -qO- --timeout=10 "http://${pod_ip}:80" &> /dev/null; then
        log_success "✓ Pod 直接访问成功"
    else
        log_error "✗ Pod 直接访问失败"
    fi
    
    # 测试 Service 访问
    log_info "测试 Service 访问..."
    if kubectl run test-pod --image=busybox --rm -i --restart=Never -- \
        wget -qO- --timeout=10 "http://myapp-pod-nodeport-service:80" &> /dev/null; then
        log_success "✓ Service 访问成功"
    else
        log_error "✗ Service 访问失败"
    fi
    
    # 测试 NodePort 访问（如果可能）
    log_info "测试 NodePort 访问..."
    local node_ip
    local node_port
    
    node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    node_port=$(kubectl get service myapp-pod-nodeport-service -o jsonpath='{.spec.ports[0].nodePort}')
    
    if kubectl run test-pod --image=busybox --rm -i --restart=Never -- \
        wget -qO- --timeout=10 "http://${node_ip}:${node_port}" &> /dev/null; then
        log_success "✓ NodePort 访问成功"
    else
        log_warning "✗ NodePort 访问失败（可能需要外部网络访问）"
    fi
    
    log_success "网络连接测试完成"
}

# 监控资源状态
monitor_resources() {
    log_step "监控资源状态..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get pods -l app=myapp -o wide
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Service 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get services -l app=myapp
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Endpoints 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get endpoints myapp-pod-nodeport-service 2>/dev/null || log_warning "Service 不存在"
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 详细信息${NC}"
    echo -e "${CYAN}========================================${NC}"
    if kubectl get pod myapp-pod &> /dev/null; then
        kubectl describe pod myapp-pod | head -20
    else
        log_warning "Pod 不存在"
    fi
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  最近事件${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get events --sort-by=.metadata.creationTimestamp | tail -10
}

# 查看日志
view_logs() {
    log_step "查看 Pod 日志..."
    
    if ! kubectl get pod myapp-pod &> /dev/null; then
        log_error "Pod 'myapp-pod' 不存在"
        return 1
    fi
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 日志 (最近 50 行)${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl logs myapp-pod --tail=50
    
    echo ""
    log_info "实时日志查看命令: kubectl logs -f myapp-pod"
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
    kubectl get all
    
    echo ""
    # Pod 详细状态
    if kubectl get pod myapp-pod &> /dev/null; then
        log_info "Pod 详细状态:"
        kubectl describe pod myapp-pod
    fi
    
    echo ""
    # Service 详细状态
    if kubectl get service myapp-pod-nodeport-service &> /dev/null; then
        log_info "Service 详细状态:"
        kubectl describe service myapp-pod-nodeport-service
    fi
    
    echo ""
    # 网络策略
    log_info "网络策略:"
    kubectl get networkpolicies 2>/dev/null || log_info "无网络策略"
    
    echo ""
    # 存储类
    log_info "存储类:"
    kubectl get storageclass 2>/dev/null || log_info "无存储类"
}

# 最佳实践演示
show_best_practices() {
    log_step "最佳实践演示..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Kubernetes YAML 最佳实践${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    echo -e "${GREEN}1. 资源配置最佳实践:${NC}"
    echo "   • 始终设置资源请求和限制"
    echo "   • 使用有意义的标签和注解"
    echo "   • 配置健康检查探针"
    echo "   • 设置适当的重启策略"
    
    echo ""
    echo -e "${GREEN}2. 安全最佳实践:${NC}"
    echo "   • 使用非 root 用户运行容器"
    echo "   • 设置只读根文件系统"
    echo "   • 禁用特权提升"
    echo "   • 使用安全上下文"
    
    echo ""
    echo -e "${GREEN}3. 网络最佳实践:${NC}"
    echo "   • 使用有意义的端口名称"
    echo "   • 配置适当的会话亲和性"
    echo "   • 考虑外部流量策略"
    echo "   • 使用网络策略限制流量"
    
    echo ""
    echo -e "${GREEN}4. 监控和日志最佳实践:${NC}"
    echo "   • 配置结构化日志"
    echo "   • 使用标准化的标签"
    echo "   • 设置适当的日志级别"
    echo "   • 配置监控指标"
    
    echo ""
    echo -e "${GREEN}5. 部署最佳实践:${NC}"
    echo "   • 使用声明式配置"
    echo "   • 版本控制 YAML 文件"
    echo "   • 使用命名空间隔离"
    echo "   • 实施 GitOps 工作流"
    
    # 显示改进的 YAML 示例
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  改进的 Pod 配置示例${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    cat << 'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod-improved
  labels:
    app: myapp
    version: v1.0
    tier: frontend
    environment: production
  annotations:
    description: "Improved web application pod"
    maintainer: "devops-team@company.com"
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
  - name: myapp
    image: grissomsh/kubenginx:1.0.0
    imagePullPolicy: IfNotPresent
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
    ports:
    - containerPort: 80
      name: http
      protocol: TCP
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
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
    env:
    - name: APP_ENV
      value: "production"
    - name: LOG_LEVEL
      value: "info"
  restartPolicy: Always
EOF
}

# 清理资源
cleanup_resources() {
    log_step "清理资源..."
    
    # 删除 Service
    if kubectl get service myapp-pod-nodeport-service &> /dev/null; then
        log_info "删除 Service..."
        kubectl delete service myapp-pod-nodeport-service
        log_success "Service 已删除"
    fi
    
    # 删除 Pod
    if kubectl get pod myapp-pod &> /dev/null; then
        log_info "删除 Pod..."
        kubectl delete pod myapp-pod
        log_success "Pod 已删除"
    fi
    
    # 清理测试 Pod（如果存在）
    kubectl delete pod test-pod --ignore-not-found=true &> /dev/null
    
    log_success "资源清理完成"
}

# 交互式菜单
interactive_menu() {
    while true; do
        echo ""
        echo -e "${CYAN}========================================${NC}"
        echo -e "${CYAN}  Kubernetes Pod YAML Demo 菜单${NC}"
        echo -e "${CYAN}========================================${NC}"
        echo "1. 验证 YAML 文件"
        echo "2. 创建 Pod"
        echo "3. 创建 Service"
        echo "4. 创建所有资源"
        echo "5. 测试网络连接"
        echo "6. 监控资源状态"
        echo "7. 查看日志"
        echo "8. 调试模式"
        echo "9. 最佳实践演示"
        echo "10. 清理资源"
        echo "0. 退出"
        echo ""
        read -p "请选择操作 (0-10): " choice
        
        case $choice in
            1) validate_yaml ;;
            2) create_pod ;;
            3) create_service ;;
            4) create_pod && create_service ;;
            5) test_connectivity ;;
            6) monitor_resources ;;
            7) view_logs ;;
            8) debug_mode ;;
            9) show_best_practices ;;
            10) cleanup_resources ;;
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
            create_pod
            create_service
            ;;
        --create-pod)
            validate_yaml
            create_pod
            ;;
        --create-service)
            validate_yaml
            create_service
            ;;
        --test)
            test_connectivity
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