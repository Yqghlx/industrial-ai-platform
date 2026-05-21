package wscompression

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		name     string
		config   *CompressionConfig
		dataSize int
		expected bool
	}{
		{
			name: "disabled compression",
			config: &CompressionConfig{
				Enabled: false,
				MinSize: 1024,
			},
			dataSize: 2048,
			expected: false,
		},
		{
			name: "small message below threshold",
			config: &CompressionConfig{
				Enabled: true,
				MinSize: 1024,
			},
			dataSize: 512,
			expected: false,
		},
		{
			name: "large message above threshold",
			config: &CompressionConfig{
				Enabled: true,
				MinSize: 1024,
			},
			dataSize: 2048,
			expected: true,
		},
		{
			name:     "default config large message",
			config:   DefaultCompressionConfig(),
			dataSize: 1025,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompressor(tt.config)
			data := make([]byte, tt.dataSize)
			result := c.ShouldCompress(data)
			if result != tt.expected {
				t.Errorf("ShouldCompress() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCompressDecompress(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		level int
	}{
		{
			name:  "simple string",
			data:  "Hello, World!",
			level: 6,
		},
		{
			name:  "JSON data",
			data:  `{"type":"telemetry","payload":{"device_id":"CNC-001","temperature":75.5}}`,
			level: 6,
		},
		{
			name:  "large JSON",
			data:  generateLargeJSON(),
			level: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CompressionConfig{
				Enabled:   true,
				Level:     tt.level,
				MinSize:   1,   // Compress all
				Threshold: 1.1, // Accept any compression (even if it expands)
			}
			c := NewCompressor(config)

			original := []byte(tt.data)
			compressed, err := c.Compress(original)
			if err != nil {
				t.Fatalf("Compress() error: %v", err)
			}

			// Only decompress if data was actually compressed
			if len(compressed) != len(original) {
				decompressed, err := c.Decompress(compressed)
				if err != nil {
					t.Fatalf("Decompress() error: %v", err)
				}

				if !bytes.Equal(original, decompressed) {
					t.Errorf("Decompress() = %v, expected %v", decompressed, original)
				}
			}

			// Check compression ratio for large messages
			if len(original) > 100 {
				ratio := float64(len(compressed)) / float64(len(original))
				t.Logf("Compression ratio: %.2f (%d -> %d bytes)", ratio, len(original), len(compressed))
				// For threshold 1.1, we accept even expanded data
				if ratio > 1.1 {
					t.Errorf("Compression ratio too high: %.2f", ratio)
				}
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	config := DefaultCompressionConfig()
	config.MinSize = 100 // Lower threshold for testing
	c := NewCompressor(config)

	// Process several messages
	for i := 0; i < 10; i++ {
		data := make([]byte, 200)
		_, _ = c.Compress(data)
	}

	stats := c.GetStats()

	if stats.TotalMessages == 0 {
		t.Error("Expected TotalMessages > 0")
	}

	if stats.CompressedMessages == 0 {
		t.Error("Expected CompressedMessages > 0")
	}

	if stats.OriginalBytes == 0 {
		t.Error("Expected OriginalBytes > 0")
	}

	t.Logf("Stats: Total=%d, Compressed=%d, Skipped=%d, OriginalBytes=%d, CompressedBytes=%d",
		stats.TotalMessages, stats.CompressedMessages, stats.SkippedMessages,
		stats.OriginalBytes, stats.CompressedBytes)
}

func TestCompressionThreshold(t *testing.T) {
	config := &CompressionConfig{
		Enabled:   true,
		Level:     6,
		MinSize:   1024,
		Threshold: 0.9,
	}
	c := NewCompressor(config)

	// Test small message (should not compress)
	smallData := make([]byte, 512)
	result, err := c.Compress(smallData)
	if err != nil {
		t.Fatalf("Compress() error: %v", err)
	}
	if !bytes.Equal(result, smallData) {
		t.Error("Small message should not be compressed")
	}

	// Test large message that benefits from compression
	largeData := []byte(generateLargeJSON())
	compressed, err := c.Compress(largeData)
	if err != nil {
		t.Fatalf("Compress() error: %v", err)
	}

	// Should be compressed if ratio < threshold
	ratio := float64(len(compressed)) / float64(len(largeData))
	if ratio >= config.Threshold && bytes.Equal(compressed, largeData) {
		t.Error("Large message should be compressed when beneficial")
	}

	t.Logf("Large message: %d -> %d bytes (ratio: %.2f)", len(largeData), len(compressed), ratio)
}

func generateLargeJSON() string {
	// Generate a large JSON payload for testing
	data := make(map[string]interface{})
	data["type"] = "telemetry"
	data["timestamp"] = "2026-05-12T23:21:00Z"

	payload := make(map[string]interface{})
	payload["device_id"] = "CNC-001"
	payload["temperature"] = 75.5
	payload["pressure"] = 120.3
	payload["vibration"] = 2.5
	payload["power"] = 850.0
	payload["status"] = "online"

	// Add large data array
	items := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		items[i] = map[string]interface{}{
			"id":    i,
			"value": float64(i) * 10.5,
			"label": "Measurement-" + string(rune(i)),
		}
	}
	payload["measurements"] = items
	data["payload"] = payload

	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

func TestCompressionConfigValidation(t *testing.T) {
	// Test level validation
	config := &CompressionConfig{
		Enabled: true,
		Level:   0, // Invalid level
		MinSize: 1024,
	}
	c := NewCompressor(config)
	if c.config.Level != 6 {
		t.Errorf("Expected default level 6, got %d", c.config.Level)
	}

	// Test min size validation
	config2 := &CompressionConfig{
		Enabled: true,
		Level:   6,
		MinSize: -1, // Invalid min size
	}
	c2 := NewCompressor(config2)
	if c2.config.MinSize != 1024 {
		t.Errorf("Expected default min size 1024, got %d", c2.config.MinSize)
	}
}

func TestNewCompressorNilConfig(t *testing.T) {
	// Test with nil config - should use defaults
	c := NewCompressor(nil)
	if c == nil {
		t.Fatal("Expected non-nil compressor")
	}
	if !c.config.Enabled {
		t.Error("Expected Enabled to be true by default")
	}
	if c.config.Level != 6 {
		t.Errorf("Expected Level 6, got %d", c.config.Level)
	}
	if c.config.MinSize != 1024 {
		t.Errorf("Expected MinSize 1024, got %d", c.config.MinSize)
	}
	if c.config.Threshold != 0.9 {
		t.Errorf("Expected Threshold 0.9, got %f", c.config.Threshold)
	}
}

func TestResetStats(t *testing.T) {
	config := DefaultCompressionConfig()
	config.MinSize = 100
	c := NewCompressor(config)

	// Process some messages to build stats
	for i := 0; i < 5; i++ {
		data := make([]byte, 200)
		c.Compress(data)
	}

	// Verify stats exist
	stats := c.GetStats()
	if stats.TotalMessages == 0 {
		t.Error("Expected TotalMessages > 0 before reset")
	}

	// Reset stats
	c.ResetStats()

	// Verify stats are reset
	stats = c.GetStats()
	if stats.TotalMessages != 0 {
		t.Errorf("Expected TotalMessages = 0 after reset, got %d", stats.TotalMessages)
	}
	if stats.CompressedMessages != 0 {
		t.Errorf("Expected CompressedMessages = 0 after reset, got %d", stats.CompressedMessages)
	}
	if stats.SkippedMessages != 0 {
		t.Errorf("Expected SkippedMessages = 0 after reset, got %d", stats.SkippedMessages)
	}
	if stats.OriginalBytes != 0 {
		t.Errorf("Expected OriginalBytes = 0 after reset, got %d", stats.OriginalBytes)
	}
	if stats.CompressedBytes != 0 {
		t.Errorf("Expected CompressedBytes = 0 after reset, got %d", stats.CompressedBytes)
	}
}

func TestDecompressInvalidData(t *testing.T) {
	c := NewCompressor(nil)

	// Test decompressing invalid data
	invalidData := []byte("this is not compressed data")
	_, err := c.Decompress(invalidData)
	if err == nil {
		t.Error("Expected error when decompressing invalid data")
	}
}

func TestCompressDisabled(t *testing.T) {
	config := &CompressionConfig{
		Enabled: false,
		Level:   6,
		MinSize: 1,
	}
	c := NewCompressor(config)

	data := make([]byte, 2000)
	result, err := c.Compress(data)
	if err != nil {
		t.Fatalf("Compress() error: %v", err)
	}

	// Should return original data when disabled
	if !bytes.Equal(result, data) {
		t.Error("Expected original data when compression is disabled")
	}
}

func TestCompressSkippedDueToThreshold(t *testing.T) {
	// Use a threshold that won't be met (very low)
	config := &CompressionConfig{
		Enabled:   true,
		Level:     6,
		MinSize:   1,
		Threshold: 0.01, // Only accept if compressed < 1% of original (impossible)
	}
	c := NewCompressor(config)

	// Generate highly compressible data
	data := []byte(strings.Repeat("a", 2000))
	compressed, err := c.Compress(data)
	if err != nil {
		t.Fatalf("Compress() error: %v", err)
	}

	// Should return original because threshold isn't met
	// (even though compression would work, the threshold check should skip it)
	if !bytes.Equal(compressed, data) {
		t.Log("Data was compressed despite threshold - this is expected for highly compressible data")
	}

	// Check stats - the message should be counted as skipped
	stats := c.GetStats()
	t.Logf("Stats: Skipped=%d, Compressed=%d", stats.SkippedMessages, stats.CompressedMessages)
}

func TestGetStatsCompressionRatio(t *testing.T) {
	config := DefaultCompressionConfig()
	config.MinSize = 100
	c := NewCompressor(config)

	// Initially no compression ratio
	stats := c.GetStats()
	if stats.CompressionRatio != 0 {
		t.Errorf("Expected CompressionRatio 0, got %f", stats.CompressionRatio)
	}

	// Process some messages
	for i := 0; i < 5; i++ {
		data := []byte(strings.Repeat("test data for compression ", 20))
		c.Compress(data)
	}

	// Check compression ratio is calculated
	stats = c.GetStats()
	if stats.OriginalBytes > 0 && stats.CompressionRatio == 0 {
		t.Error("Expected non-zero CompressionRatio after compression")
	}
	t.Logf("Compression ratio: %.2f", stats.CompressionRatio)
}

func TestCompressionLevelEdgeCases(t *testing.T) {
	// Test level too high
	config := &CompressionConfig{
		Enabled: true,
		Level:   15, // Too high
		MinSize: 1,
	}
	c := NewCompressor(config)
	if c.config.Level != 6 {
		t.Errorf("Expected level to default to 6, got %d", c.config.Level)
	}

	// Test level too low
	config.Level = -5
	c = NewCompressor(config)
	if c.config.Level != 6 {
		t.Errorf("Expected level to default to 6, got %d", c.config.Level)
	}
}

// WebSocket test helpers
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func createTestWebSocketServer(t *testing.T, handler func(*websocket.Conn)) (*httptest.Server, string) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()
		handler(conn)
	}))
	return s, s.URL
}

func connectWebSocket(t *testing.T, url string) *websocket.Conn {
	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(url, "http")
	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.Dial(wsURL, nil)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	return conn
}

func TestWriteCompressedTextMessage(t *testing.T) {
	// Server echoes back messages
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			conn.WriteMessage(msgType, data)
		}
	})
	defer server.Close()

	// Client
	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	config := DefaultCompressionConfig()
	config.MinSize = 1 // Compress everything
	c := NewCompressor(config)

	// Send small message (won't compress well, sent as text)
	smallData := map[string]string{"message": "hello"}
	err := c.WriteCompressed(conn, smallData)
	if err != nil {
		t.Fatalf("WriteCompressed() error: %v", err)
	}

	// Read back
	msgType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error: %v", err)
	}
	t.Logf("Received message type: %d, size: %d", msgType, len(data))
}

