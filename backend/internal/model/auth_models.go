package model

import (
	"fmt"
	"regexp"
	"strings"
)

// Password validation constants
const (
	MinPasswordLength = 12
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=12,max=100"`
}

// PasswordValidationError 密码验证错误
type PasswordValidationError struct {
	Errors []string
}

func (e *PasswordValidationError) Error() string {
	return "密码验证失败: " + strings.Join(e.Errors, ", ")
}

// ValidatePassword 验证密码复杂度
// 要求: 至少12位，包含大小写字母、数字和特殊字符
func ValidatePassword(password string) error {
	var errors []string

	// 检查最小长度
	if len(password) < MinPasswordLength {
		errors = append(errors, fmt.Sprintf("长度不足，至少需要%d个字符", MinPasswordLength))
	}

	// 检查大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		errors = append(errors, "缺少大写字母")
	}

	// 检查小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		errors = append(errors, "缺少小写字母")
	}

	// 检查数字
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		errors = append(errors, "缺少数字")
	}

	// 检查特殊字符
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?` + "`" + `~]`).MatchString(password)
	if !hasSpecial {
		errors = append(errors, "缺少特殊字符（如 !@#$%^&* 等）")
	}

	if len(errors) > 0 {
		return &PasswordValidationError{Errors: errors}
	}

	return nil
}

// LoginResponse 登录响应 (增强版 - 包含 Refresh Token)
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"` // Access Token 过期时间 (秒)
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

// TokenRefreshResponse Token 刷新响应
type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// LogoutRequest 注销请求 (可选提供 Refresh Token)
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"` // 可选
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=12"`
	NewPassword string `json:"new_password" binding:"required,min=12,max=100,nefield=OldPassword"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password string `json:"password" binding:"required,min=12,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"omitempty,oneof=admin operator viewer"`
}
