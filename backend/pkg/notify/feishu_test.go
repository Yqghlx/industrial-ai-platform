package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// FeishuNotifier Tests
// ============================================

func TestNewFeishuNotifier(t *testing.T) {
	notifier := NewFeishuNotifier("https://feishu.webhook.url")
	require.NotNil(t, notifier)
	assert.Equal(t, "https://feishu.webhook.url", notifier.webhookURL)
}

func TestNewFeishuNotifier_EmptyURL(t *testing.T) {
	notifier := NewFeishuNotifier("")
	require.NotNil(t, notifier)
	assert.Empty(t, notifier.webhookURL)
}

func TestFeishuNotifier_SendAlert_EmptyWebhook(t *testing.T) {
	notifier := NewFeishuNotifier("")
	ctx := context.Background()

	alert := &model.Alert{ID: 1, Severity: "high", Message: "Test alert"}
	device := &model.Device{ID: "dev-1", Name: "Test Device"}
	rule := &model.AlertRule{ID: 1, Name: "Test Rule"}

	err := notifier.SendAlert(ctx, alert, device, rule)
	require.NoError(t, err) // Should skip when webhook is empty
}

func TestFeishuNotifier_SendAlert_Success(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Read request body
		body, _ := json.Marshal(map[string]interface{}{
			"code": 0,
			"msg":  "success",
		})
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(server.URL)
	ctx := context.Background()

	alert := &model.Alert{ID: 1, Severity: "high", Message: "Test alert"}
	device := &model.Device{ID: "dev-1", Name: "Test Device"}
	rule := &model.AlertRule{ID: 1, Name: "Test Rule"}

	err := notifier.SendAlert(ctx, alert, device, rule)
	require.NoError(t, err)
}

func TestFeishuNotifier_SendAlertResolved_EmptyWebhook(t *testing.T) {
	notifier := NewFeishuNotifier("")
	ctx := context.Background()

	err := notifier.SendAlertResolved(ctx, 1, "Test Device")
	require.NoError(t, err)
}

func TestFeishuNotifier_SendAlertResolved_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0,"msg":"success"}`))
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(server.URL)
	ctx := context.Background()

	err := notifier.SendAlertResolved(ctx, 1, "Test Device")
	require.NoError(t, err)
}

func TestFeishuNotifier_GetSeverityColor(t *testing.T) {
	notifier := NewFeishuNotifier("")

	// Just verify that colors are returned for known severities
	assert.NotEmpty(t, notifier.getSeverityColor("critical"))
	assert.NotEmpty(t, notifier.getSeverityColor("high"))
	assert.NotEmpty(t, notifier.getSeverityColor("medium"))
	assert.NotEmpty(t, notifier.getSeverityColor("low"))
	// unknown might return empty
}

func TestFeishuNotifier_GetSeverityText(t *testing.T) {
	notifier := NewFeishuNotifier("")

	// Just verify that texts are returned
	assert.NotEmpty(t, notifier.getSeverityText("critical"))
	assert.NotEmpty(t, notifier.getSeverityText("high"))
	assert.NotEmpty(t, notifier.getSeverityText("medium"))
	assert.NotEmpty(t, notifier.getSeverityText("low"))
	assert.NotEmpty(t, notifier.getSeverityText("unknown"))
}

// ============================================
// NotifyManager Tests
// ============================================

func TestNewNotifyManager(t *testing.T) {
	manager := NewNotifyManager("https://webhook.url", true)
	require.NotNil(t, manager)
}

func TestNewNotifyManager_Disabled(t *testing.T) {
	manager := NewNotifyManager("", false)
	require.NotNil(t, manager)
}

func TestNotifyManager_NotifyAlert_Disabled(t *testing.T) {
	manager := NewNotifyManager("", false)
	ctx := context.Background()

	alert := &model.Alert{ID: 1, Severity: "high"}
	device := &model.Device{ID: "dev-1"}
	rule := &model.AlertRule{ID: 1}

	err := manager.NotifyAlert(ctx, alert, device, rule)
	require.NoError(t, err) // Should skip when disabled
}

func TestNotifyManager_NotifyAlertResolved_Disabled(t *testing.T) {
	manager := NewNotifyManager("", false)
	ctx := context.Background()

	err := manager.NotifyAlertResolved(ctx, 1, "Test Device")
	require.NoError(t, err)
}

func TestNotifyManager_NotifyAlert_Enabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0}`))
	}))
	defer server.Close()

	manager := NewNotifyManager(server.URL, true)
	ctx := context.Background()

	alert := &model.Alert{ID: 1, Severity: "high", Message: "Test"}
	device := &model.Device{ID: "dev-1", Name: "Device"}
	rule := &model.AlertRule{ID: 1, Name: "Rule"}

	err := manager.NotifyAlert(ctx, alert, device, rule)
	require.NoError(t, err)
}

