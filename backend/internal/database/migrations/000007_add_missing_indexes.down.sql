-- 回滚：删除 migration 7 添加的索引
DROP INDEX IF EXISTS idx_alerts_device_rule_triggered;
DROP INDEX IF EXISTS idx_work_orders_device_id;
