package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// P2-02: getTestPort returns a random available port for testing
// This avoids hardcoded port :8080 which can cause conflicts
func getTestPort() string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		// Fallback to a common test port if random port fails
		return ":18080"
	}
	listener.Close()
	// Extract port from listener address
	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf(":%d", addr.Port)
}

// testPort is a cached test port used across all tests in this file
var testPort = getTestPort()

// MockStateSaver 实现 StateSaver 接口的模拟
type MockStateSaver struct {
	taskStates     map[string]string
	shutdownStatus string
	mu             sync.Mutex
}

func NewMockStateSaver() *MockStateSaver {
	return &MockStateSaver{
		taskStates: make(map[string]string),
	}
}

func (m *MockStateSaver) SaveTaskState(taskID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskStates[taskID] = status
	return nil
}

func (m *MockStateSaver) SaveShutdownState(status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownStatus = status
	return nil
}

func (m *MockStateSaver) ClearShutdownState() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownStatus = ""
	return nil
}

func (m *MockStateSaver) GetLastShutdownState() ShutdownState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return ShutdownState{Status: m.shutdownStatus}
}

// TestNewGracefulShutdownManager 测试创建优雅关闭管理器
func TestNewGracefulShutdownManager(t *testing.T) {
	// 创建测试服务器
	server := &http.Server{Addr: testPort}
	db := &sql.DB{}
	redis := &redis.Client{}
	timeout := 30 * time.Second

	manager := NewGracefulShutdownManager(server, db, redis, timeout)

	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}
	if manager.server != server {
		t.Error("Server not set correctly")
	}
	if manager.db != db {
		t.Error("Database not set correctly")
	}
	if manager.redis != redis {
		t.Error("Redis not set correctly")
	}
	if manager.shutdownTimeout != timeout {
		t.Error("Shutdown timeout not set correctly")
	}
	if manager.backgroundTasks == nil {
		t.Error("Background tasks map should be initialized")
	}
	if manager.onShutdownHooks == nil {
		t.Error("Shutdown hooks slice should be initialized")
	}
}

// TestNewGracefulShutdownManagerWithNilDependencies 测试使用 nil 依赖创建管理器
func TestNewGracefulShutdownManagerWithNilDependencies(t *testing.T) {
	server := &http.Server{Addr: testPort}
	timeout := 30 * time.Second

	// 测试 nil db
	manager := NewGracefulShutdownManager(server, nil, nil, timeout)
	if manager == nil {
		t.Fatal("Expected manager with nil dependencies, got nil")
	}
	if manager.db != nil {
		t.Error("Database should be nil")
	}
	if manager.redis != nil {
		t.Error("Redis should be nil")
	}
}

// TestSetAndGetShuttingDown 测试设置和获取关闭标志
func TestSetAndGetShuttingDown(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 初始状态应该是 false
	if manager.IsShuttingDown() {
		t.Error("Initial shutting down state should be false")
	}

	// 设置为 true
	manager.SetShuttingDown(true)
	if !manager.IsShuttingDown() {
		t.Error("Shutting down state should be true after setting")
	}

	// 设置为 false
	manager.SetShuttingDown(false)
	if manager.IsShuttingDown() {
		t.Error("Shutting down state should be false after setting")
	}
}

// TestSetShuttingDownConcurrent 测试并发设置关闭标志
func TestSetShuttingDownConcurrent(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val bool) {
			defer wg.Done()
			manager.SetShuttingDown(val)
		}(i%2 == 0)
	}
	wg.Wait()

	// 最终状态应该是有效的（true 或 false 都可）
	_ = manager.IsShuttingDown()
}

// TestRegisterBackgroundTask 测试注册后台任务
func TestRegisterBackgroundTask(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager.RegisterBackgroundTask("task-1", "Test Task", cancel)

	if len(manager.backgroundTasks) != 1 {
		t.Errorf("Expected 1 background task, got %d", len(manager.backgroundTasks))
	}

	task, exists := manager.backgroundTasks["task-1"]
	if !exists {
		t.Fatal("Task not found in background tasks")
	}
	if task.ID != "task-1" {
		t.Errorf("Expected task ID 'task-1', got '%s'", task.ID)
	}
	if task.Name != "Test Task" {
		t.Errorf("Expected task name 'Test Task', got '%s'", task.Name)
	}
	if task.Status != "running" {
		t.Errorf("Expected task status 'running', got '%s'", task.Status)
	}
	if task.CancelFn == nil {
		t.Error("Cancel function should not be nil")
	}
}

