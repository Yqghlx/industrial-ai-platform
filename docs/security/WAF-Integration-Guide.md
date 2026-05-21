# WAF 集成指南

> **Industrial AI Platform Web 应用防火墙最佳实践**  
> **版本**: 1.0.0  
> **更新日期**: 2026-05-13

---

## 📋 WAF 概述

Phase 4 P2 WAF 集成目标：

| 指标 | 当前状态 | 目标 |
|------|---------|------|
| **SQL 注入防护** | 应用层 | WAF + 应用层 |
| **XSS 防护** | 应用层 | WAF + 应用层 |
| **DDoS 防护** | 无 | WAF 限流 |
| **异常请求检测** | 无 | WAF 规则 |

---

## 🔄 WAF 架构

### WAF 防护流程

```
┌─────────────────────────────────────────┐
│  用户请求                                │
│  - HTTP 请求                             │
│  - API 调用                              │
│  - WebSocket 连接                        │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  WAF (Web 应用防火墙)                    │
│  - 请求检查                              │
│  - 规则匹配                              │
│  - 异常检测                              │
│  - 流量限制                              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  Nginx 反向代理                          │
│  - TLS 终止                              │
│  - 请求转发                              │
│  - 日志记录                              │
└─────────────────────────────────────────┘
          ↓
┌─────────────────────────────────────────┐
│  应用服务                                │
│  - 业务逻辑                              │
│  - 数据处理                              │
│  - API 响应                              │
└─────────────────────────────────────────┘
```

---

## 🔧 WAF 规则类型

### OWASP Top 10 防护规则

| 规则类型 | 说明 | 防护内容 |
|---------|------|---------|
| **SQL Injection** | SQL 注入防护 | SELECT/INSERT/UPDATE/DELETE |
| **XSS Attack** | XSS 攻击防护 | <script>/javascript/onclick |
| **Path Traversal** | 路径遍历防护 | ..//..\ etc/passwd |
| **Command Injection** | 命令注入防护 | ; | ` $() exec |
| **SSRF Attack** | SSRF 攻击防护 | http://localhost/file:// |
| **File Upload** | 文件上传防护 | .php/.jsp/.exe |
| **Sensitive Data** | 敏感数据防护 | password/token/secret |
| **Rate Limiting** | 流量限制 | 请求频率/并发数 |

---

## 📝 ModSecurity 规则配置

### ModSecurity 核心规则集 (CRS)

```apache
# ModSecurity 配置

SecRuleEngine On
SecRequestBodyAccess On
SecResponseBodyAccess Off

# SQL 注入防护
SecRule REQUEST_ARGS "@rx (?i:select|insert|update|delete|drop|union|exec)" \
    "id:1001,phase:2,deny,status:403,msg:'SQL Injection Detected'"

# XSS 防护
SecRule REQUEST_ARGS "@rx (?i:<script|javascript:|onerror=|onclick=)" \
    "id:1002,phase:2,deny,status:403,msg:'XSS Attack Detected'"

# 路径遍历防护
SecRule REQUEST_ARGS "@rx (\.\.\/|\.\.\\|%2e%2e%2f)" \
    "id:1003,phase:2,deny,status:403,msg:'Path Traversal Detected'"

# 命令注入防护
SecRule REQUEST_ARGS "@rx (;|\|`|\$\(|exec\(|system\()" \
    "id:1004,phase:2,deny,status:403,msg:'Command Injection Detected'"

# SSRF 防护
SecRule REQUEST_ARGS "@rx (http://localhost|file://|ftp://)" \
    "id:1005,phase:2,deny,status:403,msg:'SSRF Attack Detected'"

# 文件上传防护
SecRule FILES_EXTENSION "@rx \.(php|jsp|asp|exe|sh)$" \
    "id:1006,phase:2,deny,status:403,msg:'Dangerous File Upload'"

# 敏感数据防护
SecRule RESPONSE_BODY "@rx (?i:password|token|secret|key)" \
    "id:1007,phase:4,deny,status:403,msg:'Sensitive Data Leakage'"
```

---

## 📊 Nginx WAF 配置

### Nginx 安全规则

```nginx
# Nginx WAF 配置

# 禁止危险请求方法
if ($request_method !~ ^(GET|POST|PUT|DELETE|HEAD|OPTIONS)$) {
    return 405;
}

# 禁止 SQL 注入关键字
if ($args ~* "(select|insert|update|delete|drop|union|exec)") {
    return 403;
}

# 禁止 XSS 关键字
if ($args ~* "(<script|javascript:|onerror|onclick|onload)") {
    return 403;
}

# 禁止路径遍历
if ($args ~* "(../|..\\|%2e%2e%2f|%252e%252e%252f)") {
    return 403;
}

# 禁止命令注入
if ($args ~* "(;|\\|`|\\$\\(|exec|system)") {
    return 403;
}

# 禁止 SSRF
if ($args ~* "(http://localhost|file://|ftp://|dict://)") {
    return 403;
}

# 禁止空 User-Agent
if ($http_user_agent = "") {
    return 403;
}

# 禁止爬虫 User-Agent
if ($http_user_agent ~* "(bot|crawler|spider|scan)") {
    return 403;
}

# 限制请求大小
client_max_body_size 10m;

# 限制请求速率
limit_req_zone $binary_remote_addr zone=req_limit:10m rate=10r/s;
limit_req zone=req_limit burst=20 nodelay;

# 限制连接数
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;
limit_conn conn_limit 10;
```

---

## 🔧 Go WAF 中间件实现

### WAF 中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "regexp"
)

// WAFMiddleware WAF 中间件
func WAFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查请求参数
        args := c.Request.URL.Query()
        for key, values := range args {
            for _, value := range values {
                // SQL 注入检查
                if detectSQLInjection(value) {
                    c.AbortWithStatus(403)
                    return
                }
                
                // XSS 检查
                if detectXSS(value) {
                    c.AbortWithStatus(403)
                    return
                }
                
                // 路径遍历检查
                if detectPathTraversal(value) {
                    c.AbortWithStatus(403)
                    return
                }
            }
        }
        
        c.Next()
    }
}

// detectSQLInjection 检测 SQL 注入
func detectSQLInjection(input string) bool {
    patterns := []string{
        `(?i)select.*from`,
        `(?i)insert.*into`,
        `(?i)update.*set`,
        `(?i)delete.*from`,
        `(?i)drop.*table`,
        `(?i)union.*select`,
    }
    
    for _, pattern := range patterns {
        if regexp.MustCompile(pattern).MatchString(input) {
            return true
        }
    }
    return false
}

// detectXSS 检测 XSS
func detectXSS(input string) bool {
    patterns := []string{
        `<script`,
        `javascript:`,
        `onerror=`,
        `onclick=`,
        `onload=`,
    }
    
    for _, pattern := range patterns {
        if regexp.MustCompile(pattern).MatchString(input) {
            return true
        }
    }
    return false
}
```

---

## ✅ WAF 验收标准

| 检查项 | 要求 | 验证方法 |
|--------|------|---------|
| **SQL 注入防护** | 规则生效 | 安全测试 |
| **XSS 防护** | 规则生效 | 安全测试 |
| **路径遍历防护** | 规则生效 | 安全测试 |
| **命令注入防护** | 规则生效 | 安全测试 |
| **流量限制** | 规则生效 | 压测验证 |
| **日志记录** | 全量记录 | 日志检查 |

---

**最后更新**: 2026-05-13  
**审核人**: Security Team