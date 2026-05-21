package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestPaginationParams_Defaults tests the Defaults method
func TestPaginationParams_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		input    PaginationParams
		expected PaginationParams
	}{
		{
			name:     "empty defaults",
			input:    PaginationParams{},
			expected: PaginationParams{Page: 1, PageSize: 20, SortBy: "created_at", Order: "desc"},
		},
		{
			name:     "page zero defaults to 1",
			input:    PaginationParams{Page: 0, PageSize: 10},
			expected: PaginationParams{Page: 1, PageSize: 10, SortBy: "created_at", Order: "desc"},
		},
		{
			name:     "pagesize zero defaults to 20",
			input:    PaginationParams{Page: 2, PageSize: 0},
			expected: PaginationParams{Page: 2, PageSize: 20, SortBy: "created_at", Order: "desc"},
		},
		{
			name:     "pagesize over 100 defaults to 20",
			input:    PaginationParams{Page: 1, PageSize: 200},
			expected: PaginationParams{Page: 1, PageSize: 20, SortBy: "created_at", Order: "desc"},
		},
		{
			name:     "empty sortBy defaults to created_at",
			input:    PaginationParams{Page: 1, PageSize: 10, SortBy: ""},
			expected: PaginationParams{Page: 1, PageSize: 10, SortBy: "created_at", Order: "desc"},
		},
		{
			name:     "empty order defaults to desc",
			input:    PaginationParams{Page: 1, PageSize: 10, SortBy: "name", Order: ""},
			expected: PaginationParams{Page: 1, PageSize: 10, SortBy: "name", Order: "desc"},
		},
		{
			name:     "valid params unchanged",
			input:    PaginationParams{Page: 5, PageSize: 50, SortBy: "name", Order: "asc"},
			expected: PaginationParams{Page: 5, PageSize: 50, SortBy: "name", Order: "asc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.input
			params.Defaults()

			if params.Page != tt.expected.Page {
				t.Errorf("Page = %d, want %d", params.Page, tt.expected.Page)
			}
			if params.PageSize != tt.expected.PageSize {
				t.Errorf("PageSize = %d, want %d", params.PageSize, tt.expected.PageSize)
			}
			if params.SortBy != tt.expected.SortBy {
				t.Errorf("SortBy = %s, want %s", params.SortBy, tt.expected.SortBy)
			}
			if params.Order != tt.expected.Order {
				t.Errorf("Order = %s, want %s", params.Order, tt.expected.Order)
			}
		})
	}
}

// TestPaginationParams_JSON tests JSON serialization
func TestPaginationParams_JSON(t *testing.T) {
	params := PaginationParams{Page: 2, PageSize: 50, SortBy: "name", Order: "asc"}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled PaginationParams
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Page != params.Page {
		t.Errorf("Page = %d, want %d", unmarshaled.Page, params.Page)
	}
}

// TestUser_JSON tests User struct JSON
func TestUser_JSON(t *testing.T) {
	now := time.Now()
	user := User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "admin",
		Roles:        []string{"admin", "operator"},
		TenantID:     "t-1",
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled User
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Password should not be in JSON
	if unmarshaled.Password != "" {
		t.Errorf("Password should be empty in JSON")
	}

	if unmarshaled.ID != user.ID {
		t.Errorf("ID = %d, want %d", unmarshaled.ID, user.ID)
	}
	if unmarshaled.Username != user.Username {
		t.Errorf("Username = %s, want %s", unmarshaled.Username, user.Username)
	}
	if len(unmarshaled.Roles) != len(user.Roles) {
		t.Errorf("Roles length = %d, want %d", len(unmarshaled.Roles), len(user.Roles))
	}
}

