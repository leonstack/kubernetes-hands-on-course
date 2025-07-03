// batch-scheduler.go
package scheduler

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
		bs.metrics.RecordSchedulingLatency("batch-scheduler", "default", "success", time.Since(start))
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

// BatchNodeScore 批量调度器节点评分
type BatchNodeScore struct {
	Node  *corev1.Node
	Score int64
}

// scoreNodes 对节点进行评分
func (bs *BatchScheduler) scoreNodes(pod *corev1.Pod, nodes []*corev1.Node) []BatchNodeScore {
	scores := make([]BatchNodeScore, 0, len(nodes))

	for _, node := range nodes {
		score := bs.calculateNodeScore(pod, node)
		scores = append(scores, BatchNodeScore{
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
