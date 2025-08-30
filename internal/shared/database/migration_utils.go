package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-templ-template/internal/config"
)

// MigrationInfo holds information about a migration
type MigrationInfo struct {
	Version     uint
	Name        string
	UpFile      string
	DownFile    string
	Description string
}

// MigrationManager provides utilities for managing migrations
type MigrationManager struct {
	migrationsPath string
	runner         *MigrationRunner
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(cfg *config.DatabaseConfig, migrationsPath string) (*MigrationManager, error) {
	runner, err := NewMigrationRunner(cfg, migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration runner: %w", err)
	}

	return &MigrationManager{
		migrationsPath: migrationsPath,
		runner:         runner,
	}, nil
}

// Close closes the migration manager
func (mm *MigrationManager) Close() error {
	if mm.runner != nil {
		return mm.runner.Close()
	}
	return nil
}

// CreateMigration creates a new migration file pair (up and down)
func (mm *MigrationManager) CreateMigration(name string) (*MigrationInfo, error) {
	// Validate name
	if name == "" {
		return nil, fmt.Errorf("migration name cannot be empty")
	}

	// Clean the name (remove spaces, special characters)
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "-", "_")

	// Get next version number
	version, err := mm.getNextVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get next version: %w", err)
	}

	// Create file names
	versionStr := fmt.Sprintf("%03d", version)
	upFile := fmt.Sprintf("%s_%s.up.sql", versionStr, cleanName)
	downFile := fmt.Sprintf("%s_%s.down.sql", versionStr, cleanName)

	upPath := filepath.Join(mm.migrationsPath, upFile)
	downPath := filepath.Join(mm.migrationsPath, downFile)

	// Check if files already exist
	if _, err := os.Stat(upPath); err == nil {
		return nil, fmt.Errorf("migration file already exists: %s", upFile)
	}
	if _, err := os.Stat(downPath); err == nil {
		return nil, fmt.Errorf("migration file already exists: %s", downFile)
	}

	// Create up migration file
	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: %s

-- Add your up migration SQL here

`, name, time.Now().Format("2006-01-02 15:04:05"), name)

	err = os.WriteFile(upPath, []byte(upContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downContent := fmt.Sprintf(`-- Migration rollback: %s
-- Created: %s
-- Description: Rollback for %s

-- Add your down migration SQL here

