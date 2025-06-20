# 5. Services with kubectl

## 5.0 目录

- [5. Services with kubectl](#5-services-with-kubectl)
  - [5.0 目录](#50-目录)
  - [5.1 项目概述](#51-项目概述)
  - [5.2 学习目标](#52-学习目标)
  - [5.3 应用场景](#53-应用场景)
  - [5.4 前置条件](#54-前置条件)
  - [5.5 Service 类型介绍](#55-service-类型介绍)
    - [5.5.1 Service 类型概览](#551-service-类型概览)
    - [5.5.2 本教程重点](#552-本教程重点)
  - [5.6 ClusterIP Service - 后端应用设置](#56-clusterip-service---后端应用设置)
    - [5.6.1 操作概述](#561-操作概述)
    - [5.6.2 详细操作步骤](#562-详细操作步骤)
      - [5.6.2.1 创建后端应用 Deployment](#5621-创建后端应用-deployment)
      - [5.6.2.2 创建 ClusterIP Service](#5622-创建-clusterip-service)
    - [5.6.3 重要说明](#563-重要说明)
      - [5.6.3.1 关于 Service 类型](#5631-关于-service-类型)
      - [5.6.3.2 关于端口配置](#5632-关于端口配置)
      - [5.6.3.3 验证 Service 功能](#5633-验证-service-功能)
    - [5.6.4 应用程序信息](#564-应用程序信息)
    - [5.6.5 架构说明](#565-架构说明)
  - [5.7 NodePort Service - 前端应用设置](#57-nodeport-service---前端应用设置)
    - [5.7.1 操作概述](#571-操作概述)
    - [5.7.2 架构说明](#572-架构说明)
    - [5.7.3 前端应用配置](#573-前端应用配置)
      - [5.7.3.1 Nginx 反向代理配置](#5731-nginx-反向代理配置)
    - [5.7.4 详细操作步骤](#574-详细操作步骤)
      - [5.7.4.1 创建前端应用 Deployment](#5741-创建前端应用-deployment)
      - [5.7.4.2 创建 NodePort Service](#5742-创建-nodeport-service)
      - [5.7.4.3 获取访问信息](#5743-获取访问信息)
    - [5.7.5 负载均衡测试](#575-负载均衡测试)
      - [5.7.5.1 扩展后端应用](#5751-扩展后端应用)
      - [5.7.5.2 验证负载均衡](#5752-验证负载均衡)
    - [5.7.6 架构验证](#576-架构验证)
    - [5.7.7 完整架构图](#577-完整架构图)
  - [5.8 清理资源](#58-清理资源)
    - [5.8.1 完整清理](#581-完整清理)
    - [5.8.2 选择性清理](#582-选择性清理)
  - [5.9 最佳实践和高级用法](#59-最佳实践和高级用法)
    - [5.9.1 Service 选择最佳实践](#591-service-选择最佳实践)
      - [5.9.1.1 何时使用 ClusterIP](#5911-何时使用-clusterip)
      - [5.9.1.2 何时使用 NodePort](#5912-何时使用-nodeport)
    - [5.9.2 监控和调试](#592-监控和调试)
      - [5.9.2.1 服务健康检查](#5921-服务健康检查)
      - [5.9.2.2 网络连通性测试](#5922-网络连通性测试)
      - [5.9.2.3 日志查看](#5923-日志查看)
    - [5.9.3 高级配置示例](#593-高级配置示例)
      - [5.9.3.1 会话亲和性配置](#5931-会话亲和性配置)
      - [5.9.3.2 多端口 Service](#5932-多端口-service)
  - [5.10 故障排除](#510-故障排除)
    - [5.10.1 常见问题和解决方案](#5101-常见问题和解决方案)
      - [5.10.1.1 问题 1：Service 无法访问](#51011-问题-1service-无法访问)
      - [5.10.1.2 问题 2：NodePort 无法从外部访问](#51012-问题-2nodeport-无法从外部访问)
      - [5.10.1.3 问题 3：负载均衡不工作](#51013-问题-3负载均衡不工作)
    - [5.10.2 调试命令集合](#5102-调试命令集合)
  - [5.11 总结](#511-总结)
    - [5.11.1 学习要点回顾](#5111-学习要点回顾)
    - [5.11.2 关键优势总结](#5112-关键优势总结)
    - [5.11.3 下一步学习](#5113-下一步学习)
    - [5.11.4 生产环境建议](#5114-生产环境建议)
  - [5.12 后续主题](#512-后续主题)
    - [5.12.1 LoadBalancer Service](#5121-loadbalancer-service)
    - [5.12.2 ExternalName Service](#5122-externalname-service)
  - [5.13 参考资料](#513-参考资料)

## 5.1 项目概述

本教程将深入学习 Kubernetes Services 的核心概念和实际应用。通过实际操作，您将掌握如何使用不同类型的 Service 来暴露和管理应用程序的网络访问。

## 5.2 学习目标

完成本教程后，您将能够：

- **理解 Service 概念**：掌握 Kubernetes Service 的作用和工作原理
- **掌握 Service 类型**：了解 ClusterIP、NodePort、LoadBalancer 和 ExternalName 的区别
- **实现服务发现**：使用 ClusterIP Service 实现集群内部服务发现
- **暴露外部访问**：使用 NodePort Service 暴露应用程序给外部用户
- **构建完整架构**：创建前后端分离的完整应用架构
- **负载均衡实践**：验证 Service 的负载均衡功能

## 5.3 应用场景

- **微服务架构**：为微服务提供稳定的网络访问入口
- **服务发现**：实现服务间的自动发现和通信
- **负载均衡**：在多个 Pod 实例间分发流量
- **外部访问**：为集群内应用提供外部访问能力

## 5.4 前置条件

在开始本教程之前，请确保您已经：

- ✅ 完成 Kubernetes 集群搭建
- ✅ 安装并配置 kubectl 命令行工具
- ✅ 具备基本的 Kubernetes Pod 和 Deployment 知识
- ✅ 了解基本的网络概念（端口、代理等）

## 5.5 Service 类型介绍

### 5.5.1 Service 类型概览

Kubernetes 提供四种主要的 Service 类型：

1. **ClusterIP**（默认类型）
   - 仅在集群内部可访问
   - 为 Service 分配一个集群内部 IP
   - 适用于内部服务通信

2. **NodePort**
   - 在每个节点上开放一个端口
   - 通过节点 IP 和端口从外部访问
   - 端口范围：30000-32767

3. **LoadBalancer**
   - 云提供商的负载均衡器
   - 自动分配外部 IP
   - 主要用于云环境

4. **ExternalName**
   - 将服务映射到外部 DNS 名称
   - 不分配 IP 地址
   - 需要 YAML 定义

### 5.5.2 本教程重点

- 本节将详细学习 **ClusterIP** 和 **NodePort** 类型
- LoadBalancer 类型因云提供商而异，将在相应的云平台教程中介绍
- ExternalName 类型需要 YAML 定义，将在后续 YAML 教程中涵盖

## 5.6 ClusterIP Service - 后端应用设置

### 5.6.1 操作概述

在这一步中，我们将：

- 创建后端应用的 Deployment（Spring Boot REST 应用）
- 为后端应用创建 ClusterIP Service 实现负载均衡

### 5.6.2 详细操作步骤

#### 5.6.2.1 创建后端应用 Deployment

```bash
# 创建后端 REST 应用的 Deployment
kubectl create deployment my-backend-rest-app --image=grissomsh/kube-helloworld:1.0.0

# 查看 Deployment 状态
kubectl get deploy
kubectl get pods
```

**预期输出：**

```text
NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
my-backend-rest-app    1/1     1            1           30s
```

#### 5.6.2.2 创建 ClusterIP Service

```bash
# 为后端应用创建 ClusterIP Service
kubectl expose deployment my-backend-rest-app --port=8080 --target-port=8080 --name=my-backend-service

# 查看 Service 状态
kubectl get svc
kubectl describe svc my-backend-service
```

**预期输出：**

```text
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
my-backend-service   ClusterIP   10.96.100.123   <none>        8080/TCP   15s
```

### 5.6.3 重要说明

#### 5.6.3.1 关于 Service 类型

- 🔍 **观察**：我们不需要指定 `--type=ClusterIP`，因为 ClusterIP 是默认的 Service 类型
- 📝 **默认行为**：kubectl expose 命令默认创建 ClusterIP Service

#### 5.6.3.2 关于端口配置

- **--port=8080**：Service 暴露的端口
- **--target-port=8080**：Pod 中容器的端口
- 💡 **提示**：当 Service 端口和容器端口相同时，可以省略 `--target-port`，但为了清晰起见，建议明确指定

#### 5.6.3.3 验证 Service 功能

```bash
# 查看 Service 详细信息
kubectl get svc my-backend-service -o wide

# 查看 Service 的 Endpoints
kubectl get endpoints my-backend-service

# 测试集群内部访问（从另一个 Pod 中测试）
kubectl run test-pod --image=busybox --rm -it --restart=Never -- wget -qO- http://my-backend-service:8080/hello
```

### 5.6.4 应用程序信息

- **应用类型**：Spring Boot REST API
- **容器端口**：8080
- **健康检查端点**：`/hello`
- **源代码位置**：[kube-helloworld](../00-Docker-Images/02-kube-backend-helloworld-springboot/kube-helloworld)

### 5.6.5 架构说明

```text
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │───▶│  ClusterIP       │───▶│   Backend       │
│   (Nginx)       │    │  Service         │    │   (Spring Boot) │
│                 │    │  my-backend-     │    │                 │
│                 │    │  service:8080    │    │   Port: 8080    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

这个 ClusterIP Service 将作为后端应用的稳定访问入口，为前端应用提供服务发现和负载均衡功能。

## 5.7 NodePort Service - 前端应用设置

### 5.7.1 操作概述

在这一步中，我们将：

- 创建前端应用的 Deployment（Nginx 反向代理）
- 为前端应用创建 NodePort Service 以提供外部访问
- 验证完整的前后端架构和负载均衡功能

### 5.7.2 架构说明

虽然我们在之前的教程中多次使用了 **NodePort Service**，但这次我们将构建一个完整的架构视图，展示 NodePort Service 与 ClusterIP Service 的协作关系。

### 5.7.3 前端应用配置

#### 5.7.3.1 Nginx 反向代理配置

前端使用 Nginx 作为反向代理，将请求转发到后端服务。关键配置如下：

```nginx
server {
    listen       80;
    server_name  localhost;
    location / {
        # 后端 ClusterIP Service 的名称和端口
        # proxy_pass http://<Backend-ClusterIp-Service-Name>:<Port>;
        proxy_pass http://my-backend-service:8080;
    }
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**重要说明：**

- 🔗 **服务发现**：Nginx 配置中使用 `my-backend-service` 作为后端服务名
- 🐳 **预构建镜像**：我们已经准备好了配置完成的镜像 `grissomsh/kube-frontend-nginx:1.0.0`
- 📦 **镜像位置**：[Docker Hub](https://hub.docker.com/repository/docker/grissomsh/kube-frontend-nginx)
- 📁 **源代码位置**：[kube-frontend-nginx](../00-Docker-Images/03-kube-frontend-nginx)

### 5.7.4 详细操作步骤

#### 5.7.4.1 创建前端应用 Deployment

```bash
# 创建前端 Nginx 代理的 Deployment
kubectl create deployment my-frontend-nginx-app --image=grissomsh/kube-frontend-nginx:1.0.0

# 查看 Deployment 状态
kubectl get deploy
kubectl get pods -l app=my-frontend-nginx-app
```

**预期输出：**

```text
NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
my-frontend-nginx-app    1/1     1            1           30s
```

#### 5.7.4.2 创建 NodePort Service

```bash
# 为前端应用创建 NodePort Service
kubectl expose deployment my-frontend-nginx-app --type=NodePort --port=80 --target-port=80 --name=my-frontend-service

# 查看 Service 状态
kubectl get svc
kubectl describe svc my-frontend-service
```

**预期输出：**

```text
NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
my-frontend-service   NodePort    10.96.200.456   <none>        80:31234/TCP   20s
my-backend-service    ClusterIP   10.96.100.123   <none>        8080/TCP       5m
```

#### 5.7.4.3 获取访问信息

```bash
# 获取 Service 信息
kubectl get svc my-frontend-service

# 获取节点信息
kubectl get nodes -o wide

# 查看完整的服务列表
kubectl get svc -o wide
```

**访问应用：**

```bash
# 访问格式
http://<node-ip>:<node-port>/hello

# 示例（根据实际输出替换）
http://192.168.1.100:31234/hello
```

### 5.7.5 负载均衡测试

#### 5.7.5.1 扩展后端应用

```bash
# 将后端应用扩展到 10 个副本
kubectl scale --replicas=10 deployment/my-backend-rest-app

# 验证扩展结果
kubectl get pods -l app=my-backend-rest-app
kubectl get deployment my-backend-rest-app
```

**预期输出：**

```text
NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
my-backend-rest-app    10/10   10           10          10m
```

#### 5.7.5.2 验证负载均衡

```bash
# 多次访问应用，观察负载均衡效果
for i in {1..10}; do
  curl http://<node-ip>:<node-port>/hello
  echo "Request $i completed"
  sleep 1
done
```

**观察要点：**

- 🔄 **负载分发**：请求会被分发到不同的后端 Pod
- 📊 **响应内容**：每个 Pod 可能返回不同的主机名或实例信息
- ⚡ **响应时间**：观察响应时间的一致性

### 5.7.6 架构验证

```bash
# 查看完整的应用架构
kubectl get all

# 查看服务端点
kubectl get endpoints

# 查看服务详细信息
kubectl describe svc my-frontend-service
kubectl describe svc my-backend-service
```

### 5.7.7 完整架构图

```text
外部用户
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes 集群                          │
│                                                             │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐ │
│  │   NodePort  │───▶│   Frontend   │───▶│   ClusterIP     │ │
│  │   Service   │    │   (Nginx)    │    │   Service       │ │
│  │   :31234    │    │   Port: 80   │    │   :8080         │ │
│  └─────────────┘    └──────────────┘    └─────────────────┘ │
│                                                ▼             │
│                                        ┌─────────────────┐   │
│                                        │   Backend       │   │
│                                        │   (Spring Boot) │   │
│                                        │   10 Replicas   │   │
│                                        │   Port: 8080    │   │
│                                        └─────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 5.8 清理资源

### 5.8.1 完整清理

当您完成实验后，可以清理所有创建的资源：

```bash
# 删除 Services
kubectl delete svc my-frontend-service
kubectl delete svc my-backend-service

# 删除 Deployments
kubectl delete deployment my-frontend-nginx-app
kubectl delete deployment my-backend-rest-app

# 验证清理结果
kubectl get all
```

### 5.8.2 选择性清理

如果您想保留某些资源用于进一步学习：

```bash
# 仅删除前端相关资源
kubectl delete svc my-frontend-service
kubectl delete deployment my-frontend-nginx-app

# 保留后端资源用于其他实验
# kubectl get svc my-backend-service
# kubectl get deployment my-backend-rest-app
```

## 5.9 最佳实践和高级用法

### 5.9.1 Service 选择最佳实践

#### 5.9.1.1 何时使用 ClusterIP

- ✅ **内部服务通信**：微服务之间的通信
- ✅ **数据库访问**：应用访问数据库服务
- ✅ **API 网关后端**：作为 API 网关的后端服务
- ✅ **缓存服务**：Redis、Memcached 等缓存服务

#### 5.9.1.2 何时使用 NodePort

- ✅ **开发测试**：快速暴露服务进行测试
- ✅ **简单部署**：小规模部署或概念验证
- ✅ **特定端口需求**：需要固定端口的应用
- ❌ **生产环境**：不推荐在生产环境直接使用

### 5.9.2 监控和调试

#### 5.9.2.1 服务健康检查

```bash
# 检查服务状态
kubectl get svc -o wide

# 检查端点状态
kubectl get endpoints

# 查看服务详细信息
kubectl describe svc <service-name>

# 检查 Pod 标签匹配
kubectl get pods --show-labels
```

#### 5.9.2.2 网络连通性测试

```bash
# 从集群内测试服务连通性
kubectl run debug-pod --image=busybox --rm -it --restart=Never -- sh

# 在 debug pod 中执行
wget -qO- http://my-backend-service:8080/hello
nslookup my-backend-service
```

#### 5.9.2.3 日志查看

```bash
# 查看应用日志
kubectl logs -l app=my-backend-rest-app
kubectl logs -l app=my-frontend-nginx-app

# 实时查看日志
kubectl logs -f deployment/my-backend-rest-app
```

### 5.9.3 高级配置示例

#### 5.9.3.1 会话亲和性配置

```bash
# 创建带会话亲和性的 Service
kubectl expose deployment my-backend-rest-app \
  --port=8080 \
  --target-port=8080 \
  --name=my-backend-sticky \
  --session-affinity=ClientIP
```

#### 5.9.3.2 多端口 Service

```bash
# 暴露多个端口（需要使用 YAML 配置）
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: multi-port-service
spec:
  selector:
    app: my-backend-rest-app
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
EOF
```

## 5.10 故障排除

### 5.10.1 常见问题和解决方案

#### 5.10.1.1 问题 1：Service 无法访问

**症状：**

- 无法通过 Service 名称访问应用
- 连接超时或拒绝连接

**排查步骤：**

```bash
# 1. 检查 Service 是否存在
kubectl get svc

# 2. 检查 Endpoints
kubectl get endpoints my-backend-service

# 3. 检查 Pod 标签
kubectl get pods --show-labels

# 4. 检查 Service 选择器
kubectl describe svc my-backend-service
```

**解决方案：**

- 确保 Pod 标签与 Service 选择器匹配
- 检查 Pod 是否处于 Running 状态
- 验证容器端口配置

#### 5.10.1.2 问题 2：NodePort 无法从外部访问

**症状：**

- 无法通过 NodePort 访问应用
- 浏览器显示连接失败

**排查步骤：**

```bash
# 1. 检查 NodePort Service
kubectl get svc my-frontend-service

# 2. 检查节点状态
kubectl get nodes -o wide

# 3. 检查防火墙设置
# 确保 NodePort 端口（30000-32767）未被阻止

# 4. 测试集群内访问
kubectl run test --image=busybox --rm -it --restart=Never -- wget -qO- http://my-frontend-service/hello
```

**解决方案：**

- 检查防火墙和安全组设置
- 确认使用正确的节点 IP 和端口
- 验证 Service 类型为 NodePort

#### 5.10.1.3 问题 3：负载均衡不工作

**症状：**

- 请求总是路由到同一个 Pod
- 负载分布不均匀

**排查步骤：**

```bash
# 1. 检查后端 Pod 数量
kubectl get pods -l app=my-backend-rest-app

# 2. 检查 Endpoints
kubectl get endpoints my-backend-service

# 3. 测试多次请求
for i in {1..10}; do curl http://<service-url>/hello; done
```

**解决方案：**

- 确保有多个健康的后端 Pod
- 检查应用是否支持无状态访问
- 考虑禁用会话亲和性

### 5.10.2 调试命令集合

```bash
# 服务发现调试
kubectl get svc,endpoints,pods -o wide

# 网络连通性测试
kubectl run netshoot --image=nicolaka/netshoot --rm -it --restart=Never -- bash

# DNS 解析测试
kubectl run dnsutils --image=tutum/dnsutils --rm -it --restart=Never -- nslookup my-backend-service

# 端口转发测试
kubectl port-forward svc/my-backend-service 8080:8080

# 查看 kube-proxy 日志
kubectl logs -n kube-system -l k8s-app=kube-proxy
```

## 5.11 总结

### 5.11.1 学习要点回顾

通过本教程，您已经掌握了：

1. **Service 基础概念**
   - 四种 Service 类型的特点和用途
   - ClusterIP 和 NodePort 的实际应用

2. **实际操作技能**
   - 使用 kubectl 创建和管理 Service
   - 配置前后端应用的网络通信
   - 实现负载均衡和服务发现

3. **架构设计能力**
   - 设计微服务网络架构
   - 选择合适的 Service 类型
   - 理解服务间通信模式

### 5.11.2 关键优势总结

- **服务发现**：自动发现和连接后端服务
- **负载均衡**：自动分发流量到健康的 Pod
- **高可用性**：Pod 故障时自动切换
- **解耦合**：前后端通过 Service 名称通信，而非 IP

### 5.11.3 下一步学习

建议继续学习以下主题：

1. **YAML 配置**：学习使用 YAML 文件定义 Service
2. **Ingress**：学习更高级的外部访问控制
3. **LoadBalancer**：在云环境中使用 LoadBalancer Service
4. **Service Mesh**：了解 Istio 等服务网格技术

### 5.11.4 生产环境建议

- 🔒 **安全性**：使用 Network Policies 限制网络访问
- 📊 **监控**：部署 Prometheus 监控 Service 性能
- 🚀 **性能**：根据负载调整 Service 和 Pod 配置
- 🔄 **高可用**：在多个可用区部署应用

## 5.12 后续主题

以下主题将在后续课程中详细介绍：

### 5.12.1 LoadBalancer Service

- 云提供商集成
- 自动 IP 分配
- 云平台特定配置

### 5.12.2 ExternalName Service

- 外部服务映射
- DNS 配置
- YAML 定义方式

## 5.13 参考资料

- [Kubernetes Services 官方文档](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Service 类型详解](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)
- [网络概念指南](https://kubernetes.io/docs/concepts/services-networking/)
- [kubectl 命令参考](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
