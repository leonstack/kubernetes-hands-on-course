// dynamic-resource-quota.go
// 动态资源配额管理器 - 根据实际使用情况自动调整命名空间资源配额
package scheduler

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

// DynamicResourceQuotaManager 动态资源配额管理器
// 负责监控命名空间资源使用情况并自动调整配额
type DynamicResourceQuotaManager struct {
	client  kubernetes.Interface            // Kubernetes客户端
	quotas  map[string]*ResourceQuotaConfig // 命名空间配额配置映射
	metrics *QuotaMetrics                   // 配额使用指标和历史数据
}

// ResourceQuotaConfig 资源配额配置
// 定义命名空间的基础配额、最大配额和扩缩容规则
type ResourceQuotaConfig struct {
	Namespace    string          // 命名空间名称
	BaseQuota    v1.ResourceList // 基础资源配额
	MaxQuota     v1.ResourceList // 最大资源配额限制
	ScalingRules []ScalingRule   // 自动扩缩容规则列表
	Priority     int             // 配额优先级，影响资源分配顺序
}

// ScalingRule 扩缩容规则
// 定义触发配额调整的条件和调整策略
type ScalingRule struct {
	MetricType   string        // 监控指标类型："cpu_usage", "memory_usage", "pod_count"
	Threshold    float64       // 触发扩容的阈值
	ScaleFactor  float64       // 扩容倍数，如1.5表示扩容50%
	CooldownTime time.Duration // 冷却时间，防止频繁调整
}

// QuotaMetrics 配额使用指标
// 存储历史使用数据和扩缩容时间记录
type QuotaMetrics struct {
	UsageHistory map[string][]ResourceUsage // 各命名空间的资源使用历史
	LastScaling  map[string]time.Time       // 各命名空间的最后扩缩容时间
}

// ResourceUsage 资源使用情况快照
// 记录某个时间点的资源使用状态
type ResourceUsage struct {
	Timestamp time.Time // 记录时间戳
	CPU       float64   // CPU使用量（核心数）
	Memory    float64   // 内存使用量（GB）
	Pods      int       // Pod数量
}

// NewDynamicResourceQuotaManager 创建新的动态资源配额管理器实例
// 初始化客户端连接和内部数据结构
func NewDynamicResourceQuotaManager(client kubernetes.Interface) *DynamicResourceQuotaManager {
	return &DynamicResourceQuotaManager{
		client: client,
		quotas: make(map[string]*ResourceQuotaConfig), // 初始化配额配置映射
		metrics: &QuotaMetrics{
			UsageHistory: make(map[string][]ResourceUsage), // 初始化使用历史
			LastScaling:  make(map[string]time.Time),       // 初始化扩缩容时间记录
		},
	}
}

// AddQuotaConfig 添加命名空间的配额配置
// 为指定命名空间设置动态配额管理规则
func (drqm *DynamicResourceQuotaManager) AddQuotaConfig(config *ResourceQuotaConfig) {
	drqm.quotas[config.Namespace] = config
	klog.Infof("Added quota config for namespace: %s", config.Namespace)
}

// UpdateQuotas 更新所有命名空间的资源配额
// 定期执行的主要逻辑，监控使用情况并根据规则调整配额
func (drqm *DynamicResourceQuotaManager) UpdateQuotas(ctx context.Context) error {
	for namespace, config := range drqm.quotas {
		// 获取当前命名空间的资源使用情况
		usage, err := drqm.getCurrentUsage(ctx, namespace)
		if err != nil {
			klog.Errorf("Failed to get usage for namespace %s: %v", namespace, err)
			continue
		}

		// 记录使用历史，用于趋势分析
		drqm.recordUsage(namespace, usage)

		// 根据扩缩容规则计算新的配额
		newQuota := drqm.calculateNewQuota(namespace, config, usage)
		if newQuota != nil {
			// 应用新的配额设置
			err = drqm.updateResourceQuota(ctx, namespace, newQuota)
			if err != nil {
				klog.Errorf("Failed to update quota for namespace %s: %v", namespace, err)
			} else {
				// 记录扩缩容时间，用于冷却期控制
				drqm.metrics.LastScaling[namespace] = time.Now()
				klog.Infof("Updated quota for namespace %s", namespace)
			}
		}
	}
	return nil
}

