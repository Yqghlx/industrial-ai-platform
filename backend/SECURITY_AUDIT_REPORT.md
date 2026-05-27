# Backend Security Audit Report

**审计日期**: 2026-05-27  
**审计范围**: industrial-ai-platform/backend  
**审计目标**: SQL注入、XSS、JWT问题、CORS、硬编码密钥、输入验证

---

## 执行摘要 (Executive Summary)

本次审计对 backend 代码库进行了全面的安全漏洞扫描，重点关注认证、授权、数据库操作、API handler、配置和 secrets 管理。

**总体评估**: 项目整体安全态势良好，已实施了多项安全措施（bcrypt密码哈希、JWT算法验证、WAF中间件、表名白名单等）。但仍存在若干需要修复的安全问题。

**漏洞统计**:
- CRITICAL: 1 个
- HIGH: 2 个
- MEDIUM: 4 个
- LOW: 3 个

---

## CRITICAL - 立即修复的安全漏洞

### CVE-IAP-001: docker-compose.yml 中明文密码泄露

**风险等级**: CRITICAL  
**影响范围**: 生产环境部署  
**文件位置**: `docker-compose.yml` 第10行

**问题描述**:
```yaml
DATABASE_URL: postgres://postgres:***@postgres:5432/industrial_ai?sslmode=disable
```

虽然 `***` 是遮盖后的显示，但原始配置中 `DATABASE_URL` 行包含明文密码。该密码与 `${POSTGRES_PASSWORD}` 环境变量重复暴露，增加了密码泄露风险。同时 `sslmode=disable` 表示数据库连接未强制SSL加密。

**风险分析**:
- 数据库密码可能通过配置文件泄露到版本控制系统
- 未加密的数据库连接可能被中间人攻击截获
- CWE-256: Unprotected Storage of Credentials
- CWE-319: Cleartext Transmission of Sensitive Information

**修复建议**:
1. 移除 `DATABASE_URL` 中的明文密码，使用 `${DATABASE_URL}` 环境变量占位符
2. 生产环境强制使用 `sslmode=require` 或 `verify-full`
3. 确保 `.env` 文件不被提交到版本控制
4. 使用 secrets 管理系统（如 Docker Secrets、Kubernetes Secrets）

```yaml
# 修复后的配置示例
environment:
  DATABASE_URL: ${DATABASE_URL}  # 从环境变量读取完整URL
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
```

---

## HIGH - 重要安全问题

### CVE-IAP-002: JWT密钥测试用例使用弱密钥

**风险等级**: HIGH  
**影响范围**: 开发/测试环境、潜在的生产配置误用  
**文件位置**: `pkg/auth/auth_test.go` 多处

**问题描述**:
测试代码中使用 `"test-secret"` 作为JWT密钥：
```go
secret := []byte("test-secret")  // 第27、35、109行等多处
```

虽然是测试代码，但这种模式可能被开发人员复制到生产配置中。代码中已设置 `MinSecretLength = 32`，但测试代码使用仅12字符的密钥。

**风险分析**:
- 弱密钥容易被暴力破解
- 测试模式可能被误用于生产环境
- CWE-321: Use of Hard-coded Cryptographic Key
- CWE-326: Inadequate Encryption Strength

**修复建议**:
1. 测试代码使用符合生产标准长度的随机密钥
2. 添加测试密钥长度验证的断言
3. 在CI/CD流程中添加密钥强度检测

```go
// 修复后的测试代码示例
secret := []byte("test-secret-key-minimum-32-characters-long-for-security")  // 52字符
if len(secret) < MinSecretLength {
    t.Skip("Test secret too short")
}
```

### CVE-IAP-003: 开发环境CORS过度宽松可能导致生产误部署

**风险等级**: HIGH  
**影响范围**: 跨域安全策略  
**文件位置**: `internal/middleware/cors.go` 第85-107行

**问题描述**:
```go
if len(allowedOrigins) == 0 {
    if gin.Mode() == gin.DebugMode {
        // 开发模式：允许所有 origin（但记录警告）
        c.Header("Access-Control-Allow-Origin", origin)
        ...
    }
}
```

开发模式下当 `allowedOrigins` 为空时，允许任何Origin访问API。虽然生产模式已正确拒绝，但如果部署时忘记切换 `gin.Mode()` 为 `ReleaseMode`，将导致严重的CORS漏洞。

**风险分析**:
- 任意Origin可访问API，可能导致CSRF攻击
- 敏感数据可能被恶意网站获取
- CWE-942: Permissive Cross-domain Policy with Untrusted Domains

**修复建议**:
1. 在应用启动时验证运行模式与CORS配置的一致性
2. 添加环境检测警告：生产环境必须显式配置 `CORS_ORIGINS`
3. 在config验证中添加CORS检查（已部分实现）

---

## MEDIUM - 中等风险问题

