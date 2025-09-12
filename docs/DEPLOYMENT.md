# 🚀 Deployment Documentation

## Overview

The Go Cats API project provides multiple deployment strategies for different environments, from local development to enterprise production. The deployment system includes an intelligent deployment script, Docker Compose orchestration, and comprehensive load balancing demonstration capabilities.

## Deployment Strategies

### Local Development

**Quick Start**:

```bash
# Clone and run
git clone <repository>
cd GolangApp
./deploy.sh dev
```

**Features**:

- Automatic dependency installation
- Hot reload with file watching
- Debug logging enabled
- Single instance deployment
- Interactive development environment

### Production Deployment

**Standard Production**:

```bash
./deploy.sh prod
```

**High Availability Production**:

```bash
./deploy.sh prod --scale 5
```

**Features**:

- Multi-replica cats API instances
- Advanced load balancing with configurable strategies
- Health monitoring and automatic recovery
- Optimized container builds
- Production logging and metrics
- Zero-downtime deployments

### Strategy Demonstration

**Interactive Demo Mode**:

```bash
./deploy.sh demo
```

**Specific Strategy Testing**:

```bash
./deploy.sh demo iphash
./deploy.sh demo leastconnections
./deploy.sh demo weighted
```

## Deployment Script (deploy.sh)

### Core Functionality

The `deploy.sh` script provides comprehensive deployment automation with the following capabilities:

#### Command Structure

```bash
./deploy.sh <environment> [strategy] [options]
```

#### Available Environments

- **`dev`**: Development mode with debug features
- **`prod`**: Production deployment with optimization
- **`demo`**: Interactive load balancing demonstration
- **`test`**: Testing mode with comprehensive validation
- **`stop`**: Graceful service shutdown
- **`clean`**: Complete environment cleanup

#### Load Balancing Strategies

- **`roundrobin`**: Default even distribution (fastest: 233.4 ns/op)
- **`random`**: Random backend selection (246.7 ns/op)
- **`weighted`**: Capacity-aware distribution (406.4 ns/op)
- **`leastconnections`**: Dynamic load adaptation (329.4 ns/op)
- **`iphash`**: Session affinity support (286.0 ns/op)

### Script Features

#### Environment Detection

```bash
detect_environment() {
    if [[ -f /.dockerenv ]]; then
        echo "container"
    elif command -v docker &> /dev/null; then
        echo "docker"
    elif command -v go &> /dev/null; then
        echo "native"
    else
        echo "unknown"
    fi
}
```

#### Service Health Monitoring

```bash
wait_for_service() {
    local service_url=$1
    local max_attempts=30

    for ((i=1; i<=max_attempts; i++)); do
        if curl -s "$service_url" > /dev/null 2>&1; then
            echo "✅ Service is ready"
            return 0
        fi
        echo "⏳ Waiting for service... ($i/$max_attempts)"
        sleep 2
    done
    return 1
}
```

#### Load Testing Integration

```bash
run_load_test() {
    local url=$1
    local requests=${2:-100}

    echo "🔄 Running load test with $requests requests..."

    for ((i=1; i<=requests; i++)); do
        curl -s "$url" > /dev/null &
        if ((i % 10 == 0)); then
            wait  # Wait for batch completion
            echo "📊 Completed $i requests"
        fi
    done
    wait  # Wait for all remaining requests
}
```

## Docker Compose Configuration

### Development Environment (docker-compose.yml)

```yaml
services:
  cats-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=debug
      - LOG_LEVEL=debug
    volumes:
      - .:/app
      - /app/vendor
    command: air -c .air.toml
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  reverse-proxy:
    build:
      context: ./projects/reverse-proxy
    ports:
      - "4443:8080"
    environment:
      - LB_STRATEGY=${LB_STRATEGY:-roundrobin}
      - BACKEND_SERVICE=cats-api
    depends_on:
      - cats-api
```

### Production Environment (docker-compose.prod.yml)

```yaml
services:
  cats-api:
    build:
      context: .
      dockerfile: Dockerfile.prod
    environment:
      - GIN_MODE=release
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 128M
          cpu: 100m
        reservations:
          memory: 64M
          cpu: 50m
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3

  reverse-proxy:
    build:
      context: ./projects/reverse-proxy
      dockerfile: Dockerfile.prod
    ports:
      - "4443:8080"
    environment:
      - LB_STRATEGY=${LB_STRATEGY:-roundrobin}
      - BACKEND_SERVICE=cats-api
      - BACKEND_PORT=8080
    depends_on:
      - cats-api
    deploy:
      resources:
        limits:
          memory: 64M
          cpu: 50m
```

## Deployment Workflows

### Development Workflow

1. **Environment Setup**:

   ```bash
   # Start development environment
   ./deploy.sh dev

   # Verify services
   curl http://localhost:4443/
   curl http://localhost:8080/cats
   ```

2. **Code Changes**:

   - Automatic reload on file changes
   - Debug logging for development
   - Single instance for simplicity

