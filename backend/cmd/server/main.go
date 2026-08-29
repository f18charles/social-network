package main

import (
	"log"
	"net"
	"net/http"

	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/routers"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/config"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/db"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/logger"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/storage"
)

func main() {
	// Initialize structured logger
	logger.Init(logger.Config{
		Level: logger.LevelInfo,
		JSON:  false,
	})

	// Configuration (using env variables or defaults)
	if err := config.Load(); err != nil {
		logger.Error("Failed to load configuration", "error", err)
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Route image/GIF uploads to Cloudinary when configured; otherwise
	// storage falls back to local disk automatically.
	if err := storage.Configure(config.App.CloudinaryURL); err != nil {
		log.Fatalf("Failed to configure image storage: %v", err)
	}
	if config.App.CloudinaryURL != "" {
		log.Println("Image storage: Cloudinary")
	} else {
		log.Println("Image storage: local disk (set CLOUDINARY_URL to use Cloudinary)")
	}

	// Initialize/load database
	if config.App.UsesTurso() {
		log.Printf("Initializing Turso database at %s...", config.App.TursoDatabaseURL)
	} else {
		log.Printf("Initializing database at %s...", config.App.DatabasePath)
	}
	database, err := db.InitDB(config.App.DatabasePath, config.App.MigrationsDir, config.App.TursoDatabaseURL, config.App.TursoAuthToken)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	log.Println("Database initialized and migrations applied successfully!")

	// Register Routes
	handler := routers.Router(database)

	address := net.JoinHostPort(config.App.BaseAddress, config.App.Port)
	log.Printf("Server starting on %s...", address)
	if err := http.ListenAndServe(address, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
