// scheduler-troubleshooter.go
package troubleshooter

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SchedulerTroubleshooter struct {
	client kubernetes.Interface
}

type TroubleshootingReport struct {
	PodName         string                          `json:"pod_name"`
	Namespace       string                          `json:"namespace"`
	IssueType       string                          `json:"issue_type"`
	Severity        string                          `json:"severity"`
	Description     string                          `json:"description"`
	Diagnosis       []DiagnosisStep                 `json:"diagnosis"`
	Recommendations []TroubleshootingRecommendation `json:"recommendations"`
	Timestamp       time.Time                       `json:"timestamp"`
}

type DiagnosisStep struct {
	Step     string        `json:"step"`
	Status   string        `json:"status"`
	Details  interface{}   `json:"details"`
	Duration time.Duration `json:"duration"`
}

type TroubleshootingRecommendation struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Priority    string `json:"priority"`
}

func NewSchedulerTroubleshooter(client kubernetes.Interface) *SchedulerTroubleshooter {
	return &SchedulerTroubleshooter{
		client: client,
	}
}

func (st *SchedulerTroubleshooter) DiagnosePendingPod(ctx context.Context, namespace, podName string) (*TroubleshootingReport, error) {
	report := &TroubleshootingReport{
		PodName:         podName,
		Namespace:       namespace,
		Timestamp:       time.Now(),
		Diagnosis:       []DiagnosisStep{},
		Recommendations: []TroubleshootingRecommendation{},
	}

	// 获取Pod信息
	pod, err := st.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	// 检查Pod状态
	if pod.Status.Phase != v1.PodPending {
		report.IssueType = "NotPending"
		report.Severity = "Info"
		report.Description = fmt.Sprintf("Pod is in %s phase, not pending", pod.Status.Phase)
		return report, nil
	}

	// 执行诊断步骤
	st.checkPodSchedulingConditions(ctx, pod, report)
	st.checkResourceRequirements(ctx, pod, report)
	st.checkNodeSelector(ctx, pod, report)
	st.checkAffinity(ctx, pod, report)
	st.checkTolerations(ctx, pod, report)
	st.checkPodSecurityPolicy(ctx, pod, report)
	st.checkResourceQuotas(ctx, pod, report)
	st.checkNodeAvailability(ctx, pod, report)

	// 确定问题类型和严重性
	st.categorizeIssue(report)

	// 生成建议
	st.generateRecommendations(report)

	return report, nil
}

func (st *SchedulerTroubleshooter) checkPodSchedulingConditions(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckSchedulingConditions",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	conditions := make(map[string]string)

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled {
			conditions["PodScheduled"] = fmt.Sprintf("Status: %s, Reason: %s, Message: %s",
				condition.Status, condition.Reason, condition.Message)

			if condition.Status == v1.ConditionFalse {
				step.Status = "Failed"
				step.Details = map[string]interface{}{
					"reason":  condition.Reason,
					"message": condition.Message,
				}
				return
			}
		}
	}

	step.Status = "Completed"
	step.Details = conditions
}

func (st *SchedulerTroubleshooter) checkResourceRequirements(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckResourceRequirements",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	var totalCPU, totalMemory int64
	resourceDetails := make(map[string]interface{})

	for _, container := range pod.Spec.Containers {
		if cpu := container.Resources.Requests.Cpu(); cpu != nil {
			totalCPU += cpu.MilliValue()
		}
		if memory := container.Resources.Requests.Memory(); memory != nil {
			totalMemory += memory.Value()
		}
	}

	resourceDetails["total_cpu_request"] = fmt.Sprintf("%dm", totalCPU)
	resourceDetails["total_memory_request"] = fmt.Sprintf("%dMi", totalMemory/(1024*1024))

	// 检查是否有节点能满足资源需求
	nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "Error"
		step.Details = map[string]interface{}{"error": err.Error()}
		return
	}

	suitableNodes := 0
	for _, node := range nodes.Items {
		allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
		allocatableMemory := node.Status.Allocatable.Memory().Value()

		if allocatableCPU >= totalCPU && allocatableMemory >= totalMemory {
			suitableNodes++
		}
	}

	resourceDetails["suitable_nodes"] = suitableNodes
	resourceDetails["total_nodes"] = len(nodes.Items)

	if suitableNodes == 0 {
		step.Status = "Failed"
		resourceDetails["issue"] = "No nodes have sufficient resources"
	} else {
		step.Status = "Completed"
	}

	step.Details = resourceDetails
}

