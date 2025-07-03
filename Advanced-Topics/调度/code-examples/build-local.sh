#!/bin/bash

# build-local.sh
# 本地构建脚本 - 直接使用 Go 构建二进制文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 工具列表
TOOLS=(
    "tenant-resource-manager"
    "scheduler-audit-analyzer"
    "scheduler-visualizer"
    "heatmap-generator"
    "performance-analyzer"
    "scheduler-analyzer"
)

# 函数定义
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
Kubernetes 调度器高级实践工具本地构建脚本

用法: $0 [选项] [工具名称...]

选项:
  -h, --help          显示此帮助信息
  -a, --all           构建所有工具
  --clean             清理构建产物
  --test              运行测试

工具列表:
$(printf "  %s\n" "${TOOLS[@]}")

示例:
  $0 --all                        # 构建所有工具
  $0 performance-analyzer         # 构建性能分析器
  $0 --clean                      # 清理构建产物

EOF
}

# 清理函数
clean_build() {
    log_info "清理构建产物..."
    
    # 清理 Go 缓存
    go clean -cache -modcache -testcache
    
    # 清理 bin 目录
    rm -rf bin/*
    
    log_success "清理完成"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
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
    
    log_info "构建 ${tool}..."
    
    # 检查源文件是否存在
    if [[ ! -f "cmd/${tool}/main.go" ]]; then
        log_error "源文件 cmd/${tool}/main.go 不存在"
        return 1
    fi
    
    # 创建 bin 目录
    mkdir -p bin
    
    # 构建二进制文件
    if go build -o "bin/${tool}" "./cmd/${tool}"; then
        log_success "${tool} 构建成功: bin/${tool}"
        return 0
    else
        log_error "${tool} 构建失败"
        return 1
    fi
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
    
    # 清理
    if [[ "$clean" == "true" ]]; then
        clean_build
        exit 0
    fi
    
    # 运行测试
    if [[ "$test" == "true" ]]; then
        run_tests
        exit 0
    fi
    
    # 确定要构建的工具
    if [[ "$build_all" == "true" ]]; then
        tools_to_build=("${TOOLS[@]}")
    elif [[ ${#tools_to_build[@]} -eq 0 ]]; then
        log_error "请指定要构建的工具或使用 --all 构建所有工具"
        show_help
        exit 1
    fi
    
    # 检查 Go 环境
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    log_info "使用 Go 版本: $(go version)"
    
    # 构建工具
    local failed_builds=()
    local successful_builds=()
    
    for tool in "${tools_to_build[@]}"; do
        if build_tool "$tool"; then
            successful_builds+=("$tool")
        else
            failed_builds+=("$tool")
        fi
    done
    
    # 显示构建摘要
    echo ""
    log_info "构建摘要:"
    echo "  成功: ${#successful_builds[@]}"
    echo "  失败: ${#failed_builds[@]}"
    
    if [[ ${#successful_builds[@]} -gt 0 ]]; then
        log_success "成功构建: ${successful_builds[*]}"
    fi
    
    if [[ ${#failed_builds[@]} -gt 0 ]]; then
        log_error "失败的构建: ${failed_builds[*]}"
        exit 1
    fi
    
    log_success "所有工具构建完成！"
}

# 运行主函数
main "$@"