-- TimescaleDB setup migration
-- Creates hypertable for device_telemetry if TimescaleDB is available
-- This migration is optional and will be skipped if TimescaleDB is not installed
-- 
-- NOTE: TimescaleDB functions (create_hypertable, add_retention_policy) are only
-- available when the TimescaleDB extension is installed. On standard PostgreSQL,
-- this migration will fail. To use TimescaleDB features:
-- 1. Install TimescaleDB extension: CREATE EXTENSION timescaledb;
-- 2. Run this migration manually
--
-- For standard PostgreSQL, device_telemetry will work as a regular table.

-- Empty migration (no-op for standard PostgreSQL)
SELECT 1;