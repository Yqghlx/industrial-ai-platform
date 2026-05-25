package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================
// CSRFWithConfig Full Tests
// ============================================

func TestCSRFWithConfig_PostWithValidToken(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// First, get a CSRF token via GET request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the CSRF token from the cookie
	cookies := w.Result().Cookies()
	var csrfToken string
	for _, cookie := range cookies {
		if cookie.Name == config.CookieName {
			// Decode URL-encoded cookie value
			decoded, err := url.QueryUnescape(cookie.Value)
			require.NoError(t, err)
			csrfToken = decoded
			break
		}
	}
	require.NotEmpty(t, csrfToken, "CSRF token should be set in cookie")

	// Now make a POST request with the token
	postReq := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set(config.HeaderName, csrfToken)
	// Add the cookie (use URL-encoded value)
	for _, cookie := range cookies {
		if cookie.Name == config.CookieName {
			postReq.AddCookie(cookie)
			break
		}
	}

	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	assert.Equal(t, 200, postW.Code)
}

func TestCSRFWithConfig_PostWithTokenMismatch(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Set cookie with one token, but send different token in header
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.HeaderName, "wrong-token")
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: "correct-token",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF token mismatch")
}

func TestCSRFWithConfig_PostWithoutCookie(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// POST without any CSRF cookie
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.HeaderName, "some-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF token not found in cookie")
}

func TestCSRFWithConfig_PostWithTokenInForm(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	csrfToken := "test-form-token"

	// POST with token in form field
	formData := url.Values{}
	formData.Set(config.FormFieldName, csrfToken)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: csrfToken,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_PostWithTokenInQuery(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	csrfToken := "test-query-token"

	// POST with token in query parameter (fallback)
	req := httptest.NewRequest("POST", "/test?"+config.FormFieldName+"="+csrfToken, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: csrfToken,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_CustomErrorHandler(t *testing.T) {
	customErrorMsg := "custom csrf error"
	config := &CSRFConfig{
		TokenLength:   32,
		CookieName:    "csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "_csrf",
		CookiePath:    "/",
		ErrorHandler: func(c *gin.Context, errMsg string) {
			c.JSON(400, gin.H{
				"custom_error": customErrorMsg,
				"detail":       errMsg,
			})
			c.Abort()
		},
	}

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), customErrorMsg)
}

func TestCSRFWithConfig_NilErrorHandler(t *testing.T) {
	config := &CSRFConfig{
		TokenLength:   32,
		CookieName:    "csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "_csrf",
		CookiePath:    "/",
		ErrorHandler:  nil, // Should use default
	}

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should use default error handler and return 403
	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF validation failed")
}

func TestCSRFWithConfig_EmptyCookieToken(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Cookie exists but with empty value
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.HeaderName, "some-token")
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: "",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF token not found in cookie")
}

func TestCSRFWithConfig_EmptyRequestToken(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Cookie exists but no token in request
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: "some-token",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF token not found in request")
}

