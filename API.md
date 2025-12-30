# GatewayOps API Documentation

**Version:** 1.0.0
**Base URL:** `https://api.gatewayops.com/v1`
**Gateway URL:** `https://gateway.gatewayops.com/v1`
**Last Updated:** 2025-12-30

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Rate Limiting](#rate-limiting)
4. [Error Handling](#error-handling)
5. [Gateway API](#gateway-api)
6. [Traces API](#traces-api)
7. [Costs API](#costs-api)
8. [API Keys API](#api-keys-api)
9. [Budgets API](#budgets-api)
10. [Organizations API](#organizations-api)
11. [Teams API](#teams-api)
12. [MCP Servers API](#mcp-servers-api)
13. [Alerts API](#alerts-api)
14. [Webhooks API](#webhooks-api)
15. [SDKs](#sdks)
16. [Changelog](#changelog)

---

## Overview

### API Types

GatewayOps provides two API surfaces:

| API | Base URL | Purpose |
|-----|----------|---------|
| **Gateway API** | `gateway.gatewayops.com` | Proxy MCP requests to your servers |
| **Management API** | `api.gatewayops.com` | Manage configuration, view traces, costs |

### Request Format

- All requests use HTTPS
- Request bodies are JSON (`Content-Type: application/json`)
- UTF-8 encoding required
- Maximum request size: 10MB

### Response Format

All responses follow this structure:

```json
{
  "data": { },
  "meta": {
    "request_id": "req_abc123xyz",
    "timestamp": "2025-01-15T10:30:00.000Z"
  }
}
```

Error responses:

```json
{
  "error": {
    "code": "validation_error",
    "message": "Human-readable error message",
    "details": [],
    "request_id": "req_abc123xyz"
  }
}
```

### Pagination

List endpoints use cursor-based pagination:

```json
{
  "data": [...],
  "pagination": {
    "has_more": true,
    "cursor": "eyJpZCI6MTIzfQ=="
  }
}
```

Query parameters:
- `limit` - Items per page (default: 50, max: 200)
- `cursor` - Cursor from previous response

---

## Authentication

### API Keys

API keys authenticate requests to both Gateway and Management APIs.

**Format:** `gwo_{environment}_{32_random_chars}`

**Example:** `gwo_prod_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6`

**Header:**
```http
Authorization: Bearer gwo_prod_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6
```

### Key Types

| Type | Use Case | Permissions |
|------|----------|-------------|
| **Organization Key** | Backend services, CI/CD | Full org access |
| **Team Key** | Team-specific agents | Team resources only |
| **User Key** | Personal development | User's permissions |
| **Scoped Key** | Limited access | Specific MCP servers |

### Creating API Keys

```bash
curl -X POST https://api.gatewayops.com/v1/api-keys \
  -H "Authorization: Bearer gwo_prod_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Agent",
    "type": "organization",
    "expires_at": "2026-01-01T00:00:00Z",
    "scopes": {
      "mcp_servers": ["filesystem", "database"],
      "permissions": ["gateway:proxy", "traces:read"]
    }
  }'
```

### Session Authentication (Dashboard)

For web dashboard access, use SSO authentication:

1. Redirect to `/v1/auth/sso/{provider}/authorize`
2. User authenticates with IdP
3. Callback returns session token
4. Include in requests: `Authorization: Bearer session_xxx`

### Security Best Practices

- **Never expose keys in client-side code**
- **Rotate keys regularly** (every 90 days recommended)
- **Use scoped keys** for limited access needs
- **Set expiration dates** on all keys
- **Monitor key usage** in the dashboard

---

## Rate Limiting

### Default Limits

| Scope | Limit | Window |
|-------|-------|--------|
| Per API Key | 1,000 req | 1 minute |
| Per Organization | 10,000 req | 1 minute |
| Per IP (unauthenticated) | 100 req | 1 minute |

### Rate Limit Headers

Every response includes:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1705312260
X-RateLimit-Policy: 1000;w=60
```

### Rate Limit Exceeded Response

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 30
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705312260

{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded. Retry after 30 seconds.",
    "details": {
      "limit": 1000,
      "window": "1m",
      "reset_at": "2025-01-15T10:31:00Z"
    },
    "request_id": "req_abc123"
  }
}
```

### Custom Rate Limits

Enterprise plans can configure custom limits per API key or team.

---

## Error Handling

### Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `bad_request` | Malformed request syntax |
| 400 | `validation_error` | Invalid field values |
| 401 | `unauthorized` | Missing or invalid authentication |
| 401 | `token_expired` | API key or session expired |
| 403 | `forbidden` | Insufficient permissions |
| 403 | `scope_exceeded` | Key doesn't have required scope |
| 404 | `not_found` | Resource doesn't exist |
| 409 | `conflict` | Resource already exists |
| 422 | `unprocessable_entity` | Valid syntax but can't process |
| 429 | `rate_limit_exceeded` | Too many requests |
| 500 | `internal_error` | Server error |
| 502 | `upstream_error` | MCP server returned error |
| 503 | `service_unavailable` | Service temporarily unavailable |
| 504 | `gateway_timeout` | MCP server didn't respond |

### Error Response Format

```json
{
  "error": {
    "code": "validation_error",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "email",
        "code": "invalid_format",
        "message": "Must be a valid email address"
      },
      {
        "field": "name",
        "code": "required",
        "message": "Name is required"
      }
    ],
    "request_id": "req_abc123xyz",
    "documentation_url": "https://docs.gatewayops.com/errors/validation_error"
  }
}
```

### Handling Errors

```python
import requests

try:
    response = requests.post(url, headers=headers, json=data)
    response.raise_for_status()
except requests.exceptions.HTTPError as e:
    error = response.json().get('error', {})

    if error.get('code') == 'rate_limit_exceeded':
        retry_after = int(response.headers.get('Retry-After', 60))
        time.sleep(retry_after)
        # Retry request

    elif error.get('code') == 'validation_error':
        for detail in error.get('details', []):
            print(f"Field {detail['field']}: {detail['message']}")

    elif error.get('code') == 'upstream_error':
        # MCP server issue - check server status
        pass
```

---

## Gateway API

The Gateway API proxies requests to your MCP servers with authentication, rate limiting, cost tracking, and tracing.

**Base URL:** `https://gateway.gatewayops.com/v1`

### Proxy MCP Tool Call

Invoke a tool on an MCP server.

```http
POST /mcp/{server}/tools/call
```

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `server` | string | MCP server name (configured in dashboard) |

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | Bearer token (API key) |
| `Content-Type` | Yes | `application/json` |
| `X-Request-ID` | No | Idempotency key |
| `X-Trace-ID` | No | Custom trace ID |

**Request Body:**

```json
{
  "tool": "read_file",
  "arguments": {
    "path": "/src/main.py"
  }
}
```

**Response:**

```json
{
  "data": {
    "result": {
      "content": "# Main application file\nimport os\n...",
      "type": "text"
    }
  },
  "meta": {
    "trace_id": "tr_abc123def456",
    "span_id": "sp_789xyz",
    "latency_ms": 150,
    "cost": 0.001,
    "mcp_server": "filesystem",
    "tool": "read_file",
    "request_id": "req_abc123"
  }
}
```

**Example:**

```bash
curl -X POST https://gateway.gatewayops.com/v1/mcp/filesystem/tools/call \
  -H "Authorization: Bearer gwo_prod_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "read_file",
    "arguments": {
      "path": "/src/main.py"
    }
  }'
```

---

### List MCP Tools

List available tools on an MCP server.

```http
POST /mcp/{server}/tools/list
```

**Request Body:**

```json
{}
```

**Response:**

```json
{
  "data": {
    "tools": [
      {
        "name": "read_file",
        "description": "Read contents of a file",
        "inputSchema": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "Path to the file"
            }
          },
          "required": ["path"]
        }
      },
      {
        "name": "write_file",
        "description": "Write contents to a file",
        "inputSchema": {
          "type": "object",
          "properties": {
            "path": { "type": "string" },
            "content": { "type": "string" }
          },
          "required": ["path", "content"]
        }
      }
    ]
  },
  "meta": {
    "trace_id": "tr_abc123",
    "latency_ms": 50,
    "cost": 0.0001
  }
}
```

---

### Read MCP Resource

Read a resource from an MCP server.

```http
POST /mcp/{server}/resources/read
```

**Request Body:**

```json
{
  "uri": "file:///src/main.py"
}
```

**Response:**

```json
{
  "data": {
    "contents": [
      {
        "uri": "file:///src/main.py",
        "mimeType": "text/x-python",
        "text": "# Main application file..."
      }
    ]
  },
  "meta": {
    "trace_id": "tr_abc123",
    "latency_ms": 120,
    "cost": 0.001
  }
}
```

---

### List MCP Resources

List available resources on an MCP server.

```http
POST /mcp/{server}/resources/list
```

**Response:**

```json
{
  "data": {
    "resources": [
      {
        "uri": "file:///src/",
        "name": "Source Directory",
        "description": "Application source code",
        "mimeType": "inode/directory"
      }
    ]
  }
}
```

---

### Get MCP Prompt

Retrieve a prompt template from an MCP server.

```http
POST /mcp/{server}/prompts/get
```

**Request Body:**

```json
{
  "name": "code_review",
  "arguments": {
    "language": "python",
    "focus": "security"
  }
}
```

**Response:**

```json
{
  "data": {
    "description": "Code review prompt for Python",
    "messages": [
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "Review the following Python code for security issues..."
        }
      }
    ]
  }
}
```

---

### List MCP Prompts

List available prompts on an MCP server.

```http
POST /mcp/{server}/prompts/list
```

**Response:**

```json
{
  "data": {
    "prompts": [
      {
        "name": "code_review",
        "description": "Generate code review prompts",
        "arguments": [
          {
            "name": "language",
            "description": "Programming language",
            "required": true
          }
        ]
      }
    ]
  }
}
```

---

## Traces API

View and search distributed traces.

**Base URL:** `https://api.gatewayops.com/v1`

### List Traces

```http
GET /traces
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start_time` | ISO8601 | Yes | Start of time range |
| `end_time` | ISO8601 | Yes | End of time range |
| `status` | string | No | Filter: `success`, `error`, `timeout` |
| `user_id` | string | No | Filter by user ID |
| `team_id` | string | No | Filter by team ID |
| `agent_name` | string | No | Filter by agent name |
| `mcp_server` | string | No | Filter by MCP server |
| `min_duration_ms` | integer | No | Minimum duration filter |
| `max_duration_ms` | integer | No | Maximum duration filter |
| `min_cost` | number | No | Minimum cost filter |
| `limit` | integer | No | Items per page (default: 50) |
| `cursor` | string | No | Pagination cursor |

**Response:**

```json
{
  "data": {
    "traces": [
      {
        "id": "tr_abc123def456",
        "start_time": "2025-01-15T10:30:00.000Z",
        "end_time": "2025-01-15T10:30:05.200Z",
        "duration_ms": 5200,
        "status": "success",
        "user_id": "user_123",
        "team_id": "team_eng",
        "agent_name": "code-review-agent",
        "environment": "production",
        "span_count": 5,
        "total_cost": 0.052,
        "mcp_servers": ["filesystem", "code-analysis"]
      }
    ]
  },
  "pagination": {
    "has_more": true,
    "cursor": "eyJ0IjoiMjAyNS0wMS0xNVQxMDozMDowMC4wMDBaIn0="
  }
}
```

**Example:**

```bash
curl -G https://api.gatewayops.com/v1/traces \
  -H "Authorization: Bearer gwo_prod_xxx" \
  --data-urlencode "start_time=2025-01-15T00:00:00Z" \
  --data-urlencode "end_time=2025-01-15T23:59:59Z" \
  --data-urlencode "status=error" \
  --data-urlencode "limit=20"
```

---

### Get Trace

Retrieve a single trace with all spans.

```http
GET /traces/{trace_id}
```

**Response:**

```json
{
  "data": {
    "trace": {
      "id": "tr_abc123def456",
      "start_time": "2025-01-15T10:30:00.000Z",
      "end_time": "2025-01-15T10:30:05.200Z",
      "duration_ms": 5200,
      "status": "success",
      "context": {
        "user_id": "user_123",
        "team_id": "team_eng",
        "agent_name": "code-review-agent",
        "environment": "production",
        "api_key_id": "key_456",
        "ip_address": "192.168.1.100"
      },
      "spans": [
        {
          "id": "sp_001",
          "parent_id": null,
          "operation": "mcp.tools.call",
          "mcp_server": "filesystem",
          "tool": "read_file",
          "start_time": "2025-01-15T10:30:00.100Z",
          "end_time": "2025-01-15T10:30:00.250Z",
          "duration_ms": 150,
          "status": "success",
          "request": {
            "path": "/src/main.py"
          },
          "response": {
            "content_length": 4500,
            "content_type": "text"
          },
          "cost": 0.001,
          "attributes": {
            "file.size": "4500"
          }
        },
        {
          "id": "sp_002",
          "parent_id": "sp_001",
          "operation": "mcp.tools.call",
          "mcp_server": "code-analysis",
          "tool": "analyze",
          "start_time": "2025-01-15T10:30:00.300Z",
          "end_time": "2025-01-15T10:30:02.800Z",
          "duration_ms": 2500,
          "status": "success",
          "request": {
            "language": "python",
            "code_length": 4500
          },
          "response": {
            "issues_found": 3
          },
          "cost": 0.045
        }
      ],
      "total_cost": 0.052,
      "tags": {
        "pr_number": "123",
        "repo": "acme/webapp"
      }
    }
  }
}
```

---

### Search Traces

Advanced trace search with full-text query.

```http
POST /traces/search
```

**Request Body:**

```json
{
  "query": "error connection refused",
  "filters": {
    "time_range": {
      "start": "2025-01-15T00:00:00Z",
      "end": "2025-01-15T23:59:59Z"
    },
    "status": ["error"],
    "mcp_servers": ["database"],
    "min_duration_ms": 1000
  },
  "sort": {
    "field": "duration_ms",
    "order": "desc"
  },
  "limit": 50
}
```

**Response:**

```json
{
  "data": {
    "traces": [...],
    "total_count": 127,
    "aggregations": {
      "by_status": {
        "error": 127
      },
      "by_server": {
        "database": 127
      }
    }
  },
  "pagination": {
    "has_more": true,
    "cursor": "..."
  }
}
```

---

### Export Traces

Export traces to external format.

```http
POST /traces/export
```

**Request Body:**

```json
{
  "format": "otlp",
  "filters": {
    "time_range": {
      "start": "2025-01-15T00:00:00Z",
      "end": "2025-01-15T23:59:59Z"
    },
    "trace_ids": ["tr_abc123", "tr_def456"]
  },
  "destination": {
    "type": "s3",
    "bucket": "my-traces-bucket",
    "prefix": "exports/2025-01-15/"
  }
}
```

**Response:**

```json
{
  "data": {
    "export_id": "exp_xyz789",
    "status": "pending",
    "trace_count": 1000,
    "estimated_size_bytes": 5242880
  }
}
```

---

## Costs API

View and analyze cost data.

### Get Cost Summary

```http
GET /costs/summary
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start_date` | date | Yes | Start date (YYYY-MM-DD) |
| `end_date` | date | Yes | End date (YYYY-MM-DD) |
| `group_by` | string | No | `day`, `week`, `month`, `team`, `user`, `server` |
| `team_id` | string | No | Filter by team |
| `user_id` | string | No | Filter by user |
| `mcp_server` | string | No | Filter by MCP server |

**Response:**

```json
{
  "data": {
    "summary": {
      "total_cost": 4523.45,
      "total_calls": 4523450,
      "average_cost_per_call": 0.001,
      "period": {
        "start": "2025-01-01",
        "end": "2025-01-31"
      }
    },
    "breakdown": [
      {
        "group": "2025-01-01",
        "cost": 145.23,
        "calls": 145230,
        "average_cost": 0.001
      },
      {
        "group": "2025-01-02",
        "cost": 152.10,
        "calls": 152100,
        "average_cost": 0.001
      }
    ],
    "comparison": {
      "previous_period": {
        "total_cost": 3800.00,
        "total_calls": 3800000
      },
      "change_percent": 19.04
    }
  }
}
```

---

### Get Costs by Team

```http
GET /costs/by-team
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `start_date` | date | Yes | Start date |
| `end_date` | date | Yes | End date |
| `include_users` | boolean | No | Include user breakdown |

**Response:**

```json
{
  "data": {
    "teams": [
      {
        "team_id": "team_eng",
        "team_name": "Engineering",
        "total_cost": 2150.00,
        "total_calls": 2150000,
        "percentage_of_total": 47.5,
        "budget": {
          "amount": 2500.00,
          "usage_percent": 86.0
        },
        "top_servers": [
          {"server": "code-analysis", "cost": 1200.00},
          {"server": "database", "cost": 650.00}
        ],
        "users": [
          {"user_id": "user_123", "cost": 800.00},
          {"user_id": "user_456", "cost": 650.00}
        ]
      }
    ]
  }
}
```

---

### Get Costs by Server

```http
GET /costs/by-server
```

**Response:**

```json
{
  "data": {
    "servers": [
      {
        "server": "code-analysis",
        "total_cost": 2035.00,
        "total_calls": 40700,
        "average_cost_per_call": 0.05,
        "percentage_of_total": 45.0,
        "top_tools": [
          {"tool": "analyze", "cost": 1800.00, "calls": 36000},
          {"tool": "lint", "cost": 235.00, "calls": 4700}
        ]
      },
      {
        "server": "database",
        "total_cost": 1131.00,
        "total_calls": 113100,
        "average_cost_per_call": 0.01,
        "percentage_of_total": 25.0
      }
    ]
  }
}
```

---

### Get Trace Cost

Get cost breakdown for a specific trace.

```http
GET /costs/trace/{trace_id}
```

**Response:**

```json
{
  "data": {
    "trace_id": "tr_abc123",
    "total_cost": 0.052,
    "spans": [
      {
        "span_id": "sp_001",
        "server": "filesystem",
        "tool": "read_file",
        "cost": 0.001,
        "breakdown": {
          "base_cost": 0.001,
          "token_cost": 0.0
        }
      },
      {
        "span_id": "sp_002",
        "server": "code-analysis",
        "tool": "analyze",
        "cost": 0.045,
        "breakdown": {
          "base_cost": 0.005,
          "input_tokens": 4500,
          "input_cost": 0.018,
          "output_tokens": 1200,
          "output_cost": 0.022
        }
      }
    ]
  }
}
```

---

## API Keys API

Manage API keys for authentication.

### List API Keys

```http
GET /api-keys
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | Filter: `organization`, `team`, `user`, `scoped` |
| `status` | string | Filter: `active`, `expired`, `revoked` |
| `team_id` | string | Filter by team |

**Response:**

```json
{
  "data": {
    "api_keys": [
      {
        "id": "key_abc123",
        "name": "Production Agent",
        "type": "organization",
        "prefix": "gwo_prod_a1B2",
        "status": "active",
        "scopes": {
          "mcp_servers": ["*"],
          "permissions": ["*"]
        },
        "created_at": "2025-01-01T00:00:00Z",
        "expires_at": "2026-01-01T00:00:00Z",
        "last_used_at": "2025-01-15T10:30:00Z",
        "created_by": {
          "id": "user_123",
          "email": "admin@company.com"
        },
        "usage": {
          "total_requests": 150000,
          "last_24h_requests": 5000
        }
      }
    ]
  }
}
```

---

### Create API Key

```http
POST /api-keys
```

**Request Body:**

```json
{
  "name": "Production Agent Key",
  "type": "organization",
  "expires_at": "2026-01-01T00:00:00Z",
  "scopes": {
    "mcp_servers": ["filesystem", "database", "code-analysis"],
    "permissions": [
      "gateway:proxy",
      "traces:read",
      "costs:read"
    ]
  },
  "metadata": {
    "project": "code-review",
    "environment": "production"
  },
  "rate_limit": {
    "requests_per_minute": 500
  }
}
```

**Response:**

```json
{
  "data": {
    "api_key": {
      "id": "key_xyz789",
      "name": "Production Agent Key",
      "type": "organization",
      "prefix": "gwo_prod_x1Y2",
      "secret": "gwo_prod_x1Y2z3A4b5C6d7E8f9G0h1I2j3K4l5M6",
      "status": "active",
      "scopes": {...},
      "created_at": "2025-01-15T10:30:00Z",
      "expires_at": "2026-01-01T00:00:00Z"
    }
  },
  "meta": {
    "warning": "Store this secret securely. It will not be shown again."
  }
}
```

---

### Get API Key

```http
GET /api-keys/{key_id}
```

**Response:**

```json
{
  "data": {
    "api_key": {
      "id": "key_abc123",
      "name": "Production Agent",
      "type": "organization",
      "prefix": "gwo_prod_a1B2",
      "status": "active",
      "scopes": {...},
      "created_at": "2025-01-01T00:00:00Z",
      "expires_at": "2026-01-01T00:00:00Z",
      "last_used_at": "2025-01-15T10:30:00Z",
      "usage_stats": {
        "total_requests": 150000,
        "requests_by_day": [
          {"date": "2025-01-14", "count": 4800},
          {"date": "2025-01-15", "count": 5200}
        ],
        "top_servers": [
          {"server": "filesystem", "count": 80000},
          {"server": "database", "count": 50000}
        ]
      }
    }
  }
}
```

---

### Update API Key

```http
PATCH /api-keys/{key_id}
```

**Request Body:**

```json
{
  "name": "Updated Key Name",
  "scopes": {
    "mcp_servers": ["filesystem"],
    "permissions": ["gateway:proxy"]
  },
  "rate_limit": {
    "requests_per_minute": 1000
  }
}
```

---

### Rotate API Key

Generate a new secret for an existing key.

```http
POST /api-keys/{key_id}/rotate
```

**Request Body:**

```json
{
  "grace_period_hours": 24
}
```

**Response:**

```json
{
  "data": {
    "api_key": {
      "id": "key_abc123",
      "prefix": "gwo_prod_n3W4",
      "secret": "gwo_prod_n3W4k5E6y7R8o9T0a1T2e3D4s5E6c7R8",
      "previous_key_valid_until": "2025-01-16T10:30:00Z"
    }
  },
  "meta": {
    "warning": "Store this secret securely. It will not be shown again."
  }
}
```

---

### Revoke API Key

```http
DELETE /api-keys/{key_id}
```

**Response:**

```json
{
  "data": {
    "api_key": {
      "id": "key_abc123",
      "status": "revoked",
      "revoked_at": "2025-01-15T10:30:00Z"
    }
  }
}
```

---

## Budgets API

Manage cost budgets and alerts.

### List Budgets

```http
GET /budgets
```

**Response:**

```json
{
  "data": {
    "budgets": [
      {
        "id": "bgt_abc123",
        "name": "Engineering Monthly",
        "scope": {
          "type": "team",
          "team_id": "team_eng"
        },
        "amount": 2500.00,
        "currency": "USD",
        "period": "monthly",
        "alert_thresholds": [50, 80, 100],
        "enforce_limit": false,
        "current_usage": {
          "amount": 2150.00,
          "percentage": 86.0,
          "period_start": "2025-01-01",
          "period_end": "2025-01-31"
        },
        "status": "warning",
        "created_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### Create Budget

```http
POST /budgets
```

**Request Body:**

```json
{
  "name": "Engineering Monthly Budget",
  "scope": {
    "type": "team",
    "team_id": "team_eng"
  },
  "amount": 2500.00,
  "currency": "USD",
  "period": "monthly",
  "alert_thresholds": [50, 80, 100],
  "enforce_limit": true,
  "notifications": {
    "email": ["admin@company.com", "finance@company.com"],
    "slack_channel": "#cost-alerts",
    "webhook_url": "https://hooks.company.com/budget-alerts"
  }
}
```

---

### Get Budget Usage

```http
GET /budgets/{budget_id}/usage
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `periods` | integer | Number of periods to return (default: 6) |

**Response:**

```json
{
  "data": {
    "budget_id": "bgt_abc123",
    "current_period": {
      "start": "2025-01-01",
      "end": "2025-01-31",
      "amount": 2500.00,
      "used": 2150.00,
      "remaining": 350.00,
      "percentage": 86.0,
      "projected_end": 2580.00,
      "status": "warning"
    },
    "history": [
      {
        "period": "2024-12",
        "amount": 2500.00,
        "used": 2320.00,
        "percentage": 92.8
      },
      {
        "period": "2024-11",
        "amount": 2500.00,
        "used": 2100.00,
        "percentage": 84.0
      }
    ],
    "daily_breakdown": [
      {"date": "2025-01-01", "cost": 70.00},
      {"date": "2025-01-02", "cost": 75.00}
    ]
  }
}
```

---

## Organizations API

Manage organization settings.

### Get Organization

```http
GET /organizations/current
```

**Response:**

```json
{
  "data": {
    "organization": {
      "id": "org_abc123",
      "name": "Acme Corp",
      "slug": "acme-corp",
      "plan": "pro",
      "settings": {
        "default_data_retention_days": 30,
        "require_sso": true,
        "allowed_ip_ranges": ["192.168.1.0/24"]
      },
      "usage": {
        "users": 45,
        "teams": 8,
        "api_keys": 23,
        "mcp_servers": 5
      },
      "limits": {
        "users": 100,
        "teams": 20,
        "api_keys": 100,
        "mcp_servers": 20,
        "requests_per_month": 10000000
      },
      "created_at": "2024-06-15T00:00:00Z"
    }
  }
}
```

---

### Update Organization

```http
PATCH /organizations/current
```

**Request Body:**

```json
{
  "name": "Acme Corporation",
  "settings": {
    "default_data_retention_days": 60,
    "require_sso": true,
    "session_timeout_hours": 8
  }
}
```

---

## Teams API

Manage teams within your organization.

### List Teams

```http
GET /teams
```

**Response:**

```json
{
  "data": {
    "teams": [
      {
        "id": "team_eng",
        "name": "Engineering",
        "slug": "engineering",
        "member_count": 25,
        "api_key_count": 8,
        "budget_id": "bgt_abc123",
        "created_at": "2024-06-15T00:00:00Z"
      }
    ]
  }
}
```

---

### Create Team

```http
POST /teams
```

**Request Body:**

```json
{
  "name": "Data Science",
  "slug": "data-science",
  "settings": {
    "default_mcp_servers": ["database", "ml-models"]
  }
}
```

---

### Add Team Member

```http
POST /teams/{team_id}/members
```

**Request Body:**

```json
{
  "user_id": "user_123",
  "role": "admin"
}
```

---

## MCP Servers API

Configure MCP server connections.

### List MCP Servers

```http
GET /mcp-servers
```

**Response:**

```json
{
  "data": {
    "servers": [
      {
        "id": "srv_abc123",
        "name": "filesystem",
        "url": "http://mcp-filesystem.internal:3000",
        "status": "healthy",
        "auth_type": "none",
        "timeout_ms": 30000,
        "enabled": true,
        "tools": ["read_file", "write_file", "list_directory"],
        "usage": {
          "last_24h_calls": 15000,
          "last_24h_cost": 15.00,
          "avg_latency_ms": 120
        },
        "health_check": {
          "last_check": "2025-01-15T10:29:00Z",
          "status": "healthy",
          "latency_ms": 50
        },
        "created_at": "2024-06-15T00:00:00Z"
      }
    ]
  }
}
```

---

### Create MCP Server

```http
POST /mcp-servers
```

**Request Body:**

```json
{
  "name": "database",
  "url": "http://mcp-database.internal:3000",
  "auth_type": "bearer",
  "auth_config": {
    "token": "secret_token_here"
  },
  "timeout_ms": 60000,
  "retry_config": {
    "max_retries": 3,
    "backoff_ms": 1000
  },
  "rate_limit": {
    "requests_per_second": 100
  },
  "pricing": {
    "per_call": 0.01,
    "per_input_token": 0.0,
    "per_output_token": 0.0
  }
}
```

---

### Test MCP Server Connection

```http
POST /mcp-servers/{server_id}/test
```

**Response:**

```json
{
  "data": {
    "status": "success",
    "latency_ms": 120,
    "tools_discovered": 5,
    "tools": [
      {"name": "query", "description": "Execute SQL query"},
      {"name": "schema", "description": "Get table schema"}
    ]
  }
}
```

---

## Alerts API

Manage alerts and notifications.

### List Alerts

```http
GET /alerts
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | `active`, `acknowledged`, `resolved` |
| `severity` | string | `critical`, `high`, `medium`, `low` |
| `type` | string | `budget`, `error_rate`, `latency`, `custom` |

**Response:**

```json
{
  "data": {
    "alerts": [
      {
        "id": "alt_abc123",
        "type": "budget",
        "severity": "high",
        "status": "active",
        "title": "Engineering team at 86% of budget",
        "message": "Team Engineering has used $2,150 of $2,500 monthly budget",
        "resource": {
          "type": "budget",
          "id": "bgt_abc123",
          "name": "Engineering Monthly"
        },
        "triggered_at": "2025-01-15T08:00:00Z",
        "acknowledged_at": null,
        "resolved_at": null
      }
    ]
  }
}
```

---

### Acknowledge Alert

```http
POST /alerts/{alert_id}/acknowledge
```

**Request Body:**

```json
{
  "note": "Aware of this, investigating"
}
```

---

### Create Alert Rule

```http
POST /alert-rules
```

**Request Body:**

```json
{
  "name": "High Error Rate",
  "type": "error_rate",
  "condition": {
    "metric": "error_rate",
    "operator": "greater_than",
    "threshold": 5.0,
    "window_minutes": 5
  },
  "scope": {
    "mcp_servers": ["database"]
  },
  "severity": "critical",
  "notifications": {
    "channels": ["slack", "pagerduty"],
    "slack_channel": "#alerts",
    "pagerduty_service": "gatewayops-critical"
  },
  "enabled": true
}
```

---

## Webhooks API

Configure webhooks for real-time events.

### List Webhooks

```http
GET /webhooks
```

**Response:**

```json
{
  "data": {
    "webhooks": [
      {
        "id": "whk_abc123",
        "name": "Cost Alerts",
        "url": "https://hooks.company.com/gatewayops",
        "events": ["alert.triggered", "budget.threshold_reached"],
        "status": "active",
        "secret": "whsec_****",
        "created_at": "2025-01-01T00:00:00Z",
        "last_triggered_at": "2025-01-15T08:00:00Z",
        "stats": {
          "total_deliveries": 150,
          "successful_deliveries": 148,
          "failed_deliveries": 2
        }
      }
    ]
  }
}
```

---

### Create Webhook

```http
POST /webhooks
```

**Request Body:**

```json
{
  "name": "All Events",
  "url": "https://hooks.company.com/gatewayops",
  "events": [
    "alert.triggered",
    "alert.resolved",
    "budget.threshold_reached",
    "budget.exceeded",
    "api_key.created",
    "api_key.revoked"
  ],
  "headers": {
    "X-Custom-Header": "value"
  }
}
```

**Response:**

```json
{
  "data": {
    "webhook": {
      "id": "whk_xyz789",
      "name": "All Events",
      "url": "https://hooks.company.com/gatewayops",
      "events": [...],
      "secret": "whsec_a1b2c3d4e5f6g7h8i9j0",
      "status": "active",
      "created_at": "2025-01-15T10:30:00Z"
    }
  },
  "meta": {
    "note": "Use the secret to verify webhook signatures"
  }
}
```

---

### Webhook Payload Format

```json
{
  "id": "evt_abc123",
  "type": "alert.triggered",
  "created_at": "2025-01-15T10:30:00Z",
  "data": {
    "alert": {
      "id": "alt_xyz789",
      "type": "budget",
      "severity": "high",
      "title": "Budget threshold reached",
      "message": "Engineering team at 80% of budget"
    }
  }
}
```

### Verifying Webhook Signatures

```python
import hmac
import hashlib

def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(f"sha256={expected}", signature)

# Usage
signature = request.headers.get('X-GatewayOps-Signature')
is_valid = verify_webhook(request.body, signature, webhook_secret)
```

---

## SDKs

### Python SDK

**Installation:**

```bash
pip install gatewayops
```

**Usage:**

```python
from gatewayops import GatewayOps

# Initialize client
client = GatewayOps(
    api_key="gwo_prod_xxx",
    gateway_url="https://gateway.gatewayops.com"  # Optional
)

# Proxy MCP call
result = client.mcp.call(
    server="filesystem",
    tool="read_file",
    arguments={"path": "/src/main.py"}
)
print(result.data)
print(f"Cost: ${result.meta.cost}")
print(f"Trace ID: {result.meta.trace_id}")

# List traces
traces = client.traces.list(
    start_time="2025-01-15T00:00:00Z",
    end_time="2025-01-15T23:59:59Z",
    status="error"
)
for trace in traces:
    print(f"{trace.id}: {trace.status} ({trace.duration_ms}ms)")

# Get costs
costs = client.costs.summary(
    start_date="2025-01-01",
    end_date="2025-01-31",
    group_by="team"
)
print(f"Total cost: ${costs.summary.total_cost}")

# Manage API keys
key = client.api_keys.create(
    name="New Agent Key",
    scopes={"mcp_servers": ["filesystem"]}
)
print(f"New key: {key.secret}")  # Only shown once!

# Handle errors
from gatewayops.exceptions import (
    RateLimitError,
    AuthenticationError,
    MCPError
)

try:
    result = client.mcp.call(server="database", tool="query", arguments={})
except RateLimitError as e:
    print(f"Rate limited. Retry after {e.retry_after} seconds")
except MCPError as e:
    print(f"MCP server error: {e.message}")
```

**Async Support:**

```python
from gatewayops import AsyncGatewayOps
import asyncio

async def main():
    client = AsyncGatewayOps(api_key="gwo_prod_xxx")

    # Concurrent MCP calls
    results = await asyncio.gather(
        client.mcp.call("filesystem", "read_file", {"path": "/a.py"}),
        client.mcp.call("filesystem", "read_file", {"path": "/b.py"}),
        client.mcp.call("filesystem", "read_file", {"path": "/c.py"}),
    )

    await client.close()

asyncio.run(main())
```

---

### TypeScript/JavaScript SDK

**Installation:**

```bash
npm install @gatewayops/sdk
# or
yarn add @gatewayops/sdk
```

**Usage:**

```typescript
import { GatewayOps } from '@gatewayops/sdk';

// Initialize client
const client = new GatewayOps({
  apiKey: 'gwo_prod_xxx',
  gatewayUrl: 'https://gateway.gatewayops.com'  // Optional
});

// Proxy MCP call
const result = await client.mcp.call({
  server: 'filesystem',
  tool: 'read_file',
  arguments: { path: '/src/main.py' }
});

console.log(result.data);
console.log(`Cost: $${result.meta.cost}`);
console.log(`Trace ID: ${result.meta.traceId}`);

// List traces
const traces = await client.traces.list({
  startTime: '2025-01-15T00:00:00Z',
  endTime: '2025-01-15T23:59:59Z',
  status: 'error'
});

for (const trace of traces.data.traces) {
  console.log(`${trace.id}: ${trace.status} (${trace.durationMs}ms)`);
}

// Get costs
const costs = await client.costs.summary({
  startDate: '2025-01-01',
  endDate: '2025-01-31',
  groupBy: 'team'
});

console.log(`Total cost: $${costs.data.summary.totalCost}`);

// Manage API keys
const key = await client.apiKeys.create({
  name: 'New Agent Key',
  scopes: { mcpServers: ['filesystem'] }
});

console.log(`New key: ${key.data.apiKey.secret}`);  // Only shown once!

// Error handling
import {
  RateLimitError,
  AuthenticationError,
  MCPError
} from '@gatewayops/sdk';

try {
  await client.mcp.call({ server: 'database', tool: 'query', arguments: {} });
} catch (error) {
  if (error instanceof RateLimitError) {
    console.log(`Rate limited. Retry after ${error.retryAfter} seconds`);
  } else if (error instanceof MCPError) {
    console.log(`MCP server error: ${error.message}`);
  }
}
```

**With React Query:**

```typescript
import { useQuery, useMutation } from '@tanstack/react-query';
import { useGatewayOps } from '@gatewayops/react';

function TracesPage() {
  const client = useGatewayOps();

  const { data: traces, isLoading } = useQuery({
    queryKey: ['traces', { status: 'error' }],
    queryFn: () => client.traces.list({
      startTime: new Date(Date.now() - 86400000).toISOString(),
      endTime: new Date().toISOString(),
      status: 'error'
    })
  });

  if (isLoading) return <div>Loading...</div>;

  return (
    <ul>
      {traces?.data.traces.map(trace => (
        <li key={trace.id}>{trace.id}: {trace.status}</li>
      ))}
    </ul>
  );
}
```

---

### Go SDK

**Installation:**

```bash
go get github.com/gatewayops/gatewayops-go
```

**Usage:**

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gatewayops/gatewayops-go"
)

func main() {
    // Initialize client
    client := gatewayops.NewClient("gwo_prod_xxx")

    // Proxy MCP call
    result, err := client.MCP.Call(context.Background(), &gatewayops.MCPCallRequest{
        Server: "filesystem",
        Tool:   "read_file",
        Arguments: map[string]interface{}{
            "path": "/src/main.py",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Result: %v\n", result.Data)
    fmt.Printf("Cost: $%.4f\n", result.Meta.Cost)
    fmt.Printf("Trace ID: %s\n", result.Meta.TraceID)

    // List traces
    traces, err := client.Traces.List(context.Background(), &gatewayops.TracesListParams{
        StartTime: "2025-01-15T00:00:00Z",
        EndTime:   "2025-01-15T23:59:59Z",
        Status:    gatewayops.StatusError,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, trace := range traces.Data.Traces {
        fmt.Printf("%s: %s (%dms)\n", trace.ID, trace.Status, trace.DurationMs)
    }

    // Error handling
    var rateLimitErr *gatewayops.RateLimitError
    if errors.As(err, &rateLimitErr) {
        fmt.Printf("Rate limited. Retry after %d seconds\n", rateLimitErr.RetryAfter)
    }
}
```

---

### cURL Examples

**Proxy MCP Call:**

```bash
curl -X POST https://gateway.gatewayops.com/v1/mcp/filesystem/tools/call \
  -H "Authorization: Bearer gwo_prod_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "read_file",
    "arguments": {
      "path": "/src/main.py"
    }
  }'
```

**List Traces:**

```bash
curl -G https://api.gatewayops.com/v1/traces \
  -H "Authorization: Bearer gwo_prod_xxx" \
  --data-urlencode "start_time=2025-01-15T00:00:00Z" \
  --data-urlencode "end_time=2025-01-15T23:59:59Z" \
  --data-urlencode "limit=10"
```

**Get Cost Summary:**

```bash
curl -G https://api.gatewayops.com/v1/costs/summary \
  -H "Authorization: Bearer gwo_prod_xxx" \
  --data-urlencode "start_date=2025-01-01" \
  --data-urlencode "end_date=2025-01-31" \
  --data-urlencode "group_by=team"
```

**Create API Key:**

```bash
curl -X POST https://api.gatewayops.com/v1/api-keys \
  -H "Authorization: Bearer gwo_prod_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Agent",
    "type": "organization",
    "expires_at": "2026-01-01T00:00:00Z",
    "scopes": {
      "mcp_servers": ["filesystem", "database"]
    }
  }'
```

---

## Changelog

### v1.0.0 (2025-01-15)

**Initial Release**

- Gateway API for MCP proxying
- Traces API for observability
- Costs API for cost tracking
- API Keys management
- Budgets and alerts
- Organization and team management
- Webhooks for real-time events
- Python, TypeScript, and Go SDKs

### Upcoming

**v1.1.0 (Planned)**

- OpenTelemetry export endpoint
- Trace replay for debugging
- Advanced search filters
- Bulk operations

---

## Support

- **Documentation:** https://docs.gatewayops.com
- **API Status:** https://status.gatewayops.com
- **Support Email:** support@gatewayops.com
- **Community Discord:** https://discord.gg/gatewayops
