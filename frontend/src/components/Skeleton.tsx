import React from 'react';

interface SkeletonProps {
  variant?: 'text' | 'card' | 'circle' | 'chart' | 'list' | 'avatar' | 'button';
  width?: string | number;
  height?: string | number;
  className?: string;
  lines?: number;
}

export default function Skeleton({ 
  variant = 'text', 
  width, 
  height, 
  className = '',
  lines = 1,
}: SkeletonProps) {
  const baseClasses = 'animate-pulse bg-slate-700 rounded';
  
  const variantClasses = {
    text: 'h-4 w-full',
    card: 'h-24 lg:h-32 w-full rounded-lg',
    circle: 'h-10 w-10 lg:h-12 lg:w-12 rounded-full',
    chart: 'h-40 lg:h-48 w-full rounded-lg',
    list: 'h-16 w-full rounded-lg',
    avatar: 'h-8 w-8 lg:h-10 lg:w-10 rounded-full',
    button: 'h-9 lg:h-10 w-20 lg:w-24 rounded-md',
  };

  const style = {
    width: width ? (typeof width === 'number' ? `${width}px` : width) : undefined,
    height: height ? (typeof height === 'number' ? `${height}px` : height) : undefined,
  };

  if (variant === 'text' && lines > 1) {
    return (
      <div className={`space-y-2 ${className}`}>
        {Array.from({ length: lines }).map((_, i) => (
          <div
            key={i}
            className={`${baseClasses} ${variantClasses.text} ${i === lines - 1 ? 'w-3/4' : ''}`}
          />
        ))}
      </div>
    );
  }

  return (
    <div
      className={`${baseClasses} ${variantClasses[variant]} ${className}`}
      style={style}
    />
  );
}

// Skeleton grid for dashboard - responsive
export function SkeletonGrid({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-3 lg:gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <Skeleton key={i} variant="card" />
      ))}
    </div>
  );
}

// Skeleton for stats cards
export function SkeletonStats({ count = 4 }: { count?: number }) {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="metric-card animate-pulse">
          <div className="flex items-center gap-2 mb-2">
            <Skeleton variant="circle" className="w-5 h-5 lg:w-6 lg:h-6" />
            <Skeleton variant="text" width="40%" height={12} />
          </div>
          <Skeleton variant="text" width="50%" height={24} />
        </div>
      ))}
    </div>
  );
}

// Skeleton table - mobile optimized
export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <div className="table-container overflow-x-auto">
      {/* Mobile: Card list view */}
      <div className="lg:hidden space-y-2 p-3">
        {Array.from({ length: rows }).map((_, rowIndex) => (
          <div key={rowIndex} className="p-3 bg-slate-800/50 rounded-lg border border-slate-700 animate-pulse">
            <div className="flex items-center gap-3 mb-2">
              <Skeleton variant="circle" />
              <div className="flex-1 space-y-1">
                <Skeleton variant="text" width="60%" height={14} />
                <Skeleton variant="text" width="40%" height={12} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-2">
              <Skeleton variant="text" height={12} />
              <Skeleton variant="text" height={12} />
            </div>
          </div>
        ))}
      </div>
      
      {/* Desktop: Table view */}
      <table className="table hidden lg:table">
        <thead>
          <tr>
            {Array.from({ length: 4 }).map((_, i) => (
              <th key={i}><Skeleton variant="text" width={100} /></th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rows }).map((_, rowIndex) => (
            <tr key={rowIndex}>
              {Array.from({ length: 4 }).map((_, colIndex) => (
                <td key={colIndex}><Skeleton variant="text" /></td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// Skeleton for list items
export function SkeletonList({ count = 5 }: { count?: number }) {
  return (
    <div className="space-y-2 lg:space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="list-item animate-pulse">
          <div className="flex items-center gap-3">
            <Skeleton variant="avatar" />
            <div className="flex-1 space-y-1">
              <Skeleton variant="text" width="70%" height={14} />
              <Skeleton variant="text" width="50%" height={12} />
            </div>
            <Skeleton variant="button" />
          </div>
        </div>
      ))}
    </div>
  );
}

// Skeleton for device cards
export function SkeletonDeviceCards({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 lg:gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="p-3 lg:p-4 bg-slate-800/50 rounded-lg border border-slate-700 animate-pulse">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-slate-600" />
              <Skeleton variant="text" width={80} height={16} />
            </div>
            <Skeleton variant="text" width={60} height={20} className="rounded-full" />
          </div>
          <div className="grid grid-cols-2 gap-2 mb-2">
            <Skeleton variant="text" height={12} />
            <Skeleton variant="text" height={12} />
          </div>
          <Skeleton variant="text" width="60%" height={12} />
        </div>
      ))}
    </div>
  );
}

// Skeleton for page header
export function SkeletonPageHeader() {
  return (
    <div className="animate-pulse space-y-2 mb-4 lg:mb-6">
      <Skeleton variant="text" width="40%" height={28} className="rounded-lg" />
      <Skeleton variant="text" width="60%" height={16} />
    </div>
  );
}

// Skeleton for charts/visualizations
export function SkeletonChart() {
  return (
    <div className="animate-pulse">
      <div className="h-8 lg:h-10 bg-slate-700 rounded mb-4" />
      <div className="h-40 lg:h-64 bg-slate-700/50 rounded-lg relative overflow-hidden">
        {/* Fake chart bars */}
        <div className="absolute bottom-0 left-0 right-0 flex items-end justify-around h-full px-2">
          {[40, 65, 45, 80, 55, 70, 50, 60].map((h, i) => (
            <div 
              key={i} 
              className="w-6 lg:w-8 bg-slate-600 rounded-t"
              style={{ height: `${h}%` }}
            />
          ))}
        </div>
      </div>
    </div>
  );
}