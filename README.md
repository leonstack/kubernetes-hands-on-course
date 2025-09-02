# Kubernetes 实战课程

## 1. 课程简介

欢迎来到 Kubernetes 实战课程！本课程提供了从零开始学习 Kubernetes 的完整路径，通过循序渐进的实践教程，帮助您掌握容器编排的核心技能。课程涵盖了从基础概念到生产级部署的所有关键知识点。

## 2. 课程特色

- **实战导向**：每个章节都包含完整的实践案例
- **循序渐进**：从容器基础到企业级平台，逐步深入
- **内容全面**：涵盖基础操作、高级特性和生产实践
- **工具齐全**：提供完整的配置文件和演示脚本
- **最佳实践**：融入生产环境的经验和最佳实践
- **前沿技术**：包含 DRA、HPA、Operator 等最新特性
- **企业级视角**：涵盖平台建设和扩展机制
- **故障排除**：包含常见问题的诊断和解决方案

## 3. 学习目标

完成本课程后，您将能够：

✅ **构建容器镜像**：掌握 Docker 镜像构建和优化技术  
✅ **理解 Kubernetes 架构**：掌握核心组件和工作原理  
✅ **管理容器化应用**：熟练使用 Pod、ReplicaSet、Deployment  
✅ **配置网络服务**：掌握 Service 的各种类型和使用场景  
✅ **编写 YAML 配置**：能够编写和维护复杂的 Kubernetes 配置  
✅ **实施最佳实践**：应用生产级的部署和运维策略  
✅ **掌握高级特性**：理解 HPA、DRA、CSI、Operator 等高级功能  
✅ **平台化建设**：具备企业级 Kubernetes 平台的设计和实施能力  
✅ **故障排除能力**：具备诊断和解决常见问题的能力  

## 4. 课程大纲

### 第零部分：Docker 镜像构建

#### 0.1 [Docker 镜像构建](./00-Docker-Images/)

**学习重点：** 为 Kubernetes 应用构建优化的 Docker 镜像

- **[kubenginx](./00-Docker-Images/01-kubenginx/)**：多版本 Nginx 镜像构建实践
- **[Spring Boot 后端](./00-Docker-Images/02-kube-backend-helloworld-springboot/)**：Java 应用容器化最佳实践
- **[前端 Nginx](./00-Docker-Images/03-kube-frontend-nginx/)**：静态资源容器化部署

**适用场景：** 应用容器化、镜像优化、多阶段构建

---

### 第一部分：基础概念与架构

#### 1.1 [Kubernetes 架构](./01-Kubernetes-Architecture/)

**学习重点：** 理解 Kubernetes 的整体架构和核心组件

- **传统部署挑战**：了解容器化和编排的必要性
- **核心组件介绍**：Master 节点、Worker 节点、etcd、API Server
- **网络架构**：Pod 网络、Service 网络、Ingress
- **设计理念**：声明式管理、控制器模式、自愈机制

**适用场景：** 架构设计、技术选型、团队培训

---

### 第二部分：kubectl 命令行实践

#### 2.1 [PODs with kubectl](./02-PODs-with-kubectl/)

**学习重点：** 掌握 Pod 的基本概念和 kubectl 操作

- **Pod 基础**：理解 Pod 的概念和生命周期
- **kubectl 操作**：创建、查看、删除 Pod
- **多容器 Pod**：Sidecar、Ambassador、Adapter 模式
- **调试技巧**：日志查看、容器进入、端口转发

**适用场景：** 开发调试、故障排除、基础运维

#### 2.2 [ReplicaSets with kubectl](./03-ReplicaSets-with-kubectl/)

**学习重点：** 理解副本控制和高可用性

- **副本管理**：自动维护指定数量的 Pod 副本
- **故障恢复**：Pod 失败时的自动重建机制
- **标签选择器**：基于标签的 Pod 管理
- **扩缩容操作**：动态调整副本数量

**适用场景：** 高可用部署、负载分担、故障恢复

#### 2.3 [Deployments with kubectl](./04-Deployments-with-kubectl/)

**学习重点：** 掌握应用部署和版本管理

- **创建和扩展**：Deployment 基础操作和服务暴露
- **滚动更新**：零停机时间的应用程序更新
- **版本回滚**：快速回退到之前的稳定版本
- **暂停和恢复**：控制部署过程的执行

**适用场景：** 生产部署、版本管理、持续交付

#### 2.4 [Services with kubectl](./05-Services-with-kubectl/)

**学习重点：** 实现服务发现和负载均衡

