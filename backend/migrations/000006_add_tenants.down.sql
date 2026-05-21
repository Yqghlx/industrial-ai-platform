-- Rollback multi-tenancy migration

-- 回滚迁移
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE devices DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE alert_rules DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE work_orders DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE blackbox_records DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE reports DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE device_telemetry DROP COLUMN IF EXISTS tenant_id;

DROP TABLE IF EXISTS tenants;