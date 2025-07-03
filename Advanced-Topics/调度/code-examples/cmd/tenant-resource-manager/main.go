// tenant-resource-manager.go
// Kubernetes 多租户资源管理器实现

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// TenantResourceManager 租户资源管理器
type TenantResourceManager struct {
	clientset *kubernetes.Clientset
	tenants   map[string]*TenantConfig
	mu        sync.RWMutex
	metrics   *TenantMetrics
}

// TenantConfig 租户配置
type TenantConfig struct {
	Name       string                 `json:"name"`
	Namespaces []string               `json:"namespaces"`
	Quotas     map[string]TenantQuota `json:"quotas"`
	Priorities map[string]int32       `json:"priorities"`
	Policies   TenantPolicies         `json:"policies"`
}

// TenantQuota 租户配额
type TenantQuota struct {
	CPU     resource.Quantity `json:"cpu"`
	Memory  resource.Quantity `json:"memory"`
	Storage resource.Quantity `json:"storage"`
	GPU     int64             `json:"gpu"`
	Pods    int64             `json:"pods"`
}

// TenantPolicies 租户策略
type TenantPolicies struct {
	AllowOvercommit  bool    `json:"allow_overcommit"`
	MaxBurstRatio    float64 `json:"max_burst_ratio"`
	PreemptionPolicy string  `json:"preemption_policy"`
	SchedulingPolicy string  `json:"scheduling_policy"`
}

// TenantMetrics 租户指标
type TenantMetrics struct {
	Usage      map[string]TenantUsage      `json:"usage"`
	Violations map[string][]QuotaViolation `json:"violations"`
	mu         sync.RWMutex
}

// TenantUsage 租户使用情况
type TenantUsage struct {
	CPU         resource.Quantity `json:"cpu"`
	Memory      resource.Quantity `json:"memory"`
	Storage     resource.Quantity `json:"storage"`
	GPU         int64             `json:"gpu"`
	Pods        int64             `json:"pods"`
	LastUpdated time.Time         `json:"last_updated"`
}

// QuotaViolation 配额违规
type QuotaViolation struct {
	Tenant    string            `json:"tenant"`
	Namespace string            `json:"namespace"`
	Resource  string            `json:"resource"`
	Requested resource.Quantity `json:"requested"`
	Available resource.Quantity `json:"available"`
	Timestamp time.Time         `json:"timestamp"`
}

// ResourceAllocation 资源分配结果
type ResourceAllocation struct {
	Approved    bool                `json:"approved"`
	Resources   corev1.ResourceList `json:"resources"`
	Priority    int32               `json:"priority"`
	Constraints map[string]string   `json:"constraints"`
}

// NewTenantResourceManager 创建租户资源管理器
func NewTenantResourceManager(clientset kubernetes.Interface) *TenantResourceManager {
	// 类型断言，确保是正确的客户端类型
	var kubeClient *kubernetes.Clientset
	if cs, ok := clientset.(*kubernetes.Clientset); ok {
		kubeClient = cs
	} else {
		// 如果类型断言失败，记录警告但继续运行
		klog.Warning("Clientset type assertion failed, some features may not work correctly")
		kubeClient = nil
	}

	trm := &TenantResourceManager{
		clientset: kubeClient,
		tenants:   make(map[string]*TenantConfig),
		metrics: &TenantMetrics{
			Usage:      make(map[string]TenantUsage),
			Violations: make(map[string][]QuotaViolation),
		},
	}

	// 启动后台任务定期更新租户使用情况
	go trm.startMetricsUpdater()

	return trm
}

// startMetricsUpdater 启动指标更新器
func (trm *TenantResourceManager) startMetricsUpdater() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 使用 for range 替代 for { select {} } 模式
	for range ticker.C {
		// 定期更新所有租户的指标
		trm.updateAllTenantMetrics()
	}
}

// updateAllTenantMetrics 更新所有租户的指标
func (trm *TenantResourceManager) updateAllTenantMetrics() {
	trm.mu.RLock()
	tenantNames := make([]string, 0, len(trm.tenants))
	for name := range trm.tenants {
		tenantNames = append(tenantNames, name)
	}
	trm.mu.RUnlock()

	for _, tenantName := range tenantNames {
		if _, err := trm.getTenantUsage(context.Background(), tenantName); err != nil {
			klog.Errorf("Failed to update metrics for tenant %s: %v", tenantName, err)
		}
	}
}

