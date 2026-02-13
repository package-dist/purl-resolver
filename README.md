# purl-resolver

Service for resolving pURL identifiers to an OCI artifact

## Quick Start

### Local Development

```bash
# Build and run locally
go build -v -o purl-resolver .
./purl-resolver serve

# Run tests
go test -v ./...
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
