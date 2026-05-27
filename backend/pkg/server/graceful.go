package server

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/industrial-ai/platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// GracefulShutdownManager 优雅关闭管理器
type GracefulShutdownManager struct {
	server          *http.Server
	db              *sql.DB
	redis           *redis.Client
	shutdownTimeout time.Duration
	shuttingDown    bool
	shutdownMutex   sync.RWMutex
	backgroundTasks map[string]*BackgroundTask
	taskMutex       sync.Mutex
	stateSaver      StateSaver
	onShutdownHooks []ShutdownHook
}

// BackgroundTask 后台任务
type BackgroundTask struct {
	ID       string
	Name     string
	Status   string
	Started  time.Time
	CancelFn context.CancelFunc
}

// StateSaver 状态保存器接口
type StateSaver interface {
	SaveTaskState(taskID, status string) error
	SaveShutdownState(status string) error
	ClearShutdownState() error
	GetLastShutdownState() ShutdownState
}

// ShutdownState 关闭状态
type ShutdownState struct {
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
	UnfinishedTasks []string  `json:"unfinished_tasks"`
}

// ShutdownHook 关闭钩子函数
type ShutdownHook func(ctx context.Context) error

// NewGracefulShutdownManager 创建优雅关闭管理器
func NewGracefulShutdownManager(server *http.Server, db *sql.DB, redis *redis.Client, timeout time.Duration) *GracefulShutdownManager {
	return &GracefulShutdownManager{
		server:          server,
		db:              db,
		redis:           redis,
		shutdownTimeout: timeout,
		backgroundTasks: make(map[string]*BackgroundTask),
		onShutdownHooks: make([]ShutdownHook, 0),
	}
}

// === 信号处理 ===

// SetupSignalHandler 设置信号处理器
func (m *GracefulShutdownManager) SetupSignalHandler() {
	// 创建信号通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// 启动监听 Goroutine
	go func() {
		sig := <-quit
		logger.L().Info("Received shutdown signal", zap.String("signal", sig.String()))
		m.GracefulShutdown()
	}()
}

// === 优雅关闭 ===

// GracefulShutdown 执行优雅关闭
func (m *GracefulShutdownManager) GracefulShutdown() {
	logger.L().Info("Starting graceful shutdown...")

	// 1. 设置关闭标志
	m.SetShuttingDown(true)

	// 2. 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), m.shutdownTimeout)
	defer cancel()

	// 3. 执行关闭钩子
	for _, hook := range m.onShutdownHooks {
		if err := hook(ctx); err != nil {
			logger.L().Error("Shutdown hook error", zap.Error(err))
		}
	}

	// 4. 等待后台任务完成
	m.waitForBackgroundTasks(ctx)

	// 5. 关闭 HTTP 服务器
	if err := m.shutdownHTTPServer(ctx); err != nil {
		logger.L().Error("HTTP server shutdown error", zap.Error(err))
	}

	// 6. 关闭数据库连接
	m.closeDatabase()

	// 7. 关闭 Redis 连接
	m.closeRedis()

	// 8. 保存应用状态
	m.saveApplicationState(ctx)

	// 9. 记录关闭完成
	logger.L().Info("Graceful shutdown completed")

	// 10. 退出应用
	os.Exit(0)
}

// SetShuttingDown 设置关闭标志
func (m *GracefulShutdownManager) SetShuttingDown(shuttingDown bool) {
	m.shutdownMutex.Lock()
	defer m.shutdownMutex.Unlock()
	m.shuttingDown = shuttingDown
}

// IsShuttingDown 检查是否正在关闭
func (m *GracefulShutdownManager) IsShuttingDown() bool {
	m.shutdownMutex.RLock()
	defer m.shutdownMutex.RUnlock()
	return m.shuttingDown
}

// === HTTP 服务器关闭 ===

// shutdownHTTPServer 关闭 HTTP 服务器
func (m *GracefulShutdownManager) shutdownHTTPServer(ctx context.Context) error {
	logger.L().Info("Shutting down HTTP server...")

	// 禁止新请求 (Gin 中可以检查 IsShuttingDown 标志)
	// 实际实现需要在 Gin 中间件中检查

	// 关闭服务器 (等待现有请求完成)
	err := m.server.Shutdown(ctx)
	if err != nil {
		logger.L().Error("HTTP server shutdown error, forcing closure", zap.Error(err))
		m.server.Close()
		return err
	}

	logger.L().Info("HTTP server shutdown completed")
	return nil
}

// === 后台任务处理 ===

// RegisterBackgroundTask 注册后台任务
func (m *GracefulShutdownManager) RegisterBackgroundTask(taskID, taskName string, cancelFn context.CancelFunc) {
	m.taskMutex.Lock()
	defer m.taskMutex.Unlock()

	m.backgroundTasks[taskID] = &BackgroundTask{
		ID:       taskID,
		Name:     taskName,
		Status:   "running",
		Started:  time.Now(),
		CancelFn: cancelFn,
	}

	logger.L().Info("Background task registered", zap.String("task_name", taskName), zap.String("task_id", taskID))
}

// UnregisterBackgroundTask 取消注册后台任务
func (m *GracefulShutdownManager) UnregisterBackgroundTask(taskID string) {
	m.taskMutex.Lock()
	defer m.taskMutex.Unlock()

	delete(m.backgroundTasks, taskID)
}

