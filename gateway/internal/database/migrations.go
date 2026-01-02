// Package database provides database connection management.
package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// MigrationRunner runs database migrations.
type MigrationRunner struct {
	db     *Postgres
	logger zerolog.Logger
}

// NewMigrationRunner creates a new migration runner.
func NewMigrationRunner(db *Postgres, logger zerolog.Logger) *MigrationRunner {
	return &MigrationRunner{
		db:     db,
		logger: logger,
	}
}

// Run executes all pending migrations from the embedded filesystem.
func (m *MigrationRunner) Run(ctx context.Context, migrationsFS embed.FS, path string) error {
	m.logger.Info().Msg("Starting database migrations")

	// Create migrations table if not exists
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files
	entries, err := fs.ReadDir(migrationsFS, path)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Apply pending migrations
	for _, file := range files {
		if applied[file] {
			m.logger.Debug().Str("file", file).Msg("Migration already applied, skipping")
			continue
		}

		m.logger.Info().Str("file", file).Msg("Applying migration")

		content, err := fs.ReadFile(migrationsFS, filepath.Join(path, file))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		if err := m.applyMigration(ctx, file, string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		m.logger.Info().Str("file", file).Msg("Migration applied successfully")
	}

	m.logger.Info().Msg("Database migrations completed")
	return nil
}

// RunFromStrings executes migrations from a slice of SQL strings.
func (m *MigrationRunner) RunFromStrings(ctx context.Context, migrations map[string]string) error {
	m.logger.Info().Msg("Starting database migrations")

	// Create migrations table if not exists
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migration names
	var names []string
	for name := range migrations {
		names = append(names, name)
	}
	sort.Strings(names)

	// Apply pending migrations
	for _, name := range names {
		if applied[name] {
			m.logger.Debug().Str("name", name).Msg("Migration already applied, skipping")
			continue
		}

		m.logger.Info().Str("name", name).Msg("Applying migration")

		if err := m.applyMigration(ctx, name, migrations[name]); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", name, err)
		}

		m.logger.Info().Str("name", name).Msg("Migration applied successfully")
	}

	m.logger.Info().Msg("Database migrations completed")
	return nil
}

func (m *MigrationRunner) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`
	_, err := m.db.Exec(ctx, query)
	return err
}

func (m *MigrationRunner) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func (m *MigrationRunner) applyMigration(ctx context.Context, name, content string) error {
	// Start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.ExecContext(ctx, content); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Record migration
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)",
		name, time.Now(),
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// Status returns the current migration status.
func (m *MigrationRunner) Status(ctx context.Context) ([]MigrationStatus, error) {
	query := `SELECT version, applied_at FROM schema_migrations ORDER BY version`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var status []MigrationStatus
	for rows.Next() {
		var s MigrationStatus
		if err := rows.Scan(&s.Version, &s.AppliedAt); err != nil {
			return nil, err
		}
		status = append(status, s)
	}

	return status, rows.Err()
}

// MigrationStatus represents a migration's status.
type MigrationStatus struct {
	Version   string
	AppliedAt time.Time
}
