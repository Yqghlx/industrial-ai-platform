# CORS/CSRF 安全配置指南

> **Industrial AI Platform 跨域和跨站请求安全最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 CORS/CSRF 安全概述

| 功能 | 描述 | 安全收益 |
|------|------|---------|
| **CORS 白名单** | 仅允许指定域名跨域访问 | 防止未授权跨域请求 |
| **CORS 凭证控制** | 严格控制 Cookie 发送 | 防止跨域 Cookie 泄露 |
| **CORS 方法限制** | 仅允许必要 HTTP 方法 | 减少攻击面 |
| **CORS 头部限制** | 仅允许必要请求头 | 防止头部注入 |
| **CSRF Token** | 表单请求 CSRF 验证 | 防止跨站请求伪造 |
| **SameSite Cookie** | Cookie SameSite 属性 | 防止 CSRF 攻击 |

---

## 🌐 CORS (跨域资源共享)

### CORS 配置策略

**生产环境 CORS 配置原则：**

1. **白名单域名**: 仅允许已知的可信域名
2. **精确匹配**: 不使用通配符 `*` (生产环境)
3. **凭证控制**: 仅必要时允许 `credentials`
4. **方法限制**: 仅允许必要的 HTTP 方法
5. **头部限制**: 仅允许必要的请求头

### CORS 配置示例

```yaml
# config.yaml (生产环境)
cors:
  allowed_origins:
    - "https://industrial-ai.example.com"
    - "https://admin.industrial-ai.example.com"
    - "*.industrial-ai.local"  # 内网开发环境 (支持通配符)
  
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "PATCH"
  
  allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Requested-With"
    - "X-Tenant-ID"
    - "Accept-Language"
  
  exposed_headers:
    - "Content-Length"
    - "X-Total-Count"
    - "X-Request-ID"
  
  allow_credentials: true
  max_age: 86400  # 24 小时预检缓存
```

### CORS 中间件实现

