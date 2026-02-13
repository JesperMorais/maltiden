# Codebase Concerns

**Analysis Date:** 2026-02-06

## Tech Debt

**Error message injection in HTTP responses:**
- Issue: Several handlers concatenate `err.Error()` directly into JSON response strings, e.g. `{"error":"` + err.Error() + `"}`. This breaks JSON if the error message contains a quote character, and leaks internal error details to clients.
- Files: `backend/internal/api/handlers/auth.go:30`, `backend/internal/api/handlers/recipes.go:72`, `backend/internal/api/handlers/offers.go:74,116,147`
- Impact: Malformed JSON responses, potential information disclosure of internal errors to end users.
- Fix approach: Use `json.NewEncoder` to properly serialize error responses. Create a shared `writeError(w, code, msg)` helper. Never pass `err.Error()` to clients in production; use generic messages and log the real error server-side.

**Error handling via string matching:**
- Issue: Error routing in handlers uses `err.Error()` string comparison (`switch err.Error() { case "invalid_code": ... }`). This is fragile; any change to error message text silently breaks the routing.
- Files: `backend/internal/api/handlers/household.go:82-89,140-146,169-178`
- Impact: Changing an error message in a service silently changes HTTP status codes. No compiler safety.
- Fix approach: Define sentinel errors (`var ErrInvalidCode = errors.New("invalid_code")`) in the domain package and use `errors.Is()` in handlers.

**Duplicate coordinate parsing in offers handlers:**
- Issue: The lat/lng/radius parsing logic is copy-pasted across `SearchOffers`, `GetDiscounts`, and `GetStores` handlers (lines 36-50, 84-102, 125-143). The exclude-store parsing is also duplicated across `SearchOffers` and `GetDiscounts`.
- Files: `backend/internal/api/handlers/offers.go:27-81,83-122,124-155`
- Impact: Bug fixes or new parameters must be applied in three places. Easy to miss one.
- Fix approach: Extract a `parseLocationParams(r) (lat, lng, radius)` helper and a `parseExcludeStores(r) []string` helper.

**Bubble sort used instead of `sort.Slice`:**
- Issue: Manual O(n^2) bubble sort used for sorting stores and discount offers.
- Files: `backend/internal/services/tjek_service.go:286-292,432-438`
- Impact: Minor performance concern; functionally correct but uses an anti-pattern when `sort.Slice` is already imported and used elsewhere in the codebase (e.g., `backend/internal/services/shopping_service.go:138`).
- Fix approach: Replace with `sort.Slice` for consistency and readability.

**`math/rand` used without seeding (pre-Go 1.20 pattern concern):**
- Issue: `rand.Intn` is used for picking random recipes. In Go 1.20+ this is auto-seeded, but the import `"math/rand"` without a seed is a pattern that can produce predictable results on older Go versions. The project uses Go 1.24 so this is technically fine, but worth noting.
- Files: `backend/internal/services/menu_service.go:68`
- Impact: None with Go 1.24. Menu generation is "random enough" for its purpose.
- Fix approach: No action needed, but consider using `math/rand/v2` for the modern idiomatic approach.

**Recipe recipes.GetByID uses manual path splitting instead of PathValue:**
- Issue: `GetByID` manually splits `r.URL.Path` to extract the ID, despite the router supporting `r.PathValue("id")` (used correctly in `household.go:121,161`).
- Files: `backend/internal/api/handlers/recipes.go:40-46`, `backend/internal/api/handlers/shopping.go:43-49`
- Impact: Inconsistent pattern across handlers. Manual parsing is more error-prone.
- Fix approach: Replace `strings.Split(path, "/")` with `r.PathValue("id")` to match the pattern in other handlers.

**Dashboard store silently falls back to mock data on API failure:**
- Issue: When the `/households/me` API call fails, the dashboard store catches the error and loads full mock data without clearly indicating to the user that they're seeing fake data.
- Files: `frontend/src/stores/dashboard.ts:79-88`
- Impact: In production, users could see mock data (fake meals, fake household members) if the backend is temporarily down. This is confusing and misleading.
- Fix approach: Remove mock fallback in production. Only fall back to mocks when `USE_MOCKS` is true. Show a proper error state instead.