`, name, time.Now().Format("2006-01-02 15:04:05"), name)

	err = os.WriteFile(downPath, []byte(downContent), 0644)
	if err != nil {
		// Clean up up file if down file creation fails
		os.Remove(upPath)
		return nil, fmt.Errorf("failed to create down migration file: %w", err)
	}

	return &MigrationInfo{
		Version:     version,
		Name:        cleanName,
		UpFile:      upFile,
		DownFile:    downFile,
		Description: name,
	}, nil
}

// ListMigrations lists all available migrations
func (mm *MigrationManager) ListMigrations() ([]MigrationInfo, error) {
	files, err := os.ReadDir(mm.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrations := make(map[uint]*MigrationInfo)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse migration file name
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		versionStr := parts[0]
		version, err := strconv.ParseUint(versionStr, 10, 32)
		if err != nil {
			continue
		}

		v := uint(version)

		// Initialize migration info if not exists
		if migrations[v] == nil {
			migrations[v] = &MigrationInfo{
				Version: v,
			}
		}

		migration := migrations[v]

		if strings.HasSuffix(name, ".up.sql") {
			migration.UpFile = name
			// Extract name from filename
			nameParts := parts[1:]
			if len(nameParts) > 0 {
				lastName := nameParts[len(nameParts)-1]
				if strings.HasSuffix(lastName, ".up.sql") {
					nameParts[len(nameParts)-1] = strings.TrimSuffix(lastName, ".up.sql")
				}
			}
			migration.Name = strings.Join(nameParts, "_")
		} else if strings.HasSuffix(name, ".down.sql") {
			migration.DownFile = name
		}
	}

	// Convert map to slice and sort
	result := make([]MigrationInfo, 0, len(migrations))
	for _, migration := range migrations {
		// Only include migrations that have both up and down files
		if migration.UpFile != "" && migration.DownFile != "" {
			result = append(result, *migration)
		}
	}

	// Sort by version
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Version > result[j].Version {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// GetCurrentVersion returns the current migration version
func (mm *MigrationManager) GetCurrentVersion() (uint, bool, error) {
	return mm.runner.Version()
}

// GetStatus returns the migration status
func (mm *MigrationManager) GetStatus() (*MigrationStatus, error) {
	currentVersion, dirty, err := mm.runner.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	migrations, err := mm.ListMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %w", err)
	}

	var latestVersion uint
	if len(migrations) > 0 {
		latestVersion = migrations[len(migrations)-1].Version
	}

	appliedMigrations := make([]MigrationInfo, 0)
	pendingMigrations := make([]MigrationInfo, 0)

	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			appliedMigrations = append(appliedMigrations, migration)
		} else {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	return &MigrationStatus{
		CurrentVersion:    currentVersion,
		LatestVersion:     latestVersion,
		IsDirty:           dirty,
		AppliedMigrations: appliedMigrations,
		PendingMigrations: pendingMigrations,
		TotalMigrations:   len(migrations),
		AppliedCount:      len(appliedMigrations),
		PendingCount:      len(pendingMigrations),
	}, nil
}

// MigrateUp runs migrations up
func (mm *MigrationManager) MigrateUp() error {
	return mm.runner.Up()
}

// MigrateDown runs migrations down
func (mm *MigrationManager) MigrateDown() error {
	return mm.runner.Down()
}

// MigrateSteps runs n migration steps
func (mm *MigrationManager) MigrateSteps(n int) error {
	return mm.runner.Steps(n)
}

// MigrateTo migrates to a specific version
func (mm *MigrationManager) MigrateTo(targetVersion uint) error {
	currentVersion, _, err := mm.runner.Version()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion == currentVersion {
		return nil // Already at target version
	}

	if targetVersion > currentVersion {
		// Migrate up
		steps := int(targetVersion - currentVersion)
		return mm.runner.Steps(steps)
	} else {
		// Migrate down
		steps := -int(currentVersion - targetVersion)
		return mm.runner.Steps(steps)
	}
}

// Force forces the migration version
func (mm *MigrationManager) Force(version int) error {
	return mm.runner.Force(version)
}

// ValidateMigrations validates all migration files
func (mm *MigrationManager) ValidateMigrations() error {
	migrations, err := mm.ListMigrations()
	if err != nil {
		return fmt.Errorf("failed to list migrations: %w", err)
	}

	// Check for gaps in version numbers
	for i, migration := range migrations {
		expectedVersion := uint(i + 1)
		if migration.Version != expectedVersion {
			return fmt.Errorf("migration version gap detected: expected %d, found %d", expectedVersion, migration.Version)
		}

		// Check if files exist
		upPath := filepath.Join(mm.migrationsPath, migration.UpFile)
		downPath := filepath.Join(mm.migrationsPath, migration.DownFile)

		if _, err := os.Stat(upPath); os.IsNotExist(err) {
			return fmt.Errorf("up migration file missing: %s", migration.UpFile)
		}

		if _, err := os.Stat(downPath); os.IsNotExist(err) {
			return fmt.Errorf("down migration file missing: %s", migration.DownFile)
		}

		// Check if files are not empty
		upContent, err := os.ReadFile(upPath)
		if err != nil {
			return fmt.Errorf("failed to read up migration file %s: %w", migration.UpFile, err)
		}

		downContent, err := os.ReadFile(downPath)
		if err != nil {
			return fmt.Errorf("failed to read down migration file %s: %w", migration.DownFile, err)
		}

		// Check for basic SQL content (not just comments)
		if !containsSQL(string(upContent)) {
			return fmt.Errorf("up migration file %s appears to be empty or contains no SQL", migration.UpFile)
		}

		if !containsSQL(string(downContent)) {
			return fmt.Errorf("down migration file %s appears to be empty or contains no SQL", migration.DownFile)
		}
	}

	return nil
}

// getNextVersion determines the next version number for a new migration
func (mm *MigrationManager) getNextVersion() (uint, error) {
	migrations, err := mm.ListMigrations()
	if err != nil {
		return 0, err
	}

	if len(migrations) == 0 {
		return 1, nil
	}

	// Return the highest version + 1
	highestVersion := migrations[len(migrations)-1].Version
	return highestVersion + 1, nil
}

// containsSQL checks if the content contains actual SQL statements
func containsSQL(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		// If we find a non-comment, non-empty line, assume it's SQL
		return true
	}
	return false
}

// MigrationStatus holds the current migration status
type MigrationStatus struct {
	CurrentVersion    uint            `json:"current_version"`
	LatestVersion     uint            `json:"latest_version"`
	IsDirty           bool            `json:"is_dirty"`
	AppliedMigrations []MigrationInfo `json:"applied_migrations"`
	PendingMigrations []MigrationInfo `json:"pending_migrations"`
	TotalMigrations   int             `json:"total_migrations"`
	AppliedCount      int             `json:"applied_count"`
	PendingCount      int             `json:"pending_count"`
}

// IsUpToDate returns true if all migrations are applied
func (ms *MigrationStatus) IsUpToDate() bool {
	return ms.CurrentVersion == ms.LatestVersion && !ms.IsDirty
}

// NeedsMigration returns true if there are pending migrations
func (ms *MigrationStatus) NeedsMigration() bool {
	return ms.PendingCount > 0 || ms.IsDirty
}

// String returns a human-readable status string
func (ms *MigrationStatus) String() string {
	if ms.IsDirty {
		return fmt.Sprintf("Migration state is dirty at version %d", ms.CurrentVersion)
	}

	if ms.IsUpToDate() {
		return fmt.Sprintf("All migrations applied (version %d)", ms.CurrentVersion)
	}

	return fmt.Sprintf("Migration needed: %d/%d applied (current: %d, latest: %d)",
		ms.AppliedCount, ms.TotalMigrations, ms.CurrentVersion, ms.LatestVersion)
}
