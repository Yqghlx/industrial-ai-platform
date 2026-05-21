-- RBAC tables migration
-- Creates roles, permissions, and user_roles tables for Role-Based Access Control

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    tenant_id VARCHAR(255),
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Role-Permissions association table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Roles association table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255),
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes for RBAC tables
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id ON user_roles(tenant_id);

-- Add tenant_id columns to existing tables if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE blackbox_records ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE reports ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);
ALTER TABLE agent_task_logs ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(255);

-- Create indexes for tenant_id columns
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_tenant_id ON devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_tenant_id ON device_telemetry(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_id ON alert_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_id ON alerts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_tenant_id ON work_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_tenant_id ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_blackbox_records_tenant_id ON blackbox_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reports_tenant_id ON reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agent_task_logs_tenant_id ON agent_task_logs(tenant_id);