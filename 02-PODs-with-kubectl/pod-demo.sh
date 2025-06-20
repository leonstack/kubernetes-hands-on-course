#!/bin/bash

# Kubernetes POD 演示脚本
# 作者: Grissom
# 版本: 1.0.0
# 描述: 自动化演示 Kubernetes POD 管理的各个步骤

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
POD_NAME="my-first-pod"
IMAGE_NAME="grissomsh/kubenginx:1.0.0"
SERVICE_NAME="my-first-service"
NAMESPACE="default"
WAIT_TIME=10

# 函数：打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 函数：打印步骤标题
print_step() {
    local step_num=$1
    local step_title=$2
    echo
    print_message $BLUE "=== 步骤 $step_num: $step_title ==="
    echo
}

# 函数：等待用户确认
wait_for_user() {
    if [[ "$INTERACTIVE" == "true" ]]; then
        print_message $YELLOW "按 Enter 键继续..."
        read
    else
        sleep 2
    fi
}

# 函数：检查命令是否成功
check_command() {
    if [ $? -eq 0 ]; then
        print_message $GREEN "✅ 命令执行成功"
    else
        print_message $RED "❌ 命令执行失败"
        exit 1
    fi
}

# 函数：检查 kubectl 是否可用
check_kubectl() {
    print_step "0" "环境检查"
    
    if ! command -v kubectl &> /dev/null; then
        print_message $RED "❌ kubectl 命令未找到，请先安装 kubectl"
        exit 1
    fi
    
    print_message $GREEN "✅ kubectl 已安装"
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        print_message $RED "❌ 无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    print_message $GREEN "✅ Kubernetes 集群连接正常"
    wait_for_user
}

# 函数：获取工作节点状态
get_nodes_status() {
    print_step "1" "获取工作节点状态"
    
    print_message $CYAN "获取工作节点状态..."
    kubectl get nodes
    echo
    
    print_message $CYAN "使用 wide 选项获取详细信息..."
    kubectl get nodes -o wide
    
    wait_for_user
}

# 函数：创建 Pod
create_pod() {
    print_step "2" "创建 Pod"
    
    # 检查 Pod 是否已存在
    if kubectl get pod $POD_NAME &> /dev/null; then
        print_message $YELLOW "⚠️  Pod $POD_NAME 已存在，先删除..."
        kubectl delete pod $POD_NAME --grace-period=0 --force
        sleep 5
    fi
    
    print_message $CYAN "创建 Pod: $POD_NAME"
    kubectl run $POD_NAME --image=$IMAGE_NAME
    check_command
    
    print_message $CYAN "等待 Pod 启动..."
    kubectl wait --for=condition=Ready pod/$POD_NAME --timeout=300s
    check_command
    
    wait_for_user
}

# 函数：列出 Pod
list_pods() {
    print_step "3" "列出 Pod"
    
    print_message $CYAN "列出所有 Pod..."
    kubectl get pods
    echo
    
    print_message $CYAN "使用 wide 选项列出 Pod..."
    kubectl get pods -o wide
    
    wait_for_user
}

# 函数：描述 Pod
describe_pod() {
    print_step "4" "描述 Pod"
    
    print_message $CYAN "描述 Pod $POD_NAME..."
    kubectl describe pod $POD_NAME
    
    wait_for_user
}

# 函数：创建 Service
create_service() {
    print_step "5" "创建 NodePort Service"
    
    # 检查 Service 是否已存在
    if kubectl get service $SERVICE_NAME &> /dev/null; then
        print_message $YELLOW "⚠️  Service $SERVICE_NAME 已存在，先删除..."
        kubectl delete service $SERVICE_NAME
        sleep 2
    fi
    
    print_message $CYAN "将 Pod 暴露为 NodePort Service..."
    kubectl expose pod $POD_NAME --type=NodePort --port=80 --name=$SERVICE_NAME
    check_command
    
    print_message $CYAN "获取 Service 信息..."
    kubectl get service $SERVICE_NAME
    echo
    
    # 获取 NodePort
    NODE_PORT=$(kubectl get service $SERVICE_NAME -o jsonpath='{.spec.ports[0].nodePort}')
    print_message $GREEN "✅ Service 已创建，NodePort: $NODE_PORT"
    
    # 获取节点 IP
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    if [ -z "$NODE_IP" ]; then
        NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    fi
    
    print_message $GREEN "🌐 访问地址: http://$NODE_IP:$NODE_PORT"
    
    wait_for_user
}

# 函数：测试应用程序访问
test_application() {
    print_step "6" "测试应用程序访问"
    
    # 获取 Service 信息
    SERVICE_IP=$(kubectl get service $SERVICE_NAME -o jsonpath='{.spec.clusterIP}')
    SERVICE_PORT=$(kubectl get service $SERVICE_NAME -o jsonpath='{.spec.ports[0].port}')
    
    print_message $CYAN "使用临时 Pod 测试应用程序访问..."
    kubectl run test-pod --image=busybox --rm -it --restart=Never -- wget -qO- http://$SERVICE_IP:$SERVICE_PORT || true
    
    wait_for_user
}

# 函数：查看 Pod 日志
view_logs() {
    print_step "7" "查看 Pod 日志"
    
    print_message $CYAN "查看 Pod 日志..."
    kubectl logs $POD_NAME
    
    wait_for_user
}

