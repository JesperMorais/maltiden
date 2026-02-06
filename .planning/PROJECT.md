# Maltiden Backend Code Review

## What This Is

A comprehensive code review of the Maltiden backend (Go 1.24, SQLite, JWT auth) covering security, correctness, architecture, performance, and code quality. The output is a detailed findings report with a prioritized fix plan.

## Core Value

Surface every actionable issue in the backend — bugs, security gaps, architectural debt, performance problems — and produce a prioritized plan to fix them.

## Requirements

### Validated

- ✓ Codebase mapped — `.planning/codebase/` with 7 analysis documents — existing
- ✓ Known concerns documented — `.planning/codebase/CONCERNS.md` captures initial findings — existing

### Active

- [ ] Deep review of all backend handlers for correctness, error handling, and auth coverage
- [ ] Security audit: JWT handling, input validation, auth middleware gaps, request size limits
- [ ] Architecture review: layering violations, dependency patterns, error propagation
- [ ] Performance review: N+1 queries, sequential API calls, missing caching, sorting algorithms
- [ ] Code quality: idiomatic Go patterns, consistency across handlers, dead code, duplication
- [ ] Findings report with severity ratings (critical/high/medium/low)
- [ ] Prioritized fix plan organized by severity and effort

### Out of Scope

- Frontend code — backend-only review
- Adding new features — review and plan only, no code changes
- Test writing — identified as a gap but not part of this review scope
- Infrastructure migration (e.g., PostgreSQL) — note as recommendation only

## Context

- Backend is a Go REST API serving a family meal planning app (Maltiden)
- Layered architecture: handlers → services → storage (SQLite)
- JWT auth with middleware, bcrypt passwords, prefixed UUID entity IDs
- External dependency on Tjek API for grocery offers/discounts
- Deployed on Fly.io with Docker, scales to zero
- Existing codebase map in `.planning/codebase/` already identifies several concerns
- Only one test file exists (`household_service_test.go`) — test coverage is thin
- CONCERNS.md provides a starting point but the deep review should go file-by-file

## Constraints

- **Scope**: Backend code only (`backend/` directory) plus deployment config (`Dockerfile`, `fly.toml`)
- **Output**: Report + prioritized fix plan, no code changes during review
- **Tech stack**: Go 1.24, SQLite, stdlib `net/http` — review against idiomatic patterns for these

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Review backend only, exclude frontend | Focused scope, manageable context | — Pending |
| Report + fix plan, no immediate code changes | Avoid mixing review with implementation | — Pending |
| Include deployment config in scope | Fly.io config and Dockerfile affect backend behavior | — Pending |

---
*Last updated: 2026-02-06 after initialization*
