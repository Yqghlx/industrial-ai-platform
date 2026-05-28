# Database SSL/TLS Configuration Guide

> **Document Version:** 1.0  
> **Last Updated:** 2026-05-28  
> **Author:** DevOps Team

## Table of Contents

1. [Overview](#overview)
2. [SSL Modes Explained](#ssl-modes-explained)
3. [Configuration Methods](#configuration-methods)
4. [Production Recommendations](#production-recommendations)
5. [Troubleshooting](#troubleshooting)
6. [Security Best Practices](#security-best-practices)

---

## Overview

This document provides guidance on configuring SSL/TLS for PostgreSQL database connections in the Industrial AI Platform. Secure database connections are critical for protecting sensitive data in transit.

### Why SSL/TLS for Database Connections?

- **Data Protection**: Encrypts data in transit between application and database
- **Compliance**: Required by many security standards (PCI DSS, HIPAA, GDPR)
- **Man-in-the-Middle Prevention**: Ensures you're connecting to the correct database server
- **Security Hardening**: Part of defense-in-depth security strategy

---

## SSL Modes Explained

PostgreSQL supports several SSL modes with different security levels:

| Mode | Encryption | Certificate Verification | Use Case |
|------|------------|-------------------------|----------|
| `disable` | ❌ No | N/A | Development only (NOT recommended for production) |
| `allow` | Optional | No | Legacy compatibility |
| `prefer` | Optional | No | Default, tries SSL first |
| `require` | ✅ Yes | No | Minimum recommended for production |
| `verify-ca` | ✅ Yes | CA only | High security environments |
| `verify-full` | ✅ Yes | CA + Hostname | Maximum security |

### Mode Details

#### `disable` (NOT Recommended)
```
postgres://user:pass@host:5432/db?sslmode=disable
```
- **Risk**: Data transmitted in plain text
- **Use**: Only for local development without sensitive data

#### `require` (Minimum for Production)
```
postgres://user:pass@host:5432/db?sslmode=require
```
- **Pros**: Encrypts all traffic
- **Cons**: Doesn't verify server certificate
- **Risk**: Man-in-the-middle attacks possible

#### `verify-ca` (Recommended for Production)
```
postgres://user:pass@host:5432/db?sslmode=verify-ca&sslrootcert=/path/to/ca.crt
```
- **Pros**: Verifies server certificate against trusted CA
- **Cons**: Doesn't verify hostname matches certificate

#### `verify-full` (Maximum Security)
```
postgres://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/path/to/ca.crt
```
- **Pros**: Full certificate and hostname verification
- **Cons**: Requires proper certificate management
- **Recommendation**: Use for production environments with sensitive data

---

## Configuration Methods

### Method 1: Environment Variable

```bash
# .env or .env.production
DATABASE_URL="postgres://industrial_user:password@postgres:5432/industrial_ai?sslmode=require"
```

### Method 2: Connection String with SSL Parameters

```bash
# Basic SSL (require)
DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require"

# With CA certificate verification
DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=verify-ca&sslrootcert=/etc/ssl/certs/postgres-ca.crt"

# Full verification
DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/etc/ssl/certs/postgres-ca.crt&sslcert=/etc/ssl/certs/client.crt&sslkey=/etc/ssl/private/client.key"
```

### Method 3: Kubernetes Secret

```yaml
# infra/k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: industrial-ai-secrets
  namespace: industrial-ai
type: Opaque
stringData:
  # Option 1: Basic SSL (minimum for production)
  DATABASE_URL: "postgres://user:pass@postgres:5432/industrial_ai?sslmode=require"
  
  # Option 2: With CA verification
  DATABASE_URL: "postgres://user:pass@postgres:5432/industrial_ai?sslmode=verify-ca"
---
# Separate CA certificate (if using verify-ca or verify-full)
apiVersion: v1
kind: Secret
metadata:
  name: postgres-ssl-certs
  namespace: industrial-ai
type: Opaque
data:
  # Base64 encoded CA certificate
  ca.crt: <base64-encoded-ca-certificate>
  # Optional: Client certificate for mutual TLS
  client.crt: <base64-encoded-client-certificate>
  client.key: <base64-encoded-client-key>
```

### Method 4: Volume Mount for Certificates

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  template:
    spec:
      volumes:
        - name: postgres-ssl
          secret:
            secretName: postgres-ssl-certs
            defaultMode: 0400  # Read-only for owner
      containers:
        - name: backend
          env:
            - name: DATABASE_URL
              value: "postgres://user:pass@postgres:5432/db?sslmode=verify-full&sslrootcert=/etc/ssl/postgres/ca.crt"
          volumeMounts:
            - name: postgres-ssl
              mountPath: /etc/ssl/postgres
              readOnly: true
```

---

## Production Recommendations

### Recommended Configuration

For production environments, use **`sslmode=verify-full`** with:

1. **CA Certificate**: Verify server identity
2. **Client Certificate** (optional): Mutual TLS authentication
3. **Certificate Rotation**: Automated certificate renewal

### Step-by-Step Production Setup

#### Step 1: Obtain SSL Certificates

**Option A: Let's Encrypt (for public databases)**
```bash
# Install certbot
sudo apt-get install certbot

# Obtain certificate
sudo certbot certonly --standalone -d db.yourdomain.com
```

**Option B: Internal CA**
```bash
# Generate CA (if not exists)
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt

# Generate server certificate
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# Generate client certificate (for mutual TLS)
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
```

#### Step 2: Configure PostgreSQL Server

```bash
# /etc/postgresql/15/main/postgresql.conf
ssl = on
ssl_cert_file = '/etc/postgresql/ssl/server.crt'
ssl_key_file = '/etc/postgresql/ssl/server.key'
ssl_ca_file = '/etc/postgresql/ssl/ca.crt'  # For mutual TLS
```

```bash
# /etc/postgresql/15/main/pg_hba.conf
# Require SSL for all connections
hostssl all all 0.0.0.0/0 md5

# For mutual TLS (client certificate authentication)
hostssl all all 0.0.0.0/0 cert map=cn
```

#### Step 3: Create Kubernetes Secrets

```bash
# Create secret with CA certificate
kubectl create secret generic postgres-ssl-certs \
  --namespace=industrial-ai \
  --from-file=ca.crt=./ca.crt \
  --from-file=client.crt=./client.crt \
  --from-file=client.key=./client.key

# Update database URL secret
kubectl create secret generic industrial-ai-secrets \
  --namespace=industrial-ai \
  --from-literal=database-url="postgres://industrial_user:password@postgres:5432/industrial_ai?sslmode=verify-full&sslrootcert=/etc/ssl/postgres/ca.crt" \
  --dry-run=client -o yaml | kubectl apply -f -
```

#### Step 4: Update Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  template:
    spec:
      volumes:
        - name: postgres-ssl
          secret:
            secretName: postgres-ssl-certs
            items:
              - key: ca.crt
                path: ca.crt
              - key: client.crt
                path: client.crt
              - key: client.key
                path: client.key
                mode: 0400
      containers:
        - name: backend
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: industrial-ai-secrets
                  key: database-url
          volumeMounts:
            - name: postgres-ssl
              mountPath: /etc/ssl/postgres
              readOnly: true
```

---

## Troubleshooting

### Common SSL Connection Errors

#### Error: `no pg_hba.conf entry for host`

```
FATAL: no pg_hba.conf entry for host "10.0.0.5", user "industrial_user", database "industrial_ai", SSL on
```

**Solution**: Ensure PostgreSQL is configured to accept SSL connections:
```bash
# pg_hba.conf
hostssl industrial_ai industrial_user 0.0.0.0/0 md5
```

#### Error: `certificate verify failed`

```
SSL error: certificate verify failed
```

**Solutions**:
1. Check CA certificate is correct
2. Verify certificate hasn't expired
3. Ensure hostname matches certificate CN/SAN

```bash
# Verify certificate
openssl x509 -in ca.crt -text -noout
openssl verify -CAfile ca.crt server.crt
```

#### Error: `SSL SYSCALL error`

```
SSL SYSCALL error: EOF detected
```

**Solutions**:
1. Check PostgreSQL SSL configuration
2. Verify certificates are readable
3. Check network connectivity

### Testing SSL Connection

```bash
# Test SSL connection
psql "postgres://user:pass@host:5432/db?sslmode=require"

# With certificate verification
psql "postgres://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/path/to/ca.crt"

# Check SSL status in psql
\conninfo

# Output should show:
# SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, ...)
```

---

## Security Best Practices

### Certificate Management

1. **Rotation Schedule**: Rotate certificates before expiry
   ```bash
   # Check certificate expiry
   openssl x509 -in server.crt -noout -dates
   
   # Set up monitoring for certificate expiry
   # Add to Prometheus alerts
   ```

2. **Private Key Protection**
   ```bash
   # Set correct permissions
   chmod 400 /etc/postgresql/ssl/server.key
   chown postgres:postgres /etc/postgresql/ssl/server.key
   ```

3. **Certificate Storage**
   - Never commit certificates to Git
   - Use Kubernetes Secrets or external secret management (Vault, AWS Secrets Manager)
   - Enable encryption at rest for Kubernetes secrets

### Connection Security

1. **Disable Non-SSL Connections**
   ```bash
   # postgresql.conf
   ssl = on
   
   # pg_hba.conf - remove 'host' entries, keep only 'hostssl'
   # host all all 0.0.0.0/0 md5  # REMOVE THIS
   hostssl all all 0.0.0.0/0 md5  # KEEP THIS
   ```

2. **Use Strong Cipher Suites**
   ```bash
   # postgresql.conf
   ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL'
   ssl_prefer_server_ciphers = on
   ```

3. **Configure TLS Protocol**
   ```bash
   # postgresql.conf (PostgreSQL 15+)
   ssl_min_protocol_version = 'TLSv1.2'
   ssl_max_protocol_version = 'TLSv1.3'
   ```

### Monitoring and Alerting

Add alerts for:

```yaml
# Prometheus alerting rules
groups:
  - name: postgres_ssl
    rules:
      - alert: PostgresSSLCertificateExpiringSoon
        expr: postgres_ssl_cert_expiry_days < 30
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "PostgreSQL SSL certificate expiring soon"
          description: "Certificate expires in {{ $value }} days"
      
      - alert: PostgresSSLConnectionErrors
        expr: rate(postgres_ssl_errors_total[5m]) > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "PostgreSQL SSL connection errors detected"
```

---

## Quick Reference

### Configuration Checklist

- [ ] PostgreSQL SSL enabled (`ssl = on`)
- [ ] CA certificate generated and distributed
- [ ] Server certificate installed on PostgreSQL
- [ ] Client certificate available to application (if using mutual TLS)
- [ ] Connection string includes `sslmode=require` or higher
- [ ] Non-SSL connections disabled in `pg_hba.conf`
- [ ] Certificate rotation process documented
- [ ] Monitoring for certificate expiry configured

### Connection String Templates

```bash
# Development (minimum)
postgres://user:pass@localhost:5432/db?sslmode=prefer

# Staging (recommended)
postgres://user:pass@postgres:5432/db?sslmode=require

# Production (high security)
postgres://user:pass@postgres:5432/db?sslmode=verify-full&sslrootcert=/etc/ssl/postgres/ca.crt

# Production with Mutual TLS
postgres://user:pass@postgres:5432/db?sslmode=verify-full&sslrootcert=/etc/ssl/postgres/ca.crt&sslcert=/etc/ssl/postgres/client.crt&sslkey=/etc/ssl/postgres/client.key
```

---

## References

- [PostgreSQL SSL Documentation](https://www.postgresql.org/docs/current/libpq-ssl.html)
- [PostgreSQL TLS Configuration](https://www.postgresql.org/docs/current/ssl-tcp.html)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [cert-manager for Kubernetes](https://cert-manager.io/docs/)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-28 | DevOps Team | Initial document |