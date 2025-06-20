#!/bin/bash

# 描述: 演示如何暂停 Deployment、进行批量更改、然后恢复部署
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
NAMESPACE="default"
INITIAL_IMAGE="grissomsh/kubenginx:2.0.0"
UPDATE_IMAGE="grissomsh/kubenginx:4.0.0"
REPLICAS=3

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
    
    log_success "kubectl 可用，集群连接正常"
}

# 等待用户确认
wait_for_user() {
    echo -e "\n${CYAN}按 Enter 键继续，或按 Ctrl+C 退出...${NC}"
    read -r
}

# 显示当前状态
show_status() {
    local title="$1"
    log_step "$title"
    
    echo -e "${YELLOW}Deployment 状态:${NC}"
    kubectl get deployment $DEPLOYMENT_NAME -o wide 2>/dev/null || echo "Deployment 不存在"
    
    echo -e "\n${YELLOW}Pod 状态:${NC}"
    kubectl get pods -l app=$DEPLOYMENT_NAME 2>/dev/null || echo "没有找到相关 Pod"
    
    echo -e "\n${YELLOW}ReplicaSet 状态:${NC}"
    kubectl get rs -l app=$DEPLOYMENT_NAME 2>/dev/null || echo "没有找到相关 ReplicaSet"
}

# 检查暂停状态
check_pause_status() {
    local paused=$(kubectl get deployment $DEPLOYMENT_NAME -o jsonpath='{.spec.paused}' 2>/dev/null)
    if [ "$paused" = "true" ]; then
        echo -e "${GREEN}✅ Deployment 当前处于暂停状态${NC}"
    else
        echo -e "${BLUE}ℹ️  Deployment 当前处于运行状态${NC}"
    fi
}

# 创建初始 Deployment
create_deployment() {
    log_step "创建初始 Deployment"
    
    # 检查是否已存在
    if kubectl get deployment $DEPLOYMENT_NAME &> /dev/null; then
        log_warning "Deployment $DEPLOYMENT_NAME 已存在，跳过创建"
        return
    fi
    
    log_info "创建 Deployment: $DEPLOYMENT_NAME"
    kubectl create deployment $DEPLOYMENT_NAME --image=$INITIAL_IMAGE --replicas=$REPLICAS
    
    log_info "等待 Deployment 就绪..."
    kubectl rollout status deployment/$DEPLOYMENT_NAME --timeout=300s
    
    if [ $? -eq 0 ]; then
        log_success "Deployment 创建成功并已就绪"
    else
        log_error "Deployment 创建失败或超时"
        exit 1
    fi
}

# 暂停 Deployment
pause_deployment() {
    log_step "暂停 Deployment"
    
    log_info "执行暂停命令..."
    kubectl rollout pause deployment/$DEPLOYMENT_NAME
    
    if [ $? -eq 0 ]; then
        log_success "Deployment 已暂停"
        check_pause_status
    else
        log_error "暂停 Deployment 失败"
        exit 1
    fi
}

# 在暂停状态下进行多项更改
make_changes_while_paused() {
    log_step "在暂停状态下进行批量更改"
    
    # 更改 1: 更新镜像版本
    log_info "1. 更新镜像版本: $INITIAL_IMAGE -> $UPDATE_IMAGE"
    kubectl set image deployment/$DEPLOYMENT_NAME kubenginx=$UPDATE_IMAGE --record=true
    
    # 更改 2: 设置资源限制
    log_info "2. 设置资源限制和请求"
    kubectl set resources deployment/$DEPLOYMENT_NAME -c=kubenginx --limits=cpu=50m,memory=64Mi --requests=cpu=10m,memory=32Mi
    
    # 更改 3: 添加环境变量
    log_info "3. 添加环境变量"
    kubectl set env deployment/$DEPLOYMENT_NAME APP_ENV=production
    
    # 更改 4: 更新标签
    log_info "4. 更新标签"
    kubectl label deployment $DEPLOYMENT_NAME version=v4.0.0 --overwrite
    
    # 更改 5: 添加注解
    log_info "5. 添加注解"
    kubectl annotate deployment $DEPLOYMENT_NAME deployment.kubernetes.io/change-cause="Batch update to v4.0.0 with resource limits" --overwrite
    
    log_success "所有更改已应用到 Deployment 配置"
    
    # 验证暂停状态下的行为
    log_info "验证暂停状态下的行为..."
    echo -e "\n${YELLOW}当前 Deployment 配置中的镜像:${NC}"
    kubectl get deployment $DEPLOYMENT_NAME -o jsonpath='{.spec.template.spec.containers[0].image}'
    echo
    
    echo -e "\n${YELLOW}当前运行的 Pod 中的镜像:${NC}"
    kubectl get pods -l app=$DEPLOYMENT_NAME -o jsonpath='{.items[0].spec.containers[0].image}' 2>/dev/null
    echo
    
    log_warning "注意: Pod 仍在使用旧镜像，因为 Deployment 处于暂停状态"
}

