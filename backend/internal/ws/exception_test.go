package ws

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// Ensure broadcaster is running
}

// ============================================
// P1: WebSocket 异常测试
// ============================================

func TestBroadcaster_ConcurrentBroadcast(t *testing.T) {
	broadcaster := GetBroadcaster()

	// 100 concurrent broadcasts
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			broadcaster.Broadcast(Message{
				Type:    "test",
				Payload: map[string]interface{}{"id": id},
			})
		}(i)
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)

	// Should handle concurrent broadcasts without crashing
	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_BroadcastTelemetry_Concurrent(t *testing.T) {
	broadcaster := GetBroadcaster()

	// 50 concurrent telemetry broadcasts
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			broadcaster.BroadcastTelemetry("device-test", map[string]interface{}{
				"temperature": 25.5,
				"humidity":    60,
			})
		}()
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)

	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_BroadcastAlert_Concurrent(t *testing.T) {
	broadcaster := GetBroadcaster()

	// 50 concurrent alert broadcasts
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			broadcaster.BroadcastAlert("alert-test", "high", "Test alert message")
		}()
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)

	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_LargePayload(t *testing.T) {
	broadcaster := GetBroadcaster()

	// Create large payload
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	broadcaster.Broadcast(Message{
		Type:    "large_test",
		Payload: largeData,
	})

	time.Sleep(10 * time.Millisecond)
	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_SpecialUnicodePayload(t *testing.T) {
	broadcaster := GetBroadcaster()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "Chinese characters",
			payload: map[string]interface{}{
				"message": "测试消息内容",
				"device":  "设备一号",
			},
		},
		{
			name: "Japanese characters",
			payload: map[string]interface{}{
				"message": "テストメッセージ",
				"device":  "デバイス",
			},
		},
		{
			name: "Emoji",
			payload: map[string]interface{}{
				"message": "Alert 🚨 Warning ⚠️",
				"status":  "🎉 Success",
			},
		},
		{
			name: "Mixed Unicode",
			payload: map[string]interface{}{
				"message": "测试-テスト-🎉",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcaster.Broadcast(Message{
				Type:    "unicode_test",
				Payload: tt.payload,
			})
			time.Sleep(5 * time.Millisecond)
		})
	}

	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_NilPayload(t *testing.T) {
	broadcaster := GetBroadcaster()

	broadcaster.Broadcast(Message{
		Type:    "nil_test",
		Payload: nil,
	})

	time.Sleep(5 * time.Millisecond)
	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_EmptyType(t *testing.T) {
	broadcaster := GetBroadcaster()

	broadcaster.Broadcast(Message{
		Type:    "",
		Payload: "test",
	})

	time.Sleep(5 * time.Millisecond)
	assert.NotNil(t, broadcaster)
}

func TestBroadcaster_ConnectionCount_ThreadSafe(t *testing.T) {
	broadcaster := GetBroadcaster()

	// Concurrent connection count queries
	var wg sync.WaitGroup
	results := make(chan int, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- broadcaster.ConnectionCount()
		}()
	}

	wg.Wait()
	close(results)

	// All should return valid counts
	for count := range results {
		assert.GreaterOrEqual(t, count, 0)
	}
}

func TestBroadcaster_GetCurrentTimestamp(t *testing.T) {
	ts := getCurrentTimestamp()

	assert.NotEmpty(t, ts)

	// Should be parseable as RFC3339
	_, err := time.Parse(time.RFC3339, ts)
	assert.NoError(t, err)
}

func TestMessage_JSONMarshal(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "normal message",
			message: Message{
				Type:    "test",
				Payload: "data",
			},
		},
		{
			name: "complex payload",
			message: Message{
				Type: "telemetry",
				Payload: map[string]interface{}{
					"temperature": 25.5,
					"humidity":    60.0,
					"pressure":    1013.25,
				},
			},
		},
		{
			name: "nil payload",
			message: Message{
				Type:    "nil",
				Payload: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			var unmarshaled Message
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.message.Type, unmarshaled.Type)
		})
	}
}

func TestBroadcaster_FloodTest(t *testing.T) {
	broadcaster := GetBroadcaster()

	// Flood with 500 messages rapidly
	for i := 0; i < 500; i++ {
		broadcaster.Broadcast(Message{
			Type:    "flood_test",
			Payload: i,
		})
	}

	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, broadcaster)
	assert.GreaterOrEqual(t, broadcaster.ConnectionCount(), 0)
}

func TestBroadcaster_MixedMessageTypes(t *testing.T) {
	broadcaster := GetBroadcaster()

	// Mix different message types concurrently
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			broadcaster.BroadcastTelemetry("device-1", nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			broadcaster.BroadcastAlert("alert-1", "high", "test")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			broadcaster.BroadcastDeviceUpdate("device-1", "online")
		}
	}()

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	assert.NotNil(t, broadcaster)
}
