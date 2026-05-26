// Package database provides database migration utilities
// Updated: 2026-05-26 - TimescaleDB migration made optional for standard PostgreSQL
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt *string
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// ensureMigrationTable creates the migrations tracking table if it doesn't exist
func (m *Migrator) ensureMigrationTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[int]string, error) {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure migration table: %w", err)
	}

	query := "SELECT version, name FROM schema_migrations ORDER BY version"
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]string)
	for rows.Next() {
		var version int
		var name string
		if err := rows.Scan(&version, &name); err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		applied[version] = name
	}

	return applied, rows.Err()
}

// loadMigrations loads migrations from embedded filesystem
func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Group files by version
	migrationMap := make(map[int]*Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse filename: 000001_name.up.sql or 000001_name.down.sql
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if migrationMap[version] == nil {
			migrationMap[version] = &Migration{
				Version: version,
				Name:    strings.Join(parts[1:], "_"),
			}
		}

		content, err := fs.ReadFile(migrationsFS, "migrations/"+name)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		if strings.Contains(name, ".up.sql") {
			migrationMap[version].UpSQL = string(content)
		} else if strings.Contains(name, ".down.sql") {
			migrationMap[version].DownSQL = string(content)
		}
	}

	// Convert map to sorted slice
	var migrations []Migration
	for _, m := range migrationMap {
		// Clean up the name
		m.Name = strings.TrimSuffix(m.Name, ".up")
		m.Name = strings.TrimSuffix(m.Name, ".down")
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(migrations) == 0 {
		log.Println("No migrations found")
		return nil
	}

	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; exists {
			log.Printf("Migration %d already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Name)

		if migration.UpSQL == "" {
			return fmt.Errorf("migration %d has no up SQL", migration.Version)
		}

		// Execute each migration in its own transaction
		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
		}

		// Execute migration SQL
		if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
			tx.Rollback()
			// TimescaleDB migration (version 2) is optional - continue on failure
			if migration.Version == 2 {
				log.Printf("Migration %d failed (optional TimescaleDB), skipping: %v", migration.Version, err)
				// Mark as skipped in a new transaction
				recordTx, err := m.db.BeginTx(ctx, nil)
				if err == nil {
					query := "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)"
					if _, err := recordTx.ExecContext(ctx, query, migration.Version, migration.Name+"_skipped"); err != nil {
						log.Printf("Failed to record skipped migration %d: %v", migration.Version, err)
					}
					recordTx.Commit()
				}
				continue
			}
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Record migration
		query := "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)"
		if _, err := tx.ExecContext(ctx, query, migration.Version, migration.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		log.Printf("Migration %d applied successfully", migration.Version)
	}

	log.Println("All migrations applied successfully")
	return nil
}

// Down rolls back the most recent migration
func (m *Migrator) Down(ctx context.Context) error {
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Find the latest applied migration
	var latestVersion int
	for version := range applied {
		if version > latestVersion {
			latestVersion = version
		}
	}

	if latestVersion == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Find the migration
	var migration *Migration
	for i, m := range migrations {
		if m.Version == latestVersion {
			migration = &migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %d not found", latestVersion)
	}

	if migration.DownSQL == "" {
		return fmt.Errorf("migration %d has no down SQL", migration.Version)
	}

	log.Printf("Rolling back migration %d: %s", migration.Version, migration.Name)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.ExecContext(ctx, migration.DownSQL); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
	}

	// Remove migration record
	query := "DELETE FROM schema_migrations WHERE version = $1"
	if _, err := tx.ExecContext(ctx, query, migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("Migration %d rolled back successfully", migration.Version)
	return nil
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) ([]Migration, error) {
	migrations, err := loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for i := range migrations {
		if name, exists := applied[migrations[i].Version]; exists {
			migrations[i].AppliedAt = &name // Store applied status
		}
	}

	return migrations, nil
}

// Reset rolls back all migrations and re-applies them
func (m *Migrator) Reset(ctx context.Context) error {
	log.Println("Resetting database...")

	// Keep rolling back until no more migrations
	for {
		applied, err := m.getAppliedMigrations(ctx)
		if err != nil {
			return err
		}

		if len(applied) == 0 {
			break
		}

		if err := m.Down(ctx); err != nil {
			return err
		}
	}

	// Apply all migrations
	return m.Up(ctx)
}

// RunMigrations is a convenience function to run all pending migrations
func RunMigrations(db *sql.DB) error {
	migrator := NewMigrator(db)
	return migrator.Up(context.Background())
}
