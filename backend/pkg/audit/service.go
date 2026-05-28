package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ============================================
// 审计日志核心结构
// ============================================

// AuditLog 审计日志结构
type AuditLog struct {
	AuditID       string                 `json:"audit_id" db:"audit_id"`
	Timestamp     time.Time              `json:"timestamp" db:"timestamp"`
	EventType     string                 `json:"event_type" db:"event_type"`
	EventCategory string                 `json:"event_category" db:"event_category"`
	Severity      string                 `json:"severity" db:"severity"`
	UserID        string                 `json:"user_id" db:"user_id"`
	TenantID      string                 `json:"tenant_id" db:"tenant_id"`
	SessionID     string                 `json:"session_id" db:"session_id"`
	IPAddress     string                 `json:"ip_address" db:"ip_address"`
	UserAgent     string                 `json:"user_agent" db:"user_agent"`
	ResourceType  string                 `json:"resource_type" db:"resource_type"`
	ResourceID    string                 `json:"resource_id" db:"resource_id"`
	Action        string                 `json:"action" db:"action"`
	Operation     string                 `json:"operation" db:"operation"`
	RequestID     string                 `json:"request_id" db:"request_id"`
	TraceID       string                 `json:"trace_id" db:"trace_id"`
	BeforeState   map[string]interface{} `json:"before_state" db:"before_state"`
	AfterState    map[string]interface{} `json:"after_state" db:"after_state"`
	Changes       map[string]interface{} `json:"changes" db:"changes"`
	Result        string                 `json:"result" db:"result"`
	ErrorMessage  string                 `json:"error_message" db:"error_message"`
	DurationMs    float64                `json:"duration_ms" db:"duration_ms"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// ============================================
// 审计事件类型定义
// ============================================

// EventType 事件类型常量
const (
	// 认证事件
	EventAuthLogin          = "auth.login"
	EventAuthLogout         = "auth.logout"
	EventAuthFailed         = "auth.failed"
	EventAuthTokenRefresh   = "auth.token_refresh"
	EventAuthPasswordChange = "auth.password_change"

	// 授权事件
	EventAuthzGrant  = "authz.grant"
	EventAuthzRevoke = "authz.revoke"
	EventAuthzCheck  = "authz.check"

	// 数据访问事件
	EventDataRead   = "data.read"
	EventDataWrite  = "data.write"
	EventDataDelete = "data.delete"
	EventDataExport = "data.export"

	// 配置变更事件
	EventConfigCreate = "config.create"
	EventConfigUpdate = "config.update"
	EventConfigDelete = "config.delete"

	// 管理操作事件
	EventAdminUserCreate    = "admin.user_create"
	EventAdminUserUpdate    = "admin.user_update"
	EventAdminUserDelete    = "admin.user_delete"
	EventAdminRoleAssign    = "admin.role_assign"
	EventAdminRoleRevoke    = "admin.role_revoke"
	EventAdminConfigChange  = "admin.config_change"
	EventAdminSystemRestart = "admin.system_restart"

	// 系统操作事件
	EventSystemStart   = "system.start"
	EventSystemStop    = "system.stop"
	EventSystemRestart = "system.restart"

	// 安全事件
	EventSecurityAlert     = "security.alert"
	EventSecurityViolation = "security.violation"
	EventSecurityBlocked   = "security.blocked"
	EventSecurityIncident  = "security.incident"
)

// EventCategory 事件分类常量
const (
	CategoryAuth     = "auth"
	CategoryAuthz    = "authz"
	CategoryData     = "data"
	CategoryConfig   = "config"
	CategoryAdmin    = "admin"
	CategorySystem   = "system"
	CategorySecurity = "security"
)

// Action 操作类型常量
const (
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionDelete = "delete"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionExport = "export"
)

// Severity 严重程度常量
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Result 结果类型常量
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// ============================================
// 日志级别配置
// ============================================

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelAll LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelCritical
	LogLevelNone
)

// ShouldLog 判断是否应该记录日志
func (l LogLevel) ShouldLog(severity string) bool {
	switch l {
	case LogLevelAll:
		return true
	case LogLevelInfo:
		return severity == SeverityInfo || severity == SeverityWarning || severity == SeverityCritical
	case LogLevelWarning:
		return severity == SeverityWarning || severity == SeverityCritical
	case LogLevelCritical:
		return severity == SeverityCritical
	case LogLevelNone:
		return false
	default:
		return true
	}
}

// ============================================
// 审计日志配置
// ============================================

// Config 审计日志配置
type Config struct {
	// Enabled 是否启用审计日志
	Enabled bool `json:"enabled"`

	// LogLevel 日志级别
	LogLevel LogLevel `json:"log_level"`

	// AsyncEnabled 是否启用异步写入
	AsyncEnabled bool `json:"async_enabled"`

	// QueueSize 异步队列大小
	QueueSize int `json:"queue_size"`

	// WorkerCount 工作协程数量
	WorkerCount int `json:"worker_count"`

	// BatchSize 批量写入大小
	BatchSize int `json:"batch_size"`

	// BatchTimeout 批量写入超时时间（秒）
	BatchTimeout int `json:"batch_timeout"`

	// RetentionDays 日志保留天数
	RetentionDays int `json:"retention_days"`

	// EnableMetadata 是否记录元数据
	EnableMetadata bool `json:"enable_metadata"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
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
}

