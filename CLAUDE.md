# CLAUDE.md — Måltiden

Swedish meal planning app for weekly menus and smart shopping lists.
Full-stack monorepo: **Go backend + Vue 3 frontend**, deployed on Fly.io.

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Backend | Go | 1.26 |
| Database | SQLite (go-sqlite3, CGO) | 1.14.41 |
| Auth | JWT (golang-jwt) + bcrypt | HS256, 7-day expiry |
| Frontend | Vue 3 + TypeScript strict | 3.5.x / 5.9.x |
| State | Pinia (Composition API) | 3.0.x |
| HTTP client | Axios | 1.13.x |
| Build | Vite | 7.3.x |
| Lint/Format | ESLint + Prettier | semi: false, singleQuote: true, printWidth: 100 |
| Animation | motion-v | 1.10.x |
| Testing | Vitest + @vue/test-utils + happy-dom | 4.0.x |
| Node | ^20.19.0 \|\| >=22.12.0 | LTS |
| AI | Claude API (Anthropic) | Recipe parsing |
| Hosting | Fly.io | Region: arn |

## Project Structure

```
maltiden/
├── backend/
│   ├── cmd/server/main.go              # Entry point, SPA serving
│   ├── internal/
│   │   ├── api/handlers/               # HTTP handlers (no business logic)
│   │   ├── api/router.go               # Routes + dependency injection
│   │   ├── domain/                     # Pure types, zero dependencies
│   │   ├── services/                   # Business logic & orchestration
│   │   └── storage/sqlite/             # Database access layer
│   ├── pkg/
│   │   ├── claude/client.go            # Claude API HTTP client
│   │   ├── middleware/{auth,cors,requestid,ratelimit}.go
│   │   └── utils/{jwt,password}.go     # Token + bcrypt helpers
│   └── migrations/                     # SQL migrations (001–007)
├── frontend/
│   └── src/
│       ├── api/                        # Axios client + typed API services
│       ├── components/
│       │   ├── common/                 # BaseButton, BaseCard, ProgressBar, etc.
│       │   ├── dashboard/              # HouseholdWidget, WeeklyMenuGrid, etc.
│       │   ├── landing/                # Hero, Features, CTA sections
│       │   ├── menu/                   # MenuDayCard, GenerateMenuEmptyState
│       │   ├── poc/                    # Proof-of-concept components (OfferSearch)
│       │   ├── recipe-parser/          # RecipeParseInput, RecipeEditForm, etc.
│       │   ├── recipes/                # RecipeCard, RecipeListPanel, AddRecipePanel
│       │   ├── skeleton/               # Skeleton loading components + layouts
│       │   └── vue-bits/               # Reusable animations (RotatingText, SpotlightCard, etc.)
│       ├── composables/                # Vue composables (useOptimistic, useProgressBar, useShoppingList, etc.)
│       ├── directives/                 # Custom Vue directives (vPrefetch)
│       ├── stores/                     # Pinia stores (user, dashboard, menu)
│       ├── utils/                      # Utility functions (token management)
│       ├── views/                      # Page-level components
│       ├── router/                     # Vue Router with auth guards
│       ├── mocks/                      # Mock data for offline dev
│       └── styles/                     # Global CSS + theming
├── docs/
│   ├── PROJECT.md                      # Architecture & git workflow
│   ├── API.md                          # Full API contract
│   ├── RECIPE_PARSER_PIPELINE.md       # Claude integration spec
│   ├── TJEK_API_INTEGRATION.md         # Grocery offers API
│   └── TODO.md                         # Sprint planning
├── .github/workflows/                  # CI, Claude review, API docs sync
├── Dockerfile                          # Multi-stage (Node build → Go build → slim runtime)
└── fly.toml                            # Fly.io deployment config
```

## Architecture — Backend

**Strict layered architecture with unidirectional dependency flow:**

```
Handler (api/handlers/)     ← HTTP glue only, no business logic
    ↓
Service (services/)         ← Business logic, orchestration, validation
    ↓
Storage (storage/sqlite/)   ← SQL queries, transactions
    ↓
Domain (domain/)            ← Pure types, NO imports from other layers
```

**Rules:**
- `domain/` must never import from `api/`, `services/`, or `storage/`
- Handlers decode requests, call services, encode responses — nothing else
- Services own all business rules and validation logic
- Storage handles raw SQL — no ORM, direct `database/sql` usage
- Dependencies are injected via constructors in `router.go`

