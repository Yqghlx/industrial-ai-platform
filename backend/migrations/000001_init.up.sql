-- Initial schema migration
-- Creates all core tables for the industrial AI platform

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    location VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'online',
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Device telemetry table
CREATE TABLE IF NOT EXISTS device_telemetry (
    id BIGSERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL REFERENCES devices(id),
    time TIMESTAMP NOT NULL DEFAULT NOW(),
    temperature DOUBLE PRECISION,
    pressure DOUBLE PRECISION,
    vibration DOUBLE PRECISION,
    humidity DOUBLE PRECISION,
    power DOUBLE PRECISION,
    status VARCHAR(50) NOT NULL DEFAULT 'normal',
    message TEXT
);

-- Alert rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(100) NOT NULL DEFAULT '*',
    metric VARCHAR(50) NOT NULL,
    operator VARCHAR(10) NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    actions TEXT NOT NULL DEFAULT '[{"type": "notification"}]',
    enabled BOOLEAN NOT NULL DEFAULT true,
    cooldown_sec INTEGER NOT NULL DEFAULT 300,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES alert_rules(id),
    device_id VARCHAR(255) REFERENCES devices(id),
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    triggered_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP
);

-- Work orders table
CREATE TABLE IF NOT EXISTS work_orders (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    device_id VARCHAR(255) REFERENCES devices(id),
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    assigned_to INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    device_id VARCHAR(255) REFERENCES devices(id),
    read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Blackbox records table
CREATE TABLE IF NOT EXISTS blackbox_records (
    id BIGSERIAL PRIMARY KEY,
    device_id VARCHAR(255) REFERENCES devices(id),
    trigger_type VARCHAR(100) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    summary TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Reports table
CREATE TABLE IF NOT EXISTS reports (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    device_id VARCHAR(255) REFERENCES devices(id),
    content TEXT NOT NULL,
    generated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Agent task logs table
CREATE TABLE IF NOT EXISTS agent_task_logs (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(100) NOT NULL,
    query TEXT NOT NULL,
    response TEXT,
    agent VARCHAR(100),
    executed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_telemetry_device_id ON device_telemetry(device_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_time ON device_telemetry(time);
CREATE INDEX IF NOT EXISTS idx_alerts_device_id ON alerts(device_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);