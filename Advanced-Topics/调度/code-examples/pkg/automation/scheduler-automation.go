// scheduler-automation.go
package automation

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"github.com/kubernetes-fundamentals/pkg/scheduler"
)

// AutomationEngine 自动化响应引擎
type AutomationEngine struct {
	client   kubernetes.Interface
	analyzer *scheduler.SchedulerAnalyzer
	config   *AutomationConfig
}

// AutomationConfig 自动化配置
type AutomationConfig struct {
	EnableAutoRemediation bool             `json:"enable_auto_remediation"`
	RemediationCooldown   time.Duration    `json:"remediation_cooldown"`
	MaxRemediationActions int              `json:"max_remediation_actions"`
	Thresholds            *ThresholdConfig `json:"thresholds"`
	Actions               *ActionConfig    `json:"actions"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	HighLatencyThreshold       time.Duration `json:"high_latency_threshold"`
	HighFailureRateThreshold   float64       `json:"high_failure_rate_threshold"`
	LargeQueueThreshold        int           `json:"large_queue_threshold"`
	HighFragmentationThreshold float64       `json:"high_fragmentation_threshold"`
	LowUtilizationThreshold    float64       `json:"low_utilization_threshold"`
	HighUtilizationThreshold   float64       `json:"high_utilization_threshold"`
}

// ActionConfig 动作配置
type ActionConfig struct {
	EnableSchedulerRestart bool `json:"enable_scheduler_restart"`
	EnableNodeDraining     bool `json:"enable_node_draining"`
	EnablePodEviction      bool `json:"enable_pod_eviction"`
	EnableNodeLabeling     bool `json:"enable_node_labeling"`
}

// RemediationAction 修复动作
type RemediationAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  time.Time              `json:"timestamp"`
	Status     string                 `json:"status"`
	Result     string                 `json:"result"`
	Duration   time.Duration          `json:"duration"`
}

// RemediationHistory 修复历史
type RemediationHistory struct {
	Actions []RemediationAction `json:"actions"`
	mutex   sync.RWMutex
}

func NewAutomationEngine(client kubernetes.Interface, analyzer *scheduler.SchedulerAnalyzer, config *AutomationConfig) *AutomationEngine {
	return &AutomationEngine{
		client:   client,
		analyzer: analyzer,
		config:   config,
	}
}

func (ae *AutomationEngine) StartAutomationLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	history := &RemediationHistory{
		Actions: make([]RemediationAction, 0),
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := ae.processAutomationCycle(ctx, history); err != nil {
				klog.Errorf("Automation cycle failed: %v", err)
			}
		}
	}
}

func (ae *AutomationEngine) processAutomationCycle(ctx context.Context, history *RemediationHistory) error {
	// 生成分析报告
	report, err := ae.analyzer.GenerateAnalysisReport(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate analysis report: %v", err)
	}

	// 检查是否需要自动修复
	if !ae.config.EnableAutoRemediation {
		return nil
	}

	// 检查冷却期
	if ae.isInCooldownPeriod(history) {
		klog.V(4).Info("Automation is in cooldown period, skipping remediation")
		return nil
	}

	// 检查修复动作限制
	if ae.hasExceededActionLimit(history) {
		klog.Warning("Maximum remediation actions reached, skipping further actions")
		return nil
	}

	// 执行自动修复
	actions := ae.determineRemediationActions(report)
	for _, action := range actions {
		if err := ae.executeRemediationAction(ctx, action, history); err != nil {
			klog.Errorf("Failed to execute remediation action %s: %v", action.Type, err)
		}
	}

	return nil
}

func (ae *AutomationEngine) determineRemediationActions(report *scheduler.AnalysisReport) []RemediationAction {
	var actions []RemediationAction

	// 基于性能指标的自动修复
	if report.PerformanceMetrics.P95SchedulingLatency > ae.config.Thresholds.HighLatencyThreshold {
		actions = append(actions, RemediationAction{
			Type:   "optimize_scheduler_config",
			Target: "scheduler",
			Parameters: map[string]interface{}{
				"reason":  "high_latency",
				"latency": report.PerformanceMetrics.P95SchedulingLatency.String(),
			},
			Timestamp: time.Now(),
			Status:    "pending",
		})
	}

	if report.PerformanceMetrics.FailureRate > ae.config.Thresholds.HighFailureRateThreshold {
		actions = append(actions, RemediationAction{
			Type:   "investigate_failures",
			Target: "scheduler",
			Parameters: map[string]interface{}{
				"reason":       "high_failure_rate",
				"failure_rate": report.PerformanceMetrics.FailureRate,
			},
			Timestamp: time.Now(),
			Status:    "pending",
		})
	}

	if report.PerformanceMetrics.QueueLength > ae.config.Thresholds.LargeQueueThreshold {
		actions = append(actions, RemediationAction{
			Type:   "drain_problematic_nodes",
			Target: "cluster",
			Parameters: map[string]interface{}{
				"reason":       "large_queue",
				"queue_length": report.PerformanceMetrics.QueueLength,
			},
			Timestamp: time.Now(),
			Status:    "pending",
		})
	}

	// 基于资源分析的自动修复
	if report.ResourceAnalysis.Fragmentation["cpu"] > ae.config.Thresholds.HighFragmentationThreshold {
		actions = append(actions, RemediationAction{
			Type:   "defragment_resources",
			Target: "cluster",
			Parameters: map[string]interface{}{
				"reason":        "high_fragmentation",
				"fragmentation": report.ResourceAnalysis.Fragmentation["cpu"],
				"resource":      "cpu",
			},
			Timestamp: time.Now(),
			Status:    "pending",
		})
	}

	// 处理热点节点
	if len(report.ResourceAnalysis.HotSpots) > 0 {
		for _, hotspot := range report.ResourceAnalysis.HotSpots {
			actions = append(actions, RemediationAction{
				Type:   "rebalance_node",
				Target: hotspot,
				Parameters: map[string]interface{}{
					"reason": "resource_hotspot",
					"node":   hotspot,
				},
				Timestamp: time.Now(),
				Status:    "pending",
			})
		}
	}

	// 处理低利用率节点
	if len(report.ResourceAnalysis.UnderutilizedNodes) > 2 {
		actions = append(actions, RemediationAction{
			Type:   "consolidate_workloads",
			Target: "cluster",
			Parameters: map[string]interface{}{
				"reason":              "low_utilization",
				"underutilized_nodes": report.ResourceAnalysis.UnderutilizedNodes,
			},
			Timestamp: time.Now(),
			Status:    "pending",
		})
	}

	return actions
}

func (ae *AutomationEngine) executeRemediationAction(ctx context.Context, action RemediationAction, history *RemediationHistory) error {
	start := time.Now()
	action.Status = "executing"

	defer func() {
		action.Duration = time.Since(start)
		history.mutex.Lock()
		history.Actions = append(history.Actions, action)
		history.mutex.Unlock()
	}()

	switch action.Type {
	case "optimize_scheduler_config":
		return ae.optimizeSchedulerConfig(ctx, action)
	case "investigate_failures":
		return ae.investigateFailures(ctx, action)
	case "drain_problematic_nodes":
		return ae.drainProblematicNodes(ctx, action)
	case "defragment_resources":
		return ae.defragmentResources(ctx, action)
	case "rebalance_node":
		return ae.rebalanceNode(ctx, action)
	case "consolidate_workloads":
		return ae.consolidateWorkloads(ctx, action)
	default:
		action.Status = "failed"
		action.Result = fmt.Sprintf("Unknown action type: %s", action.Type)
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (ae *AutomationEngine) optimizeSchedulerConfig(ctx context.Context, action RemediationAction) error {
	if !ae.config.Actions.EnableSchedulerRestart {
		action.Status = "skipped"
		action.Result = "Scheduler restart is disabled"
		return nil
	}

	klog.Infof("Optimizing scheduler configuration due to %s", action.Parameters["reason"])

	// 这里可以实现调度器配置优化逻辑
	// 例如：调整 percentageOfNodesToScore、修改插件配置等

	action.Status = "completed"
	action.Result = "Scheduler configuration optimized"
	return nil
}

func (ae *AutomationEngine) investigateFailures(ctx context.Context, action RemediationAction) error {
	klog.Infof("Investigating scheduling failures, failure rate: %v", action.Parameters["failure_rate"])

	// 收集失败的Pod信息
	pendingPods, err := ae.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Pending",
	})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to list pending pods: %v", err)
		return err
	}

	failureReasons := make(map[string]int)
	for _, pod := range pendingPods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionFalse {
				failureReasons[condition.Reason]++
			}
		}
	}

	klog.Infof("Scheduling failure analysis: %+v", failureReasons)

	action.Status = "completed"
	action.Result = fmt.Sprintf("Analyzed %d pending pods, failure reasons: %+v", len(pendingPods.Items), failureReasons)
	return nil
}

func (ae *AutomationEngine) drainProblematicNodes(ctx context.Context, action RemediationAction) error {
	if !ae.config.Actions.EnableNodeDraining {
		action.Status = "skipped"
		action.Result = "Node draining is disabled"
		return nil
	}

	klog.Infof("Identifying problematic nodes for draining, queue length: %v", action.Parameters["queue_length"])

	// 识别有问题的节点（例如：NotReady状态的节点）
	nodes, err := ae.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to list nodes: %v", err)
		return err
	}

	problematicNodes := []string{}
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
				problematicNodes = append(problematicNodes, node.Name)
				break
			}
		}
	}

	if len(problematicNodes) == 0 {
		action.Status = "completed"
		action.Result = "No problematic nodes found"
		return nil
	}

	// 这里可以实现节点排空逻辑
	klog.Infof("Found %d problematic nodes: %v", len(problematicNodes), problematicNodes)

	action.Status = "completed"
	action.Result = fmt.Sprintf("Identified %d problematic nodes for potential draining", len(problematicNodes))
	return nil
}

func (ae *AutomationEngine) defragmentResources(ctx context.Context, action RemediationAction) error {
	if !ae.config.Actions.EnablePodEviction {
		action.Status = "skipped"
		action.Result = "Pod eviction is disabled"
		return nil
	}

	klog.Infof("Defragmenting %s resources, fragmentation: %v",
		action.Parameters["resource"], action.Parameters["fragmentation"])

	// 这里可以实现资源碎片整理逻辑
	// 例如：识别小的Pod并重新调度到更合适的节点

	action.Status = "completed"
	action.Result = "Resource defragmentation analysis completed"
	return nil
}

func (ae *AutomationEngine) rebalanceNode(ctx context.Context, action RemediationAction) error {
	if !ae.config.Actions.EnableNodeLabeling {
		action.Status = "skipped"
		action.Result = "Node labeling is disabled"
		return nil
	}

	nodeName := action.Parameters["node"].(string)
	klog.Infof("Rebalancing node %s due to resource hotspot", nodeName)

	// 为热点节点添加标签，防止新Pod调度到该节点
	node, err := ae.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to get node %s: %v", nodeName, err)
		return err
	}

	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	node.Labels["scheduler.kubernetes.io/hotspot"] = "true"
	node.Labels["scheduler.kubernetes.io/hotspot-timestamp"] = fmt.Sprintf("%d", time.Now().Unix())

	_, err = ae.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to update node labels: %v", err)
		return err
	}

	action.Status = "completed"
	action.Result = fmt.Sprintf("Added hotspot labels to node %s", nodeName)
	return nil
}

func (ae *AutomationEngine) consolidateWorkloads(ctx context.Context, action RemediationAction) error {
	klog.Infof("Analyzing workload consolidation opportunities")

	underutilizedNodes := action.Parameters["underutilized_nodes"].([]string)

	// 这里可以实现工作负载整合逻辑
	// 例如：建议将某些节点上的Pod迁移到其他节点

	action.Status = "completed"
	action.Result = fmt.Sprintf("Analyzed %d underutilized nodes for consolidation", len(underutilizedNodes))
	return nil
}

func (ae *AutomationEngine) isInCooldownPeriod(history *RemediationHistory) bool {
	history.mutex.RLock()
	defer history.mutex.RUnlock()

	if len(history.Actions) == 0 {
		return false
	}

	lastAction := history.Actions[len(history.Actions)-1]
	return time.Since(lastAction.Timestamp) < ae.config.RemediationCooldown
}

func (ae *AutomationEngine) hasExceededActionLimit(history *RemediationHistory) bool {
	history.mutex.RLock()
	defer history.mutex.RUnlock()

	// 计算最近1小时内的动作数量
	oneHourAgo := time.Now().Add(-time.Hour)
	recentActions := 0

	for _, action := range history.Actions {
		if action.Timestamp.After(oneHourAgo) {
			recentActions++
		}
	}

	return recentActions >= ae.config.MaxRemediationActions
}

// 清理过期的热点标签
func (ae *AutomationEngine) cleanupHotspotLabels(ctx context.Context) error {
	nodes, err := ae.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "scheduler.kubernetes.io/hotspot=true",
	})
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		timestampStr, exists := node.Labels["scheduler.kubernetes.io/hotspot-timestamp"]
		if !exists {
			continue
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue
		}

		// 如果标签超过30分钟，则移除
		if time.Since(time.Unix(timestamp, 0)) > 30*time.Minute {
			delete(node.Labels, "scheduler.kubernetes.io/hotspot")
			delete(node.Labels, "scheduler.kubernetes.io/hotspot-timestamp")

			_, err = ae.client.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Failed to remove hotspot labels from node %s: %v", node.Name, err)
			} else {
				klog.Infof("Removed expired hotspot labels from node %s", node.Name)
			}
		}
	}

	return nil
}
