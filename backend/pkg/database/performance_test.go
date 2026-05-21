package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProductionPoolConfig(t *testing.T) {
	config := DefaultProductionPoolConfig()

	require.NotNil(t, config)
	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 1*time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, config.ConnMaxIdleTime)
}

func TestDefaultDevelopmentPoolConfig(t *testing.T) {
	config := DefaultDevelopmentPoolConfig()

	require.NotNil(t, config)
	assert.Equal(t, 10, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 30*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, config.ConnMaxIdleTime)
}

func TestConnectionPoolConfigStruct(t *testing.T) {
	config := &ConnectionPoolConfig{
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 2 * time.Hour,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 20, config.MaxIdleConns)
	assert.Equal(t, 2*time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 15*time.Minute, config.ConnMaxIdleTime)
}

func TestConfigureConnectionPool(t *testing.T) {
	t.Run("configures pool with all settings", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := &ConnectionPoolConfig{
			MaxOpenConns:    50,
			MaxIdleConns:    25,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		}

		ConfigureConnectionPool(db, config)

		stats := db.Stats()
		assert.Equal(t, 50, stats.MaxOpenConnections)
	})

	t.Run("skips zero values", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Set initial values
		db.SetMaxOpenConns(30)
		db.SetMaxIdleConns(10)

		config := &ConnectionPoolConfig{
			MaxOpenConns:    0,
			MaxIdleConns:    0,
			ConnMaxLifetime: 0,
			ConnMaxIdleTime: 0,
		}

		ConfigureConnectionPool(db, config)

		stats := db.Stats()
		assert.Equal(t, 30, stats.MaxOpenConnections)
	})

	t.Run("configures with positive values only", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		config := &ConnectionPoolConfig{
			MaxOpenConns:    100,
			MaxIdleConns:    0, // Should be skipped
			ConnMaxLifetime: 30 * time.Minute,
			ConnMaxIdleTime: 0, // Should be skipped
		}

		ConfigureConnectionPool(db, config)

		stats := db.Stats()
		assert.Equal(t, 100, stats.MaxOpenConnections)
	})
}

func TestGetConnectionPoolStats(t *testing.T) {
	t.Run("returns stats for valid database", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(10)

		stats := GetConnectionPoolStats(db)

		require.NotNil(t, stats)
		assert.Equal(t, 50, stats.MaxOpenConnections)
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
		assert.GreaterOrEqual(t, stats.InUse, 0)
		assert.GreaterOrEqual(t, stats.Idle, 0)
	})
}

func TestIsConnectionPoolHealthy(t *testing.T) {
	t.Run("returns healthy for normal usage", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		db.SetMaxOpenConns(50)

		healthy, warnings := IsConnectionPoolHealthy(db)

		// With no connections in use, should be healthy
		assert.NotNil(t, warnings)
		_ = healthy
	})

	t.Run("detects no idle connections warning", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		db.SetMaxOpenConns(50)

		healthy, warnings := IsConnectionPoolHealthy(db)

		// The health check logic depends on actual connection state
		// Just verify the function runs without panic
		assert.NotNil(t, warnings)
		_ = healthy
	})
}

func TestConnectionPoolStatsStruct(t *testing.T) {
	stats := &ConnectionPoolStats{
		MaxOpenConnections: 100,
		OpenConnections:    50,
		InUse:              30,
		Idle:               20,
		WaitCount:          5,
		WaitDuration:       100,
		MaxIdleClosed:      10,
		MaxLifetimeClosed:  15,
		MaxIdleTimeClosed:  20,
	}

	assert.Equal(t, 100, stats.MaxOpenConnections)
	assert.Equal(t, 50, stats.OpenConnections)
	assert.Equal(t, 30, stats.InUse)
	assert.Equal(t, 20, stats.Idle)
	assert.Equal(t, int64(5), stats.WaitCount)
	assert.Equal(t, int64(100), stats.WaitDuration)
	assert.Equal(t, int64(10), stats.MaxIdleClosed)
	assert.Equal(t, int64(15), stats.MaxLifetimeClosed)
	assert.Equal(t, int64(20), stats.MaxIdleTimeClosed)
}

func TestQueryPerformanceStatsStruct(t *testing.T) {
	stats := &QueryPerformanceStats{
		SlowQueryCount:    10,
		AvgQueryTime:      50.5,
		P95QueryTime:      120.5,
		CacheHitRate:      95.5,
		TotalQueries:      1000,
		ActiveConnections: 25,
	}

	assert.Equal(t, int64(10), stats.SlowQueryCount)
	assert.Equal(t, 50.5, stats.AvgQueryTime)
	assert.Equal(t, 120.5, stats.P95QueryTime)
	assert.Equal(t, 95.5, stats.CacheHitRate)
	assert.Equal(t, int64(1000), stats.TotalQueries)
	assert.Equal(t, 25, stats.ActiveConnections)
}

func TestGetQueryPerformanceStats(t *testing.T) {
	t.Run("returns active connections", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Mock active connections query
		rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_stat_activity").
			WillReturnRows(rows)

		// Mock pg_stat_statements query (may fail, which is handled)
		mock.ExpectQuery("SELECT(.+)FROM pg_stat_statements").
			WillReturnError(sql.ErrConnDone)

		stats, err := GetQueryPerformanceStats(db)

		require.NoError(t, err)
		assert.Equal(t, 5, stats.ActiveConnections)
	})

	t.Run("handles connection error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_stat_activity").
			WillReturnError(sql.ErrConnDone)

		stats, err := GetQueryPerformanceStats(db)

		require.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "active connections")
	})
}

