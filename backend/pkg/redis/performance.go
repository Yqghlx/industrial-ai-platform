package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 客户端配置
type RedisConfig struct {
	Addr         string        // Redis 地址 (host:port)
	Password     string        // Redis 密码
	DB           int           // Redis 数据库编号
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	MaxRetries   int           // 最大重试次数
	DialTimeout  time.Duration // 连接超时
	ReadTimeout  time.Duration // 读超时
	WriteTimeout time.Duration // 写超时
	PoolTimeout  time.Duration // 连接池超时
	IdleTimeout  time.Duration // 空闲连接超时
	MaxConnAge   time.Duration // 连接最大生命周期
}

// DefaultProductionRedisConfig 生产环境 Redis 配置
func DefaultProductionRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:         "redis:6379",
		Password:     "", // 从环境变量加载
		DB:           0,
		PoolSize:     50,               // 连接池大小
		MinIdleConns: 10,               // 最小空闲连接
		MaxRetries:   3,                // 重试次数
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读超时
		WriteTimeout: 3 * time.Second,  // 写超时
		PoolTimeout:  4 * time.Second,  // 连接池等待超时
		IdleTimeout:  5 * time.Minute,  // 空闲超时
		MaxConnAge:   30 * time.Minute, // 连接最大生命周期
	}
}

// DefaultDevelopmentRedisConfig 开发环境 Redis 配置
func DefaultDevelopmentRedisConfig() *RedisConfig {
	// P0-01: 从环境变量读取 Redis 地址，避免硬编码
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	return &RedisConfig{
		Addr:         redisAddr,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolTimeout:  2 * time.Second,
		IdleTimeout:  2 * time.Minute,
		MaxConnAge:   10 * time.Minute,
	}
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(config *RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})
}

// RedisHealthCheck Redis 健康检查
type RedisHealthCheck struct {
	Connected          bool    `json:"connected"`
	Version            string  `json:"version"`
	UptimeSeconds      int64   `json:"uptime_seconds"`
	ConnectedClients   int     `json:"connected_clients"`
	UsedMemory         string  `json:"used_memory"`
	UsedMemoryPeak     string  `json:"used_memory_peak"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	HitRate            float64 `json:"hit_rate"` // 缓存命中率
	TotalHits          int64   `json:"total_hits"`
	TotalMisses        int64   `json:"total_misses"`
	OpsPerSec          int64   `json:"ops_per_sec"`
	SlowlogLen         int64   `json:"slowlog_len"`
	KeysCount          int64   `json:"keys_count"`
}

// GetHealthCheck 获取 Redis 健康状态
func GetHealthCheck(client *redis.Client) (*RedisHealthCheck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := &RedisHealthCheck{}

	// 测试连接
	err := client.Ping(ctx).Err()
	health.Connected = err == nil
	if err != nil {
		return health, fmt.Errorf("redis ping failed: %w", err)
	}

	// 获取 INFO
	info, err := client.Info(ctx, "server", "memory", "stats", "clients").Result()
	if err != nil {
		return health, fmt.Errorf("redis info failed: %w", err)
	}

	// 解析健康检查信息
	parseHealthCheckInfo(info, health)

	// 获取键数量
	keysCount, _ := client.DBSize(ctx).Result()
	health.KeysCount = keysCount

	return health, nil
}

// parseHealthCheckInfo 解析 Redis INFO 输出并填充健康检查数据
func parseHealthCheckInfo(info string, health *RedisHealthCheck) {
	health.Version = parseRedisInfo(info, "redis_version")
	health.UptimeSeconds = parseRedisInfoInt64(info, "uptime_in_seconds")
	health.ConnectedClients = parseRedisInfoInt(info, "connected_clients")
	health.UsedMemory = parseRedisInfo(info, "used_memory_human")
	health.UsedMemoryPeak = parseRedisInfo(info, "used_memory_peak_human")
	health.TotalHits = parseRedisInfoInt64(info, "keyspace_hits")
	health.TotalMisses = parseRedisInfoInt64(info, "keyspace_misses")
	health.OpsPerSec = parseRedisInfoInt64(info, "instantaneous_ops_per_sec")
	health.SlowlogLen = parseRedisInfoInt64(info, "slowlog_len")

	// 计算缓存命中率
	if health.TotalHits+health.TotalMisses > 0 {
		health.HitRate = float64(health.TotalHits) / float64(health.TotalHits+health.TotalMisses) * 100
	}

	// 计算内存使用率 (假设 maxmemory 为 2GB)
	maxMemory := int64(2 * 1024 * 1024 * 1024) // 2GB
	usedMemoryBytes := parseRedisInfoInt64(info, "used_memory")
	if maxMemory > 0 {
		health.MemoryUsagePercent = float64(usedMemoryBytes) / float64(maxMemory) * 100
	}
}

// IsRedisHealthy 检查 Redis 是否健康
func IsRedisHealthy(client *redis.Client) (bool, []string) {
	health, err := GetHealthCheck(client)
	if err != nil {
		return false, []string{fmt.Sprintf("Redis connection failed: %v", err)}
	}

	return evaluateHealthStatus(health)
}

// evaluateHealthStatus 根据健康检查结果生成警告信息
func evaluateHealthStatus(health *RedisHealthCheck) (bool, []string) {
	warnings := []string{}

	// 1. 检查连接状态
	if !health.Connected {
		warnings = append(warnings, "Redis not connected")
	}

	// 2. 检查缓存命中率
	if health.HitRate < 80 && health.TotalHits+health.TotalMisses > 100 {
		warnings = append(warnings,
			fmt.Sprintf("Low cache hit rate: %.1f%% (hits=%d, misses=%d)",
				health.HitRate, health.TotalHits, health.TotalMisses))
	}

	// 3. 检查内存使用率
	if health.MemoryUsagePercent > 80 {
		warnings = append(warnings,
			fmt.Sprintf("High memory usage: %.1f%%", health.MemoryUsagePercent))
	}

	// 4. 检查慢查询
	if health.SlowlogLen > 50 {
		warnings = append(warnings,
			fmt.Sprintf("Many slow queries: %d", health.SlowlogLen))
	}

	// 5. 检查连接数
	maxClients := 10000 // 默认 maxclients
	if health.ConnectedClients > int(float64(maxClients)*0.8) {
		warnings = append(warnings,
			fmt.Sprintf("High client connections: %d", health.ConnectedClients))
	}

	healthy := len(warnings) == 0
	return healthy, warnings
}

// RedisSlowLogEntry 慢查询日志条目
type RedisSlowLogEntry struct {
	ID        int64    `json:"id"`
	StartTime int64    `json:"start_time"`
	Duration  int64    `json:"duration_us"` // 微秒
	Command   string   `json:"command"`
	Key       string   `json:"key"`
	Args      []string `json:"args"`
}

// GetSlowLog 获取慢查询日志
func GetSlowLog(client *redis.Client, limit int64) ([]RedisSlowLogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slowlogs, err := client.SlowLogGet(ctx, limit).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get slowlog: %w", err)
	}

	entries := []RedisSlowLogEntry{}
	for _, sl := range slowlogs {
		entry := RedisSlowLogEntry{
			ID:        sl.ID,
			StartTime: sl.Time.Unix(),
			Duration:  sl.Duration.Microseconds(),
		}
		if len(sl.Args) > 0 {
			entry.Command = sl.Args[0]
			if len(sl.Args) > 1 {
				entry.Key = sl.Args[1]
				entry.Args = sl.Args[2:]
			}
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// RedisMemoryInfo 内存信息
type RedisMemoryInfo struct {
	UsedMemory         int64   `json:"used_memory_bytes"`
	UsedMemoryPeak     int64   `json:"used_memory_peak_bytes"`
	UsedMemoryRSS      int64   `json:"used_memory_rss"`
	FragmentationRatio float64 `json:"fragmentation_ratio"`
	MaxMemory          int64   `json:"max_memory_bytes"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	UsedMemoryHuman    string  `json:"used_memory_human"`
}

