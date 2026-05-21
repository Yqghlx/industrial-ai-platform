# Health Check Enhancement

## Overview
Enhanced the `/health` endpoint to provide comprehensive multi-component health checking instead of just database ping.

## Changes Made

### 1. Created HealthService (`backend/internal/service/health_service.go`)
A new dedicated service for health checking that includes:

**Components Checked:**
- **Database**: Connection status and latency
- **LLM API**: Availability and connectivity check
- **Memory**: System memory usage with warnings for high usage
- **Disk**: Basic write accessibility check

**Features:**
- Concurrent health checks using goroutines for better performance
- Configurable timeout (default: 5 seconds overall, 2-3 seconds per component)
- Singleton pattern for global health service instance
- Comprehensive status reporting

### 2. Enhanced Server Structure (`backend/internal/handler/server.go`)
- Added `healthSvc *service.HealthService` field
- Added `startTime time.Time` for uptime tracking
- Initialize health service in `NewServer()`

### 3. Updated Health Endpoint (`backend/internal/handler/auth_handler.go`)
The `/health` endpoint now returns:

```json
{
  "status": "healthy",
  "components": {
    "database": {
      "status": "healthy",
      "message": "connected",
      "latency_ms": 5
    },
    "llm_api": {
      "status": "available",
      "message": "API reachable, model: glm-5",
      "latency_ms": 120
    },
    "memory": {
      "status": "healthy",
      "message": "45.2% used (128/284 MB)"
    },
    "disk": {
      "status": "healthy",
      "message": "disk accessible"
    }
  },
  "version": "1.0.0",
  "uptime": "2h30m",
  "timestamp": "2024-01-15T10:30:00Z",
  "ws_clients": 5
}
```

### 4. Status Codes and Levels

**Overall Status:**
- `healthy`: All critical components are operational
- `degraded`: Non-critical components have issues (e.g., LLM API unavailable)
- `unhealthy`: Critical components are failing (e.g., database disconnected)

**HTTP Status Codes:**
- 200 OK: Status is `healthy` or `degraded`
- 503 Service Unavailable: Status is `unhealthy`

**Component Status Values:**
- `healthy`: Component is working normally
- `warning`: Component is working but with concerns (e.g., high memory)
- `unhealthy`: Component is not functioning properly
- `unavailable`: Component is not configured (e.g., no LLM API key)
- `unknown`: Cannot determine status

## Performance Considerations

1. **Concurrent Checks**: All component checks run in parallel using goroutines
2. **Timeouts**: Each check has appropriate timeouts to prevent hanging
   - Overall health check: 5 seconds
   - Database ping: 2 seconds
   - LLM API check: 3 seconds
3. **Non-Blocking**: Health checks won't block the main application flow
4. **Fast Fallback**: If a component check times out, it's marked as unhealthy and the overall check continues

## Configuration

The health service automatically reads configuration from environment variables:
- `LLM_API_KEY`: API key for LLM service
- `LLM_BASE_URL`: Base URL for LLM API (default: https://coding.dashscope.aliyuncs.com/v1)
- `LLM_MODEL`: LLM model name (default: glm-5)

## Usage

Simply access the health endpoint:
```bash
curl http://localhost:8080/health
```

The response includes detailed information about all components, making it suitable for:
- Monitoring systems (Prometheus, Datadog, etc.)
- Load balancer health checks
- Kubernetes liveness/readiness probes
- Operational dashboards

## Testing

Run the health service tests:
```bash
cd backend
go test ./internal/service -v -run TestHealthService
```

## Future Enhancements

Possible improvements:
1. Add disk space usage metrics using syscall.Statfs
2. Add CPU usage monitoring
3. Add custom health check endpoints for specific components
4. Add health check history/metrics
5. Add dependency health checks (external services, message queues, etc.)