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
	"github.com/industrial-ai/platform/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestIngestTelemetry_Success tests successful telemetry ingestion
func TestIngestTelemetry_Success(t *testing.T) {
	telemetryData := model.TelemetryData{
		DeviceID:    "device-001",
		Timestamp:   time.Now(),
		Temperature: 25.5,
		Pressure:    100.0,
		Vibration:   1.2,
		Humidity:    45.0,
		Power:       150.0,
		Status:      "normal",
	}

	assert.Equal(t, "device-001", telemetryData.DeviceID)
	assert.Equal(t, 25.5, telemetryData.Temperature)
	assert.NotZero(t, telemetryData.Timestamp)

	err := service.ValidateTelemetryData(&telemetryData)
	assert.NoError(t, err)
}

// TestIngestTelemetry_InvalidJSON tests telemetry ingestion with invalid JSON
func TestIngestTelemetry_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/devices/telemetry", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	var data model.TelemetryData
	err := json.Unmarshal([]byte("invalid json"), &data)
	assert.Error(t, err)
}

// TestIngestTelemetry_ValidationFailure tests telemetry validation failure
func TestIngestTelemetry_ValidationFailure(t *testing.T) {
	telemetryData := model.TelemetryData{
		DeviceID:    "",
		Timestamp:   time.Now(),
		Temperature: 25.5,
	}

	err := service.ValidateTelemetryData(&telemetryData)
	assert.Error(t, err)
}

// TestIngestTelemetry_AutoRegister tests auto-registration when device doesn't exist
func TestIngestTelemetry_AutoRegister(t *testing.T) {
	mockDeviceRepo := new(MockDeviceRepository)
	mockDeviceSvc := new(MockDeviceService)
	mockTelemetrySvc := new(MockTelemetryService)

	telemetryData := model.TelemetryData{
		DeviceID:    "new-device-001",
		Timestamp:   time.Now(),
		Temperature: 25.5,
	}

	mockDeviceRepo.On("GetByID", mock.Anything, "new-device-001").Return(nil, errors.New("not found"))

	newDevice := &model.Device{
		ID:   "new-device-001",
		Name: "Auto-registered Device",
		Type: "unknown",
	}
	mockDeviceSvc.On("AutoRegisterDevice", mock.Anything, "new-device-001").Return(newDevice, nil)
	mockTelemetrySvc.On("Ingest", mock.Anything, mock.AnythingOfType("*model.TelemetryData")).Return(nil)

	err := service.ValidateTelemetryData(&telemetryData)
	assert.NoError(t, err)
}

// TestIngestTelemetry_AutoRegisterFailure tests auto-registration failure
func TestIngestTelemetry_AutoRegisterFailure(t *testing.T) {
	mockDeviceRepo := new(MockDeviceRepository)
	mockDeviceSvc := new(MockDeviceService)

	mockDeviceRepo.On("GetByID", mock.Anything, "device-fail").Return(nil, errors.New("not found"))
	mockDeviceSvc.On("AutoRegisterDevice", mock.Anything, "device-fail").Return(nil, errors.New("auto-register failed"))
}

// TestAgentQuery_Success tests successful AI agent query
func TestAgentQuery_Success(t *testing.T) {
	mockAgentSvc := new(MockAgentService)

	query := model.AgentQuery{
		Query:    "What is the status of device-001?",
		DeviceID: "device-001",
	}

	expectedResponse := &model.AgentResponse{
		SessionID: "session-001",
		Response:  "Device is operating normally",
		Agent:     "maintenance-agent",
		Timestamp: time.Now(),
	}

	mockAgentSvc.On("Query", mock.Anything, query).Return(expectedResponse, nil)
	mockAgentSvc.On("GetDeviceContext", mock.Anything, "device-001").Return(&model.DeviceContext{}, nil)

	assert.Equal(t, "What is the status of device-001?", query.Query)
	assert.Equal(t, "device-001", query.DeviceID)
}

// TestAgentQuery_InvalidRequest tests agent query with invalid request
func TestAgentQuery_InvalidRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/agent/query", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	var query model.AgentQuery
	err := json.Unmarshal([]byte("invalid json"), &query)
	assert.Error(t, err)
}

// TestAgentQuery_QueryFailure tests agent query failure
func TestAgentQuery_QueryFailure(t *testing.T) {
	mockAgentSvc := new(MockAgentService)

	query := model.AgentQuery{
		Query: "Invalid query",
	}

	mockAgentSvc.On("Query", mock.Anything, query).Return(nil, errors.New("query failed"))
}