**Onboarding join flow uses hardcoded mock families:**
- Issue: The "join household" flow in OnboardingView uses a hardcoded `mockFamilies` dictionary and simulates API calls with `setTimeout`. This flow never actually calls the backend.
- Files: `frontend/src/views/OnboardingView.vue:46-51,86-108,114-133`
- Impact: The "join household" feature is non-functional. Only "create household" (register) actually works against the backend.
- Fix approach: Wire up the join flow to call `householdApi.joinHousehold()` after registering the user. The backend endpoints (`POST /households/join`) already exist.

**Menu save doesn't persist the actual draft:**
- Issue: `saveMenu()` in the menu generator store calls `generateMenu()` again with default params, discarding the user's locked/customized draft. The user sees one menu, but a different random menu gets saved.
- Files: `frontend/src/stores/menuGenerator.ts:319-353`
- Impact: Users who lock days and regenerate specific days will lose their selections when saving.
- Fix approach: Implement a `PUT /menus/current` endpoint that accepts exact recipe assignments per day, or modify the save to pass the draft's recipe IDs through the existing generate endpoint.

## Known Bugs

**Shopping list and item update endpoints lack authentication:**
- Symptoms: Any unauthenticated client can view and modify shopping lists by guessing menu IDs.
- Trigger: Send `GET /shopping-list?menuId=<id>` or `PATCH /shopping-list/items/<id>` without an Authorization header.
- Files: `backend/internal/api/router.go:76-77` (uses `mux.HandleFunc` instead of `mux.Handle` with `middleware.RequireAuth`)
- Workaround: Menu IDs include UUIDs so they're hard to guess, but this is security by obscurity.
- Root cause: These routes were added with `HandleFunc` (no middleware wrapping) instead of `Handle` with `RequireAuth`.

**fly.toml has conflicting memory settings:**
- Symptoms: Unclear which memory value is used; `memory = '1gb'` and `memory_mb = 256` appear in the same `[[vm]]` block.
- Trigger: Deploying to Fly.io.
- Files: `fly.toml:19-23`
- Workaround: Fly.io likely uses one and ignores the other, but the intent is ambiguous.
- Root cause: Copy-paste from Fly.io launch template without cleanup.

## Security Considerations

**JWT secret read at init time from environment:**
- Risk: `jwtSecret` is read from `os.Getenv("JWT_SECRET")` at package init time as a package-level variable. If `JWT_SECRET` is not set, the secret is an empty byte slice, which means tokens are signed with an empty key. Any attacker can forge tokens.
- Files: `backend/pkg/utils/jwt.go:11`
- Current mitigation: In production (Fly.io), the secret is set via `fly secrets`. But there is no validation that it's non-empty.
- Recommendations: Add a startup check in `main.go` that fatals if `JWT_SECRET` is empty or too short (< 32 chars). Consider passing the secret as a dependency rather than using a package-level var.

**CORS allows only localhost origins:**
- Risk: In production, the CORS middleware only allows `localhost:5173`, `localhost:4173`, and `127.0.0.1:5173`. This means the deployed frontend (different origin) will be blocked by CORS unless they're served from the same domain.
- Files: `backend/pkg/middleware/cors.go:11`
- Current mitigation: If the frontend is served from the same domain as the backend, CORS isn't needed. If they're separate, the frontend won't work.
- Recommendations: Make allowed origins configurable via environment variable (e.g., `CORS_ORIGINS`). Add the production frontend URL.

**Recipes are publicly accessible without authentication:**
- Risk: `GET /recipes` and `GET /recipes/{id}` are public routes. All recipe data (including any future premium content) is accessible to anyone.
- Files: `backend/internal/api/router.go:45-46`
- Current mitigation: Recipes are non-sensitive community content for now.
- Recommendations: Consider if recipes should be household-scoped in the future. No immediate action needed.

**No rate limiting on authentication endpoints:**
- Risk: `POST /auth/login` and `POST /auth/register` have no rate limiting. Brute-force attacks on passwords and email enumeration are possible.
- Files: `backend/internal/api/router.go:40-41`, `backend/internal/api/handlers/auth.go`
- Current mitigation: bcrypt with cost 12 provides some natural rate limiting (~250ms per attempt).
- Recommendations: Add rate limiting middleware (e.g., per-IP token bucket) for auth endpoints. Return consistent errors for "user not found" vs "wrong password" (currently already done correctly - both return "invalid credentials").

