/**
 * API Error Messages - Internationalization Module
 * 
 * This module provides centralized error message handling with
 * internationalization support for API error responses.
 * 
 * Usage:
 *   import { getErrorMessage, ApiErrorCodes } from './errorMessages';
 *   const message = getErrorMessage(error, 'zh');
 */

// API Error Code Definitions
export enum ApiErrorCodes {
  // Authentication Errors (AUTH_*)
  AUTH_INVALID_CREDENTIALS = 'AUTH_INVALID_CREDENTIALS',
  AUTH_TOKEN_EXPIRED = 'AUTH_TOKEN_EXPIRED',
  AUTH_TOKEN_INVALID = 'AUTH_TOKEN_INVALID',
  AUTH_UNAUTHORIZED = 'AUTH_UNAUTHORIZED',
  AUTH_FORBIDDEN = 'AUTH_FORBIDDEN',
  AUTH_SESSION_EXPIRED = 'AUTH_SESSION_EXPIRED',
  AUTH_ACCOUNT_LOCKED = 'AUTH_ACCOUNT_LOCKED',
  AUTH_TOO_MANY_ATTEMPTS = 'AUTH_TOO_MANY_ATTEMPTS',
  AUTH_PASSWORD_MISMATCH = 'AUTH_PASSWORD_MISMATCH',
  AUTH_PASSWORD_TOO_SHORT = 'AUTH_PASSWORD_TOO_SHORT',
  AUTH_INVALID_EMAIL = 'AUTH_INVALID_EMAIL',

  // Device Errors (DEVICE_*)
  DEVICE_NOT_FOUND = 'DEVICE_NOT_FOUND',
  DEVICE_ALREADY_EXISTS = 'DEVICE_ALREADY_EXISTS',
  DEVICE_OFFLINE = 'DEVICE_OFFLINE',
 _DEVICE_CONNECTION_FAILED = 'DEVICE_CONNECTION_FAILED',
  DEVICE_INVALID_DATA = 'DEVICE_INVALID_DATA',
  DEVICE_OPERATION_FAILED = 'DEVICE_OPERATION_FAILED',
  DEVICE_MAINTENANCE_REQUIRED = 'DEVICE_MAINTENANCE_REQUIRED',

  // Telemetry Errors (TELEMETRY_*)
  TELEMETRY_DATA_NOT_FOUND = 'TELEMETRY_DATA_NOT_FOUND',
  TELEMETRY_QUERY_FAILED = 'TELEMETRY_QUERY_FAILED',
  TELEMETRY_INVALID_RANGE = 'TELEMETRY_INVALID_RANGE',
  TELEMETRY_EXPORT_FAILED = 'TELEMETRY_EXPORT_FAILED',
  TELEMETRY_TIMEOUT = 'TELEMETRY_TIMEOUT',

  // Alert Errors (ALERT_*)
  ALERT_NOT_FOUND = 'ALERT_NOT_FOUND',
  ALERT_RULE_NOT_FOUND = 'ALERT_RULE_NOT_FOUND',
  ALERT_RULE_ALREADY_EXISTS = 'ALERT_RULE_ALREADY_EXISTS',
  ALERT_INVALID_THRESHOLD = 'ALERT_INVALID_THRESHOLD',
  ALERT_NOTIFICATION_FAILED = 'ALERT_NOTIFICATION_FAILED',
  ALERT_ACKNOWLEDGE_FAILED = 'ALERT_ACKNOWLEDGE_FAILED',

  // Work Order Errors (WORKORDER_*)
  WORKORDER_NOT_FOUND = 'WORKORDER_NOT_FOUND',
  WORKORDER_INVALID_STATUS = 'WORKORDER_INVALID_STATUS',
  WORKORDER_ASSIGNMENT_FAILED = 'WORKORDER_ASSIGNMENT_FAILED',
  WORKORDER_UPDATE_FAILED = 'WORKORDER_UPDATE_FAILED',
  WORKORDER_ALREADY_EXISTS = 'WORKORDER_ALREADY_EXISTS',

