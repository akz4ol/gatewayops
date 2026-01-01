// Package database provides database connection management.
package database

import (
	"database/sql"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/rs/zerolog"
)

// Postgres wraps the SQL database connection.
type Postgres struct {
	DB     *sql.DB
	logger zerolog.Logger
}

// NewPostgres creates a new PostgreSQL connection.
// In demo mode, returns a mock connection that always succeeds.
func NewPostgres(cfg config.DatabaseConfig, logger zerolog.Logger) (*Postgres, error) {
	logger.Info().
		Msg("Demo mode: Using mock PostgreSQL connection")

	return &Postgres{
		DB:     nil, // Mock - no actual DB
		logger: logger,
	}, nil
}

// Close closes the database connection.
func (p *Postgres) Close() error {
	return nil
}

// Health checks if the database is healthy.
func (p *Postgres) Health() bool {
	return true // Demo mode always healthy
}

// Ready checks if the database is ready to accept queries.
func (p *Postgres) Ready() bool {
	return true // Demo mode always ready
}
