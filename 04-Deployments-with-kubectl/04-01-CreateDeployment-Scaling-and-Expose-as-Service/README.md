# Kubernetes - Deployment 管理指南

## 1. 📋 项目概述

本教程演示如何使用 kubectl 管理 Kubernetes Deployment，包含创建、扩展、暴露服务等核心操作，以及生产级别的最佳实践配置。

## 2. 🚀 Deployment 介绍

### 2.1 什么是 Deployment？

- **Deployment** 是 Kubernetes 中用于管理应用程序部署的高级控制器
- 它管理 ReplicaSet，而 ReplicaSet 管理 Pod
- 提供声明式更新、回滚、暂停和恢复等功能
- 是生产环境中部署应用程序的推荐方式

### 2.2 使用 Deployment 的优势

- **声明式管理**：描述期望状态，Kubernetes 自动维护
- **滚动更新**：零停机时间的应用程序更新
- **版本控制**：支持回滚到之前的版本
- **自动扩缩容**：根据需求调整副本数量
- **自愈能力**：自动替换失败的 Pod
- **暂停和恢复**：支持分阶段部署策略

## 3. 🚀 创建 Deployment

### 3.1 基础创建操作

创建 Deployment 来推出 ReplicaSet，并验证各层级资源的创建情况。

**Docker 镜像位置：** <https://hub.docker.com/repository/docker/grissomsh/kubenginx>

```bash
# 创建 Deployment（基础命令）
kubectl create deployment <Deployment-Name> --image=<Container-Image>
kubectl create deployment my-first-deployment --image=grissomsh/kubenginx:1.0.0

# 创建 Deployment（带更多选项）
kubectl create deployment my-first-deployment \
  --image=grissomsh/kubenginx:1.0.0 \
  --replicas=3 \
  --port=80

# 验证 Deployment
kubectl get deployments
kubectl get deploy
kubectl get deploy -o wide

# 描述 Deployment 详细信息
kubectl describe deployment <deployment-name>
kubectl describe deployment my-first-deployment

# 查看 Deployment 的 YAML 配置
kubectl get deployment my-first-deployment -o yaml

# 验证 ReplicaSet
kubectl get rs
kubectl get rs -l app=my-first-deployment

# 验证 Pod
kubectl get po
kubectl get po -l app=my-first-deployment
kubectl get po -o wide
```

### 3.2 验证部署层次结构

```bash
# 查看完整的资源层次关系
echo "=== Deployment 信息 ==="
kubectl get deployment my-first-deployment

echo "\n=== ReplicaSet 信息 ==="
kubectl get rs -l app=my-first-deployment

echo "\n=== Pod 信息 ==="
kubectl get pods -l app=my-first-deployment

# 查看标签和选择器
kubectl get deployment my-first-deployment --show-labels
kubectl get pods -l app=my-first-deployment --show-labels

# 验证 Pod 的所有者引用
POD_NAME=$(kubectl get pods -l app=my-first-deployment -o jsonpath='{.items[0].metadata.name}')
kubectl get pod $POD_NAME -o yaml | grep -A 10 ownerReferences
```

### 3.3 监控部署状态

```bash
# 实时监控 Deployment 状态
kubectl rollout status deployment/my-first-deployment

# 查看 Deployment 历史
kubectl rollout history deployment/my-first-deployment

# 查看事件
kubectl get events --sort-by=.metadata.creationTimestamp | grep my-first-deployment

# 查看资源使用情况（需要 metrics-server）
kubectl top pods -l app=my-first-deployment
```

## 4. 📈 扩展 Deployment

### 4.1 基础扩缩容操作

扩展 Deployment 以增加副本（Pod）数量，实现应用程序的水平扩展。

