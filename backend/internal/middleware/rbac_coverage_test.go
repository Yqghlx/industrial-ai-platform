package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/internal/service"
	"github.com/industrial-ai/platform/pkg/database"
)

// ============================================
// Test Helpers
// ============================================

// newTestRBACService creates a RBACService with mocked repositories for testing
func newTestRBACService(t *testing.T) (*service.RBACService, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rbacRepo := repository.NewRBACRepository(database.NewDBWrapper(db))
	userRepo := repository.NewUserRepository(database.NewDBWrapper(db))
	tenantRepo := repository.NewTenantRepo(database.NewDBWrapper(db))
	svc := service.NewRBACServiceWithRBACRepo(rbacRepo, userRepo, tenantRepo)

	t.Cleanup(func() {
		db.Close()
	})

	return svc, mock, db
}

// ============================================
// PermissionRequired Middleware Tests
// ============================================

func TestPermissionRequired_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(PermissionRequired(svc, "devices", "read"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_AUTHENTICATED")
}

func TestPermissionRequired_InvalidUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "invalid-type") // Set invalid type (string instead of int)
		c.Next()
	})
	router.Use(PermissionRequired(svc, "devices", "read"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_USER_ID")
}

func TestPermissionRequired_CheckPermissionError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnError(sql.ErrConnDone)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(PermissionRequired(svc, "devices", "read"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERMISSION_CHECK_ERROR")
}

func TestPermissionRequired_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "delete").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(PermissionRequired(svc, "devices", "delete"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_PERMISSIONS")
	assert.Contains(t, w.Body.String(), "devices")
	assert.Contains(t, w.Body.String(), "delete")
}

func TestPermissionRequired_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(PermissionRequired(svc, "devices", "read"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestPermissionRequired_WithDifferentResources(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		resource   string
		action     string
		userID     int
		count      int
		expectCode int
	}{
		{"devices read granted", "devices", "read", 1, 1, http.StatusOK},
		{"devices write granted", "devices", "write", 2, 1, http.StatusOK},
		{"users manage denied", "users", "manage", 3, 0, http.StatusForbidden},
		{"reports read granted", "reports", "read", 4, 1, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mock, _ := newTestRBACService(t)

			rows := sqlmock.NewRows([]string{"count"}).AddRow(tt.count)
			mock.ExpectQuery(`SELECT COUNT`).
				WithArgs(tt.userID, tt.resource, tt.action).
				WillReturnRows(rows)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})
			router.Use(PermissionRequired(svc, tt.resource, tt.action))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

// ============================================
// CanManageDevices Tests
// ============================================

func TestCanManageDevices_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageDevices(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageDevices_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageDevices(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageDevices_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageDevices(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanManageUsers Tests
// ============================================

func TestCanManageUsers_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "users", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageUsers(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageUsers_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "users", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageUsers(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageUsers_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageUsers(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanManageRules Tests
// ============================================

func TestCanManageRules_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "rules", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageRules(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageRules_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "rules", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageRules(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageRules_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageRules(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanViewReports Tests
// ============================================

func TestCanViewReports_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "reports", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanViewReports(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanViewReports_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "reports", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanViewReports(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanViewReports_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanViewReports(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanGenerateReports Tests
// ============================================

func TestCanGenerateReports_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "reports", "generate").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanGenerateReports(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanGenerateReports_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "reports", "generate").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanGenerateReports(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanGenerateReports_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanGenerateReports(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanManageAlerts Tests
// ============================================

func TestCanManageAlerts_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "alerts", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageAlerts(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageAlerts_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "alerts", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageAlerts(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageAlerts_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageAlerts(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanViewAlerts Tests
// ============================================

func TestCanViewAlerts_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "alerts", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanViewAlerts(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanViewAlerts_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "alerts", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanViewAlerts(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanViewAlerts_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanViewAlerts(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanManageWorkOrders Tests
// ============================================

func TestCanManageWorkOrders_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "workorders", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageWorkOrders(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageWorkOrders_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "workorders", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageWorkOrders(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageWorkOrders_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageWorkOrders(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanViewWorkOrders Tests
// ============================================

func TestCanViewWorkOrders_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "workorders", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanViewWorkOrders(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanViewWorkOrders_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "workorders", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanViewWorkOrders(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanViewWorkOrders_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanViewWorkOrders(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanManageSettings Tests
// ============================================

func TestCanManageSettings_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "settings", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageSettings(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageSettings_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "settings", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageSettings(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageSettings_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageSettings(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// CanManageTenants Tests
// ============================================

func TestCanManageTenants_PermissionGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "tenants", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageTenants(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanManageTenants_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "tenants", "manage").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(CanManageTenants(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanManageTenants_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	router := gin.New()
	router.Use(CanManageTenants(svc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================
// RoleRequired Middleware Tests
// ============================================

func TestRoleRequired_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RoleRequired("admin", "operator"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_AUTHENTICATED")
}

func TestRoleRequired_HasRequiredRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(RoleRequired("admin", "operator"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleRequired_HasSecondRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "operator")
		c.Next()
	})
	router.Use(RoleRequired("admin", "operator"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleRequired_DoesNotHaveRequiredRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "viewer")
		c.Next()
	})
	router.Use(RoleRequired("admin", "operator"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_ROLE")
	assert.Contains(t, w.Body.String(), "admin")
	assert.Contains(t, w.Body.String(), "operator")
}

func TestRoleRequired_SingleRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(RoleRequired("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleRequired_SingleRoleDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "viewer")
		c.Next()
	})
	router.Use(RoleRequired("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_ROLE")
}

// ============================================
// AnyPermissionRequired Middleware Tests
// ============================================

func TestAnyPermissionRequired_UserNotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	router := gin.New()
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_AUTHENTICATED")
}

func TestAnyPermissionRequired_InvalidUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "invalid-type")
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_USER_ID")
}

func TestAnyPermissionRequired_HasFirstPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnyPermissionRequired_HasSecondPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	// First permission denied
	rows1 := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnRows(rows1)

	// Second permission granted
	rows2 := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "write").
		WillReturnRows(rows2)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnyPermissionRequired_HasNoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	// Both permissions denied
	rows1 := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnRows(rows1)

	rows2 := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "write").
		WillReturnRows(rows2)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_PERMISSIONS")
}

func TestAnyPermissionRequired_FirstPermissionHasError_ContinueToNext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	// First permission returns error (should continue to check next)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnError(sql.ErrConnDone)

	// Second permission granted
	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "write").
		WillReturnRows(rows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnyPermissionRequired_AllPermissionsHaveErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
	}

	// All permissions return errors
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnError(sql.ErrConnDone)

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "write").
		WillReturnError(sql.ErrConnDone)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_PERMISSIONS")
}

func TestAnyPermissionRequired_ThreePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, mock, _ := newTestRBACService(t)

	permissions := []struct{ Resource, Action string }{
		{"devices", "read"},
		{"devices", "write"},
		{"devices", "manage"},
	}

	// First two denied
	rows1 := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "read").
		WillReturnRows(rows1)

	rows2 := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "write").
		WillReturnRows(rows2)

	// Third granted
	rows3 := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(1, "devices", "manage").
		WillReturnRows(rows3)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	})
	router.Use(AnyPermissionRequired(svc, permissions...))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
