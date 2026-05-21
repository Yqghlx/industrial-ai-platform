package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// TenantServiceInterface 定义租户服务接口，用于测试和依赖注入
type TenantServiceInterface interface {
	CreateTenant(name, slug, plan string, maxDevices int) (*model.Tenant, error)
	GetTenant(id string) (*model.Tenant, error)
	GetTenantBySlug(slug string) (*model.Tenant, error)
	ListTenants(limit, offset int) ([]model.Tenant, error)
	UpdateTenant(id string, updates map[string]interface{}) (*model.Tenant, error)
	DeleteTenant(id string) error
	CountTenants() (int, error)
}

// 确保 TenantService 实现了 TenantServiceInterface
var _ TenantServiceInterface = (*service.TenantService)(nil)

type TenantHandler struct {
	tenantSvc TenantServiceInterface
}

func NewTenantHandler(tenantSvc TenantServiceInterface) *TenantHandler {
	return &TenantHandler{tenantSvc: tenantSvc}
}

// CreateTenant - POST /api/v1/tenants (Admin)
func (h *TenantHandler) CreateTenant(c *gin.Context) {
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

	tenant, err := h.tenantSvc.CreateTenant(req.Name, req.Slug, req.Plan, 0) // maxDevices default 0
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
func (h *TenantHandler) ListTenants(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tenants, err := h.tenantSvc.ListTenants(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	count, err := h.tenantSvc.CountTenants()
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
func (h *TenantHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "tenant id required",
		})
		return
	}

	tenant, err := h.tenantSvc.GetTenant(id)
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
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
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

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.Plan != "" {
		updates["plan"] = req.Plan
	}
	if req.MaxDevices > 0 {
		updates["max_devices"] = req.MaxDevices
	}

	tenant, err := h.tenantSvc.UpdateTenant(id, updates)
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
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "tenant id required",
		})
		return
	}

	err := h.tenantSvc.DeleteTenant(id)
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
