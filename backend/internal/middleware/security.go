package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// FIX-056: 安全头部 TTL 常量定义
// 这些常量用于 HTTP 安全头部配置，确保一致性
const (
	// HSTSMaxAgeSeconds HSTS 最大有效期（一年）
	// 推荐值：至少 31536000 秒（一年），可设置为两年 63072000
	HSTSMaxAgeSeconds = 31536000

	// CORSMaxAgeSeconds CORS 预检请求缓存时间（24小时）
	// 浏览器缓存预检请求结果，减少 OPTIONS 请求次数
	CORSMaxAgeSeconds = 86400
)

// SecurityHeaders 安全头部中间件
// SetSecurityHeaders 统一设置安全头部
// FIX-035: 统一 Security Headers - 提供统一的安全头部设置函数
// 其他中间件 (如 CORS) 应调用此函数，避免重复设置
// FIX-056: 使用常量定义 TTL 值
func SetSecurityHeaders(c *gin.Context) {
	// 防止 MIME 类型嗅探
	c.Header("X-Content-Type-Options", "nosniff")

	// 防止点击劫持
	c.Header("X-Frame-Options", "DENY")

	// XSS 保护 (现代浏览器已废弃但仍建议设置)
	c.Header("X-XSS-Protection", "1; mode=block")

	// HTTP Strict Transport Security
	// FIX-060: 使用更严格的 HSTS 配置
	// FIX-056: 使用常量定义 max-age
	c.Header("Strict-Transport-Security", "max-age="+strconv.Itoa(HSTSMaxAgeSeconds)+"; includeSubDomains; preload")

	// Referer 策略
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// Cache-Control (防止敏感信息缓存)
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
}

// CSPConfig Content-Security-Policy 配置
// FIX-060: CSP 严格配置
type CSPConfig struct {
	DefaultSrc     []string
	ScriptSrc      []string
	StyleSrc       []string
	ImgSrc         []string
	FontSrc        []string
	ConnectSrc     []string
	FrameAncestors []string
	FormAction     []string
	BaseURI        []string
	ReportURI      string // CSP 报告端点
}

// DefaultCSPConfig 生产环境默认 CSP 配置
// FIX-060: 严格配置，移除 unsafe-inline 和 unsafe-eval
func DefaultCSPConfig() *CSPConfig {
	return &CSPConfig{
		DefaultSrc:     []string{"'self'"},
		ScriptSrc:      []string{"'self'"}, // 生产环境: 移除 unsafe-inline 和 unsafe-eval
		StyleSrc:       []string{"'self'"}, // 生产环境: 移除 unsafe-inline
		ImgSrc:         []string{"'self'", "data:", "https:"},
		FontSrc:        []string{"'self'", "data:"},
		ConnectSrc:     []string{"'self'", "wss:", "https:"},
		FrameAncestors: []string{"'self'"},
		FormAction:     []string{"'self'"},
		BaseURI:        []string{"'self'"},
	}
}

// DevelopmentCSPConfig 开发环境 CSP 配置
// 开发环境允许更宽松的配置，方便调试
func DevelopmentCSPConfig() *CSPConfig {
	return &CSPConfig{
		DefaultSrc:     []string{"'self'"},
		ScriptSrc:      []string{"'self'", "'unsafe-inline'", "'unsafe-eval'"}, // 开发环境允许
		StyleSrc:       []string{"'self'", "'unsafe-inline'"},
		ImgSrc:         []string{"'self'", "data:", "https:", "blob:"},
		FontSrc:        []string{"'self'", "data:"},
		ConnectSrc:     []string{"'self'", "wss:", "https:", "ws:", "http://localhost:*"},
		FrameAncestors: []string{"'self'"},
		FormAction:     []string{"'self'"},
		BaseURI:        []string{"'self'"},
	}
}

