// Package database provides database connection management.
package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

// Postgres wraps the SQL database connection.
type Postgres struct {
	DB     *sql.DB
	logger zerolog.Logger
	cfg    config.DatabaseConfig
}

// NewPostgres creates a new PostgreSQL connection.
func NewPostgres(cfg config.DatabaseConfig, logger zerolog.Logger) (*Postgres, error) {
	logger.Info().
		Str("url", maskDSN(cfg.URL)).
		Int("max_open_conns", cfg.MaxOpenConns).
		Int("max_idle_conns", cfg.MaxIdleConns).
		Msg("Connecting to PostgreSQL")

	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	logger.Info().Msg("PostgreSQL connected successfully")

	return &Postgres{
		DB:     db,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// Close closes the database connection.
func (p *Postgres) Close() error {
	if p.DB != nil {
		p.logger.Info().Msg("Closing PostgreSQL connection")
		return p.DB.Close()
	}
	return nil
}

// Health checks if the database is healthy.
func (p *Postgres) Health() bool {
	if p.DB == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.DB.PingContext(ctx); err != nil {
		p.logger.Warn().Err(err).Msg("PostgreSQL health check failed")
		return false
	}
	return true
}

// Ready checks if the database is ready to accept queries.
func (p *Postgres) Ready() bool {
	if p.DB == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple query
	var result int
	err := p.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		p.logger.Warn().Err(err).Msg("PostgreSQL ready check failed")
		return false
	}
	return result == 1
}

// Stats returns database connection pool statistics.
func (p *Postgres) Stats() sql.DBStats {
	if p.DB == nil {
		return sql.DBStats{}
	}
	return p.DB.Stats()
}

// Exec executes a query without returning any rows.
func (p *Postgres) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.DB.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
func (p *Postgres) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.DB.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row.
func (p *Postgres) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.DB.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction.
func (p *Postgres) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.DB.BeginTx(ctx, opts)
}

// maskDSN masks sensitive information in the DSN for logging.
func maskDSN(dsn string) string {
	// Simple masking - in production, use proper URL parsing
	if len(dsn) > 30 {
		return dsn[:20] + "..." + dsn[len(dsn)-10:]
	}
	return "***"
}
