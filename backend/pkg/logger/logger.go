package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ============================================
// 结构化日志配置
// ============================================

// Config 日志配置
type Config struct {
	Level       string // 日志级别 (debug/info/warn/error)
	Format      string // 日志格式 (json/console)
	ServiceName string // 服务名称
	Environment string // 环境 (development/staging/production)
	Version     string // 服务版本
	Output      string // 输出路径 (stdout/stderr/file)
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "industrial-ai-backend",
		Environment: "production",
		Version:     "1.0.0",
		Output:      "stdout",
	}
}

// ============================================
// Zap Logger 封装
// ============================================

// Logger 结构化日志器
type Logger struct {
	*zap.Logger
	config Config
}

// NewLogger 创建结构化日志器
func NewLogger(cfg Config) (*Logger, error) {
	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "source",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建 Core
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		level,
	)

	// 创建 Logger
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.String("service", cfg.ServiceName),
			zap.String("environment", cfg.Environment),
			zap.String("version", cfg.Version),
		),
	)

	return &Logger{
		Logger: zapLogger,
		config: cfg,
	}, nil
}

// ============================================
// 上下文日志方法
// ============================================

// WithContext 添加上下文信息
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	// 提取上下文中的 TraceID
	traceID := ctx.Value("trace_id")
	if traceID != nil {
		return l.With(zap.String("trace_id", traceID.(string)))
	}
	return l.Logger
}

// WithTraceID 添加追踪 ID
func (l *Logger) WithTraceID(traceID string) *zap.Logger {
	return l.With(zap.String("trace_id", traceID))
}

// WithRequestID 添加请求 ID
func (l *Logger) WithRequestID(requestID string) *zap.Logger {
	return l.With(zap.String("request_id", requestID))
}

// WithTenantID 添加租户 ID
func (l *Logger) WithTenantID(tenantID string) *zap.Logger {
	return l.With(zap.String("tenant_id", tenantID))
}

// WithUserID 添加用户 ID
func (l *Logger) WithUserID(userID string) *zap.Logger {
	return l.With(zap.String("user_id", userID))
}

// WithError 添加错误信息
func (l *Logger) WithError(err error) *zap.Logger {
	if err != nil {
		return l.With(zap.String("error", err.Error()))
	}
	return l.Logger
}

// ============================================
// HTTP 请求日志方法
// ============================================

// HTTPFields HTTP 请求日志字段
func HTTPFields(method, path string, statusCode int, latency time.Duration, requestSize, responseSize int64) []zap.Field {
	return []zap.Field{
		zap.String("http.method", method),
		zap.String("http.path", path),
		zap.Int("http.status_code", statusCode),
		zap.Float64("http.latency_ms", float64(latency.Milliseconds())),
		zap.Int64("http.request_size_bytes", requestSize),
		zap.Int64("http.response_size_bytes", responseSize),
	}
}

// LogHTTPRequest 记录 HTTP 请求
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, latency time.Duration, requestSize, responseSize int64) {
	fields := HTTPFields(method, path, statusCode, latency, requestSize, responseSize)

	if statusCode >= 500 {
		l.Error("HTTP request error", fields...)
	} else if statusCode >= 400 {
		l.Warn("HTTP request warning", fields...)
	} else {
		l.Info("HTTP request processed", fields...)
	}
}

// ============================================
// 全局日志器
// ============================================

var globalLogger *Logger

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger(cfg Config) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger 获取全局日志器
func GetLogger() *Logger {
	if globalLogger == nil {
		// 初始化默认日志器
		logger, _ := NewLogger(DefaultConfig())
		globalLogger = logger
	}
	return globalLogger
}

// L 获取全局日志器 (快捷方式)
func L() *Logger {
	return GetLogger()
}

// ============================================
// 动态配置方法 (BE-P2-04)
// ============================================

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(level string) error {
	_, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", level)
	}

	// 更新配置
	// 注意: zap.Logger 的 Core 通常是不可变的
	// 保存新的配置，下次重建 Logger 时会使用新级别
	l.config.Level = level

	return nil
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() string {
	return l.config.Level
}

// GetLevelZap 获取当前 zap 日志级别
func (l *Logger) GetLevelZap() zapcore.Level {
	level, _ := zapcore.ParseLevel(l.config.Level)
	return level
}

// SetFormat 动态设置日志格式
func (l *Logger) SetFormat(format string) {
	l.config.Format = format
}

// GetFormat 获取当前日志格式
func (l *Logger) GetFormat() string {
	return l.config.Format
}

// ReloadConfig 重新加载配置并重建 Logger
func (l *Logger) ReloadConfig(cfg Config) error {
	newLogger, err := NewLogger(cfg)
	if err != nil {
		return err
	}

	// 关闭旧 Logger
	l.Sync()

	// 替换 Logger
	l.Logger = newLogger.Logger
	l.config = cfg

	return nil
}

// GetConfig 获取当前配置
func (l *Logger) GetConfig() Config {
	return l.config
}

// ============================================
// 全局日志器动态配置 (BE-P2-04)
// ============================================

// SetGlobalLevel 设置全局日志级别
func SetGlobalLevel(level string) error {
	if globalLogger == nil {
		return fmt.Errorf("global logger not initialized")
	}
	return globalLogger.SetLevel(level)
}

// GetGlobalLevel 获取全局日志级别
func GetGlobalLevel() string {
	if globalLogger == nil {
		return "info"
	}
	return globalLogger.GetLevel()
}

// ReloadGlobalLogger 重新加载全局日志器配置
func ReloadGlobalLogger(cfg Config) error {
	return InitGlobalLogger(cfg)
}

// ============================================
// 辅助函数
// ============================================

// Sync 同步日志缓冲
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close 关闭日志器
func (l *Logger) Close() error {
	return l.Sync()
}
