# Library Management System Makefile

# Variables
BINARY_NAME=library
API_BINARY=bin/library-api
WORKER_BINARY=bin/library-worker
MIGRATE_BINARY=bin/library-migrate
GO=go
GOFLAGS=-v
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_FILE=deployments/docker/docker-compose.yml

# Version information
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

.DEFAULT_GOAL := help

## help: Display this help message
.PHONY: help
help:
	@echo "Library Management System - Available Commands"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'

## init: Initialize project dependencies
.PHONY: init
init:
	@echo "$(GREEN)Initializing project...$(NC)"
	$(GO) mod download
	$(GO) mod vendor
	@echo "$(GREEN)Project initialized!$(NC)"

## build: Build all binaries
.PHONY: build
build: build-api build-worker build-migrate
	@echo "$(GREEN)All binaries built successfully!$(NC)"

## build-api: Build API server binary
.PHONY: build-api
build-api:
	@echo "$(YELLOW)Building API server...$(NC)"
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(API_BINARY) ./cmd/api
	@echo "$(GREEN)API server built: $(API_BINARY)$(NC)"

## build-worker: Build worker binary
.PHONY: build-worker
build-worker:
	@echo "$(YELLOW)Building worker...$(NC)"
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(WORKER_BINARY) ./cmd/worker
	@echo "$(GREEN)Worker built: $(WORKER_BINARY)$(NC)"

## build-migrate: Build migration tool binary
.PHONY: build-migrate
build-migrate:
	@echo "$(YELLOW)Building migration tool...$(NC)"
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(MIGRATE_BINARY) ./cmd/migrate
	@echo "$(GREEN)Migration tool built: $(MIGRATE_BINARY)$(NC)"

## run: Run API server locally
.PHONY: run
run:
	@echo "$(YELLOW)Starting API server...$(NC)"
	$(GO) run ./cmd/api

## run-worker: Run worker locally
.PHONY: run-worker
run-worker:
	@echo "$(YELLOW)Starting worker...$(NC)"
	$(GO) run ./cmd/worker

## test: Run all tests
.PHONY: test
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "$(GREEN)Tests completed!$(NC)"

## test-unit: Run unit tests only
.PHONY: test-unit
test-unit:
	@echo "$(YELLOW)Running unit tests...$(NC)"
	$(GO) test -v -short ./...
	@echo "$(GREEN)Unit tests completed!$(NC)"

## test-integration: Run integration tests only
.PHONY: test-integration
test-integration:
	@echo "$(YELLOW)Running integration tests...$(NC)"
	$(GO) test -v -tags=integration ./test/integration/...
	@echo "$(GREEN)Integration tests completed!$(NC)"

## test-coverage: Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## lint: Run linters
.PHONY: lint
lint:
	@echo "$(YELLOW)Running linters...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)golangci-lint not installed. Run: make install-tools$(NC)"; \
	fi

## fmt: Format code
.PHONY: fmt
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)Vet completed!$(NC)"

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Cleaned!$(NC)"

## clean-logs: Remove all log files from source code
.PHONY: clean-logs
clean-logs:
	@echo "$(YELLOW)Removing log files from source code...$(NC)"
	@find . -name "*.log" -type f -not -path "./vendor/*" -not -path "./node_modules/*" -not -path "./logs/*" -delete
	@echo "$(GREEN)Log files cleaned!$(NC)"

## migrate-up: Run store migrations
.PHONY: migrate-up
migrate-up:
	@echo "$(YELLOW)Running migrations...$(NC)"
	$(GO) run ./cmd/migrate up

## migrate-down: Rollback store migrations
.PHONY: migrate-down
migrate-down:
	@echo "$(YELLOW)Rolling back migrations...$(NC)"
	$(GO) run ./cmd/migrate down

## migrate-create: Create a new migration (usage: make migrate-create name=migration_name)
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Please provide migration name. Usage: make migrate-create name=migration_name$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Creating migration: $(name)...$(NC)"
	$(GO) run ./cmd/migrate create $(name)

## docker-up: Start all service with docker-compose (alias: up)
.PHONY: docker-up up
docker-up up:
	@echo "$(YELLOW)Starting Docker services...$(NC)"
	cd deployments/docker && $(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started!$(NC)"

## docker-down: Stop all service with docker-compose (alias: down)
.PHONY: docker-down down
docker-down down:
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	cd deployments/docker && $(DOCKER_COMPOSE) down
	@echo "$(GREEN)Services stopped!$(NC)"

## docker-logs: Show logs from docker-compose service
.PHONY: docker-logs
docker-logs:
	cd deployments/docker && $(DOCKER_COMPOSE) logs -f

## docker-build: Build Docker images
.PHONY: docker-build
docker-build:
	@echo "$(YELLOW)Building Docker images...$(NC)"
	cd deployments/docker && $(DOCKER_COMPOSE) build
	@echo "$(GREEN)Docker images built!$(NC)"

## restart: Restart docker service (alias for backward compatibility)
.PHONY: restart
restart: docker-down docker-up

## install-tools: Install development tools
.PHONY: install-tools
install-tools:
	@echo "$(YELLOW)Installing development tools...$(NC)"
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/golang/mock/mockgen@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Tools installed!$(NC)"

## install-hooks: Install git hooks
.PHONY: install-hooks
install-hooks:
	@echo "$(YELLOW)Installing git hooks...$(NC)"
	@chmod +x .githooks/pre-commit
	@git config core.hooksPath .githooks
	@echo "$(GREEN)Git hooks installed! Pre-commit checks will now run automatically.$(NC)"

## gen-mocks: Generate mocks for testing
.PHONY: gen-mocks
gen-mocks:
	@echo "$(YELLOW)Generating mocks...$(NC)"
	$(GO) generate ./...
	@echo "$(GREEN)Mocks generated!$(NC)"

## gen-docs: Generate API documentation
.PHONY: gen-docs
gen-docs:
	@echo "$(YELLOW)Generating API documentation...$(NC)"
	@if command -v swag &> /dev/null; then \
		swag init -g cmd/api/main.go -o docs; \
		echo "$(GREEN)API documentation generated!$(NC)"; \
	else \
		echo "$(RED)swag not installed. Run: make install-tools$(NC)"; \
	fi

## dev: Start development environment
.PHONY: dev
dev: docker-up migrate-up
	@echo "$(YELLOW)Starting development server...$(NC)"
	$(GO) run ./cmd/api

## ci: Run CI pipeline locally
.PHONY: ci
ci: fmt vet lint test build
	@echo "$(GREEN)CI pipeline completed successfully!$(NC)"

## check: Run all checks (format, vet, lint)
.PHONY: check
check: fmt vet lint
	@echo "$(GREEN)All checks passed!$(NC)"

## mod-tidy: Tidy go modules
.PHONY: mod-tidy
mod-tidy:
	@echo "$(YELLOW)Tidying go modules...$(NC)"
	$(GO) mod tidy
	$(GO) mod vendor
	@echo "$(GREEN)Modules tidied!$(NC)"

## mod-update: Update go modules
.PHONY: mod-update
mod-update:
	@echo "$(YELLOW)Updating go modules...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy
	$(GO) mod vendor
	@echo "$(GREEN)Modules updated!$(NC)"

## benchmark: Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...
	@echo "$(GREEN)Benchmarks completed!$(NC)"

## security: Run security checks
.PHONY: security
security:
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if command -v gosec &> /dev/null; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)Installing gosec...$(NC)"; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi
	@echo "$(GREEN)Security checks completed!$(NC)"

## version: Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

.PHONY: all
all: clean build test