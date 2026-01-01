# GatewayOps Deployment Guide

## Quick Start (Local Development)

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development without Docker)
- Node.js 20+ (for dashboard development)

### Run with Docker Compose

```bash
# Clone the repository
git clone https://github.com/akz4ol/gatewayops.git
cd gatewayops

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f gateway

# Stop services
docker-compose down
```

Services will be available at:
- **Gateway API**: http://localhost:8080
- **Dashboard**: http://localhost:3000
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **ClickHouse**: localhost:8123

### Run Without Docker

```bash
# Start PostgreSQL, Redis, ClickHouse locally first

# Run migrations
psql -U gatewayops -d gatewayops -f migrations/postgres/001_initial.sql
psql -U gatewayops -d gatewayops -f migrations/postgres/002_sso_rbac.sql
psql -U gatewayops -d gatewayops -f migrations/postgres/003_safety_approvals.sql
psql -U gatewayops -d gatewayops -f migrations/postgres/004_alerting.sql

# Start the gateway
cd gateway
go run ./cmd/gateway

# In another terminal, start the dashboard
cd dashboard
npm install
npm run dev
```

---

## Production Deployment

### Option 1: Fly.io (Recommended for getting started)

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login

# Deploy gateway
cd gateway
flyctl launch --name gatewayops-api
flyctl deploy

# Deploy dashboard
cd ../dashboard
flyctl launch --name gatewayops-dashboard
flyctl deploy

# Set up PostgreSQL
flyctl postgres create --name gatewayops-db
flyctl postgres attach gatewayops-db --app gatewayops-api

# Set up Redis
flyctl redis create --name gatewayops-redis
```

### Option 2: Railway

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Initialize project
railway init

# Deploy
railway up
```

### Option 3: Render

1. Connect your GitHub repository to Render
2. Create a new Web Service for `gateway/`
3. Create a new Web Service for `dashboard/`
4. Create PostgreSQL and Redis databases
5. Set environment variables

### Option 4: Kubernetes

```bash
# Apply manifests
kubectl apply -f deploy/k8s/

# Or use Helm
helm install gatewayops ./deploy/helm/gatewayops
```

---

## Environment Variables

### Gateway

| Variable | Description | Default |
|----------|-------------|---------|
| `GATEWAY_ENV` | Environment (development/production) | development |
| `GATEWAY_PORT` | HTTP port | 8080 |
| `GATEWAY_LOG_LEVEL` | Log level (debug/info/warn/error) | info |
| `DATABASE_URL` | PostgreSQL connection string | required |
| `REDIS_URL` | Redis connection string | required |
| `CLICKHOUSE_URL` | ClickHouse connection string | optional |
| `JWT_SECRET` | Secret for JWT signing | required in production |
| `ENCRYPTION_KEY` | Key for encrypting secrets (32 bytes) | required in production |

### Dashboard

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Public API URL for browser | http://localhost:8080 |
| `GATEWAYOPS_API_URL` | Internal API URL for SSR | http://gateway:8080 |

---

## Database Migrations

Migrations are in `migrations/postgres/`. Run them in order:

```bash
# Using psql
psql $DATABASE_URL -f migrations/postgres/001_initial.sql
psql $DATABASE_URL -f migrations/postgres/002_sso_rbac.sql
psql $DATABASE_URL -f migrations/postgres/003_safety_approvals.sql
psql $DATABASE_URL -f migrations/postgres/004_alerting.sql
```

Or use the built-in migration command:

```bash
./gateway migrate up
```

---

## Health Checks

- **Gateway**: `GET /health` - Returns 200 if healthy
- **Gateway Ready**: `GET /ready` - Returns 200 if ready to accept traffic
- **Dashboard**: `GET /api/health` - Proxied to gateway

---

## Monitoring

### Prometheus Metrics

Gateway exposes Prometheus metrics at `/metrics`:
- `gatewayops_requests_total` - Total requests by server, operation, status
- `gatewayops_request_duration_seconds` - Request latency histogram
- `gatewayops_active_connections` - Current active connections

### OpenTelemetry

Configure OTEL export in settings or via environment:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://your-collector:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

---

## Scaling

### Horizontal Scaling

The gateway is stateless and can be scaled horizontally:

```bash
# Docker Compose
docker-compose up -d --scale gateway=3

# Kubernetes
kubectl scale deployment gatewayops-gateway --replicas=3
```

### Database Scaling

- **PostgreSQL**: Use connection pooling (PgBouncer) for many connections
- **Redis**: Use Redis Cluster for high availability
- **ClickHouse**: Use distributed tables for high volume

---

## Security Checklist

- [ ] Set strong `JWT_SECRET` (min 32 characters)
- [ ] Set strong `ENCRYPTION_KEY` (exactly 32 bytes)
- [ ] Enable TLS/HTTPS
- [ ] Configure rate limiting
- [ ] Set up SSO/OIDC for production
- [ ] Enable audit logging
- [ ] Configure safety policies
- [ ] Set up alerting

---

## Troubleshooting

### Gateway won't start

1. Check database connection: `psql $DATABASE_URL`
2. Check Redis connection: `redis-cli -u $REDIS_URL ping`
3. Check logs: `docker-compose logs gateway`

### Dashboard shows 500 errors

1. Check API URL configuration
2. Verify gateway is running: `curl http://localhost:8080/health`
3. Check browser console for CORS errors

### High latency

1. Check database query performance
2. Enable connection pooling
3. Check Redis cache hit rate
4. Review trace data for slow spans
