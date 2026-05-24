package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/mocks"
	"github.com/industrial-ai/platform/internal/service"
)

// ============================================
// 测试 Helper (简化版)
// ============================================

// TestHelper 测试辅助工具
type TestHelper struct {
	t      *testing.T
	router *gin.Engine
}

// NewTestHelper 创建测试辅助工具
func NewTestHelper(t *testing.T) *TestHelper {
	gin.SetMode(gin.TestMode)
	return &TestHelper{
		t:      t,
		router: gin.New(),
	}
}

// GetRouter 获取路由器
func (h *TestHelper) GetRouter() *gin.Engine {
	return h.router
}

// MakeRequest 发起 HTTP 请求
func (h *TestHelper) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()

	h.router.ServeHTTP(w, req)
	return w
}

// AssertSuccess 验证成功响应
func (h *TestHelper) AssertSuccess(w *httptest.ResponseRecorder) {
	require.Equal(h.t, http.StatusOK, w.Code)
}

// AssertBadRequest 验证参数错误响应
func (h *TestHelper) AssertBadRequest(w *httptest.ResponseRecorder) {
	require.Equal(h.t, http.StatusBadRequest, w.Code)
}

// AssertNotFound 验证资源不存在响应
func (h *TestHelper) AssertNotFound(w *httptest.ResponseRecorder) {
	require.Equal(h.t, http.StatusNotFound, w.Code)
}

// AssertInternalServerError 验证服务器错误响应
func (h *TestHelper) AssertInternalServerError(w *httptest.ResponseRecorder) {
	require.Equal(h.t, http.StatusInternalServerError, w.Code)
}

// ParseResponse 解析 JSON 响应
func (h *TestHelper) ParseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return response
}

// ============================================
// Mock Factory (快速创建 Mock Service)
// ============================================

// CreateMockServiceFactory 创建 Mock Service 工厂
func CreateMockServiceFactory() *service.ServiceFactory {
	sf := service.NewServiceFactory()

	sf.SetDeviceService(new(mocks.MockDeviceService))
	sf.SetAlertService(new(mocks.MockAlertService))
	sf.SetAuthService(new(mocks.MockAuthService))
	sf.SetUserService(new(mocks.MockUserService))
	sf.SetHealthService(new(mocks.MockHealthService))

	return sf
}

// CreateMockHandlerFactory 创建 Mock Handler 工厂
func CreateMockHandlerFactory() *HandlerFactory {
	sf := CreateMockServiceFactory()
	return NewHandlerFactory(sf, func(msg model.WSMessage) {}, new(MockCache))
}

// ============================================
// 测试数据生成器
// ============================================

// CreateTestDevice 创建测试设备
func CreateTestDevice(id, name string) *model.Device {
	return &model.Device{
		ID:     id,
		Name:   name,
		Type:   "sensor",
		Status: "online",
	}
}

// CreateTestDevices 创建多个测试设备
func CreateTestDevices(count int) []model.Device {
	devices := []model.Device{}
	for i := 0; i < count; i++ {
		devices = append(devices, model.Device{
			ID:     "device-" + string(rune('0'+i)),
			Name:   "Device " + string(rune('0'+i)),
			Type:   "sensor",
			Status: "online",
		})
	}
	return devices
}

// CreateTestTelemetryData 创建测试遥测数据
func CreateTestTelemetryData(deviceID string) *model.TelemetryData {
	return &model.TelemetryData{
		DeviceID:    deviceID,
		Timestamp:   time.Now(),
		Temperature: 25.5,
		Pressure:    100.0,
		Vibration:   0.1,
	}
}

// CreateTestAlert 创建测试告警
func CreateTestAlert(id int, deviceID, severity string) *model.Alert {
	return &model.Alert{
		ID:       id,
		DeviceID: deviceID,
		Severity: severity,
		Status:   "active",
		Message:  "Test alert message",
	}
}

// CreateTestUser 创建测试用户
func CreateTestUser(id int, username string) *model.User {
	return &model.User{
		ID:       id,
		Username: username,
		Email:    username + "@test.com",
		Role:     "user",
	}
}
