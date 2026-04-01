package commands

import (
	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "pganalytics",
		Short:   "pgAnalytics CLI - PostgreSQL monitoring from the command line",
		Long:    "pgAnalytics is a powerful CLI tool for PostgreSQL monitoring, analysis, and optimization",
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewIndexCmd())
	rootCmd.AddCommand(NewVacuumCmd())
	rootCmd.AddCommand(NewMCPCmd())

	// Global flags
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication")
	rootCmd.PersistentFlags().String("format", "table", "Output format (table, json, csv)")

	return rootCmd
}
