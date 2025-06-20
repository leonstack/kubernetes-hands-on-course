#!/bin/bash

# 演示 YAML 语法验证、Kubernetes 资源创建和最佳实践
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

# 脚本信息
SCRIPT_NAME="YAML 基础学习演示"
SCRIPT_VERSION="1.0.0"
SCRIPT_AUTHOR="Grissom"

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

# 显示脚本信息
show_script_info() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    $SCRIPT_NAME                    ║"
    echo "║                                                              ║"
    echo "║  版本: $SCRIPT_VERSION                                           ║"
    echo "║  作者: $SCRIPT_AUTHOR                                    ║"
    echo "║                                                              ║"
    echo "║  功能:                                                       ║"
    echo "║  • YAML 语法验证和示例                                      ║"
    echo "║  • Kubernetes 资源创建演示                                  ║"
    echo "║  • 最佳实践和常见错误演示                                   ║"
    echo "║  • 交互式学习体验                                           ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}\n"
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}使用方法:${NC}"
    echo "  $0 [选项]"
    echo ""
    echo -e "${CYAN}选项:${NC}"
    echo "  -h, --help     显示此帮助信息"
    echo "  -v, --version  显示版本信息"
    echo "  -i, --info     显示脚本信息"
    echo "  --validate     仅执行 YAML 验证"
    echo "  --k8s          仅执行 Kubernetes 演示"
    echo "  --examples     仅显示示例"
    echo "  --cleanup      清理演示资源"
    echo ""
    echo -e "${CYAN}示例:${NC}"
    echo "  $0              # 运行完整演示"
    echo "  $0 --validate   # 仅验证 YAML 语法"
    echo "  $0 --k8s        # 仅演示 Kubernetes 资源"
    echo "  $0 --cleanup    # 清理演示资源"
}

# 检查依赖
check_dependencies() {
    log_step "检查依赖工具"
    
    local missing_deps=()
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    else
        log_success "kubectl 已安装: $(kubectl version --client --short 2>/dev/null || echo '版本信息获取失败')"
    fi
    
    # 检查 Python (用于 YAML 验证)
    if ! command -v python3 &> /dev/null; then
        missing_deps+=("python3")
    else
        log_success "Python3 已安装: $(python3 --version)"
    fi
    
    # 检查 yamllint (可选)
    if command -v yamllint &> /dev/null; then
        log_success "yamllint 已安装: $(yamllint --version)"
    else
        log_warning "yamllint 未安装 (可选工具)"
        echo "  安装命令: pip install yamllint"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少必需的依赖工具: ${missing_deps[*]}"
        echo "请安装缺少的工具后重新运行脚本。"
        exit 1
    fi
    
    log_success "所有必需的依赖工具都已安装"
}

# 验证 YAML 语法
validate_yaml() {
    log_step "YAML 语法验证演示"
    
    # 创建测试 YAML 文件
    local test_files=("good-example.yaml" "bad-example.yaml")
    
    # 创建正确的 YAML 示例
    cat > good-example.yaml << 'EOF'
# 正确的 YAML 示例
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  labels:
    app: test
    version: v1.0
spec:
  containers:
    - name: test-container
      image: nginx:1.21
      ports:
        - containerPort: 80
          protocol: TCP
      env:
        - name: ENV_VAR
          value: "test-value"
EOF
    
    # 创建错误的 YAML 示例
    cat > bad-example.yaml << 'EOF'
# 错误的 YAML 示例 (缩进错误)
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
    labels:              # 缩进错误
  app: test             # 缩进错误
spec:
  containers:
    - name: test-container
      image: nginx:1.21
      ports:
        - containerPort: 80
          protocol: TCP
EOF
    
    log_info "验证正确的 YAML 文件..."
    if python3 -c "import yaml; yaml.safe_load(open('good-example.yaml'))" 2>/dev/null; then
        log_success "good-example.yaml 语法正确"
    else
        log_error "good-example.yaml 语法错误"
    fi
    
    log_info "验证错误的 YAML 文件..."
    if python3 -c "import yaml; yaml.safe_load(open('bad-example.yaml'))" 2>/dev/null; then
        log_warning "bad-example.yaml 语法正确 (意外)"
    else
        log_error "bad-example.yaml 语法错误 (预期结果)"
        echo "  这演示了缩进错误会导致 YAML 解析失败"
    fi
    
    # 使用 kubectl 验证 Kubernetes YAML
    log_info "使用 kubectl 验证 Kubernetes YAML..."
    if kubectl apply --dry-run=client -f good-example.yaml &>/dev/null; then
        log_success "good-example.yaml 通过 kubectl 验证"
    else
        log_error "good-example.yaml 未通过 kubectl 验证"
    fi
    
    # 使用 yamllint (如果可用)
    if command -v yamllint &> /dev/null; then
        log_info "使用 yamllint 检查代码风格..."
        echo "good-example.yaml 检查结果:"
        yamllint good-example.yaml || true
        echo "\nbad-example.yaml 检查结果:"
        yamllint bad-example.yaml || true
    fi
    
    # 清理测试文件
    rm -f good-example.yaml bad-example.yaml
}

