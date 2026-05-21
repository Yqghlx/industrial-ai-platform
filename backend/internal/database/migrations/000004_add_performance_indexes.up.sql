-- 性能优化索引迁移
-- Version: 000004
-- Name: add_performance_indexes

-- 设备表索引优化
-- 高频查询: WHERE tenant_id = ? AND status = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_status 
ON devices (tenant_id, status);

-- 列表排序查询: ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_created 
ON devices (tenant_id, created_at DESC);

-- 设备心跳查询: WHERE last_heartbeat < ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_last_heartbeat 
ON devices (last_heartbeat DESC);

-- 设备名称查询: WHERE tenant_id = ? AND name = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_devices_tenant_name 
ON devices (tenant_id, name);

-- 遥测数据索引优化 (最高频查询)
-- 设备遥测查询: WHERE device_id = ? ORDER BY timestamp DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_device_time 
ON device_telemetry (device_id, timestamp DESC);

-- 多租户遥测查询: WHERE tenant_id = ? AND device_id = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_tenant_device_time 
ON device_telemetry (tenant_id, device_id, timestamp DESC);

-- 遥测数据类型查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_tenant_type_time 
ON device_telemetry (tenant_id, telemetry_type, timestamp DESC);

-- 告警表索引优化
-- 告警列表查询: WHERE status = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_status_created 
ON alerts (status, created_at DESC);

-- 设备告警查询: WHERE tenant_id = ? AND device_id = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_tenant_device 
ON alerts (tenant_id, device_id);

-- 告警严重程度查询: WHERE severity = ? AND status = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_severity_status 
ON alerts (severity, status);

-- 告警时间范围查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_created 
ON alerts (created_at DESC);

-- 用户表索引优化
-- 登录查询: WHERE tenant_id = ? AND username = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_tenant_username 
ON users (tenant_id, username);

-- 用户最后登录查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_last_login 
ON users (last_login DESC);

-- 用户状态查询: WHERE tenant_id = ? AND status = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_tenant_status 
ON users (tenant_id, status);

-- 告警规则索引优化
-- 启用规则查询: WHERE enabled = true
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_enabled 
ON alert_rules (enabled);

-- 租户规则查询: WHERE tenant_id = ? AND enabled = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_tenant_enabled 
ON alert_rules (tenant_id, enabled);

-- 规则触发设备查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alert_rules_device 
ON alert_rules (device_id);

-- 工单索引优化
-- 工单列表查询: WHERE status = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_status_created 
ON work_orders (status, created_at DESC);

-- 租户工单查询: WHERE tenant_id = ? AND status = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_tenant_status 
ON work_orders (tenant_id, status);

-- 工单优先级查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_orders_priority_created 
ON work_orders (priority DESC, created_at DESC);

-- AI 查询日志索引
-- 用户查询历史: WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ai_queries_user_time 
ON ai_queries (user_id, created_at DESC);

-- 租户查询统计: WHERE tenant_id = ? AND created_at > ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ai_queries_tenant_time 
ON ai_queries (tenant_id, created_at DESC);

-- 通知索引优化
-- 未读通知查询: WHERE user_id = ? AND read = false
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_user_unread 
ON notifications (user_id, read, created_at DESC);

-- 租户通知查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_tenant_time 
ON notifications (tenant_id, created_at DESC);