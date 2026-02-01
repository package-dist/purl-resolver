package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "purl-resolver",
	Short: "Service for resolving pURL identifiers to an OCI artifact",
	Long:  `pURL Resolver is a service that resolves Package URL (pURL) identifiers to OCI artifacts.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
