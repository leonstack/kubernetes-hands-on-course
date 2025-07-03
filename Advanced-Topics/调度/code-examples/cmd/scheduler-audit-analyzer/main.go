// scheduler-audit-analyzer.go
// Kubernetes 调度器安全审计分析器实现

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// AuditEvent 审计事件结构
type AuditEvent struct {
	Kind                     string            `json:"kind"`
	APIVersion               string            `json:"apiVersion"`
	Level                    string            `json:"level"`
	AuditID                  string            `json:"auditID"`
	Stage                    string            `json:"stage"`
	RequestURI               string            `json:"requestURI"`
	Verb                     string            `json:"verb"`
	User                     User              `json:"user"`
	SourceIPs                []string          `json:"sourceIPs"`
	UserAgent                string            `json:"userAgent"`
	ObjectRef                ObjectRef         `json:"objectRef"`
	RequestObject            interface{}       `json:"requestObject,omitempty"`
	ResponseObject           interface{}       `json:"responseObject,omitempty"`
	RequestReceivedTimestamp time.Time         `json:"requestReceivedTimestamp"`
	StageTimestamp           time.Time         `json:"stageTimestamp"`
	Annotations              map[string]string `json:"annotations,omitempty"`
}

// User 用户信息
type User struct {
	Username string              `json:"username"`
	UID      string              `json:"uid"`
	Groups   []string            `json:"groups"`
	Extra    map[string][]string `json:"extra,omitempty"`
}

// ObjectRef 对象引用
type ObjectRef struct {
	Resource        string `json:"resource"`
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	UID             string `json:"uid"`
	APIGroup        string `json:"apiGroup"`
	APIVersion      string `json:"apiVersion"`
	ResourceVersion string `json:"resourceVersion"`
	Subresource     string `json:"subresource"`
}

// SchedulingAuditAnalyzer 调度审计分析器
type SchedulingAuditAnalyzer struct {
	auditEvents      []AuditEvent
	allowedIPs       []string
	allowedAgents    []string
	suspiciousIPs    map[string]int
	violations       []SecurityViolation
	patterns         []SchedulingPattern
	violationCache   map[string]bool // 防止重复违规记录
	maxEvents        int             // 最大事件数限制
	suspiciousThreshold int          // 可疑IP阈值
}

// SecurityViolation 安全违规
type SecurityViolation struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	User        string                 `json:"user"`
	SourceIP    string                 `json:"source_ip"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SchedulingPattern 调度模式
type SchedulingPattern struct {
	Type        string                 `json:"type"`
	Count       int                    `json:"count"`
	Description string                 `json:"description"`
	TimeRange   TimeRange              `json:"time_range"`
	Details     map[string]interface{} `json:"details"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	TotalEvents      int                 `json:"total_events"`
	SchedulingEvents int                 `json:"scheduling_events"`
	BindingEvents    int                 `json:"binding_events"`
	Violations       []SecurityViolation `json:"violations"`
	Patterns         []SchedulingPattern `json:"patterns"`
	Summary          AnalysisSummary     `json:"summary"`
	GeneratedAt      time.Time           `json:"generated_at"`
}

// AnalysisSummary 分析摘要
type AnalysisSummary struct {
	HighSeverityViolations   int     `json:"high_severity_violations"`
	MediumSeverityViolations int     `json:"medium_severity_violations"`
	LowSeverityViolations    int     `json:"low_severity_violations"`
	SuccessfulBindings       int     `json:"successful_bindings"`
	FailedSchedulings        int     `json:"failed_schedulings"`
	AverageSchedulingLatency float64 `json:"average_scheduling_latency_ms"`
	SuspiciousActivities     int     `json:"suspicious_activities"`
}

// NewSchedulingAuditAnalyzer 创建调度审计分析器
func NewSchedulingAuditAnalyzer() *SchedulingAuditAnalyzer {
	return &SchedulingAuditAnalyzer{
		auditEvents:         make([]AuditEvent, 0),
		allowedIPs:          []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		allowedAgents:       []string{"kube-scheduler", "kubectl", "kube-apiserver"},
		suspiciousIPs:       make(map[string]int),
		violations:          make([]SecurityViolation, 0),
		patterns:            make([]SchedulingPattern, 0),
		violationCache:      make(map[string]bool),
		maxEvents:           10000, // 默认最大事件数
		suspiciousThreshold: 10,    // 默认可疑IP阈值
	}
}

