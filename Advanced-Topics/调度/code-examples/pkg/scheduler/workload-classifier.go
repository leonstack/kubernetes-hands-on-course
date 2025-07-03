// workload-classifier.go
// 智能工作负载分类器 - 根据Pod特征自动识别工作负载类型并应用相应的调度策略
package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// WorkloadClassifier 智能工作负载分类器
// 根据Pod特征自动识别工作负载类型并应用相应的调度策略
type WorkloadClassifier struct {
	client   kubernetes.Interface        // Kubernetes API客户端
	rules    []ClassificationRule        // 分类规则列表，按优先级排序
	profiles map[string]*WorkloadProfile // 工作负载配置文件映射
}

// ClassificationRule 工作负载分类规则
// 定义如何识别特定类型的工作负载
type ClassificationRule struct {
	Name         string             // 规则名称
	Priority     int                // 规则优先级，数值越高优先级越高
	Condition    func(*v1.Pod) bool // 判断条件函数，返回true表示匹配
	WorkloadType string             // 匹配的工作负载类型
	Scheduler    string             // 推荐使用的调度器
	Labels       map[string]string  // 要添加的标签
	Annotations  map[string]string  // 要添加的注解
}

// WorkloadProfile 工作负载配置文件
// 定义特定类型工作负载的完整特征和调度要求
type WorkloadProfile struct {
	Type            string          // 工作负载类型标识
	Description     string          // 类型描述
	ResourcePattern ResourcePattern // 资源使用模式
	SchedulingHints SchedulingHints // 调度提示和约束
	SLA             SLARequirements // 服务级别协议要求
}

// ResourcePattern 资源使用模式
// 描述工作负载的资源消耗特征
type ResourcePattern struct {
	CPUIntensive     bool   // 是否为CPU密集型
	MemoryIntensive  bool   // 是否为内存密集型
	IOIntensive      bool   // 是否为IO密集型
	NetworkIntensive bool   // 是否为网络密集型
	GPURequired      bool   // 是否需要GPU资源
	StoragePattern   string // 存储访问模式："high-iops", "high-throughput", "low-latency"
}

// SchedulingHints 调度提示和约束
// 为调度器提供工作负载的调度偏好
type SchedulingHints struct {
	PreferredScheduler string                        // 首选调度器名称
	NodeAffinity       *v1.NodeAffinity              // 节点亲和性规则
	PodAntiAffinity    *v1.PodAntiAffinity           // Pod反亲和性规则
	Tolerations        []v1.Toleration               // 容忍度配置
	TopologySpread     []v1.TopologySpreadConstraint // 拓扑分布约束
}

// SLARequirements 服务级别协议要求
// 定义工作负载的性能和可用性要求
type SLARequirements struct {
	MaxLatency        time.Duration // 最大延迟要求
	Availability      float64       // 可用性要求（如99.9%）
	Throughput        int           // 吞吐量要求（每秒请求数）
	ResourceGuarantee bool          // 是否需要资源保证
}

// NewWorkloadClassifier 创建新的工作负载分类器实例
// 初始化预定义的工作负载配置文件和分类规则
func NewWorkloadClassifier(client kubernetes.Interface) *WorkloadClassifier {
	wc := &WorkloadClassifier{
		client:   client,
		profiles: make(map[string]*WorkloadProfile),
	}

	// 初始化预定义的工作负载配置文件
	wc.initializeProfiles()
	// 初始化工作负载分类规则
	wc.initializeRules()

	return wc
}

