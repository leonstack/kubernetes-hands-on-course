# Operator 开发教学指南

## 目录

- [Operator 开发教学指南](#operator-开发教学指南)
  - [目录](#目录)
  - [1. 课程概述](#1-课程概述)
    - [1.1 学习目标](#11-学习目标)
    - [1.2 前置知识](#12-前置知识)
  - [2. Kubernetes Operator 基础](#2-kubernetes-operator-基础)
    - [2.1 什么是 Operator](#21-什么是-operator)
      - [2.1.1 Operator 架构概览](#211-operator-架构概览)
      - [2.1.2 Operator Pattern 核心原理](#212-operator-pattern-核心原理)
    - [2.2 Operator 的组成部分](#22-operator-的组成部分)
      - [2.2.1 Operator 组件架构图](#221-operator-组件架构图)
      - [2.2.2 组件详细说明](#222-组件详细说明)
    - [2.3 Operator 的优势](#23-operator-的优势)
  - [3. Spring Boot 应用特点分析](#3-spring-boot-应用特点分析)
    - [3.1 Spring Boot 应用的典型架构](#31-spring-boot-应用的典型架构)
      - [3.1.1 Spring Boot 微服务架构图](#311-spring-boot-微服务架构图)
    - [3.2 Spring Boot 在 Kubernetes 中的部署挑战](#32-spring-boot-在-kubernetes-中的部署挑战)
      - [3.2.1 Spring Boot 应用部署流程图](#321-spring-boot-应用部署流程图)
      - [3.2.2 主要部署挑战分析](#322-主要部署挑战分析)
    - [3.3 为什么需要 Spring Boot Operator](#33-为什么需要-spring-boot-operator)
      - [3.3.1 传统部署 vs Operator 部署对比](#331-传统部署-vs-operator-部署对比)
      - [3.3.2 Spring Boot Operator 的核心价值](#332-spring-boot-operator-的核心价值)
  - [4. 实验驱动的 Spring Boot Operator 开发](#4-实验驱动的-spring-boot-operator-开发)
    - [4.1 实验环境准备](#41-实验环境准备)
      - [4.1.1 环境要求](#411-环境要求)
      - [4.1.2 项目初始化](#412-项目初始化)
    - [4.2 Operator 功能规划](#42-operator-功能规划)
    - [4.3 实验一：基础 Operator 设计与实现](#43-实验一基础-operator-设计与实现)
      - [4.3.1 设计目标](#431-设计目标)
      - [4.3.2 API 设计思路](#432-api-设计思路)
      - [4.3.3 实验步骤](#433-实验步骤)
      - [4.3.4 测试验证](#434-测试验证)
    - [4.4 实验二：配置管理功能](#44-实验二配置管理功能)
      - [4.4.1 设计目标](#441-设计目标)
      - [4.4.2 实验二架构设计图](#442-实验二架构设计图)
      - [4.4.3 配置变更检测流程图](#443-配置变更检测流程图)
      - [4.4.2 API 扩展设计](#442-api-扩展设计)
      - [4.4.3 实验步骤](#443-实验步骤)
      - [4.4.4 测试验证](#444-测试验证)
    - [4.5 实验三：服务暴露和 Ingress](#45-实验三服务暴露和-ingress)
      - [4.5.1 设计目标](#451-设计目标)
      - [4.5.2 实验三架构设计图](#452-实验三架构设计图)
      - [4.5.3 服务类型选择流程图](#453-服务类型选择流程图)
      - [4.5.2 API 扩展设计](#452-api-扩展设计)
      - [4.5.3 实验步骤](#453-实验步骤)
      - [4.5.4 测试验证](#454-测试验证)
    - [4.6 综合实验：完整的微服务应用](#46-综合实验完整的微服务应用)
      - [4.6.1 实验目标](#461-实验目标)
      - [4.6.2 实验架构](#462-实验架构)
      - [4.6.3 微服务通信流程图](#463-微服务通信流程图)
      - [4.6.3 实验步骤](#463-实验步骤)
  - [5. 总结](#5-总结)
    - [5.1 学习路径总览](#51-学习路径总览)
    - [5.2 技术栈总览](#52-技术栈总览)
    - [5.3 核心收获](#53-核心收获)
      - [理论知识](#理论知识)
      - [实践技能](#实践技能)
      - [工程能力](#工程能力)
    - [5.4 扩展方向](#54-扩展方向)
      - [功能增强](#功能增强)
      - [运维集成](#运维集成)
      - [生态集成](#生态集成)
    - [5.5 最佳实践总结](#55-最佳实践总结)
      - [开发阶段](#开发阶段)
      - [部署阶段](#部署阶段)
      - [运维阶段](#运维阶段)

## 1. 课程概述

### 1.1 学习目标

- 理解 Kubernetes Operator 的核心概念
- 掌握为 Spring Boot 应用创建 Operator 的方法
- 学会使用 Operator SDK 开发自定义 Operator
- 实现 Spring Boot 应用的自动化部署和管理

### 1.2 前置知识

- Kubernetes 基础概念（Pod、Service、Deployment 等）
- Spring Boot 应用开发基础
- YAML 配置文件编写
- Go 语言基础（可选，用于 Operator 开发）

## 2. Kubernetes Operator 基础

### 2.1 什么是 Operator

Kubernetes Operator 是一种扩展 Kubernetes API 的方法，它将人类操作员的知识编码到软件中，使应用程序能够自动管理自己。

#### 2.1.1 Operator 架构概览

```text
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Operator 架构                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   kubectl   │    │  Dashboard  │    │  External   │         │
│  │   (CLI)     │    │    (UI)     │    │   Tools     │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│         │                   │                   │               │
│         └───────────────────┼───────────────────┘               │
│                             │                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Kubernetes API Server                     │   │
│  │  ┌─────────────────┐  ┌─────────────────────────────┐  │   │
│  │  │   Built-in      │  │      Custom Resources       │  │   │
│  │  │   Resources     │  │        (CRDs)              │  │   │
│  │  │  - Pod          │  │  - SpringBootApp            │  │   │
│  │  │  - Service      │  │  - DatabaseCluster          │  │   │
│  │  │  - Deployment   │  │  - MonitoringConfig         │  │   │
│  │  └─────────────────┘  └─────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                             │                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                Controller Manager                       │   │
│  │  ┌─────────────────┐  ┌─────────────────────────────┐  │   │
│  │  │   Built-in      │  │      Custom Controllers     │  │   │
│  │  │  Controllers    │  │       (Operators)          │  │   │
│  │  │  - Deployment   │  │  - SpringBootApp Controller │  │   │
│  │  │  - ReplicaSet   │  │  - Database Controller      │  │   │
│  │  │  - Service      │  │  - Monitoring Controller    │  │   │
│  │  └─────────────────┘  └─────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                             │                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   Cluster State                        │   │
│  │  ┌─────────────────┐  ┌─────────────────────────────┐  │   │
│  │  │   Kubernetes    │  │      Application            │  │   │
│  │  │   Resources     │  │      Resources              │  │   │
│  │  │  - Pods         │  │  - Spring Boot Apps         │  │   │
│  │  │  - Services     │  │  - Databases                │  │   │
│  │  │  - ConfigMaps   │  │  - Monitoring Stack         │  │   │
│  │  └─────────────────┘  └─────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

**Operator 的定义和作用：**

- Operator 是一个应用程序特定的控制器，它扩展了 Kubernetes API 来创建、配置和管理复杂有状态应用程序的实例
- 它将运维人员的领域知识编码到软件中，实现应用程序的自动化管理
- Operator 可以处理应用程序的整个生命周期，包括安装、升级、备份、故障恢复等

#### 2.1.2 Operator Pattern 核心原理

```text
┌─────────────────────────────────────────────────────────────────┐
│                    Operator Pattern 工作流程                     │
└─────────────────────────────────────────────────────────────────┘

    用户操作                API 层                控制器层              资源层
┌─────────────┐         ┌─────────────┐         ┌─────────────┐      ┌─────────────┐
│             │  CRUD   │             │ Watch   │             │ CRUD │             │
│   kubectl   │────────►│ API Server  │────────►│ Controller  │─────►│ Kubernetes  │
│             │         │             │         │             │      │ Resources   │
│   apply     │         │ Custom      │         │ Reconcile   │      │             │
│   delete    │         │ Resources   │         │ Loop        │      │ Pod         │
│   get       │         │             │         │             │      │ Service     │
│   ...       │         │ CRDs        │         │             │      │ ConfigMap   │
└─────────────┘         └─────────────┘         └─────────────┘      │ ...         │
                                │                       │             └─────────────┘
                                │                       │
                                │ Status Update        │ Event
                                │◄──────────────────────┘
                                │
                                ▼
                        ┌─────────────┐
                        │   etcd      │
                        │ (State      │
                        │  Store)     │
                        └─────────────┘

控制循环详细流程：

1. 用户创建/更新 Custom Resource
   ├── kubectl apply -f springboot-app.yaml
   └── API Server 验证并存储到 etcd

2. Controller 监听资源变化
   ├── Watch API 接收变化事件
   ├── 将事件加入工作队列
   └── 触发 Reconcile 函数

3. Reconcile 协调逻辑
   ├── 获取当前资源状态
   ├── 比较期望状态 vs 实际状态
   ├── 计算需要执行的操作
   └── 执行创建/更新/删除操作

4. 状态反馈
   ├── 更新 Custom Resource Status
   ├── 记录事件日志
   └── 等待下一次协调周期
```

**Operator Pattern 的核心思想：**

- **声明式 API**：用户声明期望的状态，Operator 负责实现这个状态
- **控制循环**：持续监控实际状态与期望状态的差异，并采取行动消除差异
- **领域知识封装**：将特定应用程序的运维知识封装在代码中
- **事件驱动**：基于 Kubernetes 事件机制，响应资源变化
- **最终一致性**：通过持续协调确保系统最终达到期望状态

**Controller 和 Custom Resource 的关系：**

- Custom Resource (CR)：定义应用程序的期望状态
- Controller：监控 CR 的变化，并执行相应的操作来达到期望状态
- 两者结合形成了完整的 Operator 模式

### 2.2 Operator 的组成部分

#### 2.2.1 Operator 组件架构图

```text
┌─────────────────────────────────────────────────────────────────┐
│                    Operator 核心组件架构                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     开发时组件                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────┐  │
│  │      CRD        │    │   Controller    │    │   RBAC      │  │
│  │   Definition    │    │     Logic       │    │   Rules     │  │
│  │                 │    │                 │    │             │  │
│  │ • Schema        │    │ • Reconcile     │    │ • Roles     │  │
│  │ • Validation    │    │ • Event Watch   │    │ • Bindings  │  │
│  │ • Versions      │    │ • Error Handle  │    │ • Accounts  │  │
│  │ • Subresources  │    │ • Status Update │    │             │  │
│  └─────────────────┘    └─────────────────┘    └─────────────┘  │
│           │                       │                       │      │
│           └───────────────────────┼───────────────────────┘      │
│                                   │                              │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                     运行时组件                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────┐  │
│  │   API Server    │    │   Controller    │    │   etcd      │  │
│  │   Extension     │    │    Manager      │    │   Storage   │  │
│  │                 │    │                 │    │             │  │
│  │ • CRD Registry  │◄──►│ • Work Queue    │◄──►│ • CR State  │  │
│  │ • Validation    │    │ • Reconcile     │    │ • Status    │  │
│  │ • Admission     │    │ • Leader Elect  │    │ • Events    │  │
│  │ • Webhook       │    │ • Metrics       │    │             │  │
│  └─────────────────┘    └─────────────────┘    └─────────────┘  │
│           │                       │                       │      │
│           └───────────────────────┼───────────────────────┘      │
│                                   │                              │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                   管理和分发组件                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────┐  │
│  │      OLM        │    │   OperatorHub   │    │   Catalog   │  │
│  │   (Lifecycle    │    │   (Registry)    │    │   Source    │  │
│  │    Manager)     │    │                 │    │             │  │
│  │                 │    │ • Discovery     │    │ • Bundles   │  │
│  │ • Install       │◄──►│ • Metadata      │◄──►│ • Channels  │  │
│  │ • Upgrade       │    │ • Dependencies  │    │ • Versions  │  │
│  │ • Dependency    │    │ • Security      │    │ • Images    │  │
│  └─────────────────┘    └─────────────────┘    └─────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### 2.2.2 组件详细说明

**Custom Resource Definition (CRD)：**

- **定义新的 Kubernetes 资源类型**：扩展 Kubernetes API，使其能够理解应用程序特定的概念
- **Schema 定义**：使用 OpenAPI v3 规范定义资源结构
- **验证规则**：内置字段验证、格式检查、枚举值限制
- **版本管理**：支持多版本 API，提供版本转换机制
- **子资源支持**：Status 子资源、Scale 子资源等

```yaml
# CRD 示例结构
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: springbootapps.springboot.tutorial.example.com
spec:
  group: springboot.tutorial.example.com
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              image:
                type: string
                pattern: '^[a-zA-Z0-9._/-]+:[a-zA-Z0-9._-]+$'
              replicas:
                type: integer
                minimum: 1
                maximum: 100
```

**Custom Controller：**

- **业务逻辑核心**：实现特定应用的管理逻辑
- **事件监听**：Watch API 监听资源变化事件
- **协调循环**：Reconcile 函数实现期望状态与实际状态的协调
- **错误处理**：重试机制、指数退避、错误分类
- **状态管理**：更新 Custom Resource 的 Status 字段
- **指标暴露**：Prometheus 指标，监控 Controller 性能

**Operator Lifecycle Manager (OLM)：**

- **安装管理**：自动化 Operator 的安装和配置
- **升级策略**：支持自动升级、手动升级、回滚
- **依赖解析**：处理 Operator 之间的依赖关系
- **权限管理**：自动创建和管理 RBAC 规则
- **版本兼容性**：确保 API 版本兼容性
- **安全策略**：镜像签名验证、安全扫描

### 2.3 Operator 的优势

**自动化运维：**

- 减少手动操作，降低人为错误
- 实现 24/7 自动化监控和响应
- 提高运维效率和可靠性

**领域特定知识的封装：**

- 将专家知识编码到软件中
- 标准化最佳实践
- 降低运维门槛

**声明式配置管理：**

- 用户只需声明期望状态
- 系统自动处理实现细节
- 提供一致的用户体验

## 3. Spring Boot 应用特点分析

### 3.1 Spring Boot 应用的典型架构

Spring Boot 是构建企业级 Java 应用程序的流行框架，具有以下特点：

#### 3.1.1 Spring Boot 微服务架构图

```text
┌─────────────────────────────────────────────────────────────────┐
│                Spring Boot 微服务生态架构                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      客户端层                                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Web UI    │  │  Mobile App │  │  Third-party│  │   CLI   │  │
│  │  (React/    │  │   (iOS/     │  │    APIs     │  │  Tools  │  │
│  │   Vue.js)   │  │   Android)  │  │             │  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      网关层                                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │              Spring Cloud Gateway                      │     │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │     │
│  │  │   路由      │  │   过滤器    │  │   限流      │     │     │
│  │  │  Routing    │  │   Filters   │  │Rate Limiting│     │     │
│  │  └─────────────┘  └─────────────┘  └─────────────┘     │     │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │     │
│  │  │   认证      │  │   监控      │  │   日志      │     │     │
│  │  │    Auth     │  │ Monitoring  │  │  Logging    │     │     │
│  │  └─────────────┘  └─────────────┘  └─────────────┘     │     │
│  └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    业务服务层                                     │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   用户服务  │  │   订单服务  │  │   支付服务  │  │   ...   │  │
│  │    User     │  │   Order     │  │   Payment   │  │  Other  │  │
│  │   Service   │  │   Service   │  │   Service   │  │Services │  │
│  │             │  │             │  │             │  │         │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │  │         │  │
│  │ │REST API │ │  │ │REST API │ │  │ │REST API │ │  │         │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │  │         │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │  │         │  │
│  │ │Business │ │  │ │Business │ │  │ │Business │ │  │         │  │
│  │ │ Logic   │ │  │ │ Logic   │ │  │ │ Logic   │ │  │         │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │  │         │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │  │         │  │
│  │ │Data     │ │  │ │Data     │ │  │ │Data     │ │  │         │  │
│  │ │Access   │ │  │ │Access   │ │  │ │Access   │ │  │         │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    基础设施层                                     │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   服务发现  │  │   配置中心  │  │   消息队列  │  │  缓存   │  │
│  │   Eureka/   │  │   Config    │  │  RabbitMQ/  │  │  Redis/ │  │
│  │   Consul    │  │   Server    │  │   Kafka     │  │ Hazelcast│ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   数据库    │  │   监控      │  │   日志      │  │  安全   │  │
│  │  MySQL/     │  │ Prometheus/ │  │   ELK/      │  │  OAuth2/│  │
│  │ PostgreSQL  │  │  Grafana    │  │  Fluentd    │  │   JWT   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**微服务架构特点：**

- **独立部署**：Spring Boot 天然支持微服务架构模式，每个服务都是独立的、可部署的单元
- **服务通信**：服务间通过 REST API、gRPC 或消息队列进行通信
- **服务发现**：支持服务注册与发现（如 Eureka、Consul、Kubernetes Service Discovery）
- **数据隔离**：每个微服务拥有独立的数据存储
- **技术栈自由**：不同服务可以使用不同的技术栈和数据库

**配置管理（application.properties/yml）：**

```yaml
# application.yml 示例
server:
  port: 8080
  servlet:
    context-path: /api

spring:
  datasource:
    url: jdbc:mysql://localhost:3306/demo
    username: ${DB_USERNAME:root}
    password: ${DB_PASSWORD:password}
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true

logging:
  level:
    com.example: DEBUG
  pattern:
    console: "%d{yyyy-MM-dd HH:mm:ss} - %msg%n"

management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus
  endpoint:
    health:
      show-details: always
```

**健康检查端点：**

- Spring Boot Actuator 提供了丰富的监控端点
- `/actuator/health` - 应用健康状态
- `/actuator/info` - 应用信息
- `/actuator/metrics` - 应用指标
- 支持自定义健康检查指标

**监控和指标收集：**

- 集成 Micrometer 指标库
- 支持 Prometheus、Grafana 等监控系统
- 提供 JVM 指标、HTTP 请求指标、数据库连接池指标等
- 支持分布式链路追踪（如 Zipkin、Jaeger）

### 3.2 Spring Boot 在 Kubernetes 中的部署挑战

#### 3.2.1 Spring Boot 应用部署流程图

```text
┌─────────────────────────────────────────────────────────────────┐
│            Spring Boot 应用在 Kubernetes 中的部署流程            │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      开发阶段                                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   代码开发  │  │   单元测试  │  │   集成测试  │  │  代码   │  │
│  │    Code     │  │    Unit     │  │Integration  │  │ Review  │  │
│  │ Development │  │    Test     │  │    Test     │  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      构建阶段                                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Maven/    │  │   Docker    │  │   镜像扫描  │  │  镜像   │  │
│  │   Gradle    │  │   Build     │  │  Security   │  │ Registry│  │
│  │   Build     │  │             │  │   Scan      │  │  Push   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    配置管理阶段                                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │  ConfigMap  │  │   Secret    │  │Environment  │  │ Volume  │  │
│  │   创建      │  │    创建     │  │ Variables   │  │ Mounts  │  │
│  │             │  │             │  │    设置     │  │  配置   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    部署阶段                                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │ Deployment  │  │   Service   │  │   Ingress   │  │  HPA    │  │
│  │    创建     │  │    创建     │  │    创建     │  │  配置   │  │
│  │             │  │             │  │             │  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    运维阶段                                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   健康检查  │  │   日志收集  │  │   监控告警  │  │  故障   │  │
│  │   Health    │  │   Logging   │  │ Monitoring  │  │ Recovery│  │
│  │   Check     │  │             │  │  & Alert    │  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.2.2 主要部署挑战分析

**1. 配置文件管理挑战：**

- **多环境配置**：不同环境（开发、测试、生产）需要不同的配置，传统方式难以管理
- **敏感信息安全**：敏感信息（数据库密码、API 密钥）需要安全存储，避免明文暴露
- **配置热更新**：配置变更需要重启应用或支持热更新，影响服务可用性
- **配置版本管理**：配置文件版本管理和回滚机制复杂
- **配置一致性**：多实例部署时确保配置同步和一致性

**2. 服务发现挑战：**

- **服务注册发现**：微服务间需要相互发现和通信，传统注册中心与K8s机制冲突
- **动态实例管理**：服务实例的动态注册和注销，处理Pod重启和扩缩容
- **负载均衡策略**：负载均衡和故障转移机制需要与K8s Service集成
- **网络策略**：跨命名空间、跨集群的服务通信复杂性
- 跨命名空间的服务访问

**数据库连接管理：**

- 数据库连接池配置优化
- 数据库密码和连接信息的安全管理
- 数据库迁移和版本管理
- 多数据源配置和事务管理

**日志收集：**

- 容器化环境下的日志收集策略
- 结构化日志格式
- 日志聚合和分析
- 日志轮转和存储管理

**滚动更新策略：**

- 零停机部署
- 蓝绿部署和金丝雀发布
- 健康检查和就绪探针配置
- 回滚策略和版本管理

### 3.3 为什么需要 Spring Boot Operator

#### 3.3.1 传统部署 vs Operator 部署对比

```text
┌─────────────────────────────────────────────────────────────────┐
│                    传统 Kubernetes 部署方式                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      开发者                                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   编写      │  │   创建      │  │   配置      │  │  手动   │  │
│  │ Dockerfile  │  │ Deployment  │  │ ConfigMap   │  │ 部署    │  │
│  │             │  │    YAML     │  │  Secret     │  │ 管理    │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   创建      │  │   配置      │  │   设置      │  │  监控   │  │
│  │  Service    │  │  Ingress    │  │   HPA       │  │  告警   │  │
│  │    YAML     │  │   YAML      │  │   YAML      │  │  配置   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      问题和挑战                                  │
├─────────────────────────────────────────────────────────────────┤
│  • 配置文件繁多，容易出错                                        │
│  • 缺乏标准化，每个应用配置不一致                                │
│  • 手动运维，无法自动化处理故障                                  │
│  • 缺乏领域知识，不了解Spring Boot最佳实践                      │
│  • 升级和回滚复杂，风险高                                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                  Spring Boot Operator 部署方式                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      开发者                                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │              只需定义 SpringBootApp CRD                │     │
│  │                                                         │     │
│  │  apiVersion: springboot.example.com/v1                 │     │
│  │  kind: SpringBootApp                                    │     │
│  │  metadata:                                              │     │
│  │    name: my-app                                         │     │
│  │  spec:                                                  │     │
│  │    image: my-app:v1.0.0                                │     │
│  │    replicas: 3                                          │     │
│  │    config:                                              │     │
│  │      database:                                          │     │
│  │        url: jdbc:mysql://db:3306/mydb                  │     │
│  │    service:                                             │     │
│  │      type: ClusterIP                                    │     │
│  │      port: 8080                                         │     │
│  └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Spring Boot Operator                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   自动创建  │  │   自动配置  │  │   自动管理  │  │  自动   │  │
│  │ Deployment  │  │ ConfigMap   │  │   Service   │  │ 监控    │  │
│  │   Secret    │  │    HPA      │  │  Ingress    │  │ 告警    │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   故障      │  │   滚动      │  │   配置      │  │  最佳   │  │
│  │   自愈      │  │   更新      │  │   热更新    │  │  实践   │  │
│  │             │  │             │  │             │  │  应用   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      优势和收益                                  │
├─────────────────────────────────────────────────────────────────┤
│  • 声明式配置，简单易用                                          │
│  • 标准化部署，减少错误                                          │
│  • 自动化运维，提高效率                                          │
│  • 内置最佳实践，提升质量                                        │
│  • 一键升级回滚，降低风险                                        │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.3.2 Spring Boot Operator 的核心价值

**1. 简化部署流程：**

- **声明式配置**：将复杂的部署步骤封装成简单的声明式配置，开发者只需关注业务逻辑
- **依赖管理**：自动处理依赖关系和部署顺序，确保服务按正确顺序启动
- **一键操作**：提供一键部署和升级能力，支持批量操作和环境迁移
- **错误预防**：减少部署错误和不一致性，通过验证机制确保配置正确性

**2. 标准化配置管理：**

- **最佳实践模板**：提供 Spring Boot 应用的最佳实践配置模板，包含性能优化和安全配置
- **统一规范**：统一配置格式和命名规范，提高团队协作效率
- **自动生成**：自动生成 ConfigMap 和 Secret，支持多环境配置管理
- **版本控制**：支持配置的版本管理和回滚，确保配置变更可追溯

**3. 自动化运维任务：**

- **智能扩缩容**：基于应用指标（CPU、内存、QPS）自动扩缩容，优化资源利用率
- **故障自愈**：自动故障检测和恢复，包括健康检查失败重启、依赖服务恢复等
- **数据管理**：自动备份和数据迁移，支持数据库版本升级和迁移
- **监控集成**：自动监控告警配置，集成 Prometheus、Grafana 等监控系统

**4. 提供最佳实践：**

- **部署模式**：内置 Spring Boot 应用的部署最佳实践，包括蓝绿部署、金丝雀发布等
- **健康检查**：自动配置健康检查和就绪探针，确保服务可用性
- **性能优化**：优化资源配置和性能参数，包括 JVM 参数、连接池配置等
- **安全策略**：集成安全策略和网络策略，确保应用安全运行

## 4. 实验驱动的 Spring Boot Operator 开发

本章采用实验驱动的教学方式，通过循序渐进的实验来学习 Spring Boot Operator 的设计和实现。每个实验都包含设计思路、实现步骤和验证方法。

> **📁 完整实验代码**：本章所有实验的完整代码和配置文件已整理在 [`code-examples`](./code-examples/) 目录中，按实验分类组织。每个实验目录包含完整的源代码、配置文件和部署脚本，可直接运行验证。详细的使用说明请参考 [`code-examples/README.md`](./code-examples/README.md)。

### 4.1 实验环境准备

在开始实验之前，我们需要准备开发环境：

#### 4.1.1 环境要求

**必需软件：**

- Go 1.19+
- Docker Desktop
- kubectl
- Kind 或 Minikube（本地 Kubernetes 集群）

**安装步骤：**

1. **安装 Operator SDK**

   ```bash
   # macOS
   brew install operator-sdk
   
   # 或者直接下载
   curl -LO https://github.com/operator-framework/operator-sdk/releases/latest/download/operator-sdk_darwin_amd64
   chmod +x operator-sdk_darwin_amd64
   sudo mv operator-sdk_darwin_amd64 /usr/local/bin/operator-sdk
   ```

2. **安装 Kind**

   ```bash
   go install sigs.k8s.io/kind@v0.20.0
   ```

3. **创建本地集群**

   ```bash
   # 创建集群配置文件
   cat <<EOF > kind-config.yaml
   kind: Cluster
   apiVersion: kind.x-k8s.io/v1alpha4
   nodes:
   - role: control-plane
     kubeadmConfigPatches:
     - |
       kind: InitConfiguration
       nodeRegistration:
         kubeletExtraArgs:
           node-labels: "ingress-ready=true"
     extraPortMappings:
     - containerPort: 80
       hostPort: 80
       protocol: TCP
     - containerPort: 443
       hostPort: 443
       protocol: TCP
   EOF
   
   # 创建集群
   kind create cluster --config=kind-config.yaml --name=operator-lab
   
   # 验证集群
   kubectl cluster-info
   kubectl get nodes
   ```

#### 4.1.2 项目初始化

```bash
# 创建项目目录
mkdir springboot-operator-tutorial
cd springboot-operator-tutorial

# 初始化 Go 模块
go mod init github.com/example/springboot-operator

# 初始化 Operator 项目
operator-sdk init --domain=tutorial.example.com --repo=github.com/example/springboot-operator
```

### 4.2 Operator 功能规划

我们的 Spring Boot Operator 将提供以下核心功能：

**应用部署和更新：**

- 自动创建和管理 Deployment 资源
- 支持滚动更新和回滚
- 镜像版本管理和升级策略
- 副本数量自动调整

**配置管理：**

- 自动生成 ConfigMap 和 Secret
- 支持多环境配置切换
- 配置热更新和应用重启
- 配置模板和变量替换

**健康检查配置：**

- 自动配置 livenessProbe 和 readinessProbe
- 基于 Spring Boot Actuator 的健康检查
- 自定义健康检查端点
- 启动时间和超时配置

**服务暴露：**

- 自动创建 Service 资源
- 支持 ClusterIP、NodePort、LoadBalancer 类型
- Ingress 配置和路由规则
- 服务发现和负载均衡

**数据库连接管理：**

- 数据库连接配置自动化
- 连接池参数优化
- 数据库密码安全管理
- 多数据源支持

**监控配置：**

- Prometheus 指标暴露
- 自定义监控指标
- 告警规则配置
- 日志收集和分析

### 4.3 实验一：基础 Operator 设计与实现

> **📂 实验代码位置**：[`code-examples/experiment-1-basic-operator/`](./code-examples/experiment-1-basic-operator/)

#### 4.3.1 设计目标

在第一个实验中，我们将设计并实现一个最基础的 Spring Boot Operator，它能够：

- 定义 SpringBootApp 自定义资源
- 根据 SpringBootApp 创建对应的 Deployment
- 管理应用的基本生命周期

#### 4.3.2 API 设计思路

**设计原则：**

1. **简单性**：从最基本的功能开始
2. **可扩展性**：为后续功能预留扩展空间
3. **声明式**：用户只需声明期望状态

**API 结构设计：**

核心 API 结构包括 `SpringBootAppSpec`（期望状态）和 `SpringBootAppStatus`（当前状态）两部分：

```go
// 核心字段示例
type SpringBootAppSpec struct {
    Image    string `json:"image"`           // 应用镜像
    Replicas *int32 `json:"replicas,omitempty"` // 副本数量
    Port     int32  `json:"port,omitempty"`     // 应用端口
}

type SpringBootAppStatus struct {
    Replicas      int32  `json:"replicas"`      // 当前副本数
    ReadyReplicas int32  `json:"readyReplicas"` // 就绪副本数
    Phase         string `json:"phase,omitempty"` // 应用阶段
}
```

> **📁 完整代码参考**：详细的 API 定义请查看 [`code-examples/experiment-1/api/v1/springbootapp_types.go`](./code-examples/experiment-1/api/v1/springbootapp_types.go)

#### 4.3.3 实验步骤

**步骤 1：创建 API：**

```bash
# 创建 SpringBootApp API
operator-sdk create api --group=springboot --version=v1 --kind=SpringBootApp --resource --controller
```

**步骤 2：定义 API 结构：**

编辑 `api/v1/springbootapp_types.go`，定义核心数据结构：

```go
// 关键结构体定义
type SpringBootAppSpec struct {
    Image    string  `json:"image"`
    Replicas *int32  `json:"replicas,omitempty"`
    Port     int32   `json:"port,omitempty"`
}

type SpringBootApp struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   SpringBootAppSpec   `json:"spec,omitempty"`
    Status SpringBootAppStatus `json:"status,omitempty"`
}
```

> **📁 完整代码参考**：包含所有 kubebuilder 注解和完整结构定义的代码请查看 [`code-examples/experiment-1/api/v1/springbootapp_types.go`](./code-examples/experiment-1/api/v1/springbootapp_types.go)

**步骤 3：实现基础 Controller：**

编辑 `controllers/springbootapp_controller.go`，实现核心协调逻辑：

```go
// Controller 核心结构
type SpringBootAppReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// 主要协调逻辑
func (r *SpringBootAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. 获取 SpringBootApp 实例
    var springBootApp springbootv1.SpringBootApp
    if err := r.Get(ctx, req.NamespacedName, &springBootApp); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // 2. 协调 Deployment 和 Service
    if err := r.reconcileDeployment(ctx, &springBootApp); err != nil {
        return ctrl.Result{}, err
    }
    if err := r.reconcileService(ctx, &springBootApp); err != nil {
        return ctrl.Result{}, err
    }
    
    // 3. 更新状态
    return ctrl.Result{}, r.updateStatus(ctx, &springBootApp)
}
```

**核心功能包括：**

- `reconcileDeployment()` - 管理应用部署
- `reconcileService()` - 管理服务暴露
- `updateStatus()` - 更新资源状态

> **📁 完整代码参考**：包含完整实现细节的 Controller 代码请查看 [`code-examples/experiment-1/controllers/springbootapp_controller.go`](./code-examples/experiment-1/controllers/springbootapp_controller.go)

**步骤 4：生成 CRD 和部署文件：**

```bash
# 生成 CRD
make manifests

# 生成代码
make generate

# 构建并推送镜像（可选，用于生产环境）
make docker-build docker-push IMG=<your-registry>/springboot-operator:tag
```

**步骤 5：部署到集群：**

```bash
# 安装 CRD
make install

# 运行 Controller（开发模式）
make run
```

#### 4.3.4 测试验证

**创建测试应用：**

```yaml
# config/samples/springboot_v1_springbootapp.yaml
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: demo-app
  namespace: default
spec:
  image: "springio/gs-spring-boot-docker:latest"
  replicas: 2
  port: 8080
```

**部署测试：**

```bash
# 应用测试资源
kubectl apply -f config/samples/springboot_v1_springbootapp.yaml

# 查看创建的资源
kubectl get springbootapp
kubectl get deployment
kubectl get service
kubectl get pods

# 查看应用状态
kubectl describe springbootapp demo-app
```

**验收标准：**

1. ✅ SpringBootApp 资源创建成功
2. ✅ 自动创建对应的 Deployment 和 Service
3. ✅ Pod 正常启动并处于 Running 状态
4. ✅ SpringBootApp 状态正确反映实际情况
5. ✅ 修改 replicas 能触发 Deployment 更新

### 4.4 实验二：配置管理功能

> **📂 实验代码位置**：[`code-examples/experiment-2-config-management/`](./code-examples/experiment-2-config-management/)

#### 4.4.1 设计目标

在第二个实验中，我们将为 Operator 添加配置管理功能：

- 支持通过 ConfigMap 管理应用配置
- 支持环境变量注入
- 配置变更时自动重启应用

#### 4.4.2 实验二架构设计图

```text
┌─────────────────────────────────────────────────────────────────┐
│              实验二：配置管理功能架构                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      配置管理层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │                SpringBootApp CRD                       │     │
│  │                                                         │     │
│  │  spec:                                                  │     │
│  │    image: demo-app:v1.0.0                              │     │
│  │    replicas: 3                                          │     │
│  │    config:                                              │     │
│  │      configMapRef:                                      │     │
│  │        name: app-config                                 │     │
│  │      env:                                               │     │
│  │        - name: SPRING_PROFILES_ACTIVE                  │     │
│  │          value: "production"                           │     │
│  │      mountPath: "/app/config"                          │     │
│  └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                Spring Boot Operator Controller                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │              Enhanced Reconcile Loop                   │     │
│  │                                                         │     │
│  │  1. Watch SpringBootApp & ConfigMap Changes            │     │
│  │  2. Reconcile ConfigMap Resources                      │     │
│  │  3. Update Deployment with Config Mounts              │     │
│  │  4. Inject Environment Variables                       │     │
│  │  5. Handle Config Change Detection                     │     │
│  │  6. Trigger Rolling Update if Needed                  │     │
│  └─────────────────────────────────────────────────────────┘     │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │ ConfigMap   │  │   Volume    │  │    Env      │  │ Change  │  │
│  │ Reconciler  │  │  Manager    │  │  Injector   │  │Detector │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    配置资源管理层                                 │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │ ConfigMap   │  │   Secret    │  │Environment  │  │ Volume  │  │
│  │             │  │             │  │ Variables   │  │ Mounts  │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │             │  │         │  │
│  │ │app.yaml │ │  │ │database │ │  │ ┌─────────┐ │  │ ┌─────┐ │  │
│  │ │app.props│ │  │ │password │ │  │ │SPRING_  │ │  │ │/app/│ │  │
│  │ │log4j.xml│ │  │ │api-key  │ │  │ │PROFILES │ │  │ │config│ │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ │_ACTIVE  │ │  │ └─────┘ │  │
│  └─────────────┘  └─────────────┘  │ └─────────┘ │  └─────────┘  │
│                                    │ ┌─────────┐ │              │
│                                    │ │DATABASE │ │              │
│                                    │ │_URL     │ │              │
│                                    │ └─────────┘ │              │
│                                    └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      应用运行时层                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Pod 1     │  │   Pod 2     │  │   Pod 3     │  │   ...   │  │
│  │             │  │             │  │             │  │         │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────┐ │  │
│  │ │Spring   │ │  │ │Spring   │ │  │ │Spring   │ │  │ │App  │ │  │
│  │ │Boot App │ │  │ │Boot App │ │  │ │Boot App │ │  │ │     │ │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │  │ └─────┘ │  │
│  │             │  │             │  │             │  │         │  │
│  │ Config:     │  │ Config:     │  │ Config:     │  │         │  │
│  │ • /app/config│ │ • /app/config│ │ • /app/config│ │         │  │
│  │ • ENV vars  │  │ • ENV vars  │  │ • ENV vars  │  │         │  │
│  │ • Secrets   │  │ • Secrets   │  │ • Secrets   │  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.4.3 配置变更检测流程图

```text
┌─────────────────────────────────────────────────────────────────┐
│                配置变更检测与处理流程                             │
└─────────────────────────────────────────────────────────────────┘

     ┌─────────────┐
     │   Start     │
     └─────────────┘
            │
            ▼
     ┌─────────────┐
     │   Watch     │
     │ ConfigMap   │
     │   Events    │
     └─────────────┘
            │
            ▼
     ┌─────────────┐
     │ ConfigMap   │ ──── Create/Update/Delete
     │   Changed   │
     └─────────────┘
            │
            ▼
     ┌─────────────┐
     │   Find      │
     │ Associated  │
     │SpringBootApp│
     └─────────────┘
            │
            ▼
     ┌─────────────┐
     │ Calculate   │
     │ Config Hash │
     │ (Current)   │
     └─────────────┘
            │
            ▼
     ┌─────────────┐
     │ Compare     │
     │ with Last   │
     │ Known Hash  │
     └─────────────┘
            │
            ▼
     ┌─────────────┐      ┌─────────────┐
     │   Hash      │ Yes  │   Update    │
     │ Different?  │ ──── │ Deployment  │
     └─────────────┘      │ Annotation  │
            │ No           └─────────────┘
            ▼                     │
     ┌─────────────┐              │
     │   Skip      │              │
     │   Update    │              │
     └─────────────┘              │
            │                     │
            ▼                     ▼
     ┌─────────────┐      ┌─────────────┐
     │   Update    │      │  Trigger    │
     │ Last Known  │      │  Rolling    │
     │    Hash     │      │   Update    │
     └─────────────┘      └─────────────┘
            │                     │
            ▼                     ▼
     ┌─────────────────────────────────┐
     │            End                  │
     └─────────────────────────────────┘
```

#### 4.4.2 API 扩展设计

**扩展 SpringBootAppSpec：**

```go
type SpringBootAppSpec struct {
    // 基础字段
    Image    string  `json:"image"`
    Replicas *int32  `json:"replicas,omitempty"`
    Port     int32   `json:"port,omitempty"`
    
    // 新增配置管理字段
    Config *ConfigSpec `json:"config,omitempty"`
}

type ConfigSpec struct {
    // ConfigMap 引用
    ConfigMapRef *ConfigMapRef `json:"configMapRef,omitempty"`
    
    // 环境变量
    Env []corev1.EnvVar `json:"env,omitempty"`
    
    // 配置文件挂载路径
    MountPath string `json:"mountPath,omitempty"`
}

type ConfigMapRef struct {
    // ConfigMap 名称
    Name string `json:"name"`
    
    // 是否可选
    Optional *bool `json:"optional,omitempty"`
}
```

#### 4.4.3 实验步骤

**步骤 1：更新 API 定义：**

修改 `api/v1/springbootapp_types.go`，添加配置管理相关字段：

```go
// SpringBootAppSpec defines the desired state of SpringBootApp
type SpringBootAppSpec struct {
    // Image is the container image for the Spring Boot application
    Image string `json:"image"`
    
    // Replicas is the number of desired replicas
    // +kubebuilder:default=1
    // +optional
    Replicas *int32 `json:"replicas,omitempty"`
    
    // Port is the port that the application listens on
    // +kubebuilder:default=8080
    // +optional
    Port int32 `json:"port,omitempty"`
    
    // Config defines the configuration for the application
    // +optional
    Config *ConfigSpec `json:"config,omitempty"`
}

// ConfigSpec defines the configuration specification
type ConfigSpec struct {
    // ConfigMapRef references a ConfigMap containing application configuration
    // +optional
    ConfigMapRef *ConfigMapRef `json:"configMapRef,omitempty"`
    
    // Env defines environment variables for the application
    // +optional
    Env []corev1.EnvVar `json:"env,omitempty"`
    
    // MountPath is the path where configuration files will be mounted
    // +kubebuilder:default="/config"
    // +optional
    MountPath string `json:"mountPath,omitempty"`
}

// ConfigMapRef defines a reference to a ConfigMap
type ConfigMapRef struct {
    // Name is the name of the ConfigMap
    Name string `json:"name"`
    
    // Optional indicates whether the ConfigMap must exist
    // +optional
    Optional *bool `json:"optional,omitempty"`
}
```

**步骤 2：更新 Controller 实现：**

修改 `controllers/springbootapp_controller.go`，增加配置管理功能：

```go
// 在 reconcileDeployment 中新增配置处理逻辑
func (r *SpringBootAppReconciler) reconcileDeployment(ctx context.Context, app *springbootv1.SpringBootApp) error {
    // 基础 Deployment 创建逻辑...
    
    // 新增：配置管理处理
    if app.Spec.Config != nil {
        r.applyConfigToContainer(&container, app.Spec.Config)
        r.applyConfigMapVolume(&podTemplate, app.Spec.Config)
    }
    
    return nil
}

// 配置应用辅助函数
func (r *SpringBootAppReconciler) applyConfigToContainer(container *corev1.Container, config *springbootv1.ConfigSpec) {
    // 环境变量注入
    // ConfigMap 文件挂载
    // Spring Boot 配置路径设置
}
```

**主要增强功能：**

- 环境变量动态注入
- ConfigMap 文件挂载
- Spring Boot 配置路径自动设置

> **📁 完整代码参考**：详细的配置管理实现请查看 [`code-examples/experiment-2/controllers/springbootapp_controller.go`](./code-examples/experiment-2/controllers/springbootapp_controller.go)

**步骤 3：添加 ConfigMap 监听：**

更新 Controller 的 `SetupWithManager` 方法以监听 ConfigMap 变化：

```go
// 增加 ConfigMap 监听支持配置热更新
func (r *SpringBootAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&springbootv1.SpringBootApp{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Watches(&source.Kind{Type: &corev1.ConfigMap{}}, 
               handler.EnqueueRequestsFromMapFunc(r.findSpringBootAppsForConfigMap)).
        Complete(r)
}

// 查找使用指定 ConfigMap 的 SpringBootApp 实例
func (r *SpringBootAppReconciler) findSpringBootAppsForConfigMap(obj client.Object) []reconcile.Request {
    // 实现逻辑：遍历所有 SpringBootApp，找到引用该 ConfigMap 的实例
    // 返回需要重新协调的请求列表
}
```

**配置热更新机制：**

- 监听 ConfigMap 变化事件
- 自动触发相关应用的重新部署
- 支持配置的动态更新

> **📁 完整代码参考**：详细的 ConfigMap 监听实现请查看 [`code-examples/experiment-2/controllers/springbootapp_controller.go`](./code-examples/experiment-2/controllers/springbootapp_controller.go)

#### 4.4.4 测试验证

**步骤 1：创建配置文件：**

```yaml
# config-demo.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-config
data:
  application.yml: |
    server:
      port: 8080
    spring:
      application:
        name: demo-app
```

**步骤 2：创建带配置的应用：**

```yaml
# springboot-app-with-config.yaml
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: demo-app-with-config
spec:
  image: "springio/gs-spring-boot-docker:latest"
  replicas: 1
  port: 8080
  config:
    configMapRef:
      name: demo-config
    mountPath: "/app/config"
```

**步骤 3：部署和测试：**

```bash
# 重新生成和部署 CRD
make manifests && make install

# 部署配置和应用
kubectl apply -f config-demo.yaml
kubectl apply -f springboot-app-with-config.yaml

# 验证部署状态
kubectl get springbootapp demo-app-with-config
kubectl get pods -l app=demo-app-with-config

# 验证配置挂载
kubectl exec <pod-name> -- cat /app/config/application.yml
```

**步骤 4：测试配置热更新：**

```bash
# 更新 ConfigMap 触发重启
kubectl patch configmap demo-config --patch='{
  "data": {
    "application.yml": "server:\n  port: 8080\nspring:\n  application:\n    name: demo-app-updated"
  }
}'

# 观察应用重启和验证新配置
kubectl get pods -l app=demo-app-with-config -w
```

**验收标准：**

1. ✅ SpringBootApp 支持 ConfigMap 配置引用
2. ✅ 配置文件正确挂载到指定路径
3. ✅ 环境变量正确注入到容器
4. ✅ ConfigMap 变更触发应用重启
5. ✅ 可选配置（optional: true）正常工作

### 4.5 实验三：服务暴露和 Ingress

> **📂 实验代码位置**：[`code-examples/experiment-3-service-ingress/`](./code-examples/experiment-3-service-ingress/)

#### 4.5.1 设计目标

在第三个实验中，我们将添加服务暴露功能：

- 支持多种 Service 类型（ClusterIP、NodePort、LoadBalancer）
- 支持 Ingress 配置
- 支持自定义域名和路径

#### 4.5.2 实验三架构设计图

```text
┌─────────────────────────────────────────────────────────────────┐
│            实验三：服务暴露和 Ingress 架构                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      外部访问层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Internet  │  │   Browser   │  │   Mobile    │  │  API    │  │
│  │   Traffic   │  │   Client    │  │    App      │  │ Client  │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Ingress 控制层                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │                    Ingress Resource                    │     │
│  │                                                         │     │
│  │  spec:                                                  │     │
│  │    rules:                                               │     │
│  │    - host: demo-app.example.com                        │     │
│  │      http:                                              │     │
│  │        paths:                                           │     │
│  │        - path: /api                                     │     │
│  │          backend:                                       │     │
│  │            service:                                     │     │
│  │              name: demo-app-service                    │     │
│  │              port: 8080                                │     │
│  │    tls:                                                 │     │
│  │    - secretName: demo-app-tls                          │     │
│  │      hosts: [demo-app.example.com]                     │     │
│  └─────────────────────────────────────────────────────────┘     │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Nginx     │  │   Traefik   │  │    HAProxy  │  │  Other  │  │
│  │  Ingress    │  │  Ingress    │  │   Ingress   │  │Ingress  │  │
│  │ Controller  │  │ Controller  │  │ Controller  │  │Controller│  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Service 抽象层                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │ ClusterIP   │  │  NodePort   │  │LoadBalancer │  │ExternalName│
│  │   Service   │  │   Service   │  │   Service   │  │ Service │  │
│  │             │  │             │  │             │  │         │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────┐ │  │
│  │ │Internal │ │  │ │External │ │  │ │Cloud LB │ │  │ │DNS  │ │  │
│  │ │Only     │ │  │ │Access   │ │  │ │Public IP│ │  │ │CNAME│ │  │
│  │ │10.0.1.10│ │  │ │Node:Port│ │  │ │1.2.3.4  │ │  │ │     │ │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │  │ └─────┘ │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                Spring Boot Operator Controller                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │              Service & Ingress Reconciler             │     │
│  │                                                         │     │
│  │  1. Parse Service Configuration                        │     │
│  │  2. Create/Update Service Resource                     │     │
│  │  3. Parse Ingress Configuration                        │     │
│  │  4. Create/Update Ingress Resource                     │     │
│  │  5. Handle TLS Certificate Management                  │     │
│  │  6. Update Service Discovery Annotations               │     │
│  └─────────────────────────────────────────────────────────┘     │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │  Service    │  │   Ingress   │  │    TLS      │  │  DNS    │  │
│  │ Reconciler  │  │ Reconciler  │  │  Manager    │  │ Manager │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      应用实例层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Pod 1     │  │   Pod 2     │  │   Pod 3     │  │   ...   │  │
│  │             │  │             │  │             │  │         │  │
│  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────┐ │  │
│  │ │Spring   │ │  │ │Spring   │ │  │ │Spring   │ │  │ │App  │ │  │
│  │ │Boot App │ │  │ │Boot App │ │  │ │Boot App │ │  │ │     │ │  │
│  │ │:8080    │ │  │ │:8080    │ │  │ │:8080    │ │  │ │     │ │  │
│  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │  │ └─────┘ │  │
│  │             │  │             │  │             │  │         │  │
│  │ Labels:     │  │ Labels:     │  │ Labels:     │  │         │  │
│  │ • app=demo  │  │ • app=demo  │  │ • app=demo  │  │         │  │
│  │ • version=v1│  │ • version=v1│  │ • version=v1│  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.5.3 服务类型选择流程图

```text
┌─────────────────────────────────────────────────────────────────┐
│                    服务类型选择决策流程                           │
└─────────────────────────────────────────────────────────────────┘

     ┌─────────────┐
     │   Start     │
     │ Service     │
     │Configuration│
     └─────────────┘
            │
            ▼
     ┌─────────────┐
     │   Need      │ Yes  ┌─────────────┐
     │ External    │ ──── │   Need      │ Yes  ┌─────────────┐
     │ Access?     │      │ Static IP?  │ ──── │LoadBalancer │
     └─────────────┘      └─────────────┘      │   Service   │
            │ No                  │ No          └─────────────┘
            ▼                     ▼                    │
     ┌─────────────┐      ┌─────────────┐              │
     │ ClusterIP   │      │  NodePort   │              │
     │   Service   │      │   Service   │              │
     │             │      │             │              │
     │ • Internal  │      │ • Node IP   │              │
     │   Only      │      │ • Fixed Port│              │
     │ • Fast      │      │ • Manual LB │              │
     │ • Secure    │      │ • Dev/Test  │              │
     └─────────────┘      └─────────────┘              │
            │                     │                    │
            ▼                     ▼                    ▼
     ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
     │   Need      │ Yes  │   Need      │ Yes  │   Need      │ Yes
     │  Ingress?   │ ──── │  Ingress?   │ ──── │  Ingress?   │ ────┐
     └─────────────┘      └─────────────┘      └─────────────┘     │
            │ No                  │ No                  │ No        │
            ▼                     ▼                     ▼           ▼
     ┌─────────────┐      ┌─────────────┐      ┌─────────────┐ ┌─────────────┐
     │   Direct    │      │   Direct    │      │   Direct    │ │   Create    │
     │ ClusterIP   │      │  NodePort   │      │LoadBalancer │ │   Ingress   │
     │   Access    │      │   Access    │      │   Access    │ │             │
     └─────────────┘      └─────────────┘      └─────────────┘ │ • Domain    │
                                                               │ • Path Route│
                                                               │ • TLS/SSL   │
                                                               │ • Load LB   │
                                                               └─────────────┘
```

#### 4.5.2 API 扩展设计

**扩展 SpringBootAppSpec：**

```go
type SpringBootAppSpec struct {
    // 基础字段
    Image    string      `json:"image"`
    Replicas *int32      `json:"replicas,omitempty"`
    Port     int32       `json:"port,omitempty"`
    Config   *ConfigSpec `json:"config,omitempty"`
    
    // 新增服务暴露字段
    Service *ServiceSpec `json:"service,omitempty"`
    Ingress *IngressSpec `json:"ingress,omitempty"`
}

type ServiceSpec struct {
    Type     corev1.ServiceType `json:"type,omitempty"`     // Service 类型
    NodePort int32              `json:"nodePort,omitempty"` // NodePort 端口
    Ports    []ServicePort      `json:"ports,omitempty"`    // 额外端口
}

type IngressSpec struct {
    Enabled     bool              `json:"enabled,omitempty"`     // 是否启用
    Host        string            `json:"host,omitempty"`        // 主机名
    Path        string            `json:"path,omitempty"`        // 路径
    TLS         *IngressTLS       `json:"tls,omitempty"`         // TLS 配置
    Annotations map[string]string `json:"annotations,omitempty"` // 注解
}
```

**核心设计思路：**

- 支持多种 Service 类型（ClusterIP、NodePort、LoadBalancer）
- 灵活的 Ingress 配置，支持自定义域名和路径
- TLS 证书管理和自动化配置

> **📁 完整代码参考**：详细的 API 定义请查看 [`code-examples/experiment-3/api/v1/springbootapp_types.go`](./code-examples/experiment-3/api/v1/springbootapp_types.go)

#### 4.5.3 实验步骤

**步骤 1：更新 API 定义：**

在 `api/v1/springbootapp_types.go` 中添加服务暴露相关字段，并添加必要的导入：

```go
import (
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
)
```

**步骤 2：更新 Controller 实现：**

修改 `controllers/springbootapp_controller.go`，添加 Ingress 管理功能：

```go
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *SpringBootAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 获取 SpringBootApp 实例
    var springBootApp springbootv1.SpringBootApp
    if err := r.Get(ctx, req.NamespacedName, &springBootApp); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // 协调各种资源
    if err := r.reconcileDeployment(ctx, &springBootApp); err != nil {
        return ctrl.Result{}, err
    }
    if err := r.reconcileService(ctx, &springBootApp); err != nil {
        return ctrl.Result{}, err
    }
    if err := r.reconcileIngress(ctx, &springBootApp); err != nil {
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, r.updateStatus(ctx, &springBootApp)
}

// 更新 reconcileService 方法以支持自定义 Service 配置
func (r *SpringBootAppReconciler) reconcileService(ctx context.Context, app *springbootv1.SpringBootApp) error {
    service := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace},
    }
    
    _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
        ctrl.SetControllerReference(app, service, r.Scheme)
        
        // 配置端口和服务类型
        port := app.Spec.Port
        if port == 0 { port = 8080 }
        
        serviceType := corev1.ServiceTypeClusterIP
        if app.Spec.Service != nil && app.Spec.Service.Type != "" {
            serviceType = app.Spec.Service.Type
        }
        
        service.Spec = corev1.ServiceSpec{
            Selector: map[string]string{"app": app.Name},
            Ports: []corev1.ServicePort{{
                Name: "http", Port: port, TargetPort: intstr.FromInt(int(port)),
            }},
            Type: serviceType,
        }
        
        // 处理 NodePort 和额外端口配置...
        return nil
    })
    return err
}

// reconcileIngress 管理 Ingress 资源
func (r *SpringBootAppReconciler) reconcileIngress(ctx context.Context, app *springbootv1.SpringBootApp) error {
    // 如果未启用 Ingress，删除现有资源
    if app.Spec.Ingress == nil || !app.Spec.Ingress.Enabled {
        return r.deleteIngressIfExists(ctx, app)
    }
    
    ingress := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace},
    }
    
    _, err := ctrl.CreateOrUpdate(ctx, r.Client, ingress, func() error {
        ctrl.SetControllerReference(app, ingress, r.Scheme)
        
        // 配置基本路径和端口
        path := app.Spec.Ingress.Path
        if path == "" { path = "/" }
        
        port := app.Spec.Port
        if port == 0 { port = 8080 }
        
        // 构建 Ingress 规则
        rule := networkingv1.IngressRule{
            Host: app.Spec.Ingress.Host,
            IngressRuleValue: networkingv1.IngressRuleValue{
                HTTP: &networkingv1.HTTPIngressRuleValue{
                    Paths: []networkingv1.HTTPIngressPath{{
                        Path: path,
                        Backend: networkingv1.IngressBackend{
                            Service: &networkingv1.IngressServiceBackend{
                                Name: app.Name,
                                Port: networkingv1.ServiceBackendPort{Number: port},
                            },
                        },
                    }},
                },
            },
        }
        
        ingress.Spec = networkingv1.IngressSpec{
            Rules: []networkingv1.IngressRule{rule},
        }
        
        // 处理 TLS 和注解配置...
        return nil
    })
    return err
}

// 更新 SetupWithManager 以监听 Ingress
func (r *SpringBootAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&springbootv1.SpringBootApp{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Owns(&networkingv1.Ingress{}).
        Watches(&source.Kind{Type: &corev1.ConfigMap{}}, 
               handler.EnqueueRequestsFromMapFunc(r.findSpringBootAppsForConfigMap)).
        Complete(r)
}
```

> **📁 完整代码参考**：详细的 Controller 实现请查看 [`code-examples/experiment-3/controllers/springbootapp_controller.go`](./code-examples/experiment-3/controllers/springbootapp_controller.go)

#### 4.5.4 测试验证

**步骤 1：创建 NodePort 服务测试：**

```yaml
# springboot-app-nodeport.yaml
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: demo-app-nodeport
spec:
  image: "springio/gs-spring-boot-docker:latest"
  replicas: 1
  port: 8080
  service:
    type: NodePort
    nodePort: 30080
```

**步骤 2：创建 Ingress 测试：**

```yaml
# springboot-app-ingress.yaml
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: demo-app-ingress
spec:
  image: "springio/gs-spring-boot-docker:latest"
  replicas: 2
  port: 8080
  ingress:
    enabled: true
    host: "demo-app.local"
    path: "/api"
    annotations:
      nginx.ingress.kubernetes.io/rewrite-target: "/"
```

**步骤 3：部署和测试：**

```bash
# 重新生成和部署 CRD
make manifests && make install

# 测试 NodePort 服务
kubectl apply -f springboot-app-nodeport.yaml
kubectl get service demo-app-nodeport
kubectl port-forward service/demo-app-nodeport 8080:8080

# 测试 Ingress（需要先安装 Nginx Ingress Controller）
kubectl apply -f springboot-app-ingress.yaml
kubectl get ingress demo-app-ingress

# 配置本地域名解析并测试
echo "127.0.0.1 demo-app.local" | sudo tee -a /etc/hosts
curl -H "Host: demo-app.local" http://localhost/api/
```

**验收标准：**

1. ✅ 支持不同类型的 Service（ClusterIP、NodePort、LoadBalancer）
2. ✅ 支持自定义端口配置
3. ✅ Ingress 资源正确创建和配置
4. ✅ 支持自定义域名和路径
5. ✅ 支持 TLS 配置
6. ✅ 支持 Ingress 注解

### 4.6 综合实验：完整的微服务应用

> **📂 实验代码位置**：[`code-examples/experiment-4-microservices/`](./code-examples/experiment-4-microservices/)

#### 4.6.1 实验目标

通过一个综合实验，部署一个完整的微服务应用，包括：

- 用户服务（User Service）
- 订单服务（Order Service）
- 网关服务（Gateway Service）
- 配置管理
- 服务发现
- 监控和日志

#### 4.6.2 实验架构

```text
┌─────────────────────────────────────────────────────────────────┐
│                    实验四：完整微服务架构                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      外部访问层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Web UI    │  │   Mobile    │  │   API       │  │  Admin  │  │
│  │   Client    │  │    App      │  │  Client     │  │ Console │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Ingress 网关层                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │                    Nginx Ingress                       │     │
│  │                                                         │     │
│  │  Rules:                                                 │     │
│  │  • microservices.local/api/users/* → Gateway Service   │     │
│  │  • microservices.local/api/orders/* → Gateway Service  │     │
│  │  • microservices.local/admin/* → Gateway Service       │     │
│  │                                                         │     │
│  │  TLS: microservices-tls-secret                         │     │
│  └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API 网关层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐     │
│  │                  Gateway Service                       │     │
│  │                (Spring Cloud Gateway)                 │     │
│  │                                                         │     │
│  │  Routes:                                                │     │
│  │  • /api/users/** → User Service (Load Balanced)        │     │
│  │  • /api/orders/** → Order Service (Load Balanced)      │     │
│  │                                                         │     │
│  │  Features:                                              │     │
│  │  • Authentication & Authorization                       │     │
│  │  • Rate Limiting & Circuit Breaker                     │     │
│  │  • Request/Response Logging                             │     │
│  │  • Metrics Collection                                   │     │
│  └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      微服务层                                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│  │  User Service   │  │ Order Service   │  │ Notification    │   │
│  │                 │  │                 │  │   Service       │   │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │   │
│  │ │ Pod 1       │ │  │ │ Pod 1       │ │  │ │ Pod 1       │ │   │
│  │ │ :8080       │ │  │ │ :8080       │ │  │ │ :8080       │ │   │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │   │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │   │
│  │ │ Pod 2       │ │  │ │ Pod 2       │ │  │ │ Pod 2       │ │   │
│  │ │ :8080       │ │  │ │ :8080       │ │  │ │ :8080       │ │   │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │   │
│  │                 │  │                 │  │                 │   │
│  │ Features:       │  │ Features:       │  │ Features:       │   │
│  │ • User Auth     │  │ • Order CRUD    │  │ • Email/SMS     │   │
│  │ • Profile Mgmt  │  │ • Payment       │  │ • Push Notify   │   │
│  │ • JWT Token     │  │ • Inventory     │  │ • Event Stream  │   │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                │               │               │
                ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      数据存储层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│  │   PostgreSQL    │  │   PostgreSQL    │  │     Redis       │   │
│  │   (User DB)     │  │   (Order DB)    │  │    (Cache)      │   │
│  │                 │  │                 │  │                 │   │
│  │ Tables:         │  │ Tables:         │  │ Data:           │   │
│  │ • users         │  │ • orders        │  │ • Sessions      │   │
│  │ • roles         │  │ • order_items   │  │ • Cache Data    │   │
│  │ • permissions   │  │ • payments      │  │ • Rate Limits   │   │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      配置管理层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│  │ Gateway Config  │  │ User Service    │  │ Order Service   │   │
│  │   ConfigMap     │  │   ConfigMap     │  │   ConfigMap     │   │
│  │                 │  │                 │  │                 │   │
│  │ • Routes        │  │ • DB Config     │  │ • DB Config     │   │
│  │ • Rate Limits   │  │ • JWT Secret    │  │ • Payment API   │   │
│  │ • CORS Policy   │  │ • Email Config  │  │ • Inventory API │   │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      监控观测层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│  │   Prometheus    │  │     Grafana     │  │   ELK Stack     │   │
│  │   (Metrics)     │  │  (Dashboard)    │  │    (Logs)       │   │
│  │                 │  │                 │  │                 │   │
│  │ • App Metrics   │  │ • Service Dash  │  │ • App Logs      │   │
│  │ • JVM Metrics   │  │ • Infra Dash    │  │ • Access Logs   │   │
│  │ • Custom KPIs   │  │ • Alert Rules   │  │ • Error Logs    │   │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.6.3 微服务通信流程图

```text
┌─────────────────────────────────────────────────────────────────┐
│                    微服务请求处理流程                             │
└─────────────────────────────────────────────────────────────────┘

     ┌─────────────┐
     │   Client    │
     │  Request    │
     └─────────────┘
            │ 1. HTTP Request
            ▼
     ┌─────────────┐
     │   Ingress   │ 2. Route to Gateway
     │ Controller  │ ────────────────────┐
     └─────────────┘                     │
                                         ▼
                                  ┌─────────────┐
                                  │  Gateway    │
                                  │  Service    │
                                  └─────────────┘
                                         │
                        3. Authentication & Authorization
                                         │
                        ┌────────────────┼────────────────┐
                        ▼                ▼                ▼
                 ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
                 │User Service │  │Order Service│  │Notification │
                 │             │  │             │  │  Service    │
                 └─────────────┘  └─────────────┘  └─────────────┘
                        │                │                │
                        ▼                ▼                ▼
                 ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
                 │  User DB    │  │  Order DB   │  │   Redis     │
                 │(PostgreSQL) │  │(PostgreSQL) │  │  (Cache)    │
                 └─────────────┘  └─────────────┘  └─────────────┘

     4. Service Discovery & Load Balancing (Kubernetes Service)
     5. Database Operations & Caching
     6. Response Aggregation & Return
```

#### 4.6.3 实验步骤

**步骤 1：创建命名空间和配置：**

```yaml
# microservices-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: microservices
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-config
  namespace: microservices
data:
  application.yml: |
    spring:
      cloud:
        gateway:
          routes:
          - id: user-service
            uri: http://user-service:8080
            predicates:
            - Path=/api/users/**
          - id: order-service
            uri: http://order-service:8080
            predicates:
            - Path=/api/orders/**
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: user-service-config
  namespace: microservices
data:
  application.yml: |
    spring:
      datasource:
        url: jdbc:postgresql://postgres:5432/userdb
        username: user
        password: password
```

**步骤 2：部署微服务应用：**

```yaml
# microservices-apps.yaml
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: gateway-service
  namespace: microservices
spec:
  image: "your-registry/gateway-service:latest"
  replicas: 2
  config:
    configMapRef:
      name: gateway-config
  ingress:
    enabled: true
    host: "api.microservices.local"
---
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: user-service
  namespace: microservices
spec:
  image: "your-registry/user-service:latest"
  replicas: 2
  config:
    configMapRef:
      name: user-service-config
---
apiVersion: springboot.tutorial.example.com/v1
kind: SpringBootApp
metadata:
  name: order-service
  namespace: microservices
spec:
  image: "your-registry/order-service:latest"
  replicas: 2
  config:
    configMapRef:
      name: user-service-config  # 复用配置
```

**步骤 3：部署和测试：**

```bash
# 创建命名空间和配置
kubectl apply -f microservices-namespace.yaml

# 部署微服务应用
kubectl apply -f microservices-apps.yaml

# 查看部署状态
kubectl get springbootapp -n microservices
kubectl get pods -n microservices
kubectl get ingress -n microservices

# 测试网关访问
echo "127.0.0.1 api.microservices.local" | sudo tee -a /etc/hosts
curl -H "Host: api.microservices.local" http://localhost/api/users/health
```

## 5. 总结

通过这些实验，我们完成了一个功能完整的 Spring Boot Operator 的开发和测试：

> **🎯 完整代码**：所有实验的代码和配置文件都在 [`code-examples`](./code-examples/) 目录中。

### 5.1 学习路径总览

```text
┌─────────────────────────────────────────────────────────────────┐
│                    Spring Boot Operator 学习路径                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   理论基础      │    │   实践开发      │    │   生产应用      │
│                 │    │                 │    │                 │
│ • Operator 概念 │───►│ • 基础 Operator │───►│ • 性能优化      │
│ • CRD 设计      │    │ • 配置管理      │    │ • 安全加固      │
│ • Controller    │    │ • 服务暴露      │    │ • 监控告警      │
│ • Spring Boot   │    │ • 微服务架构    │    │ • 故障恢复      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ 知识点掌握      │    │ 技能点实现      │    │ 工程化能力      │
│                 │    │                 │    │                 │
│ ✓ Kubernetes    │    │ ✓ Go 开发       │    │ ✓ CI/CD 集成    │
│ ✓ YAML 配置     │    │ ✓ API 设计      │    │ ✓ 多环境部署    │
│ ✓ 容器化        │    │ ✓ 事件处理      │    │ ✓ 版本管理      │
│ ✓ 微服务        │    │ ✓ 状态管理      │    │ ✓ 团队协作      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 5.2 技术栈总览

```text
┌─────────────────────────────────────────────────────────────────┐
│                      技术栈架构图                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      开发工具层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │   Kubebuilder│  │   Operator  │  │    Make     │  │   Git   │  │
│  │   Framework  │  │     SDK     │  │   Build     │  │ Version │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      编程语言层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │     Go      │  │    YAML     │  │    JSON     │  │  Shell  │  │
│  │  Language   │  │ Manifests   │  │   Config    │  │ Scripts │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Kubernetes 层                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │    CRD      │  │ Controller  │  │   Service   │  │ Ingress │  │
│  │ Definition  │  │  Manager    │  │ Discovery   │  │ Gateway │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │ ConfigMap   │  │   Secret    │  │ Deployment  │  │   Pod   │  │
│  │ Management  │  │ Management  │  │ Management  │  │ Lifecycle│  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      应用服务层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │Spring Boot  │  │   Gateway   │  │    User     │  │  Order  │  │
│  │  Operator   │  │   Service   │  │  Service    │  │ Service │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      基础设施层                                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │  Container  │  │   Network   │  │   Storage   │  │ Monitor │  │
│  │   Runtime   │  │   Policy    │  │  Volumes    │  │ & Logs  │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 5.3 核心收获

#### 理论知识

- **Operator 模式**：理解了 Kubernetes Operator 的设计理念和工作原理
- **CRD 设计**：掌握了自定义资源定义的最佳实践
- **Controller 模式**：学会了事件驱动的控制器开发
- **微服务架构**：了解了云原生微服务的部署和管理

#### 实践技能

- **Go 语言开发**：使用 Kubebuilder 框架进行 Operator 开发
- **Kubernetes API**：熟练使用 client-go 操作 Kubernetes 资源
- **配置管理**：实现了 ConfigMap 和 Secret 的自动化管理
- **服务暴露**：掌握了 Service 和 Ingress 的配置和管理

#### 工程能力

- **项目结构**：学会了标准的 Operator 项目组织方式
- **测试验证**：掌握了 Operator 的测试和验证方法
- **部署运维**：了解了 Operator 的部署和生产环境运维
- **问题排查**：具备了 Kubernetes 环境下的问题诊断能力

### 5.4 扩展方向

#### 功能增强

- **自动扩缩容**：基于 HPA/VPA 实现应用的自动伸缩
- **蓝绿部署**：支持零停机的应用更新策略
- **金丝雀发布**：实现渐进式的应用发布流程
- **多环境管理**：支持开发、测试、生产环境的差异化配置

#### 运维集成

- **监控告警**：集成 Prometheus 和 Grafana 实现全面监控
- **日志聚合**：使用 ELK Stack 进行日志收集和分析
- **链路追踪**：集成 Jaeger 或 Zipkin 实现分布式追踪
- **安全加固**：实现 RBAC、网络策略和安全扫描

#### 生态集成

- **服务网格**：与 Istio 或 Linkerd 集成实现高级流量管理
- **GitOps**：与 ArgoCD 或 Flux 集成实现声明式部署
- **多集群**：支持跨集群的应用部署和管理
- **云原生**：与云厂商的托管 Kubernetes 服务深度集成

### 5.5 最佳实践总结

#### 开发阶段

1. **API 设计优先**：先设计好 CRD 结构，再实现 Controller 逻辑
2. **渐进式开发**：从简单功能开始，逐步增加复杂特性
3. **充分测试**：编写单元测试和集成测试确保代码质量
4. **文档完善**：维护清晰的 API 文档和使用说明

#### 部署阶段

1. **资源限制**：合理设置 CPU 和内存限制
2. **权限最小化**：只授予必要的 RBAC 权限
3. **健康检查**：配置适当的存活性和就绪性探针
4. **监控覆盖**：确保关键指标都有监控覆盖

#### 运维阶段

1. **版本管理**：使用语义化版本管理 Operator 发布
2. **升级策略**：制定清晰的升级和回滚策略
3. **故障恢复**：建立完善的故障处理和恢复机制
4. **性能优化**：持续监控和优化 Operator 性能

---
