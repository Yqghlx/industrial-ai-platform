// hooks.ts 测试
// 测试各种自定义 Hook

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';

// Mock timers
vi.useFakeTimers();

// Mock localStorage
const localStorageMock = {
  store: {} as Record<string, string>,
  getItem: vi.fn((key: string) => localStorageMock.store[key] || null),
  setItem: vi.fn((key: string, value: string) => {
    localStorageMock.store[key] = value;
  }),
  removeItem: vi.fn((key: string) => {
    delete localStorageMock.store[key];
  }),
  clear: vi.fn(() => {
    localStorageMock.store = {};
  }),
};

vi.stubGlobal('localStorage', localStorageMock);

// Mock matchMedia
const matchMediaMock = vi.fn((query: string) => ({
  matches: false,
  media: query,
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
}));

vi.stubGlobal('matchMedia', matchMediaMock);

// Mock navigator
vi.stubGlobal('navigator', {
  onLine: true,
});

// Import hooks after mocks
import {
  useDebounce,
  useThrottle,
  useInterval,
  useDeferredValue,
  useVirtualList,
  useOnlineStatus,
  useMediaQuery,
  useBreakpoint,
  useLocalStorage,
} from './hooks';

describe('useDebounce', () => {
  beforeEach(() => {
    vi.clearAllTimers();
    localStorageMock.store = {};
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should return initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('test', 500));
    expect(result.current).toBe('test');
  });

  it('should debounce value changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial', delay: 500 } }
    );

    expect(result.current).toBe('initial');

    // Change value
    rerender({ value: 'changed', delay: 500 });

    // Should still be initial before delay
    expect(result.current).toBe('initial');

    // Fast forward time
    act(() => {
      vi.advanceTimersByTime(500);
    });

    // Now should be changed
    expect(result.current).toBe('changed');
  });

  it('should use different delay values', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial', delay: 100 } }
    );

    rerender({ value: 'changed', delay: 100 });

    act(() => {
      vi.advanceTimersByTime(100);
    });

    expect(result.current).toBe('changed');
  });

  it('should cancel pending update on value change', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial', delay: 500 } }
    );

    rerender({ value: 'first', delay: 500 });
    
    act(() => {
      vi.advanceTimersByTime(200);
    });

    rerender({ value: 'second', delay: 500 });

    act(() => {
      vi.advanceTimersByTime(500);
    });

    expect(result.current).toBe('second');
  });
});

describe('useThrottle', () => {
  beforeEach(() => {
    vi.clearAllTimers();
    vi.setSystemTime(0);
  });

  it('should return initial value', () => {
    const { result } = renderHook(() => useThrottle('test', 500));
    expect(result.current).toBe('test');
  });

  it('should throttle rapid changes', () => {
    const { result, rerender } = renderHook(
      ({ value, interval }) => useThrottle(value, interval),
      { initialProps: { value: 'initial', interval: 500 } }
    );

    // First change after interval should update immediately
    vi.setSystemTime(600);
    rerender({ value: 'first', interval: 500 });
    expect(result.current).toBe('first');

    // Rapid change within interval should be delayed
    vi.setSystemTime(700);
    rerender({ value: 'second', interval: 500 });

    act(() => {
      vi.advanceTimersByTime(400);
    });

    expect(result.current).toBe('second');
  });
});

describe('useInterval', () => {
  beforeEach(() => {
    vi.clearAllTimers();
  });

  it('should call callback at interval', () => {
    const callback = vi.fn();
    
    renderHook(() => useInterval(callback, 1000));

    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(callback).toHaveBeenCalledTimes(1);

    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(callback).toHaveBeenCalledTimes(2);
  });

  it('should not run when delay is null', () => {
    const callback = vi.fn();
    
    renderHook(() => useInterval(callback, null));

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(callback).not.toHaveBeenCalled();
  });

  it('should cleanup on unmount', () => {
    const callback = vi.fn();
    
    const { unmount } = renderHook(() => useInterval(callback, 1000));

    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(callback).toHaveBeenCalledTimes(1);

    unmount();

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    // Should not have been called after unmount
    expect(callback).toHaveBeenCalledTimes(1);
  });
});

describe('useDeferredValue', () => {
  beforeEach(() => {
    vi.clearAllTimers();
  });

  it('should return initial value', () => {
    const { result } = renderHook(() => useDeferredValue('test', 200));
    expect(result.current).toBe('test');
  });

  it('should defer value update', () => {
    const { result, rerender } = renderHook(
      ({ value, timeout }) => useDeferredValue(value, timeout),
      { initialProps: { value: 'initial', timeout: 200 } }
    );

    rerender({ value: 'changed', timeout: 200 });

    // Should still be initial before timeout
    expect(result.current).toBe('initial');

    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(result.current).toBe('changed');
  });

  it('should use default timeout', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDeferredValue(value),
      { initialProps: { value: 'initial' } }
    );

    rerender({ value: 'changed' });

    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(result.current).toBe('changed');
  });
});

