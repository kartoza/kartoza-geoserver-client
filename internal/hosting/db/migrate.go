package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration.
type Migration struct {
	Version   int
	Name      string
	Direction string // "up" or "down"
	SQL       string
}

// MigrateUp runs all pending migrations.
func (db *DB) MigrateUp() error {
	if err := db.ensureMigrationsTable(); err != nil {
		return err
	}

	migrations, err := loadMigrations("up")
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		log.Printf("Applying migration %d: %s", m.Version, m.Name)
		if err := db.runMigration(m); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", m.Version, err)
		}
	}

	return nil
}

// MigrateDown rolls back the last migration.
func (db *DB) MigrateDown() error {
	if err := db.ensureMigrationsTable(); err != nil {
		return err
	}

	// Get the last applied migration
	var version int
	var name string
	err := db.QueryRow(`
		SELECT version, name FROM schema_migrations
		ORDER BY version DESC LIMIT 1
	`).Scan(&version, &name)
	if err == sql.ErrNoRows {
		log.Println("No migrations to roll back")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	migrations, err := loadMigrations("down")
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, m := range migrations {
		if m.Version == version {
			log.Printf("Rolling back migration %d: %s", m.Version, m.Name)
			if err := db.rollbackMigration(m); err != nil {
				return fmt.Errorf("failed to rollback migration %d: %w", m.Version, err)
			}
			return nil
		}
	}

	return fmt.Errorf("down migration not found for version %d", version)
}

// MigrateReset rolls back all migrations and reapplies them.
func (db *DB) MigrateReset() error {
	// Roll back all
	for {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
			break
		}
		if count == 0 {
			break
		}
		if err := db.MigrateDown(); err != nil {
			return err
		}
	}

	// Apply all
	return db.MigrateUp()
}

// MigrationStatus returns the status of all migrations.
func (db *DB) MigrationStatus() ([]MigrationStatusEntry, error) {
	if err := db.ensureMigrationsTable(); err != nil {
		return nil, err
	}

	migrations, err := loadMigrations("up")
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := db.getAppliedMigrationsWithTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var status []MigrationStatusEntry
	for _, m := range migrations {
		entry := MigrationStatusEntry{
			Version: m.Version,
			Name:    m.Name,
			Applied: applied[m.Version] != nil,
		}
		if entry.Applied {
			entry.AppliedAt = applied[m.Version]
		}
		status = append(status, entry)
	}

	return status, nil
}

// MigrationStatusEntry represents the status of a single migration.
type MigrationStatusEntry struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

func (db *DB) ensureMigrationsTable() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (db *DB) getAppliedMigrations() (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func (db *DB) getAppliedMigrationsWithTime() (map[int]*time.Time, error) {
	rows, err := db.Query(`SELECT version, applied_at FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]*time.Time)
	for rows.Next() {
		var version int
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = &appliedAt
	}
	return applied, rows.Err()
}

func (db *DB) runMigration(m Migration) error {
	return db.Transaction(func(tx *sql.Tx) error {
		if _, err := tx.Exec(m.SQL); err != nil {
			return err
		}
		_, err := tx.Exec(`
			INSERT INTO schema_migrations (version, name) VALUES ($1, $2)
		`, m.Version, m.Name)
		return err
	})
}

func (db *DB) rollbackMigration(m Migration) error {
	return db.Transaction(func(tx *sql.Tx) error {
		if _, err := tx.Exec(m.SQL); err != nil {
			return err
		}
		_, err := tx.Exec(`DELETE FROM schema_migrations WHERE version = $1`, m.Version)
		return err
	})
}

func loadMigrations(direction string) ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	suffix := "." + direction + ".sql"
	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}

		content, err := migrationsFS.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		// Parse filename: 001_initial.up.sql -> version=1, name=initial
		name := strings.TrimSuffix(entry.Name(), suffix)
		parts := strings.SplitN(name, "_", 2)
		if len(parts) != 2 {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		migrations = append(migrations, Migration{
			Version:   version,
			Name:      parts[1],
			Direction: direction,
			SQL:       string(content),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}
