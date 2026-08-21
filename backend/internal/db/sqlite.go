package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// InitDB initializes the database (local SQLite or remote Turso/libSQL),
// creates required schemas, and applies migrations.
//
// When tursoURL is non-empty, it connects to that remote Turso database
// using the "libsql" driver and tursoAuthToken for authentication, and
// dbPath/directory creation are skipped entirely. Otherwise it falls back
// to local SQLite at dbPath, exactly as before.
func InitDB(dbPath string, migrationsDir string, tursoURL string, tursoAuthToken string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	if tursoURL != "" {
		db, err = sql.Open("libsql", libsqlDSN(tursoURL, tursoAuthToken))
		if err != nil {
			return nil, fmt.Errorf("failed to open turso database: %w", err)
		}
	} else {
		// Ensure directory for dbPath exists if it is in a subdirectory
		dbDir := filepath.Dir(dbPath)
		if dbDir != "." && dbDir != "/" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create db directory: %w", err)
			}
		}

		db, err = sql.Open("sqlite3", sqliteDSN(dbPath))
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := runMigrations(db, migrationsDir); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func sqliteDSN(dbPath string) string {
	sep := "?"
	if strings.Contains(dbPath, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%s_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath, sep)
}

// libsqlDSN builds the DSN for a remote Turso/libSQL connection. The auth
// token is passed as a query parameter, per the libsql-client-go driver.
func libsqlDSN(tursoURL string, authToken string) string {
	sep := "?"
	if strings.Contains(tursoURL, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%sauthToken=%s", tursoURL, sep, authToken)
}

func runMigrations(db *sql.DB, migrationsDir string) error {
	// Create migration tracking table if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory %s: %w", migrationsDir, err)
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort files to run them in order (e.g. 000001 before 000002)
	sort.Strings(migrationFiles)

	for _, filename := range migrationFiles {
		var version int
		_, err := fmt.Sscanf(filename, "%d_", &version)
		if err != nil {
			log.Printf("Warning: skipping invalid migration filename %s: %v", filename, err)
			continue
		}

		// Check if already applied
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration state for version %d: %w", version, err)
		}

		if exists {
			continue
		}

		// Read and execute migration
		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		log.Printf("Applying migration: %s", filename)

		// Execute queries in transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", version, filename); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to log migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration transaction %s: %w", filename, err)
		}
	}

	return nil
}
