package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// Load populates App from variables
func Load() error {
	env := make(map[string]string)

	// Load from .env if it exists
	if file, err := os.Open(".env"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// remove quotes if any
				if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
					(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
					val = val[1 : len(val)-1]
				}
				env[key] = val
			}
		}
	}

	getEnv := func(key, defaultValue string) string {
		if val, exists := os.LookupEnv(key); exists {
			return val
		}
		if val, exists := env[key]; exists {
			return val
		}
		return defaultValue
	}

	App.Port = getEnv("PORT", "8080")
	App.DatabasePath = getEnv("DATABASE_PATH", "")
	App.BaseAddress = getEnv("BASE_ADDRESS", "localhost")
	App.AppEnv = getEnv("APP_ENV", "development")
	App.AllowedOrigin = getEnv("ALLOWED_ORIGIN", "http://localhost:5173")
	App.MigrationsDir = getEnv("MIGRATIONS_DIR", "./internal/db/migrations")

	if App.DatabasePath == "" {
		return fmt.Errorf("DATABASE_PATH is a required configuration variable")
	}

	return nil
}
