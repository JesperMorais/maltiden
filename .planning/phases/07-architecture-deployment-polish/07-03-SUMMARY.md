---
phase: 07-architecture-deployment-polish
plan: 03
subsystem: architecture
tags: [jwt, fnv, hash, dependency-injection, go]

# Dependency graph
requires:
  - phase: 07-01
    provides: Domain interfaces for services, wireDependencies pattern
  - phase: 07-02
    provides: Structured logging, graceful shutdown, embedded migrations
provides:
  - Injectable JWTService struct eliminating package-level mutable state
  - FNV-based shopping item ID generation (replacing MD5)
  - Recipe shuffle algorithm for menu variety
affects: [testing, service-initialization]

# Tech tracking
tech-stack:
  added: [hash/fnv, math/rand/v2]
  patterns: [injectable-jwt-service, token-validator-interface, recipe-shuffle-algorithm]

key-files:
  created: []
  modified:
    - backend/pkg/utils/jwt.go
    - backend/cmd/server/main.go
    - backend/internal/api/router.go
    - backend/pkg/middleware/auth.go
    - backend/internal/services/auth_service.go
    - backend/internal/services/shopping_service.go
    - backend/internal/services/menu_service.go
    - backend/internal/services/household_service_test.go

key-decisions:
  - "JWTService as injectable struct with 32-char validation in constructor"
  - "TokenValidator interface in middleware to avoid circular dependencies"
  - "FNV-64a hash replaces MD5 for shopping item IDs (non-cryptographic use case)"
  - "Recipe shuffle with modulo cycling for menus longer than recipe pool"
  - "math/rand/v2 for modern Go random number generation"

patterns-established:
  - "Service injection pattern: jwtService passed through main → router → wireDependencies → services"
  - "Middleware factory pattern: RequireAuth(validator) returns middleware function"
  - "Test helpers for shared dependencies: setupTestJWTService(t)"

issues-created: []

# Metrics
duration: 5min
completed: 2026-02-10
---

# Phase 7 Plan 3: Architecture Polish & v1.1 Completion

**Eliminated package-level JWT state with injectable service, replaced MD5 with FNV hash, and implemented recipe shuffle for menu variety**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-10T09:17:53Z
- **Completed:** 2026-02-10T09:22:43Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Refactored JWT handling from package-level mutable state to injectable `JWTService` struct
- Replaced cryptographic MD5 with non-cryptographic FNV-64a hash for shopping item IDs
- Implemented Fisher-Yates shuffle for recipe selection to prevent duplicates in menus
- Completed Phase 7 (Architecture, Deployment & Polish)
- **Completed v1.1 Fix milestone** - all prioritized code review findings addressed

## Task Commits

Each task was committed atomically:

1. **Task 1: Refactor JWT to injectable JWTService struct** - `1cc1bd9` (refactor)
2. **Task 2: Replace MD5 with FNV and add recipe shuffle for variety** - `d87dcce` (perf)

## Files Created/Modified

- `backend/pkg/utils/jwt.go` - JWTService struct with NewJWTService constructor, removed package-level var
- `backend/cmd/server/main.go` - Creates and injects JWTService into router
- `backend/internal/api/router.go` - Accepts jwtService, passes to RequireAuth and AuthService
- `backend/pkg/middleware/auth.go` - TokenValidator interface, RequireAuth factory function
- `backend/internal/services/auth_service.go` - Accepts jwtService, uses for token generation
- `backend/internal/services/shopping_service.go` - Uses hash/fnv instead of crypto/md5
- `backend/internal/services/menu_service.go` - Shuffles recipes, cycles with modulo for variety
- `backend/internal/services/household_service_test.go` - setupTestJWTService helper

## Decisions Made

**JWTService structure:**
- Struct with private `secret []byte` field instead of package-level variable
- Constructor validates 32-character minimum at application startup
- Methods use receiver's secret field instead of package-level state
- Rationale: Eliminates global mutable state, enables testing with different secrets, follows Go best practices

**Middleware injection:**
- `TokenValidator` interface defined in middleware package to avoid circular dependency
- `RequireAuth` returns a middleware factory function instead of being a direct middleware
- Rationale: Allows dependency injection while keeping clean package boundaries

**Hashing algorithm change:**
- FNV-64a replaces MD5 for shopping item ID generation
- Not a migration concern: shopping lists regenerate per menu, old IDs don't persist
- Rationale: Non-cryptographic hash is faster and appropriate for deduplication use case

**Recipe variety:**
- Fisher-Yates shuffle with modulo cycling when menu days exceed recipe pool
- Uses `math/rand/v2` instead of deprecated `math/rand`
- Rationale: Prevents duplicate recipes until all recipes used once, improves user experience

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

**Type mismatch in shuffle:**
- Issue: `GetAll()` returns `[]RecipeSummary`, not `[]Recipe`
- Resolution: Changed shuffle slice type to `[]domain.RecipeSummary`
- Impact: None - RecipeSummary contains ID field needed for assignment

## Batch 9 Findings Addressed

This plan completed the final findings from FIX-PLAN.md Batch 9:

**Addressed:**
- AUTH-11: JWT package-level state → Refactored to injectable JWTService struct
- PERF-16: MD5 hashing → Replaced with FNV-64a for shopping item IDs
- DATA-10: Recipe duplicates → Implemented shuffle-based selection with cycling

**Already completed in prior plans:**
- PERF-10: sort.Slice → Completed in Phase 6 (already using sort.Slice correctly)

**Deferred (minor improvements, not critical):**
- ARCH-10: pkg/ restructure → Unnecessary churn for current project size
- ARCH-11/12/14/15: Various architectural cleanups → Out of v1.1 scope
- DEPLOY-11/12/13: Additional deployment configs → Fly.io sufficient for now
- ERROR-04/05/06: Error wrapping enhancements → Phase 5 sentinel errors sufficient
- AUTH-13/15: Minor JWT verification enhancements → Current validation secure
- DATA-07/08/11: Edge case validations → Core cases covered, defer to future iteration
- PERF-13/14: Tjek response limits → Phase 6 caching already addresses performance

## Next Phase Readiness

**Phase 7 Complete:**
- All architecture improvements from Batch 7-9 implemented
- Deployment configuration production-ready (embedded migrations, graceful shutdown, health checks)
- Code quality improvements complete (no package-level state, modern hashing, better algorithms)

**v1.1 Milestone Complete:**
- All critical, high, and medium-severity findings addressed
- 93 backend findings triaged and implemented across 15 plans
- Project ready for production deployment
- Future work documented in FIX-PLAN.md "Deferred" sections

**No blockers or concerns.**

---

*Phase: 07-architecture-deployment-polish*
*Completed: 2026-02-10*
*v1.1 Fix Milestone: ✓ Complete*
