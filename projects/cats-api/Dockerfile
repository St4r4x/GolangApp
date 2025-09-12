# Build stage
FROM golang:1.23-alpine AS builder

# Install dependencies and security updates
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    && apk upgrade --no-cache

# Create a non-root user for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy dependency files first for better Docker layer caching
COPY go.mod ./
COPY go.sum* ./

# Download and verify dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
ARG VERSION=dev
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -a -installsuffix cgo \
    -o backend .

# Run tests during build to catch issues early
RUN go test -v ./...

# Runtime stage - minimal image
FROM scratch AS runtime

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy user information for non-root execution
COPY --from=builder /etc/passwd /etc/passwd

# Copy the compiled binary
COPY --from=builder /build/backend /backend

# Copy static assets
COPY --from=builder /build/swagger-ui /swagger-ui
COPY --from=builder /build/openapi.yml /openapi.yml

# Use non-root user for security
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/backend", "-health-check"]

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/backend"]
