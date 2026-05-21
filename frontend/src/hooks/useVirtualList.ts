/**
 * useVirtualList - 虚拟列表 Hook
 * 用于高效渲染大型列表，只渲染可视区域内的元素
 */

import { useState, useCallback, useRef, useMemo, useEffect } from 'react';

interface VirtualListOptions {
  /** 列表总数据量 */
  itemCount: number;
  /** 每个项目的估计高度（像素） */
  itemHeight: number;
  /** 容器高度（像素） */
  containerHeight: number;
  /** 缓冲区大小（在可视区域外额外渲染的项目数） */
  overscan?: number;
}

interface VirtualListResult {
  /** 起始索引 */
  startIndex: number;
  /** 结束索引 */
  endIndex: number;
  /** 虚拟列表样式 */
  containerStyle: React.CSSProperties;
  /** 内容区域样式 */
  contentStyle: React.CSSProperties;
  /** 获取项目位置样式 */
  getItemStyle: (index: number) => React.CSSProperties;
  /** 滚动事件处理 */
  onScroll: (event: React.UIEvent<HTMLElement>) => void;
  /** 滚动到指定索引 */
  scrollTo: (index: number) => void;
  /** 滚动到顶部 */
  scrollToTop: () => void;
  /** 滚动到底部 */
  scrollToBottom: () => void;
  /** 当前可视区域内的索引数组 */
  visibleIndices: number[];
  /** 总高度 */
  totalHeight: number;
}

/**
 * 虚拟列表 Hook
 * 
 * @example
 * ```tsx
 * const {
 *   startIndex,
 *   endIndex,
 *   containerStyle,
 *   contentStyle,
 *   getItemStyle,
 *   onScroll,
 * } = useVirtualList({
 *   itemCount: 10000,
 *   itemHeight: 50,
 *   containerHeight: 500,
 *   overscan: 5,
 * });
 * 
 * return (
 *   <div style={containerStyle} onScroll={onScroll}>
 *     <div style={contentStyle}>
 *       {items.slice(startIndex, endIndex + 1).map((item, i) => (
 *         <div key={startIndex + i} style={getItemStyle(startIndex + i)}>
 *           {item.content}
 *         </div>
 *       ))}
 *     </div>
 *   </div>
 * );
 * ```
 */
export function useVirtualList(options: VirtualListOptions): VirtualListResult {
  const { itemCount, itemHeight, containerHeight, overscan = 3 } = options;

  // 滚动位置
  const [scrollTop, setScrollTop] = useState(0);
  
  // 容器引用（用于程序化滚动）
  const containerRef = useRef<HTMLElement | null>(null);

  // 计算总高度 - 使用 useMemo 避免不必要的重新计算
  const totalHeight = useMemo(() => {
    return itemCount * itemHeight;
  }, [itemCount, itemHeight]);

  // 计算起始索引 - 使用 useMemo 优化性能
  const startIndex = useMemo(() => {
    const index = Math.floor(scrollTop / itemHeight);
    return Math.max(0, index - overscan);
  }, [scrollTop, itemHeight, overscan]);

  // 计算结束索引 - 使用 useMemo 优化性能
  const endIndex = useMemo(() => {
    const visibleCount = Math.ceil(containerHeight / itemHeight);
    const index = Math.floor(scrollTop / itemHeight) + visibleCount;
    return Math.min(itemCount - 1, index + overscan);
  }, [scrollTop, itemHeight, containerHeight, itemCount, overscan]);

  // 可视区域内的索引数组 - 使用 useMemo 缓存
  const visibleIndices = useMemo(() => {
    const indices: number[] = [];
    for (let i = startIndex; i <= endIndex; i++) {
      indices.push(i);
    }
    return indices;
  }, [startIndex, endIndex]);

  // 容器样式 - 使用 useMemo 缓存
  const containerStyle: React.CSSProperties = useMemo(() => ({
    overflow: 'auto',
    height: containerHeight,
    position: 'relative' as const,
  }), [containerHeight]);

  // 内容区域样式 - 使用 useMemo 缓存
  const contentStyle: React.CSSProperties = useMemo(() => ({
    height: totalHeight,
    position: 'relative' as const,
  }), [totalHeight]);

  // 获取项目样式 - 使用 useCallback 缓存函数
  const getItemStyle = useCallback((index: number): React.CSSProperties => ({
    position: 'absolute' as const,
    top: index * itemHeight,
    height: itemHeight,
    width: '100%',
  }), [itemHeight]);

  // 滚动事件处理 - 使用 useCallback 缓存函数
  const onScroll = useCallback((event: React.UIEvent<HTMLElement>) => {
    const target = event.currentTarget;
    setScrollTop(target.scrollTop);
    containerRef.current = target;
  }, []);

  // 滚动到指定索引 - 使用 useCallback 缓存函数
  const scrollTo = useCallback((index: number) => {
    if (containerRef.current) {
      const targetTop = Math.max(0, Math.min(index * itemHeight, totalHeight - containerHeight));
      containerRef.current.scrollTop = targetTop;
      setScrollTop(targetTop);
    }
  }, [itemHeight, totalHeight, containerHeight]);

  // 滚动到顶部 - 使用 useCallback 缓存函数
  const scrollToTop = useCallback(() => {
    scrollTo(0);
  }, [scrollTo]);

  // 滚动到底部 - 使用 useCallback 缓存函数
  const scrollToBottom = useCallback(() => {
    scrollTo(itemCount - 1);
  }, [scrollTo, itemCount]);

  // 清理函数
  useEffect(() => {
    return () => {
      containerRef.current = null;
    };
  }, []);

  return {
    startIndex,
    endIndex,
    containerStyle,
    contentStyle,
    getItemStyle,
    onScroll,
    scrollTo,
    scrollToTop,
    scrollToBottom,
    visibleIndices,
    totalHeight,
  };
}

export default useVirtualList;