// ============================================
// 审计日志服务（AuditLogger）
// ============================================

// AuditLogger 审计日志服务
type AuditLogger struct {
	repo   Repository
	logger *zap.Logger
	config *Config

	// 异步写入队列
	auditQueue chan *AuditLog

	// 批量写入缓冲区
	batchBuffer []*AuditLog
	batchMutex  sync.Mutex

	// 关闭信号
	closeChan chan struct{}
	waitGroup sync.WaitGroup

	// BE-P2-06: Context 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc

	// 统计信息
	stats      *AuditStats
	statsMutex sync.RWMutex
}

// AuditStats 审计统计信息
type AuditStats struct {
	TotalLogs    int64
	SuccessCount int64
	FailureCount int64
	QueueSize    int
	DroppedCount int64
	LastLogTime  time.Time
}

// NewAuditLogger 创建审计日志服务
func NewAuditLogger(repo Repository, logger *zap.Logger, config *Config) *AuditLogger {
	if config == nil {
		config = DefaultConfig()
	}

	// BE-P2-06: 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	auditLogger := &AuditLogger{
		repo:        repo,
		logger:      logger,
		config:      config,
		batchBuffer: make([]*AuditLog, 0, config.BatchSize),
		closeChan:   make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
		stats:       &AuditStats{},
	}

	// 启用异步写入
	if config.AsyncEnabled && config.QueueSize > 0 {
		auditLogger.auditQueue = make(chan *AuditLog, config.QueueSize)
		auditLogger.startWorkers()
	}

	return auditLogger
}

// ============================================
// 工作协程管理
// ============================================

// startWorkers 启动工作协程
func (a *AuditLogger) startWorkers() {
	for i := 0; i < a.config.WorkerCount; i++ {
		a.waitGroup.Add(1)
		go a.worker(i)
	}

	// 启动批量写入定时器
	a.waitGroup.Add(1)
	go a.batchTimer()

	a.logger.Info("Audit logger workers started",
		zap.Int("worker_count", a.config.WorkerCount),
		zap.Int("queue_size", a.config.QueueSize),
	)
}

