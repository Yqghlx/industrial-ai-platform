package model

import (
	"encoding/json"
	"strings"
	"testing"

	"industrial-ai-platform/pkg/constants"
)

// TestValidatePassword tests password validation
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid password",
			password: "Admin@123456!",
			wantErr:  false,
		},
		{
			name:        "too short",
			password:    "Abc@1",
			wantErr:     true,
			errContains: "长度不足",
		},
		{
			name:        "missing uppercase",
			password:    "admin@123456!",
			wantErr:     true,
			errContains: "大写字母",
		},
		{
			name:        "missing lowercase",
			password:    "ADMIN@123456!",
			wantErr:     true,
			errContains: "小写字母",
		},
		{
			name:        "missing digit",
			password:    "Admin@Password!",
			wantErr:     true,
			errContains: "数字",
		},
		{
			name:        "missing special",
			password:    "Admin12345678",
			wantErr:     true,
			errContains: "特殊字符",
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     true,
			errContains: "长度不足",
		},
		{
			name:     "complex valid",
			password: "MyP@ssw0rd!2024#",
			wantErr:  false,
		},
		{
			name:        "11 chars - just under minimum",
			password:    "Admin@12345",
			wantErr:     true,
			errContains: "长度不足",
		},
		{
			name:     "12 chars exactly",
			password: "Admin@123456",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePassword(%s) expected error, got nil", tt.password)
					return
				}

				pwdErr, ok := err.(*PasswordValidationError)
				if !ok {
					t.Errorf("Expected PasswordValidationError, got %T", err)
					return
				}

				if tt.errContains != "" && !strings.Contains(pwdErr.Error(), tt.errContains) {
					t.Errorf("Error should contain '%s', got '%s'", tt.errContains, pwdErr.Error())
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword(%s) unexpected error: %v", tt.password, err)
				}
			}
		})
	}
}

// TestPasswordValidationError_Error tests error message format
func TestPasswordValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   []string
		expected string
	}{
		{
			name:     "single error",
			errors:   []string{"长度不足"},
			expected: "密码验证失败: 长度不足",
		},
		{
			name:     "multiple errors",
			errors:   []string{"长度不足", "缺少大写字母", "缺少数字"},
			expected: "密码验证失败: 长度不足, 缺少大写字母, 缺少数字",
		},
		{
			name:     "empty errors",
			errors:   []string{},
			expected: "密码验证失败: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &PasswordValidationError{Errors: tt.errors}
			if err.Error() != tt.expected {
				t.Errorf("Error() = '%s', want '%s'", err.Error(), tt.expected)
			}
		})
	}
}

// TestValidatePassword_AllErrors tests password with all validation failures
func TestValidatePassword_AllErrors(t *testing.T) {
	password := "abc" // 3 chars, no upper, no digit, no special

	err := ValidatePassword(password)
	if err == nil {
		t.Fatal("Expected error for weak password")
	}

	pwdErr := err.(*PasswordValidationError)
	if len(pwdErr.Errors) < 4 {
		t.Errorf("Expected at least 4 errors, got %d: %v", len(pwdErr.Errors), pwdErr.Errors)
	}
}

// TestLoginRequest_JSON tests LoginRequest struct
func TestLoginRequest_JSON(t *testing.T) {
	req := LoginRequest{
		Username: "admin",
		Password: "Admin@123456!",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled LoginRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Username != req.Username {
		t.Errorf("Username = %s, want %s", unmarshaled.Username, req.Username)
	}
	if unmarshaled.Password != req.Password {
		t.Errorf("Password = %s, want %s", unmarshaled.Password, req.Password)
	}
}

// TestLoginResponse_JSON tests LoginResponse struct
func TestLoginResponse_JSON(t *testing.T) {
	resp := LoginResponse{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: UserResponse{
			ID:       1,
			Username: "admin",
			Role:     "admin",
			TenantID: "t-1",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled LoginResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.AccessToken != resp.AccessToken {
		t.Errorf("AccessToken = %s, want %s", unmarshaled.AccessToken, resp.AccessToken)
	}
	if unmarshaled.ExpiresIn != resp.ExpiresIn {
		t.Errorf("ExpiresIn = %d, want %d", unmarshaled.ExpiresIn, resp.ExpiresIn)
	}
	if unmarshaled.User.ID != resp.User.ID {
		t.Errorf("User.ID = %d, want %d", unmarshaled.User.ID, resp.User.ID)
	}
}

// TestTokenRefreshResponse_JSON tests TokenRefreshResponse
func TestTokenRefreshResponse_JSON(t *testing.T) {
	resp := TokenRefreshResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    7200,
		TokenType:    "Bearer",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled TokenRefreshResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.AccessToken != resp.AccessToken {
		t.Errorf("AccessToken = %s, want %s", unmarshaled.AccessToken, resp.AccessToken)
	}
}

// TestLogoutRequest_JSON tests LogoutRequest
func TestLogoutRequest_JSON(t *testing.T) {
	// With refresh token
	req := LogoutRequest{
		RefreshToken: "refresh-token-to-revoke",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled LogoutRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.RefreshToken != req.RefreshToken {
		t.Errorf("RefreshToken = %s, want %s", unmarshaled.RefreshToken, req.RefreshToken)
	}

	// Empty refresh token
	emptyReq := LogoutRequest{}
	emptyData, err := json.Marshal(emptyReq)
	if err != nil {
		t.Fatalf("Marshal empty failed: %v", err)
	}

	var emptyUnmarshaled LogoutRequest
	if err := json.Unmarshal(emptyData, &emptyUnmarshaled); err != nil {
		t.Fatalf("Unmarshal empty failed: %v", err)
	}
}

// TestChangePasswordRequest_JSON tests ChangePasswordRequest
func TestChangePasswordRequest_JSON(t *testing.T) {
	req := ChangePasswordRequest{
		OldPassword: "OldPass@123!",
		NewPassword: "NewPass@456!",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled ChangePasswordRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.OldPassword != req.OldPassword {
		t.Errorf("OldPassword = %s, want %s", unmarshaled.OldPassword, req.OldPassword)
	}
	if unmarshaled.NewPassword != req.NewPassword {
		t.Errorf("NewPassword = %s, want %s", unmarshaled.NewPassword, req.NewPassword)
	}
}

// TestUserResponse_JSON tests UserResponse
func TestUserResponse_JSON(t *testing.T) {
	resp := UserResponse{
		ID:       1,
		Username: "operator",
		Role:     "operator",
		TenantID: "t-1",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled UserResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ID != resp.ID {
		t.Errorf("ID = %d, want %d", unmarshaled.ID, resp.ID)
	}
	if unmarshaled.Role != resp.Role {
		t.Errorf("Role = %s, want %s", unmarshaled.Role, resp.Role)
	}
}

// TestRegisterRequest_JSON tests RegisterRequest
func TestRegisterRequest_JSON(t *testing.T) {
	req := RegisterRequest{
		Username: "newuser",
		Password: "NewUser@123!",
		Email:    "newuser@example.com",
		Role:     "viewer",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled RegisterRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Username != req.Username {
		t.Errorf("Username = %s, want %s", unmarshaled.Username, req.Username)
	}
	if unmarshaled.Email != req.Email {
		t.Errorf("Email = %s, want %s", unmarshaled.Email, req.Email)
	}
}

// TestMinPasswordLength tests the constant (P3-04: 使用 pkg/constants 中的统一常量)
func TestMinPasswordLength(t *testing.T) {
	if constants.MinPasswordLength != 12 {
		t.Errorf("MinPasswordLength = %d, want 12", constants.MinPasswordLength)
	}
}
