# Roadmap: Maltiden Backend

## Overview

Systematic improvement of the Maltiden backend — from code review through implementation of all identified fixes. The v1.0 milestone reviewed the codebase and produced findings; v1.1 implements the prioritized fixes from that review.

## Milestones

- ✅ **v1.0 Pre-Fix** — Phases 1-3 (shipped 2026-02-09)
- 🚧 **v1.1 Fix** — Phases 4-7 (in progress)

## Phases

<details>
<summary>✅ v1.0 Pre-Fix (Phases 1-3) — SHIPPED 2026-02-09</summary>

- [x] Phase 1: Deep Code Review (2/2 plans) — completed 2026-02-09
- [x] Phase 2: Architecture & Performance Review (2/2 plans) — completed 2026-02-09
- [x] Phase 3: Findings Report & Fix Plan (2/2 plans) — completed 2026-02-09

See [milestone archive](milestones/v1.0-pre-fix-ROADMAP.md) for full details.

</details>

### 🚧 v1.1 Fix (In Progress)

**Milestone Goal:** Implement all critical, high, and medium-severity fixes from the backend code review (93 findings across 9 FIX-PLAN batches)

#### Phase 4: Critical Security & Data Integrity
**Goal**: Fix all authentication bypass vulnerabilities, IDOR issues, transaction safety, and database integrity (FIX-PLAN Batches 1-2)
**Depends on**: v1.0 Pre-Fix milestone complete
**Research**: Unlikely (established Go patterns, internal code changes)
**Plans**: TBD

Plans:
- [ ] 04-01: TBD (run /gsd:plan-phase 4 to break down)

#### Phase 5: Error Handling & Input Validation
**Goal**: Establish error infrastructure (sentinel errors, error response helpers) and sweep all input validation gaps (FIX-PLAN Batches 3-4)
**Depends on**: Phase 4
**Research**: Unlikely (internal patterns, Go error conventions)
**Plans**: TBD

Plans:
- [ ] 05-01: TBD (run /gsd:plan-phase 5 to break down)

#### Phase 6: Performance & Database
**Goal**: Fix external API performance (Tjek caching, concurrent fetching) and database configuration (WAL mode, connection pool, N+1 queries) (FIX-PLAN Batches 5-6)
**Depends on**: Phase 4 (FK enforcement enables reliable query optimization)
**Research**: Unlikely (SQLite configuration, Go concurrency patterns)
**Plans**: TBD

Plans:
- [ ] 06-01: TBD (run /gsd:plan-phase 6 to break down)

#### Phase 7: Architecture, Deployment & Polish
**Goal**: Storage interfaces for testability, DI improvements, deployment config fixes, observability, and remaining medium/low findings (FIX-PLAN Batches 7-9)
**Depends on**: Phases 5-6 (error infrastructure and performance foundations)
**Research**: Unlikely (Go interface patterns, fly.io configuration)
**Plans**: TBD

Plans:
- [ ] 07-01: TBD (run /gsd:plan-phase 7 to break down)

## Domain Expertise

None

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Deep Code Review | v1.0 | 2/2 | ✅ Complete | 2026-02-09 |
| 2. Architecture & Performance Review | v1.0 | 2/2 | ✅ Complete | 2026-02-09 |
| 3. Findings Report & Fix Plan | v1.0 | 2/2 | ✅ Complete | 2026-02-09 |
| 4. Critical Security & Data Integrity | v1.1 | 0/? | Not started | - |
| 5. Error Handling & Input Validation | v1.1 | 0/? | Not started | - |
| 6. Performance & Database | v1.1 | 0/? | Not started | - |
| 7. Architecture, Deployment & Polish | v1.1 | 0/? | Not started | - |
