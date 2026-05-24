package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/industrial-ai/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestTenantHandler_CreateTenant 测试创建租户方法
func TestTenantHandler_CreateTenant(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(m *MockTenantService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "successful create tenant",
			requestBody: model.TenantCreateRequest{
				Name: "Test Tenant",
				Slug: "test-tenant",
				Plan: "pro",
			},
			mockSetup: func(m *MockTenantService) {
				m.On("CreateTenant", mock.Anything, "Test Tenant", "test-tenant", "pro", 0).Return(&model.Tenant{
					ID:         "tenant-001",
					Name:       "Test Tenant",
					Slug:       "test-tenant",
					Plan:       "pro",
					MaxDevices: 100,
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "invalid request - missing name",
			requestBody: map[string]interface{}{
				"slug": "test-tenant",
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "invalid request - missing slug",
			requestBody: map[string]interface{}{
				"name": "Test Tenant",
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "service error - slug exists",
			requestBody: model.TenantCreateRequest{
				Name: "Test Tenant",
				Slug: "existing-slug",
				Plan: "free",
			},
			mockSetup: func(m *MockTenantService) {
				m.On("CreateTenant", mock.Anything, "Test Tenant", "existing-slug", "free", 0).Return(nil, errors.New("tenant slug already exists"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name: "create tenant with default plan",
			requestBody: model.TenantCreateRequest{
				Name: "New Tenant",
				Slug: "new-tenant",
			},
			mockSetup: func(m *MockTenantService) {
				m.On("CreateTenant", mock.Anything, "New Tenant", "new-tenant", "", 0).Return(&model.Tenant{
					ID:         "tenant-002",
					Name:       "New Tenant",
					Slug:       "new-tenant",
					Plan:       "free",
					MaxDevices: 10,
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTenantService := new(MockTenantService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockTenantService)
			}

			// 创建 handler，直接传入 MockTenantService 指针
			handler := NewTenantHandler(mockTenantService)

			router := gin.New()
			router.POST("/tenants", handler.CreateTenant)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/tenants", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response model.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, response.Success)

			mockTenantService.AssertExpectations(t)
		})
	}
}

// TestTenantHandler_ListTenants 测试列出租户方法
func TestTenantHandler_ListTenants(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(m *MockTenantService)
		expectedStatus int
		expectSuccess  bool
		expectedTotal  int
	}{
		{
			name:        "successful list tenants",
			queryParams: "",
			mockSetup: func(m *MockTenantService) {
				m.On("ListTenants", mock.Anything, 50, 0).Return([]model.Tenant{
					{ID: "tenant-001", Name: "Tenant 1", Slug: "tenant-1", Plan: "pro"},
					{ID: "tenant-002", Name: "Tenant 2", Slug: "tenant-2", Plan: "free"},
				}, nil)
				m.On("CountTenants", mock.Anything).Return(2, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			expectedTotal:  2,
		},
		{
			name:        "list tenants with pagination",
			queryParams: "?limit=10&offset=20",
			mockSetup: func(m *MockTenantService) {
				m.On("ListTenants", mock.Anything, 10, 20).Return([]model.Tenant{}, nil)
				m.On("CountTenants", mock.Anything).Return(50, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			expectedTotal:  50,
		},
		{
			name:        "list tenants service error",
			queryParams: "",
			mockSetup: func(m *MockTenantService) {
				m.On("ListTenants", mock.Anything, 50, 0).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name:        "list tenants with count error",
			queryParams: "",
			mockSetup: func(m *MockTenantService) {
				m.On("ListTenants", mock.Anything, 50, 0).Return([]model.Tenant{
					{ID: "tenant-001", Name: "Tenant 1", Slug: "tenant-1", Plan: "pro"},
				}, nil)
				m.On("CountTenants", mock.Anything).Return(0, errors.New("count error"))
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			expectedTotal:  0, // count error defaults to 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTenantService := new(MockTenantService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockTenantService)
			}

			handler := NewTenantHandler(mockTenantService)

			router := gin.New()
			router.GET("/tenants", handler.ListTenants)

			req, _ := http.NewRequest("GET", "/tenants"+tt.queryParams, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response model.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, response.Success)

			if tt.expectSuccess {
				data, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(tt.expectedTotal), data["total"])
			}

			mockTenantService.AssertExpectations(t)
		})
	}
}

// TestTenantHandler_GetTenant 测试获取租户方法
func TestTenantHandler_GetTenant(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c *gin.Context, tenantID string)
		tenantID       string
		mockSetup      func(m *MockTenantService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:     "successful get tenant",
			tenantID: "tenant-001",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup: func(m *MockTenantService) {
				m.On("GetTenant", mock.Anything, "tenant-001").Return(&model.Tenant{
					ID:         "tenant-001",
					Name:       "Test Tenant",
					Slug:       "test-tenant",
					Plan:       "pro",
					MaxDevices: 100,
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:     "tenant not found",
			tenantID: "non-existent",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup: func(m *MockTenantService) {
				m.On("GetTenant", mock.Anything, "non-existent").Return(nil, errors.New("tenant not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:     "empty tenant id - handled by handler",
			tenantID: "",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTenantService := new(MockTenantService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockTenantService)
			}

			handler := NewTenantHandler(mockTenantService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/tenants/"+tt.tenantID, nil)

			tt.setupContext(c, tt.tenantID)
			handler.GetTenant(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response model.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, response.Success)

			mockTenantService.AssertExpectations(t)
		})
	}
}

// TestTenantHandler_UpdateTenant 测试更新租户方法
func TestTenantHandler_UpdateTenant(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c *gin.Context, tenantID string)
		tenantID       string
		requestBody    interface{}
		mockSetup      func(m *MockTenantService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:     "successful update tenant",
			tenantID: "tenant-001",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			requestBody: model.TenantUpdateRequest{
				Name:       "Updated Tenant",
				Plan:       "enterprise",
				MaxDevices: 500,
			},
			mockSetup: func(m *MockTenantService) {
				m.On("UpdateTenant", mock.Anything, "tenant-001", mock.AnythingOfType("map[string]interface {}")).Return(&model.Tenant{
					ID:         "tenant-001",
					Name:       "Updated Tenant",
					Slug:       "test-tenant",
					Plan:       "enterprise",
					MaxDevices: 500,
					IsActive:   true,
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:     "update with partial fields",
			tenantID: "tenant-001",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			requestBody: map[string]interface{}{
				"name": "Only Name Updated",
			},
			mockSetup: func(m *MockTenantService) {
				m.On("UpdateTenant", mock.Anything, "tenant-001", mock.AnythingOfType("map[string]interface {}")).Return(&model.Tenant{
					ID:         "tenant-001",
					Name:       "Only Name Updated",
					Slug:       "test-tenant",
					Plan:       "pro",
					MaxDevices: 100,
					IsActive:   true,
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:     "empty tenant id - handled by handler",
			tenantID: "",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			requestBody: model.TenantUpdateRequest{
				Name: "Updated Tenant",
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:     "tenant not found",
			tenantID: "non-existent",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			requestBody: model.TenantUpdateRequest{
				Name: "Updated Tenant",
			},
			mockSetup: func(m *MockTenantService) {
				m.On("UpdateTenant", mock.Anything, "non-existent", mock.AnythingOfType("map[string]interface {}")).Return(nil, errors.New("tenant not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name:     "empty request body - still valid",
			tenantID: "tenant-001",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			requestBody: map[string]interface{}{},
			mockSetup: func(m *MockTenantService) {
				m.On("UpdateTenant", mock.Anything, "tenant-001", mock.AnythingOfType("map[string]interface {}")).Return(&model.Tenant{
					ID:         "tenant-001",
					Name:       "Test Tenant",
					Slug:       "test-tenant",
					Plan:       "pro",
					MaxDevices: 100,
					IsActive:   true,
					UpdatedAt:  time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTenantService := new(MockTenantService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockTenantService)
			}

			handler := NewTenantHandler(mockTenantService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPut, "/tenants/"+tt.tenantID, bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			tt.setupContext(c, tt.tenantID)
			handler.UpdateTenant(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response model.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, response.Success)

			mockTenantService.AssertExpectations(t)
		})
	}
}

// TestTenantHandler_DeleteTenant 测试删除租户方法
func TestTenantHandler_DeleteTenant(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c *gin.Context, tenantID string)
		tenantID       string
		mockSetup      func(m *MockTenantService)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:     "successful delete tenant",
			tenantID: "tenant-001",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup: func(m *MockTenantService) {
				m.On("DeleteTenant", mock.Anything, "tenant-001").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:     "tenant not found",
			tenantID: "non-existent",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup: func(m *MockTenantService) {
				m.On("DeleteTenant", mock.Anything, "non-existent").Return(errors.New("tenant not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name:     "empty tenant id - handled by handler",
			tenantID: "",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant-002",
			setupContext: func(c *gin.Context, tenantID string) {
				c.Params = gin.Params{{Key: "id", Value: tenantID}}
			},
			mockSetup: func(m *MockTenantService) {
				m.On("DeleteTenant", mock.Anything, "tenant-002").Return(errors.New("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTenantService := new(MockTenantService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockTenantService)
			}

			handler := NewTenantHandler(mockTenantService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/tenants/"+tt.tenantID, nil)

			tt.setupContext(c, tt.tenantID)
			handler.DeleteTenant(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response model.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, response.Success)

			if tt.expectSuccess {
				assert.Equal(t, "tenant deleted", response.Message)
			}

			mockTenantService.AssertExpectations(t)
		})
	}
}

// TestTenantHandler_RegisterTenantRoutes 测试路由注册
func TestTenantHandler_RegisterTenantRoutes(t *testing.T) {
	mockTenantService := new(MockTenantService)
	handler := NewTenantHandler(mockTenantService)

	router := gin.New()
	api := router.Group("/api/v1")

	// 注册路由，使用空的中间件
	RegisterTenantRoutes(api, handler, func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() })

	// 测试路由是否正确注册
	routes := router.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path+"-"+route.Method] = true
	}

	// 验证所有租户路由都已注册
	assert.True(t, routePaths["/api/v1/tenants-POST"], "POST /tenants route should be registered")
	assert.True(t, routePaths["/api/v1/tenants-GET"], "GET /tenants route should be registered")
	assert.True(t, routePaths["/api/v1/tenants/:id-GET"], "GET /tenants/:id route should be registered")
	assert.True(t, routePaths["/api/v1/tenants/:id-PUT"], "PUT /tenants/:id route should be registered")
	assert.True(t, routePaths["/api/v1/tenants/:id-DELETE"], "DELETE /tenants/:id route should be registered")
}

// TestTenantModel 保留原有的模型测试
func TestTenantModel(t *testing.T) {
	tenant := model.Tenant{
		ID:         "tenant-001",
		Name:       "Test Tenant",
		Slug:       "test-tenant",
		Plan:       "pro",
		MaxDevices: 100,
		IsActive:   true,
		Settings:   `{"email_notifications": true}`,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	assert.Equal(t, "tenant-001", tenant.ID)
	assert.Equal(t, "Test Tenant", tenant.Name)
	assert.Equal(t, "test-tenant", tenant.Slug)
	assert.Equal(t, "pro", tenant.Plan)
	assert.Equal(t, 100, tenant.MaxDevices)
	assert.True(t, tenant.IsActive)
	assert.NotZero(t, tenant.CreatedAt)
	assert.NotZero(t, tenant.UpdatedAt)
}

// TestTenantPlanLimits 保留原有的计划限制测试
func TestTenantPlanLimits(t *testing.T) {
	tests := []struct {
		plan          model.TenantPlan
		expectedLimit int
	}{
		{model.PlanFree, 10},
		{model.PlanPro, 100},
		{model.PlanEnterprise, -1},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			limits := model.PlanLimits[tt.plan]
			assert.Equal(t, tt.expectedLimit, limits.MaxDevices)
		})
	}
}

// TestTenantSettings 保留原有的设置测试
func TestTenantSettings(t *testing.T) {
	settings := model.TenantSettings{
		EmailNotifications: true,
		SMSAlerts:          false,
		WebhookURL:         "https://example.com/webhook",
	}

	assert.True(t, settings.EmailNotifications)
	assert.False(t, settings.SMSAlerts)
	assert.Equal(t, "https://example.com/webhook", settings.WebhookURL)
}
