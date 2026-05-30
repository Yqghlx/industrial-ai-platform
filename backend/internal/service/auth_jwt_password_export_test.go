package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/model"
)

// ============================================
// JWT Service 额外测试（补充 auth_helpers_test.go 中未覆盖的场景）
// ============================================

const testJWTSecretForExtra = "test-jwt-secret-for-unit-tests-min-32-chars"

func newTestJWTServiceExtra(t *testing.T) *JWTService {
	svc, err := NewJWTService(testJWTSecretForExtra)
	require.NoError(t, err)
	return svc
}

func TestNewJWTService_EmptySecret(t *testing.T) {
	svc, err := NewJWTService("")
	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "required")
}

func TestNewJWTService_ShortSecret(t *testing.T) {
	svc, err := NewJWTService("short")
	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "at least")
}

func TestJWTService_ParseToken_ExpiredToken(t *testing.T) {
	// 手动创建一个过期的 token
	claims := Claims{
		UserID:       1,
		Username:     "testuser",
		Role:         "admin",
		TenantID:     "tenant-1",
		TokenType:    "access",
		TokenID:      "expired-token-id",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    TokenIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testJWTSecretForExtra))
	require.NoError(t, err)

	svc := newTestJWTServiceExtra(t)
	parsedClaims, err := svc.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

func TestJWTService_ParseToken_InvalidSigningMethod(t *testing.T) {
	claims := Claims{
		UserID:       1,
		Username:     "testuser",
		Role:         "admin",
		TokenType:    "access",
		TokenID:      "test-id",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    TokenIssuer,
		},
	}

	// 使用 RS256 签名方法（不是 HMAC）
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString([]byte("any-key"))

	svc := newTestJWTServiceExtra(t)
	parsedClaims, err := svc.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}

