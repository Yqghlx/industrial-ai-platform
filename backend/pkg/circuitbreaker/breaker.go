package circuitbreaker

import (
	"errors"
	"sync"
	"time"

	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// State represents circuit breaker state
type State int

const (
	StateClosed   State = 0 // Closed state (normal operation)
	StateOpen     State = 1 // Open state (circuit breaker tripped)
	StateHalfOpen State = 2 // Half-open state (probing)
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

// Errors represents circuit breaker errors
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// Config represents circuit breaker configuration
type Config struct {
	Name             string        // Circuit breaker name
	FailureThreshold int           // Failure threshold (percentage)
	MinRequests      int           // Minimum requests before checking failure rate
	OpenTimeout      time.Duration // Open state timeout duration
	HalfOpenRequests int           // Allowed requests in half-open state
	SuccessThreshold int           // Success threshold to recover from half-open state
}

// DefaultConfig returns default configuration
func DefaultConfig(name string) *Config {
	return &Config{
		Name:             name,
		FailureThreshold: 50,               // 50% failure rate
		MinRequests:      10,               // Minimum 10 requests
		OpenTimeout:      30 * time.Second, // Probe after 30 seconds
		HalfOpenRequests: 3,                // 3 requests in half-open state
		SuccessThreshold: 5,                // Recover after 5 consecutive successes
	}
}

// CircuitBreaker represents a circuit breaker
type CircuitBreaker struct {
	config Config

	// State tracking
	state           State
	failureCount    int
	successCount    int
	requestCount    int
	lastFailureTime time.Time
	lastStateChange time.Time

	// Half-open state counters
	halfOpenSuccesses int
	halfOpenFailures  int

	mutex         sync.RWMutex
	onStateChange func(name string, old, new State)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Call executes a request with circuit breaker protection
// Optimization: separate state check and state update, user function execution without lock to prevent system blocking
func (cb *CircuitBreaker) Call(fn func() error) error {
	// === Phase 1: State check - briefly hold lock ===
	cb.mutex.Lock()
	currentState := cb.state
	allowed := true

	switch currentState {
	case StateOpen:
		// Check if can transition to half-open state
		if time.Since(cb.lastStateChange) > cb.config.OpenTimeout {
			cb.transitionTo(StateHalfOpen)
			logger.L().Info("CircuitBreaker transitioning to half-open",
				zap.String("name", cb.config.Name),
				zap.String("state", "half-open"),
			)
		} else {
			allowed = false
		}

	case StateHalfOpen:
		// Check half-open state request limit
		if cb.halfOpenSuccesses+cb.halfOpenFailures >= cb.config.HalfOpenRequests {
			allowed = false
		}

	case StateClosed:
		// Normal state, allow request
	}
	cb.mutex.Unlock()

	// If not allowed, return immediately
	if !allowed {
		return ErrCircuitBreakerOpen
	}

	// === Phase 2: Execute user function - without lock ===
	err := fn()

	// === Phase 3: Update state - briefly hold lock again ===
	cb.mutex.Lock()
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
	cb.mutex.Unlock()

	return err
}

// CallWithFallback executes a request with fallback callback
func (cb *CircuitBreaker) CallWithFallback(fn func() error, fallback func() error) error {
	err := cb.Call(fn)
	if err == ErrCircuitBreakerOpen {
		// Circuit breaker open, execute fallback callback
		return fallback()
	}
	return err
}

// === State Recording ===

// recordSuccess records a successful request
// Updates success counters in different states:
//   - HalfOpen state: increment half-open success count, recover to Closed state when threshold reached
//   - Closed state: reset failure counter, maintain normal state
//
// Note: Must hold write lock before calling this method
func (cb *CircuitBreaker) recordSuccess() {
	cb.successCount++
	cb.requestCount++

	switch cb.state {
	case StateHalfOpen:
		cb.halfOpenSuccesses++
		// Check if can recover to closed state
		if cb.halfOpenSuccesses >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
			logger.L().Info("CircuitBreaker recovered to closed state",
				zap.String("name", cb.config.Name),
				zap.String("state", "closed"),
			)
		}

	case StateClosed:
		// Normal state, reset failure count
		cb.failureCount = 0
	}
}

// recordFailure records a failed request
// Updates failure counters in different states:
//   - HalfOpen state: immediately transition to Open state, half-open probe failed
//   - Closed state: increment failure count, trigger circuit breaker when failure rate threshold reached
//
// Note: Must hold write lock before calling this method
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.requestCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateHalfOpen:
		cb.halfOpenFailures++
		// Half-open state failure, immediately enter open state
		cb.transitionTo(StateOpen)
		logger.L().Warn("CircuitBreaker failure in half-open, back to open",
			zap.String("name", cb.config.Name),
			zap.String("state", "open"),
		)

	case StateClosed:
		// Check if failure rate exceeds threshold
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

// === State Transition ===

// transitionTo executes state transition
// Parameters:
//   - newState: target state (StateClosed, StateOpen, StateHalfOpen)
//
// Functions:
//   - Update circuit breaker current state
//   - Record state change time
//   - Reset related counters based on new state
//   - Trigger state change callback function (if set)
//
// Notes:
//   - If target state equals current state, no operation is performed
//   - Must hold write lock before calling this method
func (cb *CircuitBreaker) transitionTo(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters
	switch newState {
	case StateClosed:
		cb.failureCount = 0
		cb.successCount = 0
		cb.requestCount = 0

	case StateOpen:
		// Open state, wait for timeout

	case StateHalfOpen:
		cb.halfOpenSuccesses = 0
		cb.halfOpenFailures = 0
	}

	// Trigger state change callback
	if cb.onStateChange != nil {
		cb.onStateChange(cb.config.Name, oldState, newState)
	}
}

// === State Query ===

// GetState returns current state
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns statistics
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

// Stats represents circuit breaker statistics
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

// === Callback Settings ===

// OnStateChange sets state change callback
func (cb *CircuitBreaker) OnStateChange(callback func(name string, old, new State)) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.onStateChange = callback
}

// === Manual Control ===

// ForceOpen forces circuit breaker to open
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.transitionTo(StateOpen)
	logger.L().Warn("CircuitBreaker manually forced to open",
		zap.String("name", cb.config.Name),
	)
}

// ForceClose forces circuit breaker to close
func (cb *CircuitBreaker) ForceClose() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.transitionTo(StateClosed)
	logger.L().Info("CircuitBreaker manually forced to close",
		zap.String("name", cb.config.Name),
	)
}

// === Circuit Breaker Manager ===

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Register registers a circuit breaker
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

// Get retrieves a circuit breaker by name
func (m *CircuitBreakerManager) Get(name string) *CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.breakers[name]
}

// GetAllStats returns statistics for all circuit breakers
func (m *CircuitBreakerManager) GetAllStats() map[string]Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]Stats)
	for name, cb := range m.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}

// RegisterDefaultBreakers registers default circuit breakers
func RegisterDefaultBreakers(m *CircuitBreakerManager) {
	// GLM API circuit breaker
	glmConfig := Config{
		Name:             "glm_api",
		FailureThreshold: 50,
		MinRequests:      10,
		OpenTimeout:      30 * time.Second,
		HalfOpenRequests: 3,
		SuccessThreshold: 5,
	}
	m.Register(glmConfig)

	// Database circuit breaker
	dbConfig := Config{
		Name:             "database",
		FailureThreshold: 30,
		MinRequests:      5,
		OpenTimeout:      60 * time.Second,
		HalfOpenRequests: 2,
		SuccessThreshold: 3,
	}
	m.Register(dbConfig)

	// Redis circuit breaker
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
