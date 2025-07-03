// scheduler-selector.go
// 智能调度器选择器 - 根据工作负载特征自动选择最适合的调度器
package scheduler

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SchedulerSelector 调度器选择器
// 负责根据Pod的特征和标签自动选择最适合的调度器
type SchedulerSelector struct {
	client kubernetes.Interface // Kubernetes客户端
	rules  []SchedulerRule      // 调度器选择规则列表
}

// SchedulerRule 调度器选择规则
// 定义了工作负载识别条件和对应的调度器
type SchedulerRule struct {
	Name      string             // 规则名称，用于日志和调试
	Condition func(*v1.Pod) bool // 条件判断函数，返回true表示匹配
	Scheduler string             // 匹配时使用的调度器名称
	Priority  int                // 规则优先级，数值越高优先级越高
}

// NewSchedulerSelector 创建新的调度器选择器实例
// 初始化预定义的调度器选择规则
func NewSchedulerSelector(client kubernetes.Interface) *SchedulerSelector {
	return &SchedulerSelector{
		client: client,
		rules: []SchedulerRule{
			{
				Name: "GPU Workload", // GPU工作负载识别规则
				Condition: func(pod *v1.Pod) bool {
					// 检查所有容器是否有GPU资源请求
					for _, container := range pod.Spec.Containers {
						if _, hasGPU := container.Resources.Requests["nvidia.com/gpu"]; hasGPU {
							return true // 发现GPU请求，匹配GPU工作负载
						}
					}
					return false
				},
				Scheduler: "gpu-scheduler", // 使用GPU专用调度器
				Priority:  100,             // 最高优先级，GPU资源稀缺需优先处理
			},
			{
				Name: "Batch Workload", // 批处理工作负载识别规则
				Condition: func(pod *v1.Pod) bool {
					// 通过标签识别批处理工作负载
					if workloadType, exists := pod.Labels["workload.kubernetes.io/type"]; exists {
						return workloadType == "batch"
					}
					return false
				},
				Scheduler: "batch-scheduler", // 使用批处理专用调度器
				Priority:  80,                // 较高优先级
			},
			{
				Name: "Realtime Workload", // 实时工作负载识别规则
				Condition: func(pod *v1.Pod) bool {
					// 通过优先级标签识别实时工作负载
					if priority, exists := pod.Labels["workload.kubernetes.io/priority"]; exists {
						return priority == "realtime"
					}
					return false
				},
				Scheduler: "realtime-scheduler", // 使用实时专用调度器
				Priority:  90,                   // 高优先级，实时性要求高
			},
			{
				Name: "High Memory Workload", // 高内存工作负载识别规则
				Condition: func(pod *v1.Pod) bool {
					// 检查是否有高内存需求（>8Gi）
					for _, container := range pod.Spec.Containers {
						if memory := container.Resources.Requests.Memory(); memory != nil {
							if memory.Value() > 8*1024*1024*1024 { // 大于8Gi内存
								return true
							}
						}
					}
					return false
				},
				Scheduler: "memory-optimized-scheduler", // 使用内存优化调度器
				Priority:  70,                           // 中等优先级
			},
		},
	}
}

// SelectScheduler 为Pod选择最适合的调度器
// 根据规则优先级和匹配条件返回调度器名称
func (ss *SchedulerSelector) SelectScheduler(pod *v1.Pod) string {
	// 如果Pod已经指定了调度器，直接返回（尊重用户选择）
	if pod.Spec.SchedulerName != "" {
		return pod.Spec.SchedulerName
	}

	var selectedScheduler string
	var highestPriority int

	// 遍历所有规则，找到优先级最高的匹配规则
	for _, rule := range ss.rules {
		if rule.Condition(pod) && rule.Priority > highestPriority {
			selectedScheduler = rule.Scheduler
			highestPriority = rule.Priority
		}
	}

	// 如果没有匹配的规则，使用默认调度器
	if selectedScheduler == "" {
		return "default-scheduler"
	}

	return selectedScheduler
}

// UpdatePodScheduler 更新Pod的调度器配置
// 根据选择结果更新Pod的schedulerName字段
func (ss *SchedulerSelector) UpdatePodScheduler(ctx context.Context, pod *v1.Pod) error {
	selectedScheduler := ss.SelectScheduler(pod)

	// 如果调度器没有变化，无需更新
	if pod.Spec.SchedulerName == selectedScheduler {
		return nil // 无需更新，避免不必要的API调用
	}

	// 更新Pod的调度器名称
	pod.Spec.SchedulerName = selectedScheduler

	// 调用API更新Pod配置
	_, err := ss.client.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}
