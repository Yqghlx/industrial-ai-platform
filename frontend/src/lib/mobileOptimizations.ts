/**
 * 移动端性能优化工具
 */

// Network Information API type
interface NetworkInformation extends EventTarget {
  effectiveType: 'slow-2g' | '2g' | '3g' | '4g';
  downlink: number;
  rtt: number;
  saveData: boolean;
  onchange?: EventListener;
}

// Extend Navigator for network info
declare global {
  interface Navigator {
    connection?: NetworkInformation;
    mozConnection?: NetworkInformation;
    webkitConnection?: NetworkInformation;
    deviceMemory?: number;
  }
}

// 检测移动设备
export function isMobileDevice(): boolean {
  if (typeof window === 'undefined') return false;
  
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent
  ) || window.innerWidth < 768;
}

// 检测 iOS 设备
export function isIOS(): boolean {
  if (typeof window === 'undefined') return false;
  
  return /iPad|iPhone|iPod/.test(navigator.userAgent) ||
    (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);
}

// 检测是否支持触摸
export function isTouchDevice(): boolean {
  if (typeof window === 'undefined') return false;
  
  return (
    'ontouchstart' in window ||
    navigator.maxTouchPoints > 0
  );
}

// 检测是否为低端设备
export function isLowEndDevice(): boolean {
  if (typeof navigator === 'undefined') return false;
  
  const cores = navigator.hardwareConcurrency ?? 2;
  const memory = navigator.deviceMemory ?? 4;
  
  return cores <= 2 || memory <= 2;
}

// 检测网络状态
export function getConnectionInfo(): {
  effectiveType: string;
  downlink: number;
  saveData: boolean;
} {
  const connection = (navigator.connection ?? navigator.mozConnection ?? navigator.webkitConnection) as NetworkInformation | undefined;
  
  if (connection) {
    return {
      effectiveType: connection.effectiveType ?? '4g',
      downlink: connection.downlink ?? 10,
      saveData: connection.saveData ?? false,
    };
  }
  
  return {
    effectiveType: '4g',
    downlink: 10,
    saveData: false,
  };
}

// 是否应该减少动画
export function shouldReduceMotion(): boolean {
  if (typeof window === 'undefined') return false;
  
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// 是否为慢速网络
export function isSlowNetwork(): boolean {
  const { effectiveType, saveData } = getConnectionInfo();
  return effectiveType === 'slow-2g' || effectiveType === '2g' || saveData;
}

// 防抖函数
export function debounce<T extends (...args: unknown[]) => unknown>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;
  
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

// 节流函数
export function throttle<T extends (...args: unknown[]) => unknown>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle = false;
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
}

// 请求空闲回调的 polyfill
export const requestIdleCallback =
  typeof window !== 'undefined' && 'requestIdleCallback' in window
    ? window.requestIdleCallback
    : (cb: IdleRequestCallback) => setTimeout(() => cb({ didTimeout: false, timeRemaining: () => 50 } as IdleDeadline), 1);

// 取消空闲回调
export const cancelIdleCallback =
  typeof window !== 'undefined' && 'cancelIdleCallback' in window
    ? window.cancelIdleCallback
    : (id: number) => clearTimeout(id);

// 批量更新优化
export function batchUpdates(callback: () => void): void {
  // 使用 React 18 的自动批处理或手动批处理
  if (typeof requestIdleCallback !== 'undefined') {
    requestIdleCallback(() => callback());
  } else {
    setTimeout(callback, 0);
  }
}

// 图片预加载策略
export function preloadCriticalImages(urls: string[]): void {
  if (isSlowNetwork() || isLowEndDevice()) {
    // 慢速网络或低端设备不预加载
    return;
  }
  
  requestIdleCallback(() => {
    urls.forEach(url => {
      const link = document.createElement('link');
      link.rel = 'preload';
      link.as = 'image';
      link.href = url;
      document.head.appendChild(link);
    });
  });
}

// 获取合适的图片质量
export function getOptimalImageQuality(): 'low' | 'medium' | 'high' {
  if (isSlowNetwork() || isLowEndDevice()) return 'low';
  if (isMobileDevice()) return 'medium';
  return 'high';
}

// 创建响应式图片 URL
export function createResponsiveImageUrl(
  baseUrl: string,
  width: number,
  quality: 'low' | 'medium' | 'high' = 'medium'
): string {
  const qualityMap = {
    low: 50,
    medium: 75,
    high: 90,
  };
  
  // 如果 URL 已经有查询参数
  const separator = baseUrl.includes('?') ? '&' : '?';
  return `${baseUrl}${separator}w=${width}&q=${qualityMap[quality]}`;
}

// 触觉反馈
export function hapticFeedback(type: 'light' | 'medium' | 'heavy' = 'light'): void {
  if (!isTouchDevice()) return;
  
  if ('vibrate' in navigator) {
    const duration = {
      light: 10,
      medium: 20,
      heavy: 30,
    };
    navigator.vibrate?.(duration[type]);
  }
}

// 滚动优化 - 使用 CSS scroll-behavior 和 overscroll-behavior
export function optimizeScrolling(element: HTMLElement): void {
  element.style.webkitOverflowScrolling = 'touch';
  element.style.overscrollBehavior = 'contain';
}

// 内存警告监听
export function onMemoryWarning(callback: () => void): () => void {
  if (!isMobileDevice()) return () => {};
  
  // iOS 内存警告
  const handleMemoryWarning = () => {
    callback();
  };
  
  // 监听页面可见性变化（可能的内存压力）
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
      // 页面隐藏时可以释放一些内存
      handleMemoryWarning();
    }
  });
  
  return () => {
    document.removeEventListener('visibilitychange', handleMemoryWarning);
  };
}

// 获取安全区域 insets
export function getSafeAreaInsets(): {
  top: number;
  right: number;
  bottom: number;
  left: number;
} {
  if (typeof window === 'undefined') {
    return { top: 0, right: 0, bottom: 0, left: 0 };
  }
  
  const style = getComputedStyle(document.documentElement);
  
  return {
    top: parseInt(style.getPropertyValue('--safe-area-top') || '0'),
    right: parseInt(style.getPropertyValue('--safe-area-right') || '0'),
    bottom: parseInt(style.getPropertyValue('--safe-area-bottom') || '0'),
    left: parseInt(style.getPropertyValue('--safe-area-left') || '0'),
  };
}