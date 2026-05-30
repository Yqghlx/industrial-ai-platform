import { describe, it, expect } from 'vitest';
import {
  ApiErrorCodes,
  getErrorMessage,
  getApiErrorMessage,
  ErrorMessageHelper,
  defaultErrorHelper,
  httpStatusToErrorCode,
} from './errorMessages';

describe('ApiErrorCodes', () => {
  it('包含所有认证错误码', () => {
    expect(ApiErrorCodes.AUTH_INVALID_CREDENTIALS).toBe('AUTH_INVALID_CREDENTIALS');
    expect(ApiErrorCodes.AUTH_TOKEN_EXPIRED).toBe('AUTH_TOKEN_EXPIRED');
    expect(ApiErrorCodes.AUTH_TOKEN_INVALID).toBe('AUTH_TOKEN_INVALID');
    expect(ApiErrorCodes.AUTH_UNAUTHORIZED).toBe('AUTH_UNAUTHORIZED');
    expect(ApiErrorCodes.AUTH_FORBIDDEN).toBe('AUTH_FORBIDDEN');
    expect(ApiErrorCodes.AUTH_SESSION_EXPIRED).toBe('AUTH_SESSION_EXPIRED');
    expect(ApiErrorCodes.AUTH_ACCOUNT_LOCKED).toBe('AUTH_ACCOUNT_LOCKED');
    expect(ApiErrorCodes.AUTH_TOO_MANY_ATTEMPTS).toBe('AUTH_TOO_MANY_ATTEMPTS');
    expect(ApiErrorCodes.AUTH_PASSWORD_MISMATCH).toBe('AUTH_PASSWORD_MISMATCH');
    expect(ApiErrorCodes.AUTH_PASSWORD_TOO_SHORT).toBe('AUTH_PASSWORD_TOO_SHORT');
    expect(ApiErrorCodes.AUTH_INVALID_EMAIL).toBe('AUTH_INVALID_EMAIL');
  });

  it('包含所有设备错误码', () => {
    expect(ApiErrorCodes.DEVICE_NOT_FOUND).toBe('DEVICE_NOT_FOUND');
    expect(ApiErrorCodes.DEVICE_ALREADY_EXISTS).toBe('DEVICE_ALREADY_EXISTS');
    expect(ApiErrorCodes.DEVICE_OFFLINE).toBe('DEVICE_OFFLINE');
    expect(ApiErrorCodes.DEVICE_INVALID_DATA).toBe('DEVICE_INVALID_DATA');
    expect(ApiErrorCodes.DEVICE_OPERATION_FAILED).toBe('DEVICE_OPERATION_FAILED');
    expect(ApiErrorCodes.DEVICE_MAINTENANCE_REQUIRED).toBe('DEVICE_MAINTENANCE_REQUIRED');
  });

  it('包含所有告警错误码', () => {
    expect(ApiErrorCodes.ALERT_NOT_FOUND).toBe('ALERT_NOT_FOUND');
    expect(ApiErrorCodes.ALERT_RULE_NOT_FOUND).toBe('ALERT_RULE_NOT_FOUND');
    expect(ApiErrorCodes.ALERT_RULE_ALREADY_EXISTS).toBe('ALERT_RULE_ALREADY_EXISTS');
    expect(ApiErrorCodes.ALERT_INVALID_THRESHOLD).toBe('ALERT_INVALID_THRESHOLD');
    expect(ApiErrorCodes.ALERT_NOTIFICATION_FAILED).toBe('ALERT_NOTIFICATION_FAILED');
    expect(ApiErrorCodes.ALERT_ACKNOWLEDGE_FAILED).toBe('ALERT_ACKNOWLEDGE_FAILED');
  });

  it('包含所有通用错误码', () => {
    expect(ApiErrorCodes.UNKNOWN_ERROR).toBe('UNKNOWN_ERROR');
    expect(ApiErrorCodes.VALIDATION_ERROR).toBe('VALIDATION_ERROR');
    expect(ApiErrorCodes.NETWORK_ERROR).toBe('NETWORK_ERROR');
    expect(ApiErrorCodes.TIMEOUT_ERROR).toBe('TIMEOUT_ERROR');
    expect(ApiErrorCodes.RATE_LIMIT_ERROR).toBe('RATE_LIMIT_ERROR');
    expect(ApiErrorCodes.SERVER_ERROR).toBe('SERVER_ERROR');
    expect(ApiErrorCodes.BAD_REQUEST).toBe('BAD_REQUEST');
    expect(ApiErrorCodes.NOT_FOUND).toBe('NOT_FOUND');
    expect(ApiErrorCodes.CONFLICT).toBe('CONFLICT');
    expect(ApiErrorCodes.SERVICE_UNAVAILABLE).toBe('SERVICE_UNAVAILABLE');
    expect(ApiErrorCodes.GATEWAY_TIMEOUT).toBe('GATEWAY_TIMEOUT');
    expect(ApiErrorCodes.PAYLOAD_TOO_LARGE).toBe('PAYLOAD_TOO_LARGE');
    expect(ApiErrorCodes.UNSUPPORTED_MEDIA_TYPE).toBe('UNSUPPORTED_MEDIA_TYPE');
  });
});

