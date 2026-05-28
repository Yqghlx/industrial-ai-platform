package handler

// ============================================
// 测试辅助函数 - 减少重复代码
// ============================================
//
// 使用说明：
// 1. 使用 setupDeviceHandlerTest() 创建测试环境
// 2. 使用 assertHTTPResponse() 统一验证HTTP响应
// 3. 使用 mocks.AssertAllExpectations() 统一验证所有mock
//
// 示例：
// func TestDeviceHandler_ListDevices_Success(t *testing.T) {
//     router, handler, mocks := setupDeviceHandlerTest(t)
//     mocks.deviceSvc.On("List", mock.Anything, 1, 20).Return(testDevices, 2, nil)
//     router.GET("/devices", handler.ListDevices)
//
//     w := performRequest(router, "GET", "/devices?page=1&page_size=20")
//     assertHTTPResponse(t, w, http.StatusOK)
//     assertResponseBodyContains(t, w, "total", float64(2))
//     mocks.AssertAllExpectations(t)
// }

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/industrial-ai/platform/internal/model"
)

// TestMocks 封装所有mock对象，便于统一管理
type TestMocks struct {
	DeviceSvc    *MockDeviceService
	AlertSvc     *MockAlertService
	AuthSvc      *MockAuthService
	TelemetrySvc *MockTelemetryService
	UserSvc      *MockUserService
}

// AssertAllExpectations 验证所有mock期望被满足
func (m *TestMocks) AssertAllExpectations(t *testing.T) {
	if m.DeviceSvc != nil {
		m.DeviceSvc.AssertExpectations(t)
	}
	if m.AlertSvc != nil {
		m.AlertSvc.AssertExpectations(t)
	}
	if m.AuthSvc != nil {
		m.AuthSvc.AssertExpectations(t)
	}
	if m.TelemetrySvc != nil {
		m.TelemetrySvc.AssertExpectations(t)
	}
	if m.UserSvc != nil {
		m.UserSvc.AssertExpectations(t)
	}
}

// TestFixtures 测试数据常量，确保数据一致性
var TestFixtures = struct {
	Devices []model.Device
	Alerts  []model.Alert
	Rules   []model.AlertRule
	Users   []model.User
}{
	Devices: []model.Device{
		{ID: "CNC-001", Name: "数控机床001", Type: "数控机床", Status: "online"},
		{ID: "CNC-002", Name: "数控机床002", Type: "数控机床", Status: "online"},
		{ID: "INJ-001", Name: "注塑机001", Type: "注塑机", Status: "online"},
	},
	Alerts: []model.Alert{
		{ID: 1, DeviceID: "CNC-001", Severity: "high", Status: "active", Message: "温度过高"},
		{ID: 2, DeviceID: "CNC-002", Severity: "medium", Status: "resolved", Message: "振动异常"},
	},
	Rules: []model.AlertRule{
		{ID: 1, Name: "高温告警", Metric: "temperature", Operator: ">", Threshold: 100, Severity: "high"},
		{ID: 2, Name: "振动异常", Metric: "vibration", Operator: ">", Threshold: 3.0, Severity: "medium"},
	},
	Users: []model.User{
		{ID: 1, Username: "admin", Role: "admin", TenantID: "tenant-001"},
		{ID: 2, Username: "operator", Role: "operator", TenantID: "tenant-001"},
	},
}

// setupDeviceHandlerTest 创建DeviceHandler测试环境
// 返回: router, handler, mocks
func setupDeviceHandlerTest(t *testing.T) (*gin.Engine, *DeviceHandlerNew, *TestMocks) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mocks := &TestMocks{
		DeviceSvc:    new(MockDeviceService),
		AlertSvc:     new(MockAlertService),
		AuthSvc:      new(MockAuthService),
		TelemetrySvc: new(MockTelemetryService),
	}

	broadcastFunc := func(msg model.WSMessage) {}
	handler := NewDeviceHandlerNew(
		mocks.DeviceSvc,
		mocks.AlertSvc,
		mocks.AuthSvc,
		mocks.TelemetrySvc,
		broadcastFunc,
	)

	return router, handler, mocks
}

