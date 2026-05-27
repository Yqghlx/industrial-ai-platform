package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/pkg/errors"
)

// ErrorResponse represents a standard API error response
type ErrorResponse struct {
	Error  string `json:"error"`
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

// ErrCodeToHTTPStatus maps AppError codes to HTTP status codes
func ErrCodeToHTTPStatus(code string) int {
	switch code {
	case errors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case errors.ErrCodeNotFound:
		return http.StatusNotFound
	case errors.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case errors.ErrCodeForbidden:
		return http.StatusForbidden
	case errors.ErrCodeConflict:
		return http.StatusConflict
	case errors.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case errors.ErrCodeValidation:
		return http.StatusBadRequest
	case errors.ErrCodeDatabase:
		return http.StatusInternalServerError
	case errors.ErrCodeInternal:
		return http.StatusInternalServerError
	case errors.ErrCodeService:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// HandleError handles AppError and returns appropriate HTTP response
// If err is not an AppError, treats it as internal error
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's an AppError
	if appErr, ok := err.(*errors.AppError); ok {
		status := ErrCodeToHTTPStatus(appErr.Code)
		c.JSON(status, ErrorResponse{
			Error:  appErr.Message,
			Code:   appErr.Code,
			Detail: appErr.Detail,
		})
		return
	}

	// Generic error - treat as internal error
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: err.Error(),
		Code:  errors.ErrCodeInternal,
	})
}

// SuccessResponse represents a standard API success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Success returns a success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{Data: data})
}

// SuccessWithMessage returns a success response with message
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// Created returns a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{Data: data})
}

// NoContent returns a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest returns a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: message,
		Code:  errors.ErrCodeInvalidInput,
	})
}

// NotFound returns a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: message,
		Code:  errors.ErrCodeNotFound,
	})
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error: message,
		Code:  errors.ErrCodeUnauthorized,
	})
}

// Forbidden returns a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error: message,
		Code:  errors.ErrCodeForbidden,
	})
}

// InternalError returns a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: message,
		Code:  errors.ErrCodeInternal,
	})
}
