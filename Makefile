.PHONY: test coverage update-coverage clean

# Run all tests
test:
	go test ./test/...

# Run tests with coverage report
coverage:
	go test -coverprofile=coverage.out ./... -coverpkg=./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out | grep "total:"

# Update coverage badge in README
update-coverage:
	./update-coverage.sh

# Clean up generated files
clean:
	rm -f coverage.out coverage.html

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