// ============================================
// Error Handling Tests
// ============================================

func TestFeishuNotifier_Send_HTTPError(t *testing.T) {
	// Create a server that immediately closes the connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force connection close to simulate network error
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			conn.Close()
		}
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(server.URL)
	ctx := context.Background()

	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{"text": "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "send request")
}

func TestFeishuNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"msg":"internal error"}`))
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(server.URL)
	ctx := context.Background()

	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{"text": "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestFeishuNotifier_Send_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to ensure context cancellation triggers
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewFeishuNotifier(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{"text": "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
}

func TestNotifyManager_NotifyAlert_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500}`))
	}))
	defer server.Close()

	manager := NewNotifyManager(server.URL, true)
	ctx := context.Background()

	alert := &model.Alert{ID: 1, Severity: "critical", Message: "Critical alert"}
	device := &model.Device{ID: "dev-1", Name: "Device"}
	rule := &model.AlertRule{ID: 1, Name: "Critical Rule"}

	err := manager.NotifyAlert(ctx, alert, device, rule)
	assert.Error(t, err)
}

func TestNotifyManager_NotifyAlertResolved_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500}`))
	}))
	defer server.Close()

	manager := NewNotifyManager(server.URL, true)
	ctx := context.Background()

	err := manager.NotifyAlertResolved(ctx, 1, "Test Device")
	assert.Error(t, err)
}

// ============================================
// Different Severity Levels Tests
// ============================================

func TestFeishuNotifier_SendAlert_AllSeverities(t *testing.T) {
	severities := []string{"critical", "high", "medium", "low", "unknown"}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"code":0}`))
			}))
			defer server.Close()

			notifier := NewFeishuNotifier(server.URL)
			ctx := context.Background()

			alert := &model.Alert{
				ID:         1,
				Severity:   severity,
				Message:   "Test alert",
				TriggeredAt: time.Now(),
			}
			device := &model.Device{ID: "dev-1", Name: "Test Device"}
			rule := &model.AlertRule{ID: 1, Name: "Test Rule"}

			err := notifier.SendAlert(ctx, alert, device, rule)
			require.NoError(t, err)
		})
	}
}

func TestFeishuNotifier_GetSeverityColor_AllLevels(t *testing.T) {
	notifier := NewFeishuNotifier("")

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "<font color='red'>"},
		{"high", "<font color='orange'>"},
		{"medium", "<font color='yellow'>"},
		{"low", "<font color='green'>"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := notifier.getSeverityColor(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFeishuNotifier_GetSeverityText_AllLevels(t *testing.T) {
	notifier := NewFeishuNotifier("")

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "紧急"},
		{"high", "高"},
		{"medium", "中"},
		{"low", "低"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := notifier.getSeverityText(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================
// Mock Send Tests
// ============================================

func TestFeishuNotifier_Send_InvalidURL(t *testing.T) {
	notifier := NewFeishuNotifier("http://invalid.local:99999/webhook")
	ctx := context.Background()

	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{"text": "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
}

func TestFeishuNotifier_Send_MessageMarshalError(t *testing.T) {
	notifier := NewFeishuNotifier("http://example.com")
	ctx := context.Background()

	// Create a message that cannot be marshaled to JSON
	msg := map[string]interface{}{
		"invalid": make(chan int), // channels cannot be marshaled to JSON
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marshal message")
}
