-- ============================================
-- 审计日志表 Schema
-- Industrial AI Platform - 安全审计日志
-- ============================================

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    audit_id VARCHAR(64) PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    event_type VARCHAR(32) NOT NULL,
    event_category VARCHAR(32) NOT NULL,
    severity VARCHAR(16) NOT NULL DEFAULT 'info',
    user_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64),
    ip_address VARCHAR(45) NOT NULL,
    user_agent VARCHAR(256),
    resource_type VARCHAR(64),
    resource_id VARCHAR(64),
    action VARCHAR(16) NOT NULL,
    operation VARCHAR(256) NOT NULL,
    request_id VARCHAR(64),
    trace_id VARCHAR(64),
    before_state JSONB,
    after_state JSONB,
    changes JSONB,
    result VARCHAR(16) NOT NULL,
    error_message TEXT,
    duration_ms FLOAT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- 索引
-- ============================================

-- 时间索引 (最常用查询条件)
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp DESC);

-- 用户索引
CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_logs(user_id);

-- 租户索引
CREATE INDEX IF NOT EXISTS idx_audit_tenant_id ON audit_logs(tenant_id);

-- 事件类型索引
CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_logs(event_type);

-- 分类索引
CREATE INDEX IF NOT EXISTS idx_audit_category ON audit_logs(event_category);

-- 资源索引
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource_type, resource_id);

-- 结果索引
CREATE INDEX IF NOT EXISTS idx_audit_result ON audit_logs(result);

-- IP 地址索引
CREATE INDEX IF NOT EXISTS idx_audit_ip_address ON audit_logs(ip_address);

-- 组合索引 (租户+时间)
CREATE INDEX IF NOT EXISTS idx_audit_tenant_time ON audit_logs(tenant_id, timestamp DESC);

-- 组合索引 (用户+时间)
CREATE INDEX IF NOT EXISTS idx_audit_user_time ON audit_logs(user_id, timestamp DESC);

-- ============================================
-- 分区 (按季度)
-- ============================================

-- 创建分区表 (按季度)
-- 注意: 需要先删除主表,然后创建分区表

-- 2026 Q1 分区
CREATE TABLE IF NOT EXISTS audit_logs_2026_q1 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');

-- 2026 Q2 分区
CREATE TABLE IF NOT EXISTS audit_logs_2026_q2 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-07-01');

-- 2026 Q3 分区
CREATE TABLE IF NOT EXISTS audit_logs_2026_q3 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-07-01') TO ('2026-10-01');

-- 2026 Q4 分区
CREATE TABLE IF NOT EXISTS audit_logs_2026_q4 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-10-01') TO ('2027-01-01');

-- ============================================
-- 数据保留策略
-- ============================================

-- 创建清理旧审计日志的函数
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs()
RETURNS void AS $$
BEGIN
    -- 删除 365 天前的审计日志
    DELETE FROM audit_logs
    WHERE timestamp < NOW() - INTERVAL '365 days';
    
    -- 记录清理日志
    RAISE NOTICE 'Cleanup old audit logs completed';
END;
$$ LANGUAGE plpgsql;

-- 创建定时任务 (需要 pg_cron 扩展)
-- 每日凌晨 3 点执行清理
-- SELECT cron.schedule('cleanup_audit_logs', '0 3 * * *', 'SELECT cleanup_old_audit_logs()');

-- ============================================
-- 审计日志统计视图
-- ============================================

-- 每日审计日志统计视图
CREATE OR REPLACE VIEW audit_daily_stats AS
SELECT
    DATE(timestamp) as date,
    tenant_id,
    COUNT(*) as total_logs,
    COUNT(CASE WHEN result = 'failure' THEN 1 END) as failure_count,
    COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_count,
    AVG(duration_ms) as avg_duration,
    COUNT(DISTINCT user_id) as active_users
FROM audit_logs
GROUP BY DATE(timestamp), tenant_id
ORDER BY date DESC;

