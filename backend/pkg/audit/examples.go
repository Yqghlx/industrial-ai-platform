package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// ============================================
// 使用示例
// ============================================

// ExampleUsage 展示审计日志服务的基本使用
func ExampleUsage() {
	// 1. 初始化数据库连接
	// FIX-003: 使用环境变量而非硬编码连接字符串，防止敏感信息泄露
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// 仅用于示例演示，实际使用必须设置环境变量
		fmt.Println("Warning: DATABASE_URL not set. This is an example only.")
		return
	}
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		panic(err)
	}

	// 2. 创建日志记录器
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 3. 创建审计日志仓库
	repo := NewPostgresRepository(db, logger)

	// 4. 创建审计日志配置
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

	// 5. 创建审计日志服务
	auditLogger := NewAuditLogger(repo, logger, config)
	defer auditLogger.Close()

	// 6. 使用审计日志服务
	ctx := context.Background()

	// 示例 1: 记录登录事件
	err = auditLogger.LogLogin(ctx, "user-123", "tenant-456", "session-789", "192.168.1.1", "Mozilla/5.0", true)
	if err != nil {
		logger.Error("Failed to log login event", zap.Error(err))
	}

	// 示例 2: 记录数据访问
	err = auditLogger.LogDataAccess(ctx, "user-123", "tenant-456", "192.168.1.1",
		"device", "device-001", ActionRead, "Read device data", map[string]interface{}{
			"device_name": "Temperature Sensor",
			"location":    "Building A",
		})
	if err != nil {
		logger.Error("Failed to log data access", zap.Error(err))
	}

	// 示例 3: 记录管理操作
	beforeState := map[string]interface{}{
		"username": "olduser",
		"role":     "user",
	}
	afterState := map[string]interface{}{
		"username": "newuser",
		"role":     "admin",
	}
	changes := map[string]interface{}{
		"username": "olduser -> newuser",
		"role":     "user -> admin",
	}

	err = auditLogger.LogAdminAction(ctx, "admin-001", "tenant-456", "192.168.1.1",
		EventAdminUserUpdate, "user", "user-123", "Update user profile",
		beforeState, afterState, changes, map[string]interface{}{
			"approved_by": "super-admin",
			"ticket_id":   "TICKET-123",
		})
	if err != nil {
		logger.Error("Failed to log admin action", zap.Error(err))
	}

	// 示例 4: 记录安全事件
	err = auditLogger.LogSecurityViolation(ctx, "user-123", "tenant-456", "192.168.1.1",
		"unauthorized_access", "Attempted to access restricted resource",
		map[string]interface{}{
			"resource":    "/admin/settings",
			"method":      "GET",
			"status_code": 403,
		})
	if err != nil {
		logger.Error("Failed to log security event", zap.Error(err))
	}

	// 示例 5: 查询审计日志
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	query := &QueryRequest{
		TenantID:  "tenant-456",
		UserID:    "user-123",
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  20,
	}

	logs, total, err := auditLogger.Query(ctx, query)
	if err != nil {
		logger.Error("Failed to query audit logs", zap.Error(err))
	} else {
		logger.Info("Found audit logs",
			zap.Int64("total", total),
			zap.Int("count", len(logs)),
		)
	}

	// 示例 6: 导出审计日志
	jsonData, err := auditLogger.ExportAuditLogs(ctx, query, "json")
	if err != nil {
		logger.Error("Failed to export audit logs", zap.Error(err))
	} else {
		fmt.Printf("Exported logs: %s\n", string(jsonData))
	}

	// 示例 7: 获取统计信息
	stats := auditLogger.GetStats()
	logger.Info("Audit logger stats",
		zap.Int64("total_logs", stats.TotalLogs),
		zap.Int64("success_count", stats.SuccessCount),
		zap.Int64("failure_count", stats.FailureCount),
		zap.Int("queue_size", stats.QueueSize),
	)
}

// ============================================
// HTTP 中间件示例
// ============================================

