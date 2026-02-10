# Architecture

**Analysis Date:** 2026-02-06

## Pattern Overview

**Overall:** Monorepo with separate frontend SPA and backend REST API (classic client-server)

**Key Characteristics:**
- Go backend with layered architecture (handlers -> services -> storage)
- Vue 3 SPA frontend with Pinia stores and API layer
- SQLite database with file-based migrations
- JWT-based authentication with Bearer tokens
- Frontend communicates with backend via REST JSON API over HTTP
- Mock data system allows frontend development without backend running
- No shared code between frontend and backend (types duplicated)

## Layers

### Backend Layers

**HTTP Handlers (Presentation):**
- Purpose: Parse HTTP requests, call services, write HTTP responses
- Location: `backend/internal/api/handlers/`
- Contains: `auth.go`, `household.go`, `recipes.go`, `menus.go`, `shopping.go`, `offers.go`, `health.go`
- Depends on: Services layer, Domain types, Middleware
- Used by: Router (`backend/internal/api/router.go`)
- Pattern: Each handler struct receives a service via constructor injection

**Services (Business Logic):**
- Purpose: Core business rules, validation, orchestration
- Location: `backend/internal/services/`
- Contains: `auth_service.go`, `household_service.go`, `recipe_service.go`, `menu_service.go`, `shopping_service.go`, `tjek_service.go`
- Depends on: Storage layer, Domain types, Utils
- Used by: Handlers
- Pattern: Each service struct receives storage dependencies via constructor injection

**Storage (Data Access):**
- Purpose: Database CRUD operations, raw SQL queries
- Location: `backend/internal/storage/sqlite/`
- Contains: `db.go`, `user_storage.go`, `household_storage.go`, `recipe_storage.go`, `menu_storage.go`, `shopping_storage.go`
- Depends on: `database/sql`, Domain types
- Used by: Services
- Pattern: Each storage struct wraps `*sql.DB`, uses raw SQL (no ORM)

**Domain (Types/Models):**
- Purpose: Shared data types for requests, responses, and entities
- Location: `backend/internal/domain/`
- Contains: `user.go`, `auth.go`, `household.go`, `recipe.go`, `menu.go`, `shopping.go`, `offers.go`
- Depends on: Nothing (leaf package)
- Used by: All layers

**Middleware & Utilities:**
- Purpose: Cross-cutting concerns (auth, CORS) and helper functions (JWT, password hashing)
- Location: `backend/pkg/middleware/` and `backend/pkg/utils/`
- Contains: `auth.go`, `cors.go`, `jwt.go`, `password.go`
- Used by: Router and Services

### Frontend Layers

**Views (Pages):**
- Purpose: Top-level page components, route targets
- Location: `frontend/src/views/`
- Contains: `LandingView.vue`, `LoginView.vue`, `OnboardingView.vue`, `DashboardView.vue`, `GenerateMenuView.vue`, `AboutView.vue`, `OffersView.vue`, `HomeView.vue`
- Depends on: Components, Stores
- Used by: Router

**Components (UI):**
- Purpose: Reusable and feature-specific UI components
- Location: `frontend/src/components/`
- Contains subdirectories: `common/`, `dashboard/`, `landing/`, `menu/`, `poc/`
- Depends on: Stores, API types
- Used by: Views

**Stores (State Management):**
- Purpose: Centralized reactive state, async data fetching, business logic
- Location: `frontend/src/stores/`
- Contains: `user.ts`, `dashboard.ts`, `landing.ts`, `theme.ts`, `menuGenerator.ts`, `counter.ts`
- Depends on: API layer, Mocks
- Used by: Views and Components
- Pattern: Pinia composition API stores (`defineStore` with setup function)

**API Layer (HTTP Client):**
- Purpose: Backend communication, typed API calls
- Location: `frontend/src/api/`
- Contains: `client.ts`, `auth.api.ts`, `household.api.ts`, `recipes.api.ts`, `menu.api.ts`, `shopping.api.ts`, `offers.api.ts`, `landing.api.ts`
- Depends on: Axios, Mock data
- Used by: Stores
- Pattern: Each API file exports async functions that check `USE_MOCKS` flag before calling backend

**API Types:**
- Purpose: TypeScript type definitions for API contracts
- Location: `frontend/src/api/types/`
- Contains: `dashboard.types.ts`, `offers.types.ts`, `landing.types.ts`, `error.types.ts`
- Used by: API layer, Stores, Components

**Mocks:**
- Purpose: Fake data for frontend development without backend
- Location: `frontend/src/mocks/`
- Contains: `index.ts`, `auth.mock.ts`, `dashboard.mock.ts`, `household.mock.ts`, `menu.mock.ts`, `recipes.mock.ts`, `shopping.mock.ts`, `offers.mock.ts`, `landing.mock.ts`
- Used by: API layer (conditional on `USE_MOCKS`)
- Toggle: `VITE_USE_REAL_API` env var; `npm run dev:mock` uses mocks, `npm run dev` uses real API

## Data Flow

**Typical Authenticated Request (e.g., Generate Menu):**

