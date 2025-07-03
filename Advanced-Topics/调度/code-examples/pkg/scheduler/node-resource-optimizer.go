// node-resource-optimizer.go
// 节点资源优化器 - 监控节点资源使用情况并执行自动优化策略
package scheduler

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

// NodeResourceOptimizer 节点资源优化器
// 负责收集节点指标、分析资源使用趋势并执行优化策略
type NodeResourceOptimizer struct {
	client        kubernetes.Interface    // Kubernetes API客户端
	metricsClient metricsclient.Interface // Metrics API客户端
	optimizer     *ResourceOptimizer      // 资源优化器实例
}

// ResourceOptimizer 资源优化器核心
// 存储节点指标和优化规则
type ResourceOptimizer struct {
	NodeMetrics       map[string]*NodeResourceMetrics // 节点指标映射
	OptimizationRules []OptimizationRule              // 优化规则列表
}

// NodeResourceMetrics 节点资源指标
// 记录单个节点的资源使用情况和容量信息
type NodeResourceMetrics struct {
	NodeName       string        // 节点名称
	CPUUsage       float64       // CPU使用率（百分比）
	MemoryUsage    float64       // 内存使用率（百分比）
	CPUCapacity    float64       // CPU总容量（核心数）
	MemoryCapacity float64       // 内存总容量（GB）
	PodCount       int           // 当前Pod数量
	PodCapacity    int           // 最大Pod容量
	LastUpdated    time.Time     // 最后更新时间
	Trend          ResourceTrend // 资源使用趋势
}

// ResourceTrend 资源使用趋势
// 分析资源使用的变化方向和置信度
type ResourceTrend struct {
	CPUTrend    string  // CPU趋势："increasing", "decreasing", "stable"
	MemoryTrend string  // 内存趋势："increasing", "decreasing", "stable"
	Confidence  float64 // 趋势预测的置信度（0-1）
}

// OptimizationRule 优化规则
// 定义触发条件、执行动作和优先级
type OptimizationRule struct {
	Name      string                                                                    // 规则名称
	Condition func(*NodeResourceMetrics) bool                                           // 触发条件函数
	Action    func(context.Context, *NodeResourceOptimizer, *NodeResourceMetrics) error // 执行动作函数
	Priority  int                                                                       // 优先级（数值越大优先级越高）
}

// NewNodeResourceOptimizer 创建新的节点资源优化器实例
// 初始化客户端连接和预定义的优化规则
func NewNodeResourceOptimizer(client kubernetes.Interface, metricsClient metricsclient.Interface) *NodeResourceOptimizer {
	return &NodeResourceOptimizer{
		client:        client,
		metricsClient: metricsClient,
		optimizer: &ResourceOptimizer{
			NodeMetrics: make(map[string]*NodeResourceMetrics),
			OptimizationRules: []OptimizationRule{
				{
					Name: "High CPU Usage",
					// 当CPU使用率超过80%时触发
					Condition: func(metrics *NodeResourceMetrics) bool {
						return metrics.CPUUsage > 80.0
					},
					Action:   handleHighCPUUsage, // 处理高CPU使用率
					Priority: 100,                // 最高优先级
				},
				{
					Name: "High Memory Usage",
					// 当内存使用率超过85%时触发
					Condition: func(metrics *NodeResourceMetrics) bool {
						return metrics.MemoryUsage > 85.0
					},
					Action:   handleHighMemoryUsage, // 处理高内存使用率
					Priority: 95,                    // 高优先级
				},
				{
					Name: "Low Resource Utilization",
					// 当资源利用率过低时触发（CPU<20%, 内存<30%, Pod<5个）
					Condition: func(metrics *NodeResourceMetrics) bool {
						return metrics.CPUUsage < 20.0 && metrics.MemoryUsage < 30.0 && metrics.PodCount < 5
					},
					Action:   handleLowUtilization, // 处理低资源利用率
					Priority: 50,                   // 中等优先级
				},
				{
					Name: "Resource Fragmentation",
					// 当Pod数量接近上限但CPU使用率不高时触发（资源碎片化）
					Condition: func(metrics *NodeResourceMetrics) bool {
						return float64(metrics.PodCount) > float64(metrics.PodCapacity)*0.8 && metrics.CPUUsage < 60.0
					},
					Action:   handleResourceFragmentation, // 处理资源碎片化
					Priority: 70,                          // 较高优先级
				},
			},
		},
	}
}

