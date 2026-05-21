-- TimescaleDB setup migration
-- Creates hypertable for device_telemetry if TimescaleDB is available
-- This migration is optional and will be skipped if TimescaleDB is not installed

-- Create hypertable for time-series data
SELECT create_hypertable('device_telemetry', 'time', if_not_exists => TRUE);

-- Add compression policy (compress chunks older than 7 days)
SELECT add_compress_chunk_policy('device_telemetry', INTERVAL '7 days', if_not_exists => TRUE);

-- Add retention policy (keep data for 90 days)
SELECT add_retention_policy('device_telemetry', INTERVAL '90 days', if_not_exists => TRUE);