```bash
# 扩展 Deployment
kubectl scale --replicas=20 deployment/<Deployment-Name>
kubectl scale --replicas=20 deployment/my-first-deployment

# 验证扩展过程
echo "当前副本数：$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.replicas}')"
echo "可用副本数：$(kubectl get deployment my-first-deployment -o jsonpath='{.status.availableReplicas}')"

# 实时监控扩展过程
kubectl get deployment my-first-deployment -w &
WATCH_PID=$!
sleep 30
kill $WATCH_PID 2>/dev/null

# 验证 Deployment
kubectl get deploy
kubectl get deploy -o wide

# 验证 ReplicaSet
kubectl get rs
kubectl get rs -l app=my-first-deployment

# 验证 Pod 分布
kubectl get po -l app=my-first-deployment -o wide
kubectl get po -l app=my-first-deployment --show-labels

# 缩减 Deployment
kubectl scale --replicas=10 deployment/my-first-deployment
kubectl scale --replicas=3 deployment/my-first-deployment
```

### 4.2 渐进式扩缩容策略

```bash
# 渐进式扩容（避免资源突然消耗）
echo "当前副本数：$(kubectl get deployment my-first-deployment -o jsonpath='{.spec.replicas}')"

# 第一步：扩展到 5 个副本
kubectl scale --replicas=5 deployment/my-first-deployment
echo "等待扩容到 5 个副本..."
kubectl rollout status deployment/my-first-deployment

# 第二步：扩展到 10 个副本
kubectl scale --replicas=10 deployment/my-first-deployment
echo "等待扩容到 10 个副本..."
kubectl rollout status deployment/my-first-deployment

# 最终：扩展到 15 个副本
kubectl scale --replicas=15 deployment/my-first-deployment
echo "最终扩容到 15 个副本"
kubectl rollout status deployment/my-first-deployment

# 验证最终状态
kubectl get deployment my-first-deployment
kubectl get pods -l app=my-first-deployment
```

### 4.3 监控扩缩容过程

```bash
# 查看扩缩容事件
kubectl get events --sort-by=.metadata.creationTimestamp | grep -E "(Scaling|SuccessfulCreate)"

# 查看 Pod 创建时间
kubectl get pods -l app=my-first-deployment --sort-by=.metadata.creationTimestamp

# 检查资源使用情况
kubectl top nodes
kubectl top pods -l app=my-first-deployment

# 查看 Deployment 状态
kubectl describe deployment my-first-deployment | grep -A 10 "Conditions:"
```

## 5. 🌐 将 Deployment 暴露为 Service

### 5.1 创建 NodePort Service

使用 Service（NodePort Service）暴露 **Deployment** 以从外部（互联网）访问应用程序。

```bash
# 方法1：使用 kubectl expose 命令
kubectl expose deployment <Deployment-Name> --type=NodePort --port=80 --target-port=80 --name=<Service-Name-To-Be-Created>
kubectl expose deployment my-first-deployment --type=NodePort --port=80 --target-port=80 --name=my-first-deployment-service

# 方法2：使用 YAML 文件创建（推荐）
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: my-first-deployment-service
  labels:
    app: my-first-deployment
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
    name: http
  selector:
    app: my-first-deployment
EOF

# 获取 Service 信息
kubectl get svc
kubectl get svc my-first-deployment-service
kubectl get svc -o wide

# 描述 Service 详细信息
kubectl describe svc my-first-deployment-service

# 查看 Service 的端点
kubectl get endpoints my-first-deployment-service
```

### 5.2 获取访问信息

```bash
# 获取 NodePort
NODE_PORT=$(kubectl get svc my-first-deployment-service -o jsonpath='{.spec.ports[0].nodePort}')
echo "NodePort: $NODE_PORT"

# 获取工作节点的 IP
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
if [ -z "$NODE_IP" ]; then
  NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
fi
echo "Node IP: $NODE_IP"

# 显示完整的访问 URL
echo "访问 URL: http://$NODE_IP:$NODE_PORT"

# 获取所有节点的访问信息
echo "\n=== 所有节点访问信息 ==="
kubectl get nodes -o wide
echo "\n观察：记下 NodePort ($NODE_PORT) 和节点 IP，用于外部访问"
```

