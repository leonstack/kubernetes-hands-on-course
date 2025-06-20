#!/bin/bash

# Kubernetes POD 快速开始脚本
# 快速演示单容器和多容器 Pod 的基本操作

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 打印消息函数
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 检查 kubectl
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_message $RED "❌ kubectl 未安装"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        print_message $RED "❌ 无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    print_message $GREEN "✅ 环境检查通过"
}

# 显示菜单
show_menu() {
    echo
    print_message $PURPLE "🚀 Kubernetes POD 快速开始"
    echo
    echo "请选择要执行的操作:"
    echo "1) 运行完整的 POD 演示脚本"
    echo "2) 创建单容器 Pod"
    echo "3) 创建多容器 Pod (Sidecar 模式)"
    echo "4) 查看所有 Pod 状态"
    echo "5) 清理所有演示资源"
    echo "6) 退出"
    echo
}

# 运行完整演示
run_full_demo() {
    print_message $BLUE "启动完整 POD 演示..."
    if [ -f "./pod-demo.sh" ]; then
        ./pod-demo.sh
    else
        print_message $RED "❌ pod-demo.sh 脚本未找到"
    fi
}

# 创建单容器 Pod
create_single_pod() {
    print_message $BLUE "创建单容器 Pod..."
    
    kubectl run simple-pod --image=nginx:1.25-alpine --port=80
    
    print_message $CYAN "等待 Pod 启动..."
    kubectl wait --for=condition=Ready pod/simple-pod --timeout=60s
    
    print_message $GREEN "✅ 单容器 Pod 创建成功"
    kubectl get pod simple-pod -o wide
    
    # 创建 Service
    kubectl expose pod simple-pod --type=NodePort --port=80 --name=simple-service
    print_message $GREEN "✅ Service 创建成功"
    kubectl get service simple-service
}

# 创建多容器 Pod
create_multi_pod() {
    print_message $BLUE "创建多容器 Pod (Sidecar 模式)..."
    
    if [ -f "./multi-container-pod-demo.yaml" ]; then
        kubectl apply -f multi-container-pod-demo.yaml
        
        print_message $CYAN "等待 Pod 启动..."
        kubectl wait --for=condition=Ready pod/multi-container-demo --timeout=120s
        
        print_message $GREEN "✅ 多容器 Pod 创建成功"
        kubectl get pod multi-container-demo -o wide
        
        print_message $CYAN "查看容器状态..."
        kubectl describe pod multi-container-demo | grep -A 10 "Containers:"
        
        print_message $YELLOW "💡 提示: 使用以下命令查看不同容器的日志:"
        echo "  kubectl logs multi-container-demo -c web-server"
        echo "  kubectl logs multi-container-demo -c log-collector"
        echo "  kubectl logs multi-container-demo -c monitoring-agent"
    else
        print_message $RED "❌ multi-container-pod-demo.yaml 文件未找到"
    fi
}

# 查看 Pod 状态
view_pod_status() {
    print_message $BLUE "查看所有 Pod 状态..."
    
    echo "=== Pod 列表 ==="
    kubectl get pods -o wide
    echo
    
    echo "=== Service 列表 ==="
    kubectl get services
    echo
    
    # 如果存在多容器 Pod，显示详细信息
    if kubectl get pod multi-container-demo &> /dev/null; then
        echo "=== 多容器 Pod 详细信息 ==="
        kubectl describe pod multi-container-demo | grep -A 20 "Containers:"
        echo
        
        print_message $CYAN "最近的日志片段:"
        echo "--- Web Server 日志 ---"
        kubectl logs multi-container-demo -c web-server --tail=5 2>/dev/null || echo "暂无日志"
        echo
        echo "--- Log Collector 日志 ---"
        kubectl logs multi-container-demo -c log-collector --tail=5 2>/dev/null || echo "暂无日志"
        echo
        echo "--- Monitoring Agent 日志 ---"
        kubectl logs multi-container-demo -c monitoring-agent --tail=5 2>/dev/null || echo "暂无日志"
    fi
}

# 清理资源
cleanup_resources() {
    print_message $BLUE "清理演示资源..."
    
    # 清理单容器 Pod 和 Service
    kubectl delete pod simple-pod --ignore-not-found=true
    kubectl delete service simple-service --ignore-not-found=true
    
    # 清理多容器 Pod 和相关资源
    if [ -f "./multi-container-pod-demo.yaml" ]; then
        kubectl delete -f multi-container-pod-demo.yaml --ignore-not-found=true
    fi
    
    # 清理演示脚本创建的资源
    kubectl delete pod my-first-pod --ignore-not-found=true
    kubectl delete service my-first-service --ignore-not-found=true
    
    # 清理临时文件
    rm -f pod-definition.yaml service-definition.yaml
    
    print_message $GREEN "✅ 清理完成"
}

# 主函数
main() {
    check_kubectl
    
    while true; do
        show_menu
        read -p "请输入选项 (1-6): " choice
        
        case $choice in
            1)
                run_full_demo
                ;;
            2)
                create_single_pod
                ;;
            3)
                create_multi_pod
                ;;
            4)
                view_pod_status
                ;;
            5)
                cleanup_resources
                ;;
            6)
                print_message $GREEN "👋 再见！"
                exit 0
                ;;
            *)
                print_message $RED "❌ 无效选项，请输入 1-6"
                ;;
        esac
        
        echo
        read -p "按 Enter 键继续..."
    done
}

# 捕获中断信号
trap 'print_message $RED "\n❌ 脚本被中断"; exit 1' INT TERM

# 运行主函数
main "$@"