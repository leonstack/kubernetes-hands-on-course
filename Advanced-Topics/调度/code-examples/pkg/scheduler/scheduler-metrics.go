// scheduler-metrics.go
package scheduler

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