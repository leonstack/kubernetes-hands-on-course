#!/bin/bash

# 演示 Kubernetes Services 的创建、管理和使用
# 作者: Grissom
# 版本: 1.0.0
# 日期: 2025-06-20

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置变量
NAMESPACE="default"
MANIFEST_DIR="./kube-manifests"
BACKEND_DEPLOYMENT="backend-restapp"
BACKEND_SERVICE="backend-restapp-clusterip-service"
FRONTEND_DEPLOYMENT="frontend-nginxapp"
FRONTEND_SERVICE="frontend-nginxapp-nodeport-service"
NODEPORT="31234"

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

log_header() {
    echo -e "\n${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}\n"
}

# 检查前置条件
check_prerequisites() {
    log_header "检查前置条件"
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
    
    # 检查 Kubernetes 集群连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群，请检查配置"
        exit 1
    fi
    
    # 检查 YAML 文件
    if [ ! -d "$MANIFEST_DIR" ]; then
        log_error "找不到 manifest 目录: $MANIFEST_DIR"
        exit 1
    fi
    
    local required_files=(
        "01-backend-deployment.yml"
        "02-backend-clusterip-service.yml"
        "03-frontend-deployment.yml"
        "04-frontend-nodeport-service.yml"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$MANIFEST_DIR/$file" ]; then
            log_error "找不到必需的文件: $MANIFEST_DIR/$file"
            exit 1
        fi
    done
    
    log_success "前置条件检查通过"
    
    # 显示集群信息
    echo -e "\n${CYAN}集群信息:${NC}"
    kubectl cluster-info
    echo -e "\n${CYAN}节点信息:${NC}"
    kubectl get nodes -o wide
}

# 验证 YAML 文件
validate_yaml_files() {
    log_header "验证 YAML 文件"
    
    local files=("$MANIFEST_DIR"/*.yml)
    
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            log_info "验证文件: $(basename "$file")"
            if kubectl apply --dry-run=client -f "$file" &> /dev/null; then
                log_success "✓ $(basename "$file") 语法正确"
            else
                log_error "✗ $(basename "$file") 语法错误"
                kubectl apply --dry-run=client -f "$file"
                return 1
            fi
        fi
    done
    
    log_success "所有 YAML 文件验证通过"
}

# 创建后端资源
create_backend_resources() {
    log_header "创建后端资源"
    
    # 创建后端 Deployment
    log_info "创建后端 Deployment..."
    kubectl apply -f "$MANIFEST_DIR/01-backend-deployment.yml"
    
    # 等待 Deployment 就绪
    log_info "等待后端 Deployment 就绪..."
    kubectl rollout status deployment/$BACKEND_DEPLOYMENT --timeout=300s
    
    # 创建后端 ClusterIP Service
    log_info "创建后端 ClusterIP Service..."
    kubectl apply -f "$MANIFEST_DIR/02-backend-clusterip-service.yml"
    
    # 验证后端资源
    log_info "验证后端资源状态..."
    kubectl get deployment $BACKEND_DEPLOYMENT
    kubectl get service $BACKEND_SERVICE
    kubectl get pods -l app=$BACKEND_DEPLOYMENT
    
    log_success "后端资源创建完成"
}

# 创建前端资源
create_frontend_resources() {
    log_header "创建前端资源"
    
    # 创建前端 Deployment
    log_info "创建前端 Deployment..."
    kubectl apply -f "$MANIFEST_DIR/03-frontend-deployment.yml"
    
    # 等待 Deployment 就绪
    log_info "等待前端 Deployment 就绪..."
    kubectl rollout status deployment/$FRONTEND_DEPLOYMENT --timeout=300s
    
    # 创建前端 NodePort Service
    log_info "创建前端 NodePort Service..."
    kubectl apply -f "$MANIFEST_DIR/04-frontend-nodeport-service.yml"
    
    # 验证前端资源
    log_info "验证前端资源状态..."
    kubectl get deployment $FRONTEND_DEPLOYMENT
    kubectl get service $FRONTEND_SERVICE
    kubectl get pods -l app=$FRONTEND_DEPLOYMENT
    
    log_success "前端资源创建完成"
}

# 批量创建所有资源
create_all_resources() {
    log_header "批量创建所有资源"
    
    log_info "使用 kubectl apply 批量创建资源..."
    kubectl apply -f "$MANIFEST_DIR/"
    
    # 等待所有 Deployment 就绪
    log_info "等待所有 Deployment 就绪..."
    kubectl rollout status deployment/$BACKEND_DEPLOYMENT --timeout=300s
    kubectl rollout status deployment/$FRONTEND_DEPLOYMENT --timeout=300s
    
    log_success "所有资源创建完成"
}

# 测试服务连接
test_services() {
    log_header "测试服务连接"
    
    # 获取节点 IP
    local node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    if [ -z "$node_ip" ]; then
        node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    fi
    
    if [ -z "$node_ip" ]; then
        log_warning "无法获取节点 IP，使用 localhost 进行端口转发测试"
        test_with_port_forward
        return
    fi
    
    log_info "节点 IP: $node_ip"
    log_info "NodePort: $NODEPORT"
    
    # 测试前端服务
    log_info "测试前端 NodePort 服务..."
    local frontend_url="http://$node_ip:$NODEPORT"
    echo -e "${CYAN}前端访问地址: $frontend_url${NC}"
    
    if command -v curl &> /dev/null; then
        log_info "使用 curl 测试前端服务..."
        if curl -s --connect-timeout 10 "$frontend_url" > /dev/null; then
            log_success "✓ 前端服务响应正常"
        else
            log_warning "前端服务可能还未完全就绪，请稍后手动测试"
        fi
        
        # 测试后端 API
        log_info "测试后端 API..."
        if curl -s --connect-timeout 10 "$frontend_url/hello" > /dev/null; then
            log_success "✓ 后端 API 响应正常"
        else
            log_warning "后端 API 可能还未完全就绪，请稍后手动测试"
        fi
    else
        log_warning "curl 未安装，跳过自动测试"
    fi
    
    # 显示访问信息
    echo -e "\n${GREEN}服务访问信息:${NC}"
    echo -e "${CYAN}前端应用: $frontend_url${NC}"
    echo -e "${CYAN}后端 API: $frontend_url/hello${NC}"
}

# 端口转发测试
test_with_port_forward() {
    log_header "端口转发测试"
    
    log_info "启动前端服务端口转发..."
    kubectl port-forward svc/$FRONTEND_SERVICE 8080:80 &
    local pf_pid=$!
    
    sleep 5
    
    if command -v curl &> /dev/null; then
        log_info "测试端口转发连接..."
        if curl -s --connect-timeout 5 "http://localhost:8080" > /dev/null; then
            log_success "✓ 端口转发测试成功"
            echo -e "${CYAN}本地访问地址: http://localhost:8080${NC}"
        else
            log_warning "端口转发测试失败"
        fi
    fi
    
    # 停止端口转发
    kill $pf_pid 2>/dev/null || true
    log_info "端口转发已停止"
}

# 服务发现演示
demonstrate_service_discovery() {
    log_header "服务发现演示"
    
    # 创建测试 Pod
    log_info "创建测试 Pod 进行服务发现演示..."
    kubectl run test-pod --image=busybox --rm -i --tty --restart=Never -- /bin/sh -c "
        echo '=== DNS 服务发现测试 ==='
        echo '1. 测试后端服务 DNS 解析:'
        nslookup $BACKEND_SERVICE
        echo ''
        echo '2. 测试前端服务 DNS 解析:'
        nslookup $FRONTEND_SERVICE
        echo ''
        echo '3. 测试后端服务连接:'
        wget -qO- --timeout=10 http://$BACKEND_SERVICE:8080/hello || echo 'Connection failed'
        echo ''
        echo '=== 环境变量服务发现 ==='
        env | grep -E '(BACKEND|FRONTEND).*SERVICE' | sort
    " 2>/dev/null || log_warning "服务发现演示失败，可能是网络策略限制"
}

# 负载均衡测试
test_load_balancing() {
    log_header "负载均衡测试"
    
    log_info "测试后端服务负载均衡..."
    
    # 获取后端 Pod 列表
    local backend_pods=$(kubectl get pods -l app=$BACKEND_DEPLOYMENT -o jsonpath='{.items[*].metadata.name}')
    log_info "后端 Pod 列表: $backend_pods"
    
    # 创建测试脚本
    kubectl run load-test --image=busybox --rm -i --tty --restart=Never -- /bin/sh -c "
        echo '=== 负载均衡测试 ==='
        echo '发送 10 个请求到后端服务...'
        for i in \$(seq 1 10); do
            echo \"Request \$i:\"
            wget -qO- --timeout=5 http://$BACKEND_SERVICE:8080/hello 2>/dev/null || echo 'Request failed'
            sleep 1
        done
    " 2>/dev/null || log_warning "负载均衡测试失败"
}

# 监控资源状态
monitor_resources() {
    log_header "监控资源状态"
    
    echo -e "${CYAN}=== Deployments ===${NC}"
    kubectl get deployments -l 'tier in (backend,frontend)' -o wide
    
    echo -e "\n${CYAN}=== Services ===${NC}"
    kubectl get services -l 'tier in (backend,frontend)' -o wide
    
    echo -e "\n${CYAN}=== Pods ===${NC}"
    kubectl get pods -l 'tier in (backend,frontend)' -o wide
    
    echo -e "\n${CYAN}=== Endpoints ===${NC}"
    kubectl get endpoints $BACKEND_SERVICE $FRONTEND_SERVICE
    
    echo -e "\n${CYAN}=== Events ===${NC}"
    kubectl get events --sort-by=.metadata.creationTimestamp | tail -10
}

# 查看详细信息
show_detailed_info() {
    log_header "查看详细信息"
    
    echo -e "${CYAN}=== 后端 Deployment 详情 ===${NC}"
    kubectl describe deployment $BACKEND_DEPLOYMENT
    
    echo -e "\n${CYAN}=== 后端 Service 详情 ===${NC}"
    kubectl describe service $BACKEND_SERVICE
    
    echo -e "\n${CYAN}=== 前端 Deployment 详情 ===${NC}"
    kubectl describe deployment $FRONTEND_DEPLOYMENT
    
    echo -e "\n${CYAN}=== 前端 Service 详情 ===${NC}"
    kubectl describe service $FRONTEND_SERVICE
}

# 查看日志
view_logs() {
    log_header "查看应用日志"
    
    echo -e "${CYAN}=== 后端应用日志 ===${NC}"
    kubectl logs -l app=$BACKEND_DEPLOYMENT --tail=20
    
    echo -e "\n${CYAN}=== 前端应用日志 ===${NC}"
    kubectl logs -l app=$FRONTEND_DEPLOYMENT --tail=20
}

# 扩缩容演示
demonstrate_scaling() {
    log_header "扩缩容演示"
    
    # 扩容后端
    log_info "扩容后端服务到 5 个副本..."
    kubectl scale deployment $BACKEND_DEPLOYMENT --replicas=5
    kubectl rollout status deployment/$BACKEND_DEPLOYMENT --timeout=120s
    
    log_info "当前后端 Pod 状态:"
    kubectl get pods -l app=$BACKEND_DEPLOYMENT
    
    sleep 5
    
    # 缩容后端
    log_info "缩容后端服务到 2 个副本..."
    kubectl scale deployment $BACKEND_DEPLOYMENT --replicas=2
    kubectl rollout status deployment/$BACKEND_DEPLOYMENT --timeout=120s
    
    log_info "当前后端 Pod 状态:"
    kubectl get pods -l app=$BACKEND_DEPLOYMENT
    
    # 恢复原始副本数
    log_info "恢复后端服务到 3 个副本..."
    kubectl scale deployment $BACKEND_DEPLOYMENT --replicas=3
    kubectl rollout status deployment/$BACKEND_DEPLOYMENT --timeout=120s
    
    log_success "扩缩容演示完成"
}

# 故障排除演示
demonstrate_troubleshooting() {
    log_header "故障排除演示"
    
    echo -e "${CYAN}=== 常用故障排除命令 ===${NC}"
    
    echo -e "\n1. 检查 Pod 状态和事件:"
    kubectl get pods -l 'tier in (backend,frontend)' -o wide
    
    echo -e "\n2. 检查 Service 端点:"
    kubectl get endpoints
    
    echo -e "\n3. 检查网络策略:"
    kubectl get networkpolicies 2>/dev/null || echo "No network policies found"
    
    echo -e "\n4. 检查资源使用情况:"
    kubectl top pods -l 'tier in (backend,frontend)' 2>/dev/null || echo "Metrics server not available"
    
    echo -e "\n5. 检查 DNS 解析:"
    kubectl run dns-test --image=busybox --rm -i --tty --restart=Never -- nslookup kubernetes.default 2>/dev/null || echo "DNS test failed"
}

# 性能测试
performance_test() {
    log_header "性能测试"
    
    if ! command -v ab &> /dev/null; then
        log_warning "Apache Bench (ab) 未安装，跳过性能测试"
        return
    fi
    
    # 获取节点 IP 和端口
    local node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    if [ -z "$node_ip" ]; then
        log_warning "无法获取节点 IP，跳过性能测试"
        return
    fi
    
    local test_url="http://$node_ip:$NODEPORT/hello"
    
    log_info "对后端 API 进行性能测试..."
    log_info "测试 URL: $test_url"
    
    # 简单的性能测试
    ab -n 100 -c 10 "$test_url" || log_warning "性能测试失败"
}

# 清理资源
cleanup_resources() {
    log_header "清理资源"
    
    log_info "删除所有相关资源..."
    
    # 删除 Services
    kubectl delete service $BACKEND_SERVICE $FRONTEND_SERVICE --ignore-not-found=true
    
    # 删除 Deployments
    kubectl delete deployment $BACKEND_DEPLOYMENT $FRONTEND_DEPLOYMENT --ignore-not-found=true
    
    # 等待资源删除
    log_info "等待资源删除完成..."
    kubectl wait --for=delete deployment/$BACKEND_DEPLOYMENT --timeout=60s 2>/dev/null || true
    kubectl wait --for=delete deployment/$FRONTEND_DEPLOYMENT --timeout=60s 2>/dev/null || true
    
    # 验证清理结果
    log_info "验证清理结果..."
    local remaining_resources=$(kubectl get deployments,services -l 'tier in (backend,frontend)' --no-headers 2>/dev/null | wc -l)
    
    if [ "$remaining_resources" -eq 0 ]; then
        log_success "所有资源已成功清理"
    else
        log_warning "仍有资源未清理完成"
        kubectl get deployments,services -l 'tier in (backend,frontend)'
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${PURPLE}Kubernetes Services with YAML Demo Script${NC}"
    echo -e "${PURPLE}==========================================${NC}\n"
    
    echo -e "${CYAN}用法:${NC}"
    echo -e "  $0 [选项]\n"
    
    echo -e "${CYAN}选项:${NC}"
    echo -e "  ${GREEN}--check${NC}              检查前置条件"
    echo -e "  ${GREEN}--validate${NC}           验证 YAML 文件"
    echo -e "  ${GREEN}--create-backend${NC}     创建后端资源"
    echo -e "  ${GREEN}--create-frontend${NC}    创建前端资源"
    echo -e "  ${GREEN}--create-all${NC}         批量创建所有资源"
    echo -e "  ${GREEN}--test${NC}               测试服务连接"
    echo -e "  ${GREEN}--port-forward${NC}       端口转发测试"
    echo -e "  ${GREEN}--service-discovery${NC}  服务发现演示"
    echo -e "  ${GREEN}--load-balancing${NC}     负载均衡测试"
    echo -e "  ${GREEN}--monitor${NC}            监控资源状态"
    echo -e "  ${GREEN}--details${NC}            查看详细信息"
    echo -e "  ${GREEN}--logs${NC}               查看应用日志"
    echo -e "  ${GREEN}--scaling${NC}            扩缩容演示"
    echo -e "  ${GREEN}--troubleshoot${NC}       故障排除演示"
    echo -e "  ${GREEN}--performance${NC}        性能测试"
    echo -e "  ${GREEN}--cleanup${NC}            清理所有资源"
    echo -e "  ${GREEN}--demo${NC}               完整演示流程"
    echo -e "  ${GREEN}--interactive${NC}        交互式菜单"
    echo -e "  ${GREEN}--help${NC}               显示帮助信息\n"
    
    echo -e "${CYAN}示例:${NC}"
    echo -e "  $0 --demo                # 运行完整演示"
    echo -e "  $0 --create-all --test   # 创建资源并测试"
    echo -e "  $0 --interactive         # 交互式菜单"
    echo -e "  $0 --cleanup             # 清理所有资源\n"
}

# 交互式菜单
interactive_menu() {
    while true; do
        echo -e "\n${PURPLE}========================================${NC}"
        echo -e "${PURPLE}Kubernetes Services Demo - 交互式菜单${NC}"
        echo -e "${PURPLE}========================================${NC}\n"
        
        echo -e "${CYAN}请选择操作:${NC}"
        echo -e "  ${GREEN}1)${NC}  检查前置条件"
        echo -e "  ${GREEN}2)${NC}  验证 YAML 文件"
        echo -e "  ${GREEN}3)${NC}  创建后端资源"
        echo -e "  ${GREEN}4)${NC}  创建前端资源"
        echo -e "  ${GREEN}5)${NC}  批量创建所有资源"
        echo -e "  ${GREEN}6)${NC}  测试服务连接"
        echo -e "  ${GREEN}7)${NC}  服务发现演示"
        echo -e "  ${GREEN}8)${NC}  负载均衡测试"
        echo -e "  ${GREEN}9)${NC}  监控资源状态"
        echo -e "  ${GREEN}10)${NC} 查看详细信息"
        echo -e "  ${GREEN}11)${NC} 查看应用日志"
        echo -e "  ${GREEN}12)${NC} 扩缩容演示"
        echo -e "  ${GREEN}13)${NC} 故障排除演示"
        echo -e "  ${GREEN}14)${NC} 性能测试"
        echo -e "  ${GREEN}15)${NC} 清理所有资源"
        echo -e "  ${GREEN}16)${NC} 完整演示流程"
        echo -e "  ${GREEN}0)${NC}  退出\n"
        
        read -p "请输入选项 (0-16): " choice
        
        case $choice in
            1) check_prerequisites ;;
            2) validate_yaml_files ;;
            3) create_backend_resources ;;
            4) create_frontend_resources ;;
            5) create_all_resources ;;
            6) test_services ;;
            7) demonstrate_service_discovery ;;
            8) test_load_balancing ;;
            9) monitor_resources ;;
            10) show_detailed_info ;;
            11) view_logs ;;
            12) demonstrate_scaling ;;
            13) demonstrate_troubleshooting ;;
            14) performance_test ;;
            15) cleanup_resources ;;
            16) run_full_demo ;;
            0) log_info "退出程序"; exit 0 ;;
            *) log_error "无效选项，请重新选择" ;;
        esac
        
        echo -e "\n${YELLOW}按 Enter 键继续...${NC}"
        read
    done
}

# 完整演示流程
run_full_demo() {
    log_header "Kubernetes Services 完整演示"
    
    echo -e "${CYAN}本演示将展示以下内容:${NC}"
    echo -e "1. 前置条件检查"
    echo -e "2. YAML 文件验证"
    echo -e "3. 创建后端和前端资源"
    echo -e "4. 测试服务连接"
    echo -e "5. 服务发现演示"
    echo -e "6. 负载均衡测试"
    echo -e "7. 扩缩容演示"
    echo -e "8. 监控和故障排除"
    echo -e "9. 资源清理\n"
    
    read -p "是否继续? (y/N): " confirm
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        log_info "演示已取消"
        return
    fi
    
    # 执行演示步骤
    check_prerequisites
    validate_yaml_files
    create_all_resources
    sleep 10  # 等待服务启动
    test_services
    demonstrate_service_discovery
    test_load_balancing
    monitor_resources
    demonstrate_scaling
    demonstrate_troubleshooting
    
    # 询问是否清理资源
    echo -e "\n${YELLOW}演示完成！${NC}"
    read -p "是否清理演示资源? (y/N): " cleanup_confirm
    if [[ $cleanup_confirm =~ ^[Yy]$ ]]; then
        cleanup_resources
    else
        log_info "资源保留，可以继续手动测试"
        echo -e "${CYAN}访问地址: http://<NODE-IP>:$NODEPORT${NC}"
    fi
    
    log_success "演示流程完成！"
}

# 主函数
main() {
    # 如果没有参数，显示帮助
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check)
                check_prerequisites
                shift
                ;;
            --validate)
                validate_yaml_files
                shift
                ;;
            --create-backend)
                create_backend_resources
                shift
                ;;
            --create-frontend)
                create_frontend_resources
                shift
                ;;
            --create-all)
                create_all_resources
                shift
                ;;
            --test)
                test_services
                shift
                ;;
            --port-forward)
                test_with_port_forward
                shift
                ;;
            --service-discovery)
                demonstrate_service_discovery
                shift
                ;;
            --load-balancing)
                test_load_balancing
                shift
                ;;
            --monitor)
                monitor_resources
                shift
                ;;
            --details)
                show_detailed_info
                shift
                ;;
            --logs)
                view_logs
                shift
                ;;
            --scaling)
                demonstrate_scaling
                shift
                ;;
            --troubleshoot)
                demonstrate_troubleshooting
                shift
                ;;
            --performance)
                performance_test
                shift
                ;;
            --cleanup)
                cleanup_resources
                shift
                ;;
            --demo)
                run_full_demo
                shift
                ;;
            --interactive)
                interactive_menu
                shift
                ;;
            --help)
                show_help
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi

# ========================================
# 脚本使用说明
# ========================================

# 1. 基本使用:
#    chmod +x services-yaml-demo.sh
#    ./services-yaml-demo.sh --help

# 2. 快速开始:
#    ./services-yaml-demo.sh --demo

# 3. 交互式使用:
#    ./services-yaml-demo.sh --interactive

# 4. 分步执行:
#    ./services-yaml-demo.sh --check
#    ./services-yaml-demo.sh --create-all
#    ./services-yaml-demo.sh --test
#    ./services-yaml-demo.sh --cleanup

# 5. 故障排除:
#    ./services-yaml-demo.sh --troubleshoot
#    ./services-yaml-demo.sh --logs

# ========================================
# 注意事项
# ========================================

# 1. 确保 kubectl 已正确配置
# 2. 确保有足够的集群资源
# 3. 某些功能需要特定的集群配置
# 4. 在生产环境中谨慎使用
# 5. 定期清理测试资源