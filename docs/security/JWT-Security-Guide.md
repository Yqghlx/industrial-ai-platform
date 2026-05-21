# JWT 安全配置指南

> **Industrial AI Platform JWT 认证安全最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 JWT 安全增强概述

Phase 4 JWT 安全加固包含以下改进：

| 功能 | 描述 | 安全收益 |
|------|------|---------|
| **短期 Access Token** | 15 分钟有效期 | 减少被盗用风险 |
| **Refresh Token** | 7 天有效期 | 无需频繁登录 |
| **Token 黑名单** | Redis 缓存注销 Token | 注销/修改密码后立即失效 |
| **签名验证增强** | HMAC-SHA256 + Issuer 验证 | 防止伪造 Token |
| **Tenant ID 支持** | 多租户隔离 | 防止跨租户访问 |
| **Token ID (JWT ID)** | 唯一标识每个 Token | 支持精准黑名单 |

---

## 🔐 Token 有效期策略

### Access Token (访问令牌)

```yaml
有效期: 15 分钟
用途: API 请求认证
存储: 客户端内存 (不持久化)
安全:
  - 短有效期减少被盗用风险
  - 即使泄露，15 分钟后自动失效
  - 结合 HTTPS 防止中间人攻击
```

### Refresh Token (刷新令牌)

```yaml
有效期: 7 天
用途: 刷新 Access Token
存储: 客户端安全存储 (LocalStorage/SessionStorage)
安全:
  - 仅用于 /auth/refresh 端点
  - 不能用于其他 API
  - 注销时加入黑名单
  - 可配置单次使用 (rotation)
```

---

## 🔄 Token 刷新流程

```
┌─────────────┐                 ┌─────────────┐
│   Client    │                 │   Server    │
└──────┬──────┘                 └──────┬──────┘
       │                                │
       │  1. Login                      │
       │────────────────────────────────>│
       │                                │
       │  2. Access Token + Refresh Token│
       │<────────────────────────────────│
       │                                │
       │  3. API Request + Access Token  │
       │────────────────────────────────>│
       │                                │
       │  4. Response                    │
       │<────────────────────────────────│
       │                                │
       │  ... (14 分钟后)                 │
       │                                │
       │  5. Access Token Expired (401)  │
       │<────────────────────────────────│
       │                                │
       │  6. Refresh Token Request       │
       │────────────────────────────────>│
       │                                │
       │  7. New Access Token + New Refresh Token│
       │<────────────────────────────────│
       │                                │
       │  8. API Request + New Access Token│
       │────────────────────────────────>│
       │                                │
```

---

## 🛡️ Token 黑名单机制

### Redis 黑名单实现

```go
// Token 黑名单 Redis key 格式
Key: jwt_blacklist:<token_id>
Value: "1"
TTL: Token 剩余有效期
```

### 使用场景

| 场景 | 操作 | 说明 |
|------|------|------|
| **用户注销** | Access + Refresh Token 加入黑名单 | 立即失效 |
| **修改密码** | 所有用户 Token 加入黑名单 | 强制重新登录 |
| **管理员禁用用户** | 所有用户 Token 加入黑名单 | 立即失效 |
| **检测异常登录** | 相关 Token 加入黑名单 | 安全响应 |

---

## 🔧 API 端点

### 登录端点 (增强版)

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}

Response 200:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 900,         // 15 分钟 = 900 秒
  "token_type": "Bearer",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "tenant_id": "tenant-001"
  }
}
```

### 刷新 Token 端点

```http
POST /api/v1/auth/refresh
Content-Type: application/json
Authorization: Bearer <refresh_token>

Response 200:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",  // 新的 refresh token
  "expires_in": 900,
  "token_type": "Bearer"
}

Response 401 (Refresh Token 无效):
{
  "error": "invalid refresh token"
}
```

### 注销端点

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>

Response 200:
{
  "message": "logged out successfully"
}

// 同时将 Access Token 和 Refresh Token 加入黑名单
```

