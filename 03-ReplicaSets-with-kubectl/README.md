# 3. ReplicaSets with kubectl

## 3.0 目录

- [3. ReplicaSets with kubectl](#3-replicasets-with-kubectl)
  - [3.0 目录](#30-目录)
  - [3.1 项目概述](#31-项目概述)
  - [3.2 ReplicaSet 介绍](#32-replicaset-介绍)
    - [3.2.1 什么是 ReplicaSet？](#321-什么是-replicaset)
    - [3.2.2 使用 ReplicaSet 的优势](#322-使用-replicaset-的优势)
  - [3.3 创建 ReplicaSet](#33-创建-replicaset)
    - [3.3.1 创建 ReplicaSet](#331-创建-replicaset)
    - [3.3.2 replicaset-demo.yml 说明](#332-replicaset-demoyml-说明)
    - [3.3.3 配置说明](#333-配置说明)
      - [3.3.3.1 标签和注解优化](#3331-标签和注解优化)
      - [3.3.3.2 资源管理](#3332-资源管理)
      - [3.3.3.3 健康检查](#3333-健康检查)
      - [3.3.3.4 安全配置](#3334-安全配置)
    - [3.3.4 列出 ReplicaSet](#334-列出-replicaset)
    - [3.3.5 描述 ReplicaSet](#335-描述-replicaset)
    - [3.3.6 Pod 管理和监控](#336-pod-管理和监控)
    - [3.3.7 验证 Pod 的所有者关系](#337-验证-pod-的所有者关系)
    - [3.3.8 监控和健康检查](#338-监控和健康检查)
  - [3.4 将 ReplicaSet 暴露为 Service](#34-将-replicaset-暴露为-service)
    - [3.4.1 创建 Service](#341-创建-service)
    - [3.4.2 查看 Service 信息](#342-查看-service-信息)
    - [3.4.3 访问应用程序](#343-访问应用程序)
    - [3.4.4 测试 Service 连接](#344-测试-service-连接)
  - [3.5 🔧 测试 ReplicaSet 可靠性或高可用性](#35--测试-replicaset-可靠性或高可用性)
    - [3.5.1 自愈能力测试](#351-自愈能力测试)
    - [3.5.2 📊 监控自愈过程](#352--监控自愈过程)
  - [3.6 📈 扩展 ReplicaSet](#36--扩展-replicaset)
    - [3.6.1 使用 kubectl scale 命令扩容](#361-使用-kubectl-scale-命令扩容)
    - [3.6.2 使用 YAML 文件扩容](#362-使用-yaml-文件扩容)
  - [3.7 📉 缩减 ReplicaSet](#37--缩减-replicaset)
    - [3.7.1 缩容操作](#371-缩容操作)
    - [3.7.2 🎯 渐进式缩容策略](#372--渐进式缩容策略)
  - [3.8 🧹 清理资源](#38--清理资源)
    - [3.8.1 📋 清理前检查](#381--清理前检查)
    - [3.8.2 🗑️ 逐步清理](#382-️-逐步清理)
    - [3.8.3 🚀 快速清理（一键清理）](#383--快速清理一键清理)
    - [3.8.4 🔍 清理验证和故障排除](#384--清理验证和故障排除)
  - [3.9 ReplicaSet 中的待讨论概念](#39-replicaset-中的待讨论概念)
  - [3.10 📚 最佳实践](#310--最佳实践)
    - [3.10.1 🏷️ 标签和选择器](#3101-️-标签和选择器)
    - [3.10.2 🔒 安全配置](#3102--安全配置)
    - [3.10.3 📊 资源管理](#3103--资源管理)
    - [3.10.4 🏥 健康检查](#3104--健康检查)
  - [3.11 🔧 故障排除](#311--故障排除)
    - [3.11.1 常见问题和解决方案](#3111-常见问题和解决方案)
      - [3.11.1.1 Pod 无法启动](#31111-pod-无法启动)
      - [3.11.1.2 ReplicaSet 无法创建 Pod](#31112-replicaset-无法创建-pod)
      - [3.11.1.3 Service 无法访问](#31113-service-无法访问)
      - [3.11.1.4 资源清理问题](#31114-资源清理问题)
    - [3.11.2 🔍 调试命令集合](#3112--调试命令集合)
  - [3.12 📖 总结](#312--总结)
    - [3.12.1 🚀 下一步学习](#3121--下一步学习)
  - [3.13 📚 参考资料](#313--参考资料)

## 3.1 项目概述

本教程演示如何使用 kubectl 管理 Kubernetes ReplicaSet，包含生产级别的最佳实践配置。

## 3.2 ReplicaSet 介绍

### 3.2.1 什么是 ReplicaSet？

- ReplicaSet 是 Kubernetes 中用于确保指定数量的 Pod 副本始终运行的控制器
- 它是 Deployment 的底层实现机制
- 提供自愈能力：当 Pod 失败时自动创建新的 Pod

### 3.2.2 使用 ReplicaSet 的优势

- **高可用性**：确保应用程序始终有足够的副本运行
- **负载分布**：将流量分散到多个 Pod 实例
- **自动恢复**：Pod 故障时自动替换
- **水平扩展**：可以轻松调整副本数量

## 3.3 创建 ReplicaSet

### 3.3.1 创建 ReplicaSet

```bash
# 创建 ReplicaSet
kubectl create -f replicaset-demo.yml

# 或者使用 apply（推荐）
kubectl apply -f replicaset-demo.yml
```

### 3.3.2 replicaset-demo.yml 说明

我们的配置文件包含了生产级别的最佳实践：

```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: my-helloworld-rs
  labels:
    app: my-helloworld
    version: v1.0.0
    component: frontend
    tier: web
  annotations:
    description: "Hello World ReplicaSet for Kubernetes fundamentals demo"
    maintainer: "kubernetes-fundamentals-team"
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-helloworld
      version: v1.0.0
  template:
    metadata:
      labels:
        app: my-helloworld
        version: v1.0.0
        component: frontend
        tier: web
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
    spec:
      containers:
      - name: my-helloworld-app
        image: grissomsh/kube-helloworld:1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /hello
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /hello
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        env:
        - name: APP_NAME
          value: "my-helloworld"
        - name: APP_VERSION
          value: "1.0.0"
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 1000
```

### 3.3.3 配置说明

#### 3.3.3.1 标签和注解优化

- **多层标签**：`app`, `version`, `component`, `tier` 便于管理和选择
- **Prometheus 注解**：支持自动服务发现和监控
- **描述性注解**：提供配置的元信息

#### 3.3.3.2 资源管理

- **资源请求**：确保 Pod 获得最小资源保证
- **资源限制**：防止单个 Pod 消耗过多资源
- **合理配置**：64Mi-128Mi 内存，50m-100m CPU

#### 3.3.3.3 健康检查

- **存活探针**：检测应用程序是否正常运行
- **就绪探针**：确保 Pod 准备好接收流量
- **渐进式检查**：合理的延迟和间隔设置

#### 3.3.3.4 安全配置

- **非 root 用户**：提高容器安全性
- **最小权限**：禁用特权提升
- **能力限制**：移除所有不必要的 Linux 能力

### 3.3.4 列出 ReplicaSet

```bash
# 获取 ReplicaSet 列表
kubectl get replicaset
kubectl get rs

# 获取详细信息
kubectl get rs -o wide

# 使用标签选择器
kubectl get rs -l app=my-helloworld
kubectl get rs -l component=frontend
```

### 3.3.5 描述 ReplicaSet

```bash
# 描述 ReplicaSet 详细信息
kubectl describe rs/my-helloworld-rs
# 或者
kubectl describe rs my-helloworld-rs

# 查看 ReplicaSet 的 YAML 配置
kubectl get rs my-helloworld-rs -o yaml

# 查看 ReplicaSet 的 JSON 配置
kubectl get rs my-helloworld-rs -o json
```

### 3.3.6 Pod 管理和监控

```bash
# 获取 Pod 列表
kubectl get pods
kubectl get pods -l app=my-helloworld

# 获取 Pod 详细信息（包括 IP 和节点）
kubectl get pods -o wide

# 描述特定 Pod
kubectl describe pod <pod-name>

# 查看 Pod 日志
kubectl logs <pod-name>
kubectl logs -f <pod-name>  # 实时查看日志

# 查看所有 Pod 的日志
kubectl logs -l app=my-helloworld
```

### 3.3.7 验证 Pod 的所有者关系

验证 Pod 与 ReplicaSet 的关联关系：

```bash
# 查看 Pod 的所有者引用
kubectl get pods <pod-name> -o yaml | grep -A 10 ownerReferences

# 或者查看完整的 YAML
kubectl get pods <pod-name> -o yaml

# 使用 jsonpath 提取所有者信息
kubectl get pods -l app=my-helloworld -o jsonpath='{.items[*].metadata.ownerReferences[*].name}'
```

在输出中查找 `ownerReferences` 部分的 `name` 字段，确认 Pod 属于正确的 ReplicaSet。

### 3.3.8 监控和健康检查

```bash
# 检查 Pod 状态和就绪情况
kubectl get pods -l app=my-helloworld -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,READY:.status.conditions[?(@.type=="Ready")].status

# 查看资源使用情况（需要 metrics-server）
kubectl top pods -l app=my-helloworld

# 查看事件
kubectl get events --sort-by=.metadata.creationTimestamp
```

## 3.4 将 ReplicaSet 暴露为 Service

### 3.4.1 创建 Service

```bash
# 方法1：使用 kubectl expose 命令
kubectl expose rs my-helloworld-rs \
  --type=NodePort \
  --port=80 \
  --target-port=8080 \
  --name=my-helloworld-rs-service

# 方法2：使用 YAML 文件（推荐）
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: my-helloworld-rs-service
  labels:
    app: my-helloworld
    version: v1.0.0
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: my-helloworld
    version: v1.0.0
EOF
```

### 3.4.2 查看 Service 信息

```bash
# 获取 Service 列表
kubectl get service
kubectl get svc

# 获取详细信息
kubectl get svc -o wide
kubectl describe svc my-helloworld-rs-service

# 查看 Service 的端点
kubectl get endpoints my-helloworld-rs-service

# 获取节点信息
kubectl get nodes -o wide
```

### 3.4.3 访问应用程序

```bash
# 获取 NodePort
NODE_PORT=$(kubectl get svc my-helloworld-rs-service -o jsonpath='{.spec.ports[0].nodePort}')
echo "NodePort: $NODE_PORT"

# 获取节点 IP
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
if [ -z "$NODE_IP" ]; then
  NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
fi
echo "Node IP: $NODE_IP"

# 访问应用程序
echo "访问 URL: http://$NODE_IP:$NODE_PORT/hello"
curl http://$NODE_IP:$NODE_PORT/hello
```

### 3.4.4 测试 Service 连接

```bash
# 使用 kubectl port-forward 进行本地测试
kubectl port-forward svc/my-helloworld-rs-service 8080:80

# 在另一个终端测试
curl http://localhost:8080/hello

# 测试负载均衡
for i in {1..10}; do
  curl -s http://localhost:8080/hello | grep -o 'Pod Name: [^<]*'
done
```

## 3.5 🔧 测试 ReplicaSet 可靠性或高可用性

### 3.5.1 自愈能力测试

测试 Kubernetes 中如何自动实现高可用性或可靠性概念。每当 Pod 由于某些应用程序问题意外终止时，ReplicaSet 应该自动创建该 Pod 以维护配置的所需副本数量来实现高可用性。

```bash
# 获取当前 Pod 列表和状态
kubectl get pods -l app=my-helloworld -o wide

# 记录当前 Pod 数量
echo "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.replicas}')"
echo "就绪副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.readyReplicas}')"

# 选择一个 Pod 进行删除测试
POD_NAME=$(kubectl get pods -l app=my-helloworld -o jsonpath='{.items[0].metadata.name}')
echo "将要删除的 Pod: $POD_NAME"

# 删除 Pod 模拟故障
kubectl delete pod $POD_NAME

# 立即查看 ReplicaSet 状态
kubectl get rs my-helloworld-rs

# 验证新 Pod 是否自动创建
echo "等待 Pod 重新创建..."
sleep 5
kubectl get pods -l app=my-helloworld -o wide

# 验证新 Pod 的年龄和名称
echo "\n=== 自愈验证 ==="
echo "新 Pod 列表（注意创建时间）："
kubectl get pods -l app=my-helloworld --sort-by=.metadata.creationTimestamp
```

### 3.5.2 📊 监控自愈过程

```bash
# 实时监控 Pod 状态变化
kubectl get pods -l app=my-helloworld -w &
WATCH_PID=$!

# 删除多个 Pod 测试
kubectl delete pods -l app=my-helloworld --grace-period=0 --force

# 等待观察
sleep 30

# 停止监控
kill $WATCH_PID 2>/dev/null

# 查看相关事件
kubectl get events --sort-by=.metadata.creationTimestamp | grep my-helloworld
```

## 3.6 📈 扩展 ReplicaSet

### 3.6.1 使用 kubectl scale 命令扩容

```bash
# 方法1：使用 kubectl scale 命令（推荐）
kubectl scale --replicas=10 rs/my-helloworld-rs

# 方法2：使用 kubectl patch 命令
kubectl patch rs my-helloworld-rs -p '{"spec":{"replicas":10}}'

# 实时监控扩容过程
kubectl get rs my-helloworld-rs -w &
WATCH_PID=$!

# 查看 Pod 创建过程
kubectl get pods -l app=my-helloworld

# 等待扩容完成
sleep 30
kill $WATCH_PID 2>/dev/null

# 验证扩容结果
echo "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.replicas}')"
echo "就绪副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.status.readyReplicas}')"
```

### 3.6.2 使用 YAML 文件扩容

```bash
# 修改 replicaset-demo.yml 文件
sed -i 's/replicas: 3/replicas: 6/' replicaset-demo.yml

# 应用更改
kubectl apply -f replicaset-demo.yml

# 验证是否创建了新的 Pod
kubectl get pods -l app=my-helloworld -o wide

# 查看扩容事件
kubectl get events --sort-by=.metadata.creationTimestamp | grep ScalingReplicaSet
```

## 3.7 📉 缩减 ReplicaSet

### 3.7.1 缩容操作

```bash
# 缩减到 2 个副本
kubectl scale --replicas=2 rs/my-helloworld-rs

# 监控缩容过程
kubectl get pods -l app=my-helloworld -w &
WATCH_PID=$!

# 等待缩容完成
sleep 20
kill $WATCH_PID 2>/dev/null

# 验证缩容结果
kubectl get rs my-helloworld-rs
kubectl get pods -l app=my-helloworld

# 查看哪些 Pod 被终止
kubectl get events --sort-by=.metadata.creationTimestamp | grep -E "(Killing|SuccessfulDelete)"
```

### 3.7.2 🎯 渐进式缩容策略

```bash
# 渐进式缩容（避免服务中断）
echo "当前副本数：$(kubectl get rs my-helloworld-rs -o jsonpath='{.spec.replicas}')"

# 第一步：缩减到 5 个副本
kubectl scale --replicas=5 rs/my-helloworld-rs
echo "等待缩容到 5 个副本..."
sleep 15

# 第二步：缩减到 3 个副本
kubectl scale --replicas=3 rs/my-helloworld-rs
echo "等待缩容到 3 个副本..."
sleep 15

# 最终：缩减到 2 个副本
kubectl scale --replicas=2 rs/my-helloworld-rs
echo "最终缩容到 2 个副本"

# 验证最终状态
kubectl get rs my-helloworld-rs
kubectl get pods -l app=my-helloworld
```

## 3.8 🧹 清理资源

### 3.8.1 📋 清理前检查

```bash
# 查看当前资源状态
echo "=== 当前 ReplicaSet 状态 ==="
kubectl get rs -l app=my-helloworld

echo "\n=== 当前 Pod 状态 ==="
kubectl get pods -l app=my-helloworld

echo "\n=== 当前 Service 状态 ==="
kubectl get svc -l app=my-helloworld

echo "\n=== 相关事件 ==="
kubectl get events --sort-by=.metadata.creationTimestamp | grep my-helloworld | tail -5
```

### 3.8.2 🗑️ 逐步清理

```bash
# 步骤1：删除 Service（停止外部访问）
echo "删除 Service..."
kubectl delete svc my-helloworld-rs-service

# 验证 Service 删除
kubectl get svc | grep my-helloworld || echo "Service 已删除"

# 步骤2：缩减 ReplicaSet 到 0（优雅停止 Pod）
echo "\n缩减 ReplicaSet 到 0..."
kubectl scale --replicas=0 rs/my-helloworld-rs

# 等待 Pod 终止
echo "等待 Pod 终止..."
sleep 10
kubectl get pods -l app=my-helloworld

# 步骤3：删除 ReplicaSet
echo "\n删除 ReplicaSet..."
kubectl delete rs my-helloworld-rs

# 最终验证
echo "\n=== 清理验证 ==="
kubectl get rs -l app=my-helloworld || echo "ReplicaSet 已删除"
kubectl get pods -l app=my-helloworld || echo "Pod 已删除"
kubectl get svc -l app=my-helloworld || echo "Service 已删除"
```

### 3.8.3 🚀 快速清理（一键清理）

```bash
# 使用标签选择器一次性删除所有相关资源
kubectl delete all -l app=my-helloworld

# 或者删除特定资源类型
kubectl delete rs,svc -l app=my-helloworld

# 强制删除（如果资源卡住）
kubectl delete rs my-helloworld-rs --grace-period=0 --force
kubectl delete pods -l app=my-helloworld --grace-period=0 --force
```

### 3.8.4 🔍 清理验证和故障排除

```bash
# 检查是否有残留资源
echo "=== 检查残留资源 ==="
kubectl get all -l app=my-helloworld
kubectl get events | grep my-helloworld

# 检查命名空间中的所有资源
kubectl get all -n default | grep my-helloworld

# 如果发现卡住的资源，查看详细信息
# kubectl describe rs my-helloworld-rs
# kubectl describe pod <stuck-pod-name>

# 清理完成确认
if [ -z "$(kubectl get all -l app=my-helloworld 2>/dev/null)" ]; then
  echo "✅ 所有资源已成功清理"
else
  echo "⚠️  仍有残留资源，请手动检查"
  kubectl get all -l app=my-helloworld
fi
```

## 3.9 ReplicaSet 中的待讨论概念

- 我们没有讨论 **标签和选择器（Labels & Selectors）**
- 当我们学习编写 Kubernetes YAML 清单时，可以详细了解这个概念。
- 因此我们将在 **ReplicaSets-YAML** 部分了解这一点。

## 3.10 📚 最佳实践

### 3.10.1 🏷️ 标签和选择器

```bash
# 使用有意义的标签
app: my-helloworld          # 应用名称
version: v1.0.0             # 版本号
component: frontend         # 组件类型
environment: development    # 环境
tier: web                   # 层级

# 标签查询示例
kubectl get pods -l app=my-helloworld,version=v1.0.0
kubectl get pods -l 'environment in (development,staging)'
kubectl get pods -l 'tier!=database'
```

### 3.10.2 🔒 安全配置

```yaml
# 安全上下文最佳实践
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 3000
  fsGroup: 2000
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

### 3.10.3 📊 资源管理

```yaml
# 资源限制最佳实践
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "100m"
```

### 3.10.4 🏥 健康检查

```yaml
# 健康检查最佳实践
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## 3.11 🔧 故障排除

### 3.11.1 常见问题和解决方案

#### 3.11.1.1 Pod 无法启动

```bash
# 查看 Pod 状态和事件
kubectl describe pod <pod-name>
kubectl get events --sort-by=.metadata.creationTimestamp

# 查看 Pod 日志
kubectl logs <pod-name>
kubectl logs <pod-name> --previous  # 查看之前容器的日志
```

#### 3.11.1.2 ReplicaSet 无法创建 Pod

```bash
# 检查 ReplicaSet 状态
kubectl describe rs my-helloworld-rs

# 检查节点资源
kubectl top nodes
kubectl describe nodes

# 检查镜像拉取问题
kubectl get events | grep "Failed to pull image"
```

#### 3.11.1.3 Service 无法访问

```bash
# 检查 Service 和 Endpoints
kubectl describe svc my-helloworld-rs-service
kubectl get endpoints my-helloworld-rs-service

# 检查标签选择器匹配
kubectl get pods --show-labels
kubectl get svc my-helloworld-rs-service -o yaml | grep selector -A 5
```

#### 3.11.1.4 资源清理问题

```bash
# 强制删除卡住的资源
kubectl delete rs my-helloworld-rs --grace-period=0 --force

# 检查 finalizers
kubectl get rs my-helloworld-rs -o yaml | grep finalizers -A 5

# 手动编辑移除 finalizers（谨慎使用）
kubectl patch rs my-helloworld-rs -p '{"metadata":{"finalizers":[]}}' --type=merge
```

### 3.11.2 🔍 调试命令集合

```bash
# 资源状态检查
kubectl get all -l app=my-helloworld
kubectl describe rs my-helloworld-rs
kubectl get events --sort-by=.metadata.creationTimestamp

# 网络连接测试
kubectl run debug --image=busybox --rm -it --restart=Never -- /bin/sh
# 在 debug pod 中测试连接
# wget -qO- http://my-helloworld-rs-service/hello

# 资源使用监控
kubectl top pods -l app=my-helloworld
kubectl top nodes
```

## 3.12 📖 总结

通过本教程，你学会了：

✅ **ReplicaSet 基础概念**

- 理解 ReplicaSet 的作用和工作原理
- 掌握标签选择器的使用

✅ **实际操作技能**

- 创建和管理 ReplicaSet
- 配置健康检查和资源限制
- 暴露服务并进行访问测试

✅ **运维管理能力**

- 扩缩容操作和监控
- 自愈能力测试
- 资源清理和故障排除

✅ **最佳实践应用**

- 安全配置和资源管理
- 标签策略和监控方法
- 调试技巧和问题解决

### 3.12.1 🚀 下一步学习

- **Deployment**: 学习更高级的部署控制器
- **Service**: 深入了解服务发现和负载均衡
- **ConfigMap & Secret**: 配置和密钥管理
- **Ingress**: 外部访问和路由管理
- **Monitoring**: 监控和日志收集

---

## 3.13 📚 参考资料

- [Kubernetes ReplicaSet 官方文档](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/)
- [kubectl 命令参考](https://kubernetes.io/docs/reference/kubectl/)
- [Kubernetes 标签和选择器](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [Kubernetes 最佳实践](https://kubernetes.io/docs/concepts/configuration/overview/)
- [Pod 安全标准](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [资源管理指南](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
