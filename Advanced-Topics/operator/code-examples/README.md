# Spring Boot Operator 实验代码

本目录包含了 Spring Boot Operator 教程的所有实验代码和配置文件，从文档中提取并按实验分类组织。

## 目录

- [Spring Boot Operator 实验代码](#spring-boot-operator-实验代码)
  - [目录](#目录)
  - [目录结构](#目录结构)
  - [一、实验概述](#一实验概述)
    - [实验一：基础 Operator](#实验一基础-operator)
    - [实验二：配置管理](#实验二配置管理)
    - [实验三：服务暴露和 Ingress](#实验三服务暴露和-ingress)
    - [实验四：综合微服务应用](#实验四综合微服务应用)
  - [二、快速开始](#二快速开始)
    - [2.1 前置条件](#21-前置条件)
    - [2.2 环境准备](#22-环境准备)
    - [2.3 运行实验](#23-运行实验)
      - [实验一：基础 Operator 运行](#实验一基础-operator-运行)
      - [实验二：运行配置管理功能](#实验二运行配置管理功能)
      - [实验三：将服务对外暴露](#实验三将服务对外暴露)
      - [实验四：微服务应用完整部署](#实验四微服务应用完整部署)
  - [三、验证和测试](#三验证和测试)
    - [3.1 检查 Operator 状态](#31-检查-operator-状态)
    - [3.2 查看应用状态](#32-查看应用状态)
    - [3.3 查看日志](#33-查看日志)
  - [四、故障排除](#四故障排除)
    - [4.1 常见问题](#41-常见问题)
    - [4.2 清理资源](#42-清理资源)
  - [五、扩展开发](#五扩展开发)
    - [5.1 添加新功能](#51-添加新功能)
    - [5.2 最佳实践](#52-最佳实践)
  - [六、参考资源](#六参考资源)

## 目录结构

```bash
code-examples/
├── README.md                           # 本文件
├── experiment-1-basic-operator/        # 实验一：基础 Operator
├── experiment-2-config-management/     # 实验二：配置管理
├── experiment-3-service-ingress/       # 实验三：服务暴露和 Ingress
└── experiment-4-microservices/         # 实验四：综合微服务应用
```

## 一、实验概述

### 实验一：基础 Operator

**目标**: 创建一个基础的 Spring Boot Operator，能够管理 Spring Boot 应用的部署和服务。

**包含文件**:

- `springbootapp_types.go` - CRD 类型定义
- `springbootapp_controller.go` - Controller 实现
- `test-springboot-app.yaml` - 测试应用配置
- `deploy.sh` - 部署脚本

**功能特性**:

- 自定义资源定义 (CRD)
- Deployment 管理
- Service 创建
- 状态更新

### 实验二：配置管理

**目标**: 扩展 Operator 支持 ConfigMap 配置管理和热更新。

**包含文件**:

- `springbootapp_types_extended.go` - 扩展的 CRD 定义
- `config-demo.yaml` - 配置管理示例
- `test-config.sh` - 配置测试脚本

**功能特性**:

- ConfigMap 挂载
- 环境变量注入
- 配置热更新
- 配置哈希跟踪

### 实验三：服务暴露和 Ingress

**目标**: 实现多种服务暴露方式，包括 NodePort、LoadBalancer 和 Ingress。

**包含文件**:

- `springbootapp_types_full.go` - 完整的 CRD 定义
- `service-ingress-tests.yaml` - 服务暴露测试配置
- `test-service-ingress.sh` - 服务测试脚本

**功能特性**:

- 多种 Service 类型支持
- Ingress 自动创建
- TLS 证书管理
- 负载均衡配置

### 实验四：综合微服务应用

**目标**: 部署一个完整的微服务架构，展示 Operator 在实际场景中的应用。

**包含文件**:

- `microservices-namespace.yaml` - 命名空间和配置
- `microservices-apps.yaml` - 微服务应用部署
- `postgres.yaml` - 数据库部署
- `deploy-microservices.sh` - 综合部署脚本

**架构组件**:

- Gateway Service (API 网关)
- User Service (用户管理)
- Order Service (订单管理)
- PostgreSQL (数据库)

## 二、快速开始

### 2.1 前置条件

1. Kubernetes 集群 (v1.20+)
2. kubectl 命令行工具
3. Go 开发环境 (v1.19+)
4. Operator SDK
5. Docker

### 2.2 环境准备

```bash
# 安装 Operator SDK
curl -LO https://github.com/operator-framework/operator-sdk/releases/latest/download/operator-sdk_darwin_amd64
chmod +x operator-sdk_darwin_amd64 && sudo mv operator-sdk_darwin_amd64 /usr/local/bin/operator-sdk

# 验证安装
operator-sdk version
```

### 2.3 运行实验

#### 实验一：基础 Operator 运行

**说明**: 本实验创建一个基础的 Spring Boot Operator，能够管理 Spring Boot 应用的部署和服务。

**操作步骤**:

```bash
cd experiment-1-basic-operator

# 创建 Operator 项目
operator-sdk init --domain=tutorial.example.com --repo=github.com/example/springboot-operator

# 创建 API
operator-sdk create api --group=springboot --version=v1 --kind=SpringBootApp --resource --controller

# 复制代码文件
cp springbootapp_types.go api/v1/
cp springbootapp_controller.go controllers/

# 运行部署脚本
chmod +x deploy.sh
./deploy.sh
```

#### 实验二：运行配置管理功能

**说明**: 本实验扩展 Operator 支持 ConfigMap 配置管理和热更新功能。

**操作步骤**:

```bash
cd experiment-2-config-management

# 更新 API 定义
cp springbootapp_types_extended.go ../experiment-1-basic-operator/api/v1/springbootapp_types.go

# 运行配置测试
chmod +x test-config.sh
./test-config.sh
```

#### 实验三：将服务对外暴露

**说明**: 本实验将服务暴露功能扩展为支持 NodePort、LoadBalancer 和 Ingress 配置。

**操作步骤**:

```bash
cd experiment-3-service-ingress

# 更新 API 定义
cp springbootapp_types_full.go ../experiment-1-basic-operator/api/v1/springbootapp_types.go

# 运行服务测试
chmod +x test-service-ingress.sh
./test-service-ingress.sh
```

#### 实验四：微服务应用完整部署

```bash
cd experiment-4-microservices

# 部署微服务架构
chmod +x deploy-microservices.sh
./deploy-microservices.sh
```

---

## 三、验证和测试

### 3.1 检查 Operator 状态

```bash
# 查看 CRD
kubectl get crd springbootapps.springboot.tutorial.example.com

# 查看 Operator Pod
kubectl get pods -n springboot-operator-system

# 查看 SpringBootApp 资源
kubectl get springbootapp -A
```

### 3.2 查看应用状态

```bash
# 查看应用 Pod
kubectl get pods -l app.kubernetes.io/managed-by=springboot-operator

# 查看服务
kubectl get services -l app.kubernetes.io/managed-by=springboot-operator

# 查看 Ingress
kubectl get ingress -l app.kubernetes.io/managed-by=springboot-operator
```

### 3.3 查看日志

```bash
# Operator 日志
kubectl logs -n springboot-operator-system deployment/springboot-operator-controller-manager

# 应用日志
kubectl logs deployment/my-spring-app
```

## 四、故障排除

### 4.1 常见问题

1. **CRD 未创建**

   ```bash
   make install
   ```

2. **Operator 启动失败**

   ```bash
   kubectl describe pod -n springboot-operator-system
   ```

3. **应用部署失败**

   ```bash
   kubectl describe springbootapp <app-name>
   kubectl get events --sort-by=.metadata.creationTimestamp
   ```

4. **镜像拉取失败**
   - 检查镜像名称和标签
   - 确保镜像仓库可访问
   - 配置镜像拉取密钥

### 4.2 清理资源

```bash
# 删除测试应用
kubectl delete springbootapp --all -A

# 删除 Operator
make undeploy

# 删除 CRD
make uninstall

# 删除微服务命名空间
kubectl delete namespace microservices
```

## 五、扩展开发

### 5.1 添加新功能

1. **修改 API 定义**
   - 编辑 `api/v1/springbootapp_types.go`
   - 添加新的字段到 Spec 或 Status

2. **更新 Controller 逻辑**
   - 编辑 `controllers/springbootapp_controller.go`
   - 实现新功能的协调逻辑

3. **重新生成代码**

   ```bash
   make generate
   make manifests
   ```

4. **测试和部署**

   ```bash
   make test
   make docker-build docker-push IMG=<your-registry>/springboot-operator:tag
   make deploy IMG=<your-registry>/springboot-operator:tag
   ```

### 5.2 最佳实践

1. **错误处理**: 实现完善的错误处理和重试机制
2. **状态管理**: 及时更新资源状态，提供清晰的状态信息
3. **事件记录**: 记录重要操作的 Kubernetes 事件
4. **资源清理**: 实现 Finalizer 确保资源正确清理
5. **监控指标**: 暴露 Prometheus 指标用于监控
6. **日志记录**: 提供详细的结构化日志

## 六、参考资源

- [Operator SDK 文档](https://sdk.operatorframework.io/)
- [Kubernetes API 参考](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Controller Runtime 文档](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Spring Boot 官方文档](https://spring.io/projects/spring-boot)
