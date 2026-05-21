package circuitbreaker

import (
	"errors"
	"sync"
	"time"

	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// State 熔断器状态
type State int

const (
	StateClosed   State = 0 // 关闭状态 (正常)
	StateOpen     State = 1 // 打开状态 (熔断)
	StateHalfOpen State = 2 // 半开状态 (试探)
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Errors 熔断器错误
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// Config 熔断器配置
type Config struct {
	Name             string        // 熔断器名称
	FailureThreshold int           // 失败阈值 (百分比)
	MinRequests      int           // 最小请求数 (判断失败率前)
	OpenTimeout      time.Duration // 打开状态超时时间
	HalfOpenRequests int           // 半开状态允许请求数
	SuccessThreshold int           // 成功阈值 (半开状态后恢复)
}

// DefaultConfig 默认配置
func DefaultConfig(name string) *Config {
	return &Config{
		Name:             name,
		FailureThreshold: 50,               // 50% 失败率
		MinRequests:      10,               // 最少 10 次请求
		OpenTimeout:      30 * time.Second, // 30 秒后试探
		HalfOpenRequests: 3,                // 半开状态 3 次请求
		SuccessThreshold: 5,                // 连续 5 次成功恢复
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config Config

	// 状态
	state           State
	failureCount    int
	successCount    int
	requestCount    int
	lastFailureTime time.Time
	lastStateChange time.Time

	// 半开状态计数
	halfOpenSuccesses int
	halfOpenFailures  int

	mutex         sync.RWMutex
	onStateChange func(name string, old, new State)
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Call 执行请求 (带熔断保护)
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// 检查状态
	switch cb.state {
	case StateOpen:
		// 检查是否可以进入半开状态
		if time.Since(cb.lastStateChange) > cb.config.OpenTimeout {
			cb.transitionTo(StateHalfOpen)
			logger.L().Info("CircuitBreaker transitioning to half-open",
				zap.String("name", cb.config.Name),
				zap.String("state", "half-open"),
			)
		} else {
			return ErrCircuitBreakerOpen
		}

	case StateHalfOpen:
		// 检查半开状态请求限制
		if cb.halfOpenSuccesses+cb.halfOpenFailures >= cb.config.HalfOpenRequests {
			return ErrCircuitBreakerOpen
		}

	case StateClosed:
		// 正常状态，允许请求
	}

	// 执行请求
	err := fn()

	// 记录结果
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// CallWithFallback 执行请求 (带降级回调)
func (cb *CircuitBreaker) CallWithFallback(fn func() error, fallback func() error) error {
	err := cb.Call(fn)
	if err == ErrCircuitBreakerOpen {
		// 熔断状态，执行降级回调
		return fallback()
	}
	return err
}

// === 状态记录 ===

// recordSuccess 记录成功
func (cb *CircuitBreaker) recordSuccess() {
	cb.successCount++
	cb.requestCount++

	switch cb.state {
	case StateHalfOpen:
		cb.halfOpenSuccesses++
		// 检查是否可以恢复到关闭状态
		if cb.halfOpenSuccesses >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
			logger.L().Info("CircuitBreaker recovered to closed state",
				zap.String("name", cb.config.Name),
				zap.String("state", "closed"),
			)
		}

	case StateClosed:
		// 正常状态，重置失败计数
		cb.failureCount = 0
	}
}

// recordFailure 记录失败
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.requestCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateHalfOpen:
		cb.halfOpenFailures++
		// 半开状态失败，立即进入打开状态
		cb.transitionTo(StateOpen)
		logger.L().Warn("CircuitBreaker failure in half-open, back to open",
			zap.String("name", cb.config.Name),
			zap.String("state", "open"),
		)

	case StateClosed:
		// 检查失败率是否超过阈值
		if cb.requestCount >= cb.config.MinRequests {
			failureRate := cb.failureCount * 100 / cb.requestCount
			if failureRate >= cb.config.FailureThreshold {
				cb.transitionTo(StateOpen)
				logger.L().Warn("CircuitBreaker opening circuit due to high failure rate",
					zap.String("name", cb.config.Name),
					zap.Int("failure_rate", failureRate),
					zap.Int("threshold", cb.config.FailureThreshold),
				)
			}
		}
	}
}

// === 状态转换 ===

// transitionTo 状态转换
func (cb *CircuitBreaker) transitionTo(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// 重置计数器
	switch newState {
	case StateClosed:
		cb.failureCount = 0
		cb.successCount = 0
		cb.requestCount = 0

	case StateOpen:
		// 打开状态，等待超时

	case StateHalfOpen:
		cb.halfOpenSuccesses = 0
		cb.halfOpenFailures = 0
	}

	// 触发状态变更回调
	if cb.onStateChange != nil {
		cb.onStateChange(cb.config.Name, oldState, newState)
	}
}

// === 状态查询 ===

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats 获取统计信息
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	failureRate := 0
	if cb.requestCount > 0 {
		failureRate = cb.failureCount * 100 / cb.requestCount
	}

	return Stats{
		Name:            cb.config.Name,
		State:           cb.state.String(),
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		RequestCount:    cb.requestCount,
		FailureRate:     failureRate,
		LastFailureTime: cb.lastFailureTime,
		LastStateChange: cb.lastStateChange,
	}
}

