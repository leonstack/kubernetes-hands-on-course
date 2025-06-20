#!/bin/bash

# ReplicaSet kubectl 管理演示脚本
# 基于 README.md 文档生成的完整演示

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 等待函数
wait_for_input() {
    echo -e "${YELLOW}按 Enter 键继续...${NC}"
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
    
    log_success "kubectl 和集群连接正常"
}

# 步骤 01: ReplicaSet 介绍
step01_introduction() {
    log_info "=== 步骤 01: ReplicaSet 介绍 ==="
    echo "ReplicaSet 是 Kubernetes 中用于确保指定数量的 Pod 副本始终运行的控制器"
    echo "主要优势:"
    echo "- 高可用性：确保应用程序始终有足够的副本运行"
    echo "- 负载分布：将流量分散到多个 Pod 实例"
    echo "- 自动恢复：Pod 故障时自动替换"
    echo "- 水平扩展：可以轻松调整副本数量"
    wait_for_input
}

# 步骤 02: 创建 ReplicaSet
step02_create_replicaset() {
    log_info "=== 步骤 02: 创建 ReplicaSet ==="
    
    # 创建 YAML 文件
    cat > replicaset-demo.yml << 'EOF'
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: my-helloworld-rs
  labels:
    app: my-helloworld
    version: v1.0.0
    component: frontend
    tier: web
  annotations:
    description: "Hello World ReplicaSet for Kubernetes fundamentals demo"
    maintainer: "kubernetes-fundamentals-team"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-helloworld
      version: v1.0.0
  template:
    metadata:
      labels:
        app: my-helloworld
        version: v1.0.0
        component: frontend
        tier: web
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
    spec:
      containers:
      - name: my-helloworld-app
        image: grissomsh/kube-helloworld:1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /hello
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /hello
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        env:
        - name: APP_NAME
          value: "my-helloworld"
        - name: APP_VERSION
          value: "1.0.0"
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 1000
EOF

    log_info "创建 ReplicaSet YAML 文件完成"
    
    # 应用 ReplicaSet
    log_info "应用 ReplicaSet 配置..."
    kubectl apply -f replicaset-demo.yml
    
    # 等待 Pod 就绪
    log_info "等待 Pod 就绪..."
    kubectl wait --for=condition=ready pod -l app=my-helloworld --timeout=120s
    
    # 查看 ReplicaSet
    log_info "查看 ReplicaSet 状态:"
    kubectl get replicaset
    kubectl get rs -o wide
    
    # 查看 Pod
    log_info "查看 Pod 状态:"
    kubectl get pods -l app=my-helloworld -o wide
    
    # 描述 ReplicaSet
    log_info "ReplicaSet 详细信息:"
    kubectl describe rs/my-helloworld-rs
    
    log_success "ReplicaSet 创建完成"
    wait_for_input
}

# 步骤 03: 暴露为 Service
step03_expose_service() {
    log_info "=== 步骤 03: 将 ReplicaSet 暴露为 Service ==="
    
    # 创建 Service
    log_info "创建 Service..."
    kubectl expose rs my-helloworld-rs \
        --type=NodePort \
        --port=80 \
        --target-port=8080 \
        --name=my-helloworld-rs-service
    
    # 查看 Service
    log_info "查看 Service 信息:"
    kubectl get service
    kubectl get svc -o wide
    kubectl describe svc my-helloworld-rs-service
    
    # 查看端点
    log_info "查看 Service 端点:"
    kubectl get endpoints my-helloworld-rs-service
    
    # 获取访问信息
    NODE_PORT=$(kubectl get svc my-helloworld-rs-service -o jsonpath='{.spec.ports[0].nodePort}')
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    
    log_info "访问信息:"
    echo "NodePort: $NODE_PORT"
    echo "Node IP: $NODE_IP"
    echo "访问 URL: http://$NODE_IP:$NODE_PORT/hello"
    
    # 测试连接
    log_info "测试 Service 连接 (使用 port-forward):"
    kubectl port-forward svc/my-helloworld-rs-service 8080:80 &
    PORT_FORWARD_PID=$!
    sleep 5
    
    log_info "测试应用程序响应:"
    for i in {1..3}; do
        echo "测试 $i:"
        curl -s http://localhost:8080/hello || echo "连接失败"
        sleep 1
    done
    
    # 停止 port-forward
    kill $PORT_FORWARD_PID 2>/dev/null || true
    
    log_success "Service 创建和测试完成"
    wait_for_input
}

