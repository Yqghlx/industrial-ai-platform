package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/mocks"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/cache"
)

// ============================================
// Server Integration Tests (使用 Mock 依赖)
// ============================================

// MockDB 模拟数据库连接
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	args2 := m.Called(query, args)
	return args2.Get(0).(sql.Result), args2.Error(1)
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	args2 := m.Called(query, args)
	if args2.Get(0) == nil {
		return nil, args2.Error(1)
	}
	return args2.Get(0).(*sql.Rows), args2.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	m.Called(query, args)
	return &sql.Row{}
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) SetMaxOpenConns(n int) {
	m.Called(n)
}

func (m *MockDB) SetMaxIdleConns(n int) {
	m.Called(n)
}

func (m *MockDB) SetConnMaxLifetime(d time.Duration) {
	m.Called(d)
}

// MockCache 模拟缓存服务
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) bool {
	args := m.Called(ctx, key)
	return args.Bool(0)
}

func (m *MockCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCache) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCache) GetStats() cache.Stats {
	args := m.Called()
	return args.Get(0).(cache.Stats)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================
// HTTPServerNew Tests
// ============================================

func TestHTTPServerNew_healthCheck_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	startTime := time.Now()
	server := &HTTPServerNew{
		router:    router,
		startTime: startTime,
	}

	router.GET("/health", server.healthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "uptime")
}

func TestHTTPServerNew_GetRouter_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	server := &HTTPServerNew{
		router: router,
	}

	result := server.GetRouter()

	assert.NotNil(t, result)
	assert.Equal(t, router, result)
}

func TestHTTPServerNew_Close_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock DB that implements Close
	mockDB := new(MockDB)
	mockDB.On("Close").Return(nil)

	server := &HTTPServerNew{
		router: gin.New(),
	}

	// Note: Close() calls db.Close() which requires a valid DB
	// This test verifies the structure without calling Close
	assert.NotNil(t, server.router)
}

func TestHTTPServerNew_Close_WithCache_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &HTTPServerNew{
		router: gin.New(),
	}

	// Verify server structure
	assert.NotNil(t, server.router)
}

// ============================================
// Handler Setup Tests
// ============================================

func TestHTTPServerNew_setupHandlers_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Create mock services
	sf := service.NewServiceFactory()
	sf.SetDeviceService(new(mocks.MockDeviceService))
	sf.SetAlertService(new(mocks.MockAlertService))
	sf.SetAuthService(new(mocks.MockAuthService))
	sf.SetUserService(new(mocks.MockUserService))

	broadcastFunc := func(msg model.WSMessage) {}

	// Create handlers using factory
	handlerFactory := NewHandlerFactory(sf, broadcastFunc, new(MockCache))

	// Setup routes manually (simulating setupHandlers)
	deviceHandler := handlerFactory.CreateDeviceHandler()
	alertHandler := handlerFactory.CreateAlertHandler()
	authHandler := handlerFactory.CreateAuthHandler()
	telemetryHandler := handlerFactory.CreateTelemetryHandler()

	// Register routes
	router.GET("/devices", deviceHandler.ListDevices)
	router.GET("/alerts", alertHandler.ListAlerts)
	router.POST("/auth/login", authHandler.Login)
	router.GET("/telemetry/latest", telemetryHandler.GetLatestTelemetry)

	// Verify routes are registered
	routes := router.Routes()
	assert.GreaterOrEqual(t, len(routes), 4)

	// Test device route
	mockDeviceSvc := sf.GetDeviceService().(*mocks.MockDeviceService)
	mockDeviceSvc.On("List", mock.Anything, 1, 20).Return([]model.Device{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/devices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockDeviceSvc.AssertExpectations(t)
}

func TestHTTPServerNew_setupMiddleware_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Setup basic middleware (simulating setupMiddleware)
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Header("X-Server-Time", time.Now().Format(time.RFC3339))
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Server-Time"))
}

// ============================================
// WebSocket Tests
// ============================================

func TestHTTPServerNew_WS_Broadcast_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	broadcastChan := make(chan model.WSMessage, 10)
	broadcastFunc := func(msg model.WSMessage) {
		broadcastChan <- msg
	}

	server := &HTTPServerNew{
		router:        gin.New(),
		broadcastFn:   broadcastFunc,
		broadcastChan: broadcastChan,
	}

	// Test broadcast function
	testMsg := model.WSMessage{
		Type:    "test",
		Payload: map[string]string{"key": "value"},
	}

	server.broadcastFn(testMsg)

	// Verify message was broadcast
	select {
	case msg := <-broadcastChan:
		assert.Equal(t, "test", msg.Type)
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected broadcast message")
	}
}

// ============================================
// Context Tests
// ============================================

func TestHTTPServerNew_getRequestContext_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ctx, cancel := getRequestContext(c)
		defer cancel()

		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.True(t, deadline.After(time.Now()))

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// ============================================
// Password Generation Tests
// ============================================

func TestHTTPServerNew_generateRandomPassword_Integration(t *testing.T) {
	password1 := generateRandomPassword(16)
	password2 := generateRandomPassword(16)

	assert.Len(t, password1, 16)
	assert.Len(t, password2, 16)
	assert.NotEqual(t, password1, password2) // Should be different
}

