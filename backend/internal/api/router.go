package api

import (
	"database/sql"
	"log"
	"maltiden/internal/api/handlers"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/claude"
	"maltiden/pkg/email"
	"maltiden/pkg/middleware"
	"maltiden/pkg/utils"
	"net/http"
	"os"
	"strings"
)

type dependencies struct {
	auth          *handlers.AuthHandler
	household     *handlers.HouseholdHandler
	recipe        *handlers.RecipeHandler
	menu          *handlers.MenuHandler
	shopping      *handlers.ShoppingHandler
	offers        *handlers.OffersHandler
	feedback      *handlers.FeedbackHandler
	health        *handlers.HealthHandler
	passwordReset *handlers.PasswordResetHandler
	userStorage   *sqlite.UserStorage // needed for token version checks in auth middleware
}

func wireDependencies(db *sql.DB, jwtService *utils.JWTService) *dependencies {
	// Storage layer
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	feedbackStorage := sqlite.NewFeedbackStorage(db)

	// Service layer
	authService := services.NewAuthService(db, userStorage, householdStorage, jwtService)
	householdService := services.NewHouseholdService(householdStorage, userStorage)
	recipeService := services.NewRecipeService(recipeStorage)
	menuService := services.NewMenuService(menuStorage, recipeStorage)
	shoppingService := services.NewShoppingService(menuStorage, recipeStorage, shoppingStorage)
	tjekService := services.NewTjekService()
	feedbackService := services.NewFeedbackService(feedbackStorage)

	// Email sender — defaults to log-only; uses Resend if RESEND_API_KEY is set.
	var emailer email.EmailSender
	fromAddr := os.Getenv("EMAIL_FROM")
	if fromAddr == "" {
		fromAddr = "Måltiden <no-reply@maltiden.app>"
	}
	if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		emailer = email.NewResendSender(apiKey, fromAddr)
		log.Printf("Email sender: Resend (%s)", fromAddr)
	} else {
		emailer = email.NewLogSender(fromAddr)
		log.Printf("Email sender: log-only (set RESEND_API_KEY to enable real email)")
	}
	passwordResetService := services.NewPasswordResetService(db, userStorage, emailer)

	// Handler layer
	return &dependencies{
		auth:          handlers.NewAuthHandler(authService),
		household:     handlers.NewHouseholdHandler(householdService),
		recipe:        handlers.NewRecipeHandler(recipeService),
		menu:          handlers.NewMenuHandler(menuService),
		shopping:      handlers.NewShoppingHandler(shoppingService, menuStorage),
		offers:        handlers.NewOffersHandler(tjekService),
		feedback:      handlers.NewFeedbackHandler(feedbackService),
		health:        handlers.NewHealthHandler(db),
		passwordReset: handlers.NewPasswordResetHandler(passwordResetService),
		userStorage:   userStorage,
	}
}