// CollectMetrics 收集所有节点的资源使用指标
// 从Kubernetes API和Metrics API获取节点信息并计算资源使用率
func (nro *NodeResourceOptimizer) CollectMetrics(ctx context.Context) error {
	// 获取集群中所有节点的列表
	nodes, err := nro.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	// 获取所有节点的实时资源使用指标
	nodeMetrics, err := nro.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node metrics: %v", err)
	}

	// 遍历每个节点，计算并存储其资源指标
	for _, node := range nodes.Items {
		metrics := nro.calculateNodeMetrics(&node, nodeMetrics.Items)
		if metrics != nil {
			nro.optimizer.NodeMetrics[node.Name] = metrics
		}
	}

	return nil
}

// calculateNodeMetrics 计算单个节点的资源使用指标
// 结合节点容量信息和实时使用数据，计算使用率和趋势
func (nro *NodeResourceOptimizer) calculateNodeMetrics(node *v1.Node, nodeMetrics []metricsv1beta1.NodeMetrics) *NodeResourceMetrics {
	// 在指标列表中查找当前节点对应的使用数据
	var metrics *metricsv1beta1.NodeMetrics
	for _, nm := range nodeMetrics {
		if nm.Name == node.Name {
			metrics = &nm
			break
		}
	}

	// 如果没有找到对应的指标数据，返回nil
	if metrics == nil {
		return nil
	}

	// 从节点状态中获取资源容量信息
	cpuCapacity := float64(node.Status.Capacity.Cpu().MilliValue()) / 1000                  // 转换为核心数
	memoryCapacity := float64(node.Status.Capacity.Memory().Value()) / (1024 * 1024 * 1024) // 转换为GB
	podCapacity := int(node.Status.Capacity.Pods().Value())                                 // Pod容量

	// 从指标数据中获取当前资源使用量
	cpuUsage := float64(metrics.Usage.Cpu().MilliValue()) / 1000                  // 转换为核心数
	memoryUsage := float64(metrics.Usage.Memory().Value()) / (1024 * 1024 * 1024) // 转换为GB

	// 获取节点上当前运行的Pod数量
	podCount := nro.getPodCount(node.Name)

	// 构建节点资源指标对象
	nodeResourceMetrics := &NodeResourceMetrics{
		NodeName:       node.Name,
		CPUUsage:       (cpuUsage / cpuCapacity) * 100,       // 计算CPU使用率百分比
		MemoryUsage:    (memoryUsage / memoryCapacity) * 100, // 计算内存使用率百分比
		CPUCapacity:    cpuCapacity,
		MemoryCapacity: memoryCapacity,
		PodCount:       podCount,
		PodCapacity:    podCapacity,
		LastUpdated:    time.Now(),
	}

	// 基于历史数据计算资源使用趋势
	nodeResourceMetrics.Trend = nro.calculateTrend(node.Name, nodeResourceMetrics)

	return nodeResourceMetrics
}

// getPodCount 获取指定节点上运行的Pod数量
// 通过字段选择器查询调度到特定节点的所有Pod
func (nro *NodeResourceOptimizer) getPodCount(nodeName string) int {
	pods, err := nro.client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName), // 按节点名称过滤Pod
	})
	if err != nil {
		return 0 // 查询失败时返回0
	}
	return len(pods.Items)
}

// calculateTrend 计算节点资源使用趋势
// 通过比较当前指标与历史指标，分析资源使用的变化方向
func (nro *NodeResourceOptimizer) calculateTrend(nodeName string, current *NodeResourceMetrics) ResourceTrend {
	// 获取该节点的历史指标数据
	previous, exists := nro.optimizer.NodeMetrics[nodeName]
	if !exists {
		// 如果没有历史数据，返回稳定趋势
		return ResourceTrend{
			CPUTrend:    "stable",
			MemoryTrend: "stable",
			Confidence:  0.5, // 低置信度
		}
	}

	// 计算CPU和内存使用率的变化量
	cpuDiff := current.CPUUsage - previous.CPUUsage
	memoryDiff := current.MemoryUsage - previous.MemoryUsage

	var cpuTrend, memoryTrend string

	// 根据变化量判断CPU使用趋势
	if cpuDiff > 5 {
		cpuTrend = "increasing" // 增长趋势
	} else if cpuDiff < -5 {
		cpuTrend = "decreasing" // 下降趋势
	} else {
		cpuTrend = "stable" // 稳定趋势
	}

	// 根据变化量判断内存使用趋势
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
		Confidence:  0.8, // 基于历史数据的高置信度
	}
}

