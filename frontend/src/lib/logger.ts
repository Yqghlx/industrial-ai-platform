/**
 * Logger - 统一日志服务
 * 
 * FE-P3-08/09: 统一 console 调用处理
 * - 开发环境：输出到 console
 * - 生产环境：可选择上报到监控服务
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogEntry {
  level: LogLevel;
  message: string;
  data?: unknown;
  timestamp: Date;
}

// 是否为开发环境
const isDev = import.meta.env.DEV;

// 日志存储（用于调试或上报）
const logBuffer: LogEntry[] = [];
const MAX_BUFFER_SIZE = 100;

/**
 * 添加日志到缓冲区
 */
function addToBuffer(level: LogLevel, message: string, data?: unknown) {
  logBuffer.push({
    level,
    message,
    data,
    timestamp: new Date(),
  });
  
  // 限制缓冲区大小
  if (logBuffer.length > MAX_BUFFER_SIZE) {
    logBuffer.shift();
  }
}

/**
 * 获取日志缓冲区
 */
export function getLogBuffer(): LogEntry[] {
  return [...logBuffer];
}

/**
 * 清空日志缓冲区
 */
export function clearLogBuffer(): void {
  logBuffer.length = 0;
}

/**
 * 日志输出函数
 */
function log(level: LogLevel, message: string, data?: unknown) {
  // 添加到缓冲区
  addToBuffer(level, message, data);
  
  // 开发环境直接输出
  if (isDev) {
    const prefix = `[${level.toUpperCase()}]`;
    switch (level) {
      case 'error':
        console.error(prefix, message, data !== undefined ? data : '');
        break;
      case 'warn':
        console.warn(prefix, message, data !== undefined ? data : '');
        break;
      case 'info':
        // eslint-disable-next-line no-console
        console.info(prefix, message, data !== undefined ? data : '');
        break;
      case 'debug':
        // eslint-disable-next-line no-console
        console.log(prefix, message, data !== undefined ? data : '');
        break;
    }
  }
  
  // 生产环境可选择上报错误日志
  if (!isDev && level === 'error') {
    // 未来可接入错误上报服务
    // reportErrorToService(message, data);
  }
}

/**
 * Logger API
 */
export const logger = {
  debug: (message: string, data?: unknown) => log('debug', message, data),
  info: (message: string, data?: unknown) => log('info', message, data),
  warn: (message: string, data?: unknown) => log('warn', message, data),
  error: (message: string, data?: unknown) => log('error', message, data),
  
  // 便捷方法
  apiError: (endpoint: string, error: unknown) => 
    log('error', `API Error [${endpoint}]`, error),
  
  loadFailed: (resource: string, error: unknown) => 
    log('error', `Failed to load ${resource}`, error),
  
  invalidResponse: (type: string, data: unknown) => 
    log('error', `Invalid ${type} response`, data),
  
  performance: (metric: string, value: unknown) => 
    log('debug', `[Performance] ${metric}`, value),
};

// 默认导出
export default logger;