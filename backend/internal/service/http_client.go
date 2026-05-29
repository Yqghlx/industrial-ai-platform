package service

import (
	"net/http"
	"os"
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
	// 可通过 LLM_FALLBACK_URL 环境变量配置
	LLMFallbackURL string
}

// DefaultHealthServiceConfig returns default configuration
func DefaultHealthServiceConfig() HealthServiceConfig {
	// 从环境变量读取 LLM 配置，避免硬编码
	fallbackURL := os.Getenv("LLM_FALLBACK_URL")

	return HealthServiceConfig{
		CheckTimeout:  5 * time.Second,
		LLMAPIKey:     os.Getenv("LLM_API_KEY"),
		LLMBseURL:     os.Getenv("LLM_BASE_URL"),
		LLMModel:      os.Getenv("LLM_MODEL"),
		LLMFallbackURL: fallbackURL,
	}
}
