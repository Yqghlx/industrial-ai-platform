package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// AdminHandlerNew Tests
// ============================================

func TestNewAdminHandlerNew(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)

	handler := NewAdminHandlerNew(mockAuthSvc, mockTelemetrySvc)

	assert.NotNil(t, handler)
	assert.Equal(t, mockAuthSvc, handler.authSvc)
}

func TestAdminHandlerNew_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		expectedPage   int
		expectedSize   int
		setupMock      func(*MockAuthService)
		expectedStatus int
	}{
		{
			name:           "default_pagination",
			query:          "",
			expectedPage:   1,
			expectedSize:   50,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 1, 50).Return([]model.User{}, 0, nil) },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom_pagination",
			query:          "?page=2&page_size=20",
			expectedPage:   2,
			expectedSize:   20,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 2, 20).Return([]model.User{}, 0, nil) },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "negative_page_uses_default",
			query:          "?page=-1&page_size=20",
			expectedPage:   1,
			expectedSize:   20,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 1, 20).Return([]model.User{}, 0, nil) },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid_page_uses_default",
			query:          "?page=abc&page_size=20",
			expectedPage:   1,
			expectedSize:   20,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 1, 20).Return([]model.User{}, 0, nil) },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "page_size_over_limit_uses_default",
			query:          "?page=1&page_size=200",
			expectedPage:   1,
			expectedSize:   50,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 1, 50).Return([]model.User{}, 0, nil) },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "negative_page_size_uses_default",
			query:          "?page=1&page_size=-10",
			expectedPage:   1,
			expectedSize:   50,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 1, 50).Return([]model.User{}, 0, nil) },
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service_error_returns_500",
			query:          "",
			expectedPage:   1,
			expectedSize:   50,
			setupMock:      func(m *MockAuthService) { m.On("ListUsers", mock.Anything, 1, 50).Return(nil, 0, assert.AnError) },
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockAuthSvc := new(MockAuthService)
			tt.setupMock(mockAuthSvc)

			handler := NewAdminHandlerNew(mockAuthSvc, new(MockTelemetryService))
			router.GET("/users", handler.ListUsers)

			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				assert.Equal(t, float64(0), response["total"])
				assert.Equal(t, float64(tt.expectedPage), response["page"])
				assert.Equal(t, float64(tt.expectedSize), response["page_size"])
			}
		})
	}
}

func TestAdminHandlerNew_CreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc, new(MockTelemetryService))

	router.POST("/users", handler.CreateUser)

	body := map[string]string{
		"username": "newuser",
		"password": "password123",
		"email":    "new@example.com",
		"role":     "admin",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "User created (placeholder)", response["message"])
}

func TestAdminHandlerNew_CreateUser_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc, new(MockTelemetryService))

	router.POST("/users", handler.CreateUser)

	body := map[string]string{
		"username": "newuser",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandlerNew_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc, new(MockTelemetryService))

	router.DELETE("/users/:id", handler.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "User deleted (placeholder)", response["message"])
	assert.Equal(t, "123", response["id"])
}

func TestAdminHandlerNew_GetSystemStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)
	mockTelemetrySvc := new(MockTelemetryService)
	mockTelemetrySvc.On("GetSystemStatus", mock.Anything).Return(&model.SystemStatus{
		Database:    "healthy",
		Version:     "1.0.0",
		Uptime:      "running",
		DBLatency:   0,
		DeviceCount: 0,
		UserCount:   0,
	}, nil)

	handler := NewAdminHandlerNew(mockAuthSvc, mockTelemetrySvc)

	router.GET("/system/status", handler.GetSystemStatus)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "healthy", response["database"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "1.0.0", response["version"])
}

func TestAdminHandlerNew_CreateUser_WithOptionalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockAuthSvc := new(MockAuthService)

	handler := NewAdminHandlerNew(mockAuthSvc, new(MockTelemetryService))

	router.POST("/users", handler.CreateUser)

	// Test without email and role (optional fields)
	body := map[string]string{
		"username": "minimaluser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