  // Report Errors (REPORT_*)
  REPORT_GENERATION_FAILED = 'REPORT_GENERATION_FAILED',
  REPORT_NOT_FOUND = 'REPORT_NOT_FOUND',
  REPORT_EXPORT_FAILED = 'REPORT_EXPORT_FAILED',
  REPORT_SCHEDULE_FAILED = 'REPORT_SCHEDULE_FAILED',
  REPORT_INVALID_FORMAT = 'REPORT_INVALID_FORMAT',

  // AI Agent Errors (AI_*)
  AI_QUERY_FAILED = 'AI_QUERY_FAILED',
  AI_TIMEOUT = 'AI_TIMEOUT',
  AI_MODEL_UNAVAILABLE = 'AI_MODEL_UNAVAILABLE',
  AI_INVALID_INPUT = 'AI_INVALID_INPUT',
  AI_RESPONSE_PARSE_ERROR = 'AI_RESPONSE_PARSE_ERROR',

  // Black Box Errors (BLACKBOX_*)
  BLACKBOX_NOT_FOUND = 'BLACKBOX_NOT_FOUND',
  BLACKBOX_SNAPSHOT_FAILED = 'BLACKBOX_SNAPSHOT_FAILED',
  BLACKBOX_PLAYBACK_FAILED = 'BLACKBOX_PLAYBACK_FAILED',
  BLACKBOX_EXPORT_FAILED = 'BLACKBOX_EXPORT_FAILED',

  // System Errors (SYSTEM_*)
  SYSTEM_HEALTH_CHECK_FAILED = 'SYSTEM_HEALTH_CHECK_FAILED',
  SYSTEM_MAINTENANCE = 'SYSTEM_MAINTENANCE',
  SYSTEM_OVERLOADED = 'SYSTEM_OVERLOADED',

  // General Errors
  UNKNOWN_ERROR = 'UNKNOWN_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  NETWORK_ERROR = 'NETWORK_ERROR',
  TIMEOUT_ERROR = 'TIMEOUT_ERROR',
  RATE_LIMIT_ERROR = 'RATE_LIMIT_ERROR',
  SERVER_ERROR = 'SERVER_ERROR',
  BAD_REQUEST = 'BAD_REQUEST',
  NOT_FOUND = 'NOT_FOUND',
  CONFLICT = 'CONFLICT',
  SERVICE_UNAVAILABLE = 'SERVICE_UNAVAILABLE',
  GATEWAY_TIMEOUT = 'GATEWAY_TIMEOUT',
  PAYLOAD_TOO_LARGE = 'PAYLOAD_TOO_LARGE',
  UNSUPPORTED_MEDIA_TYPE = 'UNSUPPORTED_MEDIA_TYPE',
}

