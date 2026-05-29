# 工业AI平台安全审计报告

**审计日期**: 2026-05-28  
**审计范围**: API Handlers、数据库操作、认证系统、配置文件  
**项目路径**: /Users/yqgmac/yqg/project/industrial-ai-platform

---

## 漏洞汇总

| 严重级别 | 数量 | 类别 |
|---------|------|------|
| CRITICAL | 3 | 硬编码密钥泄露、敏感文件提交 |
| HIGH | 4 | SSL禁用、公端点暴露、输入验证缺失 |
| MEDIUM | 3 | CORS配置、JWT配置 |
| LOW | 2 | 信息泄露、占位实现 |

---

## CRITICAL 级别漏洞

### CVE-IAP-001: 硬编码密钥和密码泄露

**描述**: 多个敏感文件包含明文密钥和密码，这些文件存在于项目目录中。

**影响文件**:
- `.env` - 包含明文密码 `POSTGRES_PASSWORD=<REDACTED>`
- `.secrets.tmp` - 包含 `JWT_SECRET`, `ENCRYPTION_KEY`, `REDIS_PASSWORD`, `DB_PASSWORD`, `ADMIN_PASSWORD` 等敏感信息

**CVSS评分**: 9.8 (Critical)  
**攻击向量**: Network (AV:N)  
**攻击复杂度**: Low (AC:L)  
**权限要求**: None (PR:N)  
**影响**: High (IH:H)

**证据**:
```
JWT_SECRET=<REDACTED>
ENCRYPTION_KEY=<REDACTED>
REDIS_PASSWORD=<REDACTED>
DB_PASSWORD=<REDACTED>
ADMIN_PASSWORD=<REDACTED>
```

**修复建议**:
1. 立即从文件系统中删除 `.secrets.tmp` 文件
2. 使用密码管理工具或环境变量注入密钥
3. 确保 `.env` 文件不被提交到 Git（已在 `.gitignore` 中）
4. 轮换所有已暴露的密钥和密码
5. 使用 Docker Secrets 或 Kubernetes Secrets 管理敏感配置

---

### CVE-IAP-002: 数据库连接禁用SSL

**描述**: 多个配置文件中的数据库连接字符串使用 `sslmode=disable`，禁用了 SSL/TLS 加密。

**影响文件**:
- `docker-compose.yml` (line 47)
- `docker-compose.ghcr.yml` (line 36)
- `infra/k8s/deployment.yaml` (line 17)
- `README.md` (示例代码)

**CVSS评分**: 8.6 (High)  
**攻击向量**: Network (AV:N)  
**攻击复杂度**: Medium (AC:M)  
**权限要求**: Low (PR:L)

**证据**:
```yaml
DATABASE_URL=postgres://${POSTGRES_USER}:***@postgres:5432/${POSTGRES_DB:-industrial_ai}?sslmode=disable
```

**修复建议**:
1. 生产环境必须使用 `sslmode=require` 或 `sslmode=verify-full`
2. 配置 PostgreSQL 服务器证书
3. 更新所有示例文档中的连接字符串
4. 在 `scripts/verify-production-config.sh` 中添加 SSL 检查（已存在）

---

### CVE-IAP-003: 敏感信息写入临时文件

**描述**: `.secrets.tmp` 文件包含所有生产环境的敏感密钥，且文件被创建在项目根目录。

**CVSS评分**: 9.1 (Critical)  
**攻击向量**: Local (AV:L)  
**攻击复杂度**: Low (AC:L)  
**权限要求**: Low (PR:L)

**证据**: 文件 `scripts/generate-secrets.sh` 生成包含敏感信息的临时文件

**修复建议**:
1. 使用 `/dev/stdout` 或管道输出，避免写入文件
2. 或写入到临时目录并设置严格权限
3. 生成后立即删除文件
4. 禁止将此文件提交到任何版本控制系统

---

## HIGH 级别漏洞

### CVE-IAP-004: 遥测端点公开放置

**描述**: `/api/v1/devices/telemetry` 端点被配置为公端点，无需认证即可访问。

**影响文件**:
- `backend/internal/middleware/auth.go` (line 28)

**CVSS评分**: 7.5 (High)  
**攻击向量**: Network (AV:N)  

**证据**:
```go
PublicEndpoints: []string{
    "/health",
    "/api/v1/devices/telemetry", // SEC-MED-02: Intentionally public for edge device data ingestion
},
```

**风险分析**:
- 允许未认证的数据提交
- 可能被用于数据注入攻击
- 可能导致虚假数据污染系统

**修复建议**:
1. 为边缘设备实现设备认证机制（API Key 或证书）
2. 添加 IP 白名单限制
3. 实现请求速率限制
4. 添加数据签名验证机制

---

### CVE-IAP-005: 管理员接口占位实现

**描述**: `admin_handler_new.go` 中的用户创建和删除功能为占位实现，无实际安全检查。

**影响文件**:
- `backend/internal/handler/admin_handler_new.go` (lines 58-84)

**CVSS评分**: 7.2 (High)

**证据**:
```go
func (h *AdminHandlerNew) CreateUser(c *gin.Context) {
    // 占位实现 - 实际应调用 authSvc.Register
    c.JSON(http.StatusOK, gin.H{
        "message": "User created (placeholder)",
        "user":    req,  // 返回完整用户信息包括密码
    })
}
```

**风险**:
- 返回包含密码的用户数据
- 无实际验证或授权检查
- 可能导致权限提升

