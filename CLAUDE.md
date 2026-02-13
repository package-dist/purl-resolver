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

# Run all tests with Ginkgo
make test
ginkgo -v ./cmd

# Run tests with coverage
make test-coverage
ginkgo -r --cover --coverprofile=coverage.out

# View coverage report
go tool cover -func=coverage.out

# Run tests in watch mode
make test-watch
ginkgo watch ./cmd

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
ginkgo -v --tags=integration ./cmd

# Delete KinD cluster
make kind-delete

# See all available targets
make help
```

## Testing

This project uses **Ginkgo v2** (BDD test framework) and **Gomega** (assertion library) for all tests. This encourages behavior-driven testing practices with clear, readable test specifications.

### BDD Testing Philosophy

Tests should describe **behavior** (what the system does) rather than implementation details. Use descriptive test organization:

- **`Describe`** - Component or feature being tested (e.g., "Server", "Health Check Handler")
- **`When`** - Scenario or condition (e.g., "when handling /healthz requests", "when the service is deployed")
- **`It`** - Expected behavior (e.g., "should return 200 OK status", "should respond with 'OK'")
- **`BeforeEach`** - Setup code that runs before each test
- **`AfterEach`** - Cleanup code that runs after each test

### Quick Start

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run unit tests
make test
ginkgo -v ./cmd

# Run integration tests (requires deployed service)
make test-integration
ginkgo -v --tags=integration ./cmd

# Watch mode (re-runs on file changes)
make test-watch
ginkgo watch ./cmd

# Coverage
make test-coverage
ginkgo -r --cover --coverprofile=coverage.out
```

### Unit Tests

**Location**: `cmd/serve_ginkgo_test.go`
**Package**: `cmd` (white-box testing - can access internal types)
**Build Tag**: `//go:build !integration` (excluded from integration test runs)
**Suite**: `cmd/cmd_unit_suite_test.go`

Unit tests verify individual components in isolation using `httptest` for HTTP handler testing.

**Example Pattern**:
```go
var _ = Describe("Server", func() {
    Describe("Health Check Handler", func() {
        When("handling /healthz requests", func() {
            var (
                server *Server
                req    *http.Request
                w      *httptest.ResponseRecorder
            )

            BeforeEach(func() {
                server = NewServer(8080)
                req = httptest.NewRequest(http.MethodGet, "/healthz", nil)
                w = httptest.NewRecorder()
                server.mux.ServeHTTP(w, req)
            })

            It("should return 200 OK status", func() {
                Expect(w.Code).To(Equal(http.StatusOK))
            })

            It("should return 'OK' in the body", func() {
                Expect(w.Body.String()).To(Equal("OK"))
            })
        })
    })
})
```

### Integration Tests

**Location**: `cmd/serve_integration_ginkgo_test.go`
**Package**: `cmd_test` (black-box testing - tests external API contract)
**Build Tag**: `//go:build integration` (only runs with `--tags=integration`)
**Suite**: `cmd/cmd_integration_suite_test.go`

Integration tests verify the deployed service in a Kubernetes environment. They use:
- **`Eventually()`** - Gomega's polling assertion for async operations
- **`SpecContext`** - Ginkgo's context for timeout control and graceful termination
- **`SpecTimeout()`** - Decorator to set overall spec timeout
- **`WithPolling()`** - Configure polling interval for `Eventually()`

**Example Pattern**:
```go
var _ = Describe("pURL Resolver Service [Integration]", func() {
    var serviceURL string

    BeforeEach(func() {
        serviceURL = os.Getenv("PURL_RESOLVER_SERVICE_URL")
        if serviceURL == "" {
            serviceURL = "http://localhost:8080"
        }
    })

    Describe("Health Check Endpoint", func() {
        When("the service is deployed", func() {
            It("should respond to /healthz with 200 OK", func(ctx SpecContext) {
                healthzURL := fmt.Sprintf("%s/healthz", serviceURL)

                Eventually(ctx, func() int {
                    resp, err := http.Get(healthzURL)
                    if err != nil {
                        return 0
                    }
                    defer resp.Body.Close()
                    return resp.StatusCode
                }).
                    WithPolling(1 * time.Second).
                    Should(Equal(http.StatusOK))
            }, SpecTimeout(45*time.Second))
        })
    })
})
```

**Key Integration Test Patterns**:

1. **SpecContext for Timeout Control**: Pass `ctx SpecContext` to `It` blocks to enable timeout management
   ```go
   It("should respond", func(ctx SpecContext) {
       Eventually(ctx, func() int {
           // Test code
       }).Should(Equal(200))
   }, SpecTimeout(30*time.Second))
   ```

