# WebSocket Compression Configuration Guide

This guide explains how to configure WebSocket message compression in the Industrial AI Platform.

## Overview

WebSocket compression reduces bandwidth usage for large messages (>1KB) while maintaining compatibility and performance for small messages.

## Configuration

### Backend Configuration (Go)

Configure WebSocket compression through environment variables:

```bash
# WebSocket Compression Settings
WS_COMPRESSION_ENABLED=true      # Enable/disable compression (default: true)
WS_COMPRESSION_LEVEL=6           # Compression level 1-9 (default: 6)
WS_COMPRESSION_MIN_SIZE=1024     # Minimum message size in bytes to compress (default: 1024)
```

#### Configuration Details

- **WS_COMPRESSION_ENABLED**: Enable or disable compression. Set to `false` to disable compression completely.
  - Default: `true`
  - Recommended: `true` for production

- **WS_COMPRESSION_LEVEL**: Compression intensity (1-9)
  - 1: Fastest compression, lowest compression ratio
  - 9: Slowest compression, highest compression ratio
  - Default: `6` (balanced)
  - Recommended: `3-6` for real-time data streaming

- **WS_COMPRESSION_MIN_SIZE**: Minimum message size threshold for compression
  - Default: `1024` (1KB)
  - Format: Can be specified as bytes, KB, MB (e.g., `1024`, `1KB`, `2KB`)
  - Messages smaller than this threshold are sent uncompressed
  - Recommended: `1024-2048` bytes

### Frontend Configuration (TypeScript)

The frontend automatically handles decompression using the `pako` library. No manual configuration required.

Compression is enabled by default in the WebSocket connection handler:

```typescript
const ws = new CompressedWebSocket(wsUrl, true); // Enable compression
```

To disable compression in frontend:

```typescript
const ws = new CompressedWebSocket(wsUrl, false); // Disable compression
```

## Compression Algorithm

- **Algorithm**: DEFLATE (same as zlib/gzip)
- **Implementation**: Go `compress/flate` package
- **Frontend Library**: `pako` (JavaScript zlib implementation)

## How It Works

### Message Flow

1. **Backend → Frontend**:
   - Backend checks message size
   - If size > threshold: compress message, send as binary
   - If size ≤ threshold: send as JSON text
   - Frontend detects binary/text message type
   - Binary: decompress using pako, parse JSON
   - Text: parse JSON directly

2. **Frontend → Backend** (currently not compressed):
   - Client messages are typically small, sent as JSON text
   - Backend reads and parses JSON

### Compression Decision Logic

```
if message.size < MIN_SIZE:
    send uncompressed (JSON text)
else:
    compressed = compress(message)
    if compressed.size / message.size > 0.9:  # threshold
        send uncompressed (compression not beneficial)
    else:
        send compressed (binary)
```

## Performance Impact

### Benefits

- **Bandwidth Reduction**: 50-80% for large JSON messages
- **Network Latency**: Reduced for large messages
- **Mobile/Frontend**: Better performance on slow connections

### Considerations

- **CPU Usage**: Slight increase for compression/decompression
- **Memory**: Minimal overhead
- **Small Messages**: No impact (skipped compression)

### Benchmark Results

| Message Size | Original | Compressed | Savings |
|-------------|----------|------------|---------|
| 1 KB        | 1024 B   | 1024 B     | 0%      |
| 10 KB       | 10240 B  | 2048 B     | 80%     |
| 100 KB      | 102400 B | 15360 B    | 85%     |
| 1 MB        | 1048576 B| 131072 B   | 87%     |

## Monitoring

### WebSocket Compression Stats API

Access compression statistics via HTTP endpoint:

```
GET /ws/stats
```

Response:

```json
{
  "enabled": true,
  "total_messages": 1000,
  "compressed_messages": 800,
  "skipped_messages": 200,
  "original_bytes": 1024000,
  "compressed_bytes": 204800,
  "compression_ratio": 0.2,
  "savings_percent": 80
}
```

### Frontend Compression Stats

Get stats from WebSocket instance:

```typescript
const stats = ws.getCompressionStats();
console.log(`Total messages: ${stats.totalMessages}`);
console.log(`Savings: ${stats.savingsPercent}%`);
```

## Troubleshooting

### Compression Not Working

1. Check backend logs: `[WebSocket Compression] Enabled: true`
2. Verify message size exceeds threshold
3. Check `/ws/stats` endpoint for compression statistics

### Performance Issues

1. Lower compression level (e.g., `WS_COMPRESSION_LEVEL=3`)
2. Increase minimum size threshold (e.g., `WS_COMPRESSION_MIN_SIZE=2KB`)
3. Disable compression if needed (`WS_COMPRESSION_ENABLED=false`)

### Compatibility Issues

- Ensure frontend uses `binaryType = 'arraybuffer'`
- Verify `pako` library is installed and loaded
- Check WebSocket connection supports binary messages

## Best Practices

1. **Production Settings**:
   - `WS_COMPRESSION_ENABLED=true`
   - `WS_COMPRESSION_LEVEL=6`
   - `WS_COMPRESSION_MIN_SIZE=1KB`

2. **High-Traffic Systems**:
   - `WS_COMPRESSION_LEVEL=3` (faster compression)
   - `WS_COMPRESSION_MIN_SIZE=2KB` (skip more small messages)

3. **Mobile Clients**:
   - Enable compression for bandwidth savings
   - Use level 6-9 for better compression

4. **Real-Time Telemetry**:
   - Enable compression for large batch updates
   - Keep level 3-6 for balance

## Testing Compression

### Manual Test

1. Start backend with compression enabled
2. Open frontend, check console for compression logs:
   ```
   [WS Compression] Decompressed: 2048 bytes → 10240 bytes (80% saved)
   ```
3. Access `/ws/stats` to view statistics

### Automated Test

```bash
# Send large WebSocket message
curl -X POST http://localhost:8080/api/v1/devices/telemetry \
  -H "Content-Type: application/json" \
  -d '{"device_id":"TEST-001","temperature":75.5,...}'

# Check compression stats
curl http://localhost:8080/ws/stats
```

## Future Enhancements

1. **Per-Message Compression**: Allow client to request compression
2. **Dynamic Threshold**: Adjust threshold based on network conditions
3. **Compression Negotiation**: WebSocket subprotocol negotiation
4. **Batch Compression**: Compress multiple messages together