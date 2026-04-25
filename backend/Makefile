.PHONY: deps run test build docker-up docker-down migrate lint clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY=server

# Docker
DC=docker compose -f ../docker-compose.yaml

deps:
	$(GOMOD) download
	$(GOMOD) tidy

run:
	$(GOCMD) run ./cmd/server

build:
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY) ./cmd/server

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
	@echo "  make docker-up   - Start postgres and redis containers"
	@echo "  make docker-down - Stop containers"
	@echo "  make build      - Build server binary"
	@echo "  make run        - Run server (requires docker-up first)"
	@echo "  make test       - Run tests"
	@echo "  make lint       - Run golangci-lint"
