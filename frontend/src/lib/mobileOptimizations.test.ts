import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock window.navigator
const mockNavigator = {
  userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
  maxTouchPoints: 0,
  platform: 'Win32',
  deviceMemory: 8,
  connection: {
    effectiveType: '4g',
    downlink: 10,
    rtt: 50,
    saveData: false,
  },
};

Object.defineProperty(window, 'navigator', {
  value: mockNavigator,
  writable: true,
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
} from '../lib/mobileOptimizations';

describe.skip('mobileOptimizations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('isMobileDevice', () => {
    it('returns false for desktop devices', () => {
      const result = isMobileDevice();
      expect(result).toBe(false);
    });

    it('returns true for mobile devices', () => {
      Object.defineProperty(window, 'innerWidth', { value: 400, writable: true });
      const result = isMobileDevice();
      expect(result).toBe(true);
    });
  });

  describe('isIOS', () => {
    it('returns false for non-iOS devices', () => {
      const result = isIOS();
      expect(result).toBe(false);
    });

    it('returns true for iOS devices', () => {
      Object.defineProperty(navigator, 'userAgent', { 
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)', 
        writable: true 
      });
      const result = isIOS();
      expect(result).toBe(true);
    });
  });

  describe('isTouchDevice', () => {
    it('returns false for non-touch devices', () => {
      const result = isTouchDevice();
      expect(result).toBe(false);
    });

    it('returns true for touch devices', () => {
      Object.defineProperty(navigator, 'maxTouchPoints', { value: 5, writable: true });
      const result = isTouchDevice();
      expect(result).toBe(true);
    });
  });

  describe('isLowEndDevice', () => {
    it('returns false for high-end devices', () => {
      Object.defineProperty(navigator, 'deviceMemory', { value: 8, writable: true });
      const result = isLowEndDevice();
      expect(result).toBe(false);
    });

    it('returns true for low-end devices', () => {
      Object.defineProperty(navigator, 'deviceMemory', { value: 1, writable: true });
      const result = isLowEndDevice();
      expect(result).toBe(true);
    });
  });

  describe('getConnectionInfo', () => {
    it('returns connection info', () => {
      const result = getConnectionInfo();
      expect(result).toBeDefined();
      expect(result.effectiveType).toBeDefined();
    });
  });

  describe('shouldReduceMotion', () => {
    it('returns false when reduce motion is not preferred', () => {
      const result = shouldReduceMotion();
      expect(result).toBe(false);
    });
  });

  describe('isSlowNetwork', () => {
    it('returns false for fast networks', () => {
      Object.defineProperty(navigator, 'connection', { 
        value: { effectiveType: '4g' }, 
        writable: true 
      });
      const result = isSlowNetwork();
      expect(result).toBe(false);
    });

    it('returns true for slow networks', () => {
      Object.defineProperty(navigator, 'connection', { 
        value: { effectiveType: '2g' }, 
        writable: true 
      });
      const result = isSlowNetwork();
      expect(result).toBe(true);
    });
  });

  describe('getOptimalImageQuality', () => {
    it('returns high for high-end devices', () => {
      Object.defineProperty(navigator, 'deviceMemory', { value: 8, writable: true });
      Object.defineProperty(navigator, 'connection', { 
        value: { effectiveType: '4g' }, 
        writable: true 
      });
      const result = getOptimalImageQuality();
      expect(result).toBe('high');
    });

    it('returns low for low-end devices', () => {
      Object.defineProperty(navigator, 'deviceMemory', { value: 1, writable: true });
      const result = getOptimalImageQuality();
      expect(result).toBe('low');
    });
  });
});