# Codebase Structure

**Analysis Date:** 2026-02-06

## Directory Layout

```
maltiden/
├── backend/                # Go REST API server
│   ├── cmd/server/         # Application entry point
│   ├── internal/           # Private application code
│   │   ├── api/            # HTTP layer (router + handlers)
│   │   │   └── handlers/   # Request handlers per domain
│   │   ├── domain/         # Shared types (entities, DTOs)
│   │   ├── services/       # Business logic layer
│   │   └── storage/        # Data access layer
│   │       └── sqlite/     # SQLite implementations
│   ├── migrations/         # SQL migration files
│   └── pkg/                # Shared packages (public)
│       ├── middleware/      # HTTP middleware (auth, CORS)
│       └── utils/          # Utilities (JWT, password hashing)
├── frontend/               # Vue 3 SPA
│   ├── public/             # Static assets (favicon)
│   └── src/                # Application source
│       ├── api/            # Backend API client functions
│       │   └── types/      # TypeScript API type definitions
│       ├── assets/         # Images and static resources
│       │   └── images/     # Image files
│       ├── components/     # Vue components
│       │   ├── common/     # Shared reusable components
│       │   ├── dashboard/  # Dashboard-specific components
│       │   ├── landing/    # Landing page components
│       │   ├── menu/       # Menu generation components
│       │   └── poc/        # Proof-of-concept components
│       ├── composables/    # Vue composables (reusable logic)
│       ├── directives/     # Custom Vue directives
│       ├── mocks/          # Mock data for dev without backend
│       ├── router/         # Vue Router configuration
│       ├── stores/         # Pinia state management
│       ├── styles/         # Global CSS (theme variables)
│       ├── utils/          # Utility functions
│       └── views/          # Page-level Vue components
├── docs/                   # Project documentation
├── .github/                # GitHub config
│   └── workflows/          # CI/CD workflows
└── .planning/              # Planning documents
    └── codebase/           # Architecture analysis (this file)
```

## Directory Purposes

**`backend/cmd/server/`:**
- Purpose: Go application entry point
- Contains: `main.go` - reads env vars, opens DB, creates router, starts HTTP server
- Key files: `backend/cmd/server/main.go`

**`backend/internal/api/`:**
- Purpose: HTTP routing and request handling
- Contains: Router setup and handler packages
- Key files: `backend/internal/api/router.go` (route definitions and DI wiring)
- Subdirectories: `handlers/` (one file per domain: auth, household, recipes, menus, shopping, offers, health)

**`backend/internal/domain/`:**
- Purpose: Domain types shared across all backend layers
- Contains: Entity structs, request/response DTOs
- Key files: `user.go`, `auth.go`, `household.go`, `recipe.go`, `menu.go`, `shopping.go`, `offers.go`
- Convention: Request types suffixed with `Request`, response types suffixed with `Response`

**`backend/internal/services/`:**
- Purpose: Business logic, validation, orchestration
- Contains: One service per domain area
- Key files: `auth_service.go`, `household_service.go`, `recipe_service.go`, `menu_service.go`, `shopping_service.go`, `tjek_service.go`
- Pattern: Struct with storage dependencies, constructor `NewXxxService()`

**`backend/internal/storage/sqlite/`:**
- Purpose: SQLite database operations (CRUD)
- Contains: One storage struct per entity, plus DB initialization
- Key files: `db.go` (Open + migrations), `user_storage.go`, `household_storage.go`, `recipe_storage.go`, `menu_storage.go`, `shopping_storage.go`
- Pattern: Raw SQL queries, `*sql.DB` wrapper structs

**`backend/migrations/`:**
- Purpose: SQL schema migration files run at startup
- Contains: Numbered SQL files (`001_create_users.sql` through `006_household_invites_and_member_status.sql`)
- Tracked in: `schema_migrations` table (version + applied_at)
- Convention: `NNN_description.sql` format

**`backend/pkg/middleware/`:**
- Purpose: HTTP middleware (auth guard, CORS)
- Key files: `backend/pkg/middleware/auth.go` (JWT validation, context injection), `backend/pkg/middleware/cors.go` (dev CORS headers)

**`backend/pkg/utils/`:**
- Purpose: Shared utility functions
- Key files: `backend/pkg/utils/jwt.go` (token generation/validation), `backend/pkg/utils/password.go` (bcrypt hashing)

**`frontend/src/api/`:**
- Purpose: Typed API client functions for each backend endpoint
- Contains: One `.api.ts` file per domain, plus shared `client.ts`
- Key files: `client.ts` (Axios instance with interceptors), `auth.api.ts`, `household.api.ts`, `recipes.api.ts`, `menu.api.ts`, `shopping.api.ts`, `offers.api.ts`, `landing.api.ts`
- Pattern: Export async functions, check `USE_MOCKS` before calling backend

