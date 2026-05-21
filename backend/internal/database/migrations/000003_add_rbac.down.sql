-- RBAC tables rollback migration
-- Drops roles, permissions, and user_roles tables

-- Drop indexes first
DROP INDEX IF EXISTS idx_roles_tenant_id;
DROP INDEX IF EXISTS idx_roles_name;
DROP INDEX IF EXISTS idx_permissions_resource_action;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_tenant_id;

-- Drop tenant_id indexes
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP INDEX IF EXISTS idx_devices_tenant_id;
DROP INDEX IF EXISTS idx_telemetry_tenant_id;
DROP INDEX IF EXISTS idx_alert_rules_tenant_id;
DROP INDEX IF EXISTS idx_alerts_tenant_id;
DROP INDEX IF EXISTS idx_work_orders_tenant_id;
DROP INDEX IF EXISTS idx_notifications_tenant_id;
DROP INDEX IF EXISTS idx_blackbox_records_tenant_id;
DROP INDEX IF EXISTS idx_reports_tenant_id;
DROP INDEX IF EXISTS idx_agent_task_logs_tenant_id;

-- Drop association tables
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;

-- Drop RBAC tables
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

-- Remove tenant_id columns from existing tables (optional, comment out if you want to keep)
-- ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE devices DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE device_telemetry DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE alert_rules DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE alerts DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE work_orders DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE notifications DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE blackbox_records DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE reports DROP COLUMN IF EXISTS tenant_id;
-- ALTER TABLE agent_task_logs DROP COLUMN IF EXISTS tenant_id;