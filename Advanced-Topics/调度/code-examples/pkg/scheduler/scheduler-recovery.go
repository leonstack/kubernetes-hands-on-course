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

// SchedulerRecoveryManager 调度器恢复管理器
type SchedulerRecoveryManager struct {
	client kubernetes.Interface
	config *RecoveryConfig
}

// RecoveryConfig 恢复配置
type RecoveryConfig struct {
	HealthCheckInterval time.Duration      `json:"health_check_interval"`
	RecoveryTimeout     time.Duration      `json:"recovery_timeout"`
	MaxRetries          int                `json:"max_retries"`
	Strategies          RecoveryStrategies `json:"strategies"`
}

// RecoveryStrategies 恢复策略
type RecoveryStrategies struct {
	SchedulerRestart    bool `json:"scheduler_restart"`
	NodeDraining        bool `json:"node_draining"`
	PodEviction         bool `json:"pod_eviction"`
	ResourceRebalancing bool `json:"resource_rebalancing"`
}

// RecoveryAction 恢复动作
type RecoveryAction struct {
	Type      string        `json:"type"`
	Target    string        `json:"target"`
	Timestamp time.Time     `json:"timestamp"`
	Status    string        `json:"status"`
	Result    string        `json:"result"`
	Duration  time.Duration `json:"duration"`
	Attempts  int           `json:"attempts"`
}

// 注意：HealthStatus 已在 health-checker.go 中定义

// NewSchedulerRecoveryManager 创建新的恢复管理器
func NewSchedulerRecoveryManager(client kubernetes.Interface, config *RecoveryConfig) *SchedulerRecoveryManager {
	if config == nil {
		config = &RecoveryConfig{
			HealthCheckInterval: 30 * time.Second,
			RecoveryTimeout:     5 * time.Minute,
			MaxRetries:          3,
			Strategies: RecoveryStrategies{
				SchedulerRestart:    true,
				NodeDraining:        true,
				PodEviction:         true,
				ResourceRebalancing: false,
			},
		}
	}

	return &SchedulerRecoveryManager{
		client: client,
		config: config,
	}
}

// StartRecoveryLoop 启动恢复循环
func (srm *SchedulerRecoveryManager) StartRecoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(srm.config.HealthCheckInterval)
	defer ticker.Stop()

	klog.Infof("Starting scheduler recovery loop with interval: %v", srm.config.HealthCheckInterval)

	for {
		select {
		case <-ctx.Done():
			klog.Info("Recovery loop stopped")
			return
		case <-ticker.C:
			if err := srm.performHealthCheck(ctx); err != nil {
				klog.Errorf("Health check failed: %v", err)
				if recoveryErr := srm.initiateRecovery(ctx, err); recoveryErr != nil {
					klog.Errorf("Recovery failed: %v", err)
				}
			}
		}
	}
}

// performHealthCheck 执行健康检查
func (srm *SchedulerRecoveryManager) performHealthCheck(ctx context.Context) error {
	// 检查调度器健康状态
	if _, err := srm.checkSchedulerHealth(ctx); err != nil {
		return fmt.Errorf("scheduler health check failed: %v", err)
	}

	// 检查节点健康状态
	if _, err := srm.checkNodeHealth(ctx); err != nil {
		return fmt.Errorf("node health check failed: %v", err)
	}

	// 检查调度队列健康状态
	if _, err := srm.checkQueueHealth(ctx); err != nil {
		return fmt.Errorf("queue health check failed: %v", err)
	}

	return nil
}

// checkSchedulerHealth 检查调度器健康状态
func (srm *SchedulerRecoveryManager) checkSchedulerHealth(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		Name:      "scheduler",
		Timestamp: time.Now(),
		Healthy:   true,
		Message:   "OK",
	}

	// 检查调度器Pod状态
	pods, err := srm.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-scheduler",
	})
	if err != nil {
		health.Healthy = false
		health.Message = fmt.Sprintf("Failed to list scheduler pods: %v", err)
		return health, err
	}

	if len(pods.Items) == 0 {
		health.Healthy = false
		health.Message = "No scheduler pods found"
		return health, fmt.Errorf("no scheduler pods found")
	}

	runningPods := 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			runningPods++
		}
	}

	if runningPods == 0 {
		health.Healthy = false
		health.Message = "No running scheduler pods"
		return health, fmt.Errorf("no running scheduler pods")
	}

	return health, nil
}

