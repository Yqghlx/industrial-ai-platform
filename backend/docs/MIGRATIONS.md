# Database Migrations Guide

This directory contains database migration files for the Industrial AI Platform.

## Overview

The platform uses a custom migration system compatible with golang-migrate format. Migrations are automatically applied on server startup.

## Directory Structure

```
backend/
├── migrations/
│   ├── 000001_init.up.sql          # Initial schema creation
│   ├── 000001_init.down.sql        # Initial schema rollback
│   ├── 000002_timescaledb.up.sql   # TimescaleDB setup
│   └── 000002_timescaledb.down.sql # TimescaleDB rollback
├── internal/
│   └── database/
│       └── migrate.go              # Migration engine
└── docs/
    └── MIGRATIONS.md                # This file
```

## Migration Files

### Naming Convention

Migration files follow the format: `{version}_{name}.{direction}.sql`

- `version`: Sequential number (e.g., 000001, 000002)
- `name`: Descriptive name (e.g., init, add_users_table)
- `direction`: Either `up` (apply) or `down` (rollback)

Examples:
- `000001_init.up.sql` - Creates initial schema
- `000001_init.down.sql` - Drops initial schema
- `000002_add_email_index.up.sql` - Adds email index
- `000002_add_email_index.down.sql` - Removes email index

## Usage

### Automatic Migration (Recommended)

Migrations are automatically applied when the server starts:

```go
import "github.com/industrial-ai/platform/internal/database"

func main() {
    // Connect to database
    db, _ := sql.Open("postgres", databaseURL)
    
    // Run migrations automatically
    migrator := database.NewMigrator(db)
    if err := migrator.Up(context.Background()); err != nil {
        log.Printf("Migration error: %v", err)
    }
}
```

### Manual Migration

For manual control, use the Migrator directly:

```go
migrator := database.NewMigrator(db)

// Apply all pending migrations
migrator.Up(context.Background())

// Rollback last migration
migrator.Down(context.Background())

// Check migration status
migrations, _ := migrator.Status(context.Background())
for _, m := range migrations {
    applied := "pending"
    if m.AppliedAt != nil {
        applied = "applied"
    }
    fmt.Printf("%d: %s [%s]\n", m.Version, m.Name, applied)
}

// Reset database (rollback all, then apply all)
migrator.Reset(context.Background())
```

## Current Migrations

### 000001_init

**Purpose:** Creates the core database schema

**Tables:**
- `users` - User accounts and authentication
- `devices` - IoT device registry
- `device_telemetry` - Time-series sensor data
- `alert_rules` - Alert configuration
- `alerts` - Active and historical alerts
- `work_orders` - Maintenance work orders
- `notifications` - User notifications
- `blackbox_records` - Incident recording
- `reports` - Generated reports
- `agent_task_logs` - AI agent query logs

**Indexes:**
- `idx_telemetry_device_id` - Fast device telemetry lookup
- `idx_telemetry_time` - Time-range queries
- `idx_alerts_device_id` - Device-specific alerts
- `idx_alerts_status` - Active alert queries

### 000002_timescaledb

**Purpose:** Configures TimescaleDB for time-series optimization

**Features:**
- Converts `device_telemetry` to hypertable
- Adds compression policy (7-day chunks)
- Adds retention policy (90-day retention)

**Note:** This migration is optional. If TimescaleDB is not installed, the migration will be skipped and the system will use standard PostgreSQL tables.

## Migration Tracking

The system maintains a `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Best Practices

### Creating New Migrations

1. **Increment version number properly**
   ```bash
   # Check latest migration version
   ls migrations/ | sort -r | head -1
   
   # Create new migration files
   touch migrations/000003_add_feature.up.sql
   touch migrations/000003_add_feature.down.sql
   ```

2. **Always create both up and down files**
   - Every `up.sql` should have a corresponding `down.sql`
   - `down.sql` should reverse all changes from `up.sql`

3. **Use IF NOT EXISTS / IF EXISTS**
   ```sql
   -- Good: Safe for repeated execution
   CREATE TABLE IF NOT EXISTS users (...);
   DROP TABLE IF EXISTS users;
   
   -- Bad: Will fail if already exists
   CREATE TABLE users (...);
   ```

4. **Test migrations in both directions**
   ```go
   migrator.Down(ctx)  // Rollback
   migrator.Up(ctx)    // Re-apply
   ```

### Migration Guidelines

- **Atomic operations**: Each migration should be a single logical change
- **Backward compatible**: When possible, make migrations backward compatible
- **No data loss**: Include data migration if needed to prevent data loss
- **Test thoroughly**: Test on development data before production

### Example Migration

**000003_add_user_status.up.sql**
```sql
-- Add status column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';

-- Add index for active users queries
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
```

**000003_add_user_status.down.sql**
```sql
-- Remove status column and index
DROP INDEX IF EXISTS idx_users_status;
ALTER TABLE users DROP COLUMN IF EXISTS status;
```

## Troubleshooting

### Migration Failed

**Problem:** Migration fails with error
```
Error: failed to apply migration 1: pq: relation "users" already exists
```

**Solution:** The migration may have partially completed. Check the database state:
```sql
SELECT * FROM schema_migrations ORDER BY version;
```

Manual intervention may be required to fix the database state.

### TimescaleDB Not Available

**Problem:** Migration 000002 fails
```
Error: TimescaleDB not available
```

**Solution:** This is expected if TimescaleDB is not installed. The system will continue to work with standard PostgreSQL tables. To enable TimescaleDB:
```sql
CREATE EXTENSION IF NOT EXISTS timescaledb;
```

### Reset Migrations

To completely reset the migration system:
```go
migrator := database.NewMigrator(db)
migrator.Reset(context.Background())
```

**Warning:** This will delete all data!

## Environment Variables

The migration system uses the same database connection as the main application:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/industrial_ai
```

## See Also

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [TimescaleDB Documentation](https://docs.timescale.com/)
- [golang-migrate](https://github.com/golang-migrate/migrate)