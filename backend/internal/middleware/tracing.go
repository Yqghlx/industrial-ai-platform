package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/industrial-ai/platform/pkg/tracing"
)

// ============================================
// OpenTelemetry 追踪中间件
// ============================================

// TracingMiddleware 追踪中间件
// 用途: 为每个 HTTP 请求创建 OpenTelemetry Span
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// 从请求头提取 TraceContext
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 创建 Span
		spanName := c.Request.Method + " " + c.FullPath()
		if spanName == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.scheme", c.Request.URL.Scheme),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.route", c.FullPath()),
				attribute.Int64("http.request_content_length", c.Request.ContentLength),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)

		// 设置到上下文
		c.Request = c.Request.WithContext(ctx)

		// 提取 TraceID 和 SpanID
		traceID := tracing.GetTraceID(ctx)
		spanID := tracing.GetSpanID(ctx)

		// 设置到 Gin 上下文 (用于日志关联)
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)

		// 设置响应头
		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)
		latencyMs := float64(latency.Milliseconds())

		// 设置响应属性
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_content_length", c.Writer.Size()),
			attribute.Float64("http.latency_ms", latencyMs),
		)

		// 设置 Span 状态
		if c.Writer.Status() >= 500 {
			span.SetStatus(codes.Error, "HTTP Server Error")
			span.SetAttributes(
				attribute.String("error.type", "http_error"),
				attribute.String("error.message", c.Errors.String()),
			)
		} else if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP Client Error")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// 结束 Span
		span.End()
	}
}

// ============================================
// TraceContext 传播中间件
// ============================================

// TraceContextPropagationMiddleware TraceContext 传播中间件
// 用途: 在响应头中传播 TraceContext
func TraceContextPropagationMiddleware() gin.HandlerFunc {
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		// 在响应头中注入 TraceContext
		propagator.Inject(c.Request.Context(), propagation.HeaderCarrier(c.Writer.Header()))
	}
}

// ============================================
// 数据库追踪中间件
// ============================================

// DatabaseSpanMiddleware 数据库追踪中间件
// 用途: 为数据库操作创建 Span
func DatabaseSpanMiddleware(ctx context.Context, tracer trace.Tracer, operation, table, statement string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "DB: "+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
			attribute.String("db.table", table),
			attribute.String("db.statement", statement),
		),
	)
}

// ============================================
// Redis 追踪中间件
// ============================================

// RedisSpanMiddleware Redis 追踪中间件
// 用途: 为 Redis 操作创建 Span
func RedisSpanMiddleware(ctx context.Context, tracer trace.Tracer, operation, key string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "Redis: "+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", operation),
			attribute.String("redis.key", key),
		),
	)
}

// ============================================
// 外部服务追踪中间件
// ============================================

// ExternalServiceSpanMiddleware 外部服务追踪中间件
// 用途: 为外部服务调用创建 Span
func ExternalServiceSpanMiddleware(ctx context.Context, tracer trace.Tracer, serviceName, operation string) (context.Context, trace.Span) {
	return tracer.Start(ctx, serviceName+": "+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.operation", operation),
		),
	)
}

// ============================================
// GLM API 追踪中间件
// ============================================

// GLMAPISpanMiddleware GLM API 追踪中间件
// 用途: 为 GLM API 调用创建 Span
func GLMAPISpanMiddleware(ctx context.Context, tracer trace.Tracer, operation string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "GLM API: "+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("service.name", "glm-api"),
			attribute.String("service.operation", operation),
			attribute.String("service.provider", "zhipu"),
		),
	)
}

// ============================================
// 追踪错误记录
// ============================================

// RecordSpanError 记录 Span 错误
func RecordSpanError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err, trace.WithAttributes(attrs...))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// ============================================
// 追踪性能记录
// ============================================

// RecordSpanLatency 记录 Span 延迟
func RecordSpanLatency(ctx context.Context, latencyMs float64) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Float64("latency_ms", latencyMs),
	)
}

// ============================================
// 追踪计数记录
// ============================================

// RecordSpanCount 记录 Span 计数
func RecordSpanCount(ctx context.Context, count int64, name string) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64(name, count),
	)
}
