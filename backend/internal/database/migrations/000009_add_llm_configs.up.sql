-- 多模型配置表：支持多个大模型配置并切换当前使用哪个
CREATE TABLE IF NOT EXISTS llm_configs (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    api_key    TEXT NOT NULL DEFAULT '',
    base_url   TEXT NOT NULL DEFAULT '',
    model      VARCHAR(100) NOT NULL DEFAULT '',
    is_active  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 将 system_config 中的现有配置迁移到新表
INSERT INTO llm_configs (name, api_key, base_url, model, is_active)
SELECT '默认配置',
       COALESCE((SELECT value FROM system_config WHERE key='llm_api_key'), ''),
       COALESCE((SELECT value FROM system_config WHERE key='llm_base_url'), 'https://open.bigmodel.cn/api/paas/v4'),
       COALESCE((SELECT value FROM system_config WHERE key='llm_model'), 'glm-4-flash'),
       TRUE
WHERE NOT EXISTS (SELECT 1 FROM llm_configs);
