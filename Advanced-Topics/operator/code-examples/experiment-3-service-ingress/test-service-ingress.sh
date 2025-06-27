#!/bin/bash

# 实验三：服务暴露和 Ingress 测试脚本

echo "=== 实验三：服务暴露和 Ingress 测试 ==="

# 重新生成 CRD（包含新字段）
echo "1. 重新生成 CRD..."
make manifests
make install

# 部署测试应用
echo "2. 部署测试应用..."
kubectl apply -f service-ingress-tests.yaml

# 等待应用就绪
echo "3. 等待应用就绪..."
kubectl wait --for=condition=ready pod -l app=demo-app-nodeport --timeout=300s
kubectl wait --for=condition=ready pod -l app=demo-app-loadbalancer --timeout=300s
kubectl wait --for=condition=ready pod -l app=demo-app-ingress --timeout=300s
kubectl wait --for=condition=ready pod -l app=demo-app-ingress-tls --timeout=300s

# 验证 NodePort 服务
echo "4. 验证 NodePort 服务..."
echo "SpringBootApp 状态:"
kubectl get springbootapp demo-app-nodeport

echo "Service 状态:"
kubectl get service demo-app-nodeport
kubectl describe service demo-app-nodeport

NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
if [ -z "$NODE_IP" ]; then
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
fi
echo "NodePort 访问地址: http://$NODE_IP:30080"

# 验证 LoadBalancer 服务
echo "5. 验证 LoadBalancer 服务..."
echo "SpringBootApp 状态:"
kubectl get springbootapp demo-app-loadbalancer

echo "Service 状态:"
kubectl get service demo-app-loadbalancer
kubectl describe service demo-app-loadbalancer

LB_IP=$(kubectl get service demo-app-loadbalancer -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ ! -z "$LB_IP" ]; then
    echo "LoadBalancer 访问地址: http://$LB_IP:8080"
else
    echo "LoadBalancer IP 尚未分配，请稍后检查"
fi

# 验证 Ingress
echo "6. 验证 Ingress..."
echo "SpringBootApp 状态:"
kubectl get springbootapp demo-app-ingress

echo "Ingress 状态:"
kubectl get ingress demo-app-ingress
kubectl describe ingress demo-app-ingress

INGRESS_IP=$(kubectl get ingress demo-app-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ ! -z "$INGRESS_IP" ]; then
    echo "Ingress 访问地址: http://demo.example.com (IP: $INGRESS_IP)"
    echo "请确保 demo.example.com 解析到 $INGRESS_IP"
else
    echo "Ingress IP 尚未分配，请稍后检查"
fi

# 验证 TLS Ingress
echo "7. 验证 TLS Ingress..."
echo "SpringBootApp 状态:"
kubectl get springbootapp demo-app-ingress-tls

echo "Ingress 状态:"
kubectl get ingress demo-app-ingress-tls
kubectl describe ingress demo-app-ingress-tls

TLS_INGRESS_IP=$(kubectl get ingress demo-app-ingress-tls -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ ! -z "$TLS_INGRESS_IP" ]; then
    echo "TLS Ingress 访问地址: https://secure.example.com (IP: $TLS_INGRESS_IP)"
    echo "请确保 secure.example.com 解析到 $TLS_INGRESS_IP"
else
    echo "TLS Ingress IP 尚未分配，请稍后检查"
fi

# 测试服务连通性
echo "8. 测试服务连通性..."
echo "测试 NodePort 服务:"
if [ ! -z "$NODE_IP" ]; then
    curl -s -o /dev/null -w "%{http_code}" http://$NODE_IP:30080/actuator/health || echo "NodePort 服务连接失败"
fi

echo "测试 LoadBalancer 服务:"
if [ ! -z "$LB_IP" ]; then
    curl -s -o /dev/null -w "%{http_code}" http://$LB_IP:8080/actuator/health || echo "LoadBalancer 服务连接失败"
fi

echo "=== 实验三测试完成 ==="
echo "请根据您的环境配置相应的 DNS 解析和证书管理"