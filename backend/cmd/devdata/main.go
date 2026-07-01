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
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	command := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	if err := config.Load(); err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.InitDB(config.App.DatabasePath, config.App.MigrationsDir)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	defer database.Close()

	opts := devdata.Options{
		DB:     database,
		AppEnv: config.App.AppEnv,
	}

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

func printStatus(database *sql.DB, label string) {
	status, err := devdata.Inspect(database)
	if err != nil {
		log.Fatalf("inspect fixture data: %v", err)
	}
	fmt.Printf("%s E2E fixture data: users=%d posts=%d comments=%d\n", label, status.Users, status.Posts, status.Comments)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/devdata [seed|teardown|status]")
}
