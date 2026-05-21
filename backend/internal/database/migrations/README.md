# Database Migrations

This directory contains SQL migration files for the Industrial AI Platform.

## Migration Files

| Version | Name | Description |
|---------|------|-------------|
| 000001 | init | Initial database schema |
| 000002 | timescaledb | TimescaleDB hypertable configuration |

## File Structure

```
migrations/
├── 000001_init.up.sql          # Create initial schema
├── 000001_init.down.sql        # Drop initial schema
├── 000002_timescaledb.up.sql   # Setup TimescaleDB
└── 000002_timescaledb.down.sql # Teardown TimescaleDB
```

## How It Works

Migrations are automatically loaded from the `migrations/` subdirectory and applied in version order.

For detailed documentation, see [../docs/MIGRATIONS.md](../docs/MIGRATIONS.md).

## Adding New Migrations

1. Create two files with the next version number:
   ```bash
   touch 000003_your_feature.up.sql
   touch 000003_your_feature.down.sql
   ```

2. Write the migration SQL:
   - `*.up.sql` - Apply changes
   - `*.down.sql` - Rollback changes

3. Test the migration:
   ```go
   migrator.Down(ctx)  // Test rollback
   migrator.Up(ctx)    // Test apply
   ```