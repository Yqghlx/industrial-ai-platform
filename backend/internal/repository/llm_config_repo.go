package repository

import (
	"context"
	"time"

	"github.com/industrial-ai/platform/internal/model"
	"github.com/industrial-ai/platform/pkg/database"
	"github.com/industrial-ai/platform/pkg/logger"
	"go.uber.org/zap"
)

// LLMConfigRepo 多模型配置数据访问层
type LLMConfigRepo struct {
	db database.QueryExecutor
}

// NewLLMConfigRepo 创建多模型配置 Repository
func NewLLMConfigRepo(db database.QueryExecutor) *LLMConfigRepo {
	return &LLMConfigRepo{db: db}
}

// List 获取所有模型配置
func (r *LLMConfigRepo) List(ctx context.Context) ([]*model.LLMConfigItem, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, name, api_key, base_url, model, is_active, created_at, updated_at FROM llm_configs ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.LLMConfigItem
	for rows.Next() {
		var item model.LLMConfigItem
		if err := rows.Scan(&item.ID, &item.Name, &item.APIKey, &item.BaseURL, &item.Model, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

// GetByID 根据 ID 获取模型配置
func (r *LLMConfigRepo) GetByID(ctx context.Context, id int) (*model.LLMConfigItem, error) {
	var item model.LLMConfigItem
	err := r.db.QueryRow(ctx,
		"SELECT id, name, api_key, base_url, model, is_active, created_at, updated_at FROM llm_configs WHERE id = $1",
		id,
	).Scan(&item.ID, &item.Name, &item.APIKey, &item.BaseURL, &item.Model, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetActive 获取当前激活的模型配置
func (r *LLMConfigRepo) GetActive(ctx context.Context) (*model.LLMConfigItem, error) {
	var item model.LLMConfigItem
	err := r.db.QueryRow(ctx,
		"SELECT id, name, api_key, base_url, model, is_active, created_at, updated_at FROM llm_configs WHERE is_active = TRUE LIMIT 1",
	).Scan(&item.ID, &item.Name, &item.APIKey, &item.BaseURL, &item.Model, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Create 创建新的模型配置
func (r *LLMConfigRepo) Create(ctx context.Context, item *model.LLMConfigItem) error {
	err := r.db.QueryRow(ctx,
		`INSERT INTO llm_configs (name, api_key, base_url, model, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		item.Name, item.APIKey, item.BaseURL, item.Model, item.IsActive, time.Now(), time.Now(),
	).Scan(&item.ID)
	if err != nil {
		logger.L().Error("创建模型配置失败", zap.Error(err))
	}
	return err
}

// Update 更新模型配置
func (r *LLMConfigRepo) Update(ctx context.Context, item *model.LLMConfigItem) error {
	_, err := r.db.Exec(ctx,
		`UPDATE llm_configs SET name = $1, api_key = $2, base_url = $3, model = $4, updated_at = $5 WHERE id = $6`,
		item.Name, item.APIKey, item.BaseURL, item.Model, time.Now(), item.ID,
	)
	if err != nil {
		logger.L().Error("更新模型配置失败", zap.Int("id", item.ID), zap.Error(err))
	}
	return err
}

// Delete 删除模型配置
func (r *LLMConfigRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM llm_configs WHERE id = $1", id)
	if err != nil {
		logger.L().Error("删除模型配置失败", zap.Int("id", id), zap.Error(err))
	}
	return err
}

// SetActive 设为激活模型（事务内取消其他所有激活）
func (r *LLMConfigRepo) SetActive(ctx context.Context, id int) error {
	// 先取消所有激活
	_, err := r.db.Exec(ctx, "UPDATE llm_configs SET is_active = FALSE, updated_at = $1 WHERE is_active = TRUE", time.Now())
	if err != nil {
		logger.L().Error("取消激活模型失败", zap.Error(err))
		return err
	}
	// 再激活指定模型
	_, err = r.db.Exec(ctx, "UPDATE llm_configs SET is_active = TRUE, updated_at = $1 WHERE id = $2", time.Now(), id)
	if err != nil {
		logger.L().Error("激活模型失败", zap.Int("id", id), zap.Error(err))
	}
	return err
}