// LoadAuditLog 加载审计日志
func (saa *SchedulingAuditAnalyzer) LoadAuditLog(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %v", err)
	}
	defer file.Close()

	// 检查文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	klog.V(2).Infof("Loading audit log file: %s (size: %d bytes)", filename, fileInfo.Size())

	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小以处理大行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max token size

	lineCount := 0
	processedCount := 0
	skippedCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 检查是否超过最大事件数限制
		if processedCount >= saa.maxEvents {
			klog.Warningf("Reached maximum events limit (%d), stopping processing", saa.maxEvents)
			break
		}

		var event AuditEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			klog.V(3).Infof("Failed to parse audit event at line %d: %v", lineCount, err)
			skippedCount++
			continue
		}

		// 验证事件的基本字段
		if event.AuditID == "" || event.RequestReceivedTimestamp.IsZero() {
			klog.V(4).Infof("Skipping invalid event at line %d: missing required fields", lineCount)
			skippedCount++
			continue
		}

		// 只处理调度相关的事件
		if saa.isSchedulingRelated(event) {
			saa.auditEvents = append(saa.auditEvents, event)
			processedCount++
			
			// 定期输出进度
			if processedCount%1000 == 0 {
				klog.V(2).Infof("Processed %d scheduling events...", processedCount)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading audit log: %v", err)
	}

	klog.Infof("Loaded %d scheduling-related events from %d total lines (skipped: %d)", processedCount, lineCount, skippedCount)
	return nil
}

// isSchedulingRelated 判断是否为调度相关事件
func (saa *SchedulingAuditAnalyzer) isSchedulingRelated(event AuditEvent) bool {
	// 检查用户代理
	if strings.Contains(strings.ToLower(event.UserAgent), "scheduler") {
		return true
	}

	// 检查请求URI
	schedulingURIs := []string{
		"/api/v1/pods",
		"/api/v1/nodes",
		"/api/v1/bindings",
		"/apis/scheduling.k8s.io",
		"/apis/metrics.k8s.io",
	}

	for _, uri := range schedulingURIs {
		if strings.Contains(event.RequestURI, uri) {
			return true
		}
	}

	// 检查对象类型
	if event.ObjectRef.Resource == "pods" || event.ObjectRef.Resource == "nodes" ||
		event.ObjectRef.Resource == "bindings" {
		return true
	}

	// 检查注解中的调度信息
	for key := range event.Annotations {
		if strings.Contains(strings.ToLower(key), "schedul") {
			return true
		}
	}

	return false
}

// Analyze 分析审计事件
func (saa *SchedulingAuditAnalyzer) Analyze() *AnalysisResult {
	klog.Info("Starting scheduling audit analysis...")

	// 重置分析结果
	saa.violations = make([]SecurityViolation, 0)
	saa.patterns = make([]SchedulingPattern, 0)
	saa.suspiciousIPs = make(map[string]int)

	bindingCount := 0
	schedulingLatencies := make([]float64, 0)

	for _, event := range saa.auditEvents {
		// 统计绑定事件
		if event.ObjectRef.Resource == "bindings" && event.Verb == "create" {
			bindingCount++
		}

		// 计算调度延迟
		if !event.RequestReceivedTimestamp.IsZero() && !event.StageTimestamp.IsZero() {
			latency := event.StageTimestamp.Sub(event.RequestReceivedTimestamp).Seconds() * 1000
			schedulingLatencies = append(schedulingLatencies, latency)
		}

		// 检测安全违规
		saa.detectSecurityViolations(event)

		// 提取调度模式
		saa.extractSchedulingPattern(event)
	}

	// 计算平均调度延迟
	var avgLatency float64
	if len(schedulingLatencies) > 0 {
		var total float64
		for _, latency := range schedulingLatencies {
			total += latency
		}
		avgLatency = total / float64(len(schedulingLatencies))
	}

	// 统计违规严重程度
	highSeverity, mediumSeverity, lowSeverity := saa.countViolationsBySeverity()

	// 计算失败的调度事件
	failedSchedulings := saa.countFailedSchedulings()

	result := &AnalysisResult{
		TotalEvents:      len(saa.auditEvents),
		SchedulingEvents: len(saa.auditEvents),
		BindingEvents:    bindingCount,
		Violations:       saa.violations,
		Patterns:         saa.patterns,
		Summary: AnalysisSummary{
			HighSeverityViolations:   highSeverity,
			MediumSeverityViolations: mediumSeverity,
			LowSeverityViolations:    lowSeverity,
			SuccessfulBindings:       bindingCount,
			FailedSchedulings:        failedSchedulings,
			AverageSchedulingLatency: avgLatency,
			SuspiciousActivities:     len(saa.suspiciousIPs),
		},
		GeneratedAt: time.Now(),
	}

	klog.Infof("Analysis completed: %d events, %d violations, %d patterns",
		result.TotalEvents, len(result.Violations), len(result.Patterns))

	return result
}

