package errors

import (
	"errors"
	"testing"
)

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("Device", "device-123")
	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Message != "Device not found" {
		t.Errorf("expected message 'Device not found', got %s", err.Message)
	}
	if err.Detail != "device-123" {
		t.Errorf("expected detail 'device-123', got %s", err.Detail)
	}
}

func TestNewDeviceNotFoundError(t *testing.T) {
	err := NewDeviceNotFoundError("dev-001")
	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Detail != "dev-001" {
		t.Errorf("expected detail 'dev-001', got %s", err.Detail)
	}
}

func TestNewInvalidInputError(t *testing.T) {
	err := NewInvalidInputError("invalid format")
	if err.Code != ErrCodeInvalidInput {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidInput, err.Code)
	}
	if err.Message != "Invalid input provided" {
		t.Errorf("expected default message, got %s", err.Message)
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("token expired")
	if err.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %s, got %s", ErrCodeUnauthorized, err.Code)
	}
	if err.Message != "token expired" {
		t.Errorf("expected custom message, got %s", err.Message)
	}
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("delete resource")
	if err.Code != ErrCodeForbidden {
		t.Errorf("expected code %s, got %s", ErrCodeForbidden, err.Code)
	}
	if err.Message != "Access denied" {
		t.Errorf("expected default message, got %s", err.Message)
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("database connection failed")
	if err.Code != ErrCodeInternal {
		t.Errorf("expected code %s, got %s", ErrCodeInternal, err.Code)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email format invalid")
	if err.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, err.Code)
	}
}

func TestNewRateLimitedError(t *testing.T) {
	err := NewRateLimitedError()
	if err.Code != ErrCodeRateLimited {
		t.Errorf("expected code %s, got %s", ErrCodeRateLimited, err.Code)
	}
}

func TestGetAppError(t *testing.T) {
	// Test wrapping standard error
	stdErr := errors.New("standard error")
	appErr := GetAppError(stdErr)
	if appErr.Code != ErrCodeInternal {
		t.Errorf("expected code %s, got %s", ErrCodeInternal, appErr.Code)
	}

	// Test wrapping nil
	nilErr := GetAppError(nil)
	if nilErr != nil {
		t.Error("expected nil for nil input")
	}

	// Test wrapping existing AppError
	existingErr := NewNotFoundError("Test", "123")
	wrappedErr := GetAppError(existingErr)
	if wrappedErr.Code != ErrCodeNotFound {
		t.Errorf("expected preserved code %s, got %s", ErrCodeNotFound, wrappedErr.Code)
	}
}

func TestIsAppError(t *testing.T) {
	appErr := NewNotFoundError("Test", "123")
	if !IsAppError(appErr) {
		t.Error("expected true for AppError")
	}

	stdErr := errors.New("standard error")
	if IsAppError(stdErr) {
		t.Error("expected false for standard error")
	}

	if IsAppError(nil) {
		t.Error("expected false for nil")
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		err      error
		expected int
	}{
		{NewNotFoundError("Test", "123"), 404},
		{NewUnauthorizedError("test"), 401},
		{NewForbiddenError("test"), 403},
		{NewInvalidInputError("test"), 400},
		{NewValidationError("test"), 400},
		{NewInternalError("test"), 500},
		{NewRateLimitedError(), 429},
		{errors.New("unknown"), 500},
	}

	for _, tt := range tests {
		result := GetStatusCode(tt.err)
		if result != tt.expected {
			t.Errorf("GetStatusCode(%v) = %d, expected %d", tt.err, result, tt.expected)
		}
	}
}

func TestErrorString(t *testing.T) {
	// Test with detail
	err := NewNotFoundError("Device", "dev-123")
	expected := "NOT_FOUND: Device not found (dev-123)"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}

	// Test without detail
	err2 := NewUnauthorizedError("test")
	expected2 := "UNAUTHORIZED: test"
	if err2.Error() != expected2 {
		t.Errorf("expected '%s', got '%s'", expected2, err2.Error())
	}
}

func TestNewAppError(t *testing.T) {
	err := NewAppError("CUSTOM_CODE", "Custom message", "Custom detail")
	if err.Code != "CUSTOM_CODE" {
		t.Errorf("expected code 'CUSTOM_CODE', got %s", err.Code)
	}
	if err.Message != "Custom message" {
		t.Errorf("expected message 'Custom message', got %s", err.Message)
	}
	if err.Detail != "Custom detail" {
		t.Errorf("expected detail 'Custom detail', got %s", err.Detail)
	}
}

