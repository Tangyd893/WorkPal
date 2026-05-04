.PHONY: deps run-gateway run-user run-im run-file run-search run-workspace test build-services docker-up docker-down lint clean help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Docker
DC=docker compose -f ../docker/docker-compose.yaml

deps:
	$(GOMOD) download
	$(GOMOD) tidy

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

run-workspace:
	$(GOCMD) run ./cmd/workspace-service

build-services:
	mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) -o bin/gateway ./cmd/gateway
	CGO_ENABLED=0 $(GOBUILD) -o bin/user-service ./cmd/user-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/im-service ./cmd/im-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/file-service ./cmd/file-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/search-service ./cmd/search-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/workspace-service ./cmd/workspace-service

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
	rm -rf bin

help:
	@echo "WorkPal Backend Makefile"
	@echo ""
	@echo "  make deps        - Download and tidy Go dependencies"
	@echo "  make docker-up   - Start the full microservice stack in Docker"
	@echo "  make docker-down - Stop containers"
	@echo "  make build-services - Build gateway and microservice binaries"
	@echo "  make run-gateway - Run API gateway"
	@echo "  make run-user    - Run user service"
	@echo "  make run-im      - Run IM service"
	@echo "  make run-file    - Run file service"
	@echo "  make run-search  - Run search service"
	@echo "  make run-workspace - Run workspace service"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run golangci-lint"
