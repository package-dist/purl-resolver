package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	server := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expectedBody := "OK"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}

func TestNewServer(t *testing.T) {
	port := 9090
	server := NewServer(port)

	if server.port != port {
		t.Errorf("Expected port %d, got %d", port, server.port)
	}

	if server.mux == nil {
		t.Error("Expected mux to be initialized, got nil")
	}
}