// setupAlertHandlerTest 创建AlertHandler测试环境
func setupAlertHandlerTest(t *testing.T) (*gin.Engine, *AlertHandler, *TestMocks) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mocks := &TestMocks{
		AlertSvc: new(MockAlertService),
	}

	broadcastFunc := func(msg model.WSMessage) {}
	handler := NewAlertHandler(mocks.AlertSvc, broadcastFunc)

	return router, handler, mocks
}

// setupAuthHandlerTest 创建AuthHandler测试环境
func setupAuthHandlerTest(t *testing.T) (*gin.Engine, *AuthHandlerNew, *TestMocks) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mocks := &TestMocks{
		AuthSvc: new(MockAuthService),
		UserSvc: new(MockUserService),
	}

	handler := NewAuthHandlerNew(mocks.AuthSvc, mocks.UserSvc)

	return router, handler, mocks
}

// performRequest 执行HTTP请求并返回响应
func performRequest(router *gin.Engine, method, path string, body ...interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if len(body) > 0 && body[0] != nil {
		jsonBody, _ := json.Marshal(body[0])
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// performRequestWithHeaders 执行带自定义Header的HTTP请求
func performRequestWithHeaders(router *gin.Engine, method, path string, headers map[string]string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// assertHTTPResponse 验证HTTP响应状态码
func assertHTTPResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, w.Code, "HTTP status code mismatch")
}

// assertResponseBodyContains 验证响应体包含指定字段和值
func assertResponseBodyContains(t *testing.T, w *httptest.ResponseRecorder, key string, expectedValue interface{}) {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to parse response body")

	assert.Equal(t, expectedValue, response[key], "Response body key '%s' mismatch", key)
}

// assertResponseBodyContainsMessage 验证响应体包含指定消息
func assertResponseBodyContainsMessage(t *testing.T, w *httptest.ResponseRecorder, expectedMessage string) {
	assert.Contains(t, w.Body.String(), expectedMessage, "Response body should contain message")
}

// assertErrorResponse 验证错误响应格式
func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	assertHTTPResponse(t, w, expectedStatus)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedCode, response["code"], "Error code mismatch")
	assert.NotEmpty(t, response["error"], "Error message should not be empty")
}

// parseResponseBody 解析响应体为指定类型
func parseResponseBody(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	assert.NoError(t, err, "Failed to parse response body")
}

// ============================================
// Table-driven测试辅助函数
// ============================================

// HTTPTestCase HTTP测试用例结构
type HTTPTestCase struct {
	Name           string
	Method         string
	Path           string
	Body           interface{}
	Headers        map[string]string
	ExpectedStatus int
	ExpectedBody   map[string]interface{}
	MockSetup      func(mocks *TestMocks)
}

// RunHTTPTests 执行table-driven HTTP测试
func RunHTTPTests(t *testing.T, testCases []HTTPTestCase, setupFunc func(t *testing.T) (*gin.Engine, interface{}, *TestMocks)) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			router, handler, mocks := setupFunc(t)

			// Setup mocks
			if tc.MockSetup != nil {
				tc.MockSetup(mocks)
			}

			// Register route based on handler type
			registerRoute(router, handler, tc.Method, tc.Path)

			// Execute request
			w := performRequestWithHeaders(router, tc.Method, tc.Path, tc.Headers, tc.Body)

			// Verify response
			assertHTTPResponse(t, w, tc.ExpectedStatus)

			if tc.ExpectedBody != nil {
				for key, value := range tc.ExpectedBody {
					assertResponseBodyContains(t, w, key, value)
				}
			}

			mocks.AssertAllExpectations(t)
		})
	}
}

