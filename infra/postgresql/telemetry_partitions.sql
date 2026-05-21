-- ============================================
-- 遥测数据分区表 Schema
-- Industrial AI Platform - 时间分区
-- ============================================

-- ============================================
-- 分区父表定义
-- ============================================

-- 遥测数据表 (按时间分区)
CREATE TABLE IF NOT EXISTS telemetry_data (
    telemetry_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    metric_type VARCHAR(32) NOT NULL,
    metric_value FLOAT NOT NULL,
    metric_unit VARCHAR(16),
    quality VARCHAR(16) DEFAULT 'good',
    source VARCHAR(32),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- ============================================
-- 全局索引 (在父表上创建)
-- ============================================

-- 设备索引
CREATE INDEX IF NOT EXISTS idx_telemetry_device ON telemetry_data(device_id);

-- 租户索引
CREATE INDEX IF NOT EXISTS idx_telemetry_tenant ON telemetry_data(tenant_id);

-- 指标类型索引
CREATE INDEX IF NOT EXISTS idx_telemetry_type ON telemetry_data(metric_type);

-- 质量索引
CREATE INDEX IF NOT EXISTS idx_telemetry_quality ON telemetry_data(quality);

-- 组合索引 (租户+设备+时间)
CREATE INDEX IF NOT EXISTS idx_telemetry_tenant_device_time 
    ON telemetry_data(tenant_id, device_id, timestamp DESC);

-- ============================================
-- 2026 年月度分区
-- ============================================

-- 2026 年 1 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_01 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

-- 2026 年 2 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_02 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

-- 2026 年 3 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_03 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

-- 2026 年 4 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_04 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

-- 2026 年 5 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_05 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

-- 2026 年 6 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_06 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-06-01 00:00:00') TO ('2026-07-01 00:00:00');

-- 2026 年 7 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_07 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-07-01 00:00:00') TO ('2026-08-01 00:00:00');

-- 2026 年 8 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_08 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-08-01 00:00:00') TO ('2026-09-01 00:00:00');

-- 2026 年 9 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_09 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-09-01 00:00:00') TO ('2026-10-01 00:00:00');

-- 2026 年 10 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_10 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-10-01 00:00:00') TO ('2026-11-01 00:00:00');

-- 2026 年 11 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_11 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-11-01 00:00:00') TO ('2026-12-01 00:00:00');

-- 2026 年 12 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2026_12 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2026-12-01 00:00:00') TO ('2027-01-01 00:00:00');

-- ============================================
-- 2027 年月度分区 (预创建)
-- ============================================

-- 2027 年 1 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2027_01 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2027-01-01 00:00:00') TO ('2027-02-01 00:00:00');

-- 2027 年 2 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2027_02 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2027-02-01 00:00:00') TO ('2027-03-01 00:00:00');

-- 2027 年 3 月分区
CREATE TABLE IF NOT EXISTS telemetry_data_2027_03 
    PARTITION OF telemetry_data
    FOR VALUES FROM ('2027-03-01 00:00:00') TO ('2027-04-01 00:00:00');

-- ============================================
-- 分区子表索引 (可选，提高特定分区性能)
-- ============================================

-- 为每个分区创建本地索引
-- 这些索引仅在特定分区上有效

-- 2026年1月分区索引
CREATE INDEX IF NOT EXISTS idx_telemetry_2026_01_time 
    ON telemetry_data_2026_01(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_2026_01_device_time 
    ON telemetry_data_2026_01(device_id, timestamp DESC);

-- 其他分区的索引会在数据量增加后按需创建

-- ============================================
-- 分区管理函数
-- ============================================

-- 创建新分区的函数
CREATE OR REPLACE FUNCTION create_telemetry_partition(
    partition_year INT,
    partition_month INT
)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date TEXT;
    end_date TEXT;
    next_year INT;
    next_month INT;
BEGIN
    -- 计算分区名称和日期范围
    partition_name := 'telemetry_data_' || partition_year || '_' || LPAD(partition_month::TEXT, 2, '0');
    
    -- 计算开始日期
    start_date := partition_year || '-' || LPAD(partition_month::TEXT, 2, '0') || '-01 00:00:00';
    
    -- 计算结束日期 (下月第一天)
    IF partition_month = 12 THEN
        next_year := partition_year + 1;
        next_month := 1;
    ELSE
        next_year := partition_year;
        next_month := partition_month + 1;
    END IF;
    
    end_date := next_year || '-' || LPAD(next_month::TEXT, 2, '0') || '-01 00:00:00';
    
    -- 创建分区
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF telemetry_data 
         FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
    
    -- 记录创建日志
    RAISE NOTICE 'Created partition % for range % to %', 
        partition_name, start_date, end_date;
END;
$$ LANGUAGE plpgsql;

-- 删除旧分区的函数
CREATE OR REPLACE FUNCTION drop_old_telemetry_partition(
    retention_months INT
)
RETURNS VOID AS $$
DECLARE
    partition_record RECORD;
    cutoff_date TIMESTAMPTZ;
BEGIN
    -- 计算截止日期
    cutoff_date := NOW() - (retention_months || ' months')::INTERVAL;
    
    -- 查询并删除旧分区
    FOR partition_record IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE tablename LIKE 'telemetry_data_%' 
        AND schemaname = 'public'
    LOOP
        -- 检查分区是否过期
        -- 这里简化处理，实际应该解析分区名称判断日期范围
        -- 仅保留最近 retention_months 个月的分区
        
        RAISE NOTICE 'Checking partition % for retention', partition_record.tablename;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 分区统计视图
-- ============================================

-- 分区大小统计视图
CREATE OR REPLACE VIEW telemetry_partition_stats AS
SELECT 
    tablename AS partition_name,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
FROM pg_tables 
WHERE tablename LIKE 'telemetry_data_%'
ORDER BY tablename;

-- 分区行数统计视图
CREATE OR REPLACE VIEW telemetry_partition_counts AS
SELECT 
    tablename AS partition_name,
    (SELECT count(*) FROM public.telemetry_data_2026_01) AS row_count_01,
    (SELECT count(*) FROM public.telemetry_data_2026_02) AS row_count_02,
    (SELECT count(*) FROM public.telemetry_data_2026_03) AS row_count_03,
    (SELECT count(*) FROM public.telemetry_data_2026_04) AS row_count_04
FROM pg_tables 
WHERE tablename = 'telemetry_data'
LIMIT 1;

-- ============================================
-- 分区查询示例
-- ============================================

-- 查询单月数据 (分区剪枝)
-- SELECT * FROM telemetry_data
-- WHERE timestamp >= '2026-01-01' AND timestamp < '2026-02-01';

-- 查询单日数据 (高效分区查询)
-- SELECT * FROM telemetry_data
-- WHERE timestamp >= '2026-01-15 00:00:00' 
-- AND timestamp < '2026-01-16 00:00:00';

-- 查询设备数据 (分区剪枝 + 设备索引)
-- SELECT * FROM telemetry_data
-- WHERE device_id = 'device-001'
-- AND timestamp >= '2026-01-01' AND timestamp < '2026-02-01';

-- 查询租户数据 (分区剪枝 + 租户索引)
-- SELECT * FROM telemetry_data
-- WHERE tenant_id = 'tenant-001'
-- AND timestamp >= NOW() - INTERVAL '7 days';

-- ============================================
-- 数据归档策略
-- ============================================

-- 归档表 (压缩存储)
CREATE TABLE IF NOT EXISTS telemetry_archive (
    telemetry_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    metric_type VARCHAR(32) NOT NULL,
    metric_value FLOAT NOT NULL,
    metric_unit VARCHAR(16),
    quality VARCHAR(16),
    source VARCHAR(32),
    metadata JSONB,
    archived_at TIMESTAMPTZ DEFAULT NOW()
);

-- 归档索引
CREATE INDEX IF NOT EXISTS idx_archive_tenant ON telemetry_archive(tenant_id);
CREATE INDEX IF NOT EXISTS idx_archive_timestamp ON telemetry_archive(timestamp);

-- ============================================
-- 注释
-- ============================================

COMMENT ON TABLE telemetry_data IS 'Industrial AI Platform 遥测数据表 (按月分区)';
COMMENT ON COLUMN telemetry_data.telemetry_id IS '遥测数据唯一 ID';
COMMENT ON COLUMN telemetry_data.device_id IS '设备 ID';
COMMENT ON COLUMN telemetry_data.tenant_id IS '租户 ID';
COMMENT ON COLUMN telemetry_data.timestamp IS '遥测时间戳 (分区键)';
COMMENT ON COLUMN telemetry_data.metric_type IS '指标类型 (temperature/pressure/humidity)';
COMMENT ON COLUMN telemetry_data.metric_value IS '指标值';
COMMENT ON COLUMN telemetry_data.metric_unit IS '指标单位 (°C/bar/%)';
COMMENT ON COLUMN telemetry_data.quality IS '数据质量 (good/bad/uncertain)';
COMMENT ON COLUMN telemetry_data.source IS '数据来源 (sensor/gateway/api)';
COMMENT ON COLUMN telemetry_data.metadata IS '额外元数据';