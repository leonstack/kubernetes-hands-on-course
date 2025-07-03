// Package metrics provides common metrics collection functionality
package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

// MetricsCollector provides common metrics functionality
type MetricsCollector struct {
	registry *prometheus.Registry
	// Common metrics
	RequestDuration *prometheus.HistogramVec
	RequestTotal    *prometheus.CounterVec
	ErrorTotal      *prometheus.CounterVec
	HealthStatus    *prometheus.GaugeVec
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(namespace, subsystem string) *MetricsCollector {
	registry := prometheus.NewRegistry()
	
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Duration of requests in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)
	
	requestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	errorTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total number of errors",
		},
		[]string{"type", "component"},
	)
	
	healthStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "health_status",
			Help:      "Health status of the component (1=healthy, 0=unhealthy)",
		},
		[]string{"component"},
	)
	
	registry.MustRegister(requestDuration, requestTotal, errorTotal, healthStatus)
	
	return &MetricsCollector{
		registry:        registry,
		RequestDuration: requestDuration,
		RequestTotal:    requestTotal,
		ErrorTotal:      errorTotal,
		HealthStatus:    healthStatus,
	}
}

// RecordRequest records a request with duration and status
func (mc *MetricsCollector) RecordRequest(method, endpoint, status string, duration time.Duration) {
	mc.RequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	mc.RequestTotal.WithLabelValues(method, endpoint, status).Inc()
}

// RecordError records an error
func (mc *MetricsCollector) RecordError(errorType, component string) {
	mc.ErrorTotal.WithLabelValues(errorType, component).Inc()
}

// SetHealthStatus sets the health status for a component
func (mc *MetricsCollector) SetHealthStatus(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	mc.HealthStatus.WithLabelValues(component).Set(value)
}

// Handler returns the HTTP handler for metrics
func (mc *MetricsCollector) Handler() http.Handler {
	return promhttp.HandlerFor(mc.registry, promhttp.HandlerOpts{})
}

// StartMetricsServer starts the metrics HTTP server
func (mc *MetricsCollector) StartMetricsServer(port int) {
	http.Handle("/metrics", mc.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	addr := fmt.Sprintf(":%d", port)
	klog.Infof("Starting metrics server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		klog.Errorf("Failed to start metrics server: %v", err)
	}
}