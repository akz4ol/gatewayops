# GatewayOps Technical & Feature Debt

**Last Updated:** 2026-01-03

## Summary

| Category | Count | Priority |
|----------|-------|----------|
| Critical (Breaking) | 2 | P0 |
| Technical Debt | 5 | P1 |
| Feature Debt | 12 | P2 |

---

## Critical Issues (P0)

### 1. Python SDK Type Mismatches - BREAKING

**Status:** SDK fails when parsing API responses

**Problem:**
The Python SDK Pydantic models don't match the actual API response schemas.

**Traces API:**
```python
# SDK expects:
class TracePage(BaseModel):
    traces: List[Trace]           # Required, non-null
    has_more: bool = Field(alias="hasMore")  # Required

# API returns:
{
    "traces": null,  # Can be null when empty
    "total": 0,
    "limit": 20,
    "offset": 0
    # No "hasMore" field
}
```

**Costs API:**
```python
# SDK expects:
class CostSummary(BaseModel):
    total_cost: float = Field(alias="totalCost")
    period_start: datetime = Field(alias="periodStart")
    period_end: datetime = Field(alias="periodEnd")
    request_count: int = Field(alias="requestCount")

# API returns:
{
    "total_cost": 0,           # snake_case, not camelCase
    "total_requests": 0,       # different field name
    "avg_cost_per_request": 0,
    "period": "month",
    "start_date": "...",       # start_date, not periodStart
    "end_date": "..."          # end_date, not periodEnd
}
```

**Fix Required:**
- Update `sdks/python/gatewayops/types.py` to match actual API response schemas
- Use snake_case field names directly (no aliases needed)
- Make `traces` field `Optional[List[Trace]]` with default `[]`
- Remove `has_more` field or make it optional with default

---

### 2. TypeScript SDK Type Mismatches

**Status:** Types don't match API responses (may cause runtime issues)

**Problem:**
TypeScript SDK types expect camelCase but API returns snake_case.

```typescript
// SDK expects:
interface CostSummary {
  totalCost: number;
  periodStart: Date;
  periodEnd: Date;
  requestCount: number;
}

// API returns:
{
  "total_cost": 0,
  "total_requests": 0,
  "start_date": "...",
  "end_date": "..."
}
```

**Fix Required:**
- Update TypeScript types to match actual API response
- Or add response transformation layer in client

---

## Technical Debt (P1)

### 3. No Response Transformation Layer

Both SDKs directly parse API responses without transformation. Should add a mapping layer to:
- Handle null → empty array conversion
- Transform snake_case to camelCase if desired
- Validate response structure

### 4. Missing Error Handling for Empty Responses

When database has no data, API returns `null` for arrays. SDKs should handle this gracefully.

### 5. Hardcoded Demo Data in Handlers

Multiple handlers have fallback demo data when repo is nil. This is fine for demo but:
- Should be feature-flagged
- Demo mode should be explicit configuration
- Production should fail fast if repo is unavailable

**Files affected:**
- `gateway/internal/handler/trace_handler.go`
- `gateway/internal/handler/cost_handler.go`
- `gateway/internal/handler/apikey_handler.go`

### 6. Missing SDK Version Sync

SDK versions are hardcoded in User-Agent headers:
- Python: `gatewayops-python/0.1.0`
- TypeScript: `gatewayops-typescript/0.1.0`

Should be dynamically read from package version.

### 7. No SDK Tests

Neither SDK has unit or integration tests:
- `sdks/python/` - No tests directory
- `sdks/typescript/` - No tests directory

---

## Feature Debt (P2) - Missing SDK Methods

### API Endpoints Not Exposed in SDKs

| Endpoint | Description | Python SDK | TypeScript SDK |
|----------|-------------|------------|----------------|
| `/v1/api-keys` | API Key CRUD | Missing | Missing |
| `/v1/metrics/*` | Dashboard metrics | Missing | Missing |
| `/v1/safety/*` | Safety policies & detection | Missing | Missing |
| `/v1/audit-logs/*` | Audit log queries | Missing | Missing |
| `/v1/alerts/*` | Alert rules & channels | Missing | Missing |
| `/v1/telemetry/*` | OpenTelemetry config | Missing | Missing |
| `/v1/approvals/*` | Tool approval workflow | Missing | Missing |
| `/v1/tool-classifications/*` | Tool classification | Missing | Missing |
| `/v1/tool-permissions/*` | Tool permissions | Missing | Missing |
| `/v1/rbac/*` | Roles & permissions | Missing | Missing |
| `/v1/users/*` | User management | Missing | Missing |
| `/v1/sso/*` | SSO provider config | Missing | Missing |
| `/v1/settings` | Organization settings | Missing | Missing |
| `/v1/traces/stats` | Trace statistics | Missing | Missing |
| `/v1/costs/by-server` | Cost by server | Missing | Missing |
| `/v1/costs/by-team` | Cost by team | Missing | Missing |
| `/v1/costs/daily` | Daily cost data | Missing | Missing |

### SDK Methods to Add

```python
# Python SDK - Missing methods
gw.api_keys.list()
gw.api_keys.create(name, environment, permissions)
gw.api_keys.delete(key_id)
gw.api_keys.rotate(key_id)

gw.safety.policies.list()
gw.safety.policies.create(...)
gw.safety.test(input_text)
gw.safety.detections.list()

gw.audit.list(filters)
gw.audit.search(query)
gw.audit.export(format)

gw.alerts.list()
gw.alerts.rules.create(...)
gw.alerts.channels.create(...)

gw.rbac.roles.list()
gw.rbac.permissions.check(user_id, permission)

gw.users.list()
gw.users.invite(email)

gw.settings.get()
gw.settings.update(...)
```

---

## Recommended Fix Priority

### Phase 1: Fix Breaking Issues (1-2 days)
1. Fix Python SDK types to match API responses
2. Fix TypeScript SDK types to match API responses
3. Add null handling for empty arrays

### Phase 2: Add Missing Core Methods (3-4 days)
1. Add API Keys management to SDKs
2. Add Traces stats endpoint
3. Add Costs by-server, by-team, daily endpoints
4. Add SDK tests

### Phase 3: Add Admin Methods (5-7 days)
1. Add Safety policies & detection
2. Add Audit logs
3. Add Alerts
4. Add RBAC
5. Add Users & SSO
6. Add Settings

---

## Files to Modify

### Python SDK
- `sdks/python/gatewayops/types.py` - Fix type definitions
- `sdks/python/gatewayops/client.py` - Add missing clients
- `sdks/python/tests/` - Add tests (new)

### TypeScript SDK
- `sdks/typescript/src/types/traces.ts` - Fix type definitions
- `sdks/typescript/src/types/costs.ts` - Fix type definitions
- `sdks/typescript/src/client.ts` - Add missing clients
- `sdks/typescript/tests/` - Add tests (new)

### Gateway
- Consider adding response transformation to ensure consistent API shape
- Add explicit demo mode configuration
