# Phase 01 Plan 02: Service & Storage Layer Review Summary

---
phase: 01-deep-code-review
plan: 02
subsystem: backend-core
tags: [security, data-integrity, services, storage, migrations]
completed: 2026-02-09
duration: 4min

requires:
  - 01-01 (auth system and handlers baseline)

provides:
  - Complete service layer security audit
  - Storage layer SQL injection audit
  - Migration schema validation
  - Transaction safety review
  - Data integrity analysis

affects:
  - Phase 02 (architecture changes will address transaction safety)
  - Phase 03 (fixes will need testing)

tech-stack:
  added: []
  patterns:
    - Transaction safety patterns identified
    - Input validation boundaries defined
    - Batch query optimization opportunities

key-files:
  reviewed:
    - backend/internal/services/*.go (6 files)
    - backend/internal/storage/sqlite/*.go (6 files)
    - backend/migrations/*.sql (6 files)
    - backend/cmd/server/main.go
  created:
    - .planning/phases/01-deep-code-review/01-02-FINDINGS.md

decisions:
  - date: 2026-02-09
    decision: Service layer review confirms transactional safety gaps
    rationale: Registration flow and join household need transaction wrappers
    impact: Phase 2 must address transaction patterns before production
  - date: 2026-02-09
    decision: Input validation belongs in service layer, not handlers
    rationale: Handlers should do basic checks, services enforce business rules
    impact: Validation fixes will be split between handlers and services
  - date: 2026-02-09
    decision: SQLite foreign keys MUST be enabled
    rationale: All CASCADE constraints are currently unenforced
    impact: Critical fix required immediately - affects data integrity
---

## One-Liner
Service and storage review found 28 issues including critical transaction safety gaps, unenforced foreign keys, and comprehensive input validation needs

## What Was Done

### Scope
Reviewed all service business logic, storage SQL queries, migration schemas, and server initialization for:
- Business logic bugs and data integrity issues
- SQL injection vulnerabilities
- Transaction safety and atomicity
- Foreign key constraints and cascade behavior
- Input validation completeness
- Performance patterns (N+1 queries, sequential fetching)

### Files Reviewed (19 total)

**Services (6 files):**
- `auth_service.go` - Registration, login, password handling
- `household_service.go` - Invite codes, member management, join flow
- `recipe_service.go` - Recipe CRUD operations
- `menu_service.go` - Menu generation algorithm
- `shopping_service.go` - Shopping list aggregation
- `tjek_service.go` - External API integration for offers

**Storage (6 files):**
- `db.go` - Database initialization and migrations
- `user_storage.go` - User CRUD
- `household_storage.go` - Household and member operations
- `recipe_storage.go` - Recipe queries with filtering
- `menu_storage.go` - Menu and menu_days persistence
- `shopping_storage.go` - Shopping item checked state

**Migrations (6 files):**
- `001_create_users.sql`
- `002_create_households.sql`
- `003_create_recipes.sql`
- `004_seed_recipes.sql`
- `005_create_menus.sql`
- `006_household_invites_and_member_status.sql`

**Entry Point:**
- `cmd/server/main.go` - Server initialization

### Methodology
1. Read every service method for business logic correctness
2. Trace data flow through service → storage → database
3. Verify all SQL queries use parameterization
4. Check transaction boundaries and atomicity requirements
5. Review migration schemas for constraints and indexes
6. Cross-reference with CONCERNS.md to verify known issues
7. Identify net-new findings not previously documented

## Findings Summary

**Total Issues:** 28 (2 Critical, 5 High, 15 Medium, 6 Low)

### Critical Issues (2)

1. **C1: Registration Flow Not Transactional** (`auth_service.go:46-83`)
   - Creates household, user, and member in separate operations
   - Partial failures leave orphaned records in database
   - No rollback mechanism for multi-step operation
   - **Impact:** Database pollution with incomplete registrations

2. **C2: Menu Generation With Zero Recipes Returns Success** (`menu_service.go:31-33`)
   - Returns `(nil, nil)` when no recipes available
   - Handler treats this as success, not error
   - Silent failure confuses users
   - **Impact:** Broken menu generation in fresh deployments

### High Severity Issues (5)

1. **H1: Recipe Tag Filter Vulnerable to LIKE Injection** (`recipe_storage.go:28-31`)
   - Tag filter uses `LIKE '%"tag"%'` pattern
   - Special characters in tag (`%`, `_`) bypass filter
   - Not full SQL injection (parameterized query protects) but pattern injection
   - **Impact:** Tag filter bypass, performance issues with wildcards

2. **H2: Foreign Key Constraints Not Enforced** (`db.go` - missing PRAGMA)
   - SQLite requires `PRAGMA foreign_keys = ON` to enforce FKs
   - All CASCADE and SET NULL behaviors are currently ignored
   - Orphaned records accumulate (menu_days, shopping_items, members)
   - **Impact:** Complete data integrity failure

3. **H3: SQL Query String Concatenation** (`household_storage.go:205`)
   - UpdateMemberStatus builds query via string concatenation
   - Currently safe (field names hardcoded) but fragile pattern
   - Sets bad precedent for future code
   - **Impact:** Maintenance risk, potential SQL injection if copied

4. **H4: Menu Day Primary Key Collision Risk** (`menu_storage.go:47`)
   - PK generated as `menuID + "_" + date`
   - Collision possible if menu IDs are reused
   - Fragile ID generation pattern
   - **Impact:** Menu regeneration failures, data corruption

5. **H5: Password Validation Only Checks Minimum** (`auth_service.go:27-29`)
   - No maximum length check
   - bcrypt silently truncates at 72 bytes
   - Leads to confusing login behavior with long passwords
   - **Impact:** Security confusion, UX issues

### Medium Severity Issues (15)

- **M1:** Login timing attack allows email enumeration
- **M2:** No validation on recipe name length (DOS via huge names)
- **M3:** No upper bound on recipe servings (calculation overflow)
- **M4:** No limit on ingredients array size (performance degradation)
- **M5:** Shopping service N+1 query pattern (verified from CONCERNS.md)
- **M6:** Shopping item ID uses MD5 (not security issue, but code smell)
- **M7:** Menu generation days count unbounded
- **M8:** Menu generation servings count unbounded
- **M9:** Random recipe selection can duplicate (poor UX)
- **M10:** Recipe deletion cascade behavior undefined
- **M11:** Invite code entropy sufficient but undocumented
- **M12:** JoinHousehold transaction doesn't lock invite code (race condition)
- **M13:** Menu day date format not validated
- **M14:** Bubble sort used instead of sort.Slice (verified from CONCERNS.md)
- **M15:** Tjek service sequential catalog fetching (verified from CONCERNS.md)

### Low Severity Issues (6)

- **L1:** Migration file paths relative to working directory (verified from CONCERNS.md)
- **L2:** WAL mode not enabled for SQLite concurrency (verified from CONCERNS.md)
- **L3:** No graceful shutdown handling
- **L4:** Database connection pool not configured
- **L5:** Shopping storage schema has redundant column
- **L6:** API consistency in nil vs empty array returns (actually consistent)

## Verification Against CONCERNS.md

**Known Issues Verified:**
- ✓ Registration not transactional (C1) - CONCERNS.md line 144-149
- ✓ N+1 query pattern (M5) - CONCERNS.md line 117-121
- ✓ Bubble sort (M14) - CONCERNS.md line 25-29
- ✓ Sequential fetching (M15) - CONCERNS.md line 123-127
- ✓ Migration paths (L1) - CONCERNS.md line 137-142
- ✓ WAL mode (L2) - CONCERNS.md line 161-163

**Net-New Findings (22):**
All critical and high issues except C1 are newly discovered. Most medium validation issues are new.

## Connections to 01-01 Findings

**Handler findings that trace to service layer:**
- **01-01 H6** (Login timing leak) → **M1** (confirmed in service implementation)
- **01-01 M3** (Path parameter validation) → **H1** (tag filter injection at storage)
- **01-01 M4** (GenerateMenuRequest validation) → **M7, M8** (service doesn't validate)

**Service findings that affect handlers:**
- **C2** (zero recipes) → Handlers need to check for nil response
- **H5** (password length) → Handlers should validate before service call

## Deviations from Plan

None - plan executed as written.

## Blockers Encountered

None - all files accessible and reviewable.

## What's Next

### Immediate Actions (before any further development):
1. **Add `PRAGMA foreign_keys = ON`** in `db.go` immediately after opening database
2. **Wrap registration in transaction** (C1 fix)
3. **Return error on zero recipes** (C2 fix)

### Phase 1 Completion Status
This was the **last plan in Phase 1**. Combined with 01-01 findings:
- **Total files reviewed:** 31 (12 from 01-01 + 19 from 01-02)
- **Total issues found:** 58 (30 from 01-01 + 28 from 01-02)
- **Coverage:** Complete backend codebase review

**Phase 1 is now complete.**

### Phase 2 Preparation
Phase 2 (Architecture & Performance Review) should address:
- Transaction pattern architecture (solve C1 systematically)
- Input validation framework (solve all M-level validation gaps)
- Batch query patterns (solve M5, M15)
- Error handling standardization (improve sentinel errors)

### Immediate Fix Priority (before Phase 2)
**Critical (block all other work):**
1. H2: Enable foreign keys
2. C1: Transactional registration
3. C2: Zero recipes error

**High (before feature development):**
4. H1: Fix tag filter
5. H4: UUID menu day IDs
6. H5: Password max length

## Key Insights

### Transaction Safety is Systemic
Registration (C1) and join household (M12) both have transaction issues. This indicates a need for transaction patterns/helpers in Phase 2.

### Input Validation is Incomplete Throughout
Medium-severity findings M2-M4, M7-M8, M13 all show gaps in input validation. Need a systematic validation approach (Phase 2).

### SQLite Configuration is Critical
Two critical infrastructure issues (H2 foreign keys, L2 WAL mode) show SQLite setup needs immediate attention. These are one-line fixes with huge impact.

### Service Layer Quality is Generally Good
Despite 28 issues, most are validation gaps and edge cases. The core business logic is sound:
- Invite code generation is cryptographically secure
- Join household uses transactions correctly (except invite lock)
- Shopping list aggregation algorithm is correct (just slow)
- Menu generation randomness is appropriate

### Performance Issues are Documented and Real
CONCERNS.md correctly identified N+1 queries and sequential fetching. These are real bottlenecks that should be fixed in Phase 2.

## Metrics

- **Files Reviewed:** 19
- **Lines of Code Reviewed:** ~3,000
- **Issues Found:** 28
- **Critical Issues:** 2 (7%)
- **High Issues:** 5 (18%)
- **Medium Issues:** 15 (54%)
- **Low Issues:** 6 (21%)
- **Cross-referenced with CONCERNS.md:** 6 issues verified
- **Net-new findings:** 22 issues
- **SQL Injection Audit Result:** All queries parameterized ✓ (one LIKE injection edge case)

## Phase 1 Complete

All backend files have been reviewed across 01-01 and 01-02:
- ✅ Authentication system (01-01)
- ✅ HTTP handlers (01-01)
- ✅ Middleware (01-01)
- ✅ Service business logic (01-02)
- ✅ Storage data access (01-02)
- ✅ Migrations schema (01-02)
- ✅ Server initialization (01-02)

**Total Phase 1 findings:** 58 issues across 31 files
**Ready for Phase 2:** Architecture & Performance Review
