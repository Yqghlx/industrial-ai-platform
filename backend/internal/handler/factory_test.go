package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// Mock Handler 实现 (用于测试)
// ============================================

// MockHandler 模拟 Handler 实现
type MockHandler struct {
	name          string
	registerError error
	routes        []RouteDefinition
}

// RouteDefinition 定义路由
type RouteDefinition struct {
	Method      string
	Path        string
	HandlerFunc gin.HandlerFunc
}

// NewMockHandler 创建模拟 Handler
func NewMockHandler(name string, routes []RouteDefinition) *MockHandler {
	return &MockHandler{
		name:   name,
		routes: routes,
	}
}

// NewMockHandlerWithError 创建会返回错误的模拟 Handler
func NewMockHandlerWithError(name string, err error) *MockHandler {
	return &MockHandler{
		name:          name,
		registerError: err,
	}
}

// RegisterRoutes 实现 Handler 接口
func (m *MockHandler) RegisterRoutes(router *gin.RouterGroup) error {
	if m.registerError != nil {
		return m.registerError
	}

	for _, route := range m.routes {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Path, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Path, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Path, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Path, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Path, route.HandlerFunc)
		}
	}

	return nil
}

// Name 实现 Handler 接口
func (m *MockHandler) Name() string {
	return m.name
}

// ============================================
// RegisterAll 函数测试
// ============================================

func TestRegisterAll_NilRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试空路由器情况
	err := RegisterAll(nil)
	require.Error(t, err)
	require.Equal(t, ErrNilRouter, err)
}

func TestRegisterAll_EmptyHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 测试空处理器列表
	err := RegisterAll(router)
	require.NoError(t, err)

	// 验证路由列表为空（除了默认路由）
	routes := router.Routes()
	assert.Equal(t, 0, len(routes))
}

func TestRegisterAll_SingleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建模拟处理器
	handler := NewMockHandler("test-handler", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/test",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			},
		},
	})

	// 注册处理器
	err := RegisterAll(router, handler)
	require.NoError(t, err)

	// 验证路由已注册
	routes := router.Routes()
	require.Equal(t, 1, len(routes))
	assert.Equal(t, "/test", routes[0].Path)
	assert.Equal(t, http.MethodGet, routes[0].Method)

	// 测试路由是否工作
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegisterAll_MultipleHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建多个模拟处理器
	handler1 := NewMockHandler("handler-1", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/api/v1/resource1",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"handler": "handler-1"})
			},
		},
	})

	handler2 := NewMockHandler("handler-2", []RouteDefinition{
		{
			Method: http.MethodPost,
			Path:   "/api/v1/resource2",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"handler": "handler-2"})
			},
		},
		{
			Method: http.MethodGet,
			Path:   "/api/v1/resource2/:id",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "handler": "handler-2"})
			},
		},
	})

	handler3 := NewMockHandler("handler-3", []RouteDefinition{
		{
			Method: http.MethodPut,
			Path:   "/api/v1/resource3/:id",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"updated": c.Param("id")})
			},
		},
		{
			Method: http.MethodDelete,
			Path:   "/api/v1/resource3/:id",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusNoContent, nil)
			},
		},
	})

	// 注册所有处理器
	err := RegisterAll(router, handler1, handler2, handler3)
	require.NoError(t, err)

	// 验证路由数量
	routes := router.Routes()
	assert.Equal(t, 5, len(routes))

	// 测试每个路由
	testCases := []struct {
		method         string
		path           string
		expectedStatus int
	}{
		{http.MethodGet, "/api/v1/resource1", http.StatusOK},
		{http.MethodPost, "/api/v1/resource2", http.StatusCreated},
		{http.MethodGet, "/api/v1/resource2/123", http.StatusOK},
		{http.MethodPut, "/api/v1/resource3/456", http.StatusOK},
		{http.MethodDelete, "/api/v1/resource3/789", http.StatusNoContent},
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestRegisterAll_HandlerWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建会返回错误的处理器
	testError := errors.New("registration failed")
	handler := NewMockHandlerWithError("error-handler", testError)

	// 注册处理器应返回错误
	err := RegisterAll(router, handler)
	require.Error(t, err)

	// 验证错误类型
	handlerErr, ok := err.(*HandlerError)
	require.True(t, ok)
	assert.Equal(t, "error-handler", handlerErr.HandlerName)
	assert.Equal(t, testError, handlerErr.Err)
}

