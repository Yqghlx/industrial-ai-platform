package wscompression

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// CompressionStats tracks compression statistics
type CompressionStats struct {
	TotalMessages      int64
	CompressedMessages int64
	SkippedMessages    int64
	OriginalBytes      int64
	CompressedBytes    int64
	CompressionRatio   float64
}

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Enabled   bool
	Level     int     // 1-9, default is 6
	MinSize   int     // Minimum message size to compress (default 1024 bytes)
	Threshold float64 // Only compress if ratio > threshold (default 0.9, meaning at least 10% reduction)
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Enabled:   true,
		Level:     6,
		MinSize:   1024, // 1KB
		Threshold: 0.9,  // Only compress if compressed size is < 90% of original
	}
}

// Compressor manages WebSocket message compression
type Compressor struct {
	config *CompressionConfig
	stats  CompressionStats
	mu     sync.RWMutex

	// Flate writer pool for reuse
	writerPool sync.Pool
}

// NewCompressor creates a new WebSocket compressor
func NewCompressor(config *CompressionConfig) *Compressor {
	if config == nil {
		config = DefaultCompressionConfig()
	}

	// Validate compression level
	if config.Level < 1 || config.Level > 9 {
		config.Level = 6
	}

	// Validate minimum size
	if config.MinSize < 0 {
		config.MinSize = 1024
	}

	return &Compressor{
		config: config,
		writerPool: sync.Pool{
			New: func() interface{} {
				// Create a new flate writer
				return nil // Will be created dynamically with proper level
			},
		},
	}
}

// ShouldCompress determines if a message should be compressed
func (c *Compressor) ShouldCompress(data []byte) bool {
	if !c.config.Enabled {
		return false
	}

	// Don't compress small messages
	if len(data) < c.config.MinSize {
		c.mu.Lock()
		c.stats.SkippedMessages++
		c.mu.Unlock()
		return false
	}

	return true
}

// Compress compresses data using flate compression
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	if !c.ShouldCompress(data) {
		return data, nil
	}

	// Create buffer for compressed data
	var buf bytes.Buffer

	// Create flate writer with configured level
	writer, err := flate.NewWriter(&buf, c.config.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to create flate writer: %w", err)
	}

	// Write data
	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	// Close writer to flush
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close flate writer: %w", err)
	}

	compressed := buf.Bytes()

	// Check if compression is beneficial
	compressionRatio := float64(len(compressed)) / float64(len(data))
	if compressionRatio >= c.config.Threshold {
		// Compression didn't help enough, return original
		c.mu.Lock()
		c.stats.SkippedMessages++
		c.mu.Unlock()
		return data, nil
	}

	// Update stats
	c.mu.Lock()
	c.stats.CompressedMessages++
	c.stats.OriginalBytes += int64(len(data))
	c.stats.CompressedBytes += int64(len(compressed))
	c.stats.TotalMessages++
	c.mu.Unlock()

	return compressed, nil
}

// Decompress decompresses flate compressed data
func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return buf.Bytes(), nil
}

// GetStats returns compression statistics
func (c *Compressor) GetStats() CompressionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	if stats.OriginalBytes > 0 {
		stats.CompressionRatio = float64(stats.CompressedBytes) / float64(stats.OriginalBytes)
	}
	return stats
}

// ResetStats resets compression statistics
func (c *Compressor) ResetStats() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats = CompressionStats{}
}

// CompressedMessage represents a WebSocket message with compression flag
type CompressedMessage struct {
	Compressed bool        `json:"compressed"`
	Data       interface{} `json:"data"`
	Size       int         `json:"size,omitempty"`       // Original size (for monitoring)
	ComprSize  int         `json:"compr_size,omitempty"` // Compressed size (for monitoring)
}

// WriteCompressed writes a compressed JSON message to WebSocket connection
func (c *Compressor) WriteCompressed(conn *websocket.Conn, data interface{}) error {
	// Serialize data to JSON first
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Track original size
	originalSize := len(jsonData)

	// Compress if needed
	compressedData, err := c.Compress(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	// Determine if data was actually compressed
	isCompressed := len(compressedData) != originalSize

	// If compressed, send as binary message with compression flag
	if isCompressed {
		// Send compressed data as binary
		err = conn.WriteMessage(websocket.BinaryMessage, compressedData)
		if err != nil {
			return fmt.Errorf("failed to write compressed message: %w", err)
		}

		// Log compression stats
		log.Printf("[WS Compression] Original: %d bytes, Compressed: %d bytes, Ratio: %.2f%%",
			originalSize, len(compressedData),
			float64(len(compressedData))/float64(originalSize)*100)
	} else {
		// Send uncompressed as text JSON
		err = conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
	}

	// Update total messages count
	c.mu.Lock()
	c.stats.TotalMessages++
	c.mu.Unlock()

	return nil
}

// ReadCompressed reads and decompresses a WebSocket message
func (c *Compressor) ReadCompressed(conn *websocket.Conn) (messageType int, data []byte, err error) {
	messageType, data, err = conn.ReadMessage()
	if err != nil {
		return messageType, nil, err
	}

	// If binary message, decompress
	if messageType == websocket.BinaryMessage && c.config.Enabled {
		decompressed, err := c.Decompress(data)
		if err != nil {
			log.Printf("[WS Compression] Failed to decompress message: %v", err)
			return messageType, data, nil // Return original data on error
		}
		return websocket.TextMessage, decompressed, nil
	}

	return messageType, data, nil
}