# 恢复 Deployment
resume_deployment() {
    log_step "恢复 Deployment"
    
    log_info "执行恢复命令..."
    kubectl rollout resume deployment/$DEPLOYMENT_NAME
    
    if [ $? -eq 0 ]; then
        log_success "Deployment 已恢复"
        check_pause_status
        
        log_info "等待滚动更新完成..."
        kubectl rollout status deployment/$DEPLOYMENT_NAME --timeout=300s
        
        if [ $? -eq 0 ]; then
            log_success "滚动更新完成"
        else
            log_error "滚动更新失败或超时"
        fi
    else
        log_error "恢复 Deployment 失败"
        exit 1
    fi
}

# 验证更新结果
verify_updates() {
    log_step "验证更新结果"
    
    # 验证镜像版本
    log_info "验证镜像版本..."
    local current_image=$(kubectl get deployment $DEPLOYMENT_NAME -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo -e "${YELLOW}Deployment 配置中的镜像:${NC} $current_image"
    
    local pod_image=$(kubectl get pods -l app=$DEPLOYMENT_NAME -o jsonpath='{.items[0].spec.containers[0].image}' 2>/dev/null)
    echo -e "${YELLOW}Pod 中的镜像:${NC} $pod_image"
    
    if [ "$current_image" = "$UPDATE_IMAGE" ] && [ "$pod_image" = "$UPDATE_IMAGE" ]; then
        log_success "镜像版本更新成功"
    else
        log_warning "镜像版本可能未完全更新"
    fi
    
    # 验证资源限制
    log_info "验证资源限制..."
    kubectl describe deployment $DEPLOYMENT_NAME | grep -A 5 "Limits\|Requests" || log_warning "未找到资源限制信息"
    
    # 验证环境变量
    log_info "验证环境变量..."
    kubectl describe deployment $DEPLOYMENT_NAME | grep -A 5 "Environment" || log_warning "未找到环境变量信息"
    
    # 验证标签和注解
    log_info "验证标签和注解..."
    kubectl describe deployment $DEPLOYMENT_NAME | grep -A 5 "Labels\|Annotations"
}

# 显示版本历史
show_rollout_history() {
    log_step "查看版本历史"
    
    log_info "显示推出历史..."
    kubectl rollout history deployment/$DEPLOYMENT_NAME
    
    log_info "显示最新版本详情..."
    kubectl rollout history deployment/$DEPLOYMENT_NAME --revision=$(kubectl rollout history deployment/$DEPLOYMENT_NAME | tail -1 | awk '{print $1}')
}

# 测试应用程序
test_application() {
    log_step "测试应用程序"
    
    # 获取一个 Pod 进行端口转发测试
    local pod_name=$(kubectl get pods -l app=$DEPLOYMENT_NAME -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [ -z "$pod_name" ]; then
        log_warning "没有找到可用的 Pod 进行测试"
        return
    fi
    
    log_info "使用 Pod: $pod_name 进行测试"
    log_info "启动端口转发 (后台运行)..."
    
    # 启动端口转发
    kubectl port-forward pod/$pod_name 8080:80 &
    local port_forward_pid=$!
    
    # 等待端口转发启动
    sleep 3
    
    # 测试应用程序
    log_info "测试应用程序响应..."
    if curl -s http://localhost:8080 > /dev/null; then
        log_success "应用程序响应正常"
        echo -e "${CYAN}您可以在浏览器中访问: http://localhost:8080${NC}"
    else
        log_warning "应用程序可能未正常响应"
    fi
    
    # 清理端口转发
    log_info "清理端口转发..."
    kill $port_forward_pid 2>/dev/null
    wait $port_forward_pid 2>/dev/null
}

# 清理资源
cleanup_resources() {
    log_step "清理资源"
    
    echo -e "${YELLOW}选择清理选项:${NC}"
    echo "1. 完整清理 (删除 Deployment)"
    echo "2. 保留资源 (仅显示状态)"
    echo "3. 跳过清理"
    
    read -p "请选择 (1-3): " choice
    
    case $choice in
        1)
            log_info "删除 Deployment..."
            kubectl delete deployment $DEPLOYMENT_NAME
            log_success "资源已清理"
            ;;
        2)
            log_info "保留资源，显示最终状态"
            show_status "最终状态"
            ;;
        3)
            log_info "跳过清理"
            ;;
        *)
            log_warning "无效选择，跳过清理"
            ;;
    esac
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}Kubernetes Deployment 暂停和恢复演示脚本${NC}"
    echo
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -i, --info     显示脚本信息"
    echo "  -s, --status   仅显示当前状态"
    echo "  -c, --cleanup  仅执行清理操作"
    echo
    echo "演示步骤:"
    echo "  1. 创建初始 Deployment"
    echo "  2. 暂停 Deployment"
    echo "  3. 在暂停状态下进行批量更改"
    echo "  4. 恢复 Deployment"
    echo "  5. 验证更新结果"
    echo "  6. 测试应用程序"
    echo "  7. 清理资源"
}