func TestCSRFWithConfig_PutMethod(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.PUT("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	csrfToken := "test-put-token"

	// PUT requires CSRF token
	req := httptest.NewRequest("PUT", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.HeaderName, csrfToken)
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: csrfToken,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_DeleteMethod(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.DELETE("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	csrfToken := "test-delete-token"

	// DELETE requires CSRF token
	req := httptest.NewRequest("DELETE", "/test", nil)
	req.Header.Set(config.HeaderName, csrfToken)
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: csrfToken,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_PatchMethod(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.PATCH("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	csrfToken := "test-patch-token"

	// PATCH requires CSRF token
	req := httptest.NewRequest("PATCH", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.HeaderName, csrfToken)
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: csrfToken,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_HeadMethod(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.HEAD("/test", func(c *gin.Context) {
		c.Status(200)
	})

	// HEAD is safe method, should pass without CSRF token
	req := httptest.NewRequest("HEAD", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_TraceMethod(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.Handle("TRACE", "/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// TRACE is safe method, should pass without CSRF token
	req := httptest.NewRequest("TRACE", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCSRFWithConfig_DefaultConfigValues(t *testing.T) {
	// Test that zero values are properly applied
	config := &CSRFConfig{
		// Leave all fields empty to test defaults
	}

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Should work with defaults applied
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail (no token) but not panic
	assert.Equal(t, 403, w.Code)
}

func TestCSRFWithConfig_ZeroTokenLength(t *testing.T) {
	config := &CSRFConfig{
		TokenLength:   0, // Should default to 32
		CookieName:    "csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "_csrf",
		CookiePath:    "/",
		ErrorHandler:  defaultCSRFErrorHandler,
	}

	// First GET to set token
	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.GET("/csrf", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/csrf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still set a cookie (with default length)
	cookies := w.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "csrf_token" {
			found = true
			assert.NotEmpty(t, cookie.Value)
			break
		}
	}
	assert.True(t, found, "CSRF cookie should be set even with zero token length")
}

func TestCSRFWithConfig_MultipleRequests(t *testing.T) {
	config := DevelopmentCSRFConfig() // Use development config (less strict)

	router := gin.New()
	router.Use(CSRFWithConfig(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Manually set a token - simulate getting token from GET and using it in POST
	token := GenerateCSRFToken(32)

	// POST with the token in both cookie and header
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(config.HeaderName, token)
	req.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: token,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// ============================================
// setCSRFToken Tests
// ============================================

func TestSetCSRFToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCSRFConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	token := setCSRFToken(c, config)
	assert.NotEmpty(t, token)

	// Verify cookie is set
	cookies := w.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == config.CookieName {
			found = true
			// Cookie values are URL-encoded, so decode for comparison
			decodedValue, err := url.QueryUnescape(cookie.Value)
			assert.NoError(t, err)
			assert.Equal(t, token, decodedValue)
			assert.Equal(t, config.CookiePath, cookie.Path)
			assert.Equal(t, config.CookieMaxAge, cookie.MaxAge)
			break
		}
	}
	assert.True(t, found)
}

// ============================================
// CSRFTokenFromContext Tests
// ============================================

func TestCSRFTokenFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultCSRFConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// No cookie set
	token := CSRFTokenFromContext(c, config)
	assert.Empty(t, token)

	// Set cookie
	c.Request.AddCookie(&http.Cookie{
		Name:  config.CookieName,
		Value: "test-token",
	})
	token = CSRFTokenFromContext(c, config)
	assert.Equal(t, "test-token", token)
}

func TestCSRFTokenFromContext_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Nil config should use default
	token := CSRFTokenFromContext(c, nil)
	assert.Empty(t, token)

	// Set cookie with default name
	c.Request.AddCookie(&http.Cookie{
		Name:  "csrf_token", // Default cookie name
		Value: "test-token",
	})
	token = CSRFTokenFromContext(c, nil)
	assert.Equal(t, "test-token", token)
}

// ============================================
// RequireCSRF Tests
// ============================================

func TestRequireCSRF_GetWithoutToken(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(RequireCSRF(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// GET without token should pass (safe method)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRequireCSRF_PostWithoutToken(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(RequireCSRF(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// POST without token should fail
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF token required")
}

func TestRequireCSRF_PostWithValidToken(t *testing.T) {
	config := DefaultCSRFConfig()

	router := gin.New()
	router.Use(RequireCSRF(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// First GET to get token
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	cookies := w1.Result().Cookies()
	var csrfToken string
	var rawCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == config.CookieName {
			decoded, err := url.QueryUnescape(cookie.Value)
			require.NoError(t, err)
			csrfToken = decoded
			rawCookie = cookie
			break
		}
	}
	require.NotEmpty(t, csrfToken)

	// POST with valid token
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(config.HeaderName, csrfToken)
	if rawCookie != nil {
		req2.AddCookie(rawCookie)
	}

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
}

func TestRequireCSRF_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireCSRF(nil)) // Should use default config
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail (no token)
	assert.Equal(t, 403, w.Code)
}

// ============================================
// secureCompare Edge Cases
// ============================================

func TestSecureCompare_DifferentLengths(t *testing.T) {
	// Different lengths should return false immediately
	assert.False(t, secureCompare("abc", "abcd"))
	assert.False(t, secureCompare("abcd", "abc"))
}

func TestSecureCompare_EmptyStrings(t *testing.T) {
	assert.True(t, secureCompare("", ""))
	assert.False(t, secureCompare("", "a"))
	assert.False(t, secureCompare("a", ""))
}

func TestSecureCompare_SpecialChars(t *testing.T) {
	assert.True(t, secureCompare("!@#$%", "!@#$%"))
	assert.False(t, secureCompare("!@#$%", "!@#$^"))
}

// ============================================
// isSafeMethod Edge Cases
// ============================================

func TestIsSafeMethod_LowerCase(t *testing.T) {
	assert.True(t, isSafeMethod("get"))
	assert.True(t, isSafeMethod("head"))
	assert.True(t, isSafeMethod("options"))
	assert.True(t, isSafeMethod("trace"))
	assert.False(t, isSafeMethod("post"))
	assert.False(t, isSafeMethod("put"))
	assert.False(t, isSafeMethod("delete"))
}

func TestIsSafeMethod_MixedCase(t *testing.T) {
	assert.True(t, isSafeMethod("Get"))
	assert.True(t, isSafeMethod("HEAD"))
	assert.True(t, isSafeMethod("Options"))
	assert.False(t, isSafeMethod("Post"))
	assert.False(t, isSafeMethod("PUT"))
}

func TestIsSafeMethod_UnknownMethod(t *testing.T) {
	assert.False(t, isSafeMethod("CUSTOM"))
	assert.False(t, isSafeMethod("UNKNOWN"))
}

// ============================================
// GenerateCSRFToken Edge Cases
// ============================================

func TestGenerateCSRFToken_NegativeLength(t *testing.T) {
	token := GenerateCSRFToken(-1)
	assert.NotEmpty(t, token)
	// Should default to 32
	assert.True(t, len(token) > 32)
}

func TestGenerateCSRFToken_LargeLength(t *testing.T) {
	token := GenerateCSRFToken(64)
	assert.NotEmpty(t, token)
	// Base64 encoding will make it longer
	assert.True(t, len(token) >= 64)
}
