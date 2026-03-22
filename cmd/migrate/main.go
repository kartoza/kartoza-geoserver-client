// Package main provides a CLI tool for running database migrations.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	database, err := db.NewFromDSN(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	command := os.Args[1]

	switch command {
	case "up":
		fmt.Println("Running migrations...")
		if err := database.MigrateUp(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		fmt.Println("Rolling back last migration...")
		if err := database.MigrateDown(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("Rollback completed successfully!")

	case "reset":
		fmt.Println("Resetting database (rollback all and reapply)...")
		if err := database.MigrateReset(); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("Reset completed successfully!")

	case "status":
		fmt.Println("Migration status:")
		fmt.Println()
		status, err := database.MigrationStatus()
		if err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}
		for _, entry := range status {
			applied := "[ ]"
			appliedAt := ""
			if entry.Applied {
				applied = "[x]"
				if entry.AppliedAt != nil {
					appliedAt = fmt.Sprintf(" (applied: %s)", entry.AppliedAt.Format("2006-01-02 15:04:05"))
				}
			}
			fmt.Printf("%s %03d_%s%s\n", applied, entry.Version, entry.Name, appliedAt)
		}

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up      Run all pending migrations")
	fmt.Println("  down    Rollback the last migration")
	fmt.Println("  reset   Rollback all and reapply migrations")
	fmt.Println("  status  Show migration status")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  DATABASE_URL  PostgreSQL connection string (required)")
}
