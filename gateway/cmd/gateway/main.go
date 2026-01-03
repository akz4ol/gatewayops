// Package main is the entry point for the GatewayOps gateway service.
package main

import (
	"context"
	_ "embed"
	"os"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/agent"
	"github.com/akz4ol/gatewayops/gateway/internal/alerting"
	"github.com/akz4ol/gatewayops/gateway/internal/approval"
	"github.com/akz4ol/gatewayops/gateway/internal/audit"
	"github.com/akz4ol/gatewayops/gateway/internal/auth"
	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/akz4ol/gatewayops/gateway/internal/database"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/otel"
	"github.com/akz4ol/gatewayops/gateway/internal/ratelimit"
	"github.com/akz4ol/gatewayops/gateway/internal/rbac"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/akz4ol/gatewayops/gateway/internal/router"
	"github.com/akz4ol/gatewayops/gateway/internal/safety"
	"github.com/akz4ol/gatewayops/gateway/internal/server"
	"github.com/akz4ol/gatewayops/gateway/internal/sso"
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

	// Run database migrations
	if postgres.DB != nil {
		migrationRunner := database.NewMigrationRunner(postgres, logger)
		migrations := getMigrations()
		if err := migrationRunner.RunFromStrings(context.Background(), migrations); err != nil {
			logger.Fatal().Err(err).Msg("Failed to run database migrations")
		}
	}

	// Initialize repositories
	traceRepo := repository.NewTraceRepository(postgres.DB)
	costRepo := repository.NewCostRepository(postgres.DB)
	alertRepo := repository.NewAlertRepository(postgres.DB)
	safetyRepo := repository.NewSafetyRepository(postgres.DB)
	toolRepo := repository.NewToolRepository(postgres.DB)
	apiKeyRepo := repository.NewAPIKeyRepository(postgres.DB)

	// Initialize auth store
	authStore := auth.NewStore(postgres.DB, logger)

	// Initialize rate limiter
	rateLimiter := ratelimit.NewLimiter(redis, logger)

	// Initialize injection detector (with repository for persistence)
	injectionDetector := safety.NewDetector(logger, safetyRepo)

	// Initialize audit logger
	auditLogger := audit.NewLogger(logger)

	// Initialize alerting service (with repository for persistence)
	alertService := alerting.NewService(logger, alertRepo)

	// Initialize OpenTelemetry exporter
	otelExporter := otel.NewExporter(logger)

	// Initialize tool approval service (with repository for persistence)
	approvalService := approval.NewService(logger, toolRepo)

	// Initialize RBAC service
	rbacService := rbac.NewService(logger)

	// Initialize SSO service
	ssoService := sso.NewService(logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(postgres, redis, rateLimiter)
	mcpHandler := handler.NewMCPHandler(cfg, logger, traceRepo)
	traceHandler := handler.NewTraceHandler(logger, traceRepo, cfg.Server.DemoMode)
	costHandler := handler.NewCostHandler(logger, costRepo, cfg.Server.DemoMode)
	apiKeyHandler := handler.NewAPIKeyHandler(logger, apiKeyRepo, cfg.Server.DemoMode)
	metricsHandler := handler.NewMetricsHandler(logger)
	docsHandler := handler.NewDocsHandler(logger, openAPISpec)
	safetyHandler := handler.NewSafetyHandler(logger, injectionDetector)
	auditHandler := handler.NewAuditHandler(logger, auditLogger)
	alertHandler := handler.NewAlertHandler(logger, alertService)
	telemetryHandler := handler.NewTelemetryHandler(logger, otelExporter)
	approvalHandler := handler.NewApprovalHandler(logger, approvalService)
	rbacHandler := handler.NewRBACHandler(logger, rbacService)
	ssoHandler := handler.NewSSOHandler(logger, ssoService, "https://gatewayops-api.fly.dev")

	// Initialize user handler
	userRepo := repository.NewUserRepository(postgres.DB)
	userHandler := handler.NewUserHandler(logger, userRepo, rbacService)

	// Initialize settings handler
	settingsHandler := handler.NewSettingsHandler(logger)

	// Initialize agent manager and handler
	agentManager := agent.NewManager(logger)
	agentHandler := handler.NewAgentHandler(logger, agentManager, "gatewayops-api.fly.dev")

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
		UserHandler:       userHandler,
		SettingsHandler:   settingsHandler,
		AgentHandler:      agentHandler,
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

// getMigrations returns the database migrations as a map.
func getMigrations() map[string]string {
	return map[string]string{
		"001_initial_schema.sql": `
-- GatewayOps Initial Schema
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, name)
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    avatar_url VARCHAR(500),
    provider VARCHAR(50),
    provider_id VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member',
    settings JSONB DEFAULT '{}',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, email)
);

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(16) NOT NULL,
    key_hash VARCHAR(128) NOT NULL,
    environment VARCHAR(50) DEFAULT 'development',
    permissions JSONB DEFAULT '["*"]',
    rate_limit INTEGER DEFAULT 1000,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMPTZ,
    UNIQUE(key_prefix)
);

-- Traces table
CREATE TABLE IF NOT EXISTS traces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(32) NOT NULL,
    parent_id VARCHAR(32),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    mcp_server VARCHAR(100) NOT NULL,
    operation VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255),
    status VARCHAR(20) NOT NULL,
    status_code INTEGER DEFAULT 200,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    request_size INTEGER DEFAULT 0,
    response_size INTEGER DEFAULT 0,
    cost DECIMAL(12, 6) DEFAULT 0,
    error_msg TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Trace spans table
CREATE TABLE IF NOT EXISTS trace_spans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(32) NOT NULL,
    parent_id VARCHAR(32),
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    duration_ms BIGINT NOT NULL,
    attributes JSONB DEFAULT '{}'
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    trace_id VARCHAR(64),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    outcome VARCHAR(20) NOT NULL,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(64),
    duration_ms BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- SSO providers table
CREATE TABLE IF NOT EXISTS sso_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    issuer_url VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret_encrypted BYTEA NOT NULL,
    claim_mappings JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token_hash VARCHAR(128) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User role assignments
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_type VARCHAR(50),
    scope_id UUID,
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, role_id, scope_type, scope_id)
);

-- Safety policies table
CREATE TABLE IF NOT EXISTS safety_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sensitivity VARCHAR(20) DEFAULT 'moderate',
    mode VARCHAR(20) DEFAULT 'warn',
    patterns JSONB NOT NULL DEFAULT '[]',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Injection detections table
CREATE TABLE IF NOT EXISTS injection_detections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trace_id VARCHAR(64),
    policy_id UUID REFERENCES safety_policies(id) ON DELETE SET NULL,
    severity VARCHAR(20) NOT NULL,
    pattern_matched VARCHAR(255),
    input_snippet TEXT,
    action_taken VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tool classifications table
CREATE TABLE IF NOT EXISTS tool_classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    classification VARCHAR(20) NOT NULL,
    requires_approval BOOLEAN DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, mcp_server, tool_name)
);

-- Tool approvals table
CREATE TABLE IF NOT EXISTS tool_approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    requested_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    reason TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Alert rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric VARCHAR(100) NOT NULL,
    condition VARCHAR(50) NOT NULL,
    threshold DECIMAL(10, 4) NOT NULL,
    window_minutes INTEGER DEFAULT 5,
    severity VARCHAR(20) DEFAULT 'warning',
    channels JSONB NOT NULL DEFAULT '[]',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Alert channels table
CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'firing',
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metric_value DECIMAL(10, 4),
    started_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_traces_org_id ON traces(org_id);
CREATE INDEX IF NOT EXISTS idx_traces_created_at ON traces(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_traces_mcp_server ON traces(mcp_server);
CREATE INDEX IF NOT EXISTS idx_traces_status ON traces(status);
CREATE INDEX IF NOT EXISTS idx_traces_org_created ON traces(org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_traces_trace_id ON traces(trace_id);
CREATE INDEX IF NOT EXISTS idx_trace_spans_trace_id ON trace_spans(trace_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_id ON audit_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_created ON audit_logs(org_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_keys_org_id ON api_keys(org_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_alerts_org_id ON alerts(org_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(token_hash);

-- Insert built-in roles
INSERT INTO roles (id, name, description, permissions, is_builtin) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'Full administrative access', '["*"]', true),
    ('00000000-0000-0000-0000-000000000002', 'developer', 'Developer access to MCP and traces', '["mcp:read","mcp:call","traces:read","costs:read"]', true),
    ('00000000-0000-0000-0000-000000000003', 'viewer', 'Read-only access', '["traces:read","costs:read","audit:read"]', true)
ON CONFLICT DO NOTHING;

-- Insert demo organization
INSERT INTO organizations (id, name, slug) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Demo Organization', 'demo')
ON CONFLICT DO NOTHING;

-- Insert demo users
INSERT INTO users (id, org_id, email, name, role) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'sarah@acme.com', 'Sarah Chen', 'admin'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'demo@acme.com', 'Demo User', 'developer')
ON CONFLICT DO NOTHING;
`,
		"002_add_service_columns.sql": `
-- Migration 002: Add missing columns for service persistence

-- Safety policies: add mcp_servers and created_by columns
ALTER TABLE safety_policies ADD COLUMN IF NOT EXISTS mcp_servers JSONB DEFAULT '[]';
ALTER TABLE safety_policies ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);

-- Alert rules: add filters and created_by columns
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS filters JSONB DEFAULT '{}';
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);

-- Tool classifications: add created_by column
ALTER TABLE tool_classifications ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);

-- Tool approvals: add requested_at, arguments, review_note, trace_id columns
ALTER TABLE tool_approvals ADD COLUMN IF NOT EXISTS requested_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE tool_approvals ADD COLUMN IF NOT EXISTS arguments JSONB;
ALTER TABLE tool_approvals ADD COLUMN IF NOT EXISTS review_note TEXT;
ALTER TABLE tool_approvals ADD COLUMN IF NOT EXISTS trace_id VARCHAR(64);

-- Alerts: add value, threshold, labels, acked_at, acked_by columns
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS value DECIMAL(15, 6);
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS threshold DECIMAL(15, 6);
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS labels JSONB DEFAULT '{}';
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS acked_at TIMESTAMPTZ;
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS acked_by UUID REFERENCES users(id);

-- Injection detections: add extended columns
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS span_id VARCHAR(32);
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS type VARCHAR(50);
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS input TEXT;
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS mcp_server VARCHAR(100);
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS tool_name VARCHAR(255);
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS api_key_id UUID REFERENCES api_keys(id);
ALTER TABLE injection_detections ADD COLUMN IF NOT EXISTS ip_address INET;

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_safety_policies_org_enabled ON safety_policies(org_id, enabled);
CREATE INDEX IF NOT EXISTS idx_tool_approvals_requested_at ON tool_approvals(requested_at DESC);
CREATE INDEX IF NOT EXISTS idx_injection_detections_type ON injection_detections(type);
CREATE INDEX IF NOT EXISTS idx_injection_detections_mcp_server ON injection_detections(mcp_server);
`,
		"003_add_demo_users.sql": `
-- Migration 003: Add demo users for API key creation
INSERT INTO users (id, org_id, email, name, role) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'sarah@acme.com', 'Sarah Chen', 'admin'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'demo@acme.com', 'Demo User', 'developer')
ON CONFLICT DO NOTHING;
`,
	}
}
