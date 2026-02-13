# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pURL Resolver is a Go-based HTTP service that resolves Package URL (pURL) identifiers to OCI artifacts. The service uses the Cobra library for CLI command structure.

## Common Commands

```bash
# Download dependencies
go mod download

# Build the binary
go build -v -o purl-resolver .

# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -func=coverage.out

# Run the server (default port 8080)
go run . serve

# Run the server on a specific port
go run . serve --port 9090
```

## Testing

Tests are colocated with source files using the `_test.go` suffix (e.g., `cmd/serve_test.go`). The test suite includes unit tests for HTTP handlers using `httptest` for request/response simulation.

## Architecture

### Command Structure

The application uses Cobra for CLI commands:
- **main.go**: Entry point that calls `cmd.Execute()`
- **cmd/root.go**: Root command definition for `purl-resolver`
- **cmd/serve.go**: Serve command that starts the HTTP server

### HTTP Server

The Server type in `cmd/serve.go` implements an HTTP server with:
- HTTP multiplexer (`http.ServeMux`) for routing
- Configurable port via `--port` flag
- Graceful shutdown with 5-second timeout on context cancellation
- Health check endpoint at `/healthz`

The server starts in a goroutine and waits for context cancellation (e.g., SIGINT) before initiating graceful shutdown.

### Adding New Commands

1. Create a new file in `cmd/` (e.g., `cmd/newcommand.go`)
2. Define the command as a `cobra.Command`
3. Register it in `init()` using `rootCmd.AddCommand()`

### Adding New HTTP Routes

In `cmd/serve.go`, add handler methods to the `Server` type and register them in `registerRoutes()`.