// GetMemoryInfo 获取内存信息
func GetMemoryInfo(client *redis.Client) (*RedisMemoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	memInfo := &RedisMemoryInfo{
		UsedMemory:      parseRedisInfoInt64(info, "used_memory"),
		UsedMemoryPeak:  parseRedisInfoInt64(info, "used_memory_peak"),
		UsedMemoryRSS:   parseRedisInfoInt64(info, "used_memory_rss"),
		MaxMemory:       parseRedisInfoInt64(info, "maxmemory"),
		UsedMemoryHuman: parseRedisInfo(info, "used_memory_human"),
	}

	// 计算碎片率
	if memInfo.UsedMemory > 0 {
		memInfo.FragmentationRatio = float64(memInfo.UsedMemoryRSS) / float64(memInfo.UsedMemory)
	}

	// 计算内存使用率
	if memInfo.MaxMemory > 0 {
		memInfo.MemoryUsagePercent = float64(memInfo.UsedMemory) / float64(memInfo.MaxMemory) * 100
	}

	return memInfo, nil
}

// 辅助函数: 解析 Redis INFO 输出
func parseRedisInfo(info string, key string) string {
	lines := splitLines(info)
	for _, line := range lines {
		if startsWith(line, key+":") {
			return substringAfter(line, key+":")
		}
	}
	return ""
}

func parseRedisInfoInt(info string, key string) int {
	val := parseRedisInfo(info, key)
	if val == "" {
		return 0
	}
	result := 0
	for _, c := range val {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			break
		}
	}
	return result
}

func parseRedisInfoInt64(info string, key string) int64 {
	val := parseRedisInfo(info, key)
	if val == "" {
		return 0
	}
	result := int64(0)
	for _, c := range val {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		} else {
			break
		}
	}
	return result
}

func splitLines(s string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func substringAfter(s, sep string) string {
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			return s[i+len(sep):]
		}
	}
	return s
}

// CacheWarmup 缓存预热
func CacheWarmup(client *redis.Client, keys map[string]string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipe := client.Pipeline()
	for key, value := range keys {
		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("cache warmup failed: %w", err)
	}

	log.Printf("Cache warmup completed: %d keys", len(keys))
	return nil
}

// CacheStats 缓存统计
type CacheStats struct {
	KeysCount    int64   `json:"keys_count"`
	HitRate      float64 `json:"hit_rate"`
	OpsPerSec    int64   `json:"ops_per_sec"`
	MemoryUsed   string  `json:"memory_used"`
	SlowlogCount int64   `json:"slowlog_count"`
}

// GetCacheStats 获取缓存统计
func GetCacheStats(client *redis.Client) (*CacheStats, error) {
	health, err := GetHealthCheck(client)
	if err != nil {
		return nil, err
	}

	return &CacheStats{
		KeysCount:    health.KeysCount,
		HitRate:      health.HitRate,
		OpsPerSec:    health.OpsPerSec,
		MemoryUsed:   health.UsedMemory,
		SlowlogCount: health.SlowlogLen,
	}, nil
}