func TestJWTService_ParseToken_WrongSecret(t *testing.T) {
	svc := newTestJWTServiceExtra(t)

	token, _, err := svc.GenerateAccessToken(1, "testuser", "admin", "tenant-1", 1)
	require.NoError(t, err)

	wrongSvc, err := NewJWTService("different-secret-key-with-32-characters!!")
	require.NoError(t, err)

	claims, err := wrongSvc.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_SetUserTokenStore(t *testing.T) {
	svc := newTestJWTServiceExtra(t)
	// 不应该 panic
	svc.SetUserTokenStore(nil)
}

func TestGenerateTokenID_Uniqueness(t *testing.T) {
	id1 := generateTokenID()
	id2 := generateTokenID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

// ============================================
// 密码哈希和验证额外测试
// ============================================

func TestHashPassword_DifferentPasswords(t *testing.T) {
	hash1, _ := HashPassword("password1")
	hash2, _ := HashPassword("password2")
	assert.NotEqual(t, hash1, hash2, "不同密码应产生不同哈希")
}

func TestHashPassword_SamePasswordDifferentHash(t *testing.T) {
	hash1, _ := HashPassword("same-password")
	hash2, _ := HashPassword("same-password")
	// bcrypt 使用随机 salt，同一密码产生不同哈希
	assert.NotEqual(t, hash1, hash2, "bcrypt 应使用随机 salt")
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	require.NoError(t, err)
	assert.True(t, VerifyPassword("", hash))
	assert.False(t, VerifyPassword("non-empty", hash))
}

func TestBcryptCostValue(t *testing.T) {
	assert.Equal(t, 12, BcryptCost, "bcrypt 成本因子应为 12")
}

// ============================================
// Export PDF 额外测试
// ============================================

func newMinimalExportService() *ExportService {
	return &ExportService{}
}

func TestExportPDF_DevicesWithStats(t *testing.T) {
	svc := newMinimalExportService()

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary: DeviceSummary{
			TotalDevices:   10,
			OnlineDevices:  8,
			OfflineDevices: 1,
			WarningDevices: 1,
			FaultDevices:   0,
			AvgTemperature: 45.5,
			AvgVibration:   1.2,
		},
		Devices: []model.Device{
			{ID: "dev-001", Name: "CNC Machine", Type: "cnc", Location: "Workshop A", Status: "online"},
		},
		DeviceStats: []model.DeviceStats{
			{DeviceID: "dev-001", AvgTemperature: 45.0, AvgVibration: 1.1},
		},
	}

	result, err := svc.exportPDF(data, "devices", "test-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Filename, ".pdf")
	assert.Equal(t, "application/pdf", result.MimeType)

	content := string(result.Data)
	assert.Contains(t, content, "设备状态报告")
	assert.Contains(t, content, "设备运行数据")
	assert.Contains(t, content, "dev-001")
}

func TestExportPDF_AlertsWithRules(t *testing.T) {
	svc := newMinimalExportService()

	data := &AlertReportData{
		GeneratedAt: time.Now(),
		AlertStats: AlertStats{
			TotalAlerts:    50,
			ActiveAlerts:   10,
			ResolvedAlerts: 40,
			CriticalAlerts: 5,
			HighAlerts:     10,
			MediumAlerts:   20,
			LowAlerts:      15,
		},
		TopAlertRules: []AlertRuleCount{
			{RuleName: "Temperature Too High", Count: 25},
			{RuleName: "Vibration Warning", Count: 15},
		},
	}

	result, err := svc.exportPDF(data, "alerts", "alert-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.Contains(t, content, "告警统计报告")
	assert.Contains(t, content, "高频告警规则")
	assert.Contains(t, content, "Temperature Too High: 25 次")
}

func TestExportPDF_ROIWithMonthlyTrend(t *testing.T) {
	svc := newMinimalExportService()

	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats: model.ROIStats{
			TotalDevices:     20,
			PredictedSavings: 150000.50,
			UptimePercentage: 98.5,
		},
		MonthlyTrend: []MonthlyMetric{
			{Month: "2024-01", TotalSavings: 50000, UptimePercent: 99.0, AlertCount: 10},
			{Month: "2024-02", TotalSavings: 55000, UptimePercent: 98.5, AlertCount: 8},
		},
	}

	result, err := svc.exportPDF(data, "roi", "roi-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.Contains(t, content, "ROI分析报告")
	assert.Contains(t, content, "月度趋势")
	assert.Contains(t, content, "2024-01")
	assert.Contains(t, content, "2024-02")
}

func TestExportPDF_EmptyData(t *testing.T) {
	svc := newMinimalExportService()

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary:     DeviceSummary{},
		Devices:     []model.Device{},
	}

	result, err := svc.exportPDF(data, "devices", "empty-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.Contains(t, content, "设备状态报告")
	assert.Contains(t, content, "总设备数: 0")
}

func TestExportPDF_AlertNoRules(t *testing.T) {
	svc := newMinimalExportService()

	data := &AlertReportData{
		GeneratedAt: time.Now(),
		AlertStats: AlertStats{
			TotalAlerts: 5,
		},
		TopAlertRules: nil,
	}

	result, err := svc.exportPDF(data, "alerts", "alert-no-rules")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.Contains(t, content, "告警统计报告")
	assert.NotContains(t, content, "高频告警规则")
}

func TestExportPDF_ROINoTrend(t *testing.T) {
	svc := newMinimalExportService()

	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats: model.ROIStats{
			TotalDevices:     10,
			PredictedSavings: 50000,
			UptimePercentage: 95.0,
		},
		MonthlyTrend: nil,
	}

	result, err := svc.exportPDF(data, "roi", "roi-no-trend")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.NotContains(t, content, "月度趋势")
}

// ============================================
// Export XLSX 额外测试
// ============================================

func TestExportXLSX_DevicesReport(t *testing.T) {
	svc := newMinimalExportService()

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary: DeviceSummary{
			TotalDevices:  5,
			OnlineDevices: 4,
		},
		Devices: []model.Device{
			{ID: "dev-001", Name: "Pump A", Type: "pump", Location: "Room 1", Status: "online"},
		},
	}

	result, err := svc.exportXLSX(data, "devices", "test-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Filename, ".xlsx")
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.MimeType)

	content := string(result.Data)
	assert.Contains(t, content, "设备状态报告")
	assert.Contains(t, content, "Pump A")
}

func TestExportXLSX_AlertsReport(t *testing.T) {
	svc := newMinimalExportService()

	data := &AlertReportData{
		GeneratedAt: time.Now(),
		AlertStats: AlertStats{
			TotalAlerts:  30,
			ActiveAlerts: 5,
			CriticalAlerts: 2,
			HighAlerts: 5,
			MediumAlerts: 10,
			LowAlerts: 8,
		},
		Alerts: []model.Alert{
			{ID: 1, DeviceID: "dev-001", Severity: "critical", Status: "active", Message: "Temperature exceeded"},
		},
	}

	result, err := svc.exportXLSX(data, "alerts", "alert-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.Contains(t, content, "告警统计报告")
	assert.Contains(t, content, "总告警数,30")
	assert.Contains(t, content, "严重,2")
}

func TestExportXLSX_ROIReport(t *testing.T) {
	svc := newMinimalExportService()

	data := &ROIReportData{
		GeneratedAt: time.Now(),
		ROIStats: model.ROIStats{
			TotalDevices:     15,
			PredictedSavings: 100000,
			UptimePercentage: 97.5,
		},
		MonthlyTrend: []MonthlyMetric{
			{Month: "2024-01", TotalSavings: 50000, UptimePercent: 98.0},
		},
	}

	result, err := svc.exportXLSX(data, "roi", "roi-report")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	content := string(result.Data)
	assert.Contains(t, content, "ROI分析报告")
	assert.Contains(t, content, "100000")
}

// ============================================
// Export 结果类型验证
// ============================================

func TestExportResult_DataIntegrity(t *testing.T) {
	svc := newMinimalExportService()

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary:     DeviceSummary{TotalDevices: 1},
		Devices:     []model.Device{{ID: "d1", Name: "Device 1", Type: "pump", Location: "A", Status: "online"}},
	}

	result, err := svc.exportPDF(data, "devices", "integrity-test")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(result.Data)), result.Size)
	assert.NotEmpty(t, result.Filename)
}

func TestExportXLSX_ResultDataIntegrity(t *testing.T) {
	svc := newMinimalExportService()

	data := &DeviceReportData{
		GeneratedAt: time.Now(),
		Summary:     DeviceSummary{TotalDevices: 1},
		Devices:     []model.Device{{ID: "d1", Name: "Device 1", Type: "pump", Location: "A", Status: "online"}},
	}

	result, err := svc.exportXLSX(data, "devices", "integrity-test")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(result.Data)), result.Size)
}
