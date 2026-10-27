# Måltiden

Weekly meal planning for Swedish households. Plan a week of dinners, get a
shopping list that merges the week's ingredients, and cook from your phone.
Household members share one menu and one list.

Mobile-first (built for the phone in your kitchen), with a desktop view for
planning. The UI is in Swedish.

**Live:** [måltiden.se](https://måltiden.se/)

## Features

- Household accounts with invite codes and roles (owner / member / guest)
- Menu generation from a shared recipe catalogue plus your own recipes, with
  per-household preferences (excluded tags, disliked ingredients, vegetarian
  days, batch cooking)
- Shopping list built from the menu, with checkable and custom items
- Optional AI helpers: paste a recipe and have it structured (Claude), or
  describe wishes for the week and have them applied (Gemini). Both are off
  unless an API key is configured, and both have daily caps.
- Self-service privacy: data export and account deletion under *Mitt konto*,
  plus a privacy policy at `/integritet`

## Stack

- **Backend:** Go 1.26, `net/http`, SQLite (`mattn/go-sqlite3`, needs CGO and a C compiler)
- **Frontend:** Vue 3 + TypeScript, Vite, Pinia, Vue Router
- **Tests:** `go test`, Vitest, Playwright
- **Hosting:** Fly.io (one machine plus a volume). See [docs/deploy.md](docs/deploy.md).

## Local development

Requirements: Go 1.26+, a C toolchain (for SQLite), and Node 20.19+ or 22.12+.

**Backend** (API on `http://localhost:8080`):

```sh
cd backend
env JWT_SECRET=local-dev-secret-at-least-32-characters go run ./cmd/server
```

On first start, the server creates `backend/data/maltiden.db` and runs all
migrations, including the seed recipes. `JWT_SECRET` is required and must be
at least 32 characters. Any value works locally. The AI features stay off
unless you also set `ANTHROPIC_API_KEY` / `GEMINI_API_KEY`. The full list of
settings is in [docs/deploy.md](docs/deploy.md#configuration).

**Frontend** (on `http://localhost:5173`):

```sh
cd frontend
npm install
npm run dev        # talks to the backend on :8080
npm run dev:mock   # or: mocked API, no backend needed
```

## Tests and checks

These are the same checks CI runs (`.github/workflows/ci.yml`):

```sh
cd backend
env JWT_SECRET=ci-test go test -race ./...
go vet ./...
gofmt -l .          # must print nothing
```

```sh
cd frontend
npm run type-check
npm run lint
npm test
npm run build
```

End-to-end tests (Playwright, against the mocked frontend; this is what the
Playwright workflow runs on PRs), from the repo root:

```sh
npm install
npx playwright install chromium
npx playwright test
```

`tests/e2e/` has a second suite that runs against the real Go backend
(`npx playwright test --config=tests/e2e/playwright.e2e.config.ts`). It is not
part of CI, and some of its assertions are out of date.

## Project structure

- `backend/`: Go API. `cmd/server` is the entrypoint, `internal/` holds the
  handlers, services and SQLite storage, and `migrations/` has the SQL
  migrations and seed data.
- `frontend/`: the Vue 3 app
- `tests/`: Playwright specs
- `scripts/`: dev and CI helpers
- `tools/`: one-off data tooling
- `docs/`: [deployment](docs/deploy.md), [recipe provenance](docs/recipe-provenance.md)

## Contributing

Branch off `dev` and open PRs against `dev`. `main` is production: every
merge to `main` deploys to Fly automatically. Keep backend code `gofmt`-clean
and make sure CI is green.

About seed data: part of the shared recipe catalogue comes from a third-party
site and is scheduled for replacement. Don't add scraped content. See
[docs/recipe-provenance.md](docs/recipe-provenance.md).

## License

There is no license file yet, so the code is not open source for now: all
rights are reserved by the authors until the maintainers choose a license.