// Error message translations
const errorMessagesZh: Record<ApiErrorCodes | string, string> = {
  // Authentication Errors
  [ApiErrorCodes.AUTH_INVALID_CREDENTIALS]: '用户名或密码错误',
  [ApiErrorCodes.AUTH_TOKEN_EXPIRED]: '登录令牌已过期，请重新登录',
  [ApiErrorCodes.AUTH_TOKEN_INVALID]: '无效的登录令牌',
  [ApiErrorCodes.AUTH_UNAUTHORIZED]: '未授权访问，请先登录',
  [ApiErrorCodes.AUTH_FORBIDDEN]: '您没有权限执行此操作',
  [ApiErrorCodes.AUTH_SESSION_EXPIRED]: '会话已过期，请重新登录',
  [ApiErrorCodes.AUTH_ACCOUNT_LOCKED]: '账户已被锁定，请联系管理员',
  [ApiErrorCodes.AUTH_TOO_MANY_ATTEMPTS]: '登录尝试次数过多，请稍后再试',
  [ApiErrorCodes.AUTH_PASSWORD_MISMATCH]: '两次输入的密码不一致',
  [ApiErrorCodes.AUTH_PASSWORD_TOO_SHORT]: '密码长度至少需要8位',
  [ApiErrorCodes.AUTH_INVALID_EMAIL]: '邮箱格式无效',

  // Device Errors
  [ApiErrorCodes.DEVICE_NOT_FOUND]: '设备不存在',
  [ApiErrorCodes.DEVICE_ALREADY_EXISTS]: '设备已存在',
  [ApiErrorCodes.DEVICE_OFFLINE]: '设备离线，无法执行操作',
  [ApiErrorCodes._DEVICE_CONNECTION_FAILED]: '设备连接失败',
  [ApiErrorCodes.DEVICE_INVALID_DATA]: '设备数据无效',
  [ApiErrorCodes.DEVICE_OPERATION_FAILED]: '设备操作失败',
  [ApiErrorCodes.DEVICE_MAINTENANCE_REQUIRED]: '设备需要维护',

  // Telemetry Errors
  [ApiErrorCodes.TELEMETRY_DATA_NOT_FOUND]: '遥测数据不存在',
  [ApiErrorCodes.TELEMETRY_QUERY_FAILED]: '遥测数据查询失败',
  [ApiErrorCodes.TELEMETRY_INVALID_RANGE]: '无效的时间范围',
  [ApiErrorCodes.TELEMETRY_EXPORT_FAILED]: '遥测数据导出失败',
  [ApiErrorCodes.TELEMETRY_TIMEOUT]: '遥测数据请求超时',

  // Alert Errors
  [ApiErrorCodes.ALERT_NOT_FOUND]: '告警不存在',
  [ApiErrorCodes.ALERT_RULE_NOT_FOUND]: '告警规则不存在',
  [ApiErrorCodes.ALERT_RULE_ALREADY_EXISTS]: '告警规则已存在',
  [ApiErrorCodes.ALERT_INVALID_THRESHOLD]: '无效的告警阈值',
  [ApiErrorCodes.ALERT_NOTIFICATION_FAILED]: '告警通知发送失败',
  [ApiErrorCodes.ALERT_ACKNOWLEDGE_FAILED]: '告警确认失败',

  // Work Order Errors
  [ApiErrorCodes.WORKORDER_NOT_FOUND]: '工单不存在',
  [ApiErrorCodes.WORKORDER_INVALID_STATUS]: '无效的工单状态',
  [ApiErrorCodes.WORKORDER_ASSIGNMENT_FAILED]: '工单分配失败',
  [ApiErrorCodes.WORKORDER_UPDATE_FAILED]: '工单更新失败',
  [ApiErrorCodes.WORKORDER_ALREADY_EXISTS]: '工单已存在',

  // Report Errors
  [ApiErrorCodes.REPORT_GENERATION_FAILED]: '报告生成失败',
  [ApiErrorCodes.REPORT_NOT_FOUND]: '报告不存在',
  [ApiErrorCodes.REPORT_EXPORT_FAILED]: '报告导出失败',
  [ApiErrorCodes.REPORT_SCHEDULE_FAILED]: '报告定时设置失败',
  [ApiErrorCodes.REPORT_INVALID_FORMAT]: '无效的报告格式',

  // AI Agent Errors
  [ApiErrorCodes.AI_QUERY_FAILED]: 'AI查询失败，请稍后重试',
  [ApiErrorCodes.AI_TIMEOUT]: 'AI响应超时',
  [ApiErrorCodes.AI_MODEL_UNAVAILABLE]: 'AI模型暂时不可用',
  [ApiErrorCodes.AI_INVALID_INPUT]: '无效的输入内容',
  [ApiErrorCodes.AI_RESPONSE_PARSE_ERROR]: 'AI响应解析失败',

  // Black Box Errors
  [ApiErrorCodes.BLACKBOX_NOT_FOUND]: '黑匣子记录不存在',
  [ApiErrorCodes.BLACKBOX_SNAPSHOT_FAILED]: '数据快照获取失败',
  [ApiErrorCodes.BLACKBOX_PLAYBACK_FAILED]: '数据回放失败',
  [ApiErrorCodes.BLACKBOX_EXPORT_FAILED]: '黑匣子数据导出失败',

  // System Errors
  [ApiErrorCodes.SYSTEM_HEALTH_CHECK_FAILED]: '系统健康检查失败',
  [ApiErrorCodes.SYSTEM_MAINTENANCE]: '系统正在维护中',
  [ApiErrorCodes.SYSTEM_OVERLOADED]: '系统负载过高，请稍后重试',

  // General Errors
  [ApiErrorCodes.UNKNOWN_ERROR]: '未知错误，请稍后重试',
  [ApiErrorCodes.VALIDATION_ERROR]: '数据验证失败',
  [ApiErrorCodes.NETWORK_ERROR]: '网络连接失败',
  [ApiErrorCodes.TIMEOUT_ERROR]: '请求超时',
  [ApiErrorCodes.RATE_LIMIT_ERROR]: '请求过于频繁，请稍后再试',
  [ApiErrorCodes.SERVER_ERROR]: '服务器内部错误',
  [ApiErrorCodes.BAD_REQUEST]: '请求参数错误',
  [ApiErrorCodes.NOT_FOUND]: '请求的资源不存在',
  [ApiErrorCodes.CONFLICT]: '资源冲突',
  [ApiErrorCodes.SERVICE_UNAVAILABLE]: '服务暂时不可用',
  [ApiErrorCodes.GATEWAY_TIMEOUT]: '网关超时',
  [ApiErrorCodes.PAYLOAD_TOO_LARGE]: '请求数据过大',
  [ApiErrorCodes.UNSUPPORTED_MEDIA_TYPE]: '不支持的媒体类型',
};

