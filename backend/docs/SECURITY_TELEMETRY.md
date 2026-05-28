# Security Documentation: Telemetry Endpoint Authentication

## SEC-MAJOR-01: /api/v1/devices/telemetry Endpoint Security Analysis

### Overview

This document explains the security design for the `/api/v1/devices/telemetry` endpoint, which now **requires device authentication** via API Key.

### Security Change Summary (SEC-MAJOR-01)

**Previous State (SEC-MED-02)**: The endpoint was intentionally public with rate limiting as the primary defense.

**Current State (SEC-MAJOR-01)**: The endpoint now **requires device authentication** using the `DeviceAuthRequired` middleware. Devices must provide valid API keys via `X-Device-Key` header or `device_key` query parameter.

### Endpoint Purpose

The `/api/v1/devices/telemetry` endpoint is designed for **edge device data ingestion**. Industrial IoT devices deployed in manufacturing facilities upload real-time sensor data to the platform using device-specific API keys.

### Security Design

#### Mandatory Device Authentication (SEC-MAJOR-01)

| Security Measure | Implementation |
|------------------|----------------|
| **Device Authentication** | `DeviceAuthRequired(nil)` middleware validates API keys |
| **Rate Limiting** | `TelemetryRateLimit()` middleware limits requests per device/IP |
| **Input Validation** | `ValidateTelemetryData()` validates all input fields |
| **Device ID Format Check** | UUID/safe-ID format validation required |
| **Data Sanitization** | SQL injection detection on all string inputs |
| **IP-Based Throttling** | Prevents abuse from single source |

### Security Controls Implemented

#### 1. Device Authentication (Primary Defense - SEC-MAJOR-01)

```go
// In server_new.go
s.router.POST("/api/v1/devices/telemetry",
    middleware.TelemetryRateLimit(),
    middleware.DeviceAuthRequired(nil), // SEC-MAJOR-01: Mandatory device auth
    s.telemetryHandler.IngestTelemetry)
```

The `DeviceAuthRequired(nil)` middleware:
- Validates device API keys from `X-Device-Key` header or `device_key` query parameter
- Supports multiple key formats:
  - Direct API key matching `DEVICE_API_KEY` environment variable
  - Device-specific key format: `deviceID:key`
  - SHA-256 hashed key: `sha256(deviceID + ":" + secret)`
- Returns `401 Unauthorized` for invalid or missing keys

#### 2. API Key Configuration

```go
// Environment variable configuration
DEVICE_API_KEY=your-secure-api-key-here

// Generate device-specific keys
key := middleware.GenerateDeviceKey("device-001", "your-secret")
// Result: sha256("device-001:your-secret")
```

#### 3. Rate Limiting (Secondary Defense)

```go
// Rate limiting middleware
middleware.TelemetryRateLimit()
```

The rate limiting provides:
- Request per-second limits
- Burst allowance for legitimate traffic spikes
- IP-based identification for rate limit enforcement

#### 4. Input Validation (SEC-MED-04)

```go
// In telemetry_service.go
func ValidateTelemetryData(data *model.TelemetryData) error {
    if data.DeviceID == "" {
        return fmt.Errorf("device_id is required")
    }
    // SEC-MED-04: Add UUID format validation
    if err := validation.ValidateDeviceID(data.DeviceID); err != nil {
        return fmt.Errorf("invalid device_id format: %w", err)
    }
    // ... additional validation
}
```

Device ID must be:
- Valid UUID format (e.g., `550e8400-e29b-41d4-a716-446655440000`)
- Or safe alphanumeric ID (e.g., `CNC-001`, `DEV_001`)
- Maximum 100 characters
- No special characters except dash and underscore

#### 5. SQL Injection Detection (SEC-MED-05)

```go
// Enhanced SQL injection detection
import "github.com/industrial-ai/platform/internal/security"

// Validate all string inputs
if security.ContainsSQLInjection(data.Message, nil) {
    return fmt.Errorf("invalid message content")
}
```

### Authentication Methods

Devices can authenticate using:

1. **Direct API Key**: 
   ```
   X-Device-Key: your-api-key
   ```

2. **Device-Specific Key**: 
   ```
   X-Device-Key: device-001:your-api-key
   ```

3. **SHA-256 Hashed Key**:
   ```
   X-Device-Key: a1b2c3d4... (sha256 hash)
   ```

4. **Query Parameter** (fallback):
   ```
   POST /api/v1/devices/telemetry?device_key=your-api-key
   ```

### Recommended Security Configuration

#### Development Environment

```env
# Accept any key >= 8 characters (development mode)
DEVICE_API_KEY=  # Empty = development mode
TELEMETRY_RATE_LIMIT=1000  # requests per second
```

#### Production Environment

```env
# Secure API key required
DEVICE_API_KEY=your-secure-256-bit-key-here
TELEMETRY_RATE_LIMIT=100   # requests per second
```

### Threat Analysis

| Threat | Mitigation |
|--------|------------|
| **Unauthorized Data Injection** | Device authentication (API key) + rate limiting |
| **Data Flooding** | Per-IP rate limiting |
| **Malicious Data** | Input validation + SQL injection detection |
| **Data Poisoning** | Device authentication + device ID validation |
| **Denial of Service** | Rate limiting + request timeout |
| **Key Theft** | Rotate keys periodically, use HTTPS |

### Device Key Management

1. **Key Generation**: Use `GenerateDeviceKey(deviceID, secret)`
2. **Key Storage**: Store hashed keys in secure configuration
3. **Key Rotation**: Rotate device keys periodically
4. **Key Revocation**: Implement key revocation mechanism if needed

### Security Recommendations

1. **Always Use Device Auth**: Required for all telemetry requests
2. **Secure API Keys**: Use strong, unique API keys for production
3. **HTTPS Required**: Always transmit keys over HTTPS
4. **Rate Limiting**: Monitor and adjust rate limits based on traffic
5. **Key Rotation**: Rotate keys periodically (recommended: 90 days)
6. **Monitor Traffic**: Log and monitor telemetry endpoint usage

### Implementation Checklist

- [x] Device authentication middleware applied (SEC-MAJOR-01)
- [x] Rate limiting middleware applied
- [x] Input validation for device_id format (SEC-MED-04)
- [x] SQL injection detection (SEC-MED-05)
- [x] Security documentation updated
- [x] CORS origin validation
- [x] Telemetry endpoint removed from public endpoints list

### Related Security Issues

- **SEC-HIGH-02**: Device authentication middleware implementation
- **SEC-MED-01**: WebSocket authentication
- **SEC-MED-04**: Device ID format validation
- **SEC-MED-05**: SQL injection detection