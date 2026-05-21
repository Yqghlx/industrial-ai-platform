-- Performance Optimization - Database Indexes
-- Task #7: Performance Optimization
-- Created: 2026-05-19

-- ============================================
-- Devices Table Indexes
-- ============================================

-- Index for created_at DESC (common sort order in list queries)
CREATE INDEX IF NOT EXISTS idx_devices_created_at_desc 
ON devices(created_at DESC);

-- Index for tenant_id (multi-tenant filtering)
CREATE INDEX IF NOT EXISTS idx_devices_tenant_id 
ON devices(tenant_id);

-- Composite index for tenant + created_at (tenant isolation + sorting)
CREATE INDEX IF NOT EXISTS idx_devices_tenant_created 
ON devices(tenant_id, created_at DESC);

-- Index for device type filtering
CREATE INDEX IF NOT EXISTS idx_devices_type 
ON devices(type);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_devices_status 
ON devices(status);

-- ============================================
-- Alerts Table Indexes
-- ============================================

-- Index for triggered_at DESC (time-series queries)
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at_desc 
ON alerts(triggered_at DESC);

-- Index for device_id (device-specific alerts)
CREATE INDEX IF NOT EXISTS idx_alerts_device_id 
ON alerts(device_id);

-- Index for tenant_id (multi-tenant isolation)
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_id 
ON alerts(tenant_id);

-- Composite index for tenant + triggered_at
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_triggered 
ON alerts(tenant_id, triggered_at DESC);

-- Index for severity (filtering by severity level)
CREATE INDEX IF NOT EXISTS idx_alerts_severity 
ON alerts(severity);

-- Index for status (active/resolved/false_positive)
CREATE INDEX IF NOT EXISTS idx_alerts_status 
ON alerts(status);

-- Composite index for device + triggered_at (device timeline)
CREATE INDEX IF NOT EXISTS idx_alerts_device_triggered 
ON alerts(device_id, triggered_at DESC);

-- ============================================
-- Telemetry Table Indexes (TimescaleDB)
-- ============================================

-- Index for device_id + timestamp DESC (device telemetry timeline)
CREATE INDEX IF NOT EXISTS idx_telemetry_device_time_desc 
ON telemetry(device_id, timestamp DESC);

-- Index for tenant_id (multi-tenant isolation)
CREATE INDEX IF NOT EXISTS idx_telemetry_tenant_id 
ON telemetry(tenant_id);

-- Composite index for tenant + device + timestamp
CREATE INDEX IF NOT EXISTS idx_telemetry_tenant_device_time 
ON telemetry(tenant_id, device_id, timestamp DESC);

-- ============================================
-- Work Orders Table Indexes
-- ============================================

-- Index for created_at DESC
CREATE INDEX IF NOT EXISTS idx_work_orders_created_at_desc 
ON work_orders(created_at DESC);

-- Index for tenant_id
CREATE INDEX IF NOT EXISTS idx_work_orders_tenant_id 
ON work_orders(tenant_id);

-- Index for device_id
CREATE INDEX IF NOT EXISTS idx_work_orders_device_id 
ON work_orders(device_id);

-- Index for status
CREATE INDEX IF NOT EXISTS idx_work_orders_status 
ON work_orders(status);

-- Composite index for tenant + status
CREATE INDEX IF NOT EXISTS idx_work_orders_tenant_status 
ON work_orders(tenant_id, status);

-- ============================================
-- Notifications Table Indexes
-- ============================================

-- Index for created_at DESC
CREATE INDEX IF NOT EXISTS idx_notifications_created_at_desc 
ON notifications(created_at DESC);

-- Index for tenant_id
CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id 
ON notifications(tenant_id);

-- Index for is_read
CREATE INDEX IF NOT EXISTS idx_notifications_is_read 
ON notifications(is_read);

-- Composite index for tenant + is_read (unread notifications)
CREATE INDEX IF NOT EXISTS idx_notifications_tenant_unread 
ON notifications(tenant_id, is_read) WHERE is_read = false;

-- ============================================
-- Alert Rules Table Indexes
-- ============================================

-- Index for tenant_id
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_id 
ON alert_rules(tenant_id);

-- Index for enabled (active rules)
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled 
ON alert_rules(enabled);

-- Composite index for tenant + enabled
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_enabled 
ON alert_rules(tenant_id, enabled) WHERE enabled = true;

-- ============================================
-- Users Table Indexes
-- ============================================

-- Index for username (login lookup)
CREATE INDEX IF NOT EXISTS idx_users_username 
ON users(username);

-- Index for tenant_id (multi-tenant)
CREATE INDEX IF NOT EXISTS idx_users_tenant_id 
ON users(tenant_id);

-- Index for role (role-based filtering)
CREATE INDEX IF NOT EXISTS idx_users_role 
ON users(role);

-- ============================================
-- Verification Query
-- ============================================

-- Verify all indexes are created
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE schemaname = 'public'
ORDER BY tablename, indexname;