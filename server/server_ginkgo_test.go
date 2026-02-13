//go:build !integration

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
				server.ServeHTTP(w, req)
			})

			It("should return 200 OK status", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should return 'OK' in the body", func() {
				Expect(w.Body.String()).To(Equal("OK"))
			})
		})
	})

	Describe("Server Initialization", func() {
		When("creating a new server", func() {
			It("should initialize with the correct port", func() {
				port := 9090
				server := NewServer(port)
				Expect(server.Port()).To(Equal(port))
			})

			It("should initialize the HTTP multiplexer", func() {
				server := NewServer(8080)
				Expect(server.mux).NotTo(BeNil())
			})
		})
	})

	Describe("Resolve Handler", func() {
		var (
			server *Server
		)

		BeforeEach(func() {
			server = NewServer(8080)
		})

		Context("Success scenarios", func() {
			When("resolving OCI pURL with digest and repository_url", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("pkg:oci/debian@sha256:244fd47?repository_url=docker.io/library/debian")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 200 OK status", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
				})

				It("should return JSON with correct OCI reference", func() {
					var response ResolveResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Purl).To(Equal("pkg:oci/debian@sha256:244fd47?repository_url=docker.io/library/debian"))
					Expect(response.OCIReference).To(Equal("docker.io/library/debian@sha256:244fd47"))
				})
			})

			When("resolving OCI pURL with name only", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("pkg:oci/nginx")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 200 OK status", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
				})

				It("should apply defaults (docker.io registry, latest tag)", func() {
					var response ResolveResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Purl).To(Equal("pkg:oci/nginx"))
					Expect(response.OCIReference).To(Equal("docker.io/nginx:latest"))
				})
			})

			When("resolving OCI pURL with tag qualifier", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("pkg:oci/nginx?tag=alpine")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 200 OK status", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
				})

				It("should use tag from qualifier", func() {
					var response ResolveResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Purl).To(Equal("pkg:oci/nginx?tag=alpine"))
					Expect(response.OCIReference).To(Equal("docker.io/nginx:alpine"))
				})
			})

			When("resolving OCI pURL with custom registry", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("pkg:oci/app?repository_url=ghcr.io/myorg/app")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 200 OK status", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
				})

				It("should use custom registry from repository_url", func() {
					var response ResolveResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Purl).To(Equal("pkg:oci/app?repository_url=ghcr.io/myorg/app"))
					Expect(response.OCIReference).To(Equal("ghcr.io/myorg/app:latest"))
				})
			})

			When("resolving URL-encoded digest", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					// sha256:244fd47 is URL-encoded as sha256%3A244fd47
					purl := "pkg:oci/debian@sha256%3A244fd47"
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 200 OK status", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
				})

				It("should decode digest correctly", func() {
					var response ResolveResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.OCIReference).To(Equal("docker.io/debian@sha256:244fd47"))
				})
			})
		})

		Context("Error scenarios", func() {
			When("purl parameter is missing", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					req = httptest.NewRequest(http.MethodGet, "/resolve", nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 400 Bad Request status", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
				})

				It("should return error message in JSON", func() {
					var response ErrorResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Error).To(Equal("missing required parameter: purl"))
				})
			})

			When("purl format is invalid", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("not-a-valid-purl")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 400 Bad Request status", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
				})

				It("should return error message in JSON", func() {
					var response ErrorResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Error).To(ContainSubstring("invalid purl format"))
				})
			})

			When("purl type is not 'oci'", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("pkg:npm/express@4.0.0")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 400 Bad Request status", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
				})

				It("should return error message indicating unsupported type", func() {
					var response ErrorResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Error).To(Equal("unsupported purl type 'npm', only 'oci' is supported"))
				})
			})

			When("purl has another unsupported type", func() {
				var (
					req *http.Request
					w   *httptest.ResponseRecorder
				)

				BeforeEach(func() {
					purl := url.QueryEscape("pkg:pypi/django@3.2.0")
					req = httptest.NewRequest(http.MethodGet, "/resolve?purl="+purl, nil)
					w = httptest.NewRecorder()
					server.ServeHTTP(w, req)
				})

				It("should return 400 Bad Request status", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
				})

				It("should return error message indicating unsupported type", func() {
					var response ErrorResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					Expect(err).To(BeNil())
					Expect(response.Error).To(Equal("unsupported purl type 'pypi', only 'oci' is supported"))
				})
			})
		})
	})
})
