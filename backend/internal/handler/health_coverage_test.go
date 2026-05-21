package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newSQLMockWithPings creates sqlmock with ping monitoring enabled
func newSQLMockWithPings(t *testing.T) (sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return mock, db
}

// ============================================
// checkDatabaseHealth Full Tests
// ============================================

func TestCheckDatabaseHealth_PingSuccess(t *testing.T) {
	mock, db := newSQLMockWithPings(t)
	mock.ExpectPing()

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseHealth(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
	assert.NotNil(t, result.Details)
	assert.Contains(t, result.Details, "open_connections")
	assert.Contains(t, result.Details, "in_use")
	assert.Contains(t, result.Details, "idle")
}

func TestCheckDatabaseHealth_PingError(t *testing.T) {
	mock, db := newSQLMockWithPings(t)
	mock.ExpectPing().WillReturnError(errors.New("connection refused"))

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseHealth(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "connection refused")
}

func TestCheckDatabaseHealth_NilDB(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseHealth(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Error, "not initialized")
}

func TestCheckDatabaseHealth_LatencyMeasurement(t *testing.T) {
	mock, db := newSQLMockWithPings(t)
	mock.ExpectPing().WillDelayFor(10 * time.Millisecond)

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseHealth(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
}

func TestCheckDatabaseHealth_ConnectionPoolStats(t *testing.T) {
	mock, db := newSQLMockWithPings(t)
	mock.ExpectPing()

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseHealth(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.NotNil(t, result.Details)
	details := result.Details
	assert.Contains(t, details, "open_connections")
	assert.Contains(t, details, "in_use")
	assert.Contains(t, details, "idle")
	assert.Contains(t, details, "wait_count")
	assert.Contains(t, details, "wait_duration_ms")
}

// ============================================
// checkPostgreSQLDependency Full Tests
// ============================================

func TestCheckPostgreSQLDependency_VersionSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	versionRows := sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 14.2 on x86_64-pc-linux-gnu, compiled by gcc 8.5.0 64-bit")
	mock.ExpectQuery("SELECT version()").WillReturnRows(versionRows)
	mock.ExpectQuery("SELECT ssl FROM pg_stat_ssl").WillReturnRows(sqlmock.NewRows([]string{"ssl"}).AddRow(true))

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkPostgreSQLDependency(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
	assert.NotNil(t, result.Details)
	assert.Contains(t, result.Details, "version")
	assert.Contains(t, result.Details, "ssl")
}

func TestCheckPostgreSQLDependency_VersionQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT version()").WillReturnError(errors.New("query failed"))

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkPostgreSQLDependency(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "query failed")
}

func TestCheckPostgreSQLDependency_NilDB(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	ctx := context.Background()

	result := handler.checkPostgreSQLDependency(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Error, "not initialized")
}

func TestCheckPostgreSQLDependency_VersionTruncation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	longVersion := "PostgreSQL 14.2 on x86_64-pc-linux-gnu, compiled by gcc (GCC) 8.5.0 20210514 (Red Hat 8.5.0-4), 64-bit"
	mock.ExpectQuery("SELECT version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(longVersion))
	mock.ExpectQuery("SELECT ssl").WillReturnRows(sqlmock.NewRows([]string{"ssl"}))

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkPostgreSQLDependency(ctx)

	assert.Equal(t, "healthy", result.Status)
	version := result.Details["version"].(string)
	assert.LessOrEqual(t, len(version), 50)
}

func TestCheckPostgreSQLDependency_SSLStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 14.2 on x86_64-pc-linux-gnu, compiled by gcc 8.5.0 64-bit"))
	mock.ExpectQuery("SELECT ssl").WillReturnRows(sqlmock.NewRows([]string{"ssl"}).AddRow(true))

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkPostgreSQLDependency(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, true, result.Details["ssl"])
}

func TestCheckPostgreSQLDependency_SSLQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 14.2 on x86_64-pc-linux-gnu, compiled by gcc 8.5.0 64-bit"))
	mock.ExpectQuery("SELECT ssl").WillReturnError(errors.New("ssl query failed"))

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkPostgreSQLDependency(ctx)

	// Should still be healthy (SSL check is optional)
	assert.Equal(t, "healthy", result.Status)
	assert.NotNil(t, result.Details)
}

func TestCheckPostgreSQLDependency_ContextTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT version()").WillDelayFor(2 * time.Second)

	handler := NewHealthHandler(db, nil, "test-version")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result := handler.checkPostgreSQLDependency(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.NotEmpty(t, result.Error)
}

// ============================================
// checkRedisHealth Full Tests
// ============================================

func TestCheckRedisHealth_WithInfo(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(nil, redisClient, "test-version")
	ctx := context.Background()

	result := handler.checkRedisHealth(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
}

func TestCheckRedisHealth_NilRedis(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	ctx := context.Background()

	result := handler.checkRedisHealth(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Error, "not initialized")
}

func TestCheckRedisHealth_ConnectionError(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "invalid-redis-address:6379",
	})

	handler := NewHealthHandler(nil, redisClient, "test-version")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := handler.checkRedisHealth(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.NotEmpty(t, result.Error)
}

// ============================================
// checkRedisDependency Full Tests
// ============================================

func TestCheckRedisDependency_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(nil, redisClient, "test-version")
	ctx := context.Background()

	result := handler.checkRedisDependency(ctx)

	assert.Equal(t, "healthy", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, int64(0))
}

