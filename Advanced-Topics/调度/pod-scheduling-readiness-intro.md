# Pod Scheduling Readiness 简介

## 1. 背景

### 1.1 问题描述

在 Kubernetes 集群中，当 Pod 被创建时，调度器会不断尝试寻找适合它的节点。这个无限循环一直持续到调度程序为 Pod 找到节点，或者 Pod 被删除。然而，长时间无法被调度的 Pod（例如，被某些外部事件阻塞的 Pod）会浪费调度周期。

一个调度周期可能需要约 20ms 或更长时间，这取决于 Pod 的调度约束的复杂度。因此，大量浪费的调度周期会严重影响调度器的性能。

### 1.2 解决方案

Kubernetes 1.26 引入了 Pod Scheduling Readiness 特性，通过调度门控（Scheduling Gates）机制来解决这个问题。调度门控允许声明新创建的 Pod 尚未准备好进行调度。当 Pod 上设置了调度门控时，调度程序会忽略该 Pod，从而避免不必要的调度尝试。

### 1.3 特性状态

- **Kubernetes 1.26**: Alpha 版本
- **Kubernetes 1.30**: GA 版本（正式发布）

---

## 2. 使用场景

### 2.1 动态配额管理

Pod Scheduling Readiness 特性支持的一个重要使用场景是动态配额管理。传统的 Kubernetes ResourceQuota 在 API Server 层面强制执行配额，如果新的 Pod 超过了 CPU 配额，它就会被拒绝。API Server 不会对 Pod 进行排队，因此需要不断尝试重新创建 Pod。

使用调度门控，外部配额管理器可以：

1. 通过变更性质的 Webhook 为集群中创建的所有 Pod 添加调度门控
2. 当存在用于启动 Pod 的配额时，管理器移除此门控
3. 避免在资源不足时的重复调度尝试

### 2.2 外部依赖管理

当 Pod 需要等待外部资源或服务准备就绪时，可以使用调度门控暂停调度，直到依赖条件满足。

### 2.3 批处理作业控制

在批处理场景中，可以使用调度门控来控制作业的启动时机，实现更精细的作业调度策略。

---

## 3. 用法

### 3.1 基本用法

#### 3.1.1 创建带有调度门控的 Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  schedulingGates:
    - name: "example.com/quota-check"
    - name: "example.com/dependency-ready"
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.9
```

#### 3.1.2 查看 Pod 状态

```bash
kubectl get pods
```

输出示例：

```bash
NAME       READY   STATUS            RESTARTS   AGE
test-pod   0/1     SchedulingGated   0          10s
```

#### 3.1.3 移除调度门控

```bash
# 移除所有调度门控
kubectl patch pod test-pod --type='json' -p='[{"op": "remove", "path": "/spec/schedulingGates"}]'

# 或者移除特定门控
kubectl patch pod test-pod --type='json' -p='[{"op": "remove", "path": "/spec/schedulingGates/0"}]'
```

### 3.2 注意事项

1. **创建时设置**: 调度门控只能在 Pod 创建时设置，不能在 Pod 创建后添加新的门控
2. **移除操作**: 可以逐个移除门控，但只有当所有门控都移除后，调度器才会开始考虑对 Pod 进行调度
3. **唯一性**: 每个调度门控必须有唯一的名称字段

---

## 4. 代码分析

### 4.1 核心数据结构

#### 4.1.1 PodSchedulingGate 定义

**文件位置**: `staging/src/k8s.io/api/core/v1/types.go:4525-4529`

```go
// PodSchedulingGate is associated to a Pod to guard its scheduling.
type PodSchedulingGate struct {
    // Name of the scheduling gate.
    // Each scheduling gate must have a unique name field.
    Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}
```

#### 4.1.2 PodSpec 中的 SchedulingGates 字段

**文件位置**: `staging/src/k8s.io/api/core/v1/types.go` (PodSpec 结构体中)

```go
// SchedulingGates is an opaque list of values that if specified will block scheduling the pod.
// If schedulingGates is not empty, the pod will stay in the SchedulingGated state and the
// scheduler will not attempt to schedule the pod.
//
// SchedulingGates can only be set at pod creation time, and be removed only afterwards.
SchedulingGates []PodSchedulingGate `json:"schedulingGates,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,38,opt,name=schedulingGates"`
```

### 4.2 PreEnqueue 插件机制

#### 4.2.1 PreEnqueuePlugin 接口

**文件位置**: `staging/src/k8s.io/kube-scheduler/framework/interface.go:374-378`

```go
// PreEnqueuePlugin is an interface that must be implemented by "PreEnqueue" plugins.
// These plugins are called prior to adding Pods to activeQ or backoffQ.
// Note: an preEnqueue plugin is expected to be lightweight and efficient, so it's not expected to
// involve expensive calls like accessing external endpoints; otherwise it'd block other
// Pods' enqueuing in event handlers.
type PreEnqueuePlugin interface {
    Plugin
    // PreEnqueue is called prior to adding Pods to activeQ or backoffQ.
    PreEnqueue(ctx context.Context, p *v1.Pod) *Status
}
```

#### 4.2.2 SchedulingGates 插件实现

**文件位置**: `pkg/scheduler/framework/plugins/schedulinggates/scheduling_gates.go:37-53`

```go
// SchedulingGates checks if a Pod carries .spec.schedulingGates.
type SchedulingGates struct {
    enableSchedulingQueueHint bool
}

