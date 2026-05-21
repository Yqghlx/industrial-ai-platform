package cache

import (
	"context"
	"log"
	"sync"
	"time"
)

// WarmupService handles cache preloading/warmup
type WarmupService struct {
	cache    CacheService
	loaders  []WarmupLoader
	mu       sync.Mutex
	warmupAt time.Time
}

// WarmupLoader defines a function that loads data to warmup cache
type WarmupLoader func(ctx context.Context, cache CacheService) error

// WarmupConfig contains configuration for warmup loaders
type WarmupConfig struct {
	Key  string
	Load func(ctx context.Context) ([]byte, error)
	TTL  time.Duration
}

// NewWarmupService creates a new warmup service
func NewWarmupService(cache CacheService) *WarmupService {
	return &WarmupService{
		cache:   cache,
		loaders: []WarmupLoader{},
	}
}

// RegisterLoader registers a warmup loader
func (w *WarmupService) RegisterLoader(loader WarmupLoader) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.loaders = append(w.loaders, loader)
}

// RegisterConfigLoader registers a loader from WarmupConfig
func (w *WarmupService) RegisterConfigLoader(cfg WarmupConfig) {
	w.RegisterLoader(func(ctx context.Context, cache CacheService) error {
		data, err := cfg.Load(ctx)
		if err != nil {
			return err
		}
		return cache.Set(ctx, cfg.Key, data, cfg.TTL)
	})
}

// Warmup executes all registered loaders
func (w *WarmupService) Warmup(ctx context.Context) error {
	if !w.cache.IsAvailable() {
		log.Println("[Cache Warmup] Cache not available, skipping warmup")
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	start := time.Now()
	var errors []error
	var successCount int

	for i, loader := range w.loaders {
		if ctx.Err() != nil {
			break
		}

		loaderCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := loader(loaderCtx, w.cache)
		cancel()

		if err != nil {
			errors = append(errors, err)
			log.Printf("[Cache Warmup] Loader %d failed: %v", i, err)
		} else {
			successCount++
		}
	}

	w.warmupAt = time.Now()
	duration := time.Since(start)

	log.Printf("[Cache Warmup] Completed in %v: %d succeeded, %d failed",
		duration, successCount, len(errors))

	if len(errors) > 0 {
		return errors[0] // Return first error
	}

	return nil
}

// WarmupAsync executes warmup in background
func (w *WarmupService) WarmupAsync() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := w.Warmup(ctx); err != nil {
			log.Printf("[Cache Warmup] Async warmup failed: %v", err)
		}
	}()
}

// GetWarmupTime returns the last warmup time
func (w *WarmupService) GetWarmupTime() time.Time {
	return w.warmupAt
}

// IsWarmupComplete returns true if warmup has been completed
func (w *WarmupService) IsWarmupComplete() bool {
	return !w.warmupAt.IsZero()
}

// ScheduleWarmup schedules periodic warmup
func (w *WarmupService) ScheduleWarmup(interval time.Duration) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				w.Warmup(ctx)
				cancel()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}
