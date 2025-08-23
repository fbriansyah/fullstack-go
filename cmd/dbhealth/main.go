package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go-templ-template/internal/config"
	"go-templ-template/internal/shared/database"
)

func main() {
	var (
		timeout = flag.Duration("timeout", 10*time.Second, "Timeout for health check")
		format  = flag.String("format", "text", "Output format: text, json")
		wait    = flag.Bool("wait", false, "Wait for database to become available")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database, database.DefaultConnectionOptions())
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	// Create health checker
	healthChecker := database.NewHealthChecker(db)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Wait for connection if requested
	if *wait {
		log.Println("Waiting for database to become available...")
		err = healthChecker.WaitForConnection(ctx, 30, 2*time.Second)
		if err != nil {
			log.Fatalf("Database did not become available: %v", err)
		}
	}

	// Perform health check
	status := healthChecker.Check(ctx)

	// Output results
	switch *format {
	case "json":
		jsonData, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonData))

	case "text":
		fmt.Printf("Database Status: %s\n", status.Status)
		if status.Message != "" {
			fmt.Printf("Message: %s\n", status.Message)
		}
		fmt.Printf("Latency: %v\n", status.Latency)
		fmt.Printf("Timestamp: %s\n", status.Timestamp.Format(time.RFC3339))
		fmt.Printf("Connections:\n")
		fmt.Printf("  Open: %d\n", status.Connections.OpenConnections)
		fmt.Printf("  In Use: %d\n", status.Connections.InUseConnections)
		fmt.Printf("  Idle: %d\n", status.Connections.IdleConnections)
		fmt.Printf("  Wait Count: %d\n", status.Connections.WaitCount)
		fmt.Printf("  Wait Duration: %dms\n", status.Connections.WaitDuration)

	default:
		log.Fatalf("Unknown format: %s", *format)
	}

	// Exit with appropriate code
	if status.Status != "healthy" {
		fmt.Fprintf(os.Stderr, "Database is not healthy\n")
		os.Exit(1)
	}
}

// Example usage:
// go run cmd/dbhealth/main.go
// go run cmd/dbhealth/main.go -format=json
// go run cmd/dbhealth/main.go -wait -timeout=30s
