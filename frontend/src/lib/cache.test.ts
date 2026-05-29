import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// 直接测试 MemoryCache 类而不是通过 hook（避免 React 依赖复杂性）
import { cache, getCacheStats, cached } from './cache';

describe('MemoryCache (cache instance)', () => {
  beforeEach(() => {
    cache.clear();
  });

  it('sets and gets a value', () => {
    cache.set('key1', 'value1');
    expect(cache.get('key1')).toBe('value1');
  });

  it('returns null for non-existent key', () => {
    expect(cache.get('nonexistent')).toBeNull();
  });

  it('returns null for expired entries', () => {
    cache.set('short-lived', 'data', 1); // 1ms TTL
    // 等待过期
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        expect(cache.get('short-lived')).toBeNull();
        resolve();
      }, 10);
    });
  });

  it('checks if key exists with has()', () => {
    cache.set('exists', 'data');
    expect(cache.has('exists')).toBe(true);
    expect(cache.has('missing')).toBe(false);
  });

  it('deletes entries', () => {
    cache.set('to-delete', 'data');
    expect(cache.has('to-delete')).toBe(true);
    cache.delete('to-delete');
    expect(cache.has('to-delete')).toBe(false);
  });

  it('clears all entries', () => {
    cache.set('key1', 'value1');
    cache.set('key2', 'value2');
    cache.clear();
    expect(cache.get('key1')).toBeNull();
    expect(cache.get('key2')).toBeNull();
  });

  it('handles different value types', () => {
    cache.set('number', 42);
    cache.set('object', { name: 'test' });
    cache.set('array', [1, 2, 3]);
    cache.set('boolean', true);
    cache.set('null', null);

    expect(cache.get('number')).toBe(42);
    expect(cache.get('object')).toEqual({ name: 'test' });
    expect(cache.get('array')).toEqual([1, 2, 3]);
    expect(cache.get('boolean')).toBe(true);
    expect(cache.get('null')).toBeNull();
  });

  it('overwrites existing entries', () => {
    cache.set('key', 'first');
    expect(cache.get('key')).toBe('first');
    cache.set('key', 'second');
    expect(cache.get('key')).toBe('second');
  });

  it('uses default TTL when not specified', () => {
    cache.set('default-ttl', 'data');
    expect(cache.get('default-ttl')).toBe('data');
  });
});

describe('getCacheStats', () => {
  beforeEach(() => {
    cache.clear();
  });

  it('returns empty stats when cache is empty', () => {
    const stats = getCacheStats();
    expect(stats.size).toBe(0);
    expect(stats.keys).toEqual([]);
  });

  it('returns correct stats', () => {
    cache.set('key1', 'value1');
    cache.set('key2', 'value2');
    const stats = getCacheStats();
    expect(stats.size).toBe(2);
    expect(stats.keys).toContain('key1');
    expect(stats.keys).toContain('key2');
  });
});

describe('cached decorator', () => {
  beforeEach(() => {
    cache.clear();
  });

  it('caches method results', async () => {
    const mockFn = vi.fn().mockResolvedValue('result');
    const descriptor = {
      value: mockFn,
    };

    const decorator = cached('test-key');
    const decorated = decorator({}, 'methodName', descriptor);

    // 第一次调用
    await decorated.value.call({});
    expect(mockFn).toHaveBeenCalledTimes(1);

    // 第二次调用应从缓存获取
    await decorated.value.call({});
    expect(mockFn).toHaveBeenCalledTimes(1); // 仍然是 1 次
  });
});
