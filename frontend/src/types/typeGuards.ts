/**
 * Type Guard Functions
 * 用于安全地验证和转换类型，替代 as Type 类型断言
 */

import {
  Device,
  Telemetry,
  DeviceStats,
  Alert,
  AlertRule,
  WorkOrder,
  Notification,
  Report,
  User,
  ROIStats,
  SystemStatus,
  DeviceGraph,
  AgentResponse,
  DeviceStatus,
  AlertSeverity,
  UserRole,
} from './api';

// ============== 基础类型守卫 ==============

/**
 * 检查对象是否具有指定的属性
 */
export function hasProperty<K extends string>(obj: unknown, key: K): obj is Record<K, unknown> {
  return typeof obj === 'object' && obj !== null && key in obj;
}

/**
 * 检查对象是否具有所有指定的属性
 */
function hasProperties(obj: unknown, keys: string[]): boolean {
  if (typeof obj !== 'object' || obj === null) return false;
  const record = obj as Record<string, unknown>;
  return keys.every(key => key in record);
}

// ============== Device 类型守卫 ==============

const DEVICE_STATUS_VALUES: DeviceStatus[] = ['online', 'warning', 'fault', 'offline'];

export function isDeviceStatus(value: unknown): value is DeviceStatus {
  return typeof value === 'string' && DEVICE_STATUS_VALUES.includes(value as DeviceStatus);
}

