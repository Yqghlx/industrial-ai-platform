package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ============================================
// Example Tests
// ============================================

func TestNewAuditMiddleware(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	middleware := NewAuditMiddleware(auditLogger, logger)
	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.auditLogger)
	assert.NotNil(t, middleware.logger)
}

func TestAuditRequest_Success(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	middleware := NewAuditMiddleware(auditLogger, logger)

	// Test successful GET request
	middleware.AuditRequest("user-123", "tenant-456", "192.168.1.1", "GET", "/api/devices", 200, 100*time.Millisecond)
	assert.Equal(t, 1, repo.GetLogCount())

	logs, _, _ := repo.Query(context.Background(), &QueryRequest{})
	assert.Len(t, logs, 1)
	assert.Equal(t, ActionRead, logs[0].Action)
	assert.Equal(t, ResultSuccess, logs[0].Result)
	assert.Equal(t, SeverityInfo, logs[0].Severity)
}

func TestAuditRequest_ClientError(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	middleware := NewAuditMiddleware(auditLogger, logger)

	// Test client error (4xx)
	middleware.AuditRequest("user-123", "tenant-456", "192.168.1.1", "GET", "/api/devices", 404, 50*time.Millisecond)
	assert.Equal(t, 1, repo.GetLogCount())

	logs, _, _ := repo.Query(context.Background(), &QueryRequest{})
	assert.Len(t, logs, 1)
	assert.Equal(t, ResultFailure, logs[0].Result)
	assert.Equal(t, SeverityWarning, logs[0].Severity)
}

func TestAuditRequest_ServerError(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	middleware := NewAuditMiddleware(auditLogger, logger)

	// Test server error (5xx)
	middleware.AuditRequest("user-123", "tenant-456", "192.168.1.1", "GET", "/api/devices", 500, 200*time.Millisecond)
	assert.Equal(t, 1, repo.GetLogCount())

	logs, _, _ := repo.Query(context.Background(), &QueryRequest{})
	assert.Len(t, logs, 1)
	assert.Equal(t, ResultFailure, logs[0].Result)
	assert.Equal(t, SeverityCritical, logs[0].Severity)
}

func TestAuditRequest_WriteMethods(t *testing.T) {
	logger := zap.NewNop()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}

	// Test POST request
	repo1 := NewMockRepository()
	al1 := NewAuditLogger(repo1, logger, config)
	m1 := NewAuditMiddleware(al1, logger)
	m1.AuditRequest("user-123", "tenant-456", "192.168.1.1", "POST", "/api/devices", 201, 150*time.Millisecond)
	logs1, _, _ := repo1.Query(context.Background(), &QueryRequest{})
	assert.Equal(t, ActionWrite, logs1[0].Action)
	al1.Close()

	// Test PUT request
	repo2 := NewMockRepository()
	al2 := NewAuditLogger(repo2, logger, config)
	m2 := NewAuditMiddleware(al2, logger)
	m2.AuditRequest("user-123", "tenant-456", "192.168.1.1", "PUT", "/api/devices/1", 200, 120*time.Millisecond)
	logs2, _, _ := repo2.Query(context.Background(), &QueryRequest{})
	assert.Equal(t, ActionWrite, logs2[0].Action)
	al2.Close()

	// Test PATCH request
	repo3 := NewMockRepository()
	al3 := NewAuditLogger(repo3, logger, config)
	m3 := NewAuditMiddleware(al3, logger)
	m3.AuditRequest("user-123", "tenant-456", "192.168.1.1", "PATCH", "/api/devices/1", 200, 80*time.Millisecond)
	logs3, _, _ := repo3.Query(context.Background(), &QueryRequest{})
	assert.Equal(t, ActionWrite, logs3[0].Action)
	al3.Close()

	// Test DELETE request
	repo4 := NewMockRepository()
	al4 := NewAuditLogger(repo4, logger, config)
	m4 := NewAuditMiddleware(al4, logger)
	m4.AuditRequest("user-123", "tenant-456", "192.168.1.1", "DELETE", "/api/devices/1", 204, 50*time.Millisecond)
	logs4, _, _ := repo4.Query(context.Background(), &QueryRequest{})
	assert.Equal(t, ActionDelete, logs4[0].Action)
	al4.Close()
}

// ============================================
// AuthService Tests
// ============================================

func TestAuthService_Login(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	authService := &AuthService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	// Test login
	ctx := context.Background()
	_, _, _, err := authService.Login(ctx, "testuser", "password", "192.168.1.1", "Mozilla/5.0")
	// Note: Login returns empty values since it's an example
	assert.NoError(t, err)
}