// worker 工作协程
// BE-P2-06: 使用 context 控制生命周期
func (a *AuditLogger) worker(id int) {
	defer a.waitGroup.Done()

	for {
		select {
		case log := <-a.auditQueue:
			if log == nil {
				continue
			}

			// 添加到批量缓冲区
			a.batchMutex.Lock()
			a.batchBuffer = append(a.batchBuffer, log)

			// 达到批量大小，执行写入
			if len(a.batchBuffer) >= a.config.BatchSize {
				batch := a.batchBuffer
				a.batchBuffer = make([]*AuditLog, 0, a.config.BatchSize)
				a.batchMutex.Unlock()

				a.writeBatch(batch)
			} else {
				a.batchMutex.Unlock()
			}

		case <-a.closeChan:
			// 关闭前刷新剩余日志
			a.flushRemaining()
			return

		case <-a.ctx.Done():
			// BE-P2-06: Context 取消时也优雅退出
			a.flushRemaining()
			return
		}
	}
}

// batchTimer 批量写入定时器
// BE-P2-06: 使用 context 控制生命周期
func (a *AuditLogger) batchTimer() {
	defer a.waitGroup.Done()

	ticker := time.NewTicker(time.Duration(a.config.BatchTimeout) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.flushBatch()
		case <-a.closeChan:
			return
		case <-a.ctx.Done():
			// BE-P2-06: Context 取消时也优雅退出
			return
		}
	}
}

// flushBatch 刷新批量缓冲区
func (a *AuditLogger) flushBatch() {
	a.batchMutex.Lock()
	if len(a.batchBuffer) == 0 {
		a.batchMutex.Unlock()
		return
	}

	batch := a.batchBuffer
	a.batchBuffer = make([]*AuditLog, 0, a.config.BatchSize)
	a.batchMutex.Unlock()

	a.writeBatch(batch)
}