-- 每小时审计日志统计视图
CREATE OR REPLACE VIEW audit_hourly_stats AS
SELECT
    DATE_TRUNC('hour', timestamp) as hour,
    tenant_id,
    event_type,
    COUNT(*) as event_count
FROM audit_logs
GROUP BY DATE_TRUNC('hour', timestamp), tenant_id, event_type
ORDER BY hour DESC;

-- 用户审计活动视图
CREATE OR REPLACE VIEW audit_user_activity AS
SELECT
    user_id,
    tenant_id,
    COUNT(*) as total_operations,
    COUNT(CASE WHEN result = 'failure' THEN 1 END) as failure_count,
    MAX(timestamp) as last_activity,
    MIN(timestamp) as first_activity
FROM audit_logs
GROUP BY user_id, tenant_id
ORDER BY total_operations DESC;

-- ============================================
-- 审计日志查询示例
-- ============================================

-- 查询最近 1 小时的审计日志
-- SELECT * FROM audit_logs
-- WHERE timestamp >= NOW() - INTERVAL '1 hour'
-- ORDER BY timestamp DESC;

-- 查询特定用户的审计日志
-- SELECT * FROM audit_logs
-- WHERE user_id = 'user-001'
-- ORDER BY timestamp DESC;

-- 查询特定租户的审计日志
-- SELECT * FROM audit_logs
-- WHERE tenant_id = 'tenant-001'
-- ORDER BY timestamp DESC;

-- 查询失败的审计日志
-- SELECT * FROM audit_logs
-- WHERE result = 'failure'
-- ORDER BY timestamp DESC;

-- 查询安全相关审计日志
-- SELECT * FROM audit_logs
-- WHERE event_category = 'security'
-- ORDER BY timestamp DESC;

-- ============================================
-- 权限设置
-- ============================================

-- 授权审计日志表访问权限
-- GRANT SELECT, INSERT ON audit_logs TO industrial_ai_app;
-- GRANT SELECT ON audit_daily_stats TO industrial_ai_app;
-- GRANT SELECT ON audit_hourly_stats TO industrial_ai_app;
-- GRANT SELECT ON audit_user_activity TO industrial_ai_app;

-- ============================================
-- 注释
-- ============================================

COMMENT ON TABLE audit_logs IS 'Industrial AI Platform 安全审计日志表';
COMMENT ON COLUMN audit_logs.audit_id IS '审计日志唯一 ID';
COMMENT ON COLUMN audit_logs.timestamp IS '审计日志时间戳';
COMMENT ON COLUMN audit_logs.event_type IS '事件类型 (auth.login/data.read 等)';
COMMENT ON COLUMN audit_logs.event_category IS '事件分类 (auth/data/config 等)';
COMMENT ON COLUMN audit_logs.severity IS '严重程度 (info/warning/critical)';
COMMENT ON COLUMN audit_logs.user_id IS '操作用户 ID';
COMMENT ON COLUMN audit_logs.tenant_id IS '租户 ID';
COMMENT ON COLUMN audit_logs.session_id IS '会话 ID';
COMMENT ON COLUMN audit_logs.ip_address IS '客户端 IP 地址';
COMMENT ON COLUMN audit_logs.user_agent IS '用户代理';
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型 (device/user 等)';
COMMENT ON COLUMN audit_logs.resource_id IS '资源 ID';
COMMENT ON COLUMN audit_logs.action IS '操作类型 (read/write/delete)';
COMMENT ON COLUMN audit_logs.operation IS '操作详情';
COMMENT ON COLUMN audit_logs.result IS '操作结果 (success/failure)';
COMMENT ON COLUMN audit_logs.changes IS '变更内容';
COMMENT ON COLUMN audit_logs.before_state IS '操作前状态';
COMMENT ON COLUMN audit_logs.after_state IS '操作后状态';
COMMENT ON COLUMN audit_logs.metadata IS '额外元数据';