**修复建议**:
1. 完成实际的用户管理逻辑
2. 确保密码不返回给客户端
3. 添加 RBAC 权限检查
4. 实现审计日志记录

---

### CVE-IAP-006: 管理员密码未强制验证

**描述**: 配置验证中，`ADMIN_PASSWORD` 只在未设置时生成警告，未强制要求。

**影响文件**:
- `backend/internal/config/config.go` (line 413)

**CVSS评分**: 6.8 (Medium-High)

**修复建议**:
1. 生产环境强制要求设置 `ADMIN_PASSWORD`
2. 验证密码复杂度
3. 首次启动时强制修改默认密码

---

### CVE-IAP-007: JWT密钥长度未强制验证

**描述**: JWT密钥只在生产环境强制要求设置，未验证密钥长度和强度。

**影响文件**:
- `backend/internal/middleware/jwt_helpers.go`

**CVSS评分**: 6.5 (Medium)

**修复建议**:
1. 验证 JWT_SECRET 长度 >= 32 字符
2. 添加密钥强度检查
3. 禁止使用常见弱密钥

---

## MEDIUM 级别漏洞

### CVE-IAP-008: CORS开发环境通配符

**描述**: CORS中间件在开发环境允许使用通配符 `*`。

**影响文件**:
- `backend/internal/middleware/cors.go` (lines 87-107)

**CVSS评分**: 5.3 (Medium)

**证据**:
```go
if len(allowedOrigins) == 0 {
    if gin.Mode() == gin.DebugMode {
        // 开发模式：允许所有 origin（但记录警告）
        c.Header("Access-Control-Allow-Origin", origin)
```

**修复建议**:
1. 即使开发环境也建议限制 CORS origins
2. 生产环境已禁止通配符（已修复）

---

### CVE-IAP-009: 前端innerHTML使用

**描述**: 前端代码使用 `innerHTML = ''` 清空容器。

**影响文件**:
- `frontend/src/components/KnowledgeGraph.tsx` (line 43)

**CVSS评分**: 4.3 (Medium)

**证据**:
```typescript
containerRef.current.innerHTML = '';
```

**风险分析**:
- 当前用法安全（仅清空）
- 潜在XSS风险如果后续修改为插入内容

**修复建议**:
1. 使用 `textContent` 或 React 状态管理替代
2. 添加安全审查注释

---

### CVE-IAP-010: 测试代码中硬编码密码模式

**描述**: 测试文件中使用示例密码字符串。

**影响文件**:
- `backend/docs/openapi.yaml` (lines 192-198)

**CVSS评分**: 3.5 (Low-Medium)

**证据**:
```yaml
password: "SecurePass123"
password: "AdminPass456"
```

**修复建议**:
1. 使用占位符 `"<REDACTED>"` 替代示例密码
2. 添加文档说明这些仅为示例

---

## LOW 级别漏洞

### CVE-IAP-011: 错误信息可能泄露配置细节

**描述**: 配置错误时输出详细信息到 stderr。

**影响文件**:
- `backend/internal/config/config.go` (lines 301-318)

**CVSS评分**: 2.4 (Low)

**修复建议**:
1. 生产环境只输出通用错误信息
2. 详细信息仅输出到日志系统

---

### CVE-IAP-012: 数据库URL示例包含密码格式

**描述**: 文档示例中展示完整的数据库连接字符串格式。

**影响范围**: README.md, 多个文档文件

**CVSS评分**: 2.1 (Low)

**修复建议**:
1. 使用 `<PASSWORD>` 占位符
2. 添加安全警告注释

---

## 已实施的安全措施（正面发现）

### 安全设计亮点

1. **SQL注入防护** ✅
   - 使用参数化查询 (`$1, $2...`)
   - 表名白名单验证 (`base_repo.go`)
   - 列名字符验证

2. **JWT算法验证** ✅
   - 禁止 "none" 算法攻击
   - 只允许 HMAC 算法
   - 验证必要 claims

3. **WAF中间件** ✅
   - SQL注入检测模式
   - XSS检测模式
   - 路径遍历检测
   - 命令注入检测

4. **安全头部** ✅
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY
   - HSTS with includeSubDomains
   - CSP 配置（开发/生产分离）

5. **密码复杂度验证** ✅
   - 最小12字符
   - 大小写、数字、特殊字符要求

6. **输入验证** ✅
   - UUID格式验证
   - ID长度和字符验证
   - 分页参数限制

7. **CORS生产验证** ✅
   - 生产环境禁止通配符
   - 必须显式配置 origins

---

## 修复优先级建议

| 优先级 | CVE编号 | 建议修复时间 |
|-------|---------|-------------|
| P0 | CVE-IAP-001 | 立即 |
| P0 | CVE-IAP-003 | 立即 |
| P1 | CVE-IAP-002 | 24小时内 |
| P1 | CVE-IAP-004 | 48小时内 |
| P2 | CVE-IAP-005 | 1周内 |
| P2 | CVE-IAP-006 | 1周内 |
| P3 | CVE-IAP-007-012 | 2周内 |

---

## 审计结论

该项目整体安全架构设计良好，已实施多项安全措施。主要风险集中在：

1. **敏感信息管理**: 存在密钥泄露风险，需要立即处理
2. **传输层安全**: 数据库SSL需要强制启用
3. **公端点访问**: 遥测端点需要额外的设备认证机制

建议按照优先级顺序修复上述漏洞，并定期进行安全审计。

---

**审计人**: Hermes Security Scanner  
**版本**: v1.0  
**生成时间**: 2026-05-28 14:19 CST