func TestAuthService_Logout(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	authService := &AuthService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	err := authService.Logout(ctx, "user-123", "tenant-456", "session-789", "192.168.1.1")
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestAuthService_ChangePassword(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	authService := &AuthService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	err := authService.ChangePassword(ctx, "user-123", "tenant-456", "192.168.1.1", true)
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

// ============================================
// DataService Tests
// ============================================

func TestDataService_GetDevice(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	dataService := &DataService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	_, err := dataService.GetDevice(ctx, "user-123", "tenant-456", "192.168.1.1", "device-001")
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestDataService_UpdateDevice(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	dataService := &DataService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	beforeState := map[string]interface{}{"name": "old"}
	afterState := map[string]interface{}{"name": "new"}
	err := dataService.UpdateDevice(ctx, "user-123", "tenant-456", "192.168.1.1", "device-001", beforeState, afterState)
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestDataService_DeleteDevice(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	dataService := &DataService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	err := dataService.DeleteDevice(ctx, "user-123", "tenant-456", "192.168.1.1", "device-001")
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

// ============================================
// AdminService Tests
// ============================================

func TestAdminService_CreateUser(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	adminService := &AdminService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	userData := map[string]interface{}{
		"user_id":  "user-002",
		"username": "newuser",
		"role":     "user",
	}
	err := adminService.CreateUser(ctx, "admin-001", "tenant-456", "192.168.1.1", userData)
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestAdminService_AssignRole(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	adminService := &AdminService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	err := adminService.AssignRole(ctx, "admin-001", "tenant-456", "192.168.1.1", "user-002", "admin")
	assert.NoError(t, err)
	assert.Equal(t, 1, repo.GetLogCount())
}

// ============================================
// SecurityService Tests
// ============================================

func TestSecurityService_DetectBruteForce_BelowThreshold(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	securityService := &SecurityService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	// Below threshold - no alert
	securityService.DetectBruteForce(ctx, "user-123", "tenant-456", "192.168.1.1", 3)
	assert.Equal(t, 0, repo.GetLogCount())
}

func TestSecurityService_DetectBruteForce_AtThreshold(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	securityService := &SecurityService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	// At threshold - should create alert
	securityService.DetectBruteForce(ctx, "user-123", "tenant-456", "192.168.1.1", 5)
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestSecurityService_DetectBruteForce_AboveThreshold(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	securityService := &SecurityService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	// Above threshold - should create alert
	securityService.DetectBruteForce(ctx, "user-123", "tenant-456", "192.168.1.1", 10)
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestSecurityService_DetectUnauthorizedAccess(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	securityService := &SecurityService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	securityService.DetectUnauthorizedAccess(ctx, "user-123", "tenant-456", "192.168.1.1", "/admin/settings")
	assert.Equal(t, 1, repo.GetLogCount())
}

func TestSecurityService_BlockIP(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	securityService := &SecurityService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	securityService.BlockIP(ctx, "192.168.1.100", "brute force attack")
	assert.Equal(t, 1, repo.GetLogCount())
}

// ============================================
// AuditAnalyzer Tests
// ============================================

func TestAuditAnalyzer_AnalyzeUserActivity(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	// Create some logs
	_ = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	_ = auditLogger.LogLogout(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1")
	_ = auditLogger.LogDataAccess(context.Background(), "user-123", "tenant-456", "192.168.1.1", "device", "device-001", ActionRead, "Read device", nil)

	analyzer := &AuditAnalyzer{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	err := analyzer.AnalyzeUserActivity(ctx, "user-123", startTime, endTime)
	assert.NoError(t, err)
}

func TestAuditAnalyzer_AnalyzeSecurityEvents(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	// Create some security logs
	_ = auditLogger.LogSecurityViolation(context.Background(), "user-123", "tenant-456", "192.168.1.1", "unauthorized", "Test", nil)
	_ = auditLogger.LogSecurityAlert(context.Background(), "user-123", "tenant-456", "192.168.1.1", "alert", "Test alert", nil)

	analyzer := &AuditAnalyzer{
		auditLogger: auditLogger,
		logger:      logger,
	}

	ctx := context.Background()
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	err := analyzer.AnalyzeSecurityEvents(ctx, startTime, endTime)
	assert.NoError(t, err)
}

// ============================================
// Benchmark Tests for Examples
// ============================================

func BenchmarkAuditRequest(b *testing.B) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	middleware := NewAuditMiddleware(auditLogger, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.AuditRequest("user-123", "tenant-456", "192.168.1.1", "GET", "/api/devices", 200, 100*time.Millisecond)
	}
}

func BenchmarkSecurityService_DetectBruteForce(b *testing.B) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{Enabled: true, LogLevel: LogLevelAll, AsyncEnabled: false}
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	securityService := &SecurityService{
		auditLogger: auditLogger,
		logger:      logger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		securityService.DetectBruteForce(context.Background(), "user-123", "tenant-456", "192.168.1.1", 5)
	}
}

// ============================================
// Test Constants
// ============================================

func TestEventConstants(t *testing.T) {
	// Test event type constants
	assert.Equal(t, "auth.login", EventAuthLogin)
	assert.Equal(t, "auth.logout", EventAuthLogout)
	assert.Equal(t, "auth.failed", EventAuthFailed)
	assert.Equal(t, "auth.token_refresh", EventAuthTokenRefresh)
	assert.Equal(t, "auth.password_change", EventAuthPasswordChange)

	assert.Equal(t, "authz.grant", EventAuthzGrant)
	assert.Equal(t, "authz.revoke", EventAuthzRevoke)
	assert.Equal(t, "authz.check", EventAuthzCheck)

	assert.Equal(t, "data.read", EventDataRead)
	assert.Equal(t, "data.write", EventDataWrite)
	assert.Equal(t, "data.delete", EventDataDelete)
	assert.Equal(t, "data.export", EventDataExport)

	assert.Equal(t, "config.create", EventConfigCreate)
	assert.Equal(t, "config.update", EventConfigUpdate)
	assert.Equal(t, "config.delete", EventConfigDelete)

	assert.Equal(t, "admin.user_create", EventAdminUserCreate)
	assert.Equal(t, "admin.user_update", EventAdminUserUpdate)
	assert.Equal(t, "admin.user_delete", EventAdminUserDelete)
	assert.Equal(t, "admin.role_assign", EventAdminRoleAssign)
	assert.Equal(t, "admin.role_revoke", EventAdminRoleRevoke)
	assert.Equal(t, "admin.config_change", EventAdminConfigChange)
	assert.Equal(t, "admin.system_restart", EventAdminSystemRestart)

	assert.Equal(t, "system.start", EventSystemStart)
	assert.Equal(t, "system.stop", EventSystemStop)
	assert.Equal(t, "system.restart", EventSystemRestart)

	assert.Equal(t, "security.alert", EventSecurityAlert)
	assert.Equal(t, "security.violation", EventSecurityViolation)
	assert.Equal(t, "security.blocked", EventSecurityBlocked)
	assert.Equal(t, "security.incident", EventSecurityIncident)
}

func TestCategoryConstants(t *testing.T) {
	assert.Equal(t, "auth", CategoryAuth)
	assert.Equal(t, "authz", CategoryAuthz)
	assert.Equal(t, "data", CategoryData)
	assert.Equal(t, "config", CategoryConfig)
	assert.Equal(t, "admin", CategoryAdmin)
	assert.Equal(t, "system", CategorySystem)
	assert.Equal(t, "security", CategorySecurity)
}

func TestActionConstants(t *testing.T) {
	assert.Equal(t, "read", ActionRead)
	assert.Equal(t, "write", ActionWrite)
	assert.Equal(t, "delete", ActionDelete)
	assert.Equal(t, "create", ActionCreate)
	assert.Equal(t, "update", ActionUpdate)
	assert.Equal(t, "export", ActionExport)
}

func TestSeverityConstants(t *testing.T) {
	assert.Equal(t, "info", SeverityInfo)
	assert.Equal(t, "warning", SeverityWarning)
	assert.Equal(t, "critical", SeverityCritical)
}

func TestResultConstants(t *testing.T) {
	assert.Equal(t, "success", ResultSuccess)
	assert.Equal(t, "failure", ResultFailure)
}

// ============================================
// Test Config struct
// ============================================

func TestConfig_Struct(t *testing.T) {
	config := &Config{
		Enabled:        true,
		LogLevel:       LogLevelAll,
		AsyncEnabled:   true,
		QueueSize:      10000,
		WorkerCount:    3,
		BatchSize:      100,
		BatchTimeout:   5,
		RetentionDays:  90,
		EnableMetadata: true,
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, LogLevelAll, config.LogLevel)
	assert.True(t, config.AsyncEnabled)
	assert.Equal(t, 10000, config.QueueSize)
	assert.Equal(t, 3, config.WorkerCount)
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, 5, config.BatchTimeout)
	assert.Equal(t, 90, config.RetentionDays)
	assert.True(t, config.EnableMetadata)
}

// ============================================
// Test AuditStats struct
// ============================================

func TestAuditStats_Struct(t *testing.T) {
	now := time.Now()
	stats := &AuditStats{
		TotalLogs:    100,
		SuccessCount: 90,
		FailureCount: 10,
		QueueSize:    50,
		DroppedCount: 5,
		LastLogTime:  now,
	}

	assert.Equal(t, int64(100), stats.TotalLogs)
	assert.Equal(t, int64(90), stats.SuccessCount)
	assert.Equal(t, int64(10), stats.FailureCount)
	assert.Equal(t, 50, stats.QueueSize)
	assert.Equal(t, int64(5), stats.DroppedCount)
	assert.Equal(t, now, stats.LastLogTime)
}