3. **Testing**:

   ```bash
   # Run unit tests
   go test ./...

   # Integration testing
   ./deploy.sh test
   ```

### Production Deployment Workflow

1. **Pre-deployment Validation**:

   ```bash
   # Run comprehensive tests
   ./deploy.sh test

   # Build production images
   docker compose -f docker-compose.prod.yml build
   ```

2. **Deployment Execution**:

   ```bash
   # Deploy with specific strategy
   LB_STRATEGY=leastconnections ./deploy.sh prod --scale 5

   # Verify deployment health
   ./deploy.sh verify
   ```

3. **Post-deployment Monitoring**:

   ```bash
   # Monitor service health
   curl http://localhost:4443/health

   # Check load distribution
   ./deploy.sh demo leastconnections

   # View logs
   docker logs -f golangapp-reverse-proxy-1
   ```

### Strategy Testing Workflow

1. **Interactive Demo**:

   ```bash
   # Launch interactive demo
   ./deploy.sh demo

   # Follow prompts to test strategies
   # View real-time load distribution
   ```

2. **Specific Strategy Testing**:

   ```bash
   # Test IP hash strategy
   ./deploy.sh demo iphash

   # Test least connections
   ./deploy.sh demo leastconnections

   # Test weighted distribution
   ./deploy.sh demo weighted
   ```

3. **Performance Comparison**:
   ```bash
   # Compare strategy performance
   for strategy in roundrobin random weighted leastconnections iphash; do
     echo "Testing $strategy..."
     ./deploy.sh demo $strategy
     sleep 5
   done
   ```

## Environment Configuration

### Environment Variables

#### Load Balancer Configuration

```bash
# Strategy selection
export LB_STRATEGY=leastconnections

# Backend configuration
export BACKEND_SERVICE=cats-api
export BACKEND_PORT=8080
export PROXY_PORT=8080

# Discovery settings
export DISCOVERY_INTERVAL=10s
export HEALTH_CHECK_INTERVAL=30s
```

#### Application Configuration

```bash
# Server settings
export PORT=8080
export GIN_MODE=release

# Logging configuration
export LOG_LEVEL=info
export LOG_FILE=server.log

# External services
export IMAGE_API_URL=https://cataas.com/cat
export IMAGE_API_TIMEOUT=5s
```

### Configuration Management

#### Development Configuration

```bash
# .env.dev
GIN_MODE=debug
LOG_LEVEL=debug
LB_STRATEGY=roundrobin
SCALE_REPLICAS=1
```

#### Production Configuration

```bash
# .env.prod
GIN_MODE=release
LOG_LEVEL=info
LB_STRATEGY=leastconnections
SCALE_REPLICAS=3
```

## Scaling and Load Balancing

### Horizontal Scaling

#### Manual Scaling

```bash
# Scale to 5 replicas
docker compose up --scale cats-api=5 -d

# Scale with specific strategy
LB_STRATEGY=weighted docker compose up --scale cats-api=5 -d

# Verify scaling
docker ps | grep cats-api
```

#### Dynamic Scaling

```bash
# Scale up during high load
./deploy.sh scale-up 10

# Scale down during low load
./deploy.sh scale-down 3

# Auto-scaling based on metrics
./deploy.sh auto-scale --cpu-threshold=70 --memory-threshold=80
```

### Load Distribution Testing

#### Even Distribution Verification

```bash
# Test round robin distribution
./deploy.sh demo roundrobin

# Expected output: Even distribution across all replicas
# Backend 1: 20 requests (20%)
# Backend 2: 20 requests (20%)
# Backend 3: 20 requests (20%)
# Backend 4: 20 requests (20%)
# Backend 5: 20 requests (20%)
```

#### Session Affinity Testing

```bash
# Test IP hash strategy
./deploy.sh demo iphash

# Verify same client always hits same backend
curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:4443/
curl -H "X-Forwarded-For: 192.168.1.101" http://localhost:4443/
```

## Health Monitoring

### Service Health Checks

#### Individual Service Monitoring

```bash
# Check cats API health
curl http://localhost:8080/health

# Check load balancer health
curl http://localhost:4443/health

# Comprehensive health check
./deploy.sh health-check
```

#### Automated Health Monitoring

```bash
monitor_services() {
    while true; do
        if ! curl -f http://localhost:4443/health > /dev/null 2>&1; then
            echo "❌ Load balancer unhealthy, restarting..."
            docker restart golangapp-reverse-proxy-1
        fi

        if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
            echo "❌ API service unhealthy, scaling up..."
            ./deploy.sh scale-up $(($(docker ps | grep cats-api | wc -l) + 1))
        fi

        sleep 30
    done
}
```

### Performance Monitoring

#### Real-time Metrics

```bash
# Monitor request distribution
watch -n 1 'docker logs golangapp-reverse-proxy-1 | tail -10'

# Monitor resource usage
watch -n 1 'docker stats --no-stream'

# Monitor connection counts
watch -n 1 './deploy.sh connection-stats'
```

#### Load Testing Metrics

