# purl-resolver

Service for resolving pURL identifiers to an OCI artifact

## Quick Start

### Local Development

```bash
# Build and run locally
go build -v -o purl-resolver .
./purl-resolver serve

# Run tests with Ginkgo
make test

# Run tests in watch mode (re-runs on file changes)
make test-watch

# Run tests with coverage
make test-coverage
```

### Container Development

This project uses [ko](https://ko.build/) for building containers and [KinD](https://kind.sigs.k8s.io/) for local Kubernetes testing.

#### Prerequisites

- [ko](https://ko.build/install/) - Container builder for Go
- [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) - Kubernetes in Docker
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Kubernetes CLI

#### Quick Test

```bash
# Deploy to local KinD cluster and run integration tests
make test-local-full

# Clean up
make clean
```

#### Development Workflow

```bash
# Create KinD cluster
make kind-create

# Build and deploy with ko
make ko-apply

# Check deployment status
kubectl get pods

# View logs
make logs

# Forward port to access locally
make port-forward

# In another terminal, test the service
curl http://localhost:8080/healthz

# Run integration tests
make test-integration

# Clean up
make kind-delete
```

#### Available Make Targets

Run `make help` to see all available targets:

- `make test` - Run unit tests
- `make test-integration` - Run integration tests
- `make deploy-local` - Create cluster and deploy application
- `make port-forward` - Forward localhost:8080 to service
- `make test-local-full` - Full automated test (deploy + integration tests)
- `make logs` - Show application logs
- `make clean` - Clean up everything

## Testing

This project uses [Ginkgo v2](https://onsi.github.io/ginkgo/) (BDD test framework) and [Gomega](https://onsi.github.io/gomega/) (assertion library) for testing.

### Running Tests

```bash
# Run unit tests
make test

# Run unit tests with coverage
make test-coverage

# View coverage in browser
make test-coverage-html

# Run tests in watch mode (re-runs on changes)
make test-watch

# Run integration tests (requires deployed service)
make test-integration

# Full integration test (deploy + test + cleanup)
make test-local-full
```

### Using Ginkgo CLI

```bash
# Install Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run tests with verbose output
ginkgo -v ./cmd

# Run specific tests by pattern
ginkgo -focus="Health Check" ./cmd

# Skip specific tests
ginkgo -skip="Integration" ./cmd

# Run tests until failure (useful for flaky tests)
ginkgo --until-it-fails ./cmd
```

### Test Organization

- **Unit Tests** (`cmd/serve_ginkgo_test.go`)
  - Package: `cmd` (white-box testing)
  - Build tag: `!integration`
  - Tests server initialization and HTTP handlers
  - Uses `httptest` for request/response simulation

- **Integration Tests** (`cmd/serve_integration_ginkgo_test.go`)
  - Package: `cmd_test` (black-box testing)
  - Build tag: `integration`
  - Tests deployed service in Kubernetes
  - Uses `Eventually()` for retry logic and `SpecContext` for timeout control

See [CLAUDE.md](./CLAUDE.md) for detailed testing guidance and examples.

## Architecture

The application is a minimal Go HTTP service that:

- Uses [Cobra](https://github.com/spf13/cobra) for CLI structure
- Exposes a health check endpoint at `/healthz`
- Runs on port 8080 by default
- Supports graceful shutdown

## Container Image

The container image is built using ko with a distroless base:

- Base: `gcr.io/distroless/static-debian12:nonroot`
- Runs as non-root user (UID 65532)
- Minimal attack surface (no shell, package manager)
- Automatic SBOM generation

## Kubernetes Deployment

The Kubernetes manifests include:

- Security contexts (non-root, read-only filesystem, no capabilities)
- Health probes (liveness and readiness)
- Resource limits (100m/64Mi request, 200m/128Mi limit)

## CI/CD

GitHub Actions workflows:

- `build-and-test.yml` - Build and unit tests
- `container-integration.yml` - Container build and integration tests in KinD
