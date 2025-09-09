# CI/CD Pipeline Troubleshooting Guide

## 🐛 Fixed Issues

### 1. Go Version Mismatch ✅

**Problem:** Pipeline was using Go 1.25 (doesn't exist) while go.mod uses 1.23
**Fix:** Updated pipeline to use Go 1.23

### 2. Health Check Command ✅

**Problem:** Using `curl` in scratch container which doesn't have curl
**Fix:** Removed health check from service definition, rely on wait logic instead

### 3. Missing Test Directories ✅

**Problem:** Pipeline fails if test directories don't exist
**Fix:** Added directory checks before running tests

### 4. Docker Build Target ✅

**Problem:** Not specifying which Docker stage to build
**Fix:** Added `target: runtime` to build step

### 5. Artifact Pattern ✅

**Problem:** Incorrect pattern for downloading coverage artifacts
**Fix:** Updated pattern to `*-coverage*`

## 🔍 Common CI/CD Issues & Solutions

### Issue: "Go version not found"

```yaml
# Make sure GO_VERSION matches your go.mod
env:
  GO_VERSION: "1.23" # ✅ Correct
  # GO_VERSION: "1.25"  # ❌ Doesn't exist
```

### Issue: "No tests found" or test failures

```bash
# Check if test directories exist
ls -la test/
# Run tests locally first
go test ./...
```

### Issue: Docker build failures

```bash
# Test Docker build locally
docker build --target runtime -t test-image .
# Check if all required files exist
ls -la go.mod go.sum
```

### Issue: Service health check failures

```yaml
# For scratch containers, avoid curl-based health checks
services:
  backend:
    image: myimage
    ports:
      - 8080:8080
    # ❌ Don't use curl in scratch containers
    # options: --health-cmd "curl -f http://localhost:8080/"
```

### Issue: Authentication failures for private registries

```yaml
# Make sure GITHUB_TOKEN has package write permissions
- name: Login to Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }} # ✅ Should work for GHCR
```

## 🧪 Testing Pipeline Locally

### 1. Test Basic Commands

```bash
# Verify Go setup
go version
go mod verify
go mod download

# Test linting
go vet ./...
go fmt ./...

# Test builds
go build .
```

### 2. Test Docker Build

```bash
# Test runtime build
docker build --target runtime -t cats-api .

# Test development build
docker build --target development -t cats-api-dev .

# Test container run
docker run --rm -d -p 8080:8080 cats-api
```

### 3. Test All Components

```bash
# Use the comprehensive test script
./scripts/test-all.sh
```

## 📋 Debugging GitHub Actions

### Check Workflow Syntax

```bash
# Install GitHub CLI
gh workflow list
gh workflow view
```

### Common Log Locations

- **Setup Go**: Check if Go version is available
- **Install dependencies**: Look for module download errors
- **Build and push**: Check Docker build logs
- **API tests**: Check if service started correctly

### Environment Variables

```yaml
# Add debug output to see environment
- name: Debug environment
  run: |
    echo "Go version: $(go version)"
    echo "Working directory: $(pwd)"
    echo "Files: $(ls -la)"
```

## 🚨 Emergency Fixes

### Skip failing jobs temporarily

```yaml
# Add to job that's failing
if: false # Temporarily disable job
```

### Simplify workflow for debugging

```yaml
# Minimal workflow for testing
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go test ./...
```

## 📞 Getting Help

1. **Check GitHub Actions logs** - Click on failed job for details
2. **Test locally first** - Reproduce issue on your machine
3. **Check action versions** - Ensure you're using compatible versions
4. **Verify permissions** - GITHUB_TOKEN needs appropriate scopes

## ✅ Validation Checklist

- [ ] Go version matches between workflow and go.mod
- [ ] All test directories exist or are handled gracefully
- [ ] Docker build works locally
- [ ] No curl commands in scratch containers
- [ ] GITHUB_TOKEN has package permissions
- [ ] All artifact patterns are correct
- [ ] Service dependencies are properly configured