**ID convention:** Prefixed UUIDs — `usr_`, `hh_`, `rec_`, `menu_`

## Architecture — Frontend

**Composition API everywhere. No Options API.**

- **Stores (Pinia):** One per domain (`user.ts`, `dashboard.ts`, `menuGenerator.ts`)
  - All stores use `defineStore` with setup function syntax
  - Actions include `isLoading` + `error` state management
- **API layer:** Typed services per domain (`auth.api.ts`, `household.api.ts`, etc.)
  - Centralized Axios instance with JWT interceptor and 401 auto-redirect
- **Components:** `<script setup lang="ts">` + scoped CSS
  - Base components in `components/common/` (BaseButton, etc.) with variant props
- **Router:** Auth guards via `meta.requiresAuth` and `meta.requiresMember`
- **Mock mode:** `npm run dev:mock` runs without backend using mock data

## Development Commands

### Backend
```bash
cd backend
go build ./...                    # Build
go test ./... -v -race            # Test with race detector
go vet ./...                      # Static analysis
JWT_SECRET=dev-secret-for-local-development go run cmd/server/main.go   # Run locally
```

### Frontend
```bash
cd frontend
npm ci                            # Install deps (use ci, not install)
npm run dev                       # Dev server (real API at localhost:8080)
npm run dev:mock                  # Dev server (mock data, no backend needed)
npm run type-check                # TypeScript type checking (vue-tsc)
npm run lint                      # ESLint with auto-fix
npm run format                    # Prettier formatting
npm run build                     # Production build (type-check + vite build)
npm run test                      # Run tests (vitest)
```

### CI Pipeline (must pass before merge)
**Backend:** build → test (race) → vet → govulncheck
**Frontend:** npm ci → type-check → lint → build

## Environment Variables

### Backend
| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP listen port |
| `DATABASE_PATH` | No | `./data/maltiden.db` | SQLite file path |
| `JWT_SECRET` | **Yes** | — | JWT signing key |
| `ANTHROPIC_API_KEY` | No | — | Claude API key (parser degrades gracefully without it) |
| `CORS_ORIGINS` | No | `http://localhost:5173,...` | Comma-separated allowed origins |

### Frontend
| Variable | Description |
|----------|-------------|
| `VITE_API_URL` | Backend URL (default: `http://localhost:8080`) |
| `VITE_USE_REAL_API` | `true` = real backend, `false` = mock data |

## Code Conventions

### Backend (Go)
- **File naming:** `user.go` (domain), `recipe_service.go` (service), `recipe_storage.go` (storage), `recipes.go` (handler)
- **Constructors:** `func NewXxxService(deps...) *XxxService` — return struct pointer
- **Error handling:** Return `error` as last value. Check with `if err != nil`. Use `errors.New("snake_case_code")` for business errors
- **JSON tags:** `camelCase` in JSON, use `json:"-"` to hide sensitive fields (e.g., password hash)
- **Transactions:** `tx, err := db.Begin()` → `defer tx.Rollback()` → operations → `tx.Commit()`
- **HTTP responses:** `json.NewEncoder(w).Encode(data)` with explicit `w.WriteHeader(status)` and `Content-Type: application/json`
- **Error responses:** `http.Error(w, {"error":"code"}, status)` — JSON format
- **Context values:** `middleware.GetUserID(r)` and `middleware.GetHouseholdID(r.Context())`
- **Route format:** Go 1.22+ method patterns: `"GET /path"`, `"POST /path/{param}"`

### Frontend (TypeScript/Vue)
- **File naming:** `PascalCase.vue` (components/views), `lowercase.api.ts` (API), `lowercase.ts` (stores/utils), `lowercase.types.ts` (types)
- **Components:** Always `<script setup lang="ts">` — no Options API
- **Props:** `withDefaults(defineProps<Props>(), { ... })`
- **Types:** Interfaces for data shapes, union types for enums (`'owner' | 'member' | 'guest'`)
- **Type guards:** `export function isApiError(error: unknown): error is ApiError`
- **Imports:** Use `@/` path alias, `import type` for type-only imports
- **Unused vars:** Prefix with `_` (ESLint rule: `argsIgnorePattern: '^_'`)
- **CSS:** Scoped styles per component (`<style scoped>`)
- **Reactivity:** `ref()` for primitives, `computed()` for derived state, `reactive()` sparingly

