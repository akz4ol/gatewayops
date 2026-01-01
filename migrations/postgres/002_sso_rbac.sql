-- Migration 002: SSO Providers and RBAC
-- Adds support for SSO/OIDC integration and role-based access control

-- Users table (extends the system with full user management)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sso_provider_id UUID,
    sso_external_id VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, email)
);

CREATE INDEX idx_users_org_id ON users(org_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_sso_external_id ON users(sso_provider_id, sso_external_id) WHERE sso_external_id IS NOT NULL;

-- SSO Providers
CREATE TABLE IF NOT EXISTS sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- okta, azure_ad, google, onelogin, auth0, oidc
    name VARCHAR(255) NOT NULL,
    issuer_url VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret_encrypted BYTEA NOT NULL,
    authorization_url VARCHAR(500),
    token_url VARCHAR(500),
    userinfo_url VARCHAR(500),
    scopes JSONB DEFAULT '["openid", "email", "profile"]',
    claim_mappings JSONB DEFAULT '{}',
    group_mappings JSONB DEFAULT '{}', -- SSO group -> role name
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sso_providers_org_id ON sso_providers(org_id);

-- Add foreign key from users to sso_providers
ALTER TABLE users ADD CONSTRAINT fk_users_sso_provider
    FOREIGN KEY (sso_provider_id) REFERENCES sso_providers(id) ON DELETE SET NULL;

-- User Sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    access_token VARCHAR(500) NOT NULL,
    refresh_token VARCHAR(500),
    expires_at TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Roles
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE, -- NULL for builtin roles
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_builtin BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name) WHERE org_id IS NULL AND is_builtin = true
);

CREATE INDEX idx_roles_org_id ON roles(org_id);

-- User Role Assignments
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_type VARCHAR(50) NOT NULL DEFAULT 'global', -- global, team, mcp_server
    scope_id UUID, -- team_id or mcp_server_id when scoped
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    UNIQUE(user_id, role_id, scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'))
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Insert builtin roles
INSERT INTO roles (id, name, description, permissions, is_builtin)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', 'Full access to all resources and settings', '["*"]', true),
    ('00000000-0000-0000-0000-000000000002', 'developer', 'Access to MCP tools, traces, and own API keys', '["mcp:read", "mcp:call", "traces:read", "costs:read:team", "keys:read", "keys:create"]', true),
    ('00000000-0000-0000-0000-000000000003', 'viewer', 'Read-only access to traces and costs', '["traces:read:own", "costs:read:team"]', true),
    ('00000000-0000-0000-0000-000000000004', 'billing', 'Access to costs and usage data only', '["costs:read", "costs:export"]', true)
ON CONFLICT DO NOTHING;

-- Update audit_logs to include user_id
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);

-- Add triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sso_providers_updated_at BEFORE UPDATE ON sso_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
