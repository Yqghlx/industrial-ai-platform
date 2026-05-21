/**
 * 组件缓存工具
 * 用于缓存组件状态和数据
 */
import React from 'react';

type CacheKey = string;
type CacheValue = unknown;
type CacheEntry<T = CacheValue> = {
  value: T;
  timestamp: number;
  ttl: number;
};

/**
 * 内存缓存
 */
class MemoryCache {
  private cache: Map<CacheKey, CacheEntry> = new Map();
  private cleanupInterval: number | null = null;

  constructor() {
    // 每分钟清理过期缓存
    this.cleanupInterval = window.setInterval(() => this.cleanup(), 60000);
  }

  /**
   * 设置缓存
   */
  set<T>(key: CacheKey, value: T, ttl: number = 5 * 60 * 1000): void {
    this.cache.set(key, {
      value,
      timestamp: Date.now(),
      ttl,
    });
  }

  /**
   * 获取缓存
   */
  get<T>(key: CacheKey): T | null {
    const entry = this.cache.get(key);
    if (!entry) return null;

    // 检查是否过期
    if (Date.now() - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      return null;
    }

    return entry.value as T;
  }

  /**
   * 检查缓存是否存在
   */
  has(key: CacheKey): boolean {
    const entry = this.cache.get(key);
    if (!entry) return false;

    // 检查是否过期
    if (Date.now() - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      return false;
    }

    return true;
  }

  /**
   * 删除缓存
   */
  delete(key: CacheKey): boolean {
    return this.cache.delete(key);
  }

  /**
   * 清空缓存
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * 清理过期缓存
   */
  private cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of this.cache.entries()) {
      if (now - entry.timestamp > entry.ttl) {
        this.cache.delete(key);
      }
    }
  }

  /**
   * 销毁缓存
   */
  destroy(): void {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
    }
    this.cache.clear();
  }
}

// 单例缓存实例
export const cache = new MemoryCache();

/**
 * React Hook: 使用缓存的异步数据
 */
export function useCachedAsync<T>(
  key: string,
  fetcher: () => Promise<T>,
  options: {
    ttl?: number;
    enabled?: boolean;
    onSuccess?: (data: T) => void;
    onError?: (error: Error) => void;
  } = {}
): {
  data: T | null;
  loading: boolean;
  error: Error | null;
  refetch: () => void;
} {
  const { ttl = 5 * 60 * 1000, enabled = true, onSuccess, onError } = options;
  const [data, setData] = React.useState<T | null>(() => cache.get<T>(key));
  const [loading, setLoading] = React.useState(!data && enabled);
  const [error, setError] = React.useState<Error | null>(null);

  const fetchData = React.useCallback(async () => {
    // 检查缓存
    const cached = cache.get<T>(key);
    if (cached) {
      setData(cached);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const result = await fetcher();
      cache.set(key, result, ttl);
      setData(result);
      onSuccess?.(result);
    } catch (err) {
      setError(err as Error);
      onError?.(err as Error);
    } finally {
      setLoading(false);
    }
  }, [key, fetcher, ttl, onSuccess, onError]);

  React.useEffect(() => {
    if (enabled) {
      fetchData();
    }
  }, [enabled, fetchData]);

  return { data, loading, error, refetch: fetchData };
}

/**
 * React Hook: 预加载数据
 */
export function usePreload<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttl: number = 5 * 60 * 1000
): void {
  React.useEffect(() => {
    // 仅当缓存不存在时预加载
    if (!cache.has(key)) {
      fetcher().then(data => {
        cache.set(key, data, ttl);
      }).catch(console.warn);
    }
  }, [key, fetcher, ttl]);
}

/**
 * 缓存装饰器
 * 用于缓存函数结果
 */
export function cached(key: string, ttl: number = 5 * 60 * 1000) {
  return function (
    target: unknown,
    propertyKey: string,
    descriptor: PropertyDescriptor
  ) {
    const originalMethod = descriptor.value;

    descriptor.value = async function (...args: unknown[]) {
      const cacheKey = `${key}:${JSON.stringify(args)}`;
      const cachedResult = cache.get(cacheKey);

      if (cachedResult) {
        return cachedResult;
      }

      const result = await originalMethod.apply(this, args);
      cache.set(cacheKey, result, ttl);
      return result;
    };

    return descriptor;
  };
}

/**
 * 获取缓存统计信息
 */
export function getCacheStats(): {
  size: number;
  keys: string[];
} {
  return {
    size: (cache as unknown as { cache: Map<string, unknown> }).cache.size,
    keys: Array.from((cache as unknown as { cache: Map<string, unknown> }).cache.keys()),
  };
}

export default cache;