# Production Deployment Guide

This guide covers deploying the integrated Cats API + Reverse Proxy application in production environments.

## Deployment Modes

The deployment script supports two modes:

### 1. Local Build Mode (Default)

Uses locally built Docker images from the workspace.

```bash
# Deploy with local builds (default)
./deploy.sh deploy 5
./deploy.sh --local deploy 5
```

**Best for:** Development, testing, and environments where you want to build from source.

### 2. Registry Mode

Pulls pre-built images from GitHub Container Registry (ghcr.io).

```bash
# Deploy with registry images
./deploy.sh --registry deploy 5
```

**Best for:** Production deployments using CI/CD built and verified images.

## Registry Authentication

For registry mode, ensure you're authenticated with GitHub Container Registry:

```bash
# Set your GitHub credentials
export GITHUB_USERNAME="your-username"
export GITHUB_TOKEN="your-personal-access-token"

# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin
```

## Registry Images

The application expects the following images in registry mode:

- `ghcr.io/st4r4x/golangapp/cats-api:latest`
- `ghcr.io/st4r4x/golangapp/reverse-proxy:latest`

These images are automatically built and pushed by the CI/CD pipeline when code is pushed to the main branch.

## Docker Compose Files

- `docker-compose.prod.yml` - Local build mode configuration
- `docker-compose.registry.yml` - Registry mode configuration

Both configurations include:

- Health checks for all services
- Resource limits
- Automatic restart policies
- Production-ready networking

## Deployment Examples

```bash
# Local development deployment
./deploy.sh deploy 3

# Production deployment with registry images
./deploy.sh --registry deploy 10

# Scale existing deployment
./deploy.sh scale 15

# Check deployment status
./deploy.sh status

# Test load balancing
./deploy.sh test
```

## Production Considerations

1. **Resource Limits**: Configured in Docker Compose files

   - Cats API: 256MB memory limit per replica
   - Reverse Proxy: 128MB memory limit

2. **Health Checks**: Automatic health monitoring with restart on failure

3. **Load Balancing**: Custom Go reverse proxy with automatic backend discovery

4. **Scaling**: Easy horizontal scaling with automatic load balancer reload

5. **Monitoring**: Built-in resource usage and status reporting

## Troubleshooting

### Registry Authentication Issues

```bash
# Verify authentication
docker info | grep -i registry

# Re-authenticate if needed
docker login ghcr.io
```

### Image Pull Issues

```bash
# Check if images exist
docker manifest inspect ghcr.io/st4r4x/golangapp/cats-api:latest
docker manifest inspect ghcr.io/st4r4x/golangapp/reverse-proxy:latest

# Fallback to local build
./deploy.sh --local deploy 5
```

### Service Health Issues

```bash
# Check detailed status
./deploy.sh status

# View logs
./deploy.sh logs

# Test specific components
curl http://localhost:4443/health
```
