import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';

/**
 * Escape 键关闭模态框 Hook
 * 封装 Escape 键监听逻辑，避免 7+ 个组件重复实现
 */
export function useEscapeKey(callback: () => void, isActive: boolean) {
  useEffect(() => {
    if (!isActive) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') callback();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [callback, isActive]);
}

/**
 * 防抖 Hook
 */
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}

/**
 * 节流 Hook
 */
export function useThrottle<T>(value: T, interval: number): T {
  const [throttledValue, setThrottledValue] = useState<T>(value);
  const lastUpdated = useRef<number>(Date.now());

  useEffect(() => {
    const now = Date.now();
    if (now - lastUpdated.current >= interval) {
      lastUpdated.current = now;
      setThrottledValue(value);
    } else {
      const timer = setTimeout(() => {
        lastUpdated.current = Date.now();
        setThrottledValue(value);
      }, interval - (now - lastUpdated.current));
      return () => clearTimeout(timer);
    }
  }, [value, interval]);

  return throttledValue;
}

/**
 * 间歇请求 Hook
 */
export function useInterval(callback: () => void, delay: number | null) {
  const savedCallback = useRef(callback);

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  useEffect(() => {
    if (delay === null) return;

    const id = setInterval(() => savedCallback.current(), delay);
    return () => clearInterval(id);
  }, [delay]);
}

/**
 * 延迟值 Hook
 * 用于延迟更新值，避免频繁渲染
 */
export function useDeferredValue<T>(value: T, timeout: number = 200): T {
  const [deferredValue, setDeferredValue] = useState<T>(value);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    
    timeoutRef.current = window.setTimeout(() => {
      setDeferredValue(value);
    }, timeout);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [value, timeout]);

  return deferredValue;
}

/**
 * 虚拟列表 Hook
 * 用于渲染大型列表时优化性能
 * FE-P3-12: 优化大数据集性能，使用 items.length 作为依赖而非 items 引用
 */
export function useVirtualList<T>(
  items: T[],
  options: {
    itemHeight: number;
    containerHeight: number;
    overscan?: number;
  }
): {
  virtualItems: { item: T; index: number; style: React.CSSProperties }[];
  totalHeight: number;
  scrollToIndex: (index: number) => void;
} {
  const { itemHeight, containerHeight, overscan = 3 } = options;
  const [scrollTop, setScrollTop] = useState(0);

  // FE-P3-12: 使用 ref 存储最新 items，避免 useMemo 依赖数组引用
  const itemsRef = useRef(items);
  itemsRef.current = items;

  const totalHeight = items.length * itemHeight;
  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
  const endIndex = Math.min(
    items.length - 1,
    Math.floor((scrollTop + containerHeight) / itemHeight) + overscan
  );

  // FE-P3-12: 使用 itemCount 而非 items 作为依赖，优化大数据集性能
  // Note: itemsRef.current holds the latest items reference, so we only depend on computed values
  const virtualItems = useMemo(() => {
    const result: { item: T; index: number; style: React.CSSProperties }[] = [];
    
    for (let i = startIndex; i <= endIndex; i++) {
      result.push({
        item: itemsRef.current[i],
        index: i,
        style: {
          position: 'absolute' as const,
          top: i * itemHeight,
          height: itemHeight,
          width: '100%',
        },
      });
    }
    
    return result;
  }, [startIndex, endIndex, itemHeight]);

  const scrollToIndex = useCallback((index: number) => {
    setScrollTop(index * itemHeight);
  }, [itemHeight]);

  return { virtualItems, totalHeight, scrollToIndex };
}

/**
 * 网络状态 Hook
 */
export function useOnlineStatus(): boolean {
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  return isOnline;
}

/**
 * 媒体查询 Hook
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia(query).matches;
  });

  useEffect(() => {
    const mediaQuery = window.matchMedia(query);
    const handler = (event: MediaQueryListEvent) => setMatches(event.matches);

    setMatches(mediaQuery.matches);
    mediaQuery.addEventListener('change', handler);

    return () => mediaQuery.removeEventListener('change', handler);
  }, [query]);

  return matches;
}

/**
 * 响应式断点 Hook
 */
export function useBreakpoint() {
  const isMobile = useMediaQuery('(max-width: 639px)');
  const isTablet = useMediaQuery('(min-width: 640px) and (max-width: 1023px)');
  const isDesktop = useMediaQuery('(min-width: 1024px)');

  return { isMobile, isTablet, isDesktop };
}

/**
 * 本地存储 Hook
 */
export function useLocalStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((prev: T) => T)) => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch {
      return initialValue;
    }
  });

  const setValue = useCallback((value: T | ((prev: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      window.localStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.error('Failed to save to localStorage:', error);
    }
  }, [key, storedValue]);

  return [storedValue, setValue];
}

/**
 * Intersection Observer Hook
 * 用于检测元素是否在视口中
 * FIX-014: 使用 useMemo 包裹 options 防止每次新对象导致 Observer 重建
 */
export function useIntersectionObserver(
  options: IntersectionObserverInit = {}
): [React.RefObject<HTMLDivElement | null>, boolean] {
  const ref = useRef<HTMLDivElement | null>(null);
  const [isIntersecting, setIsIntersecting] = useState(false);

  // FIX-014: Memoize options to prevent unnecessary Observer recreation
  const memoizedOptions = useMemo(() => ({
    root: options.root,
    rootMargin: options.rootMargin,
    threshold: options.threshold,
  }), [options.root, options.rootMargin, options.threshold]);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    const observer = new IntersectionObserver(([entry]) => {
      setIsIntersecting(entry.isIntersecting);
    }, memoizedOptions);

    observer.observe(element);

    return () => observer.disconnect();
  }, [memoizedOptions]);

  return [ref, isIntersecting];
}

/**
 * Resize Observer Hook
 * 用于检测元素尺寸变化
 */
export function useResizeObserver(): [React.RefObject<HTMLDivElement | null>, DOMRect | null] {
  const ref = useRef<HTMLDivElement | null>(null);
  const [rect, setRect] = useState<DOMRect | null>(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    const observer = new ResizeObserver(([entry]) => {
      setRect(entry.contentRect);
    });

    observer.observe(element);

    return () => observer.disconnect();
  }, []);

  return [ref, rect];
}

export default {
  useDebounce,
  useThrottle,
  useInterval,
  useDeferredValue,
  useVirtualList,
  useOnlineStatus,
  useMediaQuery,
  useBreakpoint,
  useLocalStorage,
  useIntersectionObserver,
  useResizeObserver,
  useEscapeKey,
};