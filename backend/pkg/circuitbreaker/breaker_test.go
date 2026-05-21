package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_NewCircuitBreaker(t *testing.T) {
	t.Run("creates with default config", func(t *testing.T) {
		config := DefaultConfig("test")
		cb := NewCircuitBreaker(*config)

		require.NotNil(t, cb)
		assert.Equal(t, StateClosed, cb.GetState())
	})
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	t.Run("closed to open on failure threshold", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      5,
			OpenTimeout:      100 * time.Millisecond,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Make failures to trigger open state (50% threshold)
		// Note: success in closed state resets failureCount, so we need to trigger
		// the threshold before any success
		// With MinRequests=5 and FailureThreshold=50%, we need >=3 failures out of 5 requests
		// But if we have any success before reaching MinRequests, failureCount resets
		// So we make 5 failures first (100% failure rate) which triggers open state
		for i := 0; i < 5; i++ {
			_ = cb.Call(func() error { return errors.New("error") })
		}

		assert.Equal(t, StateOpen, cb.GetState())
	})

	t.Run("open to half-open after timeout", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      50 * time.Millisecond,
			HalfOpenRequests: 3,
			SuccessThreshold: 2,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })
		assert.Equal(t, StateOpen, cb.GetState())

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Next call should transition to half-open
		_ = cb.Call(func() error { return nil })
		assert.Equal(t, StateHalfOpen, cb.GetState())
	})

	t.Run("half-open to closed on success threshold", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      50 * time.Millisecond,
			HalfOpenRequests: 5,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })
		assert.Equal(t, StateOpen, cb.GetState())

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Make enough successes in half-open state
		for i := 0; i < 3; i++ {
			_ = cb.Call(func() error { return nil })
		}

		assert.Equal(t, StateClosed, cb.GetState())
	})

	t.Run("half-open to open on failure", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      50 * time.Millisecond,
			HalfOpenRequests: 5,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })
		assert.Equal(t, StateOpen, cb.GetState())

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Transition to half-open
		_ = cb.Call(func() error { return nil })
		assert.Equal(t, StateHalfOpen, cb.GetState())

		// Fail in half-open state
		_ = cb.Call(func() error { return errors.New("error") })

		// Should go back to open
		assert.Equal(t, StateOpen, cb.GetState())
	})
}

func TestCircuitBreaker_Call(t *testing.T) {
	t.Run("successful call in closed state", func(t *testing.T) {
		config := DefaultConfig("test")
		cb := NewCircuitBreaker(*config)

		err := cb.Call(func() error { return nil })
		assert.NoError(t, err)
		assert.Equal(t, StateClosed, cb.GetState())

		stats := cb.GetStats()
		assert.Equal(t, 1, stats.SuccessCount)
		assert.Equal(t, 1, stats.RequestCount)
	})

	t.Run("failed call in closed state", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 80, // High threshold so it doesn't open
			MinRequests:      10,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		err := cb.Call(func() error { return errors.New("error") })
		assert.Error(t, err)
		assert.Equal(t, StateClosed, cb.GetState())

		stats := cb.GetStats()
		assert.Equal(t, 1, stats.FailureCount)
		assert.Equal(t, 1, stats.RequestCount)
	})

	t.Run("call returns error when open", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Next call should return ErrCircuitBreakerOpen
		err := cb.Call(func() error { return nil })
		assert.ErrorIs(t, err, ErrCircuitBreakerOpen)
	})

	t.Run("half-open requests limit", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      50 * time.Millisecond,
			HalfOpenRequests: 2,
			SuccessThreshold: 5, // Higher than HalfOpenRequests
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })
		assert.Equal(t, StateOpen, cb.GetState())

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Make calls up to limit
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })

		// Next call should be blocked
		err := cb.Call(func() error { return nil })
		assert.ErrorIs(t, err, ErrCircuitBreakerOpen)
	})
}

func TestCircuitBreaker_CallWithFallback(t *testing.T) {
	t.Run("use fallback when open", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		fallbackCalled := false
		err := cb.CallWithFallback(
			func() error { return errors.New("should not be called") },
			func() error {
				fallbackCalled = true
				return nil
			},
		)

		assert.NoError(t, err)
		assert.True(t, fallbackCalled)
	})

	t.Run("no fallback when closed", func(t *testing.T) {
		config := DefaultConfig("test")
		cb := NewCircuitBreaker(*config)

		fallbackCalled := false
		err := cb.CallWithFallback(
			func() error { return nil },
			func() error {
				fallbackCalled = true
				return nil
			},
		)

		assert.NoError(t, err)
		assert.False(t, fallbackCalled)
	})

	t.Run("return original error when not open", func(t *testing.T) {
		config := DefaultConfig("test")
		cb := NewCircuitBreaker(*config)

		originalErr := errors.New("original error")
		err := cb.CallWithFallback(
			func() error { return originalErr },
			func() error { return nil },
		)

		assert.ErrorIs(t, err, originalErr)
	})
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	t.Run("stats are tracked correctly", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 80,
			MinRequests:      10,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Make some calls
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return errors.New("error") })

		stats := cb.GetStats()

		assert.Equal(t, "test", stats.Name)
		assert.Equal(t, "closed", stats.State)
		assert.Equal(t, 2, stats.SuccessCount)
		assert.Equal(t, 1, stats.FailureCount)
		assert.Equal(t, 3, stats.RequestCount)
		assert.Equal(t, 33, stats.FailureRate) // 1/3 = 33%
	})
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	t.Run("callback is called on state change", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      50 * time.Millisecond,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		stateChanges := []State{}
		cb.OnStateChange(func(name string, old, new State) {
			stateChanges = append(stateChanges, new)
		})

		// Trigger open
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Transition to half-open
		_ = cb.Call(func() error { return nil })

		// Make successes to close
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })

		assert.Contains(t, stateChanges, StateOpen)
		assert.Contains(t, stateChanges, StateHalfOpen)
	})
}

