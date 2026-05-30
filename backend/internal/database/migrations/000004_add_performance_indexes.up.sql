-- 性能优化索引迁移
-- Version: 000004
-- Name: add_performance_indexes
-- 修复：移除引用不存在列/表的索引，添加实际需要的索引

-- ====================
-- 设备表索引优化
-- ====================
-- 多租户设备状态查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_status
ON devices (tenant_id, status);

-- 列表排序查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_created
ON devices (tenant_id, created_at DESC);

-- 设备名称查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_name
ON devices (tenant_id, name);

-- ====================
-- 遥测数据索引优化 (最高频查询)
-- ====================
-- 设备遥测查询：修复 timestamp → time（实际列名）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_device_time
ON device_telemetry (device_id, time DESC);

-- 遥测多租户查询（tenant_id 已在 000003 中添加）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_tenant_device_time
ON device_telemetry (tenant_id, device_id, time DESC);

-- ====================
-- 告警表索引优化
-- ====================
-- 告警列表排序查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_status_created
ON alerts (status, created_at DESC);

-- 多租户设备告警查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_tenant_device
ON alerts (tenant_id, device_id);

-- 告警严重程度查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_severity_status
ON alerts (severity, status);

-- 告警时间范围查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_created
ON alerts (created_at DESC);

-- 告警 cooldown 查询：WHERE device_id = ? AND rule_id = ? AND triggered_at > ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_device_rule_triggered
ON alerts (device_id, rule_id, triggered_at DESC);

-- ====================
-- 用户表索引优化
-- ====================
-- 登录查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_tenant_username
ON users (tenant_id, username);

-- 用户状态查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_tenant_status
ON users (tenant_id, status);

-- ====================
-- 告警规则索引优化
-- ====================
-- 启用规则查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_enabled
ON alert_rules (enabled);

-- 租户规则查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_tenant_enabled
ON alert_rules (tenant_id, enabled);

-- ====================
-- 工单索引优化
-- ====================
-- 工单列表排序查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_status_created
ON work_orders (status, created_at DESC);

-- 租户工单查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_tenant_status
ON work_orders (tenant_id, status);

-- 工单优先级查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_priority_created
ON work_orders (priority DESC, created_at DESC);

-- 工单按设备查询（ROI 报告等场景）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_device_id
ON work_orders (device_id);

-- ====================
-- 通知索引优化
-- ====================
-- 未读通知查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_user_unread
ON notifications (user_id, read, created_at DESC);

-- 租户通知查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_tenant_time
ON notifications (tenant_id, created_at DESC);