```go
// backend/internal/middleware/security.go
func CORSSecurity(config *CORSConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        
        // 1. 验证 Origin 白名单
        if !isOriginAllowed(origin, config.AllowedOrigins) {
            c.AbortWithStatus(403)
            return
        }
        
        // 2. 设置 CORS 头部
        c.Header("Access-Control-Allow-Origin", origin)
        c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
        c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
        
        if config.AllowCredentials {
            c.Header("Access-Control-Allow-Credentials", "true")
        }
        
        // 3. 处理 OPTIONS 预检请求
        if c.Request.Method == "OPTIONS" {
            c.Header("Access-Control-Max-Age", itoa(config.MaxAge))
            c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

---

## 🔒 CSRF (跨站请求伪造)

### JWT 认证下的 CSRF 防护

**重要**: Industrial AI Platform 使用 JWT Token 认证，Token 存储在客户端内存或 LocalStorage 中，**不存储在 Cookie 中**。因此 CSRF 攻击风险较低。

**但以下场景仍需要 CSRF 防护：**

| 场景 | CSRF 防护方式 |
|------|-------------|
| **Session-based 认证** | CSRF Token 必需 |
| **Cookie 存储 Token** | CSRF Token 必需 |
| **表单提交** | CSRF Token 推荐 |
| **敏感操作 (删除/修改)** | CSRF Token + 双重验证 |

### CSRF Token 实现

```go
// backend/internal/middleware/csrf.go
func CSRFProtection() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 仅对需要 CSRF 防护的方法验证
        if needsCSRFProtection(c.Request.Method) {
            // 从请求中获取 CSRF Token
            csrfToken := c.GetHeader("X-CSRF-Token")
            if csrfToken == "" {
                csrfToken = c.PostForm("csrf_token")
            }
            
            // 从 Session 或 Cookie 中获取期望的 Token
            expectedToken := getCSRFTokenFromSession(c)
            
            // 验证 Token
            if csrfToken == "" || csrfToken != expectedToken {
                c.JSON(403, gin.H{"error": "CSRF token invalid"})
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}

func needsCSRFProtection(method string) bool {
    // POST, PUT, DELETE, PATCH 需要 CSRF 验证
    protectedMethods := []string{"POST", "PUT", "DELETE", "PATCH"}
    for _, m := range protectedMethods {
        if method == m {
            return true
        }
    }
    return false
}
```

### CSRF Token 生成

```go
// 生成 CSRF Token
func GenerateCSRFToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

// 设置 CSRF Token 到 Cookie
func SetCSRFTokenCookie(c *gin.Context, token string) {
    c.SetCookie("csrf_token", token, 3600, "/", "", true, true)
    // SameSite=Strict, HttpOnly=true, Secure=true
}
```

---

## 🍪 Cookie 安全配置

### SameSite Cookie 属性

| SameSite 值 | 行为 | 安全级别 |
|-------------|------|---------|
| **Strict** | 仅同站请求发送 Cookie | ✅ 最高 |
| **Lax** | 同站 + 安全跨站导航发送 | ✅ 推荐 |
| **None** | 所有跨站请求发送 Cookie | ⚠️ 需配合 Secure |

### Cookie 安全配置

```go
// 设置安全 Cookie
func SetSecureCookie(c *gin.Context, name, value string, maxAge int) {
    c.SetCookie(name, value, maxAge, "/", "", 
        true,  // Secure: 仅 HTTPS
        true,  // HttpOnly: 防止 JavaScript 读取
    )
    
    // 额外设置 SameSite (需要手动设置响应头)
    c.Header("Set-Cookie", 
        fmt.Sprintf("%s=%s; Path=/; Max-Age=%d; Secure; HttpOnly; SameSite=Strict",
            name, value, maxAge))
}
```

---

## 📝 安全检查清单

### CORS 检查清单

| 检查项 | 要求 | 状态 |
|--------|------|------|
| **Origin 白名单** | 精确域名列表，不用 `*` | ✅ 已实现 |
| **Credentials 控制** | 仅必要时 `true` | ✅ 已实现 |
| **方法限制** | 仅必要方法 | ✅ 已实现 |
| **头部限制** | 仅必要头部 | ✅ 已实现 |
| **预检缓存** | max_age ≤ 24h | ✅ 已实现 |
| **通配符子域名** | 仅开发环境使用 | ✅ 已实现 |

### CSRF 检查清单

| 检查项 | 要求 | 状态 |
|--------|------|------|
| **JWT 不存 Cookie** | Token 不在 Cookie | ✅ 已实现 |
| **敏感操作验证** | DELETE/重要修改需额外验证 | ⏳ 可选 |
| **CSRF Token** | 表单场景提供 | ✅ 已实现 |
| **SameSite Cookie** | Strict 或 Lax | ✅ 已实现 |

---

## 🔧 配置示例

### 环境变量配置

```bash
# CORS 配置
CORS_ALLOWED_ORIGINS=https://industrial-ai.example.com,https://admin.industrial-ai.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,PATCH
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Tenant-ID
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400

# CSRF 配置
CSRF_ENABLED=false  # JWT 认证下默认不启用
CSRF_TOKEN_LENGTH=32
CSRF_COOKIE_NAME=csrf_token
```

### Docker Compose 配置

```yaml
# docker-compose.yml
services:
  backend:
    environment:
      - CORS_ALLOWED_ORIGINS=https://industrial-ai.example.com
      - CORS_ALLOW_CREDENTIALS=true
      - CORS_MAX_AGE=86400
```

---

## ✅ 验证 CORS 配置

### 1️⃣ 测试允许的 Origin

```bash
# 允许的 Origin 应返回 CORS 头部
curl -I -X OPTIONS https://api.industrial-ai.example.com/api/v1/devices \
  -H "Origin: https://industrial-ai.example.com" \
  -H "Access-Control-Request-Method: GET"

# 期望: Access-Control-Allow-Origin: https://industrial-ai.example.com
```

### 2️⃣ 测试禁止的 Origin

```bash
# 禁止的 Origin 应被拒绝
curl -I -X OPTIONS https://api.industrial-ai.example.com/api/v1/devices \
  -H "Origin: https://malicious-site.com" \
  -H "Access-Control-Request-Method: GET"

# 期望: 无 CORS 头部 或 403 Forbidden
```

### 3️⃣ 测试通配符子域名

```bash
# 内网子域名应允许
curl -I -X OPTIONS https://api.industrial-ai.example.com/api/v1/devices \
  -H "Origin: https://dev.industrial-ai.local"

# 期望: Access-Control-Allow-Origin: https://dev.industrial-ai.local
```

---

**最后更新**: 2026-05-13  
**审核人**: 安全团队