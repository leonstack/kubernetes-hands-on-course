# 7. PODs with YAML

## 7.0 目录

- [7. PODs with YAML](#7-pods-with-yaml)
  - [7.0 目录](#70-目录)
  - [7.1 项目概述](#71-项目概述)
  - [7.2 学习目标](#72-学习目标)
  - [7.3 应用场景](#73-应用场景)
  - [7.4 前置条件](#74-前置条件)
    - [7.4.1 环境要求](#741-环境要求)
    - [7.4.2 验证环境](#742-验证环境)
  - [7.5 Pod YAML 基本结构](#75-pod-yaml-基本结构)
    - [7.5.1 YAML 文件基本结构](#751-yaml-文件基本结构)
    - [7.5.2 核心字段详解](#752-核心字段详解)
    - [7.5.3 常见资源类型](#753-常见资源类型)
    - [7.5.4 API 对象参考](#754-api-对象参考)
  - [7.6 使用 YAML 创建简单的 Pod 定义](#76-使用-yaml-创建简单的-pod-定义)
    - [7.6.1 Pod 基础概念](#761-pod-基础概念)
    - [7.6.2 Pod 定义文件](#762-pod-定义文件)
    - [7.6.3 Pod 字段详解](#763-pod-字段详解)
    - [7.6.4 创建和管理 Pod](#764-创建和管理-pod)
    - [7.6.5 实用技巧](#765-实用技巧)
  - [7.7 创建 NodePort Service](#77-创建-nodeport-service)
    - [7.7.1 Service 基础概念](#771-service-基础概念)
    - [7.7.2 Service 类型对比](#772-service-类型对比)
    - [7.7.3 NodePort Service 定义](#773-nodeport-service-定义)
    - [7.7.4 Service 字段详解](#774-service-字段详解)
    - [7.7.5 创建和管理 Service](#775-创建和管理-service)
    - [7.7.6 访问应用程序](#776-访问应用程序)
    - [7.7.7 网络流量路径](#777-网络流量路径)
    - [7.7.8 注意事项](#778-注意事项)
  - [7.8 最佳实践](#78-最佳实践)
    - [7.8.1 YAML 编写规范](#781-yaml-编写规范)
    - [7.8.2 安全最佳实践](#782-安全最佳实践)
    - [7.8.3 资源管理](#783-资源管理)
  - [7.9 故障排除](#79-故障排除)
    - [7.9.1 常见问题和解决方案](#791-常见问题和解决方案)
      - [7.9.1.1 Pod 状态异常](#7911-pod-状态异常)
      - [7.9.1.2 Service 连接问题](#7912-service-连接问题)
      - [7.9.1.3 镜像拉取失败](#7913-镜像拉取失败)
    - [7.9.2 调试技巧](#792-调试技巧)
  - [7.10 学习总结](#710-学习总结)
    - [7.10.1 关键要点](#7101-关键要点)
    - [7.10.2 进阶学习建议](#7102-进阶学习建议)
    - [7.10.3 下一步学习](#7103-下一步学习)
  - [7.11 API 对象参考](#711-api-对象参考)
    - [7.11.1 官方文档](#7111-官方文档)
    - [7.11.2 实用工具](#7112-实用工具)

## 7.1 项目概述

## 7.2 学习目标

- 掌握 Kubernetes YAML 文件的基本结构和语法
- 学会使用 YAML 文件创建和管理 Pod 资源
- 理解 Pod 与 Service 的关联关系
- 掌握声明式资源管理的最佳实践

## 7.3 应用场景

- 生产环境中的容器化应用部署
- 微服务架构的基础设施即代码
- DevOps 流水线中的自动化部署
- 容器编排和服务发现

## 7.4 前置条件

### 7.4.1 环境要求

- Kubernetes 集群（本地或云端）
- kubectl 命令行工具已配置
- 基本的 YAML 语法知识
- 了解容器和 Docker 基础概念

### 7.4.2 验证环境

```bash
# 检查 kubectl 连接
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查当前命名空间
kubectl config current-context
```

## 7.5 Pod YAML 基本结构

### 7.5.1 YAML 文件基本结构

Kubernetes 中的所有资源都遵循相同的 YAML 结构模式，包含四个核心顶级字段：

- **01-kube-base-definition.yml**

```yml
apiVersion: # API 版本，指定使用的 Kubernetes API 版本
kind:       # 资源类型，如 Pod、Service、Deployment 等
metadata:   # 元数据，包含名称、标签、注解等信息
  name:     # 资源名称（必需）
  labels:   # 标签（可选，用于选择和分组）
  annotations: # 注解（可选，用于存储额外信息）
spec:       # 规格说明，定义资源的期望状态
```

### 7.5.2 核心字段详解

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `apiVersion` | 字符串 | ✅ | 指定 API 版本，如 `v1`、`apps/v1` |
| `kind` | 字符串 | ✅ | 资源类型，如 `Pod`、`Service` |
| `metadata` | 对象 | ✅ | 资源元数据信息 |
| `spec` | 对象 | ✅ | 资源的具体配置规格 |

### 7.5.3 常见资源类型

```yaml
# 工作负载资源
kind: Pod          # 最小部署单元
kind: Deployment   # 应用部署管理
kind: ReplicaSet   # 副本集管理
kind: DaemonSet    # 守护进程集

# 服务发现资源
kind: Service      # 服务抽象
kind: Ingress      # 入口控制器

# 配置和存储
kind: ConfigMap    # 配置映射
kind: Secret       # 敏感信息
kind: PersistentVolume # 持久卷
```

### 7.5.4 API 对象参考

- **Pod API 参考：**  <https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/>
- **完整 API 参考：** <https://kubernetes.io/docs/reference/kubernetes-api/>

## 7.6 使用 YAML 创建简单的 Pod 定义

### 7.6.1 Pod 基础概念

Pod 是 Kubernetes 中最小的可部署单元，通常包含一个或多个紧密耦合的容器。

### 7.6.2 Pod 定义文件

- **02-pod-definition.yml**

```yml
apiVersion: v1 # API 版本：v1 是核心 API 组
kind: Pod      # 资源类型：Pod
metadata:      # 元数据部分
  name: myapp-pod    # Pod 名称（在命名空间内唯一）
  labels:            # 标签（键值对，用于选择和分组）
    app: myapp       # 应用标识
    version: v1.0    # 版本标识
    tier: frontend   # 层级标识
spec:          # Pod 规格定义
  containers:  # 容器列表
    - name: myapp              # 容器名称
      image: grissomsh/kubenginx:1.0.0  # 容器镜像
      ports:                   # 端口配置
        - containerPort: 80    # 容器监听端口
          name: http           # 端口名称
          protocol: TCP        # 协议类型
      resources:               # 资源配置
        requests:              # 资源请求
          memory: "64Mi"       # 内存请求
          cpu: "250m"          # CPU 请求
        limits:                # 资源限制
          memory: "128Mi"      # 内存限制
          cpu: "500m"          # CPU 限制
```

### 7.6.3 Pod 字段详解

| 字段路径 | 类型 | 说明 | 示例 |
|----------|------|------|------|
| `metadata.name` | 字符串 | Pod 名称 | `myapp-pod` |
| `metadata.labels` | 对象 | 标签集合 | `app: myapp` |
| `spec.containers` | 数组 | 容器列表 | 见上方示例 |
| `spec.containers[].name` | 字符串 | 容器名称 | `myapp` |
| `spec.containers[].image` | 字符串 | 容器镜像 | `nginx:1.21` |
| `spec.containers[].ports` | 数组 | 端口配置 | `containerPort: 80` |

### 7.6.4 创建和管理 Pod

```bash
# 创建 Pod（声明式）
kubectl apply -f 02-pod-definition.yml

# 创建 Pod（命令式）
kubectl create -f 02-pod-definition.yml

# 查看 Pod 列表
kubectl get pods

# 查看 Pod 详细信息
kubectl describe pod myapp-pod

# 查看 Pod 日志
kubectl logs myapp-pod

# 进入 Pod 容器
kubectl exec -it myapp-pod -- /bin/bash

# 删除 Pod
kubectl delete pod myapp-pod
# 或使用文件删除
kubectl delete -f 02-pod-definition.yml
```

### 7.6.5 实用技巧

```bash
# 实时监控 Pod 状态
kubectl get pods -w

# 查看 Pod 的 YAML 输出
kubectl get pod myapp-pod -o yaml

# 查看 Pod 事件
kubectl get events --field-selector involvedObject.name=myapp-pod

# 端口转发到本地
kubectl port-forward pod/myapp-pod 8080:80
```

## 7.7 创建 NodePort Service

### 7.7.1 Service 基础概念

Service 为一组 Pod 提供稳定的网络访问入口，解决 Pod IP 动态变化的问题。

### 7.7.2 Service 类型对比

| 类型 | 访问范围 | 使用场景 | 端口范围 |
|------|----------|----------|----------|
| `ClusterIP` | 集群内部 | 内部服务通信 | - |
| `NodePort` | 集群外部 | 开发测试环境 | 30000-32767 |
| `LoadBalancer` | 集群外部 | 生产环境 | 云厂商分配 |
| `ExternalName` | 外部服务 | 服务代理 | - |

### 7.7.3 NodePort Service 定义

- **03-pod-nodeport-service.yml**

```yml
apiVersion: v1
kind: Service
metadata:
  name: myapp-pod-nodeport-service  # Service 名称
  labels:                           # Service 标签
    app: myapp
    service-type: nodeport
spec:
  type: NodePort                    # Service 类型
  selector:                         # Pod 选择器
    app: myapp                      # 匹配标签为 app: myapp 的 Pod
  ports:                            # 端口配置
    - name: http                    # 端口名称
      port: 80                      # Service 端口（集群内访问）
      targetPort: 80                # Pod 容器端口
      nodePort: 31231               # 节点端口（集群外访问）
      protocol: TCP                 # 协议类型
```

### 7.7.4 Service 字段详解

| 字段路径 | 类型 | 说明 | 示例 |
|----------|------|------|------|
| `spec.type` | 字符串 | Service 类型 | `NodePort` |
| `spec.selector` | 对象 | Pod 选择器 | `app: myapp` |
| `spec.ports[].port` | 数字 | Service 端口 | `80` |
| `spec.ports[].targetPort` | 数字/字符串 | 目标端口 | `80` 或 `http` |
| `spec.ports[].nodePort` | 数字 | 节点端口 | `31231` |

### 7.7.5 创建和管理 Service

```bash
# 创建 Service
kubectl apply -f 03-pod-nodeport-service.yml

# 查看 Service 列表
kubectl get services
# 简写形式
kubectl get svc

# 查看 Service 详细信息
kubectl describe service myapp-pod-nodeport-service

# 查看 Service 端点
kubectl get endpoints myapp-pod-nodeport-service

# 获取节点信息
kubectl get nodes -o wide
```

### 7.7.6 访问应用程序

```bash
# 获取节点 IP 和端口信息
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
NODE_PORT=$(kubectl get service myapp-pod-nodeport-service -o jsonpath='{.spec.ports[0].nodePort}')

# 访问应用
echo "应用访问地址: http://${NODE_IP}:${NODE_PORT}"

# 使用 curl 测试
curl http://${NODE_IP}:${NODE_PORT}

# 或直接访问
http://<WorkerNode-Public-IP>:31231
```

### 7.7.7 网络流量路径

```text
外部请求 → 节点IP:NodePort → Service → Pod IP:ContainerPort
     ↓
客户端:随机端口 → 节点:31231 → Service:80 → Pod:80
```

### 7.7.8 注意事项

- **端口范围**：NodePort 默认范围是 30000-32767
- **安全考虑**：NodePort 会在所有节点上开放端口
- **生产环境**：建议使用 LoadBalancer 或 Ingress
- **防火墙**：确保 NodePort 端口在防火墙中开放

## 7.8 最佳实践

### 7.8.1 YAML 编写规范

```yaml
# ✅ 推荐的 Pod 配置
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
    version: v1.0
    environment: production
  annotations:
    description: "Web application pod"
    maintainer: "team@company.com"
spec:
  containers:
  - name: myapp
    image: grissomsh/kubenginx:1.0.0
    ports:
    - containerPort: 80
      name: http
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
    livenessProbe:          # 存活探针
      httpGet:
        path: /health
        port: 80
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:         # 就绪探针
      httpGet:
        path: /ready
        port: 80
      initialDelaySeconds: 5
      periodSeconds: 5
  restartPolicy: Always     # 重启策略
```

### 7.8.2 安全最佳实践

```yaml
# 安全配置示例
spec:
  securityContext:          # Pod 安全上下文
    runAsNonRoot: true      # 非 root 用户运行
    runAsUser: 1000         # 指定用户 ID
    fsGroup: 2000           # 文件系统组 ID
  containers:
  - name: myapp
    securityContext:        # 容器安全上下文
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
```

### 7.8.3 资源管理

```yaml
# 资源配置指南
resources:
  requests:                 # 资源请求（调度依据）
    memory: "64Mi"          # 最小内存需求
    cpu: "250m"             # 最小 CPU 需求
  limits:                   # 资源限制（硬限制）
    memory: "128Mi"         # 最大内存使用
    cpu: "500m"             # 最大 CPU 使用
```

## 7.9 故障排除

### 7.9.1 常见问题和解决方案

#### 7.9.1.1 Pod 状态异常

```bash
# 检查 Pod 状态
kubectl get pods

# 常见状态说明
# Pending: 等待调度
# Running: 正在运行
# Succeeded: 成功完成
# Failed: 执行失败
# CrashLoopBackOff: 循环崩溃

# 查看详细信息
kubectl describe pod <pod-name>

# 查看日志
kubectl logs <pod-name>
# 查看前一个容器的日志
kubectl logs <pod-name> --previous
```

#### 7.9.1.2 Service 连接问题

```bash
# 检查 Service 状态
kubectl get svc

# 检查端点
kubectl get endpoints <service-name>

# 测试 Service 连通性
kubectl run test-pod --image=busybox --rm -it -- /bin/sh
# 在测试 Pod 中执行
wget -qO- http://<service-name>:<port>
```

#### 7.9.1.3 镜像拉取失败

```bash
# 检查镜像拉取策略
# imagePullPolicy: Always | IfNotPresent | Never

# 检查镜像仓库访问
kubectl describe pod <pod-name>

# 配置镜像拉取密钥（如需要）
kubectl create secret docker-registry myregistrykey \
  --docker-server=<registry-server> \
  --docker-username=<username> \
  --docker-password=<password>
```

### 7.9.2 调试技巧

```bash
# 实时监控资源
watch kubectl get pods,svc

# 查看集群事件
kubectl get events --sort-by=.metadata.creationTimestamp

# 进入容器调试
kubectl exec -it <pod-name> -- /bin/bash

# 端口转发调试
kubectl port-forward pod/<pod-name> 8080:80

# 复制文件到/从容器
kubectl cp <pod-name>:/path/to/file ./local-file
kubectl cp ./local-file <pod-name>:/path/to/file
```

## 7.10 学习总结

### 7.10.1 关键要点

1. **YAML 结构**：掌握 apiVersion、kind、metadata、spec 四大核心字段
2. **Pod 管理**：理解 Pod 生命周期和容器配置
3. **Service 网络**：掌握不同 Service 类型的使用场景
4. **标签选择器**：理解标签在资源关联中的重要作用
5. **资源配置**：合理设置资源请求和限制

### 7.10.2 进阶学习建议

1. **探索其他工作负载**：Deployment、ReplicaSet、DaemonSet
2. **学习配置管理**：ConfigMap、Secret
3. **掌握存储概念**：Volume、PersistentVolume
4. **了解网络进阶**：Ingress、NetworkPolicy
5. **实践 GitOps**：使用 Git 管理 Kubernetes 配置

### 7.10.3 下一步学习

- [08-ReplicaSets-with-YAML](../08-ReplicaSets-with-YAML/README.md)
- [09-Deployments-with-YAML](../09-Deployments-with-YAML/README.md)
- [10-Services-with-YAML](../10-Services-with-YAML/README.md)

## 7.11 API 对象参考

### 7.11.1 官方文档

- **Pod API 参考：** <https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/>
- **Service API 参考：** <https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/>
- **完整 API 参考：** <https://kubernetes.io/docs/reference/kubernetes-api/>

### 7.11.2 实用工具

- **Kubernetes 文档：** <https://kubernetes.io/docs/>
- **YAML 验证器：** <https://kubeyaml.com/>
- **资源配置生成器：** <https://k8syaml.com/>
