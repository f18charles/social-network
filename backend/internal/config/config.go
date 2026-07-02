package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds process-wide application configuration
type Config struct {
	// Port is the HTTP server's TCP port.
	Port string
	// DatabasePath is the required SQLite database file path.
	DatabasePath string
	// BaseAddress is the HTTP server's host or IP address. 
	BaseAddress string
	// AppEnv identifies the runtime environment. 
	AppEnv string
	// AllowedOrigin is the origin permitted to make cross-origin requests.
	AllowedOrigin string
	// MigrationsDir is the directory containing database migrations.
	MigrationsDir string
}

var App Config

// Load populates App from OS environment variables, using a .env file (if
// present) to fill in anything not already set in the environment, and
// falling back to defaults for anything still unset.
func Load() error {
	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	App = Config{
		Port:          getEnv("PORT", "8080"),
		DatabasePath:  getEnv("DATABASE_PATH", ""),
		BaseAddress:   getEnv("BASE_ADDRESS", "localhost"),
		AppEnv:        getEnv("APP_ENV", "development"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "./internal/db/migrations"),
	}

	if err := checkDatabasePath(App); err != nil {
		return err
	}

	return nil
}

// getEnv returns the environment variable value or the provided fallback.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// checkDatabasePath ensures the database path is always populated
func checkDatabasePath(cfg Config) error {
	if cfg.DatabasePath == "" {
		return errors.New("DATABASE_PATH environment variable is required")
	}
	return nil
}
