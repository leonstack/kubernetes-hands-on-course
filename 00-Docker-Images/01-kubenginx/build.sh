#!/bin/bash

# 用于构建所有版本的nginx镜像
# 作者: Grissom
# 版本: v1.0.0
# 日期: 2025-06-20

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
IMAGE_NAME="grissomsh/kubenginx"
VERSIONS=("1.0.0" "2.0.0" "3.0.0" "4.0.0")
REGISTRY=""  # 可以设置为你的Docker Registry地址

# 函数：打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 函数：构建单个版本
build_version() {
    local version=$1
    local version_lower=$(echo $version | tr '[:upper:]' '[:lower:]')
    local tag="${IMAGE_NAME}:${version_lower}"
    
    print_message $BLUE "🔨 构建 ${version} 版本..."
    
    cd "${version}-Release"
    
    # 构建镜像
    if docker build -t "$tag" .; then
        print_message $GREEN "✅ ${version} 版本构建成功: $tag"
    else
        print_message $RED "❌ ${version} 版本构建失败"
        cd ..
        return 1
    fi
    
    cd ..
    return 0
}

# 函数：推送镜像到Registry
push_image() {
    local version=$1
    local version_lower=$(echo $version | tr '[:upper:]' '[:lower:]')
    local tag="${IMAGE_NAME}:${version_lower}"
    
    if [ -n "$REGISTRY" ]; then
        local registry_tag="${REGISTRY}/${tag}"
        print_message $BLUE "📤 推送 ${version} 到 Registry..."
        
        docker tag "$tag" "$registry_tag"
        if docker push "$registry_tag"; then
            print_message $GREEN "✅ ${version} 推送成功: $registry_tag"
        else
            print_message $RED "❌ ${version} 推送失败"
            return 1
        fi
    fi
}

# 函数：显示帮助信息
show_help() {
    echo "Grissom's Kubernetes Demo - Docker Build Script"
    echo ""
    echo "用法: $0 [选项] [版本]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -a, --all      构建所有版本"
    echo "  -p, --push     构建后推送到Registry"
    echo "  -c, --clean    清理本地镜像"
    echo "  -l, --list     列出所有镜像"
    echo ""
    echo "版本: V1, V2, V3, V4"
    echo ""
    echo "示例:"
    echo "  $0 -a          # 构建所有版本"
    echo "  $0 V1          # 只构建V1版本"
    echo "  $0 -a -p       # 构建所有版本并推送"
    echo "  $0 -c          # 清理所有镜像"
}

# 函数：清理镜像
clean_images() {
    print_message $YELLOW "🧹 清理本地镜像..."
    
    for version in "${VERSIONS[@]}"; do
        local version_lower=$(echo $version | tr '[:upper:]' '[:lower:]')
        local tag="${IMAGE_NAME}:${version_lower}"
        
        if docker images -q "$tag" > /dev/null; then
            docker rmi "$tag" || true
            print_message $GREEN "✅ 已删除: $tag"
        fi
    done
}

# 函数：列出镜像
list_images() {
    print_message $BLUE "📋 本地镜像列表:"
    docker images | grep "$IMAGE_NAME" || print_message $YELLOW "未找到相关镜像"
}

# 主函数
main() {
    local build_all=false
    local push_images=false
    local clean=false
    local list=false
    local specific_version=""
    
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
            -p|--push)
                push_images=true
                shift
                ;;
            -c|--clean)
                clean=true
                shift
                ;;
            -l|--list)
                list=true
                shift
                ;;
            V1|V2|V3|V4)
                specific_version=$1
                shift
                ;;
            *)
                print_message $RED "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 执行操作
    if [ "$clean" = true ]; then
        clean_images
        exit 0
    fi
    
    if [ "$list" = true ]; then
        list_images
        exit 0
    fi
    
    # 检查Docker是否运行
    if ! docker info > /dev/null 2>&1; then
        print_message $RED "❌ Docker未运行，请先启动Docker"
        exit 1
    fi
    
    print_message $GREEN "🚀 开始构建 Grissom's Kubernetes Demo 镜像"
    
    # 构建镜像
    if [ "$build_all" = true ]; then
        for version in "${VERSIONS[@]}"; do
            if build_version "$version"; then
                if [ "$push_images" = true ]; then
                    push_image "$version"
                fi
            fi
        done
    elif [ -n "$specific_version" ]; then
        if build_version "$specific_version"; then
            if [ "$push_images" = true ]; then
                push_image "$specific_version"
            fi
        fi
    else
        print_message $YELLOW "请指定要构建的版本或使用 -a 构建所有版本"
        show_help
        exit 1
    fi
    
    print_message $GREEN "🎉 构建完成！"
    list_images
}

# 运行主函数
main "$@"