package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage pgAnalytics MCP Server",
	Long:  "Start, stop, or manage the pgAnalytics Model Context Protocol server",
}

var mcpStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Execute the MCP server binary
		mcpServerPath := os.ExpandEnv("$HOME/.local/bin/pganalytics-mcp-server")
		execCmd := exec.Command(mcpServerPath)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		return execCmd.Run()
	},
}

var mcpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check MCP server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("MCP Server Status: Not yet implemented")
		return nil
	},
}

func NewMCPCmd() *cobra.Command {
	mcpCmd.AddCommand(mcpStartCmd)
	mcpCmd.AddCommand(mcpStatusCmd)
	return mcpCmd
}
