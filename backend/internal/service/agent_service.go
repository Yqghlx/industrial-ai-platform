package service

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/cache"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// AgentServiceConfig holds configuration for AgentService
// FIX-039: HTTP Timeout 和连接池配置外部化
type AgentServiceConfig struct {
	// HTTP Client 配置
	HTTPTimeout         time.Duration // HTTP 请求超时时间
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每个主机最大空闲连接数
	IdleConnTimeout     time.Duration // 空闲连接超时时间

	// LLM 配置
	LLMAPIKey  string
	LLMBaseURL string
	LLMModel   string

	// 代理配置
	HTTPProxy string
}

// DefaultAgentServiceConfig 返回默认配置
// FIX-039: 提供合理的默认值
func DefaultAgentServiceConfig() *AgentServiceConfig {
	return &AgentServiceConfig{
		HTTPTimeout:         30 * time.Second, // 默认30秒，比60秒更合理
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		LLMBaseURL:          "https://open.bigmodel.cn/api/paas/v4",
		LLMModel:            "glm-4-flash",
	}
}

// LoadAgentServiceConfigFromEnv 从环境变量加载配置
// FIX-039: 支持环境变量配置
func LoadAgentServiceConfigFromEnv() *AgentServiceConfig {
	config := DefaultAgentServiceConfig()

	// HTTP Timeout 配置
	if v := os.Getenv("LLM_HTTP_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil && timeout > 0 {
			config.HTTPTimeout = time.Duration(timeout) * time.Second
		}
	}

	// 连接池配置
	if v := os.Getenv("LLM_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxIdleConns = n
		}
	}

	if v := os.Getenv("LLM_MAX_IDLE_CONNS_PER_HOST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxIdleConnsPerHost = n
		}
	}

	if v := os.Getenv("LLM_IDLE_CONN_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil && timeout > 0 {
			config.IdleConnTimeout = time.Duration(timeout) * time.Second
		}
	}

	// LLM 配置
	config.LLMAPIKey = os.Getenv("LLM_API_KEY")

	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		config.LLMBaseURL = v
	}

	if v := os.Getenv("LLM_MODEL"); v != "" {
		config.LLMModel = v
	}

	// 代理配置
	config.HTTPProxy = os.Getenv("HTTP_PROXY")
	if config.HTTPProxy == "" {
		config.HTTPProxy = os.Getenv("HTTPS_PROXY")
	}

	return config
}

// AgentService handles AI agent queries
type AgentService struct {
	taskLogRepo   repository.AgentTaskLogRepositoryInterface
	deviceRepo    repository.DeviceRepositoryInterface
	telemetryRepo repository.TelemetryRepositoryInterface
	apiKey        string
	baseURL       string
	model         string
	httpClient    HTTPClientInterface // FIX-019: HTTP Client 接口化
	config        *AgentServiceConfig
	optimizer     *AgentOptimizer // P2-3: Queue + Cache optimization
}

// NewAgentService creates a new agent service
// OPT-002: Added cacheSvc parameter for optimizer initialization
func NewAgentService(
	taskLogRepo repository.AgentTaskLogRepositoryInterface,
	deviceRepo repository.DeviceRepositoryInterface,
	telemetryRepo repository.TelemetryRepositoryInterface,
	cacheSvc cache.CacheService, // OPT-002: Added for caching
) *AgentService {
	// FIX-039: 从环境变量加载配置
	config := LoadAgentServiceConfigFromEnv()

	// FIX-019: 创建共享 HTTP Client，配置连接池参数
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  false,
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.HTTPTimeout, // FIX-039: 使用配置的超时时间
	}

	// 设置代理（如果有环境变量配置）
	if config.HTTPProxy != "" {
		if proxy, err := url.Parse(config.HTTPProxy); err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	// OPT-002: Initialize optimizer with cache for response caching
	var optimizer *AgentOptimizer
	if cacheSvc != nil {
		optimizer = NewAgentOptimizer(cacheSvc, 10) // Max 10 concurrent LLM calls
		logger.L().Info("Agent optimizer initialized with caching",
			zap.Int("max_concurrent", 10),
			zap.Duration("cache_ttl", 30*time.Minute),
		)
	}

	return &AgentService{
		taskLogRepo:   taskLogRepo,
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		apiKey:        config.LLMAPIKey,
		baseURL:       config.LLMBaseURL,
		model:         config.LLMModel,
		httpClient:    httpClient,
		config:        config,
		optimizer:     optimizer, // OPT-002: Enable caching
	}
}

