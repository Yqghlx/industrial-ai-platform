// Package benchmarks provides performance benchmark tests for the Industrial AI Platform
// P2-001: WebSocket Connection Performance Benchmarks
// These benchmarks test WebSocket connection and message handling performance
package benchmarks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSMessage represents a WebSocket message for benchmark tests
type WSMessage struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// WebSocket test helpers
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// BenchmarkWebSocketConnection benchmarks WebSocket connection establishment
func BenchmarkWebSocketConnection(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		// Send welcome message
		conn.WriteJSON(WSMessage{
			Type:      "connected",
			Timestamp: time.Now(),
		})
		
		// Read one message then close
		conn.ReadMessage()
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			b.Fatalf("Dial failed: %v", err)
		}
		conn.Close()
	}
}

// BenchmarkWebSocketMessageSend benchmarks sending WebSocket messages
func BenchmarkWebSocketMessageSend(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		// Keep connection alive for benchmark
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	message := WSMessage{
		Type: "telemetry",
		Payload: map[string]interface{}{
			"device_id": "device-001",
			"metrics": map[string]interface{}{
				"temperature": 75.5,
				"vibration":   2.3,
			},
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conn.WriteJSON(message)
		if err != nil {
			b.Fatalf("WriteJSON failed: %v", err)
		}
	}
}

// BenchmarkWebSocketMessageReceive benchmarks receiving WebSocket messages
func BenchmarkWebSocketMessageReceive(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		// Continuously send messages
		message := WSMessage{
			Type:      "telemetry",
			Payload:   map[string]interface{}{"data": "test"},
			Timestamp: time.Now(),
		}
		
		for {
			err := conn.WriteJSON(message)
			if err != nil {
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg WSMessage
		_ = json.Unmarshal(message, &msg)
	}
}

// BenchmarkWebSocketBinaryMessage benchmarks binary message handling
func BenchmarkWebSocketBinaryMessage(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Create binary payload (1KB)
	binaryData := make([]byte, 1024)
	for i := 0; i < 1024; i++ {
		binaryData[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conn.WriteMessage(websocket.BinaryMessage, binaryData)
		if err != nil {
			b.Fatalf("WriteMessage failed: %v", err)
		}
	}
}

// BenchmarkWebSocketConcurrentConnections benchmarks concurrent WebSocket connections
func BenchmarkWebSocketConcurrentConnections(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		conn.WriteJSON(WSMessage{Type: "connected", Timestamp: time.Now()})
		conn.ReadMessage()
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				continue
			}
			conn.WriteJSON(WSMessage{Type: "ping"})
			conn.Close()
		}
	})
}

// WebSocketBroadcaster simulates broadcast functionality for benchmarks
type WebSocketBroadcaster struct {
	connections map[*websocket.Conn]bool
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	broadcast   chan WSMessage
	mu          sync.RWMutex
	stopChan    chan struct{}
}

// NewWebSocketBroadcaster creates a new broadcaster
func NewWebSocketBroadcaster() *WebSocketBroadcaster {
	b := &WebSocketBroadcaster{
		connections: make(map[*websocket.Conn]bool),
		register:    make(chan *websocket.Conn, 100),
		unregister:  make(chan *websocket.Conn, 100),
		broadcast:   make(chan WSMessage, 1000),
		stopChan:    make(chan struct{}),
	}
	go b.run()
	return b
}

func (b *WebSocketBroadcaster) run() {
	for {
		select {
		case conn := <-b.register:
			b.mu.Lock()
			b.connections[conn] = true
			b.mu.Unlock()
		case conn := <-b.unregister:
			b.mu.Lock()
			delete(b.connections, conn)
			b.mu.Unlock()
			conn.Close()
		case msg := <-b.broadcast:
			b.mu.RLock()
			data, _ := json.Marshal(msg)
			for conn := range b.connections {
				conn.WriteMessage(websocket.TextMessage, data)
			}
			b.mu.RUnlock()
		case <-b.stopChan:
			b.mu.Lock()
			for conn := range b.connections {
				conn.Close()
			}
			b.connections = make(map[*websocket.Conn]bool)
			b.mu.Unlock()
			return
		}
	}
}

