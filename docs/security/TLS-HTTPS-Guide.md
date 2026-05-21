# TLS/HTTPS 配置指南

> **Industrial AI Platform SSL/TLS 安全配置**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 概述

生产环境必须强制 HTTPS，确保：
- 所有 API 通信加密
- WebSocket 连接加密 (WSS)
- 防止中间人攻击
- 符合安全合规要求

---

## 🔐 证书方案

### 方案 A: Let's Encrypt (推荐 - 免费)

**适用场景**: 公网域名、云部署

```bash
# 1. 安装 certbot
apt install certbot python3-certbot-nginx

# 2. 申请证书 (自动化)
certbot --nginx -d industrial-ai.example.com -d api.industrial-ai.example.com

# 3. 自动续期
certbot renew --dry-run

# 4. 配置自动续期 cron
echo "0 0 1 * * certbot renew --quiet" | crontab -
```

**证书位置**:
- `/etc/letsencrypt/live/industrial-ai.example.com/fullchain.pem` (证书链)
- `/etc/letsencrypt/live/industrial-ai.example.com/privkey.pem` (私钥)

---

### 方案 B: 企业证书 (商业环境)

**适用场景**: 企业内部、需要企业 CA 签名

```bash
# 1. 准备 CSR 文件
openssl req -new -newkey rsa:4096 -nodes \
  -keyout industrial-ai.key \
  -out industrial-ai.csr \
  -subj "/C=CN/ST=Shanghai/L=Shanghai/O=IndustrialAI/CN=industrial-ai.example.com"

# 2. 提交 CSR 到企业 CA 签名

# 3. 获取签名后的证书 industrial-ai.crt

# 4. 合并证书链
cat industrial-ai.crt intermediate.crt root.crt > fullchain.pem
```

---

### 方案 C: 自签名证书 (开发/测试)

**适用场景**: 开发环境、内网测试

```bash
# 使用提供的脚本生成
./scripts/generate-dev-certs.sh

# 或手动生成
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes \
  -subj "/C=CN/ST=Shanghai/L=Shanghai/O=IndustrialAI-Dev/CN=localhost"
```

---

## 🔧 Nginx HTTPS 配置

### 配置文件: `frontend/nginx-ssl.conf`

```nginx
# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name industrial-ai.example.com api.industrial-ai.example.com;
    
    # 强制 HTTPS
    return 301 https://$host$request_uri;
}

# HTTPS 主配置
server {
    listen 443 ssl http2;
    server_name industrial-ai.example.com api.industrial-ai.example.com;
    
    # TLS 1.3 only (推荐)
    ssl_protocols TLSv1.3 TLSv1.2;
    ssl_prefer_server_ciphers on;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    
    # 证书配置
    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    
    # HSTS (强制 HTTPS 1 年)
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    
    # 安全头部
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # SSL 会话优化
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # 前端静态文件
    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
    
    # API 代理
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
    }
    
    # WebSocket 代理 (WSS)
    location /ws {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
    
    # 健康检查
    location /health {
        proxy_pass http://backend:8080/health;
        access_log off;
    }
}
```

---

## 🔧 Gin 后端 HTTPS 配置

### 方式 1: Gin 直接 HTTPS (不推荐 - 通过 Nginx 更好)

```go
// 在 server.go 中
func runTLSServer() {
    r := gin.Default()
    
    // 加载证书
    r.RunTLS(":443", 
        "/etc/ssl/industrial-ai/fullchain.pem",
        "/etc/ssl/industrial-ai/privkey.pem")
}
```

### 方式 2: 通过 Nginx 代理 (推荐)

后端监听 HTTP (8080)，Nginx 负责 HTTPS 加密：

```yaml
# docker-compose.yml
backend:
  environment:
    - TLS_ENABLED=false  # 后端不直接 TLS
    - BEHIND_PROXY=true  # 通过代理，读取 X-Forwarded-Proto
```

---

## 🛡️ 安全头部配置

### Gin Middleware: Security Headers

```go
// backend/internal/middleware/security.go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "SAMEORIGIN")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        // CSP (Content Security Policy)
        c.Header("Content-Security-Policy", 
            "default-src 'self'; " +
            "script-src 'self' 'unsafe-inline'; " +
            "style-src 'self' 'unsafe-inline'; " +
            "img-src 'self' data: https:; " +
            "connect-src 'self' wss: https:;")
        
        c.Next()
    }
}
```

