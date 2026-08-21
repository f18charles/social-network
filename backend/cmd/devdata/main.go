package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/config"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/db"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/devdata"
)

func main() {
	// Disable default log prefixes (date/time) for cleaner CLI output
	log.SetFlags(0)

	// Ensure a sub-command is provided
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	// Extract the sub-command (e.g., "seed")
	command := os.Args[1]

	// Load application configuration from environment
	if err := config.Load(); err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Initialize the database connection
	database, err := db.InitDB(config.App.DatabasePath, config.App.MigrationsDir, config.App.TursoDatabaseURL, config.App.TursoAuthToken)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	defer database.Close()

	// Prepare options shared across devdata operations
	opts := devdata.Options{
		DB:     database,
		AppEnv: config.App.AppEnv,
	}

	// Execute the requested development command
	switch command {
	case "seed":
		if err := devdata.Seed(opts); err != nil {
			log.Fatalf("seed fixture data: %v", err)
		}
		printStatus(database, "Seeded")
	case "teardown":
		if err := devdata.Teardown(opts); err != nil {
			log.Fatalf("teardown fixture data: %v", err)
		}
		printStatus(database, "Removed")
	case "status":
		printStatus(database, "Current")
	default:
		usage()
		os.Exit(2)
	}
}

// printStatus queries the database for current row counts and displays them
func printStatus(database *sql.DB, label string) {
	status, err := devdata.Inspect(database)
	if err != nil {
		log.Fatalf("inspect fixture data: %v", err)
	}
	fmt.Printf("%s E2E fixture data: users=%d posts=%d comments=%d\n", label, status.Users, status.Posts, status.Comments)
}

// usage prints the correct command syntax to stderr
func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/devdata [seed|teardown|status]")
}