// getCurrentUsage 获取指定命名空间的当前资源使用情况
// 统计所有Pod的资源请求量，作为配额调整的依据
func (drqm *DynamicResourceQuotaManager) getCurrentUsage(ctx context.Context, namespace string) (*ResourceUsage, error) {
	// 获取命名空间下的所有Pod
	pods, err := drqm.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var totalCPU, totalMemory int64
	podCount := len(pods.Items)

	// 遍历所有Pod，累计资源请求量
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			// 累计CPU请求量（毫核心）
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				totalCPU += cpu.MilliValue()
			}
			// 累计内存请求量（字节）
			if memory := container.Resources.Requests.Memory(); memory != nil {
				totalMemory += memory.Value()
			}
		}
	}

	return &ResourceUsage{
		Timestamp: time.Now(),
		CPU:       float64(totalCPU) / 1000,                    // 转换为核心数
		Memory:    float64(totalMemory) / (1024 * 1024 * 1024), // 转换为GB
		Pods:      podCount,
	}, nil
}

// calculateNewQuota 根据使用情况和扩缩容规则计算新的资源配额
// 核心的配额调整逻辑，支持CPU、内存和Pod数量的动态调整
func (drqm *DynamicResourceQuotaManager) calculateNewQuota(namespace string, config *ResourceQuotaConfig, usage *ResourceUsage) v1.ResourceList {
	// 检查冷却期，防止频繁调整配额
	lastScaling, exists := drqm.metrics.LastScaling[namespace]
	if exists && time.Since(lastScaling) < 5*time.Minute {
		return nil // 冷却期内不调整
	}

	// 基于基础配额创建新配额
	newQuota := config.BaseQuota.DeepCopy()
	needsUpdate := false

	// 遍历所有扩缩容规则
	for _, rule := range config.ScalingRules {
		var currentValue float64
		var resourceName v1.ResourceName

		// 根据监控指标类型获取当前值和对应的资源名称
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

		// 检查是否超过阈值，需要扩容
		if currentValue > rule.Threshold {
			currentQuota := newQuota[resourceName]
			var newValue resource.Quantity

			// 根据资源类型计算新的配额值
			switch resourceName {
			case v1.ResourceRequestsCPU:
				newValue = resource.MustParse(fmt.Sprintf("%.1f", currentValue*rule.ScaleFactor))
			case v1.ResourceRequestsMemory:
				newValue = resource.MustParse(fmt.Sprintf("%.1fGi", currentValue*rule.ScaleFactor))
			case v1.ResourcePods:
				newValue = resource.MustParse(fmt.Sprintf("%.0f", currentValue*rule.ScaleFactor))
			}

			// 检查是否超过最大配额限制
			maxQuota := config.MaxQuota[resourceName]
			if newValue.Cmp(maxQuota) > 0 {
				newValue = maxQuota
			}

			// 只有当新值与当前值不同时才更新
			if newValue.Cmp(currentQuota) != 0 {
				newQuota[resourceName] = newValue
				needsUpdate = true
			}
		}
	}

	// 返回新配额或nil（如果不需要更新）
	if needsUpdate {
		return newQuota
	}
	return nil
}

// updateResourceQuota 更新或创建命名空间的ResourceQuota对象
// 将计算出的新配额应用到Kubernetes集群中
func (drqm *DynamicResourceQuotaManager) updateResourceQuota(ctx context.Context, namespace string, quota v1.ResourceList) error {
	// 构建ResourceQuota对象
	resourceQuota := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dynamic-quota", // 固定名称，便于管理
			Namespace: namespace,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: quota, // 设置硬限制
		},
	}

	// 尝试更新现有的ResourceQuota
	_, err := drqm.client.CoreV1().ResourceQuotas(namespace).Update(ctx, resourceQuota, metav1.UpdateOptions{})
	if err != nil {
		// 如果ResourceQuota不存在则创建新的
		_, err = drqm.client.CoreV1().ResourceQuotas(namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
	}

	return err
}

// recordUsage 记录命名空间的资源使用历史
// 维护滑动窗口的使用数据，用于趋势分析和决策
func (drqm *DynamicResourceQuotaManager) recordUsage(namespace string, usage *ResourceUsage) {
	// 获取现有的使用历史
	history := drqm.metrics.UsageHistory[namespace]
	history = append(history, *usage)

	// 保留最近24小时的数据，避免内存无限增长
	cutoff := time.Now().Add(-24 * time.Hour)
	var filtered []ResourceUsage
	for _, record := range history {
		if record.Timestamp.After(cutoff) {
			filtered = append(filtered, record)
		}
	}

	// 更新使用历史
	drqm.metrics.UsageHistory[namespace] = filtered
}
