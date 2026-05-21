-- Add audit logs table for security auditing
-- Creates comprehensive audit trail for all security-relevant events

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    audit_id VARCHAR(255) UNIQUE NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    user_id VARCHAR(255),
    tenant_id VARCHAR(255),
    session_id VARCHAR(255),
    ip_address VARCHAR(50),
    user_agent TEXT,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    action VARCHAR(50),
    operation VARCHAR(255),
    request_id VARCHAR(255),
    trace_id VARCHAR(255),
    before_state JSONB,
    after_state JSONB,
    changes JSONB,
    result VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    duration_ms DOUBLE PRECISION,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_category ON audit_logs(event_category);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_result ON audit_logs(result);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);

-- Create composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time ON audit_logs(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_time ON audit_logs(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_category_time ON audit_logs(event_category, timestamp DESC);

-- Add comment to table
COMMENT ON TABLE audit_logs IS 'Security audit log for tracking all authentication, authorization, data access, and administrative actions';
COMMENT ON COLUMN audit_logs.audit_id IS 'Unique identifier for the audit event';
COMMENT ON COLUMN audit_logs.event_type IS 'Type of event (e.g., auth.login, data.read, admin.action)';
COMMENT ON COLUMN audit_logs.event_category IS 'Category of event (auth, authz, data, config, system, security)';
COMMENT ON COLUMN audit_logs.severity IS 'Severity level (info, warning, critical)';
COMMENT ON COLUMN audit_logs.before_state IS 'State before change (for update operations)';
COMMENT ON COLUMN audit_logs.after_state IS 'State after change (for update operations)';
COMMENT ON COLUMN audit_logs.changes IS 'Changes made (for update operations)';
COMMENT ON COLUMN audit_logs.result IS 'Result of the operation (success, failure)';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional metadata in JSON format';