// registerRoute 根据handler类型注册路由
func registerRoute(router *gin.Engine, handler interface{}, method, path string) {
	switch h := handler.(type) {
	case *DeviceHandlerNew:
		switch method {
		case "GET":
			if path == "/devices" {
				router.GET(path, h.ListDevices)
			} else if path == "/devices/graph" {
				router.GET(path, h.GetDeviceGraph)
			} else {
				router.GET(path, h.GetDevice)
			}
		case "POST":
			router.POST(path, h.CreateDevice)
		case "DELETE":
			router.DELETE(path, h.DeleteDevice)
		}
	case *AlertHandler:
		switch method {
		case "GET":
			router.GET(path, h.ListAlerts)
		case "POST":
			router.POST(path, h.AcknowledgeAlert)
		}
	}
}

// ============================================
// 边界测试辅助数据
// ============================================

// BoundaryTestData 边界测试数据
var BoundaryTestData = struct {
	// 大数据量测试
	LargeDeviceList []model.Device
	// 特殊字符测试
	SpecialCharNames []string
	// 边界值测试
	TemperatureThresholds []float64
}{
	LargeDeviceList: generateLargeDeviceList(100),
	SpecialCharNames: []string{
		"设备'测试\"特殊<字符>",
		"设备\n换行符",
		"设备\t制表符",
		"设备 中文 名字",
		"Device-With-Hyphens-123",
	},
	TemperatureThresholds: []float64{
		-999.99, // 极低值
		0.0,     // 零值
		100.0,   // 正常阈值
		999.99,  // 极高值
	},
}

// generateLargeDeviceList 生成大量测试设备
func generateLargeDeviceList(count int) []model.Device {
	devices := make([]model.Device, count)
	for i := 0; i < count; i++ {
		devices[i] = model.Device{
			ID:     fmt.Sprintf("BATCH-%04d", i),
			Name:   fmt.Sprintf("批量设备%d", i),
			Type:   "数控机床",
			Status: "online",
		}
	}
	return devices
}

// ============================================
// Mock匹配器辅助函数
// ============================================

// MatchDeviceByID 创建设备ID匹配器
func MatchDeviceByID(deviceID string) interface{} {
	return mock.MatchedBy(func(d *model.Device) bool {
		return d.ID == deviceID
	})
}

// MatchAlertBySeverity 创建告警严重级别匹配器
func MatchAlertBySeverity(severity string) interface{} {
	return mock.MatchedBy(func(a *model.Alert) bool {
		return a.Severity == severity
	})
}

// MatchRuleByMetric 创建规则指标匹配器
func MatchRuleByMetric(metric string) interface{} {
	return mock.MatchedBy(func(r *model.AlertRule) bool {
		return r.Metric == metric
	})
}

// ============================================
// 测试断言增强函数
// ============================================

// assertPaginationResponse 验证分页响应格式
func assertPaginationResponse(t *testing.T, w *httptest.ResponseRecorder, expectedTotal, expectedPage int) {
	var response map[string]interface{}
	parseResponseBody(t, w, &response)

	assert.Equal(t, float64(expectedTotal), response["total"])
	assert.Equal(t, float64(expectedPage), response["page"])
	assert.NotNil(t, response["data"])
}

// assertTokenValid 验证JWT Token有效性
func assertTokenValid(t *testing.T, token string) {
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.GreaterOrEqual(t, len(token), 20, "Token should be sufficiently long")
}

// assertDeviceFields 验证设备字段完整性
func assertDeviceFields(t *testing.T, device *model.Device) {
	assert.NotEmpty(t, device.ID, "Device ID should not be empty")
	assert.NotEmpty(t, device.Name, "Device name should not be empty")
	assert.NotEmpty(t, device.Type, "Device type should not be empty")
	assert.NotEmpty(t, device.Status, "Device status should not be empty")
	assert.NotZero(t, device.CreatedAt, "Device createdAt should be set")
	assert.NotZero(t, device.UpdatedAt, "Device updatedAt should be set")
}