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

// DingTalkNotifier handles alert notifications to DingTalk
type DingTalkNotifier struct {
	webhookURL string
	httpClient *http.Client
}

// NewDingTalkNotifier creates a new DingTalk notifier
func NewDingTalkNotifier(webhookURL string) *DingTalkNotifier {
	return &DingTalkNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DingTalkMessage represents a DingTalk webhook message
type DingTalkMessage struct {
	MsgType  string              `json:"msgtype"`
	Markdown DingTalkMarkdown    `json:"markdown,omitempty"`
	Text     DingTalkTextContent `json:"text,omitempty"`
	At       DingTalkAt          `json:"at,omitempty"`
}

// DingTalkMarkdown represents markdown content
type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingTalkTextContent represents text content
type DingTalkTextContent struct {
	Content string `json:"content"`
}

// DingTalkAt represents @ settings
type DingTalkAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// SendAlert sends an alert notification to DingTalk
func (n *DingTalkNotifier) SendAlert(ctx context.Context, alert *model.Alert, device *model.Device, rule *model.AlertRule) error {
	if n.webhookURL == "" {
		logger.L().Debug("DingTalk webhook URL not configured, skipping notification")
		return nil
	}

	// Build severity indicator
	severityIcon := n.getSeverityIcon(alert.Severity)
	severityText := n.getSeverityText(alert.Severity)

	// Build markdown message
	msg := DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: fmt.Sprintf("设备告警 - %s", rule.Name),
			Text: fmt.Sprintf(`
## %s 设备告警

**告警级别:** %s %s

**设备信息:**
- 设备名称: %s
- 设备ID: %s

**告警详情:**
- 告警内容: %s
- 触发时间: %s

**请及时登录工业智能平台查看详情并处理**
`,
				severityIcon,
				severityText,
				alert.Severity,
				device.Name,
				device.ID,
				alert.Message,
				alert.TriggeredAt.Format("2006-01-02 15:04:05"),
			),
		},
		At: DingTalkAt{
			IsAtAll: alert.Severity == "critical", // @所有人 for critical alerts
		},
	}

	// Send to DingTalk
	return n.send(ctx, msg)
}

// SendAlertResolved sends a resolved notification to DingTalk
func (n *DingTalkNotifier) SendAlertResolved(ctx context.Context, alertID int, deviceName string) error {
	if n.webhookURL == "" {
		return nil
	}

	msg := DingTalkMessage{
		MsgType: "text",
		Text: DingTalkTextContent{
			Content: fmt.Sprintf("✅ 告警已解决\n设备: %s\n告警ID: %d\n时间: %s", deviceName, alertID, time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	return n.send(ctx, msg)
}

// SendMaintenanceReminder sends a maintenance reminder to DingTalk
func (n *DingTalkNotifier) SendMaintenanceReminder(ctx context.Context, deviceID, deviceName string, daysSinceMaintenance int) error {
	if n.webhookURL == "" {
		return nil
	}

	msg := DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: "设备维护提醒",
			Text: fmt.Sprintf(`
## 🔧 设备维护提醒

**设备信息:**
- 设备名称: %s
- 设备ID: %s

**维护状态:**
- 已超过 %d 天未维护
- 建议安排检修

**请在工业智能平台查看设备详情**
`,
				deviceName,
				deviceID,
				daysSinceMaintenance,
			),
		},
	}

	return n.send(ctx, msg)
}

// send sends a message to DingTalk webhook
func (n *DingTalkNotifier) send(ctx context.Context, msg DingTalkMessage) error {
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
		logger.L().Warn("DingTalk webhook failed",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return fmt.Errorf("dingtalk webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	logger.L().Info("DingTalk notification sent successfully",
		zap.String("webhook", n.webhookURL),
	)

	return nil
}

// getSeverityIcon returns icon for severity level
func (n *DingTalkNotifier) getSeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔴" // Red
	case "high":
		return "🟠" // Orange
	case "medium":
		return "🟡" // Yellow
	case "low", "info":
		return "🟢" // Green
	default:
		return "⚪"
	}
}

// getSeverityText returns Chinese text for severity
func (n *DingTalkNotifier) getSeverityText(severity string) string {
	switch severity {
	case "critical":
		return "紧急"
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	case "info":
		return "提醒"
	default:
		return severity
	}
}