// NewAgentServiceWithConfig creates a new agent service with explicit config
// FIX-039: 支持自定义配置
func NewAgentServiceWithConfig(
	taskLogRepo repository.AgentTaskLogRepositoryInterface,
	deviceRepo repository.DeviceRepositoryInterface,
	telemetryRepo repository.TelemetryRepositoryInterface,
	config *AgentServiceConfig,
) *AgentService {
	if config == nil {
		config = DefaultAgentServiceConfig()
	}

	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  false,
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.HTTPTimeout,
	}

	if config.HTTPProxy != "" {
		if proxy, err := url.Parse(config.HTTPProxy); err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	return &AgentService{
		taskLogRepo:   taskLogRepo,
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		apiKey:        config.LLMAPIKey,
		baseURL:       config.LLMBaseURL,
		model:         config.LLMModel,
		httpClient:    httpClient,
		config:        config,
	}
}

// Query processes an AI agent query
// P2-3: Optimized with queue + cache
func (s *AgentService) Query(ctx context.Context, query model.AgentQuery) (*model.AgentResponse, error) {
	// Generate session ID if not provided
	sessionID := query.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	// Determine which agent to use
	agent := s.determineAgent(query.Query)

	// P2-3: Try cached answer first
	if s.optimizer != nil {
		if cachedAnswer, found := s.optimizer.GetCachedAnswer(ctx, query.Query); found {
			logger.L().Info("Using cached answer",
				zap.String("session_id", sessionID),
			)
			return &model.AgentResponse{
				SessionID: sessionID,
				Response:  cachedAnswer,
				Agent:     agent,
			}, nil
		}
	}

	// Try to get real response from LLM
	var response string
	var err error

	if s.apiKey != "" {
		// P2-3: Acquire slot for LLM call (queue mechanism)
		if s.optimizer != nil {
			if err := s.optimizer.AcquireSlot(ctx); err != nil {
				logger.L().Warn("Queue wait timeout, using mock response",
					zap.String("session_id", sessionID),
					zap.Error(err),
				)
				response = s.mockResponse(query.Query, agent)
			} else {
				defer s.optimizer.ReleaseSlot()
				response, err = s.callLLM(ctx, query.Query, query.Context, agent)
				if err != nil {
					logger.L().Warn("LLM call failed, falling back to mock",
						zap.String("session_id", sessionID),
						zap.Error(err),
					)
					response = s.mockResponse(query.Query, agent)
				} else {
					// Cache successful answer
					s.optimizer.CacheAnswer(ctx, query.Query, response)
				}
			}
		} else {
			// Use real LLM without optimizer
			response, err = s.callLLM(ctx, query.Query, query.Context, agent)
			if err != nil {
				logger.L().Warn("LLM call failed, falling back to mock",
					zap.String("session_id", sessionID),
					zap.Error(err),
				)
				response = s.mockResponse(query.Query, agent)
			}
		}
	} else {
		// Use mock response
		logger.L().Debug("No LLM_API_KEY set, using mock response",
			zap.String("session_id", sessionID),
		)
		response = s.mockResponse(query.Query, agent)
	}

	// Log the task
	taskLog := &model.AgentTaskLog{
		SessionID:  sessionID,
		Query:      query.Query,
		Response:   response,
		Agent:      agent,
		ExecutedAt: time.Now(),
	}
	if err := s.taskLogRepo.Create(ctx, taskLog); err != nil {
		logger.L().Error("Failed to create task log", zap.Error(err))
	}

	return &model.AgentResponse{
		SessionID: sessionID,
		Response:  response,
		Agent:     agent,
		Timestamp: time.Now(),
	}, nil
}

// determineAgent determines which agent to use based on query