const errorMessagesEn: Record<ApiErrorCodes | string, string> = {
  // Authentication Errors
  [ApiErrorCodes.AUTH_INVALID_CREDENTIALS]: 'Invalid username or password',
  [ApiErrorCodes.AUTH_TOKEN_EXPIRED]: 'Token expired, please login again',
  [ApiErrorCodes.AUTH_TOKEN_INVALID]: 'Invalid token',
  [ApiErrorCodes.AUTH_UNAUTHORIZED]: 'Unauthorized access, please login first',
  [ApiErrorCodes.AUTH_FORBIDDEN]: 'You do not have permission to perform this action',
  [ApiErrorCodes.AUTH_SESSION_EXPIRED]: 'Session expired, please login again',
  [ApiErrorCodes.AUTH_ACCOUNT_LOCKED]: 'Account locked, please contact administrator',
  [ApiErrorCodes.AUTH_TOO_MANY_ATTEMPTS]: 'Too many attempts, please try again later',
  [ApiErrorCodes.AUTH_PASSWORD_MISMATCH]: 'Passwords do not match',
  [ApiErrorCodes.AUTH_PASSWORD_TOO_SHORT]: 'Password must be at least 8 characters',
  [ApiErrorCodes.AUTH_INVALID_EMAIL]: 'Invalid email format',

  // Device Errors
  [ApiErrorCodes.DEVICE_NOT_FOUND]: 'Device not found',
  [ApiErrorCodes.DEVICE_ALREADY_EXISTS]: 'Device already exists',
  [ApiErrorCodes.DEVICE_OFFLINE]: 'Device is offline, cannot perform operation',
  [ApiErrorCodes._DEVICE_CONNECTION_FAILED]: 'Device connection failed',
  [ApiErrorCodes.DEVICE_INVALID_DATA]: 'Invalid device data',
  [ApiErrorCodes.DEVICE_OPERATION_FAILED]: 'Device operation failed',
  [ApiErrorCodes.DEVICE_MAINTENANCE_REQUIRED]: 'Device requires maintenance',

  // Telemetry Errors
  [ApiErrorCodes.TELEMETRY_DATA_NOT_FOUND]: 'Telemetry data not found',
  [ApiErrorCodes.TELEMETRY_QUERY_FAILED]: 'Telemetry query failed',
  [ApiErrorCodes.TELEMETRY_INVALID_RANGE]: 'Invalid time range',
  [ApiErrorCodes.TELEMETRY_EXPORT_FAILED]: 'Telemetry data export failed',
  [ApiErrorCodes.TELEMETRY_TIMEOUT]: 'Telemetry request timed out',

  // Alert Errors
  [ApiErrorCodes.ALERT_NOT_FOUND]: 'Alert not found',
  [ApiErrorCodes.ALERT_RULE_NOT_FOUND]: 'Alert rule not found',
  [ApiErrorCodes.ALERT_RULE_ALREADY_EXISTS]: 'Alert rule already exists',
  [ApiErrorCodes.ALERT_INVALID_THRESHOLD]: 'Invalid alert threshold',
  [ApiErrorCodes.ALERT_NOTIFICATION_FAILED]: 'Alert notification failed',
  [ApiErrorCodes.ALERT_ACKNOWLEDGE_FAILED]: 'Alert acknowledgement failed',

  // Work Order Errors
  [ApiErrorCodes.WORKORDER_NOT_FOUND]: 'Work order not found',
  [ApiErrorCodes.WORKORDER_INVALID_STATUS]: 'Invalid work order status',
  [ApiErrorCodes.WORKORDER_ASSIGNMENT_FAILED]: 'Work order assignment failed',
  [ApiErrorCodes.WORKORDER_UPDATE_FAILED]: 'Work order update failed',
  [ApiErrorCodes.WORKORDER_ALREADY_EXISTS]: 'Work order already exists',

  // Report Errors
  [ApiErrorCodes.REPORT_GENERATION_FAILED]: 'Report generation failed',
  [ApiErrorCodes.REPORT_NOT_FOUND]: 'Report not found',
  [ApiErrorCodes.REPORT_EXPORT_FAILED]: 'Report export failed',
  [ApiErrorCodes.REPORT_SCHEDULE_FAILED]: 'Report scheduling failed',
  [ApiErrorCodes.REPORT_INVALID_FORMAT]: 'Invalid report format',

  // AI Agent Errors
  [ApiErrorCodes.AI_QUERY_FAILED]: 'AI query failed, please try again later',
  [ApiErrorCodes.AI_TIMEOUT]: 'AI response timed out',
  [ApiErrorCodes.AI_MODEL_UNAVAILABLE]: 'AI model temporarily unavailable',
  [ApiErrorCodes.AI_INVALID_INPUT]: 'Invalid input',
  [ApiErrorCodes.AI_RESPONSE_PARSE_ERROR]: 'Failed to parse AI response',

  // Black Box Errors
  [ApiErrorCodes.BLACKBOX_NOT_FOUND]: 'Black box record not found',
  [ApiErrorCodes.BLACKBOX_SNAPSHOT_FAILED]: 'Failed to get data snapshot',
  [ApiErrorCodes.BLACKBOX_PLAYBACK_FAILED]: 'Playback failed',
  [ApiErrorCodes.BLACKBOX_EXPORT_FAILED]: 'Black box data export failed',

  // System Errors
  [ApiErrorCodes.SYSTEM_HEALTH_CHECK_FAILED]: 'System health check failed',
  [ApiErrorCodes.SYSTEM_MAINTENANCE]: 'System is under maintenance',
  [ApiErrorCodes.SYSTEM_OVERLOADED]: 'System overloaded, please try again later',

  // General Errors
  [ApiErrorCodes.UNKNOWN_ERROR]: 'Unknown error, please try again later',
  [ApiErrorCodes.VALIDATION_ERROR]: 'Data validation failed',
  [ApiErrorCodes.NETWORK_ERROR]: 'Network connection failed',
  [ApiErrorCodes.TIMEOUT_ERROR]: 'Request timed out',
  [ApiErrorCodes.RATE_LIMIT_ERROR]: 'Too many requests, please try again later',
  [ApiErrorCodes.SERVER_ERROR]: 'Internal server error',
  [ApiErrorCodes.BAD_REQUEST]: 'Bad request parameters',
  [ApiErrorCodes.NOT_FOUND]: 'Requested resource not found',
  [ApiErrorCodes.CONFLICT]: 'Resource conflict',
  [ApiErrorCodes.SERVICE_UNAVAILABLE]: 'Service temporarily unavailable',
  [ApiErrorCodes.GATEWAY_TIMEOUT]: 'Gateway timeout',
  [ApiErrorCodes.PAYLOAD_TOO_LARGE]: 'Request payload too large',
  [ApiErrorCodes.UNSUPPORTED_MEDIA_TYPE]: 'Unsupported media type',
};