// initializeProfiles 初始化预定义的工作负载配置文件
// 定义常见工作负载类型的资源模式和调度要求
func (wc *WorkloadClassifier) initializeProfiles() {
	// Web前端应用配置文件
	// 特点：网络密集型，低延迟要求，需要高可用性
	wc.profiles["web-frontend"] = &WorkloadProfile{
		Type:        "web-frontend",
		Description: "Web frontend applications",
		ResourcePattern: ResourcePattern{
			CPUIntensive:     false,         // 非CPU密集型
			MemoryIntensive:  false,         // 非内存密集型
			IOIntensive:      false,         // 非IO密集型
			NetworkIntensive: true,          // 网络密集型，需要处理大量HTTP请求
			GPURequired:      false,         // 不需要GPU
			StoragePattern:   "low-latency", // 需要低延迟存储访问
		},
		SchedulingHints: SchedulingHints{
			PreferredScheduler: "realtime-scheduler", // 使用实时调度器
			// 拓扑分布约束：确保Pod在不同节点上均匀分布
			TopologySpread: []v1.TopologySpreadConstraint{
				{
					MaxSkew:           1,                        // 最大偏差为1
					TopologyKey:       "kubernetes.io/hostname", // 按主机名分布
					WhenUnsatisfiable: v1.DoNotSchedule,         // 不满足时拒绝调度
				},
			},
		},
		SLA: SLARequirements{
			MaxLatency:        100 * time.Millisecond, // 最大延迟100ms
			Availability:      99.9,                   // 99.9%可用性
			Throughput:        1000,                   // 1000 RPS
			ResourceGuarantee: true,                   // 需要资源保证
		},
	}

	// 批处理作业配置文件
	// 特点：CPU和内存密集型，对延迟不敏感，可容忍较低可用性
	wc.profiles["batch-processing"] = &WorkloadProfile{
		Type:        "batch-processing",
		Description: "Batch processing jobs",
		ResourcePattern: ResourcePattern{
			CPUIntensive:     true,              // CPU密集型，需要大量计算资源
			MemoryIntensive:  true,              // 内存密集型，处理大数据集
			IOIntensive:      false,             // 非IO密集型
			NetworkIntensive: false,             // 非网络密集型
			GPURequired:      false,             // 通常不需要GPU
			StoragePattern:   "high-throughput", // 需要高吞吐量存储
		},
		SchedulingHints: SchedulingHints{
			PreferredScheduler: "batch-scheduler", // 使用批处理调度器
			// 容忍度：可以调度到专用的批处理节点
			Tolerations: []v1.Toleration{
				{
					Key:      "node.kubernetes.io/batch", // 批处理节点污点键
					Operator: v1.TolerationOpEqual,       // 等值匹配
					Value:    "true",                     // 污点值
					Effect:   v1.TaintEffectNoSchedule,   // NoSchedule效果
				},
			},
		},
		SLA: SLARequirements{
			MaxLatency:        0,     // 不关心延迟
			Availability:      95.0,  // 较低的可用性要求
			Throughput:        0,     // 不关心实时吞吐量
			ResourceGuarantee: false, // 不需要严格的资源保证
		},
	}

	// 机器学习训练配置文件
	// 特点：需要GPU，CPU/内存/IO密集型，需要资源保证
	wc.profiles["ml-training"] = &WorkloadProfile{
		Type:        "ml-training",
		Description: "Machine learning training jobs",
		ResourcePattern: ResourcePattern{
			CPUIntensive:     true,        // CPU密集型，模型训练需要大量计算
			MemoryIntensive:  true,        // 内存密集型，加载大型数据集和模型
			IOIntensive:      true,        // IO密集型，频繁读取训练数据
			NetworkIntensive: false,       // 非网络密集型（单机训练）
			GPURequired:      true,        // 需要GPU加速
			StoragePattern:   "high-iops", // 需要高IOPS存储
		},
		SchedulingHints: SchedulingHints{
			PreferredScheduler: "gpu-scheduler", // 使用GPU调度器
			// 节点亲和性：必须调度到有GPU的节点
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "accelerator",                                // GPU类型标签
									Operator: v1.NodeSelectorOpIn,                          // 包含操作符
									Values:   []string{"nvidia-tesla-v100", "nvidia-a100"}, // 支持的GPU型号
								},
							},
						},
					},
				},
			},
		},
		SLA: SLARequirements{
			MaxLatency:        0,    // 不关心延迟（批处理任务）
			Availability:      99.0, // 高可用性要求
			Throughput:        0,    // 不关心实时吞吐量
			ResourceGuarantee: true, // 需要资源保证，避免训练中断
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