# 演示 YAML 数据结构
demonstrate_yaml_structures() {
    log_step "YAML 数据结构演示"
    
    log_info "1. 标量 (Scalars)"
    cat << 'EOF'
# 字符串
name: "John Doe"
city: Beijing

# 数字
age: 30
salary: 50000.50

# 布尔值
is_student: true
is_married: false

# 空值
spouse: null
children: ~
EOF
    
    log_info "\n2. 序列 (Sequences/Lists)"
    cat << 'EOF'
# 简单列表
fruits:
  - apple
  - banana
  - orange

# 内联列表
colors: [red, green, blue]

# 复杂列表
people:
  - name: Alice
    age: 25
  - name: Bob
    age: 30
EOF
    
    log_info "\n3. 映射 (Mappings/Dictionaries)"
    cat << 'EOF'
# 简单映射
person:
  name: Alice
  age: 25
  city: Shanghai

# 嵌套映射
company:
  name: TechCorp
  address:
    street: "123 Tech Street"
    city: "Beijing"
    country: "China"
  employees:
    - name: Alice
      department: Engineering
    - name: Bob
      department: Marketing
EOF
    
    log_info "\n4. 多行字符串"
    cat << 'EOF'
# 保留换行符 (|)
description: |
  这是第一行
  这是第二行
  这是第三行

# 折叠换行符 (>)
summary: >
  这是一个很长的句子，
  会被折叠成一行，
  换行符变成空格。
EOF
}

# Kubernetes 资源演示
demonstrate_kubernetes() {
    log_step "Kubernetes 资源演示"
    
    # 检查 Kubernetes 连接
    if ! kubectl cluster-info &>/dev/null; then
        log_warning "无法连接到 Kubernetes 集群，跳过实际部署演示"
        log_info "将显示 YAML 配置示例而不实际部署"
        show_k8s_examples
        return
    fi
    
    log_success "已连接到 Kubernetes 集群"
    kubectl cluster-info
    
    # 创建演示命名空间
    log_info "创建演示命名空间..."
    kubectl create namespace yaml-demo --dry-run=client -o yaml | kubectl apply -f -
    
    # 创建 Pod
    log_info "创建演示 Pod..."
    cat << 'EOF' | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: yaml-demo-pod
  namespace: yaml-demo
  labels:
    app: yaml-demo
    component: web
spec:
  containers:
    - name: nginx
      image: nginx:1.21
      ports:
        - containerPort: 80
          name: http
      env:
        - name: DEMO_ENV
          value: "yaml-tutorial"
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"
EOF
    
    # 创建 Service
    log_info "创建演示 Service..."
    cat << 'EOF' | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: yaml-demo-service
  namespace: yaml-demo
  labels:
    app: yaml-demo
spec:
  type: ClusterIP
  selector:
    app: yaml-demo
  ports:
    - name: http
      port: 80
      targetPort: 80
      protocol: TCP
EOF
    
    # 等待 Pod 就绪
    log_info "等待 Pod 就绪..."
    kubectl wait --for=condition=Ready pod/yaml-demo-pod -n yaml-demo --timeout=60s
    
    # 显示资源状态
    log_success "演示资源创建完成！"
    echo "\nPod 状态:"
    kubectl get pods -n yaml-demo -o wide
    echo "\nService 状态:"
    kubectl get services -n yaml-demo
    
    # 显示 Pod 详细信息
    echo "\nPod 详细信息:"
    kubectl describe pod yaml-demo-pod -n yaml-demo
    
    log_info "演示完成。使用 '$0 --cleanup' 清理资源。"
}

# 显示 Kubernetes 示例
show_k8s_examples() {
    log_info "Kubernetes YAML 配置示例:"
    
    echo "\n1. Pod 配置:"
    cat << 'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: example-pod
  labels:
    app: example
spec:
  containers:
    - name: web
      image: nginx:1.21
      ports:
        - containerPort: 80
EOF
    
    echo "\n2. Service 配置:"
    cat << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: example-service
spec:
  selector:
    app: example
  ports:
    - port: 80
      targetPort: 80
EOF
    
    echo "\n3. Deployment 配置:"
    cat << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: web
          image: nginx:1.21
          ports:
            - containerPort: 80
EOF
}