var _ fwk.PreEnqueuePlugin = &SchedulingGates{}
var _ fwk.EnqueueExtensions = &SchedulingGates{}

func (pl *SchedulingGates) PreEnqueue(ctx context.Context, p *v1.Pod) *fwk.Status {
    if len(p.Spec.SchedulingGates) == 0 {
        return nil
    }
    gates := make([]string, 0, len(p.Spec.SchedulingGates))
    for _, gate := range p.Spec.SchedulingGates {
        gates = append(gates, gate.Name)
    }
    return fwk.NewStatus(fwk.UnschedulableAndUnresolvable, fmt.Sprintf("waiting for scheduling gates: %v", gates))
}
```

### 4.3 调度队列处理逻辑

#### 4.3.1 PreEnqueue 插件执行

**文件位置**: `pkg/scheduler/backend/queue/scheduling_queue.go:566-590`

```go
// runPreEnqueuePlugins iterates PreEnqueue function in each registered PreEnqueuePlugin,
// and updates pInfo.GatingPlugin and pInfo.UnschedulablePlugins.
// Note: we need to associate the failed plugin to `pInfo`, so that the pod can be moved back
// to activeQ by related cluster event.
func (p *PriorityQueue) runPreEnqueuePlugins(ctx context.Context, pInfo *framework.QueuedPodInfo) {
    var s *fwk.Status
    pod := pInfo.Pod
    startTime := p.clock.Now()
    defer func() {
        metrics.FrameworkExtensionPointDuration.WithLabelValues(preEnqueue, s.Code().String(), pod.Spec.SchedulerName).Observe(metrics.SinceInSeconds(startTime))
    }()

    shouldRecordMetric := rand.Intn(100) < p.pluginMetricsSamplePercent
    logger := klog.FromContext(ctx)
    gatingPlugin := pInfo.GatingPlugin
    if gatingPlugin != "" {
        // Run the gating plugin first
        s := p.runPreEnqueuePlugin(ctx, logger, p.preEnqueuePluginMap[pod.Spec.SchedulerName][gatingPlugin], pInfo, shouldRecordMetric)
        if !s.IsSuccess() {
            // No need to iterate other plugins
            return
        }
    }
    // ... 省略部分代码
}
```

#### 4.3.2 队列管理

**文件位置**: `pkg/scheduler/backend/queue/scheduling_queue.go:705-713`

```go
// Add adds a pod to the active queue. It should be called only when a new pod
// is added so there is no chance the pod is already in active/unschedulable/backoff queues
func (p *PriorityQueue) Add(logger klog.Logger, pod *v1.Pod) {
    p.lock.Lock()
    defer p.lock.Unlock()

    pInfo := p.newQueuedPodInfo(pod)
    if added := p.moveToActiveQ(logger, pInfo, framework.EventUnscheduledPodAdd.Label(), false); added {
        p.activeQ.broadcast()
    }
}
```

#### 4.3.3 移动到活跃队列的逻辑

**文件位置**: `pkg/scheduler/backend/queue/scheduling_queue.go:638-675`

```go
// moveToActiveQ moves the given pod to activeQ.
// If the pod doesn't pass PreEnqueue plugins, it gets added to unschedulablePods instead.
// movesFromBackoffQ should be set to true, if the pod directly moves from the backoffQ, so the PreEnqueue call can be skipped.
// It returns a boolean flag to indicate whether the pod is added successfully.
func (p *PriorityQueue) moveToActiveQ(logger klog.Logger, pInfo *framework.QueuedPodInfo, event string, movesFromBackoffQ bool) bool {
    gatedBefore := pInfo.Gated()
    // If SchedulerPopFromBackoffQ feature gate is enabled,
    // PreEnqueue plugins were called when the pod was added to the backoffQ.
    // Don't need to repeat it here when the pod is directly moved from the backoffQ.
    skipPreEnqueue := p.isPopFromBackoffQEnabled && movesFromBackoffQ
    if !skipPreEnqueue {
        p.runPreEnqueuePlugins(context.Background(), pInfo)
    }

    added := false
    p.activeQ.underLock(func(unlockedActiveQ unlockedActiveQueuer) {
        if pInfo.Gated() {
            // Add the Pod to unschedulablePods if it's not passing PreEnqueuePlugins.
            if unlockedActiveQ.has(pInfo) {
                return
            }
            if p.backoffQ.has(pInfo) {
                return
            }
            if p.unschedulablePods.get(pInfo.Pod) != nil {
                return
            }
            p.unschedulablePods.addOrUpdate(pInfo, event)
            logger.V(5).Info("Pod moved to an internal scheduling queue, because the pod is gated", "pod", klog.KObj(pInfo.Pod), "event", event, "queue", unschedulableQ)
            return
        }
        // ... 省略部分代码
        unlockedActiveQ.add(logger, pInfo, event)
        added = true
    })
    return added
}
```

### 4.4 事件处理机制

#### 4.4.1 调度门控移除事件

**文件位置**: `pkg/scheduler/framework/plugins/schedulinggates/scheduling_gates.go:80-92`

```go
// EventsToRegister returns the possible events that may make a Pod
// failed by this plugin schedulable.
func (pl *SchedulingGates) EventsToRegister(_ context.Context) ([]fwk.ClusterEventWithHint, error) {
    if !pl.enableSchedulingQueueHint {
        return nil, nil
    }
    return []fwk.ClusterEventWithHint{
        // Pods can be more schedulable once it's gates are removed
        {Event: fwk.ClusterEvent{Resource: fwk.Pod, ActionType: fwk.UpdatePodSchedulingGatesEliminated}, QueueingHintFn: pl.isSchedulableAfterUpdatePodSchedulingGatesEliminated},
    }, nil
}
```

#### 4.4.2 队列提示函数

**文件位置**: `pkg/scheduler/framework/plugins/schedulinggates/scheduling_gates.go:94-105`

```go
// isSchedulableAfterUpdatePodSchedulingGatesEliminated 检查 Pod 在调度门控移除后是否可调度
func (pl *SchedulingGates) isSchedulableAfterUpdatePodSchedulingGatesEliminated(logger klog.Logger, pod *v1.Pod, oldObj, newObj interface{}) (fwk.QueueingHint, error) {
    _, modifiedPod, err := util.As[*v1.Pod](oldObj, newObj)
    if err != nil {
        return fwk.Queue, err
    }

    if modifiedPod.UID != pod.UID {
        // If the update event is not for targetPod, it wouldn't make targetPod schedulable.
        return fwk.QueueSkip, nil
    }

    return fwk.Queue, nil
}
```

### 4.5 特性门控

**文件位置**: `staging/src/k8s.io/apiserver/pkg/util/feature/feature_gate.go`

```go
// PodSchedulingReadiness enables the PodSchedulingReadiness feature.
// owner: @Huang-Wei
// alpha: v1.26
// beta: v1.27
// GA: v1.30
PodSchedulingReadiness featuregate.Feature = "PodSchedulingReadiness"
```

---

## 5. 工作流程

### 5.1 Pod 创建流程

1. **Pod 创建**: 用户创建带有 `schedulingGates` 字段的 Pod
2. **PreEnqueue 检查**: 调度器在将 Pod 加入调度队列前，执行 PreEnqueue 插件
3. **门控检查**: SchedulingGates 插件检查 Pod 是否有调度门控
4. **队列决策**:
   - 如果有门控：Pod 被标记为 `Gated`，加入 `unschedulablePods` 队列
   - 如果无门控：Pod 正常加入 `activeQ` 队列进行调度

### 5.2 门控移除流程

1. **门控移除**: 外部控制器移除 Pod 的调度门控
2. **事件触发**: 产生 `UpdatePodSchedulingGatesEliminated` 事件
3. **队列重新评估**: 调度器重新评估 Pod 的调度资格
4. **队列迁移**: Pod 从 `unschedulablePods` 移动到 `activeQ`
5. **开始调度**: 调度器开始为 Pod 寻找合适的节点

---

## 6. 性能优化

### 6.1 避免无效调度周期

通过 PreEnqueue 插件机制，调度器可以在调度周期开始前就过滤掉不符合条件的 Pod，避免浪费调度资源。

### 6.2 队列管理优化

- **分离队列**: 将被门控阻塞的 Pod 放入单独的 `unschedulablePods` 队列
- **事件驱动**: 只有在相关事件发生时才重新评估 Pod 的调度资格
- **批量处理**: 支持批量移除门控，提高处理效率

---

## 7. 总结

Pod Scheduling Readiness 是 Kubernetes 调度器的一个重要增强特性，它通过调度门控机制解决了以下关键问题：

1. **性能优化**: 避免了对尚未准备好调度的 Pod 进行无效的调度尝试，显著提升调度器性能
2. **灵活控制**: 为外部控制器提供了精确控制 Pod 调度时机的能力
3. **扩展性**: 通过 PreEnqueue 插件机制，为调度器框架增加了新的扩展点

该特性在动态配额管理、外部依赖管理、批处理作业控制等场景中具有重要价值，为构建更加智能和高效的 Kubernetes 调度系统提供了基础能力。

从代码实现角度，该特性展现了 Kubernetes 调度器框架的良好设计：

- **插件化架构**: 通过 PreEnqueuePlugin 接口实现功能扩展
- **事件驱动**: 基于集群事件的队列管理机制
- **性能考虑**: 轻量级的门控检查逻辑，避免影响调度性能

随着该特性在 Kubernetes 1.30 中达到 GA 状态，它已经成为生产环境中可靠使用的重要功能。
