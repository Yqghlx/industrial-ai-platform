package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// TenantServiceInterface 定义租户服务接口，用于测试和依赖注入
// FIX-003: 添加 context 支持
type TenantServiceInterface interface {
	CreateTenant(ctx context.Context, name, slug, plan string, maxDevices int) (*model.Tenant, error)
	GetTenant(ctx context.Context, id string) (*model.Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error)
	ListTenants(ctx context.Context, limit, offset int) ([]model.Tenant, error)
	UpdateTenant(ctx context.Context, id string, updates *model.TenantUpdates) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, id string) error
	CountTenants(ctx context.Context) (int, error)
}

// 确保 TenantService 实现了 TenantServiceInterface
var _ TenantServiceInterface = (*service.TenantService)(nil)

type TenantHandler struct {
	tenantSvc TenantServiceInterface
}

func NewTenantHandler(tenantSvc TenantServiceInterface) *TenantHandler {
	return &TenantHandler{tenantSvc: tenantSvc}
}

// getTenantRequestContext creates a context with timeout for tenant operations
// FIX-003: 默认超时控制
func getTenantRequestContext(c *gin.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	return ctx, cancel
}

// CreateTenant - POST /api/v1/tenants (Admin)
// FIX-003: 使用带超时的请求上下文
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	ctx, cancel := getTenantRequestContext(c)
	defer cancel()

	var req model.TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Slug == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "name and slug are required",
		})
		return
	}

	tenant, err := h.tenantSvc.CreateTenant(ctx, req.Name, req.Slug, req.Plan, 0) // maxDevices default 0
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Data:    tenant,
	})
}

// ListTenants - GET /api/v1/tenants (Admin)
// FIX-003: 使用带超时的请求上下文
func (h *TenantHandler) ListTenants(c *gin.Context) {
	ctx, cancel := getTenantRequestContext(c)
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tenants, err := h.tenantSvc.ListTenants(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	count, err := h.tenantSvc.CountTenants(ctx)
	if err != nil {
		count = 0
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"tenants": tenants,
			"total":   count,
			"limit":   limit,
			"offset":  offset,
		},
	})
}

// GetTenant - GET /api/v1/tenants/:id
// FIX-003: 使用带超时的请求上下文
func (h *TenantHandler) GetTenant(c *gin.Context) {
	ctx, cancel := getTenantRequestContext(c)
	defer cancel()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "tenant id required",
		})
		return
	}

	tenant, err := h.tenantSvc.GetTenant(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "tenant not found",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    tenant,
	})
}

// UpdateTenant - PUT /api/v1/tenants/:id
// FIX-003: 使用带超时的请求上下文
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	ctx, cancel := getTenantRequestContext(c)
	defer cancel()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "tenant id required",
		})
		return
	}

	var req model.TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	updates := &model.TenantUpdates{}
	if req.Name != "" {
		updates.Name = &req.Name
	}
	if req.Slug != "" {
		updates.Slug = &req.Slug
	}
	if req.Plan != "" {
		updates.Plan = &req.Plan
	}
	if req.MaxDevices > 0 {
		updates.MaxDevices = &req.MaxDevices
	}

	tenant, err := h.tenantSvc.UpdateTenant(ctx, id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    tenant,
	})
}

// DeleteTenant - DELETE /api/v1/tenants/:id (Admin)
// FIX-003: 使用带超时的请求上下文
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	ctx, cancel := getTenantRequestContext(c)
	defer cancel()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "tenant id required",
		})
		return
	}

	err := h.tenantSvc.DeleteTenant(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "tenant deleted",
	})
}

// RegisterTenantRoutes registers tenant routes on the given router group
func RegisterTenantRoutes(r *gin.RouterGroup, h *TenantHandler, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	// All tenant routes require authentication
	tenants := r.Group("/tenants")
	tenants.Use(authMiddleware)

	// Admin-only routes
	tenants.POST("", adminMiddleware, h.CreateTenant)
	tenants.GET("", adminMiddleware, h.ListTenants)
	tenants.DELETE("/:id", adminMiddleware, h.DeleteTenant)

	// Authenticated routes
	tenants.GET("/:id", h.GetTenant)
	tenants.PUT("/:id", h.UpdateTenant)
}
