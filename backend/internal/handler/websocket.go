package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/logger"
	"github.com/industrial-ai/platform/pkg/wscompression"
	"go.uber.org/zap"
)

// wsWriteTimeout WebSocket 写操作超时时间
const wsWriteTimeout = 10 * time.Second

// writeWithDeadline 在写超时保护下执行写操作，成功后清除超时
func writeWithDeadline(conn *websocket.Conn, writeFn func() error) error {
	conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	if err := writeFn(); err != nil {
		return err
	}
	conn.SetWriteDeadline(time.Time{})
	return nil
}

// removeConnFromMap 安全地从连接集合中删除并关闭连接（需在 RLock 内调用，函数内部处理锁升级）
func removeConnFromMap(conn *websocket.Conn, clients map[*websocket.Conn]bool, mu *sync.RWMutex) {
	conn.Close()
	mu.RUnlock()
	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
	mu.RLock()
}

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	clients       map[*websocket.Conn]bool
	clientsMu     sync.RWMutex
	broadcast     chan model.WSMessage
	heartbeat     chan struct{}
	compressor    *wscompression.Compressor
	stopHeartbeat chan struct{} // 心跳 ticker 退出信号
	stopBroadcast chan struct{} // 广播循环退出信号
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(compressor *wscompression.Compressor) *WebSocketManager {
	return &WebSocketManager{
		clients:       make(map[*websocket.Conn]bool),
		broadcast:     make(chan model.WSMessage, 100),
		heartbeat:     make(chan struct{}),
		compressor:    compressor,
		stopHeartbeat: make(chan struct{}),
		stopBroadcast: make(chan struct{}),
	}
}

// AddClient adds a WebSocket client
func (m *WebSocketManager) AddClient(conn *websocket.Conn) {
	m.clientsMu.Lock()
	m.clients[conn] = true
	m.clientsMu.Unlock()
}

// RemoveClient removes a WebSocket client
func (m *WebSocketManager) RemoveClient(conn *websocket.Conn) {
	m.clientsMu.Lock()
	delete(m.clients, conn)
	m.clientsMu.Unlock()
	conn.Close()
}

// Broadcast sends a message to all WebSocket clients
func (m *WebSocketManager) Broadcast(msg model.WSMessage) {
	m.broadcast <- msg
}

// Start starts the WebSocket broadcast loop
func (m *WebSocketManager) Start() {
	go func() {
		for {
			select {
			case msg := <-m.broadcast:
				m.clientsMu.RLock()
				for conn := range m.clients {
					err := writeWithDeadline(conn, func() error {
						return m.compressor.WriteCompressed(conn, msg)
					})
					if err != nil {
						logger.L().Error("WebSocket write error", zap.Error(err))
						removeConnFromMap(conn, m.clients, &m.clientsMu)
					}
				}
				m.clientsMu.RUnlock()
			case <-m.heartbeat:
				m.clientsMu.RLock()
				for conn := range m.clients {
					err := writeWithDeadline(conn, func() error {
						return conn.WriteJSON(model.WSMessage{
							Type:      "ping",
							Timestamp: time.Now(),
						})
					})
					if err != nil {
						logger.L().Error("WebSocket ping error", zap.Error(err))
						removeConnFromMap(conn, m.clients, &m.clientsMu)
					}
				}
				m.clientsMu.RUnlock()
			case <-m.stopBroadcast:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.heartbeat <- struct{}{}
			case <-m.stopHeartbeat:
				return
			}
		}
	}()
}

// ClientCount returns the number of connected WebSocket clients
func (m *WebSocketManager) ClientCount() int {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()
	return len(m.clients)
}

// Stop gracefully stops the WebSocket manager and releases resources
// P0: Added to prevent goroutine leak
func (m *WebSocketManager) Stop() {
	close(m.stopBroadcast)
	close(m.stopHeartbeat)
}

// handleWebSocket handles WebSocket connections
// FIX-058: 从 server.go 移动到此文件
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.L().Error("WebSocket upgrade error", zap.Error(err))
		return
	}

	s.addWSClient(conn)
	defer s.removeWSClient(conn)

	// Send initial connection message (compression not needed for small messages)
	conn.WriteJSON(model.WSMessage{
		Type:      "connected",
		Payload:   map[string]string{"message": "WebSocket connected", "compression": fmt.Sprintf("%v", s.wsCompressor != nil)},
		Timestamp: time.Now(),
	})

	// Read messages from client
	for {
		// Read compressed or uncompressed message
		messageType, data, err := s.wsCompressor.ReadCompressed(conn)
		if err != nil {
			logger.L().Error("WebSocket read error", zap.Error(err))
			break
		}

		// Process the message (currently just reading, no specific handling)
		// If needed, parse and handle client messages
		if messageType == websocket.TextMessage && len(data) > 0 {
			// 仅记录消息类型和大小，不记录内容（避免敏感信息泄露）
			logger.L().Info("WebSocket message received", zap.Int("type", messageType), zap.Int("size", len(data)))
		}
	}
}

// startBroadcaster starts the WebSocket broadcast loop
func (s *Server) startBroadcaster() {
	go func() {
		for {
			select {
			case msg := <-s.broadcastChan:
				s.wsClientsMu.RLock()
				for conn := range s.wsClients {
					err := writeWithDeadline(conn, func() error {
						return s.wsCompressor.WriteCompressed(conn, msg)
					})
					if err != nil {
						logger.L().Error("WebSocket write error", zap.Error(err))
						removeConnFromMap(conn, s.wsClients, &s.wsClientsMu)
					}
				}
				s.wsClientsMu.RUnlock()
			case <-s.heartbeatChan:
				s.wsClientsMu.RLock()
				for conn := range s.wsClients {
					err := writeWithDeadline(conn, func() error {
						return conn.WriteJSON(model.WSMessage{
							Type:      "ping",
							Timestamp: time.Now(),
						})
					})
					if err != nil {
						logger.L().Error("WebSocket ping error", zap.Error(err))
						removeConnFromMap(conn, s.wsClients, &s.wsClientsMu)
					}
				}
				s.wsClientsMu.RUnlock()
			case <-s.stopTicker:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.heartbeatChan <- struct{}{}
			case <-s.stopTicker:
				return
			}
		}
	}()
}

// broadcast sends a message to all WebSocket clients
// nolint:unused -- API compatibility
func (s *Server) broadcast(msg model.WSMessage) {
	s.broadcastChan <- msg
}

// addWSClient adds a WebSocket client
func (s *Server) addWSClient(conn *websocket.Conn) {
	s.wsClientsMu.Lock()
	s.wsClients[conn] = true
	s.wsClientsMu.Unlock()
}

// removeWSClient removes a WebSocket client
func (s *Server) removeWSClient(conn *websocket.Conn) {
	s.wsClientsMu.Lock()
	delete(s.wsClients, conn)
	s.wsClientsMu.Unlock()
	conn.Close()
}

// getWSCompressionStats handles getting WebSocket compression statistics
func (s *Server) getWSCompressionStats(c *gin.Context) { // nolint:unused
	if s.wsCompressor == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "WebSocket compression not initialized",
		})
		return
	}

	stats := s.wsCompressor.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"enabled":             true,
		"total_messages":      stats.TotalMessages,
		"compressed_messages": stats.CompressedMessages,
		"skipped_messages":    stats.SkippedMessages,
		"original_bytes":      stats.OriginalBytes,
		"compressed_bytes":    stats.CompressedBytes,
		"compression_ratio":   stats.CompressionRatio,
		"savings_percent":     (1 - stats.CompressionRatio) * 100,
	})
}
