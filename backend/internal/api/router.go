package api

import (
	"database/sql"
	"log"
	"maltiden/internal/api/handlers"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/claude"
	"maltiden/pkg/middleware"
	"maltiden/pkg/utils"
	"net/http"
	"os"
	"strings"
)

type dependencies struct {
	auth      *handlers.AuthHandler
	household *handlers.HouseholdHandler
	recipe    *handlers.RecipeHandler
	menu      *handlers.MenuHandler
	shopping  *handlers.ShoppingHandler
	offers    *handlers.OffersHandler
	health    *handlers.HealthHandler
}

func wireDependencies(db *sql.DB, jwtService *utils.JWTService) *dependencies {
	// Storage layer
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)

	// Service layer
	authService := services.NewAuthService(db, userStorage, householdStorage, jwtService)
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
		health:    handlers.NewHealthHandler(db),
	}
}

func NewRouter(db *sql.DB, jwtService *utils.JWTService) http.Handler {
	mux := http.NewServeMux()

	// Wire dependencies
	deps := wireDependencies(db, jwtService)

	// Rate limiter for auth endpoints: 5 requests/sec, burst of 10
	authLimiter := middleware.NewRateLimiter(5, 10)

	// Recipe parser (Claude API) - optional, degrades gracefully if ANTHROPIC_API_KEY not set
	var parserHandler *handlers.RecipeParserHandler
	claudeClient, err := claude.NewClient()
	if err != nil {
		log.Printf("Warning: Recipe parser disabled: %v", err)
	} else {
		recipeStorage := sqlite.NewRecipeStorage(db)
		recipeService := services.NewRecipeService(recipeStorage)
		parserService := services.NewRecipeParserService(claudeClient)
		parserHandler = handlers.NewRecipeParserHandler(parserService, recipeService)
	}

	// Public routes
	mux.HandleFunc("GET /health", deps.health.Check)
	mux.Handle("POST /auth/register", authLimiter.Limit(http.HandlerFunc(deps.auth.Register)))
	mux.Handle("POST /auth/login", authLimiter.Limit(http.HandlerFunc(deps.auth.Login)))
	mux.HandleFunc("GET /offers/search", deps.offers.SearchOffers)
	mux.HandleFunc("GET /offers/discounts", deps.offers.GetDiscounts)
	mux.HandleFunc("GET /offers/stores", deps.offers.GetStores)
	mux.HandleFunc("GET /recipes", deps.recipe.GetAll)
	mux.HandleFunc("GET /recipes/{id}", deps.recipe.GetByID)

	// Protected routes
	mux.Handle("GET /households/me", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.household.GetMyHousehold),
	))
	mux.Handle("POST /households/invite", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.household.CreateInvite),
	))
	mux.Handle("POST /households/join", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.household.JoinHousehold),
	))
	mux.Handle("GET /households/members/status", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.household.GetMemberStatuses),
	))
	mux.Handle("PATCH /households/members/{id}/status", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.household.UpdateMemberStatus),
	))
	mux.Handle("DELETE /households/members/{id}", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.household.RemoveMember),
	))
	mux.Handle("POST /recipes", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.recipe.Create),
	))
	if parserHandler != nil {
		mux.Handle("POST /recipes/parse", middleware.RequireAuth(jwtService)(
			http.HandlerFunc(parserHandler.ParseRecipe),
		))
		mux.Handle("POST /recipes/parse-and-save", middleware.RequireAuth(jwtService)(
			http.HandlerFunc(parserHandler.ParseAndSave),
		))
	}
	mux.Handle("POST /menus/generate", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.menu.Generate),
	))
	mux.Handle("GET /menus/current", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.menu.GetCurrent),
	))
	mux.Handle("GET /shopping-list", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.shopping.GetShoppingList),
	))
	mux.Handle("PATCH /shopping-list/items/{id}", middleware.RequireAuth(jwtService)(
		http.HandlerFunc(deps.shopping.UpdateItem),
	))

	// Get CORS origins from environment or use development defaults
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173,http://localhost:4173,http://127.0.0.1:5173"
	}
	allowedOrigins := strings.Split(corsOrigins, ",")

	// Wrap with middleware: request ID inside CORS
	return middleware.CORS(allowedOrigins)(middleware.RequestID(mux))
}