// TestDevice_JSON tests Device struct
func TestDevice_JSON(t *testing.T) {
	now := time.Now()
	device := Device{
		ID:          "d-1",
		Name:        "Sensor A",
		Type:        "temperature",
		Location:    "Factory 1",
		Status:      "online",
		Description: "Temperature sensor",
		TenantID:    "t-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Device
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ID != device.ID {
		t.Errorf("ID = %s, want %s", unmarshaled.ID, device.ID)
	}
	if unmarshaled.Status != device.Status {
		t.Errorf("Status = %s, want %s", unmarshaled.Status, device.Status)
	}
}

// TestTelemetryData_JSON tests TelemetryData struct
func TestTelemetryData_JSON(t *testing.T) {
	now := time.Now()
	data := TelemetryData{
		ID:          1,
		DeviceID:    "d-1",
		TenantID:    "t-1",
		Timestamp:   now,
		Temperature: 25.5,
		Pressure:    101.3,
		Vibration:   0.01,
		Humidity:    60.0,
		Power:       100.0,
		Status:      "normal",
		Message:     "OK",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled TelemetryData
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.DeviceID != data.DeviceID {
		t.Errorf("DeviceID = %s, want %s", unmarshaled.DeviceID, data.DeviceID)
	}
	if unmarshaled.Temperature != data.Temperature {
		t.Errorf("Temperature = %f, want %f", unmarshaled.Temperature, data.Temperature)
	}
}

// TestAlertRule_JSON tests AlertRule struct
func TestAlertRule_JSON(t *testing.T) {
	now := time.Now()
	rule := AlertRule{
		ID:          1,
		Name:        "High Temp Alert",
		DeviceType:  "temperature",
		Metric:      "temperature",
		Operator:    ">",
		Threshold:   80.0,
		Severity:    "high",
		Actions:     "email,sms",
		Enabled:     true,
		CooldownSec: 300,
		TenantID:    "t-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled AlertRule
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Name != rule.Name {
		t.Errorf("Name = %s, want %s", unmarshaled.Name, rule.Name)
	}
	if unmarshaled.Threshold != rule.Threshold {
		t.Errorf("Threshold = %f, want %f", unmarshaled.Threshold, rule.Threshold)
	}
}

// TestAlert_JSON tests Alert struct
func TestAlert_JSON(t *testing.T) {
	now := time.Now()
	alert := Alert{
		ID:          1,
		RuleID:      1,
		DeviceID:    "d-1",
		TenantID:    "t-1",
		Message:     "Temperature exceeded threshold",
		Severity:    "high",
		Status:      "active",
		TriggeredAt: now,
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Alert
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Message != alert.Message {
		t.Errorf("Message = %s, want %s", unmarshaled.Message, alert.Message)
	}
}

// TestAlert_ResolvedAt tests Alert with ResolvedAt
func TestAlert_ResolvedAt(t *testing.T) {
	now := time.Now()
	resolved := now.Add(1 * time.Hour)
	alert := Alert{
		ID:          1,
		TriggeredAt: now,
		ResolvedAt:  &resolved,
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Alert
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ResolvedAt == nil {
		t.Error("ResolvedAt should not be nil")
	}
}

// TestWorkOrder_JSON tests WorkOrder struct
func TestWorkOrder_JSON(t *testing.T) {
	now := time.Now()
	wo := WorkOrder{
		ID:          1,
		Title:       "Fix sensor",
		Description: "Replace faulty sensor",
		DeviceID:    "d-1",
		TenantID:    "t-1",
		Priority:    "high",
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled WorkOrder
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Title != wo.Title {
		t.Errorf("Title = %s, want %s", unmarshaled.Title, wo.Title)
	}
}

// TestWorkOrder_AssignedTo tests WorkOrder with AssignedTo
func TestWorkOrder_AssignedTo(t *testing.T) {
	assignedID := 5
	wo := WorkOrder{
		ID:         1,
		AssignedTo: &assignedID,
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled WorkOrder
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.AssignedTo == nil || *unmarshaled.AssignedTo != assignedID {
		t.Errorf("AssignedTo = %v, want %d", unmarshaled.AssignedTo, assignedID)
	}
}

// TestNotification_JSON tests Notification struct
func TestNotification_JSON(t *testing.T) {
	now := time.Now()
	deviceID := "d-1"
	notif := Notification{
		ID:        1,
		Type:      "alert",
		Title:     "High Temperature",
		Message:   "Temperature exceeded 80C",
		DeviceID:  &deviceID,
		TenantID:  "t-1",
		Read:      false,
		CreatedAt: now,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Notification
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Title != notif.Title {
		t.Errorf("Title = %s, want %s", unmarshaled.Title, notif.Title)
	}
}

// TestBlackBoxRecord_JSON tests BlackBoxRecord struct
func TestBlackBoxRecord_JSON(t *testing.T) {
	now := time.Now()
	record := BlackBoxRecord{
		ID:          1,
		DeviceID:    "d-1",
		TenantID:    "t-1",
		TriggerType: "alert",
		StartTime:   now,
		EndTime:     now.Add(5 * time.Minute),
		Snapshot:    []TelemetryData{},
		Summary:     "Anomaly detected",
		CreatedAt:   now,
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled BlackBoxRecord
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.TriggerType != record.TriggerType {
		t.Errorf("TriggerType = %s, want %s", unmarshaled.TriggerType, record.TriggerType)
	}
}

// TestReport_JSON tests Report struct
func TestReport_JSON(t *testing.T) {
	now := time.Now()
	deviceID := "d-1"
	report := Report{
		ID:          1,
		Title:       "Weekly Summary",
		Type:        "weekly",
		DeviceID:    &deviceID,
		TenantID:    "t-1",
		Content:     "Report content...",
		GeneratedAt: now,
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled Report
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Title != report.Title {
		t.Errorf("Title = %s, want %s", unmarshaled.Title, report.Title)
	}
}

// TestAgentQuery_JSON tests AgentQuery struct
func TestAgentQuery_JSON(t *testing.T) {
	query := AgentQuery{
		Query:     "What is the temperature?",
		Context:   map[string]interface{}{"location": "Factory 1"},
		DeviceID:  "d-1",
		SessionID: "s-1",
		TenantID:  "t-1",
	}

	data, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled AgentQuery
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Query != query.Query {
		t.Errorf("Query = %s, want %s", unmarshaled.Query, query.Query)
	}
}

// TestAgentResponse_JSON tests AgentResponse struct
func TestAgentResponse_JSON(t *testing.T) {
	now := time.Now()
	resp := AgentResponse{
		SessionID: "s-1",
		Response:  "Temperature is 25C",
		Agent:     "assistant",
		Actions:   []map[string]interface{}{{"type": "alert"}},
		Timestamp: now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled AgentResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.SessionID != resp.SessionID {
		t.Errorf("SessionID = %s, want %s", unmarshaled.SessionID, resp.SessionID)
	}
}

// TestWSMessage_JSON tests WebSocket message
func TestWSMessage_JSON(t *testing.T) {
	now := time.Now()
	msg := WSMessage{
		Type:      "telemetry",
		Payload:   map[string]interface{}{"temperature": 25.0},
		Timestamp: now,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled WSMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Type != msg.Type {
		t.Errorf("Type = %s, want %s", unmarshaled.Type, msg.Type)
	}
}

// TestDeviceStats_JSON tests DeviceStats struct
func TestDeviceStats_JSON(t *testing.T) {
	stats := DeviceStats{
		DeviceID:       "d-1",
		AvgTemperature: 25.0,
		AvgPressure:    101.0,
		AvgVibration:   0.01,
		MaxTemperature: 50.0,
		MaxPressure:    150.0,
		MaxVibration:   0.5,
		DataPoints:     1000,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled DeviceStats
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.DeviceID != stats.DeviceID {
		t.Errorf("DeviceID = %s, want %s", unmarshaled.DeviceID, stats.DeviceID)
	}
}

// TestROIStats_JSON tests ROIStats struct
func TestROIStats_JSON(t *testing.T) {
	stats := ROIStats{
		TotalDevices:     10,
		ActiveAlerts:     5,
		OpenWorkOrders:   3,
		ResolvedIssues:   20,
		PredictedSavings: 5000.0,
		UptimePercentage: 99.5,
		AvgResponseTime:  2.5,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled ROIStats
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.TotalDevices != stats.TotalDevices {
		t.Errorf("TotalDevices = %d, want %d", unmarshaled.TotalDevices, stats.TotalDevices)
	}
}

// TestSystemStatus_JSON tests SystemStatus struct
func TestSystemStatus_JSON(t *testing.T) {
	now := time.Now()
	status := SystemStatus{
		Database:    "healthy",
		DBLatency:   10,
		Uptime:      "5d 3h",
		Version:     "1.0.0",
		Timestamp:   now,
		DeviceCount: 100,
		UserCount:   50,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var unmarshaled SystemStatus
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Database != status.Database {
		t.Errorf("Database = %s, want %s", unmarshaled.Database, status.Database)
	}
}
