#!/bin/bash

# 本脚本演示 ClusterIP 和 NodePort Service 的完整操作流程
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
    echo -e "\n${PURPLE}=== $1 ===${NC}\n"
}

# 检查 kubectl 是否可用
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在 PATH 中"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    log_success "kubectl 连接正常"
}

# 等待用户确认
wait_for_user() {
    echo -e "\n${CYAN}按 Enter 键继续，或按 Ctrl+C 退出...${NC}"
    read -r
}

# 等待资源就绪
wait_for_deployment() {
    local deployment_name=$1
    local timeout=${2:-300}
    
    log_info "等待 Deployment $deployment_name 就绪..."
    
    if kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment_name; then
        log_success "Deployment $deployment_name 已就绪"
    else
        log_error "Deployment $deployment_name 未能在 ${timeout} 秒内就绪"
        return 1
    fi
}

# 显示资源状态
show_status() {
    log_step "当前资源状态"
    
    echo -e "${YELLOW}Deployments:${NC}"
    kubectl get deployments -o wide
    
    echo -e "\n${YELLOW}Services:${NC}"
    kubectl get services -o wide
    
    echo -e "\n${YELLOW}Pods:${NC}"
    kubectl get pods -o wide
    
    echo -e "\n${YELLOW}Endpoints:${NC}"
    kubectl get endpoints
}

# 创建后端应用
create_backend() {
    log_step "第一步：创建后端应用 (ClusterIP Service)"
    
    log_info "创建后端 REST 应用的 Deployment..."
    kubectl create deployment my-backend-rest-app --image=grissomsh/kube-helloworld:1.0.0
    
    # 等待 Deployment 就绪
    wait_for_deployment "my-backend-rest-app"
    
    log_info "为后端应用创建 ClusterIP Service..."
    kubectl expose deployment my-backend-rest-app --port=8080 --target-port=8080 --name=my-backend-service
    
    log_success "后端应用和 ClusterIP Service 创建完成"
    
    # 显示创建的资源
    echo -e "\n${YELLOW}后端 Service 信息:${NC}"
    kubectl get svc my-backend-service
    kubectl describe svc my-backend-service
    
    wait_for_user
}

# 测试后端服务
test_backend() {
    log_step "测试后端服务 (集群内部访问)"
    
    log_info "从集群内部测试后端服务..."
    
    # 创建测试 Pod 并测试连接
    echo "测试命令: kubectl run test-backend --image=busybox --rm -it --restart=Never -- wget -qO- http://my-backend-service:8080/hello"
    
    if kubectl run test-backend --image=busybox --rm -it --restart=Never -- wget -qO- http://my-backend-service:8080/hello; then
        log_success "后端服务测试成功"
    else
        log_warning "后端服务测试失败，但这可能是正常的（取决于镜像内容）"
    fi
    
    wait_for_user
}

# 创建前端应用
create_frontend() {
    log_step "第二步：创建前端应用 (NodePort Service)"
    
    log_info "创建前端 Nginx 代理的 Deployment..."
    kubectl create deployment my-frontend-nginx-app --image=grissomsh/kube-frontend-nginx:1.0.0
    
    # 等待 Deployment 就绪
    wait_for_deployment "my-frontend-nginx-app"
    
    log_info "为前端应用创建 NodePort Service..."
    kubectl expose deployment my-frontend-nginx-app --type=NodePort --port=80 --target-port=80 --name=my-frontend-service
    
    log_success "前端应用和 NodePort Service 创建完成"
    
    # 显示创建的资源
    echo -e "\n${YELLOW}前端 Service 信息:${NC}"
    kubectl get svc my-frontend-service
    kubectl describe svc my-frontend-service
    
    wait_for_user
}

# 获取访问信息
get_access_info() {
    log_step "获取应用访问信息"
    
    # 获取 NodePort
    local nodeport=$(kubectl get svc my-frontend-service -o jsonpath='{.spec.ports[0].nodePort}')
    
    # 获取节点 IP
    local node_ips=$(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}')
    
    echo -e "${YELLOW}应用访问信息:${NC}"
    echo "NodePort: $nodeport"
    echo "节点 IP: $node_ips"
    echo ""
    echo -e "${GREEN}访问 URL:${NC}"
    for ip in $node_ips; do
        echo "  http://$ip:$nodeport/hello"
    done
    
    echo ""
    log_info "您可以在浏览器中访问上述任一 URL 来测试应用"
    
    wait_for_user
}

# 扩展后端应用
scale_backend() {
    log_step "第三步：扩展后端应用并测试负载均衡"
    
    log_info "将后端应用扩展到 5 个副本..."
    kubectl scale --replicas=5 deployment/my-backend-rest-app
    
    # 等待扩展完成
    log_info "等待扩展完成..."
    kubectl wait --for=condition=available --timeout=120s deployment/my-backend-rest-app
    
    # 显示扩展结果
    echo -e "\n${YELLOW}扩展后的 Pod 状态:${NC}"
    kubectl get pods -l app=my-backend-rest-app
    
    echo -e "\n${YELLOW}Service Endpoints:${NC}"
    kubectl get endpoints my-backend-service
    
    log_success "后端应用扩展完成"
    
    wait_for_user
}