### 5.3 测试服务连接

```bash
# 使用 curl 测试连接
curl http://$NODE_IP:$NODE_PORT

# 使用 kubectl port-forward 进行本地测试
kubectl port-forward svc/my-first-deployment-service 8080:80 &
PORT_FORWARD_PID=$!

# 在另一个终端或等待几秒后测试
sleep 5
curl http://localhost:8080

# 停止端口转发
kill $PORT_FORWARD_PID 2>/dev/null

# 测试负载均衡
echo "\n=== 负载均衡测试 ==="
for i in {1..10}; do
  echo "请求 $i:"
  curl -s http://$NODE_IP:$NODE_PORT | grep -o 'Pod Name: [^<]*' || echo "连接失败"
  sleep 1
done
```

### 5.4 验证 Service 配置

```bash
# 验证标签选择器匹配
echo "=== Service 选择器 ==="
kubectl get svc my-first-deployment-service -o yaml | grep -A 5 selector

echo "\n=== Pod 标签 ==="
kubectl get pods -l app=my-first-deployment --show-labels

# 验证端点
echo "\n=== Service 端点 ==="
kubectl get endpoints my-first-deployment-service -o yaml

# 检查服务发现
 echo "\n=== DNS 解析测试 ==="
 kubectl run debug --image=busybox --rm -it --restart=Never -- nslookup my-first-deployment-service
 ```

## 6. 🧹 清理资源

### 6.1 清理前检查

```bash
# 查看当前资源状态
echo "=== 当前 Deployment 状态 ==="
kubectl get deployment my-first-deployment

echo "\n=== 当前 Service 状态 ==="
kubectl get svc my-first-deployment-service

echo "\n=== 当前 Pod 状态 ==="
kubectl get pods -l app=my-first-deployment

echo "\n=== 相关事件 ==="
kubectl get events --sort-by=.metadata.creationTimestamp | grep my-first-deployment | tail -5
```

### 6.2 逐步清理

```bash
# 步骤1：删除 Service（停止外部访问）
echo "删除 Service..."
kubectl delete svc my-first-deployment-service

# 验证 Service 删除
kubectl get svc | grep my-first-deployment || echo "Service 已删除"

# 步骤2：缩减 Deployment 到 0（优雅停止 Pod）
echo "\n缩减 Deployment 到 0..."
kubectl scale --replicas=0 deployment/my-first-deployment

# 等待 Pod 终止
echo "等待 Pod 终止..."
sleep 10
kubectl get pods -l app=my-first-deployment

# 步骤3：删除 Deployment
echo "\n删除 Deployment..."
kubectl delete deployment my-first-deployment

# 最终验证
echo "\n=== 清理验证 ==="
kubectl get deployment my-first-deployment 2>/dev/null || echo "Deployment 已删除"
kubectl get pods -l app=my-first-deployment 2>/dev/null || echo "Pod 已删除"
kubectl get svc my-first-deployment-service 2>/dev/null || echo "Service 已删除"
```

### 6.3 快速清理（一键清理）

```bash
# 使用标签选择器一次性删除所有相关资源
kubectl delete all -l app=my-first-deployment

# 或者删除特定资源类型
kubectl delete deployment,svc -l app=my-first-deployment

# 强制删除（如果资源卡住）
kubectl delete deployment my-first-deployment --grace-period=0 --force
kubectl delete pods -l app=my-first-deployment --grace-period=0 --force
```

## 7. 📚 最佳实践

### 7.1 生产级 Deployment 配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-production-app
  labels:
    app: my-production-app
    version: v1.0.0
    environment: production
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: my-production-app
  template:
    metadata:
      labels:
        app: my-production-app
        version: v1.0.0
    spec:
      containers:
      - name: app
        image: grissomsh/kubenginx:1.0.0
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 7.2 标签和选择器策略

