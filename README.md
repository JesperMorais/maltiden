# Måltiden

Weekly meal planning for Swedish households — plan a week of dinners, get a smart shopping list, cook from your phone.

Mobile-first (built for the phone in your kitchen), with a desktop view for planning.

## Stack

- **Backend:** Go + SQLite
- **Frontend:** Vue 3 + TypeScript (Pinia, Vue Router)
- **Hosting:** Fly.io

## Getting started

**Frontend**

```bash
cd frontend
npm install
npm run dev        # or: npm run dev:mock  (mocked API, no backend needed)
```

**Backend**

```bash
cd backend
go run ./cmd/server
```

## Project structure

- `backend/` — Go API + SQLite
- `frontend/` — Vue 3 app
- `scripts/` — dev + CI helpers
- `tests/` — end-to-end tests

## Live

[måltiden.se](https://måltiden.se/)
