# GatewayOps

**Enterprise MCP Gateway - Authentication, Rate Limiting, Tracing & Cost Tracking**

GatewayOps is a proxy layer that sits between AI agents and MCP (Model Context Protocol) servers, providing enterprise-grade observability, security, and cost management.

## Features

- **Authentication**: API key validation with bcrypt, per-key permissions
- **Rate Limiting**: Redis sliding window algorithm, per-key limits
- **Distributed Tracing**: Trace ID propagation, span tracking
- **Cost Tracking**: Per-call pricing, usage attribution
- **Request Logging**: Structured JSON logs, data masking
- **MCP Proxy**: Tools, resources, and prompts forwarding

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (optional, for local development)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/akz4ol/gatewayops.git
cd gatewayops

# Run the setup script
./scripts/dev-setup.sh

# Or manually:
cp .env.example .env
docker-compose -f deployments/docker-compose.yml up -d

# Run the gateway
make run
```

### Generate an API Key

```bash
go run scripts/generate-key.go dev
```

### Test the Gateway

```bash
# Health check
curl http://localhost:8080/health

# List MCP tools (requires API key)
curl -X POST http://localhost:8080/v1/mcp/mock/tools/list \
  -H "Authorization: Bearer gwo_dev_YOUR_API_KEY" \
  -H "Content-Type: application/json"

# Call an MCP tool
curl -X POST http://localhost:8080/v1/mcp/mock/tools/call \
  -H "Authorization: Bearer gwo_dev_YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"tool": "read_file", "arguments": {"path": "/test.txt"}}'
```

## API Endpoints

### Health
- `GET /health` - Liveness check
- `GET /ready` - Readiness check

### MCP Proxy
- `POST /v1/mcp/{server}/tools/call` - Call an MCP tool
- `POST /v1/mcp/{server}/tools/list` - List available tools
- `POST /v1/mcp/{server}/resources/read` - Read a resource
- `POST /v1/mcp/{server}/resources/list` - List resources
- `POST /v1/mcp/{server}/prompts/get` - Get a prompt
- `POST /v1/mcp/{server}/prompts/list` - List prompts

## Project Structure

```
gatewayops/
├── gateway/
│   ├── cmd/gateway/main.go       # Entry point
│   └── internal/
│       ├── config/               # Configuration loading
│       ├── server/               # HTTP server
│       ├── router/               # Route definitions
│       ├── middleware/           # Auth, rate limit, logging, trace
│       ├── handler/              # Request handlers
│       ├── auth/                 # API key authentication
│       ├── ratelimit/            # Redis rate limiting
│       └── database/             # Database connections
├── migrations/
│   ├── postgres/                 # PostgreSQL schemas
│   └── clickhouse/               # ClickHouse schemas
├── deployments/
│   ├── docker-compose.yml        # Local development
│   └── docker/                   # Dockerfiles
├── scripts/                      # Development scripts
└── test/
    └── mock-mcp/                 # Mock MCP server
```

## Technology Stack

- **Language**: Go 1.21+
- **Router**: chi
- **Databases**: PostgreSQL, ClickHouse, Redis
- **Containerization**: Docker

## Configuration

Environment variables (see `.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `ENV` | `development` | Environment (development/production) |
| `DATABASE_URL` | - | PostgreSQL connection string |
| `REDIS_URL` | - | Redis connection string |
| `CLICKHOUSE_DSN` | - | ClickHouse connection string |
| `RATE_LIMIT_DEFAULT_RPM` | `1000` | Default requests per minute |

## Related Repositories

| Repository | Purpose |
|------------|---------|
| [gatewayops-product](https://github.com/akz4ol/gatewayops-product) | Product operations - PRD, architecture, API specs |
| [ai-saas-company](https://github.com/akz4ol/ai-saas-company) | Company framework |

## License

Proprietary - All rights reserved

---

Part of the [AI SaaS Company](https://github.com/akz4ol/ai-saas-company) organizational framework.
