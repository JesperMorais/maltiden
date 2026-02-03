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
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)

	authService := services.NewAuthService(userStorage, householdStorage)
	householdService := services.NewHouseholdService(householdStorage, userStorage)
	recipeService := services.NewRecipeService(recipeStorage)
	menuService := services.NewMenuService(menuStorage, recipeStorage)
	shoppingService := services.NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	authHandler := handlers.NewAuthHandler(authService)
	householdHandler := handlers.NewHouseholdHandler(householdService)
	recipeHandler := handlers.NewRecipeHandler(recipeService)
	menuHandler := handlers.NewMenuHandler(menuService)
	shoppingHandler := handlers.NewShoppingHandler(shoppingService)

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
	mux.Handle("POST /households/invite", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.CreateInvite),
	))
	mux.Handle("POST /households/join", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.JoinHousehold),
	))
	mux.Handle("GET /households/members/status", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.GetMemberStatuses),
	))
	mux.Handle("PATCH /households/members/{id}/status", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.UpdateMemberStatus),
	))
	mux.Handle("DELETE /households/members/{id}", middleware.RequireAuth(
		http.HandlerFunc(householdHandler.RemoveMember),
	))
	mux.Handle("POST /recipes", middleware.RequireAuth(
		http.HandlerFunc(recipeHandler.Create),
	))
	mux.Handle("POST /menus/generate", middleware.RequireAuth(
		http.HandlerFunc(menuHandler.Generate),
	))
	mux.Handle("GET /menus/current", middleware.RequireAuth(
		http.HandlerFunc(menuHandler.GetCurrent),
	))
	mux.HandleFunc("GET /shopping-list", shoppingHandler.GetShoppingList)
	mux.HandleFunc("PATCH /shopping-list/items/{id}", shoppingHandler.UpdateItem)

	// Wrap with CORS middleware for frontend development
	return middleware.CORS(mux)
}
