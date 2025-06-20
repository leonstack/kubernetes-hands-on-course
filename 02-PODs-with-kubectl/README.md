# 使用 kubectl 管理 POD

## 目录

- [步骤 01：POD 介绍](#步骤-01pod-介绍)
- [步骤 02：POD 演示](#步骤-02pod-演示)
- [步骤 03：NodePort Service 介绍](#步骤-03nodeport-service-介绍)
- [步骤 04：演示 - 使用 Service 暴露 Pod](#步骤-04演示---使用-service-暴露-pod)
- [步骤 05：与 Pod 交互](#步骤-05与-pod-交互)
- [步骤 06：获取 Pod 和 Service 的 YAML 输出](#步骤-06获取-pod-和-service-的-yaml-输出)
- [步骤 07：清理](#步骤-07清理)
- [步骤 08：故障排查指南](#步骤-08故障排查指南)
- [步骤 09：最佳实践](#步骤-09最佳实践)
- [步骤 10：自动化演示脚本](#步骤-10自动化演示脚本)

## 步骤 01：POD 介绍

### 什么是 POD？

POD 是 Kubernetes 中最小的可部署单元，它具有以下特征：

- **共享网络**：Pod 中的所有容器共享同一个网络命名空间，包括 IP 地址和端口空间
- **共享存储**：Pod 中的容器可以通过 Volume 共享数据
- **生命周期管理**：Pod 中的所有容器作为一个整体进行调度、启动和停止
- **原子性**：Pod 要么全部运行，要么全部停止

### 什么是多容器 POD？

多容器 Pod 是指在一个 Pod 中运行多个容器的模式，常见的使用场景包括：

#### 1. Sidecar 模式

```text
┌─────────────────────────────────┐
│             Pod                 │
│  ┌─────────────┐ ┌─────────────┐ │
│  │ 主应用容器   │ │ Sidecar容器  │ │
│  │ (Web Server)│ │ (Log Agent) │ │
│  └─────────────┘ └─────────────┘ │
│         共享网络和存储            │
└─────────────────────────────────┘
```

#### 2. Ambassador 模式

- 主容器：应用程序
- Ambassador 容器：代理外部服务连接

#### 3. Adapter 模式

- 主容器：应用程序
- Adapter 容器：数据格式转换

### POD 的生命周期

```text
Pending → Running → Succeeded/Failed
   ↓         ↓           ↓
调度中    运行中      已完成/失败
```

**状态说明：**

- **Pending**：Pod 已被创建，但一个或多个容器尚未启动
- **Running**：Pod 已绑定到节点，所有容器都已创建，至少一个容器正在运行
- **Succeeded**：Pod 中的所有容器都已成功终止
- **Failed**：Pod 中的所有容器都已终止，至少一个容器以失败状态终止
- **Unknown**：无法获取 Pod 状态

## 步骤 02：POD 演示

### 获取工作节点状态

- 验证 Kubernetes 工作节点是否就绪。

```bash
# 获取工作节点状态
kubectl get nodes

# 使用 wide 选项获取工作节点状态
kubectl get nodes -o wide
```

### 创建 Pod

- 创建一个 Pod

```bash
# 模板
kubectl run <desired-pod-name> --image <Container-Image> --generator=run-pod/v1

# 替换 Pod 名称和容器镜像
kubectl run my-first-pod --image grissomsh/kubenginx:1.0.0 --generator=run-pod/v1
```

- **重要说明：** 如果不使用 **--generator=run-pod/v1**，它将创建一个带有 Deployment 的 Pod，这是另一个核心 Kubernetes 概念，我们将在接下来的几分钟内学习。
- **重要说明：**
  - 在 **Kubernetes 1.18 版本** 中，**kubectl run** 命令进行了大量清理。
  - 下面的命令足以创建一个 Pod 而不创建 Deployment。我们不需要添加 **--generator=run-pod/v1**

```bash
kubectl run my-first-pod --image grissomsh/kubenginx:1.0.0
```  

### 列出 Pod

- 获取 Pod 列表

```bash
# 列出 Pod
kubectl get pods

# Pod 的别名是 po
kubectl get po
```

### 使用 wide 选项列出 Pod

- 使用 wide 选项列出 Pod，同时提供 Pod 运行所在的节点信息

```bash
kubectl get pods -o wide
```

### 后台发生了什么？

  1. Kubernetes 创建了一个 Pod
  2. 从 Docker Hub 拉取了 Docker 镜像
  3. 在 Pod 中创建了容器
  4. 启动了 Pod 中的容器

### 描述 Pod

- 描述 POD，主要在故障排除时需要。
- 显示的事件在故障排除时会有很大帮助。

```bash
# 获取 Pod 名称列表
kubectl get pods

# 描述 Pod
kubectl describe pod <Pod-Name>
kubectl describe pod my-first-pod 
```

### 访问应用程序

- 目前我们只能在工作节点内部访问此应用程序。
- 要从外部访问它，我们需要创建一个 **NodePort Service**。
- **Service** 是 Kubernetes 中一个非常重要的概念。

### 删除 Pod

```bash
# 获取 Pod 名称列表
kubectl get pods

# 删除 Pod
kubectl delete pod <Pod-Name>
kubectl delete pod my-first-pod
```

## 步骤 03：NodePort Service 介绍

- 什么是 Kubernetes 中的 Service？
- 什么是 NodePort Service？
- 它是如何工作的？

## 步骤 04：演示 - 使用 Service 暴露 Pod

- 使用 Service（NodePort Service）暴露 Pod 以从外部（互联网）访问应用程序
- **端口**
  - **port：** NodePort Service 在 Kubernetes 集群内部监听的端口
  - **targetPort：** 我们在这里定义应用程序运行的容器端口。
  - **NodePort：** 我们可以访问应用程序的工作节点端口。

```bash
# 创建 Pod
kubectl run <desired-pod-name> --image <Container-Image> --generator=run-pod/v1
kubectl run my-first-pod --image grissomsh/kubenginx:1.0.0 --generator=run-pod/v1

# 将 Pod 暴露为 Service
kubectl expose pod <Pod-Name>  --type=NodePort --port=80 --name=<Service-Name>
kubectl expose pod my-first-pod  --type=NodePort --port=80 --name=my-first-service

# 获取 Service 信息
kubectl get service
kubectl get svc

# 获取工作节点的公网 IP
kubectl get nodes -o wide
```

- **使用公网 IP 访问应用程序**

```bash
http://<node1-public-ip>:<Node-Port>
```

- **关于 target-port 的重要说明**
  - 如果未定义 target-port，默认情况下为了方便，**targetPort** 会设置为与 **port** 字段相同的值。

```bash
# 下面的命令在访问应用程序时会失败，因为服务端口 (81) 和容器端口 (80) 不同
kubectl expose pod my-first-pod  --type=NodePort --port=81 --name=my-first-service2     

# 使用容器端口 (--target-port) 将 Pod 暴露为 Service
kubectl expose pod my-first-pod  --type=NodePort --port=81 --target-port=80 --name=my-first-service3

# 获取 Service 信息
kubectl get service
kubectl get svc

# 获取工作节点的公网 IP
kubectl get nodes -o wide
```

- **使用公网 IP 访问应用程序**

```bash
http://<node1-public-ip>:<Node-Port>
```

## 步骤 05：与 Pod 交互

### 验证 Pod 日志

```bash
# 获取 Pod 名称
kubectl get po

# 输出 Pod 日志
kubectl logs <pod-name>
kubectl logs my-first-pod

```bash
# 使用 -f 选项流式输出 Pod 日志，并访问应用程序查看日志

kubectl logs <pod-name>
kubectl logs -f my-first-pod

```

- **重要说明**
  - 请参考下面的链接并搜索 **Interacting with running Pods** 以了解更多日志选项
  - 故障排除技能非常重要。请仔细阅读所有可用的日志选项并掌握它们。
  - **参考：** <https://kubernetes.io/docs/reference/kubectl/cheatsheet/>

### 连接到 POD 中的容器

- **连接到 POD 中的容器并执行命令**

```bash
# 连接到 POD 中的 Nginx 容器

kubectl exec -it <pod-name> -- /bin/bash
kubectl exec -it my-first-pod -- /bin/bash

# 在 Nginx 容器中执行一些命令

ls
cd /usr/share/nginx/html
cat index.html
exit
```

- **在容器中运行单个命令**

```bash
kubectl exec -it <pod-name> env

# 示例命令
kubectl exec -it my-first-pod env
kubectl exec -it my-first-pod ls
kubectl exec -it my-first-pod cat /usr/share/nginx/html/index.html

```

## 步骤 06：获取 Pod 和 Service 的 YAML 输出

### 获取 YAML 输出

```bash
# 获取 Pod 定义的 YAML 输出
kubectl get pod my-first-pod -o yaml

# 获取 Service 定义的 YAML 输出
kubectl get service my-first-service -o yaml
```

## 步骤 07：清理

```bash
# 获取默认命名空间中的所有对象
kubectl get all

# 删除 Service
kubectl delete svc my-first-service
kubectl delete svc my-first-service2
kubectl delete svc my-first-service3

# 删除 Pod
kubectl delete pod my-first-pod

# 获取默认命名空间中的所有对象
kubectl get all
```

## 步骤 08：故障排查指南

### 常见问题及解决方案

#### 1. Pod 状态为 Pending

**可能原因：**

- 资源不足（CPU、内存）
- 节点选择器不匹配
- 存储卷挂载问题

**排查命令：**

```bash
# 查看 Pod 详细信息
kubectl describe pod <pod-name>

# 查看节点资源使用情况
kubectl top nodes

# 查看事件
kubectl get events --sort-by=.metadata.creationTimestamp
```

#### 2. Pod 状态为 CrashLoopBackOff

**可能原因：**

- 应用程序启动失败
- 配置错误
- 依赖服务不可用

**排查命令：**

```bash
# 查看 Pod 日志
kubectl logs <pod-name> --previous

# 实时查看日志
kubectl logs -f <pod-name>

# 进入容器调试
kubectl exec -it <pod-name> -- /bin/sh
```

#### 3. 无法访问应用程序

**可能原因：**

- Service 配置错误
- 端口映射问题
- 网络策略限制

**排查命令：**

```bash
# 检查 Service 配置
kubectl describe service <service-name>

# 检查端点
kubectl get endpoints

# 测试网络连通性
kubectl run test-pod --image=busybox --rm -it -- wget -qO- <service-ip>:<port>
```

### 调试工具和技巧

#### 1. 使用临时调试容器

```bash
# 创建调试 Pod
kubectl run debug-pod --image=busybox --rm -it -- /bin/sh

# 在现有 Pod 中添加调试容器（Kubernetes 1.23+）
kubectl debug <pod-name> -it --image=busybox --target=<container-name>
```

#### 2. 网络调试

```bash
# 测试 DNS 解析
nslookup kubernetes.default.svc.cluster.local

# 测试网络连通性
ping <pod-ip>
telnet <service-ip> <port>
```

#### 3. 资源监控

```bash
# 查看资源使用情况
kubectl top pods
kubectl top nodes

# 查看资源限制
kubectl describe pod <pod-name> | grep -A 5 "Limits\|Requests"
```

## 步骤 09：最佳实践

### 1. 资源管理

#### 设置资源请求和限制

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: resource-demo
spec:
  containers:
  - name: app
    image: nginx
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
```

### 2. 健康检查

#### 配置存活性和就绪性探针

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: health-check-demo
spec:
  containers:
  - name: app
    image: nginx
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

### 3. 安全配置

#### 使用非 root 用户

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: security-demo
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
  - name: app
    image: nginx
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
```

### 4. 标签和注解

#### 使用有意义的标签

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: labeled-pod
  labels:
    app: nginx
    version: v1.0
    environment: production
    tier: frontend
  annotations:
    description: "Nginx web server for production"
    maintainer: "team@company.com"
spec:
  containers:
  - name: nginx
    image: nginx:1.20
```

### 5. 环境变量管理

#### 使用 ConfigMap 和 Secret

```bash
# 创建 ConfigMap
kubectl create configmap app-config --from-literal=database_url=mysql://localhost:3306

# 创建 Secret
kubectl create secret generic app-secret --from-literal=password=mysecretpassword
```

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: config-demo
spec:
  containers:
  - name: app
    image: nginx
    env:
    - name: DATABASE_URL
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: database_url
    - name: PASSWORD
      valueFrom:
        secretKeyRef:
          name: app-secret
          key: password
```

## 步骤 10：自动化演示脚本

为了方便演示和学习，我们提供了自动化脚本来执行本教程中的所有步骤。

### 使用方法

```bash
# 运行完整演示
./pod-demo.sh

# 运行特定步骤
./pod-demo.sh --step 2

# 清理所有资源
./pod-demo.sh --cleanup

# 查看帮助
./pod-demo.sh --help
```

### 脚本功能

- ✅ 自动检查 Kubernetes 集群状态
- ✅ 逐步演示 Pod 创建和管理
- ✅ 自动创建和测试 Service
- ✅ 演示日志查看和容器交互
- ✅ 自动清理演示资源
- ✅ 错误处理和重试机制
- ✅ 彩色输出和进度提示

---

## 参考资源

- [Kubernetes 官方文档 - Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [kubectl 命令参考](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Pod 生命周期](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [多容器 Pod 模式](https://kubernetes.io/blog/2015/06/the-distributed-system-toolkit-patterns/)

## 贡献

欢迎提交 Issue 和 Pull Request 来改进本教程！

## 许可证

本项目采用 MIT 许可证。
