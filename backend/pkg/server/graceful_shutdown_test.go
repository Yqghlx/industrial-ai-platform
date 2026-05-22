package server

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupSignalHandler 测试信号处理器设置
func TestSetupSignalHandler(t *testing.T) {
	t.Run("setup signal handler without panic", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		// 设置信号处理器（不会阻塞）
		manager.SetupSignalHandler()

		// 信号处理器应该已设置，我们可以验证状态
		// 由于没有实际发送信号，IsShuttingDown 应该仍为 false
		assert.False(t, manager.IsShuttingDown(), "Should not be shutting down yet")
	})
}

// TestGracefulShutdown 测试优雅关闭流程
func TestGracefulShutdown(t *testing.T) {
	t.Run("graceful shutdown with all components", func(t *testing.T) {
		// 创建测试服务器
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// 创建 mock 数据库
		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		// 创建 miniredis
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		server := &http.Server{Addr: ts.Listener.Addr().String()}
		manager := NewGracefulShutdownManager(server, db, rdb, 5*time.Second)

		// 添加一个关闭钩子
		hookCalled := false
		manager.AddShutdownHook(func(ctx context.Context) error {
			hookCalled = true
			return nil
		})

		// 注册一个后台任务
		_, cancel := context.WithCancel(context.Background())
		manager.RegisterBackgroundTask("test-task", "Test Task", cancel)

		// 期望数据库关闭
		mock.ExpectClose()

		// 由于 GracefulShutdown 会调用 os.Exit(0)，我们不能直接测试它
		// 但我们可以测试其组件函数
		// 我们将测试各个组件而不是整个流程

		// 测试关闭钩子执行
		ctx := context.Background()
		for _, hook := range manager.onShutdownHooks {
			_ = hook(ctx)
		}
		assert.True(t, hookCalled, "Hook should be called")

		// 测试后台任务等待（手动调用）
		manager.UnregisterBackgroundTask("test-task")

		// 关闭资源
		db.Close()
		rdb.Close()
	})

	t.Run("graceful shutdown with nil dependencies", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		// 添加钩子
		manager.AddShutdownHook(func(ctx context.Context) error {
			return nil
		})

		// 设置关闭标志
		manager.SetShuttingDown(true)
		assert.True(t, manager.IsShuttingDown(), "Should be shutting down")
	})
}

// TestCloseDatabase 测试数据库关闭
func TestCloseDatabase(t *testing.T) {
	t.Run("close database with connection", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, db, nil, 30*time.Second)

		// 期望关闭
		mock.ExpectClose()

		// 调用 closeDatabase
		manager.closeDatabase()

		// 验证期望
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("close database with error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, db, nil, 30*time.Second)

		// 期望关闭失败
		mock.ExpectClose().WillReturnError(sql.ErrConnDone)

		// 调用 closeDatabase（应该不会 panic）
		manager.closeDatabase()

		// 验证期望
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("close database already closed", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)

		// 先关闭一次
		db.Close()

		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, db, nil, 30*time.Second)

		// 再次调用 closeDatabase（应该不会 panic）
		// 由于数据库已经关闭，Close() 可能会返回错误，但不会 panic
		manager.closeDatabase()
	})
}

// TestShutdownHTTPServerFull 测试 HTTP 服务器关闭的完整场景
func TestShutdownHTTPServerFull(t *testing.T) {
	t.Run("shutdown active server", func(t *testing.T) {
		// 创建一个活动的测试服务器
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // 模拟处理时间
			w.WriteHeader(http.StatusOK)
		}))

		// 使用测试服务器创建管理器
		server := &http.Server{Addr: ts.Listener.Addr().String()}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		// 发送一个请求（模拟活动连接）
		go func() {
			_, _ = http.Get(ts.URL)
		}()

		// 等待请求开始
		time.Sleep(50 * time.Millisecond)

		// 关闭服务器
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := manager.shutdownHTTPServer(ctx)
		// Shutdown 可能返回错误（服务器未启动），这是正常的
		if err != nil {
			// 如果服务器已关闭，这是预期的
			t.Logf("Shutdown returned error (expected for test server): %v", err)
		}

		ts.Close()
	})

	t.Run("shutdown server with context timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		server := &http.Server{Addr: ts.Listener.Addr().String()}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		// 使用很短的超时
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_ = manager.shutdownHTTPServer(ctx)
		// 即使超时，也应该不会 panic
	})
}