1. User action in Vue component triggers Pinia store action (`frontend/src/stores/menuGenerator.ts` -> `generateInitialMenu()`)
2. Store calls API function (`frontend/src/api/menu.api.ts` -> `generateMenu()`)
3. API function checks `USE_MOCKS`; if false, calls `apiClient.post('/menus/generate', request)` (`frontend/src/api/client.ts`)
4. Axios request interceptor attaches JWT from localStorage (`frontend/src/utils/token.ts`)
5. Request hits Go backend router (`backend/internal/api/router.go`), matched to `POST /menus/generate`
6. `middleware.RequireAuth` validates JWT, extracts `userID` and `householdID` into request context (`backend/pkg/middleware/auth.go`)
7. Handler parses JSON body, calls service method (`backend/internal/api/handlers/menus.go` -> `menuService.Generate()`)
8. Service performs business logic, calls storage (`backend/internal/services/menu_service.go`)
9. Storage executes SQL against SQLite (`backend/internal/storage/sqlite/menu_storage.go`)
10. Response flows back: Storage -> Service -> Handler -> JSON -> Axios -> Store -> Reactive UI update

**State Management:**
- Frontend: Pinia stores with composition API pattern. Each store holds `ref()` state, `computed()` getters, and async action functions
- JWT token stored in `localStorage` under key `maltiden_token`
- Theme preference stored in `localStorage` under key `maltiden_theme`
- Backend: Stateless. All state in SQLite database. No in-memory caching

## Key Abstractions

**Backend ID Convention:**
- All entity IDs use prefixed UUIDs: `usr_`, `hh_`, `hm_`, `rec_`, `menu_`, `inv_`, `item_`
- Generated via `"prefix_" + uuid.New().String()` pattern

**Dependency Injection (Backend):**
- Manual constructor injection via `NewXxxStorage(db)` -> `NewXxxService(storage)` -> `NewXxxHandler(service)`
- All wired in `backend/internal/api/router.go` `NewRouter()` function
- No DI container; explicit wiring

**Mock Switching (Frontend):**
- Every API function checks `USE_MOCKS` constant at call time
- Defined in `frontend/src/mocks/index.ts`: `USE_MOCKS = import.meta.env.DEV && import.meta.env.VITE_USE_REAL_API !== 'true'`
- Two env files: `frontend/.env.development` (real API) and `frontend/.env.mock` (mocks)

**Route Prefetching (Frontend):**
- Custom `v-prefetch` directive at `frontend/src/directives/vPrefetch.ts`
- Composable at `frontend/src/composables/usePrefetch.ts`
- Prefetches route chunks on hover/touch/focus

**Role-Based Access:**
- Three roles: `owner`, `member`, `guest`
- Backend: `household_members.role` column with CHECK constraint
- Frontend: `useUserStore` exposes `isOwner`, `isMember`, `isGuest` computed properties
- Router guard checks `requiresAuth` and `requiresMember` meta fields (`frontend/src/router/index.ts`)

## Entry Points

**Backend:**
- Location: `backend/cmd/server/main.go`
- Triggers: `go run ./cmd/server` or Docker `CMD ["./server"]`
- Responsibilities: Read env vars (`PORT`, `DATABASE_PATH`), open SQLite (run migrations), create router, start HTTP server

**Frontend:**
- Location: `frontend/src/main.ts`
- Triggers: Vite dev server or built `index.html`
- Responsibilities: Create Vue app, install Pinia and Router, register `v-prefetch` directive, initialize theme, mount to `#app`

**Frontend HTML Shell:**
- Location: `frontend/index.html`
- Loads: `src/main.ts` via Vite

**Placeholder Landing Page:**
- Location: `index.html` (root) - static "under construction" page, served by Fly.io currently

## Error Handling

**Strategy:** Error strings returned from services; handlers map to HTTP status codes. No custom error types.

**Backend Patterns:**
- Services return `(result, error)` tuples
- Handlers use `http.Error()` with inline JSON strings: `http.Error(w, '{"error":"message"}', statusCode)`
- No structured error types; plain `errors.New("string")` in services
- Missing resources return `(nil, nil)` from storage, handlers check for nil and return 404

**Frontend Patterns:**
- API functions throw on HTTP errors (Axios default behavior)
- Stores catch errors in try/catch, set `error` ref for UI display
- Axios response interceptor handles 401 globally: removes token, redirects to `/login` (`frontend/src/api/client.ts`)
- Dashboard store falls back to mock data if backend fails (`frontend/src/stores/dashboard.ts`)

## Cross-Cutting Concerns

**Logging:**
- Backend: `log` stdlib only (no structured logging). `log.Fatal()` for startup errors, `log.Printf()` for server start message
- Frontend: `console.error()` and `console.warn()` for error cases

**Validation:**
- Backend: Manual validation in services (e.g., password length in `backend/internal/services/auth_service.go`, recipe fields in `backend/internal/services/recipe_service.go`)
- Frontend: Minimal client-side validation; relies on backend error responses
- Database: CHECK constraints on `household_members.role`, UNIQUE constraints on emails and invite codes

**Authentication:**
- JWT tokens with HS256 signing, 7-day expiry (`backend/pkg/utils/jwt.go`)
- Secret from `JWT_SECRET` env var (read at package init time)
- Token contains `user_id` and `household_id` claims
- Frontend stores token in localStorage, attaches via Axios request interceptor
- Middleware: `backend/pkg/middleware/auth.go` - `RequireAuth()` wraps protected handlers
- CORS: `backend/pkg/middleware/cors.go` - allows `localhost:5173` and `localhost:4173`

**Theming:**
- CSS custom properties in `frontend/src/styles/theme.css`
- Dark/light mode via `data-theme` attribute on `<html>`
- Persisted in localStorage, respects system preference on first visit
- Managed by `frontend/src/stores/theme.ts`

---

*Architecture analysis: 2026-02-06*
*Update when major patterns change*
