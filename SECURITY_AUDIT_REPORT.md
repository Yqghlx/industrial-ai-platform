# Industrial AI Platform - 安全审计报告

**审计日期**: 2026-05-25
**审计范围**: Backend (Go) + Frontend (React/TypeScript)
**审计版本**: 主分支最新代码

---

## 执行摘要

本次安全审计对 Industrial AI Platform 项目进行了全面的安全漏洞扫描，重点关注 SQL 注入、XSS、JWT 安全、密码存储、硬编码密钥、敏感数据泄露和 CORS 配置等关键安全问题。

**总体评估**: 项目整体安全状况 **良好**，已经实施了多项安全最佳实践，但仍存在一些需要修复的问题。

| 风险级别 | 发现数量 | 状态 |
|---------|---------|------|
| CRITICAL | 2 | 需立即修复 |
| HIGH | 4 | 需优先修复 |
| MEDIUM | 3 | 建议修复 |
| LOW | 5 | 建议改进 |

---

## 🔴 CRITICAL 级别问题

### CRITICAL-01: Kubernetes Secret 硬编码弱密钥

**文件路径**: `kubernetes/backend-deployment.yaml`
**行号范围**: 118-127
**风险级别**: CRITICAL
**CVE 参考**: CWE-798 (Use of Hard-coded Credentials)

**问题描述**:
Kubernetes Secret 配置中包含硬编码的弱密钥和占位符值，这些值被提交到版本控制系统中。攻击者可直接获取这些密钥用于生产环境。

```yaml
stringData:
  database-url: "postgresql://user:***@postgres:5432/industrial_ai"
  jwt-secret: "change-me-in-production"
  llm-api-key: "your-llm-api-key"
```

**风险影响**:
- JWT 密钥泄露可导致任意用户身份伪造
- 数据库连接字符串泄露可导致数据被盗
- API 密钥泄露可导致第三方服务滥用

**修复方案**:
1. **立即** 从版本控制中删除此文件中的 Secret 定义
2. 使用 Kubernetes external-secrets 或 Vault 管理密钥
3. 创建单独的 `secrets.yaml` 文件并添加到 `.gitignore`
4. 生产环境使用强随机密钥（至少 32 字节）

---

### CRITICAL-02: Docker-compose 环境变量默认弱密钥

**文件路径**: `docker-compose.yml`
**行号范围**: 53-55
**风险级别**: CRITICAL
**CVE 参考**: CWE-798

**问题描述**:
Docker-compose 配置中使用弱默认密钥，虽然标注需要修改，但默认值仍存在风险。

```yaml
- JWT_SECRET=${JWT_SECRET:-your-super-secret-jwt-key-change-in-production}
- ADMIN_PASSWORD=${ADMIN_PASSWORD:-Admin@123456}
```

**风险影响**:
- 开发/测试环境可能使用默认值导致安全漏洞
- 默认管理员密码 `Admin@123456` 是常见弱密码

**修复方案**:
1. 移除默认值，强制要求配置
2. 使用 Docker secrets 或环境文件
3. 启动时检查密钥强度，拒绝弱密钥

---

## 🟠 HIGH 级别问题

### HIGH-01: .env 文件包含敏感信息（已提交版本控制）

**文件路径**: `.env`
**行号范围**: 5-8
**风险级别**: HIGH
**CVE 参考**: CWE-538 (File and Directory Information Exposure)

**问题描述**:
项目的 `.env` 文件已被提交到版本控制系统，包含数据库连接字符串和 Redis URL（虽然密码已掩码，但格式暴露）。

```env
DATABASE_URL=postgres://postgres:***@localhost:5432/industrial_ai?sslmode=disable
REDIS_URL=redis://localhost:***@123456
```

**风险影响**:
- 配置结构泄露，攻击者可推断连接方式
- `sslmode=disable` 表示禁用 SSL，存在中间人攻击风险

**修复方案**:
1. 从 Git 历史中彻底删除 `.env` 文件
2. 确保 `.gitignore` 包含 `.env`
3. 使用 `.env.example` 仅提供模板
4. 生产环境强制启用 SSL (`sslmode=require`)

---

### HIGH-02: JWT 密钥长度验证不足

**文件路径**: `backend/internal/service/auth_jwt.go`
**行号范围**: 50-59
**风险级别**: HIGH
**CVE 参考**: CWE-326 (Inadequate Encryption Strength)

