package middleware

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ============================================
// WAF 中间件 - Web 应用防火墙
// ============================================

// compiledPatterns 预编译的正则表达式
// P1-11: 预编译正则表达式以提升性能
var compiledPatterns struct {
	sqlInjection     []*regexp.Regexp
	xss              []*regexp.Regexp
	pathTraversal    []*regexp.Regexp
	commandInjection []*regexp.Regexp
	ssrf             []*regexp.Regexp
	initialized      bool
}

// WAFConfig WAF 配置
type WAFConfig struct {
	Enabled                  bool     // 是否启用 WAF
	SQLInjectionPatterns     []string // SQL 注入检测模式
	XSSPatterns              []string // XSS 检测模式
	PathTraversalPatterns    []string // 路径遍历检测模式
	CommandInjectionPatterns []string // 命令注入检测模式
	SSRFPatterns             []string // SSRF 检测模式
	MaxRequestSize           int64    // 最大请求大小
	MaxArgsLength            int      // 最大参数长度
	BlockedExtensions        []string // 禁止的文件扩展名
	LogBlockedRequests       bool     // 是否记录被阻止的请求
	// FIX-040: User-Agent 黑名单配置化
	BlockedUserAgents   []string // 禁止的 User-Agent 模式列表
	BlockEmptyUserAgent bool     // 是否阻止空 User-Agent
}

// DefaultWAFConfig 默认 WAF 配置
func DefaultWAFConfig() WAFConfig {
	return WAFConfig{
		Enabled: true,
		SQLInjectionPatterns: []string{
			`(?i)select\s+.*\s+from`,
			`(?i)insert\s+.*\s+into`,
			`(?i)update\s+.*\s+set`,
			`(?i)delete\s+.*\s+from`,
			`(?i)drop\s+.*\s+table`,
			`(?i)union\s+.*\s+select`,
			`(?i)exec\s+.*`,
			`(?i)execute\s+.*`,
			`(?i)declare\s+.*`,
			`(?i)truncate\s+.*`,
		},
		XSSPatterns: []string{
			`<script`,
			`javascript:`,
			`onerror\s*=`,
			`onclick\s*=`,
			`onload\s*=`,
			`onmouseover\s*=`,
			`onfocus\s*=`,
			`onblur\s*=`,
			`alert\s*\(`,
			`eval\s*\(`,
			`document\.cookie`,
			`document\.write`,
		},
		PathTraversalPatterns: []string{
			`\.\.\/`,
			`\.\.\\`,
			`%2e%2e%2f`,
			`%252e%252e%252f`,
			`\/etc\/passwd`,
			`\/etc\/shadow`,
			`c:\\windows`,
			`c:\/windows`,
		},
		CommandInjectionPatterns: []string{
			`\s*;\s*`,
			`\s*\|\s*`,
			`\s*` + "`" + `\s*`,
			`\$\(`,
			`exec\s*\(`,
			`system\s*\(`,
			`passthru\s*\(`,
			`shell_exec\s*\(`,
			`popen\s*\(`,
		},
		SSRFPatterns: []string{
			`http:\/\/localhost`,
			`http:\/\/127\.0\.0\.1`,
			`http:\/\/0\.0\.0\.0`,
			`file:\/\/`,
			`ftp:\/\/`,
			`dict:\/\/`,
			`gopher:\/\/`,
			`ldap:\/\/`,
		},
		MaxRequestSize:     10 * 1024 * 1024, // 10MB
		MaxArgsLength:      1000,
		BlockedExtensions:  []string{".php", ".jsp", ".asp", ".aspx", ".exe", ".sh", ".bat", ".cmd"},
		LogBlockedRequests: true,
		// FIX-040: 默认 User-Agent 黑名单
		BlockedUserAgents: []string{
			"bot", "crawler", "spider", "scan", "nikto", "sqlmap", "masscan", "nmap",
			"nessus", "openvas", "acunetix", "dirbuster", "gobuster", "wpscan",
		},
		BlockEmptyUserAgent: true, // 默认阻止空 User-Agent
	}
}

