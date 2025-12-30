-- GatewayOps PostgreSQL Schema
-- Version: 001_initial

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- Organizations
-- ============================================
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create default organization for development
INSERT INTO organizations (id, name, slug, plan) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Organization', 'default', 'pro');

-- ============================================
-- Teams
-- ============================================
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, slug)
);

CREATE INDEX idx_teams_org ON teams(org_id);

-- Create default team
INSERT INTO teams (id, org_id, name, slug) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Engineering', 'engineering');

-- ============================================
-- Users
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member',
    sso_provider VARCHAR(50),
    sso_id VARCHAR(255),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    UNIQUE(org_id, email)
);

CREATE INDEX idx_users_org ON users(org_id);
CREATE INDEX idx_users_email ON users(email);

-- ============================================
-- API Keys
-- ============================================
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,       -- e.g., gwo_prod_a1B2
    key_hash VARCHAR(255) NOT NULL,         -- bcrypt hash of full key
    scopes JSONB DEFAULT '{"mcp_servers": ["*"], "permissions": ["*"]}',
    rate_limit_rpm INTEGER DEFAULT 1000,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_org ON api_keys(org_id);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);

-- Create a default development API key
-- Key: gwo_dev_DEVELOPMENT_KEY_FOR_TESTING_1234 (hash generated with bcrypt cost 12)
INSERT INTO api_keys (id, org_id, name, key_prefix, key_hash, scopes) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     'Development Key',
     'gwo_dev_DEVE',
     '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.VTtYvN0aQNc.Gq',  -- password: gwo_dev_DEVELOPMENT_KEY_FOR_TESTING_1234
     '{"mcp_servers": ["*"], "permissions": ["*"]}');

-- ============================================
-- MCP Server Configurations
-- ============================================
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255),
    url VARCHAR(500) NOT NULL,
    auth_type VARCHAR(50) DEFAULT 'none',  -- none, bearer, basic, custom
    auth_config JSONB DEFAULT '{}',         -- encrypted auth details
    timeout_ms INTEGER DEFAULT 30000,
    max_retries INTEGER DEFAULT 3,
    enabled BOOLEAN DEFAULT true,
    pricing JSONB DEFAULT '{"per_call": 0.001}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, name)
);

CREATE INDEX idx_mcp_servers_org ON mcp_servers(org_id);

-- Create mock MCP server configuration for development
INSERT INTO mcp_servers (id, org_id, name, display_name, url, pricing) VALUES
    ('00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000001',
     'mock',
     'Mock MCP Server',
     'http://mock-mcp:3000',
     '{"per_call": 0.001}');

-- ============================================
-- Budgets
-- ============================================
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    scope_type VARCHAR(50) NOT NULL,  -- org, team, user, project
    scope_id UUID,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    period VARCHAR(20) NOT NULL,      -- daily, weekly, monthly
    alert_thresholds INTEGER[] DEFAULT '{50, 80, 100}',
    enforce_limit BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_budgets_org ON budgets(org_id);
CREATE INDEX idx_budgets_scope ON budgets(scope_type, scope_id);

-- ============================================
-- Audit Logs
-- ============================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    actor_type VARCHAR(50) NOT NULL,   -- user, api_key, system
    actor_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create partitions for current and next month
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE audit_logs_2025_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE audit_logs_2025_03 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE INDEX idx_audit_logs_org_time ON audit_logs(org_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_type, actor_id);

-- ============================================
-- Functions
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mcp_servers_updated_at BEFORE UPDATE ON mcp_servers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_budgets_updated_at BEFORE UPDATE ON budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
