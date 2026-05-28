package service

import (
	"net/http"
	"time"
)

// HTTPClientInterface defines the interface for HTTP client
type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// HealthServiceConfig holds configuration for HealthService
type HealthServiceConfig struct {
	LLMAPIKey    string
	LLMBseURL    string
	LLMModel     string
	CheckTimeout time.Duration
	// P2-04: Fallback LLM URL for health checks when LLMBseURL is empty
	LLMFallbackURL string
}

// DefaultHealthServiceConfig returns default configuration
func DefaultHealthServiceConfig() HealthServiceConfig {
	return HealthServiceConfig{
		CheckTimeout: 5 * time.Second,
	}
}