export function isDevice(obj: unknown): obj is Device {
  if (!hasProperties(obj, ['id', 'name', 'type', 'status', 'location'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'string' &&
    typeof record.name === 'string' &&
    typeof record.type === 'string' &&
    isDeviceStatus(record.status) &&
    typeof record.location === 'string'
  );
}

export function isDeviceArray(arr: unknown): arr is Device[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isDevice);
}

// ============== Telemetry 类型守卫 ==============

export function isTelemetry(obj: unknown): obj is Telemetry {
  if (!hasProperties(obj, ['device_id', 'timestamp'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.device_id === 'string' &&
    typeof record.timestamp === 'string' &&
    (typeof record.temperature === 'number' || record.temperature === undefined) &&
    (typeof record.pressure === 'number' || record.pressure === undefined) &&
    (typeof record.vibration === 'number' || record.vibration === undefined) &&
    (typeof record.power === 'number' || record.power === undefined) &&
    (isDeviceStatus(record.status) || record.status === undefined)
  );
}

export function isTelemetryArray(arr: unknown): arr is Telemetry[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isTelemetry);
}

// ============== DeviceStats 类型守卫 ==============

export function isDeviceStats(obj: unknown): obj is DeviceStats {
  if (!hasProperties(obj, ['device_id', 'avg_temperature', 'avg_vibration'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.device_id === 'string' &&
    typeof record.avg_temperature === 'number' &&
    typeof record.avg_vibration === 'number' &&
    (typeof record.max_temperature === 'number' || record.max_temperature === undefined) &&
    (typeof record.max_vibration === 'number' || record.max_vibration === undefined) &&
    (typeof record.data_points === 'number' || record.data_points === undefined)
  );
}

// ============== Alert 类型守卫 ==============

const ALERT_SEVERITY_VALUES: AlertSeverity[] = ['low', 'medium', 'high', 'critical'];
const ALERT_STATUS_VALUES = ['active', 'acknowledged', 'resolved'];

export function isAlertSeverity(value: unknown): value is AlertSeverity {
  return typeof value === 'string' && ALERT_SEVERITY_VALUES.includes(value as AlertSeverity);
}

export function isAlert(obj: unknown): obj is Alert {
  // metric field is optional in backend response
  if (!hasProperties(obj, ['id', 'rule_id', 'device_id', 'value', 'threshold', 'severity', 'status', 'triggered_at'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.rule_id === 'number' &&
    typeof record.device_id === 'string' &&
    (record.metric === undefined || typeof record.metric === 'string') &&
    typeof record.value === 'number' &&
    typeof record.threshold === 'number' &&
    isAlertSeverity(record.severity) &&
    typeof record.status === 'string' &&
    ALERT_STATUS_VALUES.includes(record.status as 'active' | 'resolved') &&
    typeof record.triggered_at === 'string'
  );
}

export function isAlertArray(arr: unknown): arr is Alert[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isAlert);
}

export function isAlertStatus(value: unknown): value is Alert['status'] {
  return typeof value === 'string' && ALERT_STATUS_VALUES.includes(value as 'active' | 'resolved');
}

// ============== AlertRule 类型守卫 ==============

const ALERT_OPERATOR_VALUES = ['>', '>=', '<', '<=', '==', '!='];

export function isAlertOperator(value: unknown): value is '>' | '>=' | '<' | '<=' | '==' | '!=' {
  return typeof value === 'string' && ALERT_OPERATOR_VALUES.includes(value);
}

export function isAlertRule(obj: unknown): obj is AlertRule {
  if (!hasProperties(obj, ['id', 'name', 'device_type', 'metric', 'operator', 'threshold', 'severity', 'enabled', 'cooldown_sec'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.name === 'string' &&
    typeof record.device_type === 'string' &&
    typeof record.metric === 'string' &&
    isAlertOperator(record.operator) &&
    typeof record.threshold === 'number' &&
    isAlertSeverity(record.severity) &&
    typeof record.enabled === 'boolean' &&
    typeof record.cooldown_sec === 'number'
  );
}

export function isAlertRuleArray(arr: unknown): arr is AlertRule[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isAlertRule);
}

// ============== WorkOrder 类型守卫 ==============

const WORK_ORDER_PRIORITY_VALUES = ['urgent', 'high', 'medium', 'low'];
const WORK_ORDER_STATUS_VALUES = ['pending', 'in_progress', 'completed', 'cancelled'];

export function isWorkOrder(obj: unknown): obj is WorkOrder {
  if (!hasProperties(obj, ['id', 'title', 'description', 'device_id', 'priority', 'status', 'created_at'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.title === 'string' &&
    typeof record.description === 'string' &&
    typeof record.device_id === 'string' &&
    typeof record.priority === 'string' &&
    WORK_ORDER_PRIORITY_VALUES.includes(record.priority) &&
    typeof record.status === 'string' &&
    WORK_ORDER_STATUS_VALUES.includes(record.status) &&
    typeof record.created_at === 'string'
  );
}

export function isWorkOrderArray(arr: unknown): arr is WorkOrder[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isWorkOrder);
}

// ============== Notification 类型守卫 ==============

const NOTIFICATION_TYPE_VALUES = ['alert', 'system', 'work_order'];

export function isNotification(obj: unknown): obj is Notification {
  if (!hasProperties(obj, ['id', 'type', 'title', 'message', 'read', 'created_at'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.type === 'string' &&
    NOTIFICATION_TYPE_VALUES.includes(record.type) &&
    typeof record.title === 'string' &&
    typeof record.message === 'string' &&
    typeof record.read === 'boolean' &&
    typeof record.created_at === 'string'
  );
}

export function isNotificationArray(arr: unknown): arr is Notification[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isNotification);
}

// ============== Report 类型守卫 ==============

const REPORT_TYPE_VALUES = ['daily', 'device', 'maintenance', 'anomaly', 'comprehensive'];

export function isReport(obj: unknown): obj is Report {
  if (!hasProperties(obj, ['id', 'title', 'type', 'content', 'generated_at'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.title === 'string' &&
    typeof record.type === 'string' &&
    REPORT_TYPE_VALUES.includes(record.type) &&
    typeof record.content === 'string' &&
    typeof record.generated_at === 'string'
  );
}

export function isReportArray(arr: unknown): arr is Report[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isReport);
}

// ============== User 类型守卫 ==============

const USER_ROLE_VALUES: UserRole[] = ['admin', 'user', 'viewer'];

export function isUserRole(value: unknown): value is UserRole {
  return typeof value === 'string' && USER_ROLE_VALUES.includes(value as UserRole);
}

export function isUser(obj: unknown): obj is User {
  if (!hasProperties(obj, ['id', 'username', 'email', 'role', 'created_at'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.username === 'string' &&
    typeof record.email === 'string' &&
    isUserRole(record.role) &&
    typeof record.created_at === 'string'
  );
}

export function isUserArray(arr: unknown): arr is User[] {
  if (!Array.isArray(arr)) return false;
  return arr.every(isUser);
}

// ============== ROIStats 类型守卫 ==============

export function isROIStats(obj: unknown): obj is ROIStats {
  if (!hasProperties(obj, ['total_devices', 'active_alerts', 'open_work_orders'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.total_devices === 'number' &&
    typeof record.active_alerts === 'number' &&
    typeof record.open_work_orders === 'number' &&
    typeof record.resolved_issues === 'number' &&
    typeof record.predicted_savings === 'number' &&
    typeof record.uptime_percentage === 'number' &&
    typeof record.avg_response_time_hours === 'number'
  );
}

// ============== SystemStatus 类型守卫 ==============

const DATABASE_STATUS_VALUES = ['healthy', 'unhealthy'];

export function isSystemStatus(obj: unknown): obj is SystemStatus {
  if (!hasProperties(obj, ['database', 'db_latency_ms', 'uptime', 'version', 'timestamp'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.database === 'string' &&
    DATABASE_STATUS_VALUES.includes(record.database) &&
    typeof record.db_latency_ms === 'number' &&
    typeof record.uptime === 'string' &&
    typeof record.version === 'string' &&
    typeof record.timestamp === 'string'
  );
}

// ============== DeviceGraph 类型守卫 ==============

export function isGraphNode(obj: unknown): obj is { id: string; name: string; type: string; status: DeviceStatus } {
  if (!hasProperties(obj, ['id', 'name', 'type', 'status'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'string' &&
    typeof record.name === 'string' &&
    typeof record.type === 'string' &&
    isDeviceStatus(record.status)
  );
}

export function isGraphLink(obj: unknown): obj is { source: string; target: string; type: string } {
  if (!hasProperties(obj, ['source', 'target', 'type'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.source === 'string' &&
    typeof record.target === 'string' &&
    typeof record.type === 'string'
  );
}

export function isDeviceGraph(obj: unknown): obj is DeviceGraph {
  if (!hasProperties(obj, ['nodes', 'links'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    Array.isArray(record.nodes) &&
    Array.isArray(record.links) &&
    record.nodes.every(isGraphNode) &&
    record.links.every(isGraphLink)
  );
}

// ============== AgentResponse 类型守卫 ==============

export function isAgentResponse(obj: unknown): obj is AgentResponse {
  if (!hasProperties(obj, ['session_id', 'response', 'agent'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.session_id === 'string' &&
    typeof record.response === 'string' &&
    typeof record.agent === 'string'
  );
}

// ============== 安全类型转换函数 ==============

/**
 * 安全地转换为 Device，如果类型不匹配返回 null
 */
export function asDeviceSafe(obj: unknown): Device | null {
  return isDevice(obj) ? obj : null;
}

/**
 * 安全地转换为 Telemetry 数组，如果类型不匹配返回空数组
 */
export function asTelemetryArraySafe(arr: unknown): Telemetry[] {
  return isTelemetryArray(arr) ? arr : [];
}

/**
 * 安全地转换为 DeviceStats，如果类型不匹配返回默认值
 */
export function asDeviceStatsSafe(obj: unknown): DeviceStats | null {
  return isDeviceStats(obj) ? obj : null;
}

/**
 * 安全地转换为 AlertRule 数组，如果类型不匹配返回空数组
 */
export function asAlertRuleArraySafe(arr: unknown): AlertRule[] {
  return isAlertRuleArray(arr) ? arr : [];
}

/**
 * 安全地转换为 WorkOrder 数组，如果类型不匹配返回空数组
 */
export function asWorkOrderArraySafe(arr: unknown): WorkOrder[] {
  return isWorkOrderArray(arr) ? arr : [];
}

/**
 * 安全地转换为 Notification 数组，如果类型不匹配返回空数组
 */
export function asNotificationArraySafe(arr: unknown): Notification[] {
  return isNotificationArray(arr) ? arr : [];
}

/**
 * 安全地转换为 Report 数组，如果类型不匹配返回空数组
 */
export function asReportArraySafe(arr: unknown): Report[] {
  return isReportArray(arr) ? arr : [];
}

/**
 * 安全地转换为 User 数组，如果类型不匹配返回空数组
 */
export function asUserArraySafe(arr: unknown): User[] {
  return isUserArray(arr) ? arr : [];
}

/**
 * 安全地转换为 ROIStats，如果类型不匹配返回 null
 */
export function asROIStatsSafe(obj: unknown): ROIStats | null {
  return isROIStats(obj) ? obj : null;
}

/**
 * 安全地转换为 SystemStatus，如果类型不匹配返回 null
 */
export function asSystemStatusSafe(obj: unknown): SystemStatus | null {
  return isSystemStatus(obj) ? obj : null;
}

/**
 * 安全地转换为 DeviceGraph，如果类型不匹配返回 null
 */
export function asDeviceGraphSafe(obj: unknown): DeviceGraph | null {
  return isDeviceGraph(obj) ? obj : null;
}

/**
 * 安全地转换为 AgentResponse，如果类型不匹配返回 null
 */
export function asAgentResponseSafe(obj: unknown): AgentResponse | null {
  return isAgentResponse(obj) ? obj : null;
}

/**
 * 安全地转换为 Alert 状态
 */
export function asAlertStatusSafe(value: unknown): Alert['status'] | null {
  return isAlertStatus(value) ? value : null;
}

// ============== Alert 状态更新 Payload 类型守卫 ==============

/**
 * 检查是否为 Alert 状态更新的 payload（用于 alert_resolved, alert_acknowledged 消息）
 */
export function isAlertStatusPayload(obj: unknown): obj is { id: number; status: string } {
  if (!hasProperties(obj, ['id', 'status'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.status === 'string'
  );
}

// ============== TelemetryData 类型守卫（用于 WebSocket 消息）==============

/**
 * 检查是否为 TelemetryData 对象（用于 WebSocket telemetry 消息）
 */
export function isTelemetryData(obj: unknown): obj is { device_id: string; timestamp: string; status: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number } {
  if (!hasProperties(obj, ['device_id', 'timestamp', 'status'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.device_id === 'string' &&
    typeof record.timestamp === 'string' &&
    typeof record.status === 'string' &&
    (record.temperature === undefined || typeof record.temperature === 'number') &&
    (record.pressure === undefined || typeof record.pressure === 'number') &&
    (record.vibration === undefined || typeof record.vibration === 'number') &&
    (record.power === undefined || typeof record.power === 'number') &&
    (record.humidity === undefined || typeof record.humidity === 'number')
  );
}

/**
 * 检查是否为 TelemetryData 数组
 */
export function isTelemetryDataArray(arr: unknown): arr is Array<{ device_id: string; timestamp: string; status: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number }> {
  if (!Array.isArray(arr)) return false;
  return arr.every(isTelemetryData);
}

// ============== TelemetryHistory 类型守卫 ==============

/**
 * 检查是否为 TelemetryHistory 对象
 */
export function isTelemetryHistory(obj: unknown): obj is { id: number; device_id: string; timestamp: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number } {
  if (!hasProperties(obj, ['id', 'device_id', 'timestamp'])) return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.device_id === 'string' &&
    typeof record.timestamp === 'string' &&
    (record.temperature === undefined || typeof record.temperature === 'number') &&
    (record.pressure === undefined || typeof record.pressure === 'number') &&
    (record.vibration === undefined || typeof record.vibration === 'number') &&
    (record.power === undefined || typeof record.power === 'number') &&
    (record.humidity === undefined || typeof record.humidity === 'number')
  );
}

/**
 * 检查是否为 TelemetryHistory 数组
 */
export function isTelemetryHistoryArray(arr: unknown): arr is Array<{ id: number; device_id: string; timestamp: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number }> {
  if (!Array.isArray(arr)) return false;
  return arr.every(isTelemetryHistory);
}

// ============== 安全类型转换函数（新增）==============

/**
 * 安全地转换为 TelemetryData，如果类型不匹配返回 null
 */
export function asTelemetryDataSafe(obj: unknown): { device_id: string; timestamp: string; status: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number } | null {
  return isTelemetryData(obj) ? obj : null;
}

/**
 * 安全地转换为 TelemetryData 数组，如果类型不匹配返回空数组
 */
export function asTelemetryDataArraySafe(arr: unknown): Array<{ device_id: string; timestamp: string; status: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number }> {
  return isTelemetryDataArray(arr) ? arr : [];
}

/**
 * 安全地转换为 TelemetryHistory 数组，如果类型不匹配返回空数组
 */
export function asTelemetryHistoryArraySafe(arr: unknown): Array<{ id: number; device_id: string; timestamp: string; temperature?: number; pressure?: number; vibration?: number; power?: number; humidity?: number }> {
  return isTelemetryHistoryArray(arr) ? arr : [];
}

/**
 * 安全地转换为 AlertStatusPayload，如果类型不匹配返回 null
 */
export function asAlertStatusPayloadSafe(obj: unknown): { id: number; status: string } | null {
  return isAlertStatusPayload(obj) ? obj : null;
}

export default {
  isDeviceStatus,
  isDevice,
  isDeviceArray,
  isTelemetry,
  isTelemetryArray,
  isDeviceStats,
  isAlertSeverity,
  isAlert,
  isAlertArray,
  isAlertStatus,
  isAlertOperator,
  isAlertRule,
  isAlertRuleArray,
  isWorkOrder,
  isWorkOrderArray,
  isNotification,
  isNotificationArray,
  isReport,
  isReportArray,
  isUserRole,
  isUser,
  isUserArray,
  isROIStats,
  isSystemStatus,
  isDeviceGraph,
  isAgentResponse,
  asDeviceSafe,
  asTelemetryArraySafe,
  asDeviceStatsSafe,
  asAlertRuleArraySafe,
  asWorkOrderArraySafe,
  asNotificationArraySafe,
  asReportArraySafe,
  asUserArraySafe,
  asROIStatsSafe,
  asSystemStatusSafe,
  asDeviceGraphSafe,
  asAgentResponseSafe,
  asAlertStatusSafe,
  isAlertStatusPayload,
  isTelemetryData,
  isTelemetryDataArray,
  isTelemetryHistory,
  isTelemetryHistoryArray,
  asTelemetryDataSafe,
  asTelemetryDataArraySafe,
  asTelemetryHistoryArraySafe,
  asAlertStatusPayloadSafe,
};