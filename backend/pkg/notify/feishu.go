package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// FeishuNotifier handles alert notifications to Feishu
type FeishuNotifier struct {
	webhookURL string
	httpClient *http.Client
}

// NewFeishuNotifier creates a new Feishu notifier
func NewFeishuNotifier(webhookURL string) *FeishuNotifier {
	return &FeishuNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FeishuMessage represents a Feishu webhook message
type FeishuMessage struct {
	MsgType string                 `json:"msg_type"`
	Content map[string]interface{} `json:"content"`
}

// FeishuCard represents a Feishu card message
type FeishuCard struct {
	MsgType string         `json:"msg_type"`
	Card    FeishuCardBody `json:"card"`
}

type FeishuCardBody struct {
	Config   FeishuCardConfig    `json:"config"`
	Elements []FeishuCardElement `json:"elements"`
}

type FeishuCardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
	EnableForward  bool `json:"enable_forward"`
}

type FeishuCardElement struct {
	Tag    string        `json:"tag"`
	Text   *FeishuText   `json:"text,omitempty"`
	Fields []FeishuField `json:"fields,omitempty"`
}

type FeishuText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type FeishuField struct {
	IsShort bool       `json:"is_short"`
	Text    FeishuText `json:"text"`
}

// SendAlert sends an alert notification to Feishu
func (n *FeishuNotifier) SendAlert(ctx context.Context, alert *model.Alert, device *model.Device, rule *model.AlertRule) error {
	if n.webhookURL == "" {
		logger.L().Debug("Feishu webhook URL not configured, skipping notification")
		return nil
	}

	// Build severity color
	severityColor := n.getSeverityColor(alert.Severity)

	// Build card message
	card := FeishuCard{
		MsgType: "interactive",
		Card: FeishuCardBody{
			Config: FeishuCardConfig{
				WideScreenMode: true,
				EnableForward:  true,
			},
			Elements: []FeishuCardElement{
				// Title
				{
					Tag: "div",
					Text: &FeishuText{
						Tag:     "lark_md",
						Content: fmt.Sprintf("**🔴 设备告警 - %s**", rule.Name),
					},
				},
				// Severity indicator
				{
					Tag: "div",
					Text: &FeishuText{
						Tag:     "lark_md",
						Content: fmt.Sprintf("**级别:** %s%s%s", severityColor, n.getSeverityText(alert.Severity), severityColor),
					},
				},
				// Device info
				{
					Tag: "div",
					Fields: []FeishuField{
						{
							IsShort: true,
							Text: FeishuText{
								Tag:     "lark_md",
								Content: fmt.Sprintf("**设备:** %s", device.Name),
							},
						},
						{
							IsShort: true,
							Text: FeishuText{
								Tag:     "lark_md",
								Content: fmt.Sprintf("**设备ID:** %s", device.ID),
							},
						},
					},
				},
				// Alert message
				{
					Tag: "div",
					Text: &FeishuText{
						Tag:     "lark_md",
						Content: fmt.Sprintf("**告警内容:** %s", alert.Message),
					},
				},
				// Timestamp
				{
					Tag: "div",
					Fields: []FeishuField{
						{
							IsShort: true,
							Text: FeishuText{
								Tag:     "lark_md",
								Content: fmt.Sprintf("**触发时间:** %s", alert.TriggeredAt.Format("2006-01-02 15:04:05")),
							},
						},
						{
							IsShort: true,
							Text: FeishuText{
								Tag:     "lark_md",
								Content: fmt.Sprintf("**告警ID:** %d", alert.ID),
							},
						},
					},
				},
				// Action hint
				{
					Tag: "note",
					Text: &FeishuText{
						Tag:     "lark_md",
						Content: "请登录工业智能平台查看详情并处理告警",
					},
				},
			},
		},
	}

	// Send to Feishu
	return n.send(ctx, card)
}

// SendAlertResolved sends a resolved notification to Feishu
func (n *FeishuNotifier) SendAlertResolved(ctx context.Context, alertID int, deviceName string) error {
	if n.webhookURL == "" {
		return nil
	}

	// Build simple text message
	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{
			"text": fmt.Sprintf("✅ 告警已解决\n设备: %s\n告警ID: %d\n时间: %s", deviceName, alertID, time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	return n.send(ctx, msg)
}

// send sends a message to Feishu webhook
func (n *FeishuNotifier) send(ctx context.Context, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.L().Warn("Feishu webhook failed",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return fmt.Errorf("feishu webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	logger.L().Info("Feishu notification sent successfully")

	return nil
}

// getSeverityColor returns Feishu color code for severity
func (n *FeishuNotifier) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "<font color='red'>" // Red
	case "high":
		return "<font color='orange'>" // Orange
	case "medium":
		return "<font color='yellow'>" // Yellow
	case "low":
		return "<font color='green'>" // Green
	default:
		return ""
	}
}

// getSeverityText returns Chinese text for severity
func (n *FeishuNotifier) getSeverityText(severity string) string {
	switch severity {
	case "critical":
		return "紧急"
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	default:
		return severity
	}
}

// NotifyManager manages multiple notification channels
// P2-002: 增强通知管理器，支持飞书和钉钉
type NotifyManager struct {
	feishu   *FeishuNotifier
	dingtalk *DingTalkNotifier
	enabled  bool
}

// NewNotifyManager creates a notification manager
func NewNotifyManager(feishuWebhook string, enabled bool) *NotifyManager {
	return &NotifyManager{
		feishu:  NewFeishuNotifier(feishuWebhook),
		enabled: enabled,
	}
}

// NewNotifyManagerWithDingTalk creates a notification manager with both Feishu and DingTalk
func NewNotifyManagerWithDingTalk(feishuWebhook, dingtalkWebhook string, enabled bool) *NotifyManager {
	return &NotifyManager{
		feishu:   NewFeishuNotifier(feishuWebhook),
		dingtalk: NewDingTalkNotifier(dingtalkWebhook),
		enabled:  enabled,
	}
}

// NotifyAlert sends alert notification to all channels
func (m *NotifyManager) NotifyAlert(ctx context.Context, alert *model.Alert, device *model.Device, rule *model.AlertRule) error {
	if !m.enabled {
		return nil
	}

	// Send to Feishu
	if err := m.feishu.SendAlert(ctx, alert, device, rule); err != nil {
		logger.L().Error("Failed to send Feishu notification",
			zap.Error(err),
			zap.Int("alert_id", alert.ID),
		)
		return err
	}

	return nil
}

// NotifyAlertResolved sends resolved notification
func (m *NotifyManager) NotifyAlertResolved(ctx context.Context, alertID int, deviceName string) error {
	if !m.enabled {
		return nil
	}

	return m.feishu.SendAlertResolved(ctx, alertID, deviceName)
}