**问题描述**:
JWT 密钥最小长度要求存在，但检查位置不一致。`auth_jwt.go` 定义 `MinSecretLength`，需要确保所有初始化点都执行验证。

```go
if len(secret) < MinSecretLength {
    return nil, &JWTInitError{
        Message: fmt.Sprintf("JWT_SECRET must be at least %d characters", MinSecretLength),
    }
}
```

**修复方案**:
1. 确认 `MinSecretLength` 值 >= 32 字符
2. 生产环境建议 >= 64 字符
3. 使用 `crypto/rand` 生成密钥建议值
4. 添加密钥熵值检查

---

### HIGH-03: 缺少 CSRF 保护机制

**文件路径**: `backend/internal/handler/auth_handler_new.go`
**行号范围**: 全文件
**风险级别**: HIGH
**CVE 参考**: CWE-352 (Cross-Site Request Forgery)

**问题描述**:
认证相关 API（登录、注册、修改密码）未实施 CSRF Token 保护。虽然有 CORS 限制，但状态改变操作应额外验证 CSRF。

```go
func (h *AuthHandlerNew) ChangePassword(c *gin.Context) {
    // 无 CSRF Token 验证
    ...
}
```

**风险影响**:
- 攻击者可诱导已登录用户执行未授权操作
- 可通过恶意页面触发密码修改请求

**修复方案**:
1. 为状态改变 API 添加 CSRF Token
2. 使用 `gin-csrf` 中间件
3. 或实施 SameSite Cookie 策略 (已有，但应验证)
4. 关键操作添加二次验证（如输入当前密码）

---

### HIGH-04: 用户信息在响应中过度暴露

**文件路径**: `backend/internal/handler/auth_handler_new.go`
**行号范围**: 53-56, 92-95
**风险级别**: HIGH
**CVE 参考**: CWE-200 (Information Exposure)

**问题描述**:
登录和注册响应返回完整的用户对象，可能包含敏感字段。

```go
c.JSON(http.StatusOK, gin.H{
    "user":  user,  // 可能包含 password_hash 等敏感字段
    "token": token,
})
```

**修复方案**:
1. 创建专门的响应 DTO，仅返回必要字段
2. 过滤掉 `password`、`password_hash` 等字段
3. 使用 JSON 标签 `json:"-"` 隐藏敏感字段

---

## 🟡 MEDIUM 级别问题

### MEDIUM-01: bcrypt 使用默认成本因子

**文件路径**: `backend/internal/service/auth_password.go`
**行号范围**: 6-8
**风险级别**: MEDIUM
**CVE 参考**: CWE-916 (Use of Password Hash With Insufficient Computational Effort)

**问题描述**:
密码哈希使用 `bcrypt.DefaultCost` (成本因子 10)，对于 2026 年的安全标准建议使用更高的成本因子。

```go
bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

**风险影响**:
- 低成本因子可被暴力破解攻击
- GPU 破解速度随硬件发展不断提高

**修复方案**:
1. 将成本因子提高到 12-14
2. 或使用自适应成本因子，根据硬件性能调整
3. 定期评估并调整成本因子

---

### MEDIUM-02: Token 黑名单内存实现无大小限制

**文件路径**: `backend/internal/service/auth_blacklist.go`
**行号范围**: 84-92
**风险级别**: MEDIUM
**CVE 参考**: CWE-400 (Uncontrolled Resource Consumption)

**问题描述**:
内存实现的 Token 黑名单无最大条目限制，可能导致内存耗尽。

```go
func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
    bl := &MemoryTokenBlacklist{
        entries: make(map[string]time.Time),  // 无大小限制
        ...
    }
}
```

**风险影响**:
- 大量 Token 撤销可导致内存耗尽
- DoS 攻击风险

**修复方案**:
1. 添加最大条目数限制 (如 10000)
2. 达到限制时淘汰最旧条目
3. 或强制使用 Redis 实现

---

### MEDIUM-03: SQL 注入检测模式过于宽松

**文件路径**: `backend/internal/middleware/waf.go`
**行号范围**: 41-52
**风险级别**: MEDIUM
**CVE 参考**: CWE-89 (SQL Injection)

**问题描述**:
WAF SQL 注入检测模式为正则匹配，可能存在绕过方式。某些高级注入技术（如编码绕过）可能未被检测。

```go
SQLInjectionPatterns: []string{
    `(?i)select\s+.*\s+from`,
    // ... 基础模式
}
```

**修复方案**:
1. 项目已使用参数化查询（良好实践）
2. 增强 WAF 模式库，添加编码绕过检测
3. 定期更新检测模式
4. WAF 仅作为辅助防线，不应替代参数化查询

---

## 🔵 LOW 级别问题

### LOW-01: 日志中可能包含敏感信息

**文件路径**: 多处 `fmt.Printf` 使用
**风险级别**: LOW
**CVE 参考**: CWE-532 (Information Exposure Through Log Files)

**问题描述**:
代码中存在多处 `fmt.Printf` 输出警告信息，可能包含敏感上下文。

```go
fmt.Printf("Warning: failed to get token version: %v\n", err)
```

**修复方案**:
1. 统一使用结构化日志 (zap)
2. 避免在日志中包含用户 ID、Token 等信息
3. 使用日志脱敏处理

---

### LOW-02: Request ID 生成有备选方案不够安全

**文件路径**: `backend/internal/middleware/cors.go`
**行号范围**: 296-325
**风险级别**: LOW

**问题描述**:
当 `crypto/rand` 失败时使用备选方案，虽然已标注不安全但可能被使用。

```go
func generateFallbackRequestID() string {
    // 理论上不应该发生，但作为最后备选
    // 使用确定性但不安全的方式
}
```

**修复方案**:
1. `crypto/rand` 失败时应记录告警
2. 监控备选方案使用频率
3. 生产环境不应依赖备选方案

---

### LOW-03: 密码复杂度验证未强制特殊字符位置

**文件路径**: `backend/pkg/validation/uuid.go`
**行号范围**: 108-149
**风险级别**: LOW

**问题描述**:
密码复杂度验证要求特殊字符，但不要求分布位置。用户可能使用常见模式如 `Password123!`。

**修复方案**:
1. 检查密码不包含常见模式
2. 使用密码强度评估库
3. 添加密码黑名单检查

---

### LOW-04: CORS 开发模式允许所有 Origin

**文件路径**: `backend/internal/middleware/cors.go`
**行号范围**: 88-92
**风险级别**: LOW

**问题描述**:
开发模式下自动使用第一个允许的 Origin，可能导致开发环境安全风险。

```go
if gin.Mode() == gin.DebugMode {
    origin = allowedOrigins[0]
    allowed = true
}
```

**修复方案**:
1. 开发环境也应限制允许的 Origin
2. 使用明确的开发环境 Origin 列表
3. 添加环境检测警告

---

### LOW-05: 前端 Token 存储在 localStorage

**文件路径**: `frontend/src/lib/api.ts`
**行号范围**: 57-67
**风险级别**: LOW
**CVE 参考**: CWE-922 (Inadequate Storage of Sensitive Information)

**问题描述**:
JWT Token 存储在 localStorage，易受 XSS 攻击窃取。

```typescript
private loadToken() {
    this.token = localStorage.getItem('token');
}
```

**修复方案**:
1. 考虑使用 httpOnly Cookie 存储
2. 或使用 sessionStorage（窗口关闭自动清除）
3. 实施 Token 刷新机制，缩短有效期
4. 添加 XSS 防护（项目已有 CSP）

---

## ✅ 安全最佳实践（已实施）

### 良好实践 #1: 参数化 SQL 查询

**文件路径**: `backend/internal/repository/*.go`
**状态**: ✅ 已实施

所有数据库查询使用参数化查询，无字符串拼接 SQL：

```go
err := r.db.QueryRow(context.Background(), query, tenantID, name).Scan(...)
```

### 良好实践 #2: 表名白名单验证

**文件路径**: `backend/internal/repository/base_repo.go`
**状态**: ✅ 已实施

```go
var allowedTables = map[string]bool{
    "users": true,
    "devices": true,
    // ...
}
```

### 良好实践 #3: JWT 算法验证

**文件路径**: `backend/pkg/auth/auth.go`
**状态**: ✅ 已实施

```go
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    return nil, ErrInvalidToken  // 拒绝 none 算法
}
```

### 良好实践 #4: 安全响应头

**文件路径**: `backend/internal/middleware/security.go`
**状态**: ✅ 已实施

- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- HSTS with preload
- CSP 配置

### 良好实践 #5: Token 黑名单和版本控制

**文件路径**: `backend/internal/service/auth_blacklist.go`
**状态**: ✅ 已实施

实现了 Redis + 内存混合黑名单，支持 Token 撤销和用户级 Token 作废。

### 良好实践 #6: 密码复杂度验证

**文件路径**: `backend/pkg/validation/uuid.go`
**状态**: ✅ 已实施

12 位最小长度，要求大小写、数字、特殊字符。

### 良好实践 #7: WAF 中间件

**文件路径**: `backend/internal/middleware/waf.go`
**状态**: ✅ 已实施

SQL 注入、XSS、路径遍历、命令注入、SSRF 检测。

### 良好实践 #8: Rate Limiting

**文件路径**: `backend/internal/config/config.go`
**状态**: ✅ 已实施

支持每秒请求数、突发流量、窗口限制。

### 良好实践 #9: CORS 生产环境强制配置

**文件路径**: `backend/internal/config/config.go`
**状态**: ✅ 已实施

生产环境禁止通配符 `*`，必须显式配置 Origin。

### 良好实践 #10: 前端 XSS 防护

**状态**: ✅ 无危险 API 使用

前端未使用 `dangerouslySetInnerHTML`、`innerHTML`、`eval()` 等危险 API。

---

## 依赖安全分析

**Go 依赖分析** (`go.mod`):

| 依赖 | 版本 | 状态 | 备注 |
|------|------|------|------|
| golang.org/x/crypto | v0.49.0 | ✅ 安全 | bcrypt 实现 |
| github.com/golang-jwt/jwt/v5 | v5.2.2 | ✅ 安全 | 最新版本 |
| github.com/gin-gonic/gin | v1.9.1 | ✅ 安全 | 已知漏洞已修复 |
| github.com/lib/pq | v1.2.0 | ⚠️ 注意 | 建议升级到最新版 |
| github.com/redis/go-redis/v9 | v9.7.3 | ✅ 安全 | 最新版本 |

**建议**: 运行 `go list -m -u all` 检查可用更新，特别是 `lib/pq`。

---

## 修复优先级矩阵

| 问题编号 | 风险级别 | 修复难度 | 优先级 |
|---------|---------|---------|--------|
| CRITICAL-01 | CRITICAL | 低 | P0 |
| CRITICAL-02 | CRITICAL | 低 | P0 |
| HIGH-01 | HIGH | 中 | P1 |
| HIGH-02 | HIGH | 低 | P1 |
| HIGH-03 | HIGH | 中 | P1 |
| HIGH-04 | HIGH | 低 | P1 |
| MEDIUM-01 | MEDIUM | 低 | P2 |
| MEDIUM-02 | MEDIUM | 中 | P2 |
| MEDIUM-03 | MEDIUM | 中 | P2 |
| LOW-01~05 | LOW | 低 | P3 |

---

## 建议修复时间表

| 时间框架 | 目标 |
|---------|------|
| 立即 (24h) | CRITICAL-01, CRITICAL-02 |
| 本周内 (7d) | HIGH-01~04 |
| 本月内 (30d) | MEDIUM-01~03 |
| 下个版本 | LOW-01~05 |

---

## 审计结论

项目整体安全架构设计良好，已实施多项业界最佳实践：
- ✅ 参数化 SQL 查询防止注入
- ✅ JWT 安全实现（算法验证、黑名单）
- ✅ bcrypt 密码哈希
- ✅ 安全响应头和 CSP
- ✅ WAF 中间件
- ✅ Rate Limiting

主要风险集中在配置管理：
- ❌ 硬编码密钥在部署配置中
- ❌ 默认弱密码存在
- ❌ 环境文件泄露

**总体评分**: 75/100（良好，需要修复配置问题）

---

## 附录：安全检查清单

### 部署前检查
- [ ] JWT_SECRET 使用 >= 32 字符随机密钥
- [ ] 数据库密码强度验证
- [ ] SSL 连接强制启用
- [ ] CORS Origin 显式配置
- [ ] 环境文件不在版本控制中
- [ ] Kubernetes Secret 使用外部密钥管理

### 运行时检查
- [ ] 监控 Token 黑名单大小
- [ ] 日志脱敏验证
- [ ] Rate Limiting 统计监控
- [ ] WAF 阻断日志分析

### 定期审计
- [ ] 每季度密码哈希成本因子评估
- [ ] 每月依赖安全扫描
- [ ] 每年全面渗透测试

---

**审计完成**
**审计人员**: Hermes Security Agent
**报告版本**: v1.0