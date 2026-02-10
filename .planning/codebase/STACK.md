# Technology Stack

**Analysis Date:** 2026-02-06

## Languages

**Primary:**
- Go 1.24.1 - Backend API server (`backend/go.mod`)
- TypeScript ~5.9.3 - Frontend SPA with strict mode (`frontend/package.json`, `frontend/tsconfig.app.json`)

**Secondary:**
- SQL - Database migrations (`backend/migrations/*.sql`)
- CSS - Global theme styles (`frontend/src/styles/theme.css`)
- HTML - Entry point shell (`frontend/index.html`)

## Runtime

**Environment:**
- Go 1.24.1 (backend) - compiled to single binary with CGO_ENABLED=1 for SQLite
- Node.js ^20.19.0 || >=22.12.0 (frontend dev/build) - specified in `frontend/package.json` engines field
- CI uses Node 22, Go 1.24

**Package Manager:**
- npm (frontend) - lockfile present at `frontend/package-lock.json`
- Go modules (backend) - lockfile present at `backend/go.sum`

## Frameworks

**Core:**
- Vue 3 ^3.5.26 - Frontend SPA framework (`frontend/package.json`)
- Vue Router ^5.0.2 - Client-side routing with auth guards (`frontend/src/router/index.ts`)
- Pinia ^3.0.4 - State management (`frontend/src/stores/`)
- Go stdlib `net/http` - HTTP server, no external web framework (`backend/cmd/server/main.go`)
- Go stdlib `http.NewServeMux` with method routing (`backend/internal/api/router.go`)

**Testing:**
- Go stdlib `testing` package - Backend tests (`backend/internal/services/household_service_test.go`)
- No frontend test framework installed (no vitest, jest, or cypress in `frontend/package.json`)

**Build/Dev:**
- Vite ^7.3.0 - Frontend build tool and dev server (`frontend/vite.config.ts`)
- vue-tsc ^3.2.4 - TypeScript type checking for Vue (`frontend/package.json`)
- ESLint ^9.39.2 with `eslint-plugin-vue` ~10.7.0 - Linting (`frontend/eslint.config.ts`)
- Prettier 3.8.1 - Code formatting (`frontend/package.json`)
- Docker - Production container build (`Dockerfile`)

## Key Dependencies

**Critical (Backend):**
- `github.com/mattn/go-sqlite3` v1.14.33 - SQLite database driver (CGO) (`backend/go.mod`)
- `github.com/golang-jwt/jwt/v5` v5.3.0 - JWT authentication (`backend/pkg/utils/jwt.go`)
- `github.com/google/uuid` v1.6.0 - UUID generation for entity IDs (`backend/internal/services/auth_service.go`)
- `golang.org/x/crypto` v0.46.0 - bcrypt password hashing (`backend/pkg/utils/password.go`)

**Critical (Frontend):**
- `axios` ^1.13.2 - HTTP client with interceptors for JWT auth (`frontend/src/api/client.ts`)
- `vue` ^3.5.26 - Core UI framework
- `pinia` ^3.0.4 - Stores: user, dashboard, landing, theme, counter, menuGenerator (`frontend/src/stores/`)
- `vue-router` ^5.0.2 - Route guards with auth/member checks (`frontend/src/router/index.ts`)

**Infrastructure (Frontend DevDeps):**
- `@vitejs/plugin-vue` ^6.0.3 - Vue SFC support in Vite
- `vite-plugin-vue-devtools` ^8.0.6 - Dev tooling
- `@vue/eslint-config-typescript` ^14.6.0 - TS-aware ESLint for Vue
- `@vue/eslint-config-prettier` ^10.2.0 - Prettier integration
- `@tsconfig/node24` ^24.0.3 - TypeScript base config for Node tooling
- `@vue/tsconfig` ^0.8.1 - TypeScript base config for Vue app
- `npm-run-all2` ^8.0.4 - Parallel npm script execution
- `jiti` ^2.6.1 - TypeScript config file loader

## Configuration

**Environment:**
- Backend: `PORT` (default 8080), `DATABASE_PATH` (default `./data/maltiden.db`), `JWT_SECRET` (env var, required) - configured in `backend/cmd/server/main.go` and `backend/pkg/utils/jwt.go`
- Frontend: `VITE_API_URL` (default `http://localhost:8080`), `VITE_USE_REAL_API` (true/false) - configured in `frontend/.env.development` and `frontend/.env.mock`
- Mock mode: `npm run dev:mock` uses `frontend/.env.mock` with `VITE_USE_REAL_API=false`
- Real API mode: `npm run dev` uses `frontend/.env.development` with `VITE_USE_REAL_API=true`

**Build:**
- `frontend/vite.config.ts` - Vite config with `@` path alias to `./src`
- `frontend/tsconfig.json` - Project references to `tsconfig.node.json` and `tsconfig.app.json`
- `frontend/tsconfig.app.json` - Extends `@vue/tsconfig/tsconfig.dom.json`, paths alias `@/*` -> `./src/*`
- `frontend/tsconfig.node.json` - Extends `@tsconfig/node24/tsconfig.json`, ESNext modules
- `frontend/eslint.config.ts` - Flat config with Vue essential + TS recommended, no-unused-vars with `^_` ignore pattern
- `Dockerfile` - Multi-stage: `golang:1.24-bookworm` builder, `debian:bookworm-slim` runtime
- `fly.toml` - Fly.io deployment config, app name `maltiden`, region `arn`

## Platform Requirements

**Development:**
- Go 1.24+ with CGO support (for go-sqlite3)
- Node.js ^20.19.0 || >=22.12.0 with npm
- GCC/C compiler toolchain (required by CGO for SQLite)
- No Docker required for local dev (backend runs as Go binary, frontend via Vite)

**Production:**
- Fly.io hosting with Docker container
- Debian Bookworm slim runtime with ca-certificates
- Shared CPU (1 core), 256MB memory allocation (`fly.toml`)
- Internal port 8080, force HTTPS, auto-stop/start machines
- SQLite database file at configurable path (data persistence via Fly volumes assumed)

---

*Stack analysis: 2026-02-06*
*Update after major dependency changes*
