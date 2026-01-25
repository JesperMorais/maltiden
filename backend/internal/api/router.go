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
	recipeStorage := sqlite.NewRecipeStorage(db)

	authService := services.NewAuthService(userStorage, householdStorage)
	recipeService := services.NewRecipeService(recipeStorage)

	authHandler := handlers.NewAuthHandler(authService)
	householdHandler := handlers.NewHouseholdHandler(householdStorage)
	recipeHandler := handlers.NewRecipeHandler(recipeService)

	// Tjek API service (POC)
	tjekService := services.NewTjekService()
	offersHandler := handlers.NewOffersHandler(tjekService)

	// Public routes
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("GET /offers/search", offersHandler.SearchOffers)
	mux.HandleFunc("GET /offers/discounts", offersHandler.GetDiscounts)
	mux.HandleFunc("GET /offers/stores", offersHandler.GetStores)
	mux.HandleFunc("GET /recipes", recipeHandler.GetAll)
	mux.HandleFunc("GET /recipes/{id}", recipeHandler.GetByID)

	// Protected routes
	mux.Handle("GET /households/me", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.GetMyHousehold),
	))
	mux.Handle("POST /recipes", middleware.RequireAuth(
		http.HandlerFunc(recipeHandler.Create),
	))

	// Wrap with CORS middleware for frontend development
	return middleware.CORS(mux)
}