// Stats 熔断器统计
type Stats struct {
	Name            string    `json:"name"`
	State           string    `json:"state"`
	FailureCount    int       `json:"failure_count"`
	SuccessCount    int       `json:"success_count"`
	RequestCount    int       `json:"request_count"`
	FailureRate     int       `json:"failure_rate"`
	LastFailureTime time.Time `json:"last_failure_time"`
	LastStateChange time.Time `json:"last_state_change"`
}

// === 回调设置 ===

// OnStateChange 设置状态变更回调
func (cb *CircuitBreaker) OnStateChange(callback func(name string, old, new State)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onStateChange = callback
}

// === 手动控制 ===

// ForceOpen 强制打开熔断器
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.transitionTo(StateOpen)
	logger.L().Warn("CircuitBreaker manually forced to open",
		zap.String("name", cb.config.Name),
	)
}

// ForceClose 强制关闭熔断器
func (cb *CircuitBreaker) ForceClose() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.transitionTo(StateClosed)
	logger.L().Info("CircuitBreaker manually forced to close",
		zap.String("name", cb.config.Name),
	)
}

// === 熔断器管理器 ===

// CircuitBreakerManager 熔断器管理器
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager 创建熔断器管理器
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Register 注册熔断器
func (m *CircuitBreakerManager) Register(config Config) *CircuitBreaker {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cb := NewCircuitBreaker(config)
	m.breakers[config.Name] = cb

	logger.L().Info("CircuitBreaker registered",
		zap.String("name", config.Name),
		zap.Int("failure_threshold", config.FailureThreshold),
	)
	return cb
}

// Get 获取熔断器
func (m *CircuitBreakerManager) Get(name string) *CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.breakers[name]
}

// GetAllStats 获取所有熔断器状态
func (m *CircuitBreakerManager) GetAllStats() map[string]Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]Stats)
	for name, cb := range m.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}

// RegisterDefaultBreakers 注册默认熔断器
func RegisterDefaultBreakers(m *CircuitBreakerManager) {
	// GLM API 熔断器
	glmConfig := Config{
		Name:             "glm_api",
		FailureThreshold: 50,
		MinRequests:      10,
		OpenTimeout:      30 * time.Second,
		HalfOpenRequests: 3,
		SuccessThreshold: 5,
	}
	m.Register(glmConfig)

	// Database 熔断器
	dbConfig := Config{
		Name:             "database",
		FailureThreshold: 30,
		MinRequests:      5,
		OpenTimeout:      60 * time.Second,
		HalfOpenRequests: 2,
		SuccessThreshold: 3,
	}
	m.Register(dbConfig)

	// Redis 熔断器
	redisConfig := Config{
		Name:             "redis",
		FailureThreshold: 40,
		MinRequests:      10,
		OpenTimeout:      30 * time.Second,
		HalfOpenRequests: 3,
		SuccessThreshold: 5,
	}
	m.Register(redisConfig)
}
