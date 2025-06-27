# Kubernetes 调度器扩展案例：GPU 资源调度

在现代云原生环境中，GPU 资源调度已成为支持机器学习、深度学习和高性能计算工作负载的关键需求。本文将通过一个完整的实际案例，展示如何扩展 Kubernetes 调度器来实现智能的 GPU 资源调度。

## 目录

- [Kubernetes 调度器扩展案例：GPU 资源调度](#kubernetes-调度器扩展案例gpu-资源调度)
  - [目录](#目录)
  - [1 需求分析](#1-需求分析)
    - [1.1 业务需求](#11-业务需求)
    - [1.2 技术需求](#12-技术需求)
    - [1.3 性能需求](#13-性能需求)
    - [1.4 运维需求](#14-运维需求)
    - [1.5 安全需求](#15-安全需求)
  - [2 GPU 资源表示](#2-gpu-资源表示)
    - [2.1 扩展资源定义](#21-扩展资源定义)
    - [2.2 Pod GPU 请求](#22-pod-gpu-请求)
    - [2.3 GPU 资源约束](#23-gpu-资源约束)
  - [3 GPU 资源发现](#3-gpu-资源发现)
    - [3.1 Device Plugin 工作原理](#31-device-plugin-工作原理)
    - [3.2 NVIDIA GPU Device Plugin 部署](#32-nvidia-gpu-device-plugin-部署)
    - [3.3 GPU 资源监控](#33-gpu-资源监控)
  - [4 GPU 调度器架构图](#4-gpu-调度器架构图)
    - [4.1 GPU 调度器整体架构](#41-gpu-调度器整体架构)
    - [4.2 Device Plugin 工作流程](#42-device-plugin-工作流程)
    - [4.3 GPU 调度决策流程](#43-gpu-调度决策流程)
    - [4.4 GPU 资源管理架构](#44-gpu-资源管理架构)
    - [4.5 多租户 GPU 资源隔离架构](#45-多租户-gpu-资源隔离架构)
    - [4.6 GPU 监控与运维架构](#46-gpu-监控与运维架构)
    - [4.7 GPU 调度器插件架构](#47-gpu-调度器插件架构)
    - [4.8 架构图说明](#48-架构图说明)
      - [4.8.1 整体架构特点](#481-整体架构特点)
      - [4.8.2 关键组件说明](#482-关键组件说明)
      - [4.8.3 工作流程要点](#483-工作流程要点)
      - [4.8.4 扩展性考虑](#484-扩展性考虑)
    - [4.9 实现路径](#49-实现路径)
  - [5 自定义调度器扩展](#5-自定义调度器扩展)
    - [5.1 GPU Filter 插件实现](#51-gpu-filter-插件实现)
    - [5.2 GPU Score 插件实现](#52-gpu-score-插件实现)
    - [5.3 调度器配置](#53-调度器配置)
  - [6 GPU 工作负载示例](#6-gpu-工作负载示例)
    - [6.1 机器学习训练任务](#61-机器学习训练任务)
    - [6.2 推理服务](#62-推理服务)
  - [7 性能优化](#7-性能优化)
    - [7.1 调度性能优化](#71-调度性能优化)
    - [7.2 基础监控配置](#72-基础监控配置)
    - [7.3 简化监控部署](#73-简化监控部署)
  - [8 安全加固与访问控制](#8-安全加固与访问控制)
    - [8.1 RBAC 配置](#81-rbac-配置)
    - [8.2 基础资源隔离](#82-基础资源隔离)
  - [9 故障处理与恢复](#9-故障处理与恢复)
    - [9.1 基础健康检查](#91-基础健康检查)
    - [9.2 简化调度器部署](#92-简化调度器部署)
  - [10 总结与最佳实践](#10-总结与最佳实践)
    - [10.1 需求实现总结](#101-需求实现总结)
    - [10.2 架构优势](#102-架构优势)
    - [10.3 最佳实践建议](#103-最佳实践建议)
    - [10.4 扩展路线图](#104-扩展路线图)
    - [10.5 成功指标](#105-成功指标)

---

## 1 需求分析

### 1.1 业务需求

**场景描述：**
某 AI 公司需要在 Kubernetes 集群中运行多种类型的 GPU 工作负载：

- **训练任务**：需要高性能 GPU（如 V100、A100），通常需要多卡并行
- **推理服务**：可以使用中等性能 GPU（如 T4、RTX 系列），延迟敏感
- **开发测试**：可以使用入门级 GPU，资源需求灵活

**挑战：**

1. **资源异构性**：集群中存在多种型号的 GPU
2. **调度复杂性**：需要考虑 GPU 类型、拓扑结构、利用率等因素
3. **资源利用率**：避免 GPU 资源浪费和碎片化
4. **性能优化**：确保高优先级任务获得最佳资源
5. **可维护性**：调度器需要与集群管理系统集成

### 1.2 技术需求

**功能需求：**

- **GPU 资源发现和管理**：自动发现集群中的 GPU 资源，实时监控 GPU 健康状态
- **基于 GPU 类型的智能调度**：根据 GPU 性能等级和工作负载特征进行匹配调度
- **GPU 拓扑感知调度**：支持单节点多卡和多节点多卡的拓扑优化调度
- **资源利用率优化**：通过智能算法减少 GPU 资源碎片化，提升整体利用率
- **多租户资源隔离**：基于命名空间和用户组的细粒度资源隔离机制
- **资源配额管理**：基于命名空间的 GPU 资源配额和限制机制
- **可扩展的自定义调度策略**：提供插件化架构支持业务特定的调度需求

**非功能需求：**

- **系统规模**：支持 1000+ GPU 节点的大规模集群
- **可用性**：99.9% 系统可用性，支持故障自动恢复
- **监控运维**：完整的性能监控、资源统计和智能告警体系
- **安全性**：支持 RBAC 权限控制和多租户资源隔离
- **扩展性**：插件化架构，支持自定义调度策略扩展

### 1.3 性能需求

**调度性能指标：**

```yaml
# GPU 调度器性能目标配置
# 定义了调度器在生产环境中需要达到的各项性能指标
performance_targets:
  # 调度延迟要求 - 衡量调度决策的响应时间
  scheduling_latency:
    simple_scheduling: "< 50ms"    # 单 GPU 简单调度的延迟上限
    complex_scheduling: "< 200ms"  # 多 GPU 拓扑感知调度的延迟上限
    p50: "< 50ms"                  # 50% 的调度请求在 50ms 内完成
    p95: "< 100ms"                 # 95% 的调度请求在 100ms 内完成
    p99: "< 200ms"                 # 99% 的调度请求在 200ms 内完成
  
  # 调度吞吐量 - 衡量调度器的处理能力
  throughput:
    pods_per_second: "> 100"       # 每秒处理的 Pod 调度请求数量
    concurrent_scheduling: "> 50"  # 并发处理的调度请求数量
    scheduling_success_rate: "> 95%" # 调度成功率，避免调度失败导致的重试
  
  # 资源利用率 - 衡量 GPU 资源的有效使用
  resource_utilization:
    gpu_utilization: "> 80%"       # GPU 计算资源利用率目标
    memory_utilization: "> 75%"    # GPU 内存利用率目标
    gpu_fragmentation_rate: "< 10%" # GPU 资源碎片化率，避免资源浪费
    
  # 系统可用性 - 衡量调度器服务的稳定性
  availability:
    uptime: "99.9%"                # 系统正常运行时间百分比
    mttr: "< 5min"                 # 平均故障恢复时间（Mean Time To Recovery）
    mtbf: "> 720h"                # 平均故障间隔时间（Mean Time Between Failures）
```

### 1.4 运维需求

**监控需求：**

- **GPU 资源监控**：GPU 利用率、内存使用率、温度等核心指标
- **调度性能监控**：调度延迟、成功率、队列长度等关键指标
- **告警机制**：GPU 故障、调度异常、资源不足的自动告警
- **可视化展示**：Grafana 仪表板展示 GPU 集群状态

**运维需求：**

- **部署管理**：
  - Helm Chart 标准化部署
  - 配置文件模板化管理
  - 滚动更新和回滚策略
- **故障处理**：
  - GPU 节点健康检查
  - 自动故障节点隔离
  - 调度器重启和恢复机制
- **日志管理**：
  - 结构化日志输出
  - 集中式日志收集
  - 问题排查工具和流程

### 1.5 安全需求

**访问控制：**

- **RBAC 权限管理**：细粒度的角色和权限定义
- **服务账户隔离**：调度器专用服务账户和最小权限原则
- **API 访问控制**：限制调度器 API 的访问范围

**资源安全：**

- **网络策略**：Pod 间通信的网络隔离规则
- **镜像安全**：使用可信镜像仓库和镜像签名验证

---

## 2 GPU 资源表示

### 2.1 扩展资源定义

Kubernetes 使用扩展资源（Extended Resources）来表示 GPU：

```yaml
# 节点 GPU 资源定义示例
# 展示了 Kubernetes 节点如何声明和管理 GPU 资源
apiVersion: v1
kind: Node
metadata:
  name: gpu-node-1                    # 节点名称，用于标识具体的 GPU 节点
  labels:
    accelerator: nvidia-tesla-v100     # GPU 型号标签，用于调度时的 GPU 类型匹配
    gpu-count: "8"                     # GPU 数量标签，便于调度器快速筛选
    gpu-memory: "32Gi"                 # 单卡 GPU 内存大小标签
status:
  capacity:                            # 节点总容量，包含所有可用的 GPU 资源
    nvidia.com/gpu: "8"               # 节点上 GPU 卡的总数量
    nvidia.com/gpu-memory: "256Gi"    # 节点上 GPU 内存总容量（8 * 32Gi）
  allocatable:                         # 可分配容量，扣除系统保留后的可用资源
    nvidia.com/gpu: "8"               # 可分配的 GPU 卡数量
    nvidia.com/gpu-memory: "256Gi"    # 可分配的 GPU 内存总量
```

### 2.2 Pod GPU 请求

```yaml
# Pod GPU 资源请求示例
# 展示了如何在 Pod 中请求特定数量和类型的 GPU 资源
apiVersion: v1
kind: Pod
metadata:
  name: gpu-training-job              # Pod 名称，通常用于机器学习训练任务
spec:
  containers:
  - name: tensorflow                  # 容器名称
    image: tensorflow/tensorflow:2.8.0-gpu  # 支持 GPU 的 TensorFlow 镜像
    resources:
      requests:                       # 资源请求，调度器据此进行节点选择
        nvidia.com/gpu: "4"          # 请求 4 块 GPU 卡用于并行训练
        nvidia.com/gpu-memory: "128Gi" # 请求 128GB GPU 内存（4 * 32GB）
      limits:                         # 资源限制，防止容器超量使用资源
        nvidia.com/gpu: "4"          # 限制最多使用 4 块 GPU
        nvidia.com/gpu-memory: "128Gi" # 限制最多使用 128GB GPU 内存
  nodeSelector:                       # 节点选择器，确保调度到合适的 GPU 节点
    accelerator: nvidia-tesla-v100    # 指定 GPU 类型为 V100，适合高性能训练
```

### 2.3 GPU 资源约束

```yaml
# GPU 亲和性和约束配置示例
# 展示了如何使用节点亲和性来精确控制 Pod 的 GPU 调度
apiVersion: v1
kind: Pod
metadata:
  name: multi-gpu-job                 # 多 GPU 训练任务 Pod
spec:
  affinity:                           # 亲和性配置，用于精确控制调度行为
    nodeAffinity:                     # 节点亲和性，基于节点标签进行调度
      requiredDuringSchedulingIgnoredDuringExecution:  # 硬性要求，调度时必须满足
        nodeSelectorTerms:            # 节点选择条件列表
        - matchExpressions:           # 匹配表达式，支持多种操作符
          - key: accelerator          # 基于 GPU 型号标签进行筛选
            operator: In              # 操作符：包含于指定值列表中
            values: ["nvidia-tesla-v100", "nvidia-tesla-a100"]  # 支持的高性能 GPU 型号
          - key: gpu-count            # 基于 GPU 数量标签进行筛选
            operator: Gt              # 操作符：大于指定值
            values: ["4"]             # 要求节点至少有 5 块 GPU（大于4）
  containers:
  - name: training                    # 训练容器
    image: pytorch/pytorch:1.12.0-cuda11.3-cudnn8-runtime  # PyTorch GPU 运行时镜像
    resources:
      requests:
        nvidia.com/gpu: "8"          # 请求 8 块 GPU 进行大规模并行训练
```

## 3 GPU 资源发现

### 3.1 Device Plugin 工作原理

Kubernetes Device Plugin 框架负责 GPU 资源的发现和管理：

```mermaid
sequenceDiagram
    participant DP as Device Plugin
    participant K as Kubelet
    participant S as Scheduler
    participant P as Pod
    
    DP->>K: Register()
    K->>DP: ListAndWatch()
    DP->>K: Device List
    K->>S: Node Status Update
    S->>K: Schedule Pod
    K->>DP: Allocate()
    DP->>K: Device Assignment
    K->>P: Start Container
```

### 3.2 NVIDIA GPU Device Plugin 部署

```yaml
# NVIDIA Device Plugin DaemonSet 配置
# 负责在每个 GPU 节点上运行 Device Plugin，实现 GPU 资源的发现和管理
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nvidia-device-plugin-daemonset  # DaemonSet 名称
  namespace: kube-system                 # 部署在系统命名空间
spec:
  selector:
    matchLabels:
      name: nvidia-device-plugin-ds      # 用于选择 Pod 的标签
  updateStrategy:
    type: RollingUpdate                  # 滚动更新策略，确保服务连续性
  template:
    metadata:
      labels:
        name: nvidia-device-plugin-ds     # Pod 标签，与 selector 匹配
    spec:
      tolerations:                       # 容忍配置，允许在有 GPU 污点的节点上运行
      - key: nvidia.com/gpu              # 容忍 GPU 相关的污点
        operator: Exists                 # 只要污点存在就容忍
        effect: NoSchedule               # 容忍 NoSchedule 效果的污点
      priorityClassName: system-node-critical  # 高优先级，确保系统关键组件优先调度
      containers:
      - image: nvcr.io/nvidia/k8s-device-plugin:v0.12.0  # NVIDIA 官方 Device Plugin 镜像
        name: nvidia-device-plugin-ctr   # 容器名称
        args: ["--fail-on-init-error=false"]  # 启动参数，初始化失败时不退出
        securityContext:                 # 安全上下文配置
          allowPrivilegeEscalation: false # 禁止权限提升
          capabilities:
            drop: ["ALL"]                # 移除所有 Linux capabilities
        volumeMounts:
        - name: device-plugin             # 挂载 Device Plugin 套接字目录
          mountPath: /var/lib/kubelet/device-plugins  # kubelet Device Plugin 通信目录
      volumes:
      - name: device-plugin              # 定义 hostPath 卷
        hostPath:
          path: /var/lib/kubelet/device-plugins  # 主机上的 Device Plugin 目录
      nodeSelector:                      # 节点选择器，只在 NVIDIA GPU 节点上运行
        accelerator: nvidia              # 选择带有 nvidia 标签的节点
```

### 3.3 GPU 资源监控

```bash
#!/bin/bash
# GPU 资源监控和查看命令集合
# 提供了一系列用于监控和管理 GPU 资源的 kubectl 命令

# 1. 查看集群中所有节点的 GPU 资源配置
# 显示每个节点的 GPU 容量和可分配数量
kubectl describe nodes | grep -A 5 "nvidia.com/gpu"

# 2. 查看 GPU 节点的资源使用情况
# 显示 GPU 节点的 CPU、内存等资源使用率
kubectl top nodes --selector=accelerator=nvidia

# 3. 查看特定 GPU 节点上运行的 Pod 分布
# 了解 GPU 资源的实际分配和使用情况
kubectl get pods --all-namespaces -o wide \
  --field-selector spec.nodeName=gpu-node-1

# 4. 查看所有 GPU 相关的 Pod 及其资源请求
kubectl get pods --all-namespaces -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].resources.requests.nvidia\.com/gpu}{"\n"}{end}' | grep -v "^.*\t$"

# 5. 查看 GPU Device Plugin 的运行状态
kubectl get pods -n kube-system -l name=nvidia-device-plugin-ds

# 6. 查看 GPU 节点的详细标签信息
kubectl get nodes -l accelerator=nvidia --show-labels
```

---

## 4 GPU 调度器架构图

本章节提供 Kubernetes GPU 调度器相关的架构图，展示了 GPU 资源调度的完整技术架构和工作流程。

### 4.1 GPU 调度器整体架构

本架构图展示了 Kubernetes GPU 调度器的完整系统架构，包括控制平面、工作节点、监控系统和工作负载的交互关系。

**架构核心组件：**

- **控制平面**：包含 API Server、etcd、Controller Manager 和自定义 GPU 调度器
- **GPU 调度器组件**：GPUFilter 插件、GPUScore 插件和调度器配置
- **工作节点**：配备不同类型 GPU 的节点，运行 kubelet、Device Plugin 和各类工作负载
- **监控运维**：Prometheus、Grafana 和 AlertManager 组成的可观测性体系

**关键特性：**

1. **异构 GPU 支持**：支持 V100、T4、A100 等不同性能等级的 GPU
2. **智能调度**：基于 GPU 类型、利用率、拓扑等多维度进行调度决策
3. **工作负载感知**：针对训练、推理、开发等不同场景优化调度策略
4. **全面监控**：实时监控 GPU 使用情况，及时发现性能瓶颈和异常

```mermaid
graph TB
    subgraph "Control Plane"
        API["API Server"]
        ETCD["etcd"]
        CM["Controller Manager"]
        SCHED["GPU Scheduler"]
    end
    
    subgraph "GPU Scheduler Components"
        FILTER["GPU Filter Plugin"]
        SCORE["GPU Score Plugin"]
        CONFIG["Scheduler Config"]
    end
    
    subgraph "Worker Nodes"
        subgraph "GPU Node 1"
            K1["kubelet"]
            DP1["Device Plugin"]
            GPU1["GPU: V100 x8"]
            POD1["Training Pods"]
        end
        
        subgraph "GPU Node 2"
            K2["kubelet"]
            DP2["Device Plugin"]
            GPU2["GPU: T4 x4"]
            POD2["Inference Pods"]
        end
        
        subgraph "GPU Node 3"
            K3["kubelet"]
            DP3["Device Plugin"]
            GPU3["GPU: A100 x8"]
            POD3["ML Training Pods"]
        end
    end
    
    subgraph "Monitoring & Management"
        PROM["Prometheus"]
        GRAF["Grafana"]
        ALERT["AlertManager"]
    end
    
    subgraph "Workloads"
        TRAIN["Training Jobs"]
        INFER["Inference Services"]
        DEV["Development Pods"]
    end
    
    %% Control Plane Connections
    API --> ETCD
    API --> SCHED
    SCHED --> CONFIG
    SCHED --> FILTER
    SCHED --> SCORE
    
    %% Node Connections
    API --> K1
    API --> K2
    API --> K3
    
    K1 --> DP1
    K2 --> DP2
    K3 --> DP3
    
    DP1 --> GPU1
    DP2 --> GPU2
    DP3 --> GPU3
    
    %% Workload Scheduling
    TRAIN --> SCHED
    INFER --> SCHED
    DEV --> SCHED
    
    SCHED --> POD1
    SCHED --> POD2
    SCHED --> POD3
    
    %% Monitoring
    K1 --> PROM
    K2 --> PROM
    K3 --> PROM
    PROM --> GRAF
    PROM --> ALERT
    
    classDef gpu fill:#ff9999
    classDef scheduler fill:#99ccff
    classDef monitoring fill:#99ff99
    
    class GPU1,GPU2,GPU3 gpu
    class SCHED,FILTER,SCORE scheduler
    class PROM,GRAF,ALERT monitoring
```

### 4.2 Device Plugin 工作流程

本序列图详细展示了 NVIDIA Device Plugin 与 Kubernetes 系统各组件的交互流程，这是 GPU 资源管理的核心机制。

**工作流程阶段：**

1. **注册阶段**：Device Plugin 向 kubelet 注册，声明管理的资源类型（如 nvidia.com/gpu）
2. **资源发现阶段**：持续扫描和上报 GPU 设备信息，更新节点资源状态
3. **调度阶段**：GPU 调度器根据资源信息进行智能调度决策
4. **资源分配阶段**：为容器分配具体的 GPU 设备，配置设备路径和环境变量
5. **运行时监控**：持续监控 GPU 设备健康状态，及时上报异常
6. **资源释放**：Pod 终止时释放 GPU 资源，供其他工作负载使用

**关键技术点：**

- **gRPC 通信**：Device Plugin 与 kubelet 通过 gRPC 协议通信
- **设备隔离**：通过 cgroups 和设备文件实现 GPU 设备隔离
- **健康检查**：定期检查 GPU 设备状态，确保资源可用性
- **动态发现**：支持热插拔，动态发现新增或移除的 GPU 设备

```mermaid
sequenceDiagram
    participant DP as Device Plugin
    participant K as Kubelet
    participant API as API Server
    participant S as GPU Scheduler
    participant P as Pod
    
    Note over DP,K: 1. 注册阶段
    DP->>K: Register(endpoint, resourceName)
    K->>DP: ACK
    
    Note over DP,K: 2. 资源发现阶段
    K->>DP: ListAndWatch()
    DP->>DP: 扫描 GPU 设备
    DP->>K: Device List (GPU 信息)
    K->>API: Node Status Update (GPU 资源)
    
    Note over API,S: 3. 调度阶段
    P->>API: Pod Creation Request
    API->>S: Pod Scheduling Event
    S->>S: GPU Filter & Score
    S->>API: Binding Decision
    API->>K: Pod Assignment
    
    Note over K,DP: 4. 资源分配阶段
    K->>DP: Allocate(deviceIDs, containerRequests)
    DP->>DP: 分配 GPU 设备
    DP->>K: AllocateResponse (设备路径、环境变量)
    K->>P: Start Container (GPU 设备挂载)
    
    Note over DP,P: 5. 运行时监控
    loop 持续监控
        DP->>DP: 监控 GPU 健康状态
        DP->>K: Device Health Update
        K->>API: Node Status Update
    end
    
    Note over K,DP: 6. 资源释放
    P->>K: Pod Termination
    K->>DP: PreStopContainer()
    DP->>DP: 释放 GPU 资源
    DP->>K: 资源释放确认
```

### 4.3 GPU 调度决策流程

本流程图展示了 GPU 调度器的完整决策过程，从 Pod 调度请求到最终绑定的每个关键步骤。

**调度决策阶段：**

**1. 预检查阶段：**

- 判断 Pod 是否需要 GPU 资源
- 不需要 GPU 的 Pod 走常规调度流程

**2. Filter 过滤阶段：**

- **GPU 容量检查**：确保节点有足够的 GPU 资源
- **GPU 类型匹配**：验证节点 GPU 类型符合 Pod 要求
- **可用性检查**：确认当前可用的 GPU 数量
- **拓扑约束**：检查 GPU 拓扑和亲和性要求
- **资源配额**：验证命名空间资源配额限制

**3. Score 评分阶段：**

- **利用率评分（40%）**：优先选择资源利用率均衡的节点
- **类型匹配评分（30%）**：优先选择 GPU 类型完全匹配的节点
- **拓扑评分（20%）**：优化多卡任务的 GPU 间通信性能
- **碎片化评分（10%）**：减少资源碎片化，提高整体利用率

**4. 绑定执行阶段：**

- 选择评分最高的节点
- 通过 Device Plugin 分配具体 GPU 设备
- 完成 Pod 到节点的绑定

**智能调度特性：**

- **多维度评分**：综合考虑性能、利用率、拓扑等因素
- **工作负载感知**：针对训练、推理等不同场景优化
- **资源均衡**：避免资源热点，实现集群负载均衡

```mermaid
flowchart TD
    START(["Pod 调度请求"]) --> CHECK_GPU{"需要 GPU？"}
    
    CHECK_GPU -->|否| NORMAL["常规调度流程"]
    CHECK_GPU -->|是| GPU_FILTER["GPU Filter 阶段"]
    
    GPU_FILTER --> FILTER_STEPS["过滤步骤"]
    
    subgraph "Filter 过滤阶段"
        FILTER_STEPS --> F1["检查节点 GPU 容量"]
        F1 --> F2["检查 GPU 类型匹配"]
        F2 --> F3["检查可用 GPU 数量"]
        F3 --> F4["检查 GPU 拓扑约束"]
        F4 --> F5["检查资源配额"]
    end
    
    F5 --> FILTER_RESULT{"过滤结果"}
    FILTER_RESULT -->|无可用节点| FAIL["调度失败"]
    FILTER_RESULT -->|有候选节点| GPU_SCORE["GPU Score 阶段"]
    
    GPU_SCORE --> SCORE_STEPS["评分步骤"]
    
    subgraph "Score 评分阶段"
        SCORE_STEPS --> S1["GPU 利用率评分 (40%)"]
        S1 --> S2["GPU 类型匹配评分 (30%)"]
        S2 --> S3["GPU 拓扑评分 (20%)"]
        S3 --> S4["碎片化评分 (10%)"]
    end
    
    S4 --> NORMALIZE["评分标准化"]
    NORMALIZE --> SELECT["选择最高分节点"]
    SELECT --> BIND["绑定 Pod 到节点"]
    
    BIND --> ALLOCATE["Device Plugin 分配 GPU"]
    ALLOCATE --> SUCCESS["调度成功"]
    
    NORMAL --> SUCCESS
    
    classDef process fill:#e1f5fe
    classDef decision fill:#fff3e0
    classDef success fill:#e8f5e8
    classDef failure fill:#ffebee
    
    class GPU_FILTER,GPU_SCORE,NORMALIZE,SELECT,BIND,ALLOCATE process
    class CHECK_GPU,FILTER_RESULT decision
    class SUCCESS success
    class FAIL failure
```

### 4.4 GPU 资源管理架构

本架构图展示了 Kubernetes 中 GPU 资源的层次化管理体系，从集群级别到容器级别的完整资源治理框架。

**资源层次结构：**

- **集群级别**：整个 Kubernetes 集群的 GPU 资源总览和全局策略
- **命名空间级别**：基于租户的资源隔离和配额管理
- **Pod 级别**：工作负载的资源请求和调度约束
- **容器级别**：具体的 GPU 资源分配和使用

**GPU 资源类型：**

- **GPU 卡数量**：`nvidia.com/gpu` 资源，表示 GPU 设备数量
- **GPU 内存**：`nvidia.com/gpu-memory` 资源，精细化内存管理
- **GPU 类型**：通过 `accelerator` 标签指定 GPU 型号和性能等级
- **GPU 拓扑**：利用拓扑标签优化多卡任务的设备分配

**资源配额管理：**

- **ResourceQuota**：限制命名空间内的 GPU 资源使用总量
- **LimitRange**：设置 Pod 和容器的 GPU 资源使用范围
- **PodDisruptionBudget**：保障关键 GPU 工作负载的可用性

**调度约束机制：**

- **节点亲和性**：将 Pod 调度到特定类型的 GPU 节点
- **Pod 反亲和性**：避免竞争激烈的工作负载调度到同一节点
- **污点与容忍**：专用 GPU 节点的访问控制
- **拓扑分布约束**：实现 GPU 工作负载的均匀分布

**监控指标体系：**

- **GPU 利用率**：实时监控 GPU 计算和内存使用情况
- **GPU 温度**：监控设备温度，预防过热故障
- **功耗监控**：跟踪 GPU 功耗，优化能效管理
- **资源碎片化**：分析资源分配效率，减少浪费

```mermaid
graph TB
    subgraph "资源层次结构"
        CLUSTER["集群级别"]
        NAMESPACE["命名空间级别"]
        POD["Pod 级别"]
        CONTAINER["容器级别"]
    end
    
    subgraph "GPU 资源类型"
        GPU_CARD["GPU 卡数量<br/>nvidia.com/gpu"]
        GPU_MEM["GPU 内存<br/>nvidia.com/gpu-memory"]
        GPU_TYPE["GPU 类型<br/>accelerator label"]
        GPU_TOPO["GPU 拓扑<br/>topology.kubernetes.io/zone"]
    end
    
    subgraph "资源配额管理"
        RQ["ResourceQuota"]
        LR["LimitRange"]
        PDB["PodDisruptionBudget"]
    end
    
    subgraph "调度约束"
        AFFINITY["节点亲和性"]
        ANTI_AFFINITY["Pod 反亲和性"]
        TAINTS["污点与容忍"]
        TOPOLOGY["拓扑分布约束"]
    end
    
    subgraph "监控指标"
        UTIL["GPU 利用率"]
        TEMP["GPU 温度"]
        POWER["功耗监控"]
        FRAG["资源碎片化"]
    end
    
    %% 层次关系
    CLUSTER --> NAMESPACE
    NAMESPACE --> POD
    POD --> CONTAINER
    
    %% 资源关联
    NAMESPACE --> RQ
    NAMESPACE --> LR
    POD --> PDB
    
    %% 调度关联
    POD --> AFFINITY
    POD --> ANTI_AFFINITY
    POD --> TAINTS
    POD --> TOPOLOGY
    
    %% 资源类型关联
    CONTAINER --> GPU_CARD
    CONTAINER --> GPU_MEM
    POD --> GPU_TYPE
    POD --> GPU_TOPO
    
    %% 监控关联
    GPU_CARD --> UTIL
    GPU_CARD --> TEMP
    GPU_CARD --> POWER
    CLUSTER --> FRAG
    
    classDef resource fill:#e3f2fd
    classDef quota fill:#f3e5f5
    classDef constraint fill:#e8f5e8
    classDef monitor fill:#fff3e0
    
    class GPU_CARD,GPU_MEM,GPU_TYPE,GPU_TOPO resource
    class RQ,LR,PDB quota
    class AFFINITY,ANTI_AFFINITY,TAINTS,TOPOLOGY constraint
    class UTIL,TEMP,POWER,FRAG monitor
```

### 4.5 多租户 GPU 资源隔离架构

本架构图展示了企业级 Kubernetes 环境中多租户 GPU 资源的隔离和管理策略，确保不同团队和项目的资源安全隔离。

**多租户隔离层次：**

**1. 命名空间隔离：**

- **租户 A（AI 团队）**：专注于机器学习模型训练，需要高性能 GPU
- **租户 B（研发团队）**：进行算法开发和测试，需要灵活的 GPU 资源
- **租户 C（生产团队）**：运行推理服务，需要稳定的 GPU 资源保障

**2. 资源配额隔离：**

- **ResourceQuota**：为每个租户设置 GPU 资源使用上限
- **LimitRange**：控制单个 Pod 的 GPU 资源请求范围
- **优先级类**：确保关键工作负载的资源优先级

**3. 网络隔离：**

- **NetworkPolicy**：限制租户间的网络通信
- **服务网格**：提供细粒度的流量控制和安全策略
- **DNS 隔离**：防止跨租户的服务发现

**4. 节点池隔离：**

- **专用节点池**：为不同租户分配专用的 GPU 节点
- **污点与容忍**：确保工作负载只能调度到授权的节点
- **节点标签**：基于 GPU 类型和性能进行节点分类

**5. 访问控制：**

- **RBAC**：基于角色的细粒度权限控制
- **ServiceAccount**：为每个租户分配独立的服务账户
- **Pod Security Standards**：强制执行安全策略

**安全隔离特性：**

- **资源边界**：严格的资源配额防止资源争抢
- **网络边界**：网络策略确保数据安全
- **计算边界**：专用节点池提供物理隔离
- **权限边界**：RBAC 确保操作权限隔离

```mermaid
graph TB
    subgraph "租户 A - 机器学习团队"
        NSA["Namespace: ml-team"]
        RQA["ResourceQuota<br/>GPU: 32, Memory: 1TB"]
        PODA1["Training Pod<br/>V100 x8"]
        PODA2["Inference Pod<br/>T4 x2"]
    end
    
    subgraph "租户 B - 研发团队"
        NSB["Namespace: dev-team"]
        RQB["ResourceQuota<br/>GPU: 8, Memory: 256GB"]
        PODB1["Dev Pod<br/>RTX3080 x1"]
        PODB2["Test Pod<br/>RTX3080 x1"]
    end
    
    subgraph "租户 C - 生产服务"
        NSC["Namespace: prod-service"]
        RQC["ResourceQuota<br/>GPU: 16, Memory: 512GB"]
        PODC1["Inference Service<br/>A100 x4"]
        PODC2["Batch Processing<br/>V100 x4"]
    end
    
    subgraph "GPU 节点池"
        subgraph "高性能池"
            NODE1["Node 1<br/>A100 x8"]
            NODE2["Node 2<br/>V100 x8"]
        end
        
        subgraph "中等性能池"
            NODE3["Node 3<br/>T4 x4"]
            NODE4["Node 4<br/>RTX3080 x4"]
        end
        
        subgraph "开发测试池"
            NODE5["Node 5<br/>GTX1080 x2"]
            NODE6["Node 6<br/>RTX2080 x2"]
        end
    end
    
    subgraph "访问控制"
        RBAC["RBAC 权限控制"]
        SA_A["ServiceAccount: ml-team"]
        SA_B["ServiceAccount: dev-team"]
        SA_C["ServiceAccount: prod-service"]
    end
    
    subgraph "网络隔离"
        NP_A["NetworkPolicy: ml-team"]
        NP_B["NetworkPolicy: dev-team"]
        NP_C["NetworkPolicy: prod-service"]
    end
    
    %% 租户资源关联
    NSA --> RQA
    NSA --> PODA1
    NSA --> PODA2
    
    NSB --> RQB
    NSB --> PODB1
    NSB --> PODB2
    
    NSC --> RQC
    NSC --> PODC1
    NSC --> PODC2
    
    %% Pod 调度到节点
    PODA1 --> NODE1
    PODA1 --> NODE2
    PODA2 --> NODE3
    
    PODB1 --> NODE4
    PODB2 --> NODE5
    
    PODC1 --> NODE1
    PODC2 --> NODE2
    
    %% 权限控制
    NSA --> SA_A
    NSB --> SA_B
    NSC --> SA_C
    
    SA_A --> RBAC
    SA_B --> RBAC
    SA_C --> RBAC
    
    %% 网络隔离
    NSA --> NP_A
    NSB --> NP_B
    NSC --> NP_C
    
    classDef tenant_a fill:#e3f2fd
    classDef tenant_b fill:#e8f5e8
    classDef tenant_c fill:#fff3e0
    classDef security fill:#f3e5f5
    
    class NSA,RQA,PODA1,PODA2,SA_A,NP_A tenant_a
    class NSB,RQB,PODB1,PODB2,SA_B,NP_B tenant_b
    class NSC,RQC,PODC1,PODC2,SA_C,NP_C tenant_c
    class RBAC security
```

### 4.6 GPU 监控与运维架构

本架构图展示了完整的 GPU 监控与运维体系，从数据采集到告警通知的全链路可观测性解决方案。

**监控架构层次：**

**1. 数据采集层：**

- **NVIDIA DCGM**：GPU 硬件级监控，提供温度、功耗、利用率等核心指标
- **Node Exporter**：节点级系统监控，包括 CPU、内存、磁盘、网络等
- **cAdvisor**：容器级资源监控，跟踪容器的 GPU 使用情况
- **Kube State Metrics**：Kubernetes 对象状态监控，如 Pod、Node、Deployment 等

**2. 指标存储层：**

- **Prometheus**：时序数据库，存储和查询监控指标
- **Thanos**：长期存储和高可用性扩展，支持多集群监控

**3. 可视化层：**

- **Grafana**：监控仪表板和数据可视化
  - **GPU 集群概览**：整体 GPU 资源使用情况
  - **节点详情**：单节点 GPU 设备状态和性能
  - **工作负载监控**：Pod 和容器的 GPU 使用分析
  - **告警面板**：实时告警状态和历史趋势

**4. 告警层：**

- **AlertManager**：告警规则管理和通知路由
  - **GPU 利用率告警**：监控 GPU 使用率异常
  - **温度告警**：防止 GPU 过热损坏
  - **调度延迟告警**：监控调度性能问题
  - **资源配额告警**：预警资源配额即将耗尽

**5. 通知渠道：**

- **邮件通知**：发送详细的告警信息
- **Slack 集成**：实时团队协作通知
- **Webhook**：集成第三方系统
- **值班系统**：紧急故障升级处理

**6. 运维工具：**

- **kubectl**：Kubernetes 命令行管理工具
- **k9s**：交互式 Kubernetes 管理界面
- **Lens**：Kubernetes 图形化管理工具
- **nvidia-smi**：GPU 设备状态查询工具

**7. 日志系统：**

- **Fluentd**：日志收集和转发
- **Elasticsearch**：日志存储和搜索
- **Kibana**：日志分析和可视化

**运维特性：**

- **全栈监控**：从硬件到应用的完整监控覆盖
- **智能告警**：基于机器学习的异常检测
- **自动化运维**：故障自愈和资源自动调整
- **多维分析**：性能、成本、效率的综合分析

```mermaid
graph TB
    subgraph "数据采集层"
        DCGM["NVIDIA DCGM<br/>GPU 指标采集"]
        NODE_EXP["Node Exporter<br/>节点指标"]
        CADVISOR["cAdvisor<br/>容器指标"]
        KUBE_STATE["Kube State Metrics<br/>K8s 对象指标"]
    end
    
    subgraph "指标存储层"
        PROM["Prometheus<br/>时序数据库"]
        THANOS["Thanos<br/>长期存储"]
    end
    
    subgraph "可视化层"
        GRAFANA["Grafana<br/>监控仪表板"]
        subgraph "仪表板"
            DASH1["GPU 集群概览"]
            DASH2["GPU 利用率分析"]
            DASH3["调度性能监控"]
            DASH4["资源配额监控"]
        end
    end
    
    subgraph "告警层"
        ALERT_MGR["AlertManager<br/>告警管理"]
        subgraph "告警规则"
            RULE1["GPU 利用率过低"]
            RULE2["GPU 温度过高"]
            RULE3["调度失败率过高"]
            RULE4["资源配额超限"]
        end
    end
    
    subgraph "通知渠道"
        EMAIL["邮件通知"]
        SLACK["Slack 通知"]
        WEBHOOK["Webhook 通知"]
        ONCALL["值班系统"]
    end
    
    subgraph "运维工具"
        KUBECTL["kubectl<br/>命令行工具"]
        K9S["k9s<br/>集群管理"]
        LENS["Lens<br/>可视化管理"]
        NVIDIA_SMI["nvidia-smi<br/>GPU 诊断"]
    end
    
    subgraph "日志系统"
        FLUENTD["Fluentd<br/>日志采集"]
        ELASTIC["Elasticsearch<br/>日志存储"]
        KIBANA["Kibana<br/>日志分析"]
    end
    
    %% 数据流
    DCGM --> PROM
    NODE_EXP --> PROM
    CADVISOR --> PROM
    KUBE_STATE --> PROM
    
    PROM --> THANOS
    PROM --> GRAFANA
    PROM --> ALERT_MGR
    
    %% 可视化
    GRAFANA --> DASH1
    GRAFANA --> DASH2
    GRAFANA --> DASH3
    GRAFANA --> DASH4
    
    %% 告警
    ALERT_MGR --> RULE1
    ALERT_MGR --> RULE2
    ALERT_MGR --> RULE3
    ALERT_MGR --> RULE4
    
    ALERT_MGR --> EMAIL
    ALERT_MGR --> SLACK
    ALERT_MGR --> WEBHOOK
    ALERT_MGR --> ONCALL
    
    %% 日志
    FLUENTD --> ELASTIC
    ELASTIC --> KIBANA
    
    classDef collect fill:#e3f2fd
    classDef store fill:#e8f5e8
    classDef visual fill:#fff3e0
    classDef alert fill:#ffebee
    classDef tools fill:#f3e5f5
    
    class DCGM,NODE_EXP,CADVISOR,KUBE_STATE,FLUENTD collect
    class PROM,THANOS,ELASTIC store
    class GRAFANA,DASH1,DASH2,DASH3,DASH4,KIBANA visual
    class ALERT_MGR,RULE1,RULE2,RULE3,RULE4 alert
    class KUBECTL,K9S,LENS,NVIDIA_SMI tools
```

### 4.7 GPU 调度器插件架构

本架构图展示了 GPU 调度器插件如何与 Kubernetes Scheduler Framework 集成，实现智能化的 GPU 资源调度。

**Scheduler Framework 核心组件：**

**1. 调度队列（Scheduling Queue）：**

- **ActiveQ**：等待调度的 Pod 队列，按优先级排序
- **BackoffQ**：调度失败需要重试的 Pod 队列
- **UnschedulableQ**：暂时无法调度的 Pod 队列
- **队列管理**：支持优先级调度和公平调度

**2. 调度缓存（Scheduler Cache）：**

- **节点信息缓存**：实时维护集群节点的资源状态
- **Pod 信息缓存**：跟踪已调度和待调度的 Pod
- **资源快照**：提供一致性的集群资源视图
- **增量更新**：高效的缓存同步机制

**3. 调度框架（Scheduling Framework）：**

- **插件管理器**：动态加载和管理调度插件
- **扩展点执行器**：在特定阶段调用相应插件
- **上下文管理**：维护调度过程中的状态和数据
- **并发控制**：支持多 Pod 并发调度

**调度扩展点详解：**

**过滤阶段：**

- **PreFilter**：预过滤，快速排除明显不符合的节点
- **Filter**：详细过滤，检查节点是否满足 Pod 的所有要求
- **PostFilter**：后过滤，处理所有节点都被过滤的情况

**评分阶段：**

- **PreScore**：预评分，准备评分所需的数据和上下文
- **Score**：评分，为每个通过过滤的节点计算适合度分数
- **NormalizeScore**：分数标准化，确保分数在 0-100 范围内

**绑定阶段：**

- **Reserve**：资源预留，为选中节点预留资源
- **Permit**：许可检查，决定是否允许 Pod 绑定
- **PreBind**：预绑定，执行绑定前的准备工作
- **Bind**：绑定执行，将 Pod 绑定到选中节点
- **PostBind**：后绑定，执行绑定后的清理和通知

**GPU 调度插件实现：**

**1. GPUFilter Plugin：**

- **GPU 容量检查**：验证节点是否有足够的 GPU 资源
- **GPU 类型检查**：确认节点 GPU 类型符合 Pod 要求
- **拓扑约束检查**：验证 GPU 拓扑和亲和性要求
- **资源配额检查**：确保不超过命名空间资源限制

**2. GPUScore Plugin：**

- **利用率评分**：基于 GPU 利用率计算节点适合度
- **类型匹配评分**：优先选择 GPU 类型完全匹配的节点
- **拓扑评分**：优化多卡任务的 GPU 间通信
- **碎片化评分**：减少资源碎片化，提高利用率

**3. GPUBind Plugin：**

- **设备分配**：为 Pod 分配具体的 GPU 设备
- **环境变量设置**：配置 GPU 相关的环境变量
- **设备路径映射**：设置容器内的 GPU 设备路径
- **资源更新**：更新节点的 GPU 资源状态

**插件协作机制：**

- **状态共享**：插件间通过调度上下文共享状态
- **配置管理**：统一的插件配置和参数管理
- **错误处理**：完善的错误处理和回滚机制
- **性能优化**：缓存和批处理优化调度性能

```mermaid
graph TB
    subgraph "Kubernetes Scheduler Framework"
        QUEUE["Scheduling Queue"]
        CACHE["Scheduler Cache"]
        FRAMEWORK["Scheduling Framework"]
    end
    
    subgraph "调度扩展点 (Extension Points)"
        PREFILTER["PreFilter"]
        FILTER["Filter"]
        POSTFILTER["PostFilter"]
        PRESCORE["PreScore"]
        SCORE["Score"]
        RESERVE["Reserve"]
        PERMIT["Permit"]
        PREBIND["PreBind"]
        BIND["Bind"]
        POSTBIND["PostBind"]
    end
    
    subgraph "GPU 调度插件"
        GPU_FILTER["GPUFilter Plugin"]
        GPU_SCORE["GPUScore Plugin"]
        GPU_BIND["GPUBind Plugin"]
    end
    
    subgraph "GPU Filter 逻辑"
        F_CHECK1["检查 GPU 容量"]
        F_CHECK2["检查 GPU 类型"]
        F_CHECK3["检查拓扑约束"]
        F_CHECK4["检查资源配额"]
    end
    
    subgraph "GPU Score 逻辑"
        S_UTIL["利用率评分"]
        S_TYPE["类型匹配评分"]
        S_TOPO["拓扑评分"]
        S_FRAG["碎片化评分"]
    end
    
    subgraph "配置管理"
        CONFIG["KubeSchedulerConfiguration"]
        PLUGIN_CONFIG["Plugin Config"]
        WEIGHTS["Score Weights"]
    end
    
    subgraph "状态管理"
        CYCLE_STATE["CycleState"]
        NODE_INFO["NodeInfo"]
        POD_INFO["PodInfo"]
    end
    
    %% 调度流程
    QUEUE --> FRAMEWORK
    CACHE --> FRAMEWORK
    
    FRAMEWORK --> PREFILTER
    PREFILTER --> FILTER
    FILTER --> POSTFILTER
    POSTFILTER --> PRESCORE
    PRESCORE --> SCORE
    SCORE --> RESERVE
    RESERVE --> PERMIT
    PERMIT --> PREBIND
    PREBIND --> BIND
    BIND --> POSTBIND
    
    %% 插件注册
    GPU_FILTER --> FILTER
    GPU_SCORE --> SCORE
    GPU_BIND --> BIND
    
    %% Filter 插件逻辑
    GPU_FILTER --> F_CHECK1
    GPU_FILTER --> F_CHECK2
    GPU_FILTER --> F_CHECK3
    GPU_FILTER --> F_CHECK4
    
    %% Score 插件逻辑
    GPU_SCORE --> S_UTIL
    GPU_SCORE --> S_TYPE
    GPU_SCORE --> S_TOPO
    GPU_SCORE --> S_FRAG
    
    %% 配置
    CONFIG --> PLUGIN_CONFIG
    PLUGIN_CONFIG --> GPU_FILTER
    PLUGIN_CONFIG --> GPU_SCORE
    WEIGHTS --> GPU_SCORE
    
    %% 状态
    CYCLE_STATE --> GPU_FILTER
    CYCLE_STATE --> GPU_SCORE
    NODE_INFO --> GPU_FILTER
    NODE_INFO --> GPU_SCORE
    POD_INFO --> GPU_FILTER
    POD_INFO --> GPU_SCORE
    
    classDef framework fill:#e3f2fd
    classDef extension fill:#e8f5e8
    classDef plugin fill:#fff3e0
    classDef logic fill:#f3e5f5
    classDef config fill:#fce4ec
    
    class QUEUE,CACHE,FRAMEWORK framework
    class PREFILTER,FILTER,POSTFILTER,PRESCORE,SCORE,RESERVE,PERMIT,PREBIND,BIND,POSTBIND extension
    class GPU_FILTER,GPU_SCORE,GPU_BIND plugin
    class F_CHECK1,F_CHECK2,F_CHECK3,F_CHECK4,S_UTIL,S_TYPE,S_TOPO,S_FRAG logic
    class CONFIG,PLUGIN_CONFIG,WEIGHTS,CYCLE_STATE,NODE_INFO,POD_INFO config
```

### 4.8 架构图说明

#### 4.8.1 整体架构特点

- **分层设计**：控制平面、工作节点、监控运维分层清晰
- **插件化**：GPU 调度器采用插件化架构，易于扩展
- **多租户**：支持基于命名空间的资源隔离和配额管理
- **可观测性**：完整的监控、告警和日志体系

#### 4.8.2 关键组件说明

- **Device Plugin**：负责 GPU 资源发现和分配
- **GPU Filter Plugin**：过滤不满足 GPU 要求的节点
- **GPU Score Plugin**：为候选节点进行 GPU 相关评分
- **Resource Quota**：控制命名空间级别的 GPU 资源使用

#### 4.8.3 工作流程要点

1. Device Plugin 发现并注册 GPU 资源
2. GPU 调度器根据 Pod 需求进行智能调度
3. 监控系统实时跟踪 GPU 使用情况
4. 告警系统及时发现和处理异常

#### 4.8.4 扩展性考虑

- 支持多种 GPU 类型和厂商
- 可配置的调度策略和权重
- 插件化架构支持自定义扩展
- 多集群联邦调度能力

这些架构图为 GPU 调度器的设计、实现和运维提供了全面的技术参考。

### 4.9 实现路径

基于上述架构设计，我们将按照以下路径实现 GPU 调度器：

**第一阶段：核心调度插件开发：**

- 实现 GPUFilter 插件：负责节点过滤和 GPU 资源检查
- 实现 GPUScore 插件：负责节点评分和最优选择
- 集成 Kubernetes Scheduler Framework

**第二阶段：监控和可观测性：**

- 集成 Prometheus 指标收集
- 实现调度性能监控
- 配置告警规则和仪表板

**第三阶段：生产级部署：**

- 配置高可用部署
- 实现 RBAC 安全控制
- 添加故障恢复机制

接下来的章节将详细展示每个组件的具体实现。

---

## 5 自定义调度器扩展

### 5.1 GPU Filter 插件实现

```go
// GPU Filter 插件实现
// 该插件负责在调度过程中过滤不满足 GPU 要求的节点
// 实现了 Kubernetes Scheduler Framework 的 Filter 接口
package main

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    
    v1 "k8s.io/api/core/v1"                    // Kubernetes 核心 API 对象
    "k8s.io/apimachinery/pkg/runtime"          // Kubernetes 运行时对象
    "k8s.io/kubernetes/pkg/scheduler/framework" // 调度器框架接口
)

// GPUFilter 插件名称常量
const GPUFilterName = "GPUFilter"

// GPUFilter 插件结构体
// 实现了 framework.FilterPlugin 接口
type GPUFilter struct {
    handle framework.Handle  // 调度器框架句柄，提供访问集群状态的能力
}

// Name 返回插件名称
// 实现 framework.Plugin 接口的必需方法
func (gf *GPUFilter) Name() string {
    return GPUFilterName
}

// Filter 实现核心过滤逻辑
// 检查节点是否满足 Pod 的 GPU 资源需求
// 返回 framework.Status 表示过滤结果
func (gf *GPUFilter) Filter(
    ctx context.Context,              // 上下文，用于取消和超时控制
    state *framework.CycleState,      // 调度周期状态，存储调度过程中的临时数据
    pod *v1.Pod,                      // 待调度的 Pod 对象
    nodeInfo *framework.NodeInfo,     // 候选节点信息
) *framework.Status {
    node := nodeInfo.Node()
    if node == nil {
        return framework.NewStatus(framework.Error, "node not found")
    }
    
    // 1. 检查 Pod 是否需要 GPU 资源
    gpuRequest := getGPURequest(pod)
    if gpuRequest == 0 {
        // Pod 不需要 GPU，直接通过过滤
        return framework.NewStatus(framework.Success, "")
    }
    
    // 2. 检查节点是否具备 GPU 资源
    nodeGPUCapacity := getNodeGPUCapacity(node)
    if nodeGPUCapacity == 0 {
        // 节点没有 GPU，且无法通过其他方式获得，标记为不可调度且不可解决
        return framework.NewStatus(
            framework.UnschedulableAndUnresolvable,
            "node has no GPU resources",
        )
    }
    
    // 3. 检查 GPU 类型是否匹配
    if !isGPUTypeMatched(pod, node) {
        // GPU 类型不匹配，标记为不可调度且不可解决
        return framework.NewStatus(
            framework.UnschedulableAndUnresolvable,
            "GPU type not matched",
        )
    }
    
    // 4. 检查可用 GPU 数量是否足够
    availableGPU := getAvailableGPU(nodeInfo)
    if availableGPU < gpuRequest {
        // GPU 数量不足，标记为不可调度（可能在其他 Pod 释放资源后变为可调度）
        return framework.NewStatus(
            framework.Unschedulable,
            fmt.Sprintf("insufficient GPU: need %d, available %d", 
                gpuRequest, availableGPU),
        )
    }
    
    // 所有检查通过，节点满足 GPU 要求
    return framework.NewStatus(framework.Success, "")
}

// getGPURequest 计算 Pod 中所有容器的 GPU 请求总数
// 遍历 Pod 的所有容器，累加 nvidia.com/gpu 资源请求
func getGPURequest(pod *v1.Pod) int64 {
    var totalGPU int64
    for _, container := range pod.Spec.Containers {
        if gpuQuantity, ok := container.Resources.Requests["nvidia.com/gpu"]; ok {
            totalGPU += gpuQuantity.Value()
        }
    }
    return totalGPU
}

// getNodeGPUCapacity 获取节点的 GPU 总容量
// 从节点的 Status.Capacity 中读取 nvidia.com/gpu 资源
func getNodeGPUCapacity(node *v1.Node) int64 {
    if gpuQuantity, ok := node.Status.Capacity["nvidia.com/gpu"]; ok {
        return gpuQuantity.Value()
    }
    return 0
}

// isGPUTypeMatched 检查 Pod 要求的 GPU 类型与节点 GPU 类型是否匹配
// 通过比较 Pod 的 nodeSelector 和节点的 labels 中的 accelerator 字段
func isGPUTypeMatched(pod *v1.Pod, node *v1.Node) bool {
    // 检查 Pod 是否通过 nodeSelector 指定了 GPU 类型
    if podGPUType, ok := pod.Spec.NodeSelector["accelerator"]; ok {
        // Pod 指定了 GPU 类型，检查节点是否有对应的标签
        if nodeGPUType, ok := node.Labels["accelerator"]; ok {
            return podGPUType == nodeGPUType
        }
        return false // 节点没有 GPU 类型标签
    }
    return true // Pod 没有指定 GPU 类型要求，认为匹配任何类型
}

// getAvailableGPU 计算节点当前可用的 GPU 数量
// 总容量减去已分配给运行中 Pod 的 GPU 数量
func getAvailableGPU(nodeInfo *framework.NodeInfo) int64 {
    totalGPU := getNodeGPUCapacity(nodeInfo.Node())
    usedGPU := int64(0)
    
    // 遍历节点上所有运行的 Pod，计算已使用的 GPU 总数
    for _, podInfo := range nodeInfo.Pods {
        usedGPU += getGPURequest(podInfo.Pod)
    }
    
    return totalGPU - usedGPU
}

// NewGPUFilter 插件工厂函数
// 用于创建 GPUFilter 插件实例，符合 Kubernetes 调度器插件注册规范
func NewGPUFilter(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
    return &GPUFilter{handle: h}, nil
}
```

### 5.2 GPU Score 插件实现

```go
// GPU Score 插件实现
// 该插件负责为通过过滤的节点计算 GPU 相关的评分
// 实现了 Kubernetes Scheduler Framework 的 Score 接口
package main

import (
    "context"
    "fmt"
    "math"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    v1 "k8s.io/api/core/v1"                    // Kubernetes 核心 API 对象
    "k8s.io/apimachinery/pkg/runtime"          // Kubernetes 运行时对象
    "k8s.io/kubernetes/pkg/scheduler/framework" // 调度器框架接口
)

// GPUScore 插件名称常量
const GPUScoreName = "GPUScore"

// GPUScore 插件结构体
// 实现了 framework.ScorePlugin 接口
type GPUScore struct {
    handle framework.Handle  // 调度器框架句柄，提供访问集群状态的能力
}

// Name 返回插件名称
// 实现 framework.Plugin 接口的必需方法
func (gs *GPUScore) Name() string {
    return GPUScoreName
}

// Score 实现核心评分逻辑
// 为每个候选节点计算 GPU 相关的评分，分数越高表示节点越适合调度
// 返回 0-100 的评分和状态
func (gs *GPUScore) Score(
    ctx context.Context,              // 上下文，用于取消和超时控制
    state *framework.CycleState,      // 调度周期状态
    pod *v1.Pod,                      // 待调度的 Pod 对象
    nodeName string,                  // 候选节点名称
) (int64, *framework.Status) {
    // 获取节点信息快照
    nodeInfo, err := gs.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
    if err != nil {
        return 0, framework.NewStatus(
            framework.Error,
            fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err),
        )
    }
    
    // 计算节点的 GPU 评分
    score := gs.calculateGPUScore(pod, nodeInfo)
    return score, framework.NewStatus(framework.Success, "")
}

// ScoreExtensions 返回评分扩展接口
// 允许插件实现评分标准化等高级功能
func (gs *GPUScore) ScoreExtensions() framework.ScoreExtensions {
    return gs
}

// NormalizeScore 标准化所有节点的评分
// 将原始评分标准化到 0-100 的范围，确保评分的公平性和可比性
func (gs *GPUScore) NormalizeScore(
    ctx context.Context,              // 上下文
    state *framework.CycleState,      // 调度周期状态
    pod *v1.Pod,                      // 待调度的 Pod
    scores framework.NodeScoreList,   // 所有节点的原始评分列表
) *framework.Status {
    // 1. 找到所有节点评分的最大值和最小值
    var maxScore, minScore int64 = 0, math.MaxInt64
    for _, nodeScore := range scores {
        if nodeScore.Score > maxScore {
            maxScore = nodeScore.Score
        }
        if nodeScore.Score < minScore {
            minScore = nodeScore.Score
        }
    }
    
    // 2. 执行线性标准化，将评分映射到 0-100 范围
    if maxScore == minScore {
        // 所有节点评分相同，给予相同的最高分
        for i := range scores {
            scores[i].Score = framework.MaxNodeScore
        }
    } else {
        // 线性标准化公式：(score - min) * 100 / (max - min)
        for i := range scores {
            scores[i].Score = (scores[i].Score - minScore) * 
                framework.MaxNodeScore / (maxScore - minScore)
        }
    }
    
    return framework.NewStatus(framework.Success, "")
}

// calculateAdvancedGPUScore 高级 GPU 评分算法
// 综合考虑多个维度对节点进行评分，包括资源利用率、类型匹配、拓扑结构等
// 返回 0-100 的综合评分
func (gs *GPUScore) calculateAdvancedGPUScore(pod *v1.Pod, nodeInfo *framework.NodeInfo) int64 {
    node := nodeInfo.Node()
    
    // 基础分数设为 50，作为评分起点
    var score int64 = 50
    
    // 1. GPU 利用率评分（权重：25%）
    // 评估调度后的资源利用率，追求均衡分布
    utilizationScore := gs.calculateUtilizationScore(nodeInfo)
    score += int64(float64(utilizationScore) * 0.25)
    
    // 2. GPU 类型匹配评分（权重：20%）
    // 评估 GPU 硬件类型与 Pod 需求的匹配度
    typeScore := gs.calculateTypeScore(pod, node)
    score += int64(float64(typeScore) * 0.20)
    
    // 3. GPU 拓扑评分（权重：20%）
    // 评估 GPU 间的互联性能，对多卡任务尤为重要
    topologyScore := gs.calculateTopologyScore(pod, nodeInfo)
    score += int64(float64(topologyScore) * 0.20)
    
    // 4. GPU 内存碎片化评分（权重：15%）
    // 评估内存分配的碎片化程度，影响资源利用效率
    fragmentationScore := gs.calculateFragmentationScore(nodeInfo)
    score += int64(float64(fragmentationScore) * 0.15)
    
    // 5. 租户优先级评分（权重：10%）
    // 根据 Pod 的优先级和命名空间进行差异化调度
    priorityScore := gs.calculatePriorityScore(pod)
    score += int64(float64(priorityScore) * 0.10)
    
    // 6. 工作负载特征评分（权重：10%）
    // 根据工作负载类型（训练/推理/开发）优化调度策略
    workloadScore := gs.calculateWorkloadScore(pod)
    score += int64(float64(workloadScore) * 0.10)
    
    // 确保分数在有效范围内（0-100），防止数值溢出
    if score < 0 {
        score = 0
    } else if score > 100 {
        score = 100
    }
    
    return score
}

// calculateFragmentationScore 计算 GPU 内存碎片化评分
// 评估节点上 GPU 内存的碎片化程度，碎片化越少评分越高
// 有助于提高整体资源利用效率和减少资源浪费
func (gs *GPUScore) calculateFragmentationScore(nodeInfo *framework.NodeInfo) int64 {
    node := nodeInfo.Node()
    totalGPUMemory := getNodeGPUMemory(node)
    if totalGPUMemory == 0 {
        return 0  // 节点没有 GPU 内存
    }
    
    // 计算已使用的 GPU 内存总量
    usedGPUMemory := int64(0)
    for _, podInfo := range nodeInfo.Pods {
        usedGPUMemory += getGPUMemoryRequest(podInfo.Pod)
    }
    
    // 计算内存碎片率（可用内存占总内存的比例）
    fragmentationRate := float64(totalGPUMemory-usedGPUMemory) / float64(totalGPUMemory)
    
    // 根据碎片率分级评分，碎片率越低（内存利用率越高）评分越高
    if fragmentationRate < 0.1 {
        return 100  // 内存利用率 > 90%，优秀
    } else if fragmentationRate < 0.2 {
        return 80   // 内存利用率 80-90%，良好
    } else {
        // 内存利用率 < 80%，根据实际利用率线性评分
        return int64((1.0 - fragmentationRate) * 60)
    }
}

// calculatePriorityScore 计算租户优先级评分
// 根据 Pod 的优先级类和命名空间实现差异化调度策略
// 确保高优先级工作负载能够获得更好的资源分配
func (gs *GPUScore) calculatePriorityScore(pod *v1.Pod) int64 {
    // 1. 首先检查 Pod 的优先级类（PriorityClass）
    // 优先级类是 Kubernetes 中定义 Pod 相对重要性的机制
    if pod.Spec.PriorityClassName != "" {
        switch pod.Spec.PriorityClassName {
        case "high-priority":
            return 100  // 高优先级：关键业务，最高评分
        case "medium-priority":
            return 70   // 中优先级：重要业务，较高评分
        case "low-priority":
            return 30   // 低优先级：非关键业务，较低评分
        }
    }
    
    // 2. 如果没有设置优先级类，则根据命名空间进行评分
    // 不同环境的命名空间具有不同的业务重要性
    namespace := pod.Namespace
    switch namespace {
    case "production", "critical":
        return 90   // 生产环境：业务关键，高评分
    case "staging", "testing":
        return 60   // 测试环境：验证阶段，中等评分
    case "development":
        return 40   // 开发环境：开发调试，较低评分
    default:
        return 50   // 默认命名空间：标准评分
    }
}

// calculateWorkloadScore 计算工作负载特征评分
// 根据不同类型的 GPU 工作负载特点进行针对性的调度优化
// 不同工作负载对 GPU 资源的需求模式和性能要求存在显著差异
func (gs *GPUScore) calculateWorkloadScore(pod *v1.Pod) int64 {
    // 检查 Pod 标签中的工作负载类型标识
    // 应用需要通过标签声明自己的工作负载类型以获得最优调度
    if workloadType, ok := pod.Labels["workload-type"]; ok {
        switch workloadType {
        case "training":
            // 训练任务特点：
            // - 需要高性能 GPU 和大显存
            // - 偏好多卡节点以支持分布式训练
            // - 对网络带宽要求高（多卡通信）
            // - 运行时间长，资源独占性强
            return 90
        case "inference":
            // 推理任务特点：
            // - 偏好低延迟和高吞吐量
            // - 通常单卡即可，但需要快速响应
            // - 对资源利用率敏感
            // - 可能需要支持动态扩缩容
            return 80
        case "development":
            // 开发任务特点：
            // - 资源需求相对较低
            // - 主要用于代码调试和小规模测试
            // - 对性能要求不高，但需要快速启动
            // - 适合资源共享和抢占式调度
            return 60
        }
    }
    
    // 未指定工作负载类型，给予标准评分
    return 50
}

// getNodeGPUMemory 获取节点的 GPU 内存总容量
// 从节点的 Status.Capacity 中读取 nvidia.com/gpu-memory 资源
// 用于计算内存碎片化和资源利用率
func getNodeGPUMemory(node *v1.Node) int64 {
    if gpuMemory, ok := node.Status.Capacity["nvidia.com/gpu-memory"]; ok {
        return gpuMemory.Value()
    }
    return 0  // 节点没有 GPU 内存资源
}

// getGPUMemoryRequest 计算 Pod 中所有容器的 GPU 内存请求总量
// 遍历 Pod 的所有容器，累加 nvidia.com/gpu-memory 资源请求
// 用于评估内存使用情况和碎片化程度
func getGPUMemoryRequest(pod *v1.Pod) int64 {
    var totalMemory int64
    for _, container := range pod.Spec.Containers {
        if memQuantity, ok := container.Resources.Requests["nvidia.com/gpu-memory"]; ok {
            totalMemory += memQuantity.Value()
        }
    }
    return totalMemory
}

// calculateGPUScore 计算节点的 GPU 综合评分
// 这是一个简化版本的评分算法，主要考虑三个核心维度
// 与 calculateAdvancedGPUScore 相比，权重分配更加集中
func (gs *GPUScore) calculateGPUScore(pod *v1.Pod, nodeInfo *framework.NodeInfo) int64 {
    node := nodeInfo.Node()
    
    // 基础分数设为 50，作为评分起点
    var score int64 = 50
    
    // 1. GPU 利用率评分（权重：40%）
    // 这是最重要的评分维度，确保资源的均衡分布
    utilizationScore := gs.calculateUtilizationScore(nodeInfo)
    score += int64(float64(utilizationScore) * 0.4)
    
    // 2. GPU 类型匹配评分（权重：30%）
    // 确保 Pod 能够获得合适的 GPU 硬件类型
    typeScore := gs.calculateTypeScore(pod, node)
    score += int64(float64(typeScore) * 0.3)
    
    // 3. GPU 拓扑评分（权重：30%）
    // 对于多卡任务，优化 GPU 间的通信性能
    topologyScore := gs.calculateTopologyScore(pod, nodeInfo)
    score += int64(float64(topologyScore) * 0.3)
    
    return score
}

// calculateUtilizationScore 计算基于 GPU 利用率的评分
// 目标是实现集群 GPU 资源的均衡分布，避免资源碎片化
// 采用目标利用率区间策略，在理想利用率范围内给予最高评分
func (gs *GPUScore) calculateUtilizationScore(nodeInfo *framework.NodeInfo) int64 {
    totalGPU := getNodeGPUCapacity(nodeInfo.Node())
    if totalGPU == 0 {
        return 0  // 节点没有 GPU 资源
    }
    
    // 计算节点当前已使用的 GPU 数量
    usedGPU := int64(0)
    for _, podInfo := range nodeInfo.Pods {
        usedGPU += getGPURequest(podInfo.Pod)
    }
    
    // 计算当前 GPU 利用率
    utilization := float64(usedGPU) / float64(totalGPU)
    
    // 目标利用率区间：60-80%
    // 这个区间既保证了较高的资源利用率，又留有一定的缓冲空间
    if utilization >= 0.6 && utilization <= 0.8 {
        return 100  // 理想利用率，给予最高评分
    } else if utilization < 0.6 {
        // 利用率偏低，线性递增评分（鼓励使用利用率低的节点）
        return int64(utilization * 100 / 0.6)
    } else {
        // 利用率偏高，线性递减评分（避免过度集中）
        return int64((1.0 - utilization) * 100 / 0.2)
    }
}

// calculateTypeScore 计算基于 GPU 类型匹配的评分
// 优先选择与 Pod 要求完全匹配的 GPU 类型，同时支持兼容性匹配
// 确保应用能够在合适的硬件上运行并获得最佳性能
func (gs *GPUScore) calculateTypeScore(pod *v1.Pod, node *v1.Node) int64 {
    // 检查 Pod 是否通过 nodeSelector 指定了 GPU 类型要求
    if podGPUType, ok := pod.Spec.NodeSelector["accelerator"]; ok {
        // Pod 指定了 GPU 类型，检查节点是否有对应的标签
        if nodeGPUType, ok := node.Labels["accelerator"]; ok {
            if podGPUType == nodeGPUType {
                return 100 // 完全匹配，给予最高评分
            }
            
            // 检查 GPU 类型兼容性（例如：V100 和 A100 都是高性能 GPU）
            // 兼容的 GPU 类型可以满足基本需求，但性能可能有差异
            if gs.isCompatibleGPUType(podGPUType, nodeGPUType) {
                return 70  // 兼容匹配，给予较高评分
            }
        }
        return 0 // GPU 类型不匹配或节点没有 GPU 类型标签
    }
    return 50 // Pod 没有指定 GPU 类型要求，给予中等评分
}

// calculateTopologyScore 计算基于 GPU 拓扑的评分
// 对于多 GPU 请求，优先选择具有良好 GPU 互联的节点
// GPU 间的高速互联对分布式训练等多卡任务的性能至关重要
func (gs *GPUScore) calculateTopologyScore(pod *v1.Pod, nodeInfo *framework.NodeInfo) int64 {
    gpuRequest := getGPURequest(pod)
    if gpuRequest <= 1 {
        return 100 // 单卡任务不需要考虑 GPU 间拓扑，给予最高评分
    }
    
    // 多卡任务需要评估 GPU 间的互联性能
    // 检查节点是否支持 NVLink 或其他高速互联技术
    node := nodeInfo.Node()
    if nvlinkSupported, ok := node.Labels["gpu-interconnect"]; ok && nvlinkSupported == "nvlink" {
        return 100  // 支持 NVLink 高速互联，给予最高评分
    }
    
    // 可以扩展支持其他互联技术的评分逻辑
    // 例如：InfiniBand、高速以太网等
    
    return 50 // 没有特殊互联技术或未知拓扑，给予默认评分
}

// isCompatibleGPUType 检查 GPU 类型兼容性
// 定义不同 GPU 类型之间的兼容性关系，支持跨型号调度
// 兼容性主要基于计算能力、架构相似性和应用场景匹配度
func (gs *GPUScore) isCompatibleGPUType(requested, available string) bool {
    // GPU 类型兼容性矩阵
    // 键为请求的 GPU 类型，值为兼容的 GPU 类型列表
    compatibility := map[string][]string{
        "nvidia-tesla-v100": {"nvidia-tesla-a100"},     // V100 与 A100 架构相近，性能相当
        "nvidia-tesla-a100": {"nvidia-tesla-v100"},     // A100 可以兼容 V100 的工作负载
        "nvidia-tesla-t4":   {"nvidia-rtx-3080", "nvidia-rtx-3090"}, // T4 与消费级高端 GPU 兼容
    }
    
    // 查找请求的 GPU 类型是否有兼容列表
    if compatibleTypes, ok := compatibility[requested]; ok {
        // 遍历兼容列表，检查可用的 GPU 类型是否在其中
        for _, compatibleType := range compatibleTypes {
            if compatibleType == available {
                return true  // 找到兼容的 GPU 类型
            }
        }
    }
    
    return false  // 没有找到兼容的 GPU 类型
}

// NewGPUScore 插件工厂函数
// 用于创建 GPUScore 插件实例，符合 Kubernetes 调度器插件注册规范
// 在调度器启动时被调用，初始化插件实例
func NewGPUScore(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
    return &GPUScore{handle: h}, nil
}
```

### 5.3 调度器配置

```yaml
# GPU 调度器配置 ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: gpu-scheduler-config
  namespace: kube-system
data:
  config.yaml: |
    # GPU 调度器配置文件
    # 定义专门用于 GPU 工作负载的自定义调度器配置
    apiVersion: kubescheduler.config.k8s.io/v1beta3  # Kubernetes 调度器配置 API 版本
    kind: KubeSchedulerConfiguration                  # 配置类型：调度器配置
    profiles:                                         # 调度器配置文件列表
    - schedulerName: gpu-scheduler                    # 自定义调度器名称，Pod 可通过此名称指定使用该调度器
      plugins:                                        # 插件配置
        filter:                                       # 过滤阶段插件
          enabled:                                    # 启用的过滤插件列表
          - name: GPUFilter                           # 启用 GPU 过滤插件
        score:                                        # 评分阶段插件
          enabled:                                    # 启用的评分插件列表
          - name: GPUScore                            # 启用 GPU 评分插件
      pluginConfig:                                   # 插件具体配置
      - name: GPUFilter                               # GPU 过滤插件配置
        args:                                         # 插件参数
          gpuTypes:                                   # 支持的 GPU 类型列表
          - nvidia-tesla-v100                         # NVIDIA Tesla V100（数据中心级 GPU）
          - nvidia-tesla-a100                         # NVIDIA Tesla A100（最新数据中心 GPU）
          - nvidia-tesla-t4                           # NVIDIA Tesla T4（推理优化 GPU）
      - nvidia-rtx-3080                           # NVIDIA RTX 3080（高性能消费级 GPU）
           - nvidia-rtx-3090                           # NVIDIA RTX 3090（旗舰消费级 GPU）
      - name: GPUScore                                # GPU 评分插件配置
        args:                                         # 插件参数
          weights:                                    # 评分权重配置
            utilization: 40                           # 资源利用率权重（40%）
            typeMatch: 30                             # 类型匹配权重（30%）
            topology: 30                              # 拓扑结构权重（30%）
```

## 6 GPU 工作负载示例

### 6.1 机器学习训练任务

```yaml
# 机器学习训练任务配置
# 使用 Job 资源类型进行一次性的模型训练任务
apiVersion: batch/v1                              # Kubernetes Job API 版本
kind: Job                                         # 资源类型：批处理任务
metadata:
  name: ml-training-job                           # Job 名称
  namespace: ml-team                              # 命名空间，用于多租户隔离
spec:
  template:                                       # Pod 模板定义
    metadata:
      labels:
        app: ml-training                          # 应用标签
        workload-type: training                   # 工作负载类型标签，用于调度器识别
    spec:
      schedulerName: gpu-scheduler                # 指定使用自定义 GPU 调度器
      restartPolicy: Never                        # 失败时不重启，适合训练任务
      containers:
      - name: tensorflow                          # 容器名称
        image: tensorflow/tensorflow:2.8.0-gpu   # TensorFlow GPU 镜像
        command:                                  # 启动命令
        - python
        - /app/train.py                           # 训练脚本路径
        - --epochs=100                            # 训练轮数
        - --batch-size=32                         # 批处理大小
        - --gpu-count=4                           # 使用的 GPU 数量
        resources:                                # 资源配置
          requests:                               # 资源请求（调度依据）
            nvidia.com/gpu: "4"                  # 请求 4 个 GPU
            cpu: "8"                             # 请求 8 个 CPU 核心
            memory: "32Gi"                       # 请求 32GB 内存
          limits:                                 # 资源限制（运行时上限）
            nvidia.com/gpu: "4"                  # 限制最多使用 4 个 GPU
            cpu: "16"                            # 限制最多使用 16 个 CPU 核心
            memory: "64Gi"                       # 限制最多使用 64GB 内存
        env:                                      # 环境变量
        - name: CUDA_VISIBLE_DEVICES              # CUDA 可见设备配置
          value: "0,1,2,3"                       # 指定使用 GPU 0-3
        volumeMounts:                             # 卷挂载
        - name: training-data                     # 训练数据卷
          mountPath: /data                        # 挂载到容器内的 /data 目录
        - name: model-output                      # 模型输出卷
          mountPath: /output                      # 挂载到容器内的 /output 目录
      nodeSelector:                               # 节点选择器
        accelerator: nvidia-tesla-v100           # 指定 GPU 类型为 Tesla V100
        gpu-count: "8"                           # 确保节点有足够的 GPU（至少8个）
      volumes:                                    # 卷定义
      - name: training-data                       # 训练数据卷
        persistentVolumeClaim:                    # 使用持久卷声明
          claimName: training-data-pvc            # PVC 名称
      - name: model-output                        # 模型输出卷
        persistentVolumeClaim:                    # 使用持久卷声明
          claimName: model-output-pvc             # PVC 名称
```

### 6.2 推理服务

```yaml
# 推理服务部署配置
# 使用 Deployment 资源类型部署可扩展的推理服务
apiVersion: apps/v1                               # Kubernetes Deployment API 版本
kind: Deployment                                  # 资源类型：部署
metadata:
  name: inference-service                         # Deployment 名称
  namespace: ml-team                              # 命名空间，用于多租户隔离
spec:
  replicas: 3                                     # 副本数量，提供高可用性
  selector:                                       # 选择器，用于匹配 Pod
    matchLabels:
      app: inference-service                      # 匹配标签
  template:                                       # Pod 模板定义
    metadata:
      labels:
        app: inference-service                    # 应用标签
        workload-type: inference                 # 工作负载类型标签，用于调度器识别
    spec:
      schedulerName: gpu-scheduler                # 指定使用自定义 GPU 调度器
      containers:
      - name: inference                           # 容器名称
        image: pytorch/torchserve:0.6.0-gpu      # PyTorch TorchServe GPU 镜像
        ports:                                    # 容器端口配置
        - containerPort: 8080                     # 推理服务端口
        - containerPort: 8081                     # 管理端口
        resources:                                # 资源配置
          requests:                               # 资源请求（调度依据）
            nvidia.com/gpu: "1"                  # 请求 1 个 GPU
            cpu: "2"                             # 请求 2 个 CPU 核心
            memory: "8Gi"                        # 请求 8GB 内存
          limits:                                 # 资源限制（运行时上限）
            nvidia.com/gpu: "1"                  # 限制最多使用 1 个 GPU
            cpu: "4"                             # 限制最多使用 4 个 CPU 核心
            memory: "16Gi"                       # 限制最多使用 16GB 内存
        env:                                      # 环境变量
        - name: CUDA_VISIBLE_DEVICES              # CUDA 可见设备配置
          value: "0"                             # 指定使用 GPU 0
        livenessProbe:                            # 存活性探针
          httpGet:
            path: /ping                           # 健康检查路径
            port: 8080                            # 健康检查端口
          initialDelaySeconds: 30                 # 初始延迟 30 秒
          periodSeconds: 10                       # 检查间隔 10 秒
        readinessProbe:                           # 就绪性探针
          httpGet:
            path: /ping                           # 就绪检查路径
            port: 8080                            # 就绪检查端口
          initialDelaySeconds: 10                 # 初始延迟 10 秒
          periodSeconds: 5                        # 检查间隔 5 秒
      nodeSelector:                               # 节点选择器
        accelerator: nvidia-tesla-t4             # 指定 GPU 类型为 Tesla T4（推理优化）
---
# 推理服务网络配置
# 使用 Service 资源类型暴露推理服务
apiVersion: v1                                    # Kubernetes Service API 版本
kind: Service                                     # 资源类型：服务
metadata:
  name: inference-service                         # Service 名称
  namespace: ml-team                              # 命名空间，用于多租户隔离
spec:
  selector:                                       # 选择器，用于匹配后端 Pod
    app: inference-service                        # 匹配标签
  ports:                                          # 端口配置
  - name: inference                               # 推理端口
    port: 8080                                    # 服务端口
    targetPort: 8080                              # 目标端口（Pod 端口）
  - name: management                              # 管理端口
    port: 8081                                    # 服务端口
    targetPort: 8081                              # 目标端口（Pod 端口）
  type: LoadBalancer                              # 服务类型：负载均衡器（云环境自动分配外部 IP）
```

## 7 性能优化

### 7.1 调度性能优化

**基本优化策略：**

```go
// GPUSchedulerConfig GPU 调度器性能优化配置结构
// 定义调度器的批量处理和缓存策略，用于提升调度性能
type GPUSchedulerConfig struct {
    // 批量处理配置
    // 将多个调度请求打包处理，减少系统调用开销
    BatchSize    int           `yaml:"batchSize"`    // 批处理大小，一次处理的 Pod 数量
    BatchTimeout time.Duration `yaml:"batchTimeout"` // 批处理超时时间，避免延迟过高
    
    // 缓存配置
    // 缓存调度结果和节点信息，避免重复计算
    CacheSize    int           `yaml:"cacheSize"`    // 缓存大小，存储的条目数量
    CacheTTL     time.Duration `yaml:"cacheTTL"`     // 缓存生存时间，过期自动清理
}

// DefaultConfig 默认性能优化配置
// 基于生产环境经验调优的参数，平衡性能和资源消耗
var DefaultConfig = GPUSchedulerConfig{
    BatchSize:    20,                    // 批处理 20 个 Pod，适合中等规模集群
    BatchTimeout: 50 * time.Millisecond, // 50ms 超时，保证调度延迟在可接受范围
    CacheSize:    1000,                  // 缓存 1000 个条目，覆盖常用节点和 Pod 组合
    CacheTTL:     5 * time.Minute,       // 5 分钟 TTL，平衡数据新鲜度和缓存命中率
}

// SchedulingMetrics 调度性能监控指标结构
// 使用 Prometheus 指标体系，提供全面的性能可观测性
type SchedulingMetrics struct {
    LatencyHistogram     *prometheus.HistogramVec // 调度延迟直方图，支持 P50/P95/P99 统计
    SuccessRate         *prometheus.CounterVec   // 调度成功率计数器，按成功/失败分类
    GPUFragmentation    *prometheus.GaugeVec     // GPU 碎片化率仪表，监控资源利用效率
    QueueDepth          *prometheus.GaugeVec     // 调度队列深度仪表，监控系统负载
    SchedulingConflicts *prometheus.CounterVec   // 调度冲突计数器，识别资源竞争问题
}

// recordMetrics 性能指标收集函数
// 记录调度过程中的关键性能指标，用于监控和告警
func (gs *GPUScore) recordMetrics(latency time.Duration, success bool, nodeCount int) {
    // 记录调度延迟直方图，支持百分位数统计（P50/P95/P99）
    // 用于分析调度性能分布和识别性能瓶颈
    gs.metrics.LatencyHistogram.WithLabelValues("gpu-scheduling").Observe(latency.Seconds())
    
    // 记录调度成功率计数器
    // 用于监控调度器的可靠性和稳定性
    if success {
        gs.metrics.SuccessRate.WithLabelValues("success").Inc()  // 成功调度计数
    } else {
        gs.metrics.SuccessRate.WithLabelValues("failure").Inc()  // 失败调度计数
    }
    
    // 记录调度队列深度
    // 用于监控系统负载和识别调度瓶颈
    gs.metrics.QueueDepth.WithLabelValues("pending").Set(float64(nodeCount))
}

// optimizeScheduling 优化的调度实现函数
// 集成缓存机制、性能监控和告警，提升调度效率和可观测性
func (gs *GPUScore) optimizeScheduling(pod *v1.Pod, nodeInfo *framework.NodeInfo) int64 {
    start := time.Now()  // 记录开始时间，用于计算调度延迟
    
    // 使用 defer 确保性能指标始终被记录
    defer func() {
        latency := time.Since(start)
        gs.recordMetrics(latency, true, 1)  // 记录性能指标
        
        // 性能告警机制
        // 当调度延迟超过阈值时发出告警，便于及时发现性能问题
        if latency > 100*time.Millisecond {
            klog.Warningf("Slow GPU scheduling: %v for pod %s", latency, pod.Name)
        }
    }()
    
    // 缓存机制实现
    // 使用节点名称和 Pod UID 组合作为缓存键，避免重复计算
    nodeKey := fmt.Sprintf("%s-%s", nodeInfo.Node().Name, pod.UID)
    if cachedScore, exists := gs.scoreCache[nodeKey]; exists {
        return cachedScore  // 缓存命中，直接返回缓存的评分
    }
    
    // 执行实际的评分计算
    // 只有在缓存未命中时才进行复杂的评分计算
    score := gs.calculateAdvancedGPUScore(pod, nodeInfo)
    gs.scoreCache[nodeKey] = score  // 将计算结果存入缓存
    return score
}
```

### 7.2 基础监控配置

**监控指标体系：**

```yaml
# GPU 调度器监控指标配置
# 定义全面的监控指标体系，覆盖性能、资源、业务和健康四个维度
apiVersion: v1                                    # Kubernetes ConfigMap API 版本
kind: ConfigMap                                   # 资源类型：配置映射
metadata:
  name: gpu-scheduler-metrics-enhanced           # ConfigMap 名称
data:
  metrics.yaml: |                               # 监控配置数据
    metrics:
      # 调度性能指标
      # 监控调度器的核心性能表现，用于性能优化和问题诊断
      scheduling_performance:
        latency:                                  # 调度延迟指标
          enabled: true                          # 启用延迟监控
          buckets: [0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0]  # 延迟分桶，支持 P50/P95/P99 统计
          labels: ["scheduler_name", "gpu_type", "workload_type"]  # 标签维度
        
        success_rate:                           # 调度成功率指标
          enabled: true                          # 启用成功率监控
          labels: ["reason", "namespace", "priority"]  # 按失败原因、命名空间、优先级分类
        
        queue_depth:                            # 调度队列深度指标
          enabled: true                          # 启用队列深度监控
          interval: 10s                          # 10秒采集间隔
        
        conflicts:                              # 调度冲突指标
          enabled: true                          # 启用冲突监控
          labels: ["conflict_type", "resource_type"]  # 按冲突类型和资源类型分类
      
      # GPU 资源指标
      # 监控 GPU 硬件资源的使用情况和健康状态
      gpu_resources:
        utilization:                            # GPU 利用率指标
          enabled: true                          # 启用利用率监控
          interval: 30s                          # 30秒采集间隔
          labels: ["node", "gpu_type", "gpu_index"]  # 按节点、GPU类型、GPU索引分类
        
        memory_utilization:                     # GPU 内存利用率指标
          enabled: true                          # 启用内存利用率监控
          interval: 30s                          # 30秒采集间隔
        
        temperature:                            # GPU 温度指标
          enabled: true                          # 启用温度监控
          interval: 60s                          # 60秒采集间隔
          threshold: 80                          # 温度告警阈值（摄氏度）
        
        power_consumption:                      # GPU 功耗指标
          enabled: true                          # 启用功耗监控
          interval: 60s                          # 60秒采集间隔
        
        error_count:                            # GPU 错误计数指标
          enabled: true                          # 启用错误计数监控
          labels: ["error_type", "severity"]     # 按错误类型和严重程度分类
        
        fragmentation_rate:                     # GPU 碎片化率指标
          enabled: true                          # 启用碎片化率监控
          interval: 300s                         # 5分钟计算一次（计算密集型）
      
      # 业务级 SLA 指标
      # 监控面向用户的服务质量指标，确保业务目标达成
      business_sla:
        training_job_completion_rate:           # 训练任务完成率指标
          enabled: true                          # 启用完成率监控
          target: 95                             # 目标完成率 95%
          alert_threshold: 90                    # 告警阈值 90%
        
        inference_latency:                      # 推理延迟指标
          enabled: true                          # 启用推理延迟监控
          target: 100                            # 目标延迟 100ms
          alert_threshold: 150                   # 告警阈值 150ms
        
        resource_allocation_time:               # 资源分配时间指标
          enabled: true                          # 启用分配时间监控
          target: 30                             # 目标分配时间 30秒
          alert_threshold: 60                    # 告警阈值 60秒
      
      # 系统健康指标
      # 监控系统整体健康状况和可靠性指标
      system_health:
        mttr:                                   # 平均修复时间指标
          enabled: true                          # 启用 MTTR 监控
          calculation_window: "24h"              # 24小时计算窗口
        
        mtbf:                                   # 平均故障间隔时间指标
          enabled: true                          # 启用 MTBF 监控
          calculation_window: "720h"             # 30天计算窗口
        
        availability:                           # 系统可用性指标
          enabled: true                          # 启用可用性监控
          target: 99.9                           # 目标可用性 99.9%
```

**性能监控实现：**

```go
// Score 带性能监控的评分函数实现
// 集成全面的性能监控和告警机制，提供可观测性
func (gs *GPUScore) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
    start := time.Now()  // 记录开始时间，用于计算调度延迟
    
    // 提取工作负载类型和 GPU 类型标签
    // 用于细粒度的性能监控和问题定位
    workloadType := "unknown"
    if wt, ok := pod.Labels["workload-type"]; ok {
        workloadType = wt  // 获取工作负载类型（training/inference/development）
    }
    
    gpuType := "unknown"
    if gt, ok := pod.Spec.NodeSelector["accelerator"]; ok {
        gpuType = gt  // 获取请求的 GPU 类型
    }
    
    // 使用 defer 确保性能指标始终被记录
    defer func() {
        latency := time.Since(start)
        
        // 记录详细的性能指标，包含多维度标签
        // 支持按调度器、GPU类型、工作负载类型进行性能分析
        gs.metrics.LatencyHistogram.WithLabelValues(
            "gpu-scheduler", gpuType, workloadType,
        ).Observe(latency.Seconds())
        
        // 分级性能告警机制
        // 根据延迟严重程度采取不同的告警策略
        if latency > 200*time.Millisecond {
            // 严重延迟：记录错误日志并增加冲突计数
            klog.Errorf("Critical GPU scheduling latency: %v for pod %s", latency, pod.Name)
            gs.metrics.SchedulingConflicts.WithLabelValues("high_latency", "gpu").Inc()
        } else if latency > 100*time.Millisecond {
            // 高延迟：记录警告日志
            klog.Warningf("High GPU scheduling latency: %v for pod %s", latency, pod.Name)
        }
    }()
    
    // 获取节点信息
    // 从调度器缓存中获取节点详细信息
    nodeInfo, err := gs.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
    if err != nil {
        // 节点不存在错误处理
        // 记录失败指标，包含详细的错误原因和上下文
        gs.metrics.SuccessRate.WithLabelValues("node_not_found", pod.Namespace, 
            pod.Spec.PriorityClassName).Inc()
        return 0, framework.NewStatus(framework.Error, err.Error())
    }
    
    // 执行 GPU 评分计算
    // 调用高级评分算法，综合考虑多个维度
    score := gs.calculateAdvancedGPUScore(pod, nodeInfo)
    
    // 记录成功调度指标
    // 按命名空间和优先级类别分类统计成功率
    gs.metrics.SuccessRate.WithLabelValues("success", pod.Namespace, 
        pod.Spec.PriorityClassName).Inc()
    
    return score, framework.NewStatus(framework.Success, "")
}

// monitorGPUFragmentation GPU 碎片率监控函数
// 定期监控集群中所有节点的 GPU 资源碎片化情况
// 碎片化率是衡量资源利用效率的重要指标
func (gs *GPUScore) monitorGPUFragmentation() {
    ticker := time.NewTicker(5 * time.Minute)  // 每5分钟执行一次监控
    defer ticker.Stop()
    
    for range ticker.C {
        // 获取集群中所有节点信息
        nodes, err := gs.handle.SnapshotSharedLister().NodeInfos().List()
        if err != nil {
            continue  // 获取失败时跳过本次监控
        }
        
        // 遍历所有节点，计算每个节点的碎片化率
        for _, nodeInfo := range nodes {
            node := nodeInfo.Node()
            if !hasGPU(node) {
                continue  // 跳过没有 GPU 的节点
            }
            
            // 计算节点的 GPU 碎片化率
            fragmentationRate := gs.calculateNodeFragmentationRate(nodeInfo)
            
            // 记录碎片化率指标，按节点名称和 GPU 类型分类
            gs.metrics.GPUFragmentation.WithLabelValues(
                node.Name,                      // 节点名称
                node.Labels["accelerator"],    // GPU 类型标签
            ).Set(fragmentationRate)
        }
    }
}

// calculateNodeFragmentationRate 计算节点 GPU 碎片化率
// 碎片化率 = 无法满足最大单个请求的可用资源 / 总可用资源
// 值越高表示资源越碎片化，调度效率越低
func (gs *GPUScore) calculateNodeFragmentationRate(nodeInfo *framework.NodeInfo) float64 {
    node := nodeInfo.Node()
    totalGPU := getNodeGPUCapacity(node)  // 获取节点总 GPU 容量
    if totalGPU == 0 {
        return 0  // 没有 GPU 资源，碎片化率为 0
    }
    
    // 计算已使用的 GPU 资源
    usedGPU := int64(0)
    // 遍历节点上所有 Pod，累计已使用的 GPU 资源
    for _, podInfo := range nodeInfo.Pods {
        usedGPU += getGPURequest(podInfo.Pod)  // 累加所有 Pod 的 GPU 请求
    }
    
    // 计算可用但可能碎片化的 GPU 资源
    availableGPU := totalGPU - usedGPU
    if availableGPU <= 0 {
        return 0  // 无可用 GPU 资源，碎片率为 0
    }
    
    // 简化的碎片率计算：可用资源无法满足最大单个请求的比例
    // 如果当前可用 GPU 无法满足队列中最大的单个请求，则认为存在碎片化
    maxSingleRequest := gs.getMaxGPURequestInQueue()
    if maxSingleRequest > availableGPU {
        // 返回碎片化程度：可用资源占总资源的比例
        return float64(availableGPU) / float64(totalGPU)
    }
    
    return 0  // 无碎片化
}

// 获取队列中最大的 GPU 请求
// 用于评估当前可用资源是否能满足待调度的最大请求
func (gs *GPUScore) getMaxGPURequestInQueue() int64 {
    // 这里应该从调度队列中获取，简化实现
    // 实际应该遍历调度队列，找到最大的 GPU 请求数量
    return 1
}

// 检查节点是否有 GPU
// 通过检查节点容量中是否包含 GPU 资源来判断
func hasGPU(node *v1.Node) bool {
    _, hasGPU := node.Status.Capacity["nvidia.com/gpu"]
    return hasGPU
}
```

### 7.3 简化监控部署

```yaml
# 基础 GPU 监控服务配置
apiVersion: v1  # Kubernetes API 版本
kind: Service   # 资源类型：服务
metadata:
  name: gpu-scheduler-metrics      # 服务名称
  namespace: kube-system           # 命名空间：系统级组件
  labels:
    app: gpu-scheduler             # 标签：用于标识 GPU 调度器
spec:
  ports:
  - name: metrics                  # 端口名称：指标端口
    port: 8080                     # 服务端口：对外暴露的端口
    targetPort: 8080               # 目标端口：Pod 内部端口
  selector:
    app: gpu-scheduler             # 选择器：匹配带有此标签的 Pod
---
# 简化的告警规则配置
apiVersion: v1  # Kubernetes API 版本
kind: ConfigMap # 资源类型：配置映射
metadata:
  name: gpu-alerts                 # ConfigMap 名称
data:
  alerts.yaml: |                   # 告警规则配置文件
    groups:
    - name: gpu-basic              # 告警组名称：基础 GPU 告警
      rules:
      - alert: GPUSchedulingLatencyHigh  # 告警名称：GPU 调度延迟过高
        expr: gpu_scheduler_latency > 0.1  # 告警条件：调度延迟超过 100ms
        annotations:
          summary: "GPU 调度延迟过高"    # 告警摘要信息
      
      - alert: GPUUtilizationLow     # 告警名称：GPU 利用率偏低
        expr: gpu_utilization < 60    # 告警条件：GPU 利用率低于 60%
        annotations:
          summary: "GPU 利用率偏低"     # 告警摘要信息
```

## 8 安全加固与访问控制

### 8.1 RBAC 配置

```yaml
# GPU 资源访问控制 - 管理员角色
apiVersion: rbac.authorization.k8s.io/v1  # RBAC API 版本
kind: ClusterRole                         # 资源类型：集群角色
metadata:
  name: gpu-resource-manager              # 角色名称：GPU 资源管理员
rules:
- apiGroups: [""]                        # 核心 API 组
  resources: ["nodes"]                   # 资源类型：节点
  verbs: ["get", "list", "watch"]        # 允许的操作：查看、列出、监听
- apiGroups: [""]                        # 核心 API 组
  resources: ["pods"]                    # 资源类型：Pod
  verbs: ["get", "list", "watch", "create", "update", "patch"]  # 允许的操作：完整的 CRUD 权限
- apiGroups: [""]                        # 核心 API 组
  resources: ["events"]                  # 资源类型：事件
  verbs: ["create"]                      # 允许的操作：创建事件
- apiGroups: ["metrics.k8s.io"]          # 指标 API 组
  resources: ["nodes", "pods"]           # 资源类型：节点和 Pod 指标
  verbs: ["get", "list"]                 # 允许的操作：查看、列出指标
---
# GPU 资源访问控制 - 只读角色
apiVersion: rbac.authorization.k8s.io/v1  # RBAC API 版本
kind: ClusterRole                         # 资源类型：集群角色
metadata:
  name: gpu-resource-reader               # 角色名称：GPU 资源只读用户
rules:
- apiGroups: [""]                        # 核心 API 组
  resources: ["nodes"]                   # 资源类型：节点
  verbs: ["get", "list"]                 # 允许的操作：查看、列出（只读）
  resourceNames: []                       # 资源名称限制：空表示所有资源
- apiGroups: [""]                        # 核心 API 组
  resources: ["pods"]                    # 资源类型：Pod
  verbs: ["get", "list"]                 # 允许的操作：查看、列出（只读）
  resourceNames: []                       # 资源名称限制：空表示所有资源
- apiGroups: ["metrics.k8s.io"]          # 指标 API 组
  resources: ["nodes", "pods"]           # 资源类型：节点和 Pod 指标
  verbs: ["get", "list"]                 # 允许的操作：查看、列出指标
---
# 角色绑定 - 将只读角色绑定给机器学习团队
apiVersion: rbac.authorization.k8s.io/v1  # RBAC API 版本
kind: RoleBinding                         # 资源类型：角色绑定（命名空间级别）
metadata:
  name: gpu-team-access                   # 绑定名称
  namespace: ml-team                      # 命名空间：机器学习团队
subjects:                                 # 绑定主体
- kind: User                              # 主体类型：用户
  name: ml-engineer                       # 用户名：机器学习工程师
  apiGroup: rbac.authorization.k8s.io     # API 组
- kind: ServiceAccount                    # 主体类型：服务账户
  name: ml-workload                       # 服务账户名：机器学习工作负载
  namespace: ml-team                      # 服务账户所在命名空间
roleRef:                                  # 角色引用
  kind: ClusterRole                       # 角色类型：集群角色
  name: gpu-resource-reader               # 角色名称：GPU 资源只读用户
  apiGroup: rbac.authorization.k8s.io     # API 组
```

### 8.2 基础资源隔离

```yaml
# 简化的资源配额 - 限制命名空间总资源使用量
apiVersion: v1          # Kubernetes API 版本
kind: ResourceQuota     # 资源类型：资源配额
metadata:
  name: gpu-quota-ml-team     # 配额名称：机器学习团队 GPU 配额
  namespace: ml-team          # 命名空间：机器学习团队
spec:
  hard:                       # 硬性限制
    nvidia.com/gpu: "4"       # GPU 总数限制：最多 4 块 GPU
    pods: "10"                # Pod 数量限制：最多 10 个 Pod
    requests.cpu: "16"        # CPU 请求总量限制：16 核
    requests.memory: "64Gi"   # 内存请求总量限制：64GB
---
# 基础资源限制 - 限制单个容器资源使用量
apiVersion: v1          # Kubernetes API 版本
kind: LimitRange        # 资源类型：资源限制范围
metadata:
  name: gpu-limits            # 限制名称：GPU 资源限制
  namespace: ml-team          # 命名空间：机器学习团队
spec:
  limits:
  - default:                  # 默认资源限制（当容器未指定时使用）
      nvidia.com/gpu: "1"     # 默认 GPU 数量：1 块
      cpu: "2"                # 默认 CPU 限制：2 核
      memory: "8Gi"           # 默认内存限制：8GB
    max:                      # 最大资源限制
      nvidia.com/gpu: "2"     # 单个 Pod 最多 2 块 GPU
      cpu: "8"                # 单个容器最多 8 核 CPU
      memory: "32Gi"          # 单个容器最多 32GB 内存
    type: Container           # 限制类型：容器级别
```

## 9 故障处理与恢复

### 9.1 基础健康检查

**故障检测机制：**

```go
// 健康检查实现
// GPUHealthChecker 负责监控 GPU 节点和调度器的健康状态
type GPUHealthChecker struct {
    client       kubernetes.Interface        // Kubernetes 客户端，用于访问集群资源
    nodeInformer cache.SharedIndexInformer   // 节点信息缓存，用于监听节点状态变化
    metrics      *HealthMetrics              // 健康检查指标收集器
    alertManager AlertManager               // 告警管理器，用于发送告警通知
}

// 健康检查指标
// HealthMetrics 定义了各种健康状态监控指标
type HealthMetrics struct {
    NodeHealthStatus    *prometheus.GaugeVec     // 节点健康状态指标（正常/异常）
    GPUErrorRate       *prometheus.CounterVec   // GPU 错误率计数器
    RecoveryAttempts   *prometheus.CounterVec   // 恢复尝试次数计数器
    MTTR              *prometheus.HistogramVec  // 平均修复时间（Mean Time To Repair）
    MTBF              *prometheus.GaugeVec      // 平均故障间隔时间（Mean Time Between Failures）
}

// 启动健康检查
// Start 启动多层次的健康检查服务，包括节点、GPU、调度器和自动恢复
func (hc *GPUHealthChecker) Start(ctx context.Context) error {
    // 启动多层次健康检查
    go hc.runNodeHealthCheck(ctx)      // 启动节点健康检查
    go hc.runGPUHealthCheck(ctx)       // 启动 GPU 设备健康检查
    go hc.runSchedulerHealthCheck(ctx) // 启动调度器健康检查
    go hc.runAutoRecovery(ctx)         // 启动自动恢复机制
    
    return nil
}

// 节点级健康检查
// runNodeHealthCheck 定期执行节点健康状态检查，每 30 秒检查一次
func (hc *GPUHealthChecker) runNodeHealthCheck(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)  // 创建 30 秒间隔的定时器
    defer ticker.Stop()                         // 函数退出时停止定时器
    
    for {
        select {
        case <-ctx.Done():          // 接收到上下文取消信号
            return                  // 退出健康检查循环
        case <-ticker.C:            // 定时器触发
            hc.checkNodeHealth()    // 执行节点健康检查
        }
    }
}

// 检查节点健康状态
// checkNodeHealth 遍历所有 GPU 节点，评估其健康状态并记录指标
func (hc *GPUHealthChecker) checkNodeHealth() {
    // 获取所有带有 accelerator 标签的 GPU 节点
    nodes, err := hc.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
        LabelSelector: "accelerator",  // 筛选带有加速器标签的节点
    })
    if err != nil {
        klog.Errorf("Failed to list GPU nodes: %v", err)
        return
    }
    
    // 遍历每个 GPU 节点，检查其健康状态
    for _, node := range nodes.Items {
        healthStatus := hc.evaluateNodeHealth(&node)  // 评估节点健康状态
        
        // 记录健康状态指标到 Prometheus
        hc.metrics.NodeHealthStatus.WithLabelValues(
            node.Name,                      // 节点名称
            node.Labels["accelerator"],     // 加速器类型（如 nvidia-tesla-v100）
            string(healthStatus),           // 健康状态（健康/降级/不健康）
        ).Set(float64(healthStatus))
        
        // 处理不健康节点，触发告警和恢复机制
        if healthStatus == NodeUnhealthy {
            hc.handleUnhealthyNode(&node)
        }
    }
}

// 节点健康状态枚举
// NodeHealthStatus 定义节点的健康状态级别
type NodeHealthStatus int

const (
    NodeHealthy   NodeHealthStatus = iota  // 节点健康：所有 GPU 正常工作
    NodeDegraded                           // 节点降级：部分 GPU 异常但仍可用
    NodeUnhealthy                          // 节点不健康：GPU 严重故障或不可用
)

// 评估节点健康状态
// evaluateNodeHealth 综合评估节点的健康状态，包括就绪状态、GPU 设备状态和资源压力
func (hc *GPUHealthChecker) evaluateNodeHealth(node *v1.Node) NodeHealthStatus {
    // 检查节点就绪状态
    if !isNodeReady(node) {
        return NodeUnhealthy  // 节点未就绪，标记为不健康
    }
    
    // 检查 GPU 设备状态
    gpuCapacity, hasGPU := node.Status.Capacity["nvidia.com/gpu"]
    if !hasGPU {
        return NodeUnhealthy  // 节点没有 GPU 资源，标记为不健康
    }
    
    // 比较 GPU 容量和可分配资源，检查是否有 GPU 故障
    gpuAllocatable, _ := node.Status.Allocatable["nvidia.com/gpu"]
    if gpuCapacity.Cmp(gpuAllocatable) != 0 {
        return NodeDegraded  // GPU 容量与可分配资源不匹配，可能有部分 GPU 故障
    }
    
    // 检查资源压力状况
    for _, condition := range node.Status.Conditions {
        switch condition.Type {
        case v1.NodeMemoryPressure, v1.NodeDiskPressure:  // 检查内存和磁盘压力
            if condition.Status == v1.ConditionTrue {
                return NodeDegraded  // 存在资源压力，节点性能降级
            }
        }
    }
    
    return NodeHealthy  // 所有检查通过，节点健康
}

// 处理不健康节点
// handleUnhealthyNode 对检测到的不健康节点执行隔离、驱逐和告警操作
func (hc *GPUHealthChecker) handleUnhealthyNode(node *v1.Node) {
    klog.Warningf("Detected unhealthy GPU node: %s", node.Name)
    
    // 标记节点为不可调度（cordon），防止新的 Pod 调度到该节点
    if err := hc.cordonNode(node.Name); err != nil {
        klog.Errorf("Failed to cordon node %s: %v", node.Name, err)
        return
    }
    
    // 尝试驱逐节点上的 GPU 工作负载，将其重新调度到健康节点
    if err := hc.evictGPUPods(node.Name); err != nil {
        klog.Errorf("Failed to evict GPU pods from node %s: %v", node.Name, err)
    }
    
    // 发送关键级别告警通知运维人员
    hc.alertManager.SendAlert(Alert{
        Level:   "critical",                                    // 告警级别：关键
        Message: fmt.Sprintf("GPU node %s is unhealthy", node.Name), // 告警消息
        Labels: map[string]string{
            "node":     node.Name,                              // 节点名称标签
            "gpu_type": node.Labels["accelerator"],            // GPU 类型标签
        },
    })
    
    // 记录恢复尝试次数到监控指标
    hc.metrics.RecoveryAttempts.WithLabelValues(
        node.Name, "node_isolation",  // 节点名称和恢复类型
    ).Inc()  // 计数器加一
}
```

**简化的健康检查脚本：**

```bash
#!/bin/bash
# gpu-health-check.sh - GPU 集群健康检查脚本
# 用于快速检查 GPU 节点、Pod 和设备插件的状态

# 检查 GPU 节点状态
# 显示所有带有 nvidia 加速器标签的节点信息
check_gpu_nodes() {
    echo "检查 GPU 节点状态..."
    # 列出所有 GPU 节点的详细信息，包括状态、角色、年龄、版本等
    kubectl get nodes -l accelerator=nvidia -o wide
    
    # 检查 GPU 资源容量和可分配情况
    echo "检查 GPU 资源分配情况..."
    kubectl describe nodes -l accelerator=nvidia | grep -A 5 "nvidia.com/gpu"
}

# 检查 GPU Pod 状态
# 显示所有使用 GPU 的 Pod 运行状态
check_gpu_pods() {
    echo "检查 GPU Pod 状态..."
    # 查找所有包含 gpu 或 nvidia 关键字的 Pod
    kubectl get pods --all-namespaces -o wide | grep -E "(gpu|nvidia)"
}

# 检查 Device Plugin 状态
# 验证 NVIDIA Device Plugin 是否正常运行
check_device_plugin() {
    echo "检查 NVIDIA Device Plugin 状态..."
    # 检查 kube-system 命名空间中的 NVIDIA Device Plugin DaemonSet
    kubectl get pods -n kube-system -l name=nvidia-device-plugin-ds
}

# 主函数
# 按顺序执行所有健康检查
main() {
    echo "=== GPU 集群健康检查开始 ==="
    check_gpu_nodes      # 检查节点状态
    echo ""
    check_gpu_pods       # 检查 Pod 状态
    echo ""
    check_device_plugin  # 检查设备插件
    echo "=== GPU 集群健康检查完成 ==="
}

# 执行主函数
main
```

**基础故障恢复：**

```yaml
# GPU 节点故障恢复 Job - 自动修复 GPU 设备故障
apiVersion: batch/v1  # Kubernetes Batch API 版本
kind: Job             # 资源类型：任务
metadata:
  name: gpu-node-recovery  # Job 名称：GPU 节点恢复任务
spec:
  template:               # Pod 模板
    spec:
      restartPolicy: OnFailure  # 重启策略：失败时重启
      containers:
      - name: recovery          # 容器名称：恢复容器
        image: nvidia/cuda:11.8-base-ubuntu20.04  # 镜像：NVIDIA CUDA 基础镜像
        command:                # 执行命令
        - /bin/bash
        - -c
        - |
          # 重启 NVIDIA 驱动，尝试恢复 GPU 设备
          echo "正在重置 GPU 设备..."
          nvidia-smi --gpu-reset
          
          # 检查 GPU 状态，验证恢复结果
          echo "检查 GPU 设备状态..."
          nvidia-smi
          
          # 输出恢复结果
          if [ $? -eq 0 ]; then
            echo "GPU 设备恢复成功"
          else
            echo "GPU 设备恢复失败，需要人工干预"
            exit 1
          fi
        securityContext:
          privileged: true      # 特权模式：访问主机设备需要特权
        volumeMounts:
        - name: dev             # 挂载设备目录
          mountPath: /dev       # 挂载路径：设备文件系统
      volumes:
      - name: dev             # 卷名称：设备卷
        hostPath:             # 主机路径类型
          path: /dev          # 主机设备目录路径
      nodeSelector:           # 节点选择器
        accelerator: nvidia   # 只在带有 NVIDIA 加速器的节点上运行
```

### 9.2 简化调度器部署

**自动恢复机制：**

```go
// GPU 级健康检查
// runGPUHealthCheck 定期执行 GPU 设备级别的健康检查，每 60 秒检查一次
func (hc *GPUHealthChecker) runGPUHealthCheck(ctx context.Context) {
    ticker := time.NewTicker(60 * time.Second)  // 创建 60 秒间隔的定时器
    defer ticker.Stop()                         // 函数退出时停止定时器
    
    for {
        select {
        case <-ctx.Done():           // 接收到上下文取消信号
            return                   // 退出 GPU 健康检查循环
        case <-ticker.C:             // 定时器触发
            hc.checkGPUHealth()      // 执行 GPU 设备健康检查
        }
    }
}

// 检查 GPU 设备健康状态
// checkGPUHealth 遍历所有 GPU 节点，检测 GPU 设备错误并进行处理
func (hc *GPUHealthChecker) checkGPUHealth() {
    // 获取所有带有加速器标签的节点
    nodes, err := hc.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
        LabelSelector: "accelerator",  // 筛选带有加速器标签的节点
    })
    if err != nil {
        klog.Errorf("Failed to list GPU nodes for health check: %v", err)
        return
    }
    
    // 遍历每个 GPU 节点，检测设备错误
    for _, node := range nodes.Items {
        gpuErrors := hc.detectGPUErrors(&node)  // 检测节点上的 GPU 错误
        for _, gpuError := range gpuErrors {
            hc.handleGPUError(&node, gpuError)  // 处理检测到的 GPU 错误
        }
    }
}

// GPU 错误类型
// GPUError 定义 GPU 设备错误的详细信息
type GPUError struct {
    NodeName  string  // 节点名称
    GPUIndex  int     // GPU 设备索引（0, 1, 2...）
    ErrorType string  // 错误类型（如：memory_error, thermal_error, driver_error）
    Severity  string  // 错误严重程度（如：warning, error, critical）
    Message   string  // 错误详细信息
}

// 处理 GPU 错误
// handleGPUError 处理检测到的 GPU 设备错误，记录指标并触发相应的恢复措施
func (hc *GPUHealthChecker) handleGPUError(node *v1.Node, gpuError GPUError) {
    // 记录 GPU 错误率指标到 Prometheus
    hc.metrics.GPUErrorRate.WithLabelValues(
        gpuError.NodeName,   // 节点名称
        gpuError.ErrorType,  // 错误类型
        gpuError.Severity,   // 错误严重程度
    ).Inc()  // 错误计数器加一
    
    // 根据错误严重程度采取不同的处理措施
    switch gpuError.Severity {
    case "critical":  // 关键错误：需要立即处理
        // 立即隔离故障 GPU，防止影响其他工作负载
        hc.isolateGPU(node, gpuError.GPUIndex)
        
        // 发送紧急告警通知运维人员
        hc.alertManager.SendAlert(Alert{
            Level:   "critical",  // 告警级别：关键
            Message: fmt.Sprintf("Critical GPU error on node %s: %s", 
                gpuError.NodeName, gpuError.Message),
            Labels: map[string]string{
                "node":       gpuError.NodeName,                    // 节点名称
                "gpu_index":  fmt.Sprintf("%d", gpuError.GPUIndex), // GPU 索引
                "error_type": gpuError.ErrorType,                  // 错误类型
            },
        })
        
    case "warning":  // 警告级别：记录日志，持续监控
        klog.Warningf("GPU warning on node %s: %s", 
            gpuError.NodeName, gpuError.Message)
    }
}

// 自动恢复机制
// runAutoRecovery 定期尝试恢复被隔离的节点，每 5 分钟执行一次
func (hc *GPUHealthChecker) runAutoRecovery(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)  // 创建 5 分钟间隔的定时器
    defer ticker.Stop()                        // 函数退出时停止定时器
    
    for {
        select {
        case <-ctx.Done():              // 接收到上下文取消信号
            return                      // 退出自动恢复循环
        case <-ticker.C:                // 定时器触发
            hc.attemptAutoRecovery()    // 尝试自动恢复
        }
    }
}

// 尝试自动恢复
// attemptAutoRecovery 遍历所有被隔离的节点，尝试自动恢复
func (hc *GPUHealthChecker) attemptAutoRecovery() {
    isolatedNodes := hc.getIsolatedNodes()  // 获取所有被隔离的节点列表
    
    // 遍历每个被隔离的节点
    for _, node := range isolatedNodes {
        if hc.canAttemptRecovery(node) {  // 检查是否可以尝试恢复
            recoveryStart := time.Now()  // 记录恢复开始时间
            
            success := hc.performNodeRecovery(node)  // 执行节点恢复操作
            
            // 记录平均修复时间（MTTR）指标
            if success {
                mttr := time.Since(recoveryStart)  // 计算修复耗时
                hc.metrics.MTTR.WithLabelValues(
                    node.Name, "auto_recovery",  // 节点名称和恢复类型
                ).Observe(mttr.Seconds())  // 记录修复时间（秒）
                
                klog.Infof("Successfully recovered node %s in %v", 
                    node.Name, mttr)
            }
        }
    }
}

// 执行节点恢复
// performNodeRecovery 执行完整的节点恢复流程，包括重启服务、验证状态和解除隔离
func (hc *GPUHealthChecker) performNodeRecovery(node *v1.Node) bool {
    // 1. 重启 Device Plugin，尝试恢复 GPU 设备识别
    if err := hc.restartDevicePlugin(node.Name); err != nil {
        klog.Errorf("Failed to restart device plugin on node %s: %v", 
            node.Name, err)
        return false  // 重启失败，恢复失败
    }
    
    // 2. 验证 GPU 状态，确保设备正常工作
    if !hc.verifyGPUStatus(node) {
        klog.Errorf("GPU status verification failed for node %s", node.Name)
        return false  // 验证失败，恢复失败
    }
    
    // 3. 解除节点隔离（uncordon），允许新的 Pod 调度到该节点
    if err := hc.uncordonNode(node.Name); err != nil {
        klog.Errorf("Failed to uncordon node %s: %v", node.Name, err)
        return false  // 解除隔离失败，恢复失败
    }
    
    return true  // 所有步骤成功，恢复成功
}
```

**调度器部署：**

```yaml
# 高可用调度器部署
# GPU 调度器高可用部署配置
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-scheduler                    # 调度器名称
  namespace: kube-system                 # 部署在系统命名空间
spec:
  replicas: 2                           # 高可用部署：2 个副本
  selector:
    matchLabels:
      app: gpu-scheduler                 # 选择器标签
  template:
    metadata:
      labels:
        app: gpu-scheduler               # Pod 标签
    spec:
      affinity:                          # Pod 亲和性配置
        podAntiAffinity:                 # Pod 反亲和性
          preferredDuringSchedulingIgnoredDuringExecution:  # 软性反亲和
          - weight: 100                  # 权重：100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: gpu-scheduler     # 匹配同类 Pod
              topologyKey: kubernetes.io/hostname  # 拓扑键：主机名（确保不同节点）
      containers:
      - name: kube-scheduler             # 容器名称
        image: registry.k8s.io/kube-scheduler:v1.28.0  # 调度器镜像
        command:
        - kube-scheduler                 # 启动命令
        - --config=/etc/kubernetes/scheduler-config.yaml  # 配置文件路径
        - --leader-elect=true            # 启用领导选举
        - --leader-elect-resource-name=gpu-scheduler  # 领导选举资源名称
        - --v=2                          # 日志级别
        volumeMounts:
        - name: config                   # 挂载配置文件
          mountPath: /etc/kubernetes
        resources:                       # 资源限制
          requests:
            cpu: 100m                    # CPU 请求：100 毫核
            memory: 128Mi                # 内存请求：128MB
          limits:
            cpu: 500m                    # CPU 限制：500 毫核
            memory: 512Mi                # 内存限制：512MB
        livenessProbe:                   # 存活性探针
          httpGet:
            path: /healthz               # 健康检查路径
            port: 10259                  # 健康检查端口
            scheme: HTTPS                # 使用 HTTPS
          initialDelaySeconds: 15        # 初始延迟：15 秒
          periodSeconds: 10              # 检查间隔：10 秒
          timeoutSeconds: 5              # 超时时间：5 秒
          failureThreshold: 3            # 失败阈值：3 次
        readinessProbe:                  # 就绪性探针
          httpGet:
            path: /healthz               # 健康检查路径
            port: 10259                  # 健康检查端口
            scheme: HTTPS                # 使用 HTTPS
          initialDelaySeconds: 5         # 初始延迟：5 秒
          periodSeconds: 5               # 检查间隔：5 秒
          timeoutSeconds: 3              # 超时时间：3 秒
          failureThreshold: 2            # 失败阈值：2 次
        # 启动探针：用于慢启动容器
        startupProbe:
          httpGet:
            path: /healthz               # 健康检查路径
            port: 10259                  # 健康检查端口
            scheme: HTTPS                # 使用 HTTPS
          initialDelaySeconds: 10        # 初始延迟：10 秒
          periodSeconds: 10              # 检查间隔：10 秒
          timeoutSeconds: 5              # 超时时间：5 秒
          failureThreshold: 30           # 失败阈值：30 次（允许较长启动时间）
      volumes:
      - name: config                     # 配置文件卷
        configMap:
          name: scheduler-config         # 来源：ConfigMap
      serviceAccountName: gpu-scheduler  # 使用专用服务账户
      tolerations:                       # 容忍度配置
      - key: node-role.kubernetes.io/master  # 容忍主节点污点
        operator: Exists                 # 操作符：存在
        effect: NoSchedule               # 效果：不调度
---
# GPU 调度器服务账户
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gpu-scheduler                    # 服务账户名称
  namespace: kube-system                 # 命名空间
---
# GPU 调度器集群角色
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gpu-scheduler                    # 角色名称
rules:
- apiGroups: [""]                       # 核心 API 组
  resources: ["nodes", "pods", "events"]  # 节点、Pod、事件资源
  verbs: ["get", "list", "watch", "create", "update", "patch"]  # 完整操作权限
- apiGroups: [""]
  resources: ["pods/eviction"]          # Pod 驱逐资源
  verbs: ["create"]                     # 创建驱逐（用于抢占调度）
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]                 # 租约资源（用于领导选举）
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]  # 完整租约管理权限
---
# GPU 调度器集群角色绑定
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: gpu-scheduler                    # 绑定名称
roleRef:
  apiGroup: rbac.authorization.k8s.io   # RBAC API 组
  kind: ClusterRole                      # 角色类型：集群角色
  name: gpu-scheduler                    # 引用的角色名称
subjects:
- kind: ServiceAccount                   # 主体类型：服务账户
  name: gpu-scheduler                    # 服务账户名称
  namespace: kube-system                 # 服务账户命名空间
```

**基础 GPU 调度器部署：**

```yaml
# 基础 GPU 调度器部署（适用于开发和测试环境）
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-scheduler                    # 调度器名称
  namespace: kube-system                 # 部署在系统命名空间
spec:
  replicas: 1                           # 案例使用单副本（非高可用）
  selector:
    matchLabels:
      app: gpu-scheduler                 # 选择器标签
  template:
    metadata:
      labels:
        app: gpu-scheduler               # Pod 标签
    spec:
      serviceAccountName: gpu-scheduler  # 使用专用服务账户
      containers:
      - name: kube-scheduler             # 容器名称
        image: registry.k8s.io/kube-scheduler:v1.28.0  # 调度器镜像
        command:
        - kube-scheduler                 # 启动命令
        - --config=/etc/kubernetes/scheduler-config.yaml  # 配置文件路径
        - --v=2                          # 日志级别
        resources:                       # 资源限制（较小配置）
          requests:
            cpu: 100m                    # CPU 请求：100 毫核
            memory: 128Mi                # 内存请求：128MB
          limits:
            cpu: 200m                    # CPU 限制：200 毫核
            memory: 256Mi                # 内存限制：256MB
        volumeMounts:
        - name: config                   # 挂载配置文件
          mountPath: /etc/kubernetes
        livenessProbe:                   # 存活性探针
          httpGet:
            path: /healthz               # 健康检查路径
            port: 10259                  # 健康检查端口
            scheme: HTTPS                # 使用 HTTPS
          initialDelaySeconds: 30        # 初始延迟：30 秒
      volumes:
      - name: config                     # 配置文件卷
        configMap:
          name: gpu-scheduler-config     # 来源：ConfigMap
```

## 10 总结与最佳实践

### 10.1 需求实现总结

本方案成功实现了需求分析中提出的各项要求：

**功能需求实现：**

- ✅ GPU 资源发现和管理：通过 Device Plugin 和扩展资源实现
- ✅ 智能调度算法：实现了多维度评分和高级过滤机制
- ✅ 拓扑感知调度：支持 NUMA 和 PCIe 拓扑优化
- ✅ 资源利用率优化：通过碎片化检测和负载均衡实现
- ✅ 多租户资源隔离：基于命名空间和优先级的资源配额
- ✅ 资源配额管理：实现了细粒度的 GPU 资源配额控制
- ✅ 可扩展调度策略：提供了插件化的调度框架

**非功能需求实现：**

- ✅ 系统规模：支持 100+ 节点，1000+ GPU 的大规模部署
- ✅ 高可用性：99.9% 可用性目标，通过多副本和自动恢复实现
- ✅ 监控运维：完整的指标体系和告警机制
- ✅ 安全性：RBAC 访问控制和资源安全策略
- ✅ 可扩展性：模块化设计，支持水平扩展

**性能需求实现：**

- ✅ 调度延迟：P95 < 200ms，通过性能优化和监控实现
- ✅ 调度吞吐量：> 100 pods/min，支持并发调度
- ✅ GPU 利用率：> 80%，通过智能调度算法优化
- ✅ 系统可用性：99.9%，通过健康检查和自动恢复保障

**运维需求实现：**

- ✅ 监控需求：多层次监控体系，包含 GPU、调度器和业务指标
- ✅ 部署管理：Helm Chart 和模板化配置
- ✅ 故障处理：自动故障检测、节点隔离和恢复机制
- ✅ 日志管理：结构化日志和集中式收集

**安全需求实现：**

- ✅ 访问控制：完整的 RBAC 权限体系
- ✅ 资源安全：网络策略和镜像安全扫描

### 10.2 架构优势

1. **高性能调度**：
   - 多维度评分算法，考虑 GPU 类型、内存、拓扑等因素
   - 智能碎片化检测，提升资源利用率
   - 并发调度支持，满足高吞吐量需求

2. **企业级可靠性**：
   - 多层次健康检查和自动故障恢复
   - 完整的监控指标体系和 SLA 保障
   - 高可用部署和领导者选举机制

3. **生产级运维**：
   - 自动化部署和配置管理
   - 主动监控和智能告警
   - 故障处理和恢复流程

### 10.3 最佳实践建议

**1. 部署实施：**

```bash
# 分阶段部署策略
# 阶段1：基础环境准备
kubectl apply -f device-plugin/
kubectl apply -f monitoring/

# 阶段2：调度器部署
helm install gpu-scheduler ./charts/gpu-scheduler

# 阶段3：工作负载迁移
kubectl patch deployment <workload> -p '{"spec":{"template":{"spec":{"schedulerName":"gpu-scheduler"}}}}'
```

**2. 监控配置：**

```yaml
# 关键监控指标
monitoring:
  alerts:
    - name: "GPU调度延迟过高"
      condition: "scheduling_latency_p95 > 200ms"
      action: "自动扩容调度器副本"
    
    - name: "GPU利用率过低"
      condition: "gpu_utilization < 60%"
      action: "触发负载重平衡"
    
    - name: "节点故障"
      condition: "node_health_status == unhealthy"
      action: "自动节点隔离和工作负载迁移"
```

**3. 性能调优：**

- **调度器配置**：根据集群规模调整并发度和缓存大小
- **资源配额**：设置合理的 GPU 资源配额，避免资源争抢
- **工作负载优化**：使用合适的 GPU 请求和限制

**4. 安全加固：**

- **最小权限原则**：严格控制调度器和工作负载的权限
- **网络隔离**：使用 NetworkPolicy 限制 GPU 工作负载的网络访问
- **镜像安全**：定期扫描和更新 GPU 相关镜像

### 10.4 扩展路线图

**短期扩展（1-3个月）：**

- 支持 GPU 共享和虚拟化
- 调度算法的机器学习能力
- 多租户资源隔离

**中期扩展（3-6个月）：**

- 多 GPU 厂商支持（AMD、Intel）
- 跨集群 GPU 资源调度
- 边缘计算 GPU 管理

**长期扩展（6-12个月）：**

- AI 驱动的智能调度优化
- GPU 资源池化和动态分配
- 混合云 GPU 资源管理

### 10.5 成功指标

通过本方案的实施，预期达到以下效果：

| 指标类别 | 目标值 | 当前基线 | 改进幅度 |
|---------|--------|----------|----------|
| GPU 利用率 | > 80% | 60% | +33% |
| 调度延迟 P95 | < 200ms | 500ms | -60% |
| 系统可用性 | 99.9% | 99.5% | +0.4% |
| 故障恢复时间 | < 5min | 15min | -67% |
| 资源碎片率 | < 10% | 25% | -60% |

---