- **Service 类型**：ClusterIP、NodePort、LoadBalancer、ExternalName
- **服务发现**：集群内部服务的自动发现机制
- **负载均衡**：流量在多个 Pod 间的分发
- **完整架构**：前后端分离应用的网络架构

**适用场景：** 微服务架构、服务网格、外部访问

---

### 第三部分：YAML 声明式管理

#### 3.1 [YAML 基础](./06-YAML-Basics/)

**学习重点：** 掌握 YAML 语法和最佳实践

- **基础语法**：缩进、键值对、列表、注释
- **数据类型**：标量、序列、映射、多行字符串
- **常见陷阱**：缩进错误、特殊字符、类型转换
- **工具推荐**：编辑器配置、验证工具、格式化工具

**适用场景：** 配置管理、基础设施即代码、CI/CD 流水线

#### 3.2 [PODs with YAML](./07-PODs-with-YAML/)

**学习重点：** 使用 YAML 文件管理 Pod 资源

- **YAML 结构**：apiVersion、kind、metadata、spec
- **Pod 配置**：容器定义、资源限制、环境变量
- **网络配置**：端口映射、主机网络、DNS 设置
- **安全配置**：安全上下文、权限控制、密钥管理

**适用场景：** 声明式管理、版本控制、自动化部署

#### 3.3 [ReplicaSets with YAML](./08-ReplicaSets-with-YAML/)

**学习重点：** 声明式副本控制和标签管理

- **标签选择器**：matchLabels 和 matchExpressions
- **副本策略**：期望副本数、最大不可用数
- **Pod 模板**：统一的 Pod 配置模板
- **故障恢复**：自动检测和替换失败的 Pod

**适用场景：** 高可用架构、自动化运维、容灾备份

#### 3.4 [Deployments with YAML](./09-Deployments-with-YAML/)

**学习重点：** 生产级应用部署策略

- **部署策略**：RollingUpdate、Recreate 策略配置
- **更新控制**：maxSurge、maxUnavailable 参数
- **版本历史**：revisionHistoryLimit 和回滚机制
- **健康检查**：readinessProbe、livenessProbe 配置

**适用场景：** 生产部署、持续集成、发布管理

#### 3.5 [Services with YAML](./10-Services-with-YAML/)

**学习重点：** 完整的服务网络架构

- **ClusterIP Service**：集群内部服务发现
- **NodePort Service**：外部访问入口
- **负载均衡**：会话亲和性、流量分发策略
- **前后端架构**：完整的微服务网络拓扑

**适用场景：** 微服务架构、API 网关、服务网格

---

### 第四部分：高级主题与生产实践

#### 4.1 [容器存储接口 (CSI)](./Advanced-Topics/CSI/)

**学习重点：** 理解和使用 Kubernetes 存储扩展机制

- **存储驱动**：CSI 驱动程序的工作原理
- **本地存储**：local-path 存储类的配置和使用
- **动态供应**：PV/PVC 的自动化管理

**适用场景：** 有状态应用、数据持久化、存储扩展

#### 4.2 [动态资源分配 (DRA)](./Advanced-Topics/DRA/)

**学习重点：** 掌握 Kubernetes 新一代资源管理机制

- **DRA 架构**：理解 DRA 与传统资源管理的区别
- **CDI 集成**：容器设备接口的协作机制
- **GPU 调度**：专用硬件资源的动态分配

**适用场景：** GPU 工作负载、AI/ML 应用、专用硬件管理

#### 4.3 [水平 Pod 自动扩缩 (HPA)](./Advanced-Topics/HPA/)

**学习重点：** 实现应用的自动化弹性伸缩

- **指标驱动**：基于 CPU、内存和自定义指标的扩缩容
- **架构设计**：HPA 控制器和 Metrics Server 的工作机制
- **生产实践**：扩缩容策略和性能优化

**适用场景：** 流量波动应用、成本优化、性能调优

#### 4.4 [Operator 开发](./Advanced-Topics/operator/)

**学习重点：** 开发和部署自定义 Kubernetes 控制器

- **Operator 模式**：理解 Operator 的设计理念
- **Spring Boot Operator**：实际的 Operator 开发案例
- **CRD 设计**：自定义资源定义的最佳实践

**适用场景：** 复杂应用管理、自动化运维、平台扩展

#### 4.5 [调度器深入](./Advanced-Topics/调度/)

**学习重点：** 理解和优化 Kubernetes 调度机制

- **调度基础**：调度器的工作原理和调度流程
- **高级调度**：亲和性、反亲和性、污点和容忍
- **GPU 调度**：专用硬件的调度策略和实践

