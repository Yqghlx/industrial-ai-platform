package benchmarks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

// 简化版API基准测试

// BenchmarkHealthCheck 健康检查API基准测试
func BenchmarkHealthCheck(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkJSONMarshal JSON序列化基准测试
func BenchmarkJSONMarshal(b *testing.B) {
	data := map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"count":   100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}

// BenchmarkJSONUnmarshal JSON反序列化基准测试
func BenchmarkJSONUnmarshal(b *testing.B) {
	data := []byte(`{"status":"ok","version":"1.0.0","count":100}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		_ = json.Unmarshal(data, &result)
	}
}

// BenchmarkMapAccess map访问基准测试
func BenchmarkMapAccess(b *testing.B) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m["key1"]
	}
}

// BenchmarkMapWrite map写入基准测试
func BenchmarkMapWrite(b *testing.B) {
	m := make(map[string]string)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m["key"] = "value"
	}
}

// BenchmarkSliceAppend slice追加基准测试
func BenchmarkSliceAppend(b *testing.B) {
	s := make([]int, 0, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = append(s, i)
	}
}

// BenchmarkStringConcat 字符串拼接基准测试
func BenchmarkStringConcat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = "prefix_" + "suffix"
	}
}

// BenchmarkBufferWrite Buffer写入基准测试
func BenchmarkBufferWrite(b *testing.B) {
	buf := bytes.NewBuffer(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.WriteString("test data")
	}
}

// BenchmarkConcurrentMap 并发map基准测试 (使用sync.Map)
func BenchmarkConcurrentMap(b *testing.B) {
	m := sync.Map{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("key", 1)
		}
	})
}

// BenchmarkHTTPRequest HTTP请求基准测试
func BenchmarkHTTPRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"result": "success"})
	})

	body := []byte(`{"data":"test"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
