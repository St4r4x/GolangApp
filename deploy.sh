#!/bin/bash
# =============================================================================
# Integrated Cats API + Reverse Proxy Deployment Script
# =============================================================================
# This script deploys the complete microservices package with:
# - Cats API backend (scalable replicas)
# - Custom Go reverse proxy load balancer
# - Production-ready configuration
# =============================================================================

set -e

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
SERVICE_NAME="cats-api-stack"
DEFAULT_REPLICAS=5
DEFAULT_PORT=4443
DOCKER_COMPOSE_CMD=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get the correct docker compose command
get_docker_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    elif docker compose version &> /dev/null 2>&1; then
        echo "docker compose"
    else
        log_error "Neither 'docker-compose' nor 'docker compose' is available"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check Docker Compose availability and set command
    DOCKER_COMPOSE_CMD=$(get_docker_compose_cmd)
    export DOCKER_COMPOSE_CMD
    log_info "Using Docker Compose command: $DOCKER_COMPOSE_CMD"
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Production compose file '$COMPOSE_FILE' not found"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Deploy the integrated stack
deploy_stack() {
    local replicas=${1:-$DEFAULT_REPLICAS}
    
    log_info "Deploying integrated Cats API + Reverse Proxy stack..."
    log_info "Configuration:"
    echo "  - API Replicas: $replicas"
    echo "  - External Port: $DEFAULT_PORT"
    echo "  - Load Balancer: Custom Go Reverse Proxy"
    
    # Pull latest images
    log_info "Pulling latest Docker images..."
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" pull
    
    # Deploy with specified replicas
    log_info "Starting services with $replicas API replicas..."
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" up -d --scale cats-api="$replicas"
    
    # Wait for services to be healthy
    log_info "Waiting for services to become healthy..."
    sleep 10
    
    # Reload load balancer to discover all backends
    reload_load_balancer
    
    # Check service status
    check_services
}

# Reload load balancer to discover new backends
reload_load_balancer() {
    log_info "Reloading load balancer to discover new backends..."
    
    # Restart the reverse proxy to pick up new backend containers
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" restart reverse-proxy
    
    # Wait for reverse proxy to restart and discover backends
    log_info "Waiting for load balancer to discover new backends..."
    sleep 5
    
    # Verify load balancer is responding
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s -m 5 "http://localhost:$DEFAULT_PORT/" > /dev/null 2>&1; then
            log_success "✅ Load balancer reloaded and responding"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_error "❌ Load balancer failed to respond after reload"
            return 1
        fi
        
        if [ $((attempt % 10)) -eq 0 ]; then
            log_info "Still waiting for load balancer... (attempt $attempt/$max_attempts)"
        fi
        sleep 1
        ((attempt++))
    done
}

