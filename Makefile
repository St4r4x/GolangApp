.PHONY: test coverage update-coverage clean

# Run all tests
test:
	go test ./test/...

# Run tests with coverage report
coverage:
	go test -coverprofile=docs/coverage.out ./... -coverpkg=./...
	go tool cover -html=docs/coverage.out -o docs/coverage.html
	@echo "Coverage report generated: docs/coverage.html"
	go tool cover -func=docs/coverage.out | grep "total:"

# Clean up generated files
clean:
	rm -f docs/coverage.out docs/coverage.html
	rm -rf logs/*.log

# Build the application
build:
	go build -o bin/catapi .

# Run the application
run:
	go run .

# Run with Docker
docker-build:
	docker build -t catapi .

docker-run:
	docker run -p 8080:8080 catapi
