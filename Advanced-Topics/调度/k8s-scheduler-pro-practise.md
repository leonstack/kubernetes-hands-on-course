# Kubernetes 调度器生产最佳实践

## 目录

- [Kubernetes 调度器生产最佳实践](#kubernetes-调度器生产最佳实践)
  - [目录](#目录)
  - [1. 生产环境调度器配置](#1-生产环境调度器配置)
    - [1.1 高可用调度器部署](#11-高可用调度器部署)
      - [1.1.1 多实例部署](#111-多实例部署)
      - [1.1.2 生产级调度器配置](#112-生产级调度器配置)
    - [1.2 调度器性能调优](#12-调度器性能调优)
      - [1.2.1 调度延迟优化](#121-调度延迟优化)
      - [1.2.2 内存使用优化](#122-内存使用优化)
    - [1.3 多调度器策略](#13-多调度器策略)
      - [1.3.1 工作负载专用调度器](#131-工作负载专用调度器)
      - [1.3.2 调度器选择策略](#132-调度器选择策略)
  - [2. 资源管理与优化](#2-资源管理与优化)
    - [2.1 资源配额与限制](#21-资源配额与限制)
      - [2.1.1 动态资源配额管理](#211-动态资源配额管理)
      - [2.1.2 优先级资源分配](#212-优先级资源分配)
    - [2.2 节点资源优化](#22-节点资源优化)
      - [2.2.1 节点资源监控与调优](#221-节点资源监控与调优)
    - [2.3 工作负载分类调度](#23-工作负载分类调度)
      - [2.3.1 智能工作负载分类器](#231-智能工作负载分类器)
      - [2.3.2 工作负载调度策略配置](#232-工作负载调度策略配置)
  - [3. 高级调度策略](#3-高级调度策略)
    - [3.1 自定义调度器插件](#31-自定义调度器插件)
      - [3.1.1 延迟感知调度插件](#311-延迟感知调度插件)
      - [3.1.2 成本优化调度插件](#312-成本优化调度插件)
    - [3.2 批处理调度优化](#32-批处理调度优化)
      - [3.2.1 批处理调度器实现](#321-批处理调度器实现)
      - [3.2.2 批处理调度配置](#322-批处理调度配置)
    - [3.3 边缘计算调度](#33-边缘计算调度)
      - [3.3.1 边缘节点调度器](#331-边缘节点调度器)
      - [3.3.2 边缘调度配置](#332-边缘调度配置)
      - [3.3.3 边缘节点标签配置](#333-边缘节点标签配置)
      - [3.3.4 边缘工作负载示例](#334-边缘工作负载示例)
  - [4. 监控与可观测性](#4-监控与可观测性)
    - [4.1 调度器指标监控](#41-调度器指标监控)
      - [4.1.1 Prometheus 指标收集](#411-prometheus-指标收集)
      - [4.1.2 监控配置部署](#412-监控配置部署)
    - [4.2 性能分析与诊断](#42-性能分析与诊断)
      - [4.2.1 调度器性能分析工具](#421-调度器性能分析工具)
    - [4.3 告警与自动化](#43-告警与自动化)
      - [4.3.1 Prometheus 告警规则](#431-prometheus-告警规则)
      - [4.3.2 自动化响应系统](#432-自动化响应系统)
  - [5. 故障排除与恢复](#5-故障排除与恢复)
    - [5.1 常见调度问题](#51-常见调度问题)
      - [5.1.1 调度问题分类与诊断](#511-调度问题分类与诊断)
      - [5.1.2 故障恢复策略](#512-故障恢复策略)
    - [5.2 最佳实践总结](#52-最佳实践总结)
      - [5.2.1 调度器配置最佳实践](#521-调度器配置最佳实践)
      - [5.2.2 资源配额和限制](#522-资源配额和限制)
      - [5.2.3 调度器性能调优](#523-调度器性能调优)
    - [5.3 故障处理和恢复](#53-故障处理和恢复)
      - [5.3.1 调度器故障检测](#531-调度器故障检测)
      - [5.3.2 自动故障恢复](#532-自动故障恢复)
      - [5.3.3 监控和告警集成](#533-监控和告警集成)
  - [6. 高级调度特性](#6-高级调度特性)
    - [6.1 优先级和抢占](#61-优先级和抢占)
      - [6.1.1 优先级类配置](#611-优先级类配置)
      - [6.1.2 抢占机制实现](#612-抢占机制实现)
    - [6.2 资源配额和限制](#62-资源配额和限制)
      - [6.2.1 命名空间资源配额](#621-命名空间资源配额)
      - [6.2.2 多租户资源管理](#622-多租户资源管理)

---

## 1. 生产环境调度器配置

### 1.1 高可用调度器部署

在生产环境中，调度器的高可用性至关重要。以下是推荐的高可用部署配置：

#### 1.1.1 多实例部署

```yaml
# scheduler-ha-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-scheduler-ha
  namespace: kube-system
  labels:
    component: kube-scheduler
    tier: control-plane
spec:
  replicas: 3  # 推荐奇数个实例
  selector:
    matchLabels:
      component: kube-scheduler
      tier: control-plane
  template:
    metadata:
      labels:
        component: kube-scheduler
        tier: control-plane
    spec:
      priorityClassName: system-cluster-critical
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: kube-scheduler
        image: registry.k8s.io/kube-scheduler:v1.28.4
        command:
        - kube-scheduler
        - --config=/etc/kubernetes/scheduler-config.yaml
        - --authentication-kubeconfig=/etc/kubernetes/scheduler.conf
        - --authorization-kubeconfig=/etc/kubernetes/scheduler.conf
        - --bind-address=0.0.0.0
        - --leader-elect=true
        - --leader-elect-lease-duration=15s
        - --leader-elect-renew-deadline=10s
        - --leader-elect-retry-period=2s
        - --leader-elect-resource-lock=leases
        - --leader-elect-resource-name=kube-scheduler
        - --leader-elect-resource-namespace=kube-system
        - --profiling=false
        - --v=2
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 15
          timeoutSeconds: 15
          failureThreshold: 8
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 5
          timeoutSeconds: 5
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 2000m
            memory: 1Gi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
        volumeMounts:
        - name: config
          mountPath: /etc/kubernetes
          readOnly: true
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
        operator: Exists
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      nodeSelector:
        node-role.kubernetes.io/control-plane: ""
      volumes:
      - name: config
        configMap:
          name: scheduler-config
---
apiVersion: v1
kind: Service
metadata:
  name: kube-scheduler-metrics
  namespace: kube-system
  labels:
    component: kube-scheduler
spec:
  selector:
    component: kube-scheduler
  ports:
  - name: https-metrics
    port: 10259
    protocol: TCP
    targetPort: 10259
  type: ClusterIP
```

#### 1.1.2 生产级调度器配置

```yaml
# scheduler-config.yaml
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: /etc/kubernetes/scheduler.conf
  qps: 100
  burst: 200
leaderElection:
  leaderElect: true
  leaseDuration: 15s
  renewDeadline: 10s
  retryPeriod: 2s
  resourceLock: leases
  resourceName: kube-scheduler
  resourceNamespace: kube-system
profiles:
- schedulerName: default-scheduler
  plugins:
    # 预选插件配置
    filter:
      enabled:
      - name: NodeResourcesFit
      - name: NodeAffinity
      - name: PodTopologySpread
      - name: TaintToleration
      - name: VolumeRestrictions
      - name: VolumeBinding
      - name: VolumeZone
      - name: PodOverhead
      - name: NodePorts
      disabled:
      - name: NodeUnschedulable  # 在生产环境中可能需要禁用
    # 优选插件配置
    score:
      enabled:
      - name: NodeResourcesFit
        weight: 1
      - name: NodeAffinity
        weight: 2
      - name: PodTopologySpread
        weight: 2
      - name: InterPodAffinity
        weight: 2
      - name: NodePreferAvoidPods
        weight: 10000
      - name: TaintToleration
        weight: 1
      - name: ImageLocality
        weight: 1
    # 预留插件
    reserve:
      enabled:
      - name: VolumeBinding
    # 预绑定插件
    preBind:
      enabled:
      - name: VolumeBinding
  pluginConfig:
  # 资源适配插件配置
  - name: NodeResourcesFit
    args:
      scoringStrategy:
        type: LeastAllocated
        resources:
        - name: cpu
          weight: 1
        - name: memory
          weight: 1
        - name: nvidia.com/gpu
          weight: 5  # GPU 权重更高
  # Pod 拓扑分布配置
  - name: PodTopologySpread
    args:
      defaultConstraints:
      - maxSkew: 1
        topologyKey: topology.kubernetes.io/zone
        whenUnsatisfiable: ScheduleAnyway
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: ScheduleAnyway
      defaultingType: List
  # Pod 间亲和性配置
  - name: InterPodAffinity
    args:
      hardPodAffinityWeight: 100
  # 节点亲和性配置
  - name: NodeAffinity
    args:
      addedAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: kubernetes.io/arch
              operator: In
              values:
              - amd64
              - arm64
# 性能调优配置
percentageOfNodesToScore: 50  # 只对50%的节点进行评分
parallelism: 16  # 并行度
```

### 1.2 调度器性能调优

#### 1.2.1 调度延迟优化

```go
// scheduler-performance-tuning.go
package main

import (
    "context"
    "fmt"
    "time"
    
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/klog/v2"
)

type SchedulerPerformanceTuner struct {
    client kubernetes.Interface
    config *PerformanceConfig
}

type PerformanceConfig struct {
    // 调度队列配置
    QueueSortPlugin        string
    MaxPendingPods         int
    PodInitialBackoffSeconds int
    PodMaxBackoffSeconds   int
    
    // 节点评分配置
    PercentageOfNodesToScore int
    NodeScoreParallelism     int
    
    // 缓存配置
    NodeInfoCacheTTL       time.Duration
    PodInfoCacheTTL        time.Duration
    
    // 批处理配置
    BatchSize              int
    BatchTimeout           time.Duration
}

func NewSchedulerPerformanceTuner() *SchedulerPerformanceTuner {
    config, err := rest.InClusterConfig()
    if err != nil {
        klog.Fatalf("Failed to create config: %v", err)
    }
    
    client, err := kubernetes.NewForConfig(config)
    if err != nil {
        klog.Fatalf("Failed to create client: %v", err)
    }
    
    return &SchedulerPerformanceTuner{
        client: client,
        config: &PerformanceConfig{
            QueueSortPlugin:          "PrioritySort",
            MaxPendingPods:           5000,
            PodInitialBackoffSeconds: 1,
            PodMaxBackoffSeconds:     10,
            PercentageOfNodesToScore: 50,
            NodeScoreParallelism:     16,
            NodeInfoCacheTTL:         30 * time.Second,
            PodInfoCacheTTL:          30 * time.Second,
            BatchSize:                100,
            BatchTimeout:             100 * time.Millisecond,
        },
    }
}

func (spt *SchedulerPerformanceTuner) OptimizeForClusterSize(nodeCount int) {
    switch {
    case nodeCount < 100:
        // 小集群优化
        spt.config.PercentageOfNodesToScore = 100
        spt.config.NodeScoreParallelism = 4
        spt.config.BatchSize = 50
    case nodeCount < 1000:
        // 中等集群优化
        spt.config.PercentageOfNodesToScore = 50
        spt.config.NodeScoreParallelism = 8
        spt.config.BatchSize = 100
    default:
        // 大集群优化
        spt.config.PercentageOfNodesToScore = 30
        spt.config.NodeScoreParallelism = 16
        spt.config.BatchSize = 200
    }
    
    klog.Infof("Optimized scheduler for cluster size: %d nodes", nodeCount)
}

func (spt *SchedulerPerformanceTuner) GenerateOptimizedConfig() string {
    return fmt.Sprintf(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
profiles:
- schedulerName: optimized-scheduler
  plugins:
    queueSort:
      enabled:
      - name: %s
  pluginConfig:
  - name: DefaultBinder
    args:
      bindTimeoutSeconds: 600
  - name: PodTopologySpread
    args:
      defaultingType: List
percentageOfNodesToScore: %d
parallelism: %d
`,
        spt.config.QueueSortPlugin,
        spt.config.PercentageOfNodesToScore,
        spt.config.NodeScoreParallelism,
    )
}
```

#### 1.2.2 内存使用优化

```yaml
# scheduler-memory-optimization.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-memory-config
  namespace: kube-system
data:
  config.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    clientConnection:
      qps: 50
      burst: 100
    profiles:
    - schedulerName: memory-optimized-scheduler
      plugins:
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: NodeAffinity
          disabled:
          - name: VolumeRestrictions  # 减少内存使用
        score:
          enabled:
          - name: NodeResourcesFit
          - name: NodeAffinity
          disabled:
          - name: ImageLocality  # 减少内存使用
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: LeastAllocated
    # 内存优化配置
    percentageOfNodesToScore: 30  # 减少评分节点数量
    parallelism: 8  # 适中的并行度
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-optimized-scheduler
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: memory-optimized-scheduler
  template:
    metadata:
      labels:
        app: memory-optimized-scheduler
    spec:
      containers:
      - name: kube-scheduler
        image: registry.k8s.io/kube-scheduler:v1.28.4
        command:
        - kube-scheduler
        - --config=/etc/kubernetes/config.yaml
        - --v=2
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 500m
            memory: 256Mi  # 限制内存使用
        env:
        - name: GOGC
          value: "50"  # 更频繁的GC
        - name: GOMEMLIMIT
          value: "200MiB"  # Go内存限制
        volumeMounts:
        - name: config
          mountPath: /etc/kubernetes
      volumes:
      - name: config
        configMap:
          name: scheduler-memory-config
```

### 1.3 多调度器策略

#### 1.3.1 工作负载专用调度器

```yaml
# workload-specific-schedulers.yaml
# GPU 工作负载调度器
apiVersion: v1
kind: ConfigMap
metadata:
  name: gpu-scheduler-config
  namespace: kube-system
data:
  config.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    profiles:
    - schedulerName: gpu-scheduler
      plugins:
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: NodeAffinity
          - name: TaintToleration
        score:
          enabled:
          - name: NodeResourcesFit
            weight: 5
          - name: NodeAffinity
            weight: 3
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: MostAllocated  # GPU节点优先填满
            resources:
            - name: nvidia.com/gpu
              weight: 10
            - name: cpu
              weight: 1
            - name: memory
              weight: 1
---
# 批处理工作负载调度器
apiVersion: v1
kind: ConfigMap
metadata:
  name: batch-scheduler-config
  namespace: kube-system
data:
  config.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    profiles:
    - schedulerName: batch-scheduler
      plugins:
        queueSort:
          enabled:
          - name: PrioritySort
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: TaintToleration
        score:
          enabled:
          - name: NodeResourcesFit
            weight: 1
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: MostAllocated  # 提高资源利用率
    # 批处理优化
    percentageOfNodesToScore: 100  # 全节点评分
    parallelism: 32  # 高并行度
---
# 实时工作负载调度器
apiVersion: v1
kind: ConfigMap
metadata:
  name: realtime-scheduler-config
  namespace: kube-system
data:
  config.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    profiles:
    - schedulerName: realtime-scheduler
      plugins:
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: NodeAffinity
          - name: PodTopologySpread
        score:
          enabled:
          - name: NodeResourcesFit
            weight: 3
          - name: PodTopologySpread
            weight: 5
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: LeastAllocated  # 保证资源充足
      - name: PodTopologySpread
        args:
          defaultConstraints:
          - maxSkew: 1
            topologyKey: kubernetes.io/hostname
            whenUnsatisfiable: DoNotSchedule  # 严格分布要求
    # 实时优化
    percentageOfNodesToScore: 20  # 快速调度
    parallelism: 4  # 低延迟
```

#### 1.3.2 调度器选择策略

```go
// scheduler-selector.go
package main

import (
    "context"
    "fmt"
    "strings"
    
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

type SchedulerSelector struct {
    client kubernetes.Interface
    rules  []SchedulerRule
}

type SchedulerRule struct {
    Name        string
    Condition   func(*v1.Pod) bool
    Scheduler   string
    Priority    int
}

func NewSchedulerSelector(client kubernetes.Interface) *SchedulerSelector {
    return &SchedulerSelector{
        client: client,
        rules: []SchedulerRule{
            {
                Name: "GPU Workload",
                Condition: func(pod *v1.Pod) bool {
                    for _, container := range pod.Spec.Containers {
                        if _, hasGPU := container.Resources.Requests["nvidia.com/gpu"]; hasGPU {
                            return true
                        }
                    }
                    return false
                },
                Scheduler: "gpu-scheduler",
                Priority:  100,
            },
            {
                Name: "Batch Workload",
                Condition: func(pod *v1.Pod) bool {
                    if workloadType, exists := pod.Labels["workload.kubernetes.io/type"]; exists {
                        return workloadType == "batch"
                    }
                    return false
                },
                Scheduler: "batch-scheduler",
                Priority:  80,
            },
            {
                Name: "Realtime Workload",
                Condition: func(pod *v1.Pod) bool {
                    if priority, exists := pod.Labels["workload.kubernetes.io/priority"]; exists {
                        return priority == "realtime"
                    }
                    return false
                },
                Scheduler: "realtime-scheduler",
                Priority:  90,
            },
            {
                Name: "High Memory Workload",
                Condition: func(pod *v1.Pod) bool {
                    for _, container := range pod.Spec.Containers {
                        if memory := container.Resources.Requests.Memory(); memory != nil {
                            if memory.Value() > 8*1024*1024*1024 { // > 8Gi
                                return true
                            }
                        }
                    }
                    return false
                },
                Scheduler: "memory-optimized-scheduler",
                Priority:  70,
            },
        },
    }
}

func (ss *SchedulerSelector) SelectScheduler(pod *v1.Pod) string {
    // 如果已经指定了调度器，直接返回
    if pod.Spec.SchedulerName != "" {
        return pod.Spec.SchedulerName
    }
    
    var selectedScheduler string
    var highestPriority int
    
    for _, rule := range ss.rules {
        if rule.Condition(pod) && rule.Priority > highestPriority {
            selectedScheduler = rule.Scheduler
            highestPriority = rule.Priority
        }
    }
    
    if selectedScheduler == "" {
        return "default-scheduler"
    }
    
    return selectedScheduler
}

func (ss *SchedulerSelector) UpdatePodScheduler(ctx context.Context, pod *v1.Pod) error {
    selectedScheduler := ss.SelectScheduler(pod)
    
    if pod.Spec.SchedulerName == selectedScheduler {
        return nil // 无需更新
    }
    
    // 更新 Pod 的调度器
    pod.Spec.SchedulerName = selectedScheduler
    
    _, err := ss.client.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{})
    return err
}
```

## 2. 资源管理与优化

### 2.1 资源配额与限制

#### 2.1.1 动态资源配额管理

```go
// dynamic-resource-quota.go
package main

import (
    "context"
    "fmt"
    "time"
    
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type DynamicResourceQuotaManager struct {
    client    kubernetes.Interface
    quotas    map[string]*ResourceQuotaConfig
    metrics   *QuotaMetrics
}

type ResourceQuotaConfig struct {
    Namespace     string
    BaseQuota     v1.ResourceList
    MaxQuota      v1.ResourceList
    ScalingRules  []ScalingRule
    Priority      int
}

type ScalingRule struct {
    MetricType    string  // "cpu_usage", "memory_usage", "pod_count"
    Threshold     float64
    ScaleFactor   float64
    CooldownTime  time.Duration
}

type QuotaMetrics struct {
    UsageHistory  map[string][]ResourceUsage
    LastScaling   map[string]time.Time
}

type ResourceUsage struct {
    Timestamp time.Time
    CPU       float64
    Memory    float64
    Pods      int
}

func NewDynamicResourceQuotaManager(client kubernetes.Interface) *DynamicResourceQuotaManager {
    return &DynamicResourceQuotaManager{
        client: client,
        quotas: make(map[string]*ResourceQuotaConfig),
        metrics: &QuotaMetrics{
            UsageHistory: make(map[string][]ResourceUsage),
            LastScaling:  make(map[string]time.Time),
        },
    }
}

func (drqm *DynamicResourceQuotaManager) AddQuotaConfig(config *ResourceQuotaConfig) {
    drqm.quotas[config.Namespace] = config
    klog.Infof("Added quota config for namespace: %s", config.Namespace)
}

func (drqm *DynamicResourceQuotaManager) UpdateQuotas(ctx context.Context) error {
    for namespace, config := range drqm.quotas {
        usage, err := drqm.getCurrentUsage(ctx, namespace)
        if err != nil {
            klog.Errorf("Failed to get usage for namespace %s: %v", namespace, err)
            continue
        }
        
        // 记录使用历史
        drqm.recordUsage(namespace, usage)
        
        // 检查是否需要调整配额
        newQuota := drqm.calculateNewQuota(namespace, config, usage)
        if newQuota != nil {
            err = drqm.updateResourceQuota(ctx, namespace, newQuota)
            if err != nil {
                klog.Errorf("Failed to update quota for namespace %s: %v", namespace, err)
            } else {
                drqm.metrics.LastScaling[namespace] = time.Now()
                klog.Infof("Updated quota for namespace %s", namespace)
            }
        }
    }
    return nil
}

func (drqm *DynamicResourceQuotaManager) getCurrentUsage(ctx context.Context, namespace string) (*ResourceUsage, error) {
    pods, err := drqm.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    var totalCPU, totalMemory int64
    podCount := len(pods.Items)
    
    for _, pod := range pods.Items {
        for _, container := range pod.Spec.Containers {
            if cpu := container.Resources.Requests.Cpu(); cpu != nil {
                totalCPU += cpu.MilliValue()
            }
            if memory := container.Resources.Requests.Memory(); memory != nil {
                totalMemory += memory.Value()
            }
        }
    }
    
    return &ResourceUsage{
        Timestamp: time.Now(),
        CPU:       float64(totalCPU) / 1000, // 转换为核心数
        Memory:    float64(totalMemory) / (1024 * 1024 * 1024), // 转换为GB
        Pods:      podCount,
    }, nil
}

func (drqm *DynamicResourceQuotaManager) calculateNewQuota(namespace string, config *ResourceQuotaConfig, usage *ResourceUsage) v1.ResourceList {
    lastScaling, exists := drqm.metrics.LastScaling[namespace]
    if exists && time.Since(lastScaling) < 5*time.Minute {
        return nil // 冷却期内不调整
    }
    
    newQuota := config.BaseQuota.DeepCopy()
    needsUpdate := false
    
    for _, rule := range config.ScalingRules {
        var currentValue float64
        var resourceName v1.ResourceName
        
        switch rule.MetricType {
        case "cpu_usage":
            currentValue = usage.CPU
            resourceName = v1.ResourceRequestsCPU
        case "memory_usage":
            currentValue = usage.Memory
            resourceName = v1.ResourceRequestsMemory
        case "pod_count":
            currentValue = float64(usage.Pods)
            resourceName = v1.ResourcePods
        default:
            continue
        }
        
        if currentValue > rule.Threshold {
            currentQuota := newQuota[resourceName]
            var newValue resource.Quantity
            
            switch resourceName {
            case v1.ResourceRequestsCPU:
                newValue = resource.MustParse(fmt.Sprintf("%.1f", currentValue*rule.ScaleFactor))
            case v1.ResourceRequestsMemory:
                newValue = resource.MustParse(fmt.Sprintf("%.1fGi", currentValue*rule.ScaleFactor))
            case v1.ResourcePods:
                newValue = resource.MustParse(fmt.Sprintf("%.0f", currentValue*rule.ScaleFactor))
            }
            
            // 检查是否超过最大配额
            maxQuota := config.MaxQuota[resourceName]
            if newValue.Cmp(maxQuota) > 0 {
                newValue = maxQuota
            }
            
            if newValue.Cmp(currentQuota) != 0 {
                newQuota[resourceName] = newValue
                needsUpdate = true
            }
        }
    }
    
    if needsUpdate {
        return newQuota
    }
    return nil
}

func (drqm *DynamicResourceQuotaManager) updateResourceQuota(ctx context.Context, namespace string, quota v1.ResourceList) error {
    resourceQuota := &v1.ResourceQuota{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "dynamic-quota",
            Namespace: namespace,
        },
        Spec: v1.ResourceQuotaSpec{
            Hard: quota,
        },
    }
    
    _, err := drqm.client.CoreV1().ResourceQuotas(namespace).Update(ctx, resourceQuota, metav1.UpdateOptions{})
    if err != nil {
        // 如果不存在则创建
        _, err = drqm.client.CoreV1().ResourceQuotas(namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
    }
    
    return err
}

func (drqm *DynamicResourceQuotaManager) recordUsage(namespace string, usage *ResourceUsage) {
    history := drqm.metrics.UsageHistory[namespace]
    history = append(history, *usage)
    
    // 保留最近24小时的数据
    cutoff := time.Now().Add(-24 * time.Hour)
    var filtered []ResourceUsage
    for _, record := range history {
        if record.Timestamp.After(cutoff) {
            filtered = append(filtered, record)
        }
    }
    
    drqm.metrics.UsageHistory[namespace] = filtered
}
```

#### 2.1.2 优先级资源分配

```yaml
# priority-resource-allocation.yaml
# 高优先级工作负载的资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: high-priority-quota
  namespace: production
spec:
  hard:
    requests.cpu: "100"
    requests.memory: 200Gi
    limits.cpu: "200"
    limits.memory: 400Gi
    pods: "50"
  scopeSelector:
    matchExpressions:
    - operator: In
      scopeName: PriorityClass
      values: ["high-priority"]
---
# 中等优先级工作负载的资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: medium-priority-quota
  namespace: production
spec:
  hard:
    requests.cpu: "50"
    requests.memory: 100Gi
    limits.cpu: "100"
    limits.memory: 200Gi
    pods: "30"
  scopeSelector:
    matchExpressions:
    - operator: In
      scopeName: PriorityClass
      values: ["medium-priority"]
---
# 低优先级工作负载的资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: low-priority-quota
  namespace: production
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    pods: "20"
  scopeSelector:
    matchExpressions:
    - operator: In
      scopeName: PriorityClass
      values: ["low-priority"]
---
# 优先级类定义
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000
globalDefault: false
description: "High priority workloads"
---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: medium-priority
value: 500
globalDefault: true
description: "Medium priority workloads"
---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: low-priority
value: 100
globalDefault: false
description: "Low priority workloads"
preemptionPolicy: Never  # 不允许抢占
```

### 2.2 节点资源优化

#### 2.2.1 节点资源监控与调优

```go
// node-resource-optimizer.go
package main

import (
    "context"
    "fmt"
    "sort"
    "time"
    
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
    metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
    metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeResourceOptimizer struct {
    client        kubernetes.Interface
    metricsClient metricsclient.Interface
    optimizer     *ResourceOptimizer
}

type ResourceOptimizer struct {
    NodeMetrics     map[string]*NodeResourceMetrics
    OptimizationRules []OptimizationRule
}

type NodeResourceMetrics struct {
    NodeName        string
    CPUUsage        float64
    MemoryUsage     float64
    CPUCapacity     float64
    MemoryCapacity  float64
    PodCount        int
    PodCapacity     int
    LastUpdated     time.Time
    Trend           ResourceTrend
}

type ResourceTrend struct {
    CPUTrend    string  // "increasing", "decreasing", "stable"
    MemoryTrend string
    Confidence  float64
}

type OptimizationRule struct {
    Name        string
    Condition   func(*NodeResourceMetrics) bool
    Action      func(context.Context, *NodeResourceOptimizer, *NodeResourceMetrics) error
    Priority    int
}

func NewNodeResourceOptimizer(client kubernetes.Interface, metricsClient metricsclient.Interface) *NodeResourceOptimizer {
    return &NodeResourceOptimizer{
        client:        client,
        metricsClient: metricsClient,
        optimizer: &ResourceOptimizer{
            NodeMetrics: make(map[string]*NodeResourceMetrics),
            OptimizationRules: []OptimizationRule{
                {
                    Name: "High CPU Usage",
                    Condition: func(metrics *NodeResourceMetrics) bool {
                        return metrics.CPUUsage > 80.0
                    },
                    Action: handleHighCPUUsage,
                    Priority: 100,
                },
                {
                    Name: "High Memory Usage",
                    Condition: func(metrics *NodeResourceMetrics) bool {
                        return metrics.MemoryUsage > 85.0
                    },
                    Action: handleHighMemoryUsage,
                    Priority: 95,
                },
                {
                    Name: "Low Resource Utilization",
                    Condition: func(metrics *NodeResourceMetrics) bool {
                        return metrics.CPUUsage < 20.0 && metrics.MemoryUsage < 30.0 && metrics.PodCount < 5
                    },
                    Action: handleLowUtilization,
                    Priority: 50,
                },
                {
                    Name: "Resource Fragmentation",
                    Condition: func(metrics *NodeResourceMetrics) bool {
                        return metrics.PodCount > metrics.PodCapacity*0.8 && metrics.CPUUsage < 60.0
                    },
                    Action: handleResourceFragmentation,
                    Priority: 70,
                },
            },
        },
    }
}

func (nro *NodeResourceOptimizer) CollectMetrics(ctx context.Context) error {
    // 获取节点列表
    nodes, err := nro.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list nodes: %v", err)
    }
    
    // 获取节点指标
    nodeMetrics, err := nro.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to get node metrics: %v", err)
    }
    
    // 处理每个节点的指标
    for _, node := range nodes.Items {
        metrics := nro.calculateNodeMetrics(&node, nodeMetrics.Items)
        if metrics != nil {
            nro.optimizer.NodeMetrics[node.Name] = metrics
        }
    }
    
    return nil
}

func (nro *NodeResourceOptimizer) calculateNodeMetrics(node *v1.Node, nodeMetrics []metricsv1beta1.NodeMetrics) *NodeResourceMetrics {
    // 查找对应的指标
    var metrics *metricsv1beta1.NodeMetrics
    for _, nm := range nodeMetrics {
        if nm.Name == node.Name {
            metrics = &nm
            break
        }
    }
    
    if metrics == nil {
        return nil
    }
    
    // 计算容量
    cpuCapacity := float64(node.Status.Capacity.Cpu().MilliValue()) / 1000
    memoryCapacity := float64(node.Status.Capacity.Memory().Value()) / (1024 * 1024 * 1024)
    podCapacity := int(node.Status.Capacity.Pods().Value())
    
    // 计算使用量
    cpuUsage := float64(metrics.Usage.Cpu().MilliValue()) / 1000
    memoryUsage := float64(metrics.Usage.Memory().Value()) / (1024 * 1024 * 1024)
    
    // 获取Pod数量
    podCount := nro.getPodCount(node.Name)
    
    nodeMetrics := &NodeResourceMetrics{
        NodeName:       node.Name,
        CPUUsage:       (cpuUsage / cpuCapacity) * 100,
        MemoryUsage:    (memoryUsage / memoryCapacity) * 100,
        CPUCapacity:    cpuCapacity,
        MemoryCapacity: memoryCapacity,
        PodCount:       podCount,
        PodCapacity:    podCapacity,
        LastUpdated:    time.Now(),
    }
    
    // 计算趋势
    nodeMetrics.Trend = nro.calculateTrend(node.Name, nodeMetrics)
    
    return nodeMetrics
}

func (nro *NodeResourceOptimizer) getPodCount(nodeName string) int {
    pods, err := nro.client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
        FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
    })
    if err != nil {
        return 0
    }
    return len(pods.Items)
}

func (nro *NodeResourceOptimizer) calculateTrend(nodeName string, current *NodeResourceMetrics) ResourceTrend {
    // 简化的趋势计算，实际应该基于历史数据
    previous, exists := nro.optimizer.NodeMetrics[nodeName]
    if !exists {
        return ResourceTrend{
            CPUTrend:    "stable",
            MemoryTrend: "stable",
            Confidence:  0.5,
        }
    }
    
    cpuDiff := current.CPUUsage - previous.CPUUsage
    memoryDiff := current.MemoryUsage - previous.MemoryUsage
    
    var cpuTrend, memoryTrend string
    
    if cpuDiff > 5 {
        cpuTrend = "increasing"
    } else if cpuDiff < -5 {
        cpuTrend = "decreasing"
    } else {
        cpuTrend = "stable"
    }
    
    if memoryDiff > 5 {
        memoryTrend = "increasing"
    } else if memoryDiff < -5 {
        memoryTrend = "decreasing"
    } else {
        memoryTrend = "stable"
    }
    
    return ResourceTrend{
        CPUTrend:    cpuTrend,
        MemoryTrend: memoryTrend,
        Confidence:  0.8,
    }
}

func (nro *NodeResourceOptimizer) OptimizeResources(ctx context.Context) error {
    // 按优先级排序规则
    sort.Slice(nro.optimizer.OptimizationRules, func(i, j int) bool {
        return nro.optimizer.OptimizationRules[i].Priority > nro.optimizer.OptimizationRules[j].Priority
    })
    
    for _, metrics := range nro.optimizer.NodeMetrics {
        for _, rule := range nro.optimizer.OptimizationRules {
            if rule.Condition(metrics) {
                klog.Infof("Applying optimization rule '%s' to node %s", rule.Name, metrics.NodeName)
                err := rule.Action(ctx, nro, metrics)
                if err != nil {
                    klog.Errorf("Failed to apply rule '%s' to node %s: %v", rule.Name, metrics.NodeName, err)
                }
                break // 只应用第一个匹配的规则
            }
        }
    }
    
    return nil
}

// 优化规则实现
func handleHighCPUUsage(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
    klog.Warningf("Node %s has high CPU usage: %.2f%%", metrics.NodeName, metrics.CPUUsage)
    
    // 添加污点以防止新Pod调度到此节点
    return nro.addNodeTaint(ctx, metrics.NodeName, "high-cpu-usage", "NoSchedule")
}

func handleHighMemoryUsage(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
    klog.Warningf("Node %s has high memory usage: %.2f%%", metrics.NodeName, metrics.MemoryUsage)
    
    // 添加污点以防止新Pod调度到此节点
    return nro.addNodeTaint(ctx, metrics.NodeName, "high-memory-usage", "NoSchedule")
}

func handleLowUtilization(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
    klog.Infof("Node %s has low utilization - CPU: %.2f%%, Memory: %.2f%%, Pods: %d", 
        metrics.NodeName, metrics.CPUUsage, metrics.MemoryUsage, metrics.PodCount)
    
    // 添加标签标识低利用率节点
    return nro.addNodeLabel(ctx, metrics.NodeName, "node.kubernetes.io/utilization", "low")
}

func handleResourceFragmentation(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
    klog.Infof("Node %s has resource fragmentation - Pods: %d/%d, CPU: %.2f%%", 
        metrics.NodeName, metrics.PodCount, metrics.PodCapacity, metrics.CPUUsage)
    
    // 添加标签标识资源碎片化节点
    return nro.addNodeLabel(ctx, metrics.NodeName, "node.kubernetes.io/fragmentation", "high")
}

func (nro *NodeResourceOptimizer) addNodeTaint(ctx context.Context, nodeName, key, effect string) error {
    node, err := nro.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
    if err != nil {
        return err
    }
    
    // 检查污点是否已存在
    for _, taint := range node.Spec.Taints {
        if taint.Key == key {
            return nil // 污点已存在
        }
    }
    
    // 添加新污点
    newTaint := v1.Taint{
        Key:    key,
        Value:  "true",
        Effect: v1.TaintEffect(effect),
    }
    
    node.Spec.Taints = append(node.Spec.Taints, newTaint)
    
    _, err = nro.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
    return err
}

func (nro *NodeResourceOptimizer) addNodeLabel(ctx context.Context, nodeName, key, value string) error {
    node, err := nro.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
    if err != nil {
        return err
    }
    
    if node.Labels == nil {
        node.Labels = make(map[string]string)
    }
    
    node.Labels[key] = value
    
    _, err = nro.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
    return err
}
```

### 2.3 工作负载分类调度

#### 2.3.1 智能工作负载分类器

```go
// workload-classifier.go
package main

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"
    
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type WorkloadClassifier struct {
    client     kubernetes.Interface
    rules      []ClassificationRule
    profiles   map[string]*WorkloadProfile
}

type ClassificationRule struct {
    Name        string
    Priority    int
    Condition   func(*v1.Pod) bool
    WorkloadType string
    Scheduler   string
    Labels      map[string]string
    Annotations map[string]string
}

type WorkloadProfile struct {
    Type            string
    Description     string
    ResourcePattern ResourcePattern
    SchedulingHints SchedulingHints
    SLA             SLARequirements
}

type ResourcePattern struct {
    CPUIntensive    bool
    MemoryIntensive bool
    IOIntensive     bool
    NetworkIntensive bool
    GPURequired     bool
    StoragePattern  string // "high-iops", "high-throughput", "low-latency"
}

type SchedulingHints struct {
    PreferredScheduler string
    NodeAffinity       *v1.NodeAffinity
    PodAntiAffinity    *v1.PodAntiAffinity
    Tolerations        []v1.Toleration
    TopologySpread     []v1.TopologySpreadConstraint
}

type SLARequirements struct {
    MaxLatency      time.Duration
    Availability    float64 // 99.9%
    Throughput      int     // requests per second
    ResourceGuarantee bool
}

func NewWorkloadClassifier(client kubernetes.Interface) *WorkloadClassifier {
    wc := &WorkloadClassifier{
        client:   client,
        profiles: make(map[string]*WorkloadProfile),
    }
    
    wc.initializeProfiles()
    wc.initializeRules()
    
    return wc
}

func (wc *WorkloadClassifier) initializeProfiles() {
    wc.profiles["web-frontend"] = &WorkloadProfile{
        Type:        "web-frontend",
        Description: "Web frontend applications",
        ResourcePattern: ResourcePattern{
            CPUIntensive:     false,
            MemoryIntensive:  false,
            IOIntensive:      false,
            NetworkIntensive: true,
            GPURequired:      false,
            StoragePattern:   "low-latency",
        },
        SchedulingHints: SchedulingHints{
            PreferredScheduler: "realtime-scheduler",
            TopologySpread: []v1.TopologySpreadConstraint{
                {
                    MaxSkew:           1,
                    TopologyKey:       "kubernetes.io/hostname",
                    WhenUnsatisfiable: v1.DoNotSchedule,
                },
            },
        },
        SLA: SLARequirements{
            MaxLatency:        100 * time.Millisecond,
            Availability:      99.9,
            Throughput:        1000,
            ResourceGuarantee: true,
        },
    }
    
    wc.profiles["batch-processing"] = &WorkloadProfile{
        Type:        "batch-processing",
        Description: "Batch processing jobs",
        ResourcePattern: ResourcePattern{
            CPUIntensive:     true,
            MemoryIntensive:  true,
            IOIntensive:      false,
            NetworkIntensive: false,
            GPURequired:      false,
            StoragePattern:   "high-throughput",
        },
        SchedulingHints: SchedulingHints{
            PreferredScheduler: "batch-scheduler",
            Tolerations: []v1.Toleration{
                {
                    Key:      "node.kubernetes.io/batch",
                    Operator: v1.TolerationOpEqual,
                    Value:    "true",
                    Effect:   v1.TaintEffectNoSchedule,
                },
            },
        },
        SLA: SLARequirements{
            MaxLatency:        0, // 不关心延迟
            Availability:      95.0,
            Throughput:        0, // 不关心吞吐量
            ResourceGuarantee: false,
        },
    }
    
    wc.profiles["ml-training"] = &WorkloadProfile{
        Type:        "ml-training",
        Description: "Machine learning training jobs",
        ResourcePattern: ResourcePattern{
            CPUIntensive:     true,
            MemoryIntensive:  true,
            IOIntensive:      true,
            NetworkIntensive: false,
            GPURequired:      true,
            StoragePattern:   "high-iops",
        },
        SchedulingHints: SchedulingHints{
            PreferredScheduler: "gpu-scheduler",
            NodeAffinity: &v1.NodeAffinity{
                RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
                    NodeSelectorTerms: []v1.NodeSelectorTerm{
                        {
                            MatchExpressions: []v1.NodeSelectorRequirement{
                                {
                                    Key:      "accelerator",
                                    Operator: v1.NodeSelectorOpIn,
                                    Values:   []string{"nvidia-tesla-v100", "nvidia-a100"},
                                },
                            },
                        },
                    },
                },
            },
        },
        SLA: SLARequirements{
            MaxLatency:        0,
            Availability:      99.0,
            Throughput:        0,
            ResourceGuarantee: true,
        },
    }
    
    wc.profiles["database"] = &WorkloadProfile{
        Type:        "database",
        Description: "Database workloads",
        ResourcePattern: ResourcePattern{
            CPUIntensive:     false,
            MemoryIntensive:  true,
            IOIntensive:      true,
            NetworkIntensive: true,
            GPURequired:      false,
            StoragePattern:   "high-iops",
        },
        SchedulingHints: SchedulingHints{
            PreferredScheduler: "default-scheduler",
            PodAntiAffinity: &v1.PodAntiAffinity{
                RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
                    {
                        LabelSelector: &metav1.LabelSelector{
                            MatchLabels: map[string]string{
                                "app.kubernetes.io/component": "database",
                            },
                        },
                        TopologyKey: "kubernetes.io/hostname",
                    },
                },
            },
        },
        SLA: SLARequirements{
            MaxLatency:        50 * time.Millisecond,
            Availability:      99.99,
            Throughput:        5000,
            ResourceGuarantee: true,
        },
    }
}

func (wc *WorkloadClassifier) initializeRules() {
    wc.rules = []ClassificationRule{
        {
            Name:     "GPU Workload Detection",
            Priority: 100,
            Condition: func(pod *v1.Pod) bool {
                for _, container := range pod.Spec.Containers {
                    if _, hasGPU := container.Resources.Requests["nvidia.com/gpu"]; hasGPU {
                        return true
                    }
                }
                return false
            },
            WorkloadType: "ml-training",
            Scheduler:    "gpu-scheduler",
            Labels: map[string]string{
                "workload.kubernetes.io/type":     "ml-training",
                "workload.kubernetes.io/priority": "high",
            },
        },
        {
            Name:     "Web Frontend Detection",
            Priority: 90,
            Condition: func(pod *v1.Pod) bool {
                // 检查标签和注解
                if appType, exists := pod.Labels["app.kubernetes.io/component"]; exists {
                    return strings.Contains(strings.ToLower(appType), "frontend") ||
                           strings.Contains(strings.ToLower(appType), "web")
                }
                
                // 检查容器端口
                for _, container := range pod.Spec.Containers {
                    for _, port := range container.Ports {
                        if port.ContainerPort == 80 || port.ContainerPort == 443 || port.ContainerPort == 8080 {
                            return true
                        }
                    }
                }
                return false
            },
            WorkloadType: "web-frontend",
            Scheduler:    "realtime-scheduler",
            Labels: map[string]string{
                "workload.kubernetes.io/type":     "web-frontend",
                "workload.kubernetes.io/priority": "realtime",
            },
        },
        {
            Name:     "Batch Job Detection",
            Priority: 80,
            Condition: func(pod *v1.Pod) bool {
                // 检查Job或CronJob
                for _, owner := range pod.OwnerReferences {
                    if owner.Kind == "Job" || owner.Kind == "CronJob" {
                        return true
                    }
                }
                
                // 检查标签
                if jobType, exists := pod.Labels["app.kubernetes.io/component"]; exists {
                    return strings.Contains(strings.ToLower(jobType), "batch") ||
                           strings.Contains(strings.ToLower(jobType), "job")
                }
                return false
            },
            WorkloadType: "batch-processing",
            Scheduler:    "batch-scheduler",
            Labels: map[string]string{
                "workload.kubernetes.io/type":     "batch-processing",
                "workload.kubernetes.io/priority": "low",
            },
        },
        {
            Name:     "Database Detection",
            Priority: 85,
            Condition: func(pod *v1.Pod) bool {
                // 检查镜像名称
                imagePatterns := []string{
                    "mysql", "postgres", "mongodb", "redis", "elasticsearch",
                    "cassandra", "mariadb", "oracle", "mssql",
                }
                
                for _, container := range pod.Spec.Containers {
                    imageName := strings.ToLower(container.Image)
                    for _, pattern := range imagePatterns {
                        if strings.Contains(imageName, pattern) {
                            return true
                        }
                    }
                }
                
                // 检查标签
                if component, exists := pod.Labels["app.kubernetes.io/component"]; exists {
                    return strings.Contains(strings.ToLower(component), "database") ||
                           strings.Contains(strings.ToLower(component), "db")
                }
                return false
            },
            WorkloadType: "database",
            Scheduler:    "default-scheduler",
            Labels: map[string]string{
                "workload.kubernetes.io/type":     "database",
                "workload.kubernetes.io/priority": "high",
            },
        },
    }
}

func (wc *WorkloadClassifier) ClassifyPod(pod *v1.Pod) (*WorkloadProfile, error) {
    // 按优先级排序规则
    for _, rule := range wc.rules {
        if rule.Condition(pod) {
            profile, exists := wc.profiles[rule.WorkloadType]
            if !exists {
                return nil, fmt.Errorf("workload profile not found: %s", rule.WorkloadType)
            }
            
            klog.Infof("Classified pod %s/%s as %s using rule %s", 
                pod.Namespace, pod.Name, rule.WorkloadType, rule.Name)
            
            return profile, nil
        }
    }
    
    // 默认分类
    return wc.profiles["web-frontend"], nil
}

func (wc *WorkloadClassifier) ApplyClassification(ctx context.Context, pod *v1.Pod) error {
    profile, err := wc.ClassifyPod(pod)
    if err != nil {
        return err
    }
    
    // 应用调度器
    if profile.SchedulingHints.PreferredScheduler != "" {
        pod.Spec.SchedulerName = profile.SchedulingHints.PreferredScheduler
    }
    
    // 应用标签
    if pod.Labels == nil {
        pod.Labels = make(map[string]string)
    }
    pod.Labels["workload.kubernetes.io/type"] = profile.Type
    
    // 应用节点亲和性
    if profile.SchedulingHints.NodeAffinity != nil {
        if pod.Spec.Affinity == nil {
            pod.Spec.Affinity = &v1.Affinity{}
        }
        pod.Spec.Affinity.NodeAffinity = profile.SchedulingHints.NodeAffinity
    }
    
    // 应用Pod反亲和性
    if profile.SchedulingHints.PodAntiAffinity != nil {
        if pod.Spec.Affinity == nil {
            pod.Spec.Affinity = &v1.Affinity{}
        }
        pod.Spec.Affinity.PodAntiAffinity = profile.SchedulingHints.PodAntiAffinity
    }
    
    // 应用容忍度
    if len(profile.SchedulingHints.Tolerations) > 0 {
        pod.Spec.Tolerations = append(pod.Spec.Tolerations, profile.SchedulingHints.Tolerations...)
    }
    
    // 应用拓扑分布约束
    if len(profile.SchedulingHints.TopologySpread) > 0 {
        pod.Spec.TopologySpreadConstraints = append(pod.Spec.TopologySpreadConstraints, 
            profile.SchedulingHints.TopologySpread...)
    }
    
    return nil
}

func (wc *WorkloadClassifier) GetWorkloadStats(ctx context.Context) (map[string]int, error) {
    pods, err := wc.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    stats := make(map[string]int)
    
    for _, pod := range pods.Items {
        workloadType := pod.Labels["workload.kubernetes.io/type"]
        if workloadType == "" {
            workloadType = "unclassified"
        }
        stats[workloadType]++
    }
    
    return stats, nil
}
```

#### 2.3.2 工作负载调度策略配置

```yaml
# workload-scheduling-policies.yaml
# Web前端工作负载调度策略
apiVersion: v1
kind: ConfigMap
metadata:
  name: web-frontend-scheduling-policy
  namespace: kube-system
data:
  policy.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    profiles:
    - schedulerName: web-frontend-scheduler
      plugins:
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: NodeAffinity
          - name: PodTopologySpread
        score:
          enabled:
          - name: NodeResourcesFit
            weight: 3
          - name: PodTopologySpread
            weight: 5
          - name: NodeAffinity
            weight: 2
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: LeastAllocated  # 保证资源充足
      - name: PodTopologySpread
        args:
          defaultConstraints:
          - maxSkew: 1
            topologyKey: kubernetes.io/hostname
            whenUnsatisfiable: DoNotSchedule
    percentageOfNodesToScore: 30  # 快速调度
    parallelism: 8
---
# 批处理工作负载调度策略
apiVersion: v1
kind: ConfigMap
metadata:
  name: batch-processing-scheduling-policy
  namespace: kube-system
data:
  policy.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    profiles:
    - schedulerName: batch-processing-scheduler
      plugins:
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: TaintToleration
        score:
          enabled:
          - name: NodeResourcesFit
            weight: 1
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: MostAllocated  # 提高资源利用率
    percentageOfNodesToScore: 100  # 全节点评分
    parallelism: 16
---
# ML训练工作负载调度策略
apiVersion: v1
kind: ConfigMap
metadata:
  name: ml-training-scheduling-policy
  namespace: kube-system
data:
  policy.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta3
    kind: KubeSchedulerConfiguration
    profiles:
    - schedulerName: ml-training-scheduler
      plugins:
        filter:
          enabled:
          - name: NodeResourcesFit
          - name: NodeAffinity
        score:
          enabled:
          - name: NodeResourcesFit
            weight: 5
          - name: NodeAffinity
            weight: 3
      pluginConfig:
      - name: NodeResourcesFit
        args:
          scoringStrategy:
            type: MostAllocated
            resources:
            - name: nvidia.com/gpu
              weight: 10
            - name: cpu
              weight: 1
            - name: memory
              weight: 1
    percentageOfNodesToScore: 50
    parallelism: 4
```

## 3. 高级调度策略

### 3.1 自定义调度器插件

#### 3.1.1 延迟感知调度插件

```go
// latency-aware-plugin.go
package main

import (
    "context"
    "fmt"
    "strconv"
    "time"
    
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/klog/v2"
    "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
    LatencyAwarePluginName = "LatencyAware"
    MaxLatencyAnnotation   = "scheduler.kubernetes.io/max-latency"
    NodeLatencyLabel       = "node.kubernetes.io/network-latency"
)

type LatencyAwarePlugin struct {
    handle    framework.Handle
    latencyMap map[string]time.Duration
}

type LatencyAwareArgs struct {
    DefaultMaxLatency time.Duration `json:"defaultMaxLatency,omitempty"`
    LatencyWeight     int64         `json:"latencyWeight,omitempty"`
}

func NewLatencyAwarePlugin(args runtime.Object, handle framework.Handle) (framework.Plugin, error) {
    latencyArgs := &LatencyAwareArgs{
        DefaultMaxLatency: 100 * time.Millisecond,
        LatencyWeight:     10,
    }
    
    if args != nil {
        if err := runtime.DecodeInto(handle.Decoder(), args, latencyArgs); err != nil {
            return nil, err
        }
    }
    
    return &LatencyAwarePlugin{
        handle:     handle,
        latencyMap: make(map[string]time.Duration),
    }, nil
}

func (la *LatencyAwarePlugin) Name() string {
    return LatencyAwarePluginName
}

func (la *LatencyAwarePlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
    node := nodeInfo.Node()
    if node == nil {
        return framework.NewStatus(framework.Error, "node not found")
    }
    
    // 获取Pod的最大延迟要求
    maxLatencyStr, exists := pod.Annotations[MaxLatencyAnnotation]
    if !exists {
        return nil // 没有延迟要求，通过过滤
    }
    
    maxLatency, err := time.ParseDuration(maxLatencyStr)
    if err != nil {
        klog.Warningf("Invalid max latency annotation for pod %s/%s: %v", pod.Namespace, pod.Name, err)
        return nil
    }
    
    // 获取节点的网络延迟
    nodeLatency := la.getNodeLatency(node)
    
    if nodeLatency > maxLatency {
        return framework.NewStatus(framework.Unschedulable, 
            fmt.Sprintf("node latency %v exceeds max latency %v", nodeLatency, maxLatency))
    }
    
    return nil
}

func (la *LatencyAwarePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
    nodeInfo, err := la.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
    if err != nil {
        return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
    }
    
    node := nodeInfo.Node()
    if node == nil {
        return 0, framework.NewStatus(framework.Error, "node not found")
    }
    
    // 获取节点延迟
    nodeLatency := la.getNodeLatency(node)
    
    // 延迟越低，分数越高（最大分数100）
    // 假设最大可接受延迟为1秒
    maxAcceptableLatency := 1 * time.Second
    score := int64(100 * (1 - float64(nodeLatency)/float64(maxAcceptableLatency)))
    
    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    
    return score, nil
}

func (la *LatencyAwarePlugin) getNodeLatency(node *v1.Node) time.Duration {
    // 从缓存中获取
    if latency, exists := la.latencyMap[node.Name]; exists {
        return latency
    }
    
    // 从节点标签获取
    if latencyStr, exists := node.Labels[NodeLatencyLabel]; exists {
        if latency, err := time.ParseDuration(latencyStr); err == nil {
            la.latencyMap[node.Name] = latency
            return latency
        }
    }
    
    // 默认延迟
    defaultLatency := 50 * time.Millisecond
    la.latencyMap[node.Name] = defaultLatency
    return defaultLatency
}

func (la *LatencyAwarePlugin) ScoreExtensions() framework.ScoreExtensions {
    return nil
}
```

#### 3.1.2 成本优化调度插件

```go
// cost-optimization-plugin.go
package main

import (
    "context"
    "fmt"
    "strconv"
    
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/klog/v2"
    "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
    CostOptimizationPluginName = "CostOptimization"
    NodeCostLabel             = "node.kubernetes.io/cost-per-hour"
    PodCostBudgetAnnotation   = "scheduler.kubernetes.io/cost-budget"
)

type CostOptimizationPlugin struct {
    handle   framework.Handle
    costMap  map[string]float64
}

type CostOptimizationArgs struct {
    DefaultCostPerHour float64 `json:"defaultCostPerHour,omitempty"`
    CostWeight         int64   `json:"costWeight,omitempty"`
}

func NewCostOptimizationPlugin(args runtime.Object, handle framework.Handle) (framework.Plugin, error) {
    costArgs := &CostOptimizationArgs{
        DefaultCostPerHour: 0.1, // $0.1 per hour
        CostWeight:         5,
    }
    
    if args != nil {
        if err := runtime.DecodeInto(handle.Decoder(), args, costArgs); err != nil {
            return nil, err
        }
    }
    
    return &CostOptimizationPlugin{
        handle:  handle,
        costMap: make(map[string]float64),
    }, nil
}

func (co *CostOptimizationPlugin) Name() string {
    return CostOptimizationPluginName
}

func (co *CostOptimizationPlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
    node := nodeInfo.Node()
    if node == nil {
        return framework.NewStatus(framework.Error, "node not found")
    }
    
    // 检查Pod的成本预算
    budgetStr, exists := pod.Annotations[PodCostBudgetAnnotation]
    if !exists {
        return nil // 没有成本预算限制
    }
    
    budget, err := strconv.ParseFloat(budgetStr, 64)
    if err != nil {
        klog.Warningf("Invalid cost budget annotation for pod %s/%s: %v", pod.Namespace, pod.Name, err)
        return nil
    }
    
    // 获取节点成本
    nodeCost := co.getNodeCost(node)
    
    if nodeCost > budget {
        return framework.NewStatus(framework.Unschedulable, 
            fmt.Sprintf("node cost %.4f exceeds budget %.4f", nodeCost, budget))
    }
    
    return nil
}

func (co *CostOptimizationPlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
    nodeInfo, err := co.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
    if err != nil {
        return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
    }
    
    node := nodeInfo.Node()
    if node == nil {
        return 0, framework.NewStatus(framework.Error, "node not found")
    }
    
    // 获取节点成本
    nodeCost := co.getNodeCost(node)
    
    // 成本越低，分数越高
    // 假设最高成本为$1.0/hour
    maxCost := 1.0
    score := int64(100 * (1 - nodeCost/maxCost))
    
    if score < 0 {
        score = 0
    }
    if score > 100 {
        score = 100
    }
    
    return score, nil
}

func (co *CostOptimizationPlugin) getNodeCost(node *v1.Node) float64 {
    // 从缓存中获取
    if cost, exists := co.costMap[node.Name]; exists {
        return cost
    }
    
    // 从节点标签获取
    if costStr, exists := node.Labels[NodeCostLabel]; exists {
        if cost, err := strconv.ParseFloat(costStr, 64); err == nil {
            co.costMap[node.Name] = cost
            return cost
        }
    }
    
    // 根据节点类型估算成本
    defaultCost := co.estimateNodeCost(node)
    co.costMap[node.Name] = defaultCost
    return defaultCost
}

func (co *CostOptimizationPlugin) estimateNodeCost(node *v1.Node) float64 {
    // 基于节点资源估算成本
    cpu := float64(node.Status.Capacity.Cpu().MilliValue()) / 1000
    memory := float64(node.Status.Capacity.Memory().Value()) / (1024 * 1024 * 1024)
    
    // 简单的成本模型：$0.02/vCPU/hour + $0.01/GB/hour
    cost := cpu*0.02 + memory*0.01
    
    // 检查是否为Spot实例
    if instanceType, exists := node.Labels["node.kubernetes.io/instance-type"]; exists {
        if instanceType == "spot" {
            cost *= 0.3 // Spot实例成本降低70%
        }
    }
    
    return cost
}

func (co *CostOptimizationPlugin) ScoreExtensions() framework.ScoreExtensions {
    return nil
}
```

### 3.2 批处理调度优化

#### 3.2.1 批处理调度器实现

```go
// batch-scheduler.go
package main

import (
    "context"
    "fmt"
    "sort"
    "sync"
    "time"
    
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type BatchScheduler struct {
    client       kubernetes.Interface
    batchSize    int
    batchTimeout time.Duration
    pendingPods  []*v1.Pod
    mutex        sync.RWMutex
    stopCh       chan struct{}
    strategies   map[string]BatchStrategy
}

type BatchStrategy interface {
    Name() string
    SortPods(pods []*v1.Pod) []*v1.Pod
    SelectNodes(pods []*v1.Pod, nodes []*v1.Node) map[string]string // pod -> node
    Priority() int
}

type ResourceAwareBatchStrategy struct {
    name string
}

func (r *ResourceAwareBatchStrategy) Name() string {
    return r.name
}

func (r *ResourceAwareBatchStrategy) Priority() int {
    return 100
}

func (r *ResourceAwareBatchStrategy) SortPods(pods []*v1.Pod) []*v1.Pod {
    // 按资源需求排序：高资源需求的Pod优先调度
    sort.Slice(pods, func(i, j int) bool {
        resourceI := r.calculateResourceScore(pods[i])
        resourceJ := r.calculateResourceScore(pods[j])
        return resourceI > resourceJ
    })
    return pods
}

func (r *ResourceAwareBatchStrategy) calculateResourceScore(pod *v1.Pod) float64 {
    var totalCPU, totalMemory float64
    
    for _, container := range pod.Spec.Containers {
        if cpu := container.Resources.Requests.Cpu(); cpu != nil {
            totalCPU += float64(cpu.MilliValue())
        }
        if memory := container.Resources.Requests.Memory(); memory != nil {
            totalMemory += float64(memory.Value()) / (1024 * 1024 * 1024) // GB
        }
    }
    
    // 加权计算资源分数
    return totalCPU*0.3 + totalMemory*0.7
}

func (r *ResourceAwareBatchStrategy) SelectNodes(pods []*v1.Pod, nodes []*v1.Node) map[string]string {
    allocation := make(map[string]string)
    nodeResources := make(map[string]*NodeResourceInfo)
    
    // 初始化节点资源信息
    for _, node := range nodes {
        nodeResources[node.Name] = &NodeResourceInfo{
            Name:           node.Name,
            AvailableCPU:   float64(node.Status.Allocatable.Cpu().MilliValue()),
            AvailableMemory: float64(node.Status.Allocatable.Memory().Value()) / (1024 * 1024 * 1024),
            AllocatedPods:  0,
        }
    }
    
    // 为每个Pod选择最佳节点
    for _, pod := range pods {
        bestNode := r.findBestNode(pod, nodeResources)
        if bestNode != "" {
            allocation[pod.Name] = bestNode
            r.updateNodeResources(pod, nodeResources[bestNode])
        }
    }
    
    return allocation
}

type NodeResourceInfo struct {
    Name            string
    AvailableCPU    float64
    AvailableMemory float64
    AllocatedPods   int
}

func (r *ResourceAwareBatchStrategy) findBestNode(pod *v1.Pod, nodeResources map[string]*NodeResourceInfo) string {
    var bestNode string
    var bestScore float64
    
    podCPU, podMemory := r.getPodResourceRequests(pod)
    
    for nodeName, nodeInfo := range nodeResources {
        // 检查资源是否足够
        if nodeInfo.AvailableCPU < podCPU || nodeInfo.AvailableMemory < podMemory {
            continue
        }
        
        // 计算节点适合度分数（资源利用率 + 负载均衡）
        cpuUtilization := (nodeInfo.AvailableCPU - podCPU) / nodeInfo.AvailableCPU
        memoryUtilization := (nodeInfo.AvailableMemory - podMemory) / nodeInfo.AvailableMemory
        loadBalance := 1.0 / (float64(nodeInfo.AllocatedPods) + 1)
        
        score := (1-cpuUtilization)*0.4 + (1-memoryUtilization)*0.4 + loadBalance*0.2
        
        if bestNode == "" || score > bestScore {
            bestNode = nodeName
            bestScore = score
        }
    }
    
    return bestNode
}

func (r *ResourceAwareBatchStrategy) getPodResourceRequests(pod *v1.Pod) (float64, float64) {
    var totalCPU, totalMemory float64
    
    for _, container := range pod.Spec.Containers {
        if cpu := container.Resources.Requests.Cpu(); cpu != nil {
            totalCPU += float64(cpu.MilliValue())
        }
        if memory := container.Resources.Requests.Memory(); memory != nil {
            totalMemory += float64(memory.Value()) / (1024 * 1024 * 1024)
        }
    }
    
    return totalCPU, totalMemory
}

func (r *ResourceAwareBatchStrategy) updateNodeResources(pod *v1.Pod, nodeInfo *NodeResourceInfo) {
    podCPU, podMemory := r.getPodResourceRequests(pod)
    nodeInfo.AvailableCPU -= podCPU
    nodeInfo.AvailableMemory -= podMemory
    nodeInfo.AllocatedPods++
}

type PriorityBatchStrategy struct {
    name string
}

func (p *PriorityBatchStrategy) Name() string {
    return p.name
}

func (p *PriorityBatchStrategy) Priority() int {
    return 90
}

func (p *PriorityBatchStrategy) SortPods(pods []*v1.Pod) []*v1.Pod {
    // 按优先级排序
    sort.Slice(pods, func(i, j int) bool {
        priorityI := p.getPodPriority(pods[i])
        priorityJ := p.getPodPriority(pods[j])
        return priorityI > priorityJ
    })
    return pods
}

func (p *PriorityBatchStrategy) getPodPriority(pod *v1.Pod) int32 {
    if pod.Spec.Priority != nil {
        return *pod.Spec.Priority
    }
    return 0
}

func (p *PriorityBatchStrategy) SelectNodes(pods []*v1.Pod, nodes []*v1.Node) map[string]string {
    // 简单的轮询分配
    allocation := make(map[string]string)
    nodeIndex := 0
    
    for _, pod := range pods {
        if len(nodes) > 0 {
            allocation[pod.Name] = nodes[nodeIndex%len(nodes)].Name
            nodeIndex++
        }
    }
    
    return allocation
}

func NewBatchScheduler(client kubernetes.Interface, batchSize int, batchTimeout time.Duration) *BatchScheduler {
    bs := &BatchScheduler{
        client:       client,
        batchSize:    batchSize,
        batchTimeout: batchTimeout,
        pendingPods:  make([]*v1.Pod, 0),
        stopCh:       make(chan struct{}),
        strategies:   make(map[string]BatchStrategy),
    }
    
    // 注册策略
    bs.strategies["resource-aware"] = &ResourceAwareBatchStrategy{name: "resource-aware"}
    bs.strategies["priority"] = &PriorityBatchStrategy{name: "priority"}
    
    return bs
}

func (bs *BatchScheduler) Start(ctx context.Context) {
    ticker := time.NewTicker(bs.batchTimeout)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-bs.stopCh:
            return
        case <-ticker.C:
            bs.processBatch(ctx)
        }
    }
}

func (bs *BatchScheduler) AddPod(pod *v1.Pod) {
    bs.mutex.Lock()
    defer bs.mutex.Unlock()
    
    bs.pendingPods = append(bs.pendingPods, pod)
    
    // 如果达到批处理大小，立即处理
    if len(bs.pendingPods) >= bs.batchSize {
        go bs.processBatch(context.Background())
    }
}

func (bs *BatchScheduler) processBatch(ctx context.Context) {
    bs.mutex.Lock()
    if len(bs.pendingPods) == 0 {
        bs.mutex.Unlock()
        return
    }
    
    // 复制待处理的Pod列表
    podsToProcess := make([]*v1.Pod, len(bs.pendingPods))
    copy(podsToProcess, bs.pendingPods)
    bs.pendingPods = bs.pendingPods[:0] // 清空列表
    bs.mutex.Unlock()
    
    klog.Infof("Processing batch of %d pods", len(podsToProcess))
    
    // 获取可用节点
    nodes, err := bs.getAvailableNodes(ctx)
    if err != nil {
        klog.Errorf("Failed to get available nodes: %v", err)
        return
    }
    
    // 选择最佳策略
    strategy := bs.selectBestStrategy(podsToProcess)
    
    // 排序Pod
    sortedPods := strategy.SortPods(podsToProcess)
    
    // 分配节点
    allocation := strategy.SelectNodes(sortedPods, nodes)
    
    // 执行调度
    bs.executeBatchScheduling(ctx, allocation)
}

func (bs *BatchScheduler) getAvailableNodes(ctx context.Context) ([]*v1.Node, error) {
    nodeList, err := bs.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
        FieldSelector: "spec.unschedulable=false",
    })
    if err != nil {
        return nil, err
    }
    
    var availableNodes []*v1.Node
    for i := range nodeList.Items {
        node := &nodeList.Items[i]
        if bs.isNodeReady(node) {
            availableNodes = append(availableNodes, node)
        }
    }
    
    return availableNodes, nil
}

func (bs *BatchScheduler) isNodeReady(node *v1.Node) bool {
    for _, condition := range node.Status.Conditions {
        if condition.Type == v1.NodeReady {
            return condition.Status == v1.ConditionTrue
        }
    }
    return false
}

func (bs *BatchScheduler) selectBestStrategy(pods []*v1.Pod) BatchStrategy {
    // 根据Pod特征选择最佳策略
    hasHighPriorityPods := false
    hasResourceIntensivePods := false
    
    for _, pod := range pods {
        if pod.Spec.Priority != nil && *pod.Spec.Priority > 1000 {
            hasHighPriorityPods = true
        }
        
        for _, container := range pod.Spec.Containers {
            if cpu := container.Resources.Requests.Cpu(); cpu != nil && cpu.MilliValue() > 1000 {
                hasResourceIntensivePods = true
            }
        }
    }
    
    if hasHighPriorityPods {
        return bs.strategies["priority"]
    }
    
    if hasResourceIntensivePods {
        return bs.strategies["resource-aware"]
    }
    
    // 默认使用资源感知策略
    return bs.strategies["resource-aware"]
}

func (bs *BatchScheduler) executeBatchScheduling(ctx context.Context, allocation map[string]string) {
    var wg sync.WaitGroup
    
    for podName, nodeName := range allocation {
        wg.Add(1)
        go func(pName, nName string) {
            defer wg.Done()
            
            binding := &v1.Binding{
                ObjectMeta: metav1.ObjectMeta{
                    Name: pName,
                },
                Target: v1.ObjectReference{
                    Kind: "Node",
                    Name: nName,
                },
            }
            
            err := bs.client.CoreV1().Pods("").Bind(ctx, binding, metav1.CreateOptions{})
            if err != nil {
                klog.Errorf("Failed to bind pod %s to node %s: %v", pName, nName, err)
            } else {
                klog.Infof("Successfully bound pod %s to node %s", pName, nName)
            }
        }(podName, nodeName)
    }
    
    wg.Wait()
}

func (bs *BatchScheduler) Stop() {
    close(bs.stopCh)
}

func (bs *BatchScheduler) GetPendingPodsCount() int {
    bs.mutex.RLock()
    defer bs.mutex.RUnlock()
    return len(bs.pendingPods)
}
```

#### 3.2.2 批处理调度配置

```yaml
# batch-scheduler-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: batch-scheduler-config
  namespace: kube-system
data:
  config.yaml: |
    batchSize: 50
    batchTimeout: "30s"
    strategies:
      - name: "resource-aware"
        priority: 100
        enabled: true
        config:
          cpuWeight: 0.3
          memoryWeight: 0.7
          loadBalanceWeight: 0.2
      - name: "priority"
        priority: 90
        enabled: true
        config:
          priorityThreshold: 1000
    nodeSelection:
      filterRules:
        - type: "resource-availability"
          minCPU: "100m"
          minMemory: "128Mi"
        - type: "node-readiness"
          requiredConditions:
            - "Ready"
      scoringRules:
        - type: "resource-utilization"
          weight: 40
        - type: "load-balance"
          weight: 20
        - type: "node-affinity"
          weight: 40
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: batch-scheduler
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: batch-scheduler
  template:
    metadata:
      labels:
        app: batch-scheduler
    spec:
      serviceAccountName: batch-scheduler
      containers:
      - name: batch-scheduler
        image: k8s.gcr.io/batch-scheduler:v1.0.0
        command:
        - /usr/local/bin/batch-scheduler
        - --config=/etc/kubernetes/batch-scheduler-config.yaml
        - --v=2
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        volumeMounts:
        - name: config
          mountPath: /etc/kubernetes
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: batch-scheduler-config
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: batch-scheduler
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: batch-scheduler
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["pods/binding"]
  verbs: ["create"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: batch-scheduler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: batch-scheduler
subjects:
- kind: ServiceAccount
  name: batch-scheduler
  namespace: kube-system
```

### 3.3 边缘计算调度

#### 3.3.1 边缘节点调度器

```go
// edge-scheduler.go
package main

import (
    "context"
    "fmt"
    "math"
    "strconv"
    "strings"
    "time"
    
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

const (
    EdgeZoneLabel        = "node.kubernetes.io/edge-zone"
    EdgeLatencyLabel     = "node.kubernetes.io/edge-latency"
    EdgeBandwidthLabel   = "node.kubernetes.io/edge-bandwidth"
    EdgeReliabilityLabel = "node.kubernetes.io/edge-reliability"
    
    PodEdgeZoneAnnotation     = "scheduler.kubernetes.io/edge-zone"
    PodLatencyRequirement     = "scheduler.kubernetes.io/max-latency"
    PodBandwidthRequirement   = "scheduler.kubernetes.io/min-bandwidth"
    PodReliabilityRequirement = "scheduler.kubernetes.io/min-reliability"
)

type EdgeScheduler struct {
    client      kubernetes.Interface
    edgeZones   map[string]*EdgeZone
    nodeMetrics map[string]*EdgeNodeMetrics
}

type EdgeZone struct {
    Name         string
    Location     GeoLocation
    Nodes        []string
    Connectivity ConnectivityInfo
    Resources    EdgeResourceInfo
}

type GeoLocation struct {
    Latitude  float64
    Longitude float64
    Region    string
    Country   string
}

type ConnectivityInfo struct {
    Latency     time.Duration
    Bandwidth   int64 // Mbps
    Reliability float64 // 0-1
    Jitter      time.Duration
}

type EdgeResourceInfo struct {
    TotalCPU    int64
    TotalMemory int64
    TotalStorage int64
    UsedCPU     int64
    UsedMemory  int64
    UsedStorage int64
}

type EdgeNodeMetrics struct {
    NodeName        string
    EdgeZone        string
    Latency         time.Duration
    Bandwidth       int64
    Reliability     float64
    ResourceUsage   ResourceUsage
    NetworkQuality  NetworkQuality
    LastUpdated     time.Time
}

type ResourceUsage struct {
    CPUUsage    float64
    MemoryUsage float64
    StorageUsage float64
    NetworkUsage float64
}

type NetworkQuality struct {
    PacketLoss   float64
    Jitter       time.Duration
    Throughput   int64
    RTT          time.Duration
}

func NewEdgeScheduler(client kubernetes.Interface) *EdgeScheduler {
    es := &EdgeScheduler{
        client:      client,
        edgeZones:   make(map[string]*EdgeZone),
        nodeMetrics: make(map[string]*EdgeNodeMetrics),
    }
    
    es.initializeEdgeZones()
    return es
}

func (es *EdgeScheduler) initializeEdgeZones() {
    // 初始化边缘区域配置
    es.edgeZones["edge-zone-east"] = &EdgeZone{
        Name: "edge-zone-east",
        Location: GeoLocation{
            Latitude:  40.7128,
            Longitude: -74.0060,
            Region:    "us-east",
            Country:   "US",
        },
        Connectivity: ConnectivityInfo{
            Latency:     10 * time.Millisecond,
            Bandwidth:   1000, // 1Gbps
            Reliability: 0.99,
            Jitter:      2 * time.Millisecond,
        },
    }
    
    es.edgeZones["edge-zone-west"] = &EdgeZone{
        Name: "edge-zone-west",
        Location: GeoLocation{
            Latitude:  37.7749,
            Longitude: -122.4194,
            Region:    "us-west",
            Country:   "US",
        },
        Connectivity: ConnectivityInfo{
            Latency:     15 * time.Millisecond,
            Bandwidth:   500, // 500Mbps
            Reliability: 0.95,
            Jitter:      5 * time.Millisecond,
        },
    }
    
    es.edgeZones["edge-zone-europe"] = &EdgeZone{
        Name: "edge-zone-europe",
        Location: GeoLocation{
            Latitude:  51.5074,
            Longitude: -0.1278,
            Region:    "eu-west",
            Country:   "UK",
        },
        Connectivity: ConnectivityInfo{
            Latency:     25 * time.Millisecond,
            Bandwidth:   300, // 300Mbps
            Reliability: 0.92,
            Jitter:      8 * time.Millisecond,
        },
    }
}

func (es *EdgeScheduler) SchedulePod(ctx context.Context, pod *v1.Pod) (string, error) {
    // 获取Pod的边缘调度要求
    requirements := es.extractPodRequirements(pod)
    
    // 获取候选节点
    candidateNodes, err := es.getCandidateNodes(ctx, requirements)
    if err != nil {
        return "", err
    }
    
    if len(candidateNodes) == 0 {
        return "", fmt.Errorf("no suitable edge nodes found for pod %s/%s", pod.Namespace, pod.Name)
    }
    
    // 计算节点分数
    nodeScores := es.scoreNodes(candidateNodes, requirements)
    
    // 选择最佳节点
    bestNode := es.selectBestNode(nodeScores)
    
    klog.Infof("Selected edge node %s for pod %s/%s", bestNode, pod.Namespace, pod.Name)
    return bestNode, nil
}

type PodEdgeRequirements struct {
    PreferredZone    string
    MaxLatency       time.Duration
    MinBandwidth     int64
    MinReliability   float64
    ResourceRequests v1.ResourceList
    LocationHint     *GeoLocation
}

func (es *EdgeScheduler) extractPodRequirements(pod *v1.Pod) *PodEdgeRequirements {
    req := &PodEdgeRequirements{
        ResourceRequests: make(v1.ResourceList),
    }
    
    // 提取注解中的要求
    if zone, exists := pod.Annotations[PodEdgeZoneAnnotation]; exists {
        req.PreferredZone = zone
    }
    
    if latencyStr, exists := pod.Annotations[PodLatencyRequirement]; exists {
        if latency, err := time.ParseDuration(latencyStr); err == nil {
            req.MaxLatency = latency
        }
    }
    
    if bandwidthStr, exists := pod.Annotations[PodBandwidthRequirement]; exists {
        if bandwidth, err := strconv.ParseInt(bandwidthStr, 10, 64); err == nil {
            req.MinBandwidth = bandwidth
        }
    }
    
    if reliabilityStr, exists := pod.Annotations[PodReliabilityRequirement]; exists {
        if reliability, err := strconv.ParseFloat(reliabilityStr, 64); err == nil {
            req.MinReliability = reliability
        }
    }
    
    // 聚合资源请求
    for _, container := range pod.Spec.Containers {
        for resource, quantity := range container.Resources.Requests {
            if existing, exists := req.ResourceRequests[resource]; exists {
                existing.Add(quantity)
                req.ResourceRequests[resource] = existing
            } else {
                req.ResourceRequests[resource] = quantity
            }
        }
    }
    
    return req
}

func (es *EdgeScheduler) getCandidateNodes(ctx context.Context, req *PodEdgeRequirements) ([]*v1.Node, error) {
    nodeList, err := es.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
        LabelSelector: EdgeZoneLabel,
    })
    if err != nil {
        return nil, err
    }
    
    var candidates []*v1.Node
    
    for i := range nodeList.Items {
        node := &nodeList.Items[i]
        
        // 检查节点是否满足基本要求
        if es.nodeMatchesRequirements(node, req) {
            candidates = append(candidates, node)
        }
    }
    
    return candidates, nil
}

func (es *EdgeScheduler) nodeMatchesRequirements(node *v1.Node, req *PodEdgeRequirements) bool {
    // 检查边缘区域
    if req.PreferredZone != "" {
        if nodeZone, exists := node.Labels[EdgeZoneLabel]; !exists || nodeZone != req.PreferredZone {
            return false
        }
    }
    
    // 检查延迟要求
    if req.MaxLatency > 0 {
        if latencyStr, exists := node.Labels[EdgeLatencyLabel]; exists {
            if latency, err := time.ParseDuration(latencyStr); err == nil {
                if latency > req.MaxLatency {
                    return false
                }
            }
        }
    }
    
    // 检查带宽要求
    if req.MinBandwidth > 0 {
        if bandwidthStr, exists := node.Labels[EdgeBandwidthLabel]; exists {
            if bandwidth, err := strconv.ParseInt(bandwidthStr, 10, 64); err == nil {
                if bandwidth < req.MinBandwidth {
                    return false
                }
            }
        }
    }
    
    // 检查可靠性要求
    if req.MinReliability > 0 {
        if reliabilityStr, exists := node.Labels[EdgeReliabilityLabel]; exists {
            if reliability, err := strconv.ParseFloat(reliabilityStr, 64); err == nil {
                if reliability < req.MinReliability {
                    return false
                }
            }
        }
    }
    
    // 检查资源可用性
    return es.nodeHasEnoughResources(node, req.ResourceRequests)
}

func (es *EdgeScheduler) nodeHasEnoughResources(node *v1.Node, requests v1.ResourceList) bool {
    for resource, requested := range requests {
        if available, exists := node.Status.Allocatable[resource]; exists {
            if requested.Cmp(available) > 0 {
                return false
            }
        } else {
            return false
        }
    }
    return true
}

type NodeScore struct {
    NodeName string
    Score    float64
    Details  ScoreDetails
}

type ScoreDetails struct {
    LatencyScore     float64
    BandwidthScore   float64
    ReliabilityScore float64
    ResourceScore    float64
    LocationScore    float64
}

func (es *EdgeScheduler) scoreNodes(nodes []*v1.Node, req *PodEdgeRequirements) []NodeScore {
    var scores []NodeScore
    
    for _, node := range nodes {
        score := es.calculateNodeScore(node, req)
        scores = append(scores, score)
    }
    
    return scores
}

func (es *EdgeScheduler) calculateNodeScore(node *v1.Node, req *PodEdgeRequirements) NodeScore {
    details := ScoreDetails{}
    
    // 延迟分数 (权重: 30%)
    details.LatencyScore = es.calculateLatencyScore(node, req.MaxLatency)
    
    // 带宽分数 (权重: 25%)
    details.BandwidthScore = es.calculateBandwidthScore(node, req.MinBandwidth)
    
    // 可靠性分数 (权重: 20%)
    details.ReliabilityScore = es.calculateReliabilityScore(node, req.MinReliability)
    
    // 资源分数 (权重: 15%)
    details.ResourceScore = es.calculateResourceScore(node, req.ResourceRequests)
    
    // 位置分数 (权重: 10%)
    details.LocationScore = es.calculateLocationScore(node, req.LocationHint)
    
    // 加权总分
    totalScore := details.LatencyScore*0.3 +
                  details.BandwidthScore*0.25 +
                  details.ReliabilityScore*0.2 +
                  details.ResourceScore*0.15 +
                  details.LocationScore*0.1
    
    return NodeScore{
        NodeName: node.Name,
        Score:    totalScore,
        Details:  details,
    }
}

func (es *EdgeScheduler) calculateLatencyScore(node *v1.Node, maxLatency time.Duration) float64 {
    if maxLatency == 0 {
        return 100.0 // 没有延迟要求
    }
    
    latencyStr, exists := node.Labels[EdgeLatencyLabel]
    if !exists {
        return 50.0 // 默认分数
    }
    
    latency, err := time.ParseDuration(latencyStr)
    if err != nil {
        return 50.0
    }
    
    // 延迟越低分数越高
    if latency <= maxLatency {
        ratio := float64(latency) / float64(maxLatency)
        return 100.0 * (1.0 - ratio)
    }
    
    return 0.0 // 超过最大延迟要求
}

func (es *EdgeScheduler) calculateBandwidthScore(node *v1.Node, minBandwidth int64) float64 {
    if minBandwidth == 0 {
        return 100.0
    }
    
    bandwidthStr, exists := node.Labels[EdgeBandwidthLabel]
    if !exists {
        return 50.0
    }
    
    bandwidth, err := strconv.ParseInt(bandwidthStr, 10, 64)
    if err != nil {
        return 50.0
    }
    
    if bandwidth >= minBandwidth {
        // 带宽越高分数越高，但有上限
        ratio := float64(bandwidth) / float64(minBandwidth)
        return math.Min(100.0, 50.0+50.0*math.Log10(ratio))
    }
    
    return 0.0
}

func (es *EdgeScheduler) calculateReliabilityScore(node *v1.Node, minReliability float64) float64 {
    if minReliability == 0 {
        return 100.0
    }
    
    reliabilityStr, exists := node.Labels[EdgeReliabilityLabel]
    if !exists {
        return 50.0
    }
    
    reliability, err := strconv.ParseFloat(reliabilityStr, 64)
    if err != nil {
        return 50.0
    }
    
    if reliability >= minReliability {
        return 100.0 * reliability
    }
    
    return 0.0
}

func (es *EdgeScheduler) calculateResourceScore(node *v1.Node, requests v1.ResourceList) float64 {
    if len(requests) == 0 {
        return 100.0
    }
    
    var totalScore float64
    var resourceCount int
    
    for resource, requested := range requests {
        if allocatable, exists := node.Status.Allocatable[resource]; exists {
            ratio := float64(requested.MilliValue()) / float64(allocatable.MilliValue())
            // 资源利用率越低分数越高
            resourceScore := 100.0 * (1.0 - ratio)
            totalScore += math.Max(0, resourceScore)
            resourceCount++
        }
    }
    
    if resourceCount == 0 {
        return 50.0
    }
    
    return totalScore / float64(resourceCount)
}

func (es *EdgeScheduler) calculateLocationScore(node *v1.Node, locationHint *GeoLocation) float64 {
    if locationHint == nil {
        return 100.0
    }
    
    // 基于边缘区域计算位置分数
    zoneLabel, exists := node.Labels[EdgeZoneLabel]
    if !exists {
        return 50.0
    }
    
    zone, exists := es.edgeZones[zoneLabel]
    if !exists {
        return 50.0
    }
    
    // 计算地理距离
    distance := es.calculateDistance(locationHint, &zone.Location)
    
    // 距离越近分数越高
    maxDistance := 10000.0 // 10000km
    if distance <= maxDistance {
        return 100.0 * (1.0 - distance/maxDistance)
    }
    
    return 0.0
}

func (es *EdgeScheduler) calculateDistance(loc1, loc2 *GeoLocation) float64 {
    // 使用Haversine公式计算地理距离
    const earthRadius = 6371 // km
    
    lat1Rad := loc1.Latitude * math.Pi / 180
    lat2Rad := loc2.Latitude * math.Pi / 180
    deltaLatRad := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
    deltaLonRad := (loc2.Longitude - loc1.Longitude) * math.Pi / 180
    
    a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
        math.Cos(lat1Rad)*math.Cos(lat2Rad)*
        math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    
    return earthRadius * c
}

func (es *EdgeScheduler) selectBestNode(scores []NodeScore) string {
    if len(scores) == 0 {
        return ""
    }
    
    bestScore := scores[0]
    for _, score := range scores[1:] {
        if score.Score > bestScore.Score {
            bestScore = score
        }
    }
    
    return bestScore.NodeName
}

func (es *EdgeScheduler) UpdateNodeMetrics(nodeName string, metrics *EdgeNodeMetrics) {
    metrics.LastUpdated = time.Now()
    es.nodeMetrics[nodeName] = metrics
}

func (es *EdgeScheduler) GetEdgeZoneStatus(zoneName string) (*EdgeZone, error) {
    zone, exists := es.edgeZones[zoneName]
    if !exists {
        return nil, fmt.Errorf("edge zone %s not found", zoneName)
    }
    
    return zone, nil
}
```

#### 3.3.2 边缘调度配置

```yaml
# edge-scheduler-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: edge-scheduler-config
  namespace: kube-system
data:
  config.yaml: |
    edgeZones:
      - name: "edge-zone-east"
        location:
          latitude: 40.7128
          longitude: -74.0060
          region: "us-east"
          country: "US"
        connectivity:
          latency: "10ms"
          bandwidth: 1000
          reliability: 0.99
          jitter: "2ms"
        resources:
          totalCPU: 1000
          totalMemory: 4096
          totalStorage: 1024
      - name: "edge-zone-west"
        location:
          latitude: 37.7749
          longitude: -122.4194
          region: "us-west"
          country: "US"
        connectivity:
          latency: "15ms"
          bandwidth: 500
          reliability: 0.95
          jitter: "5ms"
        resources:
          totalCPU: 800
          totalMemory: 3072
          totalStorage: 512
    scoring:
      weights:
        latency: 30
        bandwidth: 25
        reliability: 20
        resource: 15
        location: 10
      thresholds:
        maxLatency: "100ms"
        minBandwidth: 10
        minReliability: 0.8
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: edge-scheduler
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: edge-scheduler
  template:
    metadata:
      labels:
        app: edge-scheduler
    spec:
      serviceAccountName: edge-scheduler
      containers:
      - name: edge-scheduler
        image: k8s.gcr.io/edge-scheduler:v1.0.0
        command:
        - /usr/local/bin/edge-scheduler
        - --config=/etc/kubernetes/edge-scheduler-config.yaml
        - --v=2
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 300m
            memory: 256Mi
        volumeMounts:
        - name: config
          mountPath: /etc/kubernetes
          readOnly: true
        env:
        - name: EDGE_ZONE_DISCOVERY
          value: "auto"
        - name: METRICS_UPDATE_INTERVAL
          value: "30s"
      volumes:
      - name: config
        configMap:
          name: edge-scheduler-config
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: edge-scheduler
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: edge-scheduler
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["pods/binding"]
  verbs: ["create"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["nodes", "pods"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: edge-scheduler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: edge-scheduler
subjects:
- kind: ServiceAccount
  name: edge-scheduler
  namespace: kube-system
```

#### 3.3.3 边缘节点标签配置

```yaml
# edge-node-labels.yaml
apiVersion: v1
kind: Node
metadata:
  name: edge-node-east-01
  labels:
    node.kubernetes.io/edge-zone: "edge-zone-east"
    node.kubernetes.io/edge-latency: "8ms"
    node.kubernetes.io/edge-bandwidth: "1000"
    node.kubernetes.io/edge-reliability: "0.99"
    node.kubernetes.io/instance-type: "edge"
    topology.kubernetes.io/region: "us-east-1"
    topology.kubernetes.io/zone: "us-east-1a"
---
apiVersion: v1
kind: Node
metadata:
  name: edge-node-west-01
  labels:
    node.kubernetes.io/edge-zone: "edge-zone-west"
    node.kubernetes.io/edge-latency: "12ms"
    node.kubernetes.io/edge-bandwidth: "500"
    node.kubernetes.io/edge-reliability: "0.95"
    node.kubernetes.io/instance-type: "edge"
    topology.kubernetes.io/region: "us-west-1"
    topology.kubernetes.io/zone: "us-west-1a"
```

#### 3.3.4 边缘工作负载示例

```yaml
# edge-workload-example.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iot-data-processor
  namespace: edge-apps
spec:
  replicas: 3
  selector:
    matchLabels:
      app: iot-data-processor
  template:
    metadata:
      labels:
        app: iot-data-processor
      annotations:
        scheduler.kubernetes.io/edge-zone: "edge-zone-east"
        scheduler.kubernetes.io/max-latency: "20ms"
        scheduler.kubernetes.io/min-bandwidth: "100"
        scheduler.kubernetes.io/min-reliability: "0.95"
    spec:
      schedulerName: edge-scheduler
      containers:
      - name: processor
        image: iot-processor:v1.0.0
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        env:
        - name: EDGE_ZONE
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['scheduler.kubernetes.io/edge-zone']
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node.kubernetes.io/instance-type
                operator: In
                values: ["edge"]
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            preference:
              matchExpressions:
              - key: node.kubernetes.io/edge-zone
                operator: In
                values: ["edge-zone-east"]
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: real-time-analytics
  namespace: edge-apps
spec:
  replicas: 2
  selector:
    matchLabels:
      app: real-time-analytics
  template:
    metadata:
      labels:
        app: real-time-analytics
      annotations:
        scheduler.kubernetes.io/max-latency: "10ms"
        scheduler.kubernetes.io/min-bandwidth: "500"
        scheduler.kubernetes.io/min-reliability: "0.98"
    spec:
      schedulerName: edge-scheduler
      containers:
      - name: analytics
        image: real-time-analytics:v2.0.0
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
      tolerations:
      - key: "edge-node"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
```

## 4. 监控与可观测性

### 4.1 调度器指标监控

#### 4.1.1 Prometheus 指标收集

```go
// scheduler-metrics.go
package main

import (
    "context"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    v1 "k8s.io/api/core/v1"
    "k8s.io/klog/v2"
)

type SchedulerMetrics struct {
    // 调度延迟指标
    schedulingLatency *prometheus.HistogramVec
    
    // 调度结果指标
    schedulingAttempts *prometheus.CounterVec
    schedulingFailures *prometheus.CounterVec
    
    // 队列指标
    pendingPods prometheus.Gauge
    queueWaitTime *prometheus.HistogramVec
    
    // 节点指标
    nodeUtilization *prometheus.GaugeVec
    nodeAvailability *prometheus.GaugeVec
    
    // 插件性能指标
    pluginExecutionTime *prometheus.HistogramVec
    pluginErrors *prometheus.CounterVec
    
    // 资源指标
    resourceFragmentation *prometheus.GaugeVec
    resourceWaste *prometheus.GaugeVec
    
    // 调度器健康指标
    schedulerHealth prometheus.Gauge
    schedulerUptime prometheus.Gauge
}

func NewSchedulerMetrics() *SchedulerMetrics {
    return &SchedulerMetrics{
        schedulingLatency: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "scheduler_scheduling_latency_seconds",
                Help: "Scheduling latency in seconds",
                Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
            },
            []string{"scheduler", "profile", "result"},
        ),
        
        schedulingAttempts: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "scheduler_scheduling_attempts_total",
                Help: "Total number of scheduling attempts",
            },
            []string{"scheduler", "profile", "result"},
        ),
        
        schedulingFailures: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "scheduler_scheduling_failures_total",
                Help: "Total number of scheduling failures",
            },
            []string{"scheduler", "profile", "reason"},
        ),
        
        pendingPods: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "scheduler_pending_pods",
                Help: "Number of pending pods in the scheduling queue",
            },
        ),
        
        queueWaitTime: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "scheduler_queue_wait_time_seconds",
                Help: "Time pods spend waiting in the scheduling queue",
                Buckets: prometheus.ExponentialBuckets(0.1, 2, 12), // 100ms to ~6min
            },
            []string{"scheduler", "priority_class"},
        ),
        
        nodeUtilization: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "scheduler_node_utilization_ratio",
                Help: "Node resource utilization ratio",
            },
            []string{"node", "resource"},
        ),
        
        nodeAvailability: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "scheduler_node_availability",
                Help: "Node availability status (1=available, 0=unavailable)",
            },
            []string{"node", "zone"},
        ),
        
        pluginExecutionTime: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "scheduler_plugin_execution_time_seconds",
                Help: "Plugin execution time in seconds",
                Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15), // 0.1ms to ~3s
            },
            []string{"plugin", "extension_point", "status"},
        ),
        
        pluginErrors: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "scheduler_plugin_errors_total",
                Help: "Total number of plugin errors",
            },
            []string{"plugin", "extension_point", "error_type"},
        ),
        
        resourceFragmentation: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "scheduler_resource_fragmentation_ratio",
                Help: "Resource fragmentation ratio",
            },
            []string{"resource", "zone"},
        ),
        
        resourceWaste: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "scheduler_resource_waste_ratio",
                Help: "Resource waste ratio",
            },
            []string{"resource", "reason"},
        ),
        
        schedulerHealth: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "scheduler_health_status",
                Help: "Scheduler health status (1=healthy, 0=unhealthy)",
            },
        ),
        
        schedulerUptime: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "scheduler_uptime_seconds",
                Help: "Scheduler uptime in seconds",
            },
        ),
    }
}

func (sm *SchedulerMetrics) RecordSchedulingLatency(scheduler, profile, result string, duration time.Duration) {
    sm.schedulingLatency.WithLabelValues(scheduler, profile, result).Observe(duration.Seconds())
}

func (sm *SchedulerMetrics) RecordSchedulingAttempt(scheduler, profile, result string) {
    sm.schedulingAttempts.WithLabelValues(scheduler, profile, result).Inc()
}

func (sm *SchedulerMetrics) RecordSchedulingFailure(scheduler, profile, reason string) {
    sm.schedulingFailures.WithLabelValues(scheduler, profile, reason).Inc()
}

func (sm *SchedulerMetrics) UpdatePendingPods(count float64) {
    sm.pendingPods.Set(count)
}

func (sm *SchedulerMetrics) RecordQueueWaitTime(scheduler, priorityClass string, duration time.Duration) {
    sm.queueWaitTime.WithLabelValues(scheduler, priorityClass).Observe(duration.Seconds())
}

func (sm *SchedulerMetrics) UpdateNodeUtilization(node, resource string, utilization float64) {
    sm.nodeUtilization.WithLabelValues(node, resource).Set(utilization)
}

func (sm *SchedulerMetrics) UpdateNodeAvailability(node, zone string, available bool) {
    value := 0.0
    if available {
        value = 1.0
    }
    sm.nodeAvailability.WithLabelValues(node, zone).Set(value)
}

func (sm *SchedulerMetrics) RecordPluginExecutionTime(plugin, extensionPoint, status string, duration time.Duration) {
    sm.pluginExecutionTime.WithLabelValues(plugin, extensionPoint, status).Observe(duration.Seconds())
}

func (sm *SchedulerMetrics) RecordPluginError(plugin, extensionPoint, errorType string) {
    sm.pluginErrors.WithLabelValues(plugin, extensionPoint, errorType).Inc()
}

func (sm *SchedulerMetrics) UpdateResourceFragmentation(resource, zone string, fragmentation float64) {
    sm.resourceFragmentation.WithLabelValues(resource, zone).Set(fragmentation)
}

func (sm *SchedulerMetrics) UpdateResourceWaste(resource, reason string, waste float64) {
    sm.resourceWaste.WithLabelValues(resource, reason).Set(waste)
}

func (sm *SchedulerMetrics) UpdateSchedulerHealth(healthy bool) {
    value := 0.0
    if healthy {
        value = 1.0
    }
    sm.schedulerHealth.Set(value)
}

func (sm *SchedulerMetrics) UpdateSchedulerUptime(uptime time.Duration) {
    sm.schedulerUptime.Set(uptime.Seconds())
}

// 调度器性能分析器
type SchedulerProfiler struct {
    metrics *SchedulerMetrics
    startTime time.Time
}

func NewSchedulerProfiler(metrics *SchedulerMetrics) *SchedulerProfiler {
    return &SchedulerProfiler{
        metrics: metrics,
        startTime: time.Now(),
    }
}

func (sp *SchedulerProfiler) ProfileSchedulingCycle(ctx context.Context, pod *v1.Pod, schedulerName string) func(result string, err error) {
    start := time.Now()
    
    return func(result string, err error) {
        duration := time.Since(start)
        profile := "default"
        
        if pod.Spec.SchedulerName != "" {
            profile = pod.Spec.SchedulerName
        }
        
        // 记录调度延迟
        sp.metrics.RecordSchedulingLatency(schedulerName, profile, result, duration)
        
        // 记录调度尝试
        sp.metrics.RecordSchedulingAttempt(schedulerName, profile, result)
        
        // 记录失败原因
        if err != nil {
            reason := "unknown"
            if err.Error() != "" {
                reason = err.Error()
            }
            sp.metrics.RecordSchedulingFailure(schedulerName, profile, reason)
        }
        
        klog.V(4).Infof("Scheduling cycle for pod %s/%s completed in %v with result %s", 
            pod.Namespace, pod.Name, duration, result)
    }
}

func (sp *SchedulerProfiler) ProfilePluginExecution(plugin, extensionPoint string) func(error) {
    start := time.Now()
    
    return func(err error) {
        duration := time.Since(start)
        status := "success"
        
        if err != nil {
            status = "error"
            errorType := "unknown"
            if err.Error() != "" {
                errorType = err.Error()
            }
            sp.metrics.RecordPluginError(plugin, extensionPoint, errorType)
        }
        
        sp.metrics.RecordPluginExecutionTime(plugin, extensionPoint, status, duration)
    }
}

func (sp *SchedulerProfiler) UpdateClusterMetrics(ctx context.Context, nodes []*v1.Node, pods []*v1.Pod) {
    // 更新节点可用性
    for _, node := range nodes {
        zone := node.Labels["topology.kubernetes.io/zone"]
        if zone == "" {
            zone = "unknown"
        }
        
        available := sp.isNodeAvailable(node)
        sp.metrics.UpdateNodeAvailability(node.Name, zone, available)
        
        // 计算节点资源利用率
        cpuUtil, memUtil := sp.calculateNodeUtilization(node, pods)
        sp.metrics.UpdateNodeUtilization(node.Name, "cpu", cpuUtil)
        sp.metrics.UpdateNodeUtilization(node.Name, "memory", memUtil)
    }
    
    // 更新资源碎片化指标
    sp.updateResourceFragmentation(nodes, pods)
    
    // 更新调度器健康状态
    uptime := time.Since(sp.startTime)
    sp.metrics.UpdateSchedulerUptime(uptime)
    sp.metrics.UpdateSchedulerHealth(true) // 简化实现
}

func (sp *SchedulerProfiler) isNodeAvailable(node *v1.Node) bool {
    if node.Spec.Unschedulable {
        return false
    }
    
    for _, condition := range node.Status.Conditions {
        if condition.Type == v1.NodeReady {
            return condition.Status == v1.ConditionTrue
        }
    }
    
    return false
}

func (sp *SchedulerProfiler) calculateNodeUtilization(node *v1.Node, pods []*v1.Pod) (float64, float64) {
    allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
    allocatableMemory := node.Status.Allocatable.Memory().Value()
    
    var usedCPU, usedMemory int64
    
    for _, pod := range pods {
        if pod.Spec.NodeName == node.Name && pod.Status.Phase == v1.PodRunning {
            for _, container := range pod.Spec.Containers {
                if cpu := container.Resources.Requests.Cpu(); cpu != nil {
                    usedCPU += cpu.MilliValue()
                }
                if memory := container.Resources.Requests.Memory(); memory != nil {
                    usedMemory += memory.Value()
                }
            }
        }
    }
    
    cpuUtil := float64(usedCPU) / float64(allocatableCPU)
    memUtil := float64(usedMemory) / float64(allocatableMemory)
    
    return cpuUtil, memUtil
}

func (sp *SchedulerProfiler) updateResourceFragmentation(nodes []*v1.Node, pods []*v1.Pod) {
    zoneFragmentation := make(map[string]map[string]float64)
    
    for _, node := range nodes {
        zone := node.Labels["topology.kubernetes.io/zone"]
        if zone == "" {
            zone = "unknown"
        }
        
        if zoneFragmentation[zone] == nil {
            zoneFragmentation[zone] = make(map[string]float64)
        }
        
        // 计算节点资源碎片化
        cpuFrag, memFrag := sp.calculateResourceFragmentation(node, pods)
        
        zoneFragmentation[zone]["cpu"] += cpuFrag
        zoneFragmentation[zone]["memory"] += memFrag
    }
    
    // 更新指标
    for zone, resources := range zoneFragmentation {
        for resource, fragmentation := range resources {
            sp.metrics.UpdateResourceFragmentation(resource, zone, fragmentation)
        }
    }
}

func (sp *SchedulerProfiler) calculateResourceFragmentation(node *v1.Node, pods []*v1.Pod) (float64, float64) {
    // 简化的碎片化计算：基于最大可调度Pod大小
    allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
    allocatableMemory := node.Status.Allocatable.Memory().Value()
    
    var usedCPU, usedMemory int64
    
    for _, pod := range pods {
        if pod.Spec.NodeName == node.Name {
            for _, container := range pod.Spec.Containers {
                if cpu := container.Resources.Requests.Cpu(); cpu != nil {
                    usedCPU += cpu.MilliValue()
                }
                if memory := container.Resources.Requests.Memory(); memory != nil {
                    usedMemory += memory.Value()
                }
            }
        }
    }
    
    remainingCPU := allocatableCPU - usedCPU
    remainingMemory := allocatableMemory - usedMemory
    
    // 碎片化 = 1 - (最大可用资源块 / 总可用资源)
    cpuFragmentation := 1.0 - (float64(remainingCPU) / float64(allocatableCPU))
    memFragmentation := 1.0 - (float64(remainingMemory) / float64(allocatableMemory))
    
    return cpuFragmentation, memFragmentation
}
```

#### 4.1.2 监控配置部署

```yaml
# scheduler-monitoring.yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: scheduler-metrics
  namespace: kube-system
  labels:
    app: kube-scheduler
spec:
  selector:
    matchLabels:
      component: kube-scheduler
  endpoints:
  - port: http-metrics
    interval: 30s
    path: /metrics
    scheme: http
  - port: http-metrics
    interval: 30s
    path: /metrics/resources
    scheme: http
---
apiVersion: v1
kind: Service
metadata:
  name: scheduler-metrics
  namespace: kube-system
  labels:
    component: kube-scheduler
spec:
  selector:
    component: kube-scheduler
  ports:
  - name: http-metrics
    port: 10259
    targetPort: 10259
    protocol: TCP
  clusterIP: None
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-dashboard
  namespace: monitoring
data:
  dashboard.json: |
    {
      "dashboard": {
        "id": null,
        "title": "Kubernetes Scheduler Metrics",
        "tags": ["kubernetes", "scheduler"],
        "style": "dark",
        "timezone": "browser",
        "panels": [
          {
            "id": 1,
            "title": "Scheduling Latency",
            "type": "graph",
            "targets": [
              {
                "expr": "histogram_quantile(0.99, sum(rate(scheduler_scheduling_latency_seconds_bucket[5m])) by (le, scheduler))",
                "legendFormat": "99th percentile - {{scheduler}}"
              },
              {
                "expr": "histogram_quantile(0.95, sum(rate(scheduler_scheduling_latency_seconds_bucket[5m])) by (le, scheduler))",
                "legendFormat": "95th percentile - {{scheduler}}"
              },
              {
                "expr": "histogram_quantile(0.50, sum(rate(scheduler_scheduling_latency_seconds_bucket[5m])) by (le, scheduler))",
                "legendFormat": "50th percentile - {{scheduler}}"
              }
            ],
            "yAxes": [
              {
                "label": "Latency (seconds)",
                "min": 0
              }
            ],
            "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
          },
          {
            "id": 2,
            "title": "Scheduling Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "sum(rate(scheduler_scheduling_attempts_total[5m])) by (scheduler, result)",
                "legendFormat": "{{scheduler}} - {{result}}"
              }
            ],
            "yAxes": [
              {
                "label": "Attempts per second",
                "min": 0
              }
            ],
            "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
          },
          {
            "id": 3,
            "title": "Pending Pods",
            "type": "singlestat",
            "targets": [
              {
                "expr": "scheduler_pending_pods",
                "legendFormat": "Pending Pods"
              }
            ],
            "gridPos": {"h": 4, "w": 6, "x": 0, "y": 8}
          },
          {
            "id": 4,
            "title": "Scheduler Health",
            "type": "singlestat",
            "targets": [
              {
                "expr": "scheduler_health_status",
                "legendFormat": "Health Status"
              }
            ],
            "gridPos": {"h": 4, "w": 6, "x": 6, "y": 8}
          },
          {
            "id": 5,
            "title": "Node Utilization",
            "type": "graph",
            "targets": [
              {
                "expr": "avg(scheduler_node_utilization_ratio) by (resource)",
                "legendFormat": "{{resource}}"
              }
            ],
            "yAxes": [
              {
                "label": "Utilization Ratio",
                "min": 0,
                "max": 1
              }
            ],
            "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
          },
          {
            "id": 6,
            "title": "Plugin Execution Time",
            "type": "graph",
            "targets": [
              {
                "expr": "histogram_quantile(0.95, sum(rate(scheduler_plugin_execution_time_seconds_bucket[5m])) by (le, plugin))",
                "legendFormat": "95th percentile - {{plugin}}"
              }
            ],
            "yAxes": [
              {
                "label": "Execution Time (seconds)",
                "min": 0
              }
            ],
            "gridPos": {"h": 8, "w": 24, "x": 0, "y": 16}
          }
        ],
        "time": {
          "from": "now-1h",
          "to": "now"
        },
        "refresh": "30s"
      }
    }
```

### 4.2 性能分析与诊断

#### 4.2.1 调度器性能分析工具

```go
// scheduler-analyzer.go
package main

import (
    "context"
    "fmt"
    "sort"
    "time"
    
    "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type SchedulerAnalyzer struct {
    client kubernetes.Interface
    metrics *SchedulerMetrics
}

type AnalysisReport struct {
    Timestamp time.Time `json:"timestamp"`
    
    // 性能指标
    PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
    
    // 资源分析
    ResourceAnalysis ResourceAnalysis `json:"resource_analysis"`
    
    // 调度问题
    SchedulingIssues []SchedulingIssue `json:"scheduling_issues"`
    
    // 优化建议
    Recommendations []Recommendation `json:"recommendations"`
    
    // 趋势分析
    TrendAnalysis TrendAnalysis `json:"trend_analysis"`
}

type PerformanceMetrics struct {
    AverageSchedulingLatency time.Duration `json:"average_scheduling_latency"`
    P95SchedulingLatency     time.Duration `json:"p95_scheduling_latency"`
    P99SchedulingLatency     time.Duration `json:"p99_scheduling_latency"`
    SchedulingThroughput     float64       `json:"scheduling_throughput"`
    FailureRate              float64       `json:"failure_rate"`
    QueueLength              int           `json:"queue_length"`
    PluginPerformance        map[string]time.Duration `json:"plugin_performance"`
}

type ResourceAnalysis struct {
    ClusterUtilization map[string]float64 `json:"cluster_utilization"`
    NodeUtilization    map[string]map[string]float64 `json:"node_utilization"`
    ResourceWaste      map[string]float64 `json:"resource_waste"`
    Fragmentation      map[string]float64 `json:"fragmentation"`
    HotSpots           []string `json:"hot_spots"`
    UnderutilizedNodes []string `json:"underutilized_nodes"`
}

type SchedulingIssue struct {
    Type        string    `json:"type"`
    Severity    string    `json:"severity"`
    Description string    `json:"description"`
    AffectedPods []string `json:"affected_pods"`
    Timestamp   time.Time `json:"timestamp"`
    Count       int       `json:"count"`
}

type Recommendation struct {
    Category    string `json:"category"`
    Priority    string `json:"priority"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Action      string `json:"action"`
    Impact      string `json:"impact"`
}

type TrendAnalysis struct {
    LatencyTrend      string  `json:"latency_trend"`
    ThroughputTrend   string  `json:"throughput_trend"`
    UtilizationTrend  string  `json:"utilization_trend"`
    PredictedIssues   []string `json:"predicted_issues"`
    CapacityForecast  map[string]float64 `json:"capacity_forecast"`
}

func NewSchedulerAnalyzer(client kubernetes.Interface, metrics *SchedulerMetrics) *SchedulerAnalyzer {
    return &SchedulerAnalyzer{
        client:  client,
        metrics: metrics,
    }
}

func (sa *SchedulerAnalyzer) GenerateAnalysisReport(ctx context.Context) (*AnalysisReport, error) {
    report := &AnalysisReport{
        Timestamp: time.Now(),
    }
    
    // 收集性能指标
    perfMetrics, err := sa.collectPerformanceMetrics(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to collect performance metrics: %v", err)
    }
    report.PerformanceMetrics = *perfMetrics
    
    // 分析资源使用情况
    resourceAnalysis, err := sa.analyzeResourceUsage(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze resource usage: %v", err)
    }
    report.ResourceAnalysis = *resourceAnalysis
    
    // 检测调度问题
    issues, err := sa.detectSchedulingIssues(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to detect scheduling issues: %v", err)
    }
    report.SchedulingIssues = issues
    
    // 生成优化建议
    recommendations := sa.generateRecommendations(report)
    report.Recommendations = recommendations
    
    // 趋势分析
    trendAnalysis, err := sa.analyzeTrends(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze trends: %v", err)
    }
    report.TrendAnalysis = *trendAnalysis
    
    return report, nil
}

func (sa *SchedulerAnalyzer) collectPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
    // 这里应该从Prometheus或其他监控系统获取实际指标
    // 为了演示，我们使用模拟数据
    
    return &PerformanceMetrics{
        AverageSchedulingLatency: 50 * time.Millisecond,
        P95SchedulingLatency:     200 * time.Millisecond,
        P99SchedulingLatency:     500 * time.Millisecond,
        SchedulingThroughput:     100.0, // pods per second
        FailureRate:              0.02,  // 2%
        QueueLength:              10,
        PluginPerformance: map[string]time.Duration{
            "NodeResourcesFit": 5 * time.Millisecond,
            "NodeAffinity":     3 * time.Millisecond,
            "PodTopologySpread": 8 * time.Millisecond,
            "TaintToleration":  2 * time.Millisecond,
        },
    }, nil
}

func (sa *SchedulerAnalyzer) analyzeResourceUsage(ctx context.Context) (*ResourceAnalysis, error) {
    nodes, err := sa.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    pods, err := sa.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    analysis := &ResourceAnalysis{
        ClusterUtilization: make(map[string]float64),
        NodeUtilization:    make(map[string]map[string]float64),
        ResourceWaste:      make(map[string]float64),
        Fragmentation:      make(map[string]float64),
        HotSpots:           []string{},
        UnderutilizedNodes: []string{},
    }
    
    var totalCPU, totalMemory, usedCPU, usedMemory int64
    
    for _, node := range nodes.Items {
        nodeCPU := node.Status.Allocatable.Cpu().MilliValue()
        nodeMemory := node.Status.Allocatable.Memory().Value()
        
        totalCPU += nodeCPU
        totalMemory += nodeMemory
        
        var nodeUsedCPU, nodeUsedMemory int64
        
        for _, pod := range pods.Items {
            if pod.Spec.NodeName == node.Name && pod.Status.Phase == v1.PodRunning {
                for _, container := range pod.Spec.Containers {
                    if cpu := container.Resources.Requests.Cpu(); cpu != nil {
                        nodeUsedCPU += cpu.MilliValue()
                    }
                    if memory := container.Resources.Requests.Memory(); memory != nil {
                        nodeUsedMemory += memory.Value()
                    }
                }
            }
        }
        
        usedCPU += nodeUsedCPU
        usedMemory += nodeUsedMemory
        
        // 节点利用率
        cpuUtil := float64(nodeUsedCPU) / float64(nodeCPU)
        memUtil := float64(nodeUsedMemory) / float64(nodeMemory)
        
        analysis.NodeUtilization[node.Name] = map[string]float64{
            "cpu":    cpuUtil,
            "memory": memUtil,
        }
        
        // 检测热点和低利用率节点
        if cpuUtil > 0.8 || memUtil > 0.8 {
            analysis.HotSpots = append(analysis.HotSpots, node.Name)
        }
        
        if cpuUtil < 0.2 && memUtil < 0.2 {
            analysis.UnderutilizedNodes = append(analysis.UnderutilizedNodes, node.Name)
        }
    }
    
    // 集群整体利用率
    analysis.ClusterUtilization["cpu"] = float64(usedCPU) / float64(totalCPU)
    analysis.ClusterUtilization["memory"] = float64(usedMemory) / float64(totalMemory)
    
    // 资源浪费分析
    analysis.ResourceWaste["cpu"] = sa.calculateResourceWaste("cpu", nodes.Items, pods.Items)
    analysis.ResourceWaste["memory"] = sa.calculateResourceWaste("memory", nodes.Items, pods.Items)
    
    // 碎片化分析
    analysis.Fragmentation["cpu"] = sa.calculateFragmentation("cpu", nodes.Items, pods.Items)
    analysis.Fragmentation["memory"] = sa.calculateFragmentation("memory", nodes.Items, pods.Items)
    
    return analysis, nil
}

func (sa *SchedulerAnalyzer) detectSchedulingIssues(ctx context.Context) ([]SchedulingIssue, error) {
    var issues []SchedulingIssue
    
    // 检测长时间等待的Pod
    pendingPods, err := sa.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
        FieldSelector: "status.phase=Pending",
    })
    if err != nil {
        return nil, err
    }
    
    longWaitingPods := []string{}
    for _, pod := range pendingPods.Items {
        if time.Since(pod.CreationTimestamp.Time) > 5*time.Minute {
            longWaitingPods = append(longWaitingPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
        }
    }
    
    if len(longWaitingPods) > 0 {
        issues = append(issues, SchedulingIssue{
            Type:         "LongWaitingPods",
            Severity:     "High",
            Description:  "Pods have been waiting for scheduling for more than 5 minutes",
            AffectedPods: longWaitingPods,
            Timestamp:    time.Now(),
            Count:        len(longWaitingPods),
        })
    }
    
    // 检测调度失败率过高
    // 这里应该从指标中获取实际的失败率
    failureRate := 0.05 // 5%
    if failureRate > 0.03 {
        issues = append(issues, SchedulingIssue{
            Type:        "HighFailureRate",
            Severity:    "Medium",
            Description: fmt.Sprintf("Scheduling failure rate is %.2f%%, which exceeds the threshold of 3%%", failureRate*100),
            Timestamp:   time.Now(),
            Count:       1,
        })
    }
    
    // 检测资源碎片化
    nodes, err := sa.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    pods, err := sa.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    fragmentation := sa.calculateFragmentation("cpu", nodes.Items, pods.Items)
    if fragmentation > 0.3 {
        issues = append(issues, SchedulingIssue{
            Type:        "ResourceFragmentation",
            Severity:    "Medium",
            Description: fmt.Sprintf("CPU resource fragmentation is %.2f%%, which may impact large pod scheduling", fragmentation*100),
            Timestamp:   time.Now(),
            Count:       1,
        })
    }
    
    return issues, nil
}

func (sa *SchedulerAnalyzer) generateRecommendations(report *AnalysisReport) []Recommendation {
    var recommendations []Recommendation
    
    // 基于性能指标的建议
    if report.PerformanceMetrics.P95SchedulingLatency > 100*time.Millisecond {
        recommendations = append(recommendations, Recommendation{
            Category:    "Performance",
            Priority:    "High",
            Title:       "Optimize Scheduling Latency",
            Description: "95th percentile scheduling latency is above 100ms",
            Action:      "Consider tuning scheduler configuration or reducing plugin complexity",
            Impact:      "Improved pod startup time and user experience",
        })
    }
    
    if report.PerformanceMetrics.FailureRate > 0.03 {
        recommendations = append(recommendations, Recommendation{
            Category:    "Reliability",
            Priority:    "High",
            Title:       "Reduce Scheduling Failures",
            Description: "Scheduling failure rate exceeds 3%",
            Action:      "Investigate resource constraints and node availability",
            Impact:      "Improved workload reliability and resource utilization",
        })
    }
    
    // 基于资源分析的建议
    if report.ResourceAnalysis.ClusterUtilization["cpu"] < 0.3 {
        recommendations = append(recommendations, Recommendation{
            Category:    "Resource Optimization",
            Priority:    "Medium",
            Title:       "Improve CPU Utilization",
            Description: "Cluster CPU utilization is below 30%",
            Action:      "Consider consolidating workloads or scaling down cluster",
            Impact:      "Reduced infrastructure costs",
        })
    }
    
    if len(report.ResourceAnalysis.HotSpots) > 0 {
        recommendations = append(recommendations, Recommendation{
            Category:    "Load Balancing",
            Priority:    "Medium",
            Title:       "Address Resource Hotspots",
            Description: fmt.Sprintf("%d nodes are experiencing high resource utilization", len(report.ResourceAnalysis.HotSpots)),
            Action:      "Implement pod anti-affinity or node affinity rules to distribute load",
            Impact:      "Better resource distribution and improved performance",
        })
    }
    
    if report.ResourceAnalysis.Fragmentation["cpu"] > 0.3 {
        recommendations = append(recommendations, Recommendation{
            Category:    "Resource Optimization",
            Priority:    "Medium",
            Title:       "Reduce Resource Fragmentation",
            Description: "High CPU fragmentation detected",
            Action:      "Consider using resource quotas and limits more effectively",
            Impact:      "Improved scheduling efficiency for large workloads",
        })
    }
    
    // 基于调度问题的建议
    for _, issue := range report.SchedulingIssues {
        switch issue.Type {
        case "LongWaitingPods":
            recommendations = append(recommendations, Recommendation{
                Category:    "Scheduling",
                Priority:    "High",
                Title:       "Resolve Long-Waiting Pods",
                Description: fmt.Sprintf("%d pods have been waiting for scheduling", issue.Count),
                Action:      "Check resource availability and scheduling constraints",
                Impact:      "Improved application availability",
            })
        }
    }
    
    // 排序建议（按优先级）
    sort.Slice(recommendations, func(i, j int) bool {
        priorityOrder := map[string]int{"High": 3, "Medium": 2, "Low": 1}
        return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
    })
    
    return recommendations
}

func (sa *SchedulerAnalyzer) analyzeTrends(ctx context.Context) (*TrendAnalysis, error) {
    // 这里应该分析历史数据来确定趋势
    // 为了演示，我们返回模拟的趋势分析
    
    return &TrendAnalysis{
        LatencyTrend:     "stable",
        ThroughputTrend:  "increasing",
        UtilizationTrend: "decreasing",
        PredictedIssues: []string{
            "Potential resource shortage in 2 weeks based on current growth rate",
            "Scheduling latency may increase due to cluster growth",
        },
        CapacityForecast: map[string]float64{
            "cpu_weeks_remaining":    8.5,
            "memory_weeks_remaining": 12.0,
            "nodes_needed_next_month": 3.0,
        },
    }, nil
}

func (sa *SchedulerAnalyzer) calculateResourceWaste(resourceType string, nodes []v1.Node, pods []v1.Pod) float64 {
    // 简化的资源浪费计算
    // 实际实现应该考虑请求vs限制、实际使用vs请求等
    
    var totalRequested, totalAllocated int64
    
    for _, node := range nodes {
        if resourceType == "cpu" {
            totalAllocated += node.Status.Allocatable.Cpu().MilliValue()
        } else {
            totalAllocated += node.Status.Allocatable.Memory().Value()
        }
    }
    
    for _, pod := range pods {
        if pod.Status.Phase == v1.PodRunning {
            for _, container := range pod.Spec.Containers {
                if resourceType == "cpu" {
                    if cpu := container.Resources.Requests.Cpu(); cpu != nil {
                        totalRequested += cpu.MilliValue()
                    }
                } else {
                    if memory := container.Resources.Requests.Memory(); memory != nil {
                        totalRequested += memory.Value()
                    }
                }
            }
        }
    }
    
    if totalAllocated == 0 {
        return 0
    }
    
    utilization := float64(totalRequested) / float64(totalAllocated)
    waste := 1.0 - utilization
    
    if waste < 0 {
        waste = 0
    }
    
    return waste
}

func (sa *SchedulerAnalyzer) calculateFragmentation(resourceType string, nodes []v1.Node, pods []v1.Pod) float64 {
    // 简化的碎片化计算
    // 实际实现应该考虑最大可调度Pod大小等因素
    
    nodeUtilizations := make([]float64, 0, len(nodes))
    
    for _, node := range nodes {
        var nodeAllocated, nodeUsed int64
        
        if resourceType == "cpu" {
            nodeAllocated = node.Status.Allocatable.Cpu().MilliValue()
        } else {
            nodeAllocated = node.Status.Allocatable.Memory().Value()
        }
        
        for _, pod := range pods {
            if pod.Spec.NodeName == node.Name && pod.Status.Phase == v1.PodRunning {
                for _, container := range pod.Spec.Containers {
                    if resourceType == "cpu" {
                        if cpu := container.Resources.Requests.Cpu(); cpu != nil {
                            nodeUsed += cpu.MilliValue()
                        }
                    } else {
                        if memory := container.Resources.Requests.Memory(); memory != nil {
                            nodeUsed += memory.Value()
                        }
                    }
                }
            }
        }
        
        if nodeAllocated > 0 {
            utilization := float64(nodeUsed) / float64(nodeAllocated)
            nodeUtilizations = append(nodeUtilizations, utilization)
        }
    }
    
    if len(nodeUtilizations) == 0 {
        return 0
    }
    
    // 计算利用率的标准差作为碎片化指标
    var sum, mean, variance float64
    
    for _, util := range nodeUtilizations {
        sum += util
    }
    mean = sum / float64(len(nodeUtilizations))
    
    for _, util := range nodeUtilizations {
        variance += (util - mean) * (util - mean)
    }
    variance /= float64(len(nodeUtilizations))
    
    return variance // 标准差的平方作为碎片化指标
}

// 定期分析任务
func (sa *SchedulerAnalyzer) StartPeriodicAnalysis(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            report, err := sa.GenerateAnalysisReport(ctx)
            if err != nil {
                klog.Errorf("Failed to generate analysis report: %v", err)
                continue
            }
            
            // 输出报告或发送到监控系统
            sa.processAnalysisReport(report)
        }
    }
}

func (sa *SchedulerAnalyzer) processAnalysisReport(report *AnalysisReport) {
    // 记录关键指标
    klog.Infof("Scheduler Analysis Report - Timestamp: %v", report.Timestamp)
    klog.Infof("Performance: Latency P95=%v, Throughput=%.2f pods/s, Failure Rate=%.2f%%",
        report.PerformanceMetrics.P95SchedulingLatency,
        report.PerformanceMetrics.SchedulingThroughput,
        report.PerformanceMetrics.FailureRate*100)
    
    klog.Infof("Resource Utilization: CPU=%.2f%%, Memory=%.2f%%",
        report.ResourceAnalysis.ClusterUtilization["cpu"]*100,
        report.ResourceAnalysis.ClusterUtilization["memory"]*100)
    
    // 输出高优先级建议
    for _, rec := range report.Recommendations {
        if rec.Priority == "High" {
            klog.Warningf("High Priority Recommendation: %s - %s", rec.Title, rec.Description)
        }
    }
    
    // 输出严重问题
    for _, issue := range report.SchedulingIssues {
        if issue.Severity == "High" {
            klog.Errorf("High Severity Issue: %s - %s (Count: %d)", issue.Type, issue.Description, issue.Count)
        }
    }
}
```

### 4.3 告警与自动化

#### 4.3.1 Prometheus 告警规则

```yaml
# scheduler-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: scheduler-alerts
  namespace: kube-system
  labels:
    app: kube-scheduler
spec:
  groups:
  - name: scheduler.performance
    interval: 30s
    rules:
    - alert: SchedulerHighLatency
      expr: histogram_quantile(0.95, sum(rate(scheduler_scheduling_latency_seconds_bucket[5m])) by (le)) > 0.1
      for: 2m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "Scheduler latency is high"
        description: "95th percentile scheduling latency is {{ $value }}s, which exceeds the threshold of 100ms"
        runbook_url: "https://runbooks.example.com/scheduler-high-latency"
    
    - alert: SchedulerVeryHighLatency
      expr: histogram_quantile(0.95, sum(rate(scheduler_scheduling_latency_seconds_bucket[5m])) by (le)) > 0.5
      for: 1m
      labels:
        severity: critical
        component: scheduler
      annotations:
        summary: "Scheduler latency is very high"
        description: "95th percentile scheduling latency is {{ $value }}s, which is critically high"
        runbook_url: "https://runbooks.example.com/scheduler-very-high-latency"
    
    - alert: SchedulerHighFailureRate
      expr: |
        (
          sum(rate(scheduler_scheduling_failures_total[5m]))
          /
          sum(rate(scheduler_scheduling_attempts_total[5m]))
        ) > 0.05
      for: 3m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "Scheduler failure rate is high"
        description: "Scheduler failure rate is {{ $value | humanizePercentage }}, which exceeds 5%"
        runbook_url: "https://runbooks.example.com/scheduler-high-failure-rate"
    
    - alert: SchedulerDown
      expr: up{job="kube-scheduler"} == 0
      for: 1m
      labels:
        severity: critical
        component: scheduler
      annotations:
        summary: "Scheduler is down"
        description: "Scheduler instance {{ $labels.instance }} is down"
        runbook_url: "https://runbooks.example.com/scheduler-down"
  
  - name: scheduler.queue
    interval: 30s
    rules:
    - alert: SchedulerQueueTooLarge
      expr: scheduler_pending_pods > 100
      for: 5m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "Scheduler queue is too large"
        description: "There are {{ $value }} pending pods in the scheduler queue"
        runbook_url: "https://runbooks.example.com/scheduler-queue-large"
    
    - alert: SchedulerQueueGrowthRate
      expr: increase(scheduler_pending_pods[10m]) > 50
      for: 2m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "Scheduler queue is growing rapidly"
        description: "Scheduler queue has grown by {{ $value }} pods in the last 10 minutes"
        runbook_url: "https://runbooks.example.com/scheduler-queue-growth"
  
  - name: scheduler.resources
    interval: 60s
    rules:
    - alert: NodeResourceFragmentation
      expr: avg(scheduler_resource_fragmentation_ratio) by (resource) > 0.3
      for: 10m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "High resource fragmentation detected"
        description: "{{ $labels.resource }} fragmentation is {{ $value | humanizePercentage }}"
        runbook_url: "https://runbooks.example.com/resource-fragmentation"
    
    - alert: ClusterResourceUtilizationLow
      expr: avg(scheduler_node_utilization_ratio) by (resource) < 0.2
      for: 30m
      labels:
        severity: info
        component: scheduler
      annotations:
        summary: "Low cluster resource utilization"
        description: "{{ $labels.resource }} utilization is {{ $value | humanizePercentage }}"
        runbook_url: "https://runbooks.example.com/low-utilization"
    
    - alert: ClusterResourceUtilizationHigh
      expr: avg(scheduler_node_utilization_ratio) by (resource) > 0.85
      for: 5m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "High cluster resource utilization"
        description: "{{ $labels.resource }} utilization is {{ $value | humanizePercentage }}"
        runbook_url: "https://runbooks.example.com/high-utilization"
  
  - name: scheduler.plugins
    interval: 30s
    rules:
    - alert: SchedulerPluginErrors
      expr: increase(scheduler_plugin_errors_total[5m]) > 10
      for: 2m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "Scheduler plugin errors detected"
        description: "Plugin {{ $labels.plugin }} has {{ $value }} errors in the last 5 minutes"
        runbook_url: "https://runbooks.example.com/plugin-errors"
    
    - alert: SchedulerPluginSlowExecution
      expr: histogram_quantile(0.95, sum(rate(scheduler_plugin_execution_time_seconds_bucket[5m])) by (le, plugin)) > 0.01
      for: 5m
      labels:
        severity: warning
        component: scheduler
      annotations:
        summary: "Scheduler plugin execution is slow"
        description: "Plugin {{ $labels.plugin }} 95th percentile execution time is {{ $value }}s"
        runbook_url: "https://runbooks.example.com/plugin-slow"
```

#### 4.3.2 自动化响应系统

```go
// scheduler-automation.go
package main

import (
    "context"
    "fmt"
    "time"
    
    "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type AutomationEngine struct {
    client   kubernetes.Interface
    analyzer *SchedulerAnalyzer
    config   *AutomationConfig
}

type AutomationConfig struct {
    EnableAutoRemediation bool          `yaml:"enable_auto_remediation"`
    MaxRemediationActions int           `yaml:"max_remediation_actions"`
    RemediationCooldown   time.Duration `yaml:"remediation_cooldown"`
    
    // 阈值配置
    Thresholds ThresholdConfig `yaml:"thresholds"`
    
    // 动作配置
    Actions ActionConfig `yaml:"actions"`
}

type ThresholdConfig struct {
    HighLatencyThreshold      time.Duration `yaml:"high_latency_threshold"`
    HighFailureRateThreshold  float64       `yaml:"high_failure_rate_threshold"`
    LargeQueueThreshold       int           `yaml:"large_queue_threshold"`
    HighFragmentationThreshold float64      `yaml:"high_fragmentation_threshold"`
    LowUtilizationThreshold   float64       `yaml:"low_utilization_threshold"`
    HighUtilizationThreshold  float64       `yaml:"high_utilization_threshold"`
}

type ActionConfig struct {
    EnableNodeDraining     bool `yaml:"enable_node_draining"`
    EnablePodEviction      bool `yaml:"enable_pod_eviction"`
    EnableSchedulerRestart bool `yaml:"enable_scheduler_restart"`
    EnableNodeLabeling     bool `yaml:"enable_node_labeling"`
    EnableAlertEscalation  bool `yaml:"enable_alert_escalation"`
}

type RemediationAction struct {
    Type        string                 `json:"type"`
    Target      string                 `json:"target"`
    Parameters  map[string]interface{} `json:"parameters"`
    Timestamp   time.Time              `json:"timestamp"`
    Status      string                 `json:"status"`
    Result      string                 `json:"result"`
    Duration    time.Duration          `json:"duration"`
}

type RemediationHistory struct {
    Actions []RemediationAction `json:"actions"`
    mutex   sync.RWMutex
}

func NewAutomationEngine(client kubernetes.Interface, analyzer *SchedulerAnalyzer, config *AutomationConfig) *AutomationEngine {
    return &AutomationEngine{
        client:   client,
        analyzer: analyzer,
        config:   config,
    }
}

func (ae *AutomationEngine) StartAutomationLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    history := &RemediationHistory{
        Actions: make([]RemediationAction, 0),
    }
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := ae.processAutomationCycle(ctx, history); err != nil {
                klog.Errorf("Automation cycle failed: %v", err)
            }
        }
    }
}

func (ae *AutomationEngine) processAutomationCycle(ctx context.Context, history *RemediationHistory) error {
    // 生成分析报告
    report, err := ae.analyzer.GenerateAnalysisReport(ctx)
    if err != nil {
        return fmt.Errorf("failed to generate analysis report: %v", err)
    }
    
    // 检查是否需要自动修复
    if !ae.config.EnableAutoRemediation {
        return nil
    }
    
    // 检查冷却期
    if ae.isInCooldownPeriod(history) {
        klog.V(4).Info("Automation is in cooldown period, skipping remediation")
        return nil
    }
    
    // 检查修复动作限制
    if ae.hasExceededActionLimit(history) {
        klog.Warning("Maximum remediation actions reached, skipping further actions")
        return nil
    }
    
    // 执行自动修复
    actions := ae.determineRemediationActions(report)
    for _, action := range actions {
        if err := ae.executeRemediationAction(ctx, action, history); err != nil {
            klog.Errorf("Failed to execute remediation action %s: %v", action.Type, err)
        }
    }
    
    return nil
}

func (ae *AutomationEngine) determineRemediationActions(report *AnalysisReport) []RemediationAction {
    var actions []RemediationAction
    
    // 基于性能指标的自动修复
    if report.PerformanceMetrics.P95SchedulingLatency > ae.config.Thresholds.HighLatencyThreshold {
        actions = append(actions, RemediationAction{
            Type:       "optimize_scheduler_config",
            Target:     "scheduler",
            Parameters: map[string]interface{}{
                "reason": "high_latency",
                "latency": report.PerformanceMetrics.P95SchedulingLatency.String(),
            },
            Timestamp: time.Now(),
            Status:    "pending",
        })
    }
    
    if report.PerformanceMetrics.FailureRate > ae.config.Thresholds.HighFailureRateThreshold {
        actions = append(actions, RemediationAction{
            Type:   "investigate_failures",
            Target: "scheduler",
            Parameters: map[string]interface{}{
                "reason":       "high_failure_rate",
                "failure_rate": report.PerformanceMetrics.FailureRate,
            },
            Timestamp: time.Now(),
            Status:    "pending",
        })
    }
    
    if report.PerformanceMetrics.QueueLength > ae.config.Thresholds.LargeQueueThreshold {
        actions = append(actions, RemediationAction{
            Type:   "drain_problematic_nodes",
            Target: "cluster",
            Parameters: map[string]interface{}{
                "reason":      "large_queue",
                "queue_length": report.PerformanceMetrics.QueueLength,
            },
            Timestamp: time.Now(),
            Status:    "pending",
        })
    }
    
    // 基于资源分析的自动修复
    if report.ResourceAnalysis.Fragmentation["cpu"] > ae.config.Thresholds.HighFragmentationThreshold {
        actions = append(actions, RemediationAction{
            Type:   "defragment_resources",
            Target: "cluster",
            Parameters: map[string]interface{}{
                "reason":        "high_fragmentation",
                "fragmentation": report.ResourceAnalysis.Fragmentation["cpu"],
                "resource":      "cpu",
            },
            Timestamp: time.Now(),
            Status:    "pending",
        })
    }
    
    // 处理热点节点
    if len(report.ResourceAnalysis.HotSpots) > 0 {
        for _, hotspot := range report.ResourceAnalysis.HotSpots {
            actions = append(actions, RemediationAction{
                Type:   "rebalance_node",
                Target: hotspot,
                Parameters: map[string]interface{}{
                    "reason": "resource_hotspot",
                    "node":   hotspot,
                },
                Timestamp: time.Now(),
                Status:    "pending",
            })
        }
    }
    
    // 处理低利用率节点
    if len(report.ResourceAnalysis.UnderutilizedNodes) > 2 {
        actions = append(actions, RemediationAction{
            Type:   "consolidate_workloads",
            Target: "cluster",
            Parameters: map[string]interface{}{
                "reason":               "low_utilization",
                "underutilized_nodes": report.ResourceAnalysis.UnderutilizedNodes,
            },
            Timestamp: time.Now(),
            Status:    "pending",
        })
    }
    
    return actions
}

func (ae *AutomationEngine) executeRemediationAction(ctx context.Context, action RemediationAction, history *RemediationHistory) error {
    start := time.Now()
    action.Status = "executing"
    
    defer func() {
        action.Duration = time.Since(start)
        history.mutex.Lock()
        history.Actions = append(history.Actions, action)
        history.mutex.Unlock()
    }()
    
    switch action.Type {
    case "optimize_scheduler_config":
        return ae.optimizeSchedulerConfig(ctx, action)
    case "investigate_failures":
        return ae.investigateFailures(ctx, action)
    case "drain_problematic_nodes":
        return ae.drainProblematicNodes(ctx, action)
    case "defragment_resources":
        return ae.defragmentResources(ctx, action)
    case "rebalance_node":
        return ae.rebalanceNode(ctx, action)
    case "consolidate_workloads":
        return ae.consolidateWorkloads(ctx, action)
    default:
        action.Status = "failed"
        action.Result = fmt.Sprintf("Unknown action type: %s", action.Type)
        return fmt.Errorf("unknown action type: %s", action.Type)
    }
}

func (ae *AutomationEngine) optimizeSchedulerConfig(ctx context.Context, action RemediationAction) error {
    if !ae.config.Actions.EnableSchedulerRestart {
        action.Status = "skipped"
        action.Result = "Scheduler restart is disabled"
        return nil
    }
    
    klog.Infof("Optimizing scheduler configuration due to %s", action.Parameters["reason"])
    
    // 这里可以实现调度器配置优化逻辑
    // 例如：调整 percentageOfNodesToScore、修改插件配置等
    
    action.Status = "completed"
    action.Result = "Scheduler configuration optimized"
    return nil
}

func (ae *AutomationEngine) investigateFailures(ctx context.Context, action RemediationAction) error {
    klog.Infof("Investigating scheduling failures, failure rate: %v", action.Parameters["failure_rate"])
    
    // 收集失败的Pod信息
    pendingPods, err := ae.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
        FieldSelector: "status.phase=Pending",
    })
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to list pending pods: %v", err)
        return err
    }
    
    failureReasons := make(map[string]int)
    for _, pod := range pendingPods.Items {
        for _, condition := range pod.Status.Conditions {
            if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionFalse {
                failureReasons[condition.Reason]++
            }
        }
    }
    
    klog.Infof("Scheduling failure analysis: %+v", failureReasons)
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Analyzed %d pending pods, failure reasons: %+v", len(pendingPods.Items), failureReasons)
    return nil
}

func (ae *AutomationEngine) drainProblematicNodes(ctx context.Context, action RemediationAction) error {
    if !ae.config.Actions.EnableNodeDraining {
        action.Status = "skipped"
        action.Result = "Node draining is disabled"
        return nil
    }
    
    klog.Infof("Identifying problematic nodes for draining, queue length: %v", action.Parameters["queue_length"])
    
    // 识别有问题的节点（例如：NotReady状态的节点）
    nodes, err := ae.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to list nodes: %v", err)
        return err
    }
    
    problematicNodes := []string{}
    for _, node := range nodes.Items {
        for _, condition := range node.Status.Conditions {
            if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
                problematicNodes = append(problematicNodes, node.Name)
                break
            }
        }
    }
    
    if len(problematicNodes) == 0 {
        action.Status = "completed"
        action.Result = "No problematic nodes found"
        return nil
    }
    
    // 这里可以实现节点排空逻辑
    klog.Infof("Found %d problematic nodes: %v", len(problematicNodes), problematicNodes)
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Identified %d problematic nodes for potential draining", len(problematicNodes))
    return nil
}

func (ae *AutomationEngine) defragmentResources(ctx context.Context, action RemediationAction) error {
    if !ae.config.Actions.EnablePodEviction {
        action.Status = "skipped"
        action.Result = "Pod eviction is disabled"
        return nil
    }
    
    klog.Infof("Defragmenting %s resources, fragmentation: %v", 
        action.Parameters["resource"], action.Parameters["fragmentation"])
    
    // 这里可以实现资源碎片整理逻辑
    // 例如：识别小的Pod并重新调度到更合适的节点
    
    action.Status = "completed"
    action.Result = "Resource defragmentation analysis completed"
    return nil
}

func (ae *AutomationEngine) rebalanceNode(ctx context.Context, action RemediationAction) error {
    if !ae.config.Actions.EnableNodeLabeling {
        action.Status = "skipped"
        action.Result = "Node labeling is disabled"
        return nil
    }
    
    nodeName := action.Parameters["node"].(string)
    klog.Infof("Rebalancing node %s due to resource hotspot", nodeName)
    
    // 为热点节点添加标签，防止新Pod调度到该节点
    node, err := ae.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to get node %s: %v", nodeName, err)
        return err
    }
    
    if node.Labels == nil {
        node.Labels = make(map[string]string)
    }
    
    node.Labels["scheduler.kubernetes.io/hotspot"] = "true"
    node.Labels["scheduler.kubernetes.io/hotspot-timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
    
    _, err = ae.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to update node labels: %v", err)
        return err
    }
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Added hotspot labels to node %s", nodeName)
    return nil
}

func (ae *AutomationEngine) consolidateWorkloads(ctx context.Context, action RemediationAction) error {
    klog.Infof("Analyzing workload consolidation opportunities")
    
    underutilizedNodes := action.Parameters["underutilized_nodes"].([]string)
    
    // 这里可以实现工作负载整合逻辑
    // 例如：建议将某些节点上的Pod迁移到其他节点
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Analyzed %d underutilized nodes for consolidation", len(underutilizedNodes))
    return nil
}

func (ae *AutomationEngine) isInCooldownPeriod(history *RemediationHistory) bool {
    history.mutex.RLock()
    defer history.mutex.RUnlock()
    
    if len(history.Actions) == 0 {
        return false
    }
    
    lastAction := history.Actions[len(history.Actions)-1]
    return time.Since(lastAction.Timestamp) < ae.config.RemediationCooldown
}

func (ae *AutomationEngine) hasExceededActionLimit(history *RemediationHistory) bool {
    history.mutex.RLock()
    defer history.mutex.RUnlock()
    
    // 计算最近1小时内的动作数量
    oneHourAgo := time.Now().Add(-time.Hour)
    recentActions := 0
    
    for _, action := range history.Actions {
        if action.Timestamp.After(oneHourAgo) {
            recentActions++
        }
    }
    
    return recentActions >= ae.config.MaxRemediationActions
}

// 清理过期的热点标签
func (ae *AutomationEngine) cleanupHotspotLabels(ctx context.Context) error {
    nodes, err := ae.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
        LabelSelector: "scheduler.kubernetes.io/hotspot=true",
    })
    if err != nil {
        return err
    }
    
    for _, node := range nodes.Items {
        timestampStr, exists := node.Labels["scheduler.kubernetes.io/hotspot-timestamp"]
        if !exists {
            continue
        }
        
        timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
        if err != nil {
            continue
        }
        
        // 如果标签超过30分钟，则移除
        if time.Since(time.Unix(timestamp, 0)) > 30*time.Minute {
            delete(node.Labels, "scheduler.kubernetes.io/hotspot")
            delete(node.Labels, "scheduler.kubernetes.io/hotspot-timestamp")
            
            _, err = ae.client.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
            if err != nil {
                klog.Errorf("Failed to remove hotspot labels from node %s: %v", node.Name, err)
            } else {
                klog.Infof("Removed expired hotspot labels from node %s", node.Name)
            }
        }
    }
    
    return nil
}
```

## 5. 故障排除与恢复

### 5.1 常见调度问题

#### 5.1.1 调度问题分类与诊断

```go
// scheduler-troubleshooter.go
package main

import (
    "context"
    "fmt"
    "strings"
    "time"
    
    "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type SchedulerTroubleshooter struct {
    client kubernetes.Interface
}

type TroubleshootingReport struct {
    PodName       string                 `json:"pod_name"`
    Namespace     string                 `json:"namespace"`
    IssueType     string                 `json:"issue_type"`
    Severity      string                 `json:"severity"`
    Description   string                 `json:"description"`
    Diagnosis     []DiagnosisStep        `json:"diagnosis"`
    Recommendations []TroubleshootingRecommendation `json:"recommendations"`
    Timestamp     time.Time              `json:"timestamp"`
}

type DiagnosisStep struct {
    Step        string      `json:"step"`
    Status      string      `json:"status"`
    Details     interface{} `json:"details"`
    Duration    time.Duration `json:"duration"`
}

type TroubleshootingRecommendation struct {
    Action      string `json:"action"`
    Description string `json:"description"`
    Command     string `json:"command,omitempty"`
    Priority    string `json:"priority"`
}

func NewSchedulerTroubleshooter(client kubernetes.Interface) *SchedulerTroubleshooter {
    return &SchedulerTroubleshooter{
        client: client,
    }
}

func (st *SchedulerTroubleshooter) DiagnosePendingPod(ctx context.Context, namespace, podName string) (*TroubleshootingReport, error) {
    report := &TroubleshootingReport{
        PodName:   podName,
        Namespace: namespace,
        Timestamp: time.Now(),
        Diagnosis: []DiagnosisStep{},
        Recommendations: []TroubleshootingRecommendation{},
    }
    
    // 获取Pod信息
    pod, err := st.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get pod: %v", err)
    }
    
    // 检查Pod状态
    if pod.Status.Phase != v1.PodPending {
        report.IssueType = "NotPending"
        report.Severity = "Info"
        report.Description = fmt.Sprintf("Pod is in %s phase, not pending", pod.Status.Phase)
        return report, nil
    }
    
    // 执行诊断步骤
    st.checkPodSchedulingConditions(ctx, pod, report)
    st.checkResourceRequirements(ctx, pod, report)
    st.checkNodeSelector(ctx, pod, report)
    st.checkAffinity(ctx, pod, report)
    st.checkTolerations(ctx, pod, report)
    st.checkPodSecurityPolicy(ctx, pod, report)
    st.checkResourceQuotas(ctx, pod, report)
    st.checkNodeAvailability(ctx, pod, report)
    
    // 确定问题类型和严重性
    st.categorizeIssue(report)
    
    // 生成建议
    st.generateRecommendations(report)
    
    return report, nil
}

func (st *SchedulerTroubleshooter) checkPodSchedulingConditions(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckSchedulingConditions",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    conditions := make(map[string]string)
    
    for _, condition := range pod.Status.Conditions {
        if condition.Type == v1.PodScheduled {
            conditions["PodScheduled"] = fmt.Sprintf("Status: %s, Reason: %s, Message: %s", 
                condition.Status, condition.Reason, condition.Message)
            
            if condition.Status == v1.ConditionFalse {
                step.Status = "Failed"
                step.Details = map[string]interface{}{
                    "reason":  condition.Reason,
                    "message": condition.Message,
                }
                return
            }
        }
    }
    
    step.Status = "Completed"
    step.Details = conditions
}

func (st *SchedulerTroubleshooter) checkResourceRequirements(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckResourceRequirements",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    var totalCPU, totalMemory int64
    resourceDetails := make(map[string]interface{})
    
    for _, container := range pod.Spec.Containers {
        if cpu := container.Resources.Requests.Cpu(); cpu != nil {
            totalCPU += cpu.MilliValue()
        }
        if memory := container.Resources.Requests.Memory(); memory != nil {
            totalMemory += memory.Value()
        }
    }
    
    resourceDetails["total_cpu_request"] = fmt.Sprintf("%dm", totalCPU)
    resourceDetails["total_memory_request"] = fmt.Sprintf("%dMi", totalMemory/(1024*1024))
    
    // 检查是否有节点能满足资源需求
    nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        step.Status = "Error"
        step.Details = map[string]interface{}{"error": err.Error()}
        return
    }
    
    suitableNodes := 0
    for _, node := range nodes.Items {
        allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
        allocatableMemory := node.Status.Allocatable.Memory().Value()
        
        if allocatableCPU >= totalCPU && allocatableMemory >= totalMemory {
            suitableNodes++
        }
    }
    
    resourceDetails["suitable_nodes"] = suitableNodes
    resourceDetails["total_nodes"] = len(nodes.Items)
    
    if suitableNodes == 0 {
        step.Status = "Failed"
        resourceDetails["issue"] = "No nodes have sufficient resources"
    } else {
        step.Status = "Completed"
    }
    
    step.Details = resourceDetails
}

func (st *SchedulerTroubleshooter) checkNodeSelector(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckNodeSelector",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    if len(pod.Spec.NodeSelector) == 0 {
        step.Status = "Skipped"
        step.Details = "No node selector specified"
        return
    }
    
    nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        step.Status = "Error"
        step.Details = map[string]interface{}{"error": err.Error()}
        return
    }
    
    matchingNodes := 0
    for _, node := range nodes.Items {
        matches := true
        for key, value := range pod.Spec.NodeSelector {
            if nodeValue, exists := node.Labels[key]; !exists || nodeValue != value {
                matches = false
                break
            }
        }
        if matches {
            matchingNodes++
        }
    }
    
    details := map[string]interface{}{
        "node_selector":   pod.Spec.NodeSelector,
        "matching_nodes":  matchingNodes,
        "total_nodes":     len(nodes.Items),
    }
    
    if matchingNodes == 0 {
        step.Status = "Failed"
        details["issue"] = "No nodes match the node selector"
    } else {
        step.Status = "Completed"
    }
    
    step.Details = details
}

func (st *SchedulerTroubleshooter) checkAffinity(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckAffinity",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    if pod.Spec.Affinity == nil {
        step.Status = "Skipped"
        step.Details = "No affinity rules specified"
        return
    }
    
    details := make(map[string]interface{})
    
    // 检查节点亲和性
    if pod.Spec.Affinity.NodeAffinity != nil {
        details["node_affinity"] = "present"
        
        if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
            details["required_node_affinity"] = "present"
        }
        
        if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
            details["preferred_node_affinity"] = len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
        }
    }
    
    // 检查Pod亲和性
    if pod.Spec.Affinity.PodAffinity != nil {
        details["pod_affinity"] = "present"
    }
    
    // 检查Pod反亲和性
    if pod.Spec.Affinity.PodAntiAffinity != nil {
        details["pod_anti_affinity"] = "present"
    }
    
    step.Status = "Completed"
    step.Details = details
}

func (st *SchedulerTroubleshooter) checkTolerations(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckTolerations",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    if len(pod.Spec.Tolerations) == 0 {
        step.Status = "Skipped"
        step.Details = "No tolerations specified"
        return
    }
    
    nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        step.Status = "Error"
        step.Details = map[string]interface{}{"error": err.Error()}
        return
    }
    
    tolerableNodes := 0
    taintedNodes := 0
    
    for _, node := range nodes.Items {
        if len(node.Spec.Taints) > 0 {
            taintedNodes++
            
            canTolerate := true
            for _, taint := range node.Spec.Taints {
                if !st.canTolerateTaint(pod.Spec.Tolerations, taint) {
                    canTolerate = false
                    break
                }
            }
            
            if canTolerate {
                tolerableNodes++
            }
        } else {
            tolerableNodes++
        }
    }
    
    details := map[string]interface{}{
        "tolerations":     len(pod.Spec.Tolerations),
        "tainted_nodes":   taintedNodes,
        "tolerable_nodes": tolerableNodes,
        "total_nodes":     len(nodes.Items),
    }
    
    if tolerableNodes == 0 {
        step.Status = "Failed"
        details["issue"] = "Pod cannot tolerate taints on any node"
    } else {
        step.Status = "Completed"
    }
    
    step.Details = details
}

func (st *SchedulerTroubleshooter) canTolerateTaint(tolerations []v1.Toleration, taint v1.Taint) bool {
    for _, toleration := range tolerations {
        if toleration.Key == taint.Key {
            if toleration.Operator == v1.TolerationOpExists {
                return true
            }
            if toleration.Operator == v1.TolerationOpEqual && toleration.Value == taint.Value {
                return true
            }
        }
    }
    return false
}

func (st *SchedulerTroubleshooter) checkPodSecurityPolicy(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckPodSecurityPolicy",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    // 简化的PSP检查
    details := make(map[string]interface{})
    
    if pod.Spec.SecurityContext != nil {
        details["security_context"] = "present"
        
        if pod.Spec.SecurityContext.RunAsUser != nil {
            details["run_as_user"] = *pod.Spec.SecurityContext.RunAsUser
        }
        
        if pod.Spec.SecurityContext.RunAsGroup != nil {
            details["run_as_group"] = *pod.Spec.SecurityContext.RunAsGroup
        }
        
        if pod.Spec.SecurityContext.FSGroup != nil {
            details["fs_group"] = *pod.Spec.SecurityContext.FSGroup
        }
    }
    
    step.Status = "Completed"
    step.Details = details
}

func (st *SchedulerTroubleshooter) checkResourceQuotas(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckResourceQuotas",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    quotas, err := st.client.CoreV1().ResourceQuotas(pod.Namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        step.Status = "Error"
        step.Details = map[string]interface{}{"error": err.Error()}
        return
    }
    
    if len(quotas.Items) == 0 {
        step.Status = "Skipped"
        step.Details = "No resource quotas in namespace"
        return
    }
    
    details := map[string]interface{}{
        "quotas_count": len(quotas.Items),
    }
    
    quotaIssues := []string{}
    for _, quota := range quotas.Items {
        for resource, used := range quota.Status.Used {
            if hard, exists := quota.Status.Hard[resource]; exists {
                if used.Cmp(hard) >= 0 {
                    quotaIssues = append(quotaIssues, fmt.Sprintf("%s: %s/%s", resource, used.String(), hard.String()))
                }
            }
        }
    }
    
    if len(quotaIssues) > 0 {
        step.Status = "Failed"
        details["quota_issues"] = quotaIssues
    } else {
        step.Status = "Completed"
    }
    
    step.Details = details
}

func (st *SchedulerTroubleshooter) checkNodeAvailability(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
    start := time.Now()
    step := DiagnosisStep{
        Step:   "CheckNodeAvailability",
        Status: "Running",
    }
    
    defer func() {
        step.Duration = time.Since(start)
        report.Diagnosis = append(report.Diagnosis, step)
    }()
    
    nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        step.Status = "Error"
        step.Details = map[string]interface{}{"error": err.Error()}
        return
    }
    
    readyNodes := 0
    schedulableNodes := 0
    
    for _, node := range nodes.Items {
        isReady := false
        for _, condition := range node.Status.Conditions {
            if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
                isReady = true
                readyNodes++
                break
            }
        }
        
        if isReady && !node.Spec.Unschedulable {
            schedulableNodes++
        }
    }
    
    details := map[string]interface{}{
        "total_nodes":       len(nodes.Items),
        "ready_nodes":       readyNodes,
        "schedulable_nodes": schedulableNodes,
    }
    
    if schedulableNodes == 0 {
        step.Status = "Failed"
        details["issue"] = "No schedulable nodes available"
    } else {
        step.Status = "Completed"
    }
    
    step.Details = details
}

func (st *SchedulerTroubleshooter) categorizeIssue(report *TroubleshootingReport) {
    failedSteps := []string{}
    
    for _, step := range report.Diagnosis {
        if step.Status == "Failed" {
            failedSteps = append(failedSteps, step.Step)
        }
    }
    
    if len(failedSteps) == 0 {
        report.IssueType = "Unknown"
        report.Severity = "Low"
        report.Description = "No obvious scheduling issues detected"
        return
    }
    
    // 根据失败的步骤确定问题类型
    if contains(failedSteps, "CheckNodeAvailability") {
        report.IssueType = "NodeUnavailability"
        report.Severity = "Critical"
        report.Description = "No schedulable nodes available"
    } else if contains(failedSteps, "CheckResourceRequirements") {
        report.IssueType = "InsufficientResources"
        report.Severity = "High"
        report.Description = "No nodes have sufficient resources"
    } else if contains(failedSteps, "CheckNodeSelector") {
        report.IssueType = "NodeSelectorMismatch"
        report.Severity = "Medium"
        report.Description = "No nodes match the node selector"
    } else if contains(failedSteps, "CheckTolerations") {
        report.IssueType = "TaintTolerationMismatch"
        report.Severity = "Medium"
        report.Description = "Pod cannot tolerate node taints"
    } else if contains(failedSteps, "CheckResourceQuotas") {
        report.IssueType = "ResourceQuotaExceeded"
        report.Severity = "Medium"
        report.Description = "Resource quota limits exceeded"
    } else {
        report.IssueType = "SchedulingConstraints"
        report.Severity = "Medium"
        report.Description = "Pod scheduling constraints cannot be satisfied"
    }
}

func (st *SchedulerTroubleshooter) generateRecommendations(report *TroubleshootingReport) {
    switch report.IssueType {
    case "NodeUnavailability":
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Check node status",
            Description: "Verify that nodes are ready and schedulable",
            Command:     "kubectl get nodes",
            Priority:    "High",
        })
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Check node conditions",
            Description: "Examine node conditions for issues",
            Command:     "kubectl describe nodes",
            Priority:    "High",
        })
    
    case "InsufficientResources":
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Scale cluster",
            Description: "Add more nodes to the cluster",
            Priority:    "High",
        })
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Optimize resource requests",
            Description: "Review and optimize pod resource requests",
            Priority:    "Medium",
        })
    
    case "NodeSelectorMismatch":
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Update node selector",
            Description: "Modify or remove node selector constraints",
            Priority:    "Medium",
        })
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Label nodes",
            Description: "Add required labels to nodes",
            Command:     "kubectl label nodes <node-name> <key>=<value>",
            Priority:    "Medium",
        })
    
    case "TaintTolerationMismatch":
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Add tolerations",
            Description: "Add appropriate tolerations to the pod",
            Priority:    "Medium",
        })
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Remove taints",
            Description: "Remove unnecessary taints from nodes",
            Command:     "kubectl taint nodes <node-name> <key>-",
            Priority:    "Low",
        })
    
    case "ResourceQuotaExceeded":
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Increase quota",
            Description: "Increase resource quota limits",
            Priority:    "Medium",
        })
        report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
            Action:      "Clean up resources",
            Description: "Remove unused resources to free up quota",
            Priority:    "Medium",
        })
    }
    
    // 通用建议
    report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
        Action:      "Check scheduler logs",
        Description: "Examine scheduler logs for detailed error messages",
        Command:     "kubectl logs -n kube-system -l component=kube-scheduler",
        Priority:    "Low",
    })
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

#### 5.1.2 故障恢复策略

```go
// scheduler-recovery.go
package main

import (
    "context"
    "fmt"
    "time"
    
    "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

type SchedulerRecoveryManager struct {
    client kubernetes.Interface
    config *RecoveryConfig
}

type RecoveryConfig struct {
    EnableAutoRecovery     bool          `yaml:"enable_auto_recovery"`
    RecoveryTimeout        time.Duration `yaml:"recovery_timeout"`
    MaxRecoveryAttempts    int           `yaml:"max_recovery_attempts"`
    HealthCheckInterval    time.Duration `yaml:"health_check_interval"`
    
    // 恢复策略配置
    Strategies RecoveryStrategies `yaml:"strategies"`
}

type RecoveryStrategies struct {
    SchedulerRestart    bool `yaml:"scheduler_restart"`
    NodeDraining        bool `yaml:"node_draining"`
    PodEviction         bool `yaml:"pod_eviction"`
    ResourceRebalancing bool `yaml:"resource_rebalancing"`
    ConfigRollback      bool `yaml:"config_rollback"`
}

type RecoveryAction struct {
    Type        string                 `json:"type"`
    Target      string                 `json:"target"`
    Parameters  map[string]interface{} `json:"parameters"`
    Timestamp   time.Time              `json:"timestamp"`
    Status      string                 `json:"status"`
    Result      string                 `json:"result"`
    Duration    time.Duration          `json:"duration"`
    Attempts    int                    `json:"attempts"`
}

type HealthStatus struct {
    Component   string    `json:"component"`
    Status      string    `json:"status"`
    LastCheck   time.Time `json:"last_check"`
    ErrorCount  int       `json:"error_count"`
    LastError   string    `json:"last_error,omitempty"`
}

func NewSchedulerRecoveryManager(client kubernetes.Interface, config *RecoveryConfig) *SchedulerRecoveryManager {
    return &SchedulerRecoveryManager{
        client: client,
        config: config,
    }
}

func (srm *SchedulerRecoveryManager) StartRecoveryLoop(ctx context.Context) {
    ticker := time.NewTicker(srm.config.HealthCheckInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := srm.performHealthCheck(ctx); err != nil {
                klog.Errorf("Health check failed: %v", err)
                if srm.config.EnableAutoRecovery {
                    if err := srm.initiateRecovery(ctx, err); err != nil {
                        klog.Errorf("Recovery failed: %v", err)
                    }
                }
            }
        }
    }
}

func (srm *SchedulerRecoveryManager) performHealthCheck(ctx context.Context) error {
    // 检查调度器健康状态
    schedulerHealth, err := srm.checkSchedulerHealth(ctx)
    if err != nil {
        return fmt.Errorf("scheduler health check failed: %v", err)
    }
    
    // 检查节点健康状态
    nodeHealth, err := srm.checkNodeHealth(ctx)
    if err != nil {
        return fmt.Errorf("node health check failed: %v", err)
    }
    
    // 检查调度队列健康状态
    queueHealth, err := srm.checkQueueHealth(ctx)
    if err != nil {
        return fmt.Errorf("queue health check failed: %v", err)
    }
    
    // 记录健康状态
    klog.V(4).Infof("Health check results - Scheduler: %s, Nodes: %s, Queue: %s", 
        schedulerHealth.Status, nodeHealth.Status, queueHealth.Status)
    
    // 如果任何组件不健康，返回错误
    if schedulerHealth.Status != "healthy" || nodeHealth.Status != "healthy" || queueHealth.Status != "healthy" {
        return fmt.Errorf("unhealthy components detected")
    }
    
    return nil
}

func (srm *SchedulerRecoveryManager) checkSchedulerHealth(ctx context.Context) (*HealthStatus, error) {
    health := &HealthStatus{
        Component: "scheduler",
        LastCheck: time.Now(),
        Status:    "healthy",
    }
    
    // 检查调度器Pod状态
    pods, err := srm.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
        LabelSelector: "component=kube-scheduler",
    })
    if err != nil {
        health.Status = "unhealthy"
        health.LastError = fmt.Sprintf("Failed to list scheduler pods: %v", err)
        return health, err
    }
    
    if len(pods.Items) == 0 {
        health.Status = "unhealthy"
        health.LastError = "No scheduler pods found"
        return health, fmt.Errorf("no scheduler pods found")
    }
    
    runningPods := 0
    for _, pod := range pods.Items {
        if pod.Status.Phase == v1.PodRunning {
            runningPods++
        }
    }
    
    if runningPods == 0 {
        health.Status = "unhealthy"
        health.LastError = "No running scheduler pods"
        return health, fmt.Errorf("no running scheduler pods")
    }
    
    return health, nil
}

func (srm *SchedulerRecoveryManager) checkNodeHealth(ctx context.Context) (*HealthStatus, error) {
    health := &HealthStatus{
        Component: "nodes",
        LastCheck: time.Now(),
        Status:    "healthy",
    }
    
    nodes, err := srm.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        health.Status = "unhealthy"
        health.LastError = fmt.Sprintf("Failed to list nodes: %v", err)
        return health, err
    }
    
    if len(nodes.Items) == 0 {
        health.Status = "unhealthy"
        health.LastError = "No nodes found"
        return health, fmt.Errorf("no nodes found")
    }
    
    readyNodes := 0
    for _, node := range nodes.Items {
        for _, condition := range node.Status.Conditions {
            if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
                readyNodes++
                break
            }
        }
    }
    
    // 如果少于50%的节点就绪，认为不健康
    if float64(readyNodes)/float64(len(nodes.Items)) < 0.5 {
        health.Status = "unhealthy"
        health.LastError = fmt.Sprintf("Only %d/%d nodes are ready", readyNodes, len(nodes.Items))
        return health, fmt.Errorf("insufficient ready nodes")
    }
    
    return health, nil
}

func (srm *SchedulerRecoveryManager) checkQueueHealth(ctx context.Context) (*HealthStatus, error) {
    health := &HealthStatus{
        Component: "queue",
        LastCheck: time.Now(),
        Status:    "healthy",
    }
    
    // 检查Pending状态的Pod数量
    pendingPods, err := srm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
        FieldSelector: "status.phase=Pending",
    })
    if err != nil {
        health.Status = "unhealthy"
        health.LastError = fmt.Sprintf("Failed to list pending pods: %v", err)
        return health, err
    }
    
    // 如果Pending Pod数量过多，认为队列不健康
    if len(pendingPods.Items) > 100 {
        health.Status = "unhealthy"
        health.LastError = fmt.Sprintf("Too many pending pods: %d", len(pendingPods.Items))
        return health, fmt.Errorf("queue overloaded")
    }
    
    // 检查长时间Pending的Pod
    longPendingPods := 0
    for _, pod := range pendingPods.Items {
        if time.Since(pod.CreationTimestamp.Time) > 10*time.Minute {
            longPendingPods++
        }
    }
    
    if longPendingPods > 10 {
        health.Status = "unhealthy"
        health.LastError = fmt.Sprintf("Too many long-pending pods: %d", longPendingPods)
        return health, fmt.Errorf("scheduling stalled")
    }
    
    return health, nil
}

func (srm *SchedulerRecoveryManager) initiateRecovery(ctx context.Context, healthErr error) error {
    klog.Infof("Initiating recovery due to health check failure: %v", healthErr)
    
    // 确定恢复策略
    actions := srm.determineRecoveryActions(healthErr)
    
    // 执行恢复动作
    for _, action := range actions {
        if err := srm.executeRecoveryAction(ctx, action); err != nil {
            klog.Errorf("Recovery action %s failed: %v", action.Type, err)
            continue
        }
        
        // 等待恢复生效
        time.Sleep(30 * time.Second)
        
        // 重新检查健康状态
        if err := srm.performHealthCheck(ctx); err == nil {
            klog.Infof("Recovery successful after action: %s", action.Type)
            return nil
        }
    }
    
    return fmt.Errorf("all recovery actions failed")
}

func (srm *SchedulerRecoveryManager) determineRecoveryActions(healthErr error) []RecoveryAction {
    var actions []RecoveryAction
    
    errorMsg := healthErr.Error()
    
    // 基于错误类型确定恢复策略
    if strings.Contains(errorMsg, "scheduler") {
        if srm.config.Strategies.SchedulerRestart {
            actions = append(actions, RecoveryAction{
                Type:      "restart_scheduler",
                Target:    "kube-scheduler",
                Timestamp: time.Now(),
                Status:    "pending",
            })
        }
    }
    
    if strings.Contains(errorMsg, "nodes") || strings.Contains(errorMsg, "ready") {
        if srm.config.Strategies.NodeDraining {
            actions = append(actions, RecoveryAction{
                Type:      "drain_unhealthy_nodes",
                Target:    "cluster",
                Timestamp: time.Now(),
                Status:    "pending",
            })
        }
    }
    
    if strings.Contains(errorMsg, "queue") || strings.Contains(errorMsg, "pending") {
        if srm.config.Strategies.PodEviction {
            actions = append(actions, RecoveryAction{
                Type:      "evict_stuck_pods",
                Target:    "cluster",
                Timestamp: time.Now(),
                Status:    "pending",
            })
        }
        
        if srm.config.Strategies.ResourceRebalancing {
            actions = append(actions, RecoveryAction{
                Type:      "rebalance_resources",
                Target:    "cluster",
                Timestamp: time.Now(),
                Status:    "pending",
            })
        }
    }
    
    return actions
}

func (srm *SchedulerRecoveryManager) executeRecoveryAction(ctx context.Context, action RecoveryAction) error {
    start := time.Now()
    action.Status = "executing"
    action.Attempts++
    
    defer func() {
        action.Duration = time.Since(start)
    }()
    
    switch action.Type {
    case "restart_scheduler":
        return srm.restartScheduler(ctx, action)
    case "drain_unhealthy_nodes":
        return srm.drainUnhealthyNodes(ctx, action)
    case "evict_stuck_pods":
        return srm.evictStuckPods(ctx, action)
    case "rebalance_resources":
        return srm.rebalanceResources(ctx, action)
    default:
        action.Status = "failed"
        action.Result = fmt.Sprintf("Unknown recovery action: %s", action.Type)
        return fmt.Errorf("unknown recovery action: %s", action.Type)
    }
}

func (srm *SchedulerRecoveryManager) restartScheduler(ctx context.Context, action RecoveryAction) error {
    klog.Infof("Restarting scheduler pods")
    
    // 获取调度器Pod
    pods, err := srm.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
        LabelSelector: "component=kube-scheduler",
    })
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to list scheduler pods: %v", err)
        return err
    }
    
    // 删除调度器Pod以触发重启
    for _, pod := range pods.Items {
        err := srm.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
        if err != nil {
            klog.Errorf("Failed to delete scheduler pod %s: %v", pod.Name, err)
            continue
        }
        klog.Infof("Deleted scheduler pod: %s", pod.Name)
    }
    
    // 等待新Pod启动
    time.Sleep(60 * time.Second)
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Restarted %d scheduler pods", len(pods.Items))
    return nil
}

func (srm *SchedulerRecoveryManager) drainUnhealthyNodes(ctx context.Context, action RecoveryAction) error {
    klog.Infof("Draining unhealthy nodes")
    
    nodes, err := srm.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to list nodes: %v", err)
        return err
    }
    
    unhealthyNodes := []string{}
    for _, node := range nodes.Items {
        isReady := false
        for _, condition := range node.Status.Conditions {
            if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
                isReady = true
                break
            }
        }
        
        if !isReady {
            unhealthyNodes = append(unhealthyNodes, node.Name)
            
            // 标记节点为不可调度
            node.Spec.Unschedulable = true
            _, err := srm.client.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
            if err != nil {
                klog.Errorf("Failed to mark node %s as unschedulable: %v", node.Name, err)
            } else {
                klog.Infof("Marked node %s as unschedulable", node.Name)
            }
        }
    }
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Drained %d unhealthy nodes: %v", len(unhealthyNodes), unhealthyNodes)
    return nil
}

func (srm *SchedulerRecoveryManager) evictStuckPods(ctx context.Context, action RecoveryAction) error {
    klog.Infof("Evicting stuck pods")
    
    // 获取长时间Pending的Pod
    pendingPods, err := srm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
        FieldSelector: "status.phase=Pending",
    })
    if err != nil {
        action.Status = "failed"
        action.Result = fmt.Sprintf("Failed to list pending pods: %v", err)
        return err
    }
    
    evictedPods := []string{}
    for _, pod := range pendingPods.Items {
        // 只驱逐超过10分钟的Pending Pod
        if time.Since(pod.CreationTimestamp.Time) > 10*time.Minute {
            err := srm.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
            if err != nil {
                klog.Errorf("Failed to evict pod %s/%s: %v", pod.Namespace, pod.Name, err)
                continue
            }
            evictedPods = append(evictedPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
            klog.Infof("Evicted stuck pod: %s/%s", pod.Namespace, pod.Name)
        }
    }
    
    action.Status = "completed"
    action.Result = fmt.Sprintf("Evicted %d stuck pods", len(evictedPods))
    return nil
}

func (srm *SchedulerRecoveryManager) rebalanceResources(ctx context.Context, action RecoveryAction) error {
    klog.Infof("Rebalancing cluster resources")
    
    // 这里可以实现资源重平衡逻辑
    // 例如：识别资源使用不均衡的节点，建议Pod迁移等
    
    action.Status = "completed"
    action.Result = "Resource rebalancing analysis completed"
    return nil
}
```

### 5.2 最佳实践总结

#### 5.2.1 调度器配置最佳实践

**生产环境配置模板：**

```yaml
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
metadata:
  name: production-scheduler-config
clientConnection:
  kubeconfig: "/etc/kubernetes/scheduler.conf"
  qps: 100
  burst: 200
leaderElection:
  leaderElect: true
  leaseDuration: 15s
  renewDeadline: 10s
  retryPeriod: 2s
  resourceLock: leases
  resourceName: kube-scheduler
  resourceNamespace: kube-system
profiles:
- schedulerName: production-scheduler
  plugins:
    filter:
      enabled:
      - name: NodeResourcesFit
      - name: NodeAffinity
      - name: PodTopologySpread
      - name: TaintToleration
      - name: VolumeBinding
      disabled:
      - name: NodePorts
    score:
      enabled:
      - name: NodeResourcesFit
        weight: 10
      - name: NodeAffinity
        weight: 5
      - name: PodTopologySpread
        weight: 3
      - name: InterPodAffinity
        weight: 2
  pluginConfig:
  - name: NodeResourcesFit
    args:
      scoringStrategy:
        type: LeastAllocated
        resources:
        - name: cpu
          weight: 1
        - name: memory
          weight: 1
  - name: PodTopologySpread
    args:
      defaultConstraints:
      - maxSkew: 1
        topologyKey: topology.kubernetes.io/zone
        whenUnsatisfiable: ScheduleAnyway
      - maxSkew: 2
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: ScheduleAnyway
parallelism: 16
percentageOfNodesToScore: 50
```

**安全配置要点：**

```yaml
# RBAC 配置
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:kube-scheduler
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "update", "patch"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["persistentvolumes"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses", "csinodes", "csidrivers", "csistoragecapacities"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["create", "get", "list", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:kube-scheduler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:kube-scheduler
subjects:
- kind: ServiceAccount
  name: kube-scheduler
  namespace: kube-system
```

**高可用部署方案：**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-scheduler
  namespace: kube-system
  labels:
    component: kube-scheduler
    tier: control-plane
spec:
  replicas: 3
  selector:
    matchLabels:
      component: kube-scheduler
      tier: control-plane
  template:
    metadata:
      labels:
        component: kube-scheduler
        tier: control-plane
    spec:
      serviceAccountName: kube-scheduler
      containers:
      - name: kube-scheduler
        image: k8s.gcr.io/kube-scheduler:v1.28.0
        command:
        - kube-scheduler
        - --config=/etc/kubernetes/scheduler-config.yaml
        - --v=2
        - --leader-elect=true
        - --leader-elect-lease-duration=15s
        - --leader-elect-renew-deadline=10s
        - --leader-elect-retry-period=2s
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 15
          timeoutSeconds: 15
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 5
          timeoutSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /etc/kubernetes
          readOnly: true
        - name: kubeconfig
          mountPath: /etc/kubernetes/scheduler.conf
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: scheduler-config
      - name: kubeconfig
        hostPath:
          path: /etc/kubernetes/scheduler.conf
          type: FileOrCreate
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                component: kube-scheduler
            topologyKey: kubernetes.io/hostname
      tolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
```

#### 5.2.2 资源配额和限制

**ResourceQuota 配置：**

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: production
spec:
  hard:
    # 计算资源限制
    requests.cpu: "100"
    requests.memory: 200Gi
    limits.cpu: "200"
    limits.memory: 400Gi
    
    # 存储资源限制
    requests.storage: 1Ti
    persistentvolumeclaims: "50"
    
    # 对象数量限制
    pods: "100"
    services: "20"
    secrets: "50"
    configmaps: "50"
    
    # 扩展资源限制
    requests.nvidia.com/gpu: "10"
    limits.nvidia.com/gpu: "10"
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: priority-quota
  namespace: production
spec:
  hard:
    pods: "50"
    requests.cpu: "50"
    requests.memory: 100Gi
  scopeSelector:
    matchExpressions:
    - operator: In
      scopeName: PriorityClass
      values: ["high-priority"]
```

**LimitRange 设置：**

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: resource-limits
  namespace: production
spec:
  limits:
  # 容器级别限制
  - type: Container
    default:
      cpu: "500m"
      memory: "512Mi"
    defaultRequest:
      cpu: "100m"
      memory: "128Mi"
    max:
      cpu: "2"
      memory: "4Gi"
    min:
      cpu: "50m"
      memory: "64Mi"
    maxLimitRequestRatio:
      cpu: "4"
      memory: "4"
  
  # Pod 级别限制
  - type: Pod
    max:
      cpu: "4"
      memory: "8Gi"
    min:
      cpu: "100m"
      memory: "128Mi"
  
  # PVC 限制
  - type: PersistentVolumeClaim
    max:
      storage: "100Gi"
    min:
      storage: "1Gi"
```

**多租户资源管理实现：**

```go
package quota

import (
    "context"
    "fmt"
    "time"
    
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

// TenantResourceManager 多租户资源管理器
type TenantResourceManager struct {
    client kubernetes.Interface
}

// TenantQuota 租户配额定义
type TenantQuota struct {
    TenantID    string
    Namespace   string
    CPUQuota    resource.Quantity
    MemoryQuota resource.Quantity
    GPUQuota    int64
    StorageQuota resource.Quantity
    Priority    int32
}

// CreateTenantQuota 创建租户配额
func (trm *TenantResourceManager) CreateTenantQuota(ctx context.Context, quota TenantQuota) error {
    // 创建 ResourceQuota
    resourceQuota := &corev1.ResourceQuota{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-quota", quota.TenantID),
            Namespace: quota.Namespace,
            Labels: map[string]string{
                "tenant-id": quota.TenantID,
                "managed-by": "tenant-resource-manager",
            },
        },
        Spec: corev1.ResourceQuotaSpec{
            Hard: corev1.ResourceList{
                "requests.cpu":    quota.CPUQuota,
                "requests.memory": quota.MemoryQuota,
                "limits.cpu":      quota.CPUQuota,
                "limits.memory":   quota.MemoryQuota,
                "requests.storage": quota.StorageQuota,
            },
        },
    }
    
    // 如果有 GPU 配额，添加 GPU 资源
    if quota.GPUQuota > 0 {
        resourceQuota.Spec.Hard["requests.nvidia.com/gpu"] = *resource.NewQuantity(quota.GPUQuota, resource.DecimalSI)
        resourceQuota.Spec.Hard["limits.nvidia.com/gpu"] = *resource.NewQuantity(quota.GPUQuota, resource.DecimalSI)
    }
    
    _, err := trm.client.CoreV1().ResourceQuotas(quota.Namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create resource quota: %v", err)
    }
    
    // 创建 LimitRange
    limitRange := &corev1.LimitRange{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-limits", quota.TenantID),
            Namespace: quota.Namespace,
            Labels: map[string]string{
                "tenant-id": quota.TenantID,
                "managed-by": "tenant-resource-manager",
            },
        },
        Spec: corev1.LimitRangeSpec{
            Limits: []corev1.LimitRangeItem{
                {
                    Type: corev1.LimitTypeContainer,
                    Default: corev1.ResourceList{
                        "cpu":    resource.MustParse("500m"),
                        "memory": resource.MustParse("512Mi"),
                    },
                    DefaultRequest: corev1.ResourceList{
                        "cpu":    resource.MustParse("100m"),
                        "memory": resource.MustParse("128Mi"),
                    },
                    Max: corev1.ResourceList{
                        "cpu":    quota.CPUQuota,
                        "memory": quota.MemoryQuota,
                    },
                },
            },
        },
    }
    
    _, err = trm.client.CoreV1().LimitRanges(quota.Namespace).Create(ctx, limitRange, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create limit range: %v", err)
    }
    
    klog.Infof("Created tenant quota for %s in namespace %s", quota.TenantID, quota.Namespace)
    return nil
}

// GetTenantResourceUsage 获取租户资源使用情况
func (trm *TenantResourceManager) GetTenantResourceUsage(ctx context.Context, tenantID, namespace string) (*TenantResourceUsage, error) {
    quotaName := fmt.Sprintf("%s-quota", tenantID)
    quota, err := trm.client.CoreV1().ResourceQuotas(namespace).Get(ctx, quotaName, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get resource quota: %v", err)
    }
    
    usage := &TenantResourceUsage{
        TenantID:  tenantID,
        Namespace: namespace,
        Timestamp: time.Now(),
    }
    
    // 解析资源使用情况
    if cpuUsed, ok := quota.Status.Used["requests.cpu"]; ok {
        usage.CPUUsed = cpuUsed
    }
    if memoryUsed, ok := quota.Status.Used["requests.memory"]; ok {
        usage.MemoryUsed = memoryUsed
    }
    if gpuUsed, ok := quota.Status.Used["requests.nvidia.com/gpu"]; ok {
        usage.GPUUsed = gpuUsed.Value()
    }
    
    // 解析资源配额
    if cpuHard, ok := quota.Status.Hard["requests.cpu"]; ok {
        usage.CPUQuota = cpuHard
    }
    if memoryHard, ok := quota.Status.Hard["requests.memory"]; ok {
        usage.MemoryQuota = memoryHard
    }
    if gpuHard, ok := quota.Status.Hard["requests.nvidia.com/gpu"]; ok {
        usage.GPUQuota = gpuHard.Value()
    }
    
    return usage, nil
}

// TenantResourceUsage 租户资源使用情况
type TenantResourceUsage struct {
    TenantID    string
    Namespace   string
    Timestamp   time.Time
    CPUUsed     resource.Quantity
    CPUQuota    resource.Quantity
    MemoryUsed  resource.Quantity
    MemoryQuota resource.Quantity
    GPUUsed     int64
    GPUQuota    int64
}

// GetUtilizationRate 获取资源利用率
func (usage *TenantResourceUsage) GetUtilizationRate() map[string]float64 {
    rates := make(map[string]float64)
    
    // CPU 利用率
    if !usage.CPUQuota.IsZero() {
        rates["cpu"] = float64(usage.CPUUsed.MilliValue()) / float64(usage.CPUQuota.MilliValue())
    }
    
    // 内存利用率
    if !usage.MemoryQuota.IsZero() {
        rates["memory"] = float64(usage.MemoryUsed.Value()) / float64(usage.MemoryQuota.Value())
    }
    
    // GPU 利用率
    if usage.GPUQuota > 0 {
        rates["gpu"] = float64(usage.GPUUsed) / float64(usage.GPUQuota)
    }
    
    return rates
}
```

#### 5.2.3 调度器性能调优

**性能指标监控：**

```go
package metrics

import (
    "context"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "k8s.io/klog/v2"
)

// SchedulerMetrics 调度器性能指标
type SchedulerMetrics struct {
    // 调度延迟
    schedulingLatency prometheus.Histogram
    
    // 调度吞吐量
    schedulingThroughput prometheus.Counter
    
    // 调度成功率
    schedulingSuccessRate prometheus.Gauge
    
    // 节点评分时间
    nodeScoreLatency prometheus.Histogram
    
    // 插件执行时间
    pluginExecutionLatency prometheus.HistogramVec
    
    // 调度队列长度
    schedulingQueueLength prometheus.Gauge
    
    // 资源利用率
    resourceUtilization prometheus.GaugeVec
}

// NewSchedulerMetrics 创建调度器指标
func NewSchedulerMetrics() *SchedulerMetrics {
    return &SchedulerMetrics{
        schedulingLatency: promauto.NewHistogram(prometheus.HistogramOpts{
            Name: "scheduler_scheduling_latency_seconds",
            Help: "Scheduling latency in seconds",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
        }),
        
        schedulingThroughput: promauto.NewCounter(prometheus.CounterOpts{
            Name: "scheduler_scheduling_total",
            Help: "Total number of scheduling attempts",
        }),
        
        schedulingSuccessRate: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "scheduler_scheduling_success_rate",
            Help: "Scheduling success rate",
        }),
        
        nodeScoreLatency: promauto.NewHistogram(prometheus.HistogramOpts{
            Name: "scheduler_node_score_latency_seconds",
            Help: "Node scoring latency in seconds",
            Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15),
        }),
        
        pluginExecutionLatency: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "scheduler_plugin_execution_latency_seconds",
                Help: "Plugin execution latency in seconds",
                Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15),
            },
            []string{"plugin", "extension_point"},
        ),
        
        schedulingQueueLength: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "scheduler_queue_length",
            Help: "Length of the scheduling queue",
        }),
        
        resourceUtilization: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "scheduler_resource_utilization",
                Help: "Resource utilization rate",
            },
            []string{"resource", "node"},
        ),
    }
}

// RecordSchedulingLatency 记录调度延迟
func (sm *SchedulerMetrics) RecordSchedulingLatency(duration time.Duration) {
    sm.schedulingLatency.Observe(duration.Seconds())
    sm.schedulingThroughput.Inc()
}

// RecordPluginLatency 记录插件执行延迟
func (sm *SchedulerMetrics) RecordPluginLatency(plugin, extensionPoint string, duration time.Duration) {
    sm.pluginExecutionLatency.WithLabelValues(plugin, extensionPoint).Observe(duration.Seconds())
}

// UpdateQueueLength 更新队列长度
func (sm *SchedulerMetrics) UpdateQueueLength(length int) {
    sm.schedulingQueueLength.Set(float64(length))
}

// UpdateResourceUtilization 更新资源利用率
func (sm *SchedulerMetrics) UpdateResourceUtilization(resource, node string, utilization float64) {
    sm.resourceUtilization.WithLabelValues(resource, node).Set(utilization)
}
```

**性能调优配置：**

```yaml
# 高性能调度器配置
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
metadata:
  name: high-performance-scheduler
profiles:
- schedulerName: high-performance-scheduler
  plugins:
    filter:
      enabled:
      - name: NodeResourcesFit
      - name: NodeAffinity
      - name: TaintToleration
      # 禁用不必要的插件以提高性能
      disabled:
      - name: VolumeRestrictions
      - name: EBSLimits
      - name: GCEPDLimits
    score:
      enabled:
      - name: NodeResourcesFit
        weight: 10
      - name: NodeAffinity
        weight: 5
      disabled:
      - name: NodeResourcesBalancedAllocation  # 在大规模集群中可能较慢
  pluginConfig:
  - name: NodeResourcesFit
    args:
      scoringStrategy:
        type: LeastAllocated  # 比 MostAllocated 更快
        resources:
        - name: cpu
          weight: 1
        - name: memory
          weight: 1
# 性能优化参数
parallelism: 32  # 增加并发度
percentageOfNodesToScore: 30  # 减少评分节点百分比
```

**批量调度优化：**

```go
package batch

import (
    "context"
    "fmt"
    "sort"
    "sync"
    "time"
    
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

// BatchScheduler 批量调度器
type BatchScheduler struct {
    client       kubernetes.Interface
    batchSize    int
    batchTimeout time.Duration
    metrics      *SchedulerMetrics
    mu           sync.RWMutex
    pendingPods  []*corev1.Pod
}

// NewBatchScheduler 创建批量调度器
func NewBatchScheduler(client kubernetes.Interface, batchSize int, batchTimeout time.Duration) *BatchScheduler {
    return &BatchScheduler{
        client:       client,
        batchSize:    batchSize,
        batchTimeout: batchTimeout,
        metrics:      NewSchedulerMetrics(),
        pendingPods:  make([]*corev1.Pod, 0),
    }
}

// ScheduleBatch 批量调度 Pod
func (bs *BatchScheduler) ScheduleBatch(ctx context.Context, pods []*corev1.Pod) error {
    start := time.Now()
    defer func() {
        bs.metrics.RecordSchedulingLatency(time.Since(start))
    }()
    
    // 按优先级排序
    sort.Slice(pods, func(i, j int) bool {
        return getPodPriority(pods[i]) > getPodPriority(pods[j])
    })
    
    // 获取可用节点
    nodes, err := bs.getAvailableNodes(ctx)
    if err != nil {
        return fmt.Errorf("failed to get available nodes: %v", err)
    }
    
    // 批量调度
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // 限制并发数
    
    for _, pod := range pods {
        wg.Add(1)
        go func(p *corev1.Pod) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            if err := bs.schedulePod(ctx, p, nodes); err != nil {
                klog.Errorf("Failed to schedule pod %s/%s: %v", p.Namespace, p.Name, err)
            }
        }(pod)
    }
    
    wg.Wait()
    return nil
}

// schedulePod 调度单个 Pod
func (bs *BatchScheduler) schedulePod(ctx context.Context, pod *corev1.Pod, nodes []*corev1.Node) error {
    // 过滤节点
    filteredNodes := bs.filterNodes(pod, nodes)
    if len(filteredNodes) == 0 {
        return fmt.Errorf("no suitable nodes found for pod %s/%s", pod.Namespace, pod.Name)
    }
    
    // 评分节点
    scoredNodes := bs.scoreNodes(pod, filteredNodes)
    
    // 选择最佳节点
    bestNode := scoredNodes[0].Node
    
    // 绑定 Pod 到节点
    return bs.bindPod(ctx, pod, bestNode)
}

// NodeScore 节点评分
type NodeScore struct {
    Node  *corev1.Node
    Score int64
}

// scoreNodes 对节点进行评分
func (bs *BatchScheduler) scoreNodes(pod *corev1.Pod, nodes []*corev1.Node) []NodeScore {
    scores := make([]NodeScore, 0, len(nodes))
    
    for _, node := range nodes {
        score := bs.calculateNodeScore(pod, node)
        scores = append(scores, NodeScore{
            Node:  node,
            Score: score,
        })
    }
    
    // 按分数降序排序
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].Score > scores[j].Score
    })
    
    return scores
}

// calculateNodeScore 计算节点分数
func (bs *BatchScheduler) calculateNodeScore(pod *corev1.Pod, node *corev1.Node) int64 {
    var score int64
    
    // 资源适配性评分
    score += bs.scoreResourceFit(pod, node)
    
    // 节点亲和性评分
    score += bs.scoreNodeAffinity(pod, node)
    
    // 负载均衡评分
    score += bs.scoreLoadBalance(node)
    
    return score
}

// scoreResourceFit 资源适配性评分
func (bs *BatchScheduler) scoreResourceFit(pod *corev1.Pod, node *corev1.Node) int64 {
    // 计算资源请求
    cpuRequest := getPodCPURequest(pod)
    memoryRequest := getPodMemoryRequest(pod)
    
    // 计算节点可用资源
    cpuAvailable := getNodeCPUAvailable(node)
    memoryAvailable := getNodeMemoryAvailable(node)
    
    // 计算资源利用率
    cpuUtilization := float64(cpuRequest) / float64(cpuAvailable)
    memoryUtilization := float64(memoryRequest) / float64(memoryAvailable)
    
    // 偏好资源利用率较低的节点
    score := int64((1.0 - cpuUtilization) * 50)
    score += int64((1.0 - memoryUtilization) * 50)
    
    return score
}

// filterNodes 过滤节点
func (bs *BatchScheduler) filterNodes(pod *corev1.Pod, nodes []*corev1.Node) []*corev1.Node {
    var filtered []*corev1.Node
    
    for _, node := range nodes {
        if bs.nodeFilter(pod, node) {
            filtered = append(filtered, node)
        }
    }
    
    return filtered
}

// nodeFilter 节点过滤器
func (bs *BatchScheduler) nodeFilter(pod *corev1.Pod, node *corev1.Node) bool {
    // 检查节点是否就绪
    if !isNodeReady(node) {
        return false
    }
    
    // 检查资源是否充足
    if !hasEnoughResources(pod, node) {
        return false
    }
    
    // 检查污点容忍
    if !toleratesTaints(pod, node) {
        return false
    }
    
    // 检查节点亲和性
    if !matchesNodeAffinity(pod, node) {
        return false
    }
    
    return true
}

// 辅助函数
func getPodPriority(pod *corev1.Pod) int32 {
    if pod.Spec.Priority != nil {
        return *pod.Spec.Priority
    }
    return 0
}

func getPodCPURequest(pod *corev1.Pod) int64 {
    var total int64
    for _, container := range pod.Spec.Containers {
        if cpu := container.Resources.Requests[corev1.ResourceCPU]; !cpu.IsZero() {
            total += cpu.MilliValue()
        }
    }
    return total
}

func getPodMemoryRequest(pod *corev1.Pod) int64 {
    var total int64
    for _, container := range pod.Spec.Containers {
        if memory := container.Resources.Requests[corev1.ResourceMemory]; !memory.IsZero() {
            total += memory.Value()
        }
    }
    return total
}

func getNodeCPUAvailable(node *corev1.Node) int64 {
    if cpu := node.Status.Allocatable[corev1.ResourceCPU]; !cpu.IsZero() {
        return cpu.MilliValue()
    }
    return 0
}

func getNodeMemoryAvailable(node *corev1.Node) int64 {
    if memory := node.Status.Allocatable[corev1.ResourceMemory]; !memory.IsZero() {
        return memory.Value()
    }
    return 0
}

func isNodeReady(node *corev1.Node) bool {
    for _, condition := range node.Status.Conditions {
        if condition.Type == corev1.NodeReady {
            return condition.Status == corev1.ConditionTrue
        }
    }
    return false
}

func hasEnoughResources(pod *corev1.Pod, node *corev1.Node) bool {
    cpuRequest := getPodCPURequest(pod)
    memoryRequest := getPodMemoryRequest(pod)
    
    cpuAvailable := getNodeCPUAvailable(node)
    memoryAvailable := getNodeMemoryAvailable(node)
    
    return cpuRequest <= cpuAvailable && memoryRequest <= memoryAvailable
}

func toleratesTaints(pod *corev1.Pod, node *corev1.Node) bool {
    // 简化的污点容忍检查
    for _, taint := range node.Spec.Taints {
        tolerated := false
        for _, toleration := range pod.Spec.Tolerations {
            if toleration.Key == taint.Key && toleration.Effect == taint.Effect {
                tolerated = true
                break
            }
        }
        if !tolerated {
            return false
        }
    }
    return true
}

func matchesNodeAffinity(pod *corev1.Pod, node *corev1.Node) bool {
    // 简化的节点亲和性检查
    if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil {
        return true
    }
    
    // 这里应该实现完整的节点亲和性检查逻辑
    return true
}

func (bs *BatchScheduler) scoreNodeAffinity(pod *corev1.Pod, node *corev1.Node) int64 {
    // 节点亲和性评分逻辑
    return 0
}

func (bs *BatchScheduler) scoreLoadBalance(node *corev1.Node) int64 {
    // 负载均衡评分逻辑
    return 0
}

func (bs *BatchScheduler) getAvailableNodes(ctx context.Context) ([]*corev1.Node, error) {
    nodeList, err := bs.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    var nodes []*corev1.Node
    for i := range nodeList.Items {
        nodes = append(nodes, &nodeList.Items[i])
    }
    
    return nodes, nil
}

func (bs *BatchScheduler) bindPod(ctx context.Context, pod *corev1.Pod, node *corev1.Node) error {
    // 创建绑定对象
    binding := &corev1.Binding{
        ObjectMeta: metav1.ObjectMeta{
            Name:      pod.Name,
            Namespace: pod.Namespace,
        },
        Target: corev1.ObjectReference{
            Kind: "Node",
            Name: node.Name,
        },
    }
    
    // 执行绑定
    err := bs.client.CoreV1().Pods(pod.Namespace).Bind(ctx, binding, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to bind pod %s/%s to node %s: %v", pod.Namespace, pod.Name, node.Name, err)
    }
    
    klog.Infof("Successfully bound pod %s/%s to node %s", pod.Namespace, pod.Name, node.Name)
    return nil
}
```

### 5.3 故障处理和恢复

#### 5.3.1 调度器故障检测

**健康检查配置：**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-health-config
  namespace: kube-system
data:
  health-check.yaml: |
    healthChecks:
      - name: scheduler-api
        endpoint: "https://localhost:10259/healthz"
        interval: 10s
        timeout: 5s
        retries: 3
      - name: scheduler-metrics
        endpoint: "https://localhost:10259/metrics"
        interval: 30s
        timeout: 10s
        retries: 2
      - name: leader-election
        endpoint: "https://localhost:10259/healthz/leader-election"
        interval: 15s
        timeout: 5s
        retries: 3
    alerts:
      - name: SchedulerDown
        condition: "scheduler-api == false"
        severity: critical
        message: "Scheduler API is not responding"
      - name: SchedulerNotLeader
        condition: "leader-election == false"
        severity: warning
        message: "Scheduler is not the leader"
```

**故障检测实现：**

```go
package health

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "k8s.io/klog/v2"
)

// HealthChecker 健康检查器
type HealthChecker struct {
    checks   []HealthCheck
    interval time.Duration
    alerts   AlertManager
}

// HealthCheck 健康检查定义
type HealthCheck struct {
    Name     string
    Endpoint string
    Interval time.Duration
    Timeout  time.Duration
    Retries  int
}

// HealthStatus 健康状态
type HealthStatus struct {
    Name      string
    Healthy   bool
    LastCheck time.Time
    Error     error
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(checks []HealthCheck, interval time.Duration) *HealthChecker {
    return &HealthChecker{
        checks:   checks,
        interval: interval,
        alerts:   NewAlertManager(),
    }
}

// Start 启动健康检查
func (hc *HealthChecker) Start(ctx context.Context) {
    ticker := time.NewTicker(hc.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            hc.runHealthChecks(ctx)
        }
    }
}

// runHealthChecks 执行健康检查
func (hc *HealthChecker) runHealthChecks(ctx context.Context) {
    for _, check := range hc.checks {
        go func(c HealthCheck) {
            status := hc.performCheck(ctx, c)
            hc.handleHealthStatus(status)
        }(check)
    }
}

// performCheck 执行单个健康检查
func (hc *HealthChecker) performCheck(ctx context.Context, check HealthCheck) HealthStatus {
    status := HealthStatus{
        Name:      check.Name,
        LastCheck: time.Now(),
    }
    
    client := &http.Client{
        Timeout: check.Timeout,
    }
    
    var lastErr error
    for i := 0; i <= check.Retries; i++ {
        req, err := http.NewRequestWithContext(ctx, "GET", check.Endpoint, nil)
        if err != nil {
            lastErr = err
            continue
        }
        
        resp, err := client.Do(req)
        if err != nil {
            lastErr = err
            continue
        }
        
        resp.Body.Close()
        
        if resp.StatusCode == http.StatusOK {
            status.Healthy = true
            return status
        }
        
        lastErr = fmt.Errorf("health check failed with status: %d", resp.StatusCode)
    }
    
    status.Healthy = false
    status.Error = lastErr
    return status
}

// handleHealthStatus 处理健康状态
func (hc *HealthChecker) handleHealthStatus(status HealthStatus) {
    if !status.Healthy {
        klog.Errorf("Health check failed for %s: %v", status.Name, status.Error)
        hc.alerts.TriggerAlert(AlertEvent{
            Name:      fmt.Sprintf("%sUnhealthy", status.Name),
            Severity:  "critical",
            Message:   fmt.Sprintf("Health check failed for %s: %v", status.Name, status.Error),
            Timestamp: status.LastCheck,
        })
    } else {
        klog.V(2).Infof("Health check passed for %s", status.Name)
    }
}

// AlertManager 告警管理器
type AlertManager struct {
    // 实现告警逻辑
}

// AlertEvent 告警事件
type AlertEvent struct {
    Name      string
    Severity  string
    Message   string
    Timestamp time.Time
}

// NewAlertManager 创建告警管理器
func NewAlertManager() AlertManager {
    return AlertManager{}
}

// TriggerAlert 触发告警
func (am *AlertManager) TriggerAlert(event AlertEvent) {
    // 实现告警逻辑，如发送到 Prometheus Alertmanager
    klog.Warningf("Alert triggered: %s - %s", event.Name, event.Message)
}
```

#### 5.3.2 自动故障恢复

**故障恢复策略配置：**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-recovery-config
  namespace: kube-system
data:
  recovery-policy.yaml: |
    recoveryPolicies:
      - name: scheduler-restart
        trigger:
          - condition: "scheduler-api == false"
            duration: 30s
        actions:
          - type: restart-pod
            target: "kube-scheduler"
            maxRetries: 3
            backoff: exponential
      
      - name: leader-election-recovery
        trigger:
          - condition: "leader-election == false"
            duration: 60s
        actions:
          - type: force-leader-election
            target: "kube-scheduler"
      
      - name: stuck-pods-recovery
        trigger:
          - condition: "pending-pods > 100"
            duration: 300s
        actions:
          - type: reschedule-pods
            filter: "status.phase == Pending"
            maxPods: 50
    
    escalation:
      - level: 1
        duration: 300s
        actions: ["restart-pod"]
      - level: 2
        duration: 600s
        actions: ["restart-pod", "force-leader-election"]
      - level: 3
        duration: 1200s
        actions: ["restart-pod", "force-leader-election", "reschedule-pods"]
```

**自动恢复实现：**

```go
package recovery

import (
    "context"
    "fmt"
    "time"
    
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

// RecoveryManager 故障恢复管理器
type RecoveryManager struct {
    client    kubernetes.Interface
    policies  []RecoveryPolicy
    escalator *EscalationManager
}

// RecoveryPolicy 恢复策略
type RecoveryPolicy struct {
    Name     string
    Triggers []Trigger
    Actions  []Action
}

// Trigger 触发条件
type Trigger struct {
    Condition string
    Duration  time.Duration
}

// Action 恢复动作
type Action struct {
    Type       string
    Target     string
    MaxRetries int
    Backoff    string
    Filter     string
    MaxPods    int
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager(client kubernetes.Interface, policies []RecoveryPolicy) *RecoveryManager {
    return &RecoveryManager{
        client:    client,
        policies:  policies,
        escalator: NewEscalationManager(),
    }
}

// Start 启动恢复管理器
func (rm *RecoveryManager) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            rm.checkAndRecover(ctx)
        }
    }
}

// checkAndRecover 检查并执行恢复
func (rm *RecoveryManager) checkAndRecover(ctx context.Context) {
    for _, policy := range rm.policies {
        if rm.shouldTriggerRecovery(ctx, policy) {
            klog.Infof("Triggering recovery policy: %s", policy.Name)
            rm.executeRecovery(ctx, policy)
        }
    }
}

// shouldTriggerRecovery 检查是否应该触发恢复
func (rm *RecoveryManager) shouldTriggerRecovery(ctx context.Context, policy RecoveryPolicy) bool {
    for _, trigger := range policy.Triggers {
        if rm.evaluateCondition(ctx, trigger.Condition) {
            // 检查条件持续时间
            if rm.escalator.ShouldEscalate(policy.Name, trigger.Duration) {
                return true
            }
        }
    }
    return false
}

// evaluateCondition 评估触发条件
func (rm *RecoveryManager) evaluateCondition(ctx context.Context, condition string) bool {
    switch condition {
    case "scheduler-api == false":
        return !rm.checkSchedulerAPI(ctx)
    case "leader-election == false":
        return !rm.checkLeaderElection(ctx)
    case "pending-pods > 100":
        return rm.getPendingPodsCount(ctx) > 100
    default:
        return false
    }
}

// executeRecovery 执行恢复动作
func (rm *RecoveryManager) executeRecovery(ctx context.Context, policy RecoveryPolicy) {
    for _, action := range policy.Actions {
        if err := rm.executeAction(ctx, action); err != nil {
            klog.Errorf("Failed to execute recovery action %s: %v", action.Type, err)
        } else {
            klog.Infof("Successfully executed recovery action: %s", action.Type)
        }
    }
}

// executeAction 执行单个恢复动作
func (rm *RecoveryManager) executeAction(ctx context.Context, action Action) error {
    switch action.Type {
    case "restart-pod":
        return rm.restartPod(ctx, action.Target)
    case "force-leader-election":
        return rm.forceLeaderElection(ctx, action.Target)
    case "reschedule-pods":
        return rm.reschedulePods(ctx, action.Filter, action.MaxPods)
    default:
        return fmt.Errorf("unknown action type: %s", action.Type)
    }
}

// restartPod 重启 Pod
func (rm *RecoveryManager) restartPod(ctx context.Context, target string) error {
    // 获取调度器 Pod
    pods, err := rm.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
        LabelSelector: fmt.Sprintf("component=%s", target),
    })
    if err != nil {
        return fmt.Errorf("failed to list pods: %v", err)
    }
    
    if len(pods.Items) == 0 {
        return fmt.Errorf("no pods found for target: %s", target)
    }
    
    // 删除 Pod 以触发重启
    for _, pod := range pods.Items {
        err := rm.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
        if err != nil {
            klog.Errorf("Failed to delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
        } else {
            klog.Infof("Deleted pod %s/%s for restart", pod.Namespace, pod.Name)
        }
    }
    
    return nil
}

// forceLeaderElection 强制重新选举
func (rm *RecoveryManager) forceLeaderElection(ctx context.Context, target string) error {
    // 删除 leader election lease 以触发重新选举
    err := rm.client.CoordinationV1().Leases("kube-system").Delete(ctx, target, metav1.DeleteOptions{})
    if err != nil {
        return fmt.Errorf("failed to delete leader election lease: %v", err)
    }
    
    klog.Infof("Deleted leader election lease for %s", target)
    return nil
}

// reschedulePods 重新调度 Pod
func (rm *RecoveryManager) reschedulePods(ctx context.Context, filter string, maxPods int) error {
    // 获取待调度的 Pod
    pods, err := rm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
        FieldSelector: "status.phase=Pending",
    })
    if err != nil {
        return fmt.Errorf("failed to list pending pods: %v", err)
    }
    
    count := 0
    for _, pod := range pods.Items {
        if count >= maxPods {
            break
        }
        
        // 检查 Pod 是否卡住（超过 5 分钟未调度）
        if time.Since(pod.CreationTimestamp.Time) > 5*time.Minute {
            // 删除并重新创建 Pod
            if err := rm.recreatePod(ctx, &pod); err != nil {
                klog.Errorf("Failed to recreate pod %s/%s: %v", pod.Namespace, pod.Name, err)
            } else {
                count++
                klog.Infof("Recreated stuck pod %s/%s", pod.Namespace, pod.Name)
            }
        }
    }
    
    return nil
}

// recreatePod 重新创建 Pod
func (rm *RecoveryManager) recreatePod(ctx context.Context, pod *corev1.Pod) error {
    // 创建新的 Pod 对象
    newPod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:        pod.Name + "-recovered",
            Namespace:   pod.Namespace,
            Labels:      pod.Labels,
            Annotations: pod.Annotations,
        },
        Spec: pod.Spec,
    }
    
    // 清除调度相关字段
    newPod.Spec.NodeName = ""
    newPod.Spec.SchedulerName = ""
    
    // 先创建新 Pod，避免服务中断
    createdPod, err := rm.client.CoreV1().Pods(newPod.Namespace).Create(ctx, newPod, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create new pod: %v", err)
    }
    
    // 等待新 Pod 开始调度
    time.Sleep(5 * time.Second)
    
    // 删除原 Pod
    err = rm.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
    if err != nil {
        // 如果删除失败，尝试删除新创建的 Pod 以避免重复
        rm.client.CoreV1().Pods(createdPod.Namespace).Delete(ctx, createdPod.Name, metav1.DeleteOptions{})
        return fmt.Errorf("failed to delete original pod: %v", err)
    }
    
    return nil
}

// 辅助方法
func (rm *RecoveryManager) checkSchedulerAPI(ctx context.Context) bool {
    // 实现调度器 API 健康检查
    return true
}

func (rm *RecoveryManager) checkLeaderElection(ctx context.Context) bool {
    // 实现 leader election 检查
    return true
}

func (rm *RecoveryManager) getPendingPodsCount(ctx context.Context) int {
    pods, err := rm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
        FieldSelector: "status.phase=Pending",
    })
    if err != nil {
        return 0
    }
    return len(pods.Items)
}

// EscalationManager 升级管理器
type EscalationManager struct {
    incidents map[string]*Incident
}

// Incident 事件记录
type Incident struct {
    Name      string
    StartTime time.Time
    Level     int
}

// NewEscalationManager 创建升级管理器
func NewEscalationManager() *EscalationManager {
    return &EscalationManager{
        incidents: make(map[string]*Incident),
    }
}

// ShouldEscalate 检查是否应该升级
func (em *EscalationManager) ShouldEscalate(name string, duration time.Duration) bool {
    incident, exists := em.incidents[name]
    if !exists {
        em.incidents[name] = &Incident{
            Name:      name,
            StartTime: time.Now(),
            Level:     1,
        }
        return true
    }
    
    if time.Since(incident.StartTime) > duration {
        incident.Level++
        return true
    }
    
    return false
}
```

#### 5.3.3 监控和告警集成

**Prometheus 监控配置：**

```yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: kube-scheduler-monitor
  namespace: kube-system
  labels:
    app: kube-scheduler
spec:
  selector:
    matchLabels:
      component: kube-scheduler
  endpoints:
  - port: https-metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
    interval: 30s
    path: /metrics
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: kube-scheduler-alerts
  namespace: kube-system
spec:
  groups:
  - name: kube-scheduler
    rules:
    - alert: KubeSchedulerDown
      expr: up{job="kube-scheduler"} == 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Kube-scheduler is down"
        description: "Kube-scheduler has been down for more than 5 minutes."
    
    - alert: KubeSchedulerHighLatency
      expr: histogram_quantile(0.99, rate(scheduler_scheduling_duration_seconds_bucket[5m])) > 1
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "High scheduling latency"
        description: "99th percentile scheduling latency is {{ $value }}s"
    
    - alert: KubeSchedulerPendingPods
      expr: scheduler_pending_pods > 100
      for: 15m
      labels:
        severity: warning
      annotations:
        summary: "Too many pending pods"
        description: "There are {{ $value }} pending pods for more than 15 minutes"
    
    - alert: KubeSchedulerFailedScheduling
      expr: rate(scheduler_schedule_attempts_total{result="error"}[5m]) > 0.1
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High scheduling failure rate"
        description: "Scheduling failure rate is {{ $value }} per second"
```

**Grafana 仪表板配置：**

```json
{
  "dashboard": {
    "id": null,
    "title": "Kubernetes Scheduler Monitoring",
    "tags": ["kubernetes", "scheduler"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Scheduling Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(scheduler_scheduling_duration_seconds_bucket[5m]))",
            "legendFormat": "99th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(scheduler_scheduling_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(scheduler_scheduling_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "yAxes": [
          {
            "label": "Latency (seconds)",
            "min": 0
          }
        ]
      },
      {
        "id": 2,
        "title": "Scheduling Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(scheduler_schedule_attempts_total[5m])",
            "legendFormat": "Scheduling attempts/sec"
          }
        ]
      },
      {
        "id": 3,
        "title": "Pending Pods",
        "type": "singlestat",
        "targets": [
          {
            "expr": "scheduler_pending_pods",
            "legendFormat": "Pending Pods"
          }
        ]
      },
      {
        "id": 4,
        "title": "Scheduling Success Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(scheduler_schedule_attempts_total{result=\"scheduled\"}[5m]) / rate(scheduler_schedule_attempts_total[5m])",
            "legendFormat": "Success Rate"
          }
        ],
        "yAxes": [
          {
            "min": 0,
            "max": 1,
            "unit": "percentunit"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

## 6. 高级调度特性

### 6.1 优先级和抢占

#### 6.1.1 优先级类配置

**PriorityClass 定义：**

```yaml
# 高优先级类
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000
globalDefault: false
description: "High priority class for critical workloads"
---
# 中优先级类
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: medium-priority
value: 500
globalDefault: false
description: "Medium priority class for important workloads"
---
# 低优先级类
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: low-priority
value: 100
globalDefault: true
description: "Low priority class for batch workloads"
---
# 系统级优先级类
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: system-critical
value: 2000
globalDefault: false
description: "System critical priority class"
```

**Pod 优先级配置示例：**

```yaml
# 高优先级 Web 应用
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app-high-priority
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
      priority: high
  template:
    metadata:
      labels:
        app: web-app
        priority: high
    spec:
      priorityClassName: high-priority
      containers:
      - name: web
        image: nginx:1.20
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1
            memory: 1Gi
---
# 低优先级批处理任务
apiVersion: batch/v1
kind: Job
metadata:
  name: batch-job-low-priority
  namespace: batch
spec:
  template:
    spec:
      priorityClassName: low-priority
      restartPolicy: Never
      containers:
      - name: batch-worker
        image: busybox:1.35
        command: ["sh", "-c", "echo 'Processing batch job...' && sleep 3600"]
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
```

#### 6.1.2 抢占机制实现

**抢占策略配置：**

```yaml
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
metadata:
  name: preemption-scheduler
profiles:
- schedulerName: preemption-scheduler
  plugins:
    preFilter:
      enabled:
      - name: NodeResourcesFit
      - name: NodeAffinity
    filter:
      enabled:
      - name: NodeResourcesFit
      - name: NodeAffinity
      - name: PodTopologySpread
    postFilter:
      enabled:
      - name: DefaultPreemption  # 启用抢占
    score:
      enabled:
      - name: NodeResourcesFit
      - name: NodeAffinity
      - name: PodTopologySpread
  pluginConfig:
  - name: DefaultPreemption
    args:
      minCandidateNodesPercentage: 10
      minCandidateNodesAbsolute: 100
```

**自定义抢占插件：**

```go
package preemption

import (
    "context"
    "fmt"
    "sort"
    
    corev1 "k8s.io/api/core/v1"
    policyv1 "k8s.io/api/policy/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/klog/v2"
    "k8s.io/kubernetes/pkg/scheduler/framework"
)

// CustomPreemption 自定义抢占插件
type CustomPreemption struct {
    handle framework.Handle
}

// Name 返回插件名称
func (cp *CustomPreemption) Name() string {
    return "CustomPreemption"
}

// New 创建插件实例
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
    return &CustomPreemption{handle: h}, nil
}

// PostFilter 实现抢占逻辑
func (cp *CustomPreemption) PostFilter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
    klog.V(2).Infof("Starting preemption for pod %s/%s", pod.Namespace, pod.Name)
    
    // 获取所有节点
    allNodes, err := cp.handle.SnapshotSharedLister().NodeInfos().List()
    if err != nil {
        return nil, framework.NewStatus(framework.Error, fmt.Sprintf("failed to list nodes: %v", err))
    }
    
    // 查找可以抢占的节点
    candidates := cp.findPreemptionCandidates(ctx, pod, allNodes)
    if len(candidates) == 0 {
        return nil, framework.NewStatus(framework.Unschedulable, "no preemption candidates found")
    }
    
    // 选择最佳抢占候选
    bestCandidate := cp.selectBestCandidate(candidates)
    
    // 执行抢占
    if err := cp.executePreemption(ctx, bestCandidate); err != nil {
        return nil, framework.NewStatus(framework.Error, fmt.Sprintf("failed to execute preemption: %v", err))
    }
    
    klog.Infof("Successfully preempted pods on node %s for pod %s/%s", bestCandidate.Node.Name, pod.Namespace, pod.Name)
    
    return &framework.PostFilterResult{
        NominatedNodeName: bestCandidate.Node.Name,
    }, nil
}

// PreemptionCandidate 抢占候选
type PreemptionCandidate struct {
    Node         *framework.NodeInfo
    VictimsToEvict []*corev1.Pod
    Score        int64
}

// findPreemptionCandidates 查找抢占候选
func (cp *CustomPreemption) findPreemptionCandidates(ctx context.Context, pod *corev1.Pod, nodes []*framework.NodeInfo) []PreemptionCandidate {
    var candidates []PreemptionCandidate
    
    for _, node := range nodes {
        victims := cp.findVictims(pod, node)
        if len(victims) > 0 {
            candidate := PreemptionCandidate{
                Node:           node,
                VictimsToEvict: victims,
                Score:          cp.calculatePreemptionScore(pod, node, victims),
            }
            candidates = append(candidates, candidate)
        }
    }
    
    return candidates
}

// findVictims 查找可以被抢占的 Pod
func (cp *CustomPreemption) findVictims(preemptor *corev1.Pod, node *framework.NodeInfo) []*corev1.Pod {
    var victims []*corev1.Pod
    preemptorPriority := getPodPriority(preemptor)
    
    // 按优先级排序节点上的 Pod
    pods := make([]*corev1.Pod, 0, len(node.Pods))
    for _, podInfo := range node.Pods {
        pods = append(pods, podInfo.Pod)
    }
    
    sort.Slice(pods, func(i, j int) bool {
        return getPodPriority(pods[i]) < getPodPriority(pods[j])
    })
    
    // 计算需要释放的资源
    requiredCPU := getResourceRequest(preemptor, corev1.ResourceCPU)
    requiredMemory := getResourceRequest(preemptor, corev1.ResourceMemory)
    
    var releasedCPU, releasedMemory int64
    
    for _, pod := range pods {
        podPriority := getPodPriority(pod)
        
        // 只能抢占优先级更低的 Pod
        if podPriority >= preemptorPriority {
            continue
        }
        
        // 跳过系统 Pod
        if isSystemPod(pod) {
            continue
        }
        
        victims = append(victims, pod)
        releasedCPU += getResourceRequest(pod, corev1.ResourceCPU)
        releasedMemory += getResourceRequest(pod, corev1.ResourceMemory)
        
        // 检查是否释放了足够的资源
        if releasedCPU >= requiredCPU && releasedMemory >= requiredMemory {
            break
        }
    }
    
    // 如果释放的资源不足，返回空
    if releasedCPU < requiredCPU || releasedMemory < requiredMemory {
        return nil
    }
    
    return victims
}

// calculatePreemptionScore 计算抢占分数
func (cp *CustomPreemption) calculatePreemptionScore(preemptor *corev1.Pod, node *framework.NodeInfo, victims []*corev1.Pod) int64 {
    var score int64
    
    // 优先选择抢占 Pod 数量少的节点
    score -= int64(len(victims)) * 10
    
    // 优先选择抢占优先级差距大的 Pod
    preemptorPriority := getPodPriority(preemptor)
    for _, victim := range victims {
        priorityDiff := preemptorPriority - getPodPriority(victim)
        score += int64(priorityDiff)
    }
    
    // 考虑节点资源利用率
    utilization := calculateNodeUtilization(node)
    score += int64((1.0 - utilization) * 100)
    
    return score
}

// selectBestCandidate 选择最佳抢占候选
func (cp *CustomPreemption) selectBestCandidate(candidates []PreemptionCandidate) PreemptionCandidate {
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Score > candidates[j].Score
    })
    
    return candidates[0]
}

// executePreemption 执行抢占
func (cp *CustomPreemption) executePreemption(ctx context.Context, candidate PreemptionCandidate) error {
    for _, victim := range candidate.VictimsToEvict {
        if err := cp.evictPod(ctx, victim); err != nil {
            klog.Errorf("Failed to evict pod %s/%s: %v", victim.Namespace, victim.Name, err)
            return err
        }
        klog.Infof("Evicted pod %s/%s for preemption", victim.Namespace, victim.Name)
    }
    
    return nil
}

// evictPod 驱逐 Pod
func (cp *CustomPreemption) evictPod(ctx context.Context, pod *corev1.Pod) error {
    // 使用 Eviction API 优雅地驱逐 Pod
    eviction := &policyv1.Eviction{
        ObjectMeta: metav1.ObjectMeta{
            Name:      pod.Name,
            Namespace: pod.Namespace,
        },
    }
    
    return cp.handle.ClientSet().PolicyV1().Evictions(pod.Namespace).Evict(ctx, eviction)
}

// 辅助函数
func getPodPriority(pod *corev1.Pod) int32 {
    if pod.Spec.Priority != nil {
        return *pod.Spec.Priority
    }
    return 0
}

func getResourceRequest(pod *corev1.Pod, resource corev1.ResourceName) int64 {
    var total int64
    for _, container := range pod.Spec.Containers {
        if req := container.Resources.Requests[resource]; !req.IsZero() {
            if resource == corev1.ResourceCPU {
                total += req.MilliValue()
            } else {
                total += req.Value()
            }
        }
    }
    return total
}

func isSystemPod(pod *corev1.Pod) bool {
    // 检查是否为系统 Pod
    if pod.Namespace == "kube-system" {
        return true
    }
    
    // 检查是否有系统相关的标签
    if pod.Labels != nil {
        if tier, exists := pod.Labels["tier"]; exists && tier == "control-plane" {
            return true
        }
        if component, exists := pod.Labels["component"]; exists {
            systemComponents := []string{"kube-scheduler", "kube-controller-manager", "kube-apiserver", "etcd"}
            for _, sc := range systemComponents {
                if component == sc {
                    return true
                }
            }
        }
    }
    
    return false
}

func calculateNodeUtilization(node *framework.NodeInfo) float64 {
    if node.Allocatable == nil {
        return 0.0
    }
    
    allocatableCPU := node.Allocatable.MilliCPU
    allocatableMemory := node.Allocatable.Memory
    
    var requestedCPU, requestedMemory int64
    for _, podInfo := range node.Pods {
        requestedCPU += getResourceRequest(podInfo.Pod, corev1.ResourceCPU)
        requestedMemory += getResourceRequest(podInfo.Pod, corev1.ResourceMemory)
    }
    
    cpuUtilization := float64(requestedCPU) / float64(allocatableCPU)
    memoryUtilization := float64(requestedMemory) / float64(allocatableMemory)
    
    // 返回 CPU 和内存利用率的平均值
    return (cpuUtilization + memoryUtilization) / 2.0
}
```

### 6.2 资源配额和限制

#### 6.2.1 命名空间资源配额

**ResourceQuota 配置：**

```yaml
# 生产环境资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: production-quota
  namespace: production
spec:
  hard:
    # 计算资源限制
    requests.cpu: "100"
    requests.memory: 200Gi
    limits.cpu: "200"
    limits.memory: 400Gi
    
    # 存储资源限制
    requests.storage: 1Ti
    persistentvolumeclaims: "50"
    
    # 对象数量限制
    pods: "100"
    replicationcontrollers: "20"
    secrets: "50"
    configmaps: "50"
    services: "20"
    services.loadbalancers: "5"
    services.nodeports: "10"
    
    # 扩展资源限制
    requests.nvidia.com/gpu: "10"
    limits.nvidia.com/gpu: "10"
---
# 开发环境资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: development-quota
  namespace: development
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    requests.storage: 200Gi
    persistentvolumeclaims: "20"
    pods: "50"
    services: "10"
---
# 测试环境资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: testing-quota
  namespace: testing
spec:
  hard:
    requests.cpu: "30"
    requests.memory: 60Gi
    limits.cpu: "60"
    limits.memory: 120Gi
    requests.storage: 300Gi
    persistentvolumeclaims: "30"
    pods: "75"
    services: "15"
```

**LimitRange 配置：**

```yaml
# Pod 资源限制范围
apiVersion: v1
kind: LimitRange
metadata:
  name: pod-limit-range
  namespace: production
spec:
  limits:
  # Pod 级别限制
  - type: Pod
    max:
      cpu: "8"
      memory: 16Gi
    min:
      cpu: 100m
      memory: 128Mi
  # Container 级别限制
  - type: Container
    default:
      cpu: 500m
      memory: 512Mi
    defaultRequest:
      cpu: 100m
      memory: 128Mi
    max:
      cpu: "4"
      memory: 8Gi
    min:
      cpu: 50m
      memory: 64Mi
  # PVC 存储限制
  - type: PersistentVolumeClaim
    max:
      storage: 100Gi
    min:
      storage: 1Gi
---
# 开发环境限制范围
apiVersion: v1
kind: LimitRange
metadata:
  name: dev-limit-range
  namespace: development
spec:
  limits:
  - type: Pod
    max:
      cpu: "2"
      memory: 4Gi
    min:
      cpu: 50m
      memory: 64Mi
  - type: Container
    default:
      cpu: 200m
      memory: 256Mi
    defaultRequest:
      cpu: 50m
      memory: 64Mi
    max:
      cpu: "1"
      memory: 2Gi
    min:
      cpu: 25m
      memory: 32Mi
```

#### 6.2.2 多租户资源管理

**租户资源管理器：**

```go
package quota

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

// TenantResourceManager 租户资源管理器
type TenantResourceManager struct {
    clientset kubernetes.Interface
    
    // 租户配置
    tenants map[string]*TenantConfig
    mu      sync.RWMutex
    
    // 监控指标
    metrics *TenantMetrics
}

// TenantConfig 租户配置
type TenantConfig struct {
    Name        string
    Namespaces  []string
    Quotas      map[string]TenantQuota
    Priorities  map[string]int32
    Policies    TenantPolicies
}

// TenantQuota 租户配额
type TenantQuota struct {
    CPU     resource.Quantity
    Memory  resource.Quantity
    Storage resource.Quantity
    GPU     int64
    Pods    int64
}

// TenantPolicies 租户策略
type TenantPolicies struct {
    AllowOvercommit bool
    MaxBurstRatio   float64
    PreemptionPolicy string
    SchedulingPolicy string
}

// TenantMetrics 租户指标
type TenantMetrics struct {
    Usage     map[string]TenantUsage
    Violations map[string][]QuotaViolation
    mu        sync.RWMutex
}

// TenantUsage 租户使用情况
type TenantUsage struct {
    CPU     resource.Quantity
    Memory  resource.Quantity
    Storage resource.Quantity
    GPU     int64
    Pods    int64
    
    LastUpdated time.Time
}

// QuotaViolation 配额违规
type QuotaViolation struct {
    Tenant      string
    Namespace   string
    Resource    string
    Requested   resource.Quantity
    Available   resource.Quantity
    Timestamp   time.Time
}

// NewTenantResourceManager 创建租户资源管理器
func NewTenantResourceManager(clientset kubernetes.Interface) *TenantResourceManager {
    return &TenantResourceManager{
        clientset: clientset,
        tenants:   make(map[string]*TenantConfig),
        metrics: &TenantMetrics{
            Usage:      make(map[string]TenantUsage),
            Violations: make(map[string][]QuotaViolation),
        },
    }
}

// RegisterTenant 注册租户
func (trm *TenantResourceManager) RegisterTenant(config *TenantConfig) error {
    trm.mu.Lock()
    defer trm.mu.Unlock()
    
    // 验证租户配置
    if err := trm.validateTenantConfig(config); err != nil {
        return fmt.Errorf("invalid tenant config: %v", err)
    }
    
    trm.tenants[config.Name] = config
    
    // 为租户创建资源配额
    if err := trm.createTenantQuotas(config); err != nil {
        return fmt.Errorf("failed to create tenant quotas: %v", err)
    }
    
    klog.Infof("Registered tenant %s with %d namespaces", config.Name, len(config.Namespaces))
    return nil
}

// CheckResourceRequest 检查资源请求
func (trm *TenantResourceManager) CheckResourceRequest(ctx context.Context, tenant, namespace string, request corev1.ResourceList) (*ResourceAllocation, error) {
    trm.mu.RLock()
    tenantConfig, exists := trm.tenants[tenant]
    trm.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("tenant %s not found", tenant)
    }
    
    // 检查命名空间是否属于租户
    if !trm.isNamespaceInTenant(namespace, tenantConfig) {
        return nil, fmt.Errorf("namespace %s does not belong to tenant %s", namespace, tenant)
    }
    
    // 获取当前使用情况
    usage, err := trm.getTenantUsage(ctx, tenant)
    if err != nil {
        return nil, fmt.Errorf("failed to get tenant usage: %v", err)
    }
    
    // 检查配额限制
    allocation, err := trm.checkQuotaLimits(tenantConfig, usage, request)
    if err != nil {
        // 记录配额违规
        trm.recordQuotaViolation(tenant, namespace, request, err)
        return nil, err
    }
    
    return allocation, nil
}

// ResourceAllocation 资源分配结果
type ResourceAllocation struct {
    Approved    bool
    Resources   corev1.ResourceList
    Priority    int32
    Constraints map[string]string
}

// validateTenantConfig 验证租户配置
func (trm *TenantResourceManager) validateTenantConfig(config *TenantConfig) error {
    if config.Name == "" {
        return fmt.Errorf("tenant name cannot be empty")
    }
    
    if len(config.Namespaces) == 0 {
        return fmt.Errorf("tenant must have at least one namespace")
    }
    
    // 验证配额配置
    for env, quota := range config.Quotas {
        if quota.CPU.IsZero() || quota.Memory.IsZero() {
            return fmt.Errorf("invalid quota for environment %s: CPU and Memory must be specified", env)
        }
    }
    
    return nil
}

// createTenantQuotas 为租户创建资源配额
func (trm *TenantResourceManager) createTenantQuotas(config *TenantConfig) error {
    for _, namespace := range config.Namespaces {
        // 确定环境类型
        env := trm.getEnvironmentType(namespace)
        quota, exists := config.Quotas[env]
        if !exists {
            quota = config.Quotas["default"]
        }
        
        // 创建 ResourceQuota
        resourceQuota := &corev1.ResourceQuota{
            ObjectMeta: metav1.ObjectMeta{
                Name:      fmt.Sprintf("%s-quota", config.Name),
                Namespace: namespace,
                Labels: map[string]string{
                    "tenant":      config.Name,
                    "environment": env,
                },
            },
            Spec: corev1.ResourceQuotaSpec{
                Hard: corev1.ResourceList{
                    corev1.ResourceRequestsCPU:    quota.CPU,
                    corev1.ResourceRequestsMemory: quota.Memory,
                    corev1.ResourceRequestsStorage: quota.Storage,
                    corev1.ResourcePods:           *resource.NewQuantity(quota.Pods, resource.DecimalSI),
                },
            },
        }
        
        // 添加 GPU 配额（如果有）
        if quota.GPU > 0 {
            resourceQuota.Spec.Hard["requests.nvidia.com/gpu"] = *resource.NewQuantity(quota.GPU, resource.DecimalSI)
        }
        
        _, err := trm.clientset.CoreV1().ResourceQuotas(namespace).Create(context.TODO(), resourceQuota, metav1.CreateOptions{})
        if err != nil {
            return fmt.Errorf("failed to create resource quota for namespace %s: %v", namespace, err)
        }
    }
    
    return nil
}

// getTenantUsage 获取租户使用情况
func (trm *TenantResourceManager) getTenantUsage(ctx context.Context, tenant string) (TenantUsage, error) {
    trm.mu.RLock()
    tenantConfig := trm.tenants[tenant]
    trm.mu.RUnlock()
    
    var totalUsage TenantUsage
    
    for _, namespace := range tenantConfig.Namespaces {
        // 获取命名空间中的所有 Pod
        pods, err := trm.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
            FieldSelector: "status.phase=Running",
        })
        if err != nil {
            return totalUsage, fmt.Errorf("failed to list pods in namespace %s: %v", namespace, err)
        }
        
        // 计算资源使用量
        for _, pod := range pods.Items {
            for _, container := range pod.Spec.Containers {
                if cpu := container.Resources.Requests[corev1.ResourceCPU]; !cpu.IsZero() {
                    totalUsage.CPU.Add(cpu)
                }
                if memory := container.Resources.Requests[corev1.ResourceMemory]; !memory.IsZero() {
                    totalUsage.Memory.Add(memory)
                }
            }
        }
        
        totalUsage.Pods += int64(len(pods.Items))
    }
    
    totalUsage.LastUpdated = time.Now()
    
    // 更新缓存
    trm.metrics.mu.Lock()
    trm.metrics.Usage[tenant] = totalUsage
    trm.metrics.mu.Unlock()
    
    return totalUsage, nil
}

// checkQuotaLimits 检查配额限制
func (trm *TenantResourceManager) checkQuotaLimits(config *TenantConfig, usage TenantUsage, request corev1.ResourceList) (*ResourceAllocation, error) {
    // 获取适用的配额
    quota := config.Quotas["default"]
    
    // 计算有效配额（考虑突发使用策略）
    effectiveCPUQuota := quota.CPU.DeepCopy()
    effectiveMemoryQuota := quota.Memory.DeepCopy()
    
    if config.Policies.AllowOvercommit && config.Policies.MaxBurstRatio > 1.0 {
        // 允许突发使用，增加有效配额
        burstCPU := quota.CPU.DeepCopy()
        burstCPU.Set(int64(float64(burstCPU.MilliValue()) * config.Policies.MaxBurstRatio))
        effectiveCPUQuota = burstCPU
        
        burstMemory := quota.Memory.DeepCopy()
        burstMemory.Set(int64(float64(burstMemory.Value()) * config.Policies.MaxBurstRatio))
        effectiveMemoryQuota = burstMemory
    }
    
    // 预留10%资源用于系统开销
    reservedCPU := effectiveCPUQuota.DeepCopy()
    reservedCPU.Set(int64(float64(reservedCPU.MilliValue()) * 0.9))
    
    reservedMemory := effectiveMemoryQuota.DeepCopy()
    reservedMemory.Set(int64(float64(reservedMemory.Value()) * 0.9))
    
    // 检查 CPU 配额
    if cpu := request[corev1.ResourceCPU]; !cpu.IsZero() {
        newCPUUsage := usage.CPU.DeepCopy()
        newCPUUsage.Add(cpu)
        if newCPUUsage.Cmp(reservedCPU) > 0 {
            return nil, fmt.Errorf("CPU quota exceeded: requested %v, available %v (reserved quota)", cpu, reservedCPU.DeepCopy().Sub(usage.CPU))
        }
    }
    
    // 检查内存配额
    if memory := request[corev1.ResourceMemory]; !memory.IsZero() {
        newMemoryUsage := usage.Memory.DeepCopy()
        newMemoryUsage.Add(memory)
        if newMemoryUsage.Cmp(reservedMemory) > 0 {
            return nil, fmt.Errorf("Memory quota exceeded: requested %v, available %v (reserved quota)", memory, reservedMemory.DeepCopy().Sub(usage.Memory))
        }
    }
    
    // 检查 Pod 数量配额
    if usage.Pods >= quota.Pods {
        return nil, fmt.Errorf("Pod quota exceeded: current %d, limit %d", usage.Pods, quota.Pods)
    }
    
    // 创建资源分配
    allocation := &ResourceAllocation{
        Approved:  true,
        Resources: request,
        Priority:  config.Priorities["default"],
        Constraints: map[string]string{
            "tenant": config.Name,
            "burst-allowed": fmt.Sprintf("%t", config.Policies.AllowOvercommit),
        },
    }
    
    return allocation, nil
}

// recordQuotaViolation 记录配额违规
func (trm *TenantResourceManager) recordQuotaViolation(tenant, namespace string, request corev1.ResourceList, err error) {
    violation := QuotaViolation{
        Tenant:    tenant,
        Namespace: namespace,
        Timestamp: time.Now(),
    }
    
    trm.metrics.mu.Lock()
    trm.metrics.Violations[tenant] = append(trm.metrics.Violations[tenant], violation)
    trm.metrics.mu.Unlock()
    
    klog.Warningf("Quota violation for tenant %s in namespace %s: %v", tenant, namespace, err)
}

// 辅助函数
func (trm *TenantResourceManager) isNamespaceInTenant(namespace string, config *TenantConfig) bool {
    for _, ns := range config.Namespaces {
        if ns == namespace {
            return true
        }
    }
    return false
}

func (trm *TenantResourceManager) getEnvironmentType(namespace string) string {
    if contains(namespace, "prod") {
        return "production"
    } else if contains(namespace, "dev") {
        return "development"
    } else if contains(namespace, "test") {
        return "testing"
    }
    return "default"
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || 
           (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// GetTenantMetrics 获取租户指标
func (trm *TenantResourceManager) GetTenantMetrics(tenant string) (*TenantMetrics, error) {
    trm.metrics.mu.RLock()
    defer trm.metrics.mu.RUnlock()
    
    usage, exists := trm.metrics.Usage[tenant]
    if !exists {
        return nil, fmt.Errorf("no metrics found for tenant %s", tenant)
    }
    
    violations := trm.metrics.Violations[tenant]
    
    return &TenantMetrics{
        Usage: map[string]TenantUsage{
            tenant: usage,
        },
        Violations: map[string][]QuotaViolation{
            tenant: violations,
        },
    }, nil
}
```
