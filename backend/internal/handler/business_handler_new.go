package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/cache"
	"github.com/industrial-ai/platform/pkg/response"
)

// ============================================
// Phase 3: BusinessHandler重构 - 仅依赖Service层
// ============================================

// BusinessHandlerNew 业务处理器（新架构，只依赖Service）
type BusinessHandlerNew struct {
	workOrderSvc    service.WorkOrderServiceInterface
	notificationSvc service.NotificationServiceInterface
	blackBoxSvc     service.BlackBoxServiceInterface
	reportSvc       service.ReportServiceInterface
	alertSvc        service.AlertServiceInterface
	broadcast       func(msg model.WSMessage)
	cache           cache.CacheService // 添加缓存服务
}

// NewBusinessHandlerNew 创建业务处理器（新架构）
func NewBusinessHandlerNew(
	workOrderSvc service.WorkOrderServiceInterface,
	notificationSvc service.NotificationServiceInterface,
	blackBoxSvc service.BlackBoxServiceInterface,
	reportSvc service.ReportServiceInterface,
	alertSvc service.AlertServiceInterface,
	broadcast func(msg model.WSMessage),
	cacheSvc cache.CacheService, // 添加缓存参数
) *BusinessHandlerNew {
	return &BusinessHandlerNew{
		workOrderSvc:    workOrderSvc,
		notificationSvc: notificationSvc,
		blackBoxSvc:     blackBoxSvc,
		reportSvc:       reportSvc,
		alertSvc:        alertSvc,
		broadcast:       broadcast,
		cache:           cacheSvc,
	}
}

// ListWorkOrders 列出工单
func (h *BusinessHandlerNew) ListWorkOrders(c *gin.Context) {
	ctx := c.Request.Context()

	status := c.Query("status")
	deviceID := c.Query("device_id")
	pagination := GetPagination(c)

	orders, total, err := h.workOrderSvc.List(ctx, status, deviceID, pagination.Page, pagination.PageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      orders,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

// CreateWorkOrder 创建工单
func (h *BusinessHandlerNew) CreateWorkOrder(c *gin.Context) {
	ctx := c.Request.Context()

	var order model.WorkOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = "pending"

	if err := h.workOrderSvc.Create(ctx, &order); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdateWorkOrderStatus 更新工单状态
func (h *BusinessHandlerNew) UpdateWorkOrderStatus(c *gin.Context) {
	ctx := c.Request.Context()
	orderID := c.Param("id")

	var id int
	fmt.Sscanf(orderID, "%d", &id)

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.workOrderSvc.UpdateStatus(ctx, id, req.Status); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Work order updated",
		"id":      id,
		"status":  req.Status,
	})
}

// ListNotifications 列出通知
func (h *BusinessHandlerNew) ListNotifications(c *gin.Context) {
	ctx := c.Request.Context()

	status := c.Query("status")
	pagination := GetPagination(c)

	notifications, total, err := h.notificationSvc.List(ctx, status, pagination.Page, pagination.PageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notifications,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

// MarkNotificationRead 标记通知已读
func (h *BusinessHandlerNew) MarkNotificationRead(c *gin.Context) {
	ctx := c.Request.Context()
	notificationID := c.Param("id")

	var id int
	fmt.Sscanf(notificationID, "%d", &id)

	if err := h.notificationSvc.MarkRead(ctx, id); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read", "id": id})
}

// ListBlackBox 列出黑匣子记录
func (h *BusinessHandlerNew) ListBlackBox(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Query("device_id")
	pagination := GetPagination(c)

	records, total, err := h.blackBoxSvc.List(ctx, deviceID, pagination.Page, pagination.PageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

// ListReports 列出报告
func (h *BusinessHandlerNew) ListReports(c *gin.Context) {
	ctx := c.Request.Context()
	reportType := c.Query("type")
	pagination := GetPagination(c)

	reports, total, err := h.reportSvc.ListReports(ctx, reportType, pagination.Page, pagination.PageSize)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      reports,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

// GenerateReport 生成报告
func (h *BusinessHandlerNew) GenerateReport(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		ReportType string `json:"report_type" binding:"required"`
		DeviceID   string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	report, err := h.reportSvc.GenerateReport(ctx, req.ReportType, req.DeviceID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetROIStats 获取ROI统计（带缓存）
func (h *BusinessHandlerNew) GetROIStats(c *gin.Context) {
	ctx := c.Request.Context()

	if h.reportSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"total_devices":     0,
			"active_alerts":     0,
			"open_work_orders":  0,
			"resolved_issues":   0,
			"predicted_savings": 0,
			"uptime_percentage": 99.5,
			"avg_response_time": 2.5,
		})
		return
	}

	// 使用缓存（5分钟TTL）
	cacheKey := cache.ROICachePrefix.Build("stats")
	var stats *model.ROIStats

	if h.cache != nil && h.cache.IsAvailable() {
		// 从缓存获取或查询数据库
		cachedData, err := h.cache.Get(ctx, cacheKey)
		if err == nil {
			// 缓存命中
			if err := json.Unmarshal(cachedData, &stats); err == nil {
				c.JSON(http.StatusOK, stats)
				return
			}
		}

		// 缓存未命中，查询数据库
		stats, err = h.reportSvc.GetROIStats(ctx)
		if err != nil {
			response.HandleError(c, err)
			return
		}

		// 写入缓存
		if data, err := json.Marshal(stats); err == nil {
			_ = h.cache.Set(ctx, cacheKey, data, 5*time.Minute)
		}
	} else {
		// 缓存不可用时直接查询
		var err error
		stats, err = h.reportSvc.GetROIStats(ctx)
		if err != nil {
			response.HandleError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetAlertStats 获取告警统计（通过AlertService）
func (h *BusinessHandlerNew) GetAlertStats(c *gin.Context) {
	ctx := c.Request.Context()

	alerts, _, err := h.alertSvc.GetAlerts(ctx, "all", 1, 1000)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	severityCounts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}
	statusCounts := map[string]int{"active": 0, "resolved": 0, "acknowledged": 0}

	for _, a := range alerts {
		severityCounts[a.Severity]++
		statusCounts[a.Status]++
	}

	c.JSON(http.StatusOK, gin.H{
		"active_count": statusCounts["active"],
		"by_severity":  severityCounts,
		"by_status":    statusCounts,
		"total_count":  len(alerts),
	})
}

// GetWorkOrder 获取单个工单（占位实现）
func (h *BusinessHandlerNew) GetWorkOrder(c *gin.Context) {
	orderID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":      orderID,
		"message": "GetWorkOrder requires WorkOrderService extension",
	})
}

// GetBlackBoxData 获取黑匣子数据（占位实现）
func (h *BusinessHandlerNew) GetBlackBoxData(c *gin.Context) {
	recordID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":      recordID,
		"data":    []interface{}{},
		"message": "GetBlackBoxData requires BlackBoxService extension",
	})
}

// ExportDevices 导出设备数据（占位实现）
func (h *BusinessHandlerNew) ExportDevices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ExportDevices requires ReportService extension",
	})
}

// ExportAlerts 导出告警数据（占位实现）
func (h *BusinessHandlerNew) ExportAlerts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ExportAlerts requires ReportService extension",
	})
}

// ExportROI 导出ROI数据（占位实现）
func (h *BusinessHandlerNew) ExportROI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ExportROI requires ReportService extension",
	})
}