// HTTP Status to Error Code Mapping
export const httpStatusToErrorCode: Record<number, ApiErrorCodes> = {
  400: ApiErrorCodes.BAD_REQUEST,
  401: ApiErrorCodes.AUTH_UNAUTHORIZED,
  403: ApiErrorCodes.AUTH_FORBIDDEN,
  404: ApiErrorCodes.NOT_FOUND,
  409: ApiErrorCodes.CONFLICT,
  413: ApiErrorCodes.PAYLOAD_TOO_LARGE,
  415: ApiErrorCodes.UNSUPPORTED_MEDIA_TYPE,
  429: ApiErrorCodes.RATE_LIMIT_ERROR,
  500: ApiErrorCodes.SERVER_ERROR,
  502: ApiErrorCodes.SERVICE_UNAVAILABLE,
  503: ApiErrorCodes.SERVICE_UNAVAILABLE,
  504: ApiErrorCodes.GATEWAY_TIMEOUT,
};

type Language = 'zh' | 'en';

/**
 * Get localized error message by error code
 * 
 * @param errorCode - API error code or HTTP status code
 * @param language - Target language ('zh' or 'en')
 * @returns Localized error message
 */
export function getErrorMessage(
  errorCode: ApiErrorCodes | string | number,
  language: Language = 'zh'
): string {
  const messages = language === 'zh' ? errorMessagesZh : errorMessagesEn;
  
  // Handle HTTP status codes
  if (typeof errorCode === 'number') {
    const mappedCode = httpStatusToErrorCode[errorCode];
    if (mappedCode) {
      return messages[mappedCode] || messages[ApiErrorCodes.UNKNOWN_ERROR];
    }
    return messages[ApiErrorCodes.UNKNOWN_ERROR];
  }
  
  // Return localized message or fallback to unknown error
  return messages[errorCode] || messages[ApiErrorCodes.UNKNOWN_ERROR];
}

