package service

import (
	"context"

	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// ConfigService 系统配置服务
type ConfigService struct {
	configRepo  *repository.SystemConfigRepo
	agentSvc    *AgentService
}

// NewConfigService 创建系统配置服务
func NewConfigService(configRepo *repository.SystemConfigRepo, agentSvc *AgentService) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		agentSvc:   agentSvc,
	}
}

// GetLLMConfig 获取大模型配置（API Key 脱敏）
func (s *ConfigService) GetLLMConfig(ctx context.Context) (*LLMConfigResponse, error) {
	configs, err := s.configRepo.GetByCategory(ctx, "llm")
	if err != nil {
		logger.L().Error("获取 LLM 配置失败", zap.Error(err))
		return nil, err
	}

	resp := &LLMConfigResponse{}
	for _, c := range configs {
		switch c.Key {
		case "llm_api_key":
			resp.LLMAPIKey = maskAPIKey(c.Value)
		case "llm_base_url":
			resp.LLMBaseURL = c.Value
		case "llm_model":
			resp.LLMModel = c.Value
		}
	}
	return resp, nil
}

// UpdateLLMConfig 更新大模型配置并即时刷新 AgentService
func (s *ConfigService) UpdateLLMConfig(ctx context.Context, update *LLMConfigUpdate) error {
	// 获取当前数据库中的完整 API Key（用于判断用户是否修改了 Key）
	currentKeyConfig, err := s.configRepo.GetByKey(ctx, "llm_api_key")
	if err != nil {
		logger.L().Error("获取当前 API Key 失败", zap.Error(err))
		return err
	}

	// 保存到数据库
	if err := s.configRepo.Upsert(ctx, "llm_base_url", update.LLMBaseURL, "llm"); err != nil {
		return err
	}
	if err := s.configRepo.Upsert(ctx, "llm_model", update.LLMModel, "llm"); err != nil {
		return err
	}

	// 处理 API Key：如果用户传入的是脱敏值（包含 ***），则不更新
	actualAPIKey := update.LLMAPIKey
	if isMasked(actualAPIKey) {
		actualAPIKey = currentKeyConfig.Value
	} else if actualAPIKey != "" {
		// 用户输入了新的 API Key
		if err := s.configRepo.Upsert(ctx, "llm_api_key", actualAPIKey, "llm"); err != nil {
			return err
		}
	}

	// 即时刷新 AgentService 的配置
	if s.agentSvc != nil {
		s.agentSvc.UpdateConfig(actualAPIKey, update.LLMBaseURL, update.LLMModel)
		logger.L().Info("LLM 配置已动态刷新",
			zap.String("base_url", update.LLMBaseURL),
			zap.String("model", update.LLMModel),
		)
	}

	return nil
}

// maskAPIKey 对 API Key 进行脱敏处理，只显示前 6 位
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 6 {
		return "***"
	}
	return key[:6] + "***"
}

// isMasked 判断字符串是否为脱敏后的值
func isMasked(s string) bool {
	return len(s) > 0 && (s[len(s)-3:] == "***" || s == "***")
}

// SeedFromEnv 将环境变量中的配置同步到数据库（仅当数据库值为空时）
func (s *ConfigService) SeedFromEnv(ctx context.Context, apiKey, baseURL, model string) error {
	return s.configRepo.SeedFromEnv(ctx, apiKey, baseURL, model)
}
