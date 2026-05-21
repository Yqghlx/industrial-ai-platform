package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// FIX-020: 统一 WebSocket 广播器 - 单例模式

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
}

var (
	instance     *Broadcaster
	once         sync.Once
	instanceLock sync.Mutex
)

// GetBroadcaster returns the singleton broadcaster instance
func GetBroadcaster() *Broadcaster {
	once.Do(func() {
		instance = &Broadcaster{
			connections: make(map[*websocket.Conn]bool),
			register:    make(chan *websocket.Conn, 100),
			unregister:  make(chan *websocket.Conn, 100),
			broadcast:   make(chan Message, 1000),
			stopChan:    make(chan struct{}),
		}
		go instance.run()
	})
	return instance
}

// run handles the broadcast loop
func (b *Broadcaster) run() {
	for {
		select {
		case conn := <-b.register:
			b.mu.Lock()
			b.connections[conn] = true
			b.mu.Unlock()
			log.Printf("[Broadcaster] Client connected. Total: %d", len(b.connections))

		case conn := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.connections[conn]; ok {
				delete(b.connections, conn)
				conn.Close()
			}
			b.mu.Unlock()
			log.Printf("[Broadcaster] Client disconnected. Total: %d", len(b.connections))

		case msg := <-b.broadcast:
			b.mu.RLock()
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[Broadcaster] JSON marshal error: %v", err)
				b.mu.RUnlock()
				continue
			}

			for conn := range b.connections {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("[Broadcaster] Write error: %v", err)
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
			log.Println("[Broadcaster] Stopping broadcast loop")
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
func (b *Broadcaster) Stop() {
	close(b.stopChan)
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
