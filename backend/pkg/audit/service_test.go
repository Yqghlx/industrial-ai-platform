package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockRepository 模拟仓库
type MockRepository struct {
	mu      sync.Mutex
	logs    []*AuditLog
	queries int64
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		logs: make([]*AuditLog, 0),
	}
}

func (m *MockRepository) Create(ctx context.Context, log *AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockRepository) Query(ctx context.Context, query *QueryRequest) ([]*AuditLog, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queries++
	return m.logs, int64(len(m.logs)), nil
}

func (m *MockRepository) GetByID(ctx context.Context, auditID string) (*AuditLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, log := range m.logs {
		if log.AuditID == auditID {
			return log, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) DeleteOld(ctx context.Context, retentionDays int) error {
	return nil
}

func (m *MockRepository) GetLogCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.logs)
}

// ============================================
// 测试用例
// ============================================

func TestNewAuditLogger(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := DefaultConfig()

	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	assert.NotNil(t, auditLogger)
	assert.Equal(t, LogLevelAll, auditLogger.config.LogLevel)
	assert.True(t, auditLogger.config.AsyncEnabled)
	assert.Equal(t, 10000, auditLogger.config.QueueSize)
}

func TestLogLevelShouldLog(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		severity string
		expected bool
	}{
		{"All logs Info", LogLevelAll, SeverityInfo, true},
		{"All logs Warning", LogLevelAll, SeverityWarning, true},
		{"All logs Critical", LogLevelAll, SeverityCritical, true},
		{"Info level Info", LogLevelInfo, SeverityInfo, true},
		{"Info level Warning", LogLevelInfo, SeverityWarning, true},
		{"Info level Critical", LogLevelInfo, SeverityCritical, true},
		{"Warning level Info", LogLevelWarning, SeverityInfo, false},
		{"Warning level Warning", LogLevelWarning, SeverityWarning, true},
		{"Warning level Critical", LogLevelWarning, SeverityCritical, true},
		{"Critical level Info", LogLevelCritical, SeverityInfo, false},
		{"Critical level Warning", LogLevelCritical, SeverityWarning, false},
		{"Critical level Critical", LogLevelCritical, SeverityCritical, true},
		{"None level", LogLevelNone, SeverityCritical, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.ShouldLog(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogAuthEvent(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false, // 同步测试
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 测试登录成功
	err := auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	assert.NoError(t, err)

	// 测试登录失败
	err = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "", "192.168.1.1", "Mozilla/5.0", false)
	assert.NoError(t, err)

	// 测试登出
	err = auditLogger.LogLogout(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1")
	assert.NoError(t, err)

	// 测试密码修改
	err = auditLogger.LogPasswordChange(context.Background(), "user-123", "tenant-456", "192.168.1.1", true, map[string]interface{}{
		"changed_by": "self",
	})
	assert.NoError(t, err)

	assert.Equal(t, 4, repo.GetLogCount())
}

func TestLogDataAccess(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 测试数据读取
	err := auditLogger.LogDataAccess(context.Background(), "user-123", "tenant-456", "192.168.1.1",
		"device", "device-001", ActionRead, "Read device data", nil)
	assert.NoError(t, err)

	// 测试数据写入
	err = auditLogger.LogDataAccess(context.Background(), "user-123", "tenant-456", "192.168.1.1",
		"device", "device-001", ActionWrite, "Update device configuration", nil)
	assert.NoError(t, err)

	// 测试数据删除
	err = auditLogger.LogDataAccess(context.Background(), "user-123", "tenant-456", "192.168.1.1",
		"device", "device-001", ActionDelete, "Delete device", nil)
	assert.NoError(t, err)

	assert.Equal(t, 3, repo.GetLogCount())
}

func TestLogAdminAction(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 测试创建用户
	err := auditLogger.LogAdminAction(context.Background(), "admin-001", "tenant-456", "192.168.1.1",
		EventAdminUserCreate, "user", "user-002", "Create new user",
		nil, map[string]interface{}{"username": "newuser"}, map[string]interface{}{"username": "newuser"}, nil)
	assert.NoError(t, err)

	// 测试更新用户
	err = auditLogger.LogAdminAction(context.Background(), "admin-001", "tenant-456", "192.168.1.1",
		EventAdminUserUpdate, "user", "user-002", "Update user role",
		map[string]interface{}{"role": "user"}, map[string]interface{}{"role": "admin"}, map[string]interface{}{"role": "admin -> user"}, nil)
	assert.NoError(t, err)

	// 测试删除用户
	err = auditLogger.LogAdminAction(context.Background(), "admin-001", "tenant-456", "192.168.1.1",
		EventAdminUserDelete, "user", "user-002", "Delete user",
		map[string]interface{}{"username": "olduser"}, nil, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, 3, repo.GetLogCount())
}

func TestLogSecurityEvent(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 测试安全违规
	err := auditLogger.LogSecurityViolation(context.Background(), "user-123", "tenant-456", "192.168.1.1",
		"unauthorized_access", "Attempted to access restricted resource", nil)
	assert.NoError(t, err)

	// 测试安全告警
	err = auditLogger.LogSecurityAlert(context.Background(), "user-123", "tenant-456", "192.168.1.1",
		"suspicious_activity", "Multiple failed login attempts", map[string]interface{}{
			"attempt_count": 5,
		})
	assert.NoError(t, err)

	// 测试自定义安全事件
	err = auditLogger.LogSecurityEvent(context.Background(), "user-123", "tenant-456", "192.168.1.1",
		EventSecurityBlocked, "IP blocked due to brute force", SeverityWarning, nil)
	assert.NoError(t, err)

	assert.Equal(t, 3, repo.GetLogCount())
}

func TestAsyncLogging(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:        true,
		LogLevel:       LogLevelAll,
		AsyncEnabled:   true,
		QueueSize:      100,
		WorkerCount:    2,
		BatchSize:      10,
		BatchTimeout:   1,
		RetentionDays:  90,
		EnableMetadata: true,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 写入多条日志
	for i := 0; i < 20; i++ {
		err := auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
		assert.NoError(t, err)
	}

	// 等待异步处理
	time.Sleep(2 * time.Second)

	auditLogger.Close()

	assert.Equal(t, 20, repo.GetLogCount())
}

func TestLogLevelFilter(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelWarning, // 只记录 Warning 和 Critical
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 测试 Info 级别日志（应该被过滤）
	err := auditLogger.Log(context.Background(), &AuditLog{
		EventType:     EventAuthLogin,
		EventCategory: CategoryAuth,
		Severity:      SeverityInfo,
		UserID:        "user-123",
		Operation:     "Test info log",
	})
	assert.NoError(t, err)

	// 测试 Warning 级别日志（应该记录）
	err = auditLogger.Log(context.Background(), &AuditLog{
		EventType:     EventAuthFailed,
		EventCategory: CategoryAuth,
		Severity:      SeverityWarning,
		UserID:        "user-123",
		Operation:     "Test warning log",
	})
	assert.NoError(t, err)

	// 测试 Critical 级别日志（应该记录）
	err = auditLogger.Log(context.Background(), &AuditLog{
		EventType:     EventSecurityViolation,
		EventCategory: CategorySecurity,
		Severity:      SeverityCritical,
		UserID:        "user-123",
		Operation:     "Test critical log",
	})
	assert.NoError(t, err)

	assert.Equal(t, 2, repo.GetLogCount())
}

func TestGetStats(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 写入一些日志
	_ = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	_ = auditLogger.LogLogout(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1")

	stats := auditLogger.GetStats()

	assert.Equal(t, int64(2), stats.TotalLogs)
	assert.Equal(t, int64(2), stats.SuccessCount)
	assert.Equal(t, int64(0), stats.FailureCount)
}

func TestDisabledLogging(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      false, // 禁用审计日志
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 写入日志（应该被忽略）
	err := auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	assert.NoError(t, err)

	assert.Equal(t, 0, repo.GetLogCount())
}

func TestExportAuditLogs(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 写入一些日志
	_ = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	_ = auditLogger.LogLogout(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1")

	// 测试 JSON 导出
	jsonData, err := auditLogger.ExportAuditLogs(context.Background(), &QueryRequest{}, "json")
	assert.NoError(t, err)
	assert.NotNil(t, jsonData)

	// 测试 CSV 导出
	csvData, err := auditLogger.ExportAuditLogs(context.Background(), &QueryRequest{}, "csv")
	assert.NoError(t, err)
	assert.NotNil(t, csvData)
	assert.Contains(t, string(csvData), "AuditID,Timestamp,EventType")
}

func TestConcurrentLogging(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: true,
		QueueSize:    1000,
		WorkerCount:  3,
		BatchSize:    50,
		BatchTimeout: 1,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	// 并发写入日志
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_ = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
		}(i)
	}

	wg.Wait()

	// 等待异步处理完成
	time.Sleep(3 * time.Second)

	auditLogger.Close()

	assert.Equal(t, 100, repo.GetLogCount())
}

// ============================================
// 基准测试
// ============================================

func BenchmarkLogLogin(b *testing.B) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: false,
	}

	auditLogger := NewAuditLogger(repo, logger, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	}
}

func BenchmarkLogLoginAsync(b *testing.B) {
	logger := zap.NewNop()
	repo := NewMockRepository()
	config := &Config{
		Enabled:      true,
		LogLevel:     LogLevelAll,
		AsyncEnabled: true,
		QueueSize:    10000,
		WorkerCount:  3,
		BatchSize:    100,
		BatchTimeout: 5,
	}

	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auditLogger.LogLogin(context.Background(), "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	}
}
