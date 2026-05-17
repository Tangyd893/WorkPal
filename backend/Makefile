.PHONY: deps run-gateway run-user run-im run-file run-search run-workspace run-project run-notification test build-services docker-up docker-down lint clean help migrate-install migrate-up migrate-down migrate-create

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

run-project:
	$(GOCMD) run ./cmd/project-service

run-notification:
	$(GOCMD) run ./cmd/notification-service

build-services:
	mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) -o bin/gateway ./cmd/gateway
	CGO_ENABLED=0 $(GOBUILD) -o bin/user-service ./cmd/user-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/im-service ./cmd/im-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/file-service ./cmd/file-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/search-service ./cmd/search-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/workspace-service ./cmd/workspace-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/project-service ./cmd/project-service
	CGO_ENABLED=0 $(GOBUILD) -o bin/notification-service ./cmd/notification-service

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
  @echo "  make run-project  - Run project service"
  @echo "  make run-notification - Run notification service"
  @echo "  make test            - Run tests"
  @echo "  make lint            - Run golangci-lint"
  @echo "  make migrate-install - Install golang-migrate CLI"
  @echo "  make migrate-up SERVICE=<name>       - Run migrations up for one service"
  @echo "  make migrate-down SERVICE=<name>     - Rollback last migration"
  @echo "  make migrate-create SERVICE=<name> NAME=<name> - Create a new migration pair"

# Migration parameters
PG_USER ?= workpal
PG_PASSWORD ?= workpal123
PG_HOST ?= localhost
PG_PORT ?= 5432
PG_DB ?= workpal
PG_URL := postgres://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DB)?sslmode=disable

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make migrate-up SERVICE=user-service"; exit 1; fi
	migrate -path migrations/$(SERVICE) -database "$(PG_URL)" up

migrate-down:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make migrate-down SERVICE=user-service"; exit 1; fi
	migrate -path migrations/$(SERVICE) -database "$(PG_URL)" down 1

migrate-create:
	@if [ -z "$(SERVICE)" ] || [ -z "$(NAME)" ]; then echo "Usage: make migrate-create SERVICE=user-service NAME=add_avatar"; exit 1; fi
	migrate create -ext sql -dir migrations/$(SERVICE) -seq $(NAME)