// flushRemaining 刷新剩余日志
func (a *AuditLogger) flushRemaining() {
	a.flushBatch()

	// 处理队列中剩余的日志
	for {
		select {
		case log := <-a.auditQueue:
			if log == nil {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := a.repo.Create(ctx, log); err != nil {
				a.logger.Error("Failed to write remaining audit log",
					zap.String("audit_id", log.AuditID),
					zap.Error(err),
				)
			}
			cancel()

		default:
			return
		}
	}
}

// writeBatch 批量写入日志
func (a *AuditLogger) writeBatch(batch []*AuditLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, log := range batch {
		if err := a.repo.Create(ctx, log); err != nil {
			a.logger.Error("Failed to write audit log",
				zap.String("audit_id", log.AuditID),
				zap.String("event_type", log.EventType),
				zap.Error(err),
			)

			a.updateStats(false)
		} else {
			a.updateStats(true)
		}
	}
}

// updateStats 更新统计信息
func (a *AuditLogger) updateStats(success bool) {
	a.statsMutex.Lock()
	defer a.statsMutex.Unlock()

	a.stats.TotalLogs++
	if success {
		a.stats.SuccessCount++
	} else {
		a.stats.FailureCount++
	}
	a.stats.LastLogTime = time.Now()
	if a.auditQueue != nil {
		a.stats.QueueSize = len(a.auditQueue)
	}
}

// Close 关闭审计日志服务
// BE-P2-06: 添加 Context 取消确保所有 goroutine 优雅退出
func (a *AuditLogger) Close() error {
	if a.auditQueue != nil {
		// 先取消 context，让所有 goroutine 知道需要退出
		if a.cancel != nil {
			a.cancel()
		}
		close(a.closeChan)
		a.waitGroup.Wait()
		close(a.auditQueue)
	}

	a.logger.Info("Audit logger closed")
	return nil
}

// Shutdown 优雅关闭（带超时）
// BE-P2-06: 新增带超时的关闭方法
func (a *AuditLogger) Shutdown(ctx context.Context) error {
	if a.auditQueue == nil {
		return nil
	}

	// 取消 context
	if a.cancel != nil {
		a.cancel()
	}
	close(a.closeChan)

	// 等待 goroutine 退出或超时
	done := make(chan struct{})
	go func() {
		a.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(a.auditQueue)
		a.logger.Info("Audit logger shutdown completed")
		return nil
	case <-ctx.Done():
		a.logger.Warn("Audit logger shutdown timeout")
		return ctx.Err()
	}
}

// GetStats 获取统计信息
func (a *AuditLogger) GetStats() *AuditStats {
	a.statsMutex.RLock()
	defer a.statsMutex.RUnlock()

	stats := *a.stats
	return &stats
}

// ============================================
// 核心日志记录方法
// ============================================

// Log 记录审计日志
func (a *AuditLogger) Log(ctx context.Context, log *AuditLog) error {
	// 检查是否启用
	if !a.config.Enabled {
		return nil
	}

	// 检查日志级别
	if !a.config.LogLevel.ShouldLog(log.Severity) {
		return nil
	}

	// 生成审计 ID
	if log.AuditID == "" {
		log.AuditID = "audit-" + uuid.New().String()
	}

	// 设置时间戳
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// 设置创建时间
	log.CreatedAt = time.Now()

	// 设置默认值
	if log.Severity == "" {
		log.Severity = SeverityInfo
	}
	if log.Result == "" {
		log.Result = ResultSuccess
	}

	// 异步写入
	if a.config.AsyncEnabled && a.auditQueue != nil {
		select {
		case a.auditQueue <- log:
			return nil
		default:
			// 队列已满，记录警告并尝试同步写入
			a.logger.Warn("Audit queue is full, attempting sync write",
				zap.String("audit_id", log.AuditID),
			)

			a.statsMutex.Lock()
			a.stats.DroppedCount++
			a.statsMutex.Unlock()
		}
	}

	// 同步写入
	err := a.repo.Create(ctx, log)
	if err != nil {
		a.logger.Error("Failed to create audit log",
			zap.String("audit_id", log.AuditID),
			zap.String("event_type", log.EventType),
			zap.Error(err),
		)
		return err
	}

	a.updateStats(true)

	return nil
}

// ============================================
// 认证事件日志方法
// ============================================

// LogAuthEvent 记录认证事件（登录/登出/密码修改）
func (a *AuditLogger) LogAuthEvent(ctx context.Context, eventType, userID, tenantID, sessionID, ipAddress, userAgent string, success bool, metadata map[string]interface{}) error {
	log := &AuditLog{
		EventType:     eventType,
		EventCategory: CategoryAuth,
		UserID:        userID,
		TenantID:      tenantID,
		SessionID:     sessionID,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Action:        ActionRead,
		Result:        ResultSuccess,
		Metadata:      metadata,
	}

	// 根据事件类型设置操作描述
	switch eventType {
	case EventAuthLogin:
		log.Operation = "User login"
		if !success {
			log.EventType = EventAuthFailed
			log.Result = ResultFailure
			log.Severity = SeverityWarning
		}
	case EventAuthLogout:
		log.Operation = "User logout"
	case EventAuthPasswordChange:
		log.Operation = "Password changed"
		log.Severity = SeverityWarning
		if !success {
			log.Result = ResultFailure
			log.Severity = SeverityCritical
		}
	case EventAuthTokenRefresh:
		log.Operation = "Token refresh"
		if !success {
			log.Result = ResultFailure
			log.Severity = SeverityWarning
		}
	default:
		log.Operation = "Authentication event"
	}

	return a.Log(ctx, log)
}

// LogLogin 记录登录事件
func (a *AuditLogger) LogLogin(ctx context.Context, userID, tenantID, sessionID, ipAddress, userAgent string, success bool) error {
	return a.LogAuthEvent(ctx, EventAuthLogin, userID, tenantID, sessionID, ipAddress, userAgent, success, nil)
}

// LogLogout 记录登出事件
func (a *AuditLogger) LogLogout(ctx context.Context, userID, tenantID, sessionID, ipAddress string) error {
	return a.LogAuthEvent(ctx, EventAuthLogout, userID, tenantID, sessionID, ipAddress, "", true, nil)
}

// LogPasswordChange 记录密码修改事件
func (a *AuditLogger) LogPasswordChange(ctx context.Context, userID, tenantID, ipAddress string, success bool, metadata map[string]interface{}) error {
	return a.LogAuthEvent(ctx, EventAuthPasswordChange, userID, tenantID, "", ipAddress, "", success, metadata)
}

// ============================================
// 数据访问日志方法
// ============================================

// LogDataAccess 记录数据访问事件
func (a *AuditLogger) LogDataAccess(ctx context.Context, userID, tenantID, ipAddress string, resourceType, resourceID, action, operation string, metadata map[string]interface{}) error {
	log := &AuditLog{
		EventType:     EventDataRead,
		EventCategory: CategoryData,
		UserID:        userID,
		TenantID:      tenantID,
		IPAddress:     ipAddress,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Action:        action,
		Operation:     operation,
		Metadata:      metadata,
	}

	// 根据操作类型设置事件类型
	switch action {
	case ActionWrite:
		log.EventType = EventDataWrite
	case ActionDelete:
		log.EventType = EventDataDelete
		log.Severity = SeverityWarning
	case ActionExport:
		log.EventType = EventDataExport
		log.Severity = SeverityWarning
	}

	return a.Log(ctx, log)
}

// ============================================
// 管理操作日志方法
// ============================================

// LogAdminAction 记录管理操作事件
func (a *AuditLogger) LogAdminAction(ctx context.Context, adminUserID, tenantID, ipAddress string, eventType, resourceType, resourceID, operation string, beforeState, afterState, changes map[string]interface{}, metadata map[string]interface{}) error {
	log := &AuditLog{
		EventType:     eventType,
		EventCategory: CategoryAdmin,
		Severity:      SeverityWarning,
		UserID:        adminUserID,
		TenantID:      tenantID,
		IPAddress:     ipAddress,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Operation:     operation,
		BeforeState:   beforeState,
		AfterState:    afterState,
		Changes:       changes,
		Metadata:      metadata,
	}

	// 根据事件类型设置操作
	switch eventType {
	case EventAdminUserCreate:
		log.Action = ActionCreate
		log.Operation = "Admin: Create user"
	case EventAdminUserUpdate:
		log.Action = ActionUpdate
		log.Operation = "Admin: Update user"
	case EventAdminUserDelete:
		log.Action = ActionDelete
		log.Operation = "Admin: Delete user"
		log.Severity = SeverityCritical
	case EventAdminRoleAssign:
		log.Action = ActionUpdate
		log.Operation = "Admin: Assign role"
		log.Severity = SeverityCritical
	case EventAdminRoleRevoke:
		log.Action = ActionUpdate
		log.Operation = "Admin: Revoke role"
		log.Severity = SeverityCritical
	case EventAdminConfigChange:
		log.Action = ActionUpdate
		log.Operation = "Admin: Change configuration"
		log.Severity = SeverityCritical
	case EventAdminSystemRestart:
		log.Action = ActionUpdate
		log.Operation = "Admin: System restart"
		log.Severity = SeverityCritical
	}

	return a.Log(ctx, log)
}

// ============================================
// 安全事件日志方法
// ============================================

// LogSecurityEvent 记录安全事件
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, userID, tenantID, ipAddress string, eventType, operation, severity string, metadata map[string]interface{}) error {
	log := &AuditLog{
		EventType:     eventType,
		EventCategory: CategorySecurity,
		Severity:      severity,
		UserID:        userID,
		TenantID:      tenantID,
		IPAddress:     ipAddress,
		Action:        ActionRead,
		Operation:     operation,
		Result:        ResultFailure,
		Metadata:      metadata,
	}

	// 根据事件类型设置结果
	switch eventType {
	case EventSecurityAlert:
		log.Result = ResultSuccess
	case EventSecurityViolation:
		log.Severity = SeverityCritical
	case EventSecurityBlocked:
		log.Severity = SeverityWarning
	case EventSecurityIncident:
		log.Severity = SeverityCritical
	}

	return a.Log(ctx, log)
}