func (b *WebSocketBroadcaster) Broadcast(msg WSMessage) {
	b.broadcast <- msg
}

func (b *WebSocketBroadcaster) Stop() {
	close(b.stopChan)
}

// BenchmarkWebSocketBroadcast benchmarks broadcasting messages to multiple clients
func BenchmarkWebSocketBroadcast(b *testing.B) {
	broadcaster := NewWebSocketBroadcaster()
	defer broadcaster.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.Broadcast(WSMessage{
			Type: "telemetry",
			Payload: map[string]interface{}{
				"device_id": "device-001",
				"value":     float64(i),
			},
		})
	}
}

// BenchmarkWebSocketTelemetryBroadcast benchmarks telemetry broadcast
func BenchmarkWebSocketTelemetryBroadcast(b *testing.B) {
	broadcaster := NewWebSocketBroadcaster()
	defer broadcaster.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.Broadcast(WSMessage{
			Type: "telemetry",
			Payload: map[string]interface{}{
				"device_id": "device-001",
				"data": map[string]interface{}{
					"temperature": 75.5,
					"vibration":   2.3,
					"pressure":    120.0,
				},
			},
		})
	}
}

// BenchmarkWebSocketAlertBroadcast benchmarks alert broadcast
func BenchmarkWebSocketAlertBroadcast(b *testing.B) {
	broadcaster := NewWebSocketBroadcaster()
	defer broadcaster.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.Broadcast(WSMessage{
			Type: "alert",
			Payload: map[string]interface{}{
				"alert_id":  fmt.Sprintf("alert-%d", i),
				"severity":  "high",
				"message":   "Temperature exceeded threshold",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})
	}
}

// BenchmarkWebSocketJSONMarshal benchmarks JSON marshaling for WebSocket messages
func BenchmarkWebSocketJSONMarshal(b *testing.B) {
	message := WSMessage{
		Type: "telemetry",
		Payload: map[string]interface{}{
			"device_id": "device-001",
			"metrics": map[string]interface{}{
				"temperature": 75.5,
				"vibration":   2.3,
				"pressure":    120.0,
				"power":       550.0,
			},
			"status": "running",
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(message)
	}
}

// BenchmarkWebSocketJSONUnmarshal benchmarks JSON unmarshaling for WebSocket messages
func BenchmarkWebSocketJSONUnmarshal(b *testing.B) {
	message := WSMessage{
		Type: "telemetry",
		Payload: map[string]interface{}{
			"device_id": "device-001",
			"data":      "test",
		},
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg WSMessage
		_ = json.Unmarshal(data, &msg)
	}
}

// BenchmarkWebSocketHeartbeat benchmarks WebSocket heartbeat handling
func BenchmarkWebSocketHeartbeat(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		// Respond to ping messages
		for {
			msgType, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if msgType == websocket.TextMessage {
				conn.WriteJSON(WSMessage{Type: "pong", Timestamp: time.Now()})
			}
		}
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	pingMsg := WSMessage{Type: "ping", Timestamp: time.Now()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WriteJSON(pingMsg)
		_, _, _ = conn.ReadMessage()
	}
}

// BenchmarkWebSocketLargeMessage benchmarks handling large WebSocket messages
func BenchmarkWebSocketLargeMessage(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	server := httptest.NewServer(router)
	defer server.Close()
	
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Create large message payload with many metrics
	largePayload := WSMessage{
		Type: "telemetry",
		Payload: map[string]interface{}{
			"device_id": "device-001",
			"metrics":   generateLargeMetrics(1000),
		},
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(largePayload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			b.Fatalf("WriteMessage failed: %v", err)
		}
	}
}

// generateLargeMetrics creates a large metrics map for testing
func generateLargeMetrics(count int) map[string]interface{} {
	metrics := make(map[string]interface{})
	for i := 0; i < count; i++ {
		metrics[fmt.Sprintf("metric_%d", i)] = float64(i) * 1.5
	}
	return metrics
}

// BenchmarkWebSocketConnectionPool benchmarks connection pool management
func BenchmarkWebSocketConnectionPool(b *testing.B) {
	var connections sync.Map
	var mu sync.Mutex
	count := 0

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		connID := fmt.Sprintf("conn-%d", time.Now().UnixNano())
		
		for pb.Next() {
			// Add connection
			mu.Lock()
			connections.Store(connID, true)
			count++
			mu.Unlock()
			
			// Remove connection
			mu.Lock()
			connections.Delete(connID)
			count--
			mu.Unlock()
		}
	})
}