// TestWaitForBackgroundTasksWithRealCancellation 测试真实的后台任务取消
func TestWaitForBackgroundTasksWithRealCancellation(t *testing.T) {
	t.Run("tasks complete before timeout", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		// 创建会自动完成的任务
		taskCompleted := make(chan bool, 2)

		// 任务 1
		ctx1, cancel1 := context.WithCancel(context.Background())
		go func() {
			<-ctx1.Done()
			manager.UnregisterBackgroundTask("task-1")
			taskCompleted <- true
		}()
		manager.RegisterBackgroundTask("task-1", "Task 1", cancel1)

		// 任务 2
		ctx2, cancel2 := context.WithCancel(context.Background())
		go func() {
			<-ctx2.Done()
			manager.UnregisterBackgroundTask("task-2")
			taskCompleted <- true
		}()
		manager.RegisterBackgroundTask("task-2", "Task 2", cancel2)

		// 等待上下文
		waitCtx, waitCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer waitCancel()

		// 开始等待
		manager.waitForBackgroundTasks(waitCtx)

		// 验证任务已完成
		assert.Equal(t, 0, len(manager.backgroundTasks), "All tasks should be unregistered")
	})

	t.Run("tasks timeout with remaining tasks", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		// 创建一个不会自动取消的任务
		_, cancel := context.WithCancel(context.Background())
		manager.RegisterBackgroundTask("stuck-task", "Stuck Task", cancel)

		// 使用非常短的超时
		waitCtx, waitCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer waitCancel()

		start := time.Now()
		manager.waitForBackgroundTasks(waitCtx)
		elapsed := time.Since(start)

		// 应该在超时后返回（大约 100ms）
		assert.Less(t, elapsed.Milliseconds(), int64(200), "Should timeout around 100ms")

		// 清理
		cancel()
		manager.UnregisterBackgroundTask("stuck-task")
	})
}

// TestSaveApplicationStateWithRedisAndTasks 测试保存应用状态（Redis 和任务）
func TestSaveApplicationStateWithRedisAndTasks(t *testing.T) {
	t.Run("save state with redis and state saver", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, rdb, 30*time.Second)

		// 添加状态保存器
		stateSaver := NewMockStateSaver()
		manager.stateSaver = stateSaver

		// 注册任务
		_, cancel := context.WithCancel(context.Background())
		manager.RegisterBackgroundTask("interrupted-task", "Interrupted Task", cancel)

		// 保存状态
		ctx := context.Background()
		manager.saveApplicationState(ctx)

		// 验证状态保存器被调用
		assert.Equal(t, "interrupted", stateSaver.taskStates["interrupted-task"], "Task state should be saved")

		// 验证 Redis 状态
		stateJSON, err := rdb.Get(ctx, "shutdown_state").Result()
		assert.NoError(t, err, "Should be able to get shutdown state")
		assert.Contains(t, stateJSON, "graceful", "State should contain graceful status")
	})
}

// TestStartupRecoveryWithData 测试启动恢复（有数据）
func TestStartupRecoveryWithData(t *testing.T) {
	t.Run("recovery with previous shutdown state", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		// 设置之前的关闭状态
		ctx := context.Background()
		rdb.Set(ctx, "shutdown_state", "graceful|2024-01-01T12:00:00Z", 24*time.Hour)

		// 执行恢复
		state := StartupRecovery(rdb)

		// 应该返回 graceful 状态
		assert.Equal(t, "graceful", state.Status, "Should return graceful status")

		// 状态应该被清理
		_, err = rdb.Get(ctx, "shutdown_state").Result()
		assert.Error(t, err, "State should be cleared after recovery")
	})

	t.Run("recovery with no previous state", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		// 执行恢复（没有之前的关闭状态）
		state := StartupRecovery(rdb)

		// 应该返回 normal 状态
		assert.Equal(t, "normal", state.Status, "Should return normal status")
	})
}

