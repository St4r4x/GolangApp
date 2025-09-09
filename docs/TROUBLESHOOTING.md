# Go Cats API - Troubleshooting Guide

This comprehensive guide covers common issues and their solutions for the Go Cats API CI/CD pipeline.

## 🚨 Common CI/CD Issues

### 1. Static Analysis Errors

**Issue**: `staticcheck` fails with deprecated imports or code quality issues

```bash
SA1019: Package io/ioutil is deprecated
SA4006: unused variable
U1000: unused function
```

**Solution**:

- Replace `io/ioutil` with `os` package functions
- Use underscore assignment for intentionally unused variables: `_, err := http.Get(...)`
- Remove unused functions and variables

### 2. Formatting Errors

**Issue**: `gofmt` check fails

```bash
Code is not formatted. Please run 'go fmt ./...'
test/apitests/all_test.go
```

**Solution**:

```bash
go fmt ./...
```

**Note**: Use modern build tags `//go:build integration` instead of `// +build integration`

### 3. Coverage Analysis Connection Issues

**Issue**: API tests fail during coverage with "connection refused"

```bash
dial tcp [::1]:8080: connect: connection refused
coverage: 10.3% of statements
```

**Solution**: API tests are isolated with build tags and run separately in CI/CD

- Local coverage excludes API tests: `make coverage`
- CI/CD runs API tests with live Docker service
- Coverage maintains ~65% without connection issues

### 4. Docker Repository Name Issues

**Issue**: Docker build fails with uppercase repository names

```bash
ERROR: repository name must be lowercase
```

**Solution**: Repository names are automatically converted to lowercase in CI/CD workflow

### 5. Docker Service Health Check Failures

**Issue**: GitHub Actions services fail health checks

```bash
Service container backend failed.
backend service is starting, waiting 32 seconds...
unhealthy
```

**Solution**:

- Removed problematic health check from scratch containers
- Enhanced API readiness detection with 3-minute timeout
- Added comprehensive logging and debugging

### 6. Security Scan Upload Issues

**Issue**: SARIF upload fails due to permissions

```bash
Warning: Resource not accessible by integration
Error: Resource not accessible by integration
```

**Solution**:

- Added proper workflow permissions for `security-events: write`
- Implemented fallback artifact upload
- Made security scan non-blocking with `continue-on-error`

## 🔧 Quick Fixes

### Format and Lint

```bash
# Fix formatting
go fmt ./...

# Check static analysis
staticcheck ./...

# Verify formatting
gofmt -s -l .
```

### Coverage Analysis

```bash
# Local coverage (excludes API tests)
make coverage

# Full coverage (requires running server)
make coverage-all
```

### Docker Testing

```bash
# Build production image
docker build --target runtime -t cats-api:test .

# Test container
docker run -d -p 8080:8080 cats-api:test
curl http://localhost:8080/
```

### Development Setup

```bash
# Install development tools
make dev-setup

# Run all tests
make test

# Start development server with hot reload
make dev
```

## 📋 Pipeline Status Checklist

- ✅ **Formatting**: `gofmt -s -l .` returns empty
- ✅ **Static Analysis**: `staticcheck ./...` passes
- ✅ **Tests**: All unit/integration tests pass
- ✅ **Coverage**: ~65% coverage maintained
- ✅ **Build**: Docker builds successfully
- ✅ **Security**: Trivy scans complete (with or without SARIF upload)

## 🎯 Best Practices

1. **Code Quality**: Run `make lint` before committing
2. **Testing**: Use appropriate test types (unit/integration/API)
3. **Docker**: Keep images minimal and secure
4. **CI/CD**: Monitor pipeline logs for early issue detection
5. **Documentation**: Keep this guide updated with new issues

## 🆘 Need Help?

If you encounter issues not covered here:

1. Check GitHub Actions logs for detailed error messages
2. Run local validation: `make ci-local`
3. Test Docker builds locally before pushing
4. Review recent changes that might have introduced issues

The CI/CD pipeline has comprehensive error handling and should provide clear feedback for most issues.
