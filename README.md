# GatewayOps

**Enterprise Control Plane for AI Agent Infrastructure**

GatewayOps is the missing observability, governance, and cost management layer that enterprises need to deploy AI agents at scale with confidence.

## Overview

GatewayOps MCP Gateway is a proxy layer that sits between AI agents and MCP (Model Context Protocol) servers, providing:

- **Authentication & Authorization**: SSO integration, API key management, RBAC
- **Cost Attribution**: Per-call tracking, team/project allocation, budget enforcement
- **Observability**: Distributed tracing, request logging, latency metrics
- **Security**: Request validation, rate limiting, audit trails, data masking

## Documentation

| Document | Description |
|----------|-------------|
| [PRD.md](./PRD.md) | Product Requirements Document - features, user stories, specifications |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Technical Architecture - system design, infrastructure, security |
| [API.md](./API.md) | API Documentation - endpoints, SDKs, examples |

## Target Market

- **Primary**: Engineering teams at companies with 50-500 employees running AI/ML workloads
- **Secondary**: Platform teams at enterprises (500+) standardizing AI agent infrastructure

## Key Features

### P0 (MVP)
- MCP Gateway Proxy
- Cost Tracking Engine
- Distributed Tracing
- Authentication & API Keys
- Request Logging & Audit Trail
- Rate Limiting

### P1 (Post-MVP)
- Budget Enforcement
- Real-Time Dashboard
- Anomaly Detection
- OpenTelemetry Export
- Slack Integration
- RBAC

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

- **Gateway**: Go 1.21+
- **Dashboard**: Next.js 14, shadcn/ui
- **Databases**: PostgreSQL, ClickHouse, Redis
- **Infrastructure**: AWS (EKS, RDS, ElastiCache)
- **CI/CD**: GitHub Actions, ArgoCD

## Timeline

| Phase | Timeline | Goal |
|-------|----------|------|
| Foundation | Months 1-3 | MVP with 5 design partners |
| Validation | Months 4-6 | GA launch, $100K ARR |
| Scale | Months 7-12 | $500K ARR, SOC2 certified |

## Links

- **Company**: [ai-saas-company](https://github.com/akz4ol/ai-saas-company) - Organization skill system
- **Website**: gatewayops.com (coming soon)
- **Documentation**: docs.gatewayops.com (coming soon)

---

Built with the [AI SaaS Company](https://github.com/akz4ol/ai-saas-company) organizational framework.
