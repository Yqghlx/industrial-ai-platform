# Security Documentation: CSRF Protection with JWT Authentication

## SEC-MED-07: CSRF Protection Explanation for JWT Authentication

### Overview

This document explains the CSRF (Cross-Site Request Forgery) protection strategy for the Industrial AI Platform backend, particularly for JWT-based authentication.

### What is CSRF?

CSRF (Cross-Site Request Forgery) is an attack that forces an end user to execute unwanted actions on a web application in which they're currently authenticated. The attack works by exploiting the fact that browsers automatically include cookies with requests to the same domain.

### CSRF Protection Strategies

#### Strategy 1: JWT Token in Authorization Header (Recommended)

When JWT tokens are stored in localStorage/sessionStorage and sent via the `Authorization` header (as `Bearer` token), **CSRF protection is inherently provided** because:

1. **Browser Behavior**: Browsers do NOT automatically include custom headers like `Authorization` in cross-site requests
2. **Same-Origin Policy**: JavaScript on a malicious site cannot read tokens from localStorage/sessionStorage of another origin
3. **Explicit Transmission**: The token must be explicitly attached to each request by your JavaScript code

**This is our primary CSRF protection mechanism.**

```javascript
// Frontend: JWT token sent via Authorization header (CSRF-safe)
fetch('/api/v1/devices', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
```

#### Strategy 2: Double Submit Cookie Pattern

For applications that use cookies for JWT storage (less secure), CSRF protection via the double submit cookie pattern is required:

1. Server generates a random CSRF token and sets it as a cookie
2. Client reads the cookie and includes the token in requests (header or form field)
3. Server validates that the cookie token matches the request token

See `internal/middleware/csrf.go` for implementation.

### Our CSRF Protection Implementation

#### Primary Protection: JWT via Authorization Header

The Industrial AI Platform uses JWT tokens stored in client-side storage (localStorage or sessionStorage) and transmitted via the `Authorization: Bearer <token>` header. This approach provides inherent CSRF protection.

**Why this is CSRF-safe:**

1. **No Cookie Auto-Send**: The JWT token is NOT stored in a cookie that the browser automatically sends
2. **Explicit Header**: The Authorization header is NOT automatically included by the browser
3. **JavaScript Required**: Only JavaScript code running on your authenticated origin can attach the token
4. **Cross-Origin Blocked**: A malicious site cannot execute JavaScript that reads localStorage from your origin

#### Secondary Protection: SameSite Cookies

Any cookies used by the platform are configured with:
- `SameSite=Strict` or `SameSite=Lax` - prevents cookies from being sent in cross-site requests
- `Secure=true` - ensures cookies are only sent over HTTPS
- `HttpOnly=true` (for session cookies) - prevents JavaScript access

#### Rate Limiting

The `/auth/login` endpoint has rate limiting (`LoginRateLimit()`) to prevent brute-force attacks that could be part of CSRF attack chains.

### CSRF Protection is NOT Needed for These Endpoints

| Endpoint | Reason |
|----------|--------|
| `/api/v1/auth/login` | No authentication required |
| `/api/v1/auth/register` | No authentication required |
| `/api/v1/devices/telemetry` | Device API Key authentication (not cookie-based) |
| `/ws` | WebSocket with origin validation |

### CSRF Protection IS Needed for These Scenarios

If you switch from Authorization header to cookie-based JWT:

| Action | Protection Required |
|--------|---------------------|
| POST/PUT/DELETE to authenticated endpoints | CSRF token validation |
| State-changing operations | CSRF middleware |

### Implementation Details

The `csrf.go` middleware provides:

```go
// Default configuration (secure)
func DefaultCSRFConfig() *CSRFConfig {
    return &CSRFConfig{
        CookieSecure:   true,  // HTTPS only
        CookieHTTPOnly: false, // Must be false so JS can read it
        CookieSameSite: http.SameSiteStrictMode,
        CookieMaxAge:   3600, // 1 hour
    }
}
```

### Recommendations

1. **Continue using Authorization header**: The current JWT implementation via Authorization header is CSRF-safe
2. **Document the decision**: Clearly document that CSRF protection is provided by the JWT header mechanism
3. **Monitor for cookie usage**: If any future code adds JWT to cookies, implement CSRF middleware immediately
4. **Add CSRF for session auth**: If session-based authentication is added later, CSRF middleware is required

### Testing CSRF Protection

The `auth_bypass_test.go` file contains CSRF protection tests:

```go
func TestCSRFProtection(t *testing.T) {
    // Test: CSRF token validation for cookie-based auth
    // Test: Missing CSRF token returns 403
    // Test: Invalid CSRF token returns 403
}
```

### Security Audit Checklist

- [x] JWT tokens sent via Authorization header (not cookie)
- [x] SameSite attribute set on all cookies
- [x] Rate limiting on authentication endpoints
- [x] CSRF middleware available for future cookie-based auth
- [x] Documentation of CSRF protection strategy

### References

- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [JWT Best Practices](https://auth0.com/blog/jwt-authentication-best-practices/)