// detectSecurityViolations 检测安全违规
func (saa *SchedulingAuditAnalyzer) detectSecurityViolations(event AuditEvent) {
	// 检查未经授权的调度器访问
	if strings.Contains(event.UserAgent, "scheduler") && !saa.isAuthorizedScheduler(event.User) {
		violationKey := fmt.Sprintf("unauthorized_scheduler_%s_%s", event.User.Username, strings.Join(event.SourceIPs, ","))
		if !saa.violationCache[violationKey] {
			violation := SecurityViolation{
				Type:        "unauthorized_scheduler_access",
				Severity:    "high",
				Description: "Unauthorized scheduler access detected",
				User:        event.User.Username,
				SourceIP:    strings.Join(event.SourceIPs, ","),
				Timestamp:   event.RequestReceivedTimestamp,
				Details: map[string]interface{}{
					"user_agent":  event.UserAgent,
					"request_uri": event.RequestURI,
					"verb":        event.Verb,
				},
			}
			saa.violations = append(saa.violations, violation)
			saa.violationCache[violationKey] = true
			klog.V(2).Infof("Detected unauthorized scheduler access: %s from %s", event.User.Username, strings.Join(event.SourceIPs, ","))
		}
	}

	// 检查异常源IP
	for _, ip := range event.SourceIPs {
		if ip == "" {
			continue // 跳过空IP
		}
		
		if !saa.isAllowedSourceIP(ip) {
			saa.suspiciousIPs[ip]++
			if saa.suspiciousIPs[ip] > saa.suspiciousThreshold {
				violationKey := fmt.Sprintf("suspicious_ip_%s", ip)
				if !saa.violationCache[violationKey] {
					violation := SecurityViolation{
						Type:        "suspicious_source_ip",
						Severity:    "medium",
						Description: fmt.Sprintf("Suspicious source IP detected: %s (count: %d)", ip, saa.suspiciousIPs[ip]),
						User:        event.User.Username,
						SourceIP:    ip,
						Timestamp:   event.RequestReceivedTimestamp,
						Details: map[string]interface{}{
							"access_count": saa.suspiciousIPs[ip],
							"user_agent":   event.UserAgent,
							"threshold":    saa.suspiciousThreshold,
						},
					}
					saa.violations = append(saa.violations, violation)
					saa.violationCache[violationKey] = true
					klog.V(2).Infof("Detected suspicious IP: %s (count: %d)", ip, saa.suspiciousIPs[ip])
				}
			}
		}
	}

	// 检查无效的用户代理
	if !saa.isValidUserAgent(event.UserAgent) {
		violationKey := fmt.Sprintf("invalid_agent_%s_%s", event.User.Username, event.UserAgent)
		if !saa.violationCache[violationKey] {
			violation := SecurityViolation{
				Type:        "invalid_user_agent",
				Severity:    "low",
				Description: "Invalid or suspicious user agent detected",
				User:        event.User.Username,
				SourceIP:    strings.Join(event.SourceIPs, ","),
				Timestamp:   event.RequestReceivedTimestamp,
				Details: map[string]interface{}{
					"user_agent":      event.UserAgent,
					"expected_agents": saa.allowedAgents,
				},
			}
			saa.violations = append(saa.violations, violation)
			saa.violationCache[violationKey] = true
			klog.V(3).Infof("Detected invalid user agent: %s from user %s", event.UserAgent, event.User.Username)
		}
	}
}