func (st *SchedulerTroubleshooter) checkNodeSelector(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckNodeSelector",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	if len(pod.Spec.NodeSelector) == 0 {
		step.Status = "Skipped"
		step.Details = "No node selector specified"
		return
	}

	nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "Error"
		step.Details = map[string]interface{}{"error": err.Error()}
		return
	}

	matchingNodes := 0
	for _, node := range nodes.Items {
		matches := true
		for key, value := range pod.Spec.NodeSelector {
			if nodeValue, exists := node.Labels[key]; !exists || nodeValue != value {
				matches = false
				break
			}
		}
		if matches {
			matchingNodes++
		}
	}

	details := map[string]interface{}{
		"node_selector":  pod.Spec.NodeSelector,
		"matching_nodes": matchingNodes,
		"total_nodes":    len(nodes.Items),
	}

	if matchingNodes == 0 {
		step.Status = "Failed"
		details["issue"] = "No nodes match the node selector"
	} else {
		step.Status = "Completed"
	}

	step.Details = details
}

func (st *SchedulerTroubleshooter) checkAffinity(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckAffinity",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	if pod.Spec.Affinity == nil {
		step.Status = "Skipped"
		step.Details = "No affinity rules specified"
		return
	}

	details := make(map[string]interface{})

	// 检查节点亲和性
	if pod.Spec.Affinity.NodeAffinity != nil {
		details["node_affinity"] = "present"

		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			details["required_node_affinity"] = "present"
		}

		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			details["preferred_node_affinity"] = len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
		}
	}

	// 检查Pod亲和性
	if pod.Spec.Affinity.PodAffinity != nil {
		details["pod_affinity"] = "present"
	}

	// 检查Pod反亲和性
	if pod.Spec.Affinity.PodAntiAffinity != nil {
		details["pod_anti_affinity"] = "present"
	}

	step.Status = "Completed"
	step.Details = details
}

func (st *SchedulerTroubleshooter) checkTolerations(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckTolerations",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	if len(pod.Spec.Tolerations) == 0 {
		step.Status = "Skipped"
		step.Details = "No tolerations specified"
		return
	}

	nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "Error"
		step.Details = map[string]interface{}{"error": err.Error()}
		return
	}

	tolerableNodes := 0
	taintedNodes := 0

	for _, node := range nodes.Items {
		if len(node.Spec.Taints) > 0 {
			taintedNodes++

			canTolerate := true
			for _, taint := range node.Spec.Taints {
				if !st.canTolerateTaint(pod.Spec.Tolerations, taint) {
					canTolerate = false
					break
				}
			}

			if canTolerate {
				tolerableNodes++
			}
		} else {
			tolerableNodes++
		}
	}

	details := map[string]interface{}{
		"tolerations":     len(pod.Spec.Tolerations),
		"tainted_nodes":   taintedNodes,
		"tolerable_nodes": tolerableNodes,
		"total_nodes":     len(nodes.Items),
	}

	if tolerableNodes == 0 {
		step.Status = "Failed"
		details["issue"] = "Pod cannot tolerate taints on any node"
	} else {
		step.Status = "Completed"
	}

	step.Details = details
}

func (st *SchedulerTroubleshooter) canTolerateTaint(tolerations []v1.Toleration, taint v1.Taint) bool {
	for _, toleration := range tolerations {
		if toleration.Key == taint.Key {
			if toleration.Operator == v1.TolerationOpExists {
				return true
			}
			if toleration.Operator == v1.TolerationOpEqual && toleration.Value == taint.Value {
				return true
			}
		}
	}
	return false
}

func (st *SchedulerTroubleshooter) checkPodSecurityPolicy(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckPodSecurityPolicy",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	// 简化的PSP检查
	details := make(map[string]interface{})

	if pod.Spec.SecurityContext != nil {
		details["security_context"] = "present"

		if pod.Spec.SecurityContext.RunAsUser != nil {
			details["run_as_user"] = *pod.Spec.SecurityContext.RunAsUser
		}

		if pod.Spec.SecurityContext.RunAsGroup != nil {
			details["run_as_group"] = *pod.Spec.SecurityContext.RunAsGroup
		}

		if pod.Spec.SecurityContext.FSGroup != nil {
			details["fs_group"] = *pod.Spec.SecurityContext.FSGroup
		}
	}

	step.Status = "Completed"
	step.Details = details
}

func (st *SchedulerTroubleshooter) checkResourceQuotas(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckResourceQuotas",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	quotas, err := st.client.CoreV1().ResourceQuotas(pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "Error"
		step.Details = map[string]interface{}{"error": err.Error()}
		return
	}

	if len(quotas.Items) == 0 {
		step.Status = "Skipped"
		step.Details = "No resource quotas in namespace"
		return
	}

	details := map[string]interface{}{
		"quotas_count": len(quotas.Items),
	}

	quotaIssues := []string{}
	for _, quota := range quotas.Items {
		for resource, used := range quota.Status.Used {
			if hard, exists := quota.Status.Hard[resource]; exists {
				if used.Cmp(hard) >= 0 {
					quotaIssues = append(quotaIssues, fmt.Sprintf("%s: %s/%s", resource, used.String(), hard.String()))
				}
			}
		}
	}

	if len(quotaIssues) > 0 {
		step.Status = "Failed"
		details["quota_issues"] = quotaIssues
	} else {
		step.Status = "Completed"
	}

	step.Details = details
}

