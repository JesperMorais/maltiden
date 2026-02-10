package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"maltiden/internal/api"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
)

func main() {
	// Set up structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Create JWT service with validated secret
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtService, err := utils.NewJWTService(jwtSecret)
	if err != nil {
		slog.Error("JWT_SECRET must be at least 32 characters")
		os.Exit(1)
	}

	// Get config from env variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/maltiden.db"
	}

	// Open db (runs migrations)
	db, err := sqlite.Open(dbPath)
	if err != nil {
		slog.Error("database error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create router (injects db and jwtService)
	router := api.NewRouter(db, jwtService)

	// Configure HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("server shutting down")

	// Graceful shutdown with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
