/**
 * WebSocket Singleton Hook for Industrial AI Platform
 * Provides a single shared WebSocket connection across all components
 */

import { useEffect, useRef, useCallback, useState } from 'react';
import { CompressedWebSocket, createWebSocketURL, WSMessage } from '../lib/wsCompression';

// Singleton WebSocket instance
let sharedWebSocket: CompressedWebSocket | null = null;
let connectionCount = 0;
const messageCallbacks: Set<(message: WSMessage) => void> = new Set();
const errorCallbacks: Set<(error: Event) => void> = new Set();
const connectCallbacks: Set<() => void> = new Set();
const closeCallbacks: Set<() => void> = new Set();

/**
 * Get or create the shared WebSocket connection
 */
function getSharedWebSocket(): CompressedWebSocket {
  if (!sharedWebSocket) {
    const wsUrl = createWebSocketURL();
    sharedWebSocket = new CompressedWebSocket(wsUrl, true);
    
    // Set up central message handler that dispatches to all subscribers
    sharedWebSocket.onMessage((message: WSMessage) => {
      messageCallbacks.forEach(cb => cb(message));
    });
    
    sharedWebSocket.onError((error: Event) => {
      errorCallbacks.forEach(cb => cb(error));
    });
    
    sharedWebSocket.onConnect(() => {
      connectCallbacks.forEach(cb => cb());
    });
    
    sharedWebSocket.onClose(() => {
      closeCallbacks.forEach(cb => cb());
    });
    
    sharedWebSocket.connect();
  }
  
  return sharedWebSocket;
}

/**
 * Cleanup shared WebSocket when no subscribers remain
 */
function cleanupSharedWebSocket(): void {
  connectionCount--;
  
  if (connectionCount <= 0 && sharedWebSocket) {
    // Delay cleanup to allow for quick navigation between components
    setTimeout(() => {
      if (connectionCount <= 0 && sharedWebSocket) {
        sharedWebSocket.disconnect();
        sharedWebSocket = null;
      }
    }, 5000);
  }
}

/**
 * Hook options
 */
interface UseWebSocketOptions {
  /** Enable compression (default: true) */
  compression?: boolean;
  /** Auto-connect on mount (default: true) */
  autoConnect?: boolean;
}

/**
 * Hook return type
 */
interface UseWebSocketReturn {
  /** Whether WebSocket is connected */
  isConnected: boolean;
  /** Send a message through WebSocket */
  send: (data: unknown) => void;
  /** Manually reconnect */
  reconnect: () => void;
  /** Get compression statistics */
  getCompressionStats: () => {
    totalMessages: number;
    compressedMessages: number;
    skippedMessages: number;
    originalBytes: number;
    compressedBytes: number;
    compressionRatio: number;
    savingsPercent: number;
  };
}

/**
 * useWebSocket Hook
 * Provides a single shared WebSocket connection for all components
 * 
 * @example
 * ```tsx
 * const { isConnected, send } = useWebSocket({
 *   onMessage: (message) => {
 *     if (message.type === 'telemetry') {
 *       console.log('Received telemetry:', message.payload);
 *     }
 *   },
 *   onError: (error) => console.error('WebSocket error:', error),
 * });
 * ```
 */
export function useWebSocket(
  options?: UseWebSocketOptions & {
    onMessage?: (message: WSMessage) => void;
    onError?: (error: Event) => void;
    onConnect?: () => void;
    onClose?: () => void;
  }
): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false);
  const callbacksRef = useRef<{
    onMessage?: (message: WSMessage) => void;
    onError?: (error: Event) => void;
    onConnect?: () => void;
    onClose?: () => void;
  }>(options);
  
  // Update callbacks ref when options change
  callbacksRef.current = options;
  
  useEffect(() => {
    connectionCount++;
    
    const ws = getSharedWebSocket();
    
    // Register component-specific callbacks
    const handleMessage = (message: WSMessage) => {
      if (callbacksRef.current?.onMessage) {
        callbacksRef.current.onMessage(message);
      }
    };
    
    const handleError = (error: Event) => {
      if (callbacksRef.current?.onError) {
        callbacksRef.current.onError(error);
      }
    };
    
    const handleConnect = () => {
      setIsConnected(true);
      if (callbacksRef.current?.onConnect) {
        callbacksRef.current.onConnect();
      }
    };
    
    const handleClose = () => {
      setIsConnected(false);
      if (callbacksRef.current?.onClose) {
        callbacksRef.current.onClose();
      }
    };
    
    messageCallbacks.add(handleMessage);
    errorCallbacks.add(handleError);
    connectCallbacks.add(handleConnect);
    closeCallbacks.add(handleClose);
    
    // Check current connection state
    if (ws.isConnected()) {
      setIsConnected(true);
    }
    
    // Cleanup on unmount
    return () => {
      messageCallbacks.delete(handleMessage);
      errorCallbacks.delete(handleError);
      connectCallbacks.delete(handleConnect);
      closeCallbacks.delete(handleClose);
      cleanupSharedWebSocket();
    };
  }, []); // Empty deps - connection is managed via singleton
  
  const send = useCallback((data: unknown) => {
    if (sharedWebSocket && sharedWebSocket.isConnected()) {
      sharedWebSocket.send(data);
    }
  }, []);
  
  const reconnect = useCallback(() => {
    if (sharedWebSocket) {
      sharedWebSocket.disconnect();
      sharedWebSocket = null;
    }
    connectionCount++;
    const ws = getSharedWebSocket();
    ws.connect();
  }, []);
  
  const getCompressionStats = useCallback(() => {
    if (sharedWebSocket) {
      return sharedWebSocket.getCompressionStats();
    }
    return {
      totalMessages: 0,
      compressedMessages: 0,
      skippedMessages: 0,
      originalBytes: 0,
      compressedBytes: 0,
      compressionRatio: 0,
      savingsPercent: 0,
    };
  }, []);
  
  return {
    isConnected,
    send,
    reconnect,
    getCompressionStats,
  };
}

export default useWebSocket;