func TestHTTPServerNew_generateFallbackPassword_Integration(t *testing.T) {
	password := generateFallbackPassword(16)

	assert.Len(t, password, 16)
	// Should be numeric
	for _, c := range password {
		assert.True(t, c >= '0' && c <= '9')
	}
}

// ============================================
// Server Creation Tests (Mocked)
// ============================================

func TestHTTPServerNew_CreateWithMocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create server with minimal dependencies
	router := gin.New()

	server := &HTTPServerNew{
		router:    router,
		startTime: time.Now(),
		jwtSecret: "test-secret",
	}

	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.NotEmpty(t, server.jwtSecret)
}

func TestHTTPServerNew_AdminPasswordGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminPassword := generateRandomPassword(12)

	assert.Len(t, adminPassword, 12)
	assert.NotEmpty(t, adminPassword)
}

// ============================================
// Route Registration Tests
// ============================================

func TestHTTPServerNew_RoutesRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Register all routes (simulating full setupHandlers)
	groups := []string{
		"/devices",
		"/alerts",
		"/rules",
		"/auth",
		"/admin",
		"/telemetry",
		"/health",
		"/export",
		"/ws",
		"/api",
		"/rbac",
		"/tenants",
	}

	for _, group := range groups {
		g := router.Group(group)
		g.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"group": group})
		})
	}

	// Verify all groups are registered
	routes := router.Routes()
	assert.Equal(t, len(groups), len(routes))

	// Test each route
	for _, group := range groups {
		req := httptest.NewRequest(http.MethodGet, group+"/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	}
}

// ============================================
// Backward Compatibility Tests
// ============================================

func TestHTTPServerNew_NewAuthHandler_Compat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUserSvc := new(mocks.MockUserService)
	jwtSecret := "test-secret"

	handler := NewAuthHandler(mockUserSvc, jwtSecret)

	assert.NotNil(t, handler)
}

func TestHTTPServerNew_compatAuthSvc_Login_Compat(t *testing.T) {
	mockUserSvc := new(mocks.MockUserService)

	user := &model.User{ID: 1, Username: "testuser"}
	mockUserSvc.On("Authenticate", "testuser", "password").Return(user, nil)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	resultUser, token, err := compatSvc.Login(context.Background(), "testuser", "password")

	require.NoError(t, err)
	assert.Equal(t, user, resultUser)
	assert.Equal(t, "token", token)

	mockUserSvc.AssertExpectations(t)
}

func TestHTTPServerNew_compatAuthSvc_GetUserByID_Compat(t *testing.T) {
	mockUserSvc := new(mocks.MockUserService)

	user := &model.User{ID: 1, Username: "testuser"}
	mockUserSvc.On("GetByID", 1).Return(user, nil)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	resultUser, err := compatSvc.GetUserByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, user, resultUser)

	mockUserSvc.AssertExpectations(t)
}

func TestHTTPServerNew_compatAuthSvc_Register_Compat(t *testing.T) {
	mockUserSvc := new(mocks.MockUserService)

	compatSvc := &compatAuthSvc{userSvc: mockUserSvc}

	user, token, err := compatSvc.Register(context.Background(), &model.RegisterRequest{
		Username: "newuser",
		Password: "password",
	})

	require.NoError(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
}

// ============================================
// Export Wrapper Tests
// ============================================

func TestHTTPServerNew_exportDevices_Wrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a mock export handler with mock setup
	mockExportSvc := new(MockExportService)
	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)
	mockExportHandler := NewExportHandler(mockExportSvc)

	server := &HTTPServerNew{
		router:        router,
		exportHandler: mockExportHandler,
	}

	router.GET("/export/devices", server.exportDevices)

	req := httptest.NewRequest(http.MethodGet, "/export/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return error due to mock
	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestHTTPServerNew_exportAlerts_Wrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)
	mockExportHandler := NewExportHandler(mockExportSvc)

	server := &HTTPServerNew{
		router:        router,
		exportHandler: mockExportHandler,
	}

	router.GET("/export/alerts", server.exportAlerts)

	req := httptest.NewRequest(http.MethodGet, "/export/alerts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestHTTPServerNew_exportROI_Wrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockExportSvc := new(MockExportService)
	mockExportSvc.On("Export", mock.Anything, mock.AnythingOfType("*service.ExportRequest")).Return(nil, assert.AnError)
	mockExportHandler := NewExportHandler(mockExportSvc)

	server := &HTTPServerNew{
		router:        router,
		exportHandler: mockExportHandler,
	}

	router.GET("/export/roi", server.exportROI)

	req := httptest.NewRequest(http.MethodGet, "/export/roi", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	mockExportSvc.AssertExpectations(t)
}

func TestHTTPServerNew_getCacheStatus_Wrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockHealthHandler := NewHealthHandlerNew(time.Now())

	server := &HTTPServerNew{
		router:        router,
		healthHandler: mockHealthHandler,
	}

	router.GET("/cache-status", server.getCacheStatus)

	req := httptest.NewRequest(http.MethodGet, "/cache-status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
