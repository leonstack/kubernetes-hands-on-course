#!/bin/bash

# 描述：演示 Kubernetes Deployment 的各种回滚操作
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

# 配置变量
DEPLOYMENT_NAME="my-first-deployment"
SERVICE_NAME="my-first-deployment-service"
NAMESPACE="default"
WAIT_TIME=30

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
    
    log_success "kubectl 连接正常"
}

# 检查前置条件
check_prerequisites() {
    log_step "检查前置条件"
    
    # 检查 Deployment 是否存在
    if ! kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE &> /dev/null; then
        log_error "Deployment '$DEPLOYMENT_NAME' 不存在"
        log_info "请先运行 04-01 或 04-02 教程创建 Deployment"
        exit 1
    fi
    
    # 检查是否有版本历史
    local history_count=$(kubectl rollout history deployment/$DEPLOYMENT_NAME -n $NAMESPACE --no-headers | wc -l)
    if [ $history_count -lt 2 ]; then
        log_warning "版本历史不足（当前：$history_count 个版本）"
        log_info "建议先执行一些更新操作以创建版本历史"
    else
        log_success "发现 $history_count 个历史版本"
    fi
}

# 显示当前状态
show_current_status() {
    log_step "显示当前状态"
    
    echo -e "${CYAN}=== Deployment 状态 ===${NC}"
    kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o wide
    
    echo -e "\n${CYAN}=== Pod 状态 ===${NC}"
    kubectl get pods -l app=$DEPLOYMENT_NAME -n $NAMESPACE -o wide
    
    echo -e "\n${CYAN}=== 版本历史 ===${NC}"
    kubectl rollout history deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    
    echo -e "\n${CYAN}=== 当前镜像版本 ===${NC}"
    local current_image=$(kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "当前镜像：$current_image"
}

# 回滚到上一个版本
rollback_to_previous() {
    log_step "回滚到上一个版本"
    
    # 获取当前版本号
    local current_revision=$(kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
    log_info "当前版本：$current_revision"
    
    # 执行回滚
    log_info "执行回滚操作..."
    kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    
    # 监控回滚过程
    log_info "监控回滚过程（最多等待 ${WAIT_TIME} 秒）..."
    if kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=${WAIT_TIME}s; then
        log_success "回滚完成"
    else
        log_error "回滚超时，请检查状态"
        return 1
    fi
    
    # 验证回滚结果
    verify_rollback_result
}

# 回滚到指定版本
rollback_to_revision() {
    log_step "回滚到指定版本"
    
    # 显示可用版本
    echo -e "${CYAN}=== 可用版本历史 ===${NC}"
    kubectl rollout history deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    
    # 获取用户输入
    echo -n "请输入要回滚到的版本号："
    read -r target_revision
    
    # 验证版本号
    if ! kubectl rollout history deployment/$DEPLOYMENT_NAME -n $NAMESPACE --revision=$target_revision &> /dev/null; then
        log_error "版本 $target_revision 不存在"
        return 1
    fi
    
    # 显示目标版本详细信息
    log_info "目标版本 $target_revision 的详细信息："
    kubectl rollout history deployment/$DEPLOYMENT_NAME -n $NAMESPACE --revision=$target_revision
    
    # 确认回滚
    echo -n "确认回滚到版本 $target_revision？(y/N): "
    read -r confirm
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        log_info "取消回滚操作"
        return 0
    fi
    
    # 执行回滚
    log_info "回滚到版本 $target_revision..."
    kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE --to-revision=$target_revision
    
    # 监控回滚过程
    log_info "监控回滚过程（最多等待 ${WAIT_TIME} 秒）..."
    if kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=${WAIT_TIME}s; then
        log_success "回滚到版本 $target_revision 完成"
    else
        log_error "回滚超时，请检查状态"
        return 1
    fi
    
    # 验证回滚结果
    verify_rollback_result
}

# 滚动重启
rolling_restart() {
    log_step "执行滚动重启"
    
    log_info "滚动重启不会改变应用配置，只是重新创建所有 Pod"
    echo -n "确认执行滚动重启？(y/N): "
    read -r confirm
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        log_info "取消滚动重启操作"
        return 0
    fi
    
    # 记录重启前的 Pod 信息
    log_info "重启前的 Pod 状态："
    kubectl get pods -l app=$DEPLOYMENT_NAME -n $NAMESPACE -o custom-columns=NAME:.metadata.name,AGE:.metadata.creationTimestamp
    
    # 执行滚动重启
    log_info "执行滚动重启..."
    kubectl rollout restart deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    
    # 监控重启过程
    log_info "监控重启过程（最多等待 ${WAIT_TIME} 秒）..."
    if kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=${WAIT_TIME}s; then
        log_success "滚动重启完成"
    else
        log_error "滚动重启超时，请检查状态"
        return 1
    fi
    
    # 显示重启后的状态
    log_info "重启后的 Pod 状态："
    kubectl get pods -l app=$DEPLOYMENT_NAME -n $NAMESPACE -o custom-columns=NAME:.metadata.name,AGE:.metadata.creationTimestamp
}

# 验证回滚结果
verify_rollback_result() {
    log_step "验证回滚结果"
    
    # 检查 Deployment 状态
    local ready_replicas=$(kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
    local desired_replicas=$(kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.spec.replicas}')
    
    if [ "$ready_replicas" = "$desired_replicas" ]; then
        log_success "所有副本都已就绪 ($ready_replicas/$desired_replicas)"
    else
        log_warning "副本状态：$ready_replicas/$desired_replicas"
    fi
    
    # 检查 Pod 状态
    local running_pods=$(kubectl get pods -l app=$DEPLOYMENT_NAME -n $NAMESPACE --field-selector=status.phase=Running --no-headers | wc -l)
    log_info "运行中的 Pod 数量：$running_pods"
    
    # 显示当前镜像版本
    local current_image=$(kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.spec.template.spec.containers[0].image}')
    log_info "当前镜像版本：$current_image"
}

# 测试应用程序
test_application() {
    log_step "测试应用程序"
    
    # 检查 Service 是否存在
    if ! kubectl get service $SERVICE_NAME -n $NAMESPACE &> /dev/null; then
        log_warning "Service '$SERVICE_NAME' 不存在，跳过应用程序测试"
        return 0
    fi
    
    # 获取 Service 信息
    local service_type=$(kubectl get service $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.type}')
    local service_port=$(kubectl get service $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].port}')
    
    if [ "$service_type" = "NodePort" ]; then
        local node_port=$(kubectl get service $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')
        local node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
        
        log_info "测试 NodePort 服务访问："
        echo "访问地址：http://$node_ip:$node_port"
        
        if command -v curl &> /dev/null; then
            log_info "使用 curl 测试连接..."
            if curl -s --connect-timeout 5 "http://$node_ip:$node_port" > /dev/null; then
                log_success "应用程序响应正常"
            else
                log_warning "应用程序可能未完全就绪，请稍后重试"
            fi
        fi
    else
        log_info "使用端口转发测试应用程序..."
        echo "启动端口转发（Ctrl+C 停止）："
        echo "kubectl port-forward service/$SERVICE_NAME -n $NAMESPACE 8080:$service_port"
        echo "然后访问：http://localhost:8080"
    fi
}

# 显示状态概览
show_status_overview() {
    log_step "状态概览"
    
    echo -e "${CYAN}=== 完整状态概览 ===${NC}"
    
    # Deployment 状态
    echo -e "\n${YELLOW}Deployment 状态：${NC}"
    kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o wide
    
    # ReplicaSet 状态
    echo -e "\n${YELLOW}ReplicaSet 状态：${NC}"
    kubectl get rs -l app=$DEPLOYMENT_NAME -n $NAMESPACE
    
    # Pod 状态
    echo -e "\n${YELLOW}Pod 状态：${NC}"
    kubectl get pods -l app=$DEPLOYMENT_NAME -n $NAMESPACE -o wide
    
    # 版本历史
    echo -e "\n${YELLOW}版本历史：${NC}"
    kubectl rollout history deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    
    # 最近事件
    echo -e "\n${YELLOW}最近事件：${NC}"
    kubectl get events --field-selector involvedObject.name=$DEPLOYMENT_NAME -n $NAMESPACE --sort-by=.metadata.creationTimestamp | tail -5
}

# 清理资源
cleanup_resources() {
    log_step "清理资源"
    
    echo "选择清理方式："
    echo "1) 完全清理（删除 Deployment 和 Service）"
    echo "2) 重置到第一个版本"
    echo "3) 取消"
    echo -n "请选择 (1-3): "
    read -r choice
    
    case $choice in
        1)
            log_warning "即将删除所有资源"
            echo -n "确认删除？(y/N): "
            read -r confirm
            if [[ $confirm =~ ^[Yy]$ ]]; then
                kubectl delete deployment $DEPLOYMENT_NAME -n $NAMESPACE
                kubectl delete service $SERVICE_NAME -n $NAMESPACE 2>/dev/null || true
                log_success "资源清理完成"
            fi
            ;;
        2)
            log_info "重置到第一个版本"
            kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE --to-revision=1
            kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE
            log_success "已重置到第一个版本"
            ;;
        3)
            log_info "取消清理操作"
            ;;
        *)
            log_error "无效选择"
            ;;
    esac
}

