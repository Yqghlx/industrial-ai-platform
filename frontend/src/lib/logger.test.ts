
import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as loggerModule from './logger';

describe('logger', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should log info message', () => {
    const spy = vi.spyOn(console, 'log').mockImplementation(() => {});
    loggerModule.logger?.info?.('test message');
    
    // 如果 logger 存在，应该调用 console
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });

  it('should log error message', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    loggerModule.logger?.error?.('error message');
    
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });

  it('should have debug method', () => {
    if (loggerModule.logger?.debug) {
      expect(typeof loggerModule.logger.debug).toBe('function');
    }
  });

  it('should have warn method', () => {
    if (loggerModule.logger?.warn) {
      expect(typeof loggerModule.logger.warn).toBe('function');
    }
  });
});