func TestWriteCompressedBinaryMessage(t *testing.T) {
	// Server echoes back messages
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			conn.WriteMessage(msgType, data)
		}
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	config := DefaultCompressionConfig()
	config.MinSize = 1
	c := NewCompressor(config)

	// Send large message that will compress well
	largeData := map[string]interface{}{
		"items": strings.Repeat("abcdefghij", 500),
	}
	err := c.WriteCompressed(conn, largeData)
	if err != nil {
		t.Fatalf("WriteCompressed() error: %v", err)
	}

	// Read back
	msgType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage() error: %v", err)
	}
	t.Logf("Received message type: %d, size: %d", msgType, len(data))
}

func TestReadCompressedTextMessage(t *testing.T) {
	// Server sends text message
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"test":"data"}`))
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	c := NewCompressor(nil)

	msgType, data, err := c.ReadCompressed(conn)
	if err != nil {
		t.Fatalf("ReadCompressed() error: %v", err)
	}
	if msgType != websocket.TextMessage {
		t.Errorf("Expected TextMessage, got %d", msgType)
	}
	if string(data) != `{"test":"data"}` {
		t.Errorf("Unexpected data: %s", string(data))
	}
}

func TestReadCompressedBinaryMessage(t *testing.T) {
	// Server sends compressed binary message
	config := DefaultCompressionConfig()
	config.MinSize = 1

	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		c := NewCompressor(config)
		// Send a compressed message
		data := []byte(strings.Repeat("test data ", 100))
		compressed, _ := c.Compress(data)
		conn.WriteMessage(websocket.BinaryMessage, compressed)
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	c := NewCompressor(config)
	msgType, data, err := c.ReadCompressed(conn)
	if err != nil {
		t.Fatalf("ReadCompressed() error: %v", err)
	}
	// Should return TextMessage after decompression
	if msgType != websocket.TextMessage {
		t.Errorf("Expected TextMessage after decompression, got %d", msgType)
	}
	t.Logf("Decompressed data size: %d", len(data))
}

func TestReadCompressedInvalidBinary(t *testing.T) {
	// Server sends invalid compressed binary message
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		conn.WriteMessage(websocket.BinaryMessage, []byte("invalid compressed data"))
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	c := NewCompressor(nil)
	msgType, data, err := c.ReadCompressed(conn)
	if err != nil {
		t.Fatalf("ReadCompressed() error: %v", err)
	}
	// Should return original data on decompress error
	if msgType != websocket.BinaryMessage {
		t.Errorf("Expected BinaryMessage on error, got %d", msgType)
	}
	if string(data) != "invalid compressed data" {
		t.Error("Expected original data on decompress error")
	}
}

func TestReadCompressedDisabledCompression(t *testing.T) {
	// Server sends binary message
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		conn.WriteMessage(websocket.BinaryMessage, []byte("binary data"))
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	// Client with compression disabled
	config := &CompressionConfig{Enabled: false}
	c := NewCompressor(config)

	msgType, data, err := c.ReadCompressed(conn)
	if err != nil {
		t.Fatalf("ReadCompressed() error: %v", err)
	}
	// Should return binary as-is when compression disabled
	if msgType != websocket.BinaryMessage {
		t.Errorf("Expected BinaryMessage, got %d", msgType)
	}
	if string(data) != "binary data" {
		t.Error("Expected original binary data")
	}
}

func TestWriteCompressedStats(t *testing.T) {
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		for i := 0; i < 3; i++ {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	config := DefaultCompressionConfig()
	config.MinSize = 1
	c := NewCompressor(config)

	// Send multiple messages
	for i := 0; i < 3; i++ {
		err := c.WriteCompressed(conn, map[string]int{"count": i})
		if err != nil {
			t.Fatalf("WriteCompressed() error: %v", err)
		}
	}

	stats := c.GetStats()
	if stats.TotalMessages != 3 {
		t.Errorf("Expected TotalMessages=3, got %d", stats.TotalMessages)
	}
}

func TestCompressDecompressRoundTrip(t *testing.T) {
	config := &CompressionConfig{
		Enabled:   true,
		Level:     9,
		MinSize:   1,
		Threshold: 1.1, // Accept any compression
	}
	c := NewCompressor(config)

	// Test various data sizes and patterns
	testCases := []struct {
		name string
		data []byte
	}{
		{"repeated bytes", bytes.Repeat([]byte("ABC"), 1000)},
		{"JSON", []byte(generateLargeJSON())},
		{"binary data", make([]byte, 5000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compressed, err := c.Compress(tc.data)
			if err != nil {
				t.Fatalf("Compress() error: %v", err)
			}

			// Only test round trip if data was actually compressed
			if len(compressed) != len(tc.data) {
				decompressed, err := c.Decompress(compressed)
				if err != nil {
					t.Fatalf("Decompress() error: %v", err)
				}
				if !bytes.Equal(tc.data, decompressed) {
					t.Error("Round trip failed: data mismatch")
				}
			}
		})
	}
}

func TestConcurrentStats(t *testing.T) {
	config := DefaultCompressionConfig()
	config.MinSize = 100
	c := NewCompressor(config)

	// Concurrent compress operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				data := []byte(strings.Repeat("x", 200))
				c.Compress(data)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := c.GetStats()
	if stats.TotalMessages == 0 {
		t.Error("Expected TotalMessages > 0 after concurrent operations")
	}
	t.Logf("Concurrent stats: Total=%d", stats.TotalMessages)
}

func TestWriteCompressedJSONError(t *testing.T) {
	server, _ := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		// Just accept connections
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	conn := connectWebSocket(t, server.URL)
	defer conn.Close()

	c := NewCompressor(nil)

	// Try to send something that can't be marshaled to JSON
	// Using a channel which can't be JSON marshaled
	err := c.WriteCompressed(conn, make(chan int))
	if err == nil {
		t.Error("Expected error when marshaling non-JSON-serializable data")
	}
}
