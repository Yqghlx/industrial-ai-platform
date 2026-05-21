package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/response"
)

// ============================================
// Phase 1: 新的Handler结构体 - 分层重构
// ============================================

// AlertHandler handles alert-related requests
// 只依赖Service层，不直接访问Repository
type AlertHandler struct {
	alertSvc  service.AlertServiceInterface
	broadcast func(msg model.WSMessage)
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertSvc service.AlertServiceInterface, broadcast func(msg model.WSMessage)) *AlertHandler {
	return &AlertHandler{
		alertSvc:  alertSvc,
		broadcast: broadcast,
	}
}

// ListAlerts lists alerts with filters
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	ctx, cancel := getRequestContext(c)
	defer cancel()

	status := c.Query("status")
	severity := c.Query("severity")
	deviceID := c.Query("device_id")

	pagination := GetPagination(c)

	filterStatus := status
	if filterStatus == "" {
		filterStatus = "all"
	}

	alerts, total, err := h.alertSvc.GetAlerts(ctx, filterStatus, pagination.Page, pagination.PageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Apply additional filters
	if severity != "" || deviceID != "" {
		filtered := []model.Alert{}
		for _, a := range alerts {
			if severity != "" && a.Severity != severity {
				continue
			}
			if deviceID != "" && a.DeviceID != deviceID {
				continue
			}
			filtered = append(filtered, a)
		}
		alerts = filtered
		total = len(filtered)
	}

	c.JSON(200, gin.H{
		"data":      alerts,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

// GetAlert gets a single alert by ID
func (h *AlertHandler) GetAlert(c *gin.Context) {
	ctx := c.Request.Context()
	alertID := c.Param("id")

	var id int
	fmt.Sscanf(alertID, "%d", &id)

	alert, err := h.alertSvc.GetAlertByID(ctx, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(200, alert)
}

// ResolveAlert resolves an alert
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	ctx := c.Request.Context()
	alertID := c.Param("id")

	var id int
	fmt.Sscanf(alertID, "%d", &id)

	err := h.alertSvc.ResolveAlert(ctx, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	h.broadcast(model.WSMessage{
		Type: "alert_resolved",
		Payload: map[string]interface{}{
			"id":        id,
			"status":    "resolved",
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	c.JSON(200, gin.H{
		"message": "Alert resolved",
		"id":      id,
		"status":  "resolved",
	})
}

// AcknowledgeAlert acknowledges an alert
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	ctx := c.Request.Context()
	alertID := c.Param("id")

	var id int
	fmt.Sscanf(alertID, "%d", &id)

	err := h.alertSvc.AcknowledgeAlert(ctx, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	h.broadcast(model.WSMessage{
		Type: "alert_acknowledged",
		Payload: map[string]interface{}{
			"id":        id,
			"status":    "acknowledged",
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	c.JSON(200, gin.H{
		"message": "Alert acknowledged",
		"id":      id,
		"status":  "acknowledged",
	})
}

// GetTrend 获取告警趋势报告（占位实现）
func (h *AlertHandler) GetTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "7d")

	// 占位实现 - 实际需要扩展 AlertServiceInterface
	c.JSON(200, gin.H{
		"period":  period,
		"trend":   []interface{}{},
		"message": "GetTrend requires AlertServiceInterface extension",
	})
}

// GetRanking 获取告警设备排名（占位实现）
func (h *AlertHandler) GetRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 占位实现 - 实际需要扩展 AlertServiceInterface
	c.JSON(200, gin.H{
		"data":    []interface{}{},
		"limit":   limit,
		"message": "GetRanking requires AlertServiceInterface extension",
	})
}

// GetEfficiency 获取告警处理效率（占位实现）
func (h *AlertHandler) GetEfficiency(c *gin.Context) {
	// 占位实现 - 实际需要扩展 AlertServiceInterface
	c.JSON(200, gin.H{
		"efficiency": map[string]interface{}{
			"avg_resolve_time": 0,
			"ack_rate":         0,
		},
		"message": "GetEfficiency requires AlertServiceInterface extension",
	})
}