// BuildCSPHeader 构建 CSP 头部字符串
// FIX-060: 根据环境自动选择配置
func BuildCSPHeader() string {
	var config *CSPConfig

	// 根据运行环境选择配置
	if gin.Mode() == gin.ReleaseMode {
		config = DefaultCSPConfig()
	} else {
		config = DevelopmentCSPConfig()
	}

	return BuildCSPFromConfig(config)
}

// BuildCSPFromConfig 从配置构建 CSP 头部
func BuildCSPFromConfig(config *CSPConfig) string {
	var parts []string

	if len(config.DefaultSrc) > 0 {
		parts = append(parts, "default-src "+joinSources(config.DefaultSrc))
	}
	if len(config.ScriptSrc) > 0 {
		parts = append(parts, "script-src "+joinSources(config.ScriptSrc))
	}
	if len(config.StyleSrc) > 0 {
		parts = append(parts, "style-src "+joinSources(config.StyleSrc))
	}
	if len(config.ImgSrc) > 0 {
		parts = append(parts, "img-src "+joinSources(config.ImgSrc))
	}
	if len(config.FontSrc) > 0 {
		parts = append(parts, "font-src "+joinSources(config.FontSrc))
	}
	if len(config.ConnectSrc) > 0 {
		parts = append(parts, "connect-src "+joinSources(config.ConnectSrc))
	}
	if len(config.FrameAncestors) > 0 {
		parts = append(parts, "frame-ancestors "+joinSources(config.FrameAncestors))
	}
	if len(config.FormAction) > 0 {
		parts = append(parts, "form-action "+joinSources(config.FormAction))
	}
	if len(config.BaseURI) > 0 {
		parts = append(parts, "base-uri "+joinSources(config.BaseURI))
	}
	if config.ReportURI != "" {
		parts = append(parts, "report-uri "+config.ReportURI)
	}

	return strings.Join(parts, "; ")
}

// joinSources 连接 CSP 源列表
func joinSources(sources []string) string {
	return strings.Join(sources, " ")
}

// ForceHTTPS 强制 HTTPS 中间件
// 在代理模式下检查 X-Forwarded-Proto，重定向到 HTTPS
func ForceHTTPS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否通过代理
		proto := c.GetHeader("X-Forwarded-Proto")

		// 如果 proto 是 HTTP，重定向到 HTTPS
		if proto == "http" {
			host := c.GetHeader("X-Forwarded-Host")
			if host == "" {
				host = c.Request.Host
			}

			// 构造 HTTPS URL
			httpsURL := "https://" + host + c.Request.RequestURI

			// 301 永久重定向
			c.Redirect(301, httpsURL)
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSSecurity CORS 安全配置中间件
// FIX-P1-13: 已合并到 cors.go，此处保留函数签名以向后兼容
// Deprecated: Use middleware.CORS or middleware.CORSWithConfig instead
func CORSSecurity(allowedOrigins []string) gin.HandlerFunc {
	// 调用 cors.go 中统一的 CORS 实现
	return CORS(allowedOrigins)
}

// HSTS HTTP Strict Transport Security 中间件
// 强制浏览器使用 HTTPS 连接
// 参数:
//   - maxAge: HSTS 有效期（秒），建议至少 31536000（一年）
//   - includeSubDomains: 是否包含子域名
//   - preload: 是否提交到 HSTS preload 列表
func HSTS(maxAge int, includeSubDomains bool, preload bool) gin.HandlerFunc {
	value := "max-age=" + strconv.Itoa(maxAge)
	if includeSubDomains {
		value += "; includeSubDomains"
	}
	if preload {
		value += "; preload"
	}

	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", value)
		c.Next()
	}
}

// HSTSDefault 生产环境默认 HSTS 配置
// FIX-P2: 默认启用 preload，确保最大安全保护
// 包含: max-age=1年, includeSubDomains, preload
func HSTSDefault() gin.HandlerFunc {
	return HSTS(HSTSMaxAgeSeconds, true, true)
}
