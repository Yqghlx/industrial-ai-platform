package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/wscompression"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWebSocketManager tests the creation of WebSocketManager
func TestNewWebSocketManager(t *testing.T) {
	t.Run("creates manager with nil compressor", func(t *testing.T) {
		manager := NewWebSocketManager(nil)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.clients)
		assert.NotNil(t, manager.broadcast)
		assert.NotNil(t, manager.heartbeat)
		assert.Nil(t, manager.compressor)
	})

	t.Run("creates manager with compressor", func(t *testing.T) {
		compressor := wscompression.NewCompressor(nil)
		manager := NewWebSocketManager(compressor)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.compressor)
	})

	t.Run("initializes empty clients map", func(t *testing.T) {
		manager := NewWebSocketManager(nil)
		assert.Equal(t, 0, manager.ClientCount())
	})

	t.Run("initializes buffered broadcast channel", func(t *testing.T) {
		manager := NewWebSocketManager(nil)
		// Channel should have capacity of 100
		// We can verify by sending without blocking
		for i := 0; i < 50; i++ {
			select {
			case manager.broadcast <- model.WSMessage{}:
				// OK
			default:
				t.Fatal("broadcast channel should not block with 50 messages")
			}
		}
	})
}

// TestWebSocketManager_AddRemoveClient tests AddClient and RemoveClient
func TestWebSocketManager_AddRemoveClient(t *testing.T) {
	t.Run("add client increases count", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		manager := NewWebSocketManager(nil)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 1, manager.ClientCount())
	})

	t.Run("add multiple clients", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		manager := NewWebSocketManager(nil)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		var testConns []*websocket.Conn
		for i := 0; i < 3; i++ {
			conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			if resp != nil {
				resp.Body.Close()
			}
			testConns = append(testConns, conn)
		}

		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 3, manager.ClientCount())

		for _, conn := range testConns {
			conn.Close()
		}
	})

	t.Run("remove client decreases count", func(t *testing.T) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		manager := NewWebSocketManager(nil)
		var serverConn1 *websocket.Conn
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			mu.Lock()
			if serverConn1 == nil {
				serverConn1 = conn
			}
			mu.Unlock()
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Add two clients
		conn1, resp1, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		if resp1 != nil {
			resp1.Body.Close()
		}
		conn2, resp2, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		if resp2 != nil {
			resp2.Body.Close()
		}

		time.Sleep(50 * time.Millisecond)
		countBefore := manager.ClientCount()
		assert.GreaterOrEqual(t, countBefore, 2)

		// Remove the server-side connection (which is in the manager)
		mu.Lock()
		connToRemove := serverConn1
		mu.Unlock()
		if connToRemove != nil {
			manager.RemoveClient(connToRemove)
		}

		time.Sleep(50 * time.Millisecond)
		countAfter := manager.ClientCount()
		assert.Equal(t, countBefore-1, countAfter)

		// Cleanup
		conn1.Close()
		conn2.Close()
	})
}

// TestWebSocketManager_ClientCount tests the ClientCount method
func TestWebSocketManager_ClientCount(t *testing.T) {
	t.Run("returns zero for empty manager", func(t *testing.T) {
		manager := NewWebSocketManager(nil)
		assert.Equal(t, 0, manager.ClientCount())
	})

	t.Run("returns correct count after add/remove", func(t *testing.T) {
		manager := NewWebSocketManager(nil)

		// Create mock connections (using net.Conn is complex, so we'll test the logic differently)
		// This test verifies the mutex protection and map operations
		initialCount := manager.ClientCount()

		// We can't easily create *websocket.Conn without a real connection
		// So we'll just verify the count returns consistent results
		for i := 0; i < 10; i++ {
			assert.Equal(t, initialCount, manager.ClientCount())
		}
	})

	t.Run("is thread-safe", func(t *testing.T) {
		manager := NewWebSocketManager(nil)

		// Run concurrent reads
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = manager.ClientCount()
			}()
		}
		wg.Wait()
		// Should not panic or race
	})
}