// AuditMiddleware HTTP 审计中间件
type AuditMiddleware struct {
	auditLogger *AuditLogger
	logger      *zap.Logger
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(auditLogger *AuditLogger, logger *zap.Logger) *AuditMiddleware {
	return &AuditMiddleware{
		auditLogger: auditLogger,
		logger:      logger,
	}
}

// AuditRequest 审计 HTTP 请求
func (m *AuditMiddleware) AuditRequest(userID, tenantID, ipAddress, method, path string, statusCode int, duration time.Duration) {
	ctx := context.Background()

	action := ActionRead
	switch method {
	case "POST", "PUT", "PATCH":
		action = ActionWrite
	case "DELETE":
		action = ActionDelete
	}

	result := ResultSuccess
	if statusCode >= 400 {
		result = ResultFailure
	}

	severity := SeverityInfo
	if statusCode >= 500 {
		severity = SeverityCritical
	} else if statusCode >= 400 {
		severity = SeverityWarning
	}

	err := m.auditLogger.Log(ctx, &AuditLog{
		EventType:     EventDataRead,
		EventCategory: CategoryData,
		Severity:      severity,
		UserID:        userID,
		TenantID:      tenantID,
		IPAddress:     ipAddress,
		ResourceType:  "api_endpoint",
		ResourceID:    path,
		Action:        action,
		Operation:     fmt.Sprintf("%s %s", method, path),
		Result:        result,
		DurationMs:    float64(duration.Milliseconds()),
		Metadata: map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": statusCode,
		},
	})

	if err != nil {
		m.logger.Error("Failed to audit HTTP request", zap.Error(err))
	}
}

// ============================================
// 认证集成示例
// ============================================

// AuthService 认证服务示例
type AuthService struct {
	auditLogger *AuditLogger
	logger      *zap.Logger
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, username, password, ipAddress, userAgent string) (userID, tenantID, sessionID string, err error) {
	// 执行登录逻辑
	// ...

	// 记录登录审计日志
	success := err == nil
	auditErr := s.auditLogger.LogAuthEvent(ctx, EventAuthLogin, userID, tenantID, sessionID, ipAddress, userAgent, success, map[string]interface{}{
		"username": username,
	})

	if auditErr != nil {
		s.logger.Error("Failed to audit login event", zap.Error(auditErr))
	}

	return userID, tenantID, sessionID, err
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, userID, tenantID, sessionID, ipAddress string) error {
	// 执行登出逻辑
	// ...

	// 记录登出审计日志
	err := s.auditLogger.LogLogout(ctx, userID, tenantID, sessionID, ipAddress)
	if err != nil {
		s.logger.Error("Failed to audit logout event", zap.Error(err))
	}

	return nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID, tenantID, ipAddress string, success bool) error {
	// 执行密码修改逻辑
	// ...

	// 记录密码修改审计日志
	err := s.auditLogger.LogPasswordChange(ctx, userID, tenantID, ipAddress, success, map[string]interface{}{
		"changed_by": "user",
		"timestamp":  time.Now(),
	})

	if err != nil {
		s.logger.Error("Failed to audit password change event", zap.Error(err))
	}

	return nil
}

// ============================================
// 数据访问审计示例
// ============================================

// DataService 数据服务示例
type DataService struct {
	auditLogger *AuditLogger
	logger      *zap.Logger
}

// GetDevice 获取设备信息
func (s *DataService) GetDevice(ctx context.Context, userID, tenantID, ipAddress, deviceID string) (interface{}, error) {
	// 执行数据查询逻辑
	// ...

	// 记录数据访问审计日志
	err := s.auditLogger.LogDataAccess(ctx, userID, tenantID, ipAddress,
		"device", deviceID, ActionRead, "Read device information", map[string]interface{}{
			"device_id": deviceID,
		})

	if err != nil {
		s.logger.Error("Failed to audit data access", zap.Error(err))
	}

	return nil, nil
}

