package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	pkgerrors "github.com/industrial-ai/platform/pkg/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================
// ErrCodeToHTTPStatus Tests
// ============================================

func TestErrCodeToHTTPStatus(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{pkgerrors.ErrCodeInvalidInput, http.StatusBadRequest},
		{pkgerrors.ErrCodeNotFound, http.StatusNotFound},
		{pkgerrors.ErrCodeUnauthorized, http.StatusUnauthorized},
		{pkgerrors.ErrCodeForbidden, http.StatusForbidden},
		{pkgerrors.ErrCodeConflict, http.StatusConflict},
		{pkgerrors.ErrCodeRateLimited, http.StatusTooManyRequests},
		{pkgerrors.ErrCodeDatabase, http.StatusInternalServerError},
		{pkgerrors.ErrCodeService, http.StatusInternalServerError},
		{"unknown_code", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := ErrCodeToHTTPStatus(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================
// HandleError Tests
// ============================================

func TestHandleError_AppError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := pkgerrors.NewAppError(pkgerrors.ErrCodeNotFound, "Item not found", "item_id=123")
	HandleError(c, appErr)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Item not found")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeNotFound)
}

func TestHandleError_StandardError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	stdErr := errors.New("standard error")
	HandleError(c, stdErr)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "standard error")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeInternal)
}

func TestHandleError_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleError(c, nil)

	// Should not write anything
	assert.Equal(t, 0, w.Body.Len())
}

// ============================================
// Success Tests
// ============================================

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"key": "value"}
	Success(c, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "value")
}

func TestSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================
// SuccessWithMessage Tests
// ============================================

func TestSuccessWithMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]int{"count": 5}
	SuccessWithMessage(c, data, "Operation completed")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Operation completed")
	assert.Contains(t, w.Body.String(), "count")
}

// ============================================
// Created Tests
// ============================================

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"id": "new-123"}
	Created(c, data)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "new-123")
}

// ============================================
// NoContent Tests
// ============================================

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NoContent(c)

	// gin.Status() doesn't write immediately, we need to finalize
	c.Writer.WriteHeaderNow()
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ============================================
// BadRequest Tests
// ============================================

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "Invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid input")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeInvalidInput)
}

// ============================================
// NotFound Tests
// ============================================

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "Device not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Device not found")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeNotFound)
}

// ============================================
// Unauthorized Tests
// ============================================

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "Invalid credentials")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid credentials")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeUnauthorized)
}

// ============================================
// Forbidden Tests
// ============================================

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "Access denied")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Access denied")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeForbidden)
}

// ============================================
// InternalError Tests
// ============================================

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	InternalError(c, "Database connection failed")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Database connection failed")
	assert.Contains(t, w.Body.String(), pkgerrors.ErrCodeInternal)
}

// ============================================
// Integration Tests
// ============================================

func TestHandleError_FullFlow(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		HandleError(c, pkgerrors.NewAppError(pkgerrors.ErrCodeConflict, "Resource already exists", ""))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}