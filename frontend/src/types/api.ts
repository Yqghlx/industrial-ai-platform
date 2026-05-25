/**
 * API Types for Industrial AI Platform
 * Complete interface definitions for all API responses and requests
 */

// ============== Device Types ==============

export type DeviceStatus = 'online' | 'warning' | 'fault' | 'offline';
export type DeviceType = 'pump' | 'motor' | 'compressor' | 'conveyor' | 'valve' | 'sensor' | 'other';

export interface Device {
  id: string;
  name: string;
  type: string;
  status: DeviceStatus;
  location: string;
  description?: string;
}

export interface DeviceCreateInput {
  id: string;
  name: string;
  type: string;
  status: DeviceStatus;
  location: string;
  description?: string;
}

export interface DeviceUpdateInput {
  name?: string;
  type?: string;
  status?: DeviceStatus;
  location?: string;
  description?: string;
}

// ============== Telemetry Types ==============

export interface Telemetry {
  id: number;
  device_id: string;
  timestamp: string;
  time?: string; // Alternative timestamp field
  temperature: number;
  pressure: number;
  vibration: number;
  power: number;
  status: DeviceStatus;
}

export interface TelemetryInput {
  device_id: string;
  temperature?: number;
  pressure?: number;
  vibration?: number;
  power?: number;
  status?: DeviceStatus;
}

export interface DeviceStats {
  device_id: string;
  avg_temperature: number;
  avg_vibration: number;
  max_temperature: number;
  max_vibration: number;
  data_points: number;
}

export interface LatestTelemetry {
  device_id: string;
  device_name?: string;
  temperature?: number;
  pressure?: number;
  vibration?: number;
  power?: number;
  status: DeviceStatus;
  timestamp: string;
}

// ============== Alert/Rule Types ==============

export type AlertSeverity = 'low' | 'medium' | 'high' | 'critical';
export type AlertOperator = '>' | '>=' | '<' | '<=' | '==' | '!=';

export interface AlertRule {
  id: number;
  name: string;
  device_type: string;
  metric: string;
  operator: AlertOperator;
  threshold: number;
  severity: AlertSeverity;
  enabled: boolean;
  cooldown_sec: number;
  actions?: string;
  created_at?: string;
  updated_at?: string;
}

export interface AlertRuleCreateInput {
  name: string;
  device_type: DeviceType;
  metric: string;
  operator: AlertOperator;
  threshold: number;
  severity: AlertSeverity;
  enabled?: boolean;
  cooldown_sec?: number;
  actions?: string;
}

export interface AlertRuleUpdateInput {
  name?: string;
  device_type?: DeviceType;
  metric?: string;
  operator?: AlertOperator;
  threshold?: number;
  severity?: AlertSeverity;
  enabled?: boolean;
  cooldown_sec?: number;
  actions?: string;
}

export interface Alert {
  id: number;
  rule_id: number;
  rule_name?: string;
  device_id: string;
  metric: string;
  value: number;
  threshold: number;
  message?: string;
  severity: AlertSeverity;
  status: 'active' | 'acknowledged' | 'resolved';
  triggered_at: string;
  resolved_at?: string;
}

// ============== Work Order Types ==============

export type WorkOrderPriority = 'urgent' | 'high' | 'medium' | 'low';
export type WorkOrderStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';

export interface WorkOrder {
  id: number;
  title: string;
  description: string;
  device_id: string;
  priority: WorkOrderPriority;
  status: WorkOrderStatus;
  assignee?: string;
  created_at: string;
  updated_at?: string;
}

export interface WorkOrderCreateInput {
  title: string;
  description?: string;
  device_id?: string;
  priority: WorkOrderPriority;
  assignee?: string;
}

export interface WorkOrderUpdateInput {
  status: WorkOrderStatus;
}

// ============== Notification Types ==============

export type NotificationType = 'alert' | 'system' | 'work_order';

export interface Notification {
  id: number;
  type: NotificationType;
  title: string;
  message: string;
  device_id?: string;
  read: boolean;
  created_at: string;
}

// ============== Report Types ==============

export type ReportType = 'daily' | 'device' | 'maintenance' | 'anomaly' | 'comprehensive';

export interface Report {
  id: number;
  title: string;
  type: ReportType;
  device_id?: string;
  content: string;
  generated_at: string;
}

export interface ReportCreateInput {
  type: ReportType;
  device_id?: string;
}

// ============== Black Box Types ==============

export type TriggerType = 'alert' | 'fault' | 'manual' | 'system';

export interface BlackBoxRecord {
  id: number;
  device_id: string;
  trigger_type: TriggerType;
  start_time: string;
  end_time: string;
  summary?: string;
  snapshot?: Telemetry[];
  created_at: string;
}

export interface BlackBoxData {
  record_id: number;
  data: Telemetry[];
  summary?: string;
}

// ============== Agent Types ==============

export interface AgentResponse {
  session_id: string;
  response: string;
  agent: string;
}

export interface AgentQueryInput {
  query: string;
  device_id?: string;
}

export interface AgentLog {
  id: number;
  session_id: string;
  query: string;
  response: string;
  agent: string;
  device_id?: string;
  created_at: string;
}

// ============== ROI Stats Types ==============

export interface ROIStats {
  total_devices: number;
  active_alerts: number;
  open_work_orders: number;
  resolved_issues: number;
  predicted_savings: number;
  uptime_percentage: number;
  avg_response_time_hours: number;
}

// ============== User Types ==============

export type UserRole = 'admin' | 'user' | 'viewer';

export interface User {
  id: number;
  username: string;
  email: string;
  role: UserRole;
  created_at: string;
  last_login?: string;
}

export interface UserCreateInput {
  username: string;
  password: string;
  email: string;
  role?: UserRole;
}

export interface UserUpdateInput {
  username?: string;
  email?: string;
  role?: UserRole;
  password?: string;
}

// ============== Auth Response Types ==============

export interface LoginResponse {
  token: string;
  user: User;
}

export interface RegisterResponse {
  token: string;
  user: User;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface TokenPairResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface ValidateTokenResponse {
  valid: boolean;
  user_id: number;
  username: string;
  role: string;
  expires_at: string;
}

export interface ChangePasswordResponse {
  message: string;
  token_revoked: boolean;
}

export interface LogoutResponse {
  message: string;
}

// ============== System Status Types ==============

export interface SystemStatus {
  database: 'healthy' | 'unhealthy';
  db_latency_ms: number;
  uptime: string;
  version: string;
  timestamp: string;
  device_count: number;
  user_count: number;
}

export interface HealthCheck {
  status: 'ok' | 'error';
  database: 'healthy' | 'unhealthy';
  version: string;
  timestamp: string;
}

// ============== Graph Types ==============

export interface GraphNode {
  id: string;
  name: string;
  type: string;
  status: DeviceStatus;
}

export interface GraphLink {
  source: string;
  target: string;
  type: string;
}

export interface DeviceGraph {
  nodes: GraphNode[];
  links: GraphLink[];
}

// ============== Pagination Types ==============

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page?: number;
  page_size?: number;
}

export interface PaginationParams {
  page?: number;
  page_size?: number;
}

// ============== API Error Types ==============

export interface ApiError {
  error: string;
  code: string;
  details?: Record<string, unknown>;
}

export interface MessageResponse {
  message: string;
}

// ============== Cache Types ==============

export interface CacheStatusResponse {
  available: boolean;
  backend_type: string;
  message?: string;
  stats?: {
    keys: number;
    memory_usage: number;
    hit_rate: number;
  };
}

// ============== WebSocket Types ==============

export interface WSCompressionStats {
  enabled: boolean;
  total_messages: number;
  compressed_messages: number;
  skipped_messages: number;
  original_bytes: number;
  compressed_bytes: number;
  compression_ratio: number;
  savings_percent: number;
}

export interface WSMessage {
  type: string;
  payload?: unknown;
  timestamp: string;
}

// ============== Tenant Types ==============

export interface Tenant {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at?: string;
  is_active: boolean;
}

export interface TenantCreateInput {
  name: string;
  description?: string;
}

export interface TenantUpdateInput {
  name?: string;
  description?: string;
  is_active?: boolean;
}

// ============== RBAC Types ==============

export interface Role {
  id: number;
  name: string;
  description?: string;
  permissions: Permission[];
  created_at: string;
}

export interface Permission {
  id: number;
  name: string;
  resource: string;
  action: string;
  description?: string;
}

export interface RoleCreateInput {
  name: string;
  description?: string;
  permissions?: number[];
}

export interface RoleUpdateInput {
  name?: string;
  description?: string;
  permissions?: number[];
}

// ============== Export Types ==============

export interface ExportResponse {
  data: Blob;
  filename: string;
  mimeType: string;
}

export type ExportFormat = 'pdf' | 'xlsx' | 'csv';
export type ExportReportType = 'devices' | 'alerts' | 'roi';

// ============== Rate Limit Types ==============

export interface RateLimitInfo {
  limit: number;
  remaining: number;
  reset: number;
}

// ============== Error Detail Types ==============

export interface ValidationError {
  field: string;
  message: string;
}

export interface DetailedApiError {
  error: string;
  code: string;
  details?: ValidationError[] | Record<string, unknown>;
  request_id?: string;
}