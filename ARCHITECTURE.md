# GatewayOps MCP Gateway
## Technical Architecture Document

**Document Version:** 1.0
**Last Updated:** 2025-12-30
**Status:** Draft
**Owner:** CTO Agent
**Classification:** Internal

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [System Components](#system-components)
3. [Data Architecture](#data-architecture)
4. [API Design](#api-design)
5. [Infrastructure](#infrastructure)
6. [Security Architecture](#security-architecture)
7. [Scalability & Performance](#scalability--performance)
8. [Reliability & Availability](#reliability--availability)
9. [Observability](#observability)
10. [Development & Deployment](#development--deployment)
11. [Disaster Recovery](#disaster-recovery)
12. [Technology Decisions](#technology-decisions)

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                    CLIENTS                                               │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                 │
│  │  AI Agent   │   │  AI Agent   │   │   Web UI    │   │   CLI/SDK   │                 │
│  │  (Claude)   │   │  (Custom)   │   │ (Dashboard) │   │  (DevTools) │                 │
│  └──────┬──────┘   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘                 │
│         │                 │                 │                 │                         │
│         └─────────────────┴─────────────────┴─────────────────┘                         │
│                                     │                                                    │
└─────────────────────────────────────┼────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                 EDGE LAYER                                               │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌───────────────────────────────────────────────────────────────────────────────────┐  │
│  │                              AWS CloudFront (CDN)                                  │  │
│  │                         - Static asset caching                                     │  │
│  │                         - DDoS protection (Shield)                                 │  │
│  │                         - Edge TLS termination                                     │  │
│  └───────────────────────────────────────────────────────────────────────────────────┘  │
│                                          │                                               │
│  ┌───────────────────────────────────────────────────────────────────────────────────┐  │
│  │                              AWS WAF (Web Application Firewall)                    │  │
│  │                         - Rate limiting rules                                      │  │
│  │                         - SQL injection protection                                 │  │
│  │                         - IP reputation blocking                                   │  │
│  └───────────────────────────────────────────────────────────────────────────────────┘  │
│                                          │                                               │
└──────────────────────────────────────────┼───────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              LOAD BALANCING LAYER                                        │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌────────────────────────────────┐    ┌────────────────────────────────┐               │
│  │     AWS ALB (Gateway API)      │    │     AWS ALB (Dashboard)        │               │
│  │     gateway.gatewayops.com     │    │     app.gatewayops.com         │               │
│  │     - Health checks            │    │     - Session stickiness       │               │
│  │     - TLS termination          │    │     - HTTP/2 support           │               │
│  └────────────────┬───────────────┘    └────────────────┬───────────────┘               │
│                   │                                      │                               │
└───────────────────┼──────────────────────────────────────┼───────────────────────────────┘
                    │                                      │
                    ▼                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              APPLICATION LAYER (EKS)                                     │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌─────────────────────────────────────┐  ┌─────────────────────────────────────┐       │
│  │         GATEWAY SERVICE             │  │         WEB APPLICATION             │       │
│  │         (Go, Stateless)             │  │         (Next.js, Stateless)        │       │
│  │                                     │  │                                     │       │
│  │  ┌─────────────────────────────┐   │  │  ┌─────────────────────────────┐   │       │
│  │  │      Gateway Pod (x3+)      │   │  │  │      Web Pod (x2+)          │   │       │
│  │  │  - MCP Protocol Handler     │   │  │  │  - SSR/React                │   │       │
│  │  │  - Auth Middleware          │   │  │  │  - API Routes               │   │       │
│  │  │  - Rate Limiter             │   │  │  │  - Session Management       │   │       │
│  │  │  - Cost Calculator          │   │  │  │                             │   │       │
│  │  │  - Trace Generator          │   │  │  └─────────────────────────────┘   │       │
│  │  │  - Request Logger           │   │  │                                     │       │
│  │  └─────────────────────────────┘   │  └─────────────────────────────────────┘       │
│  │                                     │                                                │
│  └─────────────────┬───────────────────┘                                                │
│                    │                                                                     │
│  ┌─────────────────┴───────────────────────────────────────────────────────────────┐   │
│  │                           INTERNAL SERVICES                                      │   │
│  │                                                                                  │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │   │
│  │  │ Auth Service │  │ Cost Service │  │Trace Service │  │Alert Service │        │   │
│  │  │   (Go)       │  │    (Go)      │  │    (Go)      │  │    (Go)      │        │   │
│  │  │              │  │              │  │              │  │              │        │   │
│  │  │ - SSO/SAML   │  │ - Aggregate  │  │ - Ingest     │  │ - Rules      │        │   │
│  │  │ - API Keys   │  │ - Budget     │  │ - Query      │  │ - Notify     │        │   │
│  │  │ - RBAC       │  │ - Report     │  │ - Export     │  │ - Escalate   │        │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘        │   │
│  │                                                                                  │   │
│  └──────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              DATA LAYER                                                  │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐                       │
│  │   PostgreSQL     │  │   ClickHouse     │  │     Redis        │                       │
│  │   (AWS RDS)      │  │   (ClickHouse    │  │   (ElastiCache)  │                       │
│  │                  │  │    Cloud)        │  │                  │                       │
│  │ - Users         │  │                  │  │ - Rate limits    │                       │
│  │ - API Keys      │  │ - Traces         │  │ - Sessions       │                       │
│  │ - Orgs/Teams    │  │ - Metrics        │  │ - Cache          │                       │
│  │ - Configs       │  │ - Cost Events    │  │ - Pub/Sub        │                       │
│  │ - Audit Logs    │  │ - Aggregations   │  │                  │                       │
│  │                  │  │                  │  │                  │                       │
│  │ Primary + Read  │  │ Distributed      │  │ Cluster Mode     │                       │
│  │ Replica         │  │ Cluster          │  │ (3 nodes)        │                       │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘                       │
│                                                                                          │
│  ┌──────────────────┐  ┌──────────────────┐                                             │
│  │      S3          │  │  Secrets Manager │                                             │
│  │                  │  │                  │                                             │
│  │ - Trace exports  │  │ - DB credentials │                                             │
│  │ - Audit archives │  │ - API secrets    │                                             │
│  │ - Backups        │  │ - Encryption key │                                             │
│  └──────────────────┘  └──────────────────┘                                             │
│                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              EXTERNAL INTEGRATIONS                                       │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                │
│  │ MCP Servers  │  │    Slack     │  │  PagerDuty   │  │   SendGrid   │                │
│  │ (Customer)   │  │              │  │              │  │              │                │
│  │              │  │ - Alerts     │  │ - Incidents  │  │ - Emails     │                │
│  │ - Filesystem │  │ - Reports    │  │ - On-call    │  │ - Invites    │                │
│  │ - Database   │  │              │  │              │  │              │                │
│  │ - Custom     │  │              │  │              │  │              │                │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘                │
│                                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                                   │
│  │   Datadog    │  │    Stripe    │  │  IdP (SSO)   │                                   │
│  │              │  │              │  │              │                                   │
│  │ - APM        │  │ - Billing    │  │ - Okta       │                                   │
│  │ - Logs       │  │ - Invoices   │  │ - Azure AD   │                                   │
│  │ - Metrics    │  │              │  │ - Google     │                                   │
│  └──────────────┘  └──────────────┘  └──────────────┘                                   │
│                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────┘
```

### Architecture Principles

```yaml
architecture_principles:

  stateless_services:
    description: "All application services are stateless"
    rationale: "Enables horizontal scaling, simplifies deployment"
    implementation:
      - "No local state in gateway or web pods"
      - "Session state in Redis"
      - "All persistence in databases"

  separation_of_concerns:
    description: "Clear boundaries between components"
    rationale: "Independent scaling, deployment, and evolution"
    implementation:
      - "Gateway handles MCP proxying only"
      - "Dedicated services for auth, costs, traces"
      - "Web app is pure presentation layer"

  defense_in_depth:
    description: "Multiple layers of security"
    rationale: "No single point of failure for security"
    implementation:
      - "Edge: WAF, DDoS protection"
      - "Network: VPC, security groups"
      - "Application: Auth, rate limiting"
      - "Data: Encryption, masking"

  observability_first:
    description: "Comprehensive visibility into system behavior"
    rationale: "Critical for debugging, operations, dogfooding"
    implementation:
      - "Structured logging everywhere"
      - "Distributed tracing for all requests"
      - "Metrics for all components"
      - "We use our own product for observability"

  fail_gracefully:
    description: "Handle failures without cascading"
    rationale: "Maintain availability during partial outages"
    implementation:
      - "Circuit breakers for external calls"
      - "Timeouts on all network operations"
      - "Graceful degradation (serve cached data)"
      - "Clear error messages to clients"
```

---

## System Components

### Gateway Service

The core component that proxies MCP traffic.

```yaml
gateway_service:
  name: "gateway"
  language: "Go 1.21+"
  repository: "github.com/gatewayops/gateway"

  responsibilities:
    - "Accept and validate MCP requests"
    - "Authenticate requests (API key or session)"
    - "Apply rate limiting"
    - "Route requests to MCP servers"
    - "Calculate and record costs"
    - "Generate and propagate trace context"
    - "Log requests and responses"

  architecture:
    ```
    ┌─────────────────────────────────────────────────────────────────┐
    │                      Gateway Service                             │
    ├─────────────────────────────────────────────────────────────────┤
    │                                                                  │
    │  ┌─────────────┐                                                │
    │  │   HTTP      │  ← Incoming MCP requests                       │
    │  │   Server    │                                                │
    │  └──────┬──────┘                                                │
    │         │                                                        │
    │         ▼                                                        │
    │  ┌─────────────────────────────────────────────────────────┐   │
    │  │                  Middleware Chain                        │   │
    │  │                                                          │   │
    │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐│   │
    │  │  │ Request  │→│   Auth   │→│   Rate   │→│  Trace   ││   │
    │  │  │   ID     │  │Middleware│  │ Limiter  │  │  Start   ││   │
    │  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘│   │
    │  │                                                          │   │
    │  └─────────────────────────┬────────────────────────────────┘   │
    │                            │                                     │
    │                            ▼                                     │
    │  ┌─────────────────────────────────────────────────────────┐   │
    │  │                   MCP Router                             │   │
    │  │                                                          │   │
    │  │  - Parse MCP request                                     │   │
    │  │  - Lookup target server                                  │   │
    │  │  - Validate request schema                               │   │
    │  │  - Apply request transformations                         │   │
    │  │                                                          │   │
    │  └─────────────────────────┬────────────────────────────────┘   │
    │                            │                                     │
    │                            ▼                                     │
    │  ┌─────────────────────────────────────────────────────────┐   │
    │  │                   MCP Proxy                              │   │
    │  │                                                          │   │
    │  │  - Connection pooling                                    │   │
    │  │  - Retry logic                                           │   │
    │  │  - Circuit breaker                                       │   │
    │  │  - Timeout handling                                      │   │
    │  │                                                          │   │
    │  └─────────────────────────┬────────────────────────────────┘   │
    │                            │                                     │
    │                            ▼                                     │
    │  ┌─────────────────────────────────────────────────────────┐   │
    │  │                Post-Processing                           │   │
    │  │                                                          │   │
    │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐│   │
    │  │  │   Cost   │  │  Trace   │  │ Request  │  │ Response ││   │
    │  │  │  Calc    │  │  End     │  │  Log     │  │  Return  ││   │
    │  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘│   │
    │  │                                                          │   │
    │  └──────────────────────────────────────────────────────────┘   │
    │                                                                  │
    └──────────────────────────────────────────────────────────────────┘
    ```

  configuration:
    ```yaml
    # config.yaml
    server:
      http_port: 8080
      grpc_port: 9090
      read_timeout: 30s
      write_timeout: 30s
      idle_timeout: 120s
      max_request_size: 10MB

    auth:
      api_key_header: "Authorization"
      api_key_prefix: "Bearer gwo_"

    rate_limiting:
      enabled: true
      default_rps: 100
      burst_size: 50
      redis_url: "${REDIS_URL}"

    mcp_servers:
      discovery: "config"  # or "dns" for service discovery
      default_timeout: 30s
      max_retries: 3
      circuit_breaker:
        threshold: 5
        timeout: 30s

    tracing:
      enabled: true
      sample_rate: 1.0  # 100% sampling
      exporter: "clickhouse"

    logging:
      level: "info"
      format: "json"
      mask_sensitive: true
    ```

  api_surface:
    ```go
    // Main MCP proxy endpoints
    POST /v1/mcp/{server}/tools/call
    POST /v1/mcp/{server}/tools/list
    POST /v1/mcp/{server}/resources/read
    POST /v1/mcp/{server}/resources/list
    POST /v1/mcp/{server}/prompts/get
    POST /v1/mcp/{server}/prompts/list

    // Health endpoints
    GET /health          // Liveness probe
    GET /ready           // Readiness probe
    GET /metrics         // Prometheus metrics
    ```

  scaling:
    min_replicas: 3
    max_replicas: 50
    target_cpu: 70%
    target_memory: 80%
```

### Auth Service

Handles authentication and authorization.

```yaml
auth_service:
  name: "auth"
  language: "Go 1.21+"

  responsibilities:
    - "API key validation and management"
    - "SSO integration (SAML, OIDC)"
    - "Session management"
    - "RBAC policy enforcement"
    - "Audit logging for auth events"

  data_flow:
    ```
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │   Gateway   │────▶│    Auth     │────▶│ PostgreSQL  │
    │             │     │   Service   │     │  (Users,    │
    │  Validate   │◀────│             │◀────│   Keys)     │
    │  API Key    │     │  Lookup &   │     │             │
    │             │     │  Validate   │     │             │
    └─────────────┘     └──────┬──────┘     └─────────────┘
                               │
                               │ Cache
                               ▼
                        ┌─────────────┐
                        │    Redis    │
                        │  (Sessions, │
                        │  Key Cache) │
                        └─────────────┘
    ```

  api_surface:
    ```go
    // Internal APIs (called by Gateway)
    POST /internal/v1/auth/validate-key
    POST /internal/v1/auth/validate-session
    POST /internal/v1/auth/check-permission

    // External APIs (called by Web App)
    POST /v1/auth/keys                  // Create API key
    GET  /v1/auth/keys                  // List API keys
    DELETE /v1/auth/keys/{id}           // Revoke key
    POST /v1/auth/keys/{id}/rotate      // Rotate key

    // SSO endpoints
    GET  /v1/auth/sso/saml/metadata
    POST /v1/auth/sso/saml/acs
    GET  /v1/auth/sso/oidc/authorize
    POST /v1/auth/sso/oidc/callback
    POST /v1/auth/sso/oidc/token
    ```

  caching_strategy:
    api_keys:
      cache_location: "Redis"
      ttl: "5 minutes"
      invalidation: "On key update/revoke"

    sessions:
      cache_location: "Redis"
      ttl: "8 hours (sliding)"
      invalidation: "On logout"

    permissions:
      cache_location: "Redis"
      ttl: "1 minute"
      invalidation: "On policy change"
```

### Cost Service

Tracks and aggregates costs.

```yaml
cost_service:
  name: "cost"
  language: "Go 1.21+"

  responsibilities:
    - "Receive cost events from Gateway"
    - "Aggregate costs by various dimensions"
    - "Manage budgets and alerts"
    - "Generate cost reports"

  data_flow:
    ```
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │   Gateway   │────▶│    Cost     │────▶│ ClickHouse  │
    │             │     │   Service   │     │  (Events,   │
    │  Cost Event │     │             │     │  Aggregates)│
    │  (async)    │     │  Ingest &   │     │             │
    │             │     │  Aggregate  │     │             │
    └─────────────┘     └──────┬──────┘     └─────────────┘
                               │
                               │ Budget Check
                               ▼
                        ┌─────────────┐
                        │ PostgreSQL  │
                        │  (Budgets,  │
                        │   Alerts)   │
                        └──────┬──────┘
                               │
                               │ Alert Trigger
                               ▼
                        ┌─────────────┐
                        │   Alert     │
                        │   Service   │
                        └─────────────┘
    ```

  ingestion:
    method: "Async via message queue"
    queue: "Redis Streams"
    batch_size: 1000
    flush_interval: "1 second"

  aggregation:
    real_time:
      - "Cost by org (hourly)"
      - "Cost by team (hourly)"
      - "Cost by server (hourly)"
    materialized_views:
      - "Daily rollups"
      - "Weekly rollups"
      - "Monthly rollups"

  api_surface:
    ```go
    // Cost queries
    GET /v1/costs/summary
    GET /v1/costs/by-team
    GET /v1/costs/by-user
    GET /v1/costs/by-server
    GET /v1/costs/trace/{trace_id}
    POST /v1/costs/export

    // Budget management
    GET  /v1/budgets
    POST /v1/budgets
    PUT  /v1/budgets/{id}
    DELETE /v1/budgets/{id}
    GET /v1/budgets/{id}/usage
    ```
```

### Trace Service

Manages distributed tracing data.

```yaml
trace_service:
  name: "trace"
  language: "Go 1.21+"

  responsibilities:
    - "Ingest trace spans from Gateway"
    - "Store and index traces"
    - "Query and search traces"
    - "Export to external systems"

  data_flow:
    ```
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │   Gateway   │────▶│   Trace     │────▶│ ClickHouse  │
    │             │     │   Service   │     │  (Traces,   │
    │  Spans      │     │             │     │   Spans)    │
    │  (async)    │     │  Ingest &   │     │             │
    │             │     │  Index      │     │             │
    └─────────────┘     └──────┬──────┘     └─────────────┘
                               │
                               │ Export (optional)
                               ▼
                        ┌─────────────┐
                        │   OTLP      │
                        │  Exporter   │
                        │  (Datadog,  │
                        │   Jaeger)   │
                        └─────────────┘
    ```

  storage_schema:
    traces_table:
      ```sql
      CREATE TABLE traces (
        trace_id UUID,
        org_id String,
        user_id String,
        team_id String,
        start_time DateTime64(3),
        end_time DateTime64(3),
        duration_ms UInt64,
        status Enum('success', 'error', 'timeout'),
        span_count UInt32,
        total_cost Decimal64(6),
        metadata Map(String, String),

        -- Partitioning and ordering
        INDEX idx_user_id user_id TYPE bloom_filter,
        INDEX idx_team_id team_id TYPE bloom_filter,
        INDEX idx_status status TYPE set(3)
      )
      ENGINE = MergeTree()
      PARTITION BY toYYYYMMDD(start_time)
      ORDER BY (org_id, start_time, trace_id)
      TTL start_time + INTERVAL 30 DAY;
      ```

    spans_table:
      ```sql
      CREATE TABLE spans (
        trace_id UUID,
        span_id UUID,
        parent_span_id Nullable(UUID),
        org_id String,
        operation String,
        mcp_server String,
        tool_name String,
        start_time DateTime64(3),
        end_time DateTime64(3),
        duration_ms UInt64,
        status Enum('success', 'error', 'timeout'),
        request_size UInt32,
        response_size UInt32,
        cost Decimal64(6),
        error_message Nullable(String),
        attributes Map(String, String)
      )
      ENGINE = MergeTree()
      PARTITION BY toYYYYMMDD(start_time)
      ORDER BY (org_id, trace_id, start_time)
      TTL start_time + INTERVAL 30 DAY;
      ```

  api_surface:
    ```go
    // Trace queries
    GET  /v1/traces
    GET  /v1/traces/{trace_id}
    GET  /v1/traces/{trace_id}/spans
    POST /v1/traces/search

    // Export
    POST /v1/traces/export
    GET  /v1/traces/export/{job_id}
    ```
```

### Web Application

Dashboard and management UI.

```yaml
web_application:
  name: "web"
  framework: "Next.js 14 (App Router)"
  repository: "github.com/gatewayops/web"

  architecture:
    ```
    ┌─────────────────────────────────────────────────────────────┐
    │                    Next.js Application                       │
    ├─────────────────────────────────────────────────────────────┤
    │                                                              │
    │  ┌─────────────────────────────────────────────────────┐   │
    │  │                    App Router                        │   │
    │  │                                                      │   │
    │  │  /                    → Dashboard (SSR)              │   │
    │  │  /traces              → Trace List (SSR + Client)    │   │
    │  │  /traces/[id]         → Trace Detail (SSR)           │   │
    │  │  /costs               → Cost Dashboard (SSR)         │   │
    │  │  /settings/*          → Settings Pages (SSR)         │   │
    │  │  /api/*               → API Routes (Server)          │   │
    │  │                                                      │   │
    │  └─────────────────────────────────────────────────────┘   │
    │                                                              │
    │  ┌─────────────────────────────────────────────────────┐   │
    │  │                   Components                         │   │
    │  │                                                      │   │
    │  │  - shadcn/ui (Radix primitives)                     │   │
    │  │  - Recharts (data visualization)                    │   │
    │  │  - TanStack Table (data tables)                     │   │
    │  │  - TanStack Query (data fetching)                   │   │
    │  │                                                      │   │
    │  └─────────────────────────────────────────────────────┘   │
    │                                                              │
    │  ┌─────────────────────────────────────────────────────┐   │
    │  │                   API Layer                          │   │
    │  │                                                      │   │
    │  │  - Server actions for mutations                     │   │
    │  │  - API routes for complex operations                │   │
    │  │  - Direct service calls (server-side)               │   │
    │  │                                                      │   │
    │  └─────────────────────────────────────────────────────┘   │
    │                                                              │
    └──────────────────────────────────────────────────────────────┘
    ```

  pages:
    - path: "/"
      name: "Dashboard"
      data_sources: ["Cost Service", "Trace Service", "Alert Service"]
      caching: "ISR (60 seconds)"

    - path: "/traces"
      name: "Trace List"
      data_sources: ["Trace Service"]
      caching: "No cache (real-time)"

    - path: "/traces/[id]"
      name: "Trace Detail"
      data_sources: ["Trace Service"]
      caching: "Static (traces are immutable)"

    - path: "/costs"
      name: "Cost Dashboard"
      data_sources: ["Cost Service"]
      caching: "ISR (60 seconds)"

    - path: "/settings/api-keys"
      name: "API Key Management"
      data_sources: ["Auth Service"]
      caching: "No cache"

  state_management:
    server_state: "TanStack Query"
    client_state: "Zustand (minimal)"
    form_state: "React Hook Form"
```

---

## Data Architecture

### Database Selection Rationale

```yaml
database_selection:

  postgresql:
    use_cases:
      - "User accounts and profiles"
      - "Organizations and teams"
      - "API keys (hashed)"
      - "MCP server configurations"
      - "Budgets and policies"
      - "Audit logs"
    rationale:
      - "ACID transactions for critical data"
      - "Rich query capabilities"
      - "Mature ecosystem"
      - "Easy to operate (RDS)"
    scaling:
      - "Vertical scaling (larger instance)"
      - "Read replicas for dashboard queries"
      - "Connection pooling (PgBouncer)"

  clickhouse:
    use_cases:
      - "Trace spans (billions of rows)"
      - "Cost events"
      - "Metrics time series"
      - "Aggregations and rollups"
    rationale:
      - "Columnar storage for analytics"
      - "Excellent compression (10-20x)"
      - "Fast aggregation queries"
      - "Native time-series support"
    scaling:
      - "Horizontal sharding"
      - "Automatic replication"
      - "Tiered storage"

  redis:
    use_cases:
      - "Rate limiting counters"
      - "Session storage"
      - "API key cache"
      - "Real-time metrics"
      - "Message queue (Streams)"
    rationale:
      - "Sub-millisecond latency"
      - "Native data structures"
      - "Pub/sub for real-time"
    scaling:
      - "Cluster mode (sharding)"
      - "Read replicas"
```

### Data Models

#### PostgreSQL Schema

```sql
-- Organizations
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(org_id, slug)
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member',
    sso_provider VARCHAR(50),
    sso_id VARCHAR(255),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(org_id, email)
);

-- Team memberships
CREATE TABLE team_memberships (
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID NOT NULL REFERENCES teams(id),
    role VARCHAR(50) DEFAULT 'member',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, team_id)
);

-- API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,  -- For identification
    key_hash VARCHAR(255) NOT NULL,   -- bcrypt hash
    scopes JSONB DEFAULT '{}',
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    INDEX idx_key_prefix (key_prefix)
);

-- MCP Server Configurations
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    auth_type VARCHAR(50),
    auth_config JSONB,  -- Encrypted
    timeout_ms INTEGER DEFAULT 30000,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(org_id, name)
);

-- Budgets
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    scope_type VARCHAR(50) NOT NULL,  -- 'org', 'team', 'user', 'project'
    scope_id UUID,
    amount DECIMAL(10, 2) NOT NULL,
    period VARCHAR(20) NOT NULL,  -- 'daily', 'weekly', 'monthly'
    alert_thresholds INTEGER[] DEFAULT '{50, 80, 100}',
    enforce_limit BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    actor_type VARCHAR(50) NOT NULL,  -- 'user', 'api_key', 'system'
    actor_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

#### ClickHouse Schema

```sql
-- Traces
CREATE TABLE traces (
    trace_id UUID,
    org_id String,
    user_id String,
    team_id String,
    agent_name String,
    environment String,
    start_time DateTime64(3, 'UTC'),
    end_time DateTime64(3, 'UTC'),
    duration_ms UInt64,
    status Enum8('success' = 1, 'error' = 2, 'timeout' = 3),
    span_count UInt32,
    total_cost Decimal64(6),
    error_message Nullable(String),
    tags Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMM(start_time))
ORDER BY (org_id, start_time, trace_id)
TTL start_time + INTERVAL 90 DAY DELETE
SETTINGS index_granularity = 8192;

-- Spans
CREATE TABLE spans (
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
    attributes Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMM(start_time))
ORDER BY (org_id, trace_id, start_time, span_id)
TTL start_time + INTERVAL 30 DAY DELETE
SETTINGS index_granularity = 8192;

-- Cost Events
CREATE TABLE cost_events (
    event_id UUID,
    trace_id UUID,
    span_id UUID,
    org_id String,
    user_id String,
    team_id String,
    project_id String,
    mcp_server String,
    tool_name String,
    timestamp DateTime64(3, 'UTC'),
    input_tokens UInt32,
    output_tokens UInt32,
    cost Decimal64(6),
    currency String DEFAULT 'USD'
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMM(timestamp))
ORDER BY (org_id, timestamp, event_id)
TTL timestamp + INTERVAL 365 DAY DELETE;

-- Materialized view for daily costs
CREATE MATERIALIZED VIEW cost_daily_mv
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

-- Metrics (for real-time dashboard)
CREATE TABLE metrics (
    org_id String,
    metric_name String,
    timestamp DateTime64(3, 'UTC'),
    value Float64,
    tags Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY (org_id, toYYYYMMDD(timestamp))
ORDER BY (org_id, metric_name, timestamp)
TTL timestamp + INTERVAL 30 DAY DELETE;
```

### Data Flow Patterns

```yaml
data_flows:

  mcp_request_flow:
    description: "Flow of an MCP request through the system"
    steps:
      1. "Client sends MCP request to Gateway"
      2. "Gateway authenticates (Redis cache → Auth Service → PostgreSQL)"
      3. "Gateway checks rate limit (Redis)"
      4. "Gateway generates trace context"
      5. "Gateway proxies to MCP server"
      6. "Gateway calculates cost"
      7. "Gateway publishes events (Redis Streams)"
      8. "Cost Service consumes events → ClickHouse"
      9. "Trace Service consumes events → ClickHouse"
      10. "Gateway returns response with trace ID"

  cost_aggregation_flow:
    description: "How costs are aggregated"
    steps:
      1. "Cost events written to ClickHouse"
      2. "Materialized views auto-aggregate by day"
      3. "Cost Service queries aggregates for dashboards"
      4. "Budget checks run against real-time sums"
      5. "Alerts triggered when thresholds exceeded"

  audit_log_flow:
    description: "How audit events are captured"
    steps:
      1. "Auth event occurs (login, key create, etc.)"
      2. "Service writes to audit_logs table (PostgreSQL)"
      3. "Partitioned by month for retention"
      4. "Queried for compliance reports"
```

---

## API Design

### API Design Principles

```yaml
api_principles:

  restful:
    - "Resource-oriented URLs"
    - "HTTP verbs for operations"
    - "Consistent response format"

  versioning:
    strategy: "URL path versioning"
    current: "v1"
    format: "/v1/resource"
    deprecation: "6 month notice, 12 month support"

  authentication:
    methods:
      - "Bearer token (API key)"
      - "Session cookie (web app)"
    header: "Authorization: Bearer <token>"

  pagination:
    style: "Cursor-based"
    parameters:
      limit: "Number of items (max 200)"
      cursor: "Opaque cursor for next page"
    response:
      items: "Array of resources"
      cursor: "Next page cursor"
      has_more: "Boolean"

  error_handling:
    format:
      ```json
      {
        "error": {
          "code": "validation_error",
          "message": "Human-readable message",
          "details": [
            {
              "field": "email",
              "message": "Invalid email format"
            }
          ],
          "request_id": "req_abc123"
        }
      }
      ```
    codes:
      - "400: validation_error, bad_request"
      - "401: unauthorized, invalid_token"
      - "403: forbidden, insufficient_permissions"
      - "404: not_found"
      - "429: rate_limited"
      - "500: internal_error"
      - "502: upstream_error"
      - "504: timeout"

  rate_limiting:
    headers:
      - "X-RateLimit-Limit"
      - "X-RateLimit-Remaining"
      - "X-RateLimit-Reset"
      - "Retry-After (on 429)"
```

### API Specification (OpenAPI)

```yaml
openapi: "3.1.0"
info:
  title: "GatewayOps API"
  version: "1.0.0"
  description: "Enterprise control plane for AI agent infrastructure"

servers:
  - url: "https://api.gatewayops.com/v1"
    description: "Production"
  - url: "https://api.staging.gatewayops.com/v1"
    description: "Staging"

security:
  - BearerAuth: []

paths:
  /mcp/{server}/tools/call:
    post:
      summary: "Proxy MCP tool call"
      tags: ["Gateway"]
      parameters:
        - name: server
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/MCPToolCallRequest"
      responses:
        "200":
          description: "Successful response"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/MCPToolCallResponse"
        "401":
          $ref: "#/components/responses/Unauthorized"
        "429":
          $ref: "#/components/responses/RateLimited"
        "502":
          $ref: "#/components/responses/UpstreamError"

  /traces:
    get:
      summary: "List traces"
      tags: ["Observability"]
      parameters:
        - name: start_time
          in: query
          required: true
          schema:
            type: string
            format: date-time
        - name: end_time
          in: query
          required: true
          schema:
            type: string
            format: date-time
        - name: status
          in: query
          schema:
            type: string
            enum: [success, error, timeout]
        - name: user_id
          in: query
          schema:
            type: string
        - name: team_id
          in: query
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
            default: 50
            maximum: 200
        - name: cursor
          in: query
          schema:
            type: string
      responses:
        "200":
          description: "List of traces"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/TraceListResponse"

  /costs/summary:
    get:
      summary: "Get cost summary"
      tags: ["Costs"]
      parameters:
        - name: start_date
          in: query
          required: true
          schema:
            type: string
            format: date
        - name: end_date
          in: query
          required: true
          schema:
            type: string
            format: date
        - name: group_by
          in: query
          schema:
            type: string
            enum: [day, week, month, team, user, server]
            default: day
      responses:
        "200":
          description: "Cost summary"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CostSummaryResponse"

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: "API Key"

  schemas:
    MCPToolCallRequest:
      type: object
      required: [tool, arguments]
      properties:
        tool:
          type: string
          description: "Tool name to invoke"
        arguments:
          type: object
          description: "Tool arguments"

    MCPToolCallResponse:
      type: object
      properties:
        result:
          type: object
          description: "Tool result"
        metadata:
          type: object
          properties:
            trace_id:
              type: string
            latency_ms:
              type: integer
            cost:
              type: number

    TraceListResponse:
      type: object
      properties:
        traces:
          type: array
          items:
            $ref: "#/components/schemas/TraceSummary"
        cursor:
          type: string
        has_more:
          type: boolean

    TraceSummary:
      type: object
      properties:
        id:
          type: string
          format: uuid
        start_time:
          type: string
          format: date-time
        duration_ms:
          type: integer
        status:
          type: string
          enum: [success, error, timeout]
        span_count:
          type: integer
        total_cost:
          type: number

    CostSummaryResponse:
      type: object
      properties:
        summary:
          type: object
          properties:
            total_cost:
              type: number
            total_calls:
              type: integer
            period:
              type: object
              properties:
                start:
                  type: string
                  format: date
                end:
                  type: string
                  format: date
        breakdown:
          type: array
          items:
            type: object
            properties:
              group:
                type: string
              cost:
                type: number
              calls:
                type: integer

  responses:
    Unauthorized:
      description: "Authentication required"
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"

    RateLimited:
      description: "Rate limit exceeded"
      headers:
        Retry-After:
          schema:
            type: integer
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"

    UpstreamError:
      description: "MCP server error"
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
```

---

## Infrastructure

### AWS Infrastructure

```yaml
aws_infrastructure:

  region_primary: "us-east-1"
  region_dr: "us-west-2"  # Future

  networking:
    vpc:
      cidr: "10.0.0.0/16"
      availability_zones: ["us-east-1a", "us-east-1b", "us-east-1c"]

    subnets:
      public:
        - "10.0.1.0/24"  # ALB, NAT Gateway
        - "10.0.2.0/24"
        - "10.0.3.0/24"
      private:
        - "10.0.10.0/24"  # EKS nodes
        - "10.0.11.0/24"
        - "10.0.12.0/24"
      database:
        - "10.0.20.0/24"  # RDS, ElastiCache
        - "10.0.21.0/24"
        - "10.0.22.0/24"

    security_groups:
      alb:
        inbound:
          - "443 from 0.0.0.0/0"
        outbound:
          - "All to EKS nodes"

      eks_nodes:
        inbound:
          - "All from ALB"
          - "All from other EKS nodes"
        outbound:
          - "All to databases"
          - "443 to Internet (NAT)"

      databases:
        inbound:
          - "5432 from EKS nodes"  # PostgreSQL
          - "6379 from EKS nodes"  # Redis
          - "9000 from EKS nodes"  # ClickHouse
        outbound:
          - "None"

  compute:
    eks:
      version: "1.29"
      node_groups:
        gateway:
          instance_types: ["c6i.xlarge", "c6i.2xlarge"]
          min_size: 3
          max_size: 20
          labels:
            workload: "gateway"
        services:
          instance_types: ["m6i.large", "m6i.xlarge"]
          min_size: 2
          max_size: 10
          labels:
            workload: "services"
        web:
          instance_types: ["t3.medium", "t3.large"]
          min_size: 2
          max_size: 5
          labels:
            workload: "web"

  databases:
    postgresql:
      engine: "aurora-postgresql"
      version: "15.4"
      instance_class: "db.r6g.large"
      instances: 2  # Writer + Reader
      storage:
        type: "aurora"
        encrypted: true
      backup:
        retention: 7
        window: "03:00-04:00"

    redis:
      engine: "redis"
      version: "7.0"
      node_type: "cache.r6g.large"
      num_cache_clusters: 3  # Cluster mode
      automatic_failover: true

  storage:
    s3:
      buckets:
        - name: "gatewayops-traces-export"
          lifecycle:
            - transition_to_ia: 30
            - transition_to_glacier: 90
            - expiration: 365
        - name: "gatewayops-audit-archive"
          lifecycle:
            - transition_to_glacier: 30
            - expiration: 2555  # 7 years
        - name: "gatewayops-backups"
          versioning: true
          replication: "us-west-2"

  observability:
    cloudwatch:
      log_groups:
        - "/aws/eks/gatewayops/gateway"
        - "/aws/eks/gatewayops/services"
        - "/aws/rds/gatewayops"
      retention: 30

    x_ray:
      enabled: false  # Using Datadog instead

  secrets:
    secrets_manager:
      - "gatewayops/prod/database"
      - "gatewayops/prod/redis"
      - "gatewayops/prod/clickhouse"
      - "gatewayops/prod/encryption-key"
```

### Kubernetes Architecture

```yaml
kubernetes_architecture:

  namespaces:
    - name: "gatewayops"
      description: "Main application workloads"
    - name: "gatewayops-system"
      description: "System components (ingress, monitoring)"

  workloads:
    gateway:
      type: "Deployment"
      replicas:
        min: 3
        max: 50
      resources:
        requests:
          cpu: "500m"
          memory: "512Mi"
        limits:
          cpu: "2000m"
          memory: "2Gi"
      hpa:
        target_cpu: 70
        target_memory: 80
      pod_disruption_budget:
        min_available: 2
      affinity:
        node_selector:
          workload: "gateway"
        pod_anti_affinity: "preferred"

    auth_service:
      type: "Deployment"
      replicas:
        min: 2
        max: 10
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "1000m"
          memory: "1Gi"

    cost_service:
      type: "Deployment"
      replicas:
        min: 2
        max: 10
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "1000m"
          memory: "1Gi"

    trace_service:
      type: "Deployment"
      replicas:
        min: 2
        max: 10
      resources:
        requests:
          cpu: "500m"
          memory: "512Mi"
        limits:
          cpu: "2000m"
          memory: "2Gi"

    web:
      type: "Deployment"
      replicas:
        min: 2
        max: 5
      resources:
        requests:
          cpu: "200m"
          memory: "256Mi"
        limits:
          cpu: "1000m"
          memory: "1Gi"

  services:
    - name: "gateway"
      type: "ClusterIP"
      ports: [8080, 9090]

    - name: "auth"
      type: "ClusterIP"
      ports: [8080]

    - name: "cost"
      type: "ClusterIP"
      ports: [8080]

    - name: "trace"
      type: "ClusterIP"
      ports: [8080]

    - name: "web"
      type: "ClusterIP"
      ports: [3000]

  ingress:
    class: "alb"
    annotations:
      alb.ingress.kubernetes.io/scheme: "internet-facing"
      alb.ingress.kubernetes.io/certificate-arn: "${ACM_CERT_ARN}"
      alb.ingress.kubernetes.io/ssl-policy: "ELBSecurityPolicy-TLS13-1-2-2021-06"

    rules:
      - host: "gateway.gatewayops.com"
        service: "gateway"
        port: 8080

      - host: "app.gatewayops.com"
        service: "web"
        port: 3000

      - host: "api.gatewayops.com"
        paths:
          - path: "/v1/auth"
            service: "auth"
          - path: "/v1/costs"
            service: "cost"
          - path: "/v1/traces"
            service: "trace"
```

### Infrastructure as Code

```yaml
infrastructure_as_code:

  tool: "Terraform"
  version: "1.6+"

  structure:
    ```
    infrastructure/
    ├── modules/
    │   ├── vpc/
    │   ├── eks/
    │   ├── rds/
    │   ├── elasticache/
    │   └── s3/
    ├── environments/
    │   ├── prod/
    │   │   ├── main.tf
    │   │   ├── variables.tf
    │   │   └── terraform.tfvars
    │   ├── staging/
    │   └── dev/
    └── global/
        ├── iam/
        ├── route53/
        └── acm/
    ```

  state:
    backend: "S3"
    locking: "DynamoDB"
    encryption: true

  ci_cd:
    tool: "GitHub Actions"
    workflow:
      - "terraform fmt -check"
      - "terraform validate"
      - "terraform plan"
      - "Manual approval for apply"
      - "terraform apply"
```

---

## Security Architecture

### Security Layers

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SECURITY LAYERS                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  LAYER 1: EDGE SECURITY                                                     │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ • AWS Shield (DDoS)        • AWS WAF (Rules)        • CloudFront      │ │
│  │ • Rate limiting            • Geo blocking           • Bot detection   │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                       │
│  LAYER 2: NETWORK SECURITY                                                  │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ • VPC isolation            • Security groups        • NACLs           │ │
│  │ • Private subnets          • NAT Gateway            • No public IPs   │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                       │
│  LAYER 3: APPLICATION SECURITY                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ • Authentication           • Authorization          • Input validation│ │
│  │ • Rate limiting (app)      • CORS                   • CSP headers     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                      │                                       │
│  LAYER 4: DATA SECURITY                                                     │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ • Encryption at rest       • Encryption in transit  • Key management  │ │
│  │ • Data masking             • Audit logging          • Backup encryption│ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Authentication Architecture

```yaml
authentication:

  api_keys:
    generation:
      algorithm: "CSPRNG"
      length: 32 bytes
      format: "gwo_{env}_{base62_encoded}"
      example: "gwo_prod_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6"

    storage:
      hash_algorithm: "bcrypt"
      cost_factor: 12
      stored_fields:
        - "key_prefix (for lookup)"
        - "key_hash (bcrypt)"
        - "metadata (JSON)"

    validation:
      flow:
        1. "Extract key from Authorization header"
        2. "Parse prefix (first 12 chars)"
        3. "Lookup in cache (Redis)"
        4. "If miss, lookup in database"
        5. "Verify bcrypt hash"
        6. "Check expiration"
        7. "Check scopes"
        8. "Cache result"

  sso:
    saml:
      version: "2.0"
      binding: "HTTP-POST"
      signing: "RSA-SHA256"
      encryption: "AES-256-CBC"

    oidc:
      flows: ["authorization_code"]
      pkce: "required"
      scopes: ["openid", "email", "profile"]

    user_provisioning:
      mode: "JIT (Just-In-Time)"
      attribute_mapping:
        email: "email"
        name: "name"
        groups: "groups"
```

### Authorization Model

```yaml
authorization:

  model: "RBAC with resource-level permissions"

  roles:
    org_admin:
      description: "Full access to organization"
      permissions: ["*"]

    org_member:
      description: "Standard member access"
      permissions:
        - "traces:read"
        - "costs:read"
        - "api_keys:read"
        - "api_keys:create (own)"

    org_viewer:
      description: "Read-only access"
      permissions:
        - "traces:read"
        - "costs:read"

    team_admin:
      description: "Full access to team resources"
      scope: "team"
      permissions:
        - "team:*"
        - "api_keys:* (team)"
        - "budgets:* (team)"

  permission_check:
    ```go
    type Permission struct {
        Resource string   // "traces", "costs", "api_keys"
        Action   string   // "read", "write", "delete"
        Scope    *Scope   // Optional: team_id, project_id
    }

    func CheckPermission(user User, perm Permission) bool {
        // 1. Get user's roles (org + team)
        // 2. Get permissions for each role
        // 3. Check if any permission matches
        // 4. Apply scope restrictions
        // 5. Return result
    }
    ```

  policy_storage:
    location: "PostgreSQL"
    caching: "Redis (1 minute TTL)"
    invalidation: "On policy change"
```

### Encryption

```yaml
encryption:

  in_transit:
    protocol: "TLS 1.3"
    ciphers:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_CHACHA20_POLY1305_SHA256"
    certificate_management: "AWS ACM"
    hsts:
      enabled: true
      max_age: 31536000
      include_subdomains: true

  at_rest:
    databases:
      postgresql: "AWS RDS encryption (AES-256)"
      clickhouse: "Disk encryption (AES-256)"
      redis: "ElastiCache encryption (AES-256)"

    storage:
      s3: "SSE-S3 (AES-256)"
      ebs: "EBS encryption (AES-256)"

    application_level:
      sensitive_fields:
        - "api_key_hash (bcrypt)"
        - "sso_credentials (AES-256-GCM)"
        - "webhook_secrets (AES-256-GCM)"

  key_management:
    service: "AWS KMS"
    key_rotation: "Annual (automatic)"
    key_types:
      - "CMK for RDS encryption"
      - "CMK for S3 encryption"
      - "CMK for application secrets"
```

### Data Masking

```yaml
data_masking:

  automatic_patterns:
    credit_card:
      pattern: '\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b'
      replacement: "****-****-****-{last4}"

    ssn:
      pattern: '\b\d{3}-\d{2}-\d{4}\b'
      replacement: "***-**-****"

    api_key:
      pattern: '\b(gwo_[a-z]+_)[a-zA-Z0-9]{28}\b'
      replacement: "{prefix}****"

    email:
      pattern: '\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b'
      replacement: "{first2}***@{domain}"
      configurable: true  # Can be disabled

  implementation:
    timing: "Before storage (not reversible)"
    location: "Gateway service, Trace service"
    bypass: "Never - masking is permanent"
```

### Audit Logging

```yaml
audit_logging:

  events_captured:
    authentication:
      - "user.login"
      - "user.logout"
      - "user.login_failed"
      - "api_key.used"
      - "api_key.invalid"

    authorization:
      - "permission.granted"
      - "permission.denied"

    resource_changes:
      - "api_key.created"
      - "api_key.rotated"
      - "api_key.revoked"
      - "user.created"
      - "user.updated"
      - "user.deleted"
      - "team.created"
      - "budget.created"
      - "mcp_server.configured"

    system:
      - "config.changed"
      - "export.requested"

  log_format:
    ```json
    {
      "timestamp": "2025-01-15T10:30:00.000Z",
      "event_type": "api_key.created",
      "actor": {
        "type": "user",
        "id": "user_123",
        "email": "admin@company.com",
        "ip_address": "192.168.1.1"
      },
      "resource": {
        "type": "api_key",
        "id": "key_456",
        "org_id": "org_789"
      },
      "details": {
        "key_name": "Production Agent",
        "scopes": ["filesystem", "database"]
      },
      "request_id": "req_abc123"
    }
    ```

  retention:
    default: "1 year"
    max: "7 years"
    storage: "PostgreSQL (partitioned) → S3 (archived)"

  immutability:
    method: "Append-only table, no UPDATE/DELETE"
    verification: "Periodic integrity checks"
```

---

## Scalability & Performance

### Performance Requirements

```yaml
performance_requirements:

  gateway:
    latency:
      p50: "< 2ms overhead"
      p95: "< 5ms overhead"
      p99: "< 10ms overhead"
    throughput:
      per_node: "10,000 requests/second"
      cluster: "100,000 requests/second"
    connections:
      max_concurrent: "50,000 per node"

  api:
    latency:
      p50: "< 50ms"
      p95: "< 200ms"
      p99: "< 500ms"
    throughput: "1,000 requests/second"

  dashboard:
    page_load: "< 2 seconds"
    time_to_interactive: "< 3 seconds"
    api_response: "< 500ms"
```

### Scaling Strategy

```yaml
scaling_strategy:

  horizontal:
    gateway:
      trigger: "CPU > 70% OR Memory > 80% OR requests/sec > 8000"
      scale_up: "+2 pods"
      scale_down: "-1 pod (after 5 min cooldown)"
      min: 3
      max: 50

    services:
      trigger: "CPU > 70%"
      scale_up: "+1 pod"
      min: 2
      max: 10

  vertical:
    databases:
      postgresql:
        current: "db.r6g.large"
        upgrade_path: ["db.r6g.xlarge", "db.r6g.2xlarge"]
        trigger: "CPU > 80% sustained"

      redis:
        current: "cache.r6g.large"
        upgrade_path: ["cache.r6g.xlarge"]
        trigger: "Memory > 80%"

  data:
    clickhouse:
      sharding: "By org_id"
      replication: "3 replicas"
      partitioning: "By month"

    postgresql:
      read_replicas: 1 (add more as needed)
      connection_pooling: "PgBouncer"
```

### Performance Optimization

```yaml
performance_optimizations:

  gateway:
    - "Connection pooling to MCP servers"
    - "HTTP/2 multiplexing"
    - "Async logging (buffered writes)"
    - "Zero-copy response forwarding"
    - "Pre-computed rate limit buckets"

  caching:
    api_keys:
      location: "Redis"
      ttl: "5 minutes"
      hit_rate_target: "> 95%"

    permissions:
      location: "Redis"
      ttl: "1 minute"

    mcp_server_config:
      location: "In-memory"
      ttl: "1 minute"
      invalidation: "Pub/sub"

    dashboard_data:
      location: "Redis"
      ttl: "60 seconds"
      strategy: "Stale-while-revalidate"

  database:
    postgresql:
      - "Connection pooling (PgBouncer)"
      - "Prepared statements"
      - "Appropriate indexes"
      - "Query optimization"

    clickhouse:
      - "Materialized views for aggregations"
      - "Appropriate ORDER BY for queries"
      - "Compression (ZSTD)"
      - "Skip indexes"

  frontend:
    - "Code splitting"
    - "Image optimization (Next.js)"
    - "CDN for static assets"
    - "Prefetching on hover"
```

### Load Testing

```yaml
load_testing:

  tool: "k6"

  scenarios:
    baseline:
      vus: 100
      duration: "10m"
      target: "Establish baseline metrics"

    stress:
      stages:
        - vus: 100, duration: "2m"
        - vus: 500, duration: "5m"
        - vus: 1000, duration: "5m"
        - vus: 100, duration: "2m"
      target: "Find breaking point"

    soak:
      vus: 200
      duration: "2h"
      target: "Check for memory leaks, degradation"

    spike:
      stages:
        - vus: 100, duration: "1m"
        - vus: 2000, duration: "30s"
        - vus: 100, duration: "1m"
      target: "Test auto-scaling"

  thresholds:
    - "http_req_duration{p(95)} < 100ms"
    - "http_req_failed < 0.1%"
    - "checks > 99%"

  frequency: "Weekly (automated), before major releases"
```

---

## Reliability & Availability

### Availability Targets

```yaml
availability_targets:

  gateway:
    target: "99.9%"
    monthly_downtime: "43.8 minutes"
    error_budget: "0.1%"

  dashboard:
    target: "99.5%"
    monthly_downtime: "3.6 hours"

  api:
    target: "99.9%"
    monthly_downtime: "43.8 minutes"
```

### High Availability Design

```yaml
high_availability:

  compute:
    multi_az: true
    min_replicas: 3
    pod_anti_affinity: "Spread across AZs"
    pod_disruption_budget: "min_available: 2"

  databases:
    postgresql:
      type: "Aurora"
      multi_az: true
      failover: "Automatic (< 30 seconds)"
      read_replicas: 1

    redis:
      type: "ElastiCache Cluster"
      nodes: 3
      automatic_failover: true

    clickhouse:
      replication: 3
      distributed_ddl: true

  load_balancing:
    external: "AWS ALB (multi-AZ)"
    internal: "Kubernetes Services"
    health_checks:
      path: "/health"
      interval: "10s"
      threshold: 3
```

### Failure Scenarios

```yaml
failure_scenarios:

  single_pod_failure:
    impact: "None (other pods handle traffic)"
    recovery: "Automatic (Kubernetes restarts)"
    time: "< 30 seconds"

  single_az_failure:
    impact: "Reduced capacity (2/3 pods)"
    recovery: "Automatic (traffic routes to other AZs)"
    time: "< 1 minute"

  database_primary_failure:
    impact: "Brief write unavailability"
    recovery: "Automatic failover to replica"
    time: "< 30 seconds (Aurora)"

  redis_node_failure:
    impact: "Brief cache miss spike"
    recovery: "Automatic failover"
    time: "< 10 seconds"

  mcp_server_failure:
    impact: "Requests to that server fail"
    recovery: "Circuit breaker opens, clear error to client"
    time: "Immediate (graceful)"

  region_failure:
    impact: "Full service outage"
    recovery: "Manual failover to DR region (future)"
    time: "< 4 hours (RTO)"
```

### Circuit Breakers

```yaml
circuit_breakers:

  mcp_server:
    failure_threshold: 5  # failures in window
    window: "30 seconds"
    open_duration: "30 seconds"
    half_open_requests: 3

  auth_service:
    failure_threshold: 3
    window: "10 seconds"
    open_duration: "10 seconds"
    fallback: "Reject with 503"

  database:
    failure_threshold: 5
    window: "30 seconds"
    open_duration: "60 seconds"
    fallback: "Serve from cache if possible"
```

---

## Observability

### Metrics

```yaml
metrics:

  collection: "Prometheus + Datadog"

  gateway_metrics:
    - name: "gateway_requests_total"
      type: "counter"
      labels: ["server", "tool", "status"]

    - name: "gateway_request_duration_seconds"
      type: "histogram"
      labels: ["server", "tool"]
      buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]

    - name: "gateway_request_size_bytes"
      type: "histogram"
      labels: ["direction"]

    - name: "gateway_active_connections"
      type: "gauge"
      labels: ["server"]

    - name: "gateway_rate_limit_hits_total"
      type: "counter"
      labels: ["scope", "scope_id"]

    - name: "gateway_cost_total"
      type: "counter"
      labels: ["org_id", "team_id", "server"]

  service_metrics:
    - name: "http_server_requests_total"
      type: "counter"
      labels: ["method", "path", "status"]

    - name: "http_server_request_duration_seconds"
      type: "histogram"

    - name: "db_query_duration_seconds"
      type: "histogram"
      labels: ["operation"]

    - name: "cache_hits_total"
      type: "counter"
      labels: ["cache"]

    - name: "cache_misses_total"
      type: "counter"
      labels: ["cache"]

  business_metrics:
    - name: "active_organizations"
      type: "gauge"

    - name: "api_keys_active"
      type: "gauge"

    - name: "traces_ingested_total"
      type: "counter"

    - name: "cost_events_total"
      type: "counter"
```

### Logging

```yaml
logging:

  format: "JSON (structured)"

  standard_fields:
    - "timestamp"
    - "level"
    - "message"
    - "service"
    - "trace_id"
    - "request_id"
    - "org_id"
    - "user_id"

  log_levels:
    production: "info"
    staging: "debug"
    development: "debug"

  example:
    ```json
    {
      "timestamp": "2025-01-15T10:30:00.000Z",
      "level": "info",
      "message": "MCP request completed",
      "service": "gateway",
      "trace_id": "tr_abc123",
      "request_id": "req_xyz789",
      "org_id": "org_456",
      "user_id": "user_123",
      "mcp_server": "filesystem",
      "tool": "read_file",
      "duration_ms": 150,
      "status": "success",
      "cost": 0.001
    }
    ```

  aggregation:
    tool: "Datadog Logs"
    retention: "30 days (hot), 90 days (archive)"
    indexing: "By trace_id, org_id, level"
```

### Tracing

```yaml
tracing:

  protocol: "OpenTelemetry"

  spans:
    gateway:
      - "gateway.request"
      - "gateway.auth"
      - "gateway.rate_limit"
      - "gateway.mcp_proxy"
      - "gateway.log"

    services:
      - "auth.validate_key"
      - "auth.check_permission"
      - "cost.record_event"
      - "trace.ingest_span"
      - "db.query"
      - "cache.get"

  context_propagation:
    headers:
      - "X-Trace-ID"
      - "X-Request-ID"
      - "traceparent (W3C)"

  sampling:
    production: "100% (we need full visibility)"
    staging: "100%"

  export:
    internal: "ClickHouse (for product features)"
    external: "Datadog APM (for operations)"
```

### Alerting

```yaml
alerting:

  tool: "PagerDuty + Datadog"

  alert_levels:
    critical:
      response_time: "5 minutes"
      escalation: "15 minutes"
      notification: "PagerDuty (phone)"

    high:
      response_time: "30 minutes"
      escalation: "1 hour"
      notification: "PagerDuty (push) + Slack"

    medium:
      response_time: "4 hours"
      notification: "Slack"

    low:
      response_time: "24 hours"
      notification: "Email"

  alerts:
    - name: "Gateway Error Rate High"
      condition: "error_rate > 1% for 5 minutes"
      severity: "critical"

    - name: "Gateway Latency Degraded"
      condition: "p99_latency > 100ms for 5 minutes"
      severity: "high"

    - name: "Database Connection Pool Exhausted"
      condition: "available_connections < 10"
      severity: "critical"

    - name: "Redis Memory High"
      condition: "memory_usage > 80%"
      severity: "high"

    - name: "Certificate Expiring"
      condition: "days_until_expiry < 14"
      severity: "medium"

    - name: "Disk Space Low"
      condition: "disk_usage > 80%"
      severity: "high"

  runbooks:
    location: "Notion/Confluence"
    linked_in_alerts: true
```

---

## Development & Deployment

### Development Environment

```yaml
development_environment:

  local_setup:
    prerequisites:
      - "Docker Desktop"
      - "Go 1.21+"
      - "Node.js 20+"
      - "kubectl"
      - "helm"

    services:
      ```yaml
      # docker-compose.yml
      services:
        postgres:
          image: postgres:15
          ports: ["5432:5432"]
          environment:
            POSTGRES_DB: gatewayops
            POSTGRES_PASSWORD: dev

        redis:
          image: redis:7
          ports: ["6379:6379"]

        clickhouse:
          image: clickhouse/clickhouse-server:23
          ports: ["8123:8123", "9000:9000"]

        # Mock MCP server for testing
        mcp-mock:
          build: ./test/mcp-mock
          ports: ["3000:3000"]
      ```

  ide_setup:
    recommended: "VS Code"
    extensions:
      - "Go"
      - "ESLint"
      - "Prettier"
      - "Docker"
      - "Kubernetes"

  testing:
    unit:
      go: "go test ./..."
      typescript: "npm test"

    integration:
      tool: "Docker Compose"
      command: "make test-integration"

    e2e:
      tool: "Playwright"
      command: "npm run test:e2e"
```

### CI/CD Pipeline

```yaml
ci_cd:

  tool: "GitHub Actions"

  workflows:
    pull_request:
      triggers: ["pull_request"]
      jobs:
        - name: "lint"
          steps:
            - "golangci-lint"
            - "eslint"
            - "prettier --check"

        - name: "test"
          steps:
            - "go test ./..."
            - "npm test"

        - name: "build"
          steps:
            - "docker build"

        - name: "security"
          steps:
            - "snyk test"
            - "trivy scan"

    main:
      triggers: ["push to main"]
      jobs:
        - name: "test"
        - name: "build"
        - name: "push"
          steps:
            - "docker push to ECR"
            - "update image tag"

        - name: "deploy-staging"
          steps:
            - "ArgoCD sync staging"
            - "Run smoke tests"

    release:
      triggers: ["tag v*"]
      jobs:
        - name: "test"
        - name: "build"
        - name: "push"
        - name: "deploy-staging"
        - name: "approval"
          type: "manual"
        - name: "deploy-production"
          steps:
            - "ArgoCD sync production"
            - "Run smoke tests"
            - "Monitor metrics"

  deployment:
    tool: "ArgoCD"
    strategy: "GitOps"
    environments:
      staging:
        auto_sync: true
        branch: "main"

      production:
        auto_sync: false  # Manual approval
        branch: "main"
        tag_required: true
```

### Release Process

```yaml
release_process:

  versioning: "Semantic Versioning"
  format: "v{major}.{minor}.{patch}"

  process:
    1. "Create release branch from main"
    2. "Update changelog"
    3. "Create PR for release"
    4. "Merge to main"
    5. "Tag release (triggers deploy)"
    6. "Deploy to staging"
    7. "Run regression tests"
    8. "Manual approval"
    9. "Deploy to production"
    10. "Monitor for 30 minutes"
    11. "Announce release"

  rollback:
    trigger: "Error rate > 5% OR manual"
    process:
      1. "Revert ArgoCD to previous revision"
      2. "Verify rollback successful"
      3. "Investigate root cause"
      4. "Create hotfix if needed"

  feature_flags:
    tool: "LaunchDarkly (or simple config)"
    use_cases:
      - "Gradual rollout"
      - "Kill switch for features"
      - "A/B testing"
```

---

## Disaster Recovery

### Backup Strategy

```yaml
backup_strategy:

  postgresql:
    type: "Aurora automated backups"
    frequency: "Continuous (point-in-time recovery)"
    retention: "7 days"
    cross_region: "Daily snapshot to us-west-2"

  clickhouse:
    type: "Full + incremental"
    frequency: "Daily full, hourly incremental"
    retention: "7 days"
    destination: "S3"

  redis:
    type: "RDB snapshots"
    frequency: "Hourly"
    retention: "24 hours"

  secrets:
    type: "AWS Secrets Manager versioning"
    retention: "30 days"

  configuration:
    type: "Git (Infrastructure as Code)"
    retention: "Indefinite"
```

### Recovery Procedures

```yaml
recovery_procedures:

  rpo: "< 1 hour"  # Recovery Point Objective
  rto: "< 4 hours"  # Recovery Time Objective

  scenarios:
    database_corruption:
      procedure:
        1. "Identify corruption scope"
        2. "Stop writes to affected tables"
        3. "Restore from point-in-time backup"
        4. "Verify data integrity"
        5. "Resume operations"
      estimated_time: "1-2 hours"

    accidental_deletion:
      procedure:
        1. "Identify deleted data"
        2. "Restore from backup to temp table"
        3. "Verify restored data"
        4. "Merge back to production"
      estimated_time: "30 minutes - 2 hours"

    region_failure:
      procedure:
        1. "Confirm region outage"
        2. "Update DNS to DR region"
        3. "Restore databases from cross-region backups"
        4. "Deploy application to DR region"
        5. "Verify functionality"
        6. "Communicate to customers"
      estimated_time: "2-4 hours"

    security_breach:
      procedure:
        1. "Contain: Revoke compromised credentials"
        2. "Assess: Determine scope of breach"
        3. "Eradicate: Remove attacker access"
        4. "Recover: Restore from clean backup if needed"
        5. "Communicate: Notify affected parties"
        6. "Review: Post-incident analysis"
      estimated_time: "Variable"

  testing:
    frequency: "Quarterly"
    scope:
      - "Database restore drill"
      - "Failover test"
      - "Runbook verification"
```

---

## Technology Decisions

### Architecture Decision Records (ADRs)

#### ADR-001: Go for Gateway Service

```yaml
adr:
  id: "ADR-001"
  title: "Use Go for Gateway Service"
  status: "Accepted"
  date: "2025-01-01"

  context: |
    The gateway service is the most performance-critical component.
    It must handle high throughput with low latency overhead.

  decision: |
    Use Go 1.21+ for the gateway service.

  rationale:
    - "Excellent performance (comparable to C/C++)"
    - "Built-in concurrency with goroutines"
    - "Low memory footprint"
    - "Fast compilation"
    - "Strong standard library for HTTP"
    - "Easy deployment (single binary)"

  alternatives_considered:
    - "Rust: Better performance but higher complexity"
    - "Node.js: Good ecosystem but GC pauses"
    - "Java: Mature but higher memory usage"

  consequences:
    positive:
      - "Sub-millisecond latency overhead"
      - "Easy horizontal scaling"
      - "Simple deployment"
    negative:
      - "Smaller ecosystem than Node.js"
      - "Error handling verbosity"
```

#### ADR-002: ClickHouse for Time-Series Data

```yaml
adr:
  id: "ADR-002"
  title: "Use ClickHouse for Traces and Metrics"
  status: "Accepted"
  date: "2025-01-01"

  context: |
    We need to store and query billions of trace spans and cost events.
    Queries are mostly analytical (aggregations, time-range filters).

  decision: |
    Use ClickHouse for traces, spans, cost events, and metrics.

  rationale:
    - "Columnar storage optimized for analytics"
    - "Excellent compression (10-20x)"
    - "Fast aggregation queries"
    - "Native time-series support"
    - "SQL interface (familiar)"
    - "Scales horizontally"

  alternatives_considered:
    - "TimescaleDB: Good but PostgreSQL limitations"
    - "Elasticsearch: Expensive, complex"
    - "Cassandra: Write-optimized, weak analytics"
    - "InfluxDB: Good for metrics, not for traces"

  consequences:
    positive:
      - "Fast dashboards and queries"
      - "Cost-effective storage"
      - "Easy aggregations"
    negative:
      - "Another database to operate"
      - "Learning curve for team"
```

#### ADR-003: Async Event Processing

```yaml
adr:
  id: "ADR-003"
  title: "Async Processing for Cost and Trace Events"
  status: "Accepted"
  date: "2025-01-01"

  context: |
    Recording costs and traces should not add latency to the critical path.
    Events must be durable (not lost if service restarts).

  decision: |
    Use Redis Streams for async event processing.

  rationale:
    - "Low latency publishing"
    - "Durable (persisted to disk)"
    - "Consumer groups for scaling"
    - "Already using Redis for caching"
    - "Simple operations"

  alternatives_considered:
    - "Kafka: More powerful but overkill for our scale"
    - "SQS: Good but higher latency"
    - "Direct writes: Adds latency to critical path"

  consequences:
    positive:
      - "Zero latency impact on gateway"
      - "Durable event delivery"
      - "Simple scaling"
    negative:
      - "Slight delay in dashboard updates (< 1 second)"
      - "Redis becomes more critical"
```

### Technology Radar

```yaml
technology_radar:
  date: "2025-01"

  adopt:
    - technology: "Go"
      use_case: "Backend services"

    - technology: "Next.js"
      use_case: "Web application"

    - technology: "ClickHouse"
      use_case: "Analytics database"

    - technology: "PostgreSQL"
      use_case: "Transactional database"

    - technology: "Redis"
      use_case: "Caching, rate limiting"

    - technology: "Kubernetes"
      use_case: "Container orchestration"

  trial:
    - technology: "OpenTelemetry"
      use_case: "Tracing export"
      evaluation: "Testing OTLP export integration"

  assess:
    - technology: "Rust"
      use_case: "Performance-critical components"
      notes: "Evaluate if Go performance becomes insufficient"

    - technology: "Neon"
      use_case: "Serverless PostgreSQL"
      notes: "Could simplify scaling"

  hold:
    - technology: "MongoDB"
      reason: "PostgreSQL and ClickHouse cover our needs"

    - technology: "GraphQL"
      reason: "REST is simpler for our use cases"
```

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-30 | CTO Agent | Initial architecture document |

---

## Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| CTO | CTO Agent | | Pending |
| CEO | CEO Agent | | Pending |
| Security | Security Agent | | Pending |
