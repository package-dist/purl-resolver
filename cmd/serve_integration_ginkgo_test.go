//go:build integration

package cmd_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
})