// extractSchedulingPattern 提取调度模式
func (saa *SchedulingAuditAnalyzer) extractSchedulingPattern(event AuditEvent) {
	// Pod 绑定模式
	if event.ObjectRef.Resource == "bindings" && event.Verb == "create" {
		pattern := SchedulingPattern{
			Type:        "pod_binding",
			Count:       1,
			Description: saa.getPatternDescription("pod_binding"),
			TimeRange: TimeRange{
				Start: event.RequestReceivedTimestamp,
				End:   event.StageTimestamp,
			},
			Details: map[string]interface{}{
				"pod_name":      event.ObjectRef.Name,
				"pod_namespace": event.ObjectRef.Namespace,
				"scheduler":     event.User.Username,
			},
		}
		saa.patterns = append(saa.patterns, pattern)
	}

	// 调度失败模式
	if event.ObjectRef.Resource == "pods" && event.Verb == "update" {
		if event.Annotations != nil {
			if reason, exists := event.Annotations["scheduler.alpha.kubernetes.io/failed-scheduling"]; exists {
				pattern := SchedulingPattern{
					Type:        "scheduling_failure",
					Count:       1,
					Description: saa.getPatternDescription("scheduling_failure"),
					TimeRange: TimeRange{
						Start: event.RequestReceivedTimestamp,
						End:   event.StageTimestamp,
					},
					Details: map[string]interface{}{
						"pod_name":       event.ObjectRef.Name,
						"pod_namespace":  event.ObjectRef.Namespace,
						"failure_reason": reason,
					},
				}
				saa.patterns = append(saa.patterns, pattern)
			}
		}
	}

	// Pod 抢占模式
	if event.ObjectRef.Resource == "pods" && event.Verb == "delete" {
		if event.Annotations != nil {
			if _, exists := event.Annotations["scheduler.alpha.kubernetes.io/preempted-by"]; exists {
				pattern := SchedulingPattern{
					Type:        "pod_preemption",
					Count:       1,
					Description: saa.getPatternDescription("pod_preemption"),
					TimeRange: TimeRange{
						Start: event.RequestReceivedTimestamp,
						End:   event.StageTimestamp,
					},
					Details: map[string]interface{}{
						"preempted_pod": event.ObjectRef.Name,
						"namespace":     event.ObjectRef.Namespace,
					},
				}
				saa.patterns = append(saa.patterns, pattern)
			}
		}
	}
}

// getPatternDescription 获取模式描述
func (saa *SchedulingAuditAnalyzer) getPatternDescription(patternType string) string {
	descriptions := map[string]string{
		"pod_binding":        "Pod successfully bound to a node",
		"scheduling_failure": "Pod failed to be scheduled",
		"pod_preemption":     "Pod was preempted by higher priority pod",
	}
	return descriptions[patternType]
}

// isAllowedSourceIP 检查源IP是否被允许
func (saa *SchedulingAuditAnalyzer) isAllowedSourceIP(ip string) bool {
	// 输入验证
	if ip == "" {
		return false
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		klog.V(4).Infof("Invalid IP format: %s", ip)
		return false
	}

	for _, allowedIP := range saa.allowedIPs {
		if allowedIP == "" {
			continue // 跳过空的允许IP
		}
		
		if strings.Contains(allowedIP, "/") {
			// CIDR 格式
			_, cidr, err := net.ParseCIDR(allowedIP)
			if err != nil {
				klog.V(3).Infof("Invalid CIDR format in allowed IPs: %s, error: %v", allowedIP, err)
				continue
			}
			if cidr.Contains(parsedIP) {
				klog.V(5).Infof("IP %s allowed by CIDR %s", ip, allowedIP)
				return true
			}
		} else {
			// 单个IP - 标准化比较
			allowedParsed := net.ParseIP(allowedIP)
			if allowedParsed != nil && allowedParsed.Equal(parsedIP) {
				klog.V(5).Infof("IP %s allowed by exact match", ip)
				return true
			}
		}
	}
	
	klog.V(4).Infof("IP %s not in allowed list", ip)
	return false
}

// isValidUserAgent 检查用户代理是否有效
func (saa *SchedulingAuditAnalyzer) isValidUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	// 检查是否包含已知的合法代理
	for _, agent := range saa.allowedAgents {
		if strings.Contains(strings.ToLower(userAgent), strings.ToLower(agent)) {
			return true
		}
	}

	// 检查是否符合 Kubernetes 组件的命名模式
	kubernetesPattern := regexp.MustCompile(`^(kube-|kubectl|kubernetes)`)
	return kubernetesPattern.MatchString(strings.ToLower(userAgent))
}

// isAuthorizedScheduler 检查是否为授权的调度器
func (saa *SchedulingAuditAnalyzer) isAuthorizedScheduler(user User) bool {
	// 检查用户名
	authorizedUsers := []string{"system:kube-scheduler", "system:serviceaccount:kube-system:default"}
	for _, authUser := range authorizedUsers {
		if user.Username == authUser {
			return true
		}
	}

	// 检查用户组
	for _, group := range user.Groups {
		if group == "system:authenticated" || group == "system:serviceaccounts" {
			return true
		}
	}

	return false
}

// countViolationsBySeverity 按严重程度统计违规
func (saa *SchedulingAuditAnalyzer) countViolationsBySeverity() (int, int, int) {
	var high, medium, low int
	for _, violation := range saa.violations {
		switch violation.Severity {
		case "high":
			high++
		case "medium":
			medium++
		case "low":
			low++
		}
	}
	return high, medium, low
}