// LogSecurityViolation 记录安全违规事件
func (a *AuditLogger) LogSecurityViolation(ctx context.Context, userID, tenantID, ipAddress string, violationType, description string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["violation_type"] = violationType
	metadata["description"] = description

	return a.LogSecurityEvent(ctx, userID, tenantID, ipAddress, EventSecurityViolation, description, SeverityCritical, metadata)
}

// LogSecurityAlert 记录安全告警事件
func (a *AuditLogger) LogSecurityAlert(ctx context.Context, userID, tenantID, ipAddress string, alertType, description string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["alert_type"] = alertType
	metadata["description"] = description

	return a.LogSecurityEvent(ctx, userID, tenantID, ipAddress, EventSecurityAlert, description, SeverityWarning, metadata)
}

// ============================================
// 查询方法
// ============================================

// Query 查询审计日志
func (a *AuditLogger) Query(ctx context.Context, query *QueryRequest) ([]*AuditLog, int64, error) {
	return a.repo.Query(ctx, query)
}

// GetByID 获取审计日志详情
func (a *AuditLogger) GetByID(ctx context.Context, auditID string) (*AuditLog, error) {
	return a.repo.GetByID(ctx, auditID)
}

// DeleteOld 删除旧审计日志
func (a *AuditLogger) DeleteOld(ctx context.Context) error {
	return a.repo.DeleteOld(ctx, a.config.RetentionDays)
}

