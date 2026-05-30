package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// SystemConfigRepo 系统配置数据访问层
type SystemConfigRepo struct {
	db database.QueryExecutor
}

// NewSystemConfigRepo 创建系统配置 Repository
func NewSystemConfigRepo(db database.QueryExecutor) *SystemConfigRepo {
	return &SystemConfigRepo{db: db}
}

// GetByKey 根据 key 获取配置项
func (r *SystemConfigRepo) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.QueryRow(ctx,
		"SELECT key, value, category, updated_at FROM system_config WHERE key = $1",
		key,
	).Scan(&config.Key, &config.Value, &config.Category, &config.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByCategory 根据 category 获取配置项列表
func (r *SystemConfigRepo) GetByCategory(ctx context.Context, category string) ([]*model.SystemConfig, error) {
	rows, err := r.db.Query(ctx,
		"SELECT key, value, category, updated_at FROM system_config WHERE category = $1",
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*model.SystemConfig
	for rows.Next() {
		var c model.SystemConfig
		if err := rows.Scan(&c.Key, &c.Value, &c.Category, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, &c)
	}
	return configs, rows.Err()
}

// Upsert 插入或更新配置项
func (r *SystemConfigRepo) Upsert(ctx context.Context, key, value, category string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO system_config (key, value, category, updated_at) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (key) DO UPDATE SET value = $2, category = $3, updated_at = $4`,
		key, value, category, time.Now(),
	)
	if err != nil {
		logger.L().Error("更新系统配置失败", zap.String("key", key), zap.Error(err))
	}
	return err
}

// SeedFromEnv 启动时将环境变量中的 LLM 配置同步到数据库（仅当数据库值为空时）
func (r *SystemConfigRepo) SeedFromEnv(ctx context.Context, apiKey, baseURL, model string) error {
	// 同步 API Key（仅当数据库中为空时）
	if apiKey != "" {
		existing, err := r.GetByKey(ctx, "llm_api_key")
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if existing == nil || existing.Value == "" {
			if err := r.Upsert(ctx, "llm_api_key", apiKey, "llm"); err != nil {
				return err
			}
		}
	}

	// 同步 Base URL
	if baseURL != "" {
		existing, err := r.GetByKey(ctx, "llm_base_url")
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if existing == nil || existing.Value == "" || existing.Value == "https://open.bigmodel.cn/api/paas/v4" {
			if err := r.Upsert(ctx, "llm_base_url", baseURL, "llm"); err != nil {
				return err
			}
		}
	}

	// 同步 Model
	if model != "" {
		existing, err := r.GetByKey(ctx, "llm_model")
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if existing == nil || existing.Value == "" || existing.Value == "glm-4-flash" {
			if err := r.Upsert(ctx, "llm_model", model, "llm"); err != nil {
				return err
			}
		}
	}

	return nil
}