// RegisterTenant 注册租户
func (trm *TenantResourceManager) RegisterTenant(config *TenantConfig) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	// 验证租户配置
	if err := trm.validateTenantConfig(config); err != nil {
		return fmt.Errorf("invalid tenant config: %v", err)
	}

	trm.tenants[config.Name] = config

	// 为租户创建资源配额
	if err := trm.createTenantQuotas(config); err != nil {
		return fmt.Errorf("failed to create tenant quotas: %v", err)
	}

	klog.Infof("Registered tenant %s with %d namespaces", config.Name, len(config.Namespaces))
	return nil
}

// HTTP API 处理函数

// handleGetTenants 获取租户列表
func handleGetTenants(w http.ResponseWriter, _ *http.Request, trm *TenantResourceManager) {
	trm.mu.RLock()
	tenants := make([]*TenantConfig, 0, len(trm.tenants))
	for _, tenant := range trm.tenants {
		tenants = append(tenants, tenant)
	}
	trm.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"tenants": tenants,
		"count":   len(tenants),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleCreateTenant 创建租户
func handleCreateTenant(w http.ResponseWriter, r *http.Request, trm *TenantResourceManager) {
	var config TenantConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := trm.RegisterTenant(&config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to register tenant: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Tenant %s created successfully", config.Name),
		"tenant":  config.Name,
	})
}

// ResourceRequest 资源请求结构
type ResourceRequest struct {
	Namespace string              `json:"namespace"`
	Resources corev1.ResourceList `json:"resources"`
}

// handleCheckResourceRequest 检查资源请求
func handleCheckResourceRequest(w http.ResponseWriter, r *http.Request, trm *TenantResourceManager) {
	// 从 URL 路径中提取租户名称
	tenantName := r.URL.Path[len("/api/v1/tenants/"):]
	if tenantName == "" {
		http.Error(w, "Tenant name is required", http.StatusBadRequest)
		return
	}

	var req ResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	allocation, err := trm.CheckResourceRequest(r.Context(), tenantName, req.Namespace, req.Resources)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"approved": false,
			"reason":   err.Error(),
			"tenant":   tenantName,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocation)
}

// handleGetTenantMetrics 获取租户指标
func handleGetTenantMetrics(w http.ResponseWriter, r *http.Request, trm *TenantResourceManager) {
	// 从 URL 路径中提取租户名称
	tenantName := r.URL.Path[len("/api/v1/metrics/tenants/"):]
	if tenantName == "" {
		http.Error(w, "Tenant name is required", http.StatusBadRequest)
		return
	}

	metrics, err := trm.GetTenantMetrics(tenantName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tenant metrics: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// CheckResourceRequest 检查资源请求
func (trm *TenantResourceManager) CheckResourceRequest(ctx context.Context, tenant, namespace string, request corev1.ResourceList) (*ResourceAllocation, error) {
	trm.mu.RLock()
	tenantConfig, exists := trm.tenants[tenant]
	trm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenant)
	}

	// 检查命名空间是否属于租户
	if !trm.isNamespaceInTenant(namespace, tenantConfig) {
		return nil, fmt.Errorf("namespace %s does not belong to tenant %s", namespace, tenant)
	}

	// 获取当前使用情况
	usage, err := trm.getTenantUsage(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant usage: %v", err)
	}

	// 检查配额限制
	allocation, err := trm.checkQuotaLimits(tenantConfig, usage, request)
	if err != nil {
		// 记录配额违规
		trm.recordQuotaViolation(tenant, namespace, request, err)
		return nil, err
	}

	return allocation, nil
}

// validateTenantConfig 验证租户配置
func (trm *TenantResourceManager) validateTenantConfig(config *TenantConfig) error {
	if config.Name == "" {
		return fmt.Errorf("tenant name cannot be empty")
	}

	if len(config.Namespaces) == 0 {
		return fmt.Errorf("tenant must have at least one namespace")
	}

	// 验证配额配置
	for env, quota := range config.Quotas {
		if quota.CPU.IsZero() || quota.Memory.IsZero() {
			return fmt.Errorf("invalid quota for environment %s: CPU and Memory must be specified", env)
		}
	}

	return nil
}