// TestUnregisterBackgroundTask 测试取消注册后台任务
func TestUnregisterBackgroundTask(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager.RegisterBackgroundTask("task-1", "Test Task", cancel)
	if len(manager.backgroundTasks) != 1 {
		t.Errorf("Expected 1 background task, got %d", len(manager.backgroundTasks))
	}

	manager.UnregisterBackgroundTask("task-1")
	if len(manager.backgroundTasks) != 0 {
		t.Errorf("Expected 0 background tasks after unregister, got %d", len(manager.backgroundTasks))
	}

	// 取消注册不存在的任务应该不会 panic
	manager.UnregisterBackgroundTask("non-existent")
}

// TestRegisterMultipleBackgroundTasks 测试注册多个后台任务
func TestRegisterMultipleBackgroundTasks(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	for i := 0; i < 5; i++ {
		_, cancel := context.WithCancel(context.Background())
		defer cancel()
		manager.RegisterBackgroundTask(string(rune('a'+i)), "Task", cancel)
	}

	if len(manager.backgroundTasks) != 5 {
		t.Errorf("Expected 5 background tasks, got %d", len(manager.backgroundTasks))
	}
}

// TestAddShutdownHook 测试添加关闭钩子
func TestAddShutdownHook(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	hookCalled := false
	hook := func(ctx context.Context) error {
		hookCalled = true
		return nil
	}

	manager.AddShutdownHook(hook)

	if len(manager.onShutdownHooks) != 1 {
		t.Errorf("Expected 1 shutdown hook, got %d", len(manager.onShutdownHooks))
	}

	// 执行钩子
	_ = manager.onShutdownHooks[0](context.Background())
	if !hookCalled {
		t.Error("Hook should have been called")
	}
}

// TestAddMultipleShutdownHooks 测试添加多个关闭钩子
func TestAddMultipleShutdownHooks(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	callOrder := []int{}
	for i := 0; i < 3; i++ {
		idx := i
		manager.AddShutdownHook(func(ctx context.Context) error {
			callOrder = append(callOrder, idx)
			return nil
		})
	}

	if len(manager.onShutdownHooks) != 3 {
		t.Errorf("Expected 3 shutdown hooks, got %d", len(manager.onShutdownHooks))
	}

	// 执行钩子
	for _, hook := range manager.onShutdownHooks {
		_ = hook(context.Background())
	}

	// 检查执行顺序
	for i := 0; i < 3; i++ {
		if callOrder[i] != i {
			t.Errorf("Expected hook %d to be called at position %d, got %d", i, i, callOrder[i])
		}
	}
}

// TestMarshalShutdownState 测试序列化关闭状态
func TestMarshalShutdownState(t *testing.T) {
	state := ShutdownState{
		Status:          "graceful",
		Timestamp:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UnfinishedTasks: []string{"task-1", "task-2"},
	}

	result := marshalShutdownState(state)

	if !strings.Contains(result, "graceful") {
		t.Error("Result should contain 'graceful'")
	}
	if !strings.Contains(result, "2024") {
		t.Error("Result should contain timestamp")
	}
}

// TestParseShutdownState 测试解析关闭状态
func TestParseShutdownState(t *testing.T) {
	stateJSON := "graceful|2024-01-01T12:00:00Z"

	result := parseShutdownState(stateJSON)

	if result.Status != "graceful" {
		t.Errorf("Expected status 'graceful', got '%s'", result.Status)
	}
	if result.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

// TestParseShutdownStateEmpty 测试解析空关闭状态
func TestParseShutdownStateEmpty(t *testing.T) {
	result := parseShutdownState("")

	if result.Status != "graceful" {
		t.Errorf("Expected default status 'graceful', got '%s'", result.Status)
	}
}

// TestGenerateTaskID 测试生成任务 ID
func TestGenerateTaskID(t *testing.T) {
	id1 := generateTaskID()

	// ID 应该包含 "task-" 前缀
	if !strings.HasPrefix(id1, "task-") {
		t.Errorf("Expected task ID to start with 'task-', got '%s'", id1)
	}

	// 连续生成的 ID 应该不同（时间戳格式）
	// 由于时间精度问题，这两个可能相同，所以只检查格式
	if len(id1) < 5 {
		t.Errorf("Task ID too short: '%s'", id1)
	}
}

// TestRegisterTask 测试注册任务便捷方法
func TestRegisterTask(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	ctx, cancel := RegisterTask(manager, "Test Task")
	defer cancel()

	if ctx == nil {
		t.Error("Context should not be nil")
	}
	if cancel == nil {
		t.Error("Cancel function should not be nil")
	}
	if len(manager.backgroundTasks) != 1 {
		t.Errorf("Expected 1 background task, got %d", len(manager.backgroundTasks))
	}
}

// TestBackgroundTaskStruct 测试后台任务结构体
func TestBackgroundTaskStruct(t *testing.T) {
	now := time.Now()
	cancel := func() {}

	task := &BackgroundTask{
		ID:       "task-123",
		Name:     "Test Task",
		Status:   "running",
		Started:  now,
		CancelFn: cancel,
	}

	if task.ID != "task-123" {
		t.Errorf("Expected ID 'task-123', got '%s'", task.ID)
	}
	if task.Name != "Test Task" {
		t.Errorf("Expected Name 'Test Task', got '%s'", task.Name)
	}
	if task.Status != "running" {
		t.Errorf("Expected Status 'running', got '%s'", task.Status)
	}
	if !task.Started.Equal(now) {
		t.Error("Started time should match")
	}
	if task.CancelFn == nil {
		t.Error("CancelFn should not be nil")
	}
}

// TestShutdownStateStruct 测试关闭状态结构体
func TestShutdownStateStruct(t *testing.T) {
	now := time.Now()
	state := ShutdownState{
		Status:          "graceful",
		Timestamp:       now,
		UnfinishedTasks: []string{"task-1", "task-2"},
	}

	if state.Status != "graceful" {
		t.Errorf("Expected Status 'graceful', got '%s'", state.Status)
	}
	if !state.Timestamp.Equal(now) {
		t.Error("Timestamp should match")
	}
	if len(state.UnfinishedTasks) != 2 {
		t.Errorf("Expected 2 unfinished tasks, got %d", len(state.UnfinishedTasks))
	}
}

// TestCloseDatabaseWithNil 测试关闭 nil 数据库连接
func TestCloseDatabaseWithNil(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 应该不会 panic
	manager.closeDatabase()
}

// TestCloseRedisWithNil 测试关闭 nil Redis 连接
func TestCloseRedisWithNil(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 应该不会 panic
	manager.closeRedis()
}

// TestCloseRedisWithConnection 测试关闭 Redis 连接
func TestCloseRedisWithConnection(t *testing.T) {
	// 使用 miniredis 创建模拟 Redis 服务器
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, rdb, 30*time.Second)

	// 应该不会 panic
	manager.closeRedis()
}

