/**
 * ErrorHelper - 统一错误处理工具
 * 
 * 提供用户友好的错误提示，替代 "Unknown error" 等模糊信息
 */

import { useI18n } from '../i18n';

// 错误类型枚举
export enum ErrorType {
  NETWORK = 'NETWORK',
  TIMEOUT = 'TIMEOUT',
  UNAUTHORIZED = 'UNAUTHORIZED', // FIX-004: 修正错误的枚举值
  FORBIDDEN = 'FORBIDDEN',
  NOT_FOUND = 'NOT_FOUND',
  VALIDATION = 'VALIDATION',
  RATE_LIMIT = 'RATE_LIMIT',
  SERVER = 'SERVER',
  UNKNOWN = 'UNKNOWN',
}

/**
 * 解析 API 错误并返回用户友好的消息
 */
export function parseApiError(error: unknown, t?: (key: string) => string): string {
  // 网络错误
  if (error instanceof TypeError && error.message === 'Failed to fetch') {
    return t ? t('errors.networkError') : '网络连接失败，请检查网络设置';
  }

  // 超时错误
  if (error instanceof Error && error.name === 'AbortError') {
    return t ? t('errors.timeout') : '请求超时，请稍后重试';
  }

  // Error 实例
  if (error instanceof Error) {
    const message = error.message.toLowerCase();

    // 认证错误
    if (message.includes('unauthorized') || error.message === 'Unauthorized') {
      return t ? t('errors.unauthorized') : '登录已过期，请重新登录';
    }

    // 禁止访问
    if (message.includes('forbidden')) {
      return t ? t('errors.forbidden') : '您没有权限执行此操作';
    }

    // 资源未找到
    if (message.includes('not found') || message.includes('404')) {
      return t ? t('errors.notFound') : '请求的资源不存在';
    }

    // 速率限制
    if (message.includes('rate limit') || message.includes('429')) {
      return t ? t('errors.rateLimit') : '请求过于频繁，请稍后再试';
    }

    // 验证错误
    if (message.includes('validation') || message.includes('invalid')) {
      return error.message; // 保留原始验证消息
    }

    // 服务器错误
    if (message.includes('500') || message.includes('internal server error')) {
      return t ? t('errors.serverError') : '服务器错误，请稍后重试';
    }

    // 其他已知错误，返回原始消息
    if (error.message && error.message !== 'Unknown error') {
      return error.message;
    }
  }

  // 默认错误消息
  return t ? t('errors.unknown') : '操作失败，请稍后重试';
}

/**
 * 根据错误类型获取错误类型枚举
 */
export function getErrorType(error: unknown): ErrorType {
  if (error instanceof TypeError && error.message === 'Failed to fetch') {
    return ErrorType.NETWORK;
  }

  if (error instanceof Error) {
    const message = error.message.toLowerCase();

    if (error.name === 'AbortError') {
      return ErrorType.TIMEOUT;
    }
    if (message.includes('unauthorized') || error.message === 'Unauthorized') {
      return ErrorType.UNAUTHORIZED;
    }
    if (message.includes('forbidden')) {
      return ErrorType.FORBIDDEN;
    }
    if (message.includes('not found') || message.includes('404')) {
      return ErrorType.NOT_FOUND;
    }
    if (message.includes('rate limit') || message.includes('429')) {
      return ErrorType.RATE_LIMIT;
    }
    if (message.includes('500') || message.includes('internal server error')) {
      return ErrorType.SERVER;
    }
    if (message.includes('validation') || message.includes('invalid')) {
      return ErrorType.VALIDATION;
    }
  }

  return ErrorType.UNKNOWN;
}

/**
 * 检查是否为可重试的错误
 */
export function isRetryableError(error: unknown): boolean {
  const type = getErrorType(error);
  return type === ErrorType.NETWORK || 
         type === ErrorType.TIMEOUT || 
         type === ErrorType.RATE_LIMIT ||
         type === ErrorType.SERVER;
}

/**
 * 自定义错误类
 */
export class AppError extends Error {
  type: ErrorType;
  originalError?: unknown;

  constructor(message: string, type: ErrorType = ErrorType.UNKNOWN, originalError?: unknown) {
    super(message);
    this.name = 'AppError';
    this.type = type;
    this.originalError = originalError;
  }
}

/**
 * 超时错误类
 */
export class TimeoutError extends Error {
  constructor() {
    super('请求超时');
    this.name = 'TimeoutError';
  }
}

/**
 * 网络错误类
 */
export class NetworkError extends Error {
  constructor() {
    super('网络连接失败');
    this.name = 'NetworkError';
  }
}

/**
 * 创建用户友好的错误对象
 */
export function createFriendlyError(error: unknown, t?: (key: string) => string): AppError {
  const type = getErrorType(error);
  const message = parseApiError(error, t);
  return new AppError(message, type, error);
}

/**
 * 错误处理 Hook（用于组件中）
 */
export function useErrorHandler() {
  const { t } = useI18n();

  const handleError = (error: unknown, fallbackMessage?: string): string => {
    const friendlyMessage = parseApiError(error, t);
    
    // 如果解析后的消息是通用的，且有自定义回退消息，使用回退消息
    if (friendlyMessage === (t ? t('errors.unknown') : '操作失败，请稍后重试') && fallbackMessage) {
      return fallbackMessage;
    }

    return friendlyMessage;
  };

  return { handleError, parseApiError: (e: unknown) => parseApiError(e, t) };
}

// 默认导出
export default {
  parseApiError,
  getErrorType,
  isRetryableError,
  createFriendlyError,
  AppError,
  TimeoutError,
  NetworkError,
  ErrorType,
};