describe('httpStatusToErrorCode', () => {
  it('映射常见 HTTP 状态码到错误码', () => {
    expect(httpStatusToErrorCode[400]).toBe(ApiErrorCodes.BAD_REQUEST);
    expect(httpStatusToErrorCode[401]).toBe(ApiErrorCodes.AUTH_UNAUTHORIZED);
    expect(httpStatusToErrorCode[403]).toBe(ApiErrorCodes.AUTH_FORBIDDEN);
    expect(httpStatusToErrorCode[404]).toBe(ApiErrorCodes.NOT_FOUND);
    expect(httpStatusToErrorCode[409]).toBe(ApiErrorCodes.CONFLICT);
    expect(httpStatusToErrorCode[429]).toBe(ApiErrorCodes.RATE_LIMIT_ERROR);
    expect(httpStatusToErrorCode[500]).toBe(ApiErrorCodes.SERVER_ERROR);
    expect(httpStatusToErrorCode[502]).toBe(ApiErrorCodes.SERVICE_UNAVAILABLE);
    expect(httpStatusToErrorCode[503]).toBe(ApiErrorCodes.SERVICE_UNAVAILABLE);
    expect(httpStatusToErrorCode[504]).toBe(ApiErrorCodes.GATEWAY_TIMEOUT);
  });

  it('502 和 503 都映射到 SERVICE_UNAVAILABLE', () => {
    expect(httpStatusToErrorCode[502]).toBe(httpStatusToErrorCode[503]);
  });
});

describe('getErrorMessage', () => {
  it('返回中文错误消息（默认语言）', () => {
    expect(getErrorMessage(ApiErrorCodes.AUTH_INVALID_CREDENTIALS)).toBe('用户名或密码错误');
    expect(getErrorMessage(ApiErrorCodes.DEVICE_NOT_FOUND)).toBe('设备不存在');
    expect(getErrorMessage(ApiErrorCodes.NETWORK_ERROR)).toBe('网络连接失败');
    expect(getErrorMessage(ApiErrorCodes.SERVER_ERROR)).toBe('服务器内部错误');
  });

  it('返回英文错误消息', () => {
    expect(getErrorMessage(ApiErrorCodes.AUTH_INVALID_CREDENTIALS, 'en')).toBe('Invalid username or password');
    expect(getErrorMessage(ApiErrorCodes.DEVICE_NOT_FOUND, 'en')).toBe('Device not found');
    expect(getErrorMessage(ApiErrorCodes.NETWORK_ERROR, 'en')).toBe('Network connection failed');
    expect(getErrorMessage(ApiErrorCodes.SERVER_ERROR, 'en')).toBe('Internal server error');
  });

  it('通过数字 HTTP 状态码返回对应消息', () => {
    expect(getErrorMessage(401)).toBe('未授权访问，请先登录');
    expect(getErrorMessage(401, 'en')).toBe('Unauthorized access, please login first');
    expect(getErrorMessage(404)).toBe('请求的资源不存在');
    expect(getErrorMessage(500)).toBe('服务器内部错误');
  });

  it('未知 HTTP 状态码返回 UNKNOWN_ERROR 消息', () => {
    expect(getErrorMessage(418)).toBe('未知错误，请稍后重试');
    expect(getErrorMessage(418, 'en')).toBe('Unknown error, please try again later');
  });

  it('未知字符串错误码返回 UNKNOWN_ERROR 消息', () => {
    expect(getErrorMessage('SOME_UNKNOWN_CODE')).toBe('未知错误，请稍后重试');
    // 不会匹配未知码，回退到 unknown
  });

  it('覆盖所有错误分类的中英文消息', () => {
    // 遥测
    expect(getErrorMessage(ApiErrorCodes.TELEMETRY_TIMEOUT)).toBe('遥测数据请求超时');
    expect(getErrorMessage(ApiErrorCodes.TELEMETRY_TIMEOUT, 'en')).toBe('Telemetry request timed out');
    // 工单
    expect(getErrorMessage(ApiErrorCodes.WORKORDER_NOT_FOUND)).toBe('工单不存在');
    expect(getErrorMessage(ApiErrorCodes.WORKORDER_NOT_FOUND, 'en')).toBe('Work order not found');
    // 报告
    expect(getErrorMessage(ApiErrorCodes.REPORT_NOT_FOUND)).toBe('报告不存在');
    expect(getErrorMessage(ApiErrorCodes.REPORT_NOT_FOUND, 'en')).toBe('Report not found');
    // AI
    expect(getErrorMessage(ApiErrorCodes.AI_QUERY_FAILED)).toBe('AI查询失败，请稍后重试');
    expect(getErrorMessage(ApiErrorCodes.AI_QUERY_FAILED, 'en')).toBe('AI query failed, please try again later');
    // 黑匣子
    expect(getErrorMessage(ApiErrorCodes.BLACKBOX_NOT_FOUND)).toBe('黑匣子记录不存在');
    // 系统
    expect(getErrorMessage(ApiErrorCodes.SYSTEM_OVERLOADED)).toBe('系统负载过高，请稍后重试');
  });
});

