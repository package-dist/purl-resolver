package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	port int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP web server",
	Long:  `Start the HTTP web server that provides the pURL resolver API endpoints.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := NewServer(port)
		server.Start(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the HTTP server on")
}

type Server struct {
	port int
	mux  *http.ServeMux
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
}

func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
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
