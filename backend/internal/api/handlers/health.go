package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"
)

type HealthHandler struct {
	db      *sql.DB
	version string
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db, version: readBuildVersion()}
}

// readBuildVersion returns the VCS revision embedded by `go build` when run
// inside a git checkout, or "unknown" if unavailable (e.g. -buildvcs=false,
// `go test` without VCS info).
func readBuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			return s.Value
		}
	}
	return "unknown"
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity with 2-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		body, _ := json.Marshal(map[string]string{
			"status":  "error",
			"db":      "disconnected",
			"version": h.version,
		})
		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusOK)
	body, _ := json.Marshal(map[string]string{
		"status":  "ok",
		"db":      "connected",
		"version": h.version,
	})
	w.Write(body)
}
