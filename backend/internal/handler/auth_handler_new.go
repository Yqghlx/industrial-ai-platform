package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/response"
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
		response.BadRequest(c, err.Error())
		return
	}

	user, token, err := h.authSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, "Invalid credentials")
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
		response.BadRequest(c, err.Error())
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
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// ChangePassword 修改密码
func (h *AuthHandlerNew) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	// Convert userID to int
	uid, ok := userID.(int)
	if !ok {
		response.Unauthorized(c, "Invalid user ID")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=12,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate new password complexity
	if err := model.ValidatePassword(req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Call service to change password
	if err := h.authSvc.ChangePassword(ctx, uid, req.OldPassword, req.NewPassword); err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// RefreshToken 刷新Token
func (h *AuthHandlerNew) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Call service to refresh token
	tokenPair, err := h.authSvc.RefreshToken(ctx, req.Token)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
		"token_type":    tokenPair.TokenType,
	})
}

// ValidateToken 验证Token
func (h *AuthHandlerNew) ValidateToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Call service to validate token
	claims, err := h.authSvc.ValidateToken(ctx, req.Token)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}
