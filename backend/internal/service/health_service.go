package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

// ComponentHealth represents health status of all components
type ComponentHealth struct {
	Database HealthStatus `json:"database"`
	LLMAPI   HealthStatus `json:"llm_api"`
	Memory   HealthStatus `json:"memory"`
	Disk     HealthStatus `json:"disk"`
}

// HealthCheckResponse represents the complete health check response
type HealthCheckResponse struct {
	Status     string          `json:"status"`
	Components ComponentHealth `json:"components"`
	Version    string          `json:"version"`
	Uptime     string          `json:"uptime"`
	Timestamp  time.Time       `json:"timestamp"`
}

// HealthService provides health checking capabilities
type HealthService struct {
	db         *sql.DB
	httpClient HTTPClientInterface
	startTime  time.Time
	version    string
	config     HealthServiceConfig
}

// InitHealthService initializes the health service instance
func InitHealthService(db *sql.DB, version string, config HealthServiceConfig) *HealthService {
	return &HealthService{
		db:         db,
		startTime:  time.Now(),
		version:    version,
		httpClient: &http.Client{Timeout: config.CheckTimeout},
		config:     config,
	}
}

// InitHealthServiceWithClient initializes health service with custom HTTP client
func InitHealthServiceWithClient(db *sql.DB, version string, config HealthServiceConfig, client HTTPClientInterface) *HealthService {
	return &HealthService{
		db:         db,
		startTime:  time.Now(),
		version:    version,
		httpClient: client,
		config:     config,
	}
}

// CheckHealth performs comprehensive health checks on all components
func (s *HealthService) CheckHealth(ctx context.Context) *HealthCheckResponse {
	// Use a shorter timeout for the overall health check
	checkCtx, cancel := context.WithTimeout(ctx, s.config.CheckTimeout)
	defer cancel()

	// Perform checks concurrently using channels
	dbChan := make(chan HealthStatus, 1)
	llmChan := make(chan HealthStatus, 1)
	memChan := make(chan HealthStatus, 1)
	diskChan := make(chan HealthStatus, 1)

	// Check database
	go func() {
		dbChan <- s.checkDatabase(checkCtx)
	}()

	// Check LLM API
	go func() {
		llmChan <- s.checkLLMAPI(checkCtx)
	}()

	// Check memory
	go func() {
		memChan <- s.checkMemory()
	}()

	// Check disk
	go func() {
		diskChan <- s.checkDisk()
	}()

	// Collect results
	components := ComponentHealth{
		Database: <-dbChan,
		LLMAPI:   <-llmChan,
		Memory:   <-memChan,
		Disk:     <-diskChan,
	}

	// Determine overall status
	overallStatus := "healthy"
	if components.Database.Status == "unhealthy" || components.Memory.Status == "unhealthy" {
		overallStatus = "unhealthy"
	} else if components.LLMAPI.Status == "unhealthy" || components.Disk.Status == "unhealthy" {
		overallStatus = "degraded"
	}

	// Calculate uptime
	uptime := time.Since(s.startTime)

	return &HealthCheckResponse{
		Status:     overallStatus,
		Components: components,
		Version:    s.version,
		Uptime:     formatUptime(uptime),
		Timestamp:  time.Now(),
	}
}

// checkDatabase checks database connectivity
func (s *HealthService) checkDatabase(ctx context.Context) HealthStatus {
	start := time.Now()

	// Create a timeout context for the ping
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := s.db.PingContext(pingCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   fmt.Sprintf("Database connection failed: %v", err),
			LatencyMs: latency,
		}
	}

	return HealthStatus{
		Status:    "healthy",
		Message:   "connected",
		LatencyMs: latency,
	}
}

// checkLLMAPI checks LLM API availability
func (s *HealthService) checkLLMAPI(ctx context.Context) HealthStatus {
	if s.config.LLMAPIKey == "" {
		return HealthStatus{
			Status:  "unavailable",
			Message: "LLM_API_KEY not configured",
		}
	}

	start := time.Now()

	// Determine the base URL
	baseURL := s.config.LLMBseURL
	if baseURL == "" {
		// P2-04: Use configurable fallback URL instead of hardcoded one
		baseURL = s.config.LLMFallbackURL
		if baseURL == "" {
			// Final fallback if not configured
			baseURL = "https://open.bigmodel.cn/api/paas/v4"
		}
	}

	// Check API endpoint (just a simple connectivity check)
	// We'll make a simple request to check if the API is reachable
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return HealthStatus{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+s.config.LLMAPIKey)

	resp, err := s.httpClient.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   fmt.Sprintf("LLM API unreachable: %v", err),
			LatencyMs: latency,
		}
	}
	defer resp.Body.Close()

	// Consider available if we get any response (even 401/403 means the API is up)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		model := s.config.LLMModel
		if model == "" {
			model = "glm-5"
		}
		return HealthStatus{
			Status:    "available",
			Message:   fmt.Sprintf("API reachable, model: %s", model),
			LatencyMs: latency,
		}
	}

	return HealthStatus{
		Status:    "unhealthy",
		Message:   fmt.Sprintf("LLM API returned status: %d", resp.StatusCode),
		LatencyMs: latency,
	}
}

// checkMemory checks system memory usage
func (s *HealthService) checkMemory() HealthStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get system memory info
	sysMemoryMB := m.Sys / 1024 / 1024
	heapAllocMB := m.HeapAlloc / 1024 / 1024

	// Calculate memory usage percentage (using heap allocation as percentage of system memory)
	var usagePercent float64
	if sysMemoryMB > 0 {
		usagePercent = float64(heapAllocMB) / float64(sysMemoryMB) * 100
	}

	// Determine status based on memory pressure
	status := "healthy"
	message := fmt.Sprintf("%.1f%% used (%d/%d MB)", usagePercent, heapAllocMB, sysMemoryMB)

	// If using more than 85% of available memory, mark as warning
	if usagePercent > 85 {
		status = "warning"
		message = fmt.Sprintf("High memory usage: %.1f%% (%d/%d MB)", usagePercent, heapAllocMB, sysMemoryMB)
	}
	// If using more than 95%, mark as unhealthy
	if usagePercent > 95 {
		status = "unhealthy"
		message = fmt.Sprintf("Critical memory usage: %.1f%% (%d/%d MB)", usagePercent, heapAllocMB, sysMemoryMB)
	}

	return HealthStatus{
		Status:  status,
		Message: message,
	}
}

// checkDisk checks disk usage (basic implementation)
func (s *HealthService) checkDisk() HealthStatus {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return HealthStatus{
			Status:  "unknown",
			Message: fmt.Sprintf("Cannot determine working directory: %v", err),
		}
	}

	// Try to get disk usage stats (Unix-like systems)
	// This is a simplified check - in production, you'd use syscall.Statfs
	// For now, we'll just check if we can write to the directory
	testFile := wd + "/.health_check_test"
	// SEC-CRITICAL-02: 使用安全文件权限0600创建临时文件，防止敏感信息泄露
	f, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return HealthStatus{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Cannot write to disk: %v", err),
		}
	}
	f.Close()
	os.Remove(testFile)

	return HealthStatus{
		Status:  "healthy",
		Message: "disk accessible",
	}
}

// formatUptime formats the uptime duration into a human-readable string
func formatUptime(d time.Duration) string {
	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	if h > 24 {
		days := h / 24
		h = h % 24
		return fmt.Sprintf("%dd%dh%dm", days, h, m)
	}

	return fmt.Sprintf("%dh%dm%ds", h, m, s)
}
