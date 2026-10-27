package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"maltiden/internal/api"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"maltiden/pkg/utils"
)

func main() {
	// Set up structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Initialize Sentry if SENTRY_DSN is set (provisioned by Fly extension).
	// Captures unhandled panics and any errors explicitly sent via sentry.CaptureException.
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      os.Getenv("FLY_APP_NAME"),
			Release:          os.Getenv("FLY_MACHINE_VERSION"),
			TracesSampleRate: 0.2,
			SendDefaultPII:   false,
		}); err != nil {
			slog.Error("sentry init failed", "error", err)
		} else {
			slog.Info("sentry initialized")
			defer sentry.Flush(2 * time.Second)
		}
	}

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

	// Wrap with static file serving and SPA fallback,
	// then apply security headers to ALL responses (API + static files)
	handler := middleware.Security(withSPA("./static", router))

	// Wrap with Sentry HTTP middleware so panics and request context flow to Sentry.
	// No-op if SENTRY_DSN was unset (the SDK degrades silently).
	handler = sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(handler)

	// Configure HTTP server with graceful shutdown and timeouts
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   90 * time.Second, // must exceed middleware.ParserTimeout (60s) so TimeoutHandler can write 503 before TCP close
		IdleTimeout:    60 * time.Second,
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
//
// Several paths are both a client route and an API route (/recipes,
// /shopping-list), so the path alone can't decide. A browser loading a page
// asks for text/html; the API client asks for JSON. Page loads get the SPA,
// everything else goes to the API unchanged, so API 404s stay 404s.
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

		if acceptsHTML(r) {
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
			// No frontend built (e.g. local dev) — let the API answer
		}

		apiHandler.ServeHTTP(w, r)
	})
}

// acceptsHTML reports whether the request is a browser page load.
func acceptsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}