---

## 📝 JWT Claims 结构

### Access Token Claims

```json
{
  "user_id": 1,
  "username": "admin",
  "role": "admin",
  "tenant_id": "tenant-001",
  "token_type": "access",
  "token_id": "1234567890-123",
  "iss": "industrial-ai-platform",
  "sub": "user:1",
  "iat": 1715623200,
  "nbf": 1715623200,
  "exp": 1715624100,    // 15 分钟后
  "jti": "1234567890-123"
}
```

### Refresh Token Claims

```json
{
  "user_id": 1,
  "username": "admin",
  "role": "admin",
  "tenant_id": "tenant-001",
  "token_type": "refresh",
  "token_id": "1234567890-456",
  "iss": "industrial-ai-platform",
  "sub": "user:1",
  "iat": 1715623200,
  "nbf": 1715623200,
  "exp": 1716228000,    // 7 天后
  "jti": "1234567890-456"
}
```

---

## 🔒 安全最佳实践

### 1️⃣ Token 存储建议

**客户端**:
- Access Token: 内存变量 (不持久化)
- Refresh Token: LocalStorage (配合 XSS 防护)

**服务端**:
- JWT Secret: 环境变量 (不硬编码)
- Token 黑名单: Redis (快速查询)

### 2️⃣ Token 传输建议

```yaml
必须: HTTPS 加密传输
禁止: URL 参数传递 Token
推荐: Authorization Header
格式: Authorization: Bearer <token>
```

### 3️⃣ Token 验证流程

```go
// 服务端验证步骤:
1. 解析 Token
2. 验证签名算法 (必须 HMAC-SHA256)
3. 验证签名
4. 验证 Issuer
5. 验证过期时间
6. 验证生效时间 (nbf)
7. 检查黑名单
8. 返回 Claims
```

### 4️⃣ 异常处理建议

| 异常 | 处理方式 | 客户端动作 |
|------|---------|-----------|
| **Token 过期** | 返回 401 + "token expired" | 使用 Refresh Token 刷新 |
| **Token 无效** | 返回 401 + "invalid token" | 重新登录 |
| **Token 已注销** | 返回 401 + "token revoked" | 重新登录 |
| **签名错误** | 返回 401 + "invalid signature" | 重新登录 |

---

## ⚙️ 配置参数

### 环境变量

```bash
# JWT 配置
JWT_SECRET=your-very-long-random-secret-key-at-least-32-characters
JWT_ACCESS_TOKEN_DURATION=15m    # Access Token 有效期
JWT_REFRESH_TOKEN_DURATION=7d    # Refresh Token 有效期
JWT_ISSUER=industrial-ai-platform

# Redis 配置 (Token 黑名单)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1                       # Token 黑名单专用 DB
```

### 配置文件

```yaml
# config.yaml
jwt:
  secret: ${JWT_SECRET}
  access_token_duration: 15m
  refresh_token_duration: 7d
  issuer: industrial-ai-platform
  
redis:
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  password: ${REDIS_PASSWORD}
  db: 1
  blacklist_prefix: jwt_blacklist:
```

---

## ✅ 安全验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **Access Token 有效期** | ≤ 15 分钟 | 检查 Claims.exp |
| **Refresh Token 有效期** | ≤ 7 天 | 检查 Claims.exp |
| **JWT Secret** | ≥ 32 字符随机 | 检查环境变量 |
| **签名算法** | HMAC-SHA256 | 检查 Header.alg |
| **Issuer 验证** | 必须 | 检查 Claims.iss |
| **黑名单生效** | 注销后立即 | 测试注销后访问 |
| **HTTPS 传输** | 必须 | 检查协议 |

---

## 🔗 相关代码

- `backend/internal/service/auth_helpers.go` - JWT 生成/验证
- `backend/internal/middleware/auth.go` - 认证中间件
- `backend/internal/handler/auth_handler.go` - 认证 API

---

**最后更新**: 2026-05-13  
**审核人**: 安全团队