```bash
# 推荐的标签策略
app: my-app                    # 应用名称
version: v1.0.0               # 版本号
component: frontend           # 组件类型
environment: production       # 环境
tier: web                     # 层级

# 标签查询示例
kubectl get pods -l app=my-app,version=v1.0.0
kubectl get pods -l 'environment in (production,staging)'
kubectl get pods -l 'tier!=database'
```

### 7.3 资源管理

```bash
# 设置资源请求和限制
kubectl patch deployment my-first-deployment -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "kubenginx",
          "resources": {
            "requests": {
              "memory": "64Mi",
              "cpu": "50m"
            },
            "limits": {
              "memory": "128Mi",
              "cpu": "100m"
            }
          }
        }]
      }
    }
  }
}'

# 查看资源使用情况
kubectl top pods -l app=my-first-deployment
kubectl describe nodes
```

## 8. 🔧 故障排除

### 8.1 常见问题和解决方案

#### 8.1.1 Deployment 无法创建 Pod

```bash
# 检查 Deployment 状态
kubectl describe deployment my-first-deployment

# 查看事件
kubectl get events --sort-by=.metadata.creationTimestamp

# 检查节点资源
kubectl top nodes
kubectl describe nodes

# 检查镜像拉取问题
kubectl get events | grep "Failed to pull image"
```

#### 8.1.2 Service 无法访问

```bash
# 检查 Service 和 Endpoints
kubectl describe svc my-first-deployment-service
kubectl get endpoints my-first-deployment-service

# 检查标签选择器匹配
kubectl get pods --show-labels
kubectl get svc my-first-deployment-service -o yaml | grep selector -A 5

# 测试内部连接
kubectl run debug --image=busybox --rm -it --restart=Never -- wget -qO- http://my-first-deployment-service
```

#### 8.1.3 扩缩容问题

```bash
# 检查扩缩容状态
kubectl rollout status deployment/my-first-deployment
kubectl describe deployment my-first-deployment

# 查看资源限制
kubectl describe nodes | grep -A 5 "Allocated resources"

# 检查 Pod 调度问题
kubectl get pods -l app=my-first-deployment -o wide
kubectl describe pod <pending-pod-name>
```

### 8.2 调试命令集合

```bash
# 资源状态检查
kubectl get all -l app=my-first-deployment
kubectl describe deployment my-first-deployment
kubectl get events --sort-by=.metadata.creationTimestamp

# 网络连接测试
kubectl run debug --image=busybox --rm -it --restart=Never -- /bin/sh
# 在 debug pod 中测试连接
# wget -qO- http://my-first-deployment-service

# 资源使用监控
kubectl top pods -l app=my-first-deployment
kubectl top nodes
```

## 9. 📖 总结

通过本教程，你学会了：

✅ **Deployment 基础概念**

- 理解 Deployment 的作用和工作原理
- 掌握 Deployment、ReplicaSet、Pod 的层次关系

✅ **实际操作技能**

- 创建和管理 Deployment
- 扩缩容操作和监控
- 暴露服务并进行访问测试

✅ **运维管理能力**

- 监控部署状态和资源使用
- 故障排除和问题解决
- 资源清理和最佳实践

✅ **生产级配置**

- 资源限制和健康检查
- 标签策略和选择器使用
- 滚动更新策略配置

### 9.1 🚀 下一步学习

- **Deployment 更新**: 学习滚动更新和蓝绿部署
- **Deployment 回滚**: 掌握版本回滚和历史管理
- **ConfigMap & Secret**: 配置和密钥管理
- **Ingress**: 高级路由和负载均衡
- **HPA**: 自动水平扩缩容

---

## 10. 📚 参考资料

- [Kubernetes Deployment 官方文档](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [kubectl 命令参考](https://kubernetes.io/docs/reference/kubectl/)
- [Kubernetes 服务发现](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes 最佳实践](https://kubernetes.io/docs/concepts/configuration/overview/)
- [资源管理指南](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
