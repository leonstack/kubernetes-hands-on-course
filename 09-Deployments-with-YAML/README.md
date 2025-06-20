# 9. Deployments with YAML

## 9.0 目录

- [9. Deployments with YAML](#9-deployments-with-yaml)
  - [9.0 目录](#90-目录)
  - [9.1 项目概述](#91-项目概述)
  - [9.2 学习目标](#92-学习目标)
  - [9.3 应用场景](#93-应用场景)
  - [9.4 前置条件](#94-前置条件)
    - [9.4.1 环境要求](#941-环境要求)
    - [9.4.2 验证环境](#942-验证环境)
  - [9.5 Deployment 基础概念](#95-deployment-基础概念)
    - [9.5.1 Deployment 简介](#951-deployment-简介)
    - [9.5.2 核心特性](#952-核心特性)
    - [9.5.3 Deployment vs ReplicaSet vs Pod](#953-deployment-vs-replicaset-vs-pod)
  - [9.6 从 ReplicaSet 模板创建 Deployment](#96-从-replicaset-模板创建-deployment)
    - [9.6.1 模板转换步骤](#961-模板转换步骤)
    - [9.6.2 Deployment YAML 结构](#962-deployment-yaml-结构)
    - [9.6.3 关键字段详解](#963-关键字段详解)
  - [9.7 创建和管理 Deployment](#97-创建和管理-deployment)
    - [9.7.1 创建 Deployment](#971-创建-deployment)
    - [9.7.2 Deployment 状态监控](#972-deployment-状态监控)
    - [9.7.3 扩缩容操作](#973-扩缩容操作)
  - [9.8 创建和配置 Service](#98-创建和配置-service)
    - [9.8.1 Service 基础概念](#981-service-基础概念)
    - [9.8.2 创建 NodePort Service](#982-创建-nodeport-service)
    - [9.8.3 网络访问测试](#983-网络访问测试)
    - [9.8.4 负载均衡验证](#984-负载均衡验证)
  - [9.9 Deployment 高级功能](#99-deployment-高级功能)
    - [9.9.1 滚动更新](#991-滚动更新)
    - [9.9.2 回滚操作](#992-回滚操作)
    - [9.9.3 暂停和恢复部署](#993-暂停和恢复部署)
    - [9.9.4 更新策略配置](#994-更新策略配置)
  - [9.10 监控和调试](#910-监控和调试)
    - [9.10.1 资源监控](#9101-资源监控)
    - [9.10.2 日志查看](#9102-日志查看)
    - [9.10.3 故障排除](#9103-故障排除)
  - [9.11 最佳实践](#911-最佳实践)
    - [9.11.1 Deployment 配置最佳实践](#9111-deployment-配置最佳实践)
    - [9.11.2 运维最佳实践](#9112-运维最佳实践)
    - [9.11.3 生产环境建议](#9113-生产环境建议)
  - [9.12 故障排除指南](#912-故障排除指南)
    - [9.12.1 常见问题](#9121-常见问题)
    - [9.12.2 调试命令](#9122-调试命令)
  - [9.13 清理资源](#913-清理资源)
  - [9.14 学习总结](#914-学习总结)
    - [9.14.1 关键要点](#9141-关键要点)
    - [9.14.2 进阶学习建议](#9142-进阶学习建议)
    - [9.14.3 下一步学习](#9143-下一步学习)
  - [9.15 API 参考和实用工具](#915-api-参考和实用工具)
    - [9.15.1 API 参考](#9151-api-参考)
    - [9.15.2 实用工具](#9152-实用工具)

## 9.1 项目概述

本教程将深入学习如何使用 YAML 文件创建和管理 Kubernetes Deployment 资源。Deployment 是 Kubernetes 中最重要的工作负载资源之一，提供了声明式的应用部署和管理能力。

## 9.2 学习目标

- 理解 Deployment 的核心概念和工作原理
- 掌握 Deployment YAML 文件的编写和配置
- 学习 Deployment 的创建、更新、扩缩容和回滚操作
- 了解 Deployment 与 ReplicaSet、Pod 的关系
- 掌握 Service 与 Deployment 的集成使用
- 学习 Deployment 的最佳实践和故障排除

## 9.3 应用场景

- 无状态应用的部署和管理
- 应用的滚动更新和版本控制
- 应用的自动扩缩容
- 应用的高可用性保障
- 生产环境的应用发布策略

## 9.4 前置条件

### 9.4.1 环境要求

- Kubernetes 集群（版本 1.18+）
- kubectl 命令行工具
- 基本的 Kubernetes 概念理解
- 熟悉 YAML 语法
- 了解 Pod 和 ReplicaSet 概念

### 9.4.2 验证环境

```bash
# 检查 kubectl 版本
kubectl version --client

# 检查集群连接
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查当前命名空间
kubectl config current-context
```

## 9.5 Deployment 基础概念

### 9.5.1 Deployment 简介

Deployment 是 Kubernetes 中用于管理无状态应用的高级控制器，它提供了声明式的应用部署和管理能力。Deployment 通过管理 ReplicaSet 来确保指定数量的 Pod 副本始终运行。

### 9.5.2 核心特性

- **声明式管理**：通过 YAML 文件描述期望状态
- **滚动更新**：支持零停机时间的应用更新
- **版本控制**：维护部署历史，支持回滚操作
- **自动扩缩容**：支持手动和自动的副本数调整
- **自愈能力**：自动替换失败的 Pod
- **暂停和恢复**：支持部署过程的暂停和恢复

### 9.5.3 Deployment vs ReplicaSet vs Pod

| 特性 | Pod | ReplicaSet | Deployment |
|------|-----|------------|------------|
| 管理级别 | 单个容器组 | Pod 副本集 | ReplicaSet 管理器 |
| 更新策略 | 手动重建 | 手动重建 | 滚动更新 |
| 版本控制 | 无 | 无 | 支持 |
| 回滚能力 | 无 | 无 | 支持 |
| 使用场景 | 测试/调试 | 简单副本管理 | 生产环境应用 |
| 推荐程度 | 低 | 中 | 高 |

## 9.6 从 ReplicaSet 模板创建 Deployment

### 9.6.1 模板转换步骤

基于之前的 ReplicaSet 模板，我们需要进行以下修改来创建 Deployment：

1. **更改资源类型**：将 `kind: ReplicaSet` 改为 `kind: Deployment`
2. **更新镜像版本**：将容器镜像版本更新到 `3.0.0`
3. **修改服务端口**：将 NodePort 服务端口更新为 `31233`
4. **更新资源名称**：将所有名称改为 Deployment 相关
5. **修改标签选择器**：将所有标签和选择器改为 `myapp3`

### 9.6.2 Deployment YAML 结构

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp3-deployment
  labels:
    app: myapp3
    version: v3.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp3
  template:
    metadata:
      labels:
        app: myapp3
        version: v3.0
    spec:
      containers:
      - name: myapp3-container
        image: grissomsh/kubenginx:3.0.0
        ports:
        - containerPort: 80
```

### 9.6.3 关键字段详解

| 字段 | 描述 | 示例值 |
|------|------|--------|
| `apiVersion` | API 版本 | `apps/v1` |
| `kind` | 资源类型 | `Deployment` |
| `metadata.name` | Deployment 名称 | `myapp3-deployment` |
| `spec.replicas` | 期望的 Pod 副本数 | `3` |
| `spec.selector` | Pod 选择器 | `matchLabels: {app: myapp3}` |
| `spec.template` | Pod 模板 | 包含 Pod 的完整定义 |
| `spec.strategy` | 更新策略 | `RollingUpdate`（默认） |

## 9.7 创建和管理 Deployment

### 9.7.1 创建 Deployment

```bash
# 应用 Deployment 配置
kubectl apply -f 02-deployment-definition.yml

# 查看 Deployment 状态
kubectl get deployments
kubectl get deploy -o wide

# 查看 ReplicaSet（由 Deployment 创建）
kubectl get replicasets
kubectl get rs -o wide

# 查看 Pod（由 ReplicaSet 管理）
kubectl get pods
kubectl get po -o wide

# 查看 Deployment 详细信息
kubectl describe deployment myapp3-deployment
```

### 9.7.2 Deployment 状态监控

```bash
# 实时监控 Deployment 状态
kubectl get deployment myapp3-deployment -w

# 查看 Deployment 事件
kubectl get events --sort-by=.metadata.creationTimestamp

# 查看 Deployment 的 ReplicaSet 历史
kubectl get rs -l app=myapp3

# 查看 Pod 分布情况
kubectl get pods -l app=myapp3 -o wide
```

### 9.7.3 扩缩容操作

```bash
# 扩容到 5 个副本
kubectl scale deployment myapp3-deployment --replicas=5

# 查看扩容过程
kubectl get pods -l app=myapp3 -w

# 缩容到 2 个副本
kubectl scale deployment myapp3-deployment --replicas=2

# 恢复到原始副本数
kubectl scale deployment myapp3-deployment --replicas=3
```

## 9.8 创建和配置 Service

### 9.8.1 Service 基础概念

Service 为 Deployment 管理的 Pod 提供稳定的网络访问入口，支持负载均衡和服务发现。

### 9.8.2 创建 NodePort Service

```bash
# 应用 Service 配置
kubectl apply -f 03-deployment-nodeport-service.yml

# 查看 Service 状态
kubectl get services
kubectl get svc -o wide

# 查看 Service 详细信息
kubectl describe service deployment-nodeport-service

# 查看 Endpoints
kubectl get endpoints deployment-nodeport-service
```

### 9.8.3 网络访问测试

```bash
# 获取节点公网 IP
kubectl get nodes -o wide

# 获取 Service 信息
kubectl get svc deployment-nodeport-service

# 访问应用（替换为实际的节点 IP）
curl http://<Worker-Node-Public-IP>:31233

# 或在浏览器中访问
http://<Worker-Node-Public-IP>:31233
```

### 9.8.4 负载均衡验证

```bash
# 多次请求验证负载均衡
for i in {1..10}; do
  curl -s http://<Worker-Node-Public-IP>:31233 | grep -i hostname
  sleep 1
done

# 集群内部访问测试
kubectl run test-pod --image=busybox --rm -i --restart=Never -- \
  wget -qO- http://deployment-nodeport-service:80
```

## 9.9 Deployment 高级功能

### 9.9.1 滚动更新

```bash
# 更新镜像版本（触发滚动更新）
kubectl set image deployment/myapp3-deployment myapp3-container=grissomsh/kubenginx:4.0.0

# 查看滚动更新状态
kubectl rollout status deployment/myapp3-deployment

# 查看滚动更新历史
kubectl rollout history deployment/myapp3-deployment

# 查看特定版本的详细信息
kubectl rollout history deployment/myapp3-deployment --revision=2
```

### 9.9.2 回滚操作

```bash
# 回滚到上一个版本
kubectl rollout undo deployment/myapp3-deployment

# 回滚到特定版本
kubectl rollout undo deployment/myapp3-deployment --to-revision=1

# 验证回滚结果
kubectl get pods -l app=myapp3
kubectl describe deployment myapp3-deployment
```

### 9.9.3 暂停和恢复部署

```bash
# 暂停部署（停止滚动更新）
kubectl rollout pause deployment/myapp3-deployment

# 进行多个更改
kubectl set image deployment/myapp3-deployment myapp3-container=grissomsh/kubenginx:5.0.0
kubectl set resources deployment/myapp3-deployment -c=myapp3-container --limits=cpu=200m,memory=256Mi

# 恢复部署（应用所有更改）
kubectl rollout resume deployment/myapp3-deployment

# 监控恢复过程
kubectl rollout status deployment/myapp3-deployment
```

### 9.9.4 更新策略配置

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1        # 最大不可用 Pod 数
      maxSurge: 1             # 最大超出期望副本数的 Pod 数
  revisionHistoryLimit: 10   # 保留的历史版本数
```

## 9.10 监控和调试

### 9.10.1 资源监控

```bash
# 查看 Deployment 状态
kubectl get deployment myapp3-deployment -o yaml

# 查看 Pod 资源使用情况（需要 metrics-server）
kubectl top pods -l app=myapp3

# 查看节点资源使用情况
kubectl top nodes

# 实时监控 Pod 状态变化
kubectl get pods -l app=myapp3 -w
```

### 9.10.2 日志查看

```bash
# 查看所有 Pod 的日志
kubectl logs -l app=myapp3

# 查看特定 Pod 的日志
kubectl logs <pod-name>

# 实时跟踪日志
kubectl logs -f -l app=myapp3

# 查看之前容器的日志
kubectl logs <pod-name> --previous
```

### 9.10.3 故障排除

```bash
# 查看 Deployment 事件
kubectl describe deployment myapp3-deployment

# 查看 Pod 详细信息
kubectl describe pod <pod-name>

# 进入 Pod 进行调试
kubectl exec -it <pod-name> -- /bin/bash

# 查看集群事件
kubectl get events --sort-by=.metadata.creationTimestamp

# 检查资源配额
kubectl describe resourcequota
```

## 9.11 最佳实践

### 9.11.1 Deployment 配置最佳实践

1. **资源管理**

   ```yaml
   resources:
     requests:
       memory: "128Mi"
       cpu: "100m"
     limits:
       memory: "256Mi"
       cpu: "200m"
   ```

2. **健康检查**

   ```yaml
   livenessProbe:
     httpGet:
       path: /health
       port: 80
     initialDelaySeconds: 30
     periodSeconds: 10
   readinessProbe:
     httpGet:
       path: /ready
       port: 80
     initialDelaySeconds: 5
     periodSeconds: 5
   ```

3. **安全配置**

   ```yaml
   securityContext:
     runAsNonRoot: true
     runAsUser: 1000
     readOnlyRootFilesystem: true
   ```

### 9.11.2 运维最佳实践

- **版本管理**：使用语义化版本标签
- **渐进式部署**：先在测试环境验证
- **监控告警**：设置关键指标监控
- **备份策略**：定期备份配置文件
- **文档维护**：保持部署文档更新

### 9.11.3 生产环境建议

- 设置合适的资源请求和限制
- 配置健康检查探针
- 使用多副本确保高可用
- 实施适当的更新策略
- 配置 HPA 自动扩缩容
- 使用 PodDisruptionBudget

## 9.12 故障排除指南

### 9.12.1 常见问题

1. **Pod 无法启动**
   - 检查镜像是否存在
   - 验证资源配额
   - 查看 Pod 事件和日志

2. **滚动更新失败**
   - 检查新镜像的健康状态
   - 验证健康检查配置
   - 查看更新策略设置

3. **Service 无法访问**
   - 验证标签选择器匹配
   - 检查端口配置
   - 确认 Endpoints 状态

### 9.12.2 调试命令

```bash
# 检查 Deployment 状态
kubectl get deployment myapp3-deployment -o wide

# 查看 ReplicaSet 状态
kubectl get rs -l app=myapp3

# 检查 Pod 状态
kubectl get pods -l app=myapp3 -o wide

# 查看详细事件
kubectl describe deployment myapp3-deployment

# 检查网络连接
kubectl run debug-pod --image=busybox --rm -i --restart=Never -- nslookup deployment-nodeport-service
```

## 9.13 清理资源

```bash
# 删除 Service
kubectl delete service deployment-nodeport-service

# 删除 Deployment（会自动删除 ReplicaSet 和 Pod）
kubectl delete deployment myapp3-deployment

# 验证清理结果
kubectl get all -l app=myapp3

# 强制删除（如果需要）
kubectl delete deployment myapp3-deployment --force --grace-period=0
```

## 9.14 学习总结

### 9.14.1 关键要点

1. **Deployment 优势**：提供声明式管理、滚动更新、版本控制等高级功能
2. **层次关系**：Deployment → ReplicaSet → Pod 的管理层次
3. **更新策略**：支持滚动更新和重建更新两种策略
4. **版本控制**：自动维护部署历史，支持快速回滚
5. **扩缩容**：支持手动和自动的副本数调整

### 9.14.2 进阶学习建议

- 学习 HorizontalPodAutoscaler (HPA) 自动扩缩容
- 了解 VerticalPodAutoscaler (VPA) 垂直扩缩容
- 掌握 PodDisruptionBudget 中断预算
- 学习 StatefulSet 有状态应用管理
- 了解 DaemonSet 守护进程管理

### 9.14.3 下一步学习

- **Services 深入学习**：ClusterIP、LoadBalancer、ExternalName
- **Ingress 控制器**：HTTP/HTTPS 路由和负载均衡
- **ConfigMap 和 Secret**：配置和敏感信息管理
- **Persistent Volumes**：持久化存储管理
- **Helm 包管理**：应用打包和部署

## 9.15 API 参考和实用工具

### 9.15.1 API 参考

- **Deployment API**: <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#deployment-v1-apps>
- **Service API**: <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#service-v1-core>
- **Pod API**: <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pod-v1-core>

### 9.15.2 实用工具

- **kubectl 备忘单**: <https://kubernetes.io/docs/reference/kubectl/cheatsheet/>
- **YAML 验证器**: <https://kubeyaml.com/>
- **资源计算器**: <https://learnk8s.io/kubernetes-resource-calculator>
- **最佳实践指南**: <https://kubernetes.io/docs/concepts/configuration/overview/>