func TestRegisterAll_MixedHandlersWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建正常处理器
	goodHandler := NewMockHandler("good-handler", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/good",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			},
		},
	})

	// 创建会返回错误的处理器
	testError := errors.New("registration failed")
	errorHandler := NewMockHandlerWithError("error-handler", testError)

	// 注册处理器，错误处理器应导致返回错误
	err := RegisterAll(router, goodHandler, errorHandler)
	require.Error(t, err)

	// 验证错误信息
	assert.Contains(t, err.Error(), "error-handler")
	assert.Contains(t, err.Error(), "registration failed")
}

func TestRegisterAll_WithNilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建正常处理器
	handler := NewMockHandler("test-handler", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/test",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			},
		},
	})

	// 注册处理器列表包含 nil，应该跳过 nil 处理器
	err := RegisterAll(router, nil, handler, nil)
	require.NoError(t, err)

	// 验证只有一个路由被注册
	routes := router.Routes()
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, "/test", routes[0].Path)
}

func TestRegisterAll_RouteMappingVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建处理器并验证路由映射
	handler := NewMockHandler("mapping-test", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/health",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "healthy"})
			},
		},
		{
			Method: http.MethodPost,
			Path:   "/api/v1/users",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"user": "created"})
			},
		},
	})

	err := RegisterAll(router, handler)
	require.NoError(t, err)

	// 验证路由列表
	routes := router.Routes()
	require.Equal(t, 2, len(routes))

	// 验证路由顺序和属性
	assert.Equal(t, "/health", routes[0].Path)
	assert.Equal(t, http.MethodGet, routes[0].Method)
	assert.Equal(t, "/api/v1/users", routes[1].Path)
	assert.Equal(t, http.MethodPost, routes[1].Method)

	// 测试路由响应正确
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.JSONEq(t, `{"status": "healthy"}`, w.Body.String())

	req = httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.JSONEq(t, `{"user": "created"}`, w.Body.String())
}

// ============================================
// RegisterAllWithGroup 函数测试
// ============================================

func TestRegisterAllWithGroup_NilGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试空路由组情况
	err := RegisterAllWithGroup(nil)
	require.Error(t, err)
	require.Equal(t, ErrNilRouterGroup, err)
}

func TestRegisterAllWithGroup_WithBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	apiGroup := router.Group("/api/v1")

	// 创建处理器
	handler := NewMockHandler("api-handler", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/users",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"users": []string{"user1", "user2"}})
			},
		},
		{
			Method: http.MethodPost,
			Path:   "/users",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"user": "new"})
			},
		},
	})

	// 注册处理器到路由组
	err := RegisterAllWithGroup(apiGroup, handler)
	require.NoError(t, err)

	// 验证路由
	routes := router.Routes()
	require.Equal(t, 2, len(routes))

	// 验证完整路径包含路由组路径
	assert.Equal(t, "/api/v1/users", routes[0].Path)
	assert.Equal(t, "/api/v1/users", routes[1].Path)

	// 测试路由工作正常
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRegisterAllWithGroup_HandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	apiGroup := router.Group("/api")

	// 创建会返回错误的处理器
	testError := errors.New("group registration failed")
	handler := NewMockHandlerWithError("error-handler", testError)

	// 注册处理器应返回错误
	err := RegisterAllWithGroup(apiGroup, handler)
	require.Error(t, err)

	handlerErr, ok := err.(*HandlerError)
	require.True(t, ok)
	assert.Equal(t, "error-handler", handlerErr.HandlerName)
}