**`frontend/src/api/types/`:**
- Purpose: TypeScript interfaces for API contracts
- Key files: `dashboard.types.ts` (User, Meal, MenuDay, Household, ShoppingList, DashboardData), `offers.types.ts`, `landing.types.ts`, `error.types.ts`

**`frontend/src/components/`:**
- Purpose: Vue single-file components organized by feature area
- Subdirectories:
  - `common/` - Shared components: `BaseButton.vue`, `BaseCard.vue`, `LockedAction.vue`
  - `dashboard/` - Dashboard widgets: `DashboardHeader.vue`, `TodaysMeal.vue`, `WeeklyMenuGrid.vue`, `HouseholdWidget.vue`, `ShoppingListWidget.vue`, `QuickActions.vue`, `SettingsModal.vue`
  - `landing/` - Landing page sections: `HeroSection.vue`, `FeaturesSection.vue`, `FeatureCard.vue`, `CtaSection.vue`
  - `menu/` - Menu generation UI: `GenerateMenuEmptyState.vue`, `MenuDayCard.vue`, `MenuGeneratorActions.vue`
  - `poc/` - Proof-of-concept features: `OfferSearch.vue`

**`frontend/src/stores/`:**
- Purpose: Pinia stores for state management
- Key files: `user.ts` (auth state, login/register), `dashboard.ts` (dashboard data aggregation), `menuGenerator.ts` (menu generation workflow), `landing.ts` (landing page content), `theme.ts` (dark/light mode)
- Pattern: Composition API with `defineStore('name', () => { ... })`

**`frontend/src/mocks/`:**
- Purpose: Mock data for frontend-only development
- Key files: `index.ts` (USE_MOCKS flag), one mock file per API domain
- Toggle: `frontend/.env.mock` sets `VITE_USE_REAL_API=false`

**`frontend/src/views/`:**
- Purpose: Page-level components (one per route)
- Key files: `LandingView.vue`, `LoginView.vue`, `OnboardingView.vue`, `DashboardView.vue`, `GenerateMenuView.vue`, `AboutView.vue`, `OffersView.vue`, `HomeView.vue`
- Convention: `XxxView.vue` naming, lazy-loaded via dynamic imports in router

**`frontend/src/router/`:**
- Purpose: Vue Router configuration and navigation guards
- Key files: `frontend/src/router/index.ts`
- Contains: Route definitions with `meta.requiresAuth` and `meta.requiresMember` for auth guards, route transitions

**`frontend/src/composables/`:**
- Purpose: Reusable Vue composition functions
- Key files: `usePrefetch.ts` (route chunk prefetching on hover)

**`frontend/src/directives/`:**
- Purpose: Custom Vue directives
- Key files: `vPrefetch.ts` (prefetch route on hover/touch/focus)

**`frontend/src/styles/`:**
- Purpose: Global CSS theme variables
- Key files: `theme.css` (CSS custom properties for light/dark mode)

**`frontend/src/utils/`:**
- Purpose: Shared TypeScript utilities
- Key files: `token.ts` (localStorage JWT token management)

## Key File Locations

**Entry Points:**
- `backend/cmd/server/main.go`: Backend server entry
- `frontend/src/main.ts`: Frontend app entry
- `frontend/index.html`: Vite HTML shell
- `index.html`: Root static placeholder page (currently served by Fly.io)

**Configuration:**
- `backend/go.mod`: Go module and dependencies
- `frontend/package.json`: Node dependencies and scripts
- `frontend/vite.config.ts`: Vite build configuration (alias: `@` -> `src/`)
- `frontend/tsconfig.json`: TypeScript configuration
- `frontend/eslint.config.ts`: ESLint rules
- `frontend/.prettierrc.json`: Prettier formatting
- `frontend/.env.development`: Dev config (real API)
- `frontend/.env.mock`: Dev config (mock data)
- `Dockerfile`: Production Docker build (Go only, no frontend build)
- `fly.toml`: Fly.io deployment config

**Core Logic:**
- `backend/internal/api/router.go`: All route definitions and dependency wiring
- `backend/internal/services/auth_service.go`: Registration and login logic
- `backend/internal/services/menu_service.go`: Menu generation algorithm
- `backend/internal/services/shopping_service.go`: Shopping list aggregation from menu
- `backend/internal/services/tjek_service.go`: External grocery offer API integration
- `frontend/src/stores/user.ts`: Auth state management
- `frontend/src/stores/dashboard.ts`: Dashboard data orchestration
- `frontend/src/stores/menuGenerator.ts`: Menu generation workflow UI state

**Testing:**
- `backend/internal/services/household_service_test.go`: Only test file in the project

**Documentation:**
- `docs/`: Project documentation directory
- `README.md`: Brief project readme
- `frontend/README.md`: Frontend-specific readme

## Naming Conventions

**Files (Backend):**
- Go files: `snake_case.go` (e.g., `auth_service.go`, `user_storage.go`)
- Migrations: `NNN_description.sql` (e.g., `001_create_users.sql`)
- One struct per file typical, file named after struct

