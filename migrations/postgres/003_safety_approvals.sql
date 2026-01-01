-- Migration 003: AI Safety & Tool Approvals
-- Adds support for prompt injection detection, safety policies, and tool approval workflows

-- Safety Policies
CREATE TABLE IF NOT EXISTS safety_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sensitivity VARCHAR(20) NOT NULL DEFAULT 'moderate', -- strict, moderate, permissive
    mode VARCHAR(20) NOT NULL DEFAULT 'warn', -- block, warn, log
    patterns JSONB NOT NULL DEFAULT '{"block": [], "allow": []}',
    mcp_servers JSONB DEFAULT '[]', -- empty means all servers
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);

CREATE INDEX idx_safety_policies_org_id ON safety_policies(org_id);
CREATE INDEX idx_safety_policies_enabled ON safety_policies(org_id, enabled) WHERE enabled = true;

-- Injection Detections
CREATE TABLE IF NOT EXISTS injection_detections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trace_id VARCHAR(100),
    span_id VARCHAR(100),
    policy_id UUID REFERENCES safety_policies(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL, -- prompt_injection, pii, secret, malicious
    severity VARCHAR(20) NOT NULL, -- low, medium, high, critical
    pattern_matched VARCHAR(500),
    input TEXT NOT NULL, -- truncated to 1000 chars
    action_taken VARCHAR(20) NOT NULL, -- block, warn, log
    mcp_server VARCHAR(100),
    tool_name VARCHAR(255),
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_injection_detections_org_id ON injection_detections(org_id);
CREATE INDEX idx_injection_detections_created_at ON injection_detections(org_id, created_at DESC);
CREATE INDEX idx_injection_detections_severity ON injection_detections(org_id, severity);
CREATE INDEX idx_injection_detections_trace_id ON injection_detections(trace_id) WHERE trace_id IS NOT NULL;

-- Tool Classifications
CREATE TABLE IF NOT EXISTS tool_classifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    classification VARCHAR(20) NOT NULL DEFAULT 'sensitive', -- safe, sensitive, dangerous
    requires_approval BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    UNIQUE(org_id, mcp_server, tool_name)
);

CREATE INDEX idx_tool_classifications_org_id ON tool_classifications(org_id);
CREATE INDEX idx_tool_classifications_mcp_server ON tool_classifications(org_id, mcp_server);

-- Tool Approvals
CREATE TABLE IF NOT EXISTS tool_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    requested_by UUID NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason TEXT,
    arguments JSONB DEFAULT '{}', -- tool arguments for context
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, denied, expired
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    review_note TEXT,
    expires_at TIMESTAMPTZ, -- for time-limited approvals
    trace_id VARCHAR(100)
);

CREATE INDEX idx_tool_approvals_org_id ON tool_approvals(org_id);
CREATE INDEX idx_tool_approvals_status ON tool_approvals(org_id, status);
CREATE INDEX idx_tool_approvals_requested_by ON tool_approvals(requested_by);
CREATE INDEX idx_tool_approvals_active ON tool_approvals(org_id, mcp_server, tool_name, status, expires_at)
    WHERE status = 'approved';

-- Tool Permissions (pre-approved permissions for users/teams)
CREATE TABLE IF NOT EXISTS tool_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID, -- if null, applies to whole team
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255) NOT NULL, -- can be "*" for all tools
    granted_by UUID NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    max_uses_day INTEGER, -- rate limit per day
    UNIQUE(org_id, COALESCE(team_id, '00000000-0000-0000-0000-000000000000'),
           COALESCE(user_id, '00000000-0000-0000-0000-000000000000'), mcp_server, tool_name)
);

CREATE INDEX idx_tool_permissions_org_id ON tool_permissions(org_id);
CREATE INDEX idx_tool_permissions_team_id ON tool_permissions(team_id);
CREATE INDEX idx_tool_permissions_user_id ON tool_permissions(user_id) WHERE user_id IS NOT NULL;

-- Create default safety policy for new organizations
CREATE OR REPLACE FUNCTION create_default_safety_policy()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO safety_policies (
        org_id,
        name,
        description,
        sensitivity,
        mode,
        patterns,
        created_by
    ) VALUES (
        NEW.id,
        'Default Security Policy',
        'Automatically created default policy for prompt injection detection',
        'moderate',
        'warn',
        '{
            "block": [
                "ignore previous instructions",
                "ignore all previous",
                "disregard all prior",
                "disregard previous instructions",
                "forget all previous",
                "forget your instructions",
                "you are now",
                "pretend you are",
                "act as if you",
                "jailbreak",
                "DAN mode",
                "developer mode",
                "ignore your programming",
                "bypass your",
                "override your"
            ],
            "allow": [
                "summarize the following",
                "please help me",
                "can you explain"
            ]
        }',
        '00000000-0000-0000-0000-000000000000'
    );
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER create_org_default_safety_policy AFTER INSERT ON organizations
    FOR EACH ROW EXECUTE FUNCTION create_default_safety_policy();

-- Add triggers for updated_at
CREATE TRIGGER update_safety_policies_updated_at BEFORE UPDATE ON safety_policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tool_classifications_updated_at BEFORE UPDATE ON tool_classifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
