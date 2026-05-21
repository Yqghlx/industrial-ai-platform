import React, { useEffect, createContext, useContext, useMemo } from 'react';
import {
  isMobileDevice,
  isIOS,
  isTouchDevice,
  isLowEndDevice,
  shouldReduceMotion,
  getConnectionInfo,
  isSlowNetwork,
  getOptimalImageQuality,
} from '../lib/mobileOptimizations';

// Network Information API type
interface NetworkInformation extends EventTarget {
  effectiveType: 'slow-2g' | '2g' | '3g' | '4g';
  downlink: number;
  rtt: number;
  saveData: boolean;
  onchange?: EventListener;
}

interface MobileContextType {
  isMobile: boolean;
  isIOS: boolean;
  isTouch: boolean;
  isLowEnd: boolean;
  reduceMotion: boolean;
  isSlowNetwork: boolean;
  imageQuality: 'low' | 'medium' | 'high';
  connectionInfo: {
    effectiveType: string;
    downlink: number;
    saveData: boolean;
  };
}

const MobileContext = createContext<MobileContextType>({
  isMobile: false,
  isIOS: false,
  isTouch: false,
  isLowEnd: false,
  reduceMotion: false,
  isSlowNetwork: false,
  imageQuality: 'medium',
  connectionInfo: {
    effectiveType: '4g',
    downlink: 10,
    saveData: false,
  },
});

export function useMobileContext(): MobileContextType {
  return useContext(MobileContext);
}

interface MobileProviderProps {
  children: React.ReactNode;
}

export function MobileProvider({ children }: MobileProviderProps) {
  const contextValue = useMemo<MobileContextType>(() => {
    return {
      isMobile: isMobileDevice(),
      isIOS: isIOS(),
      isTouch: isTouchDevice(),
      isLowEnd: isLowEndDevice(),
      reduceMotion: shouldReduceMotion(),
      isSlowNetwork: isSlowNetwork(),
      imageQuality: getOptimalImageQuality(),
      connectionInfo: getConnectionInfo(),
    };
  }, []);

  // 设置 viewport 高度变量 (iOS Safari)
  useEffect(() => {
    const setVH = () => {
      const vh = window.innerHeight * 0.01;
      document.documentElement.style.setProperty('--vh', `${vh}px`);
    };

    setVH();
    
    // 监听 resize 和 orientationchange
    window.addEventListener('resize', setVH);
    window.addEventListener('orientationchange', setVH);
    
    return () => {
      window.removeEventListener('resize', setVH);
      window.removeEventListener('orientationchange', setVH);
    };
  }, []);

  // 设置设备类型类
  useEffect(() => {
    const { isMobile, isIOS, isTouch, isLowEnd, reduceMotion } = contextValue;
    
    const html = document.documentElement;
    
    if (isMobile) html.classList.add('is-mobile');
    if (!isMobile) html.classList.add('is-desktop');
    if (isIOS) html.classList.add('is-ios');
    if (isTouch) html.classList.add('is-touch');
    if (isLowEnd) html.classList.add('is-low-end');
    if (reduceMotion) html.classList.add('reduce-motion');
    
    return () => {
      html.classList.remove('is-mobile', 'is-desktop', 'is-ios', 'is-touch', 'is-low-end', 'reduce-motion');
    };
  }, [contextValue]);

  // 监听网络变化
  useEffect(() => {
    const handleConnectionChange = () => {
      // 可以在这里触发重新渲染或通知用户
    };

const connection = (navigator as Navigator & {
  connection?: NetworkInformation;
  mozConnection?: NetworkInformation;
  webkitConnection?: NetworkInformation;
}).connection || (navigator as Navigator & {
  connection?: NetworkInformation;
  mozConnection?: NetworkInformation;
  webkitConnection?: NetworkInformation;
}).mozConnection || (navigator as Navigator & {
  connection?: NetworkInformation;
  mozConnection?: NetworkInformation;
  webkitConnection?: NetworkInformation;
}).webkitConnection;
    
    if (connection) {
      connection.addEventListener('change', handleConnectionChange);
      return () => connection.removeEventListener('change', handleConnectionChange);
    }
  }, []);

  // 防止 iOS 双击缩放
  useEffect(() => {
    if (!contextValue.isIOS) return;

    const preventZoom = (e: TouchEvent) => {
      if (e.touches.length > 1) {
        e.preventDefault();
      }
    };

    document.addEventListener('touchstart', preventZoom, { passive: false });
    
    return () => {
      document.removeEventListener('touchstart', preventZoom);
    };
  }, [contextValue.isIOS]);

  return (
    <MobileContext.Provider value={contextValue}>
      {children}
    </MobileContext.Provider>
  );
}

// 条件渲染组件 - 只在移动端显示
export function MobileOnly({ children }: { children: React.ReactNode }) {
  const { isMobile } = useMobileContext();
  return isMobile ? <>{children}</> : null;
}

// 条件渲染组件 - 只在桌面端显示
export function DesktopOnly({ children }: { children: React.ReactNode }) {
  const { isMobile } = useMobileContext();
  return !isMobile ? <>{children}</> : null;
}

// 条件渲染组件 - 根据设备性能
export function LowEndFallback({
  lowEnd: lowEndContent,
  normal: normalContent,
}: {
  lowEnd: React.ReactNode;
  normal: React.ReactNode;
}) {
  const { isLowEnd } = useMobileContext();
  return isLowEnd ? <>{lowEndContent}</> : <>{normalContent}</>;
}

// 条件渲染组件 - 根据网络速度
export function SlowNetworkFallback({
  slow: slowContent,
  fast: fastContent,
}: {
  slow: React.ReactNode;
  fast: React.ReactNode;
}) {
  const { isSlowNetwork } = useMobileContext();
  return isSlowNetwork ? <>{slowContent}</> : <>{fastContent}</>;
}

export default MobileProvider;