// countFailedSchedulings 统计失败的调度
func (saa *SchedulingAuditAnalyzer) countFailedSchedulings() int {
	count := 0
	for _, pattern := range saa.patterns {
		if pattern.Type == "scheduling_failure" {
			count++
		}
	}
	return count
}

func main() {
	// 初始化klog flags first
	klog.InitFlags(nil)
	
	// 定义命令行参数
	var (
		maxEvents = flag.Int("max-events", 10000, "Maximum number of events to process")
		suspiciousThreshold = flag.Int("suspicious-threshold", 10, "Threshold for suspicious IP detection")
		outputFormat = flag.String("output", "json", "Output format: json or summary")
		help = flag.Bool("help", false, "Show help message")
	)
	
	// 自定义usage函数
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Kubernetes Scheduler Audit Analyzer\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <audit-log-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Description:\n")
		fmt.Fprintf(os.Stderr, "  Analyzes Kubernetes audit logs for scheduling-related security events\n")
		fmt.Fprintf(os.Stderr, "  and compliance violations.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <audit-log-file>    Path to the Kubernetes audit log file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nFeatures:\n")
		fmt.Fprintf(os.Stderr, "  - Scheduling event extraction and analysis\n")
		fmt.Fprintf(os.Stderr, "  - Security violation detection with configurable thresholds\n")
		fmt.Fprintf(os.Stderr, "  - Scheduling pattern identification\n")
		fmt.Fprintf(os.Stderr, "  - Memory-efficient processing with event limits\n")
		fmt.Fprintf(os.Stderr, "  - Duplicate violation detection prevention\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -max-events=5000 -suspicious-threshold=15 -v=3 /var/log/audit/audit.log\n\n", os.Args[0])
	}
	
	flag.Parse()
	
	// 检查帮助参数
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	
	// 检查命令行参数
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: audit log file is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	auditLogFile := flag.Arg(0)
	
	// 验证参数
	if *maxEvents <= 0 {
		fmt.Fprintf(os.Stderr, "Error: max-events must be positive\n")
		os.Exit(1)
	}
	if *suspiciousThreshold <= 0 {
		fmt.Fprintf(os.Stderr, "Error: suspicious-threshold must be positive\n")
		os.Exit(1)
	}
	if *outputFormat != "json" && *outputFormat != "summary" {
		fmt.Fprintf(os.Stderr, "Error: output format must be 'json' or 'summary'\n")
		os.Exit(1)
	}
	
	// 初始化日志
	klog.Info("Starting Kubernetes Scheduler Audit Analyzer...")
	klog.V(1).Infof("Configuration: max-events=%d, suspicious-threshold=%d, output=%s", 
		*maxEvents, *suspiciousThreshold, *outputFormat)
	
	// 创建审计分析器
	analyzer := NewSchedulingAuditAnalyzer()
	analyzer.maxEvents = *maxEvents
	analyzer.suspiciousThreshold = *suspiciousThreshold
	
	// 加载审计日志
	if err := analyzer.LoadAuditLog(auditLogFile); err != nil {
		klog.Fatalf("Failed to load audit log: %v", err)
	}
	
	// 执行分析
	result := analyzer.Analyze()
	
	// 输出结果
	if *outputFormat == "json" {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			klog.Fatalf("Failed to marshal analysis result: %v", err)
		}
		fmt.Println(string(output))
	}
	
	// 输出摘要
	fmt.Printf("\n=== Analysis Summary ===\n")
	fmt.Printf("Total Events: %d\n", result.TotalEvents)
	fmt.Printf("Scheduling Events: %d\n", result.SchedulingEvents)
	fmt.Printf("Security Violations: %d (High: %d, Medium: %d, Low: %d)\n",
		len(result.Violations),
		result.Summary.HighSeverityViolations,
		result.Summary.MediumSeverityViolations,
		result.Summary.LowSeverityViolations)
	fmt.Printf("Scheduling Patterns: %d\n", len(result.Patterns))
	fmt.Printf("Failed Schedulings: %d\n", result.Summary.FailedSchedulings)
	fmt.Printf("Successful Bindings: %d\n", result.Summary.SuccessfulBindings)
	fmt.Printf("Average Scheduling Latency: %.2f ms\n", result.Summary.AverageSchedulingLatency)
	fmt.Printf("Suspicious Activities: %d\n", result.Summary.SuspiciousActivities)
	
	if result.Summary.HighSeverityViolations > 0 {
		fmt.Printf("\n⚠️  WARNING: %d high severity security violations detected!\n", result.Summary.HighSeverityViolations)
		fmt.Printf("Please review the violations and take appropriate action.\n")
		os.Exit(1)
	}
	
	klog.Info("Analysis completed successfully")
}
