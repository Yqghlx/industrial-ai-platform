-- 补充缺失的性能索引
-- 这些索引原计划在 migration 4 中添加，但因迁移版本已被标记为 applied 而未生效

-- 告警表：设备和规则组合索引（告警规则评估冷却查询使用）
CREATE INDEX IF NOT EXISTS idx_alerts_device_rule_triggered
    ON alerts (device_id, rule_id, triggered_at DESC);

-- 工单表：设备 ID 索引（ROI 报告按设备聚合查询使用）
CREATE INDEX IF NOT EXISTS idx_work_orders_device_id
    ON work_orders (device_id);
