-- Migration 002: Add missing columns for service persistence
-- Adds columns required by repositories for alerting, safety, and approval services

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
-- Note: metric_value already exists, we're adding 'value' and 'threshold' as separate columns
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS value DECIMAL(15, 6);
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS threshold DECIMAL(15, 6);
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS labels JSONB DEFAULT '{}';
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS acked_at TIMESTAMPTZ;
ALTER TABLE alerts ADD COLUMN IF NOT EXISTS acked_by UUID REFERENCES users(id);

-- Injection detections: add extended columns for detailed detection info
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
