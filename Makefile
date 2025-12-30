.PHONY: run test test-integration build docker-build docker-up docker-down migrate clean help

# Variables
BINARY_NAME=gateway
DOCKER_COMPOSE=docker-compose -f deployments/docker-compose.yml

# Help
help:
	@echo "GatewayOps MCP Gateway"
	@echo ""
	@echo "Usage:"
	@echo "  make run              Run gateway locally"
	@echo "  make test             Run unit tests"
	@echo "  make test-integration Run integration tests"
	@echo "  make build            Build binary"
	@echo "  make docker-build     Build Docker images"
	@echo "  make docker-up        Start all services"
	@echo "  make docker-down      Stop all services"
	@echo "  make migrate          Run database migrations"
	@echo "  make generate-key     Generate a new API key"
	@echo "  make clean            Clean build artifacts"

# Development
run:
	go run gateway/cmd/gateway/main.go

test:
	go test -v -race ./...

test-integration:
	$(DOCKER_COMPOSE) up -d postgres redis clickhouse mock-mcp
	sleep 5
	go test -v -tags=integration ./test/integration/...
	$(DOCKER_COMPOSE) down

# Build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY_NAME) gateway/cmd/gateway/main.go

docker-build:
	docker build -t gatewayops-gateway:latest -f deployments/docker/Dockerfile.gateway .
	docker build -t gatewayops-mock-mcp:latest -f deployments/docker/Dockerfile.mock-mcp .

# Docker Compose
docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

# Database
migrate:
	@echo "Running PostgreSQL migrations..."
	docker exec -i gatewayops-postgres psql -U postgres -d gatewayops < migrations/postgres/001_initial.sql
	@echo "Running ClickHouse migrations..."
	docker exec -i gatewayops-clickhouse clickhouse-client --multiquery < migrations/clickhouse/001_initial.sql
	@echo "Migrations complete"

# Utilities
generate-key:
	go run scripts/generate-key.go

clean:
	rm -rf bin/
	go clean

# Linting
lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .
