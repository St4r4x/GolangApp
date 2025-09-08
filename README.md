# Introduction

![Coverage](https://img.shields.io/badge/Coverage-54.6%25-yellow)

Cats demo REST API used to manage a local database of 🐈.

## Run locally

```bash
go run .
```

Then you can browse:

- the home page: http://localhost:8080
- the Swagger UI : http://localhost:8080/swagger/
- the logs : http://localhost:8080/logs

# Dev

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

## Testing

The application includes comprehensive testing with **71 total tests** across multiple categories:

- **Unit Tests** (20+ tests): `go test ./test/unit/... -v`
- **API Tests** (8 tests): `go test ./test/apitests/... -v`
- **Integration Tests** (3 tests): `go test ./test/integration/... -v`
- **Mocked Tests** (2 tests): `go test ./test/mocked/... -v`
- **Main Package Tests** (7 tests): `go test . -v`
- **Plus additional sub-tests and variations**

### Running Tests

Run all tests: `go test ./test/... -v`

### Test Coverage

Generate coverage report: `go test -coverprofile=coverage.out ./... -coverpkg=./...`

View coverage in browser: `go tool cover -html=coverage.out`

Update coverage badge: `./update-coverage.sh`

Test structure:

```
test/
├── unit/       # Component isolation testing
├── apitests/   # End-to-end HTTP testing
└── mocked/     # Dependency injection with mocks
```

# Swagger UI setup (already done)

Done following [Swagger official doc](https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/installation.md#plain-old-htmlcssjs-standalone).

## Regenerate the OpenApi file

The Swagger UI consumes only JSON api specification, the function `yml2json` has been made to convert the YML format into JSON.
