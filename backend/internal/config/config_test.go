package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPrecedence(t *testing.T) {
	chdirTemp(t)
	isolateConfigEnv(t)
	writeDotEnv(t, strings.Join([]string{
		"PORT=7000",
		"DATABASE_PATH=dotenv.db",
		"BASE_ADDRESS=dotenv-host",
		"MIGRATIONS_DIR=dotenv-migrations",
	}, "\n"))

	t.Setenv("PORT", "8000")
	t.Setenv("DATABASE_PATH", "environment.db")
	t.Setenv("BASE_ADDRESS", "environment-host")

	if err := Load(); err != nil {
		t.Fatalf("load returned an error: %v", err)
	}

	// Environment variables take precedence over .env values.
	if App.Port != "8000" {
		t.Errorf("Port = %q, want %q", App.Port, "8000")
	}
	if App.DatabasePath != "environment.db" {
		t.Errorf("DatabasePath = %q, want %q", App.DatabasePath, "environment.db")
	}
	if App.BaseAddress != "environment-host" {
		t.Errorf("BaseAddress = %q, want %q", App.BaseAddress, "environment-host")
	}
	// MIGRATIONS_DIR was only set via .env, so that value should fill in.
	if App.MigrationsDir != "dotenv-migrations" {
		t.Errorf("MigrationsDir = %q, want %q", App.MigrationsDir, "dotenv-migrations")
	}
}

func TestLoadDefaultsBaseAddressToLocalhost(t *testing.T) {
	chdirTemp(t)
	isolateConfigEnv(t)
	t.Setenv("DATABASE_PATH", "app.db")

	if err := Load(); err != nil {
		t.Fatalf("load returned an error: %v", err)
	}

	if App.BaseAddress != "localhost" {
		t.Errorf("BaseAddress = %q, want %q", App.BaseAddress, "localhost")
	}
	if App.AllowedOrigin != "http://localhost:5173" {
		t.Errorf(
			"AllowedOrigin = %q, want %q",
			App.AllowedOrigin,
			"http://localhost:5173",
		)
	}
	if App.MigrationsDir != "./internal/db/migrations" {
		t.Errorf(
			"MigrationsDir = %q, want %q",
			App.MigrationsDir,
			"./internal/db/migrations",
		)
	}
}

func TestLoadReturnsErrorForMissingRequiredConfig(t *testing.T) {
	chdirTemp(t)
	isolateConfigEnv(t)

	err := Load()
	if err == nil {
		t.Fatal("load returned nil, want an error")
	}
	if !strings.Contains(err.Error(), "DATABASE_PATH") {
		t.Errorf("error %q does not mention DATABASE_PATH", err)
	}
}

func chdirTemp(t *testing.T) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func writeDotEnv(t *testing.T, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(".", ".env"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
}

func isolateConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"PORT",
		"DATABASE_PATH",
		"BASE_ADDRESS",
		"APP_ENV",
		"ALLOWED_ORIGIN",
		"MIGRATIONS_DIR",
	} {
		value, ok := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			var err error
			if ok {
				err = os.Setenv(key, value)
			} else {
				err = os.Unsetenv(key)
			}
			if err != nil {
				t.Errorf("restore %s: %v", key, err)
			}
		})
	}
}