package api

import (
	"database/sql"
	"maltiden/internal/api/handlers"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"net/http"
)

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Setup dependencies
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	authService := services.NewAuthService(userStorage, householdStorage)
	authHandler := handlers.NewAuthHandler(authService)
	householdHandler := handlers.NewHouseholdHandler(householdStorage)

	// Public routes
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protected routes
	mux.Handle("GET /households/me", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.GetMyHousehold),
	))

	// Wrap with CORS middleware for frontend development
	return middleware.CORS(mux)
}
