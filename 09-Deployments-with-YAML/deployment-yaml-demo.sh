#!/bin/bash

# 演示如何使用 YAML 文件创建和管理 Deployment 及 Service
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
    echo -e "${CYAN}  Kubernetes Deployment with YAML Demo${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${BLUE}功能说明：${NC}"
    echo "  • 演示 Deployment YAML 文件的创建和验证"
    echo "  • 创建和管理 Deployment 资源"
    echo "  • 演示滚动更新和版本控制"
    echo "  • 测试回滚操作"
    echo "  • 演示扩缩容和暂停/恢复部署"
    echo "  • 创建和管理 Service 资源"
    echo "  • 网络连接测试和负载均衡验证"
    echo "  • 故障排除和调试技巧"
    echo "  • 最佳实践演示"
    echo ""
}

# 帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --create-all       创建所有资源（Deployment + Service）"
    echo "  --create-deploy    仅创建 Deployment"
    echo "  --create-service   仅创建 Service"
    echo "  --rolling-update   演示滚动更新"
    echo "  --rollback         演示回滚操作"
    echo "  --scale            演示扩缩容操作"
    echo "  --pause-resume     演示暂停和恢复部署"
    echo "  --test-lb          测试负载均衡"
    echo "  --monitor          监控资源状态"
    echo "  --logs             查看 Pod 日志"
    echo "  --debug            调试模式"
    echo "  --cleanup          清理所有资源"
    echo "  --validate         验证 YAML 文件"
    echo "  --best-practices   最佳实践演示"
    echo "  --help             显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --create-all      # 创建所有资源"
    echo "  $0 --rolling-update  # 演示滚动更新"
    echo "  $0 --rollback        # 演示回滚"
    echo "  $0 --cleanup         # 清理资源"
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
        "kube-manifests/02-deployment-definition.yml"
        "kube-manifests/03-deployment-nodeport-servie.yml"
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

# 创建 Deployment
create_deployment() {
    log_step "创建 Deployment..."
    
    if kubectl get deployment myapp3-deployment &> /dev/null; then
        log_warning "Deployment 'myapp3-deployment' 已存在"
        return 0
    fi
    
    log_info "应用 Deployment 配置文件..."
    kubectl apply -f kube-manifests/02-deployment-definition.yml
    
    log_info "等待 Deployment 就绪..."
    kubectl rollout status deployment/myapp3-deployment --timeout=120s
    
    log_success "Deployment 创建成功"
    kubectl get deployment myapp3-deployment
    kubectl get replicaset -l app=myapp3
    kubectl get pods -l app=myapp3 -o wide
}

