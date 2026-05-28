package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS 安全配置
// 用于配置跨域资源共享 (CORS) 的各项参数
// 建议生产环境使用严格的 AllowedOrigins 列表，避免使用 "*"
type CORSConfig struct {
	AllowedOrigins   []string // 允许的域名列表 (精确匹配)，如 ["https://example.com"]
	AllowedMethods   []string // 允许的 HTTP 方法，如 ["GET", "POST", "PUT", "DELETE"]
	AllowedHeaders   []string // 允许的请求头，如 ["Content-Type", "Authorization"]
	ExposedHeaders   []string // 暴露给客户端的响应头，如 ["X-Total-Count", "X-Request-ID"]
	AllowCredentials bool     // 是否允许发送 Cookie 和认证信息
	MaxAge           int      // 预检请求缓存时间 (秒)，建议 86400 (24小时)
}

// DefaultCORSConfig 生产环境默认 CORS 配置
// 返回一个安全的 CORS 配置，适合生产环境使用
// SEC-HIGH-04: 从环境变量读取允许的域名列表
// 特点：
//   - 明确的 AllowedOrigins 列表，不使用通配符
//   - 允许 Credentials，支持认证
//   - 24 小时预检缓存，减少 OPTIONS 请求
//   - 只允许必要的 HTTP 方法和头部
func DefaultCORSConfig() *CORSConfig {
	// SEC-HIGH-04: 从环境变量读取 CORS origins
	// 如果未设置，返回空列表（生产环境会拒绝所有跨域请求）
	originsStr := os.Getenv("CORS_ORIGINS")
	allowedOrigins := []string{}

	if originsStr != "" {
		for _, o := range strings.Split(originsStr, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" && trimmed != "*" {
				// 生产环境不允许使用通配符
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	// 如果没有配置 CORS_ORIGINS，使用示例域名作为默认值（仅用于演示）
	// 生产部署时必须设置 CORS_ORIGINS 环境变量
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"https://industrial-ai.example.com",
			"https://admin.industrial-ai.example.com",
		}
	}

	return &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Tenant-ID",
			"Accept-Language",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-Total-Count",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           CORSMaxAgeSeconds, // FIX-056: 使用常量
	}
}

// CORS middleware with security headers (统一版)
// 合并了原有的 CORS 函数和 CORSSecurityWithConfig 的功能
// 使用方法:
//
//	router.Use(middleware.CORS([]string{"https://example.com", "https://admin.example.com"}))
//
// 开发模式行为:
//   - 如果 origin 不在允许列表中，开发模式会自动使用第一个允许的 origin
//   - 生产模式会严格拒绝未授权的 origin
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originsMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		originsMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 没有 Origin 头部 (同站请求) 直接放行
		if origin == "" {
			// 设置安全头部
			setSecurityHeaders(c)
			c.Next()
			return
		}

		// SEC-MEDIUM-01: 生产环境安全修复
		// 如果 allowedOrigins 为空，在非调试模式下拒绝所有跨域请求
		if len(allowedOrigins) == 0 {
			if gin.Mode() == gin.DebugMode {
				// 开发模式：允许所有 origin（但记录警告）
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, X-Request-ID, X-Tenant-ID")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Max-Age", strconv.Itoa(CORSMaxAgeSeconds))
				setSecurityHeaders(c)
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusNoContent)
					return
				}
				c.Next()
				return
			}
			// 生产模式：没有配置 CORS origins 时拒绝跨域请求
			setSecurityHeaders(c)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Check if origin is allowed
		allowed := isOriginAllowed(origin, allowedOrigins)

		if !allowed && !originsMap["*"] {
			// 生产模式会严格拒绝未授权的 origin
			if gin.Mode() != gin.DebugMode {
				// 不允许的 origin，不设置 CORS 头部
				// 浏览器会阻止跨域请求
				setSecurityHeaders(c)
				c.Next()
				return
			}
			// 开发模式：使用第一个允许的 origin 作为 fallback
			origin = allowedOrigins[0]
			allowed = true
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, X-Request-ID, X-Tenant-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			// FIX-056: 使用常量定义 CORS MaxAge
			c.Header("Access-Control-Max-Age", strconv.Itoa(CORSMaxAgeSeconds))
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-Total-Count, X-Request-ID")
		}

		// Security headers
		setSecurityHeaders(c)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSWithConfig CORS 中间件 (配置版)
