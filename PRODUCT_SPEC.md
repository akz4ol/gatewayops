# GatewayOps Product Specification
## From Demo to Production-Ready

**Document Version:** 1.0
**Date:** January 2026
**Status:** Implementation Roadmap

---

## Executive Summary

GatewayOps is an Enterprise MCP Gateway that provides authentication, authorization, observability, and security for Model Context Protocol deployments. The current implementation is ~70% complete architecturally but operates entirely in demo/mock mode.

**Goal:** Transform GatewayOps from a demo system to a fully functional, production-ready enterprise product.

---

## Current State Analysis

### What's Production-Ready (Works Today)
| Component | Status | Notes |
|-----------|--------|-------|
| MCP Proxy | ✅ Working | HTTP forwarding to MCP servers |
| Dashboard UI | ✅ Working | All pages render, properly wired to APIs |
| Database Schema | ✅ Complete | 4 migrations with full schema |
| Alerting Service | ✅ In-memory | Full CRUD + notifications (Slack, PagerDuty) |
| Safety Detector | ✅ In-memory | Prompt injection detection patterns |
| Approval Service | ✅ In-memory | Tool classification & approval workflows |
| SSO Service | ✅ In-memory | OIDC providers (Okta, Azure AD, Google, Auth0) |
| Audit Logger | ✅ In-memory | Full audit trail with export |
| API Client (Dashboard) | ✅ Working | SWR hooks, proper error handling |

### What's Demo/Mock Only (Needs Real Implementation)
| Component | Current State | Required Work |
|-----------|---------------|---------------|
| PostgreSQL | Returns `nil` connection | Connect real DB |
| Redis | Returns `nil` client | Connect for rate limiting |
| Trace Handler | Generates fake data | Query real traces from DB |
| Cost Handler | Generates fake data | Query real costs from DB |
| Metrics Handler | Generates fake data | Query real metrics from DB |
| API Key Handler | List/Get return fake data | Query real keys from DB |
| Trace Publishing | TODO in MCP handler | Publish to Redis Streams |
| Cost Publishing | TODO in MCP handler | Publish to database |

### What's Completely Missing
| Feature | Priority | Notes |
|---------|----------|-------|
| User/Team Management | P1 | CRUD for users, teams, invitations |
| Real-time Updates | P2 | WebSocket for live trace streaming |
| Metrics Aggregation | P2 | Background job for rollups |
| Data Retention | P2 | Automatic cleanup of old data |
| Backup/Restore | P3 | Data backup mechanisms |

---

## Architecture Overview

### Target Production Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Clients                               │
│            (Claude Desktop, Claude Code, Custom Apps)            │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      GatewayOps API                              │
│                    (Go / Chi Router)                             │
├─────────────────────────────────────────────────────────────────┤
│  Middleware Stack:                                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ Auth/SSO │→│   RBAC   │→│Rate Limit│→│ Injection│→ Handler  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
├─────────────────────────────────────────────────────────────────┤
│  Services:                                                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │   SSO    │ │ Alerting │ │  Safety  │ │ Approval │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
└──────────┬──────────────────────────────────────────────────────┘
           │
           ├──────────────────┬──────────────────┐
           ▼                  ▼                  ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│   PostgreSQL     │ │      Redis       │ │   MCP Servers    │
│  (Primary Data)  │ │  (Cache/Streams) │ │  (Filesystem,    │
│                  │ │                  │ │   GitHub, etc.)  │
└──────────────────┘ └──────────────────┘ └──────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Dashboard (Next.js)                           │
│                 https://dashboard.gatewayops.io                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Detailed Feature Specifications

### Phase 1: Database Integration (Critical Path)

#### 1.1 PostgreSQL Connection
**Priority:** P0 - Blocking everything else
**Effort:** 2 days

**Current State:**
```go
// gateway/internal/database/postgres.go:21
// Demo mode: Using mock PostgreSQL connection
return &PostgresDB{db: nil}, nil
```

**Required Changes:**
1. Remove mock connection, use real `sql.Open()`
2. Add connection pooling configuration
3. Add health check with actual query
4. Run migrations on startup

