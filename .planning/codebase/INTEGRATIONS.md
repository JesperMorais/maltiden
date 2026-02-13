# External Integrations

**Analysis Date:** 2026-02-06

## APIs & External Services

**Tjek / eTilbudsavis API (Grocery Offers):**
- Purpose: Fetch grocery store catalogs, offers/discounts, and store locations in Sweden
- Base URL: `https://api.etilbudsavis.dk/v2`
- SDK/Client: Custom Go HTTP client (`backend/internal/services/tjek_service.go`)
- Auth: None (public API, no API key)
- Endpoints consumed:
  - `GET /catalogs` - Weekly flyers near a location (lat/lng/radius)
  - `GET /catalogs/{id}/hotspots` - Offers within a specific catalog
  - `GET /stores` - Physical store locations near a location
- Timeout: 30 seconds per request
- Backend routes exposing this:
  - `GET /offers/search` - Search offers by keyword (`backend/internal/api/handlers/offers.go`)
  - `GET /offers/discounts` - Top discounted offers sorted by discount %
  - `GET /offers/stores` - List of available grocery stores
- Filtering: Swedish grocery stores only (Willys, ICA, Coop, Lidl, Hemkop, etc.) - hardcoded lists in `backend/internal/services/tjek_service.go`

**No other external APIs detected.** No payment providers, analytics, email services, or third-party auth providers.

## Data Storage

**Database:**
- SQLite 3 via `github.com/mattn/go-sqlite3` (CGO driver)
- Connection: `DATABASE_PATH` env var, default `./data/maltiden.db` (`backend/cmd/server/main.go`)
- Client: Go stdlib `database/sql` (`backend/internal/storage/sqlite/db.go`)
- ORM: None - raw SQL queries in storage layer

**Schema (6 migrations in `backend/migrations/`):**
- `001_create_users.sql` - Users table with email, password_hash, household_id
- `002_create_households.sql` - Households and household_members tables
- `003_create_recipes.sql` - Recipes with ingredients and instructions (JSON columns)
- `004_seed_recipes.sql` - Seed data with Swedish recipes
- `005_create_menus.sql` - Weekly menus with menu_days
- `006_household_invites_and_member_status.sql` - Invite codes, member eating/lunchbox status

**Migration System:**
- Custom file-based migrations tracked in `schema_migrations` table (`backend/internal/storage/sqlite/db.go`)
- Migrations run automatically on server startup
- Migration files loaded from `migrations/` directory relative to working directory

**File Storage:** None - no file uploads or blob storage

**Caching:** None - no Redis, Memcached, or in-memory cache layer

## Authentication & Identity

**Auth Provider:** Custom JWT-based authentication (no external auth service)

**Implementation:**
- Password hashing: bcrypt (cost 12) via `golang.org/x/crypto/bcrypt` (`backend/pkg/utils/password.go`)
- Token: JWT HS256 with 7-day expiry (`backend/pkg/utils/jwt.go`)
- JWT claims: `user_id`, `household_id`, standard registered claims
- JWT secret: `JWT_SECRET` env var (read once at package init)
- Middleware: `RequireAuth` extracts Bearer token, validates, injects user_id/household_id into request context (`backend/pkg/middleware/auth.go`)
- Frontend: Token stored in localStorage under key `maltiden_token` (`frontend/src/utils/token.ts`)
- Frontend interceptor: Axios request interceptor adds `Authorization: Bearer <token>` to all requests (`frontend/src/api/client.ts`)
- Auto-logout: Axios response interceptor redirects to `/login` on 401 responses

**Role System:**
- Roles: owner, member, guest (stored in `household_members` table)
- Frontend route guards: `requiresAuth` and `requiresMember` meta fields (`frontend/src/router/index.ts`)
- Role checked via `useUserStore` computed properties: `isOwner`, `isMember`, `isGuest`

## Monitoring & Observability

**Error Tracking:** None - no Sentry, Datadog, or error reporting service

**Logs:**
- Backend: Go stdlib `log` package, basic Printf/Fatal logging (`backend/cmd/server/main.go`)
- Frontend: `console.error` for network errors (`frontend/src/api/client.ts`)
- No structured logging framework

**Health Check:**
- `GET /health` endpoint (`backend/internal/api/handlers/health.go`)

## CI/CD & Deployment

**Hosting:** Fly.io
- App name: `maltiden`
- Region: `arn` (Stockholm/Arlanda)
- Config: `fly.toml`
- VM: Shared CPU, 1 core, 256MB RAM
- Auto-stop/start machines, min 0 running
- Force HTTPS

**Container:**
- Multi-stage Dockerfile at project root (`Dockerfile`)
- Builder: `golang:1.24-bookworm`
- Runtime: `debian:bookworm-slim` with ca-certificates
- Only backend binary deployed; frontend not built into container (likely served separately or not yet integrated)

**CI Pipeline:** GitHub Actions (`.github/workflows/ci.yml`)
- Triggers: push/PR to `main` and `dev` branches
- Backend job: Go build, test (with -race), vet, govulncheck
- Frontend job: npm ci, type-check, lint, build
- Action versions: checkout@v6, setup-go@v6, setup-node@v6

**Claude AI Workflows:**
- `claude.yml` - Claude Code action for @claude mentions in issues/PRs
- `claude-code-review.yml` - Auto PR review on open/sync with progress tracking
- `claude-api-docs-sync.yml` - Auto-update docs when API files change

**Dependency Management:**
- Dependabot (`.github/dependabot.yml`) - Weekly updates for gomod, npm, github-actions
- Auto-labeling PRs by changed path (`.github/labeler.yml`, `.github/workflows/labeler.yml`)

## Environment Configuration

**Required env vars (backend):**
- `JWT_SECRET` - HMAC key for JWT signing (CRITICAL - must be set in production)
- `PORT` - HTTP listen port (optional, default `8080`)
- `DATABASE_PATH` - SQLite file path (optional, default `./data/maltiden.db`)

**Required env vars (frontend build):**
- `VITE_API_URL` - Backend API base URL
- `VITE_USE_REAL_API` - `true` for real backend, `false` for mock data

**Secrets location:**
- CI: GitHub repository secrets (`CLAUDE_CODE_OAUTH_TOKEN` for Claude workflows, `JWT_SECRET` for tests)
- Production: Fly.io secrets (assumed, not visible in config)
- Local dev: Not committed - no `.env` in git (only `.env.development` and `.env.mock` for frontend Vite config)

## Webhooks & Callbacks

**Incoming:**
- None - no webhook endpoints

**Outgoing:**
- None - no outbound webhook calls

## Mock Data System

**Frontend Mock Layer:**
- Toggle: `VITE_USE_REAL_API` env var controls mock vs real API (`frontend/src/mocks/index.ts`)
- Pattern: Each API module checks `USE_MOCKS` and returns mock data instead of HTTP calls
- Mock files: `frontend/src/mocks/*.mock.ts` (auth, dashboard, household, landing, menu, offers, recipes, shopping)
- Dev command: `npm run dev:mock` for mock mode, `npm run dev` for real API

---

*Integration audit: 2026-02-06*
*Update when adding/removing external services*
