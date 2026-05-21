package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================
// TenantAdminRequired Tests
// ============================================

func TestTenantAdminRequired_NotAuthenticated(t *testing.T) {
	router := gin.New()
	router.Use(TenantAdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_AUTHENTICATED")
}

func TestTenantAdminRequired_AdminRole(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock setting user role
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(TenantAdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTenantAdminRequired_TenantAdminRole(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "tenant_admin")
		c.Next()
	})
	router.Use(TenantAdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTenantAdminRequired_InsufficientPermissions(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "user")
		c.Next()
	})
	router.Use(TenantAdminRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_PERMISSIONS")
}

func TestTenantAdminRequired_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "empty role",
			role:       "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "NOT_AUTHENTICATED",
		},
		{
			name:       "admin role",
			role:       "admin",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:       "tenant_admin role",
			role:       "tenant_admin",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:       "user role",
			role:       "user",
			wantStatus: http.StatusForbidden,
			wantBody:   "INSUFFICIENT_PERMISSIONS",
		},
		{
			name:       "viewer role",
			role:       "viewer",
			wantStatus: http.StatusForbidden,
			wantBody:   "INSUFFICIENT_PERMISSIONS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			if tt.role != "" {
				router.Use(func(c *gin.Context) {
					c.Set("user_role", tt.role)
					c.Next()
				})
			}
			router.Use(TenantAdminRequired())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ============================================
// BuildTenantFilter Tests
// ============================================

func TestBuildTenantFilter_EmptyTenantID(t *testing.T) {
	filter := BuildTenantFilter("", "SELECT * FROM devices")

	assert.Equal(t, "SELECT * FROM devices", filter.WhereClause)
	assert.Empty(t, filter.Args)
}

func TestBuildTenantFilter_WithTenantID(t *testing.T) {
	filter := BuildTenantFilter("tenant-123", "SELECT * FROM devices WHERE active = true")

	assert.Equal(t, "SELECT * FROM devices WHERE active = true AND tenant_id = ?", filter.WhereClause)
	assert.Len(t, filter.Args, 1)
	assert.Equal(t, "tenant-123", filter.Args[0])
}

func TestBuildTenantFilter_WithAdditionalArgs(t *testing.T) {
	filter := BuildTenantFilter("tenant-456", "SELECT * FROM devices WHERE status = ?", "active")

	assert.Contains(t, filter.WhereClause, "tenant_id = ?")
	assert.Len(t, filter.Args, 2)
	assert.Equal(t, "active", filter.Args[0])
	assert.Equal(t, "tenant-456", filter.Args[1])
}

func TestBuildTenantFilter_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		tenantID    string
		baseQuery   string
		args        []interface{}
		wantClause  string
		wantArgsLen int
	}{
		{
			name:        "empty tenant",
			tenantID:    "",
			baseQuery:   "SELECT * FROM users",
			args:        nil,
			wantClause:  "SELECT * FROM users",
			wantArgsLen: 0,
		},
		{
			name:        "with tenant",
			tenantID:    "t1",
			baseQuery:   "SELECT * FROM users",
			args:        nil,
			wantClause:  "SELECT * FROM users AND tenant_id = ?",
			wantArgsLen: 1,
		},
		{
			name:        "with tenant and args",
			tenantID:    "t2",
			baseQuery:   "SELECT * FROM users WHERE age > ?",
			args:        []interface{}{18},
			wantClause:  "SELECT * FROM users WHERE age > ? AND tenant_id = ?",
			wantArgsLen: 2,
		},
		{
			name:        "multiple args",
			tenantID:    "t3",
			baseQuery:   "SELECT * FROM orders WHERE status = ? AND created > ?",
			args:        []interface{}{"pending", "2024-01-01"},
			wantClause:  "SELECT * FROM orders WHERE status = ? AND created > ? AND tenant_id = ?",
			wantArgsLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := BuildTenantFilter(tt.tenantID, tt.baseQuery, tt.args...)

			assert.Equal(t, tt.wantClause, filter.WhereClause)
			assert.Len(t, filter.Args, tt.wantArgsLen)
		})
	}
}

// ============================================
// GetUserRole Tests
// ============================================

func TestGetUserRole_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	role := GetUserRole(c)
	assert.Empty(t, role)
}

func TestGetUserRole_Set(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("user_role", "admin")

	role := GetUserRole(c)
	assert.Equal(t, "admin", role)
}
