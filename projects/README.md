# Multi-Service Docker Setup

This is a multi-service application with separate containers for the API and reverse proxy.

## Project Structure

```text
projects/
├── cats-api/          # Go Cats API service
│   ├── *.go          # Go source files
│   ├── go.mod        # Go module
│   ├── test/         # Test files
│   ├── docs/         # Documentation
│   └── swagger-ui/   # Swagger UI assets
├── reverse-proxy/     # Custom reverse proxy service
│   ├── main.go       # Go reverse proxy
│   ├── go.mod        # Go module
│   └── certs/        # SSL certificates
└── examples/          # Configuration examples
```

## Services

### Cats API

- **Port:** 8080 (internal)
- **Technology:** Go 1.23
- **Features:** REST API, Swagger UI, Hot reload

### Reverse Proxy

- **Port:** 4443 (external)
- **Technology:** Custom Go proxy or Nginx
- **Features:** Load balancing, SSL termination

## Quick Start

```bash
# Start all services
make docker-multi-up

# Scale API replicas
make docker-multi-scale REPLICAS=5

# Test load balancing
make docker-multi-load-test
```

## Access Points

- **Load Balancer:** `http://localhost:8080`
- **Direct API:** Internal network only
- **Swagger UI:** `http://localhost:8080/swagger/`

## Development

Each project can be developed independently:

```bash
# Cats API development
cd projects/cats-api
make dev

# Reverse Proxy development
cd projects/reverse-proxy
go run .
```
