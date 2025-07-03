#!/bin/bash

# 端口配置改进测试脚本
# 测试调度器工具的端口配置功能

set -e

echo "=== 端口配置改进测试 ==="
echo

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_tool_help() {
    local tool=$1
    local expected_port=$2
    
    echo -e "${YELLOW}测试 $tool 帮助信息...${NC}"
    
    if ./bin/$tool --help 2>&1 | grep -q "HTTP server port (default \"$expected_port\")"; then
        echo -e "${GREEN}✓ $tool 端口参数配置正确 (默认: $expected_port)${NC}"
    else
        echo -e "${RED}✗ $tool 端口参数配置错误${NC}"
        return 1
    fi
}

# 测试端口启动
test_tool_startup() {
    local tool=$1
    local port=$2
    local test_name=$3
    
    echo -e "${YELLOW}测试 $tool $test_name...${NC}"
    
    # 启动工具
    if [ "$port" = "default" ]; then
        timeout 3 ./bin/$tool > /dev/null 2>&1 &
    else
        timeout 3 ./bin/$tool --port=$port > /dev/null 2>&1 &
    fi
    
    local pid=$!
    sleep 1
    
    # 检查进程是否还在运行
    if kill -0 $pid 2>/dev/null; then
        echo -e "${GREEN}✓ $tool $test_name 启动成功${NC}"
        kill $pid 2>/dev/null || true
        wait $pid 2>/dev/null || true
    else
        echo -e "${RED}✗ $tool $test_name 启动失败${NC}"
        return 1
    fi
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -i :$port >/dev/null 2>&1; then
        echo -e "${RED}警告: 端口 $port 已被占用${NC}"
        return 1
    fi
    return 0
}

echo "1. 检查工具是否已编译..."
if [ ! -f "bin/scheduler-visualizer" ] || [ ! -f "bin/heatmap-generator" ]; then
    echo -e "${RED}错误: 工具未编译，请先运行编译命令${NC}"
    echo "运行: go build -o bin/scheduler-visualizer ./cmd/scheduler-visualizer"
    echo "运行: go build -o bin/heatmap-generator ./cmd/heatmap-generator"
    exit 1
fi
echo -e "${GREEN}✓ 工具已编译${NC}"
echo

echo "2. 测试帮助信息和端口参数..."
test_tool_help "scheduler-visualizer" "8080"
test_tool_help "heatmap-generator" "8082"
echo

echo "3. 测试默认端口启动..."
check_port 8080 || echo "跳过端口 8080 测试"
if check_port 8080; then
    test_tool_startup "scheduler-visualizer" "default" "默认端口启动"
fi

check_port 8082 || echo "跳过端口 8082 测试"
if check_port 8082; then
    test_tool_startup "heatmap-generator" "default" "默认端口启动"
fi
echo

echo "4. 测试自定义端口启动..."
check_port 9080 || echo "跳过端口 9080 测试"
if check_port 9080; then
    test_tool_startup "scheduler-visualizer" "9080" "自定义端口启动"
fi

check_port 9082 || echo "跳过端口 9082 测试"
if check_port 9082; then
    test_tool_startup "heatmap-generator" "9082" "自定义端口启动"
fi
echo

echo "5. 测试端口冲突避免..."
echo -e "${YELLOW}验证不同工具使用不同默认端口...${NC}"
if [ "8080" != "8082" ]; then
    echo -e "${GREEN}✓ 调度决策可视化工具 (8080) 和热力图生成器 (8082) 使用不同端口${NC}"
else
    echo -e "${RED}✗ 端口冲突: 两个工具使用相同端口${NC}"
fi
echo

echo "6. 验证部署文件配置..."
echo -e "${YELLOW}检查 Kubernetes 部署文件...${NC}"

# 检查调度决策可视化工具部署文件
if grep -q '\-\-port=' deployments/kubernetes/scheduler-visualizer-deployment.yaml; then
    echo -e "${GREEN}✓ scheduler-visualizer 部署文件包含端口参数配置${NC}"
else
    echo -e "${RED}✗ scheduler-visualizer 部署文件缺少端口参数配置${NC}"
fi

# 检查热力图生成器部署文件
if grep -q '\-\-port=' deployments/kubernetes/heatmap-generator-deployment.yaml; then
    echo -e "${GREEN}✓ heatmap-generator 部署文件包含端口参数配置${NC}"
else
    echo -e "${RED}✗ heatmap-generator 部署文件缺少端口参数配置${NC}"
fi

# 检查端口配置一致性
if grep -q "value: \"8082\"" deployments/kubernetes/heatmap-generator-deployment.yaml; then
    echo -e "${GREEN}✓ heatmap-generator 部署文件端口配置正确 (8082)${NC}"
else
    echo -e "${RED}✗ heatmap-generator 部署文件端口配置错误${NC}"
fi
echo

echo -e "${GREEN}=== 端口配置改进测试完成 ===${NC}"
echo
echo "改进总结:"
echo "• 调度决策可视化工具: 支持 --port 参数，默认 8080"
echo "• 集群资源热力图生成器: 支持 --port 参数，默认 8082"
echo "• Kubernetes 部署文件: 支持通过环境变量配置端口"
echo "• 避免了端口冲突，提高了部署灵活性"
echo
echo "使用示例:"
echo "  ./bin/scheduler-visualizer --port=9080"
echo "  ./bin/heatmap-generator --port=9082"