**适用场景：** 性能优化、资源规划、多租户环境

#### 4.6 [KubeSphere 平台](./Advanced-Topics/KubeSphere-架构设计与扩展机制深度分析.md)

**学习重点：** 企业级 Kubernetes 平台的架构和扩展

- **架构设计**：KubeSphere 的多租户架构
- **扩展机制**：插件系统和自定义扩展
- **最佳实践**：企业级部署和运维经验

**适用场景：** 企业平台建设、多租户管理、DevOps 流水线

---

### 第五部分：工具与辅助资源

#### 5.1 [Kubernetes Dashboard](./99-other/K8s%20Dashboard/)

**学习重点：** 部署和使用 Kubernetes Web 界面

- **Dashboard 部署**：安全的 Dashboard 安装配置
- **权限管理**：RBAC 和只读访问控制
- **运维监控**：通过 Web 界面进行集群管理

**适用场景：** 可视化管理、团队协作、运维监控

#### 5.2 [课程演示](./presentation/)

**学习重点：** 课程配套的演示文稿和培训材料

- **Kubernetes 基础演示**：适用于团队培训和知识分享
- **架构图表**：可视化的技术架构说明
- **最佳实践总结**：生产环境的经验汇总

**适用场景：** 团队培训、技术分享、架构评审

---

## 4. 环境要求

### 4.1 基础环境

- **Kubernetes 集群**：v1.20+ （推荐 v1.25+）
- **kubectl 工具**：与集群版本兼容
- **操作系统**：Linux、macOS 或 Windows（WSL2）
- **内存要求**：至少 4GB RAM
- **存储空间**：至少 10GB 可用空间

### 4.2 推荐工具

- **代码编辑器**：VS Code + Kubernetes 扩展
- **终端工具**：支持多标签的终端（iTerm2、Windows Terminal）
- **容器运行时**：Docker Desktop 或 Podman
- **网络工具**：curl、wget、telnet

### 4.3 环境验证

```bash
# 检查 kubectl 版本
kubectl version --client

# 检查集群连接
kubectl cluster-info

# 检查节点状态
kubectl get nodes

# 检查当前上下文
kubectl config current-context
```

---

## 5. 快速开始

### 5.1 克隆课程仓库

```bash
git clone <repository-url>
cd kubernetes-fundamentals
```

### 5.2 验证环境

```bash
# 运行环境检查脚本
./scripts/check-environment.sh
```

### 5.3 开始学习

建议按照以下顺序学习：

1. **容器基础**：从 `00-Docker-Images` 开始，了解容器镜像构建
2. **理论基础**：学习 `01-Kubernetes-Architecture`，理解整体架构
3. **命令行实践**：完成 `02-PODs-with-kubectl` 到 `05-Services-with-kubectl`
4. **声明式管理**：学习 `06-YAML-Basics` 到 `10-Services-with-YAML`
5. **高级主题**：根据需要选择 `Advanced-Topics` 中的专题学习
6. **工具辅助**：配置 `99-other` 中的管理工具

---

## 6. 学习建议

### 6.1 学习路径

#### 6.1.1 初学者路径（6-8 周）

- 第 1 周：Docker 镜像构建 + 架构理解
- 第 2 周：Pod 基础 + kubectl 实践
- 第 3 周：ReplicaSet + Deployment
- 第 4 周：Service + 网络基础
- 第 5 周：YAML 基础 + 声明式管理
- 第 6 周：综合实践 + Dashboard 配置
- 第 7-8 周：选择性学习高级主题

#### 6.1.2 进阶路径（3-4 周）

- 第 1 周：快速回顾基础 + kubectl/YAML 实践
- 第 2 周：高级调度 + HPA + 存储管理
- 第 3 周：Operator 开发 + DRA 实践
- 第 4 周：生产部署 + 平台建设

#### 6.1.3 专家路径（2-3 周）

- 第 1 周：高级主题深入（DRA、调度器、Operator）
- 第 2 周：企业级平台（KubeSphere、扩展机制）
- 第 3 周：性能优化 + 故障排除

### 6.2 实践建议

- **动手实践**：每个概念都要亲自操作验证
- **记录笔记**：记录重要命令和配置模式
- **问题驱动**：遇到问题时深入研究原理
- **项目应用**：将学到的知识应用到实际项目中

---

**开始您的 Kubernetes 学习之旅吧！**

每一个伟大的容器编排专家都是从第一个 Pod 开始的。让我们一起探索 Kubernetes 的精彩世界！
