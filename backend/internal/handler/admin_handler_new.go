package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/service"
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
// TODO: 需要扩展 AuthServiceInterface 添加 List 方法
func (h *AdminHandlerNew) ListUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data":      []interface{}{},
		"total":     0,
		"page":      1,
		"page_size": 50,
		"message":   "ListUsers requires AuthServiceInterface extension",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