# 显示脚本信息
show_info() {
    echo -e "${CYAN}脚本信息:${NC}"
    echo "  名称: Kubernetes Deployment 暂停和恢复演示"
    echo "  版本: 1.0"
    echo "  Deployment: $DEPLOYMENT_NAME"
    echo "  命名空间: $NAMESPACE"
    echo "  初始镜像: $INITIAL_IMAGE"
    echo "  更新镜像: $UPDATE_IMAGE"
    echo "  副本数: $REPLICAS"
}

# 主函数
main() {
    # 解析命令行参数
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -i|--info)
            show_info
            exit 0
            ;;
        -s|--status)
            check_kubectl
            show_status "当前状态"
            exit 0
            ;;
        -c|--cleanup)
            check_kubectl
            cleanup_resources
            exit 0
            ;;
    esac
    
    # 显示欢迎信息
    echo -e "${CYAN}🚀 Kubernetes Deployment 暂停和恢复演示${NC}"
    echo -e "${YELLOW}本脚本将演示如何暂停 Deployment、进行批量更改、然后恢复部署${NC}"
    echo
    
    # 检查前置条件
    check_kubectl
    
    # 显示脚本信息
    show_info
    wait_for_user
    
    # 执行演示步骤
    show_status "初始状态"
    wait_for_user
    
    create_deployment
    wait_for_user
    
    show_status "Deployment 创建后状态"
    wait_for_user
    
    pause_deployment
    wait_for_user
    
    make_changes_while_paused
    wait_for_user
    
    show_status "暂停状态下更改后的状态"
    wait_for_user
    
    resume_deployment
    wait_for_user
    
    show_status "恢复后状态"
    wait_for_user
    
    verify_updates
    wait_for_user
    
    show_rollout_history
    wait_for_user
    
    test_application
    wait_for_user
    
    cleanup_resources
    
    log_success "演示完成！"
    echo -e "${CYAN}感谢使用 Kubernetes Deployment 暂停和恢复演示脚本${NC}"
}

# 执行主函数
main "$@"