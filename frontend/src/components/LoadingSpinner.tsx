import React from 'react';
import { useI18n } from '../i18n';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  text?: string;
  fullScreen?: boolean;
}

/**
 * 加载状态组件
 * 支持不同尺寸和全屏模式
 * 移动端优化版本
 */
export default function LoadingSpinner({
  size = 'md',
  text,
  fullScreen = true
}: LoadingSpinnerProps) {
  const { t } = useI18n();
  const displayText = text ?? t('common.loading');
  const sizeClasses = {
    sm: 'w-6 h-6 lg:w-8 lg:h-8',
    md: 'w-12 h-12 lg:w-16 lg:h-16',
    lg: 'w-16 h-16 lg:w-24 lg:h-24',
  };

  const spinner = (
    <div className="relative">
      {/* 外圈旋转 */}
      <div 
        className={`${sizeClasses[size]} border-4 border-primary-500/30 rounded-full animate-spin`}
        style={{ borderTopColor: 'rgb(59, 130, 246)' }}
      />
      {/* 内圈脉冲 */}
      <div className={`absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2`}>
        <div className={`${size === 'sm' ? 'w-2 h-2 lg:w-3 lg:h-3' : size === 'md' ? 'w-4 h-4 lg:w-5 lg:h-5' : 'w-6 h-6 lg:w-8 lg:h-8'} bg-primary-500 rounded-full animate-pulse`} />
      </div>
      {/* 加载文字 */}
      {displayText && (
        <div className={`absolute top-full left-1/2 transform -translate-x-1/2 mt-3 lg:mt-4 whitespace-nowrap`}>
          <span className="text-xs lg:text-sm text-slate-400 animate-pulse">{displayText}</span>
        </div>
      )}
    </div>
  );

  if (fullScreen) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-slate-900 safe-area-top safe-area-bottom">
        {spinner}
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center p-4 lg:p-8">
      {spinner}
    </div>
  );
}

/**
 * 内联加载指示器
 */
export function InlineLoader({ className = '' }: { className?: string }) {
  const { t } = useI18n();
  return (
    <div className={`inline-flex items-center gap-2 ${className}`}>
      <div className="w-4 h-4 border-2 border-primary-500/30 rounded-full animate-spin"
           style={{ borderTopColor: 'rgb(59, 130, 246)' }} />
      <span className="text-sm text-slate-400">{t('common.processing')}</span>
    </div>
  );
}

/**
 * 页面级加载骨架 - 移动端优化
 */
export function PageLoader() {
  return (
    <div className="min-h-screen bg-slate-900 p-3 lg:p-4 space-y-3 lg:space-y-4 safe-area-top safe-area-bottom">
      {/* Header skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="space-y-2">
          <div className="h-6 lg:h-8 w-36 lg:w-48 bg-slate-700 rounded animate-pulse" />
          <div className="h-3 lg:h-4 w-24 lg:w-32 bg-slate-700 rounded animate-pulse" />
        </div>
        <div className="h-9 lg:h-10 w-full sm:w-28 lg:w-32 bg-slate-700 rounded animate-pulse" />
      </div>
      
      {/* Stats skeleton - 2 columns on mobile */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 lg:gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-slate-800 rounded-lg p-3 lg:p-4 space-y-2 lg:space-y-3">
            <div className="h-3 lg:h-4 w-20 lg:w-24 bg-slate-700 rounded animate-pulse" />
            <div className="h-6 lg:h-8 w-12 lg:w-16 bg-slate-700 rounded animate-pulse" />
          </div>
        ))}
      </div>
      
      {/* Content skeleton */}
      <div className="bg-slate-800 rounded-lg p-3 lg:p-4 space-y-2 lg:space-y-3">
        <div className="h-3 lg:h-4 w-full bg-slate-700 rounded animate-pulse" />
        <div className="h-3 lg:h-4 w-3/4 bg-slate-700 rounded animate-pulse" />
        <div className="h-3 lg:h-4 w-1/2 bg-slate-700 rounded animate-pulse" />
      </div>
    </div>
  );
}

/**
 * 路由加载组件 - 用于 Suspense fallback
 */
export function RouteLoader() {
  const { t } = useI18n();
  return (
    <div className="flex items-center justify-center min-h-[50vh]">
      <div className="text-center">
        <div className="inline-block w-12 h-12 border-4 border-primary-500/30 rounded-full animate-spin"
             style={{ borderTopColor: 'rgb(59, 130, 246)' }} />
        <p className="mt-3 text-sm text-slate-400">{t('common.loadingPage')}</p>
      </div>
    </div>
  );
}