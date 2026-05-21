/**
 * Global Type Extensions
 * Extends built-in TypeScript interfaces for browser APIs
 */

// Network Information API types
interface NetworkInformation extends EventTarget {
  effectiveType: 'slow-2g' | '2g' | '3g' | '4g';
  downlink: number;
  rtt: number;
  saveData: boolean;
  onchange?: EventListener;
  addEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void;
  removeEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions): void;
}

declare global {
  // Extend Navigator interface for Network Information API
  interface Navigator {
    connection?: NetworkInformation;
    mozConnection?: NetworkInformation;
    webkitConnection?: NetworkInformation;
    deviceMemory?: number;
    vibrate?: (pattern: number | number[]) => boolean;
  }

  // Extend CSSStyleDeclaration for webkit-specific properties
  interface CSSStyleDeclaration {
    webkitOverflowScrolling?: 'auto' | 'touch';
  }

  // Extend Window interface
  interface Window {
    webkitRequestAnimationFrame?: (callback: FrameRequestCallback) => number;
  }

  // NodeJS namespace for setTimeout return type
  namespace NodeJS {
    interface Timeout {
      ref(): Timeout;
      unref(): Timeout;
      hasRef(): boolean;
    }
  }
}

export {};