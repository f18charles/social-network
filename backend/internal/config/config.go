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
	// CloudinaryURL is the Cloudinary account URL in the form
	// cloudinary://<api_key>:<api_secret>@<cloud_name>. When empty, uploaded
	// images and GIFs fall back to local disk storage.
	CloudinaryURL string
	// TursoDatabaseURL is the remote libSQL/Turso connection URL, e.g.
	// libsql://<db-name>-<org>.turso.io. When empty, the backend falls back
	// to local SQLite at DatabasePath.
	TursoDatabaseURL string
	// TursoAuthToken is the auth token for the Turso database. Required
	// whenever TursoDatabaseURL is set.
	TursoAuthToken string
}

// UsesTurso reports whether the app is configured to connect to a remote
// Turso database instead of local SQLite.
func (c *Config) UsesTurso() bool {
	return c.TursoDatabaseURL != ""
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
	App.CloudinaryURL = getEnv("CLOUDINARY_URL", "")
	App.TursoDatabaseURL = getEnv("TURSO_DATABASE_URL", "")
	App.TursoAuthToken = getEnv("TURSO_AUTH_TOKEN", "")

	if App.TursoDatabaseURL == "" && App.DatabasePath == "" {
		return fmt.Errorf("DATABASE_PATH is a required configuration variable when TURSO_DATABASE_URL is not set")
	}
	if App.TursoDatabaseURL != "" && App.TursoAuthToken == "" {
		return fmt.Errorf("TURSO_AUTH_TOKEN is required when TURSO_DATABASE_URL is set")
	}

	return nil
}

// IsProduction reports whether the application is running in production mode.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

// IsOriginAllowed checks if the given origin matches the allowed origins.
// It supports comma-separated origins (e.g. "http://localhost:5173,https://my-app.vercel.app")
// as well as wildcard "*".
func (c *Config) IsOriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	if c.AllowedOrigin == "*" {
		return true
	}
	for _, allowed := range strings.Split(c.AllowedOrigin, ",") {
		trimmed := strings.TrimSpace(allowed)
		if trimmed == "" {
			continue
		}
		if strings.EqualFold(trimmed, origin) || trimmed == "*" {
			return true
		}
	}
	return false
}
