---
phase: 04-critical-security-data-integrity
plan: 02
subsystem: security
tags: [idor, transactions, authorization, sqlite, go]

# Dependency graph
requires:
  - phase: 04-01
    provides: JWT authentication hardening and protected endpoints
provides:
  - IDOR protection for menu/shopping operations via household ownership verification
  - Atomic registration flow preventing orphaned records
  - Menu ownership verification method in MenuStorage
  - Transaction-aware storage methods (CreateTx)
affects: [05-error-handling-validation, 06-concurrency-edge-cases, 07-architecture-performance]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - IDOR protection pattern - verify resource ownership before data access
    - Transaction pattern for multi-table atomic operations
    - Tx method pattern for transaction-aware storage operations

key-files:
  created: []
  modified:
    - backend/internal/api/handlers/shopping.go
    - backend/internal/storage/sqlite/menu_storage.go
    - backend/internal/services/auth_service.go
    - backend/internal/storage/sqlite/user_storage.go
    - backend/internal/storage/sqlite/household_storage.go
    - backend/internal/api/router.go

key-decisions:
  - "Shopping handlers verify menu belongs to user's household before data access (403 if mismatch)"
  - "Registration flow wrapped in database transaction to prevent partial failures"
  - "MenuStorage injected into ShoppingHandler for IDOR verification"

patterns-established:
  - "IDOR protection: GetHouseholdIDByMenuID + compare with auth context householdID"
  - "CreateTx methods follow existing Tx pattern (tx *sql.Tx parameter, tx.Exec)"
  - "Transaction pattern: Begin, defer Rollback, operations, Commit"

issues-created: []

# Metrics
duration: 2min
completed: 2026-02-09
---

# Phase 4 Plan 2: IDOR Protection & Transaction Safety Summary

**Menu/shopping IDOR protection with household verification, atomic registration via transactions, and zero-recipe guard verification**

## Performance

- **Duration:** 2 min 7 sec
- **Started:** 2026-02-09T14:09:35Z
- **Completed:** 2026-02-09T14:11:42Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Added IDOR protection to shopping endpoints - users can only access menus belonging to their household
- Wrapped registration in atomic transaction preventing orphaned households or users on failure
- Verified zero-recipe menu generation guard already present and correct
- Established IDOR protection pattern for resource ownership verification

## Task Commits

Each task was committed atomically:

1. **Task 1: Add IDOR protection to menu/shopping operations** - `e1214cd` (fix)
2. **Task 2: Wrap registration in transaction and guard zero-recipe menu generation** - `4545997` (fix)

## Files Created/Modified
- `backend/internal/storage/sqlite/menu_storage.go` - Added GetHouseholdIDByMenuID for IDOR verification
- `backend/internal/api/handlers/shopping.go` - Added household verification in GetShoppingList and UpdateItem (403 if mismatch)
- `backend/internal/api/router.go` - Pass menuStorage to ShoppingHandler, pass db to AuthService
- `backend/internal/services/auth_service.go` - Added db field, wrapped Register in Begin/Commit/Rollback
- `backend/internal/storage/sqlite/household_storage.go` - Added CreateTx method
- `backend/internal/storage/sqlite/user_storage.go` - Added CreateTx method

## Decisions Made

**1. Shopping IDOR protection approach**
- MenuStorage.GetHouseholdIDByMenuID returns household_id for menu, returns empty string if not found
- Shopping handlers compare menu's household_id with auth context household_id
- 404 for menu not found, 403 for household mismatch (clear distinction)

**2. Registration transaction scope**
- Transaction wraps household creation, user creation, and member addition
- JWT generation happens AFTER transaction commit (doesn't need rollback)
- Follows established Begin/defer Rollback/Commit pattern from household_service.go

**3. Tx method pattern**
- CreateTx methods follow existing pattern: tx *sql.Tx first parameter, tx.Exec instead of s.db.Exec
- Consistent with AddMemberTx, RemoveMemberTx, etc. already in codebase

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

**Ready for Phase 5 (Error Handling & Validation):**
- IDOR protection establishes pattern for resource ownership checks
- Transaction pattern ready for additional multi-table operations
- AUTH-04 (IDOR) resolved - shopping endpoints verify menu ownership
- DATA-01 (registration not transactional) resolved - atomic registration prevents orphaned records
- DATA-02 (zero recipes) verified - already correctly returns 400 error

**No blockers or concerns.**

---
*Phase: 04-critical-security-data-integrity*
*Completed: 2026-02-09*