// ============================================
// HandlerError 测试
// ============================================

func TestHandlerError_Error(t *testing.T) {
	innerErr := errors.New("inner error")
	handlerErr := NewHandlerError("test-handler", innerErr)

	// 验证错误消息格式
	assert.Equal(t, "handler [test-handler] registration failed: inner error", handlerErr.Error())
}

func TestHandlerError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	handlerErr := NewHandlerError("test-handler", innerErr)

	// 验证 Unwrap 返回底层错误
	unwrapped := handlerErr.Unwrap()
	assert.Equal(t, innerErr, unwrapped)

	// 验证 errors.Is 和 errors.As 工作
	assert.True(t, errors.Is(handlerErr, innerErr))
}

// ============================================
// Handler 接口验证测试
// ============================================

func TestHandlerInterface_MockHandlerImplementation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 验证 MockHandler 实现了 Handler 接口
	var _ Handler = (*MockHandler)(nil)

	// 验证接口方法存在
	handler := NewMockHandler("test", []RouteDefinition{})
	assert.Equal(t, "test", handler.Name())
}

func TestHandlerInterface_NilHandlerSafety(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建 nil 处理器切片
	var handlers []Handler = nil

	// 注册 nil 处理器切片应该不报错
	err := RegisterAll(router, handlers...)
	require.NoError(t, err)

	// 验证没有路由被注册
	routes := router.Routes()
	assert.Equal(t, 0, len(routes))
}

// ============================================
// 路径冲突测试
// ============================================

func TestRegisterAll_PathConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建两个处理器注册相同路径（不同方法）
	handler1 := NewMockHandler("handler-1", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/same-path",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"handler": "handler-1"})
			},
		},
	})

	handler2 := NewMockHandler("handler-2", []RouteDefinition{
		{
			Method: http.MethodPost,
			Path:   "/same-path",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"handler": "handler-2"})
			},
		},
	})

	// 注册处理器
	err := RegisterAll(router, handler1, handler2)
	require.NoError(t, err)

	// 验证两个路由都被注册（不同方法）
	routes := router.Routes()
	assert.Equal(t, 2, len(routes))

	// 验证 GET 和 POST 都工作正常
	req := httptest.NewRequest(http.MethodGet, "/same-path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"handler": "handler-1"}`, w.Body.String())

	req = httptest.NewRequest(http.MethodPost, "/same-path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"handler": "handler-2"}`, w.Body.String())
}

// ============================================
// 边界条件测试
// ============================================

func TestRegisterAll_EmptyRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建没有路由的处理器
	handler := NewMockHandler("empty-handler", []RouteDefinition{})

	// 注册处理器
	err := RegisterAll(router, handler)
	require.NoError(t, err)

	// 验证没有路由被注册
	routes := router.Routes()
	assert.Equal(t, 0, len(routes))
}

func TestRegisterAll_LargeNumberOfRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建有大量路由的处理器
	routes := make([]RouteDefinition, 50)
	for i := 0; i < 50; i++ {
		routes[i] = RouteDefinition{
			Method: http.MethodGet,
			Path:   "/route/" + string(rune('a'+i)),
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"index": i})
			},
		}
	}

	handler := NewMockHandler("large-handler", routes)

	// 注册处理器
	err := RegisterAll(router, handler)
	require.NoError(t, err)

	// 验证所有路由都被注册
	registeredRoutes := router.Routes()
	assert.Equal(t, 50, len(registeredRoutes))
}

func TestRegisterAll_WithWildcardPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// 创建带通配符路径的处理器
	handler := NewMockHandler("wildcard-handler", []RouteDefinition{
		{
			Method: http.MethodGet,
			Path:   "/files/*filepath",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"filepath": c.Param("filepath")})
			},
		},
	})

	// 注册处理器
	err := RegisterAll(router, handler)
	require.NoError(t, err)

	// 测试通配符路由
	req := httptest.NewRequest(http.MethodGet, "/files/path/to/file.txt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}