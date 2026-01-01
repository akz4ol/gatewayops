-- Migration 004: Alerting System
-- Adds support for alert rules, channels, and alert management

-- Alert Rules
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric VARCHAR(100) NOT NULL, -- error_rate, latency_p50, latency_p95, latency_p99, request_rate, cost_per_hour, etc.
    condition VARCHAR(50) NOT NULL, -- gt, lt, gte, lte, eq, neq
    threshold DECIMAL(15,4) NOT NULL,
    window_minutes INTEGER NOT NULL DEFAULT 5,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning', -- info, warning, critical
    channels JSONB NOT NULL DEFAULT '[]', -- array of channel IDs
    filters JSONB DEFAULT '{}', -- mcp_servers, teams, environments
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);

CREATE INDEX idx_alert_rules_org_id ON alert_rules(org_id);
CREATE INDEX idx_alert_rules_enabled ON alert_rules(org_id, enabled) WHERE enabled = true;

-- Alert Channels
CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- slack, pagerduty, opsgenie, webhook, email, teams
    config JSONB NOT NULL DEFAULT '{}', -- channel-specific configuration
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_channels_org_id ON alert_channels(org_id);

-- Alerts (active and historical)
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'firing', -- firing, resolved, acknowledged
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    value DECIMAL(15,4) NOT NULL, -- the actual metric value
    threshold DECIMAL(15,4) NOT NULL, -- the threshold that was breached
    labels JSONB DEFAULT '{}', -- additional context labels
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    acked_at TIMESTAMPTZ,
    acked_by UUID
);

CREATE INDEX idx_alerts_org_id ON alerts(org_id);
CREATE INDEX idx_alerts_rule_id ON alerts(rule_id);
CREATE INDEX idx_alerts_status ON alerts(org_id, status);
CREATE INDEX idx_alerts_started_at ON alerts(org_id, started_at DESC);
CREATE INDEX idx_alerts_firing ON alerts(rule_id, status) WHERE status = 'firing';

-- Alert Notifications (track sent notifications)
CREATE TABLE IF NOT EXISTS alert_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, sent, failed
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_notifications_alert_id ON alert_notifications(alert_id);
CREATE INDEX idx_alert_notifications_status ON alert_notifications(status) WHERE status = 'pending';

-- OpenTelemetry Export Configuration
CREATE TABLE IF NOT EXISTS otel_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    protocol VARCHAR(20) NOT NULL DEFAULT 'grpc', -- grpc, http
    headers JSONB DEFAULT '{}', -- authorization headers (encrypted in practice)
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id)
);

CREATE INDEX idx_otel_configs_org_id ON otel_configs(org_id);

-- Add triggers for updated_at
CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_channels_updated_at BEFORE UPDATE ON alert_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_otel_configs_updated_at BEFORE UPDATE ON otel_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create sample alert channels for new organizations
CREATE OR REPLACE FUNCTION create_default_alert_examples()
RETURNS TRIGGER AS $$
BEGIN
    -- This is just a placeholder - actual channels should be configured by the user
    -- No default channels are created automatically
    RETURN NEW;
END;
$$ language 'plpgsql';
