package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// createTestServer creates a test WebSocket server
func createTestServer(t *testing.T) (*httptest.Server, string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Keep connection open and read messages
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	return server, wsURL
}

// connectClient creates a WebSocket client connection
func connectClient(t *testing.T, wsURL string) *websocket.Conn {
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if resp != nil {
		resp.Body.Close()
	}
	assert.NoError(t, err)
	return conn
}

func TestMain(m *testing.M) {
	// Reset broadcaster before tests
	Reset()
	m.Run()
}

func TestGetBroadcaster_Singleton(t *testing.T) {
	Reset()
	defer Reset()

	b1 := GetBroadcaster()
	b2 := GetBroadcaster()

	assert.NotNil(t, b1)
	assert.NotNil(t, b2)
	assert.Same(t, b1, b2, "GetBroadcaster should return the same instance")
}

func TestBroadcaster_RegisterUnregister(t *testing.T) {
	Reset()
	defer Reset()

	server, wsURL := createTestServer(t)
	defer server.Close()

	b := GetBroadcaster()

	// Create client connection
	conn := connectClient(t, wsURL)
	defer conn.Close()

	// Register connection
	b.Register(conn)

	// Wait for registration to process
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, b.ConnectionCount())

	// Unregister connection
	b.Unregister(conn)

	// Wait for unregistration to process
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, b.ConnectionCount())
}

func TestBroadcaster_MultipleConnections(t *testing.T) {
	Reset()
	defer Reset()

	server, wsURL := createTestServer(t)
	defer server.Close()

	b := GetBroadcaster()

	// Connect multiple clients
	var connections []*websocket.Conn
	for i := 0; i < 3; i++ {
		conn := connectClient(t, wsURL)
		connections = append(connections, conn)
		b.Register(conn)
	}

	// Wait for registrations
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 3, b.ConnectionCount())

	// Cleanup
	for _, conn := range connections {
		conn.Close()
	}
}

func TestBroadcaster_Broadcast(t *testing.T) {
	Reset()
	defer Reset()

	// Create a server that echoes messages to a channel for testing
	msgChan := make(chan []byte, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read messages and send to channel
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case msgChan <- msg:
			default:
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	b := GetBroadcaster()

	// Connect client
	conn := connectClient(t, wsURL)
	defer conn.Close()

	b.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Broadcast a message
	testMsg := Message{
		Type:    "test",
		Payload: map[string]string{"hello": "world"},
	}
	b.Broadcast(testMsg)

	// Wait for message to be received
	select {
	case msg := <-msgChan:
		var received Message
		err := json.Unmarshal(msg, &received)
		assert.NoError(t, err)
		assert.Equal(t, "test", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("Expected to receive broadcast message")
	}
}

func TestBroadcaster_BroadcastTelemetry(t *testing.T) {
	Reset()
	defer Reset()

	msgChan := make(chan []byte, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case msgChan <- msg:
			default:
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	b := GetBroadcaster()

	conn := connectClient(t, wsURL)
	defer conn.Close()

	b.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Broadcast telemetry
	b.BroadcastTelemetry("device-001", map[string]float64{"temperature": 25.5})

	select {
	case msg := <-msgChan:
		var received Message
		err := json.Unmarshal(msg, &received)
		assert.NoError(t, err)
		assert.Equal(t, "telemetry", received.Type)

		payload, ok := received.Payload.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "device-001", payload["device_id"])
	case <-time.After(1 * time.Second):
		t.Fatal("Expected to receive telemetry message")
	}
}

func TestBroadcaster_BroadcastAlert(t *testing.T) {
	Reset()
	defer Reset()

	msgChan := make(chan []byte, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case msgChan <- msg:
			default:
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	b := GetBroadcaster()

	conn := connectClient(t, wsURL)
	defer conn.Close()

	b.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Broadcast alert
	b.BroadcastAlert("alert-001", "critical", "Temperature exceeded threshold")

	select {
	case msg := <-msgChan:
		var received Message
		err := json.Unmarshal(msg, &received)
		assert.NoError(t, err)
		assert.Equal(t, "alert", received.Type)

		payload, ok := received.Payload.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "alert-001", payload["alert_id"])
		assert.Equal(t, "critical", payload["severity"])
		assert.Equal(t, "Temperature exceeded threshold", payload["message"])
		assert.NotEmpty(t, payload["timestamp"])
	case <-time.After(1 * time.Second):
		t.Fatal("Expected to receive alert message")
	}
}

func TestBroadcaster_BroadcastDeviceUpdate(t *testing.T) {
	Reset()
	defer Reset()

	msgChan := make(chan []byte, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case msgChan <- msg:
			default:
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	b := GetBroadcaster()

	conn := connectClient(t, wsURL)
	defer conn.Close()

	b.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Broadcast device update
	b.BroadcastDeviceUpdate("device-001", "online")

	select {
	case msg := <-msgChan:
		var received Message
		err := json.Unmarshal(msg, &received)
		assert.NoError(t, err)
		assert.Equal(t, "device_update", received.Type)

		payload, ok := received.Payload.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "device-001", payload["device_id"])
		assert.Equal(t, "online", payload["status"])
		assert.NotEmpty(t, payload["timestamp"])
	case <-time.After(1 * time.Second):
		t.Fatal("Expected to receive device update message")
	}
}

func TestBroadcaster_ConnectionCount(t *testing.T) {
	Reset()
	defer Reset()

	b := GetBroadcaster()
	assert.Equal(t, 0, b.ConnectionCount())

	server, wsURL := createTestServer(t)
	defer server.Close()

	conn := connectClient(t, wsURL)
	defer conn.Close()

	b.Register(conn)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, b.ConnectionCount())
}

func TestBroadcaster_Stop(t *testing.T) {
	Reset()

	server, wsURL := createTestServer(t)
	defer server.Close()

	b := GetBroadcaster()

	conn := connectClient(t, wsURL)
	defer conn.Close()

	b.Register(conn)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, b.ConnectionCount())

	// Stop the broadcaster
	b.Stop()

	// Wait for stop to process
	time.Sleep(50 * time.Millisecond)

	// Connection should be cleared
	assert.Equal(t, 0, b.ConnectionCount())

	// Reset the once so next GetBroadcaster creates new instance
	once = sync.Once{}
	instance = nil
}

