-- 系统配置表，用于存储可动态修改的配置项
CREATE TABLE IF NOT EXISTS system_config (
    key        VARCHAR(100) PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    category   VARCHAR(50) NOT NULL DEFAULT 'general',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 初始化 LLM 默认配置（仅当记录不存在时插入）
INSERT INTO system_config (key, value, category) VALUES
    ('llm_api_key', '', 'llm'),
    ('llm_base_url', 'https://open.bigmodel.cn/api/paas/v4', 'llm'),
    ('llm_model', 'glm-4-flash', 'llm')
ON CONFLICT (key) DO NOTHING;
