# 本地开发环境安全指南

## SEC-LOW-04: 环境变量处理说明

本文档说明本地开发环境中环境变量的安全处理方式。

---

## ⚠️ 重要安全警告

**环境变量在本地开发环境中以明文形式存储，这仅适用于开发测试场景。**

在生产环境中：
- 使用 Kubernetes Secrets 或 HashiCorp Vault 等密钥管理服务
- 使用加密存储敏感配置
- 通过安全配置注入机制传递环境变量

---

## 🔐 本地开发环境变量安全实践

### 1. 环境变量文件处理

本地开发使用 `.env` 文件存储配置：

```bash
# 正确做法
.env.example  # 示例文件，提交到 Git（不含真实密钥）
.env          # 本地配置文件，不提交到 Git（在 .gitignore 中排除）
```

### 2. 安全配置检查清单

| 检查项 | 本地开发 | 生产环境 |
|--------|----------|----------|
| JWT_SECRET | 最少32字符 | 从 Vault 获取 |
| DATABASE_URL | 本地测试数据库 | 加密存储 |
| REDIS_URL | 本地 Redis | 安全网络 |
| CORS_ORIGINS | localhost | 严格域名列表 |

### 3. 开发环境特殊处理

#### 3.1 CORS 开发配置
- 开发模式允许 localhost 和通配符来源
- 生产模式必须使用严格的域名列表

#### 3.2 CSP 内容安全策略
- 开发环境：允许 `unsafe-inline` 和 `unsafe-eval`
- 生产环境：严格 CSP，禁止 unsafe-* 指令

#### 3.3 JWT 密钥长度
- 开发环境：建议32字符以上
- 生产环境：必须32字符以上，使用强随机生成

### 4. 禁止操作

**本地开发环境中绝对禁止：**
1. 提交 `.env` 文件到 Git
2. 在日志中打印密钥内容或长度
3. 在测试代码中使用真实生产密钥
4. 在容器镜像中硬编码密钥

### 5. 推荐开发实践

```bash
# 生成安全的开发密钥
openssl rand -base64 32

# 或使用 go 标准库
go run -c 'package main; import ("crypto/rand", "encoding/base64", "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

### 6. 安全检查命令

运行安全检查:

```bash
# 检查是否有密钥泄露
grep -r "secret\|password\|key" --include="*.go" | grep -v "// \|/\*\|vendor"

# 检查 .env 是否被正确忽略
git check-ignore .env

# 运行漏洞扫描
make vulncheck
```

---

## 📋 配置示例

### .env.example 示例

```bash
# 应用配置
GIN_MODE=debug
PORT=8080

# 数据库配置 (本地开发)
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable

# Redis 配置 (本地开发)
REDIS_URL=localhost:6379

# JWT 配置 - 使用 openssl rand -base64 32 生成
JWT_SECRET=your-development-secret-key-min-32-characters

# CORS 配置 (开发环境)
CORS_ORIGINS=http://localhost:3000,http://localhost:8080

# 日志配置
LOG_LEVEL=debug
LOG_FORMAT=console
```

---

## 🔧 故障排查

### 问题：密钥长度不足
**现象**: JWT 签名失败
**解决**: 确保 JWT_SECRET 至少32字符

### 问题：CORS 跨域错误
**现象**: 前端请求被拒绝
**解决**: 
- 开发模式确保 CORS_ORIGINS 包含前端地址
- 生产模式使用严格域名列表

### 问题：数据库连接失败
**现象**: 连接超时
**解决**: 检查 DATABASE_URL 格式和网络可达性

---

## 📚 参考资料

- [OWASP 环境变量安全](https://owasp.org/www-community/vulnerabilities/Improper_Data_Handling)
- [12 Factor App Config](https://12factor.net/config)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [HashiCorp Vault](https://www.vaultproject.io/)

---

**最后更新**: 2026-05-22
**安全级别**: SEC-LOW-04