package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTracing creates a test tracer provider for testing
func setupTracing(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	otel.SetTracerProvider(tp)
	return tp, spanRecorder
}

// ============================================
// TracingMiddleware Tests
// ============================================

func TestTracingMiddleware_BasicRequest(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	router := gin.New()
	router.Use(TracingMiddleware("test-service"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Check that trace headers are set
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
	assert.NotEmpty(t, w.Header().Get("X-Span-ID"))

	// Verify span was created
	spans := recorder.Ended()
	assert.Len(t, spans, 1)
}

func TestTracingMiddleware_WithPath(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	router := gin.New()
	router.Use(TracingMiddleware("test-service"))
	router.GET("/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// Verify span was created with correct name
	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Contains(t, spans[0].Name(), "GET")
}

func TestTracingMiddleware_ErrorStatus(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	router := gin.New()
	router.Use(TracingMiddleware("test-service"))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)

	// Verify span has error status
	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestTracingMiddleware_ClientError(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	router := gin.New()
	router.Use(TracingMiddleware("test-service"))
	router.GET("/notfound", func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "not found"})
	})

	req := httptest.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	// Verify span has error status for 4xx
	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestTracingMiddleware_ContextPropagation(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	router := gin.New()
	router.Use(TracingMiddleware("test-service"))
	router.GET("/test", func(c *gin.Context) {
		// Check that trace_id and span_id are set in context
		traceID, exists := c.Get("trace_id")
		assert.True(t, exists)
		assert.NotEmpty(t, traceID)

		spanID, exists := c.Get("span_id")
		assert.True(t, exists)
		assert.NotEmpty(t, spanID)

		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	spans := recorder.Ended()
	assert.Len(t, spans, 1)
}

// ============================================
// TraceContextPropagationMiddleware Tests
// ============================================

func TestTraceContextPropagationMiddleware_Basic(t *testing.T) {
	tp, _ := setupTracing(t)
	defer tp.Shutdown(context.Background())

	router := gin.New()
	router.Use(TraceContextPropagationMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// DatabaseSpanMiddleware Tests
// ============================================

func TestDatabaseSpanMiddleware_Basic(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	_, span := DatabaseSpanMiddleware(context.Background(), tracer, "SELECT", "users", "SELECT * FROM users")

	assert.NotNil(t, span)
	assert.NotEmpty(t, span.SpanContext().TraceID())

	// End the span before checking recorder
	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Contains(t, spans[0].Name(), "DB: SELECT")
}

// ============================================
// RedisSpanMiddleware Tests
// ============================================

func TestRedisSpanMiddleware_Basic(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	_, span := RedisSpanMiddleware(context.Background(), tracer, "GET", "user:123")

	assert.NotNil(t, span)

	// End the span before checking recorder
	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Contains(t, spans[0].Name(), "Redis: GET")
}

// ============================================
// ExternalServiceSpanMiddleware Tests
// ============================================

func TestExternalServiceSpanMiddleware_Basic(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	_, span := ExternalServiceSpanMiddleware(context.Background(), tracer, "payment-gateway", "process_payment")

	assert.NotNil(t, span)

	// End the span before checking recorder
	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Contains(t, spans[0].Name(), "payment-gateway")
}

// ============================================
// GLMAPISpanMiddleware Tests
// ============================================

func TestGLMAPISpanMiddleware_Basic(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	_, span := GLMAPISpanMiddleware(context.Background(), tracer, "chat_completion")

	assert.NotNil(t, span)

	// End the span before checking recorder
	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Contains(t, spans[0].Name(), "GLM API")
}

// ============================================
// RecordSpanError Tests
// ============================================

func TestRecordSpanError_WithError(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")

	testErr := assert.AnError
	RecordSpanError(ctx, testErr, attribute.String("key", "value"))

	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestRecordSpanError_WithNilError(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")

	RecordSpanError(ctx, nil)

	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Ok, spans[0].Status().Code)
}

// ============================================
// RecordSpanLatency Tests
// ============================================

func TestRecordSpanLatency(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")

	RecordSpanLatency(ctx, 123.45)

	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	// Verify latency attribute was set
	attrs := spans[0].Attributes()
	var foundLatency bool
	for _, attr := range attrs {
		if attr.Key == "latency_ms" {
			foundLatency = true
			assert.Equal(t, 123.45, attr.Value.AsFloat64())
		}
	}
	assert.True(t, foundLatency)
}

// ============================================
// RecordSpanCount Tests
// ============================================

func TestRecordSpanCount(t *testing.T) {
	tp, recorder := setupTracing(t)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")

	RecordSpanCount(ctx, 42, "request_count")

	span.End()

	spans := recorder.Ended()
	assert.Len(t, spans, 1)
	// Verify count attribute was set
	attrs := spans[0].Attributes()
	var foundCount bool
	for _, attr := range attrs {
		if attr.Key == "request_count" {
			foundCount = true
			assert.Equal(t, int64(42), attr.Value.AsInt64())
		}
	}
	assert.True(t, foundCount)
}