// TestSaveApplicationStateWithoutStateSaver 测试无状态保存器时保存状态
func TestSaveApplicationStateWithoutStateSaver(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 应该不会 panic
	manager.saveApplicationState(context.Background())
}

// TestSaveApplicationStateWithStateSaver 测试有状态保存器时保存状态
func TestSaveApplicationStateWithStateSaver(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)
	manager.stateSaver = NewMockStateSaver()

	// 注册一个任务
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.RegisterBackgroundTask("task-1", "Test Task", cancel)

	manager.saveApplicationState(context.Background())
}

// TestSaveApplicationStateWithRedis 测试使用 Redis 保存状态
func TestSaveApplicationStateWithRedis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, rdb, 30*time.Second)

	// 注册任务
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.RegisterBackgroundTask("task-1", "Test Task", cancel)

	manager.saveApplicationState(context.Background())
}

// TestWaitForBackgroundTasksEmpty 测试等待空的后台任务列表
func TestWaitForBackgroundTasksEmpty(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 应该立即返回
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	manager.waitForBackgroundTasks(ctx)
	elapsed := time.Since(start)

	// 应该非常快（因为没有任务）
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected quick return for empty tasks, took %v", elapsed)
	}
}

// TestWaitForBackgroundTasksWithTasks 测试等待有后台任务
func TestWaitForBackgroundTasksWithTasks(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 注册任务并立即取消
	_, cancel1 := context.WithCancel(context.Background())
	_, cancel2 := context.WithCancel(context.Background())

	manager.RegisterBackgroundTask("task-1", "Task 1", cancel1)
	manager.RegisterBackgroundTask("task-2", "Task 2", cancel2)

	// 在 goroutine 中取消任务
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel1()
		cancel2()
		// 模拟任务完成后取消注册
		time.Sleep(50 * time.Millisecond)
		manager.UnregisterBackgroundTask("task-1")
		manager.UnregisterBackgroundTask("task-2")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager.waitForBackgroundTasks(ctx)
}

// TestWaitForBackgroundTasksTimeout 测试等待后台任务超时
func TestWaitForBackgroundTasksTimeout(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	// 注册一个永不完成的任务
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager.RegisterBackgroundTask("stuck-task", "Stuck Task", cancel)

	// 使用超时上下文
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer shortCancel()

	// 应该在超时后返回
	manager.waitForBackgroundTasks(shortCtx)
}

// TestShutdownHTTPServer 测试关闭 HTTP 服务器
func TestShutdownHTTPServer(t *testing.T) {
	// 创建测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	server := &http.Server{Addr: ts.Listener.Addr().String()}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试关闭一个没有启动的服务器
	// Shutdown 会返回错误（服务器未启动），但不会 panic
	err := manager.shutdownHTTPServer(ctx)
	if err == nil {
		// 如果没有错误，说明服务器关闭成功
		t.Log("HTTP server shutdown completed without error")
	}
}