// createTenantQuotas 为租户创建资源配额
func (trm *TenantResourceManager) createTenantQuotas(config *TenantConfig) error {
	// 如果没有 Kubernetes 客户端，跳过实际的配额创建
	if trm.clientset == nil {
		klog.V(4).Infof("No Kubernetes client available, skipping quota creation for tenant %s", config.Name)
		return nil
	}

	for _, namespace := range config.Namespaces {
		// 确定环境类型
		env := trm.getEnvironmentType(namespace)
		quota, exists := config.Quotas[env]
		if !exists {
			quota = config.Quotas["default"]
		}

		// 创建 ResourceQuota
		resourceQuota := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-quota", config.Name),
				Namespace: namespace,
				Labels: map[string]string{
					"tenant":      config.Name,
					"environment": env,
					"managed-by":  "tenant-resource-manager",
				},
				Annotations: map[string]string{
					"tenant-resource-manager/created-at": time.Now().Format(time.RFC3339),
					"tenant-resource-manager/version":    "1.0.0",
				},
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{
					corev1.ResourceRequestsCPU:     quota.CPU,
					corev1.ResourceRequestsMemory:  quota.Memory,
					corev1.ResourceRequestsStorage: quota.Storage,
					corev1.ResourcePods:            *resource.NewQuantity(quota.Pods, resource.DecimalSI),
				},
			},
		}

		// 添加 GPU 配额（如果有）
		if quota.GPU > 0 {
			resourceQuota.Spec.Hard["requests.nvidia.com/gpu"] = *resource.NewQuantity(quota.GPU, resource.DecimalSI)
		}

		// 尝试创建资源配额
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := trm.clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
		if err != nil {
			// 如果配额已存在，尝试更新
			if errors.IsAlreadyExists(err) {
				klog.Infof("ResourceQuota already exists for namespace %s, attempting to update", namespace)
				// 获取现有配额
				existingQuota, getErr := trm.clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuota.Name, metav1.GetOptions{})
				if getErr != nil {
					return fmt.Errorf("failed to get existing resource quota for namespace %s: %v", namespace, getErr)
				}
				// 更新配额
				existingQuota.Spec = resourceQuota.Spec
				existingQuota.Labels = resourceQuota.Labels
				existingQuota.Annotations = resourceQuota.Annotations
				_, updateErr := trm.clientset.CoreV1().ResourceQuotas(namespace).Update(ctx, existingQuota, metav1.UpdateOptions{})
				if updateErr != nil {
					return fmt.Errorf("failed to update resource quota for namespace %s: %v", namespace, updateErr)
				}
				klog.Infof("Updated ResourceQuota for namespace %s", namespace)
			} else {
				return fmt.Errorf("failed to create resource quota for namespace %s: %v", namespace, err)
			}
		} else {
			klog.Infof("Created ResourceQuota for namespace %s", namespace)
		}
	}

	return nil
}

// getTenantUsage 获取租户使用情况
func (trm *TenantResourceManager) getTenantUsage(ctx context.Context, tenant string) (TenantUsage, error) {
	trm.mu.RLock()
	tenantConfig, exists := trm.tenants[tenant]
	trm.mu.RUnlock()

	if !exists {
		return TenantUsage{}, fmt.Errorf("tenant %s not found", tenant)
	}

	var totalUsage TenantUsage

	// 如果没有 Kubernetes 客户端，返回模拟数据
	if trm.clientset == nil {
		klog.V(4).Infof("No Kubernetes client available, returning mock usage data for tenant %s", tenant)
		// 返回模拟的使用数据用于测试
		totalUsage = TenantUsage{
			CPU:         *resource.NewMilliQuantity(500, resource.DecimalSI),        // 0.5 CPU
			Memory:      *resource.NewQuantity(1024*1024*1024, resource.BinarySI),   // 1Gi
			Storage:     *resource.NewQuantity(5*1024*1024*1024, resource.BinarySI), // 5Gi
			GPU:         0,
			Pods:        int64(len(tenantConfig.Namespaces) * 2), // 每个命名空间2个Pod
			LastUpdated: time.Now(),
		}
	} else {
		// 实际从 Kubernetes API 获取使用情况
		for _, namespace := range tenantConfig.Namespaces {
			// 获取命名空间中的所有 Pod
			pods, err := trm.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				FieldSelector: "status.phase=Running",
			})
			if err != nil {
				klog.Errorf("Failed to list pods in namespace %s: %v", namespace, err)
				continue // 继续处理其他命名空间
			}

			// 计算资源使用量
			for _, pod := range pods.Items {
				for _, container := range pod.Spec.Containers {
					if cpu := container.Resources.Requests[corev1.ResourceCPU]; !cpu.IsZero() {
						totalUsage.CPU.Add(cpu)
					}
					if memory := container.Resources.Requests[corev1.ResourceMemory]; !memory.IsZero() {
						totalUsage.Memory.Add(memory)
					}
					// 检查 GPU 资源
					if gpu := container.Resources.Requests["nvidia.com/gpu"]; !gpu.IsZero() {
						totalUsage.GPU += gpu.Value()
					}
				}
			}

			totalUsage.Pods += int64(len(pods.Items))
		}

		totalUsage.LastUpdated = time.Now()
	}

	// 更新缓存
	trm.metrics.mu.Lock()
	trm.metrics.Usage[tenant] = totalUsage
	trm.metrics.mu.Unlock()

	return totalUsage, nil
}