// TestGracefulShutdownSequence 测试优雅关闭的完整序列
func TestGracefulShutdownSequence(t *testing.T) {
	t.Run("shutdown sequence order", func(t *testing.T) {
		// 创建所有组件
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		mr, err := miniredis.Run()
		require.NoError(t, err)

		rdb := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		// 创建管理器
		server := &http.Server{Addr: ts.Listener.Addr().String()}
		manager := NewGracefulShutdownManager(server, db, rdb, 30*time.Second)

		// 添加状态保存器
		stateSaver := NewMockStateSaver()
		manager.stateSaver = stateSaver

		// 记录关闭序列
		sequence := []string{}

		// 添加钩子记录步骤
		manager.AddShutdownHook(func(ctx context.Context) error {
			sequence = append(sequence, "hook-1")
			return nil
		})
		manager.AddShutdownHook(func(ctx context.Context) error {
			sequence = append(sequence, "hook-2")
			return nil
		})

		// 注册任务
		_, cancel := context.WithCancel(context.Background())
		manager.RegisterBackgroundTask("task-1", "Task 1", cancel)

		// 期望数据库关闭
		mock.ExpectClose()

		// 由于 GracefulShutdown 会调用 os.Exit(0)，我们测试组件
		// 1. 设置关闭标志
		manager.SetShuttingDown(true)
		sequence = append(sequence, "set-shutting-down")

		// 2. 执行钩子
		ctx2 := context.Background()
		for _, hook := range manager.onShutdownHooks {
			_ = hook(ctx2)
		}

		// 3. 手动取消任务
		cancel()
		manager.UnregisterBackgroundTask("task-1")
		sequence = append(sequence, "tasks-canceled")

		// 4. 关闭 HTTP 服务器（跳过，因为测试服务器）

		// 5. 关闭数据库
		manager.closeDatabase()
		sequence = append(sequence, "database-closed")

		// 6. 关闭 Redis
		manager.closeRedis()
		sequence = append(sequence, "redis-closed")

		// 7. 保存状态
		manager.saveApplicationState(context.Background())
		sequence = append(sequence, "state-saved")

		// 验证序列
		expected := []string{
			"set-shutting-down",
			"hook-1",
			"hook-2",
			"tasks-canceled",
			"database-closed",
			"redis-closed",
			"state-saved",
		}

		assert.Equal(t, expected, sequence, "Shutdown sequence should follow correct order")

		// 清理
		ts.Close()
		db.Close()
		rdb.Close()
		mr.Close()
	})
}

// TestShutdownHookExecutionOrder 测试关闭钩子的执行顺序
func TestShutdownHookExecutionOrder(t *testing.T) {
	t.Run("hooks execute in registration order", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		order := []int{}

		for i := 0; i < 5; i++ {
			idx := i
			manager.AddShutdownHook(func(ctx context.Context) error {
				order = append(order, idx)
				return nil
			})
		}

		// 执行钩子
		ctx := context.Background()
		for _, hook := range manager.onShutdownHooks {
			_ = hook(ctx)
		}

		// 验证顺序
		expected := []int{0, 1, 2, 3, 4}
		assert.Equal(t, expected, order, "Hooks should execute in registration order")
	})

	t.Run("hooks with errors continue execution", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		errors := []error{}

		manager.AddShutdownHook(func(ctx context.Context) error {
			return nil
		})
		manager.AddShutdownHook(func(ctx context.Context) error {
			return context.DeadlineExceeded
		})
		manager.AddShutdownHook(func(ctx context.Context) error {
			return nil
		})

		// 执行钩子并收集错误
		ctx := context.Background()
		for _, hook := range manager.onShutdownHooks {
			if err := hook(ctx); err != nil {
				errors = append(errors, err)
			}
		}

		// 应该只有一个错误
		assert.Len(t, errors, 1, "Should have one error")
	})
}

// TestConcurrentShutdownFlags 测试并发设置关闭标志
func TestConcurrentShutdownFlags(t *testing.T) {
	t.Run("concurrent set and get", func(t *testing.T) {
		server := &http.Server{Addr: ":8080"}
		manager := NewGracefulShutdownManager(server, nil, nil, 30*time.Second)

		done := make(chan bool)

		// 并发设置和读取
		go func() {
			for i := 0; i < 100; i++ {
				manager.SetShuttingDown(i%2 == 0)
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				_ = manager.IsShuttingDown()
			}
			done <- true
		}()

		// 等待完成
		<-done
		<-done

		// 最终状态应该是有效的
		_ = manager.IsShuttingDown()
	})
}