// ============================================
// 审计日志查询请求
// ============================================

// QueryRequest 审计日志查询请求
type QueryRequest struct {
	TenantID     string     `json:"tenant_id"`
	UserID       string     `json:"user_id"`
	EventType    string     `json:"event_type"`
	Category     string     `json:"category"`
	ResourceType string     `json:"resource_type"`
	ResourceID   string     `json:"resource_id"`
	Result       string     `json:"result"`
	IPAddress    string     `json:"ip_address"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Page         int        `json:"page"`
	PageSize     int        `json:"page_size"`
}

// ============================================
// 审计日志仓库接口
// ============================================

// Repository 审计日志仓库接口
type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	Query(ctx context.Context, query *QueryRequest) ([]*AuditLog, int64, error)
	GetByID(ctx context.Context, auditID string) (*AuditLog, error)
	DeleteOld(ctx context.Context, retentionDays int) error
}

// ============================================
// 兼容性支持 - Service 别名
// ============================================

// Service 审计日志服务（兼容旧代码）
type Service = AuditLogger

// NewService 创建审计日志服务（兼容旧代码）
func NewService(repo Repository, logger *zap.Logger) *Service {
	return NewAuditLogger(repo, logger, DefaultConfig())
}

// ============================================
// 审计日志导出
// ============================================

// ExportAuditLogs 导出审计日志
func (a *AuditLogger) ExportAuditLogs(ctx context.Context, query *QueryRequest, format string) ([]byte, error) {
	logs, _, err := a.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.Marshal(logs)
	case "csv":
		return exportCSV(logs)
	default:
		return json.Marshal(logs)
	}
}

// exportCSV 导出 CSV 格式
func exportCSV(logs []*AuditLog) ([]byte, error) {
	var result string
	result += "AuditID,Timestamp,EventType,UserID,TenantID,IPAddress,Action,Operation,Result\n"

	for _, log := range logs {
		result += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			log.AuditID,
			log.Timestamp.Format(time.RFC3339),
			log.EventType,
			log.UserID,
			log.TenantID,
			log.IPAddress,
			log.Action,
			log.Operation,
			log.Result,
		)
	}

	return []byte(result), nil
}