## Auth System

- **Registration:** email + password → bcrypt hash (cost 12) → create user + household → return JWT
- **Login:** email + password → verify hash → return JWT with `userID` + `householdID` claims
- **Token:** 7-day expiry, HS256 signing, stored in localStorage on frontend
- **Middleware:** `RequireAuth` extracts claims, injects into request context
- **Roles:** `owner` (full + delete household), `member` (full access), `guest` (read-only)
- **Frontend guard:** `router.beforeEach` checks `meta.requiresAuth`, redirects to `/login`

## Database

- **Engine:** SQLite with `go-sqlite3` (requires CGO)
- **Migrations:** Sequential SQL files in `backend/migrations/` (001–007), auto-applied on startup
- **Tables:** `users`, `households`, `household_members`, `invite_codes`, `recipes`, `menus`, `menu_days`, `shopping_items`
- **Note:** Ingredients and instructions are stored as JSON columns in the `recipes` table, not separate tables
- **No ORM:** Direct `database/sql` with `QueryRow`, `Query`, `Exec`, manual `Scan`
- **Indexes:** On `email`, `household_id`, and other frequently queried columns

## API Routes

**Public:** `GET /health`, `POST /auth/register`, `POST /auth/login`, `GET /recipes`, `GET /recipes/{id}`, `GET /offers/search`, `GET /offers/discounts`, `GET /offers/stores`

**Protected (JWT required):** `GET /households/me`, `POST /households/invite`, `POST /households/join`, `*/households/members/*`, `POST /recipes`, `POST /recipes/parse`, `POST /recipes/parse-and-save`, `POST /menus/generate`, `GET /menus/current`, `GET /shopping-list`, `PATCH /shopping-list/items/{id}`

See `docs/API.md` for full request/response contracts.

## Key Integrations

### Claude API (Recipe Parser)
- Converts unstructured recipe text → structured JSON with ingredients, instructions, tags
- Uses `claude-sonnet-4.5-20250929` with structured output
- Optional — app runs without `ANTHROPIC_API_KEY`
- Full spec: `docs/RECIPE_PARSER_PIPELINE.md`

### Tjek API (Grocery Offers)
- Fetches Swedish grocery store offers from etilbudsavis.dk
- No auth required
- Word-boundary matching for ingredient → offer linking
- Full spec: `docs/TJEK_API_INTEGRATION.md`

## Git Workflow

- **Branches:** `main` (production), `dev` (development), `feat/be_*` (backend), `feat/fe_*` (frontend)
- **Merge strategy:** Rebase only — no merge commits
- **Flow:** feature branch from `dev` → develop → `git rebase dev` → `git merge --ff-only` into `dev`
- **CI triggers:** Push/PR to `main` and `dev`

## Principles

- **Scale-ready:** Monolith with clean layer separation, prepared for extraction
- **No over-engineering:** Solve the current problem, not hypothetical future ones
- **Concrete code over theory:** Ship working code, not abstractions
- **Swedish UI, English code:** User-facing text in Swedish, code/comments/commits in English
- **Type safety everywhere:** TypeScript strict mode, Go's type system, typed API contracts
- **Graceful degradation:** Features like recipe parser work when available, app functions without
- **Security first:** bcrypt, JWT, input validation, CORS, OWASP awareness

## Version Control — Never Guess Versions

Always verify latest versions before modifying workflows, Dockerfiles, or configs:
```bash
gh api repos/actions/checkout/releases/latest --jq '.tag_name'
gh api repos/actions/setup-go/releases/latest --jq '.tag_name'
gh api repos/actions/setup-node/releases/latest --jq '.tag_name'
head -3 backend/go.mod
jq '.engines' frontend/package.json
```

## Documentation Reference

| Doc | Purpose |
|-----|---------|
| `docs/PROJECT.md` | Architecture, team, decisions (Swedish) |
| `docs/API.md` | Full REST API contract with examples |
| `docs/RECIPE_PARSER_PIPELINE.md` | Claude integration spec (36KB, comprehensive) |
| `docs/TJEK_API_INTEGRATION.md` | Grocery offers integration |
| `docs/TODO.md` | Sprint planning and feature roadmap |
