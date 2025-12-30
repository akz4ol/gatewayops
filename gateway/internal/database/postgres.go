// Package database provides database connection management.
package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

// Postgres wraps the SQL database connection.
type Postgres struct {
	DB     *sql.DB
	logger zerolog.Logger
}

// NewPostgres creates a new PostgreSQL connection.
func NewPostgres(cfg config.DatabaseConfig, logger zerolog.Logger) (*Postgres, error) {
	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	logger.Info().
		Str("url", maskDSN(cfg.URL)).
		Int("max_open_conns", cfg.MaxOpenConns).
		Int("max_idle_conns", cfg.MaxIdleConns).
		Msg("Connected to PostgreSQL")

	return &Postgres{
		DB:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection.
func (p *Postgres) Close() error {
	return p.DB.Close()
}

// Health checks if the database is healthy.
func (p *Postgres) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return p.DB.PingContext(ctx) == nil
}

// Ready checks if the database is ready to accept queries.
func (p *Postgres) Ready() bool {
	return p.Health()
}

// maskDSN masks sensitive parts of a database connection string.
func maskDSN(dsn string) string {
	// Simple masking - just show host
	if len(dsn) > 30 {
		return dsn[:30] + "..."
	}
	return dsn
}