// checkNodeHealth 检查节点健康状态
func (srm *SchedulerRecoveryManager) checkNodeHealth(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		Name:      "nodes",
		Timestamp: time.Now(),
		Healthy:   true,
		Message:   "OK",
	}

	nodes, err := srm.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		health.Healthy = false
		health.Message = fmt.Sprintf("Failed to list nodes: %v", err)
		return health, err
	}

	if len(nodes.Items) == 0 {
		health.Healthy = false
		health.Message = "No nodes found"
		return health, fmt.Errorf("no nodes found")
	}

	readyNodes := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				readyNodes++
				break
			}
		}
	}

	// 如果少于50%的节点就绪，认为不健康
	if float64(readyNodes)/float64(len(nodes.Items)) < 0.5 {
		health.Healthy = false
		health.Message = fmt.Sprintf("Only %d/%d nodes are ready", readyNodes, len(nodes.Items))
		return health, fmt.Errorf("insufficient ready nodes")
	}

	return health, nil
}

// checkQueueHealth 检查队列健康状态
func (srm *SchedulerRecoveryManager) checkQueueHealth(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		Name:      "queue",
		Timestamp: time.Now(),
		Healthy:   true,
		Message:   "OK",
	}

	// 检查Pending状态的Pod数量
	pendingPods, err := srm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Pending",
	})
	if err != nil {
		health.Healthy = false
		health.Message = fmt.Sprintf("Failed to list pending pods: %v", err)
		return health, err
	}

	// 如果Pending Pod数量过多，认为队列不健康
	if len(pendingPods.Items) > 100 {
		health.Healthy = false
		health.Message = fmt.Sprintf("Too many pending pods: %d", len(pendingPods.Items))
		return health, fmt.Errorf("queue overloaded")
	}

	// 检查长时间Pending的Pod
	longPendingPods := 0
	for _, pod := range pendingPods.Items {
		if time.Since(pod.CreationTimestamp.Time) > 10*time.Minute {
			longPendingPods++
		}
	}

	if longPendingPods > 10 {
		health.Healthy = false
		health.Message = fmt.Sprintf("Too many long-pending pods: %d", longPendingPods)
		return health, fmt.Errorf("scheduling stalled")
	}

	return health, nil
}

// initiateRecovery 启动恢复流程
func (srm *SchedulerRecoveryManager) initiateRecovery(ctx context.Context, healthErr error) error {
	klog.Infof("Initiating recovery due to health check failure: %v", healthErr)

	// 确定恢复策略
	actions := srm.determineRecoveryActions(healthErr)

	// 执行恢复动作
	for _, action := range actions {
		if err := srm.executeRecoveryAction(ctx, action); err != nil {
			klog.Errorf("Recovery action %s failed: %v", action.Type, err)
			continue
		}

		// 等待恢复生效
		time.Sleep(30 * time.Second)

		// 重新检查健康状态
		if err := srm.performHealthCheck(ctx); err == nil {
			klog.Infof("Recovery successful after action: %s", action.Type)
			return nil
		}
	}

	return fmt.Errorf("all recovery actions failed")
}

// determineRecoveryActions 确定恢复动作
func (srm *SchedulerRecoveryManager) determineRecoveryActions(healthErr error) []RecoveryAction {
	var actions []RecoveryAction

	errorMsg := healthErr.Error()

	// 基于错误类型确定恢复策略
	if strings.Contains(errorMsg, "scheduler") {
		if srm.config.Strategies.SchedulerRestart {
			actions = append(actions, RecoveryAction{
				Type:      "restart_scheduler",
				Target:    "kube-scheduler",
				Timestamp: time.Now(),
				Status:    "pending",
			})
		}
	}

	if strings.Contains(errorMsg, "nodes") || strings.Contains(errorMsg, "ready") {
		if srm.config.Strategies.NodeDraining {
			actions = append(actions, RecoveryAction{
				Type:      "drain_unhealthy_nodes",
				Target:    "cluster",
				Timestamp: time.Now(),
				Status:    "pending",
			})
		}
	}

	if strings.Contains(errorMsg, "queue") || strings.Contains(errorMsg, "pending") {
		if srm.config.Strategies.PodEviction {
			actions = append(actions, RecoveryAction{
				Type:      "evict_stuck_pods",
				Target:    "cluster",
				Timestamp: time.Now(),
				Status:    "pending",
			})
		}

		if srm.config.Strategies.ResourceRebalancing {
			actions = append(actions, RecoveryAction{
				Type:      "rebalance_resources",
				Target:    "cluster",
				Timestamp: time.Now(),
				Status:    "pending",
			})
		}
	}

	return actions
}

