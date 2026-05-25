import pako from 'pako';

/**
 * WebSocket Compression Handler
 * Handles compression/decompression of WebSocket messages using pako (zlib/flate)
 */

export interface CompressionStats {
  totalMessages: number;
  compressedMessages: number;
  skippedMessages: number;
  originalBytes: number;
  compressedBytes: number;
  compressionRatio: number;
  savingsPercent: number;
}

export interface WSMessage {
  type: string;
  payload: unknown;
  timestamp: string;
}

export class WebSocketCompressionHandler {
  private enabled: boolean = true;
  private stats: CompressionStats = {
    totalMessages: 0,
    compressedMessages: 0,
    skippedMessages: 0,
    originalBytes: 0,
    compressedBytes: 0,
    compressionRatio: 0,
    savingsPercent: 0,
  };

  constructor(enabled: boolean = true) {
    this.enabled = enabled;
  }

  /**
   * Check if compression is enabled
   */
  isEnabled(): boolean {
    return this.enabled;
  }

  /**
   * Enable/disable compression
   */
  setEnabled(enabled: boolean): void {
    this.enabled = enabled;
  }

  /**
   * Decompress WebSocket message
   * @param data - Raw data from WebSocket (may be compressed or uncompressed)
   * @param isBinary - Whether the message is binary (indicates compression)
   * @returns Decompressed/parsed message
   */
  decompress(data: ArrayBuffer | string, isBinary: boolean): WSMessage | null {
    this.stats.totalMessages++;

    if (!this.enabled || !isBinary) {
      // Not compressed, parse directly
      if (typeof data === 'string') {
        try {
          const message = JSON.parse(data);
          this.stats.skippedMessages++;
          return message;
        } catch (e) {
          console.error('[WS Compression] Failed to parse uncompressed message:', e);
          return null;
        }
      }
      // Binary but not compressed (shouldn't happen)
      return null;
    }

    // Decompress binary data
    try {
      const compressedData = new Uint8Array(data as ArrayBuffer);
      const decompressedData = pako.inflate(compressedData);
      
      // Update stats
      this.stats.compressedMessages++;
      this.stats.compressedBytes += compressedData.length;
      this.stats.originalBytes += decompressedData.length;
      this.updateCompressionRatio();

      // Decode to string and parse JSON
      const jsonString = new TextDecoder('utf-8').decode(decompressedData);
      const message = JSON.parse(jsonString);
      
      // Decompression stats: ${compressedData.length} bytes → ${decompressedData.length} bytes
      
      return message;
    } catch (e) {
      console.error('[WS Compression] Failed to decompress message:', e);
      this.stats.skippedMessages++;
      return null;
    }
  }

  /**
   * Get compression statistics
   */
  getStats(): CompressionStats {
    return { ...this.stats };
  }

  /**
   * Reset statistics
   */
  resetStats(): void {
    this.stats = {
      totalMessages: 0,
      compressedMessages: 0,
      skippedMessages: 0,
      originalBytes: 0,
      compressedBytes: 0,
      compressionRatio: 0,
      savingsPercent: 0,
    };
  }

  /**
   * Update compression ratio
   */
  private updateCompressionRatio(): void {
    if (this.stats.originalBytes > 0) {
      this.stats.compressionRatio = this.stats.compressedBytes / this.stats.originalBytes;
      this.stats.savingsPercent = (1 - this.stats.compressionRatio) * 100;
    }
  }
}

/**
 * Create a WebSocket connection with compression support
 */
export class CompressedWebSocket {
  private ws: WebSocket | null = null;
  private compressionHandler: WebSocketCompressionHandler;
  private url: string;
  private onMessageCallback: ((message: WSMessage) => void) | null = null;
  private onErrorCallback: ((error: Event) => void) | null = null;
  private onCloseCallback: (() => void) | null = null;
  private onConnectCallback: (() => void) | null = null;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 1000;
  private shouldReconnect: boolean = true;

  constructor(url: string, compressionEnabled: boolean = true) {
    this.url = url;
    this.compressionHandler = new WebSocketCompressionHandler(compressionEnabled);
  }

  /**
   * Connect to WebSocket
   */
  connect(): void {
    if (this.ws) {
      this.ws.close();
    }

    this.ws = new WebSocket(this.url);
    this.ws.binaryType = 'arraybuffer'; // Important for handling compressed data

    this.ws.onopen = () => {
      // WebSocket connected
      this.reconnectAttempts = 0;
      if (this.onConnectCallback) {
        this.onConnectCallback();
      }
    };

    this.ws.onmessage = (event) => {
      const isBinary = event.data instanceof ArrayBuffer;
      const message = this.compressionHandler.decompress(event.data, isBinary);
      
      if (message && this.onMessageCallback) {
        this.onMessageCallback(message);
      }
    };

    this.ws.onerror = (error) => {
      console.error('[WebSocket] Error:', error);
      if (this.onErrorCallback) {
        this.onErrorCallback(error);
      }
    };

    this.ws.onclose = () => {
      // WebSocket disconnected
      if (this.onCloseCallback) {
        this.onCloseCallback();
      }
      
      // Attempt reconnect if enabled
      if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        // Reconnecting (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})
        setTimeout(() => this.connect(), this.reconnectDelay);
        this.reconnectDelay *= 2; // Exponential backoff
      }
    };
  }

  /**
   * Disconnect WebSocket
   */
  disconnect(): void {
    this.shouldReconnect = false;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Send message (currently not compressed from client)
   */
  send(data: unknown): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  /**
   * Set message handler
   */
  onMessage(callback: (message: WSMessage) => void): void {
    this.onMessageCallback = callback;
  }

  /**
   * Set error handler
   */
  onError(callback: (error: Event) => void): void {
    this.onErrorCallback = callback;
  }

  /**
   * Set close handler
   */
  onClose(callback: () => void): void {
    this.onCloseCallback = callback;
  }

  /**
   * Set connect handler
   */
  onConnect(callback: () => void): void {
    this.onConnectCallback = callback;
  }

  /**
   * Get compression statistics
   */
  getCompressionStats(): CompressionStats {
    return this.compressionHandler.getStats();
  }

  /**
   * Check if WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

/**
 * Utility function to create WebSocket URL from current location
 */
export function createWebSocketURL(): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  // Get token from localStorage (set by API class on login)
  const token = localStorage.getItem('token');
  if (token) {
    return `${protocol}//${host}/ws?token=${encodeURIComponent(token)}`;
  }
  return `${protocol}//${host}/ws`;
}