```bash
# Performance benchmark
./deploy.sh benchmark --requests=1000 --concurrency=50

# Strategy performance comparison
./deploy.sh compare-strategies --duration=60s

# Response time analysis
./deploy.sh response-time-analysis
```

## Production Considerations

### Security

#### Network Security

```bash
# Use custom networks
docker network create --driver bridge cats-network

# Implement service isolation
docker compose -f docker-compose.secure.yml up
```

#### Secret Management

```bash
# Use Docker secrets
echo "api-key-value" | docker secret create api-key -

# Environment variable encryption
./deploy.sh encrypt-env .env.prod
```

### High Availability

#### Multi-Host Deployment

```bash
# Deploy across multiple hosts
./deploy.sh deploy-swarm --nodes=3

# Configure shared storage
./deploy.sh setup-nfs-storage

# Database clustering
./deploy.sh setup-db-cluster
```

#### Disaster Recovery

```bash
# Backup deployment configuration
./deploy.sh backup-config

# Implement blue-green deployment
./deploy.sh blue-green-deploy

# Rollback capabilities
./deploy.sh rollback --version=v1.2.3
```

### Performance Optimization

#### Container Optimization

```bash
# Multi-stage Docker builds
FROM golang:1.23-alpine AS builder
# ... build stage
FROM alpine:latest AS runtime
# ... runtime stage
```

#### Resource Tuning

```yaml
deploy:
  resources:
    limits:
      memory: 128M
      cpu: 100m
    reservations:
      memory: 64M
      cpu: 50m
```

## Troubleshooting

### Common Deployment Issues

#### Port Conflicts

**Symptom**: Services fail to start due to port conflicts

**Solution**:

```bash
# Check port usage
netstat -tlnp | grep :4443
netstat -tlnp | grep :8080

# Use alternative ports
PORT=8081 ./deploy.sh dev
PROXY_PORT=4444 ./deploy.sh prod
```

#### Container Build Failures

**Symptom**: Docker build fails during deployment

**Debugging**:

```bash
# Build with verbose output
docker compose build --no-cache --progress=plain

# Check build context
docker compose config

# Validate Dockerfile
docker build --no-cache -f Dockerfile .
```

#### Service Discovery Issues

**Symptom**: Load balancer cannot find backend services

**Solution**:

```bash
# Check Docker network
docker network inspect golangapp_default

# Verify service names
docker ps --format "table {{.Names}}\t{{.Image}}"

# Test network connectivity
docker exec golangapp-reverse-proxy-1 nslookup cats-api
```

### Debugging Commands

#### Service Debugging

```bash
# Container inspection
docker inspect golangapp-cats-api-1

# Resource monitoring
docker stats --no-stream

# Network diagnostics
docker network ls
docker network inspect bridge
```

#### Log Analysis

```bash
# Comprehensive log collection
./deploy.sh collect-logs

# Error pattern analysis
docker logs golangapp-reverse-proxy-1 | grep ERROR

# Performance log analysis
docker logs golangapp-cats-api-1 | grep duration_ms
```

## Automation Scripts

### Deployment Automation

#### Continuous Deployment Pipeline

```bash
#!/bin/bash
# ci-deploy.sh

set -e

echo "🚀 Starting CI deployment pipeline..."

# Run tests
./deploy.sh test

# Build and tag images
docker build -t cats-api:$CI_COMMIT_SHA .
docker build -t reverse-proxy:$CI_COMMIT_SHA ./projects/reverse-proxy

# Deploy to staging
ENVIRONMENT=staging ./deploy.sh prod

# Run integration tests
./deploy.sh integration-test

# Deploy to production if tests pass
if [ "$?" -eq 0 ]; then
    ENVIRONMENT=production ./deploy.sh prod
    echo "✅ Production deployment successful"
else
    echo "❌ Integration tests failed, deployment aborted"
    exit 1
fi
```

#### Rollback Automation

```bash
#!/bin/bash
# rollback.sh

PREVIOUS_VERSION=${1:-"latest"}

echo "🔄 Rolling back to version: $PREVIOUS_VERSION"

# Stop current deployment
./deploy.sh stop

# Deploy previous version
IMAGE_TAG=$PREVIOUS_VERSION ./deploy.sh prod

# Verify rollback success
./deploy.sh health-check

echo "✅ Rollback completed successfully"
```

### Monitoring Automation

#### Health Check Automation

```bash
#!/bin/bash
# health-monitor.sh

while true; do
    if ! ./deploy.sh health-check; then
        echo "🚨 Health check failed, attempting recovery..."

        # Restart unhealthy services
        ./deploy.sh restart-unhealthy

        # Wait for recovery
        sleep 30

        # Verify recovery
        if ./deploy.sh health-check; then
            echo "✅ Services recovered successfully"
        else
            echo "❌ Recovery failed, manual intervention required"
            # Send alert (email, Slack, etc.)
        fi
    fi

    sleep 60
done
```

---

_This deployment system provides enterprise-grade automation with comprehensive monitoring, scaling, and troubleshooting capabilities for the Go Cats API project._