**Files to Modify:**
- `gateway/internal/database/postgres.go`
- `gateway/cmd/gateway/main.go`

**Environment Variables:**
```bash
DATABASE_URL=postgres://user:pass@host:5432/gatewayops?sslmode=require
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE=5
```

#### 1.2 Redis Connection
**Priority:** P0
**Effort:** 1 day

**Current State:**
```go
// gateway/internal/database/redis.go:18
// Demo mode: Using mock Redis connection
return &RedisDB{client: nil}, nil
```

**Required Changes:**
1. Remove mock connection, use real `redis.NewClient()`
2. Configure for rate limiting
3. Configure for trace stream publishing

**Files to Modify:**
- `gateway/internal/database/redis.go`
- `gateway/cmd/gateway/main.go`

---

### Phase 2: Data Persistence Layer

#### 2.1 Trace Storage & Retrieval
**Priority:** P0
**Effort:** 3 days

**Current State:**
- MCP handler calculates traces but doesn't store them
- Trace handler generates fake data

**Database Schema (Already Exists):**
```sql
CREATE TABLE traces (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    trace_id VARCHAR(100) NOT NULL,
    parent_span_id VARCHAR(100),
    span_id VARCHAR(100) NOT NULL,
    operation_name VARCHAR(255) NOT NULL,
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255),
    status VARCHAR(20) NOT NULL,
    duration_ms INTEGER NOT NULL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    cost_usd DECIMAL(10,6),
    user_id UUID,
    api_key_id UUID,
    metadata JSONB,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Implementation:**

1. **MCP Handler - Publish Trace** (`gateway/internal/handler/mcp.go`):
```go
// After proxying request, publish trace span
span := &domain.TraceSpan{
    ID:            uuid.New(),
    TraceID:       traceID,
    SpanID:        spanID,
    OperationName: operationName,
    MCPServer:     server,
    Status:        status,
    DurationMs:    duration.Milliseconds(),
    StartedAt:     startTime,
    EndedAt:       time.Now(),
}
h.traceRepo.Create(ctx, span)
```

2. **Trace Handler - Query Real Data** (`gateway/internal/handler/trace_handler.go`):
```go
func (h *TraceHandler) List(w http.ResponseWriter, r *http.Request) {
    // Replace generateSampleTraces() with:
    traces, total, err := h.repo.List(ctx, filters, limit, offset)
}
```

**Files to Modify:**
- `gateway/internal/handler/mcp.go` - Add trace publishing
- `gateway/internal/handler/trace_handler.go` - Use real queries
- `gateway/internal/repository/trace_repository.go` - Implement methods

#### 2.2 Cost Tracking & Retrieval
**Priority:** P0
**Effort:** 2 days

**Current State:**
- MCP handler calculates cost but doesn't store it
- Cost handler generates fake data

**Database Schema (Already Exists):**
```sql
CREATE TABLE costs (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    trace_id UUID REFERENCES traces(id),
    mcp_server VARCHAR(100) NOT NULL,
    tool_name VARCHAR(255),
    user_id UUID,
    team_id UUID,
    api_key_id UUID,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd DECIMAL(10,6) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Implementation:**

1. **MCP Handler - Record Cost**:
```go
cost := &domain.Cost{
    TraceID:      traceID,
    MCPServer:    server,
    InputTokens:  inputTokens,
    OutputTokens: outputTokens,
    CostUSD:      calculatedCost,
}
h.costRepo.Create(ctx, cost)
```

2. **Cost Handler - Query Real Data**:
```go
func (h *CostHandler) Summary(w http.ResponseWriter, r *http.Request) {
    // Replace hardcoded data with:
    summary, err := h.repo.GetSummary(ctx, orgID, period)
}
```

**Files to Modify:**
- `gateway/internal/handler/mcp.go` - Add cost recording
- `gateway/internal/handler/cost_handler.go` - Use real queries
- `gateway/internal/repository/cost_repository.go` - Implement methods

#### 2.3 API Key Persistence
**Priority:** P0
**Effort:** 1 day

**Current State:**
- Create generates real keys but doesn't persist
- List/Get return fake data

**Implementation:**
```go
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Current: generates key but only returns it
    // Add: h.repo.Create(ctx, apiKey)
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
    // Replace hardcoded keys with:
    keys, err := h.repo.ListByOrg(ctx, orgID)
}
```

**Files to Modify:**
- `gateway/internal/handler/apikey_handler.go`
- `gateway/internal/repository/apikey_repository.go` (already has SQL)

#### 2.4 Service Persistence (Alerts, Safety, Approvals)
**Priority:** P1
**Effort:** 3 days

**Current State:**
All services store data in-memory maps with mutex locks.

**Required Changes:**
Add database persistence while keeping in-memory cache for performance.

```go
// Example: Alerting Service
func (s *Service) CreateRule(rule *domain.AlertRule) error {
    // 1. Save to database
    err := s.repo.Create(ctx, rule)
    if err != nil {
        return err
    }
    // 2. Update in-memory cache
    s.mu.Lock()
    s.rules[rule.ID] = rule
    s.mu.Unlock()
    return nil
}
```

**Files to Modify:**
- `gateway/internal/alerting/service.go`
- `gateway/internal/safety/detector.go`
- `gateway/internal/approval/service.go`
- `gateway/internal/audit/logger.go`
- `gateway/internal/sso/service.go`

---

### Phase 3: Dashboard Integration

#### 3.1 Alerts Page - Connect to Backend
**Priority:** P1
**Effort:** 1 day

**Current State:**
```typescript
// dashboard/app/(dashboard)/alerts/page.tsx
const alertRules = [
  { id: 1, name: 'High Error Rate', ... }, // Hardcoded
];
```

**Required Changes:**

1. **Create API hooks** (`dashboard/lib/hooks/use-api.ts`):
```typescript
export function useAlertRules() {
  return useSWR<AlertRulesResponse>(
    'alert-rules',
    () => fetcher(() => api.listAlertRules())
  );
}

export function useActiveAlerts() {
  return useSWR<AlertsResponse>(
    'alerts',
    () => fetcher(() => api.listAlerts())
  );
}
```

2. **Add API client methods** (`dashboard/lib/api/client.ts`):
```typescript
async listAlertRules(): Promise<AlertRulesResponse> {
  return this.get('/alerts/rules');
}

async listAlerts(): Promise<AlertsResponse> {
  return this.get('/alerts');
}

async createAlertRule(data: CreateAlertRuleRequest): Promise<AlertRule> {
  return this.post('/alerts/rules', data);
}
```

3. **Update page to use hooks**:
```typescript
export default function AlertsPage() {
  const { data: rulesData, isLoading: rulesLoading } = useAlertRules();
  const { data: alertsData, isLoading: alertsLoading } = useActiveAlerts();

  // Replace hardcoded arrays with API data
  const alertRules = rulesData?.rules || [];
  const activeAlerts = alertsData?.alerts || [];
}
```

**Files to Modify:**
- `dashboard/lib/api/client.ts` - Add alert endpoints
- `dashboard/lib/hooks/use-api.ts` - Add alert hooks
- `dashboard/app/(dashboard)/alerts/page.tsx` - Use hooks

#### 3.2 Safety Page - Connect to Backend
**Priority:** P1
**Effort:** 1 day

**Current State:** Hardcoded mock data

**Required Changes:**
Same pattern as Alerts - add hooks and wire to API.

**API Endpoints Already Exist:**
- `GET /v1/safety-policies` - List policies
- `POST /v1/safety-policies` - Create policy
- `GET /v1/safety/detections` - List detections

**Files to Modify:**
- `dashboard/lib/api/client.ts`
- `dashboard/lib/hooks/use-api.ts`
- `dashboard/app/(dashboard)/safety/page.tsx`

#### 3.3 Team Page - Connect to Backend
**Priority:** P1
**Effort:** 2 days

**Current State:** Hardcoded mock data

**Required Work:**

1. **Backend**: Implement user/team handlers
   - `GET /v1/users` - List users
   - `POST /v1/users/invite` - Invite user
   - `GET /v1/teams` - List teams
   - `GET /v1/roles` - List roles (already exists)

2. **Frontend**: Wire to API

**Files to Create/Modify:**
- `gateway/internal/handler/user_handler.go` (new)
- `gateway/internal/handler/team_handler.go` (new)
- `dashboard/lib/api/client.ts`
- `dashboard/lib/hooks/use-api.ts`
- `dashboard/app/(dashboard)/team/page.tsx`

#### 3.4 Settings Page - Make Functional
**Priority:** P1
**Effort:** 2 days

**Current State:** Forms render but don't submit

**Required Changes:**

1. **Add form submission handlers**:
```typescript
const handleSaveOrganization = async () => {
  await api.updateOrganization({
    name: orgName,
    billing_email: billingEmail,
  });
};

const handleSaveSSO = async () => {
  await api.createSSOProvider({
    type: ssoProvider,
    issuer_url: issuerUrl,
    client_id: clientId,
    client_secret: clientSecret,
  });
};
```

2. **Add loading/success states**
3. **Add form validation**

**Files to Modify:**
- `dashboard/app/(dashboard)/settings/page.tsx`
- `dashboard/lib/api/client.ts` (add settings endpoints)

---

### Phase 4: Production Hardening

#### 4.1 Real-time Updates
**Priority:** P2
**Effort:** 3 days

**Implementation Options:**

**Option A: WebSocket (Recommended)**
```go
// gateway/internal/handler/websocket.go
func (h *WebSocketHandler) TraceStream(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    // Subscribe to Redis channel
    pubsub := h.redis.Subscribe(ctx, "traces:"+orgID)
    for msg := range pubsub.Channel() {
        conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
    }
}
```

**Option B: Server-Sent Events**
```go
func (h *SSEHandler) TraceStream(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    flusher := w.(http.Flusher)

    // Subscribe and stream
    for trace := range traceChannel {
        fmt.Fprintf(w, "data: %s\n\n", json.Marshal(trace))
        flusher.Flush()
    }
}
```

#### 4.2 Metrics Aggregation Pipeline
**Priority:** P2
**Effort:** 2 days

**Implementation:**
Background goroutine that periodically aggregates metrics:

```go
func (s *MetricsService) StartAggregator() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        // Aggregate last minute's traces
        s.aggregateTraces()
        // Roll up costs
        s.rollupCosts()
        // Update dashboard cache
        s.updateCache()
    }
}
```

#### 4.3 Data Retention
**Priority:** P2
**Effort:** 1 day

**Implementation:**
```go
func (s *RetentionService) Cleanup() {
    // Delete traces older than retention period
    s.db.Exec(`
        DELETE FROM traces
        WHERE created_at < NOW() - INTERVAL '90 days'
    `)
    // Archive to cold storage if needed
}
```

---

## Implementation Roadmap

### Week 1: Database Foundation
| Day | Task | Owner |
|-----|------|-------|
| 1-2 | Connect PostgreSQL, run migrations | Backend |
| 2-3 | Connect Redis for rate limiting | Backend |
| 3-4 | Implement trace persistence in MCP handler | Backend |
| 4-5 | Implement cost persistence in MCP handler | Backend |

### Week 2: Handler Updates
| Day | Task | Owner |
|-----|------|-------|
| 1-2 | Update trace handler to query real data | Backend |
| 2-3 | Update cost handler to query real data | Backend |
| 3-4 | Update API key handler with persistence | Backend |
| 4-5 | Add persistence to alerting/safety services | Backend |

### Week 3: Dashboard Completion
| Day | Task | Owner |
|-----|------|-------|
| 1 | Wire Alerts page to backend API | Frontend |
| 2 | Wire Safety page to backend API | Frontend |
| 3 | Implement Team page with user management | Full-stack |
| 4 | Make Settings page functional | Frontend |
| 5 | Testing and bug fixes | QA |

### Week 4: Production Hardening
| Day | Task | Owner |
|-----|------|-------|
| 1-2 | Implement WebSocket for real-time updates | Backend |
| 2-3 | Add metrics aggregation pipeline | Backend |
| 3-4 | Add data retention policies | Backend |
| 4-5 | Load testing and optimization | DevOps |

---

## Database Schema Reference

### Core Tables (Migration 001)
- `organizations` - Multi-tenant organizations
- `teams` - Teams within organizations
- `users` - User accounts
- `api_keys` - API key management
- `traces` - MCP request traces
- `costs` - Cost tracking

### SSO/RBAC Tables (Migration 002)
- `sso_providers` - OIDC provider configurations
- `user_sessions` - Active sessions
- `roles` - Permission roles
- `user_roles` - Role assignments

### Safety Tables (Migration 003)
- `safety_policies` - Injection detection policies
- `injection_detections` - Detection events
- `tool_classifications` - Tool risk levels
- `tool_approvals` - Approval requests

### Alerting Tables (Migration 004)
- `alert_rules` - Alert rule definitions
- `alert_channels` - Notification channels
- `alerts` - Active/historical alerts

---

## Environment Variables (Production)

```bash
# Database
DATABASE_URL=postgres://gatewayops:password@db.example.com:5432/gatewayops?sslmode=require
DB_MAX_CONNECTIONS=50
DB_MAX_IDLE=10

# Redis
REDIS_URL=redis://:password@redis.example.com:6379/0

# Auth
JWT_SECRET=<32-byte-secret>
API_KEY_BCRYPT_COST=12
SESSION_DURATION=24h

# MCP Servers
MCP_SERVER_FILESYSTEM_URL=http://mcp-filesystem:8080
MCP_SERVER_GITHUB_URL=http://mcp-github:8080
MCP_SERVER_DATABASE_URL=http://mcp-database:8080

# Observability
LOG_LEVEL=info
LOG_FORMAT=json
OTEL_EXPORTER_OTLP_ENDPOINT=https://otel-collector:4317

# External Services
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
PAGERDUTY_ROUTING_KEY=...
```

---

## Success Criteria

### Functional Requirements
- [ ] All data persists across restarts
- [ ] Traces show real MCP request data
- [ ] Costs accurately reflect actual usage
- [ ] API keys persist and validate correctly
- [ ] Alerts fire based on real metrics
- [ ] Safety policies block actual injection attempts
- [ ] Team management allows real user CRUD
- [ ] Settings changes persist

### Non-Functional Requirements
- [ ] P99 latency < 100ms for API requests
- [ ] Support 1000 concurrent MCP connections
- [ ] 99.9% uptime SLA
- [ ] Data retention for 90 days minimum
- [ ] Automatic failover for database

---

## Appendix: File Inventory

### Backend Files to Modify
```
gateway/internal/database/postgres.go     # Connect real DB
gateway/internal/database/redis.go        # Connect real Redis
gateway/internal/handler/mcp.go           # Add trace/cost publishing
gateway/internal/handler/trace_handler.go # Query real traces
gateway/internal/handler/cost_handler.go  # Query real costs
gateway/internal/handler/apikey_handler.go # Persist keys
gateway/internal/alerting/service.go      # Add DB persistence
gateway/internal/safety/detector.go       # Add DB persistence
gateway/internal/approval/service.go      # Add DB persistence
gateway/internal/sso/service.go           # Add DB persistence
gateway/cmd/gateway/main.go               # Wire real connections
```

### Dashboard Files to Modify
```
dashboard/lib/api/client.ts               # Add missing endpoints
dashboard/lib/hooks/use-api.ts            # Add missing hooks
dashboard/app/(dashboard)/alerts/page.tsx # Wire to API
dashboard/app/(dashboard)/safety/page.tsx # Wire to API
dashboard/app/(dashboard)/team/page.tsx   # Wire to API
dashboard/app/(dashboard)/settings/page.tsx # Add form handlers
```

### New Files to Create
```
gateway/internal/handler/user_handler.go  # User management
gateway/internal/handler/team_handler.go  # Team management
gateway/internal/handler/websocket.go     # Real-time updates
```
