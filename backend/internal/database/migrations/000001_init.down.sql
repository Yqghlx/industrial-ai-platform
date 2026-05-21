-- Rollback initial schema migration
-- Drops all tables in reverse order to handle foreign key constraints

-- Drop indexes first
DROP INDEX IF EXISTS idx_alerts_status;
DROP INDEX IF EXISTS idx_alerts_device_id;
DROP INDEX IF EXISTS idx_telemetry_time;
DROP INDEX IF EXISTS idx_telemetry_device_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS agent_task_logs;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS blackbox_records;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS work_orders;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS alert_rules;
DROP TABLE IF EXISTS device_telemetry;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS users;