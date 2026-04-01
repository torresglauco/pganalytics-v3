package main

import (
	"log"
	"os"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

var (
	Version = "0.1.0"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	tr := transport.NewStdioTransport(os.Stdin, os.Stdout)
	server := NewMCPServer(tr)

	// Register tools (will be added in later tasks)
	// server.RegisterTool("table_stats", handlers.TableStats)
	// server.RegisterTool("query_analysis", handlers.QueryAnalysis)
	// server.RegisterTool("index_suggest", handlers.IndexSuggest)

	log.SetOutput(os.Stderr)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
