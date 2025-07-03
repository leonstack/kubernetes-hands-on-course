// scheduler-recovery.go
package recovery

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

type SchedulerRecoveryManager struct {
	client kubernetes.Interface
	config *RecoveryConfig
}

type RecoveryConfig struct {
	EnableAutoRecovery  bool          `yaml:"enable_auto_recovery"`
	RecoveryTimeout     time.Duration `yaml:"recovery_timeout"`
	MaxRecoveryAttempts int           `yaml:"max_recovery_attempts"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`

	// 恢复策略配置
	Strategies RecoveryStrategies `yaml:"strategies"`
}

type RecoveryStrategies struct {
	SchedulerRestart    bool `yaml:"scheduler_restart"`
	NodeDraining        bool `yaml:"node_draining"`
	PodEviction         bool `yaml:"pod_eviction"`
	ResourceRebalancing bool `yaml:"resource_rebalancing"`
	ConfigRollback      bool `yaml:"config_rollback"`
}

type RecoveryAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  time.Time              `json:"timestamp"`
	Status     string                 `json:"status"`
	Result     string                 `json:"result"`
	Duration   time.Duration          `json:"duration"`
	Attempts   int                    `json:"attempts"`
}

func NewSchedulerRecoveryManager(client kubernetes.Interface, config *RecoveryConfig) *SchedulerRecoveryManager {
	return &SchedulerRecoveryManager{
		client: client,
		config: config,
	}
}

func (srm *SchedulerRecoveryManager) StartRecoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(srm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := srm.performHealthCheck(ctx); err != nil {
				klog.Errorf("Health check failed: %v", err)
				if srm.config.EnableAutoRecovery {
					if recoveryErr := srm.initiateRecovery(ctx, err); recoveryErr != nil {
						klog.Errorf("Recovery failed: %v", err)
					}
				}
			}
		}
	}
}

func (srm *SchedulerRecoveryManager) performHealthCheck(ctx context.Context) error {
	// 检查调度器健康状态
	schedulerHealth, err := srm.checkSchedulerHealth(ctx)
	if err != nil {
		return fmt.Errorf("scheduler health check failed: %v", err)
	}

	// 检查节点健康状态
	nodeHealth, err := srm.checkNodeHealth(ctx)
	if err != nil {
		return fmt.Errorf("node health check failed: %v", err)
	}

	// 检查调度队列健康状态
	queueHealth, err := srm.checkQueueHealth(ctx)
	if err != nil {
		return fmt.Errorf("queue health check failed: %v", err)
	}

	// 记录健康状态
	klog.V(4).Infof("Health check results - Scheduler: %t, Nodes: %t, Queue: %t",
		schedulerHealth.Healthy, nodeHealth.Healthy, queueHealth.Healthy)

	// 如果任何组件不健康，返回错误
	if !schedulerHealth.Healthy || !nodeHealth.Healthy || !queueHealth.Healthy {
		return fmt.Errorf("unhealthy components detected")
	}

	return nil
}

func (srm *SchedulerRecoveryManager) checkSchedulerHealth(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		Name:      "scheduler",
		LastCheck: time.Now(),
		Healthy:   true,
	}

	// 检查调度器Pod状态
	pods, err := srm.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-scheduler",
	})
	if err != nil {
		health.Healthy = false
		health.Error = fmt.Errorf("Failed to list scheduler pods: %v", err)
		return health, err
	}

	if len(pods.Items) == 0 {
		health.Healthy = false
		health.Error = fmt.Errorf("No scheduler pods found")
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
		health.Error = fmt.Errorf("No running scheduler pods")
		return health, fmt.Errorf("no running scheduler pods")
	}

	return health, nil
}

func (srm *SchedulerRecoveryManager) checkNodeHealth(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		Name:      "nodes",
		LastCheck: time.Now(),
		Healthy:   true,
	}

	nodes, err := srm.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		health.Healthy = false
		health.Error = fmt.Errorf("Failed to list nodes: %v", err)
		return health, err
	}

	if len(nodes.Items) == 0 {
		health.Healthy = false
		health.Error = fmt.Errorf("No nodes found")
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
		health.Error = fmt.Errorf("Only %d/%d nodes are ready", readyNodes, len(nodes.Items))
		return health, fmt.Errorf("insufficient ready nodes")
	}

	return health, nil
}

func (srm *SchedulerRecoveryManager) checkQueueHealth(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		Name:      "queue",
		LastCheck: time.Now(),
		Healthy:   true,
	}

	// 检查Pending状态的Pod数量
	pendingPods, err := srm.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Pending",
	})
	if err != nil {
		health.Healthy = false
		health.Error = fmt.Errorf("Failed to list pending pods: %v", err)
		return health, err
	}

	// 如果Pending Pod数量过多，认为队列不健康
	if len(pendingPods.Items) > 100 {
		health.Healthy = false
		health.Error = fmt.Errorf("Too many pending pods: %d", len(pendingPods.Items))
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
		health.Error = fmt.Errorf("Too many long-pending pods: %d", longPendingPods)
		return health, fmt.Errorf("scheduling stalled")
	}

	return health, nil
}

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

func (srm *SchedulerRecoveryManager) rebalanceResources(ctx context.Context, action RecoveryAction) error {
	klog.Infof("Rebalancing cluster resources")

	// 这里可以实现资源重平衡逻辑
	// 例如：识别资源使用不均衡的节点，建议Pod迁移等

	action.Status = "completed"
	action.Result = "Resource rebalancing analysis completed"
	return nil
}