### CVE-IAP-004: WebSocket Origin检查开发环境过于宽松

**风险等级**: MEDIUM  
**影响范围**: WebSocket连接安全  
**文件位置**: `internal/handler/server_new.go` 第213-224行

**问题描述**:
```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    // 开发环境允许所有 origin
    if !isProduction {
        return true
    }
    ...
}
```

开发环境WebSocket允许任意Origin连接，这可能导致跨站WebSocket劫持（CSWSH）攻击。

**风险分析**:
- CSWSH攻击可能窃取实时数据
- 恶意网站可建立WebSocket连接
- CWE-1021: Improper Restriction of Renderable UI Elements

**修复建议**:
1. 即使开发环境也应限制WebSocket Origin
2. 添加WebSocket连接来源日志记录
3. 生产环境添加Origin白名单验证

### CVE-IAP-005: Telemetry端点公开访问缺乏设备认证

**风险等级**: MEDIUM  
**影响范围**: IoT数据接入安全  
**文件位置**: `internal/handler/server_new.go` 第342行

**问题描述**:
```go
// SEC-MED-02: Telemetry endpoint with rate limiting and input validation
s.router.POST("/api/v1/devices/telemetry", middleware.TelemetryRateLimit(), s.telemetryHandler.IngestTelemetry)
```

遥测数据端点为公开端点，仅依赖速率限制和输入验证。虽然设计为设备接入端点，但缺乏设备身份验证机制。

**风险分析**:
- 伪造设备可提交虚假数据
- 数据污染可能影响AI分析准确性
- CWE-287: Improper Authentication

**修复建议**:
1. 实现设备API密钥认证机制
2. 添加设备注册和认证流程
3. 使用设备证书或预共享密钥

### CVE-IAP-006: 测试配置中的硬编码密码示例

**风险等级**: MEDIUM  
**影响范围**: 配置文件示例  
**文件位置**: `internal/config/config_test.go` 第133行

**问题描述**:
```go
_ = os.Setenv("ADMIN_PASSWORD", "admin123")  // 测试中使用简单密码
```

测试代码中使用 `admin123` 作为管理员密码示例，这种弱密码模式可能被复制到实际配置。

**风险分析**:
- 弱密码容易被猜测或暴力破解
- 测试模式可能影响开发环境配置
- CWE-798: Use of Hard-coded Credentials

**修复建议**:
1. 测试代码使用随机生成的强密码
2. 添加密码强度验证警告
3. 文档中强调生产环境密码要求

### CVE-IAP-007: 连接字符串构建日志可能泄露密码

**风险等级**: MEDIUM  
**影响范围**: 日志安全  
**文件位置**: `pkg/database/connection.go` 第89行

**问题描述**:
```go
log.Printf("Database connection established: %s:%d/%s", config.Host, config.Port, config.Database)
```

虽然日志中不直接输出密码，但连接字符串构建过程（第133-136行）包含密码拼接，如果在其他地方添加调试日志，可能泄露密码。

**风险分析**:
- 连接字符串可能被意外记录到日志
- 日志文件可能被未授权访问
- CWE-532: Insertion of Sensitive Information into Log File

**修复建议**:
1. 确保所有数据库相关日志过滤敏感信息
2. 使用结构化日志（zap）并添加敏感字段过滤
3. 定期审计日志输出内容

---

## LOW - 低风险问题

### CVE-IAP-008: 密码验证复杂度常量不一致

**风险等级**: LOW  
**影响范围**: 密码验证逻辑  
**文件位置**: `pkg/constants/constants.go` 第90行 vs `internal/model/auth_models.go` 第11行

**问题描述**:
`constants.go` 中定义 `MinPasswordLength = 12`，但 `auth_models.go` 也定义了相同常量。存在常量重复定义，可能导致维护不一致。

**风险分析**:
- 常量分散可能导致修改遗漏
- 代码维护复杂度增加
- CWE-1258: Exposure of Sensitive Information through Discrepancy

**修复建议**:
1. 统一密码相关常量到 `pkg/constants`
2. 其他模块引用统一常量
3. 添加常量一致性检测

### CVE-IAP-009: 部分输入验证缺失XSS输出编码

**风险等级**: LOW  
**影响范围**: API响应安全  
**文件位置**: `internal/handler/telemetry_handler_new.go` 第130-136行

**问题描述**:
```go
if strings.Contains(strings.ToLower(data.Message), "' or ") ||
    strings.Contains(strings.ToLower(data.Message), "' and ") ...
```

虽然实现了基础的SQL注入检测，但XSS防护依赖WAF中间件，在handler层缺乏显式的输出编码。

**风险分析**:
- 存储型XSS可能通过API响应传播
- 用户输入可能包含恶意脚本
- CWE-79: Cross-site Scripting

**修复建议**:
1. 对存储和输出的用户输入进行HTML编码
2. 添加Content-Security-Policy头部（已实现）
3. 审计所有用户输入存储和显示路径

### CVE-IAP-010: 错误消息可能暴露内部信息

**风险等级**: LOW  
**影响范围**: 信息泄露  
**文件位置**: 多处error响应

**问题描述**:
部分API错误响应包含详细的错误信息，如：
```go
response.BadRequest(c, err.Error())  // 可能暴露内部错误详情
```

**风险分析**:
- 详细错误信息可能暴露系统内部结构
- 有助于攻击者进行针对性攻击
- CWE-209: Generation of Error Message Containing Sensitive Information

**修复建议**:
1. 生产环境返回通用错误消息
2. 详细错误记录到日志而非响应
3. 区分开发和生产环境错误响应策略

---

## 已实施的安全措施 (Good Security Practices)

项目已实施以下安全措施，值得肯定：

### 1. bcrypt密码哈希 ✓
- `internal/service/auth_password.go` 使用成本因子12
- 符合2026年安全标准

### 2. JWT安全 ✓
- `pkg/auth/auth.go` 实现算法验证防止算法混淆攻击
- `internal/service/auth_init.go` 强制最小密钥长度32字符
- 使用HS256签名方法，拒绝"none"算法
- AccessToken 15分钟有效期，RefreshToken 7天有效期

### 3. SQL注入防护 ✓
- `internal/repository/base_repo.go` 实现表名白名单验证
- 所有数据库查询使用参数化查询（`$1, $2`）

### 4. WAF中间件 ✓
- `internal/middleware/waf.go` 实现完整的WAF
- 检测SQL注入、XSS、路径遍历、命令注入、SSRF
- User-Agent黑名单配置化

### 5. 速率限制 ✓
- `internal/middleware/ratelimit.go` 实现Token Bucket算法
- 登录、注册、遥测、WebSocket等端点均有特定限流

### 6. 安全头部 ✓
- `internal/middleware/security.go` 实现完整安全头部
- HSTS、X-Frame-Options、X-Content-Type-Options等

### 7. CORS安全配置 ✓
- `internal/config/config.go` 生产环境强制显式CORS配置
- 禁止生产环境使用通配符"*"

### 8. Token黑名单机制 ✓
- `internal/service/auth_blacklist.go` 实现Token撤销
- 支持Redis持久化和内存fallback

### 9. 密码复杂度验证 ✓
- `pkg/validation/uuid.go` 和 `internal/model/auth_models.go`
- 要求大小写字母、数字、特殊字符

### 10. Secrets生成脚本 ✓
- `scripts/generate-secrets.sh` 安全生成密钥
- 密钥不输出到stdout，写入权限限制的文件

---

## 修复优先级建议

| 优先级 | CVE编号 | 描述 | 预计修复时间 |
|--------|---------|------|--------------|
| P1 (立即) | CVE-IAP-001 | docker-compose.yml密码泄露 | 1小时 |
| P2 (1周内) | CVE-IAP-002 | JWT测试密钥强度 | 2小时 |
| P2 (1周内) | CVE-IAP-003 | CORS部署模式检查 | 4小时 |
| P3 (2周内) | CVE-IAP-004 | WebSocket Origin限制 | 4小时 |
| P3 (2周内) | CVE-IAP-005 | 设备认证机制 | 1-2天 |
| P4 (1月内) | CVE-IAP-006-010 | 低风险问题 | 1天 |

---

## 附录

### A. 审计工具使用
- 代码静态分析：手动审计 + grep 搜索模式
- 关键文件检查：认证、授权、数据库操作、配置

### B. 参考资料
- CWE (Common Weakness Enumeration): https://cwe.mitre.org/
- OWASP Top 10: https://owasp.org/www-project-top-ten/
- JWT Best Practices: https://datatracker.ietf.org/doc/html/rfc8725

### C. 审计文件清单
```
docker-compose.yml
pkg/auth/auth.go
pkg/auth/auth_test.go
internal/service/auth_jwt.go
internal/service/auth_init.go
internal/service/auth_password.go
internal/service/auth_service.go
internal/service/auth_blacklist.go
internal/repository/base_repo.go
internal/repository/device_repo.go
internal/middleware/auth.go
internal/middleware/cors.go
internal/middleware/waf.go
internal/middleware/ratelimit.go
internal/middleware/security.go
internal/middleware/jwt_helpers.go
internal/config/config.go
internal/config/config_test.go
internal/handler/server_new.go
internal/handler/telemetry_handler_new.go
internal/model/auth_models.go
pkg/constants/constants.go
pkg/validation/uuid.go
pkg/database/connection.go
pkg/database/security.go
scripts/generate-secrets.sh
```

---

**审计完成日期**: 2026-05-27  
**审计人**: Hermes Security Audit Agent  
**报告版本**: 1.0