# API Security Guide

This document provides comprehensive guidance on API security for the Industrial AI Platform, including CSRF protection, authentication mechanisms, and best practices.

## Table of Contents

1. [Authentication Overview](#authentication-overview)
2. [CSRF Protection](#csrf-protection)
3. [JWT Authentication](#jwt-authentication)
4. [Session Authentication](#session-authentication)
5. [Device Authentication](#device-authentication)
6. [WebSocket Security](#websocket-security)
7. [Best Practices](#best-practices)

---

## Authentication Overview

The Industrial AI Platform uses multiple authentication mechanisms depending on the use case:

| Mechanism | Use Case | CSRF Required |
|-----------|----------|---------------|
| JWT Tokens | User API access | No |
| Session Cookies | Web UI (optional) | Yes |
| Device API Keys | IoT device telemetry | No |
| WebSocket | Real-time data | Rate Limiting |

---

## CSRF Protection

### What is CSRF?

Cross-Site Request Forgery (CSRF) is an attack that forces an authenticated user to execute unwanted actions on a web application. The attack exploits the fact that browsers automatically include cookies for every request to a domain.

### When CSRF Protection is Required

CSRF protection is **REQUIRED** when:
- Using session-based authentication with cookies
- The application uses traditional form submissions
- Cookies are used for authentication tokens

CSRF protection is **NOT REQUIRED** when:
- Using JWT tokens stored in localStorage/sessionStorage
- Tokens are sent via Authorization header
- Using the Authorization header with Bearer tokens

### JWT vs CSRF

The Industrial AI Platform primarily uses JWT (JSON Web Token) authentication:

```
Authorization: Bearer <jwt_token>
```

**Why JWT doesn't need CSRF protection:**

1. **Storage**: JWT tokens are stored in localStorage or sessionStorage (not cookies)
2. **Transmission**: Tokens are explicitly added to the Authorization header by JavaScript
3. **Browser Behavior**: Browsers do NOT automatically include Authorization headers in cross-site requests
4. **Attack Prevention**: An attacker cannot read or include the JWT token from a victim's browser

### CSRF Middleware (For Session Authentication)

If session-based authentication is used, CSRF protection is implemented via the Double Submit Cookie pattern:

```go
// Apply CSRF middleware for session-based routes
sessionRoutes.Use(middleware.CSRF())
```

The middleware:
1. Generates a random CSRF token
2. Sets it as a cookie (readable by JavaScript)
3. Validates the token from the request header/form against the cookie
4. Rejects requests that don't match

### CSRF Token Flow

```
┌─────────┐                    ┌─────────────┐
│ Client  │                    │   Server    │
└────┬────┘                    └──────┬──────┘
     │                                │
     │  GET /api/v1/auth/login        │
     │───────────────────────────────>│
     │                                │ Set CSRF cookie
     │  Response + CSRF cookie        │
     │<───────────────────────────────│
     │                                │
     │  POST /api/v1/action           │
     │  + X-CSRF-Token header         │
     │───────────────────────────────>│
     │                                │ Validate token
     │  Success                       │
     │<───────────────────────────────│
     │                                │
```

---

## JWT Authentication

### Token Structure

JWT tokens have three parts:
- **Header**: Token type and signing algorithm
- **Payload**: User claims (ID, role, expiration)
- **Signature**: Cryptographic signature

### Access Token

```json
{
  "user_id": 123,
  "username": "operator1",
  "role": "operator",
  "tenant_id": "tenant-001",
  "token_type": "access",
  "exp": 1700000000,
  "iat": 1699990000
}
```

**Expiration**: 15 minutes (configurable)

### Refresh Token

```json
{
  "user_id": 123,
  "token_type": "refresh",
  "exp": 1700500000,
  "iat": 1699990000
}
```

**Expiration**: 7 days (configurable)

### Token Refresh Flow

```
┌─────────┐                    ┌─────────────┐
│ Client  │                    │   Server    │
└────┬────┘                    └──────┬──────┘
     │                                │
     │  POST /auth/refresh            │
     │  + refresh_token               │
     │───────────────────────────────>│
     │                                │ Validate refresh token
     │                                │ Check token blacklist
     │                                │ Generate new tokens
     │  new access_token              │
     │  new refresh_token             │
     │<───────────────────────────────│
     │                                │
```

### Token Blacklist

When a user logs out, their tokens are blacklisted to prevent reuse:

```go
// Logout invalidates tokens
auth.POST("/auth/logout", authHandler.Logout)
```

---

## Session Authentication

### When to Use Sessions

Sessions should be used when:
- Building a traditional web application
- Users need persistent login across tabs
- Cookie-based authentication is preferred

### Session CSRF Protection

```go
// Apply CSRF for session routes
sessionGroup := router.Group("/session")
sessionGroup.Use(middleware.CSRF())
sessionGroup.POST("/login", sessionLoginHandler)
sessionGroup.POST("/action", sessionActionHandler)
```

### CSRF Token Handling

**Frontend Implementation:**

```javascript
// Get CSRF token from cookie
function getCSRFToken() {
  const name = 'csrf_token=';
  const cookies = document.cookie.split(';');
  for (let cookie of cookies) {
    cookie = cookie.trim();
    if (cookie.startsWith(name)) {
      return cookie.substring(name.length);
    }
  }
  return null;
}

// Include token in request headers
fetch('/api/v1/action', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': getCSRFToken()
  },
  body: JSON.stringify(data)
});
```

---

## Device Authentication

### Device API Keys

IoT devices authenticate using API keys instead of JWT tokens:

```http
POST /api/v1/devices/telemetry
X-Device-Key: <device_api_key>
Content-Type: application/json

{
  "device_id": "sensor-001",
  "temperature": 25.5,
  "pressure": 101.3,
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### Device Key Generation

Device keys are generated using a deterministic algorithm:

```go
key := sha256(deviceID + ":" + deviceSecret)
```

This allows:
- Offline key generation for devices
- Key verification without database lookup
- Secure key distribution during device provisioning

### Device Authentication Middleware

```go
// Apply device authentication
public.POST("/devices/telemetry", 
    middleware.DeviceAuthRequired(config),
    telemetryHandler.Ingest)
```

---

## WebSocket Security

### WebSocket Rate Limiting

WebSocket connections are rate-limited to prevent abuse:

```go
// SEC-001: WebSocket rate limiting
s.router.GET("/ws", middleware.WebSocketRateLimit(), s.handleWebSocket)
```

**Rate Limit Configuration:**
- **Capacity**: 10 connection attempts per IP
- **Refill Rate**: 1 token every 2 seconds
- **Purpose**: Prevent connection flooding while allowing legitimate monitoring

### WebSocket Authentication Options

1. **Public WebSocket** (Current implementation):
   - No authentication required
   - Rate limiting prevents abuse
   - Suitable for public monitoring dashboards

2. **Authenticated WebSocket** (Alternative):
   ```go
   // If authentication is needed
   wsGroup := router.Group("/ws/auth")
   wsGroup.Use(middleware.AuthRequired(jwtSecret))
   wsGroup.GET("", handleAuthenticatedWebSocket)
   ```

### WebSocket Connection Flow

```
┌─────────┐                    ┌─────────────┐
│ Client  │                    │   Server    │
└────┬────┘                    └──────┬──────┘
     │                                │
     │  GET /ws (upgrade request)     │
     │───────────────────────────────>│
     │                                │ Rate limit check
     │                                │ Upgrade to WebSocket
     │  WebSocket connected           │
     │<───────────────────────────────│
     │                                │
     │  Subscribe to topics           │
     │───────────────────────────────>│
     │                                │
     │  Real-time data stream         │
     │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │
     │                                │
```

---

## Best Practices

### Frontend Best Practices

1. **Always use Authorization header**:
   ```javascript
   // Good
   fetch('/api/v1/devices', {
     headers: {
       'Authorization': `Bearer ${accessToken}`
     }
   });
   
   // Avoid (can lead to CSRF with cookies)
   // Do NOT use cookies for JWT storage
   ```

2. **Secure token storage**:
   - Store tokens in sessionStorage (cleared on tab close)
   - Or use in-memory storage with refresh mechanism
   - Avoid localStorage for sensitive tokens (XSS risk)

3. **Token refresh handling**:
   ```javascript
   // Automatic token refresh before expiration
   setInterval(async () => {
     if (tokenExpiresSoon()) {
       const newTokens = await refreshToken();
       updateStoredTokens(newTokens);
     }
   }, 60000); // Check every minute
   ```

### Backend Best Practices

1. **Token validation**:
   ```go
   // Always validate token type
   if claims.TokenType != "access" {
       return Unauthorized("Invalid token type")
   }
   ```

2. **Role-based access control**:
   ```go
   // Check user role before sensitive operations
   adminRoutes.Use(middleware.AdminRequired())
   ```

3. **Rate limiting**:
   ```go
   // Apply appropriate rate limits
   auth.POST("/login", middleware.LoginRateLimit(), handler)
   auth.POST("/agent/query", middleware.AgentQueryRateLimit(), handler)
   ```

4. **Security headers**:
   ```go
   // Apply security headers to all responses
   router.Use(middleware.SecurityHeaders())
   ```

### API Security Checklist

| Check | Description |
|-------|-------------|
| ✅ | JWT tokens sent via Authorization header |
| ✅ | Rate limiting on all public endpoints |
| ✅ | Role-based access control for sensitive operations |
| ✅ | Token blacklist for logout |
| ✅ | Device authentication for telemetry |
| ✅ | WebSocket rate limiting |
| ✅ | CORS properly configured |
| ✅ | Security headers (HSTS, X-Frame-Options, etc.) |

---

## Configuration Reference

### JWT Configuration

```yaml
jwt:
  access_token_expiry: 15m
  refresh_token_expiry: 7d
  issuer: "industrial-ai-platform"
  signing_algorithm: "HS256"
```

### CSRF Configuration

```yaml
csrf:
  enabled: false  # Disabled for JWT-only auth
  cookie_name: "csrf_token"
  header_name: "X-CSRF-Token"
  token_length: 32
  cookie_secure: true
  cookie_same_site: "strict"
```

### Rate Limit Configuration

```yaml
rate_limits:
  login:
    capacity: 5
    refill_rate: 1/s
  telemetry:
    capacity: 200
    refill_rate: 50/s
  websocket:
    capacity: 10
    refill_rate: 0.5/s
  agent_query:
    capacity: 20
    refill_rate: 2/s
```

---

## Security Audit Log

All authentication events are logged for audit purposes:

```json
{
  "event": "auth_login",
  "user_id": 123,
  "username": "operator1",
  "ip_address": "192.168.1.100",
  "timestamp": "2024-01-01T10:00:00Z",
  "success": true
}
```

---

## References

- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [JWT Best Practices](https://auth0.com/blog/jwt-best-practices/)
- [WebSocket Security](https://owasp.org/www-community/vulnerabilities/WebSocket_Security)

---

*Last Updated: January 2024*
*Document Version: 1.0*
*Security Review: SEC-004*