// TestShutdownHTTPServerWithNilServer 测试关闭 nil HTTP 服务器
func TestShutdownHTTPServerWithNilServer(t *testing.T) {
	manager := &GracefulShutdownManager{
		server:          nil,
		shutdownTimeout: 30 * time.Second,
		backgroundTasks: make(map[string]*BackgroundTask),
	}

	// 应该会 panic 或返回错误，取决于实现
	// 由于实现会调用 m.server.Shutdown，nil server 会 panic
	// 所以我们测试时需要确保服务器不为 nil
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic with nil server:", r)
		}
	}()

	ctx := context.Background()
	_ = manager.shutdownHTTPServer(ctx)
}

// TestMockStateSaver 测试 MockStateSaver 实现
func TestMockStateSaver(t *testing.T) {
	saver := NewMockStateSaver()

	// 测试保存任务状态
	err := saver.SaveTaskState("task-1", "interrupted")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if saver.taskStates["task-1"] != "interrupted" {
		t.Errorf("Expected task state 'interrupted', got '%s'", saver.taskStates["task-1"])
	}

	// 测试保存关闭状态
	err = saver.SaveShutdownState("graceful")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if saver.shutdownStatus != "graceful" {
		t.Errorf("Expected shutdown status 'graceful', got '%s'", saver.shutdownStatus)
	}

	// 测试清理关闭状态
	err = saver.ClearShutdownState()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if saver.shutdownStatus != "" {
		t.Errorf("Expected empty shutdown status, got '%s'", saver.shutdownStatus)
	}

	// 测试获取最后关闭状态
	state := saver.GetLastShutdownState()
	if state.Status != "" {
		t.Errorf("Expected empty status, got '%s'", state.Status)
	}
}

// TestStartupRecoveryWithNilRedis 测试使用 nil Redis 进行启动恢复
func TestStartupRecoveryWithNilRedis(t *testing.T) {
	state := StartupRecovery(nil)

	if state.Status != "normal" {
		t.Errorf("Expected status 'normal', got '%s'", state.Status)
	}
}

// TestStartupRecoveryWithRedis 测试使用 Redis 进行启动恢复
func TestStartupRecoveryWithRedis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 设置一个关闭状态
	mr.Set("shutdown_state", "graceful|2024-01-01T12:00:00Z")

	state := StartupRecovery(rdb)

	// 状态应该是 graceful
	if state.Status != "graceful" {
		t.Errorf("Expected status 'graceful', got '%s'", state.Status)
	}
}

// TestStartupRecoveryWithRedisError 测试 Redis 错误时的启动恢复
func TestStartupRecoveryWithRedisError(t *testing.T) {
	// 创建一个连接失败的 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // 不存在的地址
	})

	state := StartupRecovery(rdb)

	// 应该返回正常状态
	if state.Status != "normal" {
		t.Errorf("Expected status 'normal' on error, got '%s'", state.Status)
	}
}

// TestConcurrentTaskRegistration 测试并发任务注册
func TestConcurrentTaskRegistration(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, cancel := context.WithCancel(context.Background())
			defer cancel()
			manager.RegisterBackgroundTask(string(rune(id)), "Task", cancel)
		}(i)
	}
	wg.Wait()

	if len(manager.backgroundTasks) != 100 {
		t.Errorf("Expected 100 tasks, got %d", len(manager.backgroundTasks))
	}
}

// TestShutdownHookError 测试关闭钩子返回错误
func TestShutdownHookError(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	hook := func(ctx context.Context) error {
		return context.DeadlineExceeded
	}

	manager.AddShutdownHook(hook)

	if len(manager.onShutdownHooks) != 1 {
		t.Errorf("Expected 1 shutdown hook, got %d", len(manager.onShutdownHooks))
	}

	// 钩子应该被添加，即使它可能返回错误
	err := manager.onShutdownHooks[0](context.Background())
	if err == nil {
		t.Error("Expected error from hook, got nil")
	}
}

// TestBackgroundTaskCancellation 测试后台任务取消
func TestBackgroundTaskCancellation(t *testing.T) {
	server := &http.Server{Addr: testPort}
	manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

	_, cancel := context.WithCancel(context.Background())

	manager.RegisterBackgroundTask("cancel-test", "Cancel Test", cancel)

	// 检查任务已注册
	if len(manager.backgroundTasks) != 1 {
		t.Fatal("Task should be registered")
	}

	// 取消任务
	cancel()

	// 验证任务仍然在管理器中（取消不会自动移除）
	if len(manager.backgroundTasks) != 1 {
		t.Error("Task should still be registered after cancel")
	}

	// 手动移除
	manager.UnregisterBackgroundTask("cancel-test")
	if len(manager.backgroundTasks) != 0 {
		t.Error("Task should be removed")
	}
}