# 显示帮助信息
show_help() {
    echo "Kubernetes Deployment 回滚演示脚本"
    echo ""
    echo "用法：$0 [选项]"
    echo ""
    echo "选项："
    echo "  -h, --help              显示帮助信息"
    echo "  -s, --status            显示当前状态"
    echo "  -p, --previous          回滚到上一个版本"
    echo "  -r, --revision          回滚到指定版本"
    echo "  -R, --restart           执行滚动重启"
    echo "  -t, --test              测试应用程序"
    echo "  -o, --overview          显示状态概览"
    echo "  -c, --cleanup           清理资源"
    echo "  -i, --interactive       交互式模式（默认）"
    echo ""
    echo "示例："
    echo "  $0                      # 交互式模式"
    echo "  $0 -s                   # 显示当前状态"
    echo "  $0 -p                   # 回滚到上一个版本"
    echo "  $0 -r                   # 回滚到指定版本"
    echo "  $0 -R                   # 执行滚动重启"
}

# 交互式菜单
interactive_menu() {
    while true; do
        echo ""
        echo -e "${CYAN}=== Kubernetes Deployment 回滚演示 ===${NC}"
        echo "1) 显示当前状态"
        echo "2) 回滚到上一个版本"
        echo "3) 回滚到指定版本"
        echo "4) 滚动重启"
        echo "5) 测试应用程序"
        echo "6) 状态概览"
        echo "7) 清理资源"
        echo "8) 退出"
        echo -n "请选择操作 (1-8): "
        read -r choice
        
        case $choice in
            1) show_current_status ;;
            2) rollback_to_previous ;;
            3) rollback_to_revision ;;
            4) rolling_restart ;;
            5) test_application ;;
            6) show_status_overview ;;
            7) cleanup_resources ;;
            8) 
                log_info "退出演示脚本"
                exit 0
                ;;
            *) 
                log_error "无效选择，请输入 1-8"
                ;;
        esac
        
        echo ""
        echo -n "按 Enter 键继续..."
        read -r
    done
}

# 主函数
main() {
    # 检查参数
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--status)
            check_kubectl
            check_prerequisites
            show_current_status
            exit 0
            ;;
        -p|--previous)
            check_kubectl
            check_prerequisites
            rollback_to_previous
            exit 0
            ;;
        -r|--revision)
            check_kubectl
            check_prerequisites
            rollback_to_revision
            exit 0
            ;;
        -R|--restart)
            check_kubectl
            check_prerequisites
            rolling_restart
            exit 0
            ;;
        -t|--test)
            check_kubectl
            check_prerequisites
            test_application
            exit 0
            ;;
        -o|--overview)
            check_kubectl
            check_prerequisites
            show_status_overview
            exit 0
            ;;
        -c|--cleanup)
            check_kubectl
            check_prerequisites
            cleanup_resources
            exit 0
            ;;
        -i|--interactive|"")
            # 交互式模式（默认）
            check_kubectl
            check_prerequisites
            interactive_menu
            ;;
        *)
            log_error "未知选项：$1"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi