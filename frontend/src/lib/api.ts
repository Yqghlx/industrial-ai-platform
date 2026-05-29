/**
 * API Client for Industrial AI Platform
 * 
 * FE-P3-04: Added comprehensive documentation for API client
 * FE-P3-03: Token storage uses sessionStorage with 'token' key consistently
 * SEC-LOW-02: Use sessionStorage instead of localStorage for better security
 * 
 * @description Handles all API requests with authentication, timeout support,
 * and request cancellation capabilities.
 * 
 * Security Note: Token is stored in sessionStorage to reduce XSS attack window.
 * SessionStorage is cleared when the browser tab is closed, limiting exposure time.
 */

import {
  Device,
  DeviceCreateInput,
  DeviceUpdateInput,
  Telemetry,
  TelemetryInput,
  DeviceStats,
  LatestTelemetry,
  AlertRule,
  AlertRuleCreateInput,
  AlertRuleUpdateInput,
  WorkOrder,
  WorkOrderCreateInput,
  Notification,
  Report,
  BlackBoxRecord,
  BlackBoxData,
  AgentResponse,
  AgentLog,
  ROIStats,
  User,
  UserCreateInput,
  SystemStatus,
  HealthCheck,
  DeviceGraph,
  PaginatedResponse,
  LoginResponse,
  RegisterResponse,
  MessageResponse,
  ApiError,
} from '../types/api';

/** Base API path for all endpoints */
const API_BASE = '/api/v1';

/** Timeout constants (in milliseconds) */
const DEFAULT_TIMEOUT = 30000; // 30 seconds for regular requests
const AGENT_TIMEOUT = 60000;   // 60 seconds for AI Agent queries (longer response time)

/**
 * Custom error class for timeout errors
 * @extends Error
 */
class TimeoutError extends Error {
  constructor(message: string = '请求超时，请稍后重试') {
    super(message);
    this.name = 'TimeoutError';
  }
}

/**
 * API Client class
 * 
 * @description Main API client with authentication token management,
 * request cancellation, and timeout handling.
 * 
 * Token Storage: Uses sessionStorage with key 'token' for consistency.
 * (FE-P3-03: Unified token storage)
 */
