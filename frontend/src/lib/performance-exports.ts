/**
 * 性能优化工具导出
 */

// 懒加载相关
export { LazyWrapper, createLazyComponent, usePreloadRoute, PreloadLink } from '../components/LazyWrapper';
export type { LoadingVariant } from '../components/LazyWrapper';

// 图片懒加载
export { LazyImage, LazyBackground, preloadImages } from '../components/LazyImage';

// 加载状态
export { default as LoadingSpinner, InlineLoader, PageLoader, RouteLoader } from '../components/LoadingSpinner';

// 性能监控
export {
  performanceMonitor,
  usePerformance,
  withPerformance,
  reportMetrics,
} from './performance';
export type { PerformanceMetrics, ComponentPerformance } from './performance';

// 性能面板
export { PerformancePanel, PerformanceButton } from '../components/PerformancePanel';

// 缓存工具
export { cache, useCachedAsync, usePreload, cached, getCacheStats } from './cache';

// 性能优化 Hooks
export {
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
} from './hooks';