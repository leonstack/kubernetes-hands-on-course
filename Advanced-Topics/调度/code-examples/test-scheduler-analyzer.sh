#!/bin/bash

# 测试调度器分析器的改进功能
echo "=== 测试调度器分析器 ==="

# 设置测试目录
TEST_DIR="./test-results"
mkdir -p $TEST_DIR

echo "1. 测试JSON格式输出到文件..."
./bin/scheduler-analyzer -output="$TEST_DIR/analysis-report.json" -format="json"
if [ $? -eq 0 ]; then
    echo "✅ JSON报告生成成功"
    echo "📄 报告文件: $TEST_DIR/analysis-report.json"
    if [ -f "$TEST_DIR/analysis-report.json" ]; then
        echo "📊 报告大小: $(wc -c < "$TEST_DIR/analysis-report.json") bytes"
        echo "🔍 报告预览:"
        head -20 "$TEST_DIR/analysis-report.json"
    fi
else
    echo "❌ JSON报告生成失败"
fi

echo ""
echo "2. 测试文本格式输出到文件..."
./bin/scheduler-analyzer -output="$TEST_DIR/analysis-report.txt" -format="text"
if [ $? -eq 0 ]; then
    echo "✅ 文本报告生成成功"
    echo "📄 报告文件: $TEST_DIR/analysis-report.txt"
    if [ -f "$TEST_DIR/analysis-report.txt" ]; then
        echo "📊 报告大小: $(wc -c < "$TEST_DIR/analysis-report.txt") bytes"
        echo "🔍 报告预览:"
        head -20 "$TEST_DIR/analysis-report.txt"
    fi
else
    echo "❌ 文本报告生成失败"
fi

echo ""
echo "3. 测试HTTP服务器模式（后台运行5秒）..."
./bin/scheduler-analyzer -port="8081" &
SERVER_PID=$!
sleep 2

echo "📡 测试HTTP接口..."
curl -s "http://localhost:8081/" > "$TEST_DIR/http-response.json" 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✅ HTTP接口响应成功"
    echo "📊 响应大小: $(wc -c < "$TEST_DIR/http-response.json") bytes"
else
    echo "❌ HTTP接口测试失败（可能是因为没有Kubernetes集群连接）"
fi

# 停止服务器
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "=== 测试完成 ==="
echo "📁 测试结果保存在: $TEST_DIR/"
ls -la $TEST_DIR/

echo ""
echo "💡 提示:"
echo "   - 如果看到连接错误，这是正常的，因为可能没有可用的Kubernetes集群"
echo "   - 改进后的分析器现在包含真实的API集成、重试机制和错误处理"
echo "   - 在有Kubernetes集群的环境中，分析器将提供真实的集群分析数据"