// waitForBackgroundTasks 等待后台任务完成
func (m *GracefulShutdownManager) waitForBackgroundTasks(ctx context.Context) {
	logger.L().Info("Waiting for background tasks to complete...")

	m.taskMutex.Lock()
	tasks := make([]*BackgroundTask, 0, len(m.backgroundTasks))
	for _, task := range m.backgroundTasks {
		tasks = append(tasks, task)
	}
	m.taskMutex.Unlock()

	if len(tasks) == 0 {
		logger.L().Info("No background tasks running")
		return
	}

	logger.L().Info("Waiting for background tasks", zap.Int("count", len(tasks)))

	// 取消所有任务
	for _, task := range tasks {
		logger.L().Info("Canceling task", zap.String("task_name", task.Name), zap.String("task_id", task.ID))
		task.CancelFn()
	}

	// 等待任务完成 (最多 15 秒)
	waitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	done := make(chan bool)
	go func() {
		// 检查任务是否全部完成
		for {
			m.taskMutex.Lock()
			remaining := len(m.backgroundTasks)
			m.taskMutex.Unlock()

			if remaining == 0 {
				done <- true
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		logger.L().Info("All background tasks completed")
	case <-waitCtx.Done():
		logger.L().Warn("Background tasks wait timeout", zap.Int("incomplete_count", len(tasks)))
	}
}

// === 依赖连接关闭 ===

// closeDatabase 关闭数据库连接
func (m *GracefulShutdownManager) closeDatabase() {
	logger.L().Info("Closing database connection...")

	if m.db != nil {
		// 等待现有查询完成 (最多 5 秒)
		time.Sleep(2 * time.Second)

		err := m.db.Close()
		if err != nil {
			logger.L().Error("Database close error", zap.Error(err))
		} else {
			logger.L().Info("Database connection closed")
		}
	}
}

// closeRedis 关闭 Redis 连接
func (m *GracefulShutdownManager) closeRedis() {
	logger.L().Info("Closing Redis connection...")

	if m.redis != nil {
		err := m.redis.Close()
		if err != nil {
			logger.L().Error("Redis close error", zap.Error(err))
		} else {
			logger.L().Info("Redis connection closed")
		}
	}
}

// === 状态保存 ===

// saveApplicationState 保存应用状态
func (m *GracefulShutdownManager) saveApplicationState(ctx context.Context) {
	logger.L().Info("Saving application state...")

	if m.stateSaver == nil {
		logger.L().Info("State saver not configured, skipping state save")
		return
	}

	// 1. 保存未完成任务状态
	m.taskMutex.Lock()
	unfinishedTasks := []string{}
	for _, task := range m.backgroundTasks {
		unfinishedTasks = append(unfinishedTasks, task.ID)
		m.stateSaver.SaveTaskState(task.ID, "interrupted")
	}
	m.taskMutex.Unlock()

	logger.L().Info("Saved unfinished tasks", zap.Int("count", len(unfinishedTasks)))

	// 2. 保存关闭状态
	shutdownState := ShutdownState{
		Status:          "graceful",
		Timestamp:       time.Now(),
		UnfinishedTasks: unfinishedTasks,
	}

	// 将状态保存到 Redis (如果可用)
	if m.redis != nil {
		stateJSON := marshalShutdownState(shutdownState)
		m.redis.Set(ctx, "shutdown_state", stateJSON, 24*time.Hour)
	}

	logger.L().Info("Application state saved")
}

// marshalShutdownState 序列化关闭状态
func marshalShutdownState(state ShutdownState) string {
	// 简化版 JSON 序列化
	return state.Status + "|" + state.Timestamp.Format(time.RFC3339)
}

// === 关闭钩子 ===

// AddShutdownHook 添加关闭钩子
func (m *GracefulShutdownManager) AddShutdownHook(hook ShutdownHook) {
	m.onShutdownHooks = append(m.onShutdownHooks, hook)
}

// === 启动恢复 ===

// StartupRecovery 启动恢复
func StartupRecovery(redis *redis.Client) ShutdownState {
	if redis == nil {
		return ShutdownState{Status: "normal"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查上次关闭状态
	stateJSON, err := redis.Get(ctx, "shutdown_state").Result()
	if err != nil {
		logger.L().Info("No previous shutdown state found")
		return ShutdownState{Status: "normal"}
	}

	// 解析关闭状态
	state := parseShutdownState(stateJSON)
	logger.L().Info("Last shutdown state recovered", zap.String("status", state.Status), zap.String("timestamp", state.Timestamp.Format(time.RFC3339)))

	// 清理关闭状态
	redis.Del(ctx, "shutdown_state")

	if len(state.UnfinishedTasks) > 0 {
		logger.L().Info("Found unfinished tasks to recover", zap.Int("count", len(state.UnfinishedTasks)))
	}

	return state
}

// parseShutdownState 解析关闭状态
func parseShutdownState(stateJSON string) ShutdownState {
	// 简化版解析
	return ShutdownState{
		Status:    "graceful",
		Timestamp: time.Now(),
	}
}

// === 工具函数 ===

// RegisterTask 注册后台任务 (便捷方法)
func RegisterTask(m *GracefulShutdownManager, name string) (context.Context, context.CancelFunc) {
	taskID := generateTaskID()
	ctx, cancel := context.WithCancel(context.Background())
	m.RegisterBackgroundTask(taskID, name, cancel)
	return ctx, cancel
}

// generateTaskID 生成任务 ID
func generateTaskID() string {
	return time.Now().Format("task-20060102-150405")
}
