// scheduler-performance-tuning.go
// 调度器性能调优工具 - 根据集群规模动态优化调度器配置
package scheduler

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// SchedulerPerformanceTuner 调度器性能调优器
// 负责根据集群规模和工作负载特征动态调整调度器配置
type SchedulerPerformanceTuner struct {
	client kubernetes.Interface // Kubernetes客户端，用于与API服务器通信
	config *PerformanceConfig   // 性能配置参数
}

// PerformanceConfig 性能配置结构体
// 包含所有影响调度器性能的关键参数
type PerformanceConfig struct {
	// 调度队列配置 - 控制Pod排队和处理策略
	QueueSortPlugin          string // 队列排序插件名称，如"PrioritySort"
	MaxPendingPods           int    // 最大待调度Pod数量，防止内存溢出
	PodInitialBackoffSeconds int    // Pod调度失败后的初始退避时间（秒）
	PodMaxBackoffSeconds     int    // Pod调度失败后的最大退避时间（秒）

	// 节点评分配置 - 控制调度决策的性能和质量平衡
	PercentageOfNodesToScore int // 参与评分的节点百分比，影响调度延迟
	NodeScoreParallelism     int // 节点评分并行度，提高大集群调度吞吐量

	// 缓存配置 - 减少API调用，提高性能
	NodeInfoCacheTTL time.Duration // 节点信息缓存生存时间
	PodInfoCacheTTL  time.Duration // Pod信息缓存生存时间

	// 批处理配置 - 优化调度吞吐量
	BatchSize    int           // 批处理大小，一次处理的Pod数量
	BatchTimeout time.Duration // 批处理超时时间
}

// NewSchedulerPerformanceTuner 创建新的调度器性能调优器实例
// 初始化Kubernetes客户端和默认性能配置
func NewSchedulerPerformanceTuner() *SchedulerPerformanceTuner {
	// 获取集群内配置，用于与API服务器通信
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to create config: %v", err)
	}

	// 创建Kubernetes客户端
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}

	return &SchedulerPerformanceTuner{
		client: client,
		config: &PerformanceConfig{
			// 默认配置 - 适用于中等规模集群
			QueueSortPlugin:          "PrioritySort",         // 使用优先级排序
			MaxPendingPods:           5000,                   // 最大5000个待调度Pod
			PodInitialBackoffSeconds: 1,                      // 初始退避1秒
			PodMaxBackoffSeconds:     10,                     // 最大退避10秒
			PercentageOfNodesToScore: 50,                     // 评分50%的节点
			NodeScoreParallelism:     16,                     // 16个并行评分线程
			NodeInfoCacheTTL:         30 * time.Second,       // 节点信息缓存30秒
			PodInfoCacheTTL:          30 * time.Second,       // Pod信息缓存30秒
			BatchSize:                100,                    // 批处理100个Pod
			BatchTimeout:             100 * time.Millisecond, // 批处理超时100ms
		},
	}
}

// OptimizeForClusterSize 根据集群规模优化调度器配置
// 不同规模的集群需要不同的性能参数以达到最佳效果
func (spt *SchedulerPerformanceTuner) OptimizeForClusterSize(nodeCount int) {
	switch {
	case nodeCount < 100:
		// 小集群优化策略 - 优先调度质量
		spt.config.PercentageOfNodesToScore = 100 // 评分所有节点，确保最优选择
		spt.config.NodeScoreParallelism = 4       // 较低并行度，减少资源消耗
		spt.config.BatchSize = 50                 // 较小批处理，降低延迟
	case nodeCount < 1000:
		// 中等集群优化策略 - 平衡性能和质量
		spt.config.PercentageOfNodesToScore = 50 // 评分一半节点，平衡性能
		spt.config.NodeScoreParallelism = 8      // 适中并行度
		spt.config.BatchSize = 100               // 标准批处理大小
	default:
		// 大集群优化策略 - 优先调度性能
		spt.config.PercentageOfNodesToScore = 30 // 仅评分30%节点，提高速度
		spt.config.NodeScoreParallelism = 16     // 高并行度，提高吞吐量
		spt.config.BatchSize = 200               // 大批处理，提高效率
	}

	klog.Infof("Optimized scheduler for cluster size: %d nodes", nodeCount)
}

// GenerateOptimizedConfig 生成优化后的调度器配置文件
// 返回YAML格式的KubeSchedulerConfiguration配置
func (spt *SchedulerPerformanceTuner) GenerateOptimizedConfig() string {
	return fmt.Sprintf(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
profiles:
- schedulerName: optimized-scheduler
  plugins:
    queueSort:
      enabled:
      - name: %s                    # 队列排序插件
  pluginConfig:
  - name: DefaultBinder
    args:
      bindTimeoutSeconds: 600       # 绑定超时时间10分钟
  - name: PodTopologySpread
    args:
      defaultingType: List          # 使用列表模式应用拓扑约束
percentageOfNodesToScore: %d        # 参与评分的节点百分比
parallelism: %d                     # 并行评分线程数
`,
		spt.config.QueueSortPlugin,
		spt.config.PercentageOfNodesToScore,
		spt.config.NodeScoreParallelism,
	)
}
