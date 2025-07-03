// scheduler-visualizer.go
// Kubernetes 调度决策流程可视化工具实现

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog/v2"
)

// SchedulingDecision 调度决策
type SchedulingDecision struct {
	PodName       string            `json:"pod_name"`
	Namespace     string            `json:"namespace"`
	NodeName      string            `json:"node_name"`
	SchedulerName string            `json:"scheduler_name"`
	DecisionTime  time.Time         `json:"decision_time"`
	Latency       time.Duration     `json:"latency"`
	Success       bool              `json:"success"`
	Reason        string            `json:"reason"`
	FilterPhase   FilterPhaseResult `json:"filter_phase"`
	ScorePhase    ScorePhaseResult  `json:"score_phase"`
	BindPhase     BindPhaseResult   `json:"bind_phase"`
	Metadata      map[string]string `json:"metadata"`
}

// FilterPhaseResult 过滤阶段结果
type FilterPhaseResult struct {
	TotalNodes    int                     `json:"total_nodes"`
	FilteredNodes int                     `json:"filtered_nodes"`
	FailedFilters map[string]int          `json:"failed_filters"`
	Duration      time.Duration           `json:"duration"`
	NodeResults   map[string]FilterResult `json:"node_results"`
}

// FilterResult 单个节点的过滤结果
type FilterResult struct {
	NodeName      string   `json:"node_name"`
	Passed        bool     `json:"passed"`
	FailedPlugins []string `json:"failed_plugins"`
	Reason        string   `json:"reason"`
}

// ScorePhaseResult 评分阶段结果
type ScorePhaseResult struct {
	ScoredNodes  int                    `json:"scored_nodes"`
	Duration     time.Duration          `json:"duration"`
	NodeScores   map[string]NodeScore   `json:"node_scores"`
	PluginScores map[string]PluginScore `json:"plugin_scores"`
}

// NodeScore 节点评分
type NodeScore struct {
	NodeName        string  `json:"node_name"`
	TotalScore      int64   `json:"total_score"`
	NormalizedScore float64 `json:"normalized_score"`
	Rank            int     `json:"rank"`
}

// PluginScore 插件评分
type PluginScore struct {
	PluginName string           `json:"plugin_name"`
	NodeScores map[string]int64 `json:"node_scores"`
	Weight     int32            `json:"weight"`
}

// BindPhaseResult 绑定阶段结果
type BindPhaseResult struct {
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	BindTime time.Time     `json:"bind_time"`
}

// SchedulingVisualizer 调度可视化器
type SchedulingVisualizer struct {
	decisions []SchedulingDecision
	stats     SchedulingStats
}

// SchedulingStats 调度统计
type SchedulingStats struct {
	TotalDecisions     int           `json:"total_decisions"`
	SuccessfulBindings int           `json:"successful_bindings"`
	FailedSchedulings  int           `json:"failed_schedulings"`
	AverageLatency     time.Duration `json:"average_latency"`
	P95Latency         time.Duration `json:"p95_latency"`
	P99Latency         time.Duration `json:"p99_latency"`
	Throughput         float64       `json:"throughput"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// NewSchedulingVisualizer 创建调度可视化器
func NewSchedulingVisualizer() *SchedulingVisualizer {
	return &SchedulingVisualizer{
		decisions: make([]SchedulingDecision, 0),
		stats:     SchedulingStats{},
	}
}

// CollectDecisions 收集调度决策数据（模拟）
func (sv *SchedulingVisualizer) CollectDecisions(count int) {
	klog.Infof("Collecting %d scheduling decisions...", count)

	for i := 0; i < count; i++ {
		decision := sv.simulateSchedulingDecision(i)
		sv.decisions = append(sv.decisions, decision)
	}

	// 更新统计信息
	sv.updateStats()
	klog.Infof("Collected %d decisions, success rate: %.2f%%",
		len(sv.decisions), float64(sv.stats.SuccessfulBindings)/float64(sv.stats.TotalDecisions)*100)
}

// simulateSchedulingDecision 模拟调度决策过程
func (sv *SchedulingVisualizer) simulateSchedulingDecision(index int) SchedulingDecision {
	startTime := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)
	success := rand.Float32() > 0.1 // 90% 成功率

	// 模拟过滤阶段
	totalNodes := 10 + rand.Intn(40)                      // 10-50个节点
	filteredNodes := totalNodes - rand.Intn(totalNodes/2) // 过滤掉一部分节点
	if !success {
		filteredNodes = 0 // 失败时没有可用节点
	}

	filterDuration := time.Duration(rand.Intn(50)+10) * time.Millisecond
	filterPhase := FilterPhaseResult{
		TotalNodes:    totalNodes,
		FilteredNodes: filteredNodes,
		FailedFilters: map[string]int{
			"NodeResourcesFit":  rand.Intn(5),
			"NodeAffinity":      rand.Intn(3),
			"PodTopologySpread": rand.Intn(2),
			"TaintToleration":   rand.Intn(4),
		},
		Duration:    filterDuration,
		NodeResults: make(map[string]FilterResult),
	}

	// 模拟评分阶段
	scoreDuration := time.Duration(rand.Intn(30)+5) * time.Millisecond
	nodeScores := make(map[string]NodeScore)
	pluginScores := make(map[string]PluginScore)

	if success && filteredNodes > 0 {
		for i := 0; i < filteredNodes; i++ {
			nodeName := fmt.Sprintf("node-%d", i+1)
			totalScore := int64(rand.Intn(100))
			nodeScores[nodeName] = NodeScore{
				NodeName:        nodeName,
				TotalScore:      totalScore,
				NormalizedScore: float64(totalScore) / 100.0,
				Rank:            i + 1,
			}
		}

		// 模拟插件评分
		plugins := []string{"NodeResourcesFit", "NodeAffinity", "PodTopologySpread", "ImageLocality"}
		for _, plugin := range plugins {
			scores := make(map[string]int64)
			for nodeName := range nodeScores {
				scores[nodeName] = int64(rand.Intn(100))
			}
			pluginScores[plugin] = PluginScore{
				PluginName: plugin,
				NodeScores: scores,
				Weight:     int32(rand.Intn(10) + 1),
			}
		}
	}

	scorePhase := ScorePhaseResult{
		ScoredNodes:  filteredNodes,
		Duration:     scoreDuration,
		NodeScores:   nodeScores,
		PluginScores: pluginScores,
	}

	// 模拟绑定阶段
	bindDuration := time.Duration(rand.Intn(20)+5) * time.Millisecond
	bindPhase := BindPhaseResult{
		Success:  success,
		Duration: bindDuration,
		BindTime: startTime.Add(filterDuration + scoreDuration + bindDuration),
	}

	if !success {
		bindPhase.Error = "No suitable nodes found"
	}

	// 选择最佳节点
	var selectedNode string
	var reason string
	if success && len(nodeScores) > 0 {
		// 选择评分最高的节点
		var bestScore int64 = -1
		for nodeName, score := range nodeScores {
			if score.TotalScore > bestScore {
				bestScore = score.TotalScore
				selectedNode = nodeName
			}
		}
		reason = "Successfully scheduled"
	} else {
		reason = "No suitable nodes available"
	}

	totalLatency := filterDuration + scoreDuration + bindDuration

	return SchedulingDecision{
		PodName:       fmt.Sprintf("pod-%d", index+1),
		Namespace:     fmt.Sprintf("namespace-%d", rand.Intn(5)+1),
		NodeName:      selectedNode,
		SchedulerName: "default-scheduler",
		DecisionTime:  startTime,
		Latency:       totalLatency,
		Success:       success,
		Reason:        reason,
		FilterPhase:   filterPhase,
		ScorePhase:    scorePhase,
		BindPhase:     bindPhase,
		Metadata: map[string]string{
			"priority_class": fmt.Sprintf("priority-%d", rand.Intn(3)),
			"workload_type":  []string{"web", "batch", "ml", "database"}[rand.Intn(4)],
		},
	}
}

// updateStats 更新统计信息
func (sv *SchedulingVisualizer) updateStats() {
	if len(sv.decisions) == 0 {
		return
	}

	successCount := 0
	var totalLatency time.Duration
	latencies := make([]time.Duration, 0, len(sv.decisions))

	for _, decision := range sv.decisions {
		if decision.Success {
			successCount++
		}
		totalLatency += decision.Latency
		latencies = append(latencies, decision.Latency)
	}

	// 计算百分位数
	p95Latency, p99Latency := calculatePercentiles(latencies)

	// 计算吞吐量（每秒决策数）
	if len(sv.decisions) > 1 {
		firstTime := sv.decisions[0].DecisionTime
		lastTime := sv.decisions[len(sv.decisions)-1].DecisionTime
		duration := lastTime.Sub(firstTime).Seconds()
		if duration > 0 {
			sv.stats.Throughput = float64(len(sv.decisions)) / duration
		}
	}

	sv.stats = SchedulingStats{
		TotalDecisions:     len(sv.decisions),
		SuccessfulBindings: successCount,
		FailedSchedulings:  len(sv.decisions) - successCount,
		AverageLatency:     totalLatency / time.Duration(len(sv.decisions)),
		P95Latency:         p95Latency,
		P99Latency:         p99Latency,
		Throughput:         sv.stats.Throughput,
		LastUpdated:        time.Now(),
	}
}

// calculatePercentiles 计算百分位数
func calculatePercentiles(latencies []time.Duration) (time.Duration, time.Duration) {
	if len(latencies) == 0 {
		return 0, 0
	}

	// 简单排序
	for i := 0; i < len(latencies)-1; i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	p95Index := int(float64(len(latencies)) * 0.95)
	p99Index := int(float64(len(latencies)) * 0.99)

	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}

	return latencies[p95Index], latencies[p99Index]
}

// GenerateFlowChart 生成 Mermaid 流程图
func (sv *SchedulingVisualizer) GenerateFlowChart() string {
	mermaid := `graph TD
`
	mermaid += `    A[Pod 提交] --> B[调度器接收]
`
	mermaid += `    B --> C[过滤阶段]
`
	mermaid += `    C --> D{是否有可用节点?}
`
	mermaid += `    D -->|是| E[评分阶段]
`
	mermaid += `    D -->|否| F[调度失败]
`
	mermaid += `    E --> G[选择最佳节点]
`
	mermaid += `    G --> H[绑定阶段]
`
	mermaid += `    H --> I{绑定成功?}
`
	mermaid += `    I -->|是| J[调度成功]
`
	mermaid += `    I -->|否| K[绑定失败]
`
	mermaid += `    F --> L[记录事件]
`
	mermaid += `    J --> M[Pod 运行]
`
	mermaid += `    K --> L
`

	// 添加样式
	mermaid += `    classDef success fill:#d4edda,stroke:#155724,stroke-width:2px
`
	mermaid += `    classDef failure fill:#f8d7da,stroke:#721c24,stroke-width:2px
`
	mermaid += `    classDef process fill:#cce5ff,stroke:#004085,stroke-width:2px
`
	mermaid += `    class J,M success
`
	mermaid += `    class F,K,L failure
`
	mermaid += `    class B,C,E,G,H process
`

	return mermaid
}

// ServeHTTP 提供 HTTP 服务
func (sv *SchedulingVisualizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/decisions":
		sv.handleDecisions(w, r)
	case "/api/stats":
		sv.handleStats(w, r)
	case "/api/flowchart":
		sv.handleFlowChart(w, r)
	case "/":
		sv.handleIndex(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleDecisions 处理决策数据请求
func (sv *SchedulingVisualizer) handleDecisions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 支持分页
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 获取数据切片
	start := offset
	end := offset + limit
	if start >= len(sv.decisions) {
		start = len(sv.decisions)
	}
	if end > len(sv.decisions) {
		end = len(sv.decisions)
	}

	response := map[string]interface{}{
		"decisions": sv.decisions[start:end],
		"total":     len(sv.decisions),
		"offset":    offset,
		"limit":     limit,
	}

	json.NewEncoder(w).Encode(response)
}

// handleStats 处理统计数据请求
func (sv *SchedulingVisualizer) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sv.stats)
}

// handleFlowChart 处理流程图请求
func (sv *SchedulingVisualizer) handleFlowChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(sv.GenerateFlowChart()))
}

// handleIndex 处理主页请求
func (sv *SchedulingVisualizer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Kubernetes 调度器可视化</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .stat-card { background: #f8f9fa; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-value { font-size: 2em; font-weight: bold; color: #007bff; }
        .stat-label { color: #6c757d; margin-top: 5px; }
        .chart-container { background: white; padding: 20px; border-radius: 8px; margin-bottom: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .mermaid { text-align: center; }
        h1, h2 { color: #333; }
        .refresh-btn { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #0056b3; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Kubernetes 调度器可视化</h1>
        <button class="refresh-btn" onclick="refreshData()">刷新数据</button>
        
        <div class="stats" id="stats">
            <!-- 统计卡片将在这里动态生成 -->
        </div>
        
        <div class="chart-container">
            <h2>调度决策流程图</h2>
            <div class="mermaid" id="flowchart">
                <!-- Mermaid 流程图将在这里显示 -->
            </div>
        </div>
        
        <div class="chart-container">
            <h2>调度延迟分布</h2>
            <canvas id="latencyChart" width="400" height="200"></canvas>
        </div>
        
        <div class="chart-container">
            <h2>过滤器失败统计</h2>
            <canvas id="filterChart" width="400" height="200"></canvas>
        </div>
    </div>

    <script>
        let latencyChart, filterChart;
        
        // 初始化 Mermaid
        mermaid.initialize({ startOnLoad: true });
        
        // 加载数据
        async function loadData() {
            try {
                // 加载统计数据
                const statsResponse = await fetch('/api/stats');
                const stats = await statsResponse.json();
                updateStats(stats);
                
                // 加载决策数据
                const decisionsResponse = await fetch('/api/decisions?limit=100');
                const decisionsData = await decisionsResponse.json();
                updateCharts(decisionsData.decisions);
                
                // 加载流程图
                const flowchartResponse = await fetch('/api/flowchart');
                const flowchartData = await flowchartResponse.text();
                updateFlowChart(flowchartData);
                
            } catch (error) {
                console.error('Error loading data:', error);
            }
        }
        
        // 更新统计卡片
        function updateStats(stats) {
            const statsContainer = document.getElementById('stats');
            statsContainer.innerHTML = ` + "`" + `
                <div class="stat-card">
                    <div class="stat-value">${stats.total_decisions || 0}</div>
                    <div class="stat-label">总调度决策</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">${((stats.successful_bindings / stats.total_decisions) * 100).toFixed(1)}%</div>
                    <div class="stat-label">成功率</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">${(stats.average_latency / 1000000).toFixed(1)}ms</div>
                    <div class="stat-label">平均延迟</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">${stats.throughput.toFixed(1)}</div>
                    <div class="stat-label">吞吐量 (决策/秒)</div>
                </div>
            ` + "`" + `;
        }
        
        // 更新图表
        function updateCharts(decisions) {
            updateLatencyChart(decisions);
            updateFilterChart(decisions);
        }
        
        // 更新延迟分布图
        function updateLatencyChart(decisions) {
            const latencies = decisions.map(d => d.latency / 1000000); // 转换为毫秒
            const bins = createHistogramBins(latencies, 10);
            
            const ctx = document.getElementById('latencyChart').getContext('2d');
            
            if (latencyChart) {
                latencyChart.destroy();
            }
            
            latencyChart = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: bins.labels,
                    datasets: [{
                        label: '调度延迟分布',
                        data: bins.counts,
                        backgroundColor: 'rgba(54, 162, 235, 0.6)',
                        borderColor: 'rgba(54, 162, 235, 1)',
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: '频次'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: '延迟 (ms)'
                            }
                        }
                    }
                }
            });
        }
        
        // 更新过滤器失败统计图
        function updateFilterChart(decisions) {
            const filterFailures = {};
            
            decisions.forEach(decision => {
                if (decision.filter_phase && decision.filter_phase.failed_filters) {
                    Object.entries(decision.filter_phase.failed_filters).forEach(([filter, count]) => {
                        filterFailures[filter] = (filterFailures[filter] || 0) + count;
                    });
                }
            });
            
            const ctx = document.getElementById('filterChart').getContext('2d');
            
            if (filterChart) {
                filterChart.destroy();
            }
            
            filterChart = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: Object.keys(filterFailures),
                    datasets: [{
                        data: Object.values(filterFailures),
                        backgroundColor: [
                            'rgba(255, 99, 132, 0.6)',
                            'rgba(54, 162, 235, 0.6)',
                            'rgba(255, 205, 86, 0.6)',
                            'rgba(75, 192, 192, 0.6)',
                            'rgba(153, 102, 255, 0.6)'
                        ]
                    }]
                },
                options: {
                    responsive: true,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        },
                        title: {
                            display: true,
                            text: '过滤器失败原因分布'
                        }
                    }
                }
            });
        }
        
        // 更新流程图
        function updateFlowChart(flowchartData) {
            const flowchartContainer = document.getElementById('flowchart');
            flowchartContainer.innerHTML = flowchartData;
            mermaid.init(undefined, flowchartContainer);
        }
        
        // 创建直方图分箱
        function createHistogramBins(data, binCount) {
            if (data.length === 0) return { labels: [], counts: [] };
            
            const min = Math.min(...data);
            const max = Math.max(...data);
            const binWidth = (max - min) / binCount;
            
            const bins = Array(binCount).fill(0);
            const labels = [];
            
            for (let i = 0; i < binCount; i++) {
                const binStart = min + i * binWidth;
                const binEnd = min + (i + 1) * binWidth;
                labels.push(` + "`${binStart.toFixed(1)}-${binEnd.toFixed(1)}`" + `);
            }
            
            data.forEach(value => {
                const binIndex = Math.min(Math.floor((value - min) / binWidth), binCount - 1);
                bins[binIndex]++;
            });
            
            return { labels, counts: bins };
        }
        
        // 刷新数据
        function refreshData() {
            loadData();
        }
        
        // 页面加载时初始化
        window.onload = function() {
            loadData();
            // 每30秒自动刷新
            setInterval(loadData, 30000);
        };
    </script>
</body>
</html>
`

	t := template.Must(template.New("index").Parse(htmlTemplate))
	t.Execute(w, nil)
}

// main 函数
func main() {
	// 解析命令行参数
	var (
		port = flag.String("port", "8080", "HTTP server port")
	)
	flag.Parse()

	// 创建可视化器
	visualizer := NewSchedulingVisualizer()

	// 收集模拟数据
	visualizer.CollectDecisions(200)

	// 启动 HTTP 服务
	http.Handle("/", visualizer)

	addr := ":" + *port
	klog.Infof("Starting scheduler visualizer on %s", addr)
	klog.Infof("Access the scheduler visualizer at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		klog.Fatalf("Failed to start server: %v", err)
	}
}
