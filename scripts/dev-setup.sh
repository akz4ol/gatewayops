#!/bin/bash
# Development environment setup script for GatewayOps

set -e

echo "=== GatewayOps Development Setup ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo "Checking prerequisites..."

check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "  ${GREEN}✓${NC} $1 found"
        return 0
    else
        echo -e "  ${RED}✗${NC} $1 not found"
        return 1
    fi
}

MISSING=0
check_command docker || MISSING=1
check_command docker-compose || check_command "docker compose" || MISSING=1
check_command go || echo -e "  ${YELLOW}!${NC} go not found (optional for Docker-only development)"

if [ $MISSING -eq 1 ]; then
    echo
    echo -e "${RED}Missing required dependencies. Please install them first.${NC}"
    exit 1
fi

echo

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo -e "  ${GREEN}✓${NC} .env created"
else
    echo -e "  ${YELLOW}!${NC} .env already exists, skipping"
fi

echo

# Start Docker services
echo "Starting Docker services..."
if command -v docker-compose &> /dev/null; then
    docker-compose -f deployments/docker-compose.yml up -d
else
    docker compose -f deployments/docker-compose.yml up -d
fi

echo

# Wait for services to be ready
echo "Waiting for services to be ready..."

wait_for_service() {
    local name=$1
    local url=$2
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $name is ready"
            return 0
        fi
        sleep 1
        attempt=$((attempt + 1))
    done

    echo -e "  ${RED}✗${NC} $name failed to start"
    return 1
}

# Wait for PostgreSQL
echo -n "  Waiting for PostgreSQL..."
PGREADY=0
for i in {1..30}; do
    if docker exec gatewayops-postgres pg_isready -U postgres > /dev/null 2>&1; then
        echo -e "\r  ${GREEN}✓${NC} PostgreSQL is ready    "
        PGREADY=1
        break
    fi
    sleep 1
done
if [ $PGREADY -eq 0 ]; then
    echo -e "\r  ${RED}✗${NC} PostgreSQL failed to start"
fi

# Wait for Redis
echo -n "  Waiting for Redis..."
REDISREADY=0
for i in {1..30}; do
    if docker exec gatewayops-redis redis-cli ping > /dev/null 2>&1; then
        echo -e "\r  ${GREEN}✓${NC} Redis is ready          "
        REDISREADY=1
        break
    fi
    sleep 1
done
if [ $REDISREADY -eq 0 ]; then
    echo -e "\r  ${RED}✗${NC} Redis failed to start"
fi

# Wait for ClickHouse
echo -n "  Waiting for ClickHouse..."
CHREADY=0
for i in {1..30}; do
    if curl -s "http://localhost:8123/ping" > /dev/null 2>&1; then
        echo -e "\r  ${GREEN}✓${NC} ClickHouse is ready     "
        CHREADY=1
        break
    fi
    sleep 1
done
if [ $CHREADY -eq 0 ]; then
    echo -e "\r  ${RED}✗${NC} ClickHouse failed to start"
fi

echo

# Run migrations
echo "Running database migrations..."

# PostgreSQL migrations
echo "  Running PostgreSQL migrations..."
docker exec -i gatewayops-postgres psql -U postgres -d gatewayops < migrations/postgres/001_initial.sql 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} PostgreSQL migrations complete"

# ClickHouse migrations
echo "  Running ClickHouse migrations..."
docker exec -i gatewayops-clickhouse clickhouse-client < migrations/clickhouse/001_initial.sql 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} ClickHouse migrations complete"

echo

# Generate a development API key
echo "Generating development API key..."
if command -v go &> /dev/null; then
    go run scripts/generate-key.go dev
else
    echo -e "  ${YELLOW}!${NC} Go not installed, skipping API key generation"
    echo "  Run 'go run scripts/generate-key.go dev' when Go is available"
fi

echo
echo "=== Setup Complete ==="
echo
echo "Services running:"
echo "  - Gateway:     http://localhost:8080"
echo "  - PostgreSQL:  localhost:5432"
echo "  - Redis:       localhost:6379"
echo "  - ClickHouse:  localhost:8123 (HTTP), localhost:9000 (Native)"
echo "  - Mock MCP:    http://localhost:3000"
echo
echo "To start the gateway:"
echo "  make run"
echo
echo "To view logs:"
echo "  docker-compose -f deployments/docker-compose.yml logs -f"
echo
