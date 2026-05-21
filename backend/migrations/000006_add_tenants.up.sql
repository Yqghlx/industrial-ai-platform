-- Add multi-tenancy support
-- Creates tenants table and adds tenant_id to existing tables

-- 创建租户表
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    max_devices INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 为现有表添加 tenant_id
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
ALTER TABLE blackbox_records ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
ALTER TABLE reports ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_tenant ON devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant ON alert_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_tenant ON work_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_tenant ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_blackbox_records_tenant ON blackbox_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reports_tenant ON reports(tenant_id);

-- 创建默认租户
INSERT INTO tenants (id, name, slug, plan, max_devices)
VALUES ('default-tenant-0000-0000-0000-000000000001', 'Default Tenant', 'default', 'enterprise', 1000)
ON CONFLICT (id) DO NOTHING;

-- 更新现有数据关联到默认租户
UPDATE users SET tenant_id = 'default-tenant-0000-0000-0000-000000000001' WHERE tenant_id IS NULL;
UPDATE devices SET tenant_id = 'default-tenant-0000-0000-0000-000000000001' WHERE tenant_id IS NULL;
UPDATE alert_rules SET tenant_id = 'default-tenant-0000-0000-0000-000000000001' WHERE tenant_id IS NULL;
UPDATE work_orders SET tenant_id = 'default-tenant-0000-0000-0000-000000000001' WHERE tenant_id IS NULL;

-- 设备遥测表 (TimescaleDB)
ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
CREATE INDEX IF NOT EXISTS idx_device_telemetry_tenant ON device_telemetry(tenant_id);