**No input size limits on request bodies:**
- Risk: `json.NewDecoder(r.Body).Decode()` reads the entire request body. An attacker can send a multi-GB body to exhaust memory.
- Files: All handlers in `backend/internal/api/handlers/` that decode JSON bodies.
- Current mitigation: None.
- Recommendations: Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, maxSize)` before decoding. A 1MB limit is reasonable.

**No email validation on registration:**
- Risk: Registration only checks password length (>= 8 chars). No email format validation, no email uniqueness check on format (case sensitivity), no verification email.
- Files: `backend/internal/services/auth_service.go:27-38`
- Current mitigation: SQLite UNIQUE constraint on email column prevents duplicates (exact match).
- Recommendations: Normalize email (lowercase, trim whitespace) before storage. Add basic format validation. Email verification can come later.

## Performance Bottlenecks

**N+1 query pattern in shopping list generation:**
- Problem: `GetShoppingList` fetches each recipe individually inside a loop over menu days. For a 5-day menu, this makes 5 separate `SELECT` queries to the recipes table.
- Files: `backend/internal/services/shopping_service.go:92-100`
- Cause: Each `day.RecipeID` triggers a separate `recipeStorage.GetByID(day.RecipeID)` call.
- Improvement path: Batch-fetch all recipe IDs in a single query: `SELECT ... FROM recipes WHERE id IN (?, ?, ?, ?, ?)`. For a 5-day menu this is minor, but it's a pattern to avoid as the system grows.

**Tjek API sequential catalog fetching:**
- Problem: `SearchOffers` and `GetTopDiscounts` fetch offers from each catalog sequentially. With 20+ grocery store catalogs, this means 20+ sequential HTTP requests to the Tjek API.
- Files: `backend/internal/services/tjek_service.go:324-355,398-429`
- Cause: Each catalog's hotspots are fetched in a `for` loop with no concurrency.
- Improvement path: Use `sync.WaitGroup` or `errgroup.Group` to fetch catalog offers concurrently (with a semaphore to limit concurrent requests to ~5).

**No caching for Tjek API responses:**
- Problem: Every call to the offers endpoints triggers fresh API calls to Tjek. Catalog data and offers change at most weekly.
- Files: `backend/internal/services/tjek_service.go`
- Cause: No caching layer exists.
- Improvement path: Add in-memory caching with TTL (e.g., 1 hour for catalogs, 15 minutes for offers). Use `sync.Map` or a simple cache struct with expiry.

## Fragile Areas

**Migration system relies on relative file paths:**
- Files: `backend/internal/storage/sqlite/db.go:52-58,74`
- Why fragile: Migration SQL files are read via relative paths (`migrations/001_create_users.sql`). The working directory must be the `backend/` directory at runtime. Tests work around this by `os.Chdir()` to the backend root.
- Common failures: Running the server from the wrong directory causes "file not found" on migrations. Tests that change working directory can interfere with each other.
- Safe modification: Always run the server from the `backend/` directory or the Docker container (which sets `WORKDIR /app` and copies migrations there).
- Test coverage: The test helper `setupTestDB` manually changes directory, which works but is fragile.

**Household creation during registration is not transactional:**
- Files: `backend/internal/services/auth_service.go:46-83`
- Why fragile: Registration creates a household, then a user, then adds the user as a member. If any step fails after the first succeeds, orphaned records are left in the database.
- Common failures: If `userStorage.Create` fails after `householdStorage.Create` succeeds, an empty household is left in the DB.
- Safe modification: Wrap the entire registration flow in a database transaction.
- Test coverage: Not tested for partial failure scenarios.

**Token expiry requires re-login; no refresh token mechanism:**
- Files: `backend/pkg/utils/jwt.go:24` (7-day expiry), `frontend/src/api/client.ts:42-61` (401 -> redirect to login)
- Why fragile: When the JWT expires after 7 days, the user is abruptly redirected to login with no warning. No refresh token flow exists.
- Common failures: Users lose their session after 7 days with no way to extend it.
- Safe modification: Implement a refresh token endpoint or extend expiry. Add a frontend interceptor that attempts refresh before redirecting.
- Test coverage: No tests for token expiry handling.

## Scaling Limits

**SQLite single-writer bottleneck:**
- Current capacity: SQLite handles one write at a time. For a family meal planning app with a handful of concurrent users, this is fine.
- Limit: Under ~100 concurrent write operations per second, SQLite will queue and slow down. Journal mode WAL helps for concurrent reads.
- Symptoms at limit: "database is locked" errors under high write concurrency.
- Scaling path: Enable WAL mode (`PRAGMA journal_mode=WAL`), or migrate to PostgreSQL when needed. The storage layer abstraction (concrete structs, not interfaces) would need refactoring to support this.

**Fly.io min_machines_running = 0 (cold starts):**
- Current capacity: The app scales to zero when not in use.
- Limit: First request after scale-down has a cold start delay (container boot + Go binary startup + SQLite open + migrations check).
- Symptoms at limit: 2-5 second delay on first request after idle period.
- Scaling path: Set `min_machines_running = 1` if cold starts are unacceptable. Current setting is fine for a personal/family project.

## Dependencies at Risk

**No risky dependencies detected.**
- `go-sqlite3 v1.14.33`: Actively maintained, stable CGO-based SQLite driver.
- `golang-jwt/jwt v5.3.0`: Standard JWT library, well-maintained.
- `google/uuid v1.6.0`: Stable, widely used.
- `golang.org/x/crypto v0.46.0`: Official Go supplementary crypto package.
- Frontend dependencies (Vue 3, Pinia 3, Vite 7, axios) are all current and actively maintained.

## Missing Critical Features

**No old menu cleanup / data retention:**
- Problem: Every `POST /menus/generate` creates a new menu. Old menus are never deleted. The `GetCurrent` query always returns the latest, but previous menus accumulate indefinitely.
- Current workaround: None. Data grows over time.
- Blocks: Storage growth on production. Shopping list items also accumulate per menu.
- Implementation complexity: Low. Add a cleanup that deletes menus older than N weeks, or limit to N menus per household.

**No password reset / forgot password flow:**
- Problem: If a user forgets their password, there is no way to recover the account.
- Current workaround: None. Account is lost.
- Blocks: Any real user adoption.
- Implementation complexity: Medium. Requires email sending infrastructure (e.g., SMTP or a service like Resend/SendGrid).

**No logging beyond stdout:**
- Problem: Errors are returned to clients but not logged server-side. When handlers return `http.Error(w, ..., 500)`, the actual error is silently swallowed.
- Current workaround: None. Debug by reading client error responses.
- Blocks: Debugging production issues.
- Implementation complexity: Low. Add `log.Printf` or structured logging (e.g., `slog`) before error responses.

## Test Coverage Gaps

**Frontend: Zero test files exist:**
- What's not tested: All Vue components, stores, composables, API services, and router guards.
- Files: Entire `frontend/src/` directory.
- Risk: UI regressions, broken state management, broken API integration go undetected. The mock system provides some manual testing support but no automated verification.
- Priority: Medium. Add Vitest + Vue Test Utils. Start with stores (`frontend/src/stores/user.ts`, `frontend/src/stores/menuGenerator.ts`) since they contain the most logic.
- Difficulty: Low for stores (pure logic), medium for components (need rendering).

**Backend: Only household service has tests:**
- What's not tested: `AuthService`, `RecipeService`, `MenuService`, `ShoppingService`, `TjekService`, all handlers, all storage layers, middleware, JWT utils.
- Files: Only `backend/internal/services/household_service_test.go` exists (482 lines, 13 tests - good coverage for that one service).
- Risk: Auth bugs (e.g., accepting empty passwords, JWT validation bypass), recipe creation validation errors, menu generation edge cases, and shopping list aggregation bugs would go undetected.
- Priority: High. Auth service and JWT utilities are the most critical to test. Shopping service aggregation logic is complex and benefits from unit tests.
- Difficulty: Low. The existing test file provides a good pattern to follow with `setupTestDB`.

**No integration/E2E tests:**
- What's not tested: Full request lifecycle (HTTP request -> handler -> service -> storage -> response).
- Risk: Handler-level bugs (wrong status codes, missing auth checks, response format issues) are not caught.
- Priority: Medium. Start with auth flow (register, login, access protected endpoint) and menu generation flow.
- Difficulty: Medium. Need to set up test HTTP server with test database.

---

*Concerns audit: 2026-02-06*
*Update as issues are fixed or new ones discovered*