// checkQuotaLimits 检查配额限制
func (trm *TenantResourceManager) checkQuotaLimits(config *TenantConfig, usage TenantUsage, request corev1.ResourceList) (*ResourceAllocation, error) {
	// 获取适用的配额
	quota := config.Quotas["default"]

	// 计算有效配额（考虑突发使用策略）
	effectiveCPUQuota := quota.CPU.DeepCopy()
	effectiveMemoryQuota := quota.Memory.DeepCopy()

	if config.Policies.AllowOvercommit && config.Policies.MaxBurstRatio > 1.0 {
		// 允许突发使用，增加有效配额
		burstCPU := quota.CPU.DeepCopy()
		burstCPU.Set(int64(float64(burstCPU.MilliValue()) * config.Policies.MaxBurstRatio))
		effectiveCPUQuota = burstCPU

		burstMemory := quota.Memory.DeepCopy()
		burstMemory.Set(int64(float64(burstMemory.Value()) * config.Policies.MaxBurstRatio))
		effectiveMemoryQuota = burstMemory
	}

	// 预留10%资源用于系统开销
	reservedCPU := effectiveCPUQuota.DeepCopy()
	reservedCPU.Set(int64(float64(reservedCPU.MilliValue()) * 0.9))

	reservedMemory := effectiveMemoryQuota.DeepCopy()
	reservedMemory.Set(int64(float64(reservedMemory.Value()) * 0.9))

	// 检查 CPU 配额
	if cpu := request[corev1.ResourceCPU]; !cpu.IsZero() {
		newCPUUsage := usage.CPU.DeepCopy()
		newCPUUsage.Add(cpu)
		if newCPUUsage.Cmp(reservedCPU) > 0 {
			available := reservedCPU.DeepCopy()
			available.Sub(usage.CPU)
			return nil, fmt.Errorf("CPU quota exceeded: requested %v, available %v (reserved quota)", cpu, available)
		}
	}

	// 检查内存配额
	if memory := request[corev1.ResourceMemory]; !memory.IsZero() {
		newMemoryUsage := usage.Memory.DeepCopy()
		newMemoryUsage.Add(memory)
		if newMemoryUsage.Cmp(reservedMemory) > 0 {
			available := reservedMemory.DeepCopy()
			available.Sub(usage.Memory)
			return nil, fmt.Errorf("Memory quota exceeded: requested %v, available %v (reserved quota)", memory, available)
		}
	}

	// 检查 Pod 数量配额
	if usage.Pods >= quota.Pods {
		return nil, fmt.Errorf("Pod quota exceeded: current %d, limit %d", usage.Pods, quota.Pods)
	}

	// 创建资源分配
	allocation := &ResourceAllocation{
		Approved:  true,
		Resources: request,
		Priority:  config.Priorities["default"],
		Constraints: map[string]string{
			"tenant":        config.Name,
			"burst-allowed": fmt.Sprintf("%t", config.Policies.AllowOvercommit),
		},
	}

	return allocation, nil
}