func TestCheckRedisDependency_NilRedis(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	ctx := context.Background()

	result := handler.checkRedisDependency(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.Contains(t, result.Error, "not initialized")
}

func TestCheckRedisDependency_ConnectionError(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{Addr: "invalid:6379"})

	handler := NewHealthHandler(nil, redisClient, "test-version")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := handler.checkRedisDependency(ctx)

	assert.Equal(t, "unhealthy", result.Status)
	assert.NotEmpty(t, result.Error)
}

// ============================================
// DependenciesCheck Full Tests
// ============================================

func TestDependenciesCheck_AllHealthy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mock.ExpectQuery("SELECT version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 14.2 on x86_64-pc-linux-gnu, compiled by gcc 8.5.0 64-bit"))
	mock.ExpectQuery("SELECT ssl").WillReturnRows(sqlmock.NewRows([]string{"ssl"}))

	handler := NewHealthHandler(db, redisClient, "test-version")

	router := gin.New()
	router.GET("/health/dependencies", handler.DependenciesCheck)

	req := httptest.NewRequest("GET", "/health/dependencies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestDependenciesCheck_DatabaseUnhealthy(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(nil, redisClient, "test-version")

	router := gin.New()
	router.GET("/health/dependencies", handler.DependenciesCheck)

	req := httptest.NewRequest("GET", "/health/dependencies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "degraded")
}

func TestDependenciesCheck_AllNil(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	router := gin.New()
	router.GET("/health/dependencies", handler.DependenciesCheck)

	req := httptest.NewRequest("GET", "/health/dependencies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "degraded")
}

// ============================================
// DetailedHealthCheck Full Tests
// ============================================

func TestDetailedHealthCheck_WithDatabase(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	handler := NewHealthHandler(db, nil, "test-version")

	router := gin.New()
	router.GET("/health", handler.DetailedHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "database")
}

func TestDetailedHealthCheck_WithRedis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(nil, redisClient, "test-version")

	router := gin.New()
	router.GET("/health", handler.DetailedHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "redis")
}

func TestDetailedHealthCheck_BothDeps(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(db, redisClient, "test-version")

	router := gin.New()
	router.GET("/health", handler.DetailedHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

// ============================================
// ReadinessCheck Full Tests
// ============================================

func TestReadinessCheck_WithDatabase(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	handler := NewHealthHandler(db, nil, "test-version")

	router := gin.New()
	router.GET("/health/ready", handler.ReadinessCheck)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 503, w.Code)
	assert.Contains(t, w.Body.String(), "not_ready")
}

func TestReadinessCheck_AllDeps(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(db, redisClient, "test-version")

	router := gin.New()
	router.GET("/health/ready", handler.ReadinessCheck)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ready")
}

// ============================================
// StartupCheck Full Tests
// ============================================

func TestStartupCheck_WithDatabase(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	handler := NewHealthHandler(db, nil, "test-version")

	router := gin.New()
	router.GET("/health/startup", handler.StartupCheck)

	req := httptest.NewRequest("GET", "/health/startup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 503, w.Code)
	assert.Contains(t, w.Body.String(), "starting")
}

func TestStartupCheck_AllDeps(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(db, redisClient, "test-version")

	router := gin.New()
	router.GET("/health/startup", handler.StartupCheck)

	req := httptest.NewRequest("GET", "/health/startup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "started")
}

// ============================================
// checkDatabaseReady Tests
// ============================================

func TestCheckDatabaseReady_PingSuccess(t *testing.T) {
	mock2, db := newSQLMockWithPings(t)
	mock2.ExpectPing()

	handler := NewHealthHandler(db, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseReady(ctx)
	assert.True(t, result)
}

func TestCheckDatabaseReady_NilDB(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	ctx := context.Background()

	result := handler.checkDatabaseReady(ctx)
	assert.False(t, result)
}

// ============================================
// checkRedisReady Tests
// ============================================

func TestCheckRedisReady_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	handler := NewHealthHandler(nil, redisClient, "test-version")
	ctx := context.Background()

	result := handler.checkRedisReady(ctx)
	assert.True(t, result)
}

func TestCheckRedisReady_NilRedis(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	ctx := context.Background()

	result := handler.checkRedisReady(ctx)
	assert.False(t, result)
}

// ============================================
// Additional Coverage Tests
// ============================================

func TestCheckDiskHealth_VerifyDetails(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	result := handler.checkDiskHealth()

	assert.Equal(t, "healthy", result.Status)
	assert.Contains(t, result.Details, "check_available")
	assert.Contains(t, result.Details, "message")
	assert.Equal(t, false, result.Details["check_available"])
}

func TestCheckSystemHealth_VerifyDetails(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")
	result := handler.checkSystemHealth()

	assert.Equal(t, "healthy", result.Status)
	goVersion := result.Details["go_version"].(string)
	assert.NotEmpty(t, goVersion)
	goroutines := result.Details["goroutines"].(int)
	assert.Greater(t, goroutines, 0)
}

func TestHealthCheckerInterface_Full(t *testing.T) {
	customChecker := &MockHealthChecker{
		name: "custom_full",
		result: HealthCheckResult{
			Status:    "healthy",
			LatencyMS: 10,
			Details:   map[string]interface{}{"custom": true},
		},
	}

	assert.Equal(t, "custom_full", customChecker.Name())
	result := customChecker.Check(context.Background())
	assert.Equal(t, "healthy", result.Status)
}

func TestHealthCheck_ConcurrentRequests(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "test-version")

	router := gin.New()
	router.GET("/health/live", handler.LivenessCheck)

	for i := 0; i < 20; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health/live", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}()
	}
}