class ApiClient {
  /** API base URL */
  private baseUrl: string;
  /** Current authentication token */
  private token: string | null = null;
  /** Map of active requests for cancellation support */
  private activeControllers: Map<string, AbortController> = new Map();

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
    this.loadToken();
  }

  /** Load token from sessionStorage on initialization */
  private loadToken() {
    this.token = sessionStorage.getItem('token');
  }

  /** 
   * Set authentication token
   * @param token - JWT token string or null to clear
   * 
   * Security: Token stored in sessionStorage (cleared on tab close)
   */
  setToken(token: string | null) {
    this.token = token;
    if (token) {
      sessionStorage.setItem('token', token);
    } else {
      sessionStorage.removeItem('token');
    }
  }

  /** Get current authentication token */
  getToken(): string | null {
    return this.token;
  }

  /**
   * Cancel an active request by its request ID
   * @param requestId - Unique identifier for the request
   */
  cancelRequest(requestId: string): void {
    const controller = this.activeControllers.get(requestId);
    if (controller) {
      controller.abort();
      this.activeControllers.delete(requestId);
    }
  }

  /**
   * Cancel all active requests
   */
  cancelAllRequests(): void {
    this.activeControllers.forEach((controller) => controller.abort());
    this.activeControllers.clear();
  }

  private async request<T>(
    method: string,
    path: string,
    data?: unknown,
    params?: Record<string, string>,
    timeout: number = DEFAULT_TIMEOUT
  ): Promise<T> {
    const url = new URL(`${this.baseUrl}${path}`, window.location.origin);
    
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.set(key, value);
      });
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    // Create AbortController for timeout and cancellation
    const controller = new AbortController();
    const requestId = `${method}-${path}-${Date.now()}`;
    this.activeControllers.set(requestId, controller);
    
    const timeoutId = setTimeout(() => {
      controller.abort();
    }, timeout);

    const options: RequestInit = {
      method,
      headers,
      signal: controller.signal,
    };

    if (data && method !== 'GET') {
      options.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url.toString(), options);

      if (response.status === 401) {
        this.setToken(null);
        window.dispatchEvent(new CustomEvent('auth:logout'));
        throw new Error('Unauthorized');
      }

      if (!response.ok) {
        const error: ApiError = await response.json();
        throw new Error(error.error || 'Request failed');
      }

      return response.json();
    } catch (error) {
      // Handle abort/timeout errors with friendly message
      if (error instanceof Error && error.name === 'AbortError') {
        throw new TimeoutError();
      }
      throw error;
    } finally {
      clearTimeout(timeoutId);
      this.activeControllers.delete(requestId);
    }
  }

  // Auth
  async login(username: string, password: string): Promise<LoginResponse> {
    // 后端返回格式: { token, user }
    const response = await this.request<{
      token: string;
      user: {
        id: number;
        username: string;
        email?: string;
        role: string;
        tenant_id?: string;
      };
    }>(
      'POST',
      '/auth/login',
      { username, password }
    );
    
    // 后端返回 token 字段
    const token = response.token;
    if (!token) {
      throw new Error('Login failed: no token in response');
    }
    
    this.setToken(token);
    
    // 后端返回了 user 信息，直接使用
    const user = {
      id: response.user?.id || 1,
      username: response.user?.username || username,
      email: `${response.user?.username || username}@example.com`,
      role: (response.user?.role || (username === 'admin' ? 'admin' : 'operator')) as User['role'],
      created_at: new Date().toISOString()
    };
    
    return { token, user };
  }
  
  // 从 JWT token 解析用户信息
  private parseUserFromToken(token: string): User | null {
    try {
      // JWT 格式: header.payload.signature
      const payload = token.split('.')[1];
      if (!payload) return null;
      
      // Base64 解码
      const decoded = JSON.parse(atob(payload));
      
      // 后端 Claims 字段: user_id, username, role, tenant_id, token_type, sub="user:{id}"
      return {
        id: decoded.user_id || (decoded.sub ? parseInt(decoded.sub.split(':')[1]) : 0),
        username: decoded.username || (decoded.sub ? decoded.sub.split(':')[1] : 'unknown'),
        email: `${decoded.username || decoded.user_id || 'unknown'}@example.com`,
        role: decoded.role || 'operator',
        created_at: new Date().toISOString()
      };
    } catch {
      return null;
    }
  }

  async register(data: { username: string; password: string; email: string; role?: string }): Promise<RegisterResponse> {
    const response = await this.request<RegisterResponse>(
      'POST',
      '/auth/register',
      data
    );
    this.setToken(response.token);
    return response;
  }

  // Devices
  async getDevices(page = 1, pageSize = 20): Promise<PaginatedResponse<Device>> {
    const response = await this.request<{ data: Device[]; total: number; page: number; page_size: number }>(
      'GET',
      '/devices',
      undefined,
      { page: String(page), page_size: String(pageSize) }
    );
    // Backend now returns 'data' field directly
    return {
      data: response.data,
      total: response.total,
      page: response.page,
      page_size: response.page_size,
    };
  }

  async getDevice(id: string): Promise<Device> {
    return this.request<Device>('GET', `/devices/${id}`);
  }

  async createDevice(data: DeviceCreateInput): Promise<Device> {
    return this.request<Device>('POST', '/devices', data);
  }

  async updateDevice(id: string, data: DeviceUpdateInput): Promise<Device> {
    return this.request<Device>('PUT', `/devices/${id}`, data);
  }

  async deleteDevice(id: string): Promise<MessageResponse> {
    return this.request<MessageResponse>('DELETE', `/devices/${id}`);
  }

  async getLatestTelemetry(): Promise<{ data: LatestTelemetry[] }> {
    return this.request<{ data: LatestTelemetry[] }>('GET', '/devices/latest');
  }

  async getDeviceTelemetry(id: string, range = '1h', limit = 1000): Promise<{ data: Telemetry[] }> {
    return this.request<{ data: Telemetry[] }>(
      'GET',
      `/devices/${id}/telemetry`,
      undefined,
      { range, limit: String(limit) }
    );
  }

  async getDeviceStats(id: string, range = '24h'): Promise<DeviceStats> {
    return this.request<DeviceStats>('GET', `/devices/${id}/stats`, undefined, { range });
  }

  async getDeviceGraph(): Promise<DeviceGraph> {
    return this.request<DeviceGraph>('GET', '/devices/graph');
  }

  // Telemetry (public)
  async ingestTelemetry(data: TelemetryInput): Promise<MessageResponse> {
    return this.request<MessageResponse>('POST', '/devices/telemetry', data);
  }

  // Rules
  async getRules(): Promise<{ data: AlertRule[] }> {
    const response = await this.request<{ data: AlertRule[]; total?: number; page?: number; page_size?: number }>('GET', '/rules');
    return { data: response.data };
  }

  async getRule(id: number): Promise<AlertRule> {
    return this.request<AlertRule>('GET', `/rules/${id}`);
  }

  async createRule(data: AlertRuleCreateInput): Promise<AlertRule> {
    return this.request<AlertRule>('POST', '/rules', data);
  }

  async updateRule(id: number, data: AlertRuleUpdateInput): Promise<AlertRule> {
    return this.request<AlertRule>('PUT', `/rules/${id}`, data);
  }

  async deleteRule(id: number): Promise<MessageResponse> {
    return this.request<MessageResponse>('DELETE', `/rules/${id}`);
  }

  async toggleRule(id: number, enabled: boolean): Promise<MessageResponse> {
    return this.request<MessageResponse>('PUT', `/rules/${id}/toggle`, { enabled });
  }

  // Alerts
  async getAlerts(params?: { status?: string; severity?: string; page?: number; page_size?: number }): Promise<{ data: unknown[]; total?: number }> {
    const queryParams: Record<string, string> = {};
    if (params?.status) queryParams.status = params.status;
    if (params?.severity) queryParams.severity = params.severity;
    if (params?.page) queryParams.page = String(params.page);
    if (params?.page_size) queryParams.page_size = String(params.page_size);
    return this.request<{ data: unknown[]; total?: number }>('GET', '/alerts', undefined, queryParams);
  }

  async getAlertStats(): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('GET', '/alerts/stats');
  }

  async resolveAlert(id: number): Promise<MessageResponse> {
    return this.request<MessageResponse>('PUT', `/alerts/${id}/resolve`);
  }

  async acknowledgeAlert(id: number): Promise<MessageResponse> {
    return this.request<MessageResponse>('PUT', `/alerts/${id}/acknowledge`);
  }

  // Alert Reports
  async getTrendReport(days: number = 7): Promise<{ data: unknown[] }> {
    return this.request<{ data: unknown[] }>('GET', '/alerts/report/trend', undefined, { days: String(days) });
  }

  async getRankingReport(limit: number = 10): Promise<{ data: unknown[] }> {
    return this.request<{ data: unknown[] }>('GET', '/alerts/report/ranking', undefined, { limit: String(limit) });
  }

  async getEfficiencyReport(): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('GET', '/alerts/report/efficiency');
  }

  // Agent
  async agentQuery(query: string, deviceId?: string): Promise<AgentResponse> {
    return this.request<AgentResponse>(
      'POST',
      '/agent/query',
      { query, device_id: deviceId },
      undefined,
      AGENT_TIMEOUT // Use extended timeout for AI Agent queries
    );
  }

  async getAgentLogs(limit = 50): Promise<{ data: AgentLog[] }> {
    return this.request<{ data: AgentLog[] }>('GET', '/ai/status', undefined, { limit: String(limit) });
  }

  // Work Orders
  async getWorkOrders(params?: { status?: string; device_id?: string; page?: number; page_size?: number }): Promise<PaginatedResponse<WorkOrder>> {
    const queryParams: Record<string, string> = {};
    if (params?.status) queryParams.status = params.status;
    if (params?.device_id) queryParams.device_id = params.device_id;
    if (params?.page) queryParams.page = String(params.page);
    if (params?.page_size) queryParams.page_size = String(params.page_size);
    return this.request<PaginatedResponse<WorkOrder>>('GET', '/work-orders', undefined, queryParams);
  }

  async createWorkOrder(data: WorkOrderCreateInput): Promise<WorkOrder> {
    return this.request<WorkOrder>('POST', '/work-orders', data);
  }

  async updateWorkOrderStatus(id: number, status: string): Promise<MessageResponse> {
    return this.request<MessageResponse>('PUT', `/work-orders/${id}/status`, { status });
  }

  // Notifications
  async getNotifications(params?: { type?: string; unread?: boolean; page?: number; page_size?: number }): Promise<PaginatedResponse<Notification>> {
    const queryParams: Record<string, string> = {};
    if (params?.type) queryParams.type = params.type;
    if (params?.unread) queryParams.unread = 'true';
    if (params?.page) queryParams.page = String(params.page);
    if (params?.page_size) queryParams.page_size = String(params.page_size);
    return this.request<PaginatedResponse<Notification>>('GET', '/notifications', undefined, queryParams);
  }

  async markNotificationRead(id: number): Promise<MessageResponse> {
    return this.request<MessageResponse>('PUT', `/notifications/${id}/read`);
  }

  // Black Box
  async getBlackBoxRecords(params?: { device_id?: string; page?: number; page_size?: number }): Promise<PaginatedResponse<BlackBoxRecord>> {
    const queryParams: Record<string, string> = {};
    if (params?.device_id) queryParams.device_id = params.device_id;
    if (params?.page) queryParams.page = String(params.page);
    if (params?.page_size) queryParams.page_size = String(params.page_size);
    return this.request<PaginatedResponse<BlackBoxRecord>>('GET', '/blackbox', undefined, queryParams);
  }

  async getBlackBoxData(id: number): Promise<BlackBoxData> {
    return this.request<BlackBoxData>('GET', `/blackbox/${id}/data`);
  }

  // Reports
  async getReports(params?: { type?: string; page?: number; page_size?: number }): Promise<PaginatedResponse<Report>> {
    const queryParams: Record<string, string> = {};
    if (params?.type) queryParams.type = params.type;
    if (params?.page) queryParams.page = String(params.page);
    if (params?.page_size) queryParams.page_size = String(params.page_size);
    return this.request<PaginatedResponse<Report>>('GET', '/reports', undefined, queryParams);
  }

  async generateReport(type: string, deviceId?: string): Promise<Report> {
    return this.request<Report>('POST', '/reports/generate', { type, device_id: deviceId });
  }

  // ROI
  async getROIStats(): Promise<ROIStats> {
    return this.request<ROIStats>('GET', '/roi/stats');
  }

  // Admin
  async getUsers(page = 1, pageSize = 20): Promise<PaginatedResponse<User>> {
    const response = await this.request<{ data: User[]; total: number; page: number; page_size: number }>(
      'GET',
      '/admin/users',
      undefined,
      { page: String(page), page_size: String(pageSize) }
    );
    // Backend now returns 'data' field directly
    return {
      data: response.data,
      total: response.total,
      page: response.page,
      page_size: response.page_size,
    };
  }

  async createUser(data: UserCreateInput): Promise<User> {
    return this.request<User>('POST', '/admin/users', data);
  }

  async deleteUser(id: number): Promise<MessageResponse> {
    return this.request<MessageResponse>('DELETE', `/admin/users/${id}`);
  }

  async getSystemStatus(): Promise<SystemStatus> {
    return this.request<SystemStatus>('GET', '/system/status');
  }

  // Health
  async healthCheck(): Promise<HealthCheck> {
    return this.request<HealthCheck>('GET', '/health');
  }

  // Export
  async exportReport(
    reportType: 'devices' | 'alerts' | 'roi',
    format: 'pdf' | 'xlsx',
    startDate?: string,
    endDate?: string
  ): Promise<{ data: Blob; filename: string; mimeType: string }> {
    const params: Record<string, string> = { format };
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;

    const url = new URL(`${this.baseUrl}/reports/${reportType}/export`, window.location.origin);
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.set(key, value);
    });

    const headers: Record<string, string> = {};
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    // 导出文件通常耗时较长，使用 120 秒超时
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 120000);

    let response: Response;
    try {
      response = await fetch(url.toString(), { headers, signal: controller.signal });
    } catch (err) {
      if (controller.signal.aborted) {
        throw new TimeoutError('导出请求超时，请缩小时间范围后重试');
      }
      throw err;
    } finally {
      clearTimeout(timeout);
    }

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Export failed');
    }

    const data = await response.blob();
    const filename = response.headers.get('Content-Disposition')?.split('filename="')[1]?.split('"')[0] || `${reportType}_report.${format}`;
    const mimeType = response.headers.get('Content-Type') || 'application/octet-stream';

    return { data, filename, mimeType };
  }
}

export const api = new ApiClient();

export { ApiClient, TimeoutError };

// Export timeout constants for external use
export { DEFAULT_TIMEOUT, AGENT_TIMEOUT };

export default api;