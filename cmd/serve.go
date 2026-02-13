package cmd

import (
	"github.com/package-dist/purl-resolver/server"
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
		srv := server.NewServer(port)
		srv.Start(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the HTTP server on")
}
