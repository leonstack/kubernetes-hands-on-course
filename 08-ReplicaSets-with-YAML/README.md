# 8. 使用 YAML 管理 ReplicaSets

## 8.0 目录

- [8. 使用 YAML 管理 ReplicaSets](#8-使用-yaml-管理-replicasets)
  - [8.0 目录](#80-目录)
  - [8.1 项目概述](#81-项目概述)
    - [8.1.1 学习目标](#811-学习目标)
    - [8.1.2 应用场景](#812-应用场景)
  - [8.2 前置条件](#82-前置条件)
    - [8.2.1 环境要求](#821-环境要求)
    - [8.2.2 验证环境](#822-验证环境)
  - [8.3 ReplicaSet 基础概念](#83-replicaset-基础概念)
    - [8.3.1 ReplicaSet 简介](#831-replicaset-简介)
    - [8.3.2 核心特性](#832-核心特性)
    - [8.3.3 ReplicaSet vs Pod](#833-replicaset-vs-pod)
  - [8.4 创建 ReplicaSet 定义](#84-创建-replicaset-定义)
    - [8.4.1 ReplicaSet YAML 结构](#841-replicaset-yaml-结构)
    - [8.4.2 关键字段详解](#842-关键字段详解)
    - [8.4.3 标签选择器机制](#843-标签选择器机制)
  - [8.5 创建和管理 ReplicaSet](#85-创建和管理-replicaset)
    - [8.5.1 创建 ReplicaSet](#851-创建-replicaset)
    - [8.5.2 ReplicaSet 状态说明](#852-replicaset-状态说明)
    - [8.5.3 测试自动恢复机制](#853-测试自动恢复机制)
    - [8.5.4 扩缩容操作](#854-扩缩容操作)
    - [8.5.5 常用管理命令](#855-常用管理命令)
    - [8.5.6 故障排除技巧](#856-故障排除技巧)
  - [8.6 创建 NodePort Service](#86-创建-nodeport-service)
    - [8.6.1 Service 基础概念](#861-service-基础概念)
    - [8.6.2 Service 类型对比](#862-service-类型对比)
    - [8.6.3 NodePort Service 配置](#863-nodeport-service-配置)
    - [8.6.4 Service 字段详解](#864-service-字段详解)
    - [8.6.5 创建和测试 Service](#865-创建和测试-service)
    - [8.6.6 网络访问测试](#866-网络访问测试)
    - [8.6.7 负载均衡验证](#867-负载均衡验证)
    - [8.6.8 网络流量路径](#868-网络流量路径)
  - [8.7 最佳实践](#87-最佳实践)
    - [8.7.1 ReplicaSet 配置最佳实践](#871-replicaset-配置最佳实践)
      - [1. 标签和选择器](#1-标签和选择器)
      - [2. 资源配置](#2-资源配置)
      - [3. 健康检查](#3-健康检查)
    - [8.7.2 安全最佳实践](#872-安全最佳实践)
      - [1. 安全上下文](#1-安全上下文)
      - [2. 镜像安全](#2-镜像安全)
    - [8.7.3 运维最佳实践](#873-运维最佳实践)
      - [1. 监控和日志](#1-监控和日志)
      - [2. 备份和版本控制](#2-备份和版本控制)
  - [8.8 故障排除](#88-故障排除)
    - [8.8.1 常见问题和解决方案](#881-常见问题和解决方案)
      - [1. Pod 无法启动](#1-pod-无法启动)
      - [2. ReplicaSet 副本数不匹配](#2-replicaset-副本数不匹配)
      - [3. Service 无法访问](#3-service-无法访问)
      - [4. 网络连接问题](#4-网络连接问题)
    - [8.8.2 调试技巧](#882-调试技巧)
      - [1. 实时监控](#1-实时监控)
      - [2. 详细日志](#2-详细日志)
  - [8.9 清理资源](#89-清理资源)
  - [8.10 学习总结](#810-学习总结)
    - [8.10.1 关键要点](#8101-关键要点)
    - [8.10.2 进阶学习建议](#8102-进阶学习建议)
    - [8.10.3 下一步学习](#8103-下一步学习)
  - [8.11 API 参考和实用工具](#811-api-参考和实用工具)
    - [8.11.1 官方文档](#8111-官方文档)
    - [8.11.2 实用工具](#8112-实用工具)

## 8.1 项目概述

### 8.1.1 学习目标

- 理解 ReplicaSet 的核心概念和工作原理
- 掌握使用 YAML 文件创建和管理 ReplicaSet
- 学习 ReplicaSet 的标签选择器机制
- 了解 ReplicaSet 与 Pod 的关系
- 实践 ReplicaSet 的扩缩容和故障恢复
- 配置 Service 为 ReplicaSet 提供网络访问

### 8.1.2 应用场景

- **高可用性应用**：确保应用始终有指定数量的副本运行
- **负载分担**：通过多个 Pod 副本分担应用负载
- **故障恢复**：自动替换失败的 Pod 实例
- **水平扩展**：根据需求动态调整 Pod 副本数量
- **滚动更新基础**：为 Deployment 提供底层支持

## 8.2 前置条件

### 8.2.1 环境要求

- Kubernetes 集群（v1.18+）
- kubectl 命令行工具
- 基础的 YAML 语法知识
- 了解 Pod 和 Service 概念

### 8.2.2 验证环境

```bash
# 检查集群连接
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查当前命名空间
kubectl config get-contexts
```

## 8.3 ReplicaSet 基础概念

### 8.3.1 ReplicaSet 简介

ReplicaSet 是 Kubernetes 中用于确保指定数量的 Pod 副本始终运行的控制器。它通过标签选择器来管理 Pod，并在 Pod 失败时自动创建新的替代实例。

### 8.3.2 核心特性

- **副本保证**：维护指定数量的 Pod 副本
- **标签选择**：通过标签选择器管理 Pod
- **自动恢复**：自动替换失败的 Pod
- **水平扩展**：支持动态调整副本数量

### 8.3.3 ReplicaSet vs Pod

| 特性 | Pod | ReplicaSet |
|------|-----|------------|
| 生命周期 | 单一实例 | 管理多个 Pod |
| 故障恢复 | 手动重启 | 自动恢复 |
| 扩展性 | 不支持 | 支持扩缩容 |
| 使用场景 | 测试/调试 | 生产环境 |

## 8.4 创建 ReplicaSet 定义

### 8.4.1 ReplicaSet YAML 结构

ReplicaSet 的 YAML 配置包含四个主要部分：

```yaml
apiVersion: apps/v1      # API 版本
kind: ReplicaSet         # 资源类型
metadata:                # 元数据
  name: myapp2-rs
  labels:
    app: myapp2
    tier: frontend
    version: v2.0
  annotations:
    description: "Web application ReplicaSet"
spec:                    # 规格说明
  replicas: 3            # 期望的 Pod 副本数量
  selector:              # 标签选择器
    matchLabels:
      app: myapp2
  template:              # Pod 模板
    metadata:
      labels:
        app: myapp2      # 必须匹配 selector
        version: v2.0
    spec:
      containers:
      - name: myapp2-container
        image: grissomsh/kubenginx:2.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
          name: http
          protocol: TCP
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 8.4.2 关键字段详解

| 字段 | 描述 | 重要性 |
|------|------|--------|
| `replicas` | 期望的 Pod 副本数量 | 核心 |
| `selector.matchLabels` | 用于选择管理的 Pod | 核心 |
| `template.metadata.labels` | Pod 模板的标签，必须匹配 selector | 核心 |
| `template.spec` | Pod 的具体配置 | 核心 |
| `resources` | 资源请求和限制 | 推荐 |
| `livenessProbe` | 存活性探针 | 推荐 |
| `readinessProbe` | 就绪性探针 | 推荐 |

### 8.4.3 标签选择器机制

```yaml
# ReplicaSet 选择器
selector:
  matchLabels:
    app: myapp2
    
# Pod 模板标签（必须包含选择器中的所有标签）
template:
  metadata:
    labels:
      app: myapp2        # 匹配选择器
      version: v2.0      # 额外标签
      tier: frontend     # 额外标签
```

**重要提示**：

- Pod 模板的标签必须包含选择器中的所有标签
- Pod 可以有额外的标签，但不能缺少选择器要求的标签
- 标签选择器一旦创建就不能修改

## 8.5 创建和管理 ReplicaSet

### 8.5.1 创建 ReplicaSet

```bash
# 应用 ReplicaSet 配置
kubectl apply -f kube-manifests/02-replicaset-definition.yml

# 查看 ReplicaSet 状态
kubectl get replicasets
kubectl get rs  # 简写形式

# 查看详细信息
kubectl describe replicaset myapp2-rs

# 查看 ReplicaSet 管理的 Pod
kubectl get pods -l app=myapp2
kubectl get pods --show-labels
```

### 8.5.2 ReplicaSet 状态说明

```bash
# 输出示例
NAME        DESIRED   CURRENT   READY   AGE
myapp2-rs   3         3         3       2m
```

| 字段 | 说明 |
|------|------|
| DESIRED | 期望的副本数量 |
| CURRENT | 当前存在的副本数量 |
| READY | 就绪的副本数量 |
| AGE | ReplicaSet 创建时间 |

### 8.5.3 测试自动恢复机制

ReplicaSet 的核心功能是维护指定数量的 Pod 副本。当 Pod 被删除时，ReplicaSet 会自动创建新的 Pod 来替代。

```bash
# 查看当前 Pod
kubectl get pods -l app=myapp2

# 删除一个 Pod（模拟故障）
kubectl delete pod <Pod-Name>

# 立即查看 Pod 状态（观察自动恢复）
kubectl get pods -l app=myapp2 -w

# 查看 ReplicaSet 事件
kubectl describe rs myapp2-rs
```

### 8.5.4 扩缩容操作

```bash
# 方法1：使用 kubectl scale 命令
kubectl scale replicaset myapp2-rs --replicas=5

# 方法2：编辑 ReplicaSet 配置
kubectl edit replicaset myapp2-rs

# 方法3：更新 YAML 文件后重新应用
# 修改 02-replicaset-definition.yml 中的 replicas 值
kubectl apply -f kube-manifests/02-replicaset-definition.yml

# 验证扩缩容结果
kubectl get rs myapp2-rs
kubectl get pods -l app=myapp2
```

### 8.5.5 常用管理命令

```bash
# 查看 ReplicaSet 列表
kubectl get rs
kubectl get rs -o wide
kubectl get rs -o yaml

# 查看特定 ReplicaSet
kubectl get rs myapp2-rs
kubectl describe rs myapp2-rs

# 查看 ReplicaSet 管理的 Pod
kubectl get pods -l app=myapp2
kubectl get pods --selector=app=myapp2

# 查看 ReplicaSet 事件
kubectl get events --field-selector involvedObject.name=myapp2-rs

# 实时监控 Pod 状态
kubectl get pods -l app=myapp2 -w
```

### 8.5.6 故障排除技巧

```bash
# 检查 ReplicaSet 状态
kubectl describe rs myapp2-rs

# 检查 Pod 状态
kubectl describe pod <pod-name>

# 查看 Pod 日志
kubectl logs <pod-name>

# 查看所有相关事件
kubectl get events --sort-by=.metadata.creationTimestamp

# 检查标签选择器
kubectl get pods --show-labels
kubectl get rs myapp2-rs -o yaml | grep -A 5 selector
```

## 8.6 创建 NodePort Service

### 8.6.1 Service 基础概念

Service 为 ReplicaSet 管理的 Pod 提供稳定的网络访问入口，实现负载均衡和服务发现。

### 8.6.2 Service 类型对比

| 类型 | 访问方式 | 使用场景 |
|------|----------|----------|
| ClusterIP | 集群内部 | 内部服务通信 |
| NodePort | 节点端口 | 外部访问（开发/测试） |
| LoadBalancer | 云负载均衡器 | 生产环境外部访问 |
| ExternalName | DNS 映射 | 外部服务代理 |

### 8.6.3 NodePort Service 配置

```yaml
apiVersion: v1
kind: Service
metadata:
  name: replicaset-nodeport-service
  labels:
    app: myapp2
    service-type: nodeport
  annotations:
    description: "NodePort service for ReplicaSet"
spec:
  type: NodePort
  selector:
    app: myapp2              # 必须匹配 ReplicaSet Pod 标签
  ports:
  - name: http
    port: 80                 # Service 端口
    targetPort: 80           # Pod 端口
    nodePort: 31232          # 节点端口（30000-32767）
    protocol: TCP
  sessionAffinity: None      # 会话亲和性
  externalTrafficPolicy: Cluster  # 外部流量策略
```

### 8.6.4 Service 字段详解

| 字段 | 描述 | 默认值 |
|------|------|--------|
| `type` | Service 类型 | ClusterIP |
| `selector` | Pod 标签选择器 | 无 |
| `port` | Service 暴露的端口 | 无 |
| `targetPort` | Pod 容器端口 | port 值 |
| `nodePort` | 节点端口 | 自动分配 |
| `sessionAffinity` | 会话亲和性 | None |
| `externalTrafficPolicy` | 外部流量策略 | Cluster |

### 8.6.5 创建和测试 Service

```bash
# 创建 NodePort Service
kubectl apply -f kube-manifests/03-replicaset-nodeport-servie.yml

# 查看 Service 状态
kubectl get services
kubectl get svc  # 简写形式

# 查看 Service 详细信息
kubectl describe service replicaset-nodeport-service

# 查看 Endpoints（Service 后端 Pod）
kubectl get endpoints replicaset-nodeport-service
```

### 8.6.6 网络访问测试

```bash
# 获取节点信息
kubectl get nodes -o wide

# 获取 Service 信息
kubectl get svc replicaset-nodeport-service

# 方法1：通过节点 IP + NodePort 访问
curl http://<Node-IP>:31232

# 方法2：通过集群内部访问
kubectl run test-pod --image=busybox --rm -i --restart=Never -- \
  wget -qO- http://replicaset-nodeport-service:80

# 方法3：端口转发测试
kubectl port-forward service/replicaset-nodeport-service 8080:80
# 然后在浏览器访问 http://localhost:8080
```

### 8.6.7 负载均衡验证

```bash
# 多次访问验证负载均衡
for i in {1..10}; do
  curl -s http://<Node-IP>:31232 | grep -i hostname
done

# 查看 Pod 访问日志
kubectl logs -l app=myapp2 --tail=20
```

### 8.6.8 网络流量路径

```text
外部客户端 → 节点IP:31232 → Service → Pod:80
     ↓
1. 客户端访问任意节点的 31232 端口
2. kube-proxy 将流量转发到 Service
3. Service 根据标签选择器找到后端 Pod
4. 流量负载均衡到健康的 Pod
```

## 8.7 最佳实践

### 8.7.1 ReplicaSet 配置最佳实践

#### 1. 标签和选择器

```yaml
# 推荐：使用多个标签提高选择精度
selector:
  matchLabels:
    app: myapp2
    version: v2.0
    tier: frontend

# 避免：使用过于宽泛的标签
selector:
  matchLabels:
    env: production  # 可能匹配过多 Pod
```

#### 2. 资源配置

```yaml
resources:
  requests:
    memory: "64Mi"     # 保证最小资源
    cpu: "250m"
  limits:
    memory: "128Mi"    # 防止资源滥用
    cpu: "500m"
```

#### 3. 健康检查

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 80
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 80
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### 8.7.2 安全最佳实践

#### 1. 安全上下文

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
```

#### 2. 镜像安全

```yaml
containers:
- name: myapp2-container
  image: grissomsh/kubenginx:2.0.0
  imagePullPolicy: IfNotPresent  # 避免 Always
```

### 8.7.3 运维最佳实践

#### 1. 监控和日志

```bash
# 设置资源监控
kubectl top pods -l app=myapp2

# 配置日志收集
kubectl logs -l app=myapp2 --tail=100 -f

# 监控 ReplicaSet 事件
kubectl get events --field-selector involvedObject.name=myapp2-rs
```

#### 2. 备份和版本控制

```bash
# 导出当前配置
kubectl get rs myapp2-rs -o yaml > backup-replicaset.yaml

# 版本控制 YAML 文件
git add kube-manifests/
git commit -m "Update ReplicaSet configuration"
```

## 8.8 故障排除

### 8.8.1 常见问题和解决方案

#### 1. Pod 无法启动

```bash
# 问题诊断
kubectl describe rs myapp2-rs
kubectl describe pod <pod-name>
kubectl logs <pod-name>

# 常见原因
# - 镜像拉取失败
# - 资源不足
# - 配置错误
# - 存储挂载问题
```

#### 2. ReplicaSet 副本数不匹配

```bash
# 检查 ReplicaSet 状态
kubectl get rs myapp2-rs
kubectl describe rs myapp2-rs

# 可能原因
# - 节点资源不足
# - 标签选择器冲突
# - Pod 调度失败
# - 镜像问题
```

#### 3. Service 无法访问

```bash
# 检查 Service 和 Endpoints
kubectl get svc replicaset-nodeport-service
kubectl get endpoints replicaset-nodeport-service
kubectl describe svc replicaset-nodeport-service

# 验证标签匹配
kubectl get pods --show-labels
kubectl get svc replicaset-nodeport-service -o yaml | grep -A 5 selector
```

#### 4. 网络连接问题

```bash
# 测试集群内部连接
kubectl run debug-pod --image=busybox --rm -i --restart=Never -- \
  nslookup replicaset-nodeport-service

# 测试端口连通性
kubectl run debug-pod --image=busybox --rm -i --restart=Never -- \
  telnet replicaset-nodeport-service 80

# 检查防火墙和网络策略
kubectl get networkpolicies
```

### 8.8.2 调试技巧

#### 1. 实时监控

```bash
# 监控 Pod 状态变化
kubectl get pods -l app=myapp2 -w

# 监控 ReplicaSet 事件
kubectl get events -w --field-selector involvedObject.name=myapp2-rs

# 监控资源使用
watch kubectl top pods -l app=myapp2
```

#### 2. 详细日志

```bash
# 查看所有容器日志
kubectl logs -l app=myapp2 --all-containers=true

# 查看历史日志
kubectl logs <pod-name> --previous

# 实时日志流
kubectl logs -l app=myapp2 -f --tail=50
```

## 8.9 清理资源

```bash
# 删除 Service
kubectl delete service replicaset-nodeport-service

# 删除 ReplicaSet（会同时删除管理的 Pod）
kubectl delete replicaset myapp2-rs

# 或者删除所有相关资源
kubectl delete -f kube-manifests/

# 验证清理结果
kubectl get rs,svc,pods -l app=myapp2
```

## 8.10 学习总结

### 8.10.1 关键要点

1. **ReplicaSet 核心功能**
   - 维护指定数量的 Pod 副本
   - 通过标签选择器管理 Pod
   - 提供自动故障恢复能力
   - 支持水平扩缩容

2. **标签选择器机制**
   - Pod 模板标签必须匹配选择器
   - 选择器创建后不可修改
   - 使用精确的标签提高管理精度

3. **Service 集成**
   - 为 ReplicaSet 提供稳定的网络入口
   - 实现负载均衡和服务发现
   - 支持多种访问方式

4. **运维管理**
   - 使用声明式配置管理
   - 实施监控和日志收集
   - 遵循安全最佳实践

### 8.10.2 进阶学习建议

1. **深入学习 Deployment**
   - Deployment 是 ReplicaSet 的高级抽象
   - 提供滚动更新和回滚功能
   - 生产环境推荐使用 Deployment

2. **探索其他控制器**
   - DaemonSet：每个节点运行一个 Pod
   - StatefulSet：有状态应用管理
   - Job/CronJob：批处理任务

3. **网络和存储**
   - Ingress 控制器
   - 持久化存储
   - 网络策略

4. **监控和可观测性**
   - Prometheus + Grafana
   - 日志聚合系统
   - 分布式追踪

### 8.10.3 下一步学习

- **09-Deployments-with-YAML**：学习 Deployment 的高级功能
- **10-Services-with-YAML**：深入学习 Service 的各种类型
- **ConfigMaps 和 Secrets**：配置和密钥管理
- **Persistent Volumes**：持久化存储解决方案

## 8.11 API 参考和实用工具

### 8.11.1 官方文档

- **ReplicaSet API**: <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#replicaset-v1-apps>
- **Service API**: <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#service-v1-core>
- **Pod API**: <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pod-v1-core>

### 8.11.2 实用工具

- **kubectl 备忘单**: <https://kubernetes.io/docs/reference/kubectl/cheatsheet/>
- **YAML 验证器**: <https://kubeyaml.com/>
- **Kubernetes 文档**: <https://kubernetes.io/docs/>
- **官方教程**: <https://kubernetes.io/docs/tutorials/>