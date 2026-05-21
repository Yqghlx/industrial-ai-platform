-- 性能优化索引回滚
-- Version: 000004
-- Name: add_performance_indexes

-- 设备表索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_devices_tenant_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_devices_tenant_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_devices_last_heartbeat;
DROP INDEX CONCURRENTLY IF EXISTS idx_devices_tenant_name;

-- 遥测数据索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_telemetry_device_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_telemetry_tenant_device_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_telemetry_tenant_type_time;

-- 告警表索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_alerts_status_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_alerts_tenant_device;
DROP INDEX CONCURRENTLY IF EXISTS idx_alerts_severity_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_alerts_created;

-- 用户表索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_users_tenant_username;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_last_login;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_tenant_status;

-- 告警规则索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_alert_rules_enabled;
DROP INDEX CONCURRENTLY IF EXISTS idx_alert_rules_tenant_enabled;
DROP INDEX CONCURRENTLY IF EXISTS idx_alert_rules_device;

-- 工单索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_work_orders_status_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_work_orders_tenant_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_work_orders_priority_created;

-- AI 查询日志索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_ai_queries_user_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_ai_queries_tenant_time;

-- 通知索引删除
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_user_unread;
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_tenant_time;