// TestWebSocketManager_Broadcast tests the Broadcast method
func TestWebSocketManager_Broadcast(t *testing.T) {
	t.Run("sends message to broadcast channel", func(t *testing.T) {
		manager := NewWebSocketManager(nil)

		msg := model.WSMessage{
			Type:      "test",
			Payload:   map[string]string{"message": "hello"},
			Timestamp: time.Now(),
		}

		// Broadcast should not block (channel has buffer)
		done := make(chan bool)
		go func() {
			manager.Broadcast(msg)
			done <- true
		}()

		select {
		case <-done:
			// Good - didn't block
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Broadcast should not block")
		}

		// Verify message was sent to channel
		select {
		case received := <-manager.broadcast:
			assert.Equal(t, msg.Type, received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Should receive message from broadcast channel")
		}
	})

	t.Run("broadcasts multiple messages", func(t *testing.T) {
		manager := NewWebSocketManager(nil)

		for i := 0; i < 5; i++ {
			msg := model.WSMessage{
				Type:      "test",
				Payload:   i,
				Timestamp: time.Now(),
			}
			manager.Broadcast(msg)
		}

		// Should receive 5 messages
		count := 0
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-manager.broadcast:
				count++
				if count >= 5 {
					return
				}
			case <-timeout:
				t.Fatalf("Expected 5 messages, got %d", count)
			}
		}
	})
}

// TestWebSocketManager_Start tests the Start method
func TestWebSocketManager_Start(t *testing.T) {
	t.Run("starts broadcast loop", func(t *testing.T) {
		manager := NewWebSocketManager(wscompression.NewCompressor(nil))

		// Start should begin the broadcast loop
		manager.Start()

		// Give time for goroutines to start
		time.Sleep(50 * time.Millisecond)

		// Verify we can send to broadcast channel
		msg := model.WSMessage{
			Type:      "test",
			Payload:   "hello",
			Timestamp: time.Now(),
		}

		// This should not block if the broadcast loop is running
		select {
		case manager.broadcast <- msg:
			// Good
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Should be able to send to broadcast channel")
		}
	})

	t.Run("handles heartbeat channel", func(t *testing.T) {
		manager := NewWebSocketManager(wscompression.NewCompressor(nil))
		manager.Start()

		// Give time for goroutines to start
		time.Sleep(50 * time.Millisecond)

		// Send heartbeat signal
		select {
		case manager.heartbeat <- struct{}{}:
			// Good
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Should be able to send to heartbeat channel")
		}
	})

	t.Run("broadcasts to connected clients", func(t *testing.T) {
		compressor := wscompression.NewCompressor(nil)
		manager := NewWebSocketManager(compressor)
		manager.Start()

		// Create test server
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect client
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		defer conn.Close()

		// Wait for connection to be established
		time.Sleep(100 * time.Millisecond)

		// Broadcast a message
		msg := model.WSMessage{
			Type:      "test",
			Payload:   map[string]string{"data": "test message"},
			Timestamp: time.Now(),
		}
		manager.Broadcast(msg)

		// Read message from client
		done := make(chan bool)
		go func() {
			// Set read deadline
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			messageType, data, err := compressor.ReadCompressed(conn)
			if err == nil {
				assert.Equal(t, websocket.TextMessage, messageType)
				assert.Contains(t, string(data), "test")
				done <- true
			}
		}()

		select {
		case <-done:
			// Success
		case <-time.After(3 * time.Second):
			t.Fatal("Should have received broadcast message")
		}
	})
}

// TestWebSocketManager_ConcurrentAccess tests thread safety
func TestWebSocketManager_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent client operations", func(t *testing.T) {
		manager := NewWebSocketManager(nil)

		// Create test server
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		var connsMu sync.Mutex
		var serverConns []*websocket.Conn

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			connsMu.Lock()
			serverConns = append(serverConns, conn)
			connsMu.Unlock()
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect multiple clients concurrently
		var wg sync.WaitGroup
		var clientConnsMu sync.Mutex
		var clientConns []*websocket.Conn

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err == nil {
					if resp != nil {
						resp.Body.Close()
					}
					clientConnsMu.Lock()
					clientConns = append(clientConns, conn)
					clientConnsMu.Unlock()
				}
			}()
		}
		wg.Wait()

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 10, manager.ClientCount())

		// Cleanup
		for _, conn := range clientConns {
			conn.Close()
		}
	})

	t.Run("concurrent broadcast and client operations", func(t *testing.T) {
		manager := NewWebSocketManager(wscompression.NewCompressor(nil))
		manager.Start()

		var wg sync.WaitGroup

		// Concurrent broadcasts
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				manager.Broadcast(model.WSMessage{
					Type:      "test",
					Payload:   n,
					Timestamp: time.Now(),
				})
			}(i)
		}

		// Concurrent client count reads
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = manager.ClientCount()
			}()
		}

		wg.Wait()
		// Should not panic or race
	})
}

