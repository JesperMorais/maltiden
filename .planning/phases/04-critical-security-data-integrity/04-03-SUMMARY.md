---
phase: 04-critical-security-data-integrity
plan: 03
subsystem: database
tags: [sqlite, sql-injection, foreign-keys, uuid, indexing]

# Dependency graph
requires:
  - phase: 04-01
    provides: JWT security fixes, auth bypass fixes
  - phase: 04-02
    provides: IDOR protection, transaction safety
provides:
  - Foreign key enforcement enabled via DSN and PRAGMA
  - SQL injection protection via json_each() for tag filtering
  - Safe parameterized queries in UpdateMemberStatus
  - UUID-based menu day IDs (no collision risk)
  - Composite index on menu_days(menu_id, date) for range queries
affects: [future database queries, menu operations, recipe filtering]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Foreign keys enforced via DSN parameter + PRAGMA verification"
    - "JSON array filtering via json_each() with parameterized queries"
    - "UUID-based IDs with 'md_' prefix for menu days"
    - "Explicit query variants instead of dynamic string building"

key-files:
  created:
    - backend/migrations/007_add_menu_date_index.sql
  modified:
    - backend/internal/storage/sqlite/db.go
    - backend/internal/storage/sqlite/recipe_storage.go
    - backend/internal/storage/sqlite/household_storage.go
    - backend/internal/storage/sqlite/menu_storage.go

key-decisions:
  - "Foreign keys enabled via DSN _foreign_keys=on for reliability across pooled connections"
  - "Tag filter uses json_each() subquery for exact matching without injection risk"
  - "UpdateMemberStatus uses explicit if/else queries instead of string concatenation"
  - "Menu day IDs use UUID with 'md_' prefix following codebase ID convention"
  - "Composite index on (menu_id, date) optimizes menu day range queries"

patterns-established:
  - "DSN parameters + PRAGMA for critical SQLite settings (defense in depth)"
  - "json_each() pattern for querying JSON arrays in SQLite"
  - "Explicit query variants for conditional updates (no string building)"

issues-created: []

# Metrics
duration: 1min 47s
completed: 2026-02-09
---

# Phase 04 Plan 03: Database Integrity Summary

**Foreign key enforcement enabled, SQL/JSON injection vulnerabilities fixed, UUID-based menu day IDs prevent collisions, composite index added for menu date queries**

## Performance

- **Duration:** 1min 47s
- **Started:** 2026-02-09T14:14:46Z
- **Completed:** 2026-02-09T14:16:33Z
- **Tasks:** 2
- **Files modified:** 5 (4 modified, 1 created)

## Accomplishments
- Foreign key enforcement enabled at database connection level (DSN + PRAGMA)
- SQL injection vulnerability in recipe tag filter eliminated via json_each()
- SQL string concatenation in UpdateMemberStatus replaced with explicit queries
- Menu day ID collision risk eliminated via UUID-based IDs
- Composite index added for efficient menu day range queries

## Task Commits

Each task was committed atomically:

1. **Task 1: Enable foreign keys and fix SQL safety issues** - `3434545` (fix)
2. **Task 2: Fix menu day IDs and add date index migration** - `94d5d20` (feat)

## Files Created/Modified
- `backend/internal/storage/sqlite/db.go` - Added `_foreign_keys=on` to DSN, PRAGMA verification, fmt import, registered migration 007
- `backend/internal/storage/sqlite/recipe_storage.go` - Replaced LIKE-based tag filter with json_each() subquery
- `backend/internal/storage/sqlite/household_storage.go` - Replaced dynamic query building with explicit if/else variants, removed strings import
- `backend/internal/storage/sqlite/menu_storage.go` - Changed menu day ID from concatenation to UUID-based generation, added uuid import
- `backend/migrations/007_add_menu_date_index.sql` - Created composite index on menu_days(menu_id, date)

## Decisions Made

**Foreign key enforcement strategy:**
- Used both DSN parameter (`?_foreign_keys=on`) and PRAGMA statement for defense in depth
- DSN parameter ensures enforcement even for pooled connections
- PRAGMA provides explicit verification step during Open()

**Tag filter security:**
- Replaced string concatenation (`%\""+filter.Tag+"\"%"`) with json_each() subquery
- Uses parameterized query with exact match on json_each.value
- Eliminates both SQL injection and JSON injection vectors

**UpdateMemberStatus refactoring:**
- Replaced dynamic string building (strings.Join) with explicit query variants
- Three paths: both fields, eating only, lunch only
- Helper function boolToInt() for consistent bool→int conversion
- Maintains single atomic UPDATE per call

**Menu day ID strategy:**
- Changed from `menu.ID+"_"+day.Date` to `"md_"+uuid.New().String()`
- Eliminates collision risk when multiple menus share same date
- Follows existing ID prefix convention (usr_, hh_, hm_, rec_, menu_)

**Index optimization:**
- Composite index on (menu_id, date) for efficient menu day range queries
- Covers common query pattern: fetch days for menu within date range

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None. All fixes implemented successfully following established codebase patterns.

## Next Phase Readiness

**Phase 4 completion:**
- All 11 critical/high findings from FIX-PLAN Batches 1-2 now addressed
- Batch 1 (AUTH-*): Fixed in 04-01
- Batch 2 (IDOR-*, TX-*, DATA-*): Fixed in 04-02 and 04-03
- Database integrity now enforced at multiple levels (FK, parameterized queries, UUIDs)

**Ready for Phase 5:**
- Phase 5 will address FIX-PLAN Batches 3-4 (validation, error handling)
- Database foundation is now secure and correct
- No blockers or concerns

---
*Phase: 04-critical-security-data-integrity*
*Completed: 2026-02-09*