2. **Eventually() with Context**: Always pass `ctx` as first argument to `Eventually()` for proper timeout propagation
   ```go
   Eventually(ctx, func() int { return statusCode }).Should(Equal(200))
   ```

3. **Polling Configuration**: Use `WithPolling()` to set retry interval
   ```go
   Eventually(ctx, func() bool { return ready }).
       WithPolling(1 * time.Second).
       Should(BeTrue())
   ```

4. **Graceful Termination**: SpecContext enables cleanup on timeout/interrupt via `DeferCleanup`
   ```go
   BeforeEach(func(ctx SpecContext) {
       resource := setupResource()
       DeferCleanup(func() { resource.Cleanup() })
   })
   ```

### Gomega Matchers

Common assertion patterns:

```go
// Equality
Expect(actual).To(Equal(expected))
Expect(actual).NotTo(Equal(unexpected))

// Nil checks
Expect(value).To(BeNil())
Expect(value).NotTo(BeNil())

// Numeric comparisons
Expect(value).To(BeNumerically(">", 100))
Expect(value).To(BeNumerically("<=", 200))

// String matchers
Expect(str).To(ContainSubstring("error"))
Expect(str).To(HavePrefix("http://"))
Expect(str).To(MatchRegexp(`\d{3}`))

// Boolean
Expect(condition).To(BeTrue())
Expect(condition).To(BeFalse())

// Collections
Expect(slice).To(HaveLen(3))
Expect(slice).To(ContainElement("foo"))
Expect(slice).To(BeEmpty())

// HTTP Status
Expect(w.Code).To(Equal(http.StatusOK))
```

### Ginkgo CLI Tips

```bash
# Run specific tests by pattern
ginkgo -focus="Health Check" ./cmd
ginkgo -focus="Integration" --tags=integration ./cmd

# Skip tests by pattern
ginkgo -skip="slow" ./cmd

# Run tests until failure (find flaky tests)
ginkgo --until-it-fails ./cmd

# Randomize test order (find dependencies)
ginkgo --randomize-all ./cmd

# Run tests with race detector
ginkgo --race ./cmd

# Verbose output with trace
ginkgo -v --trace ./cmd

# Run with coverage
ginkgo --cover --coverprofile=coverage.out ./cmd

# Parallel execution (be careful with integration tests)
ginkgo -p ./cmd
```

### Writing New Tests

**For Unit Tests**:
1. Add test specs to `cmd/serve_ginkgo_test.go`
2. Use `Describe` for components, `When` for scenarios, `It` for behaviors
3. Use `BeforeEach` for shared setup
4. Use Gomega matchers for assertions
5. No need for `SpecContext` in unit tests

**For Integration Tests**:
1. Add test specs to `cmd/serve_integration_ginkgo_test.go`
2. Use `Eventually(ctx, ...)` for async operations
3. Pass `ctx SpecContext` to `It` blocks
4. Set appropriate timeout with `SpecTimeout()` decorator
5. Configure polling interval with `WithPolling()`
6. Test external API behavior, not internal implementation

### Comparison: Standard Testing vs Ginkgo

**Before (Standard Testing)**:
```go
func TestHealthzHandler(t *testing.T) {
    server := NewServer(8080)
    req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    w := httptest.NewRecorder()
    server.mux.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
    }
    if w.Body.String() != "OK" {
        t.Errorf("Expected body %q, got %q", "OK", w.Body.String())
    }
}
```

**After (Ginkgo + Gomega)**:
```go
var _ = Describe("Server", func() {
    Describe("Health Check Handler", func() {
        When("handling /healthz requests", func() {
            var (
                server *Server
                req    *http.Request
                w      *httptest.ResponseRecorder
            )

            BeforeEach(func() {
                server = NewServer(8080)
                req = httptest.NewRequest(http.MethodGet, "/healthz", nil)
                w = httptest.NewRecorder()
                server.mux.ServeHTTP(w, req)
            })

            It("should return 200 OK status", func() {
                Expect(w.Code).To(Equal(http.StatusOK))
            })

            It("should return 'OK' in the body", func() {
                Expect(w.Body.String()).To(Equal("OK"))
            })
        })
    })
})
```

**Benefits**:
- Better readability and organization
- Clear separation of setup, execution, and assertions
- Expressive matchers with better failure messages
- Built-in support for async/polling operations
- Rich CLI for focusing, skipping, and filtering tests

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