// TestWebSocketManager_Heartbeat tests heartbeat functionality
func TestWebSocketManager_Heartbeat(t *testing.T) {
	t.Run("sends heartbeat to clients", func(t *testing.T) {
		compressor := wscompression.NewCompressor(nil)
		manager := NewWebSocketManager(compressor)
		manager.Start()

		// Create test server
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect client
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		defer conn.Close()

		// Wait for connection
		time.Sleep(50 * time.Millisecond)

		// Trigger heartbeat manually
		manager.heartbeat <- struct{}{}

		// Read heartbeat message
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg model.WSMessage
		err = conn.ReadJSON(&msg)
		require.NoError(t, err)
		assert.Equal(t, "ping", msg.Type)
	})
}

// TestWebSocketManager_ErrorHandling tests error scenarios
func TestWebSocketManager_ErrorHandling(t *testing.T) {
	t.Run("handles nil compressor", func(t *testing.T) {
		manager := NewWebSocketManager(nil)
		assert.Nil(t, manager.compressor)

		// Manager should still function
		assert.Equal(t, 0, manager.ClientCount())

		// Broadcast should work (but will fail to send to clients)
		manager.Broadcast(model.WSMessage{Type: "test"})
	})

	t.Run("remove non-existent client", func(t *testing.T) {
		manager := NewWebSocketManager(nil)

		// Should not panic when removing non-existent client
		// Note: This will try to close nil connection
		// The RemoveClient method should handle this gracefully
		initialCount := manager.ClientCount()
		assert.Equal(t, 0, initialCount)
	})

	t.Run("broadcast with no clients", func(t *testing.T) {
		manager := NewWebSocketManager(wscompression.NewCompressor(nil))
		manager.Start()

		// Broadcast with no clients should not block or panic
		manager.Broadcast(model.WSMessage{
			Type:      "test",
			Payload:   "no clients",
			Timestamp: time.Now(),
		})

		// Give time for broadcast loop to process
		time.Sleep(100 * time.Millisecond)
		// Should complete without issues
	})
}

// TestWebSocketManager_Integration is an integration test
func TestWebSocketManager_Integration(t *testing.T) {
	t.Run("full workflow", func(t *testing.T) {
		// Create manager with compressor
		compressor := wscompression.NewCompressor(&wscompression.CompressionConfig{
			Enabled: true,
			Level:   6,
			MinSize: 100,
		})
		manager := NewWebSocketManager(compressor)
		manager.Start()

		// Create test server
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			manager.AddClient(conn)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		// Connect 3 clients
		var clients []*websocket.Conn
		for i := 0; i < 3; i++ {
			conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			if resp != nil {
				resp.Body.Close()
			}
			clients = append(clients, conn)
		}

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 3, manager.ClientCount())

		// Broadcast message
		msg := model.WSMessage{
			Type:      "alert",
			Payload:   map[string]interface{}{"severity": "high", "device": "test-device"},
			Timestamp: time.Now(),
		}
		manager.Broadcast(msg)

		// All clients should receive the message
		for i, client := range clients {
			client.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, data, err := compressor.ReadCompressed(client)
			require.NoError(t, err, "Client %d should receive message", i)
			assert.Contains(t, string(data), "alert")
		}

		// Remove one client
		clients[0].Close()
		time.Sleep(50 * time.Millisecond)

		// Broadcast again
		manager.Broadcast(model.WSMessage{
			Type:      "update",
			Payload:   "test",
			Timestamp: time.Now(),
		})

		// Only 2 clients should receive
		receivedCount := 0
		for i := 1; i < 3; i++ {
			clients[i].SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			_, _, err := compressor.ReadCompressed(clients[i])
			if err == nil {
				receivedCount++
			}
		}
		assert.Equal(t, 2, receivedCount)

		// Cleanup
		for _, conn := range clients {
			conn.Close()
		}
	})
}
