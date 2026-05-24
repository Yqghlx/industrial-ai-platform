/**
 * 性能监控工具
 * 收集和报告前端性能指标
 */
import React from 'react';

// Layout Shift Entry 类型扩展
interface LayoutShiftEntry extends PerformanceEntry {
  value: number;
  hadRecentInput: boolean;
}

// 类型守卫：检查是否为 LayoutShift Entry
function isLayoutShiftEntry(entry: PerformanceEntry): entry is LayoutShiftEntry {
  return entry.entryType === 'layout-shift' && 
    'value' in entry && 
    'hadRecentInput' in entry;
}

// 性能指标类型
export interface PerformanceMetrics {
  // 首次内容绘制
  fcp: number;
  // 最大内容绘制
  lcp: number;
  // 首次输入延迟
  fid: number;
  // 累积布局偏移
  cls: number;
  // 交互时间
  tti: number;
  // 首字节时间
  ttfb: number;
  // DOM 内容加载时间
  domContentLoaded: number;
  // 页面完全加载时间
  loadComplete: number;
  // JavaScript 堆大小
  jsHeapSize?: number;
  // 组件渲染时间
  componentRenderTimes: Record<string, number>;
}

// 组件性能数据
export interface ComponentPerformance {
  name: string;
  renderTime: number;
  mountTime: number;
  updateTime?: number;
  renderCount: number;
}

// 性能观察者
type MetricCallback = (metrics: PerformanceMetrics) => void;

class PerformanceMonitor {
  private metrics: Partial<PerformanceMetrics> = {
    componentRenderTimes: {},
  };
  private observers: PerformanceObserver[] = [];
  private callbacks: MetricCallback[] = [];
  private componentMetrics: Map<string, ComponentPerformance> = new Map();

  constructor() {
    if (typeof window !== 'undefined') {
      this.initObservers();
      this.collectNavigationTiming();
    }
  }

