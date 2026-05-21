package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/validation"
)

// ============================================
// Phase 2: DeviceHandlerNew重构 - 仅依赖Service层
// ============================================

// DeviceHandlerNew 设备处理器（新架构，只依赖Service）
type DeviceHandlerNew struct {
	deviceSvc    service.DeviceServiceInterface
	alertSvc     service.AlertServiceInterface
	authSvc      service.AuthServiceInterface
	telemetrySvc service.TelemetryServiceInterface
	broadcast    func(msg model.WSMessage)
}

// NewDeviceHandlerNew 创建设备处理器（新架构）
func NewDeviceHandlerNew(
	deviceSvc service.DeviceServiceInterface,
	alertSvc service.AlertServiceInterface,
	authSvc service.AuthServiceInterface,
	telemetrySvc service.TelemetryServiceInterface,
	broadcast func(msg model.WSMessage),
) *DeviceHandlerNew {
	return &DeviceHandlerNew{
		deviceSvc:    deviceSvc,
		alertSvc:     alertSvc,
		authSvc:      authSvc,
		telemetrySvc: telemetrySvc,
		broadcast:    broadcast,
	}
}

// ListDevices 列出设备
func (h *DeviceHandlerNew) ListDevices(c *gin.Context) {
	ctx := c.Request.Context()
	pagination := GetPagination(c)

	devices, total, err := h.deviceSvc.List(ctx, pagination.Page, pagination.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "DATABASE_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      devices,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

// GetDevice 获取单个设备
func (h *DeviceHandlerNew) GetDevice(c *gin.Context) {
	ctx := c.Request.Context()
	deviceID := c.Param("id")

	if err := validation.ValidateDeviceID(deviceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_DEVICE_ID"})
		return
	}

	device, err := h.deviceSvc.GetByID(ctx, deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found", "code": "NOT_FOUND"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// CreateDevice 创建设备
func (h *DeviceHandlerNew) CreateDevice(c *gin.Context) {
	ctx := c.Request.Context()

	var device model.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	if err := validation.ValidateDeviceName(device.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_NAME"})
		return
	}

	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()

	if err := h.deviceSvc.Create(ctx, &device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "CREATE_FAILED"})
		return
	}

	h.broadcast(model.WSMessage{
		Type: "device_created",
		Payload: map[string]interface{}{
			"device_id": device.ID,
			"name":      device.Name,
			"type":      device.Type,
		},
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, device)
}

// UpdateDevice 更新设备
func (h *DeviceHandlerNew) UpdateDevice(c *gin.Context) {
	ctx := c.Request.Context()
	deviceID := c.Param("id")

	var device model.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	device.ID = deviceID
	device.UpdatedAt = time.Now()

	if err := h.deviceSvc.Update(ctx, &device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "UPDATE_FAILED"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice 删除设备
func (h *DeviceHandlerNew) DeleteDevice(c *gin.Context) {
	ctx := c.Request.Context()
	deviceID := c.Param("id")

	if err := h.deviceSvc.Delete(ctx, deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "DELETE_FAILED"})
		return
	}

	h.broadcast(model.WSMessage{
		Type: "device_deleted",
		Payload: map[string]interface{}{
			"device_id": deviceID,
		},
		Timestamp: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted", "id": deviceID})
}

// GetDeviceGraph 获取设备关系图
func (h *DeviceHandlerNew) GetDeviceGraph(c *gin.Context) {
	ctx := c.Request.Context()

	graph, err := h.deviceSvc.GetGraph(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "GRAPH_ERROR"})
		return
	}

	c.JSON(http.StatusOK, graph)
}

// ListRules 列出告警规则
func (h *DeviceHandlerNew) ListRules(c *gin.Context) {
	ctx := c.Request.Context()

	rules, err := h.alertSvc.GetRules(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "DATABASE_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// CreateRule 创建告警规则
func (h *DeviceHandlerNew) CreateRule(c *gin.Context) {
	ctx := c.Request.Context()

	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := h.alertSvc.CreateRule(ctx, &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "CREATE_FAILED"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateUser 创建用户
func (h *DeviceHandlerNew) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}

	// 转换为 RegisterRequest 格式
	registerReq := &model.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	user, token, err := h.authSvc.Register(ctx, registerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "CREATE_FAILED"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetLatestTelemetry 获取最新遥测数据
func (h *DeviceHandlerNew) GetLatestTelemetry(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.telemetrySvc.GetLatest(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetDeviceTelemetry 获取设备遥测数据
func (h *DeviceHandlerNew) GetDeviceTelemetry(c *gin.Context) {
	ctx := c.Request.Context()
	deviceID := c.Param("id")

	data, err := h.telemetrySvc.GetLatestByDevice(ctx, deviceID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "device_id": deviceID})
}

// GetDeviceStats 获取设备统计（占位实现）
func (h *DeviceHandlerNew) GetDeviceStats(c *gin.Context) {
	deviceID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"device_id": deviceID,
		"stats":     map[string]interface{}{},
		"message":   "GetDeviceStats requires DeviceServiceInterface extension",
	})
}

// GetRule 获取单个规则（占位实现）
func (h *DeviceHandlerNew) GetRule(c *gin.Context) {
	ruleID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":      ruleID,
		"message": "GetRule requires AlertServiceInterface extension",
	})
}

// UpdateRule 更新规则
func (h *DeviceHandlerNew) UpdateRule(c *gin.Context) {
	ctx := c.Request.Context()
	ruleID := c.Param("id")

	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	fmt.Sscanf(ruleID, "%d", &id)
	rule.ID = id
	rule.UpdatedAt = time.Now()

	if err := h.alertSvc.UpdateRule(ctx, &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// ToggleRule 启用/禁用规则（占位实现）
func (h *DeviceHandlerNew) ToggleRule(c *gin.Context) {
	ruleID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Rule toggled (placeholder)",
		"id":      ruleID,
	})
}

// DeleteRule 删除规则
func (h *DeviceHandlerNew) DeleteRule(c *gin.Context) {
	ctx := c.Request.Context()
	ruleID := c.Param("id")

	var id int
	fmt.Sscanf(ruleID, "%d", &id)

	if err := h.alertSvc.DeleteRule(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted", "id": id})
}
