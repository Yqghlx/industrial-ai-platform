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
// DingTalkNotifier Tests
// ============================================

func TestNewDingTalkNotifier(t *testing.T) {
	notifier := NewDingTalkNotifier("https://dingtalk.webhook.url")
	require.NotNil(t, notifier)
	assert.Equal(t, "https://dingtalk.webhook.url", notifier.webhookURL)
}

func TestNewDingTalkNotifier_EmptyURL(t *testing.T) {
	notifier := NewDingTalkNotifier("")
	require.NotNil(t, notifier)
	assert.Empty(t, notifier.webhookURL)
}

func TestDingTalkNotifier_SendAlert_EmptyWebhook(t *testing.T) {
	notifier := NewDingTalkNotifier("")
	ctx := context.Background()

	alert := &model.Alert{ID: 1, Severity: "high", Message: "Test alert"}
	device := &model.Device{ID: "dev-1", Name: "Test Device"}
	rule := &model.AlertRule{ID: 1, Name: "Test Rule"}

	err := notifier.SendAlert(ctx, alert, device, rule)
	require.NoError(t, err) // Should skip when webhook is empty
}

func TestDingTalkNotifier_SendAlert_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := json.Marshal(map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		})
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx := context.Background()

	alert := &model.Alert{
		ID:          1,
		Severity:    "high",
		Message:     "Test alert",
		TriggeredAt: time.Now(),
	}
	device := &model.Device{ID: "dev-1", Name: "Test Device"}
	rule := &model.AlertRule{ID: 1, Name: "Test Rule"}

	err := notifier.SendAlert(ctx, alert, device, rule)
	require.NoError(t, err)
}

func TestDingTalkNotifier_SendAlert_Critical(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx := context.Background()

	alert := &model.Alert{
		ID:          1,
		Severity:    "critical",
		Message:     "Critical alert",
		TriggeredAt: time.Now(),
	}
	device := &model.Device{ID: "dev-1", Name: "Critical Device"}
	rule := &model.AlertRule{ID: 1, Name: "Critical Rule"}

	err := notifier.SendAlert(ctx, alert, device, rule)
	require.NoError(t, err)
}

func TestDingTalkNotifier_SendAlertResolved_EmptyWebhook(t *testing.T) {
	notifier := NewDingTalkNotifier("")
	ctx := context.Background()

	err := notifier.SendAlertResolved(ctx, 1, "Test Device")
	require.NoError(t, err)
}

func TestDingTalkNotifier_SendAlertResolved_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx := context.Background()

	err := notifier.SendAlertResolved(ctx, 1, "Test Device")
	require.NoError(t, err)
}

func TestDingTalkNotifier_SendMaintenanceReminder_EmptyWebhook(t *testing.T) {
	notifier := NewDingTalkNotifier("")
	ctx := context.Background()

	err := notifier.SendMaintenanceReminder(ctx, "dev-1", "Test Device", 30)
	require.NoError(t, err)
}

func TestDingTalkNotifier_SendMaintenanceReminder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx := context.Background()

	err := notifier.SendMaintenanceReminder(ctx, "dev-1", "Test Device", 30)
	require.NoError(t, err)
}

// ============================================
// Severity Tests
// ============================================

func TestDingTalkNotifier_GetSeverityIcon(t *testing.T) {
	notifier := NewDingTalkNotifier("")

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "🔴"},
		{"high", "🟠"},
		{"medium", "🟡"},
		{"low", "🟢"},
		{"info", "🟢"},
		{"unknown", "⚪"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := notifier.getSeverityIcon(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDingTalkNotifier_GetSeverityText(t *testing.T) {
	notifier := NewDingTalkNotifier("")

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "紧急"},
		{"high", "高"},
		{"medium", "中"},
		{"low", "低"},
		{"info", "提醒"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := notifier.getSeverityText(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDingTalkNotifier_SendAlert_AllSeverities(t *testing.T) {
	severities := []string{"critical", "high", "medium", "low", "info", "unknown"}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"errcode":0}`))
			}))
			defer server.Close()

			notifier := NewDingTalkNotifier(server.URL)
			ctx := context.Background()

			alert := &model.Alert{
				ID:          1,
				Severity:    severity,
				Message:     "Test alert",
				TriggeredAt: time.Now(),
			}
			device := &model.Device{ID: "dev-1", Name: "Test Device"}
			rule := &model.AlertRule{ID: 1, Name: "Test Rule"}

			err := notifier.SendAlert(ctx, alert, device, rule)
			require.NoError(t, err)
		})
	}
}

// ============================================
// Error Handling Tests
// ============================================

func TestDingTalkNotifier_Send_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			conn.Close()
		}
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx := context.Background()

	msg := DingTalkMessage{
		MsgType: "text",
		Text:    DingTalkTextContent{Content: "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "send request")
}

func TestDingTalkNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errcode":500,"errmsg":"internal error"}`))
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx := context.Background()

	msg := DingTalkMessage{
		MsgType: "text",
		Text:    DingTalkTextContent{Content: "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestDingTalkNotifier_Send_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewDingTalkNotifier(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := DingTalkMessage{
		MsgType: "text",
		Text:    DingTalkTextContent{Content: "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
}

func TestDingTalkNotifier_Send_InvalidURL(t *testing.T) {
	notifier := NewDingTalkNotifier("http://invalid.local:99999/webhook")
	ctx := context.Background()

	msg := DingTalkMessage{
		MsgType: "text",
		Text:    DingTalkTextContent{Content: "test"},
	}

	err := notifier.send(ctx, msg)
	assert.Error(t, err)
}

// ============================================
// NotifyManager with DingTalk Tests
// ============================================

func TestNewNotifyManagerWithDingTalk(t *testing.T) {
	manager := NewNotifyManagerWithDingTalk("https://feishu.webhook", "https://dingtalk.webhook", true)
	require.NotNil(t, manager)
	require.NotNil(t, manager.feishu)
	require.NotNil(t, manager.dingtalk)
	assert.True(t, manager.enabled)
}

func TestNewNotifyManagerWithDingTalk_Disabled(t *testing.T) {
	manager := NewNotifyManagerWithDingTalk("", "", false)
	require.NotNil(t, manager)
	assert.False(t, manager.enabled)
}
