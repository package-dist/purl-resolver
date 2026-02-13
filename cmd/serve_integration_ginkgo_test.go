//go:build integration

package cmd_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Response structs for testing (package is cmd_test, not cmd)
type ResolveResponse struct {
	Purl         string `json:"purl"`
	OCIReference string `json:"oci_reference"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Purl  string `json:"purl,omitempty"`
}

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
			var healthzURL string

			BeforeEach(func() {
				healthzURL = fmt.Sprintf("%s/healthz", serviceURL)
			})

			It("should respond to /healthz with 200 OK", func(ctx SpecContext) {
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

			It("should return 'OK' in the response body", func(ctx SpecContext) {
				Eventually(ctx, func() string {
					resp, err := http.Get(healthzURL)
					if err != nil {
						return ""
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return ""
					}
					return string(body)
				}).
					WithPolling(1 * time.Second).
					Should(Equal("OK"))
			}, SpecTimeout(45*time.Second))
		})
	})

	Describe("Resolve Endpoint [Integration]", func() {
		When("the service is deployed", func() {
			It("should resolve a valid OCI pURL with 200 OK", func(ctx SpecContext) {
				purl := url.QueryEscape("pkg:oci/nginx")
				resolveURL := fmt.Sprintf("%s/resolve?purl=%s", serviceURL, purl)

				Eventually(ctx, func() int {
					resp, err := http.Get(resolveURL)
					if err != nil {
						return 0
					}
					defer resp.Body.Close()
					return resp.StatusCode
				}).
					WithPolling(1 * time.Second).
					Should(Equal(http.StatusOK))
			}, SpecTimeout(45*time.Second))

			It("should return correct JSON response for valid OCI pURL", func(ctx SpecContext) {
				purl := url.QueryEscape("pkg:oci/nginx")
				resolveURL := fmt.Sprintf("%s/resolve?purl=%s", serviceURL, purl)

				Eventually(ctx, func() *ResolveResponse {
					resp, err := http.Get(resolveURL)
					if err != nil {
						return nil
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						return nil
					}

					var result ResolveResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						return nil
					}
					return &result
				}).
					WithPolling(1 * time.Second).
					Should(And(
						Not(BeNil()),
						HaveField("Purl", "pkg:oci/nginx"),
						HaveField("OCIReference", "docker.io/nginx:latest"),
					))
			}, SpecTimeout(45*time.Second))

			It("should return correct OCI reference with custom registry", func(ctx SpecContext) {
				purl := url.QueryEscape("pkg:oci/app?repository_url=ghcr.io/myorg/app")
				resolveURL := fmt.Sprintf("%s/resolve?purl=%s", serviceURL, purl)

				Eventually(ctx, func() *ResolveResponse {
					resp, err := http.Get(resolveURL)
					if err != nil {
						return nil
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						return nil
					}

					var result ResolveResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						return nil
					}
					return &result
				}).
					WithPolling(1 * time.Second).
					Should(And(
						Not(BeNil()),
						HaveField("OCIReference", "ghcr.io/myorg/app:latest"),
					))
			}, SpecTimeout(45*time.Second))

			It("should return 400 when purl parameter is missing", func(ctx SpecContext) {
				resolveURL := fmt.Sprintf("%s/resolve", serviceURL)

				Eventually(ctx, func() int {
					resp, err := http.Get(resolveURL)
					if err != nil {
						return 0
					}
					defer resp.Body.Close()
					return resp.StatusCode
				}).
					WithPolling(1 * time.Second).
					Should(Equal(http.StatusBadRequest))
			}, SpecTimeout(45*time.Second))

			It("should return 400 for invalid pURL format", func(ctx SpecContext) {
				purl := url.QueryEscape("not-a-valid-purl")
				resolveURL := fmt.Sprintf("%s/resolve?purl=%s", serviceURL, purl)

				Eventually(ctx, func() int {
					resp, err := http.Get(resolveURL)
					if err != nil {
						return 0
					}
					defer resp.Body.Close()
					return resp.StatusCode
				}).
					WithPolling(1 * time.Second).
					Should(Equal(http.StatusBadRequest))
			}, SpecTimeout(45*time.Second))

			It("should return 400 for unsupported pURL type", func(ctx SpecContext) {
				purl := url.QueryEscape("pkg:npm/express@4.0.0")
				resolveURL := fmt.Sprintf("%s/resolve?purl=%s", serviceURL, purl)

				var errorResp *ErrorResponse
				Eventually(ctx, func() *ErrorResponse {
					resp, err := http.Get(resolveURL)
					if err != nil {
						return nil
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusBadRequest {
						return nil
					}

					var result ErrorResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						return nil
					}
					errorResp = &result
					return &result
				}).
					WithPolling(1 * time.Second).
					Should(Not(BeNil()))

				Expect(errorResp.Error).To(Equal("unsupported purl type 'npm', only 'oci' is supported"))
			}, SpecTimeout(45*time.Second))
		})
	})
})
