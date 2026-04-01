package main

import (
	"fmt"
	"os"
	"pganalytics-cli/commands"
)

var (
	Version = "0.1.0"
	Commit  = "dev"
	Date    = "unknown"
)

func main() {
	rootCmd := commands.NewRootCmd(Version)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
