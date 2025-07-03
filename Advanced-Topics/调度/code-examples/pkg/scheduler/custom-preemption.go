// custom-preemption.go
// 自定义抢占插件 - 提供智能的 Pod 抢占功能
package scheduler

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CustomPreemption 自定义抢占管理器
type CustomPreemption struct {
	client kubernetes.Interface
}

// PreemptionCandidate 抢占候选
type PreemptionCandidate struct {
	Node    *corev1.Node
	Victims []*corev1.Pod
	Score   int64
}

// Name 返回插件名称
func (cp *CustomPreemption) Name() string {
	return "CustomPreemption"
}

// NewCustomPreemption 创建自定义抢占管理器
func NewCustomPreemption(client kubernetes.Interface) *CustomPreemption {
	return &CustomPreemption{
		client: client,
	}
}

// PreemptPod 为指定Pod执行抢占
func (cp *CustomPreemption) PreemptPod(ctx context.Context, pod *corev1.Pod) (string, error) {
	klog.V(2).Infof("Starting preemption for pod %s/%s", pod.Namespace, pod.Name)

	// 查找抢占候选
	candidates, err := cp.findPreemptionCandidates(ctx, pod)
	if err != nil {
		return "", fmt.Errorf("failed to find preemption candidates: %v", err)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no preemption candidates found")
	}

	// 选择最佳候选
	bestCandidate := cp.selectBestCandidate(candidates)

	// 执行抢占
	if err := cp.executePreemption(ctx, pod, bestCandidate); err != nil {
		return "", fmt.Errorf("failed to execute preemption: %v", err)
	}

	klog.Infof("Successfully preempted %d pods on node %s for pod %s/%s", 
		len(bestCandidate.Victims), bestCandidate.Node.Name, pod.Namespace, pod.Name)

	return bestCandidate.Node.Name, nil
}

// findPreemptionCandidates 查找抢占候选
func (cp *CustomPreemption) findPreemptionCandidates(ctx context.Context, pod *corev1.Pod) ([]PreemptionCandidate, error) {
	var candidates []PreemptionCandidate

	// 获取所有节点
	nodes, err := cp.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	for _, node := range nodes.Items {
		// 检查节点是否可调度
		if node.Spec.Unschedulable {
			continue
		}

		// 查找可被抢占的 Pod
		victims, err := cp.findVictims(ctx, pod, &node)
		if err != nil {
			klog.Errorf("Failed to find victims on node %s: %v", node.Name, err)
			continue
		}

		if len(victims) == 0 {
			continue
		}

		// 计算抢占分数
		score := cp.calculatePreemptionScore(pod, &node, victims)

		candidates = append(candidates, PreemptionCandidate{
			Node:    &node,
			Victims: victims,
			Score:   score,
		})
	}

	// 按分数排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates, nil
}

// findVictims 查找可被抢占的 Pod
func (cp *CustomPreemption) findVictims(ctx context.Context, preemptor *corev1.Pod, node *corev1.Node) ([]*corev1.Pod, error) {
	// 获取节点上的所有 Pod
	pods, err := cp.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods on node %s: %v", node.Name, err)
	}

	var victims []*corev1.Pod
	preemptorPriority := getPodPriority(preemptor)

	for i := range pods.Items {
		pod := &pods.Items[i]

		// 跳过系统 Pod
		if isSystemPod(pod) {
			continue
		}

		// 跳过已经在终止的 Pod
		if pod.DeletionTimestamp != nil {
			continue
		}

		// 只抢占优先级较低的 Pod
	podPriority := getPodPriority(pod)
		if podPriority >= preemptorPriority {
			continue
		}

		// 检查是否可以抢占
		if cp.canPreempt(preemptor, pod) {
			victims = append(victims, pod)
		}
	}

	// 按优先级排序，优先抢占优先级最低的
	sort.Slice(victims, func(i, j int) bool {
		return getPodPriority(victims[i]) < getPodPriority(victims[j])
	})

	// 计算需要抢占的最小 Pod 集合
	return cp.selectMinimalVictims(preemptor, node, victims), nil
}

// canPreempt 检查是否可以抢占
func (cp *CustomPreemption) canPreempt(preemptor, victim *corev1.Pod) bool {
	// 检查 PDB (Pod Disruption Budget)
	if !cp.respectsPDB(victim) {
		return false
	}

	// 检查抢占策略
	if !cp.allowsPreemption(preemptor, victim) {
		return false
	}

	return true
}

