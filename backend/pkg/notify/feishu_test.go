package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
