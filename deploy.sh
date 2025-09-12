#!/bin/bash
# =============================================================================
# Integrated Cats API + Reverse Proxy Deployment Script with Strategy Demo
# =============================================================================
# This script deploys the complete microservices package with:
# - Cats API backend (scalable replicas)
# - Custom Go reverse proxy load balancer (5 strategies)
# - Production-ready configuration
# - Interactive strategy demonstration
# =============================================================================

set -e

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
REGISTRY_COMPOSE_FILE="docker-compose.registry.yml"
SERVICE_NAME="cats-api-stack"
DEFAULT_REPLICAS=5
DEFAULT_PORT=4443
DOCKER_COMPOSE_CMD=""
USE_REGISTRY=false

# Load Balancing Strategies
declare -A STRATEGIES=(
    ["roundrobin"]="Round Robin - Equal distribution"
    ["random"]="Random - Prevents cache hotspots"
    ["weighted"]="Weighted Round Robin - Different capacities"
    ["leastconnections"]="Least Connections - Varying request times"
    ["iphash"]="IP Hash - Session affinity"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

log_strategy() {
    echo -e "${PURPLE}[STRATEGY]${NC} $1"
}

log_demo() {
    echo -e "${CYAN}[DEMO]${NC} $1"
}

# Strategy demonstration functions
show_strategies() {
    echo ""
    log_strategy "Available Load Balancing Strategies:"
    echo "======================================"
    local i=1
    for strategy in "${!STRATEGIES[@]}"; do
        echo -e "${CYAN}$i.${NC} ${YELLOW}$strategy${NC} - ${STRATEGIES[$strategy]}"
        ((i++))
    done
    echo ""
}

demo_strategy() {
    local strategy=${1:-"roundrobin"}
    local replicas=${2:-5}
    
    if [[ ! "${!STRATEGIES[@]}" =~ "${strategy}" ]]; then
        log_error "Invalid strategy: $strategy"
        show_strategies
        return 1
    fi
    
    log_demo "Demonstrating $strategy strategy with $replicas replicas..."
    echo ""
    log_info "Strategy: ${YELLOW}$strategy${NC} - ${STRATEGIES[$strategy]}"
    
    # Set environment variable for load balancing strategy
    export LB_STRATEGY=$strategy
    
    # Deploy with the specified strategy
    deploy_stack $replicas
    
    if [ $? -eq 0 ]; then
        echo ""
        log_success "Load balancer deployed with $strategy strategy!"
        log_info "Testing load balancing..."
        test_load_balancing
        
        echo ""
        log_demo "Strategy demonstration complete!"
        echo -e "${CYAN}Access your application at: ${YELLOW}http://localhost:$DEFAULT_PORT${NC}"
        echo -e "${CYAN}View logs with: ${YELLOW}docker logs -f golangapp-reverse-proxy-1${NC}"
    fi
}

interactive_strategy_demo() {
    show_strategies
    echo -e "${CYAN}Choose a strategy to demonstrate (1-5) or 'q' to quit:${NC}"
    read -r choice
    
    case $choice in
        1) demo_strategy "roundrobin" ;;
        2) demo_strategy "random" ;;
        3) demo_strategy "weighted" ;;
        4) demo_strategy "leastconnections" ;;
        5) demo_strategy "iphash" ;;
        q|Q) log_info "Exiting strategy demo."; exit 0 ;;
        *) log_error "Invalid choice. Please run demo again."; exit 1 ;;
    esac
}

