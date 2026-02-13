package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/package-url/packageurl-go"
)

type Server struct {
	port int
	mux  *http.ServeMux
}

type ResolveResponse struct {
	Purl         string `json:"purl"`
	OCIReference string `json:"oci_reference"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Purl  string `json:"purl,omitempty"`
}

func NewServer(port int) *Server {
	mux := http.NewServeMux()
	s := &Server{
		port: port,
		mux:  mux,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", s.healthzHandler)
	s.mux.HandleFunc("/resolve", s.resolveHandler)
}

func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (s *Server) resolveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract purl query parameter
	purlParam := r.URL.Query().Get("purl")
	if purlParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "missing required parameter: purl",
		})
		return
	}

	// Decode URL-encoded purl
	decodedPurl, err := url.QueryUnescape(purlParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("failed to decode purl parameter: %v", err),
			Purl:  purlParam,
		})
		return
	}

	// Parse pURL
	pkg, err := packageurl.FromString(decodedPurl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("invalid purl format: %v", err),
			Purl:  decodedPurl,
		})
		return
	}

	// Validate pURL type is "oci"
	if pkg.Type != "oci" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("unsupported purl type '%s', only 'oci' is supported", pkg.Type),
			Purl:  decodedPurl,
		})
		return
	}

	// Validate name is not empty
	if pkg.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "purl is missing required component: name",
			Purl:  decodedPurl,
		})
		return
	}

	// Convert pURL to OCI reference
	ociRef := purlToOCIReference(pkg)

	// Return success response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ResolveResponse{
		Purl:         decodedPurl,
		OCIReference: ociRef,
	})
}

func purlToOCIReference(pkg packageurl.PackageURL) string {
	qualifiers := pkg.Qualifiers.Map()

	// Determine the base reference (registry + repository)
	var baseRef string
	if repoURL, ok := qualifiers["repository_url"]; ok {
		// repository_url contains the full path (registry/repository)
		// Strip http:// or https:// prefixes if present
		baseRef = strings.TrimPrefix(repoURL, "https://")
		baseRef = strings.TrimPrefix(baseRef, "http://")
	} else {
		// No repository_url, use default registry + name
		baseRef = fmt.Sprintf("docker.io/%s", pkg.Name)
	}

	// Determine tag or digest
	tagOrDigest := ":latest"
	if pkg.Version != "" {
		// Version field contains digest
		tagOrDigest = "@" + pkg.Version
	} else if tag, ok := qualifiers["tag"]; ok {
		tagOrDigest = ":" + tag
	}

	// Construct OCI reference
	return baseRef + tagOrDigest
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start(ctx context.Context) {
	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting server on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	fmt.Println("\nShutting down server...")

	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped")
}
