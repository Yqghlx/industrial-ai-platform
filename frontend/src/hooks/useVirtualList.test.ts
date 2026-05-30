import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useVirtualList } from './useVirtualList';

describe('useVirtualList', () => {
  const defaultOptions = {
    itemCount: 1000,
    itemHeight: 50,
    containerHeight: 500,
    overscan: 3,
  };

  it('计算总高度 = itemCount × itemHeight', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    expect(result.current.totalHeight).toBe(50000);
  });

  it('初始状态 startIndex = 0', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    expect(result.current.startIndex).toBe(0);
  });

  it('初始 endIndex = 可视数量 + overscan', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    // visibleCount = ceil(500/50) = 10, baseEnd = 10 + 3(overscan) → 13
    expect(result.current.endIndex).toBe(13);
  });

  it('visibleIndices 包含 startIndex 到 endIndex 的所有索引', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    expect(result.current.visibleIndices[0]).toBe(0);
    expect(result.current.visibleIndices[result.current.visibleIndices.length - 1]).toBe(13);
    expect(result.current.visibleIndices.length).toBe(14);
  });

  it('空列表时返回安全值', () => {
    const { result } = renderHook(() =>
      useVirtualList({ itemCount: 0, itemHeight: 50, containerHeight: 500 })
    );
    expect(result.current.totalHeight).toBe(0);
    expect(result.current.startIndex).toBe(0);
    // endIndex 可能是 -1 或 0，不应大于 0
    expect(result.current.endIndex).toBeLessThanOrEqual(0);
    expect(result.current.visibleIndices).toEqual([]);
  });

  it('overscan=0 时只渲染可视区域', () => {
    const { result } = renderHook(() =>
      useVirtualList({ ...defaultOptions, overscan: 0 })
    );
    expect(result.current.startIndex).toBe(0);
    expect(result.current.endIndex).toBe(10);
  });

  it('overscan 不超过边界', () => {
    const { result } = renderHook(() =>
      useVirtualList({ itemCount: 5, itemHeight: 50, containerHeight: 500, overscan: 10 })
    );
    expect(result.current.startIndex).toBe(0);
    expect(result.current.endIndex).toBe(4);
  });

  it('containerStyle 包含正确的高度和 overflow', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    expect(result.current.containerStyle).toEqual({
      overflow: 'auto',
      height: 500,
      position: 'relative',
    });
  });

  it('contentStyle 包含正确的总高度', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    expect(result.current.contentStyle).toEqual({
      height: 50000,
      position: 'relative',
    });
  });

  it('getItemStyle 返回正确的 absolute 定位样式', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    const style0 = result.current.getItemStyle(0);
    expect(style0).toEqual({ position: 'absolute', top: 0, height: 50, width: '100%' });

    const style10 = result.current.getItemStyle(10);
    expect(style10.top).toBe(500);
  });

  it('onScroll 更新 scrollTop 并改变 startIndex/endIndex', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));

    // 模拟滚动到 scrollTop=1000
    act(() => {
      result.current.onScroll({
        currentTarget: { scrollTop: 1000 },
      } as unknown as React.UIEvent<HTMLElement>);
    });

    // startIndex = max(0, floor(1000/50) - 3) = 17
    expect(result.current.startIndex).toBe(17);
    // endIndex = min(999, floor(1000/50) + 10 + 3) = 33
    expect(result.current.endIndex).toBe(33);
  });

  it('滚动到接近底部时 endIndex 不超过 itemCount-1', () => {
    const { result } = renderHook(() =>
      useVirtualList({ itemCount: 20, itemHeight: 50, containerHeight: 500, overscan: 3 })
    );

    act(() => {
      result.current.onScroll({
        currentTarget: { scrollTop: 750 },
      } as unknown as React.UIEvent<HTMLElement>);
    });

    expect(result.current.endIndex).toBeLessThanOrEqual(19);
  });

  it('选项变化时重新计算', () => {
    const { result, rerender } = renderHook(
      (opts: typeof defaultOptions) => useVirtualList(opts),
      { initialProps: defaultOptions }
    );

    expect(result.current.totalHeight).toBe(50000);

    rerender({ ...defaultOptions, itemCount: 500 });
    expect(result.current.totalHeight).toBe(25000);
  });

  it('containerHeight 变化时 endIndex 更新', () => {
    const { result, rerender } = renderHook(
      (opts: typeof defaultOptions) => useVirtualList(opts),
      { initialProps: defaultOptions }
    );

    const endSmall = result.current.endIndex;

    rerender({ ...defaultOptions, containerHeight: 1000 });
    expect(result.current.endIndex).toBeGreaterThan(endSmall);
  });

  it('scrollTo 函数存在且可调用', () => {
    const { result } = renderHook(() => useVirtualList(defaultOptions));
    expect(typeof result.current.scrollTo).toBe('function');
    expect(typeof result.current.scrollToTop).toBe('function');
    expect(typeof result.current.scrollToBottom).toBe('function');
  });

  it('itemCount=1 时正常工作', () => {
    const { result } = renderHook(() =>
      useVirtualList({ itemCount: 1, itemHeight: 50, containerHeight: 500, overscan: 3 })
    );
    expect(result.current.totalHeight).toBe(50);
    expect(result.current.startIndex).toBe(0);
    expect(result.current.endIndex).toBe(0);
    expect(result.current.visibleIndices).toEqual([0]);
  });

  it('大列表性能正常（100000 项）', () => {
    const { result } = renderHook(() =>
      useVirtualList({ itemCount: 100000, itemHeight: 30, containerHeight: 600, overscan: 5 })
    );
    expect(result.current.totalHeight).toBe(3000000);
    // 只渲染少量元素
    expect(result.current.visibleIndices.length).toBeLessThan(50);

    act(() => {
      result.current.onScroll({
        currentTarget: { scrollTop: 50000 },
      } as unknown as React.UIEvent<HTMLElement>);
    });

    expect(result.current.startIndex).toBeGreaterThan(0);
    expect(result.current.visibleIndices.length).toBeLessThan(50);
  });
});
