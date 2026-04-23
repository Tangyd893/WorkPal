.PHONY: deps run test build docker-up docker-down migrate lint clean swag

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_SERVER=bin/server
BINARY_MIGRATOR=bin/migrator

# Docker
DC=docker compose -f deployments/docker/docker-compose.yaml

deps:
	$(GOMOD) download
	$(GOMOD) tidy

run:
	$(GOCMD) run ./cmd/server

run:migrate
	@echo "Make sure postgres and redis are running: make docker-up"

build:
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_SERVER) ./cmd/server

build-migrator:
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_MIGRATOR) ./cmd/migrator

test:
	$(GOTEST) -v -race -cover ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

migrate:
	$(GOCMD) run ./cmd/migrator

docker-up:
	$(DC) up -d
	@echo "Waiting for postgres..."
	@sleep 3
	$(DC) ps

docker-down:
	$(DC) down

docker-restart:
	$(DC) restart

swag:
	swag init -g cmd/server/main.go -o internal/common/docs

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# 生成随机密钥（用于首次运行）
gen-secret:
	@openssl rand -base64 32

# 初始化数据库（创建用户/数据库）
init-db:
	@echo "Creating postgres user and database..."
	@docker exec -it workpal-postgres psql -U workpal -c "CREATE DATABASE workpal;" || echo "DB may already exist"

help:
	@echo "WorkPal Makefile"
	@echo ""
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make docker-up     - Start postgres and redis containers"
	@echo "  make docker-down   - Stop containers"
	@echo "  make migrate       - Run database migrations"
	@echo "  make build        - Build server binary"
	@echo "  make run          - Run server (requires docker-up first)"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make swag         - Generate Swagger docs"
