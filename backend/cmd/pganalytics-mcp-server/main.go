package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/handlers"
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/server"
	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

var (
	Version = "0.1.0"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost/pganalytics"
	}

	// Connect to database using pgx driver
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		// Continue without DB - will use mock data
		db = nil
	} else {
		defer db.Close()
		if err := db.Ping(); err != nil {
			log.Printf("Warning: Database ping failed: %v", err)
			db = nil
		}
	}

	// Create transport and server
	tr := transport.NewStdioTransport(os.Stdin, os.Stdout)
	mcp := server.NewMCPServer(tr)

	// Initialize handler context
	handlerCtx := handlers.NewHandlerContext(db)

	// Register all tools
	mcp.RegisterDefaultHandlers(handlerCtx)
	mcp.RegisterTool("anomaly_detect", func(params map[string]interface{}) (interface{}, error) {
		return handlerCtx.DetectAnomalies(params)
	})

	log.SetOutput(os.Stderr)
	log.Printf("pgAnalytics MCP Server v%s starting", Version)

	if err := mcp.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