describe('getApiErrorMessage', () => {
  it('从 response.data.code 获取错误消息', () => {
    const result = getApiErrorMessage({
      response: { data: { code: 'DEVICE_NOT_FOUND' } },
    });
    expect(result).toBe('设备不存在');
  });

  it('从 error.code 获取错误消息', () => {
    const result = getApiErrorMessage({
      code: 'AUTH_TOKEN_EXPIRED',
    });
    expect(result).toBe('登录令牌已过期，请重新登录');
  });

  it('从 response.status 获取错误消息', () => {
    const result = getApiErrorMessage({
      response: { status: 403 },
    });
    expect(result).toBe('您没有权限执行此操作');
  });

  it('从 error.status 获取错误消息', () => {
    const result = getApiErrorMessage({
      status: 404,
    });
    expect(result).toBe('请求的资源不存在');
  });

  it('从 error.message 获取已知错误码的消息', () => {
    const result = getApiErrorMessage({
      message: 'NETWORK_ERROR',
    });
    expect(result).toBe('网络连接失败');
  });

  it('从 error.message 返回原始消息（非已知码）', () => {
    const result = getApiErrorMessage({
      message: 'Custom error message',
    });
    expect(result).toBe('Custom error message');
  });

  it('无任何信息时返回 UNKNOWN_ERROR', () => {
    const result = getApiErrorMessage({});
    expect(result).toBe('未知错误，请稍后重试');
  });

  it('优先级：response.data.code > error.code > response.status > error.status', () => {
    const result = getApiErrorMessage({
      code: 'AUTH_TOKEN_EXPIRED',
      response: { data: { code: 'DEVICE_NOT_FOUND' } },
    });
    // response.data.code 优先
    expect(result).toBe('设备不存在');
  });

  it('支持英文语言', () => {
    const result = getApiErrorMessage(
      { code: 'AUTH_TOKEN_EXPIRED' },
      'en'
    );
    expect(result).toBe('Token expired, please login again');
  });
});

describe('ErrorMessageHelper', () => {
  it('构造函数默认中文', () => {
    const helper = new ErrorMessageHelper();
    expect(helper.get(ApiErrorCodes.DEVICE_NOT_FOUND)).toBe('设备不存在');
  });

  it('setLanguage 切换语言', () => {
    const helper = new ErrorMessageHelper();
    helper.setLanguage('en');
    expect(helper.get(ApiErrorCodes.DEVICE_NOT_FOUND)).toBe('Device not found');
  });

  it('get 方法支持数字状态码', () => {
    const helper = new ErrorMessageHelper();
    expect(helper.get(401)).toBe('未授权访问，请先登录');
  });

  it('fromApiError 处理 Error 实例', () => {
    const helper = new ErrorMessageHelper();
    const result = helper.fromApiError(new Error('NETWORK_ERROR'));
    expect(result).toBe('网络连接失败');
  });

  it('fromApiError 处理普通对象', () => {
    const helper = new ErrorMessageHelper();
    const result = helper.fromApiError({ code: 'DEVICE_OFFLINE' });
    expect(result).toBe('设备离线，无法执行操作');
  });

  it('fromApiError 处理 null 和 undefined', () => {
    const helper = new ErrorMessageHelper();
    expect(helper.fromApiError(null)).toBe('未知错误，请稍后重试');
    expect(helper.fromApiError(undefined)).toBe('未知错误，请稍后重试');
  });
});

describe('defaultErrorHelper', () => {
  it('是 ErrorMessageHelper 的实例', () => {
    expect(defaultErrorHelper).toBeInstanceOf(ErrorMessageHelper);
  });

  it('默认使用中文', () => {
    expect(defaultErrorHelper.get(ApiErrorCodes.SERVER_ERROR)).toBe('服务器内部错误');
  });
});