# 创建 Service
create_service() {
    log_step "创建 Service..."
    
    if kubectl get service deployment-nodeport-service &> /dev/null; then
        log_warning "Service 'deployment-nodeport-service' 已存在"
        return 0
    fi
    
    log_info "应用 Service 配置文件..."
    kubectl apply -f kube-manifests/03-deployment-nodeport-servie.yml
    
    log_success "Service 创建成功"
    kubectl get service deployment-nodeport-service
    
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
    node_port=$(kubectl get service deployment-nodeport-service -o jsonpath='{.spec.ports[0].nodePort}')
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  应用访问信息${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${GREEN}外部访问地址:${NC} http://${node_ip}:${node_port}"
    echo -e "${GREEN}集群内访问:${NC} http://deployment-nodeport-service:80"
    echo -e "${GREEN}Service IP:${NC} $(kubectl get service deployment-nodeport-service -o jsonpath='{.spec.clusterIP}')"
    echo -e "${GREEN}Pod 数量:${NC} $(kubectl get pods -l app=myapp3 --no-headers | wc -l)"
    echo -e "${GREEN}当前版本:${NC} $(kubectl get deployment myapp3-deployment -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')"
    echo ""
}

# 演示滚动更新
demonstrate_rolling_update() {
    log_step "演示滚动更新..."
    
    # 检查 Deployment 是否存在
    if ! kubectl get deployment myapp3-deployment &> /dev/null; then
        log_error "Deployment 'myapp3-deployment' 不存在，请先创建"
        return 1
    fi
    
    log_info "当前 Deployment 状态:"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3 -o wide
    
    # 记录当前版本
    local current_image
    current_image=$(kubectl get deployment myapp3-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    log_info "当前镜像版本: $current_image"
    
    # 更新到新版本
    log_info "更新镜像到 4.0.0 版本..."
    kubectl set image deployment/myapp3-deployment myapp3-container=grissomsh/kubenginx:4.0.0
    
    # 监控滚动更新过程
    log_info "监控滚动更新过程..."
    kubectl rollout status deployment/myapp3-deployment --timeout=120s
    
    log_success "滚动更新完成"
    kubectl get deployment myapp3-deployment
    kubectl get replicaset -l app=myapp3
    kubectl get pods -l app=myapp3 -o wide
    
    # 显示更新历史
    log_info "部署历史:"
    kubectl rollout history deployment/myapp3-deployment
    
    # 验证新版本
    log_info "验证新版本..."
    local new_image
    new_image=$(kubectl get deployment myapp3-deployment -o jsonpath='{.spec.template.spec.containers[0].image}')
    log_success "新镜像版本: $new_image"
}

# 演示回滚操作
demonstrate_rollback() {
    log_step "演示回滚操作..."
    
    # 检查 Deployment 是否存在
    if ! kubectl get deployment myapp3-deployment &> /dev/null; then
        log_error "Deployment 'myapp3-deployment' 不存在，请先创建"
        return 1
    fi
    
    # 显示当前状态
    log_info "当前 Deployment 状态:"
    kubectl get deployment myapp3-deployment
    
    # 显示部署历史
    log_info "部署历史:"
    kubectl rollout history deployment/myapp3-deployment
    
    # 记录当前版本
    local current_revision
    current_revision=$(kubectl get deployment myapp3-deployment -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
    log_info "当前版本: $current_revision"
    
    # 回滚到上一个版本
    log_info "回滚到上一个版本..."
    kubectl rollout undo deployment/myapp3-deployment
    
    # 监控回滚过程
    log_info "监控回滚过程..."
    kubectl rollout status deployment/myapp3-deployment --timeout=120s
    
    log_success "回滚完成"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3 -o wide
    
    # 验证回滚结果
    local new_revision
    new_revision=$(kubectl get deployment myapp3-deployment -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
    log_success "回滚后版本: $new_revision"
    
    # 显示更新后的历史
    log_info "回滚后的部署历史:"
    kubectl rollout history deployment/myapp3-deployment
}

# 演示扩缩容操作
demonstrate_scaling() {
    log_step "演示 Deployment 扩缩容操作..."
    
    # 检查 Deployment 是否存在
    if ! kubectl get deployment myapp3-deployment &> /dev/null; then
        log_error "Deployment 'myapp3-deployment' 不存在，请先创建"
        return 1
    fi
    
    log_info "当前 Deployment 状态:"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3
    
    # 扩容到 5 个副本
    log_info "扩容到 5 个副本..."
    kubectl scale deployment myapp3-deployment --replicas=5
    
    log_info "等待扩容完成..."
    kubectl rollout status deployment/myapp3-deployment --timeout=60s
    
    log_success "扩容完成"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3 -o wide
    
    sleep 3
    
    # 缩容到 2 个副本
    log_info "缩容到 2 个副本..."
    kubectl scale deployment myapp3-deployment --replicas=2
    
    sleep 5
    
    log_success "缩容完成"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3 -o wide
    
    sleep 3
    
    # 恢复到原始副本数
    log_info "恢复到 3 个副本..."
    kubectl scale deployment myapp3-deployment --replicas=3
    
    kubectl rollout status deployment/myapp3-deployment --timeout=60s
    
    log_success "副本数恢复完成"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3 -o wide
}

# 演示暂停和恢复部署
demonstrate_pause_resume() {
    log_step "演示暂停和恢复部署..."
    
    # 检查 Deployment 是否存在
    if ! kubectl get deployment myapp3-deployment &> /dev/null; then
        log_error "Deployment 'myapp3-deployment' 不存在，请先创建"
        return 1
    fi
    
    log_info "当前 Deployment 状态:"
    kubectl get deployment myapp3-deployment
    
    # 暂停部署
    log_info "暂停部署..."
    kubectl rollout pause deployment/myapp3-deployment
    
    # 进行多个更改
    log_info "在暂停状态下进行多个更改..."
    kubectl set image deployment/myapp3-deployment myapp3-container=grissomsh/kubenginx:5.0.0
    kubectl set resources deployment/myapp3-deployment -c=myapp3-container --limits=cpu=200m,memory=256Mi
    
    log_info "暂停期间的状态（更改不会立即应用）:"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3
    
    sleep 3
    
    # 恢复部署
    log_info "恢复部署（应用所有更改）..."
    kubectl rollout resume deployment/myapp3-deployment
    
    # 监控恢复过程
    log_info "监控恢复过程..."
    kubectl rollout status deployment/myapp3-deployment --timeout=120s
    
    log_success "暂停和恢复演示完成"
    kubectl get deployment myapp3-deployment
    kubectl get pods -l app=myapp3 -o wide
    
    # 显示最终状态
    log_info "最终镜像版本:"
    kubectl get deployment myapp3-deployment -o jsonpath='{.spec.template.spec.containers[0].image}'
    echo ""
}

# 测试负载均衡
test_load_balancing() {
    log_step "测试负载均衡..."
    
    # 检查 Service 是否存在
    if ! kubectl get service deployment-nodeport-service &> /dev/null; then
        log_error "Service 'deployment-nodeport-service' 不存在，请先创建"
        return 1
    fi
    
    # 检查 Pod 是否就绪
    local ready_pods
    ready_pods=$(kubectl get pods -l app=myapp3 --field-selector=status.phase=Running --no-headers | wc -l)
    
    if [[ $ready_pods -eq 0 ]]; then
        log_error "没有就绪的 Pod"
        return 1
    fi
    
    log_info "当前就绪的 Pod 数量: $ready_pods"
    kubectl get pods -l app=myapp3 -o wide
    
    # 测试集群内部访问
    log_info "测试集群内部负载均衡..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  负载均衡测试结果${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    for i in {1..10}; do
        local response
        response=$(kubectl run test-pod-$i --image=busybox --rm -i --restart=Never --quiet -- \
            wget -qO- --timeout=5 "http://deployment-nodeport-service:80" 2>/dev/null | \
            grep -i "hostname\|server" | head -1 || echo "请求 $i: 连接失败")
        
        if [[ -n "$response" ]]; then
            echo -e "${GREEN}请求 $i:${NC} $response"
        else
            echo -e "${RED}请求 $i:${NC} 无响应"
        fi
        
        sleep 1
    done
    
    echo ""
    log_info "负载均衡测试完成"
    
    # 显示 Endpoints
    log_info "Service Endpoints:"
    kubectl get endpoints deployment-nodeport-service
}

# 监控资源状态
monitor_resources() {
    log_step "监控资源状态..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Deployment 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get deployments -l app=myapp3 -o wide
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  ReplicaSet 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get replicasets -l app=myapp3 -o wide
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get pods -l app=myapp3 -o wide
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Service 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get services -l app=myapp3
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Endpoints 状态${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get endpoints deployment-nodeport-service 2>/dev/null || log_warning "Service 不存在"
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  部署历史${NC}"
    echo -e "${CYAN}========================================${NC}"
    if kubectl get deployment myapp3-deployment &> /dev/null; then
        kubectl rollout history deployment/myapp3-deployment
    else
        log_warning "Deployment 不存在"
    fi
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  资源使用情况${NC}"
    echo -e "${CYAN}========================================${NC}"
    if command -v kubectl &> /dev/null && kubectl top pods -l app=myapp3 &> /dev/null; then
        kubectl top pods -l app=myapp3
    else
        log_warning "metrics-server 未安装或不可用"
    fi
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  最近事件${NC}"
    echo -e "${CYAN}========================================${NC}"
    kubectl get events --sort-by=.metadata.creationTimestamp --field-selector involvedObject.name=myapp3-deployment | tail -10
}

# 查看日志
view_logs() {
    log_step "查看 Pod 日志..."
    
    local pods
    pods=$(kubectl get pods -l app=myapp3 -o jsonpath='{.items[*].metadata.name}')
    
    if [[ -z "$pods" ]]; then
        log_error "没有找到相关的 Pod"
        return 1
    fi
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Pod 日志 (最近 20 行)${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    for pod in $pods; do
        echo -e "${GREEN}--- Pod: $pod ---${NC}"
        kubectl logs "$pod" --tail=20 || log_warning "无法获取 Pod $pod 的日志"
        echo ""
    done
    
    log_info "实时日志查看命令: kubectl logs -l app=myapp3 -f --tail=50"
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
    kubectl get all -l app=myapp3
    
    echo ""
    # Deployment 详细状态
    if kubectl get deployment myapp3-deployment &> /dev/null; then
        log_info "Deployment 详细状态:"
        kubectl describe deployment myapp3-deployment
    fi
    
    echo ""
    # Service 详细状态
    if kubectl get service deployment-nodeport-service &> /dev/null; then
        log_info "Service 详细状态:"
        kubectl describe service deployment-nodeport-service
    fi
    
    echo ""
    # Pod 详细状态
    log_info "Pod 详细状态:"
    local pods
    pods=$(kubectl get pods -l app=myapp3 -o jsonpath='{.items[*].metadata.name}')
    for pod in $pods; do
        echo -e "${GREEN}--- Pod: $pod ---${NC}"
        kubectl describe pod "$pod" | head -30
        echo ""
    done
    
    echo ""
    # 事件信息
    log_info "相关事件:"
    kubectl get events --sort-by=.metadata.creationTimestamp | grep -E "myapp3|deployment" | tail -20
}

# 最佳实践演示
show_best_practices() {
    log_step "最佳实践演示..."
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Kubernetes Deployment 最佳实践${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    echo -e "${GREEN}1. Deployment 配置最佳实践:${NC}"
    echo "   • 设置资源请求和限制"
    echo "   • 配置健康检查探针"
    echo "   • 使用有意义的标签和注解"
    echo "   • 设置合适的更新策略"
    echo "   • 配置安全上下文"
    
    echo ""
    echo -e "${GREEN}2. 滚动更新最佳实践:${NC}"
    echo "   • 使用具体的镜像标签而非 latest"
    echo "   • 设置合适的 maxUnavailable 和 maxSurge"
    echo "   • 配置就绪性探针确保新 Pod 就绪"
    echo "   • 监控更新过程并准备回滚"
    echo "   • 在非生产环境先验证更新"
    
    echo ""
    echo -e "${GREEN}3. 扩缩容最佳实践:${NC}"
    echo "   • 根据实际负载设置副本数"
    echo "   • 使用 HPA 实现自动扩缩容"
    echo "   • 考虑节点资源容量"
    echo "   • 设置 PodDisruptionBudget"
    echo "   • 监控资源使用情况"
    
    echo ""
    echo -e "${GREEN}4. 安全最佳实践:${NC}"
    echo "   • 使用非 root 用户运行容器"
    echo "   • 设置只读根文件系统"
    echo "   • 禁用特权提升"
    echo "   • 使用最小权限原则"
    echo "   • 定期更新镜像版本"
    
    echo ""
    echo -e "${GREEN}5. 监控和运维最佳实践:${NC}"
    echo "   • 实施全面的监控和告警"
    echo "   • 收集和分析应用日志"
    echo "   • 定期备份配置文件"
    echo "   • 使用版本控制管理 YAML 文件"
    echo "   • 建立完善的发布流程"
    
    echo ""
    echo -e "${GREEN}6. 网络最佳实践:${NC}"
    echo "   • 使用 Service 提供稳定的网络入口"
    echo "   • 生产环境使用 LoadBalancer 或 Ingress"
    echo "   • 配置适当的网络策略"
    echo "   • 考虑服务网格（如 Istio）"
    echo "   • 实施 TLS 加密"
    
    # 显示改进的配置示例
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  生产级 Deployment 配置示例${NC}"
    echo -e "${CYAN}========================================${NC}"
    
    cat << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: production-app-deployment
  labels:
    app: production-app
    version: v1.0
    tier: frontend
    environment: production
  annotations:
    description: "Production web application deployment"
    maintainer: "devops-team@company.com"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: production-app
      version: v1.0
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  revisionHistoryLimit: 10
  progressDeadlineSeconds: 600
  template:
    metadata:
      labels:
        app: production-app
        version: v1.0
        tier: frontend
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 2000
      containers:
      - name: app-container
        image: myapp:1.0.0
        imagePullPolicy: IfNotPresent
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        ports:
        - name: http
          containerPort: 8080
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
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
EOF
}

# 清理资源
cleanup_resources() {
    log_step "清理资源..."
    
    # 删除 Service
    if kubectl get service deployment-nodeport-service &> /dev/null; then
        log_info "删除 Service..."
        kubectl delete service deployment-nodeport-service
        log_success "Service 已删除"
    fi
    
    # 删除 Deployment（会同时删除 ReplicaSet 和 Pod）
    if kubectl get deployment myapp3-deployment &> /dev/null; then
        log_info "删除 Deployment..."
        kubectl delete deployment myapp3-deployment
        log_success "Deployment 已删除"
    fi
    
    # 清理测试 Pod（如果存在）
    kubectl delete pod --selector=run=test-pod --ignore-not-found=true &> /dev/null
    
    # 等待资源完全删除
    log_info "等待资源完全删除..."
    sleep 5
    
    # 验证清理结果
    local remaining_resources
    remaining_resources=$(kubectl get deploy,rs,svc,pods -l app=myapp3 --no-headers 2>/dev/null | wc -l)
    
    if [[ $remaining_resources -eq 0 ]]; then
        log_success "所有资源已成功清理"
    else
        log_warning "仍有 $remaining_resources 个相关资源存在"
        kubectl get deploy,rs,svc,pods -l app=myapp3
    fi
}

# 交互式菜单
interactive_menu() {
    while true; do
        echo ""
        echo -e "${CYAN}========================================${NC}"
        echo -e "${CYAN}  Deployment YAML Demo 菜单${NC}"
        echo -e "${CYAN}========================================${NC}"
        echo "1. 验证 YAML 文件"
        echo "2. 创建 Deployment"
        echo "3. 创建 Service"
        echo "4. 创建所有资源"
        echo "5. 演示滚动更新"
        echo "6. 演示回滚操作"
        echo "7. 演示扩缩容操作"
        echo "8. 演示暂停和恢复部署"
        echo "9. 测试负载均衡"
        echo "10. 监控资源状态"
        echo "11. 查看日志"
        echo "12. 调试模式"
        echo "13. 最佳实践演示"
        echo "14. 清理资源"
        echo "0. 退出"
        echo ""
        read -p "请选择操作 (0-14): " choice
        
        case $choice in
            1) validate_yaml ;;
            2) create_deployment ;;
            3) create_service ;;
            4) create_deployment && create_service ;;
            5) demonstrate_rolling_update ;;
            6) demonstrate_rollback ;;
            7) demonstrate_scaling ;;
            8) demonstrate_pause_resume ;;
            9) test_load_balancing ;;
            10) monitor_resources ;;
            11) view_logs ;;
            12) debug_mode ;;
            13) show_best_practices ;;
            14) cleanup_resources ;;
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
            create_deployment
            create_service
            ;;
        --create-deploy)
            validate_yaml
            create_deployment
            ;;
        --create-service)
            validate_yaml
            create_service
            ;;
        --rolling-update)
            demonstrate_rolling_update
            ;;
        --rollback)
            demonstrate_rollback
            ;;
        --scale)
            demonstrate_scaling
            ;;
        --pause-resume)
            demonstrate_pause_resume
            ;;
        --test-lb)
            test_load_balancing
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