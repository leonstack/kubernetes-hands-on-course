#!/bin/bash

# 实验四：综合微服务应用部署脚本

echo "=== 实验四：综合微服务应用部署 ==="

# 创建命名空间和配置
echo "1. 创建命名空间和配置..."
kubectl apply -f microservices-namespace.yaml

# 部署数据库
echo "2. 部署 PostgreSQL 数据库..."
kubectl apply -f postgres.yaml

# 等待数据库就绪
echo "3. 等待数据库就绪..."
kubectl wait --for=condition=ready pod -l app=postgres -n microservices --timeout=300s

echo "验证数据库连接:"
kubectl exec -n microservices deployment/postgres -- psql -U user -d microservices -c "\l"

# 部署微服务应用
echo "4. 部署微服务应用..."
kubectl apply -f microservices-apps.yaml

# 等待应用就绪
echo "5. 等待应用就绪..."
echo "等待 Gateway Service..."
kubectl wait --for=condition=ready pod -l app=gateway-service -n microservices --timeout=300s

echo "等待 User Service..."
kubectl wait --for=condition=ready pod -l app=user-service -n microservices --timeout=300s

echo "等待 Order Service..."
kubectl wait --for=condition=ready pod -l app=order-service -n microservices --timeout=300s

# 验证部署
echo "6. 验证部署状态..."
echo "SpringBootApp 状态:"
kubectl get springbootapp -n microservices

echo "\nPod 状态:"
kubectl get pods -n microservices

echo "\nService 状态:"
kubectl get services -n microservices

echo "\nIngress 状态:"
kubectl get ingress -n microservices

# 测试服务连通性
echo "7. 测试服务连通性..."

# 测试数据库连接
echo "测试数据库连接:"
kubectl exec -n microservices deployment/postgres -- psql -U user -d userdb -c "SELECT COUNT(*) FROM users;"
kubectl exec -n microservices deployment/postgres -- psql -U user -d orderdb -c "SELECT COUNT(*) FROM orders;"

# 测试服务健康检查
echo "\n测试服务健康检查:"
echo "Gateway Service:"
kubectl exec -n microservices deployment/gateway-service -- curl -s http://localhost:8080/actuator/health

echo "\nUser Service:"
kubectl exec -n microservices deployment/user-service -- curl -s http://localhost:8080/actuator/health

echo "\nOrder Service:"
kubectl exec -n microservices deployment/order-service -- curl -s http://localhost:8080/actuator/health

# 测试服务间通信
echo "\n8. 测试服务间通信..."
echo "通过 Gateway 访问 User Service:"
kubectl exec -n microservices deployment/gateway-service -- curl -s http://localhost:8080/api/users/health

echo "\n通过 Gateway 访问 Order Service:"
kubectl exec -n microservices deployment/gateway-service -- curl -s http://localhost:8080/api/orders/health

# 获取访问信息
echo "\n9. 获取访问信息..."
INGRESS_IP=$(kubectl get ingress gateway-service -n microservices -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ ! -z "$INGRESS_IP" ]; then
    echo "外部访问地址: http://api.microservices.local (IP: $INGRESS_IP)"
    echo "请确保 api.microservices.local 解析到 $INGRESS_IP"
    echo "\n测试端点:"
    echo "  - 健康检查: http://api.microservices.local/actuator/health"
    echo "  - 用户服务: http://api.microservices.local/api/users/"
    echo "  - 订单服务: http://api.microservices.local/api/orders/"
else
    echo "Ingress IP 尚未分配，请稍后检查"
fi

# 端口转发设置
echo "\n10. 设置端口转发（可选）..."
echo "如果需要本地访问，可以运行以下命令:"
echo "kubectl port-forward -n microservices service/gateway-service 8080:8080"
echo "然后访问: http://localhost:8080"

echo "\n=== 实验四部署完成 ==="
echo "\n微服务架构部署成功！"
echo "架构组件:"
echo "  - Gateway Service: API 网关，路由请求"
echo "  - User Service: 用户管理服务"
echo "  - Order Service: 订单管理服务"
echo "  - PostgreSQL: 数据库服务"
echo "\n请查看各服务的日志以确保正常运行:"
echo "kubectl logs -n microservices deployment/gateway-service"
echo "kubectl logs -n microservices deployment/user-service"
echo "kubectl logs -n microservices deployment/order-service"