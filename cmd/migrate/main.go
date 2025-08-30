package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"go-templ-template/internal/config"
	"go-templ-template/internal/shared/database"
)

func main() {
	var (
		action         = flag.String("action", "up", "Migration action: up, down, steps, version, force, status, list, create, validate, to")
		steps          = flag.Int("steps", 0, "Number of steps for 'steps' action")
		version        = flag.Int("version", 0, "Version for 'force' or 'to' action")
		migrationsPath = flag.String("migrations", "./migrations", "Path to migrations directory")
		name           = flag.String("name", "", "Name for 'create' action")
		format         = flag.String("format", "text", "Output format: text, json")
		verbose        = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Execute the requested action
	switch *action {
	case "up":
		runMigrationsUp(cfg, *migrationsPath, *verbose)
	case "down":
		runMigrationsDown(cfg, *migrationsPath, *verbose)
	case "steps":
		runMigrationSteps(cfg, *migrationsPath, *steps, *verbose)
	case "version":
		showVersion(cfg, *migrationsPath, *format)
	case "force":
		forceMigrationVersion(cfg, *migrationsPath, *version, *verbose)
	case "status":
		showStatus(cfg, *migrationsPath, *format)
	case "list":
		listMigrations(cfg, *migrationsPath, *format)
	case "create":
		createMigration(cfg, *migrationsPath, *name, *verbose)
	case "validate":
		validateMigrations(cfg, *migrationsPath, *verbose)
	case "to":
		migrateTo(cfg, *migrationsPath, *version, *verbose)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func runMigrationsUp(cfg *config.Config, migrationsPath string, verbose bool) {
	if verbose {
		log.Println("Creating migration runner...")
	}

	runner, err := database.NewMigrationRunner(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	log.Println("Running migrations up...")
	err = runner.Up()
	if err != nil {
		log.Fatalf("Failed to run migrations up: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Show final version
	showFinalVersion(runner, verbose)
}

func runMigrationsDown(cfg *config.Config, migrationsPath string, verbose bool) {
	if verbose {
		log.Println("Creating migration runner...")
	}

	runner, err := database.NewMigrationRunner(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	log.Println("Running migrations down...")
	err = runner.Down()
	if err != nil {
		log.Fatalf("Failed to run migrations down: %v", err)
	}
	log.Println("Migrations rolled back successfully")

	// Show final version
	showFinalVersion(runner, verbose)
}

func runMigrationSteps(cfg *config.Config, migrationsPath string, steps int, verbose bool) {
	if steps == 0 {
		log.Fatal("Steps count must be specified with -steps flag")
	}

	if verbose {
		log.Println("Creating migration runner...")
	}

	runner, err := database.NewMigrationRunner(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	direction := "up"
	if steps < 0 {
		direction = "down"
	}

	log.Printf("Running %d migration steps %s...", abs(steps), direction)
	err = runner.Steps(steps)
	if err != nil {
		log.Fatalf("Failed to run migration steps: %v", err)
	}
	log.Println("Migration steps completed successfully")

	// Show final version
	showFinalVersion(runner, verbose)
}

func showVersion(cfg *config.Config, migrationsPath string, format string) {
	runner, err := database.NewMigrationRunner(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	v, dirty, err := runner.Version()
	if err != nil {
		log.Fatalf("Failed to get migration version: %v", err)
	}

	if format == "json" {
		result := map[string]interface{}{
			"version": v,
			"dirty":   dirty,
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		fmt.Printf("Current migration version: %d (dirty: %t)\n", v, dirty)
	}
}

func forceMigrationVersion(cfg *config.Config, migrationsPath string, version int, verbose bool) {
	if version < 0 {
		log.Fatal("Version must be non-negative")
	}

	if verbose {
		log.Println("Creating migration runner...")
	}

	runner, err := database.NewMigrationRunner(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	log.Printf("Forcing migration version to %d...", version)
	err = runner.Force(version)
	if err != nil {
		log.Fatalf("Failed to force migration version: %v", err)
	}
	log.Println("Migration version forced successfully")

	// Show final version
	showFinalVersion(runner, verbose)
}

func showStatus(cfg *config.Config, migrationsPath string, format string) {
	manager, err := database.NewMigrationManager(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer manager.Close()

	status, err := manager.GetStatus()
	if err != nil {
		log.Fatalf("Failed to get migration status: %v", err)
	}

	if format == "json" {
		json.NewEncoder(os.Stdout).Encode(status)
	} else {
		fmt.Println("Migration Status:")
		fmt.Printf("  Current Version: %d\n", status.CurrentVersion)
		fmt.Printf("  Latest Version:  %d\n", status.LatestVersion)
		fmt.Printf("  Is Dirty:        %t\n", status.IsDirty)
		fmt.Printf("  Applied:         %d/%d migrations\n", status.AppliedCount, status.TotalMigrations)
		fmt.Printf("  Pending:         %d migrations\n", status.PendingCount)
		fmt.Printf("  Status:          %s\n", status.String())

		if len(status.PendingMigrations) > 0 {
			fmt.Println("\nPending Migrations:")
			for _, migration := range status.PendingMigrations {
				fmt.Printf("  - %03d: %s\n", migration.Version, migration.Name)
			}
		}
	}
}

func listMigrations(cfg *config.Config, migrationsPath string, format string) {
	manager, err := database.NewMigrationManager(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer manager.Close()

	migrations, err := manager.ListMigrations()
	if err != nil {
		log.Fatalf("Failed to list migrations: %v", err)
	}

	currentVersion, _, err := manager.GetCurrentVersion()
	if err != nil {
		log.Printf("Warning: could not get current version: %v", err)
	}

	if format == "json" {
		result := map[string]interface{}{
			"migrations":      migrations,
			"current_version": currentVersion,
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		fmt.Printf("Available Migrations (current version: %d):\n", currentVersion)
		for _, migration := range migrations {
			status := "pending"
			if migration.Version <= currentVersion {
				status = "applied"
			}
			fmt.Printf("  %03d: %-30s [%s]\n", migration.Version, migration.Name, status)
		}
	}
}

func createMigration(cfg *config.Config, migrationsPath string, name string, verbose bool) {
	if name == "" {
		log.Fatal("Migration name must be specified with -name flag")
	}

	if verbose {
		log.Println("Creating migration manager...")
	}

	manager, err := database.NewMigrationManager(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer manager.Close()

	migration, err := manager.CreateMigration(name)
	if err != nil {
		log.Fatalf("Failed to create migration: %v", err)
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  Up:   %s\n", migration.UpFile)
	fmt.Printf("  Down: %s\n", migration.DownFile)
	fmt.Printf("Version: %d\n", migration.Version)

	if verbose {
		fmt.Printf("Migration files created in: %s\n", migrationsPath)
	}
}

func validateMigrations(cfg *config.Config, migrationsPath string, verbose bool) {
	if verbose {
		log.Println("Creating migration manager...")
	}

	manager, err := database.NewMigrationManager(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer manager.Close()

	if verbose {
		log.Println("Validating migrations...")
	}

	err = manager.ValidateMigrations()
	if err != nil {
		log.Fatalf("Migration validation failed: %v", err)
	}

	fmt.Println("All migrations are valid")

	if verbose {
		migrations, err := manager.ListMigrations()
		if err != nil {
			log.Printf("Warning: could not list migrations: %v", err)
		} else {
			fmt.Printf("Validated %d migrations\n", len(migrations))
		}
	}
}

func migrateTo(cfg *config.Config, migrationsPath string, targetVersion int, verbose bool) {
	if targetVersion < 0 {
		log.Fatal("Target version must be non-negative")
	}

	if verbose {
		log.Println("Creating migration manager...")
	}

	manager, err := database.NewMigrationManager(&cfg.Database, migrationsPath)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer manager.Close()

	currentVersion, _, err := manager.GetCurrentVersion()
	if err != nil {
		log.Fatalf("Failed to get current version: %v", err)
	}

	if uint(targetVersion) == currentVersion {
		fmt.Printf("Already at target version %d\n", targetVersion)
		return
	}

	direction := "up"
	if uint(targetVersion) < currentVersion {
		direction = "down"
	}

	log.Printf("Migrating %s from version %d to %d...", direction, currentVersion, targetVersion)
	err = manager.MigrateTo(uint(targetVersion))
	if err != nil {
		log.Fatalf("Failed to migrate to version %d: %v", targetVersion, err)
	}

	fmt.Printf("Successfully migrated to version %d\n", targetVersion)
}

func showFinalVersion(runner *database.MigrationRunner, verbose bool) {
	v, dirty, err := runner.Version()
	if err != nil {
		if verbose {
			log.Printf("Warning: could not get final migration version: %v", err)
		}
	} else {
		if verbose {
			log.Printf("Final migration version: %d (dirty: %t)", v, dirty)
		} else {
			fmt.Printf("Current version: %d\n", v)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Migration tool for Go Templ Template\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Actions:\n")
		fmt.Fprintf(os.Stderr, "  up        - Run all pending migrations\n")
		fmt.Fprintf(os.Stderr, "  down      - Rollback all migrations\n")
		fmt.Fprintf(os.Stderr, "  steps     - Run N migration steps (use -steps flag)\n")
		fmt.Fprintf(os.Stderr, "  version   - Show current migration version\n")
		fmt.Fprintf(os.Stderr, "  force     - Force migration to specific version (use -version flag)\n")
		fmt.Fprintf(os.Stderr, "  status    - Show detailed migration status\n")
		fmt.Fprintf(os.Stderr, "  list      - List all available migrations\n")
		fmt.Fprintf(os.Stderr, "  create    - Create new migration files (use -name flag)\n")
		fmt.Fprintf(os.Stderr, "  validate  - Validate all migration files\n")
		fmt.Fprintf(os.Stderr, "  to        - Migrate to specific version (use -version flag)\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -action=up\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -action=status -format=json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -action=create -name=\"add user roles\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -action=steps -steps=2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -action=to -version=3\n", os.Args[0])
	}
}