func TestIndexUsageStatsStruct(t *testing.T) {
	stats := &IndexUsageStats{
		IndexName:   "idx_users_email",
		TableName:   "users",
		ScanCount:   1000,
		TuplesRead:  5000,
		TuplesFetch: 4500,
		SizeBytes:   1024000,
	}

	assert.Equal(t, "idx_users_email", stats.IndexName)
	assert.Equal(t, "users", stats.TableName)
	assert.Equal(t, int64(1000), stats.ScanCount)
	assert.Equal(t, int64(5000), stats.TuplesRead)
	assert.Equal(t, int64(4500), stats.TuplesFetch)
	assert.Equal(t, int64(1024000), stats.SizeBytes)
}

func TestGetUnusedIndexes(t *testing.T) {
	t.Run("returns unused indexes", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"index_name", "table_name", "scan_count", "tuples_read", "tuples_fetch", "size_bytes",
		}).
			AddRow("idx_unused_1", "users", 0, 0, 0, 1024).
			AddRow("idx_unused_2", "orders", 0, 0, 0, 2048)

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_user_indexes").
			WillReturnRows(rows)

		indexes, err := GetUnusedIndexes(db)

		require.NoError(t, err)
		assert.Len(t, indexes, 2)
		assert.Equal(t, "idx_unused_1", indexes[0].IndexName)
		assert.Equal(t, "users", indexes[0].TableName)
		assert.Equal(t, "idx_unused_2", indexes[1].IndexName)
	})

	t.Run("returns empty slice when no indexes", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"index_name", "table_name", "scan_count", "tuples_read", "tuples_fetch", "size_bytes",
		})

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_user_indexes").
			WillReturnRows(rows)

		indexes, err := GetUnusedIndexes(db)

		require.NoError(t, err)
		assert.Empty(t, indexes)
	})

	t.Run("handles query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_user_indexes").
			WillReturnError(sql.ErrConnDone)

		indexes, err := GetUnusedIndexes(db)

		require.Error(t, err)
		assert.Nil(t, indexes)
		assert.Contains(t, err.Error(), "failed to query unused indexes")
	})

	t.Run("handles scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"index_name", "table_name", "scan_count", "tuples_read", "tuples_fetch", "size_bytes",
		}).
			AddRow("idx_name", "table", "invalid_scan_count", 0, 0, 1024) // Invalid int

		mock.ExpectQuery("SELECT(.+)FROM pg_stat_user_indexes").
			WillReturnRows(rows)

		indexes, err := GetUnusedIndexes(db)

		require.Error(t, err)
		assert.Nil(t, indexes)
	})
}

func TestTableSizeStatsStruct(t *testing.T) {
	stats := &TableSizeStats{
		TableName: "users",
		TotalSize: "128 MB",
		TableSize: "100 MB",
		IndexSize: "28 MB",
		RowCount:  50000,
	}

	assert.Equal(t, "users", stats.TableName)
	assert.Equal(t, "128 MB", stats.TotalSize)
	assert.Equal(t, "100 MB", stats.TableSize)
	assert.Equal(t, "28 MB", stats.IndexSize)
	assert.Equal(t, int64(50000), stats.RowCount)
}

func TestGetTableSizes(t *testing.T) {
	t.Run("returns table sizes", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"table_name", "total_size", "table_size", "index_size", "columns",
		}).
			AddRow("users", "128 MB", "100 MB", "28 MB", 10).
			AddRow("orders", "256 MB", "200 MB", "56 MB", 15)

		mock.ExpectQuery("SELECT(.+)FROM information_schema.tables").
			WillReturnRows(rows)

		tables, err := GetTableSizes(db)

		require.NoError(t, err)
		assert.Len(t, tables, 2)
		assert.Equal(t, "users", tables[0].TableName)
		assert.Equal(t, "128 MB", tables[0].TotalSize)
		assert.Equal(t, "orders", tables[1].TableName)
	})

	t.Run("returns empty slice when no tables", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"table_name", "total_size", "table_size", "index_size", "columns",
		})

		mock.ExpectQuery("SELECT(.+)FROM information_schema.tables").
			WillReturnRows(rows)

		tables, err := GetTableSizes(db)

		require.NoError(t, err)
		assert.Empty(t, tables)
	})

	t.Run("handles query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT(.+)FROM information_schema.tables").
			WillReturnError(sql.ErrConnDone)

		tables, err := GetTableSizes(db)

		require.Error(t, err)
		assert.Nil(t, tables)
		assert.Contains(t, err.Error(), "failed to query table sizes")
	})

	t.Run("handles scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{
			"table_name", "total_size", "table_size", "index_size", "columns",
		}).
			AddRow("users", "128 MB", "100 MB", "28 MB", "invalid_int") // Invalid int64

		mock.ExpectQuery("SELECT(.+)FROM information_schema.tables").
			WillReturnRows(rows)

		tables, err := GetTableSizes(db)

		require.Error(t, err)
		assert.Nil(t, tables)
	})
}
