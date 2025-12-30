-- GatewayOps ClickHouse Schema
-- Version: 001_initial

-- ============================================
-- Database
-- ============================================
CREATE DATABASE IF NOT EXISTS gatewayops;

USE gatewayops;

-- ============================================
-- Traces Table
-- ============================================
CREATE TABLE IF NOT EXISTS traces (
    trace_id UUID,
    org_id String,
    user_id String,
    team_id String,
    api_key_id String,
    agent_name String,
    environment String DEFAULT 'production',
    start_time DateTime64(3, 'UTC'),
    end_time DateTime64(3, 'UTC'),
    duration_ms UInt64,
    status Enum8('success' = 1, 'error' = 2, 'timeout' = 3),
    span_count UInt32,
    total_cost Decimal64(6),
    error_message Nullable(String),
    tags Map(String, String),

    -- Indexes
    INDEX idx_user_id user_id TYPE bloom_filter GRANULARITY 4,
    INDEX idx_team_id team_id TYPE bloom_filter GRANULARITY 4,
    INDEX idx_status status TYPE set(3) GRANULARITY 4,
    INDEX idx_agent_name agent_name TYPE bloom_filter GRANULARITY 4
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMM(start_time))
ORDER BY (org_id, start_time, trace_id)
TTL start_time + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

-- ============================================
-- Spans Table
-- ============================================
CREATE TABLE IF NOT EXISTS spans (
    trace_id UUID,
    span_id UUID,
    parent_span_id Nullable(UUID),
    org_id String,
    operation String,
    mcp_server String,
    tool_name String,
    start_time DateTime64(3, 'UTC'),
    end_time DateTime64(3, 'UTC'),
    duration_ms UInt64,
    status Enum8('success' = 1, 'error' = 2, 'timeout' = 3),
    request_body String CODEC(ZSTD(3)),
    response_body String CODEC(ZSTD(3)),
    request_size UInt32,
    response_size UInt32,
    cost Decimal64(6),
    error_message Nullable(String),
    attributes Map(String, String),

    -- Indexes
    INDEX idx_mcp_server mcp_server TYPE bloom_filter GRANULARITY 4,
    INDEX idx_tool_name tool_name TYPE bloom_filter GRANULARITY 4,
    INDEX idx_status status TYPE set(3) GRANULARITY 4
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMM(start_time))
ORDER BY (org_id, trace_id, start_time, span_id)
TTL start_time + INTERVAL 30 DAY DELETE
SETTINGS index_granularity = 8192;

-- ============================================
-- Cost Events Table
-- ============================================
CREATE TABLE IF NOT EXISTS cost_events (
    event_id UUID,
    trace_id UUID,
    span_id UUID,
    org_id String,
    user_id String,
    team_id String,
    api_key_id String,
    mcp_server String,
    tool_name String,
    timestamp DateTime64(3, 'UTC'),
    input_tokens UInt32 DEFAULT 0,
    output_tokens UInt32 DEFAULT 0,
    cost Decimal64(6),
    currency String DEFAULT 'USD',

    -- Indexes
    INDEX idx_mcp_server mcp_server TYPE bloom_filter GRANULARITY 4,
    INDEX idx_team_id team_id TYPE bloom_filter GRANULARITY 4
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMM(timestamp))
ORDER BY (org_id, timestamp, event_id)
TTL timestamp + INTERVAL 365 DAY DELETE
SETTINGS index_granularity = 8192;

-- ============================================
-- Materialized Views for Aggregations
-- ============================================

-- Daily costs by team
CREATE MATERIALIZED VIEW IF NOT EXISTS cost_daily_by_team_mv
ENGINE = SummingMergeTree()
PARTITION BY (org_id, toYYYYMM(date))
ORDER BY (org_id, date, team_id, mcp_server)
AS SELECT
    org_id,
    toDate(timestamp) AS date,
    team_id,
    mcp_server,
    sum(cost) AS total_cost,
    count() AS call_count,
    sum(input_tokens) AS total_input_tokens,
    sum(output_tokens) AS total_output_tokens
FROM cost_events
GROUP BY org_id, date, team_id, mcp_server;

-- Daily costs by server
CREATE MATERIALIZED VIEW IF NOT EXISTS cost_daily_by_server_mv
ENGINE = SummingMergeTree()
PARTITION BY (org_id, toYYYYMM(date))
ORDER BY (org_id, date, mcp_server, tool_name)
AS SELECT
    org_id,
    toDate(timestamp) AS date,
    mcp_server,
    tool_name,
    sum(cost) AS total_cost,
    count() AS call_count
FROM cost_events
GROUP BY org_id, date, mcp_server, tool_name;

-- Hourly metrics for real-time dashboard
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_hourly_mv
ENGINE = SummingMergeTree()
PARTITION BY (org_id, toYYYYMM(hour))
ORDER BY (org_id, hour, mcp_server)
AS SELECT
    org_id,
    toStartOfHour(timestamp) AS hour,
    mcp_server,
    count() AS request_count,
    countIf(status = 'error') AS error_count,
    sum(cost) AS total_cost,
    avg(duration_ms) AS avg_duration_ms
FROM (
    SELECT
        org_id,
        start_time AS timestamp,
        mcp_server,
        status,
        total_cost AS cost,
        duration_ms
    FROM traces
)
GROUP BY org_id, hour, mcp_server;

-- ============================================
-- Request Logs Table (for detailed logging)
-- ============================================
CREATE TABLE IF NOT EXISTS request_logs (
    log_id UUID,
    trace_id UUID,
    org_id String,
    api_key_id String,
    timestamp DateTime64(3, 'UTC'),
    method String,
    path String,
    status_code UInt16,
    latency_ms UInt32,
    request_headers Map(String, String),
    response_headers Map(String, String),
    ip_address String,
    user_agent String,
    error_message Nullable(String)
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMMDD(timestamp))
ORDER BY (org_id, timestamp, log_id)
TTL timestamp + INTERVAL 30 DAY DELETE
SETTINGS index_granularity = 8192;