describe('useVirtualList', () => {
  it('should calculate virtual items correctly', () => {
    const items = Array.from({ length: 100 }, (_, i) => ({ id: i, name: `Item ${i}` }));
    
    const { result } = renderHook(() =>
      useVirtualList(items, {
        itemHeight: 50,
        containerHeight: 500,
        overscan: 3,
      })
    );

    expect(result.current.totalHeight).toBe(5000); // 100 * 50
    expect(result.current.virtualItems.length).toBeGreaterThan(0);
  });

  it('should handle empty items', () => {
    const { result } = renderHook(() =>
      useVirtualList([], {
        itemHeight: 50,
        containerHeight: 500,
      })
    );

    expect(result.current.totalHeight).toBe(0);
    expect(result.current.virtualItems.length).toBe(0);
  });

  it('should provide scrollToIndex function', () => {
    const items = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    
    const { result } = renderHook(() =>
      useVirtualList(items, {
        itemHeight: 50,
        containerHeight: 500,
      })
    );

    act(() => {
      result.current.scrollToIndex(50);
    });

    // Check that scrollToIndex is callable
    expect(result.current.scrollToIndex).toBeDefined();
  });

  it('should calculate correct visible range', () => {
    const items = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    
    const { result } = renderHook(() =>
      useVirtualList(items, {
        itemHeight: 50,
        containerHeight: 200,
        overscan: 2,
      })
    );

    // Should show items 0-6 (4 visible + 2 overscan on each side)
    const firstItem = result.current.virtualItems[0];
    expect(firstItem.index).toBeGreaterThanOrEqual(0);
    expect(firstItem.style.position).toBe('absolute');
    expect(firstItem.style.height).toBe(50);
  });
});

describe('useOnlineStatus', () => {
  beforeEach(() => {
    vi.stubGlobal('navigator', { onLine: true });
  });

  it('should return initial online status', () => {
    const { result } = renderHook(() => useOnlineStatus());
    expect(result.current).toBe(true);
  });

  it('should handle offline events', () => {
    const { result } = renderHook(() => useOnlineStatus());

    act(() => {
      vi.stubGlobal('navigator', { onLine: false });
      window.dispatchEvent(new Event('offline'));
    });

    expect(result.current).toBe(false);
  });

  it('should handle online events', () => {
    vi.stubGlobal('navigator', { onLine: false });
    
    const { result } = renderHook(() => useOnlineStatus());
    expect(result.current).toBe(false);

    act(() => {
      vi.stubGlobal('navigator', { onLine: true });
      window.dispatchEvent(new Event('online'));
    });

    expect(result.current).toBe(true);
  });
});

describe('useMediaQuery', () => {
  beforeEach(() => {
    matchMediaMock.mockClear();
  });

  it('should return false by default', () => {
    const { result } = renderHook(() => useMediaQuery('(max-width: 600px)'));
    expect(result.current).toBe(false);
  });

  it('should call matchMedia with query', () => {
    renderHook(() => useMediaQuery('(max-width: 600px)'));
    expect(matchMediaMock).toHaveBeenCalledWith('(max-width: 600px)');
  });
});

describe('useBreakpoint', () => {
  it('should return breakpoint states', () => {
    const { result } = renderHook(() => useBreakpoint());

    expect(result.current.isMobile).toBeDefined();
    expect(result.current.isTablet).toBeDefined();
    expect(result.current.isDesktop).toBeDefined();
  });
});

describe('useLocalStorage', () => {
  beforeEach(() => {
    localStorageMock.store = {};
    vi.clearAllMocks();
  });

  it('should return initial value when no stored value', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('default');
  });

  it('should return stored value', () => {
    localStorageMock.store['test-key'] = JSON.stringify('stored');

    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('stored');
  });

  it('should update localStorage on setValue', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));

    act(() => {
      result.current[1]('new value');
    });

    expect(localStorageMock.setItem).toHaveBeenCalledWith('test-key', JSON.stringify('new value'));
    expect(result.current[0]).toBe('new value');
  });

  it('should handle function updater', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 10));

    act(() => {
      result.current[1]((prev) => prev + 5);
    });

    expect(result.current[0]).toBe(15);
  });

  it('should handle JSON parse errors', () => {
    localStorageMock.store['test-key'] = 'invalid json';

    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('default');
  });

  it('should handle complex objects', () => {
    const initialValue = { name: 'test', count: 0 };
    const { result } = renderHook(() => useLocalStorage('test-key', initialValue));

    expect(result.current[0]).toEqual(initialValue);

    act(() => {
      result.current[1]({ name: 'updated', count: 5 });
    });

    expect(result.current[0]).toEqual({ name: 'updated', count: 5 });
  });
});