// recordQuotaViolation 记录配额违规
func (trm *TenantResourceManager) recordQuotaViolation(tenant, namespace string, request corev1.ResourceList, err error) {
	// 解析错误信息以提取资源类型和数量
	resourceType := "unknown"
	var requested, available resource.Quantity

	// 尝试从错误信息中解析资源类型
	errorStr := err.Error()
	if contains(errorStr, "CPU") {
		resourceType = "cpu"
		if cpu := request[corev1.ResourceCPU]; !cpu.IsZero() {
			requested = cpu
		}
	} else if contains(errorStr, "Memory") {
		resourceType = "memory"
		if memory := request[corev1.ResourceMemory]; !memory.IsZero() {
			requested = memory
		}
	} else if contains(errorStr, "Pod") {
		resourceType = "pods"
		requested = *resource.NewQuantity(1, resource.DecimalSI)
	}

	violation := QuotaViolation{
		Tenant:    tenant,
		Namespace: namespace,
		Resource:  resourceType,
		Requested: requested,
		Available: available, // 这里可以进一步改进以获取实际可用量
		Timestamp: time.Now(),
	}

	trm.metrics.mu.Lock()
	if trm.metrics.Violations[tenant] == nil {
		trm.metrics.Violations[tenant] = make([]QuotaViolation, 0)
	}
	// 限制违规记录数量，避免内存泄漏
	if len(trm.metrics.Violations[tenant]) >= 100 {
		// 保留最近的99条记录
		trm.metrics.Violations[tenant] = trm.metrics.Violations[tenant][1:]
	}
	trm.metrics.Violations[tenant] = append(trm.metrics.Violations[tenant], violation)
	trm.metrics.mu.Unlock()

	klog.Warningf("Quota violation for tenant %s in namespace %s: resource=%s, requested=%v, error=%v",
		tenant, namespace, resourceType, requested, err)
}

// 辅助函数
func (trm *TenantResourceManager) isNamespaceInTenant(namespace string, config *TenantConfig) bool {
	for _, ns := range config.Namespaces {
		if ns == namespace {
			return true
		}
	}
	return false
}