func TestCircuitBreaker_ForceOpen(t *testing.T) {
	t.Run("force open transitions to open", func(t *testing.T) {
		config := DefaultConfig("test")
		cb := NewCircuitBreaker(*config)

		assert.Equal(t, StateClosed, cb.GetState())

		cb.ForceOpen()

		assert.Equal(t, StateOpen, cb.GetState())
	})
}

func TestCircuitBreaker_ForceClose(t *testing.T) {
	t.Run("force close transitions to closed", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      2,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })
		assert.Equal(t, StateOpen, cb.GetState())

		cb.ForceClose()

		assert.Equal(t, StateClosed, cb.GetState())
	})
}

func TestCircuitBreaker_Concurrent_Access(t *testing.T) {
	t.Run("concurrent calls are safe", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 80,
			MinRequests:      10,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 10,
			SuccessThreshold: 5,
		}
		cb := NewCircuitBreaker(config)

		var wg sync.WaitGroup
		numOps := 100

		// Concurrent successes
		wg.Add(numOps)
		for i := 0; i < numOps; i++ {
			go func() {
				defer wg.Done()
				_ = cb.Call(func() error { return nil })
			}()
		}

		// Concurrent failures
		wg.Add(numOps)
		for i := 0; i < numOps; i++ {
			go func() {
				defer wg.Done()
				_ = cb.Call(func() error { return errors.New("error") })
			}()
		}

		// Concurrent state checks
		wg.Add(numOps)
		for i := 0; i < numOps; i++ {
			go func() {
				defer wg.Done()
				_ = cb.GetState()
				_ = cb.GetStats()
			}()
		}

		wg.Wait()
		// If we reach here without panic or race condition, test passes
	})
}

func TestCircuitBreakerManager_Register(t *testing.T) {
	t.Run("register circuit breaker", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		config := DefaultConfig("test-breaker")
		cb := mgr.Register(*config)

		require.NotNil(t, cb)
		assert.Equal(t, cb, mgr.Get("test-breaker"))
	})

	t.Run("get non-existent breaker returns nil", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		cb := mgr.Get("non-existent")
		assert.Nil(t, cb)
	})
}

func TestCircuitBreakerManager_GetAllStats(t *testing.T) {
	t.Run("get all stats", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		config1 := DefaultConfig("breaker1")
		config2 := DefaultConfig("breaker2")

		_ = mgr.Register(*config1)
		_ = mgr.Register(*config2)

		// Make some calls to breaker1
		cb1 := mgr.Get("breaker1")
		_ = cb1.Call(func() error { return nil })

		stats := mgr.GetAllStats()

		assert.Len(t, stats, 2)
		assert.Contains(t, stats, "breaker1")
		assert.Contains(t, stats, "breaker2")
		assert.Equal(t, 1, stats["breaker1"].SuccessCount)
	})
}

func TestCircuitBreakerManager_Concurrent_Access(t *testing.T) {
	t.Run("concurrent register and get", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()

		var wg sync.WaitGroup
		numOps := 50

		// Concurrent register
		wg.Add(numOps)
		for i := 0; i < numOps; i++ {
			go func(idx int) {
				defer wg.Done()
				config := Config{
					Name:             "breaker-" + string(rune(idx)),
					FailureThreshold: 50,
					MinRequests:      5,
					OpenTimeout:      time.Second,
					HalfOpenRequests: 3,
					SuccessThreshold: 3,
				}
				_ = mgr.Register(config)
			}(i)
		}

		// Concurrent get
		wg.Add(numOps)
		for i := 0; i < numOps; i++ {
			go func(idx int) {
				defer wg.Done()
				_ = mgr.Get("breaker-" + string(rune(idx)))
			}(i)
		}

		wg.Wait()
	})
}

func TestState_String(t *testing.T) {
	t.Run("state string representation", func(t *testing.T) {
		assert.Equal(t, "closed", StateClosed.String())
		assert.Equal(t, "open", StateOpen.String())
		assert.Equal(t, "half-open", StateHalfOpen.String())
		assert.Equal(t, "unknown", State(99).String())
	})
}

func TestRegisterDefaultBreakers(t *testing.T) {
	t.Run("registers default breakers", func(t *testing.T) {
		mgr := NewCircuitBreakerManager()
		RegisterDefaultBreakers(mgr)

		assert.NotNil(t, mgr.Get("glm_api"))
		assert.NotNil(t, mgr.Get("database"))
		assert.NotNil(t, mgr.Get("redis"))
	})
}

func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	t.Run("exact threshold triggers open", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 50,
			MinRequests:      4,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// 2 successes, 2 failures = 50% failure rate
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		assert.Equal(t, StateOpen, cb.GetState())
	})

	t.Run("below threshold stays closed", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 80,
			MinRequests:      5,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// 1 failure, 4 successes = 20% failure rate
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })

		assert.Equal(t, StateClosed, cb.GetState())
	})
}

func TestCircuitBreaker_SuccessCount_Reset(t *testing.T) {
	t.Run("success resets failure count in closed state", func(t *testing.T) {
		config := Config{
			Name:             "test",
			FailureThreshold: 80,
			MinRequests:      5,
			OpenTimeout:      time.Second,
			HalfOpenRequests: 3,
			SuccessThreshold: 3,
		}
		cb := NewCircuitBreaker(config)

		// Make some failures
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		stats := cb.GetStats()
		assert.Equal(t, 2, stats.FailureCount)

		// Make a success
		_ = cb.Call(func() error { return nil })

		stats = cb.GetStats()
		assert.Equal(t, 0, stats.FailureCount)
	})
}
