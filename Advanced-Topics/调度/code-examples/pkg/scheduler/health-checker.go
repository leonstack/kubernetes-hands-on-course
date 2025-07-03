// health-checker.go
// 调度器健康检查器 - 提供全面的健康监控功能
package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// HealthCheck 健康检查定义
type HealthCheck struct {
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	Retries     int           `json:"retries"`
	ExpectedCode int          `json:"expectedCode"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Name      string    `json:"name"`
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks      []HealthCheck
	statuses    map[string]HealthStatus
	mu          sync.RWMutex
	alertMgr    *AlertManager
	httpClient  *http.Client
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(checks []HealthCheck, alertMgr *AlertManager) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		checks:   checks,
		statuses: make(map[string]HealthStatus),
		alertMgr: alertMgr,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动健康检查
func (hc *HealthChecker) Start() {
	klog.Info("Starting health checker...")

	for _, check := range hc.checks {
		go hc.runHealthCheck(check)
	}
}

// Stop 停止健康检查
func (hc *HealthChecker) Stop() {
	klog.Info("Stopping health checker...")
	hc.cancel()
}

// runHealthCheck 运行单个健康检查
func (hc *HealthChecker) runHealthCheck(check HealthCheck) {
	ticker := time.NewTicker(check.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.performCheck(check)
		}
	}
}

// performCheck 执行健康检查
func (hc *HealthChecker) performCheck(check HealthCheck) {
	start := time.Now()
	var lastErr error

	// 重试机制
	for i := 0; i <= check.Retries; i++ {
		ctx, cancel := context.WithTimeout(hc.ctx, check.Timeout)
		req, err := http.NewRequestWithContext(ctx, "GET", check.URL, nil)
		if err != nil {
			lastErr = err
			cancel()
			continue
		}

		resp, err := hc.httpClient.Do(req)
		cancel()

		if err != nil {
			lastErr = err
			if i < check.Retries {
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
		} else {
			resp.Body.Close()
			if resp.StatusCode == check.ExpectedCode {
				// 健康检查成功
				status := HealthStatus{
					Name:      check.Name,
					Healthy:   true,
					Message:   "OK",
					Timestamp: time.Now(),
					Latency:   time.Since(start),
				}
				hc.updateStatus(status)
				return
			} else {
				lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
		}
	}

	// 健康检查失败
	status := HealthStatus{
		Name:      check.Name,
		Healthy:   false,
		Message:   lastErr.Error(),
		Timestamp: time.Now(),
		Latency:   time.Since(start),
	}
	hc.updateStatus(status)
}

// updateStatus 更新健康状态
func (hc *HealthChecker) updateStatus(status HealthStatus) {
	hc.mu.Lock()
	prevStatus, exists := hc.statuses[status.Name]
	hc.statuses[status.Name] = status
	hc.mu.Unlock()

	// 状态变化时发送告警
	if !exists || prevStatus.Healthy != status.Healthy {
		hc.handleHealthStatusChange(status, prevStatus)
	}

	klog.V(2).Infof("Health check %s: %s (latency: %v)", status.Name, status.Message, status.Latency)
}

// handleHealthStatusChange 处理健康状态变化
func (hc *HealthChecker) handleHealthStatusChange(current, previous HealthStatus) {
	if hc.alertMgr == nil {
		return
	}

	var alertType string
	var message string

	if !current.Healthy {
		alertType = "critical"
		message = fmt.Sprintf("Health check %s failed: %s", current.Name, current.Message)
	} else {
		alertType = "resolved"
		message = fmt.Sprintf("Health check %s recovered", current.Name)
	}

	alert := AlertEvent{
		Type:      alertType,
		Message:   message,
		Timestamp: current.Timestamp,
		Labels: map[string]string{
			"component": "scheduler",
			"check":     current.Name,
		},
	}

	hc.alertMgr.SendAlert(alert)
}

// GetStatus 获取健康状态
func (hc *HealthChecker) GetStatus(name string) (HealthStatus, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	status, exists := hc.statuses[name]
	return status, exists
}

// GetAllStatuses 获取所有健康状态
func (hc *HealthChecker) GetAllStatuses() map[string]HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	statuses := make(map[string]HealthStatus)
	for name, status := range hc.statuses {
		statuses[name] = status
	}
	return statuses
}

// IsHealthy 检查整体健康状态
func (hc *HealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	for _, status := range hc.statuses {
		if !status.Healthy {
			return false
		}
	}
	return true
}

// AlertManager 告警管理器
type AlertManager struct {
	webhookURL string
	httpClient *http.Client
}

// AlertEvent 告警事件
type AlertEvent struct {
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

// NewAlertManager 创建告警管理器
func NewAlertManager(webhookURL string) *AlertManager {
	return &AlertManager{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert 发送告警
func (am *AlertManager) SendAlert(alert AlertEvent) error {
	if am.webhookURL == "" {
		klog.V(2).Infof("Alert: %s - %s", alert.Type, alert.Message)
		return nil
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %v", err)
	}

	resp, err := am.httpClient.Post(am.webhookURL, "application/json", 
		bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send alert: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("alert webhook returned status %d", resp.StatusCode)
	}

	klog.V(2).Infof("Alert sent successfully: %s", alert.Message)
	return nil
}