# 演示最佳实践
demonstrate_best_practices() {
    log_step "YAML 最佳实践演示"
    
    log_info "1. 缩进规范 (使用 2 个空格)"
    cat << 'EOF'
# ✅ 正确
metadata:
  name: good-example
  labels:
    app: myapp

# ❌ 错误 (使用 Tab 或不一致缩进)
metadata:
	name: bad-example
  labels:
      app: myapp
EOF
    
    log_info "\n2. 引号使用规范"
    cat << 'EOF'
# 简单字符串可以不用引号
name: myapp

# 包含特殊字符时使用引号
description: "包含冒号: 和空格的字符串"

# 数字字符串需要引号
version: "1.0"
port: "8080"
EOF
    
    log_info "\n3. 注释规范"
    cat << 'EOF'
# 文件头部注释
# 作者: DevOps Team
# 用途: 生产环境配置

apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod          # Pod 名称
  labels:
    app: myapp             # 应用标识
    version: v1.0          # 版本号
EOF
    
    log_info "\n4. 安全最佳实践"
    cat << 'EOF'
# ❌ 不要硬编码敏感信息
env:
  - name: DB_PASSWORD
    value: "hardcoded-password"

# ✅ 使用 Secret 引用
env:
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-secret
        key: password
EOF
}

# 演示常见错误
demonstrate_common_errors() {
    log_step "常见错误演示"
    
    log_error "1. 缩进错误"
    cat << 'EOF'
# 错误示例
metadata:
  name: test
    labels:        # 缩进过多
app: test         # 缩进不足
EOF
    
    log_error "\n2. 冒号后缺少空格"
    cat << 'EOF'
# 错误示例
name:test         # 缺少空格
app:myapp         # 缺少空格
EOF
    
    log_error "\n3. 列表格式错误"
    cat << 'EOF'
# 错误示例
containers:
- name: app1      # 缺少空格
 -name: app2      # 位置错误
EOF
    
    log_error "\n4. 引号使用错误"
    cat << 'EOF'
# 错误示例
version: 1.0      # 数字会被解析为浮点数
port: 8080        # 数字会被解析为整数

# 正确示例
version: "1.0"    # 字符串
port: "8080"      # 字符串
EOF
}

# 清理演示资源
cleanup_demo() {
    log_step "清理演示资源"
    
    if kubectl get namespace yaml-demo &>/dev/null; then
        log_info "删除演示命名空间和所有资源..."
        kubectl delete namespace yaml-demo
        log_success "演示资源已清理完成"
    else
        log_info "没有找到演示资源，无需清理"
    fi
}

# 交互式菜单
show_interactive_menu() {
    while true; do
        echo -e "\n${CYAN}=== YAML 基础学习菜单 ===${NC}"
        echo "1. YAML 语法验证演示"
        echo "2. YAML 数据结构演示"
        echo "3. Kubernetes 资源演示"
        echo "4. 最佳实践演示"
        echo "5. 常见错误演示"
        echo "6. 查看示例文件"
        echo "7. 清理演示资源"
        echo "0. 退出"
        echo -e "${CYAN}请选择 (0-7): ${NC}"
        
        read -r choice
        
        case $choice in
            1) validate_yaml ;;
            2) demonstrate_yaml_structures ;;
            3) demonstrate_kubernetes ;;
            4) demonstrate_best_practices ;;
            5) demonstrate_common_errors ;;
            6) 
                log_info "查看示例文件内容:"
                if [ -f "sample-file.yml" ]; then
                    echo "\n=== sample-file.yml ==="
                    cat sample-file.yml
                else
                    log_warning "sample-file.yml 文件不存在"
                fi
                ;;
            7) cleanup_demo ;;
            0) 
                log_info "感谢使用 YAML 基础学习演示脚本！"
                exit 0
                ;;
            *) 
                log_error "无效选择，请输入 0-7"
                ;;
        esac
        
        echo -e "\n${YELLOW}按 Enter 键继续...${NC}"
        read -r
    done
}

# 主函数
main() {
    # 解析命令行参数
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            echo "$SCRIPT_NAME v$SCRIPT_VERSION"
            exit 0
            ;;
        -i|--info)
            show_script_info
            exit 0
            ;;
        --validate)
            show_script_info
            check_dependencies
            validate_yaml
            exit 0
            ;;
        --k8s)
            show_script_info
            check_dependencies
            demonstrate_kubernetes
            exit 0
            ;;
        --examples)
            demonstrate_yaml_structures
            demonstrate_best_practices
            demonstrate_common_errors
            exit 0
            ;;
        --cleanup)
            cleanup_demo
            exit 0
            ;;
        "")
            # 无参数，运行交互式菜单
            show_script_info
            check_dependencies
            show_interactive_menu
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi