package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// ============================================
// OpenTelemetry 追踪配置
// ============================================

// Config 追踪配置
type Config struct {
	ServiceName    string  // 服务名称
	ServiceVersion string  // 服务版本
	Environment    string  // 环境 (development/staging/production)
	CollectorURL   string  // Collector 地址 (e.g., localhost:4317)
	CollectorType  string  // Collector 类型 (grpc/http)
	SampleRate     float64 // 采样率 (0.0 - 1.0)
	Enabled        bool    // 是否启用追踪
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		ServiceName:    "industrial-ai-backend",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		CollectorURL:   "localhost:4317",
		CollectorType:  "grpc",
		SampleRate:     0.1, // 10% 采样率
		Enabled:        true,
	}
}

// ============================================
// OpenTelemetry Tracer Provider
// ============================================

// TracerProvider 追踪器提供者
type TracerProvider struct {
	*sdktrace.TracerProvider
	config Config
	tracer trace.Tracer
}

// NewTracerProvider 创建追踪器提供者
func NewTracerProvider(ctx context.Context, cfg Config) (*TracerProvider, error) {
	if !cfg.Enabled {
		// 返回 Noop TracerProvider
		noopTP := sdktrace.NewTracerProvider()
		return &TracerProvider{
			TracerProvider: noopTP,
			config:         cfg,
			tracer:         noopTP.Tracer(cfg.ServiceName),
		}, nil
	}

	// 创建 exporter
	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// 创建资源
	res, err := createResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithExportTimeout(30*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
		sdktrace.WithSpanLimits(sdktrace.SpanLimits{
			AttributeValueLengthLimit: 4096,
			AttributeCountLimit:       128,
			EventCountLimit:           128,
			LinkCountLimit:            128,
		}),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 设置全局 Propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 创建 Tracer
	tracer := tp.Tracer(cfg.ServiceName)

	return &TracerProvider{
		TracerProvider: tp,
		config:         cfg,
		tracer:         tracer,
	}, nil
}

// createExporter 创建 exporter
func createExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	switch cfg.CollectorType {
	case "grpc":
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.CollectorURL),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithTimeout(30*time.Second),
		)
		if err != nil {
			return nil, err
		}
		return exp, nil
	case "http":
		exp, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.CollectorURL),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithTimeout(30*time.Second),
		)
		if err != nil {
			return nil, err
		}
		return exp, nil
	default:
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.CollectorURL),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
		return exp, nil
	}
}

// createResource 创建资源
func createResource(cfg Config) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
			attribute.String("service.instance.id", os.Getenv("HOSTNAME")),
		),
	)
}

// ============================================
// Tracer 方法
// ============================================

// GetTracer 获取 Tracer
func (tp *TracerProvider) GetTracer() trace.Tracer {
	return tp.tracer
}

// StartSpan 开始 Span
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// ============================================
// 上下文追踪方法
// ============================================

// GetTraceID 从上下文获取 TraceID
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID 从上下文获取 SpanID
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// IsSampled 检查是否被采样
func IsSampled(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().IsSampled()
}

// ============================================
// Span 辅助方法
// ============================================

// AddAttributeToSpan 添加属性到 Span
func AddAttributeToSpan(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// AddEventToSpan 添加事件到 Span
func AddEventToSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanStatus 设置 Span 状态
func SetSpanStatus(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// ============================================
// HTTP 追踪辅助
// ============================================

// HTTPAttributes HTTP 请求属性
func HTTPAttributes(method, path, host, scheme string, statusCode int, latencyMs float64) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(method),
		attribute.String("http.url", path),
		attribute.String("http.host", host),
		attribute.String("http.scheme", scheme),
		semconv.HTTPResponseStatusCodeKey.Int(statusCode),
		attribute.Float64("http.latency_ms", latencyMs),
	}
}

// DatabaseAttributes 数据库操作属性
func DatabaseAttributes(operation, table, statement string, rowCount int64, latencyMs float64) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemKey.String("postgresql"),
		semconv.DBOperationKey.String(operation),
		semconv.DBSQLTableKey.String(table),
		attribute.String("db.statement", statement),
		attribute.Int64("db.row_count", rowCount),
		attribute.Float64("db.latency_ms", latencyMs),
	}
}

// RedisAttributes Redis 操作属性
func RedisAttributes(operation, key string, latencyMs float64) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemKey.String("redis"),
		semconv.DBOperationKey.String(operation),
		attribute.String("redis.key", key),
		attribute.Float64("redis.latency_ms", latencyMs),
	}
}

// ============================================
// 全局追踪器
// ============================================

var globalTracerProvider *TracerProvider

// InitGlobalTracer 初始化全局追踪器
func InitGlobalTracer(ctx context.Context, cfg Config) error {
	tp, err := NewTracerProvider(ctx, cfg)
	if err != nil {
		return err
	}
	globalTracerProvider = tp
	return nil
}

// GetGlobalTracerProvider 获取全局追踪器提供者
func GetGlobalTracerProvider() *TracerProvider {
	if globalTracerProvider == nil {
		// 返回 Noop TracerProvider
		tp, _ := NewTracerProvider(context.Background(), Config{Enabled: false})
		globalTracerProvider = tp
	}
	return globalTracerProvider
}

// Shutdown 关闭追踪器
func Shutdown(ctx context.Context) error {
	if globalTracerProvider != nil {
		return globalTracerProvider.Shutdown(ctx)
	}
	return nil
}
