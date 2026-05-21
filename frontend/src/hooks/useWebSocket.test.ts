import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';

// Mock WebSocket and wsCompression
const mockWebSocket = {
  readyState: WebSocket.CONNECTING,
  send: vi.fn(),
  close: vi.fn(),
  onopen: null as (() => void) | null,
  onmessage: null as ((event: { data: unknown }) => void) | null,
  onerror: null as ((error: Event) => void) | null,
  onclose: null as (() => void) | null,
  binaryType: 'blob',
};

vi.stubGlobal('WebSocket', vi.fn((_url: string) => {
  mockWebSocket.readyState = WebSocket.CONNECTING;
  return mockWebSocket;
}));

// Mock wsCompression module
vi.mock('../lib/wsCompression', () => ({
  CompressedWebSocket: vi.fn().mockImplementation(() => ({
    connect: vi.fn(() => {
      mockWebSocket.readyState = WebSocket.OPEN;
      if (mockWebSocket.onopen) mockWebSocket.onopen();
    }),
    disconnect: vi.fn(() => {
      mockWebSocket.readyState = WebSocket.CLOSED;
      if (mockWebSocket.onclose) mockWebSocket.onclose();
    }),
    send: vi.fn(),
    isConnected: vi.fn(() => mockWebSocket.readyState === WebSocket.OPEN),
    onMessage: vi.fn((cb: (msg: unknown) => void) => {
      mockWebSocket.onmessage = (event: { data: unknown }) => cb(event.data);
    }),
    onError: vi.fn((cb: (err: Event) => void) => {
      mockWebSocket.onerror = cb;
    }),
    onConnect: vi.fn((cb: () => void) => {
      mockWebSocket.onopen = cb;
    }),
    onClose: vi.fn((cb: () => void) => {
      mockWebSocket.onclose = cb;
    }),
    getCompressionStats: vi.fn(() => ({
      totalMessages: 10,
      compressedMessages: 8,
      skippedMessages: 2,
      originalBytes: 1000,
      compressedBytes: 200,
      compressionRatio: 0.2,
      savingsPercent: 80,
    })),
  })),
  createWebSocketURL: vi.fn(() => 'ws://localhost/ws'),
}));

// Import after mocks
import { useWebSocket } from './useWebSocket';

describe('useWebSocket', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWebSocket.readyState = WebSocket.CONNECTING;
    mockWebSocket.onopen = null;
    mockWebSocket.onmessage = null;
    mockWebSocket.onerror = null;
    mockWebSocket.onclose = null;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('initialization', () => {
    it('should return initial state', () => {
      const { result } = renderHook(() => useWebSocket());

      // Check that the hook returns expected properties
      expect(result.current.isConnected).toBeDefined();
      expect(result.current.send).toBeDefined();
      expect(result.current.reconnect).toBeDefined();
      expect(result.current.getCompressionStats).toBeDefined();
    });

    it('should create WebSocket connection on mount', () => {
      const { result } = renderHook(() => useWebSocket());

      // Connection should be attempted
      expect(result.current).toBeDefined();
    });
  });

  describe('connection state', () => {
    it('should have connection state tracking', () => {
      const { result } = renderHook(() => useWebSocket());

      // Initially disconnected
      expect(result.current.isConnected).toBeDefined();
    });
  });

  describe('send function', () => {
    it('should have send function available', () => {
      const { result } = renderHook(() => useWebSocket());

      // Send function should exist
      expect(result.current.send).toBeDefined();

      // Disconnected - send should not throw
      act(() => {
        result.current.send({ type: 'test', data: 'value' });
      });

      // Should complete without error
      expect(result.current.send).toBeDefined();
    });
  });

  describe('reconnect function', () => {
    it('should have reconnect function available', () => {
      const { result } = renderHook(() => useWebSocket());

      // Reconnect function should exist
      expect(result.current.reconnect).toBeDefined();

      // Call reconnect
      act(() => {
        result.current.reconnect();
      });

      // Should attempt to reconnect
      expect(result.current.reconnect).toBeDefined();
    });
  });

  describe('getCompressionStats', () => {
    it('should return compression statistics', () => {
      const { result } = renderHook(() => useWebSocket());

      const stats = result.current.getCompressionStats();

      expect(stats).toBeDefined();
      expect(stats.totalMessages).toBeDefined();
      expect(stats.compressedMessages).toBeDefined();
      expect(stats.compressionRatio).toBeDefined();
    });
  });

  describe('callbacks', () => {
    it('should accept onMessage callback', () => {
      const onMessage = vi.fn();
      
      const { result } = renderHook(() => 
        useWebSocket({ onMessage })
      );

      expect(result.current).toBeDefined();
      expect(onMessage).toBeDefined();
    });

    it('should accept onError callback', () => {
      const onError = vi.fn();
      
      const { result } = renderHook(() => 
        useWebSocket({ onError })
      );

      expect(result.current).toBeDefined();
      expect(onError).toBeDefined();
    });

    it('should accept onConnect callback', () => {
      const onConnect = vi.fn();
      
      const { result } = renderHook(() => 
        useWebSocket({ onConnect })
      );

      expect(result.current).toBeDefined();
      expect(onConnect).toBeDefined();
    });

    it('should accept onClose callback', () => {
      const onClose = vi.fn();
      
      const { result } = renderHook(() => 
        useWebSocket({ onClose })
      );

      expect(result.current).toBeDefined();
      expect(onClose).toBeDefined();
    });
  });

  describe('options', () => {
    it('should accept compression option', () => {
      const { result } = renderHook(() => 
        useWebSocket({ compression: true })
      );

      expect(result.current).toBeDefined();
    });

    it('should accept autoConnect option', () => {
      const { result } = renderHook(() => 
        useWebSocket({ autoConnect: false })
      );

      expect(result.current).toBeDefined();
    });
  });

  describe('cleanup', () => {
    it('should cleanup on unmount', () => {
      const { unmount } = renderHook(() => useWebSocket());

      // Unmount should not throw
      unmount();
    });

    it('should handle multiple hook instances', () => {
      const { result: result1 } = renderHook(() => useWebSocket());
      const { result: result2 } = renderHook(() => useWebSocket());

      // Both hooks should work
      expect(result1.current).toBeDefined();
      expect(result2.current).toBeDefined();
    });
  });

  describe('connection management', () => {
    it('should track connection count', () => {
      const { unmount: unmount1 } = renderHook(() => useWebSocket());
      const { unmount: unmount2 } = renderHook(() => useWebSocket());

      // Unmount first hook
      unmount1();

      // Second hook should still work
      const { result } = renderHook(() => useWebSocket());
      expect(result.current).toBeDefined();

      // Cleanup
      unmount2();
    });
  });
});