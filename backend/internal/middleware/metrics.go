package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsService handles Prometheus metrics collection
type MetricsService struct {
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	activeConnections   prometheus.Gauge
	userLogins          *prometheus.CounterVec
	questionsAnswered   *prometheus.CounterVec
	achievementsEarned  *prometheus.CounterVec
	levelsCompleted     *prometheus.CounterVec
	errorOccurred       *prometheus.CounterVec
	dbConnections       prometheus.Gauge
	memoryUsage         prometheus.Gauge
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	return &MetricsService{
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "paperplay_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "paperplay_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		activeConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "paperplay_active_connections",
				Help: "Number of active connections",
			},
		),
		userLogins: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "paperplay_user_logins_total",
				Help: "Total number of user login attempts",
			},
			[]string{"status"}, // success, failed
		),
		questionsAnswered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "paperplay_questions_answered_total",
				Help: "Total number of questions answered",
			},
			[]string{"subject", "correct"}, // correct: true/false
		),
		achievementsEarned: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "paperplay_achievements_earned_total",
				Help: "Total number of achievements earned",
			},
			[]string{"achievement_type", "level"},
		),
		levelsCompleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "paperplay_levels_completed_total",
				Help: "Total number of levels completed",
			},
			[]string{"subject"},
		),
		errorOccurred: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "paperplay_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "component"},
		),
		dbConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "paperplay_db_connections",
				Help: "Number of database connections",
			},
		),
		memoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "paperplay_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
		),
	}
}

// HTTPMetrics creates a Gin middleware for HTTP metrics collection
func (m *MetricsService) HTTPMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment active connections
		m.activeConnections.Inc()
		defer m.activeConnections.Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()

		// If endpoint is empty (404), use path
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Record request count
		m.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()

		// Record request duration
		m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

// PrometheusHandler returns a handler for Prometheus metrics endpoint
func (m *MetricsService) PrometheusHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordUserLogin records user login metrics
func (m *MetricsService) RecordUserLogin(success bool) {
	status := "failed"
	if success {
		status = "success"
	}
	m.userLogins.WithLabelValues(status).Inc()
}

// RecordQuestionAnswered records question answering metrics
func (m *MetricsService) RecordQuestionAnswered(subject string, correct bool) {
	correctStr := "false"
	if correct {
		correctStr = "true"
	}
	m.questionsAnswered.WithLabelValues(subject, correctStr).Inc()
}

// RecordAchievementEarned records achievement earning metrics
func (m *MetricsService) RecordAchievementEarned(achievementType string, level int) {
	levelStr := strconv.Itoa(level)
	m.achievementsEarned.WithLabelValues(achievementType, levelStr).Inc()
}

// RecordLevelCompleted records level completion metrics
func (m *MetricsService) RecordLevelCompleted(subject string) {
	m.levelsCompleted.WithLabelValues(subject).Inc()
}

// RecordError records error metrics
func (m *MetricsService) RecordError(errorType, component string) {
	m.errorOccurred.WithLabelValues(errorType, component).Inc()
}

// SetDBConnections sets the current database connections count
func (m *MetricsService) SetDBConnections(count int) {
	m.dbConnections.Set(float64(count))
}

// SetMemoryUsage sets the current memory usage
func (m *MetricsService) SetMemoryUsage(bytes int64) {
	m.memoryUsage.Set(float64(bytes))
}

// GetMetrics returns all registered metrics
func (m *MetricsService) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Get HTTP request metrics
	httpMetrics := make(map[string]float64)
	metricFamilies, _ := prometheus.DefaultGatherer.Gather()

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "paperplay_http_requests_total":
			for _, metric := range mf.Metric {
				labels := make(map[string]string)
				for _, label := range metric.Label {
					labels[*label.Name] = *label.Value
				}
				key := labels["method"] + "_" + labels["endpoint"] + "_" + labels["status"]
				httpMetrics[key] = *metric.Counter.Value
			}
		}
	}

	metrics["http_requests"] = httpMetrics
	return metrics
}

// HealthCheck performs a health check for metrics service
func (m *MetricsService) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"status":          "healthy",
		"metrics_enabled": true,
		"registry_status": "active",
	}
}
