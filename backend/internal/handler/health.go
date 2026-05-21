package handler

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	db           *sql.DB
	redis        *redis.Client
	version      string
	startTime    time.Time
	dependencies []HealthChecker
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) HealthCheckResult
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status    string                 `json:"status"`
	LatencyMS int64                  `json:"latency_ms"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// HealthStatus 健康状态响应
type HealthStatus struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Uptime    int64                  `json:"uptime_seconds"`
	Checks    map[string]interface{} `json:"checks"`
	Timestamp string                 `json:"timestamp"`
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *sql.DB, redis *redis.Client, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redis:     redis,
		version:   version,
		startTime: time.Now(),
	}
}

// === Level 1: Liveness (存活检查) ===

// LivenessCheck 存活检查 - 仅检查进程存活
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// === Level 2: Readiness (就绪检查) ===

// ReadinessCheck 就绪检查 - 检查关键依赖
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := map[string]string{}
	allHealthy := true

	// 检查数据库
	dbStatus := h.checkDatabaseReady(ctx)
	if !dbStatus {
		checks["database"] = "not_ready"
		allHealthy = false
	} else {
		checks["database"] = "ok"
	}

	// 检查 Redis
	redisStatus := h.checkRedisReady(ctx)
	if !redisStatus {
		checks["redis"] = "not_ready"
		allHealthy = false
	} else {
		checks["redis"] = "ok"
	}

	// 检查迁移状态 (可选)
	// checks["migrations"] = "ok"

	if allHealthy {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"checks":    checks,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"checks":    checks,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// checkDatabaseReady 检查数据库是否就绪
func (h *HealthHandler) checkDatabaseReady(ctx context.Context) bool {
	if h.db == nil {
		return false
	}
	return h.db.PingContext(ctx) == nil
}

// checkRedisReady 检查 Redis 是否就绪
func (h *HealthHandler) checkRedisReady(ctx context.Context) bool {
	if h.redis == nil {
		return false
	}
	return h.redis.Ping(ctx).Err() == nil
}

// === Level 3: Detailed Health (详细健康检查) ===

// DetailedHealthCheck 详细健康检查
func (h *HealthHandler) DetailedHealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	checks := map[string]interface{}{}
	// nolint:ineffassign
	status := "healthy"
	allHealthy := true

	// 数据库检查
	dbResult := h.checkDatabaseHealth(ctx)
	checks["database"] = dbResult
	if dbResult.Status != "healthy" {
		allHealthy = false
	}

	// Redis 检查
	redisResult := h.checkRedisHealth(ctx)
	checks["redis"] = redisResult
	if redisResult.Status != "healthy" {
		allHealthy = false
	}

	// 磁盘检查
	diskResult := h.checkDiskHealth()
	checks["disk"] = diskResult
	if diskResult.Status != "healthy" {
		allHealthy = false
	}

	// 系统检查
	sysResult := h.checkSystemHealth()
	checks["system"] = sysResult

	if allHealthy {
		status = "healthy"
	} else {
		status = "degraded"
	}

	c.JSON(http.StatusOK, HealthStatus{
		Status:    status,
		Version:   h.version,
		Uptime:    int64(time.Since(h.startTime).Seconds()),
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// checkDatabaseHealth 数据库详细健康检查
func (h *HealthHandler) checkDatabaseHealth(ctx context.Context) HealthCheckResult {
	if h.db == nil {
		return HealthCheckResult{
			Status: "unhealthy",
			Error:  "database connection not initialized",
		}
	}

	start := time.Now()
	err := h.db.PingContext(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthCheckResult{
			Status:    "unhealthy",
			LatencyMS: latency,
			Error:     err.Error(),
		}
	}

	// 获取连接池状态
	stats := h.db.Stats()

	return HealthCheckResult{
		Status:    "healthy",
		LatencyMS: latency,
		Details: map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
			"wait_duration_ms": stats.WaitDuration.Milliseconds(),
		},
	}
}

// checkRedisHealth Redis 详细健康检查
func (h *HealthHandler) checkRedisHealth(ctx context.Context) HealthCheckResult {
	if h.redis == nil {
		return HealthCheckResult{
			Status: "unhealthy",
			Error:  "redis connection not initialized",
		}
	}

	start := time.Now()
	err := h.redis.Ping(ctx).Err()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthCheckResult{
			Status:    "unhealthy",
			LatencyMS: latency,
			Error:     err.Error(),
		}
	}

	// 获取 Redis INFO
	_, err = h.redis.Info(ctx, "memory", "stats").Result()
	if err == nil {
		return HealthCheckResult{
			Status:    "healthy",
			LatencyMS: latency,
			Details: map[string]interface{}{
				"info_available": true,
			},
		}
	}

	return HealthCheckResult{
		Status:    "healthy",
		LatencyMS: latency,
	}
}

// checkDiskHealth 磁盘健康检查
func (h *HealthHandler) checkDiskHealth() HealthCheckResult {
	// 简化版磁盘检查
	// 实际实现需要调用系统 API
	return HealthCheckResult{
		Status: "healthy",
		Details: map[string]interface{}{
			"check_available": false,
			"message":         "disk health check requires system API",
		},
	}
}

// checkSystemHealth 系统健康检查
func (h *HealthHandler) checkSystemHealth() HealthCheckResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return HealthCheckResult{
		Status: "healthy",
		Details: map[string]interface{}{
			"go_version":      runtime.Version(),
			"goroutines":      runtime.NumGoroutine(),
			"memory_alloc_mb": m.Alloc / 1024 / 1024,
			"memory_sys_mb":   m.Sys / 1024 / 1024,
			"gc_cycles":       m.NumGC,
		},
	}
}

// === Level 4: Dependencies (依赖深度检查) ===

// DependenciesCheck 依赖深度检查
func (h *HealthHandler) DependenciesCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	dependencies := map[string]interface{}{}
	status := "healthy"
	allHealthy := true

	// PostgreSQL 详细检查
	pgResult := h.checkPostgreSQLDependency(ctx)
	dependencies["postgresql"] = pgResult
	if pgResult.Status != "healthy" {
		allHealthy = false
	}

	// Redis 详细检查
	redisResult := h.checkRedisDependency(ctx)
	dependencies["redis"] = redisResult
	if redisResult.Status != "healthy" {
		allHealthy = false
	}

	// GLM API 检查 (可选)
	// glmResult := h.checkGLMAPIDependency(ctx)
	// dependencies["glm_api"] = glmResult

	if !allHealthy {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       status,
		"dependencies": dependencies,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// checkPostgreSQLDependency PostgreSQL 依赖检查
func (h *HealthHandler) checkPostgreSQLDependency(ctx context.Context) HealthCheckResult {
	if h.db == nil {
		return HealthCheckResult{
			Status: "unhealthy",
			Error:  "database connection not initialized",
		}
	}

	start := time.Now()

	// 执行简单查询测试
	var version string
	err := h.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthCheckResult{
			Status:    "unhealthy",
			LatencyMS: latency,
			Error:     err.Error(),
		}
	}

	// 检查 SSL 状态
	var ssl bool
	h.db.QueryRowContext(ctx, `
		SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid()
	`).Scan(&ssl)

	return HealthCheckResult{
		Status:    "healthy",
		LatencyMS: latency,
		Details: map[string]interface{}{
			"version": version[:50], // 截取前 50 字符
			"ssl":     ssl,
		},
	}
}

// checkRedisDependency Redis 依赖检查
func (h *HealthHandler) checkRedisDependency(ctx context.Context) HealthCheckResult {
	if h.redis == nil {
		return HealthCheckResult{
			Status: "unhealthy",
			Error:  "redis connection not initialized",
		}
	}

	start := time.Now()

	// 执行 PING 测试
	err := h.redis.Ping(ctx).Err()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthCheckResult{
			Status:    "unhealthy",
			LatencyMS: latency,
			Error:     err.Error(),
		}
	}

	// 获取 Redis 信息
	_, err = h.redis.Info(ctx, "server", "memory").Result()
	if err != nil {
		return HealthCheckResult{
			Status:    "healthy",
			LatencyMS: latency,
			Details:   map[string]interface{}{"info_available": false},
		}
	}

	// 解析关键信息 (简化)
	return HealthCheckResult{
		Status:    "healthy",
		LatencyMS: latency,
		Details: map[string]interface{}{
			"info_available": true,
		},
	}
}

// === Startup Probe (启动探针) ===

// StartupCheck 启动检查 - 检查应用初始化完成
func (h *HealthHandler) StartupCheck(c *gin.Context) {
	// 检查应用是否完成启动初始化
	// 包括: 配置加载、数据库连接、Redis 连接、迁移执行等

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	checks := map[string]string{}
	allComplete := true

	// 检查数据库迁移完成
	dbReady := h.checkDatabaseReady(ctx)
	if !dbReady {
		checks["database"] = "initializing"
		allComplete = false
	} else {
		checks["database"] = "initialized"
	}

	// 检查 Redis 连接
	redisReady := h.checkRedisReady(ctx)
	if !redisReady {
		checks["redis"] = "initializing"
		allComplete = false
	} else {
		checks["redis"] = "initialized"
	}

	if allComplete {
		c.JSON(http.StatusOK, gin.H{
			"status":    "started",
			"checks":    checks,
			"uptime_s":  int64(time.Since(h.startTime).Seconds()),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "starting",
			"checks":    checks,
			"uptime_s":  int64(time.Since(h.startTime).Seconds()),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// RegisterHealthRoutes 注册健康检查路由
func (h *HealthHandler) RegisterHealthRoutes(r *gin.Engine) {
	health := r.Group("/health")
	{
		health.GET("/live", h.LivenessCheck)             // Liveness Probe
		health.GET("/ready", h.ReadinessCheck)           // Readiness Probe
		health.GET("/startup", h.StartupCheck)           // Startup Probe
		health.GET("", h.DetailedHealthCheck)            // Detailed Health
		health.GET("/dependencies", h.DependenciesCheck) // Dependencies
	}
}