func TestBroadcaster_UnregisterNonExistent(t *testing.T) {
	Reset()
	defer Reset()

	b := GetBroadcaster()

	server, wsURL := createTestServer(t)
	defer server.Close()

	conn := connectClient(t, wsURL)
	defer conn.Close()

	// Unregister a connection that wasn't registered
	b.Unregister(conn)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Should not panic, count should be 0
	assert.Equal(t, 0, b.ConnectionCount())
}

func TestReset(t *testing.T) {
	Reset()

	b1 := GetBroadcaster()
	assert.NotNil(t, b1)

	Reset()

	b2 := GetBroadcaster()
	assert.NotNil(t, b2)
	// b1 and b2 should be different instances after reset
	assert.NotSame(t, b1, b2)

	Reset()
}

func TestGetCurrentTimestamp(t *testing.T) {
	ts := getCurrentTimestamp()
	assert.NotEmpty(t, ts)

	// Verify it's a valid RFC3339 timestamp
	_, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err)
}

func TestBroadcaster_BroadcastNoConnections(t *testing.T) {
	Reset()
	defer Reset()

	b := GetBroadcaster()

	// Broadcast when no connections - should not panic
	b.Broadcast(Message{
		Type:    "test",
		Payload: "test",
	})

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Should complete without error
	assert.Equal(t, 0, b.ConnectionCount())
}

func TestBroadcaster_ConcurrentOperations(t *testing.T) {
	Reset()
	defer Reset()

	server, wsURL := createTestServer(t)
	defer server.Close()

	b := GetBroadcaster()

	var wg sync.WaitGroup
	numOps := 10

	// Concurrent registrations
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := connectClient(t, wsURL)
			defer conn.Close()
			b.Register(conn)
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, numOps, b.ConnectionCount())

	// Concurrent broadcasts
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b.Broadcast(Message{
				Type:    "test",
				Payload: i,
			})
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Concurrent unregistrations
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Unregister(nil) // Unregister nil should not panic
		}()
	}

	wg.Wait()
}

func TestBroadcaster_MessageJSONMarshal(t *testing.T) {
	// Test Message marshaling
	msg := Message{
		Type: "test",
		Payload: map[string]interface{}{
			"key": "value",
			"num": 123,
		},
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"type":"test"`)
	assert.Contains(t, string(data), `"key":"value"`)

	// Unmarshal and verify
	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, msg.Type, unmarshaled.Type)
}
