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

	// 默认告警冷却时间（秒）
	DefaultAlertCooldownSec = 300

	// 短告警冷却时间（秒）
	ShortAlertCooldownSec = 180

	// 长告警冷却时间（秒）
	LongAlertCooldownSec = 600

	// WebSocket 广播通道容量
	WSBroadcastChannelSize = 100
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
// 告警阈值常量
// ============================================

const (
	// 高温告警阈值
	HighTemperatureThreshold = 100

	// 严重高温告警阈值
	CriticalTemperatureThreshold = 120

	// 振动异常阈值
	AbnormalVibrationThreshold = 3.0

	// 严重振动阈值
	CriticalVibrationThreshold = 5.0

	// 压力异常阈值
	AbnormalPressureThreshold = 150
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
// 密码常量
// ============================================

const (
	// 密码最小长度
	PasswordMinLength = 8

	// 密码最大长度
	PasswordMaxLength = 128
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