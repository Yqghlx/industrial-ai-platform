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

// FIX-058: Server 结构体拆分
// 将 WebSocket 相关处理逻辑从 server.go 拆分到此文件
// WebSocket 处理逻辑独立，便于维护和扩展

// WebSocketManager manages WebSocket connections
// FIX-058: 封装 WebSocket 管理逻辑
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	broadcast  chan model.WSMessage
	heartbeat  chan struct{}
	compressor *wscompression.Compressor
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(compressor *wscompression.Compressor) *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan model.WSMessage, 100),
		heartbeat:  make(chan struct{}),
		compressor: compressor,
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
	// Broadcast loop
	go func() {
		for {
			select {
			case msg := <-m.broadcast:
				m.clientsMu.RLock()
				for conn := range m.clients {
					// Use compression for broadcasting messages
					err := m.compressor.WriteCompressed(conn, msg)
					if err != nil {
						logger.L().Error("WebSocket write error", zap.Error(err))
						conn.Close()
						m.clientsMu.RUnlock()
						m.clientsMu.Lock()
						delete(m.clients, conn)
						m.clientsMu.Unlock()
						m.clientsMu.RLock()
					}
				}
				m.clientsMu.RUnlock()
			case <-m.heartbeat:
				// Heartbeat ping (usually small, no compression needed)
				m.clientsMu.RLock()
				for conn := range m.clients {
					// Heartbeat messages are small, use regular JSON
					err := conn.WriteJSON(model.WSMessage{
						Type:      "ping",
						Timestamp: time.Now(),
					})
					if err != nil {
						logger.L().Error("WebSocket ping error", zap.Error(err))
					}
				}
				m.clientsMu.RUnlock()
			}
		}
	}()

	// Heartbeat ticker
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			m.heartbeat <- struct{}{}
		}
	}()
}

// ClientCount returns the number of connected WebSocket clients
func (m *WebSocketManager) ClientCount() int {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()
	return len(m.clients)
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
// FIX-058: 从 server.go 移动到此文件
func (s *Server) startBroadcaster() {
	go func() {
		for {
			select {
			case msg := <-s.broadcastChan:
				s.wsClientsMu.RLock()
				for conn := range s.wsClients {
					// Use compression for broadcasting messages
					err := s.wsCompressor.WriteCompressed(conn, msg)
					if err != nil {
						logger.L().Error("WebSocket write error", "error", err)
						conn.Close()
						s.wsClientsMu.RUnlock()
						s.wsClientsMu.Lock()
						delete(s.wsClients, conn)
						s.wsClientsMu.Unlock()
						s.wsClientsMu.RLock()
					}
				}
				s.wsClientsMu.RUnlock()
			case <-s.heartbeatChan:
				// Heartbeat ping (usually small, no compression needed)
				s.wsClientsMu.RLock()
				for conn := range s.wsClients {
					// Heartbeat messages are small, use regular JSON
					err := conn.WriteJSON(model.WSMessage{
						Type:      "ping",
						Timestamp: time.Now(),
					})
					if err != nil {
						logger.L().Error("WebSocket ping error", "error", err)
					}
				}
				s.wsClientsMu.RUnlock()
			}
		}
	}()

	// Start heartbeat ticker
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			s.heartbeatChan <- struct{}{}
		}
	}()
}

// broadcast sends a message to all WebSocket clients
// FIX-058: 从 server.go 移动到此文件
// broadcast sends a message to all WebSocket clients
// nolint:unused -- API compatibility
func (s *Server) broadcast(msg model.WSMessage) {
	s.broadcastChan <- msg
}

// addWSClient adds a WebSocket client
// FIX-058: 从 server.go 移动到此文件
func (s *Server) addWSClient(conn *websocket.Conn) {
	s.wsClientsMu.Lock()
	s.wsClients[conn] = true
	s.wsClientsMu.Unlock()
}

// removeWSClient removes a WebSocket client
// FIX-058: 从 server.go 移动到此文件
func (s *Server) removeWSClient(conn *websocket.Conn) {
	s.wsClientsMu.Lock()
	delete(s.wsClients, conn)
	s.wsClientsMu.Unlock()
	conn.Close()
}

// getWSCompressionStats handles getting WebSocket compression statistics
// FIX-058: 从 server.go 移动到此文件
// getWSCompressionStats handles WebSocket compression stats
// FIX-058: 从 server.go 移动到此文件
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
