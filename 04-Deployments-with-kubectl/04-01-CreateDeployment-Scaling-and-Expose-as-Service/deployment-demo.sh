#!/bin/bash

# 本脚本演示 Deployment 的创建、扩展、暴露服务等操作
# 作者: Grissom
# 版本: 1.0.0
# 日期: 2025-06-20

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
    echo -e "\n${PURPLE}=== $1 ===${NC}"
}

# 等待用户输入
wait_for_user() {
    echo -e "\n${CYAN}按 Enter 键继续...${NC}"
    read -r
}

# 检查 kubectl 是否可用
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 命令未找到，请先安装 kubectl"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群，请检查配置"
        exit 1
    fi
    
    log_success "kubectl 可用，已连接到 Kubernetes 集群"
}

# 清理现有资源
cleanup_existing() {
    log_step "清理现有资源"
    
    # 检查是否存在同名资源
    if kubectl get deployment my-first-deployment &> /dev/null; then
        log_warning "发现现有的 Deployment，正在清理..."
        kubectl delete deployment my-first-deployment --ignore-not-found=true
    fi
    
    if kubectl get service my-first-deployment-service &> /dev/null; then
        log_warning "发现现有的 Service，正在清理..."
        kubectl delete service my-first-deployment-service --ignore-not-found=true
    fi
    
    # 等待资源完全删除
    log_info "等待资源完全删除..."
    sleep 5
    
    log_success "清理完成"
}

# 创建 Deployment
create_deployment() {
    log_step "步骤 1: 创建 Deployment"
    
    log_info "创建名为 my-first-deployment 的 Deployment"
    kubectl create deployment my-first-deployment --image=grissomsh/kubenginx:1.0.0
    
    if [ $? -eq 0 ]; then
        log_success "Deployment 创建成功"
    else
        log_error "Deployment 创建失败"
        exit 1
    fi
    
    wait_for_user
    
    log_info "验证 Deployment 状态"
    kubectl get deployments
    kubectl get deploy -o wide
    
    wait_for_user
    
    log_info "查看 Deployment 详细信息"
    kubectl describe deployment my-first-deployment
    
    wait_for_user
    
    log_info "验证 ReplicaSet"
    kubectl get rs
    kubectl get rs -l app=my-first-deployment
    
    wait_for_user
    
    log_info "验证 Pod"
    kubectl get pods
    kubectl get pods -l app=my-first-deployment -o wide
    
    wait_for_user
    
    log_info "查看部署层次结构"
    echo "=== Deployment 信息 ==="
    kubectl get deployment my-first-deployment
    
    echo -e "\n=== ReplicaSet 信息 ==="
    kubectl get rs -l app=my-first-deployment
    
    echo -e "\n=== Pod 信息 ==="
    kubectl get pods -l app=my-first-deployment
    
    wait_for_user
    
    log_info "监控部署状态"
    kubectl rollout status deployment/my-first-deployment
    
    log_success "Deployment 创建和验证完成"
}

# 扩展 Deployment
scale_deployment() {
    log_step "步骤 2: 扩展 Deployment"
    
    log_info "当前副本数：$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.replicas}')"
    
    log_info "扩展到 5 个副本"
    kubectl scale --replicas=5 deployment/my-first-deployment
    
    log_info "等待扩展完成..."
    kubectl rollout status deployment/my-first-deployment
    
    wait_for_user
    
    log_info "验证扩展结果"
    kubectl get deployment my-first-deployment
    kubectl get pods -l app=my-first-deployment -o wide
    
    wait_for_user
    
    log_info "继续扩展到 10 个副本"
    kubectl scale --replicas=10 deployment/my-first-deployment
    
    log_info "实时监控扩展过程（10秒）"
    kubectl get pods -l app=my-first-deployment -w &
    WATCH_PID=$!
    sleep 10
    kill $WATCH_PID 2>/dev/null
    
    wait_for_user
    
    log_info "查看扩展事件"
    kubectl get events --sort-by=.metadata.creationTimestamp | grep -E "(Scaling|SuccessfulCreate)" | tail -10
    
    wait_for_user
    
    log_info "缩减到 3 个副本"
    kubectl scale --replicas=3 deployment/my-first-deployment
    kubectl rollout status deployment/my-first-deployment
    
    log_info "最终副本状态"
    kubectl get deployment my-first-deployment
    kubectl get pods -l app=my-first-deployment
    
    log_success "Deployment 扩缩容完成"
}

# 暴露为 Service
expose_service() {
    log_step "步骤 3: 将 Deployment 暴露为 Service"
    
    log_info "创建 NodePort Service"
    kubectl expose deployment my-first-deployment --type=NodePort --port=80 --target-port=80 --name=my-first-deployment-service
    
    if [ $? -eq 0 ]; then
        log_success "Service 创建成功"
    else
        log_error "Service 创建失败"
        exit 1
    fi
    
    wait_for_user
    
    log_info "查看 Service 信息"
    kubectl get svc
    kubectl get svc my-first-deployment-service -o wide
    
    wait_for_user
    
    log_info "查看 Service 详细信息"
    kubectl describe svc my-first-deployment-service
    
    wait_for_user
    
    log_info "获取访问信息"
    NODE_PORT=$(kubectl get svc my-first-deployment-service -o jsonpath='{.spec.ports[0].nodePort}')
    echo "NodePort: $NODE_PORT"
    
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    if [ -z "$NODE_IP" ]; then
        NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    fi
    echo "Node IP: $NODE_IP"
    
    echo -e "\n${GREEN}访问 URL: http://$NODE_IP:$NODE_PORT${NC}"
    
    wait_for_user
    
    log_info "验证 Service 端点"
    kubectl get endpoints my-first-deployment-service
    
    wait_for_user
    
    log_info "测试服务连接"
    echo "尝试访问服务..."
    
    # 使用 port-forward 进行本地测试
    log_info "启动端口转发进行本地测试"
    kubectl port-forward svc/my-first-deployment-service 8080:80 &
    PORT_FORWARD_PID=$!
    
    sleep 5
    
    echo "测试本地连接 http://localhost:8080"
    if command -v curl &> /dev/null; then
        curl -s http://localhost:8080 | head -10 || echo "连接测试失败"
    else
        echo "curl 命令未找到，跳过连接测试"
    fi
    
    # 停止端口转发
    kill $PORT_FORWARD_PID 2>/dev/null
    
    wait_for_user
    
    log_info "验证标签选择器匹配"
    echo "=== Service 选择器 ==="
    kubectl get svc my-first-deployment-service -o yaml | grep -A 5 selector
    
    echo -e "\n=== Pod 标签 ==="
    kubectl get pods -l app=my-first-deployment --show-labels
    
    log_success "Service 暴露和测试完成"
}

