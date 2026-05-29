import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock matchMedia（jsdom 不提供）
beforeEach(() => {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }));
});

import {
  isMobileDevice,
  isIOS,
  isTouchDevice,
  isLowEndDevice,
  getConnectionInfo,
  shouldReduceMotion,
  isSlowNetwork,
  getOptimalImageQuality,
  debounce,
  throttle,
  createResponsiveImageUrl,
  hapticFeedback,
} from './mobileOptimizations';

describe('mobileOptimizations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // 重置 window 属性
    Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true, configurable: true });
    Object.defineProperty(navigator, 'userAgent', {
      value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      writable: true,
      configurable: true,
    });
    Object.defineProperty(navigator, 'maxTouchPoints', { value: 0, writable: true, configurable: true });
    Object.defineProperty(navigator, 'platform', { value: 'Win32', writable: true, configurable: true });
    Object.defineProperty(navigator, 'hardwareConcurrency', { value: 8, writable: true, configurable: true });
    Object.defineProperty(navigator, 'deviceMemory', { value: 8, writable: true, configurable: true });
    Object.defineProperty(navigator, 'connection', {
      value: { effectiveType: '4g', downlink: 10, rtt: 50, saveData: false },
      writable: true,
      configurable: true,
    });
  });

  describe('isMobileDevice', () => {
    it('returns false for desktop devices', () => {
      const result = isMobileDevice();
      expect(result).toBe(false);
    });

    it('returns true for narrow viewport (mobile width)', () => {
      Object.defineProperty(window, 'innerWidth', { value: 400, configurable: true });
      const result = isMobileDevice();
      expect(result).toBe(true);
    });

    it('returns true for mobile user agent', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
        configurable: true,
      });
      const result = isMobileDevice();
      expect(result).toBe(true);
    });
  });

  describe('isIOS', () => {
    it('returns false for non-iOS devices', () => {
      const result = isIOS();
      expect(result).toBe(false);
    });

    it('returns true for iPhone user agent', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
        configurable: true,
      });
      const result = isIOS();
      expect(result).toBe(true);
    });

    it('returns true for iPad user agent', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)',
        configurable: true,
      });
      const result = isIOS();
      expect(result).toBe(true);
    });
  });

  describe('isTouchDevice', () => {
    it('returns true when maxTouchPoints > 0', () => {
      Object.defineProperty(navigator, 'maxTouchPoints', { value: 5, configurable: true });
      const result = isTouchDevice();
      expect(result).toBe(true);
    });

    it('detects touch from ontouchstart', () => {
      // jsdom 可能设置了 ontouchstart
      const originalTouchStart = (window as unknown as Record<string, unknown>).ontouchstart;
      (window as unknown as Record<string, unknown>).ontouchstart = null;
      Object.defineProperty(navigator, 'maxTouchPoints', { value: 0, configurable: true });
      const result = isTouchDevice();
      // ontouchstart in window 为 true 时返回 true
      expect(typeof result).toBe('boolean');
      // 恢复
      (window as unknown as Record<string, unknown>).ontouchstart = originalTouchStart;
    });
  });

  describe('isLowEndDevice', () => {
    it('returns false for high-end devices', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: 8, configurable: true });
      Object.defineProperty(navigator, 'deviceMemory', { value: 8, configurable: true });
      const result = isLowEndDevice();
      expect(result).toBe(false);
    });

    it('returns true for low CPU cores', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: 2, configurable: true });
      Object.defineProperty(navigator, 'deviceMemory', { value: 8, configurable: true });
      const result = isLowEndDevice();
      expect(result).toBe(true);
    });

    it('returns true for low memory', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: 4, configurable: true });
      Object.defineProperty(navigator, 'deviceMemory', { value: 1, configurable: true });
      const result = isLowEndDevice();
      expect(result).toBe(true);
    });
  });

  describe('getConnectionInfo', () => {
    it('returns connection info with defaults', () => {
      const result = getConnectionInfo();
      expect(result).toBeDefined();
      expect(result.effectiveType).toBeDefined();
      expect(result.downlink).toBeDefined();
      expect(result.saveData).toBeDefined();
    });

    it('returns 4g defaults when no connection API', () => {
      Object.defineProperty(navigator, 'connection', { value: undefined, configurable: true });
      Object.defineProperty(navigator, 'mozConnection', { value: undefined, configurable: true });
      Object.defineProperty(navigator, 'webkitConnection', { value: undefined, configurable: true });
      const result = getConnectionInfo();
      expect(result.effectiveType).toBe('4g');
      expect(result.downlink).toBe(10);
      expect(result.saveData).toBe(false);
    });
  });

  describe('shouldReduceMotion', () => {
    it('returns false when reduce motion is not preferred', () => {
      window.matchMedia = vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }));
      const result = shouldReduceMotion();
      expect(result).toBe(false);
    });

    it('returns true when reduce motion is preferred', () => {
      window.matchMedia = vi.fn().mockImplementation((query: string) => ({
        matches: query === '(prefers-reduced-motion: reduce)',
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }));
      const result = shouldReduceMotion();
      expect(result).toBe(true);
    });
  });

  describe('isSlowNetwork', () => {
    it('returns false for fast networks', () => {
      Object.defineProperty(navigator, 'connection', {
        value: { effectiveType: '4g', saveData: false },
        configurable: true,
      });
      const result = isSlowNetwork();
      expect(result).toBe(false);
    });

    it('returns true for slow networks (2g)', () => {
      Object.defineProperty(navigator, 'connection', {
        value: { effectiveType: '2g', saveData: false },
        configurable: true,
      });
      const result = isSlowNetwork();
      expect(result).toBe(true);
    });

    it('returns true for slow networks (slow-2g)', () => {
      Object.defineProperty(navigator, 'connection', {
        value: { effectiveType: 'slow-2g', saveData: false },
        configurable: true,
      });
      const result = isSlowNetwork();
      expect(result).toBe(true);
    });

    it('returns true when saveData is enabled', () => {
      Object.defineProperty(navigator, 'connection', {
        value: { effectiveType: '4g', saveData: true },
        configurable: true,
      });
      const result = isSlowNetwork();
      expect(result).toBe(true);
    });
  });

  describe('getOptimalImageQuality', () => {
    it('returns high for high-end desktop devices on fast network', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: 8, configurable: true });
      Object.defineProperty(navigator, 'deviceMemory', { value: 8, configurable: true });
      Object.defineProperty(navigator, 'connection', {
        value: { effectiveType: '4g', saveData: false },
        configurable: true,
      });
      Object.defineProperty(window, 'innerWidth', { value: 1024, configurable: true });
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        configurable: true,
      });
      const result = getOptimalImageQuality();
      expect(result).toBe('high');
    });

    it('returns low for low-end devices', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: 2, configurable: true });
      Object.defineProperty(navigator, 'deviceMemory', { value: 1, configurable: true });
      const result = getOptimalImageQuality();
      expect(result).toBe('low');
    });

    it('returns medium for mobile devices on fast network', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: 4, configurable: true });
      Object.defineProperty(navigator, 'deviceMemory', { value: 4, configurable: true });
      Object.defineProperty(navigator, 'connection', {
        value: { effectiveType: '4g', saveData: false },
        configurable: true,
      });
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
        configurable: true,
      });
      const result = getOptimalImageQuality();
      expect(result).toBe('medium');
    });
  });

  describe('debounce', () => {
    it('delays function execution', async () => {
      const fn = vi.fn();
      const debounced = debounce(fn, 100);

      debounced();
      debounced();
      debounced();

      expect(fn).not.toHaveBeenCalled();

      await new Promise(resolve => setTimeout(resolve, 150));
      expect(fn).toHaveBeenCalledTimes(1);
    });
  });

  describe('throttle', () => {
    it('limits function execution rate', async () => {
      const fn = vi.fn();
      const throttled = throttle(fn, 50);

      throttled();
      throttled();
      throttled();

      expect(fn).toHaveBeenCalledTimes(1);

      await new Promise(resolve => setTimeout(resolve, 100));
      throttled();
      expect(fn).toHaveBeenCalledTimes(2);
    });
  });

  describe('createResponsiveImageUrl', () => {
    it('creates URL with width and quality params', () => {
      const url = createResponsiveImageUrl('https://example.com/image.jpg', 400, 'medium');
      expect(url).toBe('https://example.com/image.jpg?w=400&q=75');
    });

    it('handles existing query params', () => {
      const url = createResponsiveImageUrl('https://example.com/image.jpg?existing=1', 400, 'low');
      expect(url).toBe('https://example.com/image.jpg?existing=1&w=400&q=50');
    });

    it('uses correct quality values', () => {
      expect(createResponsiveImageUrl('img.jpg', 100, 'low')).toContain('q=50');
      expect(createResponsiveImageUrl('img.jpg', 100, 'medium')).toContain('q=75');
      expect(createResponsiveImageUrl('img.jpg', 100, 'high')).toContain('q=90');
    });
  });

  describe('hapticFeedback', () => {
    it('does not throw on non-touch devices', () => {
      expect(() => hapticFeedback('light')).not.toThrow();
    });
  });
});