// LoadWAFConfigFromEnv 从环境变量加载 WAF 配置
// FIX-040: 支持环境变量配置 User-Agent 黑名单
func LoadWAFConfigFromEnv() WAFConfig {
	config := DefaultWAFConfig()

	// 从环境变量加载额外的 User-Agent 黑名单
	// 格式: WAF_BLOCKED_USER_AGENTS=bot,crawler,spider,custom-bot
	if v := os.Getenv("WAF_BLOCKED_USER_AGENTS"); v != "" {
		additionalAgents := strings.Split(v, ",")
		for _, agent := range additionalAgents {
			agent = strings.TrimSpace(strings.ToLower(agent))
			if agent != "" && !containsString(config.BlockedUserAgents, agent) {
				config.BlockedUserAgents = append(config.BlockedUserAgents, agent)
			}
		}
	}

	// 是否阻止空 User-Agent
	// 格式: WAF_BLOCK_EMPTY_USER_AGENT=true/false
	if v := os.Getenv("WAF_BLOCK_EMPTY_USER_AGENT"); v != "" {
		config.BlockEmptyUserAgent = strings.ToLower(v) == "true"
	}

	// 最大请求大小
	// 格式: WAF_MAX_REQUEST_SIZE=10MB
	if v := os.Getenv("WAF_MAX_REQUEST_SIZE"); v != "" {
		if size, err := parseWAFSize(v); err == nil && size > 0 {
			config.MaxRequestSize = size
		}
	}

	// 最大参数长度
	// 格式: WAF_MAX_ARGS_LENGTH=1000
	if v := os.Getenv("WAF_MAX_ARGS_LENGTH"); v != "" {
		if length, err := strconv.Atoi(v); err == nil && length > 0 {
			config.MaxArgsLength = length
		}
	}

	// FIX-P2: 生产环境强制启用 WAF，不允许禁用
	// 格式: WAF_ENABLED=true/false (仅开发环境有效)
	if gin.Mode() == gin.ReleaseMode {
		// 生产环境强制启用 WAF
		config.Enabled = true
	} else {
		// 开发环境允许通过环境变量控制
		if v := os.Getenv("WAF_ENABLED"); v != "" {
			config.Enabled = strings.ToLower(v) == "true"
		}
	}

	return config
}

// parseWAFSize 解析大小字符串 (如 "10MB", "1KB")
func parseWAFSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	multiplier := int64(1)

	if strings.HasSuffix(s, "kb") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "kb")
	} else if strings.HasSuffix(s, "mb") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	} else if strings.HasSuffix(s, "gb") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	}

	size, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return size * multiplier, nil
}

// containsString 检查字符串是否在列表中
func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// WAFMiddleware WAF 中间件
// 用途: 检查请求中的恶意攻击特征
// P1-11: 使用预编译正则表达式提升性能
func WAFMiddleware(config WAFConfig, logger *zap.Logger) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// P1-11: 初始化时预编译所有正则表达式
	initWAFPatterns(config)

	return func(c *gin.Context) {
		// 1. 检查请求方法
		method := c.Request.Method
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
		if !isAllowedMethod(method, allowedMethods) {
			blockRequest(c, logger, "method_not_allowed", "Method not allowed: "+method)
			return
		}

		// 2. 检查请求大小
		if c.Request.ContentLength > config.MaxRequestSize {
			blockRequest(c, logger, "request_too_large", "Request size exceeds limit")
			return
		}

		// 3. 检查 User-Agent
		// FIX-040: 使用配置化的 User-Agent 黑名单
		userAgent := c.Request.UserAgent()
		if isBlockedUserAgentWithConfig(userAgent, config.BlockedUserAgents, config.BlockEmptyUserAgent) {
			blockRequest(c, logger, "blocked_user_agent", "Blocked User-Agent: "+userAgent)
			return
		}

		// 4. 检查 URL 参数
		queryParams := c.Request.URL.Query()
		for key, values := range queryParams {
			// 检查参数长度
			if len(key) > config.MaxArgsLength || len(values) > 100 {
				blockRequest(c, logger, "args_too_long", "Arguments too long")
				return
			}

			for _, value := range values {
				// SQL 注入检测 - P1-11: 使用预编译正则表达式
				if detectAttack(value, compiledPatterns.sqlInjection) {
					blockRequest(c, logger, "sql_injection", "SQL Injection detected in query param: "+key)
					return
				}

				// XSS 检测 - P1-11: 使用预编译正则表达式
				if detectAttack(value, compiledPatterns.xss) {
					blockRequest(c, logger, "xss_attack", "XSS Attack detected in query param: "+key)
					return
				}

				// 路径遍历检测 - P1-11: 使用预编译正则表达式
				if detectAttack(value, compiledPatterns.pathTraversal) {
					blockRequest(c, logger, "path_traversal", "Path Traversal detected in query param: "+key)
					return
				}

				// 命令注入检测 - P1-11: 使用预编译正则表达式
				if detectAttack(value, compiledPatterns.commandInjection) {
					blockRequest(c, logger, "command_injection", "Command Injection detected in query param: "+key)
					return
				}

				// SSRF 检测 - P1-11: 使用预编译正则表达式
				if detectAttack(value, compiledPatterns.ssrf) {
					blockRequest(c, logger, "ssrf_attack", "SSRF Attack detected in query param: "+key)
					return
				}
			}
		}

		// 5. 检查请求体 (POST/PUT/DELETE)
		if method == "POST" || method == "PUT" || method == "DELETE" {
			// 读取请求体
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			body := string(bodyBytes)

			// SQL 注入检测 - P1-11: 使用预编译正则表达式
			if detectAttack(body, compiledPatterns.sqlInjection) {
				blockRequest(c, logger, "sql_injection", "SQL Injection detected in request body")
				return
			}

			// XSS 检测 - P1-11: 使用预编译正则表达式
			if detectAttack(body, compiledPatterns.xss) {
				blockRequest(c, logger, "xss_attack", "XSS Attack detected in request body")
				return
			}

			// 路径遍历检测 - P1-11: 使用预编译正则表达式
			if detectAttack(body, compiledPatterns.pathTraversal) {
				blockRequest(c, logger, "path_traversal", "Path Traversal detected in request body")
				return
			}

			// 命令注入检测 - P1-11: 使用预编译正则表达式
			if detectAttack(body, compiledPatterns.commandInjection) {
				blockRequest(c, logger, "command_injection", "Command Injection detected in request body")
				return
			}

			// SSRF 检测 - P1-11: 使用预编译正则表达式
			if detectAttack(body, compiledPatterns.ssrf) {
				blockRequest(c, logger, "ssrf_attack", "SSRF Attack detected in request body")
				return
			}
		}

		// 6. 检查路径
		path := c.Request.URL.Path

		// 检查敏感路径
		if isSensitivePath(path) {
			blockRequest(c, logger, "sensitive_path", "Sensitive path access blocked: "+path)
			return
		}

		// 路径遍历检测 - P1-11: 使用预编译正则表达式
		if detectAttack(path, compiledPatterns.pathTraversal) {
			blockRequest(c, logger, "path_traversal", "Path Traversal detected in URL path")
			return
		}

		// 7. 检查文件上传扩展名
		if strings.Contains(path, "/upload") || strings.Contains(path, "/file") {
			// 从 Content-Disposition 获取文件名
			contentDisposition := c.GetHeader("Content-Disposition")
			if contentDisposition != "" {
				filename := extractFilename(contentDisposition)
				if isBlockedExtension(filename, config.BlockedExtensions) {
					blockRequest(c, logger, "dangerous_upload", "Dangerous file upload blocked: "+filename)
					return
				}
			}
		}

		// 所有检查通过，继续处理请求
		c.Next()
	}
}