# Check service health and status
check_services() {
    log_info "Checking service health..."
    
    # Check if containers are running
    local running_containers
    running_containers=$($DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" ps -q | wc -l)
    
    if [ "$running_containers" -eq 0 ]; then
        log_error "No containers are running"
        return 1
    fi
    
    # Check reverse proxy health
    log_info "Testing reverse proxy endpoint..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s -m 5 "http://localhost:$DEFAULT_PORT/" > /dev/null 2>&1; then
            log_success "✅ Reverse proxy is responding on port $DEFAULT_PORT"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_error "❌ Reverse proxy failed to respond after $max_attempts attempts"
            return 1
        fi
        
        log_info "Attempt $attempt/$max_attempts - waiting for reverse proxy..."
        sleep 2
        ((attempt++))
    done
    
    # Test load balancing
    test_load_balancing
}

# Test load balancing across replicas
test_load_balancing() {
    log_info "Testing load balancing across replicas..."
    
    local test_requests=20
    declare -A server_counts
    
    echo "Making $test_requests requests to test load distribution..."
    
    for i in $(seq 1 $test_requests); do
        local server_id
        server_id=$(curl -s -I "http://localhost:$DEFAULT_PORT/" 2>/dev/null | grep -i "x-server-id" | cut -d' ' -f2- | tr -d '\r\n')
        
        if [ -n "$server_id" ]; then
            # Extract just the container ID part (first 12 characters)
            local container_id=${server_id:0:12}
            server_counts["$container_id"]=$((${server_counts["$container_id"]} + 1))
            if [ $((i % 5)) -eq 0 ]; then
                echo "  Requests 1-$i: Load balancing active..."
            fi
        else
            echo "  Request $i: Server ID not found"
        fi
    done
    
    echo ""
    echo "📊 Load Distribution Summary:"
    local unique_servers=${#server_counts[@]}
    local total_containers
    total_containers=$($DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" ps -q cats-api | wc -l)
    
    if [ $unique_servers -gt 0 ]; then
        echo "  📈 Active Backends: $unique_servers servers"
        echo "  🐳 Total Containers: $total_containers containers"
        echo ""
        for server in "${!server_counts[@]}"; do
            local percentage=$((${server_counts[$server]} * 100 / test_requests))
            echo "  - Server $server: ${server_counts[$server]} requests ($percentage%)"
        done
        echo ""
        if [ $unique_servers -gt 1 ]; then
            log_success "✅ Perfect round-robin load balancing detected across $unique_servers healthy servers!"
            if [ $unique_servers -lt $total_containers ]; then
                log_warning "ℹ️  Note: $((total_containers - unique_servers)) containers may be starting up or unhealthy"
            fi
        else
            log_warning "⚠️  All requests went to the same server - check load balancer configuration"
        fi
    else
        log_error "❌ Could not detect server distribution"
    fi
    
    log_success "Load balancing test completed"
}

# Show deployment status
show_status() {
    log_info "=== Deployment Status ==="
    
    echo ""
    echo "🐳 Container Status:"
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" ps
    
    echo ""
    echo "📊 Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
    
    echo ""
    echo "🌐 Service Endpoints:"
    echo "  - API (via Load Balancer): http://localhost:$DEFAULT_PORT/"
    echo "  - Load Balancer Health: http://localhost:$DEFAULT_PORT/health"
    
    echo ""
    echo "📝 Quick Commands:"
    echo "  - View logs: ./deploy.sh logs"
    echo "  - Scale API: ./deploy.sh scale N"
    echo "  - Stop services: ./deploy.sh stop"
}

# Scale the API service
scale_service() {
    local new_replicas=$1
    
    if [ -z "$new_replicas" ] || ! [[ "$new_replicas" =~ ^[0-9]+$ ]]; then
        log_error "Please provide a valid number of replicas"
        exit 1
    fi
    
    log_info "Scaling cats-api service to $new_replicas replicas..."
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" up -d --scale cats-api="$new_replicas"
    
    # Give containers time to start
    sleep 5
    
    # Reload load balancer to discover new backends
    reload_load_balancer
    
    log_success "Service scaled to $new_replicas replicas"
    show_status
}

# Stop the integrated stack
stop_stack() {
    log_info "Stopping integrated stack..."
    $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" down
    log_success "Stack stopped successfully"
}

# Show usage information
show_usage() {
    echo "Integrated Cats API + Reverse Proxy Deployment Script"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  deploy [replicas]  Deploy the integrated stack (default: $DEFAULT_REPLICAS replicas)"
    echo "  status            Show deployment status and health"
    echo "  scale <replicas>  Scale the API service to specified replicas"
    echo "  test              Test load balancing functionality"
    echo "  logs              Show service logs"
    echo "  stop              Stop all services"
    echo "  help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 deploy 3       Deploy with 3 API replicas"
    echo "  $0 scale 10       Scale to 10 API replicas"
    echo "  $0 test           Test load balancing"
    echo "  $0 status         Show current status"
}

# Main script logic
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            deploy_stack "$2"
            show_status
            ;;
        "status")
            check_prerequisites
            show_status
            ;;
        "scale")
            check_prerequisites
            scale_service "$2"
            ;;
        "test")
            check_prerequisites
            test_load_balancing
            ;;
        "logs")
            check_prerequisites
            $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" logs -f
            ;;
        "stop")
            check_prerequisites
            stop_stack
            ;;
        "help"|"-h"|"--help")
            show_usage
            ;;
        *)
            log_error "Unknown command: $1"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
