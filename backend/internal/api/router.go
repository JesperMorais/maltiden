package api

import (
	"maltiden/internal/api/handlers"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)

	return mux
}
