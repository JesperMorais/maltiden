package api

import (
	"database/sql"
	"maltiden/internal/api/handlers"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"net/http"
)

type dependencies struct {
	auth      *handlers.AuthHandler
	household *handlers.HouseholdHandler
	recipe    *handlers.RecipeHandler
	menu      *handlers.MenuHandler
	shopping  *handlers.ShoppingHandler
	offers    *handlers.OffersHandler
}

func wireDependencies(db *sql.DB) *dependencies {
	// Storage layer
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)

	// Service layer
	authService := services.NewAuthService(db, userStorage, householdStorage)
	householdService := services.NewHouseholdService(householdStorage, userStorage)
	recipeService := services.NewRecipeService(recipeStorage)
	menuService := services.NewMenuService(menuStorage, recipeStorage)
	shoppingService := services.NewShoppingService(menuStorage, recipeStorage, shoppingStorage)
	tjekService := services.NewTjekService()

	// Handler layer
	return &dependencies{
		auth:      handlers.NewAuthHandler(authService),
		household: handlers.NewHouseholdHandler(householdService),
		recipe:    handlers.NewRecipeHandler(recipeService),
		menu:      handlers.NewMenuHandler(menuService),
		shopping:  handlers.NewShoppingHandler(shoppingService, menuStorage),
		offers:    handlers.NewOffersHandler(tjekService),
	}
}

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Wire dependencies
	deps := wireDependencies(db)

	// Public routes
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /auth/register", deps.auth.Register)
	mux.HandleFunc("POST /auth/login", deps.auth.Login)
	mux.HandleFunc("GET /offers/search", deps.offers.SearchOffers)
	mux.HandleFunc("GET /offers/discounts", deps.offers.GetDiscounts)
	mux.HandleFunc("GET /offers/stores", deps.offers.GetStores)
	mux.HandleFunc("GET /recipes", deps.recipe.GetAll)
	mux.HandleFunc("GET /recipes/{id}", deps.recipe.GetByID)

	// Protected routes
	mux.Handle("GET /households/me", middleware.RequireAuth(
		http.HandlerFunc(deps.household.GetMyHousehold),
	))
	mux.Handle("POST /households/invite", middleware.RequireAuth(
		http.HandlerFunc(deps.household.CreateInvite),
	))
	mux.Handle("POST /households/join", middleware.RequireAuth(
		http.HandlerFunc(deps.household.JoinHousehold),
	))
	mux.Handle("GET /households/members/status", middleware.RequireAuth(
		http.HandlerFunc(deps.household.GetMemberStatuses),
	))
	mux.Handle("PATCH /households/members/{id}/status", middleware.RequireAuth(
		http.HandlerFunc(deps.household.UpdateMemberStatus),
	))
	mux.Handle("DELETE /households/members/{id}", middleware.RequireAuth(
		http.HandlerFunc(deps.household.RemoveMember),
	))
	mux.Handle("POST /recipes", middleware.RequireAuth(
		http.HandlerFunc(deps.recipe.Create),
	))
	mux.Handle("POST /menus/generate", middleware.RequireAuth(
		http.HandlerFunc(deps.menu.Generate),
	))
	mux.Handle("GET /menus/current", middleware.RequireAuth(
		http.HandlerFunc(deps.menu.GetCurrent),
	))
	mux.Handle("GET /shopping-list", middleware.RequireAuth(
		http.HandlerFunc(deps.shopping.GetShoppingList),
	))
	mux.Handle("PATCH /shopping-list/items/{id}", middleware.RequireAuth(
		http.HandlerFunc(deps.shopping.UpdateItem),
	))

	// Wrap with CORS middleware for frontend development
	return middleware.CORS(mux)
}
