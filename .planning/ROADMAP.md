# Roadmap: Maltiden Backend

## Overview

Systematic improvement of the Maltiden backend — from code review through implementation of all identified fixes. The v1.0 milestone reviewed the codebase and produced findings; v1.1 implements the prioritized fixes from that review.

## Milestones

- ✅ **v1.0 Pre-Fix** — Phases 1-3 (shipped 2026-02-09)
- ✅ **v1.1 Fix** — Phases 4-7 (shipped 2026-02-10)

## Phases

<details>
<summary>✅ v1.0 Pre-Fix (Phases 1-3) — SHIPPED 2026-02-09</summary>

- [x] Phase 1: Deep Code Review (2/2 plans) — completed 2026-02-09
- [x] Phase 2: Architecture & Performance Review (2/2 plans) — completed 2026-02-09
- [x] Phase 3: Findings Report & Fix Plan (2/2 plans) — completed 2026-02-09

See [milestone archive](milestones/v1.0-pre-fix-ROADMAP.md) for full details.

</details>

### ✅ v1.1 Fix (Shipped 2026-02-10)

**Milestone Goal:** Implement all critical, high, and medium-severity fixes from the backend code review (93 findings across 9 FIX-PLAN batches)

#### Phase 4: Critical Security & Data Integrity ✅
**Goal**: Fix all authentication bypass vulnerabilities, IDOR issues, transaction safety, and database integrity (FIX-PLAN Batches 1-2)
**Depends on**: v1.0 Pre-Fix milestone complete
**Verified**: 10/10 must-haves passed

Plans:
- [x] 04-01: JWT Auth Hardening (AUTH-01, AUTH-02, AUTH-03)
- [x] 04-02: IDOR Protection & Transaction Safety (AUTH-04, DATA-01, DATA-02)
- [x] 04-03: Database Integrity (DATA-03, DATA-04, DATA-05, DATA-06, DATA-09)

#### Phase 5: Error Handling & Input Validation ✅
**Goal**: Establish error infrastructure (sentinel errors, error response helpers) and sweep all input validation gaps (FIX-PLAN Batches 3-4)
**Depends on**: Phase 4
**Verified**: 12/12 must-haves passed

Plans:
- [x] 05-01: Error Infrastructure (ERROR-01, ERROR-02, ERROR-03, VALID-01)
- [x] 05-02: Input Validation Sweep (VALID-02 through VALID-16)

#### Phase 6: Performance & Database ✅
**Goal**: Fix external API performance (Tjek caching, concurrent fetching) and database configuration (WAL mode, connection pool, N+1 queries) (FIX-PLAN Batches 5-6)
**Depends on**: Phase 4 (FK enforcement enables reliable query optimization)
**Verified**: 14/14 must-haves passed

Plans:
- [x] 06-01: Tjek API Optimization (PERF-02, PERF-03, PERF-08, PERF-09, PERF-10)
- [x] 06-02: Database Configuration & Query Optimization (PERF-01, PERF-04, PERF-05, PERF-06, PERF-12)

#### Phase 7: Architecture, Deployment & Polish ✅
**Goal**: Storage interfaces for testability, DI improvements, deployment config fixes, observability, and remaining medium/low findings (FIX-PLAN Batches 7-9)
**Depends on**: Phases 5-6 (error infrastructure and performance foundations)
**Verified**: 17/17 must-haves passed

Plans:
- [x] 07-01: Architecture & Testability (ARCH-01, ARCH-02, ARCH-03, ARCH-04, ARCH-09, ARCH-13)
- [x] 07-02: Deployment & Observability (DEPLOY-01 through DEPLOY-10, AUTH-06)
- [x] 07-03: Code Quality & Polish (AUTH-11, PERF-16, DATA-10)

## Domain Expertise

None

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Deep Code Review | v1.0 | 2/2 | ✅ Complete | 2026-02-09 |
| 2. Architecture & Performance Review | v1.0 | 2/2 | ✅ Complete | 2026-02-09 |
| 3. Findings Report & Fix Plan | v1.0 | 2/2 | ✅ Complete | 2026-02-09 |
| 4. Critical Security & Data Integrity | v1.1 | 3/3 | ✅ Complete | 2026-02-09 |
| 5. Error Handling & Input Validation | v1.1 | 2/2 | ✅ Complete | 2026-02-09 |
| 6. Performance & Database | v1.1 | 2/2 | ✅ Complete | 2026-02-09 |
| 7. Architecture, Deployment & Polish | v1.1 | 3/3 | ✅ Complete | 2026-02-10 |