# 步骤 04: 测试高可用性
step04_test_reliability() {
    log_info "=== 步骤 04: 测试 ReplicaSet 可靠性/高可用性 ==="
    
    # 显示当前状态
    log_info "当前 Pod 状态:"
    kubectl get pods -l app=my-helloworld -o wide
    
    echo "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.replicas}')"
    echo "就绪副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.readyReplicas}')"
    
    # 选择一个 Pod 进行删除测试
    POD_NAME=$(kubectl get pods -l app=my-helloworld -o jsonpath='{.items[0].metadata.name}')
    log_warning "将要删除 Pod: $POD_NAME 来测试自愈能力"
    
    # 启动监控
    log_info "启动 Pod 状态监控..."
    kubectl get pods -l app=my-helloworld -w &
    WATCH_PID=$!
    
    sleep 2
    
    # 删除 Pod
    log_info "删除 Pod 模拟故障..."
    kubectl delete pod $POD_NAME
    
    # 等待自愈
    log_info "等待 ReplicaSet 自动创建新 Pod..."
    sleep 10
    
    # 停止监控
    kill $WATCH_PID 2>/dev/null || true
    
    # 验证自愈结果
    log_info "验证自愈结果:"
    kubectl get pods -l app=my-helloworld --sort-by=.metadata.creationTimestamp
    kubectl get rs my-helloworld-rs
    
    # 查看相关事件
    log_info "相关事件:"
    kubectl get events --sort-by=.metadata.creationTimestamp | grep my-helloworld | tail -5
    
    log_success "高可用性测试完成 - ReplicaSet 自动恢复了被删除的 Pod"
    wait_for_input
}

# 步骤 05: 扩展 ReplicaSet
step05_scale_up() {
    log_info "=== 步骤 05: 扩展 ReplicaSet ==="
    
    # 当前状态
    log_info "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.spec.replicas}')"
    
    # 扩容到 6 个副本
    log_info "扩容到 6 个副本..."
    kubectl scale --replicas=6 rs/my-helloworld-rs
    
    # 监控扩容过程
    log_info "监控扩容过程..."
    kubectl get rs my-helloworld-rs -w &
    WATCH_PID=$!
    
    sleep 15
    kill $WATCH_PID 2>/dev/null || true
    
    # 验证扩容结果
    log_info "扩容后状态:"
    kubectl get rs my-helloworld-rs
    kubectl get pods -l app=my-helloworld -o wide
    
    echo "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.replicas}')"
    echo "就绪副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.readyReplicas}')"
    
    # 再次扩容到 10 个副本
    log_info "进一步扩容到 10 个副本..."
    kubectl scale --replicas=10 rs/my-helloworld-rs
    
    # 等待扩容完成
    log_info "等待扩容完成..."
    sleep 20
    
    log_info "最终扩容状态:"
    kubectl get rs my-helloworld-rs
    kubectl get pods -l app=my-helloworld
    
    log_success "扩容操作完成"
    wait_for_input
}

# 步骤 06: 缩减 ReplicaSet
step06_scale_down() {
    log_info "=== 步骤 06: 缩减 ReplicaSet ==="
    
    log_info "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.spec.replicas}')"
    
    # 渐进式缩容
    log_info "渐进式缩容 - 第一步：缩减到 5 个副本"
    kubectl scale --replicas=5 rs/my-helloworld-rs
    sleep 10
    
    log_info "第二步：缩减到 3 个副本"
    kubectl scale --replicas=3 rs/my-helloworld-rs
    sleep 10
    
    log_info "最终：缩减到 2 个副本"
    kubectl scale --replicas=2 rs/my-helloworld-rs
    sleep 10
    
    # 验证缩容结果
    log_info "缩容后状态:"
    kubectl get rs my-helloworld-rs
    kubectl get pods -l app=my-helloworld
    
    # 查看缩容事件
    log_info "缩容相关事件:"
    kubectl get events --sort-by=.metadata.creationTimestamp | grep -E "(Killing|SuccessfulDelete)" | tail -5
    
    log_success "缩容操作完成"
    wait_for_input
}

