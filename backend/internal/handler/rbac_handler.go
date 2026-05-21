package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/middleware"
	"github.com/industrial-ai/platform/internal/model"
)

// FIX-003: Context 超时控制 - 使用 telemetry_handler.go 中定义的 getRequestContext

// RBACServiceInterface defines the interface for RBAC service
type RBACServiceInterface interface {
	CreateRole(ctx context.Context, tenantID, name, displayName, description string) (*model.Role, error)
	GetRole(ctx context.Context, id int) (*model.Role, error)
	GetRoleWithPermissions(ctx context.Context, id int) (*model.RoleResponse, error)
	ListRoles(ctx context.Context, tenantID string) ([]model.Role, error)
	UpdateRole(ctx context.Context, id int, updates map[string]interface{}) error
	DeleteRole(ctx context.Context, id int) error
	AssignRole(ctx context.Context, userID, roleID int, tenantID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	GetUserRoles(ctx context.Context, userID int) ([]model.Role, error)
	GetUserPermissions(ctx context.Context, userID int) ([]model.Permission, error)
	CheckPermission(ctx context.Context, userID int, resource, action string) (bool, error)
	CreatePermission(ctx context.Context, name, resource, action, description string) (*model.Permission, error)
	GetPermission(ctx context.Context, id int) (*model.Permission, error)
	ListPermissions(ctx context.Context) ([]model.Permission, error)
	DeletePermission(ctx context.Context, id int) error
	AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error)
}

// RBACHandler handles RBAC API requests
type RBACHandler struct {
	rbacSvc RBACServiceInterface
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(rbacSvc RBACServiceInterface) *RBACHandler {
	return &RBACHandler{rbacSvc: rbacSvc}
}

// CreateRole handles POST /api/v1/roles
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	// Get tenant ID from context if not specified
	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = middleware.GetTenantID(c)
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	role, err := h.rbacSvc.CreateRole(ctx, tenantID, req.Name, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Data:    role,
	})
}

// ListRoles handles GET /api/v1/roles
func (h *RBACHandler) ListRoles(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = middleware.GetTenantID(c)
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	roles, err := h.rbacSvc.ListRoles(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"roles": roles,
			"total": len(roles),
		},
	})
}

// GetRole handles GET /api/v1/roles/:id
func (h *RBACHandler) GetRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid role id",
		})
		return
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	roleResponse, err := h.rbacSvc.GetRole(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    roleResponse,
	})
}

// UpdateRole handles PUT /api/v1/roles/:id
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid role id",
		})
		return
	}

	var req model.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}
	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	if err := h.rbacSvc.UpdateRole(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Get the updated role to return in response
	role, err := h.rbacSvc.GetRole(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    role,
	})
}

// DeleteRole handles DELETE /api/v1/roles/:id
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid role id",
		})
		return
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	if err := h.rbacSvc.DeleteRole(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "role deleted successfully",
	})
}

// AssignRole handles POST /api/v1/users/:id/roles
func (h *RBACHandler) AssignRole(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid user id",
		})
		return
	}

	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	// Get tenant ID from context if not specified
	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = middleware.GetTenantID(c)
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	if err := h.rbacSvc.AssignRole(ctx, userID, req.RoleID, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "role assigned successfully",
	})
}

// RemoveRole handles DELETE /api/v1/users/:id/roles/:role_id
func (h *RBACHandler) RemoveRole(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid user id",
		})
		return
	}

	roleID, err := strconv.Atoi(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid role id",
		})
		return
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	if err := h.rbacSvc.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "role removed successfully",
	})
}

// GetUserRoles handles GET /api/v1/users/:id/roles
func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid user id",
		})
		return
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	roles, err := h.rbacSvc.GetUserRoles(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"roles": roles,
			"total": len(roles),
		},
	})
}

// ListPermissions handles GET /api/v1/permissions
// FIX-003: 使用带超时的请求上下文
func (h *RBACHandler) ListPermissions(c *gin.Context) {
	ctx, cancel := getRequestContext(c)
	defer cancel()
	permissions, err := h.rbacSvc.ListPermissions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"permissions": permissions,
			"total":       len(permissions),
		},
	})
}

// AssignPermission handles POST /api/v1/roles/:id/permissions
func (h *RBACHandler) AssignPermission(c *gin.Context) {
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid role id",
		})
		return
	}

	var req model.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	if err := h.rbacSvc.AssignPermissionToRole(ctx, roleID, req.PermissionID); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "permission assigned successfully",
	})
}

// RemovePermission handles DELETE /api/v1/roles/:id/permissions/:perm_id
func (h *RBACHandler) RemovePermission(c *gin.Context) {
	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid role id",
		})
		return
	}

	permissionID, err := strconv.Atoi(c.Param("perm_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid permission id",
		})
		return
	}

	// FIX-003: 使用带超时的请求上下文
	ctx, cancel := getRequestContext(c)
	defer cancel()
	if err := h.rbacSvc.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "permission removed successfully",
	})
}
