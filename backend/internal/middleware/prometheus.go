package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics collectors
var (
	// HTTP metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests in flight",
		},
		[]string{"method"},
	)

	// WebSocket metrics
	wsConnectionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "websocket_connections_total",
			Help: "Total number of WebSocket connections",
		},
	)

	wsConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Current number of active WebSocket connections",
		},
	)

	wsMessagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_received_total",
			Help: "Total number of WebSocket messages received",
		},
		[]string{"type"},
	)

	wsMessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_sent_total",
			Help: "Total number of WebSocket messages sent",
		},
		[]string{"type"},
	)

	// Business metrics
	devicesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "devices_total",
			Help: "Total number of devices",
		},
		[]string{"tenant_id", "type"},
	)

	devicesOnline = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "devices_online",
			Help: "Number of online devices",
		},
		[]string{"tenant_id", "type"},
	)

	telemetryReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telemetry_received_total",
			Help: "Total number of telemetry data points received",
		},
		[]string{"tenant_id", "device_type"},
	)

	alertsTriggered = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alerts_triggered_total",
			Help: "Total number of alerts triggered",
		},
		[]string{"tenant_id", "severity"},
	)

	alertsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "alerts_active",
			Help: "Current number of active alerts",
		},
		[]string{"tenant_id", "severity"},
	)

	// AI Agent metrics
	aiQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_queries_total",
			Help: "Total number of AI queries",
		},
		[]string{"tenant_id", "model"},
	)

	aiQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_query_duration_seconds",
			Help:    "AI query duration in seconds",
			Buckets: []float64{1, 2, 5, 10, 15, 20, 30, 45, 60, 120},
		},
		[]string{"tenant_id", "model"},
	)

	aiTokensUsed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_tokens_used_total",
			Help: "Total number of AI tokens used",
		},
		[]string{"tenant_id", "model", "type"}, // type: input/output
	)

	// Database metrics
	dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation", "table"},
	)

	dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Current number of active database connections",
		},
	)

	// Redis metrics
	redisCommandsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_commands_total",
			Help: "Total number of Redis commands",
		},
		[]string{"command"},
	)

	redisCacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	redisCacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

// Initialize Prometheus metrics
func InitPrometheus() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
		wsConnectionsTotal,
		wsConnectionsActive,
		wsMessagesReceived,
		wsMessagesSent,
		devicesTotal,
		devicesOnline,
		telemetryReceived,
		alertsTriggered,
		alertsActive,
		aiQueriesTotal,
		aiQueryDuration,
		aiTokensUsed,
		dbQueriesTotal,
		dbQueryDuration,
		dbConnectionsActive,
		redisCommandsTotal,
		redisCacheHits,
		redisCacheMisses,
	)
}

// PrometheusMiddleware records HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method
		path := c.FullPath()

		// Track in-flight requests
		httpRequestsInFlight.WithLabelValues(method).Inc()

		// Process request
		c.Next()

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsInFlight.WithLabelValues(method).Dec()
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// SetupPrometheusEndpoint registers /metrics endpoint
func SetupPrometheusEndpoint(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// Metric recording helper functions

func RecordWSConnection() {
	wsConnectionsTotal.Inc()
	wsConnectionsActive.Inc()
}

func RecordWSDisconnection() {
	wsConnectionsActive.Dec()
}

func RecordWSMessageReceived(msgType string) {
	wsMessagesReceived.WithLabelValues(msgType).Inc()
}

func RecordWSMessageSent(msgType string) {
	wsMessagesSent.WithLabelValues(msgType).Inc()
}

func UpdateDeviceMetrics(tenantID, deviceType string, total, online int) {
	devicesTotal.WithLabelValues(tenantID, deviceType).Set(float64(total))
	devicesOnline.WithLabelValues(tenantID, deviceType).Set(float64(online))
}

func RecordTelemetryReceived(tenantID, deviceType string) {
	telemetryReceived.WithLabelValues(tenantID, deviceType).Inc()
}

func RecordAlertTriggered(tenantID, severity string) {
	alertsTriggered.WithLabelValues(tenantID, severity).Inc()
}

func UpdateActiveAlerts(tenantID, severity string, count int) {
	alertsActive.WithLabelValues(tenantID, severity).Set(float64(count))
}

func RecordAIQuery(tenantID, model string, duration float64, inputTokens, outputTokens int) {
	aiQueriesTotal.WithLabelValues(tenantID, model).Inc()
	aiQueryDuration.WithLabelValues(tenantID, model).Observe(duration)
	aiTokensUsed.WithLabelValues(tenantID, model, "input").Add(float64(inputTokens))
	aiTokensUsed.WithLabelValues(tenantID, model, "output").Add(float64(outputTokens))
}

func RecordDBQuery(operation, table string, duration float64) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

func UpdateDBConnections(count int) {
	dbConnectionsActive.Set(float64(count))
}

func RecordRedisCommand(command string) {
	redisCommandsTotal.WithLabelValues(command).Inc()
}

func RecordCacheHit() {
	redisCacheHits.Inc()
}

func RecordCacheMiss() {
	redisCacheMisses.Inc()
}
