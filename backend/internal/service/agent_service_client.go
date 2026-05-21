package service

import (
	"net/http"

	"github.com/industrial-ai/platform/internal/repository"
)

// NewAgentServiceWithClient creates a new agent service with custom HTTP client
// 用于测试场景，注入 Mock HTTP Client
func NewAgentServiceWithClient(
	taskLogRepo repository.AgentTaskLogRepositoryInterface,
	deviceRepo    repository.DeviceRepositoryInterface,
	telemetryRepo repository.TelemetryRepositoryInterface,
	config *AgentServiceConfig,
	client HTTPClientInterface,
) *AgentService {
	if config == nil {
		config = DefaultAgentServiceConfig()
	}
	if client == nil {
		// 创建默认 HTTP Client
		transport := &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			IdleConnTimeout:     config.IdleConnTimeout,
			DisableCompression:  false,
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		}
		client = &http.Client{
			Transport: transport,
			Timeout:   config.HTTPTimeout,
		}
	}
	return &AgentService{
		taskLogRepo:   taskLogRepo,
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		apiKey:        config.LLMAPIKey,
		baseURL:       config.LLMBaseURL,
		model:         config.LLMModel,
		httpClient:    client,
		config:        config,
	}
}