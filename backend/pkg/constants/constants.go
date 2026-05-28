// Package constants provides application-wide constants
// BE-P2-02: 魔法数字定义为常量
package constants

// ============================================
// 分页和限制常量
// ============================================

const (
	// 默认分页大小
	DefaultPageSize = 20

	// 最大分页大小
	MaxPageSize = 100

	// 默认数据查询限制
	DefaultLimit = 1000

	// 最大数据查询限制
	MaxLimit = 10000

	// 默认遥测数据限制
	DefaultTelemetryLimit = 100

	// 最大遥测数据限制
	MaxTelemetryLimit = 1000
)

// ============================================
// 时间和超时常量
// ============================================

const (
	// 默认请求超时时间（秒）
	DefaultRequestTimeoutSec = 30

	// 默认租户请求超时时间（秒）
	DefaultTenantTimeoutSec = 30

	// FIX-019: 默认服务层 Context 超时时间（秒）
	// 用于 Service 层方法，防止数据库操作无限等待
	DefaultServiceTimeoutSec = 30

	// 默认告警冷却时间（秒）
	DefaultAlertCooldownSec = 300

	// 短告警冷却时间（秒）
	ShortAlertCooldownSec = 180

	// 长告警冷却时间（秒）
	LongAlertCooldownSec = 600

	// WebSocket 广播通道容量
	WSBroadcastChannelSize = 100

	// 异步告警评估超时时间（秒）
	// 用于异步 goroutine 中的告警规则评估，防止无限等待
	AlertEvaluationTimeoutSec = 30
)

// ============================================
// 验证常量
// ============================================

const (
	// ID 最小长度
	MinIDLength = 1

	// ID 最大长度
	MaxIDLength = 100

	// 设备名称最大长度
	MaxDeviceNameLength = 200

	// 设备描述最大长度
	MaxDeviceDescriptionLength = 1000

	// 租户名称最小长度
	MinTenantNameLength = 2

	// 租户名称最大长度
	MaxTenantNameLength = 100

	// 租户 Slug 最大长度
	MaxTenantSlugLength = 50

	// 用户名最小长度
	MinUsernameLength = 3

	// 用户名最大长度
	MaxUsernameLength = 50

	// 密码最小长度
	MinPasswordLength = 12

	// 密码最大长度
	MaxPasswordLength = 100
)

// ============================================
// 告警阈值常量 - P2-002: 监控告警调优
// ============================================

const (
	// 温度告警阈值（单位：摄氏度）
	// 优化建议：根据工业设备特性分层设置
	HighTemperatureThreshold     = 80  // 高温警告阈值（从100调整为80，提前预警）
	CriticalTemperatureThreshold = 100 // 严重高温阈值（从120调整为100）
	WarningTemperatureThreshold  = 70  // 温度预警阈值（新增）
	NormalTemperatureThreshold   = 60  // 正常温度上限（新增）

	// 振动告警阈值（单位：mm/s）
	// 优化建议：符合ISO 10816标准
	AbnormalVibrationThreshold = 2.8 // 振动异常阈值（从3.0调整为2.8）
	CriticalVibrationThreshold = 4.5 // 严重振动阈值（从5.0调整为4.5）
	WarningVibrationThreshold  = 1.8 // 振动预警阈值（新增）
	NormalVibrationThreshold   = 1.0 // 正常振动上限（新增）

	// 压力告警阈值（单位：kPa）
	// 优化建议：根据设备压力等级分层
	AbnormalPressureThreshold = 120 // 压力异常阈值（从150调整为120）
	CriticalPressureThreshold = 150 // 严重压力阈值（新增）
	WarningPressureThreshold  = 100 // 压力预警阈值（新增）
	NormalPressureThreshold   = 80  // 正常压力上限（新增）

	// 湿度告警阈值（单位：%）
	HighHumidityThreshold     = 85 // 高湿度警告阈值（新增）
	CriticalHumidityThreshold = 95 // 严重湿度阈值（新增）
	LowHumidityThreshold      = 30 // 低湿度警告阈值（新增）

	// 功率告警阈值（单位：W）
	HighPowerThreshold     = 5000 // 高功率警告阈值（新增）
	CriticalPowerThreshold = 8000 // 严重功率阈值（新增）
	LowPowerThreshold      = 100  // 低功率警告阈值（新增）
)

// ============================================
// 租户计划限制常量
// ============================================

const (
	// Free 计划设备限制
	FreePlanMaxDevices = 10

	// Free 计划用户限制
	FreePlanMaxUsers = 3

	// Free 计划告警限制
	FreePlanMaxAlerts = 20

	// Pro 计划设备限制
	ProPlanMaxDevices = 100

	// Pro 计划用户限制
	ProPlanMaxUsers = 20

	// Pro 计划告警限制
	ProPlanMaxAlerts = 200
)

// ============================================
// 黑匣子快照常量
// ============================================

const (
	// 黑匣子快照时间范围（分钟）
	BlackBoxSnapshotMinutes = 5

	// 黑匣子快照数据点限制
	BlackBoxSnapshotLimit = 100
)

// ============================================
// ROI 计算常量 - P2-05: 魔法数字提取
// ============================================

const (
	// 设备月度基础节省金额（美元）
	// 每台设备每月监控带来的基础价值
	ROIBaseSavingsPerDeviceMonthly = 1000.0

	// 解决问题带来的节省金额（美元）
	// 预防性维护带来的价值
	ROIResolvedIssueSavings = 500.0

	// 活跃告警成本（美元）
	// 运营中断成本
	ROIActiveAlertCost = 100.0

	// 默认平均响应时间（小时）
	// 无历史数据时的默认响应时间
	ROIDefaultAvgResponseTimeHours = 2.5

	// 响应时间计算基础值（小时）
	// 有已解决告警时的基础响应时间
	ROIBaseResponseTimeHours = 1.5

	// 响应时间计算乘数
	// 活跃/已解决告警比率的影响系数
	ROIResponseTimeMultiplier = 2.0
)

// ============================================
// Token 和认证常量
// ============================================

const (
	// AccessToken 有效期（小时）
	AccessTokenDurationHours = 1

	// RefreshToken 有效期（小时）
	RefreshTokenDurationHours = 168 // 7 天

	// Token 密钥最小长度
	TokenMinSecretLength = 32

	// 分页最小大小
	MinPageSize = 1
)

// ============================================
// 有效值验证常量
// ============================================

// ValidTenantPlans 有效租户计划列表
var ValidTenantPlans = []string{"free", "basic", "pro", "enterprise"}

// ValidTenantStatuses 有效租户状态列表
var ValidTenantStatuses = []string{"active", "suspended", "deleted"}

// ValidDeviceStatuses 有效设备状态列表
var ValidDeviceStatuses = []string{"online", "offline", "maintenance", "error"}

// ValidUserRoles 有效用户角色列表
var ValidUserRoles = []string{"admin", "operator", "viewer"}

// ============================================
// 验证函数
// ============================================

// IsValidTenantPlan 验证租户计划是否有效
func IsValidTenantPlan(plan string) bool {
	for _, p := range ValidTenantPlans {
		if p == plan {
			return true
		}
	}
	return false
}

// IsValidTenantStatus 验证租户状态是否有效
func IsValidTenantStatus(status string) bool {
	for _, s := range ValidTenantStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidDeviceStatus 验证设备状态是否有效
func IsValidDeviceStatus(status string) bool {
	for _, s := range ValidDeviceStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidUserRole 验证用户角色是否有效
func IsValidUserRole(role string) bool {
	for _, r := range ValidUserRoles {
		if r == role {
			return true
		}
	}
	return false
}

// GetRoleLevel 获取角色权限级别
// admin=3, operator=2, viewer=1, unknown=0
func GetRoleLevel(role string) int {
	switch role {
	case "admin":
		return 3
	case "operator":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

// HasPermission 检查角色是否有足够权限
func HasPermission(role, required string) bool {
	return GetRoleLevel(role) >= GetRoleLevel(required)
}

// ============================================
// 常量命名规范 (P3-04)
// ============================================
//
// 命名约定:
// 1. 使用 Min/Max 前缀表示范围限制 (如 MinPasswordLength, MaxPasswordLength)
// 2. 避免使用重复或相似的命名，保持一致性
// 3. 避免在其他包重复定义相同语义的常量
//
// 已解决的重复定义:
// - MinPasswordLength (原在 auth_models.go) -> 统一使用 pkg/constants.MinPasswordLength
// - PasswordMinLength (已删除) -> 使用 MinPasswordLength
// - PasswordMaxLength (已删除) -> 使用 MaxPasswordLength
//
// 密码常量规范:
// - MinPasswordLength = 12 (NIST 推荐的最低密码长度)
// - MaxPasswordLength = 100 (数据库字段限制)
