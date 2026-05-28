package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// FIX-020: 统一 WebSocket 广播器 - 单例模式
// BE-P2-07: 添加 Context 生命周期管理

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Broadcaster manages WebSocket connections and broadcasts messages
type Broadcaster struct {
	connections map[*websocket.Conn]bool
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	broadcast   chan Message
	mu          sync.RWMutex
	stopChan    chan struct{}

	// BE-P2-07: Context 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var (
	instance     *Broadcaster
	once         sync.Once
	instanceLock sync.Mutex
)

// GetBroadcaster returns the singleton broadcaster instance
func GetBroadcaster() *Broadcaster {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		instance = &Broadcaster{
			connections: make(map[*websocket.Conn]bool),
			register:    make(chan *websocket.Conn, 100),
			unregister:  make(chan *websocket.Conn, 100),
			broadcast:   make(chan Message, 1000),
			stopChan:    make(chan struct{}),
			ctx:         ctx,
			cancel:      cancel,
		}
		instance.wg.Add(1)
		go instance.run()
	})
	return instance
}

// run handles the broadcast loop
// BE-P2-07: 使用 Context 控制生命周期
func (b *Broadcaster) run() {
	defer b.wg.Done()

	for {
		select {
		case conn := <-b.register:
			b.mu.Lock()
			b.connections[conn] = true
			b.mu.Unlock()
			logger.L().Info("Client connected", zap.Int("total", len(b.connections)))

		case conn := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.connections[conn]; ok {
				delete(b.connections, conn)
				conn.Close()
			}
			b.mu.Unlock()
			logger.L().Info("Client disconnected", zap.Int("total", len(b.connections)))

		case msg := <-b.broadcast:
			b.mu.RLock()
			data, err := json.Marshal(msg)
			if err != nil {
				logger.L().Error("JSON marshal error", zap.Error(err))
				b.mu.RUnlock()
				continue
			}

			for conn := range b.connections {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logger.L().Error("Write error", zap.Error(err))
					conn.Close()
					b.mu.RUnlock()
					b.mu.Lock()
					delete(b.connections, conn)
					b.mu.Unlock()
					b.mu.RLock()
				}
			}
			b.mu.RUnlock()

		case <-b.stopChan:
			logger.L().Info("Stopping broadcast loop via stopChan")
			b.closeAllConnections()
			return

		case <-b.ctx.Done():
			// BE-P2-07: Context 取消时优雅退出
			logger.L().Info("Stopping broadcast loop via context")
			b.closeAllConnections()
			return
		}
	}
}

// closeAllConnections 关闭所有连接
// BE-P2-07: 提取公共逻辑
func (b *Broadcaster) closeAllConnections() {
	b.mu.Lock()
	for conn := range b.connections {
		conn.Close()
	}
	b.connections = make(map[*websocket.Conn]bool)
	b.mu.Unlock()
}

// Register adds a new WebSocket connection
func (b *Broadcaster) Register(conn *websocket.Conn) {
	b.register <- conn
}

// Unregister removes a WebSocket connection
func (b *Broadcaster) Unregister(conn *websocket.Conn) {
	b.unregister <- conn
}

// Broadcast sends a message to all connected clients
func (b *Broadcaster) Broadcast(msg Message) {
	b.broadcast <- msg
}

// BroadcastTelemetry sends telemetry data to all clients
func (b *Broadcaster) BroadcastTelemetry(deviceID string, data interface{}) {
	b.Broadcast(Message{
		Type: "telemetry",
		Payload: map[string]interface{}{
			"device_id": deviceID,
			"data":      data,
		},
	})
}

// BroadcastAlert sends an alert to all clients
func (b *Broadcaster) BroadcastAlert(alertID string, severity string, message string) {
	b.Broadcast(Message{
		Type: "alert",
		Payload: map[string]interface{}{
			"alert_id":  alertID,
			"severity":  severity,
			"message":   message,
			"timestamp": getCurrentTimestamp(),
		},
	})
}

// BroadcastDeviceUpdate sends device status update to all clients
func (b *Broadcaster) BroadcastDeviceUpdate(deviceID string, status string) {
	b.Broadcast(Message{
		Type: "device_update",
		Payload: map[string]interface{}{
			"device_id": deviceID,
			"status":    status,
			"timestamp": getCurrentTimestamp(),
		},
	})
}

// ConnectionCount returns the number of active connections
func (b *Broadcaster) ConnectionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.connections)
}

// Stop stops the broadcaster
// BE-P2-07: 添加 Context 取消确保优雅退出
func (b *Broadcaster) Stop() {
	// 先取消 context
	if b.cancel != nil {
		b.cancel()
	}
	close(b.stopChan)
	b.wg.Wait()
}

// Shutdown 优雅关闭（带超时）
// BE-P2-07: 新增带超时的关闭方法
func (b *Broadcaster) Shutdown(ctx context.Context) error {
	// 取消 context
	if b.cancel != nil {
		b.cancel()
	}
	close(b.stopChan)

	// 等待 goroutine 退出或超时
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.L().Info("Broadcaster shutdown completed")
		return nil
	case <-ctx.Done():
		logger.L().Warn("Broadcaster shutdown timeout")
		return ctx.Err()
	}
}

// Reset resets the singleton (for testing)
func Reset() {
	instanceLock.Lock()
	defer instanceLock.Unlock()
	if instance != nil {
		instance.Stop()
	}
	instance = nil
	once = sync.Once{}
}

func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
