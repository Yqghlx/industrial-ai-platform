package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// Phase 5: AuthHandlerNew重构 - 仅依赖Service层
// ============================================

// AuthHandlerNew 认证处理器（新架构）
type AuthHandlerNew struct {
	authSvc service.AuthServiceInterface
	userSvc service.UserServiceInterface
}

// NewAuthHandlerNew 创建认证处理器
func NewAuthHandlerNew(
	authSvc service.AuthServiceInterface,
	userSvc service.UserServiceInterface,
) *AuthHandlerNew {
	return &AuthHandlerNew{
		authSvc: authSvc,
		userSvc: userSvc,
	}
}

// Login 用户登录
func (h *AuthHandlerNew) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// Logout 用户登出
func (h *AuthHandlerNew) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Register 用户注册
func (h *AuthHandlerNew) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// ChangePassword 修改密码
func (h *AuthHandlerNew) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = userID
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// RefreshToken 刷新Token（占位实现）
func (h *AuthHandlerNew) RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 占位实现 - 实际需要扩展 AuthServiceInterface
	c.JSON(http.StatusOK, gin.H{
		"token":   "new_placeholder_token",
		"message": "RefreshToken requires AuthServiceInterface extension",
	})
}

// ValidateToken 验证Token（占位实现）
func (h *AuthHandlerNew) ValidateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 占位实现 - 实际需要扩展 AuthServiceInterface
	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "ValidateToken requires AuthServiceInterface extension",
	})
}