func NewRouter(db *sql.DB, jwtService *utils.JWTService) http.Handler {
	mux := http.NewServeMux()

	// Wire dependencies
	deps := wireDependencies(db, jwtService)

	// Rate limiter for auth endpoints: 5 requests/sec, burst of 10
	authLimiter := middleware.NewRateLimiter(5, 10)

	// Rate limiter for password reset: ~3 requests/hour per IP (burst 3)
	// 3/3600 ≈ 0.000833 tokens/sec
	passwordResetLimiter := middleware.NewRateLimiter(3.0/3600.0, 3)

	// Rate limiter for invite code join: 3 requests/sec, burst of 5 (brute-force protection)
	joinLimiter := middleware.NewRateLimiter(3, 5)

	// Rate limiter for recipe parser: 2 requests/sec, burst of 5 (API credit protection)
	parserLimiter := middleware.NewRateLimiter(2, 5)

	// Rate limiter for custom shopping item creation: 0.5 req/sec, burst of 30
	// (~30 req/min sustained, prevents abuse of unbounded list growth).
	customItemLimiter := middleware.NewRateLimiter(0.5, 30)

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
	mux.Handle("POST /auth/password-reset/request", passwordResetLimiter.Limit(http.HandlerFunc(deps.auth.PasswordResetRequest)))
	mux.Handle("POST /auth/forgot-password", passwordResetLimiter.Limit(http.HandlerFunc(deps.passwordReset.ForgotPassword)))
	mux.Handle("POST /auth/reset-password", passwordResetLimiter.Limit(http.HandlerFunc(deps.passwordReset.ResetPassword)))
	mux.HandleFunc("GET /offers/search", deps.offers.SearchOffers)
	mux.HandleFunc("GET /offers/discounts", deps.offers.GetDiscounts)
	mux.HandleFunc("GET /offers/stores", deps.offers.GetStores)
	mux.Handle("GET /recipes", middleware.OptionalAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.recipe.GetAll),
	))
	mux.HandleFunc("GET /recipes/{id}", deps.recipe.GetByID)

	// Protected routes
	mux.Handle("GET /households/me", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.GetMyHousehold),
	))
	mux.Handle("PATCH /households/me", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.UpdateMyHousehold),
	))
	mux.Handle("POST /households/invite", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.CreateInvite),
	))
	mux.Handle("POST /households/join", joinLimiter.Limit(middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.JoinHousehold),
	)))
	mux.Handle("GET /households/members/status", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.GetMemberStatuses),
	))
	mux.Handle("PATCH /households/members/{id}/status", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.UpdateMemberStatus),
	))
	mux.Handle("DELETE /households/members/{id}", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.household.RemoveMember),
	))
	mux.Handle("POST /recipes", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.recipe.Create),
	))
	mux.Handle("PUT /recipes/{id}", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.recipe.Update),
	))
	mux.Handle("DELETE /recipes/{id}", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.recipe.Delete),
	))
	if parserHandler != nil {
		// Recipe parser routes call the Claude API which can routinely take
		// 20–40s for long recipes. Apply a longer per-route timeout so the
		// global 10s wrapper (added below) does not turn legitimate parses
		// into 503s. The longer timeout is applied as the OUTERMOST wrapper
		// here; the global Timeout middleware below skips these routes.
		parseTimeout := middleware.TimeoutWith(middleware.ParserTimeout)
		mux.Handle("POST /recipes/parse", parseTimeout(parserLimiter.Limit(middleware.RequireAuth(jwtService, deps.userStorage)(
			http.HandlerFunc(parserHandler.ParseRecipe),
		))))
		mux.Handle("POST /recipes/parse-and-save", parseTimeout(parserLimiter.Limit(middleware.RequireAuth(jwtService, deps.userStorage)(
			http.HandlerFunc(parserHandler.ParseAndSave),
		))))
	}
	mux.Handle("GET /menus", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.menu.List),
	))
	mux.Handle("POST /menus/generate", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.menu.Generate),
	))
	mux.Handle("PUT /menus/current", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.menu.UpdateCurrent),
	))
	mux.Handle("GET /menus/current", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.menu.GetCurrent),
	))
	mux.Handle("GET /shopping-list", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.shopping.GetShoppingList),
	))
	mux.Handle("PATCH /shopping-list/items/{id}", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.shopping.UpdateItem),
	))
	mux.Handle("POST /shopping-list/items", customItemLimiter.Limit(middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.shopping.AddCustomItem),
	)))
	mux.Handle("DELETE /shopping-list/items/{id}", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.shopping.DeleteCustomItem),
	))
	mux.Handle("POST /feedback", middleware.RequireAuth(jwtService, deps.userStorage)(
		http.HandlerFunc(deps.feedback.Create),
	))

	// Get CORS origins from environment or use development defaults
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173,http://localhost:4173,http://127.0.0.1:5173"
	}
	allowedOrigins := strings.Split(corsOrigins, ",")

	// Wrap with middleware: CORS → request ID → timeout → handler
	// Note: Security headers are applied in main.go to cover both API and static files
	//
	// The global 10s Timeout is bypassed for recipe parser routes — those
	// already have a per-route 60s timeout applied above, and wrapping them
	// in an outer 10s deadline would defeat that.
	timeoutHandler := middleware.Timeout(mux)
	timed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && (r.URL.Path == "/recipes/parse" || r.URL.Path == "/recipes/parse-and-save") {
			mux.ServeHTTP(w, r)
			return
		}
		timeoutHandler.ServeHTTP(w, r)
	})
	return middleware.CORS(allowedOrigins)(middleware.RequestID(timed))
}
