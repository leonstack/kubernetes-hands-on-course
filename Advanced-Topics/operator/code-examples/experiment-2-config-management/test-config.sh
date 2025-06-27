#!/bin/bash

# 实验二：配置管理功能测试脚本

echo "=== 实验二：配置管理功能测试 ==="

# 重新生成 CRD（包含新字段）
echo "1. 重新生成 CRD..."
make manifests
make install

# 部署配置
echo "2. 部署配置 ConfigMap..."
kubectl apply -f config-demo.yaml

# 验证部署
echo "3. 验证部署..."
echo "SpringBootApp 状态:"
kubectl get springbootapp demo-app-with-config
kubectl describe springbootapp demo-app-with-config

# 检查 Pod 配置
echo "4. 检查 Pod 配置..."
POD_NAME=$(kubectl get pods -l app=demo-app-with-config -o jsonpath='{.items[0].metadata.name}')
echo "Pod Name: $POD_NAME"

if [ ! -z "$POD_NAME" ]; then
    echo "验证配置挂载:"
    kubectl exec $POD_NAME -- ls -la /app/config
    
    echo "查看配置文件内容:"
    kubectl exec $POD_NAME -- cat /app/config/application.yml
    
    echo "验证环境变量:"
    kubectl exec $POD_NAME -- env | grep -E "(JAVA_OPTS|SPRING_)"
else
    echo "Pod 尚未就绪，等待中..."
    kubectl wait --for=condition=ready pod -l app=demo-app-with-config --timeout=300s
    POD_NAME=$(kubectl get pods -l app=demo-app-with-config -o jsonpath='{.items[0].metadata.name}')
    
    if [ ! -z "$POD_NAME" ]; then
        echo "验证配置挂载:"
        kubectl exec $POD_NAME -- ls -la /app/config
        
        echo "查看配置文件内容:"
        kubectl exec $POD_NAME -- cat /app/config/application.yml
        
        echo "验证环境变量:"
        kubectl exec $POD_NAME -- env | grep -E "(JAVA_OPTS|SPRING_)"
    fi
fi

# 测试配置热更新
echo "5. 测试配置热更新..."
echo "更新 ConfigMap..."
kubectl patch configmap demo-config --patch='{
  "data": {
    "application.yml": "server:\n  port: 8080\nspring:\n  application:\n    name: demo-app-updated\nlogging:\n  level:\n    com.example: INFO"
  }
}'

echo "等待 Pod 重启..."
sleep 10

echo "验证配置更新:"
POD_NAME=$(kubectl get pods -l app=demo-app-with-config -o jsonpath='{.items[0].metadata.name}')
if [ ! -z "$POD_NAME" ]; then
    kubectl exec $POD_NAME -- cat /app/config/application.yml
fi

echo "=== 实验二测试完成 ==="