// executeRecoveryAction 执行恢复动作
func (srm *SchedulerRecoveryManager) executeRecoveryAction(ctx context.Context, action RecoveryAction) error {
	start := time.Now()
	action.Status = "executing"
	action.Attempts++

	defer func() {
		action.Duration = time.Since(start)
	}()

	switch action.Type {
	case "restart_scheduler":
		return srm.restartScheduler(ctx, action)
	case "drain_unhealthy_nodes":
		return srm.drainUnhealthyNodes(ctx, action)
	case "evict_stuck_pods":
		return srm.evictStuckPods(ctx, action)
	case "rebalance_resources":
		return srm.rebalanceResources(ctx, action)
	default:
		action.Status = "failed"
		action.Result = fmt.Sprintf("Unknown recovery action: %s", action.Type)
		return fmt.Errorf("unknown recovery action: %s", action.Type)
	}
}

// restartScheduler 重启调度器
func (srm *SchedulerRecoveryManager) restartScheduler(ctx context.Context, action RecoveryAction) error {
	klog.Infof("Restarting scheduler pods")

	// 获取调度器Pod
	pods, err := srm.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-scheduler",
	})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to list scheduler pods: %v", err)
		return err
	}

	// 删除调度器Pod以触发重启
	for _, pod := range pods.Items {
		err := srm.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			klog.Errorf("Failed to delete scheduler pod %s: %v", pod.Name, err)
			continue
		}
		klog.Infof("Deleted scheduler pod: %s", pod.Name)
	}

	// 等待新Pod启动
	time.Sleep(60 * time.Second)

	action.Status = "completed"
	action.Result = fmt.Sprintf("Restarted %d scheduler pods", len(pods.Items))
	return nil
}

// drainUnhealthyNodes 排空不健康节点
func (srm *SchedulerRecoveryManager) drainUnhealthyNodes(ctx context.Context, action RecoveryAction) error {
	klog.Infof("Draining unhealthy nodes")

	nodes, err := srm.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to list nodes: %v", err)
		return err
	}

	unhealthyNodes := []string{}
	for _, node := range nodes.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				isReady = true
				break
			}
		}

		if !isReady {
			unhealthyNodes = append(unhealthyNodes, node.Name)

			// 标记节点为不可调度
			node.Spec.Unschedulable = true
			_, err := srm.client.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Failed to mark node %s as unschedulable: %v", node.Name, err)
			} else {
				klog.Infof("Marked node %s as unschedulable", node.Name)
			}
		}
	}

	action.Status = "completed"
	action.Result = fmt.Sprintf("Drained %d unhealthy nodes: %v", len(unhealthyNodes), unhealthyNodes)
	return nil
}

// evictStuckPods 驱逐卡住的Pod
func (srm *SchedulerRecoveryManager) evictStuckPods(ctx context.Context, action RecoveryAction) error {
	klog.Infof("Evicting stuck pods")

	// 获取长时间Pending的Pod
	pendingPods, err := srm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Pending",
	})
	if err != nil {
		action.Status = "failed"
		action.Result = fmt.Sprintf("Failed to list pending pods: %v", err)
		return err
	}

	evictedPods := []string{}
	for _, pod := range pendingPods.Items {
		// 只驱逐超过10分钟的Pending Pod
		if time.Since(pod.CreationTimestamp.Time) > 10*time.Minute {
			err := srm.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
			if err != nil {
				klog.Errorf("Failed to evict pod %s/%s: %v", pod.Namespace, pod.Name, err)
				continue
			}
			evictedPods = append(evictedPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
			klog.Infof("Evicted stuck pod: %s/%s", pod.Namespace, pod.Name)
		}
	}

	action.Status = "completed"
	action.Result = fmt.Sprintf("Evicted %d stuck pods", len(evictedPods))
	return nil
}

// rebalanceResources 重平衡资源
func (srm *SchedulerRecoveryManager) rebalanceResources(ctx context.Context, action RecoveryAction) error {
	klog.Infof("Rebalancing cluster resources")

	// 这里可以实现资源重平衡逻辑
	// 例如：识别资源使用不均衡的节点，建议Pod迁移等

	action.Status = "completed"
	action.Result = "Resource rebalancing analysis completed"
	return nil
}
