package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 空闲连接超时
}

// DefaultProductionPoolConfig 生产环境连接池配置
func DefaultProductionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxOpenConns:    50,               // 防止连接耗尽
		MaxIdleConns:    10,               // 保持一定空闲连接
		ConnMaxLifetime: 1 * time.Hour,    // 连接最大 1 小时
		ConnMaxIdleTime: 10 * time.Minute, // 空闲 10 分钟关闭
	}
}

// DefaultDevelopmentPoolConfig 开发环境连接池配置
func DefaultDevelopmentPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// ConfigureConnectionPool 配置连接池参数
func ConfigureConnectionPool(db *sql.DB, config *ConnectionPoolConfig) {
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
		log.Printf("Connection pool: MaxOpenConns=%d", config.MaxOpenConns)
	}

	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
		log.Printf("Connection pool: MaxIdleConns=%d", config.MaxIdleConns)
	}

	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
		log.Printf("Connection pool: ConnMaxLifetime=%v", config.ConnMaxLifetime)
	}

	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
		log.Printf("Connection pool: ConnMaxIdleTime=%v", config.ConnMaxIdleTime)
	}
}

// ConnectionPoolStats 连接池状态
type ConnectionPoolStats struct {
	MaxOpenConnections int   `json:"max_open_connections"`
	OpenConnections    int   `json:"open_connections"`
	InUse              int   `json:"in_use"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"wait_count"`           // 等待连接的次数
	WaitDuration       int64 `json:"wait_duration_ms"`     // 等待总时间 (ms)
	MaxIdleClosed      int64 `json:"max_idle_closed"`      // 因超过 MaxIdleConns 关闭的连接数
	MaxLifetimeClosed  int64 `json:"max_lifetime_closed"`  // 因超过 ConnMaxLifetime 关闭的连接数
	MaxIdleTimeClosed  int64 `json:"max_idle_time_closed"` // 因超过 ConnMaxIdleTime 关闭的连接数
}

// GetConnectionPoolStats 获取连接池状态
func GetConnectionPoolStats(db *sql.DB) *ConnectionPoolStats {
	stats := db.Stats()
	return &ConnectionPoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration.Milliseconds(),
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
		MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
	}
}

// IsConnectionPoolHealthy 检查连接池是否健康
func IsConnectionPoolHealthy(db *sql.DB) (bool, []string) {
	stats := GetConnectionPoolStats(db)
	warnings := []string{}

	// 1. 检查连接池是否接近耗尽
	usagePercent := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
	if usagePercent > 80 {
		warnings = append(warnings,
			fmt.Sprintf("Connection pool usage high: %.1f%% (%d/%d)",
				usagePercent, stats.InUse, stats.MaxOpenConnections))
	}

	// 2. 检查等待次数是否过多
	if stats.WaitCount > 1000 {
		warnings = append(warnings,
			fmt.Sprintf("High wait count: %d connections had to wait", stats.WaitCount))
	}

	// 3. 检查等待时间是否过长
	if stats.WaitDuration > 5000 {
		warnings = append(warnings,
			fmt.Sprintf("Long wait duration: %dms total", stats.WaitDuration))
	}

	// 4. 检查是否有空闲连接
	if stats.Idle == 0 && stats.OpenConnections > 0 {
		warnings = append(warnings,
			"No idle connections available")
	}

	healthy := len(warnings) == 0
	return healthy, warnings
}

// QueryPerformanceStats 查询性能统计
type QueryPerformanceStats struct {
	SlowQueryCount    int64   `json:"slow_query_count"`   // 慢查询数量
	AvgQueryTime      float64 `json:"avg_query_time_ms"`  // 平均查询时间
	P95QueryTime      float64 `json:"p95_query_time_ms"`  // P95 查询时间
	CacheHitRate      float64 `json:"cache_hit_rate"`     // 缓存命中率
	TotalQueries      int64   `json:"total_queries"`      // 总查询数
	ActiveConnections int     `json:"active_connections"` // 活动连接数
}

// GetQueryPerformanceStats 获取查询性能统计 (使用 pg_stat_statements)
func GetQueryPerformanceStats(db *sql.DB) (*QueryPerformanceStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := &QueryPerformanceStats{}

	// 获取活动连接数
	err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM pg_stat_activity WHERE state != 'idle'
	`).Scan(&stats.ActiveConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}

	// 使用 pg_stat_statements 获取查询统计 (需要扩展已启用)
	err = db.QueryRowContext(ctx, `
		SELECT 
			count(*) as total_queries,
			avg(total_time/calls) as avg_time,
			percentile_cont(0.95) within group (order by total_time/calls) as p95_time
		FROM pg_stat_statements
	`).Scan(&stats.TotalQueries, &stats.AvgQueryTime, &stats.P95QueryTime)

	// 如果 pg_stat_statements 未启用，忽略错误
	if err != nil {
		log.Printf("pg_stat_statements not available: %v", err)
	}

	return stats, nil
}

// IndexUsageStats 索引使用统计
type IndexUsageStats struct {
	IndexName   string `json:"index_name"`
	TableName   string `json:"table_name"`
	ScanCount   int64  `json:"scan_count"`
	TuplesRead  int64  `json:"tuples_read"`
	TuplesFetch int64  `json:"tuples_fetch"`
	SizeBytes   int64  `json:"size_bytes"`
}

// GetUnusedIndexes 获取未使用的索引
func GetUnusedIndexes(db *sql.DB) ([]IndexUsageStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT 
			indexrelname as index_name,
			relname as table_name,
			idx_scan as scan_count,
			idx_tup_read as tuples_read,
			idx_tup_fetch as tuples_fetch,
			pg_relation_size(indexrelid) as size_bytes
		FROM pg_stat_user_indexes
		WHERE idx_scan = 0
		AND indexrelname NOT LIKE '%_pkey'
		ORDER BY size_bytes DESC
		LIMIT 20
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unused indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexUsageStats
	for rows.Next() {
		var idx IndexUsageStats
		if err := rows.Scan(&idx.IndexName, &idx.TableName, &idx.ScanCount,
			&idx.TuplesRead, &idx.TuplesFetch, &idx.SizeBytes); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		indexes = append(indexes, idx)
	}

	return indexes, rows.Err()
}

// TableSizeStats 表大小统计
type TableSizeStats struct {
	TableName string `json:"table_name"`
	TotalSize string `json:"total_size"`
	TableSize string `json:"table_size"`
	IndexSize string `json:"index_size"`
	RowCount  int64  `json:"row_count"`
}

// GetTableSizes 获取表大小统计
func GetTableSizes(db *sql.DB) ([]TableSizeStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT 
			table_name,
			pg_size_pretty(pg_total_relation_size(table_name::regclass)) as total_size,
			pg_size_pretty(pg_relation_size(table_name::regclass)) as table_size,
			pg_size_pretty(pg_indexes_size(table_name::regclass)) as index_size,
			(SELECT count(*) FROM information_schema.columns WHERE table_name = t.table_name) as columns
		FROM information_schema.tables t
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		ORDER BY pg_total_relation_size(table_name::regclass) DESC
		LIMIT 20
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table sizes: %w", err)
	}
	defer rows.Close()

	var tables []TableSizeStats
	for rows.Next() {
		var tbl TableSizeStats
		if err := rows.Scan(&tbl.TableName, &tbl.TotalSize, &tbl.TableSize,
			&tbl.IndexSize, &tbl.RowCount); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, tbl)
	}

	return tables, rows.Err()
}
