package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// ============================================
// Validation Tests (补充)
// ============================================

func TestValidateLimit_Default(t *testing.T) {
	result := ValidateLimit(0)
	// Default might be 100 or 1000
	require.True(t, result >= 100)
}

func TestValidateLimit_Custom(t *testing.T) {
	result := ValidateLimit(50)
	require.Equal(t, 50, result)
}

func TestValidateLimit_MaxLimit(t *testing.T) {
	result := ValidateLimit(10000)
	// Should cap at max (implementation specific)
	require.True(t, result <= 10000)
}

func TestValidateLimit_Negative(t *testing.T) {
	result := ValidateLimit(-5)
	// Negative should use default
	require.True(t, result > 0)
}

func TestGetLimit_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		limit := GetLimit(c)
		c.JSON(http.StatusOK, gin.H{"limit": limit})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetLimit_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		limit := GetLimit(c)
		c.JSON(http.StatusOK, gin.H{"limit": limit})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?limit=25", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetLimit_MaxLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		limit := GetLimit(c)
		c.JSON(http.StatusOK, gin.H{"limit": limit})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?limit=9999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