# 步骤 07: 清理资源
step07_cleanup() {
    log_info "=== 步骤 07: 清理资源 ==="
    
    # 清理前检查
    log_info "清理前资源状态:"
    echo "=== ReplicaSet 状态 ==="
    kubectl get rs -l app=my-helloworld
    
    echo "\n=== Pod 状态 ==="
    kubectl get pods -l app=my-helloworld
    
    echo "\n=== Service 状态 ==="
    kubectl get svc -l app=my-helloworld
    
    log_warning "即将开始清理资源..."
    wait_for_input
    
    # 逐步清理
    log_info "步骤1：删除 Service"
    kubectl delete svc my-helloworld-rs-service
    kubectl get svc | grep my-helloworld || log_success "Service 已删除"
    
    log_info "步骤2：缩减 ReplicaSet 到 0"
    kubectl scale --replicas=0 rs/my-helloworld-rs
    sleep 10
    kubectl get pods -l app=my-helloworld
    
    log_info "步骤3：删除 ReplicaSet"
    kubectl delete rs my-helloworld-rs
    
    # 最终验证
    log_info "清理验证:"
    kubectl get rs -l app=my-helloworld 2>/dev/null || log_success "ReplicaSet 已删除"
    kubectl get pods -l app=my-helloworld 2>/dev/null || log_success "Pod 已删除"
    kubectl get svc -l app=my-helloworld 2>/dev/null || log_success "Service 已删除"
    
    # 清理 YAML 文件
    if [ -f "replicaset-demo.yml" ]; then
        rm replicaset-demo.yml
        log_info "清理 YAML 文件"
    fi
    
    log_success "所有资源清理完成"
}

# 故障排除函数
troubleshooting() {
    log_info "=== 故障排除信息 ==="
    
    echo "如果遇到问题，可以使用以下命令进行调试:"
    echo ""
    echo "1. 查看 Pod 状态和事件:"
    echo "   kubectl describe pod <pod-name>"
    echo "   kubectl get events --sort-by=.metadata.creationTimestamp"
    echo ""
    echo "2. 查看 Pod 日志:"
    echo "   kubectl logs <pod-name>"
    echo "   kubectl logs <pod-name> --previous"
    echo ""
    echo "3. 检查 ReplicaSet 状态:"
    echo "   kubectl describe rs my-helloworld-rs"
    echo ""
    echo "4. 检查 Service 和端点:"
    echo "   kubectl describe svc my-helloworld-rs-service"
    echo "   kubectl get endpoints my-helloworld-rs-service"
    echo ""
    echo "5. 强制清理资源:"
    echo "   kubectl delete all -l app=my-helloworld"
    echo "   kubectl delete rs my-helloworld-rs --grace-period=0 --force"
}

# 主函数
main() {
    log_info "开始 ReplicaSet kubectl 管理演示"
    log_info "本演示基于 README.md 文档内容"
    echo ""
    
    # 检查环境
    check_kubectl
    
    # 询问用户是否继续
    echo "演示将包含以下步骤:"
    echo "1. ReplicaSet 介绍"
    echo "2. 创建 ReplicaSet"
    echo "3. 暴露为 Service"
    echo "4. 测试高可用性"
    echo "5. 扩展 ReplicaSet"
    echo "6. 缩减 ReplicaSet"
    echo "7. 清理资源"
    echo ""
    
    read -p "是否开始演示？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "演示已取消"
        exit 0
    fi
    
    # 执行演示步骤
    step01_introduction
    step02_create_replicaset
    step03_expose_service
    step04_test_reliability
    step05_scale_up
    step06_scale_down
    step07_cleanup
    
    # 显示故障排除信息
    troubleshooting
    
    log_success "ReplicaSet kubectl 管理演示完成！"
    log_info "感谢使用本演示脚本"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi