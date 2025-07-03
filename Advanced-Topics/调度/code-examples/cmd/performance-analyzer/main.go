// performance-analyzer.go
// Kubernetes 调度性能趋势分析器实现

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"time"

	k8slog "k8s.io/klog/v2"
)

// PerformanceMetric 性能指标
type PerformanceMetric struct {
	Timestamp         time.Time `json:"timestamp"`
	SchedulingLatency float64   `json:"scheduling_latency_ms"`
	Throughput        float64   `json:"throughput_pods_per_sec"`
	SuccessRate       float64   `json:"success_rate_percent"`
	QueueLength       int       `json:"queue_length"`
	NodeUtilization   float64   `json:"node_utilization_percent"`
	FilterLatency     float64   `json:"filter_latency_ms"`
	ScoreLatency      float64   `json:"score_latency_ms"`
	BindLatency       float64   `json:"bind_latency_ms"`
}

// TrendAnalysis 趋势分析结果
type TrendAnalysis struct {
	TimeRange       TimeRange           `json:"time_range"`
	Summary         TrendSummary        `json:"summary"`
	Anomalies       []Anomaly           `json:"anomalies"`
	Recommendations []string            `json:"recommendations"`
	Predictions     []Prediction        `json:"predictions"`
	Metrics         []PerformanceMetric `json:"metrics"`
	GeneratedAt     time.Time           `json:"generated_at"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TrendSummary 趋势摘要
type TrendSummary struct {
	AverageLatency     float64 `json:"average_latency_ms"`
	MaxLatency         float64 `json:"max_latency_ms"`
	MinLatency         float64 `json:"min_latency_ms"`
	P95Latency         float64 `json:"p95_latency_ms"`
	P99Latency         float64 `json:"p99_latency_ms"`
	LatencyTrend       string  `json:"latency_trend"`
	ThroughputTrend    string  `json:"throughput_trend"`
	SuccessRateTrend   string  `json:"success_rate_trend"`
	AverageThroughput  float64 `json:"average_throughput"`
	AverageSuccessRate float64 `json:"average_success_rate"`
}

// Anomaly 异常检测结果
type Anomaly struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Impact      string    `json:"impact"`
}

// Prediction 性能预测
type Prediction struct {
	Timestamp      time.Time `json:"timestamp"`
	MetricType     string    `json:"metric_type"`
	PredictedValue float64   `json:"predicted_value"`
	Confidence     float64   `json:"confidence_percent"`
	Trend          string    `json:"trend"`
}

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	metrics        []PerformanceMetric
	analysisWindow time.Duration
	lastAnalysis   time.Time
	cachedAnalysis *TrendAnalysis
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metrics:        make([]PerformanceMetric, 0),
		analysisWindow: 1 * time.Hour, // 分析最近1小时的数据
	}
}

// CollectMetrics 收集性能指标（模拟）
func (pa *PerformanceAnalyzer) CollectMetrics(duration time.Duration, interval time.Duration) {
	k8slog.Infof("Collecting performance metrics for %v with %v interval", duration, interval)

	startTime := time.Now().Add(-duration)
	currentTime := startTime

	for currentTime.Before(time.Now()) {
		metric := pa.simulateMetric(currentTime)
		pa.metrics = append(pa.metrics, metric)
		currentTime = currentTime.Add(interval)
	}

	// 保持最近24小时的数据
	cutoffTime := time.Now().Add(-24 * time.Hour)
	pa.pruneOldMetrics(cutoffTime)

	k8slog.Infof("Collected %d performance metrics", len(pa.metrics))
}

// simulateMetric 模拟性能数据
func (pa *PerformanceAnalyzer) simulateMetric(timestamp time.Time) PerformanceMetric {
	// 模拟一天中的性能变化模式
	hour := timestamp.Hour()

	// 基础延迟：夜间较低，白天较高
	baseLatency := 50.0 + 30.0*math.Sin(float64(hour)*math.Pi/12.0)

	// 添加随机波动
	latencyVariation := (rand.Float64() - 0.5) * 20.0
	schedulingLatency := math.Max(10.0, baseLatency+latencyVariation)

	// 偶尔添加异常峰值
	if rand.Float64() < 0.05 { // 5% 概率出现异常
		schedulingLatency += rand.Float64() * 200.0
	}

	// 吞吐量与延迟成反比
	baseThroughput := 100.0 - (schedulingLatency-50.0)/10.0
	throughput := math.Max(10.0, baseThroughput+rand.Float64()*20.0-10.0)

	// 成功率通常很高，但在高延迟时会下降
	successRate := 98.0
	if schedulingLatency > 150.0 {
		successRate -= (schedulingLatency - 150.0) / 10.0
	}
	successRate = math.Max(85.0, math.Min(100.0, successRate+rand.Float64()*4.0-2.0))

	// 队列长度与延迟相关
	queueLength := int(schedulingLatency/10.0) + rand.Intn(20)

	// 节点利用率
	nodeUtilization := 60.0 + 20.0*math.Sin(float64(hour)*math.Pi/12.0) + rand.Float64()*20.0 - 10.0
	nodeUtilization = math.Max(30.0, math.Min(95.0, nodeUtilization))

	// 各阶段延迟
	filterLatency := schedulingLatency * (0.4 + rand.Float64()*0.2) // 40-60%
	scoreLatency := schedulingLatency * (0.3 + rand.Float64()*0.2)  // 30-50%
	bindLatency := schedulingLatency - filterLatency - scoreLatency
	bindLatency = math.Max(5.0, bindLatency)

	return PerformanceMetric{
		Timestamp:         timestamp,
		SchedulingLatency: schedulingLatency,
		Throughput:        throughput,
		SuccessRate:       successRate,
		QueueLength:       queueLength,
		NodeUtilization:   nodeUtilization,
		FilterLatency:     filterLatency,
		ScoreLatency:      scoreLatency,
		BindLatency:       bindLatency,
	}
}

// pruneOldMetrics 清理旧指标
func (pa *PerformanceAnalyzer) pruneOldMetrics(cutoffTime time.Time) {
	var filteredMetrics []PerformanceMetric
	for _, metric := range pa.metrics {
		if metric.Timestamp.After(cutoffTime) {
			filteredMetrics = append(filteredMetrics, metric)
		}
	}
	pa.metrics = filteredMetrics
}

// AnalyzeTrends 分析性能趋势
func (pa *PerformanceAnalyzer) AnalyzeTrends() *TrendAnalysis {
	// 检查缓存
	if pa.cachedAnalysis != nil && time.Since(pa.lastAnalysis) < 5*time.Minute {
		return pa.cachedAnalysis
	}

	k8slog.Info("Analyzing performance trends...")

	if len(pa.metrics) == 0 {
		return &TrendAnalysis{
			GeneratedAt: time.Now(),
			Metrics:     []PerformanceMetric{},
		}
	}

	// 获取分析窗口内的数据
	windowStart := time.Now().Add(-pa.analysisWindow)
	windowMetrics := pa.getMetricsInWindow(windowStart, time.Now())

	if len(windowMetrics) == 0 {
		return &TrendAnalysis{
			GeneratedAt: time.Now(),
			Metrics:     []PerformanceMetric{},
		}
	}

	// 计算趋势摘要
	summary := pa.calculateSummary(windowMetrics)

	// 检测异常
	anomalies := pa.detectAnomalies(windowMetrics)

	// 生成建议
	recommendations := pa.generateRecommendations(summary, anomalies)

	// 生成预测
	predictions := pa.generatePredictions(windowMetrics)

	analysis := &TrendAnalysis{
		TimeRange: TimeRange{
			Start: windowMetrics[0].Timestamp,
			End:   windowMetrics[len(windowMetrics)-1].Timestamp,
		},
		Summary:         summary,
		Anomalies:       anomalies,
		Recommendations: recommendations,
		Predictions:     predictions,
		Metrics:         windowMetrics,
		GeneratedAt:     time.Now(),
	}

	// 更新缓存
	pa.cachedAnalysis = analysis
	pa.lastAnalysis = time.Now()

	k8slog.Infof("Analysis completed: %d metrics, %d anomalies, %d recommendations",
		len(windowMetrics), len(anomalies), len(recommendations))

	return analysis
}

// getMetricsInWindow 获取时间窗口内的指标
func (pa *PerformanceAnalyzer) getMetricsInWindow(start, end time.Time) []PerformanceMetric {
	var windowMetrics []PerformanceMetric
	for _, metric := range pa.metrics {
		if metric.Timestamp.After(start) && metric.Timestamp.Before(end) {
			windowMetrics = append(windowMetrics, metric)
		}
	}
	return windowMetrics
}

// calculateSummary 计算趋势摘要
func (pa *PerformanceAnalyzer) calculateSummary(metrics []PerformanceMetric) TrendSummary {
	if len(metrics) == 0 {
		return TrendSummary{}
	}

	// 提取延迟数据
	latencies := make([]float64, len(metrics))
	var totalLatency, totalThroughput, totalSuccessRate float64

	for i, metric := range metrics {
		latencies[i] = metric.SchedulingLatency
		totalLatency += metric.SchedulingLatency
		totalThroughput += metric.Throughput
		totalSuccessRate += metric.SuccessRate
	}

	// 排序以计算百分位数
	sort.Float64s(latencies)

	// 计算百分位数
	p95Index := int(float64(len(latencies)) * 0.95)
	p99Index := int(float64(len(latencies)) * 0.99)
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}

	// 计算趋势
	latencyTrend := pa.calculateTrend(metrics, "latency")
	throughputTrend := pa.calculateTrend(metrics, "throughput")
	successRateTrend := pa.calculateTrend(metrics, "success_rate")

	return TrendSummary{
		AverageLatency:     totalLatency / float64(len(metrics)),
		MaxLatency:         latencies[len(latencies)-1],
		MinLatency:         latencies[0],
		P95Latency:         latencies[p95Index],
		P99Latency:         latencies[p99Index],
		LatencyTrend:       latencyTrend,
		ThroughputTrend:    throughputTrend,
		SuccessRateTrend:   successRateTrend,
		AverageThroughput:  totalThroughput / float64(len(metrics)),
		AverageSuccessRate: totalSuccessRate / float64(len(metrics)),
	}
}

// calculateTrend 计算趋势方向
func (pa *PerformanceAnalyzer) calculateTrend(metrics []PerformanceMetric, metricType string) string {
	if len(metrics) < 2 {
		return "stable"
	}

	// 简单线性回归计算趋势
	n := float64(len(metrics))
	var sumX, sumY, sumXY, sumX2 float64

	for i, metric := range metrics {
		x := float64(i)
		var y float64
		switch metricType {
		case "latency":
			y = metric.SchedulingLatency
		case "throughput":
			y = metric.Throughput
		case "success_rate":
			y = metric.SuccessRate
		}

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// 判断趋势
	if math.Abs(slope) < 0.1 {
		return "stable"
	} else if slope > 0 {
		if metricType == "latency" {
			return "worsening" // 延迟增加是变差
		} else {
			return "improving" // 吞吐量和成功率增加是改善
		}
	} else {
		if metricType == "latency" {
			return "improving" // 延迟减少是改善
		} else {
			return "worsening" // 吞吐量和成功率减少是变差
		}
	}
}

// detectAnomalies 检测异常
func (pa *PerformanceAnalyzer) detectAnomalies(metrics []PerformanceMetric) []Anomaly {
	var anomalies []Anomaly

	for _, metric := range metrics {
		// 检测高延迟异常
		if metric.SchedulingLatency > 200.0 {
			severity := "medium"
			if metric.SchedulingLatency > 500.0 {
				severity = "high"
			}
			anomalies = append(anomalies, Anomaly{
				Timestamp:   metric.Timestamp,
				Type:        "high_latency",
				Severity:    severity,
				Description: fmt.Sprintf("调度延迟异常高: %.1fms", metric.SchedulingLatency),
				Value:       metric.SchedulingLatency,
				Threshold:   200.0,
				Impact:      "可能影响 Pod 启动时间和用户体验",
			})
		}

		// 检测低吞吐量异常
		if metric.Throughput < 30.0 {
			anomalies = append(anomalies, Anomaly{
				Timestamp:   metric.Timestamp,
				Type:        "low_throughput",
				Severity:    "medium",
				Description: fmt.Sprintf("调度吞吐量过低: %.1f pods/sec", metric.Throughput),
				Value:       metric.Throughput,
				Threshold:   30.0,
				Impact:      "可能导致 Pod 调度积压",
			})
		}

		// 检测低成功率异常
		if metric.SuccessRate < 95.0 {
			severity := "medium"
			if metric.SuccessRate < 90.0 {
				severity = "high"
			}
			anomalies = append(anomalies, Anomaly{
				Timestamp:   metric.Timestamp,
				Type:        "low_success_rate",
				Severity:    severity,
				Description: fmt.Sprintf("调度成功率过低: %.1f%%", metric.SuccessRate),
				Value:       metric.SuccessRate,
				Threshold:   95.0,
				Impact:      "可能导致 Pod 无法正常调度",
			})
		}
	}

	return anomalies
}

// generateRecommendations 生成优化建议
func (pa *PerformanceAnalyzer) generateRecommendations(summary TrendSummary, anomalies []Anomaly) []string {
	var recommendations []string

	// 基于趋势的建议
	if summary.LatencyTrend == "worsening" {
		recommendations = append(recommendations, "调度延迟呈上升趋势，建议检查节点资源状况和调度器配置")
	}

	if summary.ThroughputTrend == "worsening" {
		recommendations = append(recommendations, "调度吞吐量呈下降趋势，建议优化调度器性能或增加调度器实例")
	}

	if summary.SuccessRateTrend == "worsening" {
		recommendations = append(recommendations, "调度成功率呈下降趋势，建议检查资源约束和节点可用性")
	}

	// 基于平均值的建议
	if summary.AverageLatency > 150.0 {
		recommendations = append(recommendations, "平均调度延迟较高，建议优化调度算法或增加节点资源")
	}

	if summary.AverageThroughput < 50.0 {
		recommendations = append(recommendations, "平均调度吞吐量较低，建议检查调度器瓶颈")
	}

	if summary.AverageSuccessRate < 98.0 {
		recommendations = append(recommendations, "平均调度成功率较低，建议检查资源配额和节点标签")
	}

	// 基于异常的建议
	highSeverityCount := 0
	for _, anomaly := range anomalies {
		if anomaly.Severity == "high" {
			highSeverityCount++
		}
	}

	if highSeverityCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("检测到 %d 个高严重性异常，建议立即检查系统状态", highSeverityCount))
	}

	// 通用建议
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "系统运行正常，建议继续监控关键指标")
	}

	return recommendations
}

// generatePredictions 生成性能预测
func (pa *PerformanceAnalyzer) generatePredictions(metrics []PerformanceMetric) []Prediction {
	var predictions []Prediction

	if len(metrics) < 10 {
		return predictions
	}

	// 预测未来30分钟的延迟
	latencyTrend := pa.calculateTrend(metrics, "latency")
	currentLatency := metrics[len(metrics)-1].SchedulingLatency

	// 简单的线性预测
	var predictedLatency float64
	var confidence float64

	switch latencyTrend {
	case "improving":
		predictedLatency = currentLatency * 0.95 // 预测改善5%
		confidence = 75.0
	case "worsening":
		predictedLatency = currentLatency * 1.1 // 预测恶化10%
		confidence = 70.0
	default:
		predictedLatency = currentLatency // 保持稳定
		confidence = 85.0
	}

	predictions = append(predictions, Prediction{
		Timestamp:      time.Now().Add(30 * time.Minute),
		MetricType:     "scheduling_latency",
		PredictedValue: predictedLatency,
		Confidence:     confidence,
		Trend:          latencyTrend,
	})

	return predictions
}

// ServeAnalysis 提供性能分析 HTTP 服务
func (pa *PerformanceAnalyzer) ServeAnalysis(w http.ResponseWriter, r *http.Request) {
	// 执行趋势分析
	analysis := pa.AnalyzeTrends()

	// 检查请求格式
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		// 返回 JSON 数据
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analysis)
		return
	}

	// 返回 HTML 页面
	w.Header().Set("Content-Type", "text/html")

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Kubernetes 调度器性能趋势分析</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .summary-card { background: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .summary-value { font-size: 1.5em; font-weight: bold; color: #007bff; }
        .summary-label { color: #6c757d; margin-top: 5px; }
        .trend { font-size: 0.9em; margin-top: 5px; }
        .trend.improving { color: #28a745; }
        .trend.worsening { color: #dc3545; }
        .trend.stable { color: #6c757d; }
        .chart-container { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .anomalies { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .anomaly { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .anomaly.high { background: #f8d7da; border-left: 4px solid #dc3545; }
        .anomaly.medium { background: #fff3cd; border-left: 4px solid #ffc107; }
        .anomaly.low { background: #d1ecf1; border-left: 4px solid #17a2b8; }
        .recommendations { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .recommendation { padding: 10px; margin: 10px 0; background: #e7f3ff; border-left: 4px solid #007bff; border-radius: 4px; }
        h1, h2 { color: #333; }
        .refresh-btn { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #0056b3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Kubernetes 调度器性能趋势分析</h1>
            <p>分析时间范围: {{.TimeRange.Start.Format "2006-01-02 15:04:05"}} - {{.TimeRange.End.Format "2006-01-02 15:04:05"}}</p>
            <p>生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
            <button class="refresh-btn" onclick="refreshData()">刷新数据</button>
        </div>
        
        <div class="summary">
            <div class="summary-card">
                <div class="summary-value">{{printf "%.1f" .Summary.AverageLatency}}ms</div>
                <div class="summary-label">平均调度延迟</div>
                <div class="trend {{.Summary.LatencyTrend}}">趋势: {{.Summary.LatencyTrend}}</div>
            </div>
            <div class="summary-card">
                <div class="summary-value">{{printf "%.1f" .Summary.P95Latency}}ms</div>
                <div class="summary-label">P95 延迟</div>
            </div>
            <div class="summary-card">
                <div class="summary-value">{{printf "%.1f" .Summary.AverageThroughput}}</div>
                <div class="summary-label">平均吞吐量 (pods/sec)</div>
                <div class="trend {{.Summary.ThroughputTrend}}">趋势: {{.Summary.ThroughputTrend}}</div>
            </div>
            <div class="summary-card">
                <div class="summary-value">{{printf "%.1f" .Summary.AverageSuccessRate}}%</div>
                <div class="summary-label">平均成功率</div>
                <div class="trend {{.Summary.SuccessRateTrend}}">趋势: {{.Summary.SuccessRateTrend}}</div>
            </div>
        </div>
        
        <div class="chart-container">
            <h2>调度延迟趋势</h2>
            <canvas id="latencyChart" width="400" height="200"></canvas>
        </div>
        
        <div class="chart-container">
            <h2>吞吐量和成功率趋势</h2>
            <canvas id="throughputChart" width="400" height="200"></canvas>
        </div>
        
        {{if .Anomalies}}
        <div class="anomalies">
            <h2>异常检测 ({{len .Anomalies}} 个异常)</h2>
            {{range .Anomalies}}
            <div class="anomaly {{.Severity}}">
                <strong>{{.Timestamp.Format "15:04:05"}}</strong> - {{.Description}}
                <br><small>影响: {{.Impact}}</small>
            </div>
            {{end}}
        </div>
        {{end}}
        
        <div class="recommendations">
            <h2>优化建议</h2>
            {{range .Recommendations}}
            <div class="recommendation">{{.}}</div>
            {{end}}
        </div>
    </div>

    <script>
        const analysisData = {{.}};
        
        // 渲染延迟趋势图
        function renderLatencyChart() {
            const ctx = document.getElementById('latencyChart').getContext('2d');
            const timestamps = analysisData.metrics.map(m => new Date(m.timestamp).toLocaleTimeString());
            const latencies = analysisData.metrics.map(m => m.scheduling_latency_ms);
            
            new Chart(ctx, {
                type: 'line',
                data: {
                    labels: timestamps,
                    datasets: [{
                        label: '调度延迟 (ms)',
                        data: latencies,
                        borderColor: 'rgb(75, 192, 192)',
                        backgroundColor: 'rgba(75, 192, 192, 0.2)',
                        tension: 0.1
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: '延迟 (ms)'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: '时间'
                            }
                        }
                    }
                }
            });
        }
        
        // 渲染吞吐量和成功率图
        function renderThroughputChart() {
            const ctx = document.getElementById('throughputChart').getContext('2d');
            const timestamps = analysisData.metrics.map(m => new Date(m.timestamp).toLocaleTimeString());
            const throughput = analysisData.metrics.map(m => m.throughput_pods_per_sec);
            const successRate = analysisData.metrics.map(m => m.success_rate_percent);
            
            new Chart(ctx, {
                type: 'line',
                data: {
                    labels: timestamps,
                    datasets: [{
                        label: '吞吐量 (pods/sec)',
                        data: throughput,
                        borderColor: 'rgb(255, 99, 132)',
                        backgroundColor: 'rgba(255, 99, 132, 0.2)',
                        yAxisID: 'y'
                    }, {
                        label: '成功率 (%)',
                        data: successRate,
                        borderColor: 'rgb(54, 162, 235)',
                        backgroundColor: 'rgba(54, 162, 235, 0.2)',
                        yAxisID: 'y1'
                    }]
                },
                options: {
                    responsive: true,
                    interaction: {
                        mode: 'index',
                        intersect: false,
                    },
                    scales: {
                        x: {
                            display: true,
                            title: {
                                display: true,
                                text: '时间'
                            }
                        },
                        y: {
                            type: 'linear',
                            display: true,
                            position: 'left',
                            title: {
                                display: true,
                                text: '吞吐量 (pods/sec)'
                            }
                        },
                        y1: {
                            type: 'linear',
                            display: true,
                            position: 'right',
                            title: {
                                display: true,
                                text: '成功率 (%)'
                            },
                            grid: {
                                drawOnChartArea: false,
                            },
                        }
                    }
                }
            });
        }
        
        function refreshData() {
            window.location.reload();
        }
        
        // 初始化图表
        renderLatencyChart();
        renderThroughputChart();
        
        // 每60秒自动刷新
        setInterval(refreshData, 60000);
    </script>
</body>
</html>
`

	t := template.Must(template.New("analysis").Parse(htmlTemplate))
	t.Execute(w, analysis)
}

// main 主函数
func main() {
	// 解析命令行参数
	var (
		port = flag.String("port", "8081", "HTTP server port")
	)
	flag.Parse()

	k8slog.Info("Starting Kubernetes Scheduler Performance Analyzer...")

	// 创建性能分析器
	analyzer := NewPerformanceAnalyzer()

	// 收集初始数据
	analyzer.CollectMetrics(2*time.Hour, 30*time.Second)

	// 启动数据收集协程
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			analyzer.CollectMetrics(30*time.Second, 30*time.Second)
		}
	}()

	// 设置HTTP路由
	http.HandleFunc("/", analyzer.ServeAnalysis)
	http.HandleFunc("/analysis", analyzer.ServeAnalysis)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// 启动HTTP服务器
	addr := ":" + *port
	k8slog.Infof("Performance analyzer server starting on %s", addr)
	k8slog.Infof("Access the dashboard at: http://localhost%s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		k8slog.Fatalf("Failed to start server: %v", err)
	}
}
