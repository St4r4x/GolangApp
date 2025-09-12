# =============================================================================
# Makefile for Multi-Service Go Application
# Simplified and organized for microservices architecture
# =============================================================================

# Variables
APP_NAME := cats-api
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
DOCKER_IMAGE := ghcr.io/st4r4x/golangapp
REPLICAS := 2

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help up down scale logs test clean

# Default target
all: clean test up

# =============================================================================
# Help
# =============================================================================
help: ## Show this help message
	@echo "$(GREEN)Multi-Service Go Application - Available Commands$(NC)"
	@echo "=================================================="
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(BLUE)<target>$(NC)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Core Operations
up: ## Start all services (load balancer + API replicas)
	@echo "$(BLUE)Starting all services...$(NC)"
	docker compose up --build -d
	@echo "$(GREEN)Services running!$(NC)"
	@echo "$(YELLOW)Access: http://localhost:4443$(NC)"
	@echo "$(YELLOW)Swagger: http://localhost:4443/swagger/$(NC)"

down: ## Stop all services
	@echo "$(BLUE)Stopping all services...$(NC)"
	docker compose down
	@echo "$(GREEN)All services stopped!$(NC)"

scale: ## Scale API replicas (usage: make scale REPLICAS=5)
	@echo "$(BLUE)Scaling to $(REPLICAS) replicas...$(NC)"
	@if [ "$(REPLICAS)" -lt 1 ] || [ "$(REPLICAS)" -gt 10 ]; then \
		echo "$(RED)Error: REPLICAS must be between 1 and 10$(NC)"; \
		exit 1; \
	fi
	@sed -i "s/replicas: [0-9]/replicas: $(REPLICAS)/" docker-compose.yml
	docker compose up --build -d
	@echo "$(GREEN)Scaled to $(REPLICAS) replicas!$(NC)"

restart: ## Restart all services
	@echo "$(BLUE)Restarting all services...$(NC)"
	docker compose restart
	@echo "$(GREEN)All services restarted!$(NC)"

##@ Development
dev: ## Start development with hot reload
	@echo "$(BLUE)Starting development environment...$(NC)"
	cd projects/cats-api && air -c .air.toml

build: ## Build the application
	@echo "$(BLUE)Building $(APP_NAME)...$(NC)"
	cd projects/cats-api && mkdir -p bin
	cd projects/cats-api && CGO_ENABLED=0 GOOS=linux go build \
		-ldflags="-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" \
		-o bin/$(APP_NAME) .
	@echo "$(GREEN)Build complete: projects/cats-api/bin/$(APP_NAME)$(NC)"

##@ Testing
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(NC)"
	cd projects/cats-api && go test -v ./...

coverage: ## Generate coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	cd projects/cats-api && go test -coverprofile=docs/coverage.out ./... -coverpkg=./...
	cd projects/cats-api && go tool cover -html=docs/coverage.out -o docs/coverage.html
	@echo "$(GREEN)Coverage report: projects/cats-api/docs/coverage.html$(NC)"
	cd projects/cats-api && go tool cover -func=docs/coverage.out | grep "total:"

test-load: ## Test load balancing
	@echo "$(BLUE)Testing load balancing...$(NC)"
	@for i in 1 2 3 4 5 6; do \
		echo "Request $$i: $$(date '+%H:%M:%S')"; \
		curl -s -I http://localhost:4443/ | grep "X-Server-Id" | sed 's/X-Server-Id: /  → Server: /'; \
		sleep 0.5; \
	done
	@echo "$(GREEN)Load balancing test completed!$(NC)"

##@ Monitoring
logs: ## Show logs from all services
	@echo "$(BLUE)Showing logs from all services...$(NC)"
	docker compose logs -f

status: ## Show service status
	@echo "$(BLUE)Service status:$(NC)"
	docker compose ps

health: ## Check service health
	@echo "$(BLUE)Checking service health...$(NC)"
	@curl -f http://localhost:4443/ && echo "$(GREEN)✓ API healthy$(NC)" || echo "$(RED)✗ API unhealthy$(NC)"

##@ Maintenance
clean: ## Clean up containers and images
	@echo "$(BLUE)Cleaning up...$(NC)"
	docker compose down --remove-orphans
	docker system prune -f
	cd projects/cats-api && rm -rf bin/ docs/coverage.*
	@echo "$(GREEN)Cleanup complete!$(NC)"

update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	cd projects/cats-api && go mod download && go mod tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

version: ## Show version information
	@echo "$(GREEN)Version Information:$(NC)"
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Current Replicas: $$(grep 'replicas:' docker-compose.yml | awk '{print $$2}')"
