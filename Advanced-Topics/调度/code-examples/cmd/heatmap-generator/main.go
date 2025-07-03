// heatmap-generator.go
// Kubernetes 集群资源热力图生成器实现

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NodeResourceUsage 节点资源使用情况
type NodeResourceUsage struct {
	NodeName          string            `json:"node_name"`
	CPUCapacity       int64             `json:"cpu_capacity_millicores"`
	MemoryCapacity    int64             `json:"memory_capacity_bytes"`
	CPUUsage          int64             `json:"cpu_usage_millicores"`
	MemoryUsage       int64             `json:"memory_usage_bytes"`
	CPUUtilization    float64           `json:"cpu_utilization_percent"`
	MemoryUtilization float64           `json:"memory_utilization_percent"`
	PodCount          int               `json:"pod_count"`
	SchedulablePods   int               `json:"schedulable_pods"`
	NodeStatus        string            `json:"node_status"`
	Labels            map[string]string `json:"labels"`
	Taints            []corev1.Taint    `json:"taints"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// HeatmapData 热力图数据
type HeatmapData struct {
	Nodes          []NodeResourceUsage `json:"nodes"`
	ClusterMetrics ClusterMetrics      `json:"cluster_metrics"`
	GeneratedAt    time.Time           `json:"generated_at"`
	UpdateInterval time.Duration       `json:"update_interval"`
}

// ClusterMetrics 集群指标
type ClusterMetrics struct {
	ClusterName              string  `json:"cluster_name"`
	TotalNodes               int     `json:"total_nodes"`
	HealthyNodes             int     `json:"healthy_nodes"`
	TotalCPUCapacity         int64   `json:"total_cpu_capacity_millicores"`
	TotalMemoryCapacity      int64   `json:"total_memory_capacity_bytes"`
	TotalCPUUsage            int64   `json:"total_cpu_usage_millicores"`
	TotalMemoryUsage         int64   `json:"total_memory_usage_bytes"`
	AverageCPUUtilization    float64 `json:"average_cpu_utilization_percent"`
	AverageMemoryUtilization float64 `json:"average_memory_utilization_percent"`
	HighLoadNodes            int     `json:"high_load_nodes"`
	TotalPods                int     `json:"total_pods"`
	SchedulablePods          int     `json:"schedulable_pods"`
}

// HeatmapGenerator 热力图生成器
type HeatmapGenerator struct {
	clientset      kubernetes.Interface
	metricsClient  metricsclientset.Interface
	clusterName    string
	updateInterval time.Duration
	lastUpdate     time.Time
	cachedData     *HeatmapData
	// 性能统计
	generationCount int64
	lastGenerationDuration time.Duration
	errorCount     int64
}

// NewHeatmapGenerator 创建热力图生成器
func NewHeatmapGenerator(clientset kubernetes.Interface, metricsClient metricsclientset.Interface, clusterName string) *HeatmapGenerator {
	return &HeatmapGenerator{
		clientset:      clientset,
		metricsClient:  metricsClient,
		clusterName:    clusterName,
		updateInterval: 30 * time.Second,
	}
}

// listNodesWithRetry 带重试机制的节点列表获取
func (hg *HeatmapGenerator) listNodesWithRetry(ctx context.Context, maxRetries int) (*corev1.NodeList, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		nodes, err := hg.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err == nil {
			return nodes, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * time.Second
			klog.V(4).Infof("Failed to list nodes (attempt %d/%d), retrying in %v: %v", i+1, maxRetries, waitTime, err)
			time.Sleep(waitTime)
		}
	}
	return nil, lastErr
}

// GenerateHeatmapData 生成热力图数据
func (hg *HeatmapGenerator) GenerateHeatmapData(ctx context.Context) (*HeatmapData, error) {
	// 检查缓存是否有效
	if hg.cachedData != nil && time.Since(hg.lastUpdate) < hg.updateInterval {
		klog.V(4).Info("Using cached heatmap data")
		return hg.cachedData, nil
	}

	startTime := time.Now()
	klog.Info("Generating cluster resource heatmap data...")
	defer func() {
		hg.lastGenerationDuration = time.Since(startTime)
		hg.generationCount++
		klog.Infof("Heatmap generation completed in %v (total generations: %d)", hg.lastGenerationDuration, hg.generationCount)
	}()

	// 获取所有节点（带重试机制）
	nodes, err := hg.listNodesWithRetry(ctx, 3)
	if err != nil {
		hg.errorCount++
		return nil, fmt.Errorf("failed to list nodes after retries: %v", err)
	}

	// 获取节点指标
	nodeUsages := make([]NodeResourceUsage, 0, len(nodes.Items))
	var clusterMetrics ClusterMetrics
	failedNodes := 0

	for _, node := range nodes.Items {
		usage, err := hg.getNodeResourceUsage(ctx, &node)
		if err != nil {
			klog.Warningf("Failed to get resource usage for node %s: %v", node.Name, err)
			failedNodes++
			// 如果失败节点过多，返回错误
			if failedNodes > len(nodes.Items)/2 {
				hg.errorCount++
				return nil, fmt.Errorf("too many nodes failed (%d/%d), cluster data may be unreliable", failedNodes, len(nodes.Items))
			}
			continue
		}
		nodeUsages = append(nodeUsages, usage)
	}

	if len(nodeUsages) == 0 {
		hg.errorCount++
		return nil, fmt.Errorf("no valid node data collected")
	}

	// 计算集群指标
	clusterMetrics = hg.calculateClusterMetrics(nodeUsages)

	heatmapData := &HeatmapData{
		Nodes:          nodeUsages,
		ClusterMetrics: clusterMetrics,
		GeneratedAt:    time.Now(),
		UpdateInterval: hg.updateInterval,
	}

	// 更新缓存
	hg.cachedData = heatmapData
	hg.lastUpdate = time.Now()

	klog.Infof("Generated heatmap data for %d nodes", len(nodeUsages))
	return heatmapData, nil
}

// getNodeResourceUsage 获取节点资源使用情况
func (hg *HeatmapGenerator) getNodeResourceUsage(ctx context.Context, node *corev1.Node) (NodeResourceUsage, error) {
	usage := NodeResourceUsage{
		NodeName:    node.Name,
		Labels:      node.Labels,
		Taints:      node.Spec.Taints,
		LastUpdated: time.Now(),
	}

	// 获取节点容量
	if cpu := node.Status.Capacity[corev1.ResourceCPU]; !cpu.IsZero() {
		usage.CPUCapacity = cpu.MilliValue()
	}
	if memory := node.Status.Capacity[corev1.ResourceMemory]; !memory.IsZero() {
		usage.MemoryCapacity = memory.Value()
	}

	// 获取节点状态
	usage.NodeStatus = "Unknown"
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				usage.NodeStatus = "Ready"
			} else {
				usage.NodeStatus = "NotReady"
			}
			break
		}
	}

	// 获取节点上的 Pod 数量
	pods, err := hg.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
	})
	if err != nil {
		return usage, fmt.Errorf("failed to list pods on node %s: %v", node.Name, err)
	}

	usage.PodCount = len(pods.Items)

	// 计算资源使用量 - 优先使用真实指标，否则基于Pod请求量估算
	var cpuUsage, memoryUsage int64
	var useRealMetrics bool

	// 尝试获取真实指标
	if hg.metricsClient != nil {
		if nodeMetrics, err := hg.metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{}); err == nil {
			if cpu := nodeMetrics.Usage[corev1.ResourceCPU]; !cpu.IsZero() {
				cpuUsage = cpu.MilliValue()
				useRealMetrics = true
			}
			if memory := nodeMetrics.Usage[corev1.ResourceMemory]; !memory.IsZero() {
				memoryUsage = memory.Value()
			}
		} else {
			klog.V(4).Infof("Failed to get real metrics for node %s: %v", node.Name, err)
		}
	}

	// 如果没有真实指标，基于Pod请求量估算
	if !useRealMetrics {
		var requestedCPU, requestedMemory int64
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
				for _, container := range pod.Spec.Containers {
					if cpu := container.Resources.Requests[corev1.ResourceCPU]; !cpu.IsZero() {
						requestedCPU += cpu.MilliValue()
					}
					if memory := container.Resources.Requests[corev1.ResourceMemory]; !memory.IsZero() {
						requestedMemory += memory.Value()
					}
				}
			}
		}

		// 基于请求量估算实际使用量（通常为请求量的75%）+ 系统开销
		cpuUsage = int64(float64(requestedCPU)*0.75) + int64(200+rand.Intn(100)) // 200-300m 系统开销
		memoryUsage = int64(float64(requestedMemory)*0.75) + int64(512*1024*1024+rand.Intn(512*1024*1024)) // 512MB-1GB 系统开销
		
		klog.V(4).Infof("Using estimated metrics for node %s (CPU: %dm, Memory: %dMB)", 
			node.Name, cpuUsage, memoryUsage/(1024*1024))
	}

	usage.CPUUsage = cpuUsage
	usage.MemoryUsage = memoryUsage

	// 计算利用率
	if usage.CPUCapacity > 0 {
		usage.CPUUtilization = float64(usage.CPUUsage) / float64(usage.CPUCapacity) * 100
	}
	if usage.MemoryCapacity > 0 {
		usage.MemoryUtilization = float64(usage.MemoryUsage) / float64(usage.MemoryCapacity) * 100
	}

	// 计算可调度的 Pod 数量
	usage.SchedulablePods = hg.calculateSchedulablePods(usage)

	return usage, nil
}

// calculateSchedulablePods 计算可调度的 Pod 数量
func (hg *HeatmapGenerator) calculateSchedulablePods(usage NodeResourceUsage) int {
	if usage.NodeStatus != "Ready" {
		return 0
	}

	// 检查节点是否有NoSchedule污点
	for _, taint := range usage.Taints {
		if taint.Effect == corev1.TaintEffectNoSchedule {
			return 0
		}
	}

	// 动态计算平均Pod资源需求（基于当前节点上的Pod）
	avgPodCPU := int64(100)                  // 默认100m
	avgPodMemory := int64(128 * 1024 * 1024) // 默认128MB
	
	if usage.PodCount > 0 {
		// 基于当前Pod使用量估算平均需求
		avgPodCPU = usage.CPUUsage / int64(usage.PodCount)
		avgPodMemory = usage.MemoryUsage / int64(usage.PodCount)
		
		// 设置合理的边界值
		if avgPodCPU < 50 {
			avgPodCPU = 50 // 最小50m
		} else if avgPodCPU > 1000 {
			avgPodCPU = 1000 // 最大1000m
		}
		
		if avgPodMemory < 64*1024*1024 {
			avgPodMemory = 64 * 1024 * 1024 // 最小64MB
		} else if avgPodMemory > 2*1024*1024*1024 {
			avgPodMemory = 2 * 1024 * 1024 * 1024 // 最大2GB
		}
	}

	// 计算剩余资源
	remainingCPU := usage.CPUCapacity - usage.CPUUsage
	remainingMemory := usage.MemoryCapacity - usage.MemoryUsage

	// 保留15%的资源作为缓冲（生产环境建议）
	bufferRatio := 0.85
	remainingCPU = int64(float64(remainingCPU) * bufferRatio)
	remainingMemory = int64(float64(remainingMemory) * bufferRatio)

	if remainingCPU <= 0 || remainingMemory <= 0 {
		return 0
	}

	// 计算基于CPU和内存的可调度Pod数量
	cpuBasedPods := remainingCPU / avgPodCPU
	memoryBasedPods := remainingMemory / avgPodMemory

	// 取较小值（资源瓶颈）
	schedulablePods := cpuBasedPods
	if memoryBasedPods < cpuBasedPods {
		schedulablePods = memoryBasedPods
	}

	// 获取节点最大Pod数限制（从节点标签或默认110）
	maxPodsPerNode := int64(110) // 默认值
	if maxPodsLabel, exists := usage.Labels["node.kubernetes.io/max-pods"]; exists {
		if parsedMax, err := strconv.ParseInt(maxPodsLabel, 10, 64); err == nil && parsedMax > 0 {
			maxPodsPerNode = parsedMax
		}
	}

	// 检查Pod数量限制
	if int64(usage.PodCount) >= maxPodsPerNode {
		return 0
	}

	remainingPodSlots := maxPodsPerNode - int64(usage.PodCount)
	if schedulablePods > remainingPodSlots {
		schedulablePods = remainingPodSlots
	}

	return int(math.Max(0, float64(schedulablePods)))
}

// calculateClusterMetrics 计算集群指标
func (hg *HeatmapGenerator) calculateClusterMetrics(nodeUsages []NodeResourceUsage) ClusterMetrics {
	metrics := ClusterMetrics{
		ClusterName: hg.clusterName,
		TotalNodes:  len(nodeUsages),
	}

	var totalCPUUsage, totalMemoryUsage int64
	var cpuUtilSum, memoryUtilSum float64
	healthyNodes := 0
	highLoadNodes := 0
	totalPods := 0
	totalSchedulablePods := 0

	for _, usage := range nodeUsages {
		// 累计容量
		metrics.TotalCPUCapacity += usage.CPUCapacity
		metrics.TotalMemoryCapacity += usage.MemoryCapacity

		// 累计使用量
		totalCPUUsage += usage.CPUUsage
		totalMemoryUsage += usage.MemoryUsage

		// 累计利用率
		cpuUtilSum += usage.CPUUtilization
		memoryUtilSum += usage.MemoryUtilization

		// 统计健康节点
		if usage.NodeStatus == "Ready" {
			healthyNodes++
		}

		// 统计高负载节点（CPU 或内存利用率超过 80%）
		if usage.CPUUtilization > 80 || usage.MemoryUtilization > 80 {
			highLoadNodes++
		}

		// 累计 Pod 数量
		totalPods += usage.PodCount
		totalSchedulablePods += usage.SchedulablePods
	}

	metrics.TotalCPUUsage = totalCPUUsage
	metrics.TotalMemoryUsage = totalMemoryUsage
	metrics.HealthyNodes = healthyNodes
	metrics.HighLoadNodes = highLoadNodes
	metrics.TotalPods = totalPods
	metrics.SchedulablePods = totalSchedulablePods

	// 计算平均利用率
	if len(nodeUsages) > 0 {
		metrics.AverageCPUUtilization = cpuUtilSum / float64(len(nodeUsages))
		metrics.AverageMemoryUtilization = memoryUtilSum / float64(len(nodeUsages))
	}

	return metrics
}

// ServeHeatmap 提供热力图 HTTP 服务
func (hg *HeatmapGenerator) ServeHeatmap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 生成热力图数据
	data, err := hg.GenerateHeatmapData(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate heatmap data: %v", err), http.StatusInternalServerError)
		return
	}

	// 检查请求格式
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		// 返回 JSON 数据
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	// 返回 HTML 页面
	w.Header().Set("Content-Type", "text/html")

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Kubernetes 集群资源热力图</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .metric-card { background: white; padding: 15px; border-radius: 8px; text-align: center; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-value { font-size: 1.8em; font-weight: bold; color: #007bff; }
        .metric-label { color: #6c757d; margin-top: 5px; font-size: 0.9em; }
        .heatmap-container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .node-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 10px; margin-top: 20px; }
        .node-card { padding: 15px; border-radius: 8px; text-align: center; color: white; font-weight: bold; transition: transform 0.2s; }
        .node-card:hover { transform: scale(1.05); cursor: pointer; }
        .node-name { font-size: 0.9em; margin-bottom: 5px; }
        .node-stats { font-size: 0.8em; opacity: 0.9; }
        .legend { display: flex; align-items: center; justify-content: center; margin: 20px 0; }
        .legend-item { display: flex; align-items: center; margin: 0 10px; }
        .legend-color { width: 20px; height: 20px; margin-right: 5px; border-radius: 3px; }
        .refresh-btn { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; margin-bottom: 20px; }
        .refresh-btn:hover { background: #0056b3; }
        .tooltip { position: absolute; background: rgba(0,0,0,0.8); color: white; padding: 10px; border-radius: 4px; font-size: 12px; pointer-events: none; z-index: 1000; }
        h1 { color: #333; margin: 0; }
        .cluster-info { color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.ClusterMetrics.ClusterName}} 集群资源热力图</h1>
            <div class="cluster-info">更新时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</div>
            <button class="refresh-btn" onclick="refreshData()">刷新数据</button>
        </div>
        
        <div class="metrics">
            <div class="metric-card">
                <div class="metric-value">{{.ClusterMetrics.TotalNodes}}</div>
                <div class="metric-label">总节点数</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{.ClusterMetrics.HealthyNodes}}</div>
                <div class="metric-label">健康节点</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{printf "%.1f" .ClusterMetrics.AverageCPUUtilization}}%</div>
                <div class="metric-label">平均 CPU 使用率</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{printf "%.1f" .ClusterMetrics.AverageMemoryUtilization}}%</div>
                <div class="metric-label">平均内存使用率</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{.ClusterMetrics.HighLoadNodes}}</div>
                <div class="metric-label">高负载节点</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">{{.ClusterMetrics.TotalPods}}</div>
                <div class="metric-label">总 Pod 数</div>
            </div>
        </div>
        
        <div class="heatmap-container">
            <h2>节点资源使用热力图</h2>
            <div class="legend">
                <div class="legend-item">
                    <div class="legend-color" style="background: #28a745;"></div>
                    <span>低负载 (0-50%)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-color" style="background: #ffc107;"></div>
                    <span>中等负载 (50-80%)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-color" style="background: #fd7e14;"></div>
                    <span>高负载 (80-95%)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-color" style="background: #dc3545;"></div>
                    <span>过载 (95%+)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-color" style="background: #6c757d;"></div>
                    <span>不可用</span>
                </div>
            </div>
            
            <div class="node-grid" id="nodeGrid">
                <!-- 节点卡片将在这里动态生成 -->
            </div>
        </div>
    </div>
    
    <div class="tooltip" id="tooltip" style="display: none;"></div>

    <script>
        const heatmapData = {{.}};
        
        function getNodeColor(node) {
            if (node.node_status !== 'Ready') {
                return '#6c757d'; // 灰色 - 不可用
            }
            
            const maxUtilization = Math.max(node.cpu_utilization_percent, node.memory_utilization_percent);
            
            if (maxUtilization >= 95) {
                return '#dc3545'; // 红色 - 过载
            } else if (maxUtilization >= 80) {
                return '#fd7e14'; // 橙色 - 高负载
            } else if (maxUtilization >= 50) {
                return '#ffc107'; // 黄色 - 中等负载
            } else {
                return '#28a745'; // 绿色 - 低负载
            }
        }
        
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }
        
        function formatCPU(millicores) {
            if (millicores >= 1000) {
                return (millicores / 1000).toFixed(1) + ' cores';
            }
            return millicores + 'm';
        }
        
        function renderNodes() {
            const nodeGrid = document.getElementById('nodeGrid');
            nodeGrid.innerHTML = '';
            
            heatmapData.nodes.forEach(node => {
                const nodeCard = document.createElement('div');
                nodeCard.className = 'node-card';
                nodeCard.style.backgroundColor = getNodeColor(node);
                
                nodeCard.innerHTML = ` + "`" + `
                    <div class="node-name">${node.node_name}</div>
                    <div class="node-stats">
                        CPU: ${node.cpu_utilization_percent.toFixed(1)}%<br>
                        内存: ${node.memory_utilization_percent.toFixed(1)}%<br>
                        Pods: ${node.pod_count}
                    </div>
                ` + "`" + `;
                
                // 添加工具提示
                nodeCard.addEventListener('mouseenter', (e) => {
                    const tooltip = document.getElementById('tooltip');
                    tooltip.innerHTML = ` + "`" + `
                        <strong>${node.node_name}</strong><br>
                        状态: ${node.node_status}<br>
                        CPU 容量: ${formatCPU(node.cpu_capacity_millicores)}<br>
                        CPU 使用: ${formatCPU(node.cpu_usage_millicores)} (${node.cpu_utilization_percent.toFixed(1)}%)<br>
                        内存容量: ${formatBytes(node.memory_capacity_bytes)}<br>
                        内存使用: ${formatBytes(node.memory_usage_bytes)} (${node.memory_utilization_percent.toFixed(1)}%)<br>
                        Pod 数量: ${node.pod_count}<br>
                        可调度 Pods: ${node.schedulable_pods}
                    ` + "`" + `;
                    tooltip.style.display = 'block';
                });
                
                nodeCard.addEventListener('mousemove', (e) => {
                    const tooltip = document.getElementById('tooltip');
                    tooltip.style.left = e.pageX + 10 + 'px';
                    tooltip.style.top = e.pageY + 10 + 'px';
                });
                
                nodeCard.addEventListener('mouseleave', () => {
                    document.getElementById('tooltip').style.display = 'none';
                });
                
                nodeGrid.appendChild(nodeCard);
            });
        }
        
        function refreshData() {
            window.location.reload();
        }
        
        // 初始化
        renderNodes();
        
        // 每30秒自动刷新
        setInterval(refreshData, 30000);
    </script>
</body>
</html>
`

	t := template.Must(template.New("heatmap").Parse(htmlTemplate))
	t.Execute(w, data)
}

func main() {
	// 解析命令行参数
	var (
		port = flag.String("port", "8082", "HTTP server port")
	)
	flag.Parse()
	
	// 这是一个模拟的 main 函数，实际使用时需要配置 Kubernetes 客户端
	klog.Info("Starting Kubernetes Cluster Resource Heatmap Generator...")
	
	// 创建模拟的热力图生成器
	hg := &HeatmapGenerator{
		clusterName:    "demo-cluster",
		updateInterval: 30 * time.Second,
	}
	
	// 设置 HTTP 路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 生成模拟数据并缓存
		data := hg.generateMockData()
		hg.cachedData = data
		hg.lastUpdate = time.Now()
		hg.ServeHeatmap(w, r)
	})
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		healthData := map[string]interface{}{
			"status": "healthy",
			"cluster_name": hg.clusterName,
			"generation_count": hg.generationCount,
			"error_count": hg.errorCount,
			"last_generation_duration_ms": hg.lastGenerationDuration.Milliseconds(),
			"last_update": hg.lastUpdate.Format(time.RFC3339),
			"cache_valid": hg.cachedData != nil && time.Since(hg.lastUpdate) < hg.updateInterval,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(healthData)
	})
	
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "# HELP heatmap_generation_total Total number of heatmap generations\n")
		fmt.Fprintf(w, "# TYPE heatmap_generation_total counter\n")
		fmt.Fprintf(w, "heatmap_generation_total %d\n", hg.generationCount)
		fmt.Fprintf(w, "# HELP heatmap_errors_total Total number of heatmap generation errors\n")
		fmt.Fprintf(w, "# TYPE heatmap_errors_total counter\n")
		fmt.Fprintf(w, "heatmap_errors_total %d\n", hg.errorCount)
		fmt.Fprintf(w, "# HELP heatmap_generation_duration_seconds Last heatmap generation duration\n")
		fmt.Fprintf(w, "# TYPE heatmap_generation_duration_seconds gauge\n")
		fmt.Fprintf(w, "heatmap_generation_duration_seconds %.3f\n", hg.lastGenerationDuration.Seconds())
	})
	
	// 启动 HTTP 服务器
	addr := ":" + *port
	klog.Infof("Starting HTTP server on %s", addr)
	klog.Infof("Access the heatmap at http://localhost%s", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		klog.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// generateMockData 生成模拟数据
func (hg *HeatmapGenerator) generateMockData() *HeatmapData {
	nodes := make([]NodeResourceUsage, 0)
	
	// 定义不同类型的节点配置
	nodeConfigs := []struct {
		name string
		cpuCores int64
		memoryGB int64
		nodeType string
		zone string
	}{
		{"master-1", 4000, 8, "master", "zone-a"},
		{"worker-1", 8000, 16, "worker", "zone-a"},
		{"worker-2", 8000, 16, "worker", "zone-b"},
		{"worker-3", 16000, 32, "worker", "zone-b"},
		{"worker-4", 4000, 8, "worker", "zone-c"},
		{"gpu-worker-1", 16000, 64, "gpu-worker", "zone-c"},
	}
	
	// 生成模拟节点数据
	for i, config := range nodeConfigs {
		// 根据节点类型设置不同的负载模式
		var cpuUsageRatio, memoryUsageRatio float64
		var podCount int
		var nodeStatus string = "Ready"
		
		switch config.nodeType {
		case "master":
			cpuUsageRatio = 0.3 + rand.Float64()*0.2 // 30%-50%
			memoryUsageRatio = 0.4 + rand.Float64()*0.2 // 40%-60%
			podCount = 5 + rand.Intn(10) // 5-15个系统Pod
		case "worker":
			cpuUsageRatio = 0.4 + rand.Float64()*0.4 // 40%-80%
			memoryUsageRatio = 0.5 + rand.Float64()*0.3 // 50%-80%
			podCount = 15 + rand.Intn(25) // 15-40个Pod
		case "gpu-worker":
			cpuUsageRatio = 0.6 + rand.Float64()*0.3 // 60%-90%
			memoryUsageRatio = 0.7 + rand.Float64()*0.2 // 70%-90%
			podCount = 8 + rand.Intn(12) // 8-20个Pod（GPU工作负载通常Pod数较少）
		}
		
		// 偶尔模拟节点故障
		if rand.Float64() < 0.1 { // 10%概率
			nodeStatus = "NotReady"
			cpuUsageRatio = 0
			memoryUsageRatio = 0
			podCount = 0
		}
		
		memoryCapacity := config.memoryGB * 1024 * 1024 * 1024
		cpuUsage := int64(float64(config.cpuCores) * cpuUsageRatio)
		memoryUsage := int64(float64(memoryCapacity) * memoryUsageRatio)
		
		node := NodeResourceUsage{
			NodeName:          config.name,
			CPUCapacity:       config.cpuCores,
			MemoryCapacity:    memoryCapacity,
			CPUUsage:          cpuUsage,
			MemoryUsage:       memoryUsage,
			PodCount:          podCount,
			NodeStatus:        nodeStatus,
			Labels: map[string]string{
				"zone": config.zone,
				"node-type": config.nodeType,
				"kubernetes.io/arch": "amd64",
				"kubernetes.io/os": "linux",
			},
			Taints: []corev1.Taint{},
			LastUpdated: time.Now(),
		}
		
		// 为master节点添加污点
		if config.nodeType == "master" {
			node.Taints = []corev1.Taint{
				{
					Key:    "node-role.kubernetes.io/master",
					Effect: corev1.TaintEffectNoSchedule,
				},
			}
		}
		
		// 计算利用率
		if node.CPUCapacity > 0 {
			node.CPUUtilization = float64(node.CPUUsage) / float64(node.CPUCapacity) * 100
		}
		if node.MemoryCapacity > 0 {
			node.MemoryUtilization = float64(node.MemoryUsage) / float64(node.MemoryCapacity) * 100
		}
		
		// 计算可调度Pod数
		node.SchedulablePods = hg.calculateSchedulablePods(node)
		
		nodes = append(nodes, node)
		
		// 添加一些延迟模拟真实环境
		if i < len(nodeConfigs)-1 {
			time.Sleep(time.Millisecond * time.Duration(10+rand.Intn(20)))
		}
	}
	
	// 计算集群指标
	clusterMetrics := hg.calculateClusterMetrics(nodes)
	clusterMetrics.ClusterName = hg.clusterName
	
	return &HeatmapData{
		Nodes:          nodes,
		ClusterMetrics: clusterMetrics,
		GeneratedAt:    time.Now(),
		UpdateInterval: hg.updateInterval,
	}
}