test_load_balancing() {
    log_info "Testing load balancing distribution..."
    
    # Wait for services to be ready
    sleep 5
    
    # Test 10 requests to see distribution
    for i in {1..10}; do
        response=$(curl -s http://localhost:$DEFAULT_PORT/ 2>/dev/null)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓${NC} Request $i successful"
        else
            echo -e "${RED}✗${NC} Request $i failed"
        fi
    done
    
    echo ""
    log_info "Check load balancer logs for distribution details:"
    echo -e "${YELLOW}docker logs golangapp-reverse-proxy-1 | grep strategy${NC}"
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
    echo "  - Deployment Mode: $(if [ "$USE_REGISTRY" = true ]; then echo "Registry (ghcr.io)"; else echo "Local Build"; fi)"
    echo "  - Compose File: $COMPOSE_FILE"
    echo "  - API Replicas: $replicas"
    echo "  - External Port: $DEFAULT_PORT"
    echo "  - Load Balancer: Custom Go Reverse Proxy"
    
    # Registry-specific preparations
    if [ "$USE_REGISTRY" = true ]; then
        log_info "Registry mode: Checking for required images..."
        
        # Check if registry images exist locally
        local cats_api_exists=$(docker images -q ghcr.io/st4r4x/golangapp:latest 2>/dev/null)
        local reverse_proxy_exists=$(docker images -q ghcr.io/st4r4x/golangapp/reverse-proxy:latest 2>/dev/null)
        
        if [ -n "$cats_api_exists" ] && [ -n "$reverse_proxy_exists" ]; then
            log_success "✅ Registry images found locally, skipping pull"
        else
            log_info "Pulling latest images from ghcr.io..."
            log_warning "Make sure you're authenticated with GitHub Container Registry"
            log_info "If not authenticated, run: echo \$GITHUB_TOKEN | docker login ghcr.io -u \$GITHUB_USERNAME --password-stdin"
            
            # Pull latest images
            $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" pull
        fi
    else
        log_info "Local build mode: Pulling/building latest Docker images..."
        # Pull latest images
        $DOCKER_COMPOSE_CMD -f "$COMPOSE_FILE" pull
    fi
    
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
    echo "Integrated Cats API + Reverse Proxy Deployment Script with Strategy Demo"
    echo ""
    echo "Usage: $0 [OPTIONS] [COMMAND] [ARGUMENTS]"
    echo ""
    echo "Options:"
    echo "  --registry         Use registry-based deployment (pull from ghcr.io)"
    echo "  --local           Use local build deployment (default)"
    echo ""
    echo "Commands:"
    echo "  deploy [replicas]  Deploy the integrated stack (default: $DEFAULT_REPLICAS replicas)"
    echo "  demo              Interactive load balancing strategy demonstration"
    echo "  demo <strategy>   Demo specific strategy (roundrobin, random, weighted, leastconnections, iphash)"
    echo "  status            Show deployment status and health"
    echo "  scale <replicas>  Scale the API service to specified replicas"
    echo "  test              Test load balancing functionality"
    echo "  logs              Show service logs"
    echo "  stop              Stop all services"
    echo "  help              Show this help message"
    echo ""
    echo "Load Balancing Strategies:"
    echo "  roundrobin        Equal distribution (default)"
    echo "  random           Prevents cache hotspots"
    echo "  weighted         Different backend capacities"
    echo "  leastconnections Varying request times"
    echo "  iphash           Session affinity"
    echo ""
    echo "Examples:"
    echo "  $0 deploy 3                    Deploy with 3 API replicas using local builds"
    echo "  $0 --registry deploy 5         Deploy with 5 replicas using registry images"
    echo "  $0 demo                        Interactive strategy demonstration"
    echo "  $0 demo leastconnections       Demo least connections strategy"
    echo "  LB_STRATEGY=iphash $0 deploy 5 Deploy with IP hash strategy"
    echo "  $0 scale 10                    Scale to 10 API replicas"
    echo "  $0 test                        Test load balancing"
    echo "  $0 status                      Show current status"
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --registry)
                USE_REGISTRY=true
                COMPOSE_FILE="$REGISTRY_COMPOSE_FILE"
                log_info "Using registry-based deployment"
                shift
                ;;
            --local)
                USE_REGISTRY=false
                COMPOSE_FILE="docker-compose.prod.yml"
                log_info "Using local build deployment"
                shift
                ;;
            *)
                # Pass remaining arguments to main function
                break
                ;;
        esac
    done
}

# Main script logic
main() {
    # Parse arguments first
    parse_arguments "$@"
    
    # Shift past parsed options to get the command
    while [[ $# -gt 0 ]]; do
        case $1 in
            --registry|--local)
                shift
                ;;
            *)
                break
                ;;
        esac
    done
    
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            deploy_stack "$2"
            show_status
            ;;
        "demo")
            check_prerequisites
            if [ -n "$2" ]; then
                demo_strategy "$2"
            else
                interactive_strategy_demo
            fi
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
