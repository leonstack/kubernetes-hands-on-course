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

// SchedulerTroubleshooter 调度器故障排除器
type SchedulerTroubleshooter struct {
	client kubernetes.Interface
}

// TroubleshootingReport 故障排除报告
type TroubleshootingReport struct {
	PodName      string            `json:"pod_name"`
	Namespace    string            `json:"namespace"`
	Timestamp    time.Time         `json:"timestamp"`
	Diagnosis    []DiagnosisStep   `json:"diagnosis"`
	Recommendations []TroubleshootingRecommendation `json:"recommendations"`
	Severity     string            `json:"severity"`
}

// DiagnosisStep 诊断步骤
type DiagnosisStep struct {
	Step        string `json:"step"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Details     string `json:"details"`
}

// TroubleshootingRecommendation 故障排除建议
type TroubleshootingRecommendation struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Command     string `json:"command,omitempty"`
}

// NewSchedulerTroubleshooter 创建新的故障排除器
func NewSchedulerTroubleshooter(client kubernetes.Interface) *SchedulerTroubleshooter {
	return &SchedulerTroubleshooter{
		client: client,
	}
}

// DiagnosePendingPod 诊断待定Pod
func (st *SchedulerTroubleshooter) DiagnosePendingPod(ctx context.Context, namespace, podName string) (*TroubleshootingReport, error) {
	report := &TroubleshootingReport{
		PodName:   podName,
		Namespace: namespace,
		Timestamp: time.Now(),
		Severity:  "medium",
	}

	// 获取Pod信息
	pod, err := st.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	// 检查Pod调度条件
	if err := st.checkPodSchedulingConditions(ctx, pod, report); err != nil {
		klog.Errorf("Failed to check pod scheduling conditions: %v", err)
	}

	// 检查节点资源
	if err := st.checkNodeResources(ctx, pod, report); err != nil {
		klog.Errorf("Failed to check node resources: %v", err)
	}

	// 检查亲和性和反亲和性
	if err := st.checkAffinityConstraints(ctx, pod, report); err != nil {
		klog.Errorf("Failed to check affinity constraints: %v", err)
	}

	// 检查污点和容忍度
	if err := st.checkTaintsAndTolerations(ctx, pod, report); err != nil {
		klog.Errorf("Failed to check taints and tolerations: %v", err)
	}

	// 检查资源配额
	if err := st.checkResourceQuotas(ctx, pod, report); err != nil {
		klog.Errorf("Failed to check resource quotas: %v", err)
	}

	// 生成建议
	st.generateRecommendations(report)

	return report, nil
}

// checkPodSchedulingConditions 检查Pod调度条件
func (st *SchedulerTroubleshooter) checkPodSchedulingConditions(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) error {
	step := DiagnosisStep{
		Step:        "pod_conditions",
		Description: "检查Pod调度条件",
	}

	// 检查Pod状态
	if pod.Status.Phase == v1.PodPending {
		// 检查调度失败原因
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionFalse {
				step.Status = "failed"
				step.Details = fmt.Sprintf("调度失败: %s - %s", condition.Reason, condition.Message)
				
				// 根据失败原因生成具体建议
				if strings.Contains(condition.Message, "Insufficient") {
					report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
						Action:      "scale_cluster",
						Description: "集群资源不足，考虑添加更多节点或调整Pod资源请求",
						Priority:    "high",
					})
				}
				
				if strings.Contains(condition.Message, "node(s) didn't match Pod's node affinity") {
					report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
						Action:      "check_node_affinity",
						Description: "检查Pod的节点亲和性配置是否过于严格",
						Priority:    "medium",
					})
				}
				
				if strings.Contains(condition.Message, "node(s) had untolerated taint") {
					report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
						Action:      "add_tolerations",
						Description: "为Pod添加适当的污点容忍度",
						Priority:    "medium",
					})
				}
				
				if strings.Contains(condition.Message, "exceeded quota") {
					report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
						Action:      "check_quota",
						Description: "检查命名空间资源配额限制",
						Priority:    "high",
						Command:     fmt.Sprintf("kubectl describe quota -n %s", pod.Namespace),
					})
				}
				break
			}
		}
	} else {
		step.Status = "passed"
		step.Details = fmt.Sprintf("Pod状态: %s", pod.Status.Phase)
	}

	report.Diagnosis = append(report.Diagnosis, step)
	return nil
}

// checkNodeResources 检查节点资源
func (st *SchedulerTroubleshooter) checkNodeResources(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) error {
	step := DiagnosisStep{
		Step:        "node_resources",
		Description: "检查节点资源可用性",
	}

	nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "error"
		step.Details = fmt.Sprintf("获取节点列表失败: %v", err)
		report.Diagnosis = append(report.Diagnosis, step)
		return err
	}

	// 计算Pod资源需求
	cpuRequest := int64(0)
	memoryRequest := int64(0)
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				cpuRequest += cpu.MilliValue()
			}
			if memory := container.Resources.Requests.Memory(); memory != nil {
				memoryRequest += memory.Value()
			}
		}
	}

	suitableNodes := 0
	for _, node := range nodes.Items {
		// 检查节点是否可调度
		if node.Spec.Unschedulable {
			continue
		}

		// 检查节点是否就绪
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				isReady = true
				break
			}
		}
		if !isReady {
			continue
		}

		// 检查资源容量
		allocatableCPU := node.Status.Allocatable.Cpu()
		allocatableMemory := node.Status.Allocatable.Memory()

		if allocatableCPU != nil && allocatableMemory != nil {
			if allocatableCPU.MilliValue() >= cpuRequest && allocatableMemory.Value() >= memoryRequest {
				suitableNodes++
			}
		}
	}

	if suitableNodes == 0 {
		step.Status = "failed"
		step.Details = fmt.Sprintf("没有节点满足资源需求 (CPU: %dm, Memory: %dMi)", cpuRequest, memoryRequest/(1024*1024))
		report.Severity = "high"
	} else {
		step.Status = "passed"
		step.Details = fmt.Sprintf("找到 %d 个满足资源需求的节点", suitableNodes)
	}

	report.Diagnosis = append(report.Diagnosis, step)
	return nil
}

// checkAffinityConstraints 检查亲和性约束
func (st *SchedulerTroubleshooter) checkAffinityConstraints(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) error {
	step := DiagnosisStep{
		Step:        "affinity_constraints",
		Description: "检查亲和性和反亲和性约束",
	}

	if pod.Spec.Affinity == nil {
		step.Status = "passed"
		step.Details = "未配置亲和性约束"
	} else {
		// 检查节点亲和性
		if pod.Spec.Affinity.NodeAffinity != nil {
			step.Details += "配置了节点亲和性; "
		}

		// 检查Pod亲和性
		if pod.Spec.Affinity.PodAffinity != nil {
			step.Details += "配置了Pod亲和性; "
		}

		// 检查Pod反亲和性
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			step.Details += "配置了Pod反亲和性; "
		}

		step.Status = "warning"
		step.Details += "建议检查亲和性配置是否过于严格"
	}

	report.Diagnosis = append(report.Diagnosis, step)
	return nil
}

// checkTaintsAndTolerations 检查污点和容忍度
func (st *SchedulerTroubleshooter) checkTaintsAndTolerations(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) error {
	step := DiagnosisStep{
		Step:        "taints_tolerations",
		Description: "检查污点和容忍度配置",
	}

	nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "error"
		step.Details = fmt.Sprintf("获取节点列表失败: %v", err)
		report.Diagnosis = append(report.Diagnosis, step)
		return err
	}

	taintedNodes := 0
	toleratedNodes := 0

	for _, node := range nodes.Items {
		if len(node.Spec.Taints) > 0 {
			taintedNodes++
			
			// 检查Pod是否能容忍节点污点
			canTolerate := true
			for _, taint := range node.Spec.Taints {
				tolerated := false
				for _, toleration := range pod.Spec.Tolerations {
					if st.tolerationMatches(toleration, taint) {
						tolerated = true
						break
					}
				}
				if !tolerated {
					canTolerate = false
					break
				}
			}
			
			if canTolerate {
				toleratedNodes++
			}
		}
	}

	if taintedNodes > 0 && toleratedNodes == 0 {
		step.Status = "failed"
		step.Details = fmt.Sprintf("集群中有 %d 个污点节点，但Pod无法容忍任何污点", taintedNodes)
	} else {
		step.Status = "passed"
		step.Details = fmt.Sprintf("污点检查通过，可调度到 %d 个节点", toleratedNodes)
	}

	report.Diagnosis = append(report.Diagnosis, step)
	return nil
}

// checkResourceQuotas 检查资源配额
func (st *SchedulerTroubleshooter) checkResourceQuotas(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) error {
	step := DiagnosisStep{
		Step:        "resource_quotas",
		Description: "检查命名空间资源配额",
	}

	quotas, err := st.client.CoreV1().ResourceQuotas(pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "error"
		step.Details = fmt.Sprintf("获取资源配额失败: %v", err)
		report.Diagnosis = append(report.Diagnosis, step)
		return err
	}

	if len(quotas.Items) == 0 {
		step.Status = "passed"
		step.Details = "命名空间未配置资源配额"
	} else {
		quotaIssues := []string{}
		for _, quota := range quotas.Items {
			for resourceName, used := range quota.Status.Used {
				if hard, exists := quota.Status.Hard[resourceName]; exists {
					usedQuantity := used.DeepCopy()
					hardQuantity := hard.DeepCopy()
					
					// 检查是否接近或超过配额
					if usedQuantity.Cmp(hardQuantity) >= 0 {
						quotaIssues = append(quotaIssues, fmt.Sprintf("%s: %s/%s (已满)", resourceName, usedQuantity.String(), hardQuantity.String()))
					}
				}
			}
		}
		
		if len(quotaIssues) > 0 {
			step.Status = "failed"
			step.Details = fmt.Sprintf("资源配额限制: %s", strings.Join(quotaIssues, ", "))
			report.Severity = "high"
		} else {
			step.Status = "passed"
			step.Details = "资源配额检查通过"
		}
	}

	report.Diagnosis = append(report.Diagnosis, step)
	return nil
}

// generateRecommendations 生成故障排除建议
func (st *SchedulerTroubleshooter) generateRecommendations(report *TroubleshootingReport) {
	// 基于诊断结果生成通用建议
	for _, step := range report.Diagnosis {
		if step.Status == "failed" {
			switch step.Step {
			case "node_resources":
				report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
					Action:      "optimize_resources",
					Description: "优化Pod资源请求或扩展集群容量",
					Priority:    "high",
				})
			case "taints_tolerations":
				report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
					Action:      "add_tolerations",
					Description: "为Pod添加适当的污点容忍度",
					Priority:    "medium",
				})
			case "resource_quotas":
				report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
					Action:      "increase_quota",
					Description: "增加命名空间资源配额或清理未使用的资源",
					Priority:    "high",
					Command:     fmt.Sprintf("kubectl describe quota -n %s", report.Namespace),
				})
			}
		}
	}

	// 添加通用建议
	report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
		Action:      "check_events",
		Description: "查看Pod和节点事件以获取更多信息",
		Priority:    "low",
		Command:     fmt.Sprintf("kubectl describe pod %s -n %s", report.PodName, report.Namespace),
	})
}

// tolerationMatches 检查容忍度是否匹配污点
func (st *SchedulerTroubleshooter) tolerationMatches(toleration v1.Toleration, taint v1.Taint) bool {
	// 检查key匹配
	if toleration.Key != "" && toleration.Key != taint.Key {
		return false
	}

	// 检查operator
	if toleration.Operator == v1.TolerationOpExists {
		return true
	}

	// 检查value匹配
	if toleration.Operator == v1.TolerationOpEqual {
		return toleration.Value == taint.Value
	}

	return false
}

// contains 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}