  /**
   * 初始化性能观察者
   */
  private initObservers() {
    // 观察绘制性能
    if ('PerformanceObserver' in window) {
      try {
        const paintObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.name === 'first-contentful-paint') {
              this.metrics.fcp = entry.startTime;
            }
          }
          this.notifyCallbacks();
        });
        paintObserver.observe({ type: 'paint', buffered: true });
        this.observers.push(paintObserver);
      } catch (e) {
        console.warn('Paint observer not supported');
      }

      // 观察最大内容绘制
      try {
        const lcpObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries();
          const lastEntry = entries[entries.length - 1];
          this.metrics.lcp = lastEntry.startTime;
          this.notifyCallbacks();
        });
        lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });
        this.observers.push(lcpObserver);
      } catch (e) {
        console.warn('LCP observer not supported');
      }

      // 观察首次输入延迟
      try {
        const fidObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.entryType === 'first-input') {
              this.metrics.fid = (entry as PerformanceEventTiming).processingStart - entry.startTime;
            }
          }
          this.notifyCallbacks();
        });
        fidObserver.observe({ type: 'first-input', buffered: true });
        this.observers.push(fidObserver);
      } catch (e) {
        console.warn('FID observer not supported');
      }

      // 观察累积布局偏移
      try {
        let clsValue = 0;
        const clsObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (isLayoutShiftEntry(entry) && !entry.hadRecentInput) {
              clsValue += entry.value;
            }
          }
          this.metrics.cls = clsValue;
          this.notifyCallbacks();
        });
        clsObserver.observe({ type: 'layout-shift', buffered: true });
        this.observers.push(clsObserver);
      } catch (e) {
        console.warn('CLS observer not supported');
      }

      // 观察长任务
      try {
        const longTaskObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            console.warn('Long task detected:', entry.duration, 'ms');
          }
        });
        longTaskObserver.observe({ type: 'longtask', buffered: true });
        this.observers.push(longTaskObserver);
      } catch (e) {
        console.warn('Long task observer not supported');
      }
    }
  }

  /**
   * 收集导航计时信息
   */
  private collectNavigationTiming() {
    window.addEventListener('load', () => {
      setTimeout(() => {
        const timing = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        if (timing) {
          this.metrics.ttfb = timing.responseStart - timing.requestStart;
          this.metrics.domContentLoaded = timing.domContentLoadedEventEnd - timing.startTime;
          this.metrics.loadComplete = timing.loadEventEnd - timing.startTime;
          this.metrics.tti = timing.domInteractive - timing.startTime;
        }

        // 内存使用情况
        if ('memory' in performance) {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const memory = (performance as any).memory;
          this.metrics.jsHeapSize = memory.usedJSHeapSize;
        }

        this.notifyCallbacks();
      }, 0);
    });
  }

  /**
   * 注册回调函数
   */
  onMetrics(callback: MetricCallback) {
    this.callbacks.push(callback);
    // 立即发送当前指标
    if (Object.keys(this.metrics).length > 0) {
      callback(this.getMetrics());
    }
    return () => {
      this.callbacks = this.callbacks.filter(cb => cb !== callback);
    };
  }

  /**
   * 通知所有回调
   */
  private notifyCallbacks() {
    const metrics = this.getMetrics();
    this.callbacks.forEach(cb => cb(metrics));
  }

  /**
   * 获取当前指标
   */
  getMetrics(): PerformanceMetrics {
    return this.metrics as PerformanceMetrics;
  }

  /**
   * 记录组件渲染时间
   */
  recordComponentRender(name: string, renderTime: number, type: 'mount' | 'update' = 'mount') {
    const existing = this.componentMetrics.get(name);
    
    if (existing) {
      existing.renderTime = renderTime;
      existing.renderCount++;
      if (type === 'update') {
        existing.updateTime = renderTime;
      }
    } else {
      this.componentMetrics.set(name, {
        name,
        renderTime,
        mountTime: renderTime,
        renderCount: 1,
      });
    }

    // 更新总指标
    this.metrics.componentRenderTimes![name] = renderTime;
  }

  /**
   * 获取组件性能数据
   */
  getComponentMetrics(): ComponentPerformance[] {
    return Array.from(this.componentMetrics.values());
  }

  /**
   * 开始计时
   */
  startMeasure(name: string): () => number {
    const start = performance.now();
    return () => {
      const duration = performance.now() - start;
      this.recordComponentRender(name, duration);
      return duration;
    };
  }

  /**
   * 使用 Performance API 标记
   */
  mark(name: string) {
    performance.mark(name);
  }

  /**
   * 测量两个标记之间的时间
   */
  measure(name: string, startMark: string, endMark: string): number {
    try {
      performance.measure(name, startMark, endMark);
      const entries = performance.getEntriesByName(name, 'measure');
      return entries[entries.length - 1]?.duration || 0;
    } catch (e) {
      console.warn('Measure failed:', e);
      return 0;
    }
  }

  /**
   * 清除标记和测量
   */
  clearMarks() {
    performance.clearMarks();
    performance.clearMeasures();
  }

  /**
   * 销毁观察者
   */
  destroy() {
    this.observers.forEach(observer => observer.disconnect());
    this.observers = [];
    this.callbacks = [];
  }
}

// 单例实例
export const performanceMonitor = new PerformanceMonitor();

/**
 * React Hook: 用于测量组件性能
 */
export function usePerformance(componentName: string) {
  const renderStartTime = React.useRef(0);
  const mountTime = React.useRef(0);

  // 记录渲染开始时间
  renderStartTime.current = performance.now();

  React.useEffect(() => {
    const renderTime = performance.now() - renderStartTime.current;
    
    if (mountTime.current === 0) {
      mountTime.current = renderTime;
      performanceMonitor.recordComponentRender(componentName, renderTime, 'mount');
    } else {
      performanceMonitor.recordComponentRender(componentName, renderTime, 'update');
    }
  });
}

/**
 * 高阶组件：用于测量组件性能
 */
export function withPerformance<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  componentName: string
): React.FC<P> {
  return function PerformanceWrapper(props: P) {
    usePerformance(componentName);
    return <WrappedComponent {...props} />;
  };
}

/**
 * 报告性能指标到服务器
 */
export async function reportMetrics(endpoint: string, metrics: PerformanceMetrics) {
  try {
    await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ...metrics,
        timestamp: new Date().toISOString(),
        url: window.location.href,
        userAgent: navigator.userAgent,
      }),
    });
  } catch (error) {
    console.error('Failed to report metrics:', error);
  }
}

/**
 * 导出默认实例
 */
export default performanceMonitor;