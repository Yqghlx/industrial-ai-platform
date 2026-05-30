package service

import (
	"context"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/internal/repository"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// ConfigService 系统配置服务
type ConfigService struct {
	configRepo  *repository.SystemConfigRepo
	llmRepo     *repository.LLMConfigRepo
	agentSvc    *AgentService
}

// NewConfigService 创建系统配置服务
func NewConfigService(configRepo *repository.SystemConfigRepo, llmRepo *repository.LLMConfigRepo, agentSvc *AgentService) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		llmRepo:    llmRepo,
		agentSvc:   agentSvc,
	}
}

// ============================================
// 单模型配置（兼容旧接口）
// ============================================

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
	currentKeyConfig, err := s.configRepo.GetByKey(ctx, "llm_api_key")
	if err != nil {
		logger.L().Error("获取当前 API Key 失败", zap.Error(err))
		return err
	}

	if err := s.configRepo.Upsert(ctx, "llm_base_url", update.LLMBaseURL, "llm"); err != nil {
		return err
	}
	if err := s.configRepo.Upsert(ctx, "llm_model", update.LLMModel, "llm"); err != nil {
		return err
	}

	actualAPIKey := update.LLMAPIKey
	if isMasked(actualAPIKey) {
		actualAPIKey = currentKeyConfig.Value
	} else if actualAPIKey != "" {
		if err := s.configRepo.Upsert(ctx, "llm_api_key", actualAPIKey, "llm"); err != nil {
			return err
		}
	}

	if s.agentSvc != nil {
		s.agentSvc.UpdateConfig(actualAPIKey, update.LLMBaseURL, update.LLMModel)
		logger.L().Info("LLM 配置已动态刷新",
			zap.String("base_url", update.LLMBaseURL),
			zap.String("model", update.LLMModel),
		)
	}

	return nil
}

// ============================================
// 多模型配置 CRUD
// ============================================

// ListLLMConfigs 获取所有模型配置列表（API Key 脱敏）
func (s *ConfigService) ListLLMConfigs(ctx context.Context) ([]*model.LLMConfigItemResponse, error) {
	items, err := s.llmRepo.List(ctx)
	if err != nil {
		logger.L().Error("获取模型配置列表失败", zap.Error(err))
		return nil, err
	}

	var resp []*model.LLMConfigItemResponse
	for _, item := range items {
		resp = append(resp, &model.LLMConfigItemResponse{
			ID:           item.ID,
			Name:         item.Name,
			APIKeyMasked: maskAPIKey(item.APIKey),
			BaseURL:      item.BaseURL,
			Model:        item.Model,
			IsActive:     item.IsActive,
			UpdatedAt:    item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}

// CreateLLMConfig 创建新的模型配置
func (s *ConfigService) CreateLLMConfig(ctx context.Context, req *model.LLMConfigCreateRequest) (*model.LLMConfigItemResponse, error) {
	item := &model.LLMConfigItem{
		Name:     req.Name,
		APIKey:   req.APIKey,
		BaseURL:  req.BaseURL,
		Model:    req.Model,
		IsActive: false,
	}
	if err := s.llmRepo.Create(ctx, item); err != nil {
		logger.L().Error("创建模型配置失败", zap.Error(err))
		return nil, err
	}

	return &model.LLMConfigItemResponse{
		ID:           item.ID,
		Name:         item.Name,
		APIKeyMasked: maskAPIKey(item.APIKey),
		BaseURL:      item.BaseURL,
		Model:        item.Model,
		IsActive:     item.IsActive,
		UpdatedAt:    item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateLLMConfigByID 更新指定模型配置
func (s *ConfigService) UpdateLLMConfigByID(ctx context.Context, id int, req *model.LLMConfigUpdateRequest) error {
	item, err := s.llmRepo.GetByID(ctx, id)
	if err != nil {
		logger.L().Error("获取模型配置失败", zap.Int("id", id), zap.Error(err))
		return err
	}

	// 处理 API Key：如果传入脱敏值，保留原值
	actualAPIKey := item.APIKey
	if req.APIKey != "" && !isMasked(req.APIKey) {
		actualAPIKey = req.APIKey
	}

	if req.Name != "" {
		item.Name = req.Name
	}
	if req.BaseURL != "" {
		item.BaseURL = req.BaseURL
	}
	if req.Model != "" {
		item.Model = req.Model
	}
	item.APIKey = actualAPIKey

	if err := s.llmRepo.Update(ctx, item); err != nil {
		return err
	}

	// 如果更新的是当前激活的模型，即时刷新 AgentService
	if item.IsActive && s.agentSvc != nil {
		s.agentSvc.UpdateConfig(item.APIKey, item.BaseURL, item.Model)
		logger.L().Info("激活模型配置已更新，AgentService 已刷新",
			zap.Int("id", id),
			zap.String("model", item.Model),
		)
	}

	return nil
}

// DeleteLLMConfig 删除模型配置（不允许删除激活中的模型）
func (s *ConfigService) DeleteLLMConfig(ctx context.Context, id int) error {
	item, err := s.llmRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item.IsActive {
		return ErrCannotDeleteActive
	}

	return s.llmRepo.Delete(ctx, id)
}

// SetActiveLLMConfig 切换激活的模型配置
func (s *ConfigService) SetActiveLLMConfig(ctx context.Context, id int) error {
	if err := s.llmRepo.SetActive(ctx, id); err != nil {
		return err
	}

	// 获取激活的模型配置并刷新 AgentService
	active, err := s.llmRepo.GetActive(ctx)
	if err != nil {
		logger.L().Error("获取激活模型配置失败", zap.Error(err))
		return err
	}

	if s.agentSvc != nil {
		s.agentSvc.UpdateConfig(active.APIKey, active.BaseURL, active.Model)
		logger.L().Info("已切换激活模型，AgentService 已刷新",
			zap.Int("id", active.ID),
			zap.String("name", active.Name),
			zap.String("model", active.Model),
		)
	}

	return nil
}

// ============================================
// 辅助方法
// ============================================

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

// EnsureActiveConfig 确保至少有一个激活的模型配置
func (s *ConfigService) EnsureActiveConfig(ctx context.Context) error {
	_, err := s.llmRepo.GetActive(ctx)
	if err != nil {
		// 没有激活配置，尝试激活第一个
		items, listErr := s.llmRepo.List(ctx)
		if listErr != nil {
			return listErr
		}
		if len(items) > 0 {
			return s.llmRepo.SetActive(ctx, items[0].ID)
		}
	}
	return nil
}
