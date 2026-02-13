.PHONY: help test test-integration kind-create kind-delete ko-apply deploy-local port-forward test-local-full logs clean

# Default target
.DEFAULT_GOAL := help

# Variables
CLUSTER_NAME ?= purl-resolver
NAMESPACE ?= default
SERVICE_NAME ?= purl-resolver
LOCAL_PORT ?= 8080

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

test: ## Run unit tests
	go test -v ./...

test-integration: ## Run integration tests (requires service to be running)
	go test -v -tags=integration ./cmd

kind-create: ## Create a KinD cluster
	kind create cluster --name $(CLUSTER_NAME)

kind-delete: ## Delete the KinD cluster
	kind delete cluster --name $(CLUSTER_NAME)

ko-apply: ## Build and deploy to KinD cluster using ko
	KIND_CLUSTER_NAME=$(CLUSTER_NAME) KO_DOCKER_REPO=kind.local ko apply -f deploy/

deploy-local: kind-create ## Create cluster, deploy application, and wait for readiness
	@echo "Building and deploying application..."
	@KIND_CLUSTER_NAME=$(CLUSTER_NAME) KO_DOCKER_REPO=kind.local ko apply -f deploy/
	@echo "Waiting for deployment to be ready..."
	@kubectl wait --for=condition=available --timeout=120s deployment/$(SERVICE_NAME) -n $(NAMESPACE)
	@echo "Deployment ready!"
	@kubectl get pods -n $(NAMESPACE) -l app=$(SERVICE_NAME)

port-forward: ## Forward localhost:8080 to the service (runs in foreground)
	@echo "Port forwarding $(LOCAL_PORT) -> service/$(SERVICE_NAME):80"
	@echo "Press Ctrl+C to stop"
	@kubectl port-forward -n $(NAMESPACE) service/$(SERVICE_NAME) $(LOCAL_PORT):80

port-forward-bg: ## Forward localhost:8080 to the service (runs in background)
	@echo "Starting port forward in background..."
	@kubectl port-forward -n $(NAMESPACE) service/$(SERVICE_NAME) $(LOCAL_PORT):80 > /dev/null 2>&1 & echo $$! > .port-forward.pid
	@sleep 2
	@echo "Port forward running (PID: $$(cat .port-forward.pid))"

stop-port-forward: ## Stop background port forwarding
	@if [ -f .port-forward.pid ]; then \
		kill $$(cat .port-forward.pid) 2>/dev/null || true; \
		rm .port-forward.pid; \
		echo "Port forward stopped"; \
	else \
		echo "No port forward PID file found"; \
	fi

test-local-full: deploy-local port-forward-bg ## Deploy and run full integration tests
	@echo "Running integration tests..."
	@PURL_RESOLVER_SERVICE_URL=http://localhost:$(LOCAL_PORT) go test -v -tags=integration ./cmd || ($(MAKE) stop-port-forward && exit 1)
	@$(MAKE) stop-port-forward
	@echo "Integration tests passed!"

logs: ## Show application logs
	@kubectl logs -n $(NAMESPACE) -l app=$(SERVICE_NAME) --tail=100 -f

clean: stop-port-forward kind-delete ## Clean up everything (stop port forward, delete cluster)
	@echo "Cleanup complete"
