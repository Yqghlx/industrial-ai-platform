package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================
// HealthHandlerNew Tests
// ============================================

func TestHealthHandlerNew_GetCacheStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandlerNew(time.Now())

	router := gin.New()
	router.GET("/cache/status", handler.GetCacheStatus)

	req := httptest.NewRequest(http.MethodGet, "/cache/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHealthHandlerNew_GetWSStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandlerNew(time.Now())

	router := gin.New()
	router.GET("/ws/stats", handler.GetWSStats)

	req := httptest.NewRequest(http.MethodGet, "/ws/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
