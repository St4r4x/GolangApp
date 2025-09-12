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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not available"
        exit 1
    fi
    
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
    docker-compose -f "$COMPOSE_FILE" pull
    
    # Deploy with specified replicas
    log_info "Starting services with $replicas API replicas..."
    docker-compose -f "$COMPOSE_FILE" up -d --scale cats-api="$replicas"
    
    # Wait for services to be healthy
    log_info "Waiting for services to become healthy..."
    sleep 10
    
    # Check service status
    check_services
}

# Check service health and status
check_services() {
    log_info "Checking service health..."
    
    # Check if containers are running
    local running_containers
    running_containers=$(docker-compose -f "$COMPOSE_FILE" ps -q | wc -l)
    
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
    
    local test_requests=10
    local unique_servers=0
    
    echo "Making $test_requests requests to test load distribution..."
    
    for i in $(seq 1 $test_requests); do
        local response
        response=$(curl -s "http://localhost:$DEFAULT_PORT/" | grep -o 'Server ID: [^<]*' || echo "Server ID: unknown")
        echo "  Request $i: $response"
    done
    
    log_success "Load balancing test completed"
}

# Show deployment status
show_status() {
    log_info "=== Deployment Status ==="
    
    echo ""
    echo "🐳 Container Status:"
    docker-compose -f "$COMPOSE_FILE" ps
    
    echo ""
    echo "📊 Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
    
    echo ""
    echo "🌐 Service Endpoints:"
    echo "  - API (via Load Balancer): http://localhost:$DEFAULT_PORT/"
    echo "  - Load Balancer Health: http://localhost:$DEFAULT_PORT/health"
    
    echo ""
    echo "📝 Quick Commands:"
    echo "  - View logs: docker-compose -f $COMPOSE_FILE logs -f"
    echo "  - Scale API: docker-compose -f $COMPOSE_FILE up -d --scale cats-api=N"
    echo "  - Stop services: docker-compose -f $COMPOSE_FILE down"
}

# Scale the API service
scale_service() {
    local new_replicas=$1
    
    if [ -z "$new_replicas" ] || ! [[ "$new_replicas" =~ ^[0-9]+$ ]]; then
        log_error "Please provide a valid number of replicas"
        exit 1
    fi
    
    log_info "Scaling cats-api service to $new_replicas replicas..."
    docker-compose -f "$COMPOSE_FILE" up -d --scale cats-api="$new_replicas"
    
    sleep 5
    log_success "Service scaled to $new_replicas replicas"
    show_status
}

# Stop the integrated stack
stop_stack() {
    log_info "Stopping integrated stack..."
    docker-compose -f "$COMPOSE_FILE" down
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
            show_status
            ;;
        "scale")
            check_prerequisites
            scale_service "$2"
            ;;
        "test")
            test_load_balancing
            ;;
        "logs")
            docker-compose -f "$COMPOSE_FILE" logs -f
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
