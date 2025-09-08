# Introduction

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

The application includes comprehensive testing with 34 passing tests:

- **Unit Tests** (7 tests): `go test ./test/unit/... -v`
- **API Tests** (8 tests): `go test ./test/apitests/... -v`
- **Mocked Tests** (19 tests): `go test ./test/mocked/... -v`

Run all tests: `go test ./test/... -v`

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
