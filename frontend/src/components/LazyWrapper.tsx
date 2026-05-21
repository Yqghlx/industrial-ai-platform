import React, { Suspense, ComponentType, LazyExoticComponent } from 'react';
import LoadingSpinner from './LoadingSpinner';
import Skeleton, { SkeletonGrid, SkeletonTable } from './Skeleton';

// 加载状态类型
export type LoadingVariant = 'spinner' | 'card' | 'table' | 'dashboard' | 'minimal';

interface LazyWrapperProps {
  children: React.ReactNode;
  variant?: LoadingVariant;
  delay?: number; // 延迟显示加载状态，避免闪烁
}

// 加载状态组件映射
const LoadingComponents: Record<LoadingVariant, React.FC> = {
  spinner: () => <LoadingSpinner />,
  card: () => (
    <div className="card">
      <div className="card-body">
        <Skeleton variant="card" />
      </div>
    </div>
  ),
  table: () => <SkeletonTable rows={10} />,
  dashboard: () => (
    <div className="space-y-6 p-4">
      <SkeletonGrid count={4} />
      <div className="card">
        <div className="card-body">
          <SkeletonGrid count={6} />
        </div>
      </div>
    </div>
  ),
  minimal: () => (
    <div className="flex items-center justify-center p-4">
      <div className="w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full animate-spin" />
    </div>
  ),
};

/**
 * 懒加载包装器组件
 * 用于包装懒加载组件，提供更好的加载状态
 */
export function LazyWrapper({ 
  children, 
  variant = 'spinner',
  delay = 200 
}: LazyWrapperProps) {
  const [showFallback, setShowFallback] = React.useState(false);

  React.useEffect(() => {
    const timer = setTimeout(() => setShowFallback(true), delay);
    return () => clearTimeout(timer);
  }, [delay]);

  const LoadingComponent = LoadingComponents[variant];

  return (
    <Suspense fallback={showFallback ? <LoadingComponent /> : null}>
      {children}
    </Suspense>
  );
}

/**
 * 创建带预加载功能的懒加载组件
 * @param importFn 动态导入函数
 * @param variant 加载状态类型
 */
export function createLazyComponent<T extends ComponentType<Record<string, unknown>>>(
  importFn: () => Promise<{ default: T }>,
  _variant: LoadingVariant = 'spinner'
): LazyExoticComponent<T> & { preload: () => Promise<void> } {
  const LazyComponent = React.lazy(importFn) as LazyExoticComponent<T>;
  
  // 添加预加载方法
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (LazyComponent as any).preload = importFn;
  
  return LazyComponent as LazyExoticComponent<T> & { preload: () => Promise<void> };
}

/**
 * 预加载路由组件
 * 在用户悬停在链接上时预加载
 */
export function usePreloadRoute(
  lazyComponent: LazyExoticComponent<ComponentType<Record<string, unknown>>> & { preload?: () => Promise<void> }
) {
  const preload = React.useCallback(() => {
    if ('preload' in lazyComponent && typeof lazyComponent.preload === 'function') {
      lazyComponent.preload();
    }
  }, [lazyComponent]);

  return { preload };
}

/**
 * 带预加载功能的 Link 组件
 */
interface PreloadLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  to: string;
  onPreload?: () => void;
  children: React.ReactNode;
}

export const PreloadLink: React.FC<PreloadLinkProps> = ({ 
  to, 
  onPreload, 
  children, 
  ...props 
}) => {
  const [isPreloaded, setIsPreloaded] = React.useState(false);

  const handleMouseEnter = React.useCallback(() => {
    if (!isPreloaded && onPreload) {
      onPreload();
      setIsPreloaded(true);
    }
  }, [isPreloaded, onPreload]);

  return (
    <a 
      href={to} 
      onMouseEnter={handleMouseEnter}
      {...props}
    >
      {children}
    </a>
  );
};

export default LazyWrapper;