package main

import (
	"flag"
	"fmt"
	"log"

	"go-templ-template/internal/config"
	"go-templ-template/internal/shared/database"
)

func main() {
	var (
		action         = flag.String("action", "up", "Migration action: up, down, steps, version, force")
		steps          = flag.Int("steps", 0, "Number of steps for 'steps' action")
		version        = flag.Int("version", 0, "Version for 'force' action")
		migrationsPath = flag.String("migrations", "./migrations", "Path to migrations directory")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create migration runner
	runner, err := database.NewMigrationRunner(&cfg.Database, *migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	// Execute the requested action
	switch *action {
	case "up":
		log.Println("Running migrations up...")
		err = runner.Up()
		if err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
		log.Println("Migrations completed successfully")

	case "down":
		log.Println("Running migrations down...")
		err = runner.Down()
		if err != nil {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
		log.Println("Migrations rolled back successfully")

	case "steps":
		if *steps == 0 {
			log.Fatal("Steps count must be specified with -steps flag")
		}
		log.Printf("Running %d migration steps...", *steps)
		err = runner.Steps(*steps)
		if err != nil {
			log.Fatalf("Failed to run migration steps: %v", err)
		}
		log.Println("Migration steps completed successfully")

	case "version":
		v, dirty, err := runner.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("Current migration version: %d (dirty: %t)\n", v, dirty)

	case "force":
		if *version == 0 {
			log.Fatal("Version must be specified with -version flag")
		}
		log.Printf("Forcing migration version to %d...", *version)
		err = runner.Force(*version)
		if err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
		log.Println("Migration version forced successfully")

	default:
		log.Fatalf("Unknown action: %s", *action)
	}

	// Show final version
	v, dirty, err := runner.Version()
	if err != nil {
		log.Printf("Warning: could not get final migration version: %v", err)
	} else {
		log.Printf("Final migration version: %d (dirty: %t)", v, dirty)
	}
}

// Example usage:
// go run cmd/migrate/main.go -action=up
// go run cmd/migrate/main.go -action=version
// go run cmd/migrate/main.go -action=steps -steps=1
// go run cmd/migrate/main.go -action=down
