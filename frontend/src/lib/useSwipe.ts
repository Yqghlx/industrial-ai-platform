import { useRef, useCallback, useState, useEffect } from 'react';

interface SwipeConfig {
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  threshold?: number;
  preventDefaultOnSwipe?: boolean;
}

interface SwipeState {
  startX: number;
  startY: number;
  isSwiping: boolean;
  direction: 'left' | 'right' | null;
}

export function useSwipe(config: SwipeConfig) {
  const { 
    onSwipeLeft, 
    onSwipeRight, 
    preventDefaultOnSwipe = true 
  } = config;
  
  const stateRef = useRef<SwipeState>({
    startX: 0,
    startY: 0,
    isSwiping: false,
    direction: null,
  });

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    const touch = e.touches[0];
    stateRef.current = {
      startX: touch.clientX,
      startY: touch.clientY,
      isSwiping: true,
      direction: null,
    };
  }, []);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!stateRef.current.isSwiping) return;
    
    const touch = e.touches[0];
    const deltaX = touch.clientX - stateRef.current.startX;
    const deltaY = touch.clientY - stateRef.current.startY;
    
    // Determine if horizontal swipe (ignore vertical scrolls)
    if (Math.abs(deltaX) > Math.abs(deltaY)) {
      stateRef.current.direction = deltaX > 0 ? 'right' : 'left';
      
      if (preventDefaultOnSwipe) {
        e.preventDefault();
      }
    }
  }, [preventDefaultOnSwipe]);

  const handleTouchEnd = useCallback(() => {
    if (!stateRef.current.isSwiping) return;
    
    const touch = stateRef.current;
    // Check if swipe meets threshold
    if (touch.direction === 'left' && onSwipeLeft) {
      onSwipeLeft();
    } else if (touch.direction === 'right' && onSwipeRight) {
      onSwipeRight();
    }
    
    stateRef.current.isSwiping = false;
    stateRef.current.direction = null;
  }, [onSwipeLeft, onSwipeRight]);

  return {
    onTouchStart: handleTouchStart,
    onTouchMove: handleTouchMove,
    onTouchEnd: handleTouchEnd,
  };
}

// Hook for detecting mobile devices
export function useIsMobile() {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 1024);
    };

    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  return isMobile;
}

// Hook for detecting iOS
export function useIsIOS() {
  const [isIOS, setIsIOS] = useState(false);

  useEffect(() => {
    const checkIOS = () => {
      return /iPad|iPhone|iPod/.test(navigator.userAgent) ||
        (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);
    };
    setIsIOS(checkIOS());
  }, []);

  return isIOS;
}

// Hook for viewport height (handles iOS Safari address bar)
export function useViewportHeight() {
  useEffect(() => {
    const setVH = () => {
      const vh = window.innerHeight * 0.01;
      document.documentElement.style.setProperty('--vh', `${vh}px`);
    };

    setVH();
    window.addEventListener('resize', setVH);
    window.addEventListener('orientationchange', setVH);
    
    return () => {
      window.removeEventListener('resize', setVH);
      window.removeEventListener('orientationchange', setVH);
    };
  }, []);
}

// Hook for pull-to-refresh
export function usePullToRefresh(onRefresh: () => Promise<void>) {
  const [isRefreshing, setIsRefreshing] = useState(false);
  // startYRef reserved for pull-to-refresh Y-axis tracking
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let startY = 0;
    let pullDistance = 0;
    const threshold = 80;

    const handleTouchStart = (e: TouchEvent) => {
      if (container.scrollTop === 0) {
        startY = e.touches[0].clientY;
      }
    };

    const handleTouchMove = (e: TouchEvent) => {
      if (startY === 0 || isRefreshing) return;
      
      const currentY = e.touches[0].clientY;
      pullDistance = currentY - startY;

      if (pullDistance > 0 && container.scrollTop === 0) {
        container.style.transform = `translateY(${Math.min(pullDistance * 0.5, threshold)}px)`;
      }
    };

    const handleTouchEnd = async () => {
      if (pullDistance >= threshold && !isRefreshing) {
        setIsRefreshing(true);
        await onRefresh();
        setIsRefreshing(false);
      }
      
      container.style.transform = '';
      startY = 0;
      pullDistance = 0;
    };

    container.addEventListener('touchstart', handleTouchStart, { passive: true });
    container.addEventListener('touchmove', handleTouchMove, { passive: true });
    container.addEventListener('touchend', handleTouchEnd);

    return () => {
      container.removeEventListener('touchstart', handleTouchStart);
      container.removeEventListener('touchmove', handleTouchMove);
      container.removeEventListener('touchend', handleTouchEnd);
    };
  }, [onRefresh, isRefreshing]);

  return { containerRef, isRefreshing };
}