// UpdateDevice 更新设备信息
func (s *DataService) UpdateDevice(ctx context.Context, userID, tenantID, ipAddress, deviceID string, beforeState, afterState map[string]interface{}) error {
	// 执行数据更新逻辑
	// ...

	// 计算变更
	changes := make(map[string]interface{})
	for key := range afterState {
		if beforeValue, ok := beforeState[key]; ok {
			if beforeValue != afterState[key] {
				changes[key] = fmt.Sprintf("%v -> %v", beforeValue, afterState[key])
			}
		} else {
			changes[key] = fmt.Sprintf("added: %v", afterState[key])
		}
	}

	// 记录数据访问审计日志
	err := s.auditLogger.Log(ctx, &AuditLog{
		EventType:     EventDataWrite,
		EventCategory: CategoryData,
		Severity:      SeverityInfo,
		UserID:        userID,
		TenantID:      tenantID,
		IPAddress:     ipAddress,
		ResourceType:  "device",
		ResourceID:    deviceID,
		Action:        ActionUpdate,
		Operation:     "Update device configuration",
		BeforeState:   beforeState,
		AfterState:    afterState,
		Changes:       changes,
	})

	if err != nil {
		s.logger.Error("Failed to audit data update", zap.Error(err))
	}

	return nil
}

// DeleteDevice 删除设备
func (s *DataService) DeleteDevice(ctx context.Context, userID, tenantID, ipAddress, deviceID string) error {
	// 执行数据删除逻辑
	// ...

	// 记录数据删除审计日志
	err := s.auditLogger.LogDataAccess(ctx, userID, tenantID, ipAddress,
		"device", deviceID, ActionDelete, "Delete device", map[string]interface{}{
			"device_id": deviceID,
		})

	if err != nil {
		s.logger.Error("Failed to audit data deletion", zap.Error(err))
	}

	return nil
}

// ============================================
// 管理操作审计示例
// ============================================

// AdminService 管理服务示例
type AdminService struct {
	auditLogger *AuditLogger
	logger      *zap.Logger
}

// CreateUser 创建用户
func (s *AdminService) CreateUser(ctx context.Context, adminUserID, tenantID, ipAddress string, userData map[string]interface{}) error {
	// 执行创建用户逻辑
	// ...

	// 记录管理操作审计日志
	err := s.auditLogger.LogAdminAction(ctx, adminUserID, tenantID, ipAddress,
		EventAdminUserCreate, "user", userData["user_id"].(string), "Create new user",
		nil, userData, userData, map[string]interface{}{
			"username": userData["username"],
			"role":     userData["role"],
		})

	if err != nil {
		s.logger.Error("Failed to audit user creation", zap.Error(err))
	}

	return nil
}

// AssignRole 分配角色
func (s *AdminService) AssignRole(ctx context.Context, adminUserID, tenantID, ipAddress, targetUserID, newRole string) error {
	// 执行角色分配逻辑
	// ...

	// 记录管理操作审计日志
	err := s.auditLogger.LogAdminAction(ctx, adminUserID, tenantID, ipAddress,
		EventAdminRoleAssign, "user", targetUserID, "Assign role to user",
		map[string]interface{}{"role": "user"},
		map[string]interface{}{"role": newRole},
		map[string]interface{}{"role": fmt.Sprintf("user -> %s", newRole)},
		map[string]interface{}{
			"target_user_id": targetUserID,
			"new_role":       newRole,
		})

	if err != nil {
		s.logger.Error("Failed to audit role assignment", zap.Error(err))
	}

	return nil
}

// ============================================
// 安全事件审计示例
// ============================================

// SecurityService 安全服务示例
type SecurityService struct {
	auditLogger *AuditLogger
	logger      *zap.Logger
}

// DetectBruteForce 检测暴力破解
func (s *SecurityService) DetectBruteForce(ctx context.Context, userID, tenantID, ipAddress string, attemptCount int) {
	if attemptCount >= 5 {
		// 记录安全告警
		err := s.auditLogger.LogSecurityAlert(ctx, userID, tenantID, ipAddress,
			"brute_force_detected", "Multiple failed login attempts detected",
			map[string]interface{}{
				"attempt_count": attemptCount,
				"time_window":   "5m",
				"blocked":       true,
			})

		if err != nil {
			s.logger.Error("Failed to audit security alert", zap.Error(err))
		}
	}
}

