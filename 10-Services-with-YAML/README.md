# 10. Services with YAML

## 10.0 目录

- [10. Services with YAML](#10-services-with-yaml)
  - [10.0 目录](#100-目录)
  - [10.1 项目概述](#101-项目概述)
  - [10.2 学习目标](#102-学习目标)
  - [10.3 应用场景](#103-应用场景)
  - [10.4 环境准备](#104-环境准备)
    - [10.4.1 环境要求](#1041-环境要求)
    - [10.4.2 验证环境](#1042-验证环境)
  - [10.5 项目架构和 Service 类型](#105-项目架构和-service-类型)
    - [10.5.1 Service 简介](#1051-service-简介)
    - [10.5.2 核心特性](#1052-核心特性)
    - [10.5.3 Service 类型对比](#1053-service-类型对比)
    - [10.5.4 教程架构](#1054-教程架构)
  - [10.6 创建后端应用和 ClusterIP Service](#106-创建后端应用和-clusterip-service)
    - [10.6.1 后端应用架构](#1061-后端应用架构)
    - [10.6.2 ClusterIP Service 特点](#1062-clusterip-service-特点)
    - [10.6.3 后端 Deployment 配置要点](#1063-后端-deployment-配置要点)
    - [10.6.4 ClusterIP Service 配置要点](#1064-clusterip-service-配置要点)
    - [10.6.5 重要说明](#1065-重要说明)
    - [10.6.6 创建后端资源](#1066-创建后端资源)
    - [10.6.7 验证后端服务](#1067-验证后端服务)
  - [10.7 创建前端应用和 NodePort Service](#107-创建前端应用和-nodeport-service)
    - [10.7.1 前端应用架构](#1071-前端应用架构)
    - [10.7.2 NodePort Service 特点](#1072-nodeport-service-特点)
    - [10.7.3 前端 Deployment 配置要点](#1073-前端-deployment-配置要点)
    - [10.7.4 NodePort Service 配置要点](#1074-nodeport-service-配置要点)
    - [10.7.5 网络通信流程](#1075-网络通信流程)
    - [10.7.6 创建前端资源](#1076-创建前端资源)
    - [10.7.7 访问应用](#1077-访问应用)
      - [获取访问信息](#获取访问信息)
      - [测试应用访问](#测试应用访问)
    - [10.7.8 验证端到端通信](#1078-验证端到端通信)
  - [10.8 批量管理 YAML 资源](#108-批量管理-yaml-资源)
    - [10.8.1 批量操作的优势](#1081-批量操作的优势)
    - [10.8.2 单文件操作](#1082-单文件操作)
    - [10.8.3 批量文件操作](#1083-批量文件操作)
    - [10.8.4 批量删除操作](#1084-批量删除操作)
    - [10.8.5 使用标签选择器管理资源](#1085-使用标签选择器管理资源)
      - [标签策略](#标签策略)
      - [基于标签的批量操作](#基于标签的批量操作)
  - [10.9 Service 高级功能](#109-service-高级功能)
    - [10.9.1 服务发现机制](#1091-服务发现机制)
      - [DNS 解析](#dns-解析)
      - [环境变量](#环境变量)
    - [10.9.2 负载均衡测试](#1092-负载均衡测试)
      - [测试后端负载均衡](#测试后端负载均衡)
      - [监控流量分发](#监控流量分发)
    - [10.9.3 Session Affinity（会话亲和性）](#1093-session-affinity会话亲和性)
      - [配置示例](#配置示例)
  - [10.10 监控和调试](#1010-监控和调试)
    - [10.10.1 资源状态监控](#10101-资源状态监控)
    - [10.10.2 网络连接测试](#10102-网络连接测试)
    - [10.10.3 日志分析](#10103-日志分析)
    - [10.10.4 故障排除](#10104-故障排除)
  - [10.11 最佳实践](#1011-最佳实践)
    - [10.11.1 Service 配置最佳实践](#10111-service-配置最佳实践)
    - [10.11.2 生产环境建议](#10112-生产环境建议)
    - [10.11.3 安全最佳实践](#10113-安全最佳实践)
  - [10.12 故障排除指南](#1012-故障排除指南)
    - [10.12.1 常见问题](#10121-常见问题)
      - [Service 无法访问](#service-无法访问)
      - [Pod 无法启动](#pod-无法启动)
      - [网络连接问题](#网络连接问题)
  - [10.13 清理资源](#1013-清理资源)
    - [10.13.1 完整清理](#10131-完整清理)
    - [10.13.2 选择性清理](#10132-选择性清理)
  - [10.14 学习总结](#1014-学习总结)
    - [10.14.1 关键要点](#10141-关键要点)
    - [10.14.2 进阶学习建议](#10142-进阶学习建议)
    - [10.14.3 下一步学习](#10143-下一步学习)
  - [10.15 API 参考和实用工具](#1015-api-参考和实用工具)
    - [10.15.1 官方文档](#10151-官方文档)
    - [10.15.2 实用工具](#10152-实用工具)
    - [10.15.3 标签选择器参考](#10153-标签选择器参考)

## 10.1 项目概述

本教程将深入学习 Kubernetes Services 的 YAML 配置和管理，通过前后端分离的实际应用案例，掌握不同类型 Service 的使用场景和配置方法。

## 10.2 学习目标

- 理解 Kubernetes Service 的核心概念和工作原理
- 掌握 ClusterIP 和 NodePort Service 的配置和使用
- 学习前后端应用的服务发现和负载均衡
- 实践 Service 与 Deployment 的协同工作
- 掌握 YAML 文件的批量管理和操作

## 10.3 应用场景

- 微服务架构中的服务发现
- 前后端分离应用的网络通信
- 负载均衡和高可用性配置
- 生产环境的服务暴露策略

## 10.4 环境准备

### 10.4.1 环境要求

- Kubernetes 集群（版本 1.20+）
- kubectl 命令行工具
- 基础的 YAML 语法知识
- 了解 Deployment 和 Pod 概念

### 10.4.2 验证环境

```bash
# 检查 kubectl 连接
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查当前命名空间
kubectl config current-context
```

## 10.5 项目架构和 Service 类型

### 10.5.1 Service 简介

Kubernetes Service 是一个抽象层，它定义了一组 Pod 的逻辑集合以及访问这些 Pod 的策略。Service 为 Pod 提供稳定的网络端点，即使 Pod 被重新创建或重新调度，Service 的 IP 地址和 DNS 名称也保持不变。

### 10.5.2 核心特性

1. **服务发现**：通过 DNS 或环境变量自动发现服务
2. **负载均衡**：在多个 Pod 实例之间分发流量
3. **故障转移**：自动将流量路由到健康的 Pod
4. **网络抽象**：隐藏底层网络复杂性

### 10.5.3 Service 类型对比

| 类型 | 用途 | 访问方式 | 适用场景 |
|------|------|----------|----------|
| ClusterIP | 集群内部通信 | 集群内 DNS/IP | 微服务间通信 |
| NodePort | 外部访问 | 节点IP:端口 | 开发测试环境 |
| LoadBalancer | 云环境外部访问 | 云负载均衡器 | 生产环境 |
| ExternalName | 外部服务映射 | DNS CNAME | 外部服务集成 |

### 10.5.4 教程架构

本教程将创建一个完整的前后端应用架构：

- **后端**：Spring Boot REST API（ClusterIP Service）
- **前端**：Nginx 反向代理（NodePort Service）
- **通信**：前端通过 Service 名称访问后端

## 10.6 创建后端应用和 ClusterIP Service

### 10.6.1 后端应用架构

后端是一个 Spring Boot REST API 应用，提供 `/hello` 端点服务。我们将使用 ClusterIP Service 来暴露后端服务，使其只能在集群内部访问。

### 10.6.2 ClusterIP Service 特点

- **默认类型**：不指定 type 时的默认 Service 类型
- **集群内访问**：只能从集群内部访问
- **DNS 解析**：自动创建 DNS 记录
- **负载均衡**：自动分发请求到后端 Pod

### 10.6.3 后端 Deployment 配置要点

```yaml
# 01-backend-deployment.yml 关键配置
spec:
  replicas: 3                    # 3个副本确保高可用
  selector:
    matchLabels:
      app: backend-restapp        # 标签选择器
  template:
    spec:
      containers:
      - name: backend-restapp
        image: grissomsh/kube-helloworld:1.0.0
        ports:
        - containerPort: 8080     # 容器端口
```

### 10.6.4 ClusterIP Service 配置要点

```yaml
# 02-backend-clusterip-service.yml 关键配置
metadata:
  name: my-backend-service        # 重要：前端 Nginx 配置中引用此名称
spec:
  # type: ClusterIP 是默认类型，可以省略
  selector:
    app: backend-restapp          # 选择后端 Pod
  ports:
  - port: 8080                    # Service 端口
    targetPort: 8080              # 容器端口
```

### 10.6.5 重要说明

- **Service 名称**：`my-backend-service` 必须与前端 Nginx 反向代理配置中的上游服务器名称一致
- **端口映射**：Service 端口和容器端口都是 8080
- **标签选择**：通过 `app: backend-restapp` 标签选择后端 Pod

### 10.6.6 创建后端资源

```bash
# 进入配置文件目录
cd /Users/wangtianqing/Project/kubernetes-fundamentals/10-Services-with-YAML/kube-manifests

# 查看当前资源状态
kubectl get all

# 创建后端 Deployment 和 Service
kubectl apply -f 01-backend-deployment.yml -f 02-backend-clusterip-service.yml

# 验证创建结果
kubectl get all
kubectl get deployment backend-restapp
kubectl get service my-backend-service
kubectl get pods -l app=backend-restapp
```

### 10.6.7 验证后端服务

```bash
# 查看 Service 详细信息
kubectl describe service my-backend-service

# 查看 Endpoints（Service 关联的 Pod IP）
kubectl get endpoints my-backend-service

# 测试集群内访问（创建临时 Pod 测试）
kubectl run test-pod --image=busybox --rm -i --restart=Never -- \
  wget -qO- http://my-backend-service:8080/hello
```

## 10.7 创建前端应用和 NodePort Service

### 10.7.1 前端应用架构

前端是一个 Nginx 应用，配置了反向代理，将 `/hello` 请求转发到后端 Service。我们使用 NodePort Service 来暴露前端服务，使其可以从集群外部访问。

### 10.7.2 NodePort Service 特点

- **外部访问**：可以从集群外部通过节点 IP 访问
- **端口范围**：默认端口范围 30000-32767
- **自动分配**：可以自动分配端口或手动指定
- **负载均衡**：在所有节点上开放相同端口

### 10.7.3 前端 Deployment 配置要点

```yaml
# 03-frontend-deployment.yml 关键配置
spec:
  replicas: 3                    # 3个副本确保高可用
  selector:
    matchLabels:
      app: frontend-nginxapp      # 标签选择器
  template:
    spec:
      containers:
      - name: frontend-nginxapp
        image: grissomsh/kube-frontend-nginx:1.0.0
        ports:
        - containerPort: 80         # Nginx 默认端口
```

### 10.7.4 NodePort Service 配置要点

```yaml
# 04-frontend-nodeport-service.yml 关键配置
metadata:
  name: frontend-nginxapp-nodeport-service
spec:
  type: NodePort                  # 指定为 NodePort 类型
  selector:
    app: frontend-nginxapp        # 选择前端 Pod
  ports:
  - port: 80                      # Service 端口
    targetPort: 80                # 容器端口
    nodePort: 31234               # 节点端口（可选，不指定则自动分配）
```

### 10.7.5 网络通信流程

1. **外部请求** → 节点IP:31234
2. **NodePort Service** → 前端 Pod:80
3. **Nginx 反向代理** → my-backend-service:8080
4. **ClusterIP Service** → 后端 Pod:8080

### 10.7.6 创建前端资源

```bash
# 进入配置文件目录
cd /Users/wangtianqing/Project/kubernetes-fundamentals/10-Services-with-YAML/kube-manifests

# 查看当前资源状态
kubectl get all

# 创建前端 Deployment 和 Service
kubectl apply -f 03-frontend-deployment.yml -f 04-frontend-nodeport-service.yml

# 验证创建结果
kubectl get all
kubectl get deployment frontend-nginxapp
kubectl get service frontend-nginxapp-nodeport-service
kubectl get pods -l app=frontend-nginxapp
```

### 10.7.7 访问应用

#### 获取访问信息

```bash
# 获取节点外部 IP
kubectl get nodes -o wide

# 查看 NodePort Service 详细信息
kubectl describe service frontend-nginxapp-nodeport-service

# 获取 NodePort 端口
kubectl get service frontend-nginxapp-nodeport-service -o jsonpath='{.spec.ports[0].nodePort}'
```

#### 测试应用访问

```bash
# 方式1：通过节点 IP 和 NodePort 访问
# 替换 <node-ip> 为实际的节点 IP
curl http://<node-ip>:31234/hello

# 方式2：使用 kubectl port-forward（开发测试）
kubectl port-forward service/frontend-nginxapp-nodeport-service 8080:80
# 然后在另一个终端访问：curl http://localhost:8080/hello

# 方式3：在浏览器中访问
# http://<node-ip>:31234/hello
```

### 10.7.8 验证端到端通信

```bash
# 查看前端 Pod 日志
kubectl logs -l app=frontend-nginxapp

# 查看后端 Pod 日志
kubectl logs -l app=backend-restapp

# 查看 Service Endpoints
kubectl get endpoints
```

## 10.8 批量管理 YAML 资源

### 10.8.1 批量操作的优势

- **效率提升**：一次操作多个资源
- **原子性**：确保相关资源同时创建或删除
- **版本控制**：便于配置文件的统一管理
- **自动化**：支持 CI/CD 流水线集成

### 10.8.2 单文件操作

```bash
# 逐个删除资源文件
kubectl delete -f 01-backend-deployment.yml
kubectl delete -f 02-backend-clusterip-service.yml
kubectl delete -f 03-frontend-deployment.yml
kubectl delete -f 04-frontend-nodeport-service.yml

# 验证删除结果
kubectl get all
```

### 10.8.3 批量文件操作

```bash
# 方式1：指定多个文件
kubectl apply -f 01-backend-deployment.yml -f 02-backend-clusterip-service.yml -f 03-frontend-deployment.yml -f 04-frontend-nodeport-service.yml

# 方式2：使用目录（推荐）
cd /Users/wangtianqing/Project/kubernetes-fundamentals/10-Services-with-YAML
kubectl apply -f kube-manifests/

# 验证创建结果
kubectl get all
```

### 10.8.4 批量删除操作

```bash
# 删除目录中的所有资源
cd /Users/wangtianqing/Project/kubernetes-fundamentals/10-Services-with-YAML
kubectl delete -f kube-manifests/

# 验证删除结果
kubectl get all
```

### 10.8.5 使用标签选择器管理资源

#### 标签策略

```bash
# 查看所有带有特定标签的资源
kubectl get all -l tier=backend
kubectl get all -l tier=frontend

# 查看所有相关资源
kubectl get all -l 'tier in (frontend,backend)'
```

#### 基于标签的批量操作

```bash
# 删除所有后端资源
kubectl delete all -l tier=backend

# 删除所有前端资源
kubectl delete all -l tier=frontend

# 删除整个应用的所有资源
kubectl delete all -l 'tier in (frontend,backend)'
```

## 10.9 Service 高级功能

### 10.9.1 服务发现机制

#### DNS 解析

```bash
# 在集群内测试 DNS 解析
kubectl run dns-test --image=busybox --rm -i --restart=Never -- nslookup my-backend-service

# 查看完整的 DNS 名称
echo "my-backend-service.default.svc.cluster.local"
```

#### 环境变量

```bash
# 查看 Pod 中的 Service 环境变量
kubectl exec -it <frontend-pod-name> -- env | grep SERVICE
```

### 10.9.2 负载均衡测试

#### 测试后端负载均衡

```bash
# 创建测试脚本
for i in {1..10}; do
  kubectl run test-$i --image=busybox --rm -i --restart=Never -- \
    wget -qO- http://my-backend-service:8080/hello
  echo "Request $i completed"
done
```

#### 监控流量分发

```bash
# 查看后端 Pod 日志
kubectl logs -l app=backend-restapp -f --tail=50

# 在另一个终端发送请求
for i in {1..20}; do
  curl http://<node-ip>:31234/hello
  sleep 1
done
```

### 10.9.3 Session Affinity（会话亲和性）

#### 配置示例

```yaml
# Service 配置中添加会话亲和性
spec:
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800  # 3小时
```

## 10.10 监控和调试

### 10.10.1 资源状态监控

```bash
# 实时监控资源状态
watch kubectl get all

# 查看资源详细信息
kubectl describe deployment backend-restapp
kubectl describe service my-backend-service

# 查看 Pod 详细状态
kubectl get pods -o wide
kubectl describe pod <pod-name>
```

### 10.10.2 网络连接测试

```bash
# 测试 Service 连通性
kubectl run network-test --image=nicolaka/netshoot --rm -i --restart=Never -- \
  curl -v http://my-backend-service:8080/hello

# 测试端口连通性
kubectl run port-test --image=busybox --rm -i --restart=Never -- \
  telnet my-backend-service 8080
```

### 10.10.3 日志分析

```bash
# 查看应用日志
kubectl logs -l app=backend-restapp --tail=100
kubectl logs -l app=frontend-nginxapp --tail=100

# 实时跟踪日志
kubectl logs -l app=backend-restapp -f

# 查看之前的日志
kubectl logs <pod-name> --previous
```

### 10.10.4 故障排除

```bash
# 检查 Service Endpoints
kubectl get endpoints

# 检查 DNS 解析
kubectl run dns-debug --image=busybox --rm -i --restart=Never -- \
  nslookup kubernetes.default

# 检查网络策略
kubectl get networkpolicies

# 查看事件
kubectl get events --sort-by=.metadata.creationTimestamp
```

## 10.11 最佳实践

### 10.11.1 Service 配置最佳实践

1. **命名规范**：使用描述性的 Service 名称
2. **标签管理**：合理使用标签进行资源分组
3. **端口命名**：为端口指定有意义的名称
4. **健康检查**：配置适当的健康检查探针
5. **资源限制**：设置合理的资源请求和限制

### 10.11.2 生产环境建议

1. **使用 LoadBalancer**：生产环境避免使用 NodePort
2. **TLS 终止**：在 Ingress 或 Service 层配置 TLS
3. **监控告警**：实施全面的监控和告警
4. **备份策略**：定期备份配置文件
5. **版本控制**：使用 Git 管理 YAML 文件

### 10.11.3 安全最佳实践

1. **网络策略**：限制 Pod 间的网络通信
2. **服务账户**：使用最小权限原则
3. **镜像安全**：使用可信的镜像源
4. **密钥管理**：使用 Secret 管理敏感信息
5. **RBAC**：实施基于角色的访问控制

## 10.12 故障排除指南

### 10.12.1 常见问题

#### Service 无法访问

```bash
# 检查 Service 是否存在
kubectl get service

# 检查 Endpoints
kubectl get endpoints <service-name>

# 检查标签选择器
kubectl get pods --show-labels
```

#### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl get pods

# 查看 Pod 详细信息
kubectl describe pod <pod-name>

# 查看 Pod 日志
kubectl logs <pod-name>
```

#### 网络连接问题

```bash
# 检查 DNS 解析
kubectl exec -it <pod-name> -- nslookup <service-name>

# 检查端口连通性
kubectl exec -it <pod-name> -- telnet <service-name> <port>

# 检查防火墙规则
kubectl get networkpolicies
```

## 10.13 清理资源

### 10.13.1 完整清理

```bash
# 删除所有相关资源
kubectl delete -f kube-manifests/

# 或使用标签删除
kubectl delete all -l 'tier in (frontend,backend)'

# 验证清理结果
kubectl get all
```

### 10.13.2 选择性清理

```bash
# 只删除 Service
kubectl delete service my-backend-service frontend-nginxapp-nodeport-service

# 只删除 Deployment
kubectl delete deployment backend-restapp frontend-nginxapp
```

## 10.14 学习总结

### 10.14.1 关键要点

1. **Service 类型**：理解不同 Service 类型的使用场景
2. **服务发现**：掌握 DNS 和环境变量两种服务发现机制
3. **负载均衡**：了解 Service 的负载均衡工作原理
4. **网络通信**：理解前后端应用的网络通信流程
5. **批量管理**：掌握 YAML 文件的批量操作技巧

### 10.14.2 进阶学习建议

1. **Ingress 控制器**：学习更高级的流量管理
2. **Service Mesh**：了解 Istio 等服务网格技术
3. **网络策略**：深入学习 Kubernetes 网络安全
4. **监控体系**：集成 Prometheus 和 Grafana
5. **CI/CD 集成**：将 Service 管理集成到部署流水线

### 10.14.3 下一步学习

- [11-ConfigMaps-and-Secrets](../11-ConfigMaps-and-Secrets/README.md)
- [12-Ingress-Controllers](../12-Ingress-Controllers/README.md)
- [13-Persistent-Volumes](../13-Persistent-Volumes/README.md)

## 10.15 API 参考和实用工具

### 10.15.1 官方文档

- [Kubernetes Services](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Service Types](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)
- [DNS for Services](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

### 10.15.2 实用工具

- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [YAML Validator](https://kubeyaml.com/)
- [Kubernetes Dashboard](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/)
- [Lens IDE](https://k8slens.dev/)

### 10.15.3 标签选择器参考

- [Using Labels Effectively](https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/#using-labels-effectively)
- [Label Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors)