// ============================================
// 辅助函数
// ============================================

// isAllowedMethod 检查是否是允许的方法
func isAllowedMethod(method string, allowedMethods []string) bool {
	for _, m := range allowedMethods {
		if strings.ToUpper(method) == m {
			return true
		}
	}
	return false
}

// isBlockedUserAgentWithConfig 检查是否是禁止的 User-Agent (使用配置)
// FIX-040: 支持配置化的 User-Agent 黑名单
func isBlockedUserAgentWithConfig(userAgent string, blockedAgents []string, blockEmpty bool) bool {
	// 检查空 User-Agent
	if userAgent == "" {
		return blockEmpty
	}

	// 检查是否包含黑名单中的关键词
	userAgentLower := strings.ToLower(userAgent)
	for _, agent := range blockedAgents {
		if strings.Contains(userAgentLower, strings.ToLower(agent)) {
			return true
		}
	}

	return false
}

// isBlockedUserAgent is kept for backward compatibility
// Deprecated: Use isBlockedUserAgentWithConfig instead
// nolint:unused
func isBlockedUserAgent(userAgent string) bool {
	// 使用默认配置
	defaultConfig := DefaultWAFConfig()
	return isBlockedUserAgentWithConfig(userAgent, defaultConfig.BlockedUserAgents, defaultConfig.BlockEmptyUserAgent)
}

// initWAFPatterns 初始化预编译的正则表达式
// P1-11: 在初始化时预编译所有正则表达式以提升性能
func initWAFPatterns(config WAFConfig) {
	if compiledPatterns.initialized {
		return
	}

	// SQL注入模式预编译
	for _, p := range config.SQLInjectionPatterns {
		compiledPatterns.sqlInjection = append(compiledPatterns.sqlInjection, regexp.MustCompile(p))
	}

	// XSS模式预编译
	for _, p := range config.XSSPatterns {
		compiledPatterns.xss = append(compiledPatterns.xss, regexp.MustCompile(p))
	}

	// 路径遍历模式预编译
	for _, p := range config.PathTraversalPatterns {
		compiledPatterns.pathTraversal = append(compiledPatterns.pathTraversal, regexp.MustCompile(p))
	}

	// 命令注入模式预编译
	for _, p := range config.CommandInjectionPatterns {
		compiledPatterns.commandInjection = append(compiledPatterns.commandInjection, regexp.MustCompile(p))
	}

	// SSRF模式预编译
	for _, p := range config.SSRFPatterns {
		compiledPatterns.ssrf = append(compiledPatterns.ssrf, regexp.MustCompile(p))
	}

	compiledPatterns.initialized = true
}

// detectAttack 检测攻击特征
// P1-11: 使用预编译的正则表达式
func detectAttack(input string, patterns []*regexp.Regexp) bool {
	// URL 解码
	decodedInput, err := url.QueryUnescape(input)
	if err != nil {
		decodedInput = input
	}

	// 使用预编译的正则表达式进行匹配
	for _, re := range patterns {
		if re.MatchString(decodedInput) {
			return true
		}
	}

	return false
}

// isBlockedExtension 检查是否是禁止的文件扩展名
func isBlockedExtension(filename string, blockedExtensions []string) bool {
	if filename == "" {
		return false
	}

	for _, ext := range blockedExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}

	return false
}

// extractFilename 从 Content-Disposition 提取文件名
func extractFilename(contentDisposition string) string {
	// 解析 Content-Disposition: attachment; filename="test.php"
	parts := strings.Split(contentDisposition, ";")
	for _, part := range parts {
		if strings.Contains(part, "filename") {
			// 提取文件名
			filename := strings.TrimPrefix(strings.TrimSpace(part), "filename=")
			filename = strings.Trim(filename, "\"")
			return filename
		}
	}
	return ""
}

// blockRequest 阻止请求
func blockRequest(c *gin.Context, logger *zap.Logger, rule, reason string) {
	// 记录被阻止的请求
	logger.Warn("WAF blocked request",
		zap.String("rule", rule),
		zap.String("reason", reason),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	// 设置 WAF 阻止状态
	c.Set("waf_blocked", true)
	c.Set("waf_rule", rule)
	c.Set("waf_reason", reason)

	// 返回 403 Forbidden
	c.AbortWithStatusJSON(403, gin.H{
		"error":   "Forbidden",
		"message": "Request blocked by WAF",
		"rule":    rule,
		"reason":  reason,
	})
}

// ============================================
// WAF 统计
// ============================================

// WAFStats WAF 统计
type WAFStats struct {
	TotalRequests     int64 `json:"total_requests"`
	BlockedRequests   int64 `json:"blocked_requests"`
	SQLInjection      int64 `json:"sql_injection"`
	XSSAttacks        int64 `json:"xss_attacks"`
	PathTraversal     int64 `json:"path_traversal"`
	CommandInjection  int64 `json:"command_injection"`
	SSRFAttacks       int64 `json:"ssrf_attacks"`
	DangerousUploads  int64 `json:"dangerous_uploads"`
	MethodNotAllowed  int64 `json:"method_not_allowed"`
	BlockedAgents     int64 `json:"blocked_agents"`
	RateLimitExceeded int64 `json:"rate_limit_exceeded"`
}

// WAFStatsMiddleware WAF 统计中间件
// 用途: 统计 WAF 检查结果
func WAFStatsMiddleware(stats *WAFStats) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 增加总请求计数
		atomic.AddInt64(&stats.TotalRequests, 1)

		// 处理请求
		c.Next()

		// 检查是否被 WAF 阻止
		if blocked, exists := c.Get("waf_blocked"); exists && blocked == true {
			atomic.AddInt64(&stats.BlockedRequests, 1)

			// 根据规则增加对应计数
			rule := c.GetString("waf_rule")
			switch rule {
			case "sql_injection":
				atomic.AddInt64(&stats.SQLInjection, 1)
			case "xss_attack":
				atomic.AddInt64(&stats.XSSAttacks, 1)
			case "path_traversal":
				atomic.AddInt64(&stats.PathTraversal, 1)
			case "command_injection":
				atomic.AddInt64(&stats.CommandInjection, 1)
			case "ssrf_attack":
				atomic.AddInt64(&stats.SSRFAttacks, 1)
			case "dangerous_upload":
				atomic.AddInt64(&stats.DangerousUploads, 1)
			case "method_not_allowed":
				atomic.AddInt64(&stats.MethodNotAllowed, 1)
			case "blocked_user_agent":
				atomic.AddInt64(&stats.BlockedAgents, 1)
			case "rate_limit":
				atomic.AddInt64(&stats.RateLimitExceeded, 1)
			}
		}
	}
}

// isSensitivePath 检查是否是敏感路径
func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/hosts",
		"/proc/self",
		"/sys/kernel",
		"/.env",
		"/.git",
		"/.svn",
		"/.htaccess",
		"/config.php",
		"/wp-config.php",
		"/web.config",
		"/.htpasswd",
	}

	pathLower := strings.ToLower(path)
	for _, sensitive := range sensitivePaths {
		if strings.Contains(pathLower, strings.ToLower(sensitive)) {
			return true
		}
	}

	return false
}
