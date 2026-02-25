package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	// Wrap with static file serving and SPA fallback
	handler := withSPA("./static", router)

	// Configure HTTP server with graceful shutdown and timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second, // higher to allow Claude API proxy calls
		IdleTimeout:  60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB max header size
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

// withSPA wraps an API handler to also serve frontend static files and
// fall back to index.html for client-side routes (SPA routing).
func withSPA(staticDir string, apiHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Non-GET/HEAD requests always go to the API
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			apiHandler.ServeHTTP(w, r)
			return
		}

		// Try serving a static file (skip root path)
		if r.URL.Path != "/" {
			filePath := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				http.ServeFile(w, r, filePath)
				return
			}
		}

		// Try the API handler with a buffered writer to detect 404s
		buf := &bufferedWriter{header: make(http.Header), code: 200}
		apiHandler.ServeHTTP(buf, r)

		if buf.code != http.StatusNotFound {
			// Forward the API response
			for k, vals := range buf.header {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(buf.code)
			w.Write(buf.body)
			return
		}

		// No static file and no API route matched — serve index.html for SPA
		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			// No frontend built (e.g. local dev) — return plain 404
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}

// bufferedWriter captures an HTTP response so we can inspect the status code
// before deciding whether to forward it or serve the SPA fallback.
type bufferedWriter struct {
	header http.Header
	body   []byte
	code   int
}

func (b *bufferedWriter) Header() http.Header { return b.header }

func (b *bufferedWriter) Write(data []byte) (int, error) {
	b.body = append(b.body, data...)
	return len(data), nil
}

func (b *bufferedWriter) WriteHeader(code int) { b.code = code }