---

## 📦 Docker Compose HTTPS 配置

### `docker-compose.prod.yml`

```yaml
services:
  frontend:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./frontend/nginx-ssl.conf:/etc/nginx/conf.d/default.conf:ro
      - ./frontend/dist:/usr/share/nginx/html:ro
      - ./ssl:/etc/nginx/ssl:ro  # SSL 证书目录
    environment:
      - NGINX_SSL_ENABLED=true

  backend:
    environment:
      - TLS_MODE=behind_proxy  # 通过代理
      - FORCE_HTTPS_REDIRECT=true  # 内部重定向检查
```

---

## 🔗 Kubernetes TLS 配置

### Ingress TLS

```yaml
# infra/k8s/ingress-tls.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: industrial-ai-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-ssl-verify: "on"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - industrial-ai.example.com
        - api.industrial-ai.example.com
      secretName: industrial-ai-tls-secret
  rules:
    - host: industrial-ai.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: frontend
                port:
                  number: 80
    - host: api.industrial-ai.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: backend
                port:
                  number: 8080
```

### 创建 TLS Secret

```bash
# 从证书文件创建 Secret
kubectl create secret tls industrial-ai-tls-secret \
  --cert=fullchain.pem \
  --key=privkey.pem \
  -n industrial-ai

# 或使用 cert-manager 自动管理
kubectl apply -f infra/k8s/cert-manager-issuer.yaml
```

---

## ✅ 验证 HTTPS

### 1️⃣ 证书验证

```bash
# 检查证书有效性
openssl s_client -connect industrial-ai.example.com:443 -servername industrial-ai.example.com

# 检查证书详情
openssl x509 -in fullchain.pem -text -noout

# 检查证书过期时间
openssl x509 -in fullchain.pem -noout -dates
```

### 2️⃣ TLS 版本验证

```bash
# 验证 TLS 1.3 支持
openssl s_client -connect industrial-ai.example.com:443 -tls1_3

# 验证 TLS 1.2 支持
openssl s_client -connect industrial-ai.example.com:443 -tls1_2

# 验证 TLS 1.0/1.1 不支持 (应失败)
openssl s_client -connect industrial-ai.example.com:443 -tls1
# Expected: handshake failure
```

### 3️⃣ 安全测试

```bash
# 使用 testssl.sh 测试
testssl.sh industrial-ai.example.com

# 或使用 SSL Labs 测试
https://www.ssllabs.com/ssltest/analyze.html?d=industrial-ai.example.com

# 期望结果: A 或 A+ 评级
```

### 4️⃣ 强制 HTTPS 验证

```bash
# HTTP 应重定向到 HTTPS
curl -I http://industrial-ai.example.com
# Expected: 301 Redirect to https://

# 验证 HTTPS 可访问
curl -I https://industrial-ai.example.com
# Expected: 200 OK
```

---

## 🔒 安全最佳实践

| 配置项 | 推荐值 | 说明 |
|--------|--------|------|
| **TLS 版本** | TLS 1.3 only | 禁用 TLS 1.0/1.1 |
| **加密套件** | ECDHE + AES-GCM | 前向保密 + 强加密 |
| **HSTS** | max-age=31536000 | 1 年强制 HTTPS |
| **证书有效期** | 90 天 (Let's Encrypt) | 短期证书更安全 |
| **私钥保护** | 400 权限 | 仅 root 可读 |
| **OCSP Stapling** | 启用 | 减少证书验证延迟 |

---

## 📝 证书管理检查清单

```bash
# 1. 证书文件权限
chmod 400 privkey.pem
chmod 644 fullchain.pem

# 2. 证书自动续期
certbot renew --dry-run  # 测试续期

# 3. 证书监控告警
# Prometheus 监控证书过期时间

# 4. 证书备份
cp /etc/letsencrypt/live/industrial-ai.example.com/*.pem /backup/ssl/

# 5. 灾难恢复
# 证书丢失时，重新申请
certbot certonly --nginx -d industrial-ai.example.com
```

---

**最后更新**: 2026-05-13  
**审核人**: 安全团队