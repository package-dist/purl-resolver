//go:build integration

package cmd_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestHealthzEndpointInKubernetes(t *testing.T) {
	// Get service URL from environment variable or default to localhost
	serviceURL := os.Getenv("PURL_RESOLVER_SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "http://localhost:8080"
	}

	healthzURL := fmt.Sprintf("%s/healthz", serviceURL)

	// Retry logic to wait for service to be ready
	var lastErr error
	maxRetries := 30
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(healthzURL)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect: %w", err)
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			break
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			break
		}

		// Check response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		expectedBody := "OK"
		if string(body) != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, string(body))
		}

		// Success!
		t.Logf("Health check successful after %d attempts", i+1)
		return
	}

	// If we got here, all retries failed
	t.Fatalf("health check failed after %d retries: %v", maxRetries, lastErr)
}
