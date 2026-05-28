package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================
// HealthHandlerNew - 健康检查Handler（新架构）
// ============================================

// CacheStatusProvider 缓存状态查询接口
type CacheStatusProvider interface {
	IsAvailable() bool
	GetStats() CacheStats
}

// CacheStats 缓存统计信息
type CacheStats struct {
	BackendType string
	Available   bool
	KeysStored  int64
	Hits        int64
	Misses      int64
	Errors      int64
}

// WSStatusProvider WebSocket状态查询接口
type WSStatusProvider interface {
	ClientCount() int
}

// HealthHandlerNew 健康检查处理器
type HealthHandlerNew struct {
	startTime time.Time
	// P1-06: 添加依赖注入支持，用于真实状态查询
	cacheProvider CacheStatusProvider
	wsProvider    WSStatusProvider
	wsCompressor  bool
}

// NewHealthHandlerNew 创建健康检查处理器
func NewHealthHandlerNew(startTime time.Time) *HealthHandlerNew {
	return &HealthHandlerNew{startTime: startTime}
}

// NewHealthHandlerNewWithDeps 创建健康检查处理器（带依赖注入）
// P1-06: 生产环境应使用此构造函数注入真实依赖
func NewHealthHandlerNewWithDeps(startTime time.Time, cacheProvider CacheStatusProvider, wsProvider WSStatusProvider, wsCompressor bool) *HealthHandlerNew {
	return &HealthHandlerNew{
		startTime:     startTime,
		cacheProvider: cacheProvider,
		wsProvider:    wsProvider,
		wsCompressor:  wsCompressor,
	}
}

// GetCacheStatus 获取缓存状态
// P1-06: 实现真实状态查询，支持依赖注入
func (h *HealthHandlerNew) GetCacheStatus(c *gin.Context) {
	uptime := int64(time.Since(h.startTime).Seconds())

	// 如果有注入缓存提供者，使用真实状态查询
	if h.cacheProvider != nil {
		stats := h.cacheProvider.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"status":       "ok",
			"backend":      stats.BackendType,
			"connected":    stats.Available,
			"keys_stored":  stats.KeysStored,
			"hits":         stats.Hits,
			"misses":       stats.Misses,
			"errors":       stats.Errors,
			"uptime":       uptime,
			"data_source":  "live", // 标记数据来源为实时数据
		})
		return
	}

	// Mock实现：当没有注入依赖时返回默认值
	// 注意：生产环境应使用 NewHealthHandlerNewWithDeps 注入真实依赖
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"backend":      "mock", // 明确标记为Mock数据
		"connected":    true,
		"uptime":       uptime,
		"data_source":  "mock", // 标记数据来源为Mock
		"note":         "inject CacheStatusProvider for live status",
	})
}

// GetWSStats 获取WebSocket状态
// P1-06: 实现真实状态查询，支持依赖注入
func (h *HealthHandlerNew) GetWSStats(c *gin.Context) {
	uptime := int64(time.Since(h.startTime).Seconds())

	// 如果有注入WebSocket提供者，使用真实状态查询
	if h.wsProvider != nil {
		clientCount := h.wsProvider.ClientCount()
		c.JSON(http.StatusOK, gin.H{
			"status":        "ok",
			"clients_count": clientCount,
			"compression":   h.wsCompressor,
			"uptime":        uptime,
			"data_source":   "live", // 标记数据来源为实时数据
		})
		return
	}

	// Mock实现：当没有注入依赖时返回默认值
	// 注意：生产环境应使用 NewHealthHandlerNewWithDeps 注入真实依赖
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"clients_count": 0,
		"compression":   true,
		"uptime":        uptime,
		"data_source":   "mock", // 标记数据来源为Mock
		"note":          "inject WSStatusProvider for live status",
	})
}
