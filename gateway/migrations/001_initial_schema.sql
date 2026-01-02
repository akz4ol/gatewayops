-- GatewayOps Initial Schema
-- Migration 001: Core tables for traces, API keys, organizations, and teams

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
    provider VARCHAR(50), -- local, google, okta, azure_ad
    provider_id VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member', -- admin, member, viewer
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
    key_prefix VARCHAR(16) NOT NULL, -- First 8 chars for identification
    key_hash VARCHAR(128) NOT NULL, -- SHA-256 hash of the key
    environment VARCHAR(50) DEFAULT 'development', -- production, staging, development
    permissions JSONB DEFAULT '["*"]',
    rate_limit INTEGER DEFAULT 1000, -- Requests per minute
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
    operation VARCHAR(100) NOT NULL, -- tools/call, resources/read, prompts/get
    tool_name VARCHAR(255),
    status VARCHAR(20) NOT NULL, -- success, error, timeout
    status_code INTEGER DEFAULT 200,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    request_size INTEGER DEFAULT 0,
    response_size INTEGER DEFAULT 0,
    cost DECIMAL(12, 6) DEFAULT 0,
    error_msg TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Trace spans table for detailed tracing
CREATE TABLE IF NOT EXISTS trace_spans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(32) NOT NULL,
    parent_id VARCHAR(32),
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(20) NOT NULL, -- client, server, internal
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
    action VARCHAR(100) NOT NULL, -- mcp.tool.call, api_key.create, etc.
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    outcome VARCHAR(20) NOT NULL, -- success, failure, blocked
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
    provider_type VARCHAR(50) NOT NULL, -- okta, azure_ad, google, auth0
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
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_type VARCHAR(50), -- NULL for org-wide, 'team' for team-scoped
    scope_id UUID, -- team_id if scoped to team
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id, COALESCE(scope_type, ''), COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'))
);

-- Safety policies table
CREATE TABLE IF NOT EXISTS safety_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sensitivity VARCHAR(20) DEFAULT 'moderate', -- low, moderate, high
    mode VARCHAR(20) DEFAULT 'warn', -- block, warn, log
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
    severity VARCHAR(20) NOT NULL, -- low, medium, high, critical
    pattern_matched VARCHAR(255),
    input_snippet TEXT,
    action_taken VARCHAR(20) NOT NULL, -- blocked, warned, logged
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tool classifications table
CREATE TABLE IF NOT EXISTS tool_classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    classification VARCHAR(20) NOT NULL, -- safe, sensitive, dangerous
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
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, denied
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
    metric VARCHAR(100) NOT NULL, -- error_rate, latency_p99, cost
    condition VARCHAR(50) NOT NULL, -- gt, lt, gte, lte
    threshold DECIMAL(10, 4) NOT NULL,
    window_minutes INTEGER DEFAULT 5,
    severity VARCHAR(20) DEFAULT 'warning', -- info, warning, critical
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
    type VARCHAR(50) NOT NULL, -- slack, pagerduty, webhook, email
    config JSONB NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Alerts table (fired alerts)
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'firing', -- firing, resolved
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metric_value DECIMAL(10, 4),
    started_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ
);

-- Indexes for performance
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