// selectMinimalVictims 选择最小的受害者集合
func (cp *CustomPreemption) selectMinimalVictims(preemptor *corev1.Pod, node *corev1.Node, candidates []*corev1.Pod) []*corev1.Pod {
	// 计算抢占者的资源需求
	cpuRequest := getPodCPURequest(preemptor)
	memoryRequest := getPodMemoryRequest(preemptor)

	// 计算节点当前可用资源
	cpuAvailable := getNodeCPUAvailable(node)
	memoryAvailable := getNodeMemoryAvailable(node)

	// 计算需要释放的资源
	cpuNeeded := cpuRequest - cpuAvailable
	memoryNeeded := memoryRequest - memoryAvailable

	if cpuNeeded <= 0 && memoryNeeded <= 0 {
		return nil // 不需要抢占
	}

	var victims []*corev1.Pod
	var cpuReleased, memoryReleased int64

	// 贪心算法选择最小的受害者集合
	for _, candidate := range candidates {
		if cpuReleased >= cpuNeeded && memoryReleased >= memoryNeeded {
			break
		}

		victims = append(victims, candidate)
		cpuReleased += getPodCPURequest(candidate)
		memoryReleased += getPodMemoryRequest(candidate)
	}

	return victims
}

// calculatePreemptionScore 计算抢占分数
func (cp *CustomPreemption) calculatePreemptionScore(preemptor *corev1.Pod, node *corev1.Node, victims []*corev1.Pod) int64 {
	var score int64

	// 受害者数量越少越好
	score += int64((100 - len(victims)) * 10)

	// 受害者优先级越低越好
	for _, victim := range victims {
		score += int64(1000 - getPodPriority(victim))
	}

	// 节点资源利用率
	cpuUtilization := float64(getNodeCPUUsed(node)) / float64(getNodeCPUAvailable(node))
	memoryUtilization := float64(getNodeMemoryUsed(node)) / float64(getNodeMemoryAvailable(node))
	avgUtilization := (cpuUtilization + memoryUtilization) / 2

	// 偏好利用率较低的节点
	score += int64((1.0 - avgUtilization) * 100)

	// 节点亲和性加分
	if matchesNodeAffinity(preemptor, node) {
		score += 200
	}

	return score
}

// selectBestCandidate 选择最佳候选
func (cp *CustomPreemption) selectBestCandidate(candidates []PreemptionCandidate) PreemptionCandidate {
	// 已经按分数排序，返回第一个
	return candidates[0]
}

// executePreemption 执行抢占
func (cp *CustomPreemption) executePreemption(ctx context.Context, preemptor *corev1.Pod, candidate PreemptionCandidate) error {
	// 删除受害者 Pod
	for _, victim := range candidate.Victims {
		err := cp.client.CoreV1().Pods(victim.Namespace).Delete(ctx, victim.Name, metav1.DeleteOptions{
			GracePeriodSeconds: int64Ptr(30),
		})
		if err != nil {
			return fmt.Errorf("failed to delete victim pod %s/%s: %v", victim.Namespace, victim.Name, err)
		}
		klog.Infof("Preempted pod %s/%s on node %s", victim.Namespace, victim.Name, candidate.Node.Name)
	}

	// 等待 Pod 删除
	time.Sleep(5 * time.Second)

	return nil
}

// 辅助函数
func isSystemPod(pod *corev1.Pod) bool {
	// 检查是否为系统 Pod
	if pod.Namespace == "kube-system" {
		return true
	}

	// 检查是否为 DaemonSet Pod
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}

	// 检查是否为静态 Pod
	if pod.Annotations["kubernetes.io/config.source"] == "file" {
		return true
	}

	return false
}

func (cp *CustomPreemption) respectsPDB(pod *corev1.Pod) bool {
	// 简化的 PDB 检查
	// 实际实现应该检查 PodDisruptionBudget
	return true
}

func (cp *CustomPreemption) allowsPreemption(preemptor, victim *corev1.Pod) bool {
	// 检查抢占策略
	// 例如检查注解、标签等
	if victim.Annotations["scheduler.alpha.kubernetes.io/preemption-policy"] == "Never" {
		return false
	}

	return true
}

func getNodeCPUUsed(node *corev1.Node) int64 {
	// 简化实现，实际应该从 metrics 获取
	return 0
}

func getNodeMemoryUsed(node *corev1.Node) int64 {
	// 简化实现，实际应该从 metrics 获取
	return 0
}

func int64Ptr(i int64) *int64 {
	return &i
}