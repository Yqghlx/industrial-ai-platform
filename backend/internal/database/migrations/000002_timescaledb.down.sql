-- Rollback TimescaleDB setup
-- Note: This will remove the hypertable but keep the regular table

-- Remove retention policy
SELECT remove_retention_policy('device_telemetry', if_exists => TRUE);

-- Remove compression policy
SELECT remove_compress_chunk_policy('device_telemetry', if_exists => TRUE);

-- Note: TimescaleDB doesn't provide a simple way to convert hypertable back to regular table
-- The table remains as hypertable but policies are removed