// OptimizeResources 执行资源优化策略
// 根据节点指标应用相应的优化规则，按优先级顺序处理
func (nro *NodeResourceOptimizer) OptimizeResources(ctx context.Context) error {
	// 按优先级从高到低排序优化规则
	sort.Slice(nro.optimizer.OptimizationRules, func(i, j int) bool {
		return nro.optimizer.OptimizationRules[i].Priority > nro.optimizer.OptimizationRules[j].Priority
	})

	// 遍历所有节点的指标数据
	for _, metrics := range nro.optimizer.NodeMetrics {
		// 对每个节点应用匹配的优化规则
		for _, rule := range nro.optimizer.OptimizationRules {
			if rule.Condition(metrics) {
				klog.Infof("Applying optimization rule '%s' to node %s", rule.Name, metrics.NodeName)
				err := rule.Action(ctx, nro, metrics)
				if err != nil {
					klog.Errorf("Failed to apply rule '%s' to node %s: %v", rule.Name, metrics.NodeName, err)
				}
				break // 每个节点只应用第一个匹配的规则，避免冲突
			}
		}
	}

	return nil
}

// 优化规则实现函数

// handleHighCPUUsage 处理高CPU使用率的节点
// 通过添加污点防止新Pod调度到高负载节点
func handleHighCPUUsage(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
	klog.Warningf("Node %s has high CPU usage: %.2f%%", metrics.NodeName, metrics.CPUUsage)

	// 添加NoSchedule污点，阻止新Pod调度但不影响现有Pod
	return nro.addNodeTaint(ctx, metrics.NodeName, "high-cpu-usage", "NoSchedule")
}

// handleHighMemoryUsage 处理高内存使用率的节点
// 通过添加污点防止新Pod调度到内存紧张的节点
func handleHighMemoryUsage(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
	klog.Warningf("Node %s has high memory usage: %.2f%%", metrics.NodeName, metrics.MemoryUsage)

	// 添加NoSchedule污点，防止内存不足导致OOM
	return nro.addNodeTaint(ctx, metrics.NodeName, "high-memory-usage", "NoSchedule")
}

// handleLowUtilization 处理低资源利用率的节点
// 标记低利用率节点，便于后续的资源整合和成本优化
func handleLowUtilization(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
	klog.Infof("Node %s has low utilization - CPU: %.2f%%, Memory: %.2f%%, Pods: %d",
		metrics.NodeName, metrics.CPUUsage, metrics.MemoryUsage, metrics.PodCount)

	// 添加标签标识低利用率节点，用于自动缩容决策
	return nro.addNodeLabel(ctx, metrics.NodeName, "node.kubernetes.io/utilization", "low")
}

// handleResourceFragmentation 处理资源碎片化的节点
// 标记资源碎片化严重的节点，便于进行Pod重新调度
func handleResourceFragmentation(ctx context.Context, nro *NodeResourceOptimizer, metrics *NodeResourceMetrics) error {
	klog.Infof("Node %s has resource fragmentation - Pods: %d/%d, CPU: %.2f%%",
		metrics.NodeName, metrics.PodCount, metrics.PodCapacity, metrics.CPUUsage)

	// 添加标签标识资源碎片化节点，触发Pod整合策略
	return nro.addNodeLabel(ctx, metrics.NodeName, "node.kubernetes.io/fragmentation", "high")
}

// addNodeTaint 为指定节点添加污点
// 用于标记节点状态，影响Pod调度行为
func (nro *NodeResourceOptimizer) addNodeTaint(ctx context.Context, nodeName, key, effect string) error {
	// 获取节点对象
	node, err := nro.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// 检查污点是否已存在，避免重复添加
	for _, taint := range node.Spec.Taints {
		if taint.Key == key {
			return nil // 污点已存在，无需重复添加
		}
	}

	// 创建新污点对象
	newTaint := v1.Taint{
		Key:    key,                    // 污点键名
		Value:  "true",                 // 污点值
		Effect: v1.TaintEffect(effect), // 污点效果：NoSchedule/PreferNoSchedule/NoExecute
	}

	// 将新污点添加到节点规格中
	node.Spec.Taints = append(node.Spec.Taints, newTaint)

	// 更新节点对象到Kubernetes API
	_, err = nro.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return err
}

// addNodeLabel 为指定节点添加标签
// 用于标记节点属性，便于调度器和其他组件识别
func (nro *NodeResourceOptimizer) addNodeLabel(ctx context.Context, nodeName, key, value string) error {
	// 获取节点对象
	node, err := nro.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// 初始化标签映射（如果不存在）
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	// 设置标签键值对
	node.Labels[key] = value

	// 更新节点对象到Kubernetes API
	_, err = nro.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return err
}
