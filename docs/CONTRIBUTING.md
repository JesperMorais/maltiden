# Contributing to Måltiden

Conventions for contributors. See `CLAUDE.md` for the full reference on architecture, code style, and tooling.

## Branch naming

- Feature branches: `feat/be_*` (backend), `feat/fe_*` (frontend)
- Bug fixes: `fix/*`
- Always branch off `dev`, never `main`
- Merge strategy: **rebase only** — no merge commits

## Commit style

- One-line messages preferred; English only
- One logical purpose per commit — unrelated changes go in separate commits
- Never use `git add -A` or `git add .` — stage files explicitly
- Never reference AI tools in commit messages

## PR checklist

- [ ] Branched from `dev`
- [ ] Backend passes: `go build ./... && go test ./... -race`
- [ ] Frontend passes: `npm run type-check && npm run lint && npm run build`
- [ ] No secrets, credentials, or `.env` files committed
- [ ] Swedish UI text, English code and comments
- [ ] Rebased on latest `dev` before review
