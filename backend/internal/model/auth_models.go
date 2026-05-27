package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/industrial-ai/platform/pkg/constants"
)

// P3-04: 使用 pkg/constants 中的统一常量，避免重复定义
// MinPasswordLength 和 MaxPasswordLength 定义在 pkg/constants/constants.go

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=12,max=100"`
}

// PasswordValidationError represents password validation errors
type PasswordValidationError struct {
	Errors []string
}

func (e *PasswordValidationError) Error() string {
	return "password validation failed: " + strings.Join(e.Errors, ", ")
}

// ValidatePassword validates password complexity
// Requirements: at least 12 characters, contains uppercase, lowercase, digits, and special characters
func ValidatePassword(password string) error {
	var errors []string

	// Check minimum length
	if len(password) < constants.MinPasswordLength {
		errors = append(errors, fmt.Sprintf("length too short, requires at least %d characters", constants.MinPasswordLength))
	}

	// Check uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		errors = append(errors, "missing uppercase letter")
	}

	// Check lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		errors = append(errors, "missing lowercase letter")
	}

	// Check digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		errors = append(errors, "missing digit")
	}

	// Check special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?` + "`" + `~]`).MatchString(password)
	if !hasSpecial {
		errors = append(errors, "missing special character (e.g., !@#$%^&*)")
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