func (st *SchedulerTroubleshooter) checkNodeAvailability(ctx context.Context, pod *v1.Pod, report *TroubleshootingReport) {
	start := time.Now()
	step := DiagnosisStep{
		Step:   "CheckNodeAvailability",
		Status: "Running",
	}

	defer func() {
		step.Duration = time.Since(start)
		report.Diagnosis = append(report.Diagnosis, step)
	}()

	nodes, err := st.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		step.Status = "Error"
		step.Details = map[string]interface{}{"error": err.Error()}
		return
	}

	readyNodes := 0
	schedulableNodes := 0

	for _, node := range nodes.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				isReady = true
				readyNodes++
				break
			}
		}

		if isReady && !node.Spec.Unschedulable {
			schedulableNodes++
		}
	}

	details := map[string]interface{}{
		"total_nodes":       len(nodes.Items),
		"ready_nodes":       readyNodes,
		"schedulable_nodes": schedulableNodes,
	}

	if schedulableNodes == 0 {
		step.Status = "Failed"
		details["issue"] = "No schedulable nodes available"
	} else {
		step.Status = "Completed"
	}

	step.Details = details
}

func (st *SchedulerTroubleshooter) categorizeIssue(report *TroubleshootingReport) {
	failedSteps := []string{}

	for _, step := range report.Diagnosis {
		if step.Status == "Failed" {
			failedSteps = append(failedSteps, step.Step)
		}
	}

	if len(failedSteps) == 0 {
		report.IssueType = "Unknown"
		report.Severity = "Low"
		report.Description = "No obvious scheduling issues detected"
		return
	}

	// 根据失败的步骤确定问题类型
	if contains(failedSteps, "CheckNodeAvailability") {
		report.IssueType = "NodeUnavailability"
		report.Severity = "Critical"
		report.Description = "No schedulable nodes available"
	} else if contains(failedSteps, "CheckResourceRequirements") {
		report.IssueType = "InsufficientResources"
		report.Severity = "High"
		report.Description = "No nodes have sufficient resources"
	} else if contains(failedSteps, "CheckNodeSelector") {
		report.IssueType = "NodeSelectorMismatch"
		report.Severity = "Medium"
		report.Description = "No nodes match the node selector"
	} else if contains(failedSteps, "CheckTolerations") {
		report.IssueType = "TaintTolerationMismatch"
		report.Severity = "Medium"
		report.Description = "Pod cannot tolerate node taints"
	} else if contains(failedSteps, "CheckResourceQuotas") {
		report.IssueType = "ResourceQuotaExceeded"
		report.Severity = "Medium"
		report.Description = "Resource quota limits exceeded"
	} else {
		report.IssueType = "SchedulingConstraints"
		report.Severity = "Medium"
		report.Description = "Pod scheduling constraints cannot be satisfied"
	}
}

func (st *SchedulerTroubleshooter) generateRecommendations(report *TroubleshootingReport) {
	switch report.IssueType {
	case "NodeUnavailability":
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Check node status",
			Description: "Verify that nodes are ready and schedulable",
			Command:     "kubectl get nodes",
			Priority:    "High",
		})
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Check node conditions",
			Description: "Examine node conditions for issues",
			Command:     "kubectl describe nodes",
			Priority:    "High",
		})

	case "InsufficientResources":
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Scale cluster",
			Description: "Add more nodes to the cluster",
			Priority:    "High",
		})
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Optimize resource requests",
			Description: "Review and optimize pod resource requests",
			Priority:    "Medium",
		})

	case "NodeSelectorMismatch":
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Update node selector",
			Description: "Modify or remove node selector constraints",
			Priority:    "Medium",
		})
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Label nodes",
			Description: "Add required labels to nodes",
			Command:     "kubectl label nodes <node-name> <key>=<value>",
			Priority:    "Medium",
		})

	case "TaintTolerationMismatch":
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Add tolerations",
			Description: "Add appropriate tolerations to the pod",
			Priority:    "Medium",
		})
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Remove taints",
			Description: "Remove unnecessary taints from nodes",
			Command:     "kubectl taint nodes <node-name> <key>-",
			Priority:    "Low",
		})

	case "ResourceQuotaExceeded":
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Increase quota",
			Description: "Increase resource quota limits",
			Priority:    "Medium",
		})
		report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
			Action:      "Clean up resources",
			Description: "Remove unused resources to free up quota",
			Priority:    "Medium",
		})
	}

	// 通用建议
	report.Recommendations = append(report.Recommendations, TroubleshootingRecommendation{
		Action:      "Check scheduler logs",
		Description: "Examine scheduler logs for detailed error messages",
		Command:     "kubectl logs -n kube-system -l component=kube-scheduler",
		Priority:    "Low",
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
