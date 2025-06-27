#!/bin/bash

# 实验一：基础 Operator 部署脚本

echo "=== 实验一：基础 Spring Boot Operator 部署 ==="

# 创建 API
echo "1. 创建 SpringBootApp API..."
operator-sdk create api --group=springboot --version=v1 --kind=SpringBootApp --resource --controller

# 生成 CRD 和 RBAC
echo "2. 生成 CRD 和 RBAC..."
make manifests

# 安装 CRD
echo "3. 安装 CRD..."
make install

# 构建和部署 Operator
echo "4. 构建 Operator 镜像..."
make docker-build IMG=springboot-operator:v1.0.0

# 部署 Operator
echo "5. 部署 Operator..."
make deploy IMG=springboot-operator:v1.0.0

# 等待 Operator 就绪
echo "6. 等待 Operator 就绪..."
kubectl wait --for=condition=available --timeout=300s deployment/springboot-operator-controller-manager -n springboot-operator-system

# 部署测试应用
echo "7. 部署测试应用..."
kubectl apply -f test-springboot-app.yaml

# 验证部署
echo "8. 验证部署..."
echo "SpringBootApp 状态:"
kubectl get springbootapp demo-app

echo "Deployment 状态:"
kubectl get deployment demo-app

echo "Service 状态:"
kubectl get service demo-app

echo "Pod 状态:"
kubectl get pods -l app=demo-app

echo "=== 实验一部署完成 ==="