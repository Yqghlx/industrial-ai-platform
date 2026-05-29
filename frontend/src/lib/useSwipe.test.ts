import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';

import { useSwipe, useIsMobile, useIsIOS, useViewportHeight } from './useSwipe';

// Mock matchMedia
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

describe('useSwipe', () => {
  it('returns touch event handlers', () => {
    const { result } = renderHook(() =>
      useSwipe({
        onSwipeLeft: vi.fn(),
        onSwipeRight: vi.fn(),
      })
    );

    expect(result.current.onTouchStart).toBeDefined();
    expect(result.current.onTouchMove).toBeDefined();
    expect(result.current.onTouchEnd).toBeDefined();
    expect(typeof result.current.onTouchStart).toBe('function');
    expect(typeof result.current.onTouchMove).toBe('function');
    expect(typeof result.current.onTouchEnd).toBe('function');
  });

  it('calls onSwipeLeft when swiping left', () => {
    const onSwipeLeft = vi.fn();
    const { result } = renderHook(() =>
      useSwipe({ onSwipeLeft, onSwipeRight: vi.fn() })
    );

    // 模拟触摸开始
    act(() => {
      result.current.onTouchStart({
        touches: [{ clientX: 200, clientY: 100 }],
      } as React.TouchEvent);
    });

    // 模拟触摸移动（向左）
    act(() => {
      result.current.onTouchMove({
        touches: [{ clientX: 50, clientY: 100 }],
        preventDefault: vi.fn(),
      } as unknown as React.TouchEvent);
    });

    // 模拟触摸结束
    act(() => {
      result.current.onTouchEnd({} as React.TouchEvent);
    });

    expect(onSwipeLeft).toHaveBeenCalledTimes(1);
  });

  it('calls onSwipeRight when swiping right', () => {
    const onSwipeRight = vi.fn();
    const { result } = renderHook(() =>
      useSwipe({ onSwipeLeft: vi.fn(), onSwipeRight })
    );

    act(() => {
      result.current.onTouchStart({
        touches: [{ clientX: 50, clientY: 100 }],
      } as React.TouchEvent);
    });

    act(() => {
      result.current.onTouchMove({
        touches: [{ clientX: 200, clientY: 100 }],
        preventDefault: vi.fn(),
      } as unknown as React.TouchEvent);
    });

    act(() => {
      result.current.onTouchEnd({} as React.TouchEvent);
    });

    expect(onSwipeRight).toHaveBeenCalledTimes(1);
  });

  it('does not call callbacks for vertical swipes', () => {
    const onSwipeLeft = vi.fn();
    const onSwipeRight = vi.fn();
    const { result } = renderHook(() =>
      useSwipe({ onSwipeLeft, onSwipeRight })
    );

    act(() => {
      result.current.onTouchStart({
        touches: [{ clientX: 100, clientY: 50 }],
      } as React.TouchEvent);
    });

    // 垂直移动
    act(() => {
      result.current.onTouchMove({
        touches: [{ clientX: 100, clientY: 200 }],
        preventDefault: vi.fn(),
      } as unknown as React.TouchEvent);
    });

    act(() => {
      result.current.onTouchEnd({} as React.TouchEvent);
    });

    expect(onSwipeLeft).not.toHaveBeenCalled();
    expect(onSwipeRight).not.toHaveBeenCalled();
  });

  it('does not crash without callbacks', () => {
    const { result } = renderHook(() => useSwipe({}));

    act(() => {
      result.current.onTouchStart({
        touches: [{ clientX: 200, clientY: 100 }],
      } as React.TouchEvent);
    });

    act(() => {
      result.current.onTouchMove({
        touches: [{ clientX: 50, clientY: 100 }],
        preventDefault: vi.fn(),
      } as unknown as React.TouchEvent);
    });

    // 不应抛出错误
    expect(() => {
      act(() => {
        result.current.onTouchEnd({} as React.TouchEvent);
      });
    }).not.toThrow();
  });
});

describe('useIsMobile', () => {
  it('returns boolean value', () => {
    const { result } = renderHook(() => useIsMobile());
    expect(typeof result.current).toBe('boolean');
  });

  it('returns true for narrow viewport', () => {
    Object.defineProperty(window, 'innerWidth', { value: 500, writable: true, configurable: true });
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);
  });
});

describe('useIsIOS', () => {
  it('returns boolean value', () => {
    const { result } = renderHook(() => useIsIOS());
    expect(typeof result.current).toBe('boolean');
  });
});

describe('useViewportHeight', () => {
  it('sets CSS custom property --vh', () => {
    const setPropertySpy = vi.fn();
    vi.spyOn(document.documentElement.style, 'setProperty').mockImplementation(setPropertySpy);

    renderHook(() => useViewportHeight());

    expect(setPropertySpy).toHaveBeenCalledWith('--vh', expect.stringContaining('px'));
  });
});