func TestNewAppErrorWithoutDetail(t *testing.T) {
	err := NewAppError("CODE", "message", "")
	if err.Detail != "" {
		t.Errorf("expected empty detail, got %s", err.Detail)
	}
	expected := "CODE: message"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNewTenantNotFoundError(t *testing.T) {
	err := NewTenantNotFoundError("tenant-001")
	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Message != "Tenant not found" {
		t.Errorf("expected message 'Tenant not found', got %s", err.Message)
	}
	if err.Detail != "tenant-001" {
		t.Errorf("expected detail 'tenant-001', got %s", err.Detail)
	}
}

func TestNewUserNotFoundError(t *testing.T) {
	err := NewUserNotFoundError("user-123")
	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Message != "User not found" {
		t.Errorf("expected message 'User not found', got %s", err.Message)
	}
	if err.Detail != "user-123" {
		t.Errorf("expected detail 'user-123', got %s", err.Detail)
	}
}

func TestNewTenantInvalidIDError(t *testing.T) {
	err := NewTenantInvalidIDError("invalid-id")
	if err.Code != ErrCodeInvalidInput {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidInput, err.Code)
	}
	if err.Message != "Invalid input provided" {
		t.Errorf("expected default message, got %s", err.Message)
	}
	expectedDetail := "Invalid tenant ID format: invalid-id"
	if err.Detail != expectedDetail {
		t.Errorf("expected detail '%s', got '%s'", expectedDetail, err.Detail)
	}
}

func TestNewDeviceInvalidIDError(t *testing.T) {
	err := NewDeviceInvalidIDError("bad-device-id")
	if err.Code != ErrCodeInvalidInput {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidInput, err.Code)
	}
	if err.Message != "Invalid input provided" {
		t.Errorf("expected default message, got %s", err.Message)
	}
	expectedDetail := "Invalid device ID format: bad-device-id"
	if err.Detail != expectedDetail {
		t.Errorf("expected detail '%s', got '%s'", expectedDetail, err.Detail)
	}
}

func TestNewAuthFailedError(t *testing.T) {
	err := NewAuthFailedError()
	if err.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %s, got %s", ErrCodeUnauthorized, err.Code)
	}
	if err.Message != "Authentication failed" {
		t.Errorf("expected message 'Authentication failed', got %s", err.Message)
	}
}

func TestNewDatabaseError(t *testing.T) {
	err := NewDatabaseError("connection timeout")
	if err.Code != ErrCodeDatabase {
		t.Errorf("expected code %s, got %s", ErrCodeDatabase, err.Code)
	}
	if err.Message != "Database operation failed" {
		t.Errorf("expected message 'Database operation failed', got %s", err.Message)
	}
	if err.Detail != "connection timeout" {
		t.Errorf("expected detail 'connection timeout', got %s", err.Detail)
	}
}

func TestGetStatusCodeAllCases(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected int
	}{
		{"NotFound", NewNotFoundError("Test", "123"), 404},
		{"Unauthorized", NewUnauthorizedError("test"), 401},
		{"Forbidden", NewForbiddenError("test"), 403},
		{"InvalidInput", NewInvalidInputError("test"), 400},
		{"Validation", NewValidationError("test"), 400},
		{"Internal", NewInternalError("test"), 500},
		{"RateLimited", NewRateLimitedError(), 429},
		{"Database", NewDatabaseError("test"), 500},
		{"Conflict", NewAppError(ErrCodeConflict, "conflict", ""), 409},
		{"Service", NewAppError(ErrCodeService, "service error", ""), 500},
		{"UnknownCode", NewAppError("UNKNOWN_CODE", "unknown", ""), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusCode(tt.err)
			if result != tt.expected {
				t.Errorf("GetStatusCode(%v) = %d, expected %d", tt.err, result, tt.expected)
			}
		})
	}
}

func TestErrorStringWithDetail(t *testing.T) {
	err := &AppError{
		Code:    "TEST_CODE",
		Message: "Test message",
		Detail:  "Test detail",
	}
	expected := "TEST_CODE: Test message (Test detail)"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorStringWithoutDetail(t *testing.T) {
	err := &AppError{
		Code:    "TEST_CODE",
		Message: "Test message",
	}
	expected := "TEST_CODE: Test message"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestIsAppErrorNil(t *testing.T) {
	result := IsAppError(nil)
	if result {
		t.Error("IsAppError(nil) should return false")
	}
}

func TestCommonErrors(t *testing.T) {
	// Test predefined errors exist
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
	if ErrUnauthorized == nil {
		t.Error("ErrUnauthorized should not be nil")
	}
	if ErrForbidden == nil {
		t.Error("ErrForbidden should not be nil")
	}
	if ErrInternal == nil {
		t.Error("ErrInternal should not be nil")
	}
	if ErrInvalidInput == nil {
		t.Error("ErrInvalidInput should not be nil")
	}
	if ErrRateLimited == nil {
		t.Error("ErrRateLimited should not be nil")
	}
}
