package config

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

// Load populates App fromiables
func Load() error {
	App = Config{
		Port:          "8080",
		DatabasePath:  "./db.sqlite",
		BaseAddress:   "localhost",
		AppEnv:        "development",
		AllowedOrigin: "http://localhost:5173",
		MigrationsDir: "./internal/db/migrations",
	}

	return nil
}