# 函数：与 Pod 交互
interact_with_pod() {
    print_step "8" "与 Pod 交互"
    
    print_message $CYAN "执行容器内命令..."
    kubectl exec $POD_NAME -- ls -la /usr/share/nginx/html/
    echo
    
    print_message $CYAN "查看 nginx 配置..."
    kubectl exec $POD_NAME -- cat /etc/nginx/nginx.conf | head -20
    echo
    
    print_message $CYAN "查看环境变量..."
    kubectl exec $POD_NAME -- env | grep -E "(KUBERNETES|POD|SERVICE)"
    
    wait_for_user
}

# 函数：获取 YAML 输出
get_yaml_output() {
    print_step "9" "获取 YAML 输出"
    
    print_message $CYAN "获取 Pod YAML 定义..."
    kubectl get pod $POD_NAME -o yaml > pod-definition.yaml
    print_message $GREEN "✅ Pod YAML 已保存到 pod-definition.yaml"
    echo
    
    print_message $CYAN "获取 Service YAML 定义..."
    kubectl get service $SERVICE_NAME -o yaml > service-definition.yaml
    print_message $GREEN "✅ Service YAML 已保存到 service-definition.yaml"
    
    wait_for_user
}

# 函数：演示故障排查
demonstrate_troubleshooting() {
    print_step "10" "故障排查演示"
    
    print_message $CYAN "演示常用的故障排查命令..."
    
    echo "1. 查看 Pod 详细信息:"
    kubectl describe pod $POD_NAME | head -30
    echo
    
    echo "2. 查看事件:"
    kubectl get events --field-selector involvedObject.name=$POD_NAME
    echo
    
    echo "3. 查看资源使用情况:"
    kubectl top pod $POD_NAME 2>/dev/null || print_message $YELLOW "⚠️  metrics-server 未安装，无法显示资源使用情况"
    
    wait_for_user
}

# 函数：清理资源
cleanup() {
    print_step "11" "清理资源"
    
    print_message $CYAN "清理演示资源..."
    
    # 删除 Service
    if kubectl get service $SERVICE_NAME &> /dev/null; then
        kubectl delete service $SERVICE_NAME
        print_message $GREEN "✅ Service $SERVICE_NAME 已删除"
    fi
    
    # 删除 Pod
    if kubectl get pod $POD_NAME &> /dev/null; then
        kubectl delete pod $POD_NAME --grace-period=0 --force
        print_message $GREEN "✅ Pod $POD_NAME 已删除"
    fi
    
    # 删除生成的 YAML 文件
    rm -f pod-definition.yaml service-definition.yaml
    print_message $GREEN "✅ 临时文件已清理"
    
    print_message $GREEN "🎉 清理完成！"
}

# 函数：显示帮助信息
show_help() {
    cat << EOF
Kubernetes POD 演示脚本

用法: $0 [选项]

选项:
  --step <number>     运行特定步骤 (1-11)
  --cleanup          只执行清理操作
  --interactive      交互模式 (默认)
  --non-interactive  非交互模式
  --help             显示此帮助信息

步骤说明:
  1  - 获取工作节点状态
  2  - 创建 Pod
  3  - 列出 Pod
  4  - 描述 Pod
  5  - 创建 NodePort Service
  6  - 测试应用程序访问
  7  - 查看 Pod 日志
  8  - 与 Pod 交互
  9  - 获取 YAML 输出
  10 - 故障排查演示
  11 - 清理资源

示例:
  $0                    # 运行完整演示
  $0 --step 2           # 只运行步骤 2
  $0 --cleanup          # 清理所有资源
  $0 --non-interactive  # 非交互模式运行

EOF
}

# 主函数
main() {
    local step=""
    local cleanup_only=false
    INTERACTIVE=true
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --step)
                step="$2"
                shift 2
                ;;
            --cleanup)
                cleanup_only=true
                shift
                ;;
            --interactive)
                INTERACTIVE=true
                shift
                ;;
            --non-interactive)
                INTERACTIVE=false
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_message $RED "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 显示欢迎信息
    print_message $PURPLE "🚀 Kubernetes POD 演示脚本"
    print_message $PURPLE "📚 本脚本将演示 Kubernetes POD 管理的各个步骤"
    echo
    
    # 如果只是清理，直接执行清理并退出
    if [[ "$cleanup_only" == "true" ]]; then
        cleanup
        exit 0
    fi
    
    # 环境检查
    check_kubectl
    
    # 如果指定了特定步骤
    if [[ -n "$step" ]]; then
        case $step in
            1) get_nodes_status ;;
            2) create_pod ;;
            3) list_pods ;;
            4) describe_pod ;;
            5) create_service ;;
            6) test_application ;;
            7) view_logs ;;
            8) interact_with_pod ;;
            9) get_yaml_output ;;
            10) demonstrate_troubleshooting ;;
            11) cleanup ;;
            *)
                print_message $RED "无效的步骤号: $step (有效范围: 1-11)"
                exit 1
                ;;
        esac
    else
        # 运行完整演示
        get_nodes_status
        create_pod
        list_pods
        describe_pod
        create_service
        test_application
        view_logs
        interact_with_pod
        get_yaml_output
        demonstrate_troubleshooting
        
        # 询问是否清理
        if [[ "$INTERACTIVE" == "true" ]]; then
            echo
            print_message $YELLOW "是否要清理演示资源？(y/N)"
            read -r response
            if [[ "$response" =~ ^[Yy]$ ]]; then
                cleanup
            else
                print_message $CYAN "资源保留，您可以稍后运行 '$0 --cleanup' 来清理"
            fi
        else
            cleanup
        fi
    fi
    
    print_message $GREEN "🎉 演示完成！"
}

# 捕获中断信号，确保清理
trap 'print_message $RED "\n❌ 脚本被中断，正在清理..."; cleanup; exit 1' INT TERM

# 运行主函数
main "$@"