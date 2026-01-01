// Package main is the entry point for the GatewayOps gateway service.
package main

import (
	_ "embed"
	"os"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/alerting"
	"github.com/akz4ol/gatewayops/gateway/internal/approval"
	"github.com/akz4ol/gatewayops/gateway/internal/audit"
	"github.com/akz4ol/gatewayops/gateway/internal/otel"
	"github.com/akz4ol/gatewayops/gateway/internal/rbac"
	"github.com/akz4ol/gatewayops/gateway/internal/sso"
	"github.com/akz4ol/gatewayops/gateway/internal/auth"
	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/akz4ol/gatewayops/gateway/internal/database"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/ratelimit"
	"github.com/akz4ol/gatewayops/gateway/internal/router"
	"github.com/akz4ol/gatewayops/gateway/internal/safety"
	"github.com/akz4ol/gatewayops/gateway/internal/server"
	"github.com/rs/zerolog"
)

//go:embed openapi.yaml
var openAPISpec []byte

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

	// Initialize injection detector
	injectionDetector := safety.NewDetector(logger)

	// Initialize audit logger
	auditLogger := audit.NewLogger(logger)

	// Initialize alerting service
	alertService := alerting.NewService(logger)

	// Initialize OpenTelemetry exporter
	otelExporter := otel.NewExporter(logger)

	// Initialize tool approval service
	approvalService := approval.NewService(logger)

	// Initialize RBAC service
	rbacService := rbac.NewService(logger)

	// Initialize SSO service
	ssoService := sso.NewService(logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(postgres, redis, rateLimiter)
	mcpHandler := handler.NewMCPHandler(cfg, logger)
	traceHandler := handler.NewTraceHandler(logger)
	costHandler := handler.NewCostHandler(logger)
	apiKeyHandler := handler.NewAPIKeyHandler(logger)
	metricsHandler := handler.NewMetricsHandler(logger)
	docsHandler := handler.NewDocsHandler(logger, openAPISpec)
	safetyHandler := handler.NewSafetyHandler(logger, injectionDetector)
	auditHandler := handler.NewAuditHandler(logger, auditLogger)
	alertHandler := handler.NewAlertHandler(logger, alertService)
	telemetryHandler := handler.NewTelemetryHandler(logger, otelExporter)
	approvalHandler := handler.NewApprovalHandler(logger, approvalService)
	rbacHandler := handler.NewRBACHandler(logger, rbacService)
	ssoHandler := handler.NewSSOHandler(logger, ssoService, "https://gatewayops-api.fly.dev")

	// Create router with dependencies
	deps := router.Dependencies{
		Config:            cfg,
		Logger:            logger,
		AuthStore:         authStore,
		RateLimiter:       rateLimiter,
		InjectionDetector: injectionDetector,
		AuditLogger:       auditLogger,
		MCPHandler:        mcpHandler,
		HealthHandler:     healthHandler,
		TraceHandler:      traceHandler,
		CostHandler:       costHandler,
		APIKeyHandler:     apiKeyHandler,
		MetricsHandler:    metricsHandler,
		DocsHandler:       docsHandler,
		SafetyHandler:     safetyHandler,
		AuditHandler:      auditHandler,
		AlertHandler:      alertHandler,
		TelemetryHandler:  telemetryHandler,
		ApprovalHandler:   approvalHandler,
		RBACHandler:       rbacHandler,
		SSOHandler:        ssoHandler,
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
