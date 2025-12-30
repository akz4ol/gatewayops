// Package main is the entry point for the GatewayOps gateway service.
package main

import (
	"os"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/auth"
	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/akz4ol/gatewayops/gateway/internal/database"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/ratelimit"
	"github.com/akz4ol/gatewayops/gateway/internal/router"
	"github.com/akz4ol/gatewayops/gateway/internal/server"
	"github.com/rs/zerolog"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Setup logger
	logger := setupLogger(cfg)

	logger.Info().
		Str("env", cfg.Server.Env).
		Str("port", cfg.Server.Port).
		Msg("Starting GatewayOps Gateway")

	// Connect to PostgreSQL
	postgres, err := database.NewPostgres(cfg.Database, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer postgres.Close()

	// Connect to Redis
	redis, err := database.NewRedis(cfg.Redis, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	// Initialize auth store
	authStore := auth.NewStore(postgres.DB, logger)

	// Initialize rate limiter
	rateLimiter := ratelimit.NewLimiter(redis.Client, logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(postgres, redis, rateLimiter)
	mcpHandler := handler.NewMCPHandler(cfg, logger)

	// Create router with dependencies
	deps := router.Dependencies{
		Config:        cfg,
		Logger:        logger,
		AuthStore:     authStore,
		RateLimiter:   rateLimiter,
		MCPHandler:    mcpHandler,
		HealthHandler: healthHandler,
	}

	r := router.New(deps)

	// Create and start server
	srv := server.New(cfg, r, logger)

	logger.Info().
		Str("addr", srv.Addr()).
		Int("mcp_servers", len(cfg.MCPServers)).
		Msg("Gateway ready to accept connections")

	if err := srv.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Server error")
	}

	logger.Info().Msg("Gateway shutdown complete")
}

// setupLogger configures zerolog based on environment.
func setupLogger(cfg *config.Config) zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var logger zerolog.Logger
	if cfg.Logging.Format == "console" || cfg.IsDevelopment() {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return logger
}
