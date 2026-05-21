# Database Migration Quick Start

## Overview

The Industrial AI Platform uses an embedded migration system that automatically applies database migrations on server startup. This separates schema management from application code.

## Quick Start

### Automatic Migration (Default)

Just start the server - migrations run automatically:

```bash
./server
# or
go run main.go
```

### Manual Migration Control

For more control, use the migration tool directly:

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    "github.com/industrial-ai/platform/internal/database"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "your-database-url")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create migrator
    migrator := database.NewMigrator(db)
    
    // Apply all pending migrations
    if err := migrator.Up(context.Background()); err != nil {
        log.Printf("Migration error: %v", err)
    }
    
    // Check status
    migrations, _ := migrator.Status(context.Background())
    for _, m := range migrations {
        status := "pending"
        if m.AppliedAt != nil {
            status = "applied"
        }
        log.Printf("Migration %d: %s [%s]", m.Version, m.Name, status)
    }
    
    // Rollback last migration (use with caution!)
    // migrator.Down(context.Background())
    
    // Reset database (rollback all, then apply all)
    // migrator.Reset(context.Background())
}
```

## Migration Files

Migrations are stored in `internal/database/migrations/`:

```
internal/database/migrations/
├── 000001_init.up.sql          # Initial schema
├── 000001_init.down.sql        # Rollback initial schema
├── 000002_timescaledb.up.sql   # TimescaleDB setup
└── 000002_timescaledb.down.sql # TimescaleDB rollback
```

## Create a New Migration

1. Create migration files:
```bash
# In backend/internal/database/migrations/
touch 000003_add_feature.up.sql
touch 000003_add_feature.down.sql
```

2. Write the SQL:
```sql
-- 000003_add_feature.up.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login);
```

```sql
-- 000003_add_feature.down.sql
DROP INDEX IF EXISTS idx_users_last_login;
ALTER TABLE users DROP COLUMN IF EXISTS last_login;
```

3. Restart the server to apply the migration.

## Migration Best Practices

1. **Always provide rollback**: Every `.up.sql` should have a corresponding `.down.sql`
2. **Use IF EXISTS/IF NOT EXISTS**: Make migrations idempotent and safe to re-run
3. **Test both directions**: Always test rollback (Down) after applying (Up)
4. **One change per migration**: Keep migrations focused on a single logical change
5. **Never modify applied migrations**: Create a new migration to fix issues

## Current Schema

The database contains the following tables:

- **users** - User accounts and authentication
- **devices** - IoT device registry
- **device_telemetry** - Time-series sensor data (TimescaleDB hypertable if available)
- **alert_rules** - Alert configuration
- **alerts** - Active and historical alerts
- **work_orders** - Maintenance work orders
- **notifications** - User notifications
- **blackbox_records** - Incident recording
- **reports** - Generated reports
- **agent_task_logs** - AI agent query logs

## Full Documentation

For complete documentation, see:
- [Migration System Guide](../docs/MIGRATIONS.md)
- [Migration Files README](../migrations/README.md)