// BenchmarkWebSocketMessageQueue benchmarks message queue performance
func BenchmarkWebSocketMessageQueue(b *testing.B) {
	queue := make(chan WSMessage, 1000)
	
	// Consumer
	go func() {
		for msg := range queue {
			_ = msg // Process message
		}
	}()

	message := WSMessage{
		Type:      "telemetry",
		Payload:   map[string]interface{}{"data": "test"},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case queue <- message:
		default:
			// Queue full, skip
		}
	}
	
	close(queue)
}

// BenchmarkWebSocketCompression benchmarks WebSocket message compression
func BenchmarkWebSocketCompression(b *testing.B) {
	// Simulate compression overhead
	message := WSMessage{
		Type: "telemetry",
		Payload: map[string]interface{}{
			"device_id": "device-001",
			"metrics":   generateLargeMetrics(100),
		},
		Timestamp: time.Now(),
	}
	
	data, _ := json.Marshal(message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate compression check (skip if < threshold)
		if len(data) > 500 {
			// Simulated compression: just copy for benchmark
			compressed := make([]byte, len(data))
			copy(compressed, data)
		}
	}
}

// BenchmarkWebSocketMultipleConnections benchmarks handling multiple connections
func BenchmarkWebSocketMultipleConnections(b *testing.B) {
	var mu sync.RWMutex
	connections := make(map[string]bool)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connID := fmt.Sprintf("conn-%d", i)
		
		// Register connection
		mu.Lock()
		connections[connID] = true
		mu.Unlock()
		
		// Check connection count
		mu.RLock()
		_ = len(connections)
		mu.RUnlock()
		
		// Unregister connection
		mu.Lock()
		delete(connections, connID)
		mu.Unlock()
	}
}

// BenchmarkWebSocketBroadcastWithClients benchmarks broadcasting with simulated clients
func BenchmarkWebSocketBroadcastWithClients(b *testing.B) {
	broadcaster := NewWebSocketBroadcaster()
	defer broadcaster.Stop()
	
	// Simulate message processing
	var processedCount int64
	var mu sync.Mutex
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := WSMessage{
				Type: "telemetry",
				Payload: map[string]interface{}{
					"device_id": fmt.Sprintf("device-%d", time.Now().UnixNano()%100),
					"value":     75.5,
				},
			}
			broadcaster.Broadcast(msg)
			
			mu.Lock()
			processedCount++
			mu.Unlock()
		}
	})
}

// BenchmarkWebSocketMessageTypeCheck benchmarks message type checking
func BenchmarkWebSocketMessageTypeCheck(b *testing.B) {
	types := []string{"telemetry", "alert", "ping", "pong", "error", "status"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgType := types[i%len(types)]
		
		switch msgType {
		case "telemetry":
			_ = fmt.Sprintf("process telemetry")
		case "alert":
			_ = fmt.Sprintf("process alert")
		case "ping":
			_ = fmt.Sprintf("send pong")
		case "pong":
			_ = fmt.Sprintf("ping received")
		case "error":
			_ = fmt.Sprintf("handle error")
		case "status":
			_ = fmt.Sprintf("update status")
		}
	}
}

// BenchmarkWebSocketRateLimit benchmarks WebSocket rate limiting
func BenchmarkWebSocketRateLimit(b *testing.B) {
	var messageCount int64
	limit := 100 // messages per second
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		messageCount++
		
		// Simulate rate limit check
		if messageCount > limit {
			messageCount = 0 // Reset after limit
		}
	}
}

// BenchmarkWebSocketChannelBuffer benchmarks channel buffer performance
func BenchmarkWebSocketChannelBuffer(b *testing.B) {
	channel := make(chan WSMessage, 100)
	
	go func() {
		for {
			_, ok := <-channel
			if !ok {
				return
			}
		}
	}()
	
	msg := WSMessage{Type: "test", Timestamp: time.Now()}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case channel <- msg:
		default:
			// Channel full, skip
		}
	}
	
	close(channel)
}