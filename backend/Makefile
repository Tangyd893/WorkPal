.PHONY: deps run run-server run-gateway run-user run-im run-file run-search test build build-services docker-up docker-down migrate lint clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY=server

# Docker
DC=docker compose -f ../docker/docker-compose.yaml

deps:
	$(GOMOD) download
	$(GOMOD) tidy

run:
	$(GOCMD) run ./cmd/server

run-server:
	$(GOCMD) run ./cmd/server

run-gateway:
	$(GOCMD) run ./cmd/gateway

run-user:
	$(GOCMD) run ./cmd/user-service

run-im:
	$(GOCMD) run ./cmd/im-service

run-file:
	$(GOCMD) run ./cmd/file-service

run-search:
	$(GOCMD) run ./cmd/search-service

build:
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY) ./cmd/server

build-services:
	mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) -o bin/gateway ./cmd/gateway
	CGO_ENABLED=0 $(GOBUILD) -o bin/user-service ./cmd/user-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/im-service ./cmd/im-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/file-service ./cmd/file-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/search-service ./cmd/search-service

test:
	$(GOTEST) -v -race -cover ./...

lint:
	golangci-lint run ./...

docker-up:
	$(DC) up -d
	@echo "Waiting for postgres..."
	@sleep 3
	$(DC) ps

docker-down:
	$(DC) down

clean:
	rm -f $(BINARY)

help:
	@echo "WorkPal Backend Makefile"
	@echo ""
	@echo "  make deps        - Download and tidy Go dependencies"
	@echo "  make docker-up   - Start postgres, redis, and minio containers"
	@echo "  make docker-down - Stop containers"
	@echo "  make build       - Build all-in-one server binary"
	@echo "  make build-services - Build gateway and microservice binaries"
	@echo "  make run         - Run all-in-one server (compatibility mode)"
	@echo "  make run-gateway - Run API gateway"
	@echo "  make run-user    - Run user service"
	@echo "  make run-im      - Run IM service"
	@echo "  make run-file    - Run file service"
	@echo "  make run-search  - Run search service"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run golangci-lint"