**Files (Frontend):**
- Vue components: `PascalCase.vue` (e.g., `DashboardHeader.vue`, `BaseCard.vue`)
- TypeScript files: `camelCase.ts` or `kebab-case.ts`
- API files: `domain.api.ts` (e.g., `auth.api.ts`, `menu.api.ts`)
- Type files: `domain.types.ts` (e.g., `dashboard.types.ts`)
- Mock files: `domain.mock.ts` (e.g., `auth.mock.ts`)
- Store files: `domain.ts` (e.g., `user.ts`, `dashboard.ts`)
- View files: `XxxView.vue` suffix

**Directories:**
- Backend: `snake_case` following Go convention
- Frontend: `kebab-case` or `camelCase` (e.g., `components/`, `composables/`)
- Component subdirs by feature: `common/`, `dashboard/`, `landing/`, `menu/`, `poc/`

**Special Patterns:**
- Backend entity IDs: Prefixed UUIDs (`usr_`, `hh_`, `hm_`, `rec_`, `menu_`, `inv_`, `item_`)
- Backend constructors: `NewXxxType()` pattern
- Frontend stores: `useXxxStore` naming via `defineStore`
- Frontend API: `USE_MOCKS` conditional at top of each API function

## Where to Add New Code

**New Backend Feature (e.g., new domain entity):**
1. Domain types: Add `backend/internal/domain/{entity}.go`
2. Migration: Add `backend/migrations/NNN_{description}.sql` and register in `backend/internal/storage/sqlite/db.go` migrations list
3. Storage: Add `backend/internal/storage/sqlite/{entity}_storage.go`
4. Service: Add `backend/internal/services/{entity}_service.go`
5. Handler: Add `backend/internal/api/handlers/{entity}.go`
6. Routes: Register in `backend/internal/api/router.go` `NewRouter()` function

**New Backend API Endpoint (existing domain):**
1. Add handler method to existing handler struct in `backend/internal/api/handlers/{domain}.go`
2. Add service method if needed in `backend/internal/services/{domain}_service.go`
3. Register route in `backend/internal/api/router.go`

**New Frontend Page/View:**
1. Create view: `frontend/src/views/XxxView.vue`
2. Add route: `frontend/src/router/index.ts` (use lazy import)
3. Add prefetch entry: `frontend/src/composables/usePrefetch.ts` routeImports map
4. If needs state: Create store at `frontend/src/stores/xxx.ts`
5. If needs API: Create `frontend/src/api/xxx.api.ts` and mock at `frontend/src/mocks/xxx.mock.ts`

**New Frontend Component:**
- Shared/reusable: `frontend/src/components/common/XxxComponent.vue`
- Feature-specific: `frontend/src/components/{feature}/XxxComponent.vue`
- Create new feature subdirectory if needed

**New Frontend API Integration:**
1. Add types: `frontend/src/api/types/{domain}.types.ts`
2. Add API functions: `frontend/src/api/{domain}.api.ts` (include `USE_MOCKS` check)
3. Add mock: `frontend/src/mocks/{domain}.mock.ts`
4. Export from mock index if needed: `frontend/src/mocks/index.ts`

**New Utility:**
- Backend: `backend/pkg/utils/{name}.go`
- Frontend: `frontend/src/utils/{name}.ts`

**New Middleware (Backend):**
- Add to: `backend/pkg/middleware/{name}.go`
- Apply in: `backend/internal/api/router.go`

**New Composable (Frontend):**
- Add to: `frontend/src/composables/use{Name}.ts`

**New Directive (Frontend):**
- Add to: `frontend/src/directives/v{Name}.ts`
- Register in: `frontend/src/main.ts`

## Special Directories

**`backend/migrations/`:**
- Purpose: SQL schema files executed sequentially on startup
- Generated: No (hand-written)
- Committed: Yes
- Note: Migration versions tracked in `schema_migrations` DB table. New migrations must be added to the `migrations` slice in `backend/internal/storage/sqlite/db.go`

**`frontend/src/mocks/`:**
- Purpose: Mock data for frontend development without backend
- Generated: No (hand-written)
- Committed: Yes
- Note: Controlled by `VITE_USE_REAL_API` env var. Run `npm run dev:mock` to use mocks

**`frontend/public/`:**
- Purpose: Static assets served as-is by Vite
- Generated: No
- Committed: Yes

**`.planning/`:**
- Purpose: Project planning and codebase analysis documents
- Generated: By analysis tools
- Committed: Yes

**`.github/workflows/`:**
- Purpose: GitHub Actions CI/CD pipelines
- Contains: `ci.yml`, `claude.yml`, `claude-api-docs-sync.yml`, `claude-code-review.yml`, `labeler.yml`
- Committed: Yes

---

*Structure analysis: 2026-02-06*
*Update when directory structure changes*
