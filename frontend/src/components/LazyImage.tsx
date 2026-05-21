import React, { useState, useRef, useEffect } from 'react';

interface LazyImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  src: string;
  alt: string;
  placeholder?: string;
  fallback?: string;
  threshold?: number; // Intersection Observer 阈值
  rootMargin?: string; // 提前加载距离
  onLoad?: () => void;
  onError?: () => void;
  // Mobile optimization props
  srcSet?: string; // Responsive images
  sizes?: string; // Image sizes for srcset
  aspectRatio?: string; // Maintain aspect ratio (e.g., "16/9", "1/1")
  lowQualityPlaceholder?: boolean; // Use low quality image placeholder
}

/**
 * 懒加载图片组件
 * 使用 Intersection Observer 实现图片懒加载
 * 支持移动端优化和响应式图片
 */
export function LazyImage({
  src,
  alt,
  placeholder,
  fallback,
  threshold = 0.1,
  rootMargin = '100px', // 增加提前加载距离以改善移动端体验
  onLoad,
  onError,
  srcSet,
  sizes,
  aspectRatio,
  lowQualityPlaceholder = true,
  className = '',
  style,
  ...props
}: LazyImageProps) {
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);
  const [hasError, setHasError] = useState(false);
  const imgRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    const img = imgRef.current;
    if (!img) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true);
          observer.disconnect();
        }
      },
      {
        threshold,
        rootMargin,
      }
    );

    observer.observe(img);

    return () => {
      observer.disconnect();
    };
  }, [threshold, rootMargin]);

  const handleLoad = () => {
    setIsLoaded(true);
    onLoad?.();
  };

  const handleError = () => {
    setHasError(true);
    onError?.();
  };

  const currentSrc = hasError && fallback 
    ? fallback 
    : isInView 
      ? src 
      : placeholder || '';

  // 容器样式，支持固定宽高比
  const containerStyle: React.CSSProperties = {
    ...style,
    aspectRatio: aspectRatio,
  };

  return (
    <div 
      className={`relative overflow-hidden ${aspectRatio ? 'w-full' : ''} ${className}`} 
      style={containerStyle}
    >
      {/* 占位符 */}
      {(!isLoaded || !isInView) && !hasError && (
        <div 
          className={`
            absolute inset-0 bg-slate-700 rounded
            ${lowQualityPlaceholder ? 'animate-pulse' : ''}
          `}
          style={{ minHeight: aspectRatio ? 'auto' : '100px' }}
        />
      )}
      
      <img
        ref={imgRef}
        src={currentSrc}
        srcSet={isInView ? srcSet : undefined}
        sizes={sizes}
        alt={alt}
        onLoad={handleLoad}
        onError={handleError}
        className={`
          transition-opacity duration-300 w-full h-full object-cover
          ${isLoaded || hasError ? 'opacity-100' : 'opacity-0'}
        `}
        loading="lazy" // 原生懒加载作为后备
        decoding="async" // 异步解码，不阻塞渲染
        {...props}
      />
      
      {/* 错误状态 */}
      {hasError && fallback && (
        <div className="absolute inset-0 flex items-center justify-center bg-slate-700 rounded">
          <span className="text-slate-400 text-sm">图片加载失败</span>
        </div>
      )}
    </div>
  );
}

/**
 * 懒加载背景图组件
 */
interface LazyBackgroundProps {
  src: string;
  children?: React.ReactNode;
  className?: string;
  placeholder?: string;
  threshold?: number;
  rootMargin?: string;
}

export function LazyBackground({
  src,
  children,
  className = '',
  placeholder,
  threshold = 0.1,
  rootMargin = '50px',
}: LazyBackgroundProps) {
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true);
          observer.disconnect();
        }
      },
      {
        threshold,
        rootMargin,
      }
    );

    observer.observe(container);

    return () => {
      observer.disconnect();
    };
  }, [threshold, rootMargin]);

  useEffect(() => {
    if (!isInView || !src) return;

    const img = new Image();
    img.src = src;
    img.onload = () => setIsLoaded(true);
  }, [isInView, src]);

  const backgroundStyle: React.CSSProperties = {
    backgroundImage: isLoaded ? `url(${src})` : placeholder ? `url(${placeholder})` : 'none',
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    transition: 'background-image 0.3s ease',
  };

  return (
    <div 
      ref={containerRef} 
      className={className} 
      style={backgroundStyle}
    >
      {!isLoaded && isInView && (
        <div className="absolute inset-0 bg-slate-700 animate-pulse" />
      )}
      {children}
    </div>
  );
}

/**
 * 图片预加载工具
 */
export function preloadImages(urls: string[]): Promise<void[]> {
  return Promise.all(
    urls.map(
      (url) =>
        new Promise<void>((resolve, reject) => {
          const img = new Image();
          img.src = url;
          img.onload = () => resolve();
          img.onerror = () => reject(new Error(`Failed to load: ${url}`));
        })
    )
  );
}

export default LazyImage;