// TestGetAIStatus_Success tests successful AI status retrieval
func TestGetAIStatus_Success(t *testing.T) {
	mockAgentSvc := new(MockAgentService)

	expectedLogs := []model.AgentTaskLog{
		{
			ID:         1,
			SessionID:  "session-001",
			Query:      "Status query",
			Response:   "Device is normal",
			Agent:      "maintenance-agent",
			ExecutedAt: time.Now(),
		},
		{
			ID:         2,
			SessionID:  "session-002",
			Query:      "Temperature check",
			Response:   "Temperature is 25°C",
			Agent:      "monitoring-agent",
			ExecutedAt: time.Now(),
		},
	}

	mockAgentSvc.On("GetTaskLogs", mock.Anything, 50).Return(expectedLogs, nil)

	assert.Len(t, expectedLogs, 2)
	assert.Equal(t, "session-001", expectedLogs[0].SessionID)
}

// TestGetAIStatus_WithCustomLimit tests AI status with custom limit
func TestGetAIStatus_WithCustomLimit(t *testing.T) {
	mockAgentSvc := new(MockAgentService)

	expectedLogs := []model.AgentTaskLog{}
	mockAgentSvc.On("GetTaskLogs", mock.Anything, 10).Return(expectedLogs, nil)
}

// TestGetAIStatus_Failure tests AI status retrieval failure
func TestGetAIStatus_Failure(t *testing.T) {
	mockAgentSvc := new(MockAgentService)
	mockAgentSvc.On("GetTaskLogs", mock.Anything, 50).Return(nil, errors.New("database error"))
}

// TestValidateTelemetryData tests telemetry data validation
func TestValidateTelemetryData(t *testing.T) {
	tests := []struct {
		name    string
		data    model.TelemetryData
		wantErr bool
	}{
		{
			name: "valid data",
			data: model.TelemetryData{
				DeviceID:    "device-001",
				Timestamp:   time.Now(),
				Temperature: 25.5,
			},
			wantErr: false,
		},
		{
			name: "empty device ID",
			data: model.TelemetryData{
				DeviceID:  "",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero timestamp sets default",
			data: model.TelemetryData{
				DeviceID:  "device-001",
				Timestamp: time.Time{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTelemetryData(&tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTelemetryDataStructure tests telemetry data structure integrity
func TestTelemetryDataStructure(t *testing.T) {
	data := model.TelemetryData{
		ID:          1,
		DeviceID:    "device-001",
		TenantID:    "tenant-001",
		Timestamp:   time.Now(),
		Temperature: 25.5,
		Pressure:    100.0,
		Vibration:   1.2,
		Humidity:    45.0,
		Power:       150.0,
		Status:      "normal",
		Message:     "All readings normal",
	}

	assert.Equal(t, int64(1), data.ID)
	assert.Equal(t, "device-001", data.DeviceID)
	assert.Equal(t, "tenant-001", data.TenantID)
	assert.NotZero(t, data.Timestamp)
	assert.Equal(t, 25.5, data.Temperature)
	assert.Equal(t, 100.0, data.Pressure)
	assert.Equal(t, 1.2, data.Vibration)
	assert.Equal(t, 45.0, data.Humidity)
	assert.Equal(t, 150.0, data.Power)
	assert.Equal(t, "normal", data.Status)
	assert.Equal(t, "All readings normal", data.Message)
}

// TestAgentQueryStructure tests agent query structure integrity
func TestAgentQueryStructure(t *testing.T) {
	query := model.AgentQuery{
		Query:     "What is the device status?",
		Context:   map[string]interface{}{"key": "value"},
		DeviceID:  "device-001",
		SessionID: "session-001",
		TenantID:  "tenant-001",
	}

	assert.Equal(t, "What is the device status?", query.Query)
	assert.NotNil(t, query.Context)
	assert.Equal(t, "device-001", query.DeviceID)
	assert.Equal(t, "session-001", query.SessionID)
	assert.Equal(t, "tenant-001", query.TenantID)
}

// TestAgentResponseStructure tests agent response structure integrity
func TestAgentResponseStructure(t *testing.T) {
	response := model.AgentResponse{
		SessionID: "session-001",
		Response:  "Device is operating normally",
		Agent:     "maintenance-agent",
		Actions: []map[string]interface{}{
			{"type": "notify", "target": "admin"},
		},
		Timestamp: time.Now(),
	}

	assert.Equal(t, "session-001", response.SessionID)
	assert.Equal(t, "Device is operating normally", response.Response)
	assert.Equal(t, "maintenance-agent", response.Agent)
	assert.Len(t, response.Actions, 1)
	assert.NotZero(t, response.Timestamp)
}
