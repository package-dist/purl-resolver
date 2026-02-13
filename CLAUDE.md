# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pURL Resolver is a Go-based HTTP service that resolves Package URL (pURL) identifiers to OCI artifacts. The service uses the Cobra library for CLI command structure.

## Common Commands

### Local Development

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

### Container Development

```bash
# Deploy to local KinD cluster and run integration tests
make test-local-full

# Create KinD cluster
make kind-create

# Build and deploy with ko
make ko-apply

# View application logs
make logs

# Forward port for local access
make port-forward

# Run integration tests (requires service to be running)
make test-integration

# Delete KinD cluster
make kind-delete

# See all available targets
make help
```

## Testing

### Unit Tests

Tests are colocated with source files using the `_test.go` suffix (e.g., `cmd/serve_test.go`). The test suite includes unit tests for HTTP handlers using `httptest` for request/response simulation.

Run unit tests:
```bash
go test -v ./...
```

### Integration Tests

Integration tests verify the containerized application in a Kubernetes environment. These tests use the `//go:build integration` build tag and are located in `cmd/serve_integration_test.go`.

Run integration tests:
```bash
# Requires the service to be running (via port-forward or deployed)
PURL_RESOLVER_SERVICE_URL=http://localhost:8080 go test -v -tags=integration ./cmd
```

The integration tests:
- Use the `cmd_test` package for black-box testing
- Support the `PURL_RESOLVER_SERVICE_URL` environment variable for flexibility
- Include retry logic to wait for service startup
- Test the `/healthz` endpoint returns 200 OK with body "OK"

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

## Containerization

### Ko Configuration

The project uses [ko](https://ko.build/) for building container images:
- Configuration in `.ko.yaml` specifies the base image
- Uses `gcr.io/distroless/static-debian12:nonroot` for minimal, secure images
- No Dockerfile required - ko handles everything

Build and push to a registry:
```bash
KO_DOCKER_REPO=gcr.io/my-project ko build .
```

Build and load into KinD:
```bash
KO_DOCKER_REPO=kind.local ko apply -f deploy/
```

### Kubernetes Manifests

Located in `deploy/`:
- `deployment.yaml` - Defines the Deployment with security contexts and health probes
- `service.yaml` - Exposes the application on port 80 (routes to container port 8080)

The Deployment uses the special image reference `ko://github.com/package-dist/purl-resolver` which ko resolves during deployment.

### Security

The container runs with:
- Non-root user (UID 65532)
- Read-only root filesystem
- All capabilities dropped
- No privilege escalation allowed

### Local Testing with KinD

KinD (Kubernetes in Docker) provides a local Kubernetes cluster for testing:

```bash
# Create cluster
make kind-create

# Deploy application
make ko-apply

# Check status
kubectl get pods
kubectl get svc

# Access via port-forward
make port-forward

# View logs
make logs

# Clean up
make kind-delete
```

### CI/CD

The `.github/workflows/container-integration.yml` workflow:
1. Builds the container with ko
2. Deploys to a KinD cluster
3. Runs integration tests
4. Shows logs on failure

Runs on push to `main` or `containerize-with-ko` branches and all PRs to `main`.
