package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"https://example.com", "https://admin.example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORS_NoOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin_ReleaseMode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin_DebugMode(t *testing.T) {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_OptionsPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_Wildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"*"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://any-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSWithConfig_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSWithConfig(nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://industrial-ai.example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://industrial-ai.example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSWithConfig_CustomConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowedOrigins:   []string{"https://custom.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Custom-Header"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://custom.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://custom.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "X-Custom-Header", w.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSWithConfig_NoCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET"},
		AllowCredentials: false,
	}

	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSWithConfig_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCORSConfig()

	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSWithConfig_ZeroMaxAge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
		MaxAge:         0,
	}

	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSWithConfig_OptionsPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCORSWithConfig_NoOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
	}

	router := gin.New()
	router.Use(CORSWithConfig(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestIsOriginAllowed_ExactMatch(t *testing.T) {
	assert.True(t, isOriginAllowed("https://example.com", []string{"https://example.com"}))
	assert.False(t, isOriginAllowed("https://evil.com", []string{"https://example.com"}))
}

func TestIsOriginAllowed_Wildcard(t *testing.T) {
	assert.True(t, isOriginAllowed("https://anything.com", []string{"*"}))
}

func TestIsOriginAllowed_SubdomainWildcard(t *testing.T) {
	assert.True(t, isOriginAllowed("https://sub.example.com", []string{"*.example.com"}))
	// The base domain itself doesn't match *.example.com pattern
	// because the wildcard requires at least one subdomain character before the domain
}

func TestIsOriginAllowed_EmptyList(t *testing.T) {
	assert.False(t, isOriginAllowed("https://example.com", []string{}))
}

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{"*"}},
		{"single", "https://example.com", []string{"https://example.com"}},
		{"multiple", "https://a.com, https://b.com", []string{"https://a.com", "https://b.com"}},
		{"with spaces", " https://a.com , https://b.com ", []string{"https://a.com", "https://b.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseOrigins(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.AllowedOrigins)
	assert.NotEmpty(t, config.AllowedMethods)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, CORSMaxAgeSeconds, config.MaxAge)
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestSetSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SetSecurityHeaders(c)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
}

func TestBuildCSPHeader(t *testing.T) {
	header := BuildCSPHeader()
	assert.NotEmpty(t, header)
	assert.Contains(t, header, "default-src")
	assert.Contains(t, header, "script-src")
}

func TestBuildCSPFromConfig(t *testing.T) {
	config := &CSPConfig{
		DefaultSrc: []string{"'self'"},
		ScriptSrc:  []string{"'self'", "'unsafe-inline'"},
		ReportURI:  "/csp-report",
	}
	header := BuildCSPFromConfig(config)
	assert.Contains(t, header, "default-src 'self'")
	assert.Contains(t, header, "script-src 'self' 'unsafe-inline'")
	assert.Contains(t, header, "report-uri /csp-report")
}

func TestHSTS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		maxAge           int
		includeSubDomain bool
		preload          bool
		expectedContains []string
	}{
		{"basic", 3600, false, false, []string{"max-age=3600"}},
		{"with subdomain", 3600, true, false, []string{"max-age=3600", "includeSubDomains"}},
		{"with preload", 3600, false, true, []string{"max-age=3600", "preload"}},
		{"full", 31536000, true, true, []string{"max-age=31536000", "includeSubDomains", "preload"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(HSTS(tt.maxAge, tt.includeSubDomain, tt.preload))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			hstsValue := w.Header().Get("Strict-Transport-Security")
			for _, expected := range tt.expectedContains {
				assert.Contains(t, hstsValue, expected)
			}
		})
	}
}

func TestForceHTTPS_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ForceHTTPS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("X-Forwarded-Host", "example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 301, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "https://example.com/test")
}

func TestForceHTTPS_NoRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ForceHTTPS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCORSSecurity_BackwardCompat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSSecurity([]string{"https://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRequestID_New(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(200, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestID_Existing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "existing-id", w.Header().Get("X-Request-ID"))
}

func TestDefaultCSPConfig(t *testing.T) {
	config := DefaultCSPConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.DefaultSrc)
}

func TestDevelopmentCSPConfig(t *testing.T) {
	config := DevelopmentCSPConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.DefaultSrc)
}
