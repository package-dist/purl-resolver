//go:build !integration

package cmd

import (
	"net/http"
	"net/http/httptest"

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

	Describe("Server Initialization", func() {
		When("creating a new server", func() {
			It("should initialize with the correct port", func() {
				port := 9090
				server := NewServer(port)
				Expect(server.port).To(Equal(port))
			})

			It("should initialize the HTTP multiplexer", func() {
				server := NewServer(8080)
				Expect(server.mux).NotTo(BeNil())
			})
		})
	})
})
