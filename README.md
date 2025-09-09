# Go Cats API

![Coverage](https://img.shields.io/badge/Coverage-86.6%25-brightgreen)

A comprehensive REST API for managing cats 🐈 with full CRUD operations, built with Go.

## Quick Start

```bash
# Clone and run
git clone <repository-url>
cd GolangApp
go run .
```

Browse the application:

- **Home page:** http://localhost:8080
- **Swagger UI:** http://localhost:8080/swagger/
- **API endpoints:** http://localhost:8080/api/cats

## Project Structure

```text
├── docs/              # Documentation and coverage reports
├── test/              # All test files organized by type
│   ├── unit/          # Unit tests with mocks
│   ├── integration/   # Integration tests
│   ├── apitests/      # API endpoint tests
│   └── mocked/        # Mocked component tests
├── logs/              # Application logs
├── swagger-ui/        # Swagger UI assets
├── *_test.go          # Main package tests (root level)
└── *.go               # Go source files
```

## Development

## Compiling

The go CLI supports `go build` to produce an exacutable and will guide you through compilation errors.

## Docker

A Dockerfile needs to be created at the repository root.
You can derive from `scratch`, then `COPY` the sources into the image and build them.
The main command of the image should simply execute the executable obtained after building.

A more advanced solution can be achieved with a staged build.

Build command:

```bash
docker build -t my-image-name .
```

Listing the images:

```bash
docker images
```

Running a container:

```bash
docker run -it <imageID>
```

Play also with:

```bash
inspect ps stop rm rmi
```

## Testing & Coverage

The application features **86.6% test coverage** with comprehensive testing across multiple categories:

- **Unit Tests:** Component isolation with mocks
- **Integration Tests:** Real function testing  
- **API Tests:** End-to-end HTTP endpoint testing
- **Mocked Tests:** Dependency injection testing

### Quick Test Commands

```bash
# Run all tests
make test

# Generate coverage report  
make coverage

# Run specific test types
go test ./test/unit/ -v        # Unit tests
go test . -v                   # Main package tests (root level)
go test ./test/apitests/ -v    # API tests (requires server running)
go test ./test/integration/ -v # Integration tests
```

### Coverage Reports

- View detailed coverage: Open `docs/coverage.html`
- Test documentation: See `docs/tp-tests.txt`
- Testing guidelines: See `docs/TESTING.md`

## Swagger UI Setup

Done following [Swagger official doc](https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/installation.md#plain-old-htmlcssjs-standalone).

## Regenerate the OpenApi file

The Swagger UI consumes only JSON api specification, the function `yml2json` has been made to convert the YML format into JSON.
