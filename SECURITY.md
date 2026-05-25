# Security Policy

## Industrial AI Platform Security Documentation

This document outlines the security policy and best practices for the Industrial AI Platform.

---

## Table of Contents

1. [Security Overview](#security-overview)
2. [Authentication & Authorization](#authentication--authorization)
3. [Token Storage Security](#token-storage-security)
4. [Data Protection](#data-protection)
5. [Logging & Monitoring](#logging--monitoring)
6. [Security Best Practices](#security-best-practices)
7. [Vulnerability Reporting](#vulnerability-reporting)

---

## Security Overview

The Industrial AI Platform implements multiple security layers to protect user data and system integrity:

- **JWT-based Authentication**: Secure token-based authentication with short-lived access tokens
- **Role-based Access Control**: Admin, operator, and viewer roles with granular permissions
- **Rate Limiting**: Protection against brute-force and DDoS attacks
- **Input Validation**: Comprehensive validation on all API endpoints
- **Security Headers**: XSS, CSRF, and clickjacking protection

---

## Authentication & Authorization

### JWT Authentication

The platform uses JWT (JSON Web Token) authentication:

- **Access Tokens**: 15-minute expiration (configurable)
- **Refresh Tokens**: 7-day expiration (configurable)
- **Token Blacklisting**: Revoked tokens are tracked in Redis

### Authentication Flow

```
┌─────────┐                    ┌─────────────┐
│ Client  │                    │   Server    │
└────┬────┘                    └──────┬──────┘
     │                                │
     │  POST /auth/login              │
     │───────────────────────────────>│
     │                                │ Validate credentials
     │                                │ Generate tokens
     │  access_token + refresh_token  │
     │<───────────────────────────────│
     │                                │
```

### Role-based Access Control

| Role | Permissions |
|------|-------------|
| `admin` | Full system access, user management, configuration |
| `operator` | Device management, alert handling, report generation |
| `viewer` | Read-only access to dashboards and reports |

---

## Token Storage Security

### SEC-LOW-02: Token Storage Best Practices

**Important**: Authentication tokens should be stored in `sessionStorage` instead of `localStorage`.

#### Why sessionStorage is More Secure

| Feature | localStorage | sessionStorage |
|---------|--------------|----------------|
| Persistence | Survives browser restart | Cleared on tab close |
| XSS Attack Window | Extended exposure | Limited exposure time |
| Cross-tab access | All tabs share data | Isolated per tab |
| Data recovery | Attackers can recover after reboot | Data lost on session end |

#### Implementation

```javascript
// Secure: Use sessionStorage
sessionStorage.setItem('token', accessToken);

// Avoid: localStorage (XSS risk)
localStorage.setItem('token', accessToken);  // ❌ Not recommended
```

#### Alternative: HttpOnly Cookies (Best Security)

For maximum security, consider using HttpOnly cookies:

```go
// Backend sets HttpOnly cookie
c.SetCookie("token", token, 3600, "/", "", true, true)
```

Benefits:
- JavaScript cannot read the cookie (immune to XSS)
- Automatic inclusion in requests
- Can be configured with SameSite=Strict

### Token Handling Guidelines

1. **Never log tokens**: Avoid including tokens in log messages
2. **Clear on logout**: Remove tokens from storage immediately
3. **Short expiration**: Use short-lived access tokens (≤15 minutes)
4. **Refresh securely**: Implement secure token refresh mechanism

---

## Data Protection

### Sensitive Data Handling

The platform implements the following data protection measures:

- **Password Hashing**: bcrypt with configurable cost factor
- **Encryption**: AES-256 for sensitive configuration data
- **Data Masking**: Sensitive fields are masked in logs and API responses

### Never Log Sensitive Information

```go
// Good: Log operation failure without sensitive data
log.Printf("Failed to hash password: %v", err)

// Bad: Log contains password
log.Printf("Password attempt: %s", password)  // ❌ NEVER DO THIS
```

### Fields to Never Log

- Passwords (plain or hashed)
- JWT tokens
- API keys
- Session IDs
- Personal identifiable information (PII)

---

## Logging & Monitoring

### SEC-LOW-01: Secure Logging Practices

All logs are designed to avoid sensitive information leakage:

#### What We Log

- Authentication events (success/failure, user ID, IP address)
- Authorization failures
- Rate limit violations
- System errors (without sensitive data)

#### What We Never Log

- Passwords or password hashes
- JWT tokens or API keys
- User credentials
- Encryption keys

### Audit Log Format

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

## Security Best Practices

### For Developers

1. **Input Validation**
   - Validate all user input on both frontend and backend
   - Use whitelist validation where possible
   - Sanitize HTML output to prevent XSS

2. **API Security**
   - Always authenticate API requests
   - Use rate limiting on public endpoints
   - Return generic error messages (avoid information leakage)

3. **Dependencies**
   - Regularly update dependencies
   - Use `npm audit` and `go mod` security scanning
   - Pin dependency versions

### For Operators

1. **Environment Configuration**
   - Never expose secrets in environment variables (use secret management)
   - Enable HTTPS in production
   - Configure proper CORS policies

2. **Monitoring**
   - Monitor authentication failures
   - Track rate limit violations
   - Set up alerts for suspicious activity

### Security Headers Applied

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-XSS-Protection` | `1; mode=block` | XSS filter (legacy browsers) |
| `Content-Security-Policy` | `default-src 'self'` | CSP protection |
| `Strict-Transport-Security` | `max-age=31536000` | Force HTTPS |

---

## Vulnerability Reporting

### How to Report

If you discover a security vulnerability, please report it responsibly:

1. **Email**: security@industrial-ai-platform.example.com
2. **Subject**: [Security] Vulnerability Report
3. **Content**: 
   - Description of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

### Response Timeline

| Stage | Timeline |
|-------|----------|
| Acknowledgment | 24 hours |
| Initial Assessment | 72 hours |
| Fix Development | 7-14 days |
| Security Advisory | After fix deployment |

### What to Expect

- We will acknowledge your report within 24 hours
- We will provide an estimated timeline for fix
- We will credit you in the security advisory (if desired)
- We do not pursue legal action against responsible disclosure

---

## Security Checklist

### Pre-deployment Security Checklist

- [ ] HTTPS enabled with valid certificate
- [ ] JWT secret is strong and properly stored
- [ ] Rate limiting configured on all endpoints
- [ ] Security headers are applied
- [ ] Database credentials are not hardcoded
- [ ] Logs do not contain sensitive information
- [ ] Tokens stored in sessionStorage (not localStorage)
- [ ] CORS configured for production domain only
- [ ] Input validation on all endpoints
- [ ] Dependencies audited for vulnerabilities

---

## Related Documentation

- [API Security Guide](docs/API_SECURITY.md)
- [Local Development Security](backend/docs/LOCAL_DEVELOPMENT_SECURITY.md)

---

*Last Updated: January 2024*
*Document Version: 1.0*
*Security Review: SEC-LOW-01, SEC-LOW-02, SEC-LOW-03*