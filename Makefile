# =============================================================================
# Makefile for Go Cats API
# Enhanced with CI/CD, Docker, and comprehensive testing support
# =============================================================================

# Variables
APP_NAME := cats-api
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
DOCKER_IMAGE := ghcr.io/st4r4x/golangapp
GO_VERSION := 1.21

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help test coverage clean build run docker lint security dev-setup ci-local

# Default target
all: clean lint test build

# =============================================================================
# Help
# =============================================================================
help: ## Show this help message
	@echo "$(GREEN)Go Cats API - Available Commands$(NC)"
	@echo "=================================="
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(BLUE)<target>$(NC)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
dev-setup: ## Set up development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	go mod download
	go mod verify
	go install github.com/cosmtrek/air@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "$(GREEN)Development environment ready!$(NC)"

dev: ## Start development server with hot reload
	@echo "$(BLUE)Starting development server with hot reload...$(NC)"
	air -c .air.toml

run: ## Run the application
	@echo "$(BLUE)Starting Go Cats API...$(NC)"
	go run .

build: ## Build the application
	@echo "$(BLUE)Building $(APP_NAME)...$(NC)"
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build \
		-ldflags="-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" \
		-o bin/$(APP_NAME) .
	@echo "$(GREEN)Build complete: bin/$(APP_NAME)$(NC)"

##@ Testing
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(NC)"
	go test -v ./...

test-unit: ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(NC)"
	go test -v ./test/unit/...

test-integration: ## Run integration tests only
	@echo "$(BLUE)Running integration tests...$(NC)"
	go test -v ./test/integration/...

test-api: ## Run API tests (requires running server)
	@echo "$(BLUE)Running API tests...$(NC)"
	go test -v ./test/apitests/...

test-mocked: ## Run mocked tests only
	@echo "$(BLUE)Running mocked tests...$(NC)"
	go test -v ./test/mocked/...

test-main: ## Run main package tests only
	@echo "$(BLUE)Running main package tests...$(NC)"
	go test -v .

##@ Coverage
coverage: ## Generate comprehensive coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	go test -coverprofile=docs/coverage.out ./... -coverpkg=./...
	go tool cover -html=docs/coverage.out -o docs/coverage.html
	@echo "$(GREEN)Coverage report generated: docs/coverage.html$(NC)"
	go tool cover -func=docs/coverage.out | grep "total:"

coverage-unit: ## Generate unit test coverage
	@echo "$(BLUE)Generating unit test coverage...$(NC)"
	go test -coverprofile=unit-coverage.out ./test/unit/... -coverpkg=./...
	go tool cover -func=unit-coverage.out

coverage-integration: ## Generate integration test coverage
	@echo "$(BLUE)Generating integration test coverage...$(NC)"
	go test -coverprofile=integration-coverage.out ./test/integration/... -coverpkg=./...
	go tool cover -func=integration-coverage.out

##@ Code Quality
lint: ## Run linting and formatting checks
	@echo "$(BLUE)Running linting checks...$(NC)"
	go fmt ./...
	go vet ./...
	staticcheck ./...
	@echo "$(GREEN)Linting complete!$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

security: ## Run security checks
	@echo "$(BLUE)Running security checks...$(NC)"
	go mod verify
	@echo "$(GREEN)Security checks complete!$(NC)"

##@ Docker
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(VERSION)$(NC)"

docker-build-dev: ## Build development Docker image
	@echo "$(BLUE)Building development Docker image...$(NC)"
	docker build --target development -t $(DOCKER_IMAGE):dev .
	@echo "$(GREEN)Development Docker image built: $(DOCKER_IMAGE):dev$(NC)"

docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 8080:8080 --name $(APP_NAME) $(DOCKER_IMAGE):latest

docker-run-dev: ## Run development Docker container
	@echo "$(BLUE)Running development Docker container...$(NC)"
	docker run -p 8080:8080 -v $(PWD):/app --name $(APP_NAME)-dev $(DOCKER_IMAGE):dev

docker-stop: ## Stop Docker container
	@echo "$(BLUE)Stopping Docker container...$(NC)"
	-docker stop $(APP_NAME) $(APP_NAME)-dev
	-docker rm $(APP_NAME) $(APP_NAME)-dev

##@ Docker Compose
compose-up: ## Start all services with Docker Compose
	@echo "$(BLUE)Starting services with Docker Compose...$(NC)"
	docker-compose up -d api-dev

compose-up-prod: ## Start production services
	@echo "$(BLUE)Starting production services...$(NC)"
	docker-compose up -d api-prod

compose-up-test: ## Start test services
	@echo "$(BLUE)Starting test services...$(NC)"
	docker-compose up api-test

compose-up-monitoring: ## Start with monitoring
	@echo "$(BLUE)Starting services with monitoring...$(NC)"
	docker-compose --profile monitoring up -d

compose-down: ## Stop all Docker Compose services
	@echo "$(BLUE)Stopping Docker Compose services...$(NC)"
	docker-compose down --remove-orphans

compose-logs: ## Show Docker Compose logs
	docker-compose logs -f

##@ CI/CD
ci-local: ## Run CI pipeline locally
	@echo "$(BLUE)Running CI pipeline locally...$(NC)"
	@echo "$(YELLOW)1. Linting and formatting...$(NC)"
	$(MAKE) lint
	@echo "$(YELLOW)2. Running unit tests...$(NC)"
	$(MAKE) test-unit
	@echo "$(YELLOW)3. Running integration tests...$(NC)"
	$(MAKE) test-integration
	@echo "$(YELLOW)4. Running main package tests...$(NC)"
	$(MAKE) test-main
	@echo "$(YELLOW)5. Generating coverage report...$(NC)"
	$(MAKE) coverage
	@echo "$(YELLOW)6. Building application...$(NC)"
	$(MAKE) build
	@echo "$(YELLOW)7. Building Docker image...$(NC)"
	$(MAKE) docker-build
	@echo "$(GREEN)CI pipeline completed successfully!$(NC)"

pre-commit: ## Run pre-commit checks
	@echo "$(BLUE)Running pre-commit checks...$(NC)"
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

##@ Utilities
clean: ## Clean up generated files
	@echo "$(BLUE)Cleaning up...$(NC)"
	rm -rf bin/
	rm -f docs/coverage.out docs/coverage.html
	rm -f *-coverage.out coverage.out
	rm -rf logs/*.log
	rm -rf tmp/
	docker system prune -f
	@echo "$(GREEN)Cleanup complete!$(NC)"

deps: ## Download and verify dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	go mod verify
	go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	$(MAKE) coverage
	@echo "$(GREEN)Documentation generated in docs/ directory$(NC)"

health-check: ## Check if the API is healthy
	@echo "$(BLUE)Checking API health...$(NC)"
	@curl -f http://localhost:8080/ || (echo "$(RED)API is not responding$(NC)" && exit 1)
	@echo "$(GREEN)API is healthy!$(NC)"

version: ## Show version information
	@echo "$(GREEN)Version Information:$(NC)"
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"
