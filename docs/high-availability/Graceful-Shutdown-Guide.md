# 优雅关闭与故障恢复指南

> **Industrial AI Platform 优雅关闭与故障恢复最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 优雅关闭概述

Phase 4 P1 高可用优雅关闭目标：

| 指标 | 当前状态 | 目标值 |
|------|---------|--------|
| **关闭超时处理** | 无 | 30 秒优雅关闭 |
| **连接清理** | 无 | 主动关闭连接 |
| **请求处理** | 立即中断 | 等待完成 |
| **状态保存** | 无 | 关闭前保存 |
| **恢复时间** | ~5min | <30s |

---

## 🔄 优雅关闭流程

### 关闭流程图

```
┌─────────────────────────────────────────┐
│  1. 收到关闭信号 (SIGTERM/SIGINT)        │
│  - Kubernetes: SIGTERM                  │
│  - Docker: SIGTERM                      │
│  - 手动: Ctrl+C (SIGINT)                │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  2. 进入优雅关闭模式                      │
│  - 设置 shuttingDown 标志                │
│  - 停止接收新请求                        │
│  - Readiness 探针返回 not_ready          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  3. 等待现有请求完成                      │
│  - 等待 HTTP 请求完成                    │
│  - 等待 WebSocket 连接关闭               │
│  - 等待后台任务完成                      │
│  - 最大等待时间: 30 秒                   │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  4. 关闭依赖连接                          │
│  - 关闭数据库连接池                      │
│  - 关闭 Redis 连接                       │
│  - 关闭消息队列连接                      │
│  - 关闭外部 API 连接                     │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  5. 保存应用状态                          │
│  - 保存后台任务状态                      │
│  - 保存缓存状态                          │
│  - 保存未完成操作                        │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  6. 退出应用                              │
│  - 记录关闭日志                          │
│  - 返回退出码 0                          │
└─────────────────────────────────────────┘
```

---

## 🔧 Kubernetes 优雅关闭配置

### Pod 终止流程

```yaml
# K8s Pod 终止流程
spec:
  terminationGracePeriodSeconds: 30  # 优雅关闭等待 30 秒
  
  # PreStop Hook (可选)
  lifecycle:
    preStop:
      exec:
        command: ["/bin/sh", "-c", "sleep 10"]  # 预停止等待
```

### Pod 终止时间线

| 时间 | 事件 |
|------|------|
| **T+0s** | Pod 收到 SIGTERM |
| **T+0s** | Service 从 Endpoints 移除 Pod |
| **T+0s** | 应用进入优雅关闭模式 |
| **T+0-30s** | 等待现有请求完成 |
| **T+30s** | 如果未完成，发送 SIGKILL |
| **T+30s+** | 容器强制终止 |

---

## 💻 Go 优雅关闭实现

### 信号处理

```go
// backend/pkg/server/graceful.go

func SetupGracefulShutdown(server *http.Server, shutdownTimeout time.Duration) {
    // 创建信号通道
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
    
    // 等待信号
    sig := <-quit
    log.Printf("Received signal: %v", sig)
    
    // 进入优雅关闭模式
    gracefulShutdown(server, shutdownTimeout)
}

func gracefulShutdown(server *http.Server, timeout time.Duration) {
    // 1. 创建关闭上下文
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    // 2. 停止接收新请求 (设置 shuttingDown 标志)
    SetShuttingDown(true)
    
    // 3. 关闭 HTTP 服务器
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
        server.Close() // 强制关闭
    }
    
    // 4. 关闭依赖连接
    closeDatabase()
    closeRedis()
    closeBackgroundTasks()
    
    // 5. 保存状态
    saveApplicationState()
    
    // 6. 记录关闭日志
    log.Printf("Server gracefully shutdown completed")
}
```

---

## 📊 状态保存机制

### 需保存的状态

| 状态类型 | 保存位置 | 示例 |
|----------|---------|------|
| **后台任务状态** | Redis | 运行中的批处理任务 |
| **WebSocket 连接** | 数据库 | 活跃连接列表 |
| **缓存状态** | Redis | 热点数据缓存 |
| **未完成操作** | 数据库 | 待处理事务 |

### 状态保存代码

```go
// 保存应用状态
func saveApplicationState() {
    // 1. 保存后台任务状态
    for task := range runningTasks {
        SaveTaskState(task.ID, task.Status)
    }
    
    // 2. 保存 WebSocket 连接状态
    SaveActiveConnections()
    
    // 3. 保存缓存热点数据
    SaveHotDataList()
    
    // 4. 记录关闭时间
    RecordShutdownTime()
}
```

---

## 🔄 故障恢复机制

### 启动恢复流程

```
┌─────────────────────────────────────────┐
│  1. 应用启动                              │
│  - 加载配置                              │
│  - 初始化依赖                            │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  2. 检查上次关闭状态                      │
│  - 是否正常关闭                          │
│  - 是否有未完成任务                      │
│  - 是否有待恢复数据                      │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  3. 恢复未完成任务                        │
│  - 恢复后台任务                          │
│  - 恢复 WebSocket 连接                   │
│  - 恢复缓存数据                          │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  4. 健康检查就绪                          │
│  - 数据库连接正常                        │
│  - Redis 连接正常                       │
│  - 返回 readiness ok                    │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  5. 接收流量                              │
│  - 加入 Service Endpoints                │
│  - 开始处理请求                          │
└─────────────────────────────────────────┘
```

### 启动恢复代码

```go
// 应用启动恢复
func StartupRecovery() {
    // 1. 检查上次关闭状态
    lastShutdown := GetLastShutdownState()
    
    if lastShutdown.Status == "abnormal" {
        log.Printf("Last shutdown was abnormal, recovering...")
        
        // 2. 恢复未完成任务
        unfinishedTasks := GetUnfinishedTasks()
        for task := range unfinishedTasks {
            ResumeTask(task)
        }
        
        // 3. 恢复缓存
        WarmupCacheFromSavedState()
        
        // 4. 通知监控
        NotifyRecoveryComplete()
    }
    
    // 5. 清理关闭状态
    ClearShutdownState()
}
```

---

## 🛡️ 关闭超时处理

### 超时策略

| 阶段 | 超时时间 | 处理方式 |
|------|---------|---------|
| **HTTP 请求等待** | 20 秒 | 超时后强制关闭 |
| **WebSocket 关闭** | 10 秒 | 主动发送关闭消息 |
| **后台任务等待** | 15 秒 | 保存状态后退出 |
| **依赖连接关闭** | 5 秒 | 立即关闭 |

### 超时监控

```go
// 关闭超时监控
func ShutdownWithTimeout(server *http.Server) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 启动监控 Goroutine
    go func() {
        time.Sleep(25 * time.Second)
        log.Printf("Shutdown approaching timeout, forcing closure...")
    }()
    
    server.Shutdown(ctx)
}
```

---

## ✅ 优雅关闭验收

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **信号处理** | 捕获 SIGTERM/SIGINT | 手动发送信号 |
| **请求等待** | 等待现有请求完成 | 请求计数监控 |
| **连接清理** | 关闭所有连接 | 连接计数监控 |
| **状态保存** | 保存未完成任务 | Redis 检查 |
| **关闭日志** | 记录关闭过程 | 日志检查 |
| **恢复时间** | <30s | 启动计时 |

---

**最后更新**: 2026-05-13  
**审核人**: DevOps Team