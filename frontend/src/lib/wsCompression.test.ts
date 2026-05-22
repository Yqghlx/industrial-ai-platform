import { describe, it, expect, beforeEach } from 'vitest';
import pako from 'pako';
import { WebSocketCompressionHandler, CompressedWebSocket, createWebSocketURL } from './wsCompression';

describe('WebSocketCompressionHandler', () => {
  let handler: WebSocketCompressionHandler;

  beforeEach(() => {
    handler = new WebSocketCompressionHandler(true);
  });

  describe('compression enabled', () => {
    it('should be enabled by default', () => {
      expect(handler.isEnabled()).toBe(true);
    });

    it('should allow toggling', () => {
      handler.setEnabled(false);
      expect(handler.isEnabled()).toBe(false);
      handler.setEnabled(true);
      expect(handler.isEnabled()).toBe(true);
    });
  });

  describe('decompress uncompressed text', () => {
    it('should parse JSON string directly', () => {
      const message = '{"type":"test","payload":{"data":"hello"},"timestamp":"2026-05-12T23:21:00Z"}';
      const result = handler.decompress(message, false);

      expect(result).not.toBeNull();
      expect(result?.type).toBe('test');
      expect(result?.payload.data).toBe('hello');
    });

    it('should skip compression stats for text messages', () => {
      const message = '{"type":"test","payload":{},"timestamp":""}';
      handler.decompress(message, false);

      const stats = handler.getStats();
      expect(stats.totalMessages).toBe(1);
      expect(stats.skippedMessages).toBe(1);
    });
  });

  describe('decompress binary data', () => {
    it('should decompress compressed JSON', () => {
      // Create and compress test data
      const testData = JSON.stringify({
        type: 'telemetry',
        payload: { device_id: 'CNC-001', temperature: 75.5 },
        timestamp: '2026-05-12T23:21:00Z',
      });
      const jsonBytes = new TextEncoder().encode(testData);
      const compressed = pako.deflate(jsonBytes);

      // Decompress
      const result = handler.decompress(compressed.buffer as ArrayBuffer, true);

      expect(result).not.toBeNull();
      expect(result?.type).toBe('telemetry');
      expect(result?.payload.device_id).toBe('CNC-001');
    });

    it('should update compression stats', () => {
      const testData = JSON.stringify({
        type: 'test',
        payload: { data: 'large data here' },
        timestamp: '',
      });
      const jsonBytes = new TextEncoder().encode(testData);
      const compressed = pako.deflate(jsonBytes);

      handler.decompress(compressed.buffer as ArrayBuffer, true);

      const stats = handler.getStats();
      expect(stats.totalMessages).toBe(1);
      expect(stats.compressedMessages).toBe(1);
      expect(stats.compressedBytes).toBe(compressed.length);
      expect(stats.originalBytes).toBe(jsonBytes.length);
    });

    it('should handle decompression errors gracefully', () => {
      const invalidData = new Uint8Array([1, 2, 3, 4, 5]);
      const result = handler.decompress(invalidData.buffer as ArrayBuffer, true);

      expect(result).toBeNull();
      const stats = handler.getStats();
      expect(stats.skippedMessages).toBe(1);
    });
  });

  describe('stats management', () => {
    it('should track multiple messages', () => {
      // Send 3 text messages
      for (let i = 0; i < 3; i++) {
        handler.decompress('{}', false);
      }

      // Send 2 compressed messages
      const compressed = pako.deflate(new TextEncoder().encode('{}'));
      for (let i = 0; i < 2; i++) {
        handler.decompress(compressed.buffer as ArrayBuffer, true);
      }

      const stats = handler.getStats();
      expect(stats.totalMessages).toBe(5);
      expect(stats.skippedMessages).toBe(3);
      expect(stats.compressedMessages).toBe(2);
    });

    it('should reset stats', () => {
      handler.decompress('{}', false);
      handler.decompress('{}', false);

      handler.resetStats();
      const stats = handler.getStats();
      expect(stats.totalMessages).toBe(0);
    });

    it('should calculate compression ratio', () => {
      const largeData = JSON.stringify({ type: 'test', payload: { data: 'x'.repeat(1000) } });
      const jsonBytes = new TextEncoder().encode(largeData);
      const compressed = pako.deflate(jsonBytes);

      handler.decompress(compressed.buffer as ArrayBuffer, true);

      const stats = handler.getStats();
      expect(stats.compressionRatio).toBeLessThan(1);
      expect(stats.savingsPercent).toBeGreaterThan(0);
    });
  });

  describe('disabled compression', () => {
    it('should parse all messages as text when disabled', () => {
      handler.setEnabled(false);

      const compressed = pako.deflate(new TextEncoder().encode('{}'));
      const result = handler.decompress(compressed.buffer as ArrayBuffer, true);

      // Should treat as text when disabled
      expect(result).toBeNull();
    });
  });
});

describe('createWebSocketURL', () => {
  it('should create ws:// URL for http', () => {
    // Mock window.location using Object.defineProperty
    const mockLocation = {
      ...window.location,
      protocol: 'http:',
      host: 'localhost:8080',
    };
    Object.defineProperty(window, 'location', {
      value: mockLocation,
      writable: true,
      configurable: true,
    });

    const url = createWebSocketURL();
    expect(url).toBe('ws://localhost:8080/ws');
  });

  it('should create wss:// URL for https', () => {
    const mockLocation = {
      ...window.location,
      protocol: 'https:',
      host: 'example.com',
    };
    Object.defineProperty(window, 'location', {
      value: mockLocation,
      writable: true,
      configurable: true,
    });

    const url = createWebSocketURL();
    expect(url).toBe('wss://example.com/ws');
  });
});

describe('CompressedWebSocket', () => {
  // Note: These tests require a WebSocket mock or actual WebSocket connection
  // For full integration tests, use a WebSocket mock library

  it('should initialize with compression enabled', () => {
    const ws = new CompressedWebSocket('ws://localhost/ws', true);
    expect(ws.getCompressionStats().totalMessages).toBe(0);
  });

  it('should track compression stats', () => {
    const ws = new CompressedWebSocket('ws://localhost/ws', true);
    const stats = ws.getCompressionStats();
    expect(stats).toHaveProperty('totalMessages');
    expect(stats).toHaveProperty('compressedMessages');
    expect(stats).toHaveProperty('savingsPercent');
  });

  // Add more integration tests with WebSocket mocking if needed
});