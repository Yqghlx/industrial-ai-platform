import { describe, it, expect } from 'vitest';

import {
  parseApiError,
  getErrorType,
  isRetryableError,
  createFriendlyError,
  AppError,
  TimeoutError,
  NetworkError,
  ErrorType,
} from './errorHelper';

describe('errorHelper', () => {
  describe('parseApiError', () => {
    it('handles network errors (Failed to fetch)', () => {
      const error = new TypeError('Failed to fetch');
      expect(parseApiError(error)).toBe('网络连接失败，请检查网络设置');
    });

    it('handles timeout errors (AbortError)', () => {
      const error = new Error('aborted');
      error.name = 'AbortError';
      expect(parseApiError(error)).toBe('请求超时，请稍后重试');
    });

    it('handles unauthorized errors', () => {
      const error = new Error('unauthorized');
      expect(parseApiError(error)).toBe('登录已过期，请重新登录');
    });

    it('handles exact Unauthorized message', () => {
      const error = new Error('Unauthorized');
      expect(parseApiError(error)).toBe('登录已过期，请重新登录');
    });

    it('handles forbidden errors', () => {
      const error = new Error('forbidden access');
      expect(parseApiError(error)).toBe('您没有权限执行此操作');
    });

    it('handles not found errors', () => {
      const error = new Error('resource not found');
      expect(parseApiError(error)).toBe('请求的资源不存在');
    });

    it('handles 404 errors', () => {
      const error = new Error('Error 404: page not found');
      expect(parseApiError(error)).toBe('请求的资源不存在');
    });

    it('handles rate limit errors', () => {
      const error = new Error('rate limit exceeded');
      expect(parseApiError(error)).toBe('请求过于频繁，请稍后再试');
    });

    it('handles 429 errors', () => {
      const error = new Error('HTTP 429 Too Many Requests');
      expect(parseApiError(error)).toBe('请求过于频繁，请稍后再试');
    });

    it('handles validation errors', () => {
      const error = new Error('validation failed: email is required');
      expect(parseApiError(error)).toBe('validation failed: email is required');
    });

    it('handles invalid input errors', () => {
      const error = new Error('invalid input data');
      expect(parseApiError(error)).toBe('invalid input data');
    });

    it('handles server errors (500)', () => {
      const error = new Error('HTTP 500 internal server error');
      expect(parseApiError(error)).toBe('服务器错误，请稍后重试');
    });

    it('handles internal server error message', () => {
      const error = new Error('internal server error');
      expect(parseApiError(error)).toBe('服务器错误，请稍后重试');
    });

    it('returns original message for other known errors', () => {
      const error = new Error('Something specific went wrong');
      expect(parseApiError(error)).toBe('Something specific went wrong');
    });

    it('handles Unknown error as default', () => {
      const error = new Error('Unknown error');
      expect(parseApiError(error)).toBe('操作失败，请稍后重试');
    });

    it('handles non-Error values', () => {
      expect(parseApiError('string error')).toBe('操作失败，请稍后重试');
      expect(parseApiError(null)).toBe('操作失败，请稍后重试');
      expect(parseApiError(undefined)).toBe('操作失败，请稍后重试');
      expect(parseApiError(42)).toBe('操作失败，请稍后重试');
    });

    it('uses translation function when provided', () => {
      const t = (key: string) => `[${key}]`;
      expect(parseApiError(new TypeError('Failed to fetch'), t)).toBe('[errors.networkError]');
      expect(parseApiError(new Error('Unknown error'), t)).toBe('[errors.unknown]');
    });
  });

  describe('getErrorType', () => {
    it('identifies network errors', () => {
      expect(getErrorType(new TypeError('Failed to fetch'))).toBe(ErrorType.NETWORK);
    });

    it('identifies timeout errors', () => {
      const error = new Error('aborted');
      error.name = 'AbortError';
      expect(getErrorType(error)).toBe(ErrorType.TIMEOUT);
    });

    it('identifies unauthorized errors', () => {
      expect(getErrorType(new Error('unauthorized'))).toBe(ErrorType.UNAUTHORIZED);
    });

    it('identifies forbidden errors', () => {
      expect(getErrorType(new Error('forbidden'))).toBe(ErrorType.FORBIDDEN);
    });

    it('identifies not found errors', () => {
      expect(getErrorType(new Error('not found'))).toBe(ErrorType.NOT_FOUND);
    });

    it('identifies rate limit errors', () => {
      expect(getErrorType(new Error('rate limit'))).toBe(ErrorType.RATE_LIMIT);
    });

    it('identifies server errors', () => {
      expect(getErrorType(new Error('500 error'))).toBe(ErrorType.SERVER);
    });

    it('identifies validation errors', () => {
      expect(getErrorType(new Error('validation failed'))).toBe(ErrorType.VALIDATION);
    });

    it('returns UNKNOWN for unrecognized errors', () => {
      expect(getErrorType(new Error('something random'))).toBe(ErrorType.UNKNOWN);
      expect(getErrorType('string')).toBe(ErrorType.UNKNOWN);
    });
  });

  describe('isRetryableError', () => {
    it('network errors are retryable', () => {
      expect(isRetryableError(new TypeError('Failed to fetch'))).toBe(true);
    });

    it('timeout errors are retryable', () => {
      const error = new Error('aborted');
      error.name = 'AbortError';
      expect(isRetryableError(error)).toBe(true);
    });

    it('rate limit errors are retryable', () => {
      expect(isRetryableError(new Error('rate limit exceeded'))).toBe(true);
    });

    it('server errors are retryable', () => {
      expect(isRetryableError(new Error('500 error'))).toBe(true);
    });

    it('unauthorized errors are not retryable', () => {
      expect(isRetryableError(new Error('unauthorized'))).toBe(false);
    });

    it('forbidden errors are not retryable', () => {
      expect(isRetryableError(new Error('forbidden'))).toBe(false);
    });

    it('validation errors are not retryable', () => {
      expect(isRetryableError(new Error('validation failed'))).toBe(false);
    });
  });

  describe('AppError', () => {
    it('creates AppError with message and type', () => {
      const error = new AppError('test message', ErrorType.NETWORK);
      expect(error.message).toBe('test message');
      expect(error.type).toBe(ErrorType.NETWORK);
      expect(error.name).toBe('AppError');
    });

    it('stores original error', () => {
      const original = new Error('original');
      const error = new AppError('wrapped', ErrorType.UNKNOWN, original);
      expect(error.originalError).toBe(original);
    });

    it('defaults to UNKNOWN type', () => {
      const error = new AppError('message');
      expect(error.type).toBe(ErrorType.UNKNOWN);
    });
  });

  describe('TimeoutError', () => {
    it('creates TimeoutError with correct message', () => {
      const error = new TimeoutError();
      expect(error.message).toBe('请求超时');
      expect(error.name).toBe('TimeoutError');
    });
  });

  describe('NetworkError', () => {
    it('creates NetworkError with correct message', () => {
      const error = new NetworkError();
      expect(error.message).toBe('网络连接失败');
      expect(error.name).toBe('NetworkError');
    });
  });

  describe('createFriendlyError', () => {
    it('creates AppError from raw error', () => {
      const error = createFriendlyError(new TypeError('Failed to fetch'));
      expect(error).toBeInstanceOf(AppError);
      expect(error.type).toBe(ErrorType.NETWORK);
      expect(error.message).toBe('网络连接失败，请检查网络设置');
    });

    it('creates AppError from unknown error', () => {
      const error = createFriendlyError('something');
      expect(error).toBeInstanceOf(AppError);
      expect(error.type).toBe(ErrorType.UNKNOWN);
    });

    it('stores original error', () => {
      const original = new Error('test');
      const error = createFriendlyError(original);
      expect(error.originalError).toBe(original);
    });
  });
});
