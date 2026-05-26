package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/response"
)

// ============================================
// AdminHandlerNew - 管理员Handler（新架构）
// ============================================

// AdminHandlerNew 管理员处理器
type AdminHandlerNew struct {
	authSvc service.AuthServiceInterface
}

// NewAdminHandlerNew 创建管理员处理器
func NewAdminHandlerNew(authSvc service.AuthServiceInterface) *AdminHandlerNew {
	return &AdminHandlerNew{authSvc: authSvc}
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

	// 占位实现 - 实际应调用 authSvc.Register
	c.JSON(http.StatusOK, gin.H{
		"message": "User created (placeholder)",
		"user":    req,
	})
}

// DeleteUser 删除用户
func (h *AdminHandlerNew) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{"message": "User deleted (placeholder)", "id": userID})
}

// GetSystemStatus 获取系统状态
func (h *AdminHandlerNew) GetSystemStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "running",
	})
}