func (trm *TenantResourceManager) getEnvironmentType(namespace string) string {
	if contains(namespace, "prod") {
		return "production"
	} else if contains(namespace, "dev") {
		return "development"
	} else if contains(namespace, "test") {
		return "testing"
	}
	return "default"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// GetTenantMetrics 获取租户指标
func (trm *TenantResourceManager) GetTenantMetrics(tenant string) (*TenantMetrics, error) {
	trm.metrics.mu.RLock()
	defer trm.metrics.mu.RUnlock()

	usage, exists := trm.metrics.Usage[tenant]
	if !exists {
		return nil, fmt.Errorf("no metrics found for tenant %s", tenant)
	}

	violations := trm.metrics.Violations[tenant]

	return &TenantMetrics{
		Usage: map[string]TenantUsage{
			tenant: usage,
		},
		Violations: map[string][]QuotaViolation{
			tenant: violations,
		},
	}, nil
}

func main() {
	// 解析命令行参数
	var (
		kubeconfig  = flag.String("kubeconfig", "", "Path to kubeconfig file")
		httpPort    = flag.String("http-port", getEnvOrDefault("HTTP_PORT", "8080"), "HTTP server port")
		metricsPort = flag.String("metrics-port", getEnvOrDefault("METRICS_PORT", "8081"), "Metrics server port")
		logLevel    = flag.String("log-level", getEnvOrDefault("LOG_LEVEL", "info"), "Log level")
	)
	flag.Parse()

	// 初始化日志
	klog.InitFlags(nil)
	if *logLevel != "" {
		// 设置日志级别
		log.Printf("Log level set to: %s", *logLevel)
	}
	klog.Infof("Starting Kubernetes Tenant Resource Manager on port %s...", *httpPort)

	// 创建 Kubernetes 客户端
	clientset, err := createKubernetesClient(*kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// 创建租户资源管理器
	trm := NewTenantResourceManager(clientset)

	// 注册示例租户
	if err := trm.registerExampleTenants(); err != nil {
		klog.Fatalf("Failed to register example tenants: %v", err)
	}

	// 启动 HTTP 服务器
	server := &http.Server{
		Addr:    ":" + *httpPort,
		Handler: setupHTTPRoutes(trm),
	}

	// 启动指标服务器
	metricsServer := &http.Server{
		Addr:    ":" + *metricsPort,
		Handler: setupMetricsRoutes(trm),
	}

	// 启动服务器
	go func() {
		klog.Infof("Starting HTTP server on port %s", *httpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		klog.Infof("Starting metrics server on port %s", *metricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Fatalf("Metrics server failed: %v", err)
		}
	}()

	klog.Info("Tenant Resource Manager started successfully")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	klog.Info("Shutting down servers...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		klog.Errorf("HTTP server shutdown error: %v", err)
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		klog.Errorf("Metrics server shutdown error: %v", err)
	}

	klog.Info("Servers stopped")
}

// createKubernetesClient 创建 Kubernetes 客户端
func createKubernetesClient(kubeconfig string) (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		// 使用指定的 kubeconfig 文件
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// 尝试使用集群内配置
		config, err = rest.InClusterConfig()
		if err != nil {
			// 如果集群内配置失败，尝试使用默认的 kubeconfig
			homeDir, _ := os.UserHomeDir()
			defaultKubeconfig := homeDir + "/.kube/config"
			if _, statErr := os.Stat(defaultKubeconfig); statErr == nil {
				config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return clientset, nil
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupHTTPRoutes 设置 HTTP 路由
func setupHTTPRoutes(trm *TenantResourceManager) http.Handler {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// 租户管理 API
	mux.HandleFunc("/api/v1/tenants", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTenants(w, r, trm)
		case http.MethodPost:
			handleCreateTenant(w, r, trm)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 租户资源检查 API
	mux.HandleFunc("/api/v1/tenants/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path[len("/api/v1/tenants/"):] != "" {
			handleCheckResourceRequest(w, r, trm)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	// 租户指标 API
	mux.HandleFunc("/api/v1/metrics/tenants/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetTenantMetrics(w, r, trm)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

// setupMetricsRoutes 设置指标路由
func setupMetricsRoutes(trm *TenantResourceManager) http.Handler {
	mux := http.NewServeMux()

	// Prometheus 指标端点
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		
		// 基本信息指标
		fmt.Fprintf(w, "# HELP tenant_resource_manager_info Information about tenant resource manager\n")
		fmt.Fprintf(w, "# TYPE tenant_resource_manager_info gauge\n")
		fmt.Fprintf(w, "tenant_resource_manager_info{version=\"1.0.0\"} 1\n")
		
		// 租户数量指标
		trm.mu.RLock()
		tenantCount := len(trm.tenants)
		trm.mu.RUnlock()
		
		fmt.Fprintf(w, "# HELP tenant_resource_manager_tenants_total Total number of registered tenants\n")
		fmt.Fprintf(w, "# TYPE tenant_resource_manager_tenants_total gauge\n")
		fmt.Fprintf(w, "tenant_resource_manager_tenants_total %d\n", tenantCount)
	})

	return mux
}

// registerExampleTenants 注册示例租户
func (trm *TenantResourceManager) registerExampleTenants() error {
	// 示例租户配置
	tenants := []*TenantConfig{
		{
			Name:       "team-alpha",
			Namespaces: []string{"alpha-prod", "alpha-dev"},
			Quotas: map[string]TenantQuota{
				"production": {
					CPU:     resource.MustParse("10"),
					Memory:  resource.MustParse("20Gi"),
					Storage: resource.MustParse("100Gi"),
					GPU:     2,
					Pods:    50,
				},
				"development": {
					CPU:     resource.MustParse("5"),
					Memory:  resource.MustParse("10Gi"),
					Storage: resource.MustParse("50Gi"),
					GPU:     1,
					Pods:    25,
				},
				"default": {
					CPU:     resource.MustParse("2"),
					Memory:  resource.MustParse("4Gi"),
					Storage: resource.MustParse("20Gi"),
					GPU:     0,
					Pods:    10,
				},
			},
			Priorities: map[string]int32{
				"default": 100,
			},
			Policies: TenantPolicies{
				AllowOvercommit:  true,
				MaxBurstRatio:    1.5,
				PreemptionPolicy: "LowerPriority",
				SchedulingPolicy: "BestEffort",
			},
		},
		{
			Name:       "team-beta",
			Namespaces: []string{"beta-prod", "beta-test"},
			Quotas: map[string]TenantQuota{
				"production": {
					CPU:     resource.MustParse("8"),
					Memory:  resource.MustParse("16Gi"),
					Storage: resource.MustParse("80Gi"),
					GPU:     1,
					Pods:    40,
				},
				"testing": {
					CPU:     resource.MustParse("4"),
					Memory:  resource.MustParse("8Gi"),
					Storage: resource.MustParse("40Gi"),
					GPU:     0,
					Pods:    20,
				},
				"default": {
					CPU:     resource.MustParse("2"),
					Memory:  resource.MustParse("4Gi"),
					Storage: resource.MustParse("20Gi"),
					GPU:     0,
					Pods:    10,
				},
			},
			Priorities: map[string]int32{
				"default": 80,
			},
			Policies: TenantPolicies{
				AllowOvercommit:  false,
				MaxBurstRatio:    1.0,
				PreemptionPolicy: "Never",
				SchedulingPolicy: "Guaranteed",
			},
		},
	}

	// 注册租户
	for _, tenant := range tenants {
		trm.mu.Lock()
		trm.tenants[tenant.Name] = tenant
		trm.mu.Unlock()
		klog.Infof("Registered tenant: %s with %d namespaces", tenant.Name, len(tenant.Namespaces))
	}

	return nil
}