# 监控和状态检查
monitor_status() {
    log_step "步骤 4: 监控和状态检查"
    
    log_info "查看所有相关资源"
    kubectl get all -l app=my-first-deployment
    
    wait_for_user
    
    log_info "查看 Deployment 历史"
    kubectl rollout history deployment/my-first-deployment
    
    wait_for_user
    
    log_info "查看相关事件"
    kubectl get events --sort-by=.metadata.creationTimestamp | grep my-first-deployment | tail -10
    
    wait_for_user
    
    # 检查是否有 metrics-server
    if kubectl top nodes &> /dev/null; then
        log_info "查看资源使用情况"
        kubectl top nodes
        kubectl top pods -l app=my-first-deployment
    else
        log_warning "metrics-server 未安装，跳过资源使用情况检查"
    fi
    
    log_success "监控和状态检查完成"
}

# 清理资源
cleanup_resources() {
    log_step "步骤 5: 清理资源"
    
    log_info "清理前检查当前资源状态"
    echo "=== 当前 Deployment 状态 ==="
    kubectl get deployment my-first-deployment
    
    echo -e "\n=== 当前 Service 状态 ==="
    kubectl get svc my-first-deployment-service
    
    echo -e "\n=== 当前 Pod 状态 ==="
    kubectl get pods -l app=my-first-deployment
    
    wait_for_user
    
    log_info "删除 Service"
    kubectl delete svc my-first-deployment-service
    
    log_info "缩减 Deployment 到 0"
    kubectl scale --replicas=0 deployment/my-first-deployment
    
    log_info "等待 Pod 终止..."
    sleep 10
    kubectl get pods -l app=my-first-deployment
    
    wait_for_user
    
    log_info "删除 Deployment"
    kubectl delete deployment my-first-deployment
    
    log_info "验证清理结果"
    echo "=== 清理验证 ==="
    kubectl get deployment my-first-deployment 2>/dev/null || echo "Deployment 已删除"
    kubectl get pods -l app=my-first-deployment 2>/dev/null || echo "Pod 已删除"
    kubectl get svc my-first-deployment-service 2>/dev/null || echo "Service 已删除"
    
    log_success "资源清理完成"
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}Kubernetes Deployment 演示脚本${NC}"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -c, --cleanup  仅执行清理操作"
    echo "  -s, --skip     跳过清理现有资源"
    echo ""
    echo "本脚本将演示以下操作:"
    echo "  1. 创建 Deployment"
    echo "  2. 扩展 Deployment"
    echo "  3. 将 Deployment 暴露为 Service"
    echo "  4. 监控和状态检查"
    echo "  5. 清理资源"
    echo ""
    echo "注意: 脚本需要有效的 kubectl 配置和 Kubernetes 集群连接"
}

# 主函数
main() {
    echo -e "${CYAN}"
    echo "================================================"
    echo "    Kubernetes Deployment 演示脚本"
    echo "================================================"
    echo -e "${NC}"
    
    # 解析命令行参数
    CLEANUP_ONLY=false
    SKIP_CLEANUP=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--cleanup)
                CLEANUP_ONLY=true
                shift
                ;;
            -s|--skip)
                SKIP_CLEANUP=true
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查环境
    check_kubectl
    
    if [ "$CLEANUP_ONLY" = true ]; then
        cleanup_existing
        log_success "清理完成"
        exit 0
    fi
    
    # 执行演示步骤
    if [ "$SKIP_CLEANUP" = false ]; then
        cleanup_existing
        wait_for_user
    fi
    
    create_deployment
    wait_for_user
    
    scale_deployment
    wait_for_user
    
    expose_service
    wait_for_user
    
    monitor_status
    wait_for_user
    
    echo -e "\n${YELLOW}是否要清理创建的资源？(y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        cleanup_resources
    else
        log_info "保留资源，可以稍后手动清理"
        echo "清理命令: kubectl delete all -l app=my-first-deployment"
    fi
    
    echo -e "\n${GREEN}"
    echo "================================================"
    echo "    Deployment 演示完成！"
    echo "================================================"
    echo -e "${NC}"
    
    echo "学习要点总结:"
    echo "✅ Deployment 创建和管理"
    echo "✅ 扩缩容操作和监控"
    echo "✅ Service 暴露和访问测试"
    echo "✅ 资源监控和状态检查"
    echo "✅ 资源清理和最佳实践"
    
    echo -e "\n下一步学习建议:"
    echo "- Deployment 滚动更新"
    echo "- Deployment 回滚操作"
    echo "- ConfigMap 和 Secret 管理"
    echo "- Ingress 和高级网络配置"
}

# 错误处理
trap 'log_error "脚本执行中断"; exit 1' INT TERM

# 执行主函数
main "$@"