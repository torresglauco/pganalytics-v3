package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"pganalytics-cli/internal/config"
	"github.com/spf13/cobra"
)

var configStore *config.FileStore

func init() {
	// Initialize config store
	configDir := filepath.Join(os.Getenv("HOME"), ".pganalytics")
	configFile := filepath.Join(configDir, "config.json")
	configStore = config.NewFileStore(configFile)
}

func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage pgAnalytics configuration",
		Long:  "Get, set, and list configuration values",
	}

	// Subcommand: config set
	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := configStore.Set(args[0], args[1]); err != nil {
				return fmt.Errorf("failed to set config: %w", err)
			}
			fmt.Printf("✓ Set %s = %s\n", args[0], args[1])
			return nil
		},
	}

	// Subcommand: config get
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			val, err := configStore.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}
			fmt.Println(val)
			return nil
		},
	}

	// Subcommand: config list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			all := configStore.GetAll()
			if len(all) == 0 {
				fmt.Println("No configuration values set")
				return nil
			}

			fmt.Println("Configuration:")
			for key, val := range all {
				fmt.Printf("  %s = %s\n", key, val)
			}
			return nil
		},
	}

	configCmd.AddCommand(setCmd, getCmd, listCmd)
	return configCmd
}
