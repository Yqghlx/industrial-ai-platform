package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/response"
)

// ============================================
// AdminHandlerNew - 管理员Handler（新架构）
// ============================================

// AdminHandlerNew 管理员处理器
type AdminHandlerNew struct {
	authSvc      service.AuthServiceInterface
	telemetrySvc service.TelemetryServiceInterface
}

// NewAdminHandlerNew 创建管理员处理器
func NewAdminHandlerNew(authSvc service.AuthServiceInterface, telemetrySvc service.TelemetryServiceInterface) *AdminHandlerNew {
	return &AdminHandlerNew{authSvc: authSvc, telemetrySvc: telemetrySvc}
}

// ListUsers 列出所有用户
func (h *AdminHandlerNew) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 50

	// 从查询参数获取分页信息
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	users, total, err := h.authSvc.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list users")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateUser 创建用户（管理员）
// SEC-HIGH-03: 完整实现用户创建逻辑，不再使用占位实现
// 管理员创建用户时可以指定角色，无需验证邮箱
func (h *AdminHandlerNew) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// SEC-HIGH-03: 验证必填字段
	if req.Username == "" {
		response.BadRequest(c, "username is required")
		return
	}
	if req.Password == "" {
		response.BadRequest(c, "password is required")
		return
	}

	// 验证密码强度
	if err := model.ValidatePassword(req.Password); err != nil {
		response.BadRequest(c, "password validation failed: "+err.Error())
		return
	}

	// 设置默认角色
	role := req.Role
	if role == "" {
		role = "user"
	}

	// 验证角色是否合法
	validRoles := map[string]bool{"admin": true, "user": true, "operator": true, "viewer": true}
	if !validRoles[role] {
		response.BadRequest(c, "invalid role: must be one of admin, user, operator, viewer")
		return
	}

	// 构建注册请求
	registerReq := &model.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Role:     role,
	}

	// SEC-HIGH-03: 调用 AuthService 实际创建用户
	ctx := c.Request.Context()
	user, token, err := h.authSvc.Register(ctx, registerReq)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"token": token,
	})
}

// DeleteUser 删除用户
// SEC-HIGH-03: 完整实现用户删除逻辑，不再使用占位实现
// 需要通过 AuthService 调用 UserRepository 进行实际删除
func (h *AdminHandlerNew) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")

	// SEC-HIGH-03: 验证用户ID
	if userIDStr == "" {
		response.BadRequest(c, "user id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid user id format")
		return
	}

	// 检查是否尝试删除当前登录的管理员用户
	currentUserID, exists := c.Get("user_id")
	if exists {
		if currentID, ok := currentUserID.(int); ok && currentID == userID {
			response.BadRequest(c, "cannot delete currently logged in user")
			return
		}
	}

	// SEC-HIGH-03: 先检查用户是否存在
	ctx := c.Request.Context()
	user, err := h.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// 防止删除最后一个管理员
	if user.Role == "admin" {
		// 检查是否是最后一个管理员
		users, _, err := h.authSvc.ListUsers(ctx, 1, 100)
		if err == nil {
			adminCount := 0
			for _, u := range users {
				if u.Role == "admin" {
					adminCount++
				}
			}
			if adminCount <= 1 {
				response.BadRequest(c, "cannot delete the last admin user")
				return
			}
		}
	}

	// SEC-HIGH-03: 通过 AuthService 删除用户
	if err := h.authSvc.DeleteUser(ctx, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"id":      userID,
	})
}

// GetSystemStatus 获取系统状态
func (h *AdminHandlerNew) GetSystemStatus(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.telemetrySvc.GetSystemStatus(ctx)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}
