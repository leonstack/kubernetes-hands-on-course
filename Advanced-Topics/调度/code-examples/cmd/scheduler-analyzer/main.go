package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kubernetes-fundamentals/pkg/scheduler"
	"k8s.io/klog/v2"
)

func main() {
	// 解析命令行参数
	var (
		port       = flag.String("port", "8081", "HTTP server port")
		outputFile = flag.String("output", "", "Output file for analysis report (optional)")
		format     = flag.String("format", "json", "Output format: json or text")
	)
	flag.Parse()

	// 初始化日志
	klog.InitFlags(nil)

	klog.Info("Starting Kubernetes Scheduler Analyzer...")

	// 创建调度器分析器
	analyzer, err := scheduler.NewSchedulerAnalyzerWithConfig()
	if err != nil {
		klog.Fatalf("Failed to create scheduler analyzer: %v", err)
	}

	// 如果指定了输出文件，生成报告并退出
	if *outputFile != "" {
		report, err := analyzer.GenerateAnalysisReport(context.Background())
		if err != nil {
			klog.Fatalf("Failed to generate analysis report: %v", err)
		}

		var output []byte

		if *format == "json" {
			var marshalErr error
			output, marshalErr = json.MarshalIndent(report, "", "  ")
			if marshalErr != nil {
				klog.Fatalf("Failed to marshal report: %v", marshalErr)
			}
		} else {
			output = []byte(formatReportAsText(report))
		}

		if err := os.WriteFile(*outputFile, output, 0644); err != nil {
			klog.Fatalf("Failed to write report to file: %v", err)
		}

		klog.Infof("Analysis report written to %s", *outputFile)
		return
	}

	// 设置 HTTP 路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		report, err := analyzer.GenerateAnalysisReport(context.Background())
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate report: %v", err), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(report)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, formatReportAsText(report))
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// 启动 HTTP 服务器
	addr := ":" + *port
	klog.Infof("Starting HTTP server on %s", addr)
	klog.Infof("Access the scheduler analysis at http://localhost%s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		klog.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func formatReportAsText(report *scheduler.AnalysisReport) string {
	var result string

	result += "Scheduler Analysis Report\n"
	result += fmt.Sprintf("Generated at: %s\n\n", report.Timestamp.Format(time.RFC3339))

	result += "Performance Metrics:\n"
	result += fmt.Sprintf("  Average Latency: %v\n", report.PerformanceMetrics.AverageSchedulingLatency)
	result += fmt.Sprintf("  P95 Latency: %v\n", report.PerformanceMetrics.P95SchedulingLatency)
	result += fmt.Sprintf("  P99 Latency: %v\n", report.PerformanceMetrics.P99SchedulingLatency)
	result += fmt.Sprintf("  Throughput: %.2f pods/sec\n", report.PerformanceMetrics.SchedulingThroughput)
	result += fmt.Sprintf("  Failure Rate: %.2f%%\n\n", report.PerformanceMetrics.FailureRate*100)

	result += "Resource Analysis:\n"
	for resource, utilization := range report.ResourceAnalysis.ClusterUtilization {
		result += fmt.Sprintf("  %s: %.2f%%\n", resource, utilization*100)
	}
	result += "\n"

	if len(report.Recommendations) > 0 {
		result += "Recommendations:\n"
		for i, rec := range report.Recommendations {
			result += fmt.Sprintf("  %d. %s\n", i+1, rec)
		}
	}

	return result
}
