#!/bin/bash

# build.sh
# Kubernetes 调度器高级实践工具构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
REGISTRY=${REGISTRY:-"localhost:5000"}
TAG=${TAG:-"latest"}
PUSH=${PUSH:-"false"}
DEPLOY=${DEPLOY:-"false"}

# 工具列表
TOOLS=(
    "tenant-resource-manager"
    "scheduler-audit-analyzer"
    "scheduler-visualizer"
    "heatmap-generator"
    "performance-analyzer"
)

# 函数定义
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

# 显示帮助信息
show_help() {
    cat << EOF
Kubernetes 调度器高级实践工具构建脚本

用法: $0 [选项] [工具名称...]

选项:
  -h, --help          显示此帮助信息
  -r, --registry      设置镜像仓库地址 (默认: localhost:5000)
  -t, --tag           设置镜像标签 (默认: latest)
  -p, --push          构建后推送镜像到仓库
  -d, --deploy        构建并部署到 Kubernetes
  -a, --all           构建所有工具
  --clean             清理构建缓存
  --test              运行测试

工具列表:
$(printf "  %s\n" "${TOOLS[@]}")

示例:
  $0 --all                                    # 构建所有工具
  $0 performance-analyzer                     # 构建性能分析器
  $0 -r docker.io/myorg -t v1.0.0 --push --all  # 构建、标记并推送所有工具
  $0 --deploy performance-analyzer            # 构建并部署性能分析器

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 kubectl (如果需要部署)
    if [[ "$DEPLOY" == "true" ]] && ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在 PATH 中，无法部署"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 清理函数
clean_build() {
    log_info "清理构建缓存..."
    
    # 清理 Go 缓存
    go clean -cache -modcache -testcache
    
    # 清理 Docker 构建缓存
    docker builder prune -f
    
    # 清理悬挂镜像
    docker image prune -f
    
    log_success "清理完成"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
    # 运行 Go 测试
    if go test -v ./...; then
        log_success "所有测试通过"
    else
        log_error "测试失败"
        exit 1
    fi
}

# 构建单个工具
build_tool() {
    local tool=$1
    local image_name="${REGISTRY}/${tool}:${TAG}"
    
    log_info "构建 ${tool}..."
    
    # 检查源文件是否存在
    if [[ ! -f "cmd/${tool}/main.go" ]]; then
        log_error "源文件 cmd/${tool}/main.go 不存在"
        return 1
    fi
    
    # 构建 Docker 镜像
    if docker build \
        --network=host \
        --build-arg TOOL_NAME="${tool}" \
        --tag "${image_name}" \
        --label "build.timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --label "build.version=${TAG}" \
        --label "build.tool=${tool}" \
        .; then
        log_success "${tool} 构建成功: ${image_name}"
    else
        log_error "${tool} 构建失败"
        return 1
    fi
    
    # 推送镜像
    if [[ "$PUSH" == "true" ]]; then
        log_info "推送 ${image_name}..."
        if docker push "${image_name}"; then
            log_success "${tool} 推送成功"
        else
            log_error "${tool} 推送失败"
            return 1
        fi
    fi
    
    # 部署到 Kubernetes
    if [[ "$DEPLOY" == "true" ]]; then
        deploy_tool "${tool}" "${image_name}"
    fi
}

# 部署工具到 Kubernetes
deploy_tool() {
    local tool=$1
    local image_name=$2
    local deployment_file="${tool}-deployment.yaml"
    
    log_info "部署 ${tool} 到 Kubernetes..."
    
    # 检查部署文件是否存在
    if [[ ! -f "${deployment_file}" ]]; then
        log_warning "部署文件 ${deployment_file} 不存在，跳过部署"
        return 0
    fi
    
    # 更新镜像名称
    local temp_file=$(mktemp)
    sed "s|image: ${tool}:latest|image: ${image_name}|g" "${deployment_file}" > "${temp_file}"
    
    # 应用部署
    if kubectl apply -f "${temp_file}"; then
        log_success "${tool} 部署成功"
        
        # 等待部署就绪
        log_info "等待 ${tool} 部署就绪..."
        kubectl rollout status deployment/${tool} -n kube-system --timeout=300s
        
        # 显示服务信息
        log_info "${tool} 服务信息:"
        kubectl get svc ${tool} -n kube-system
    else
        log_error "${tool} 部署失败"
        rm -f "${temp_file}"
        return 1
    fi
    
    rm -f "${temp_file}"
}

# 验证工具名称
validate_tool() {
    local tool=$1
    for valid_tool in "${TOOLS[@]}"; do
        if [[ "$tool" == "$valid_tool" ]]; then
            return 0
        fi
    done
    return 1
}

# 主函数
main() {
    local build_all=false
    local tools_to_build=()
    local clean=false
    local test=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -t|--tag)
                TAG="$2"
                shift 2
                ;;
            -p|--push)
                PUSH="true"
                shift
                ;;
            -d|--deploy)
                DEPLOY="true"
                shift
                ;;
            -a|--all)
                build_all=true
                shift
                ;;
            --clean)
                clean=true
                shift
                ;;
            --test)
                test=true
                shift
                ;;
            -*)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
            *)
                if validate_tool "$1"; then
                    tools_to_build+=("$1")
                else
                    log_error "未知工具: $1"
                    log_info "可用工具: ${TOOLS[*]}"
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    # 显示配置
    log_info "构建配置:"
    echo "  Registry: ${REGISTRY}"
    echo "  Tag: ${TAG}"
    echo "  Push: ${PUSH}"
    echo "  Deploy: ${DEPLOY}"
    echo ""
    
    # 检查依赖
    check_dependencies
    
    # 清理
    if [[ "$clean" == "true" ]]; then
        clean_build
    fi
    
    # 运行测试
    if [[ "$test" == "true" ]]; then
        run_tests
    fi
    
    # 确定要构建的工具
    if [[ "$build_all" == "true" ]]; then
        tools_to_build=("${TOOLS[@]}")
    elif [[ ${#tools_to_build[@]} -eq 0 ]]; then
        log_error "请指定要构建的工具或使用 --all 构建所有工具"
        show_help
        exit 1
    fi
    
    # 构建工具
    local failed_builds=()
    for tool in "${tools_to_build[@]}"; do
        if ! build_tool "$tool"; then
            failed_builds+=("$tool")
        fi
    done
    
    # 显示结果
    echo ""
    log_info "构建摘要:"
    echo "  成功: $((${#tools_to_build[@]} - ${#failed_builds[@]}))"
    echo "  失败: ${#failed_builds[@]}"
    
    if [[ ${#failed_builds[@]} -gt 0 ]]; then
        log_error "失败的构建: ${failed_builds[*]}"
        exit 1
    else
        log_success "所有构建成功完成!"
    fi
    
    # 显示后续步骤
    if [[ "$DEPLOY" == "true" ]]; then
        echo ""
        log_info "部署完成! 使用以下命令访问服务:"
        for tool in "${tools_to_build[@]}"; do
            echo "  kubectl port-forward -n kube-system svc/${tool} 8080:80"
        done
    fi
}

# 执行主函数
main "$@"