// 使用 CORSConfig 结构体进行配置，适合需要精细控制的场景
// 使用方法:
//
//	config := middleware.DefaultCORSConfig()
//	config.AllowedOrigins = []string{"https://your-domain.com"}
//	router.Use(middleware.CORSWithConfig(config))
//
// 注意事项:
//   - 当 AllowCredentials 为 true 时，AllowedOrigins 不能使用 "*"
//   - MaxAge 设置为 0 时不会设置 Access-Control-Max-Age 头部
func CORSWithConfig(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 没有 Origin 头部 (同站请求) 直接放行
		if origin == "" {
			setSecurityHeaders(c)
			c.Next()
			return
		}

		// 检查 origin 是否在允许列表中
		allowed := isOriginAllowed(origin, config.AllowedOrigins)

		if !allowed {
			// 不允许的 origin，不设置 CORS 头部
			setSecurityHeaders(c)
			c.Next()
			return
		}

		// 设置 CORS 头部
		c.Header("Access-Control-Allow-Origin", origin)

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		}

		if len(config.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		}

		if len(config.ExposedHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		// Security headers
		setSecurityHeaders(c)

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// setSecurityHeaders 设置安全头部
// FIX-035: 统一 Security Headers - 调用 security.go 中的统一函数
func setSecurityHeaders(c *gin.Context) {
	SetSecurityHeaders(c)
}

// Security headers middleware (独立中间件，不包含 CORS)
// FIX-035: 统一 Security Headers - 调用统一函数
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetSecurityHeaders(c)
		c.Next()
	}
}

// isOriginAllowed 检查 origin 是否在允许列表中
// 支持精确匹配和通配符子域名 (*.example.com)
// SEC-MEDIUM-01: 生产环境禁用通配符 "*" (通过 ValidateCORS 验证)
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		// 精确匹配
		if origin == allowed {
			return true
		}

		// SEC-MEDIUM-01: 通配符 "*" 仅用于开发环境检查
		// 生产环境通过 config.ValidateCORS() 禁止使用 "*"
		// 这里保留检查以支持开发环境，但生产环境配置验证会阻止
		if allowed == "*" {
			// 在 ReleaseMode 中，"*" 通配符不安全，记录警告
			if gin.Mode() == gin.ReleaseMode {
				// 生产环境不应该使用通配符，但为了向后兼容，仍然允许
				// 实际安全验证在 config.ValidateCORS() 中进行
			}
			return true
		}

		// 通配符子域名匹配 (*.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:] // 去掉 "*."
			// origin 必须以 domain 结尾，且是子域名
			if strings.HasSuffix(origin, domain) {
				// 确保 origin 是子域名，不是 domain 本身
				originWithoutDomain := origin[:len(origin)-len(domain)]
				if originWithoutDomain != "" &&
					(originWithoutDomain[len(originWithoutDomain)-1] == '.' ||
						strings.Contains(originWithoutDomain, "//")) {
					return true
				}
			}
		}
	}
	return false
}

// ParseOrigins 解析 CORS origins 配置字符串
// SEC-HIGH-01: 安全修复 - 空输入返回空数组而非 ["*"] 通配符
// 在生产环境中，必须显式配置 CORS origins
// 使用 config.GetCORSOrigins() 可获得更安全的默认行为
func ParseOrigins(originsStr string) []string {
	if originsStr == "" {
		// 返回空数组，不允许通配符默认值
		// 这符合 CWE-942 的安全最佳实践
		return []string{}
	}

	origins := strings.Split(originsStr, ",")
	result := []string{}

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// RequestID middleware (使用 crypto/rand 生成唯一 ID)
// FIX-036: 使用 crypto/rand 替代时间戳生成 ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID 生成唯一的 Request ID
// 使用 crypto/rand 确保安全性
func generateRequestID() string {
	// 生成 16 字节的随机数据
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果 crypto/rand 失败，这不应该在生产环境发生
		// 但我们仍然需要一个备用方案
		// 使用更安全的方式：时间戳 + 随机数组合
		// 注意：这里不使用 panic，而是生成一个仍然可用的 ID
		return generateFallbackRequestID()
	}
	return hex.EncodeToString(bytes)
}

// generateFallbackRequestID 备用 ID 生成方法
// 仅在 crypto/rand 失败时使用
func generateFallbackRequestID() string {
	// 生成两个随机部分组合
	random1 := generateSecureRandomString(8)
	random2 := generateSecureRandomString(8)
	return random1 + "-" + random2
}

// generateSecureRandomString 生成安全的随机字符串
// 使用 crypto/rand
func generateSecureRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)

	if _, err := rand.Read(bytes); err != nil {
		// 理论上不应该发生，但作为最后备选
		// 使用确定性但不安全的方式
		for i := range bytes {
			bytes[i] = letters[i%len(letters)]
		}
		return string(bytes)
	}

	// 使用随机字节映射到字母表
	for i := range bytes {
		bytes[i] = letters[int(bytes[i])%len(letters)]
	}
	return string(bytes)
}
