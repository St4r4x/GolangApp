# 🐈 Go Cats API

![Coverage](https://img.shields.io/badge/Coverage-64.6%25-green)
![CI/CD](https://img.shields.io/badge/CI%2FCD-Passing-brightgreen)
![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![Docker](https://img.shields.io/badge/Docker-Multi--Instance-blue)
![Load Balancing](https://img.shields.io/badge/Load%20Balancing-Ready-orange)

A production-ready REST API for managing cats 🐈 with full CRUD operations, enterprise-grade CI/CD pipeline, and multi-instance load balancing support.

## ✨ Features

- 🔄 **Load Balanced Multi-Instance Setup**
- 🔍 **Enhanced Request Monitoring** with server identification
- 🚀 **Enterprise CI/CD Pipeline** with automated versioning
- 📊 **Comprehensive Test Coverage** (64.6%)
- 🐳 **Docker Multi-Stage Builds** optimized for production
- 📝 **Interactive Swagger UI** documentation
- 🔐 **Security Scanning** with Trivy
- 📈 **Hot Reload Development** with Air

## 🚀 Quick Start

### Single Instance

```bash
git clone <repository-url>
cd GolangApp
go run .
```

### Multi-Instance Load Balanced Setup

```bash
# Start both instances (ports 8081, 8082)
make docker-multi-up

# Test load balancing
make docker-multi-load-test

# Monitor requests in real-time
make docker-multi-monitor
```

## 🌐 Access Points

- **Home page:** http://localhost:8080
- **Swagger UI:** http://localhost:8080/swagger/
- **API endpoints:** http://localhost:8080/api/cats
- **Instance 1:** http://localhost:8081 (when using multi-instance)
- **Instance 2:** http://localhost:8082 (when using multi-instance)

## 📁 Project Structure

```text
├── .github/workflows/     # CI/CD pipeline with 9 parallel stages
├── docs/                  # Comprehensive documentation
│   ├── CICD-DEEP-DIVE.md # Complete CI/CD explanation
│   ├── TESTING.md        # Testing strategy guide
│   └── *.md             # Additional documentation
├── test/                  # Organized test suite
│   ├── unit/             # Unit tests with mocks
│   ├── integration/      # Integration tests
│   ├── apitests/         # API endpoint tests (build tags)
│   └── mocked/           # Mocked component tests
├── examples/             # Configuration examples
│   ├── nginx-reverse-proxy.conf
│   ├── haproxy.cfg
│   └── docker-compose.traefik.yml
├── swagger-ui/           # Swagger UI assets
├── scripts/              # Development scripts
├── Dockerfile            # Multi-stage production build
├── docker-compose.yml    # Development environment
├── docker-compose.multi.yml # Multi-instance setup
├── .air.toml            # Hot reload configuration
├── Makefile             # 25+ development commands
└── *.go                 # Go source files
```

## 🔧 Development Commands

### Core Development

```bash
make dev-setup          # Setup development environment
make dev               # Start with hot reload (Air)
make run               # Standard run
make build             # Build application
```

### Testing & Quality

```bash
make test              # Run all tests
make coverage          # Generate coverage report
make lint              # Run linting checks
make security          # Security scanning
make ci-local          # Full CI pipeline locally
```

### Docker Operations

```bash
make docker-build      # Build production image
make docker-run        # Run single container
make compose-up        # Development environment
```

### Multi-Instance Management

```bash
make docker-multi-up           # Start instances (8081, 8082)
make docker-multi-down         # Stop all instances
make docker-multi-test         # Test both instances
make docker-multi-load-test    # Load balancing test
make docker-multi-monitor      # Real-time monitoring
make docker-multi-logs         # View combined logs
```

## 🏗️ Architecture

### Load Balancing Setup

```
Reverse Proxy (Port 4443) → Round Robin
    ├── Container 1 (Port 8081) → Internal 8080
    └── Container 2 (Port 8082) → Internal 8080
```

### Enhanced Monitoring

- **Server Identification:** Each request logged with container ID
- **Response Headers:** `X-Server-Id`, `X-Container-Name`, `X-Server-Port`
- **Load Distribution:** Visual confirmation of round-robin balancing
- **Real-time Logs:** Live monitoring with server identification

### Example Log Output

```
2025-09-11 13:00:52.728 dev I app.go:29 🌐 [Server: 066c8b84a819:8080] New request to: 'GET /' from 172.18.0.1:38036
2025-09-11 13:00:53.245 dev I app.go:29 🌐 [Server: ee20f5b8b7e1:8080] New request to: 'GET /' from 172.18.0.1:56862
```

## 🧪 Testing & Coverage

**64.6% test coverage** across multiple testing strategies:

- **Unit Tests:** Component isolation with mocks
- **Integration Tests:** Real function testing
- **API Tests:** End-to-end HTTP testing with build tags
- **Mocked Tests:** Dependency injection testing

### Test Commands

```bash
make test-unit         # Unit tests only
make test-integration  # Integration tests only
make test-api          # API tests (requires server)
make test-mocked       # Mocked tests only
make coverage          # Generate HTML coverage report
```

### Coverage Reports

- **HTML Report:** `docs/coverage.html`
- **Console Output:** Real-time coverage percentages
- **CI Integration:** Automated coverage tracking

## 🚀 CI/CD Pipeline

**9-stage parallel pipeline** with enterprise features:

1. **Code Quality** - Linting, formatting, staticcheck
2. **Unit Testing** - Isolated component tests
3. **Integration Testing** - Real function verification
4. **Coverage Analysis** - Comprehensive reporting
5. **Versioning** - Automated version management
6. **Docker Build** - Multi-platform images
7. **API Testing** - Live endpoint validation
8. **Security Scanning** - Trivy vulnerability analysis
9. **Version Update** - Auto-update deploy-dev branch

### Pipeline Features

- **Parallel Execution:** Optimized for speed
- **Security Integration:** GHCR + Trivy scanning
- **Auto-versioning:** Timestamp + commit hash
- **Multi-platform:** AMD64 + ARM64 support
- **Branch Protection:** Deploy-dev version tracking

## 🐳 Docker

### Multi-Stage Production Build

```dockerfile
# Builder stage with Go 1.23
FROM golang:1.23-alpine AS builder

# Runtime stage from scratch (~10MB)
FROM scratch AS runtime
```

### Container Features

- **Optimized Size:** ~10MB final image
- **Security:** Non-root user, minimal attack surface
- **Health Checks:** Built-in endpoint monitoring
- **Multi-platform:** AMD64 and ARM64 support

## 🔄 Load Balancing Integration

### Supported Reverse Proxies

- **Nginx** - Configuration in `examples/nginx-reverse-proxy.conf`
- **HAProxy** - Configuration in `examples/haproxy.cfg`
- **Traefik** - Docker Compose in `examples/docker-compose.traefik.yml`
- **Custom Go Proxy** - Round-robin implementation ready

### Monitoring Load Distribution

```bash
# Test load balancing
for i in {1..6}; do
  port=$((8080 + (i % 2) + 1))
  echo "Request $i → Port $port"
  curl -s -I http://localhost:$port/ | grep "X-Server-Id"
done
```

## 📖 Documentation

Comprehensive documentation available:

- **[CI/CD Deep Dive](docs/CICD-DEEP-DIVE.md)** - Complete pipeline explanation
- **[Testing Guide](docs/TESTING.md)** - Testing strategies and best practices
- **[Project Status](docs/PROJECT-STATUS.md)** - Current implementation status
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

## 🔧 Development Tools

### Hot Reload with Air

```bash
make dev               # Start with hot reload
# Configuration in .air.toml
```

### Comprehensive Makefile

25+ commands for all development needs:

```bash
make help              # Show all available commands
make pre-commit        # Pre-commit validation
make clean             # Clean all artifacts
make version           # Show version info
```

## 🎯 Production Ready

✅ **Multi-instance deployment**  
✅ **Load balancing support**  
✅ **Security scanning**  
✅ **Automated CI/CD**  
✅ **Comprehensive monitoring**  
✅ **Docker optimization**  
✅ **Version management**

## 📝 API Documentation

Interactive Swagger UI available at `/swagger/` with complete API specification.

### Regenerate OpenAPI

```bash
# Convert YAML to JSON for Swagger UI
go run . -convert-openapi
```

---

**Built with ❤️ using Go 1.23, Docker, and enterprise CI/CD practices.**
