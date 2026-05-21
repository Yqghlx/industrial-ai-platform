/**
 * Type Guards and API Response Utilities
 * FIX-013: 消除类型断言，使用类型守卫进行安全的类型检查
 */

/**
 * Generic API Response interface
 */
export interface ApiResponse<T> {
  data: T;
  success?: boolean;
  total?: number;
  message?: string;
}

/**
 * Type guard to check if an object is an ApiResponse
 */
export function isApiResponse<T>(
  obj: unknown,
  dataValidator?: (data: unknown) => boolean
): obj is ApiResponse<T> {
  if (!obj || typeof obj !== 'object') return false;
  
  const response = obj as Record<string, unknown>;
  
  // Check required 'data' field
  if (!('data' in response)) return false;
  
  // If a data validator is provided, use it
  if (dataValidator && !dataValidator(response.data)) return false;
  
  return true;
}

/**
 * Type guard for array data in API response
 */
export function isApiResponseArray<T>(
  obj: unknown,
  itemValidator?: (item: unknown) => boolean
): obj is ApiResponse<T[]> {
  return isApiResponse(obj, (data) => {
    if (!Array.isArray(data)) return false;
    if (itemValidator) {
      return data.every(itemValidator);
    }
    return true;
  });
}

/**
 * Validator for Device objects
 */
export function isDevice(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const device = obj as Record<string, unknown>;
  return (
    typeof device.id === 'string' &&
    typeof device.name === 'string' &&
    typeof device.type === 'string' &&
    typeof device.status === 'string' &&
    typeof device.location === 'string'
  );
}

/**
 * Validator for AlertRule objects
 */
export function isAlertRule(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const rule = obj as Record<string, unknown>;
  return (
    typeof rule.id === 'number' &&
    typeof rule.name === 'string' &&
    typeof rule.metric === 'string'
  );
}

/**
 * Validator for WorkOrder objects
 */
export function isWorkOrder(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const order = obj as Record<string, unknown>;
  return (
    typeof order.id === 'number' &&
    typeof order.title === 'string'
  );
}

/**
 * Validator for Notification objects
 */
export function isNotification(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const notification = obj as Record<string, unknown>;
  return (
    typeof notification.id === 'number' &&
    typeof notification.type === 'string' &&
    typeof notification.title === 'string'
  );
}

/**
 * Validator for Report objects
 */
export function isReport(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const report = obj as Record<string, unknown>;
  return (
    typeof report.id === 'number' &&
    typeof report.title === 'string' &&
    typeof report.type === 'string'
  );
}

/**
 * Validator for BlackBoxRecord objects
 */
export function isBlackBoxRecord(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const record = obj as Record<string, unknown>;
  return (
    typeof record.id === 'number' &&
    typeof record.device_id === 'string' &&
    typeof record.trigger_type === 'string'
  );
}

/**
 * Validator for Telemetry objects
 */
export function isTelemetry(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const telemetry = obj as Record<string, unknown>;
  return (
    typeof telemetry.device_id === 'string' &&
    (typeof telemetry.temperature === 'number' || telemetry.temperature === undefined) &&
    (typeof telemetry.vibration === 'number' || telemetry.vibration === undefined)
  );
}

/**
 * Validator for DeviceStats objects
 */
export function isDeviceStats(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const stats = obj as Record<string, unknown>;
  return (
    typeof stats.avg_temperature === 'number' ||
    typeof stats.avg_vibration === 'number' ||
    typeof stats.data_points === 'number'
  );
}

/**
 * Validator for ROIStats objects
 */
export function isROIStats(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const stats = obj as Record<string, unknown>;
  return (
    typeof stats.total_devices === 'number' &&
    typeof stats.active_alerts === 'number'
  );
}

/**
 * Validator for DeviceGraph objects
 */
export function isDeviceGraph(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const graph = obj as Record<string, unknown>;
  return (
    Array.isArray(graph.nodes) &&
    Array.isArray(graph.links)
  );
}

/**
 * Validator for SystemStatus objects
 */
export function isSystemStatus(obj: unknown): boolean {
  if (!obj || typeof obj !== 'object') return false;
  const status = obj as Record<string, unknown>;
  return (
    typeof status.database === 'string' &&
    typeof status.version === 'string'
  );
}

/**
 * Safely extract data from API response with validation
 * Returns null if validation fails
 */
export function safeExtractData<T>(
  response: unknown,
  validator: (data: unknown) => boolean
): T | null {
  if (isApiResponse(response, validator)) {
    return response.data as T;
  }
  return null;
}

/**
 * Safely extract array data from API response
 */
export function safeExtractArray<T>(
  response: unknown,
  itemValidator?: (item: unknown) => boolean
): T[] {
  if (isApiResponseArray(response, itemValidator)) {
    return response.data as T[];
  }
  return [];
}