# 测试负载均衡
test_load_balancing() {
    log_step "测试负载均衡功能"
    
    # 获取访问信息
    local nodeport=$(kubectl get svc my-frontend-service -o jsonpath='{.spec.ports[0].nodePort}')
    local node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    local url="http://$node_ip:$nodeport/hello"
    
    log_info "测试负载均衡 - 发送 10 个请求到: $url"
    
    echo -e "\n${YELLOW}负载均衡测试结果:${NC}"
    for i in {1..10}; do
        echo -n "请求 $i: "
        if command -v curl &> /dev/null; then
            curl -s --connect-timeout 5 "$url" || echo "请求失败"
        else
            echo "curl 未安装，跳过测试"
            break
        fi
        sleep 1
    done
    
    echo ""
    log_info "观察上述输出，如果看到不同的响应内容，说明负载均衡正在工作"
    
    wait_for_user
}

# 显示架构概览
show_architecture() {
    log_step "完整架构概览"
    
    echo -e "${CYAN}"
    cat << 'EOF'
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes 集群                          │
│                                                             │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐ │
│  │   NodePort  │───▶│   Frontend   │───▶│   ClusterIP     │ │
│  │   Service   │    │   (Nginx)    │    │   Service       │ │
│  │   :3xxxx    │    │   Port: 80   │    │   :8080         │ │
│  └─────────────┘    └──────────────┘    └─────────────────┘ │
│                                                ▼             │
│                                        ┌─────────────────┐   │
│                                        │   Backend       │   │
│                                        │   (Spring Boot) │   │
│                                        │   5 Replicas    │   │
│                                        │   Port: 8080    │   │
│                                        └─────────────────┘   │
└─────────────────────────────────────────────────────────────┘
EOF
    echo -e "${NC}"
    
    show_status
    
    wait_for_user
}

# 清理资源
cleanup() {
    log_step "清理演示资源"
    
    echo -e "${YELLOW}即将删除以下资源:${NC}"
    echo "- Service: my-frontend-service"
    echo "- Service: my-backend-service"
    echo "- Deployment: my-frontend-nginx-app"
    echo "- Deployment: my-backend-rest-app"
    
    echo -e "\n${RED}确认删除? (y/N):${NC}"
    read -r confirm
    
    if [[ $confirm =~ ^[Yy]$ ]]; then
        log_info "删除 Services..."
        kubectl delete svc my-frontend-service my-backend-service 2>/dev/null || true
        
        log_info "删除 Deployments..."
        kubectl delete deployment my-frontend-nginx-app my-backend-rest-app 2>/dev/null || true
        
        log_success "资源清理完成"
        
        # 验证清理结果
        echo -e "\n${YELLOW}清理后的资源状态:${NC}"
        kubectl get deployments,services,pods | grep -E "(my-frontend|my-backend)" || echo "所有演示资源已清理"
    else
        log_info "取消清理操作"
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}Kubernetes Services 演示脚本${NC}"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -c, --cleanup  仅执行清理操作"
    echo "  -s, --status   显示当前资源状态"
    echo ""
    echo "功能:"
    echo "  1. 创建后端应用和 ClusterIP Service"
    echo "  2. 创建前端应用和 NodePort Service"
    echo "  3. 演示负载均衡功能"
    echo "  4. 显示完整架构"
    echo "  5. 清理演示资源"
    echo ""
    echo "前置条件:"
    echo "  - Kubernetes 集群正在运行"
    echo "  - kubectl 已配置并可访问集群"
    echo "  - 集群有足够资源运行演示应用"
}

# 主函数
main() {
    # 解析命令行参数
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--cleanup)
            check_kubectl
            cleanup
            exit 0
            ;;
        -s|--status)
            check_kubectl
            show_status
            exit 0
            ;;
        "")
            # 正常执行流程
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
    
    # 显示脚本信息
    echo -e "${CYAN}=== Kubernetes Services 完整演示 ===${NC}\n"
    echo "本脚本将演示:"
    echo "1. ClusterIP Service - 集群内部服务发现"
    echo "2. NodePort Service - 外部访问"
    echo "3. 负载均衡功能验证"
    echo "4. 完整的前后端架构"
    
    wait_for_user
    
    # 检查环境
    log_step "环境检查"
    check_kubectl
    
    # 执行演示步骤
    create_backend
    test_backend
    create_frontend
    get_access_info
    scale_backend
    test_load_balancing
    show_architecture
    
    # 询问是否清理
    echo -e "\n${YELLOW}演示完成！${NC}"
    echo "是否要清理演示资源? (y/N):"
    read -r cleanup_confirm
    
    if [[ $cleanup_confirm =~ ^[Yy]$ ]]; then
        cleanup
    else
        log_info "保留演示资源，您可以稍后手动清理"
        echo "清理命令: $0 --cleanup"
    fi
    
    log_success "演示脚本执行完成！"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi