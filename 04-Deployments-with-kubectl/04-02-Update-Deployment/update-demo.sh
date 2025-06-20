#!/bin/bash

# 本脚本演示两种 Deployment 更新方法：kubectl set image 和 kubectl edit
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

log_substep() {
    echo -e "\n${CYAN}--- $1 ---${NC}"
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

# 检查前置条件
check_prerequisites() {
    log_step "检查前置条件"
    
    # 检查 Deployment 是否存在
    if ! kubectl get deployment my-first-deployment &> /dev/null; then
        log_error "未找到 my-first-deployment，请先运行创建 Deployment 的教程"
        echo "提示：请先执行 04-01-CreateDeployment-Scaling-and-Expose-as-Service 教程"
        exit 1
    fi
    
    # 检查 Service 是否存在
    if ! kubectl get service my-first-deployment-service &> /dev/null; then
        log_warning "未找到 my-first-deployment-service，将跳过服务测试"
        SERVICE_EXISTS=false
    else
        SERVICE_EXISTS=true
    fi
    
    log_success "前置条件检查完成"
}

# 显示当前状态
show_current_status() {
    log_step "当前状态概览"
    
    echo "=== Deployment 状态 ==="
    kubectl get deployment my-first-deployment -o wide
    
    echo -e "\n=== 当前镜像版本 ==="
    CURRENT_IMAGE=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "当前镜像: $CURRENT_IMAGE"
    
    echo -e "\n=== ReplicaSet 状态 ==="
    kubectl get rs -l app=my-first-deployment
    
    echo -e "\n=== Pod 状态 ==="
    kubectl get pods -l app=my-first-deployment
    
    if [ "$SERVICE_EXISTS" = true ]; then
        echo -e "\n=== Service 状态 ==="
        kubectl get svc my-first-deployment-service
    fi
    
    echo -e "\n=== 更新历史 ==="
    kubectl rollout history deployment/my-first-deployment
}

# 方法一：使用 kubectl set image 更新
update_with_set_image() {
    log_step "方法一：使用 kubectl set image 更新 (V1 → V2)"
    
    log_substep "准备更新"
    
    # 获取容器名称
    CONTAINER_NAME=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].name}')
    log_info "容器名称: $CONTAINER_NAME"
    
    # 备份当前配置
    kubectl get deployment my-first-deployment -o yaml > deployment-backup-v1.yaml
    log_info "当前配置已备份到 deployment-backup-v1.yaml"
    
    wait_for_user
    
    log_substep "执行镜像更新"
    
    # 执行更新
    log_info "更新镜像版本从 1.0.0 到 2.0.0"
    kubectl set image deployment/my-first-deployment $CONTAINER_NAME=grissomsh/kubenginx:2.0.0 --record=true
    
    if [ $? -eq 0 ]; then
        log_success "更新命令执行成功"
    else
        log_error "更新命令执行失败"
        return 1
    fi
    
    wait_for_user
    
    log_substep "监控更新过程"
    
    # 监控更新状态
    log_info "监控滚动更新过程..."
    kubectl rollout status deployment/my-first-deployment --timeout=300s
    
    wait_for_user
    
    log_substep "验证更新结果"
    
    # 验证新镜像版本
    NEW_IMAGE=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "新镜像版本: $NEW_IMAGE"
    
    # 查看 ReplicaSet 变化
    echo -e "\n=== ReplicaSet 状态 ==="
    kubectl get rs -l app=my-first-deployment -o wide
    
    # 查看 Pod 状态
    echo -e "\n=== Pod 状态 ==="
    kubectl get pods -l app=my-first-deployment --show-labels
    
    # 验证 Pod 镜像
    echo -e "\n=== Pod 镜像验证 ==="
    kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[*].spec.containers[0].image}'
    echo
    
    wait_for_user
    
    log_substep "查看更新历史"
    kubectl rollout history deployment/my-first-deployment
    
    wait_for_user
    
    # 测试应用程序
    if [ "$SERVICE_EXISTS" = true ]; then
        test_application "V2"
    fi
    
    log_success "方法一更新完成"
}

# 方法二：使用 kubectl edit 更新
update_with_edit() {
    log_step "方法二：使用 kubectl edit 更新 (V2 → V3)"
    
    log_substep "准备编辑环境"
    
    # 设置编辑器
    if [ -z "$EDITOR" ]; then
        export EDITOR=nano
        log_info "设置默认编辑器为 nano"
    else
        log_info "使用编辑器: $EDITOR"
    fi
    
    # 备份当前配置
    kubectl get deployment my-first-deployment -o yaml > deployment-backup-v2.yaml
    log_info "当前配置已备份到 deployment-backup-v2.yaml"
    
    # 显示编辑指南
    echo -e "\n${YELLOW}编辑指南：${NC}"
    echo "1. 找到 image: grissomsh/kubenginx:2.0.0 行"
    echo "2. 将 2.0.0 修改为 3.0.0"
    echo "3. 保存并退出编辑器"
    echo "   - nano: Ctrl+X, Y, Enter"
    echo "   - vim: Esc, :wq, Enter"
    
    wait_for_user
    
    log_substep "启动交互式编辑"
    
    # 创建一个临时的编辑脚本（用于自动化演示）
    if [ "$AUTO_EDIT" = "true" ]; then
        log_info "自动编辑模式：直接更新镜像版本"
        kubectl patch deployment my-first-deployment -p '{
            "spec": {
                "template": {
                    "spec": {
                        "containers": [{
                            "name": "kubenginx",
                            "image": "grissomsh/kubenginx:3.0.0"
                        }]
                    }
                }
            }
        }' --record=true
    else
        log_info "启动交互式编辑器..."
        kubectl edit deployment/my-first-deployment --record=true
    fi
    
    if [ $? -eq 0 ]; then
        log_success "编辑完成"
    else
        log_error "编辑失败或被取消"
        return 1
    fi
    
    wait_for_user
    
    log_substep "监控更新过程"
    
    # 监控更新状态
    log_info "监控滚动更新过程..."
    kubectl rollout status deployment/my-first-deployment --timeout=300s
    
    wait_for_user
    
    log_substep "验证更新结果"
    
    # 验证新镜像版本
    NEW_IMAGE=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "新镜像版本: $NEW_IMAGE"
    
    # 查看所有 ReplicaSet（应该有 3 个）
    echo -e "\n=== 所有 ReplicaSet 状态 ==="
    kubectl get rs -l app=my-first-deployment -o wide
    
    # 查看 Pod 状态
    echo -e "\n=== Pod 状态 ==="
    kubectl get pods -l app=my-first-deployment --show-labels
    
    wait_for_user
    
    log_substep "查看完整更新历史"
    kubectl rollout history deployment/my-first-deployment
    
    # 查看各个版本详情
    echo -e "\n=== 版本详情对比 ==="
    for revision in 1 2 3; do
        echo "--- 版本 $revision ---"
        kubectl rollout history deployment/my-first-deployment --revision=$revision 2>/dev/null || echo "版本 $revision 不存在"
    done
    
    wait_for_user
    
    # 测试应用程序
    if [ "$SERVICE_EXISTS" = true ]; then
        test_application "V3"
    fi
    
    log_success "方法二更新完成"
}

# 测试应用程序
test_application() {
    local expected_version=$1
    log_substep "测试应用程序 (期望版本: $expected_version)"
    
    # 获取访问信息
    NODE_PORT=$(kubectl get svc my-first-deployment-service -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null)
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null)
    
    if [ -n "$NODE_PORT" ] && [ -n "$NODE_IP" ]; then
        echo "应用程序访问地址: http://$NODE_IP:$NODE_PORT"
    fi
    
    # 使用端口转发进行测试
    log_info "启动端口转发进行本地测试..."
    kubectl port-forward svc/my-first-deployment-service 8080:80 &
    PORT_FORWARD_PID=$!
    sleep 5
    
    # 测试连接
    if command -v curl &> /dev/null; then
        log_info "测试应用程序响应..."
        RESPONSE=$(curl -s http://localhost:8080 2>/dev/null || echo "连接失败")
        
        if echo "$RESPONSE" | grep -i "version" > /dev/null; then
            VERSION_INFO=$(echo "$RESPONSE" | grep -i "version" | head -1)
            echo "版本信息: $VERSION_INFO"
            
            if echo "$VERSION_INFO" | grep -i "$expected_version" > /dev/null; then
                log_success "版本验证成功：应用程序显示 $expected_version"
            else
                log_warning "版本验证失败：期望 $expected_version，实际显示 $VERSION_INFO"
            fi
        else
            log_warning "未找到版本信息，但应用程序响应正常"
        fi
    else
        log_info "curl 命令不可用，请在浏览器中访问 http://localhost:8080"
        echo "期望看到: Application Version:$expected_version"
    fi
    
    # 停止端口转发
    kill $PORT_FORWARD_PID 2>/dev/null
    sleep 2
}

# 演示回滚操作
demo_rollback() {
    log_step "演示回滚操作"
    
    log_info "当前版本历史："
    kubectl rollout history deployment/my-first-deployment
    
    wait_for_user
    
    log_substep "回滚到上一个版本"
    
    # 执行回滚
    kubectl rollout undo deployment/my-first-deployment
    
    # 监控回滚过程
    log_info "监控回滚过程..."
    kubectl rollout status deployment/my-first-deployment
    
    # 验证回滚结果
    ROLLBACK_IMAGE=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "回滚后镜像版本: $ROLLBACK_IMAGE"
    
    wait_for_user
    
    log_substep "回滚到特定版本 (版本 3)"
    
    # 回滚到版本 3
    kubectl rollout undo deployment/my-first-deployment --to-revision=3
    
    # 监控回滚过程
    kubectl rollout status deployment/my-first-deployment
    
    # 验证结果
    FINAL_IMAGE=$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "最终镜像版本: $FINAL_IMAGE"
    
    log_success "回滚演示完成"
}

# 显示最终状态
show_final_status() {
    log_step "最终状态总览"
    
    echo "=== Deployment 状态 ==="
    kubectl get deployment my-first-deployment -o wide
    
    echo -e "\n=== 所有 ReplicaSet ==="
    kubectl get rs -l app=my-first-deployment
    
    echo -e "\n=== 活跃 Pod ==="
    kubectl get pods -l app=my-first-deployment
    
    echo -e "\n=== 完整更新历史 ==="
    kubectl rollout history deployment/my-first-deployment
    
    echo -e "\n=== 当前镜像版本 ==="
    kubectl get deployment my-first-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'
    echo
    
    if [ "$SERVICE_EXISTS" = true ]; then
        echo -e "\n=== Service 信息 ==="
        kubectl get svc my-first-deployment-service
    fi
}

# 清理演示资源
cleanup_demo() {
    log_step "清理演示资源"
    
    echo -e "\n${YELLOW}是否要清理创建的备份文件？(y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        rm -f deployment-backup-v*.yaml
        log_success "备份文件已清理"
    else
        log_info "保留备份文件：deployment-backup-v1.yaml, deployment-backup-v2.yaml"
    fi
    
    echo -e "\n${YELLOW}是否要重置 Deployment 到初始状态？(y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        log_info "重置到版本 1.0.0..."
        kubectl set image deployment/my-first-deployment kubenginx=grissomsh/kubenginx:1.0.0
        kubectl rollout status deployment/my-first-deployment
        log_success "已重置到初始状态"
    else
        log_info "保持当前状态"
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}Kubernetes Deployment 更新演示脚本${NC}"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo "  -a, --auto-edit   自动编辑模式（跳过交互式编辑器）"
    echo "  -s, --skip-test   跳过应用程序测试"
    echo "  -r, --rollback    仅演示回滚操作"
    echo ""
    echo "本脚本将演示以下操作:"
    echo "  1. 检查前置条件和当前状态"
    echo "  2. 方法一：使用 kubectl set image 更新 (V1→V2)"
    echo "  3. 方法二：使用 kubectl edit 更新 (V2→V3)"
    echo "  4. 演示回滚操作"
    echo "  5. 显示最终状态"
    echo ""
    echo "前置条件:"
    echo "  - 需要运行中的 my-first-deployment"
    echo "  - 建议先执行 04-01 教程创建基础环境"
}

# 主函数
main() {
    echo -e "${CYAN}"
    echo "================================================"
    echo "    Kubernetes Deployment 更新演示脚本"
    echo "================================================"
    echo -e "${NC}"
    
    # 解析命令行参数
    AUTO_EDIT=false
    SKIP_TEST=false
    ROLLBACK_ONLY=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -a|--auto-edit)
                AUTO_EDIT=true
                shift
                ;;
            -s|--skip-test)
                SKIP_TEST=true
                shift
                ;;
            -r|--rollback)
                ROLLBACK_ONLY=true
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
    check_prerequisites
    
    if [ "$ROLLBACK_ONLY" = true ]; then
        demo_rollback
        show_final_status
        exit 0
    fi
    
    # 显示当前状态
    show_current_status
    wait_for_user
    
    # 执行更新演示
    update_with_set_image
    wait_for_user
    
    update_with_edit
    wait_for_user
    
    # 演示回滚
    demo_rollback
    wait_for_user
    
    # 显示最终状态
    show_final_status
    wait_for_user
    
    # 清理
    cleanup_demo
    
    echo -e "\n${GREEN}"
    echo "================================================"
    echo "    Deployment 更新演示完成！"
    echo "================================================"
    echo -e "${NC}"
    
    echo "学习要点总结:"
    echo "✅ kubectl set image 快速镜像更新"
    echo "✅ kubectl edit 交互式配置修改"
    echo "✅ 滚动更新过程监控和验证"
    echo "✅ ReplicaSet 版本管理"
    echo "✅ 更新历史追踪和回滚操作"
    echo "✅ 应用程序功能测试"
    
    echo -e "\n下一步学习建议:"
    echo "- 04-03-Rollback-Deployment: 深入学习回滚操作"
    echo "- 04-04-Pause-and-Resume-Deployment: 学习暂停和恢复"
    echo "- 高级部署策略：蓝绿部署、金丝雀发布"
    echo "- 生产环境最佳实践和监控"
}

# 错误处理
trap 'log_error "脚本执行中断"; exit 1' INT TERM

# 执行主函数
main "$@"