// DetectUnauthorizedAccess 检测未授权访问
func (s *SecurityService) DetectUnauthorizedAccess(ctx context.Context, userID, tenantID, ipAddress, resource string) {
	// 记录安全违规
	err := s.auditLogger.LogSecurityViolation(ctx, userID, tenantID, ipAddress,
		"unauthorized_access", "Attempted to access restricted resource",
		map[string]interface{}{
			"resource":     resource,
			"blocked":      true,
			"action_taken": "access_denied",
		})

	if err != nil {
		s.logger.Error("Failed to audit security violation", zap.Error(err))
	}
}

// BlockIP 阻断 IP
func (s *SecurityService) BlockIP(ctx context.Context, ipAddress, reason string) {
	// 记录安全事件
	err := s.auditLogger.LogSecurityEvent(ctx, "", "", ipAddress,
		EventSecurityBlocked, fmt.Sprintf("IP blocked: %s", reason),
		SeverityWarning, map[string]interface{}{
			"blocked_ip": ipAddress,
			"reason":     reason,
			"timestamp":  time.Now(),
		})

	if err != nil {
		s.logger.Error("Failed to audit IP blocking", zap.Error(err))
	}
}

// ============================================
// 定时任务示例
// ============================================

// AuditMaintenance 审计日志维护
func AuditMaintenance(auditLogger *AuditLogger, logger *zap.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()

		// 清理旧日志
		err := auditLogger.DeleteOld(ctx)
		if err != nil {
			logger.Error("Failed to delete old audit logs", zap.Error(err))
		}

		// 获取统计信息
		stats := auditLogger.GetStats()
		logger.Info("Audit logger statistics",
			zap.Int64("total_logs", stats.TotalLogs),
			zap.Int64("success_count", stats.SuccessCount),
			zap.Int64("failure_count", stats.FailureCount),
			zap.Int("queue_size", stats.QueueSize),
			zap.Int64("dropped_count", stats.DroppedCount),
		)
	}
}

// ============================================
// 查询和分析示例
// ============================================

// AuditAnalyzer 审计日志分析器
type AuditAnalyzer struct {
	auditLogger *AuditLogger
	logger      *zap.Logger
}

// AnalyzeUserActivity 分析用户活动
func (a *AuditAnalyzer) AnalyzeUserActivity(ctx context.Context, userID string, startTime, endTime time.Time) error {
	query := &QueryRequest{
		UserID:    userID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  1000,
	}

	logs, total, err := a.auditLogger.Query(ctx, query)
	if err != nil {
		return err
	}

	// 分析用户活动
	activityCount := make(map[string]int64)
	for _, log := range logs {
		activityCount[log.EventType]++
	}

	a.logger.Info("User activity analysis",
		zap.String("user_id", userID),
		zap.Int64("total_events", total),
		zap.Any("activity_breakdown", activityCount),
	)

	// 转换为 JSON 用于报告
	report, _ := json.Marshal(activityCount)
	fmt.Printf("User Activity Report: %s\n", string(report))

	return nil
}

// AnalyzeSecurityEvents 分析安全事件
func (a *AuditAnalyzer) AnalyzeSecurityEvents(ctx context.Context, startTime, endTime time.Time) error {
	query := &QueryRequest{
		Category:  CategorySecurity,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  1000,
	}

	logs, total, err := a.auditLogger.Query(ctx, query)
	if err != nil {
		return err
	}

	// 分析安全事件
	severityCount := make(map[string]int64)
	eventTypeCount := make(map[string]int64)

	for _, log := range logs {
		severityCount[log.Severity]++
		eventTypeCount[log.EventType]++
	}

	a.logger.Info("Security events analysis",
		zap.Int64("total_events", total),
		zap.Any("severity_breakdown", severityCount),
		zap.Any("event_type_breakdown", eventTypeCount),
	)

	// 转换为 JSON 用于报告
	report, _ := json.Marshal(map[string]interface{}{
		"total_events":         total,
		"severity_breakdown":   severityCount,
		"event_type_breakdown": eventTypeCount,
	})
	fmt.Printf("Security Events Report: %s\n", string(report))

	return nil
}