/**
 * Get error message from API response
 * 
 * @param error - API error object
 * @param language - Target language
 * @returns Localized error message
 */
export function getApiErrorMessage(
  error: {
    code?: string | number;
    message?: string;
    status?: number;
    response?: { status?: number; data?: { code?: string; message?: string } };
  },
  language: Language = 'zh'
): string {
  // Priority: Custom error code from response
  if (error.response?.data?.code) {
    return getErrorMessage(error.response.data.code, language);
  }
  
  // Priority: Custom error code from error object
  if (error.code) {
    return getErrorMessage(error.code, language);
  }
  
  // Priority: HTTP status code
  if (error.response?.status) {
    return getErrorMessage(error.response.status, language);
  }
  
  if (error.status) {
    return getErrorMessage(error.status, language);
  }
  
  // Fallback: Use original message if available
  if (error.message) {
    // Check if message is a known error code
    const knownMessage = getErrorMessage(error.message, language);
    if (knownMessage !== getErrorMessage(ApiErrorCodes.UNKNOWN_ERROR, language)) {
      return knownMessage;
    }
    return error.message;
  }
  
  return getErrorMessage(ApiErrorCodes.UNKNOWN_ERROR, language);
}

/**
 * Error message helper class for React components
 */
export class ErrorMessageHelper {
  private language: Language;
  
  constructor(language: Language = 'zh') {
    this.language = language;
  }
  
  setLanguage(lang: Language): void {
    this.language = lang;
  }
  
  get(code: ApiErrorCodes | string | number): string {
    return getErrorMessage(code, this.language);
  }
  
  fromApiError(error: unknown): string {
    if (error instanceof Error) {
      return getApiErrorMessage(
        { message: error.message },
        this.language
      );
    }
    
    if (typeof error === 'object' && error !== null) {
      return getApiErrorMessage(error as Record<string, unknown>, this.language);
    }
    
    return getErrorMessage(ApiErrorCodes.UNKNOWN_ERROR, this.language);
  }
}

// Default helper instance
export const defaultErrorHelper = new ErrorMessageHelper();

export default {
  ApiErrorCodes,
  getErrorMessage,
  getApiErrorMessage,
  ErrorMessageHelper,
  defaultErrorHelper,
  httpStatusToErrorCode,
};