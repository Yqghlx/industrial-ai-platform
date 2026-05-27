// Package errors provides unified error handling
// FIX-029: Unified error types
package errors

import "fmt"

// Error codes
const (
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeRateLimited  = "RATE_LIMITED"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeDatabase     = "DATABASE_ERROR"
	ErrCodeService      = "SERVICE_ERROR"
)

// AppError represents a unified application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new AppError
func NewAppError(code, message, detail string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// Specific error constructors

func NewNotFoundError(entity, id string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", entity),
		Detail:  id,
	}
}

func NewDeviceNotFoundError(id string) *AppError {
	return NewNotFoundError("Device", id)
}

func NewTenantNotFoundError(id string) *AppError {
	return NewNotFoundError("Tenant", id)
}

func NewUserNotFoundError(id string) *AppError {
	return NewNotFoundError("User", id)
}

func NewInvalidInputError(detail string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidInput,
		Message: "Invalid input provided",
		Detail:  detail,
	}
}

func NewTenantInvalidIDError(id string) *AppError {
	return NewInvalidInputError(fmt.Sprintf("Invalid tenant ID format: %s", id))
}

func NewDeviceInvalidIDError(id string) *AppError {
	return NewInvalidInputError(fmt.Sprintf("Invalid device ID format: %s", id))
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

func NewAuthFailedError() *AppError {
	return NewUnauthorizedError("Authentication failed")
}

func NewForbiddenError(action string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: "Access denied",
		Detail:  action,
	}
}

func NewInternalError(detail string) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: "Internal server error",
		Detail:  detail,
	}
}

func NewValidationError(detail string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: "Validation failed",
		Detail:  detail,
	}
}

func NewDatabaseError(detail string) *AppError {
	return &AppError{
		Code:    ErrCodeDatabase,
		Message: "Database operation failed",
		Detail:  detail,
	}
}

func NewRateLimitedError() *AppError {
	return &AppError{
		Code:    ErrCodeRateLimited,
		Message: "Too many requests",
	}
}

// GetAppError wraps a standard error into AppError
func GetAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewInternalError(err.Error())
}

// IsAppError checks if an error is AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetStatusCode returns HTTP status code for AppError
func GetStatusCode(err error) int {
	appErr := GetAppError(err)
	switch appErr.Code {
	case ErrCodeNotFound:
		return 404
	case ErrCodeUnauthorized:
		return 401
	case ErrCodeForbidden:
		return 403
	case ErrCodeInvalidInput, ErrCodeValidation:
		return 400
	case ErrCodeConflict:
		return 409
	case ErrCodeRateLimited:
		return 429
	case ErrCodeInternal, ErrCodeDatabase, ErrCodeService:
		return 500
	default:
		return 500
	}
}

// Common errors for reuse
var (
	ErrNotFound     = NewNotFoundError("Resource", "")
	ErrUnauthorized = NewUnauthorizedError("Unauthorized")
	ErrForbidden    = NewForbiddenError("")
	ErrInternal     = NewInternalError("")
	ErrInvalidInput = NewInvalidInputError("")
	ErrRateLimited  = NewRateLimitedError()
)
