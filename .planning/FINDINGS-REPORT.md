# Maltiden Backend: Comprehensive Findings Report

**Date:** 2026-02-09
**Scope:** Complete backend codebase analysis from Phases 1 & 2
**Source Documents:** 01-01-FINDINGS.md, 01-02-FINDINGS.md, 02-01-FINDINGS.md, 02-02-FINDINGS.md

---

## Executive Summary

### Overall Status

This report consolidates **93 unique findings** discovered across two comprehensive review phases:
- **Phase 1 (Deep Code Review):** 58 findings across HTTP handlers, auth system, services, and storage layers
- **Phase 2 (Architecture & Performance Review):** 38 findings covering systemic patterns, database configuration, and external API integration
- **Deduplication:** 3 findings were duplicates or marked as non-issues, resulting in 93 actionable findings

**Total Findings by Severity:**
- **Critical:** 6 findings (6.5%) - Require immediate action before any production deployment
- **High:** 26 findings (28.0%) - Must be addressed before production launch
- **Medium:** 51 findings (54.8%) - Should be resolved within current sprint
- **Low:** 10 findings (10.8%) - Polish and optimization opportunities

### Top 5 Most Critical Issues

1. **AUTH-01 (Critical):** Empty JWT secret allows complete authentication bypass and token forgery
2. **AUTH-02 (Critical):** Shopping list endpoints completely lack authentication - full IDOR vulnerability
3. **DATA-01 (Critical):** Registration flow not transactional - creates orphaned database records on failure
4. **DATA-02 (Critical):** Menu generation with zero recipes silently returns success with null response
5. **AUTH-03 (Critical):** JWT signing algorithm not pinned - vulnerable to "none" algorithm attacks

### Overall Backend Health Assessment

**Security Posture:** 🔴 Critical vulnerabilities present - not production-ready
- 6 critical security bypasses and 16 high-severity security gaps must be fixed immediately
- Current state allows unauthorized access to all shopping list data and arbitrary token forging

**Data Integrity:** 🟡 Significant risks identified
- Foreign key enforcement disabled in SQLite - all CASCADE constraints ignored
- Non-transactional operations create orphaned records
- No validation framework leads to data corruption potential

**Performance:** 🟡 Multiple bottlenecks present
- External API makes 20+ sequential calls (6s latency) with no caching (600x speedup possible)
- N+1 query patterns and missing database indexes identified
- SQLite not configured for production (WAL mode disabled)

**Architecture:** 🟡 Technical debt accumulating
- No abstraction layers block testing and future database migration
- Fragile error handling via string matching in 15+ locations
- Package-level state prevents proper dependency injection

**Recommendation:** Address all 6 critical and 26 high-severity findings before any production deployment. Estimated effort: 60-85 hours total remediation.

---

## Finding Statistics

### By Area and Severity

| Area | Critical | High | Medium | Low | Total |
|------|----------|------|--------|-----|-------|
| Authentication & Authorization | 4 | 8 | 3 | 0 | **15** |
| Data Integrity & Transactions | 2 | 5 | 4 | 0 | **11** |
| Input Validation | 0 | 2 | 13 | 1 | **16** |
| Error Handling | 0 | 1 | 3 | 2 | **6** |
| Architecture & Code Quality | 0 | 3 | 10 | 2 | **15** |
| Performance | 0 | 6 | 10 | 1 | **17** |
| Deployment & Observability | 0 | 1 | 8 | 4 | **13** |
| **TOTAL** | **6** | **26** | **51** | **10** | **93** |

---

## Top 10 Most Critical Findings

1. **AUTH-01** (Critical) - Empty JWT secret allows token forgery → `backend/pkg/utils/jwt.go:11`
2. **AUTH-02** (Critical) - Shopping list endpoints lack authentication → `backend/internal/api/router.go:76-77`
3. **DATA-01** (Critical) - Registration flow not transactional → `backend/internal/services/auth_service.go:46-83`
4. **DATA-02** (Critical) - Menu generation with zero recipes returns success → `backend/internal/services/menu_service.go:31-33`
5. **AUTH-03** (Critical) - JWT algorithm not pinned → `backend/pkg/utils/jwt.go:34-37`
6. **AUTH-04** (Critical) - No IDOR protection in menu/shopping operations → `backend/internal/api/handlers/menus.go:33-36`
7. **DATA-03** (High) - Recipe tag filter vulnerable to JSON injection → `backend/internal/storage/sqlite/recipe_storage.go:28-31`
8. **DATA-04** (High) - Foreign key enforcement disabled in SQLite → `backend/internal/storage/sqlite/db.go`
9. **VALID-01** (High) - Error message injection in HTTP responses → Multiple handler files
10. **VALID-02** (High) - No request body size limits → All JSON handlers

---

## Cross-Reference Index

Maps unified finding IDs to original phase/plan finding references:

| Unified ID | Original Finding(s) | Area |
|------------|---------------------|------|
| AUTH-01 | 01-01-C1 | Authentication |
| AUTH-02 | 01-01-C2 | Authentication |
| AUTH-03 | 01-01-C3 | Authentication |
| AUTH-04 | 01-01-C4 | Authentication |
| AUTH-05 | 01-01-H3, 02-01-M17 | Authentication |
| AUTH-06 | 01-01-H4, 02-01-L6 | Authentication |
| AUTH-07 | 01-01-H5 | Authentication |
| AUTH-08 | 01-01-H6, 01-02-M1 | Authentication |
| AUTH-09 | 01-01-H7 | Authentication |
| AUTH-10 | 01-02-H5 | Authentication |
| AUTH-11 | 02-01-H2 | Authentication |
| AUTH-12 | 02-01-H5 | Authentication |
| AUTH-13 | 01-01-M11 | Authentication |
| AUTH-14 | 01-01-M12 | Authentication |
| AUTH-15 | 01-01-M13 | Authentication |
| DATA-01 | 01-02-C1 | Data Integrity |
| DATA-02 | 01-02-C2 | Data Integrity |
| DATA-03 | 01-02-H1 | Data Integrity |
| DATA-04 | 01-02-H2 | Data Integrity |
| DATA-05 | 01-02-H3 | Data Integrity |
| DATA-06 | 01-02-H4 | Data Integrity |
| DATA-07 | 01-02-M10 | Data Integrity |
| DATA-08 | 01-02-M12 | Data Integrity |
| DATA-09 | 02-02-H2 | Data Integrity |
| DATA-10 | 01-02-M9 | Data Integrity |
| DATA-11 | 01-02-M13 | Data Integrity |
| VALID-01 | 01-01-H1 | Input Validation |
| VALID-02 | 01-01-H2, 02-02-L2 | Input Validation |
| VALID-03 | 01-01-M2 | Input Validation |
| VALID-04 | 01-01-M3 | Input Validation |
| VALID-05 | 01-01-M4 | Input Validation |
| VALID-06 | 01-01-M5 | Input Validation |
| VALID-07 | 01-01-M7 | Input Validation |
| VALID-08 | 01-01-M8 | Input Validation |
| VALID-09 | 01-01-M10 | Input Validation |
| VALID-10 | 01-02-M2 | Input Validation |
| VALID-11 | 01-02-M3 | Input Validation |
| VALID-12 | 01-02-M4 | Input Validation |
| VALID-13 | 01-02-M7 | Input Validation |
| VALID-14 | 01-02-M8 | Input Validation |
| VALID-15 | 01-01-L4 | Input Validation |
| VALID-16 | 01-01-M9, 01-02-L6 | Input Validation |
| ERROR-01 | 01-01-M1, 02-01-H4 | Error Handling |
| ERROR-02 | 02-01-M7 | Error Handling |
| ERROR-03 | 02-01-M8 | Error Handling |
| ERROR-04 | 02-01-M9 | Error Handling |
| ERROR-05 | 01-01-L5 | Error Handling |
| ERROR-06 | 02-01-L2 | Error Handling |
| ARCH-01 | 02-01-H1 | Architecture |
| ARCH-02 | 01-01-H8 | Architecture |
| ARCH-03 | 02-01-H3 | Architecture |
| ARCH-04 | 02-01-M1 | Architecture |
| ARCH-05 | 02-01-M2 | Architecture |
| ARCH-06 | 02-01-M3 | Architecture |
| ARCH-07 | 02-01-M4 | Architecture |
| ARCH-08 | 02-01-M5 | Architecture |
| ARCH-09 | 02-01-M6 | Architecture |
| ARCH-10 | 02-01-M10 | Architecture |
| ARCH-11 | 02-01-M11 | Architecture |
| ARCH-12 | 02-01-M12 | Architecture |
| ARCH-13 | 01-01-M6 | Architecture |
| ARCH-14 | 02-01-L1 | Architecture |
| ARCH-15 | 02-01-L3 | Architecture |
| PERF-01 | 01-02-M5, 02-02-H1 | Performance |
| PERF-02 | 02-02-H3 | Performance |
| PERF-03 | 02-02-H4 | Performance |
| PERF-04 | 02-02-H5 | Performance |
| PERF-05 | 01-02-L4, 02-02-H6 | Performance |
| PERF-06 | 02-02-M1 | Performance |
| PERF-07 | 02-02-M2 | Performance |
| PERF-08 | 02-02-M3 | Performance |
| PERF-09 | 02-02-M4 | Performance |
| PERF-10 | 01-02-M14, 02-02-M5 | Performance |
| PERF-11 | 02-02-M6 | Performance |
| PERF-12 | 02-02-M7 | Performance |
| PERF-13 | 02-02-M8 | Performance |
| PERF-14 | 02-02-M9 | Performance |
| PERF-15 | 01-02-M15, 02-02-H3 | Performance |
| PERF-16 | 01-02-M6 | Performance |
| PERF-17 | 02-02-L1 | Performance |
| DEPLOY-01 | 02-02-H7 | Deployment |
| DEPLOY-02 | 01-02-L1 | Deployment |
| DEPLOY-03 | 02-01-M13 | Deployment |
| DEPLOY-04 | 02-01-M14 | Deployment |
| DEPLOY-05 | 01-02-L3, 02-01-L4 | Deployment |
| DEPLOY-06 | 01-01-L2, 02-01-L5 | Deployment |
| DEPLOY-07 | 02-02-M11 | Deployment |
| DEPLOY-08 | 02-01-M15 | Deployment |
| DEPLOY-09 | 02-01-M16 | Deployment |
| DEPLOY-10 | 01-01-L1 | Deployment |
| DEPLOY-11 | 01-01-L3 | Deployment |
| DEPLOY-12 | 02-02-M10 | Deployment |
| DEPLOY-13 | 02-01-L7 | Deployment |

---

## 1. Authentication & Authorization

### AUTH-01: Empty JWT Secret Allows Token Forgery
**Severity:** Critical
**Source:** 01-01-C1
**File:** `backend/pkg/utils/jwt.go:11`

JWT secret read from environment variable with no validation. If `JWT_SECRET` is empty, all tokens signed with empty key, allowing complete authentication bypass. Any attacker can forge valid tokens with arbitrary user/household IDs and gain full system access.

**Recommendation:** Add startup validation requiring JWT_SECRET to be set with minimum 32 characters. Fatal error if missing or too short.

---

### AUTH-02: Shopping List Endpoints Lack Authentication
**Severity:** Critical
**Source:** 01-01-C2
**File:** `backend/internal/api/router.go:76-77`

Shopping list GET and PATCH endpoints use `HandleFunc` instead of `Handle` with `RequireAuth` middleware. Any unauthenticated user can view and modify any household's shopping lists by guessing menu IDs. While UUIDs are hard to guess, this is security by obscurity.

**Recommendation:** Wrap both endpoints with `middleware.RequireAuth` and add householdID verification in handlers.

---

### AUTH-03: JWT Algorithm Not Pinned
**Severity:** Critical
**Source:** 01-01-C3
**File:** `backend/pkg/utils/jwt.go:34-37`

JWT validation doesn't verify signing algorithm. Vulnerable to algorithm confusion attacks. Code should explicitly validate that token uses expected HS256 algorithm before accepting it.

**Recommendation:** Add algorithm validation in token parsing callback to reject unexpected algorithms.

---

### AUTH-04: No IDOR Protection in Menu/Shopping Operations
**Severity:** Critical
**Source:** 01-01-C4
**Files:** `backend/internal/api/handlers/menus.go:33-36`, `backend/internal/api/handlers/shopping.go:26-29`

Handlers accept menuId parameters but never verify that menus belong to authenticated user's household. After adding authentication (AUTH-02 fix), IDOR vulnerability persists without ownership validation.

**Recommendation:** Add storage-layer method `GetShoppingListForHousehold(menuID, householdID)` that returns 403 if menu doesn't belong to household.

---

### AUTH-05: No Rate Limiting on Authentication Endpoints
**Severity:** High
**Source:** 01-01-H3, 02-01-M17
**Files:** `backend/internal/api/router.go:40-41`

Auth endpoints have no rate limiting. Attackers can brute force passwords, enumerate emails via registration failures, and perform credential stuffing. While bcrypt cost 12 provides natural rate limiting (~250ms per attempt), distributed attacks bypass this.

**Recommendation:** Add per-IP rate limiting middleware (5 requests per minute for auth endpoints).

---

### AUTH-06: CORS Restricted to Localhost Only
**Severity:** High
**Source:** 01-01-H4, 02-01-L6
**File:** `backend/pkg/middleware/cors.go:11`

Hardcoded localhost origins. Production frontend will be blocked by CORS, making API unusable. Deployment blocker.

**Recommendation:** Make allowed origins configurable via `CORS_ORIGINS` environment variable.

---

### AUTH-07: No Email Format Validation on Registration
**Severity:** High
**Source:** 01-01-H5
**Files:** `backend/internal/api/handlers/auth.go:18-24`

Registration accepts any string as email with no format validation. Creates accounts with invalid emails that break password reset flows and allow case-sensitive duplicates.

**Recommendation:** Normalize email (lowercase, trim), validate basic format with regex.

---

### AUTH-08: Login Endpoint Timing Attack Possible
**Severity:** High
**Source:** 01-01-H6, 01-02-M1
**Files:** `backend/internal/api/handlers/auth.go:49-52`, `backend/internal/services/auth_service.go:99-110`

Handler returns generic error correctly, but service layer has timing difference: fast return for non-existent user (~50ms) vs slow bcrypt check for wrong password (~250ms). Timing side-channel allows email enumeration.

**Recommendation:** Always run bcrypt comparison even for non-existent users using dummy hash.

---

### AUTH-09: Context Key Type Collision Risk
**Severity:** High
**Source:** 01-01-H7
**File:** `backend/pkg/middleware/auth.go:10-13`

Context keys use simple string values ("user_id", "household_id"). If other middleware or libraries use same strings, collision overwrites auth context, potentially causing privilege escalation.

**Recommendation:** Use unexported struct type for keys instead of strings (idiomatic Go pattern).

---

### AUTH-10: Password Validation Only Checks Minimum Length
**Severity:** High
**Source:** 01-02-H5
**File:** `backend/internal/services/auth_service.go:27-29`

No maximum password length check. bcrypt silently truncates passwords longer than 72 bytes, creating confusing behavior where users can login with truncated passwords they don't remember setting.

**Recommendation:** Enforce maximum 72 characters to match bcrypt limit.

---

### AUTH-11: Package-Level JWT State
**Severity:** High
**Source:** 02-01-H2
**File:** `backend/pkg/utils/jwt.go:11`

JWT secret stored as package-level variable read at init time. Prevents testing with different secrets, blocks configuration validation, and creates security risk (see AUTH-01).

**Recommendation:** Create JWTService struct accepting secret as dependency, validate in main.go before starting server.

---

### AUTH-12: No Validation of Required Environment Variables
**Severity:** High
**Source:** 02-01-H5
**File:** `backend/cmd/server/main.go:14-22`

Server reads environment variables but doesn't validate them at startup. Empty JWT_SECRET (AUTH-01) isn't caught until runtime. Should fail fast with clear error message.

**Recommendation:** Add validation block at start of main() checking all required env vars.

---

### AUTH-13: CreateInvite Checks Role but Not Member Status
**Severity:** Medium
**Source:** 01-01-M11
**File:** `backend/internal/api/handlers/household.go:49-54`

Checks member role but not if member has been removed. Removed member's JWT (valid for 7 days) can still create invites until token expires.

**Recommendation:** Check member status in database, return error if member record doesn't exist or is inactive.

---

### AUTH-14: Shopping List menuId Has No Household Verification
**Severity:** Medium
**Source:** 01-01-M12
**File:** `backend/internal/api/handlers/shopping.go:20-24`

After adding authentication (AUTH-02 fix), handler still needs householdID verification. No check that menuID belongs to authenticated user's household.

**Recommendation:** Service layer should verify menu belongs to household before returning data.

---

### AUTH-15: UpdateItem Accepts Any menuId for Path itemId
**Severity:** Medium
**Source:** 01-01-M13
**File:** `backend/internal/api/handlers/shopping.go:51-55`

No validation that itemID belongs to specified menuID. Could pass itemID from one menu and menuID from another, causing data inconsistency.

**Recommendation:** Service layer should validate itemID belongs to menuID, return 404 if mismatch.

---

## 2. Data Integrity & Transactions

### DATA-01: Registration Flow Not Transactional
**Severity:** Critical
**Source:** 01-02-C1
**File:** `backend/internal/services/auth_service.go:46-83`

Registration creates household, user, and member records separately without transaction. If any step fails after first succeeds, orphaned records remain in database permanently with no cleanup mechanism.

**Recommendation:** Wrap entire registration in database transaction with rollback on any failure.

---

### DATA-02: Menu Generation With Zero Recipes Returns Success
**Severity:** Critical
**Source:** 01-02-C2
**File:** `backend/internal/services/menu_service.go:31-33`

When recipes table is empty, returns `(nil, nil)` which handlers treat as success. Frontend crashes trying to access nil menu. Fresh deployments before seed data have broken menu generation.

**Recommendation:** Return error "no_recipes_available" when recipes table is empty.

---

### DATA-03: Recipe Tag Filter Vulnerable to JSON Injection
**Severity:** High
**Source:** 01-02-H1
**File:** `backend/internal/storage/sqlite/recipe_storage.go:28-31`

Tag filter uses `LIKE '%"tag"%'` pattern. Tag containing `"` can match unintended tags or break pattern. More critically, tag `%` matches ALL recipes (wildcard injection).

**Recommendation:** Use SQLite JSON1 extension with `json_each()` for proper JSON querying, or escape LIKE special characters.

---

### DATA-04: No Foreign Key Enforcement in SQLite
**Severity:** High
**Source:** 01-02-H2
**File:** `backend/internal/storage/sqlite/db.go`

SQLite requires `PRAGMA foreign_keys = ON` to enforce constraints. Without it, all `FOREIGN KEY` and `ON DELETE CASCADE` declarations are ignored. Leads to orphaned menu_days, shopping_items, and household_members.

**Recommendation:** Execute `PRAGMA foreign_keys = ON` immediately after opening database.

---

### DATA-05: UpdateMemberStatus Builds SQL via String Concatenation
**Severity:** High
**Source:** 01-02-H3
**File:** `backend/internal/storage/sqlite/household_storage.go:205`

SQL query built via string concatenation. Currently safe (field names hardcoded), but pattern is dangerous and could lead to SQL injection if copied with user-controlled inputs.

**Recommendation:** Use explicit query variants for each update combination instead of string building.

---

### DATA-06: Menu Day Primary Key Collision Risk
**Severity:** High
**Source:** 01-02-H4
**File:** `backend/internal/storage/sqlite/menu_storage.go:47`

Menu day ID generated as `menuID + "_" + date`. If menu generation logic changes to reuse menu IDs, collision causes "UNIQUE constraint failed" errors.

**Recommendation:** Use UUID for menu day IDs to guarantee uniqueness.

---

### DATA-07: Recipe Deletion Leaves Orphaned Menu Days
**Severity:** Medium
**Source:** 01-02-M10
**File:** Recipe storage (missing deletion method)

No `Delete` method exists yet. When added, needs to handle menu_days referencing the recipe. Without foreign key enforcement (DATA-04), orphaned records accumulate.

**Recommendation:** Prevent deletion of recipes used in active menus, or add proper cascade behavior.

---

### DATA-08: JoinHousehold Transaction Doesn't Lock Invite Code
**Severity:** Medium
**Source:** 01-02-M12
**File:** `backend/internal/services/household_service.go:85-132`

Invite code read before transaction starts. Race condition allows two users to read same unused invite simultaneously and both join household.

**Recommendation:** Move invite validation inside transaction or use SELECT FOR UPDATE.

---

### DATA-09: No Index on menu_days.date for Range Queries
**Severity:** High
**Source:** 02-02-H2
**File:** `backend/migrations/005_create_menus.sql`

Date filtering on menu_days has no index. As menus accumulate without cleanup mechanism, queries become O(n) table scans.

**Recommendation:** Add composite index on `(menu_id, date)`.

---

### DATA-10: Menu Generation Random Selection Allows Duplicates
**Severity:** Medium
**Source:** 01-02-M9
**File:** `backend/internal/services/menu_service.go:68`

Random recipe selection doesn't prevent duplicates. With 10 recipes and 5-day menu, ~40% chance of duplicate recipe. Poor user experience.

**Recommendation:** Use random selection without replacement (shuffle algorithm).

---

### DATA-11: Menu Day Date Format Not Validated
**Severity:** Medium
**Source:** 01-02-M13
**File:** `backend/internal/services/menu_service.go:56`

Dates generated correctly now, but if API ever accepts user-provided dates, no validation ensures YYYY-MM-DD format. Future-proofing needed.

**Recommendation:** Add date format validation helper for use in future APIs.

---

## 3. Input Validation

### VALID-01: Error Message Injection in HTTP Responses
**Severity:** High
**Source:** 01-01-H1
**Files:** Multiple handlers (`auth.go:30`, `recipes.go:72`, `offers.go:74,116,147`)

Direct string concatenation of `err.Error()` into JSON responses. If error contains `"`, JSON becomes malformed. Also leaks internal details (SQL errors, file paths).

**Recommendation:** Use `json.NewEncoder` for all error responses, create `writeError(w, code, message)` helper.

---

### VALID-02: No Request Body Size Limits
**Severity:** High
**Source:** 01-01-H2, 02-02-L2
**Files:** All handlers that decode JSON

No `http.MaxBytesReader` wrapper. Attacker can send multi-GB request body to exhaust memory and crash server.

**Recommendation:** Add `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` before all JSON decoding.

---

### VALID-03: Manual Path Splitting Instead of PathValue
**Severity:** Medium
**Source:** 01-01-M2
**Files:** `backend/internal/api/handlers/recipes.go:40-46`, `backend/internal/api/handlers/shopping.go:43-49`

Manual path parsing with `strings.Split()` when Go 1.22+ provides `r.PathValue("id")`. Other handlers use PathValue correctly.

**Recommendation:** Replace manual parsing with `r.PathValue("id")` for consistency.

---

### VALID-04: No Validation of Path Parameters
**Severity:** Medium
**Source:** 01-01-M3
**Files:** `household.go:121-125,161-165`, `recipes.go:46`, `shopping.go:49`

Path parameters extracted but not validated. No UUID format check, only empty string check. Invalid formats passed to database.

**Recommendation:** Validate UUID format with `uuid.Parse()` before passing to services.

---

### VALID-05: No Validation of GenerateMenuRequest Fields
**Severity:** Medium
**Source:** 01-01-M4
**File:** `backend/internal/api/handlers/menus.go:27-31`

Days count could be 0, negative, or 10,000. ServingSize could be 0 or negative. No bounds checking.

**Recommendation:** Validate days (1-31 range) and servings (1-100 range).

---

### VALID-06: Offers Handlers Silently Ignore Invalid Coordinates
**Severity:** Medium
**Source:** 01-01-M5
**Files:** `backend/internal/api/handlers/offers.go:36-50,88-102,129-143`

Invalid coordinate parameters silently fall back to defaults. User thinks they set coordinates but default used instead.

**Recommendation:** Return 400 error for invalid coordinate parameters instead of silent fallback.

---

### VALID-07: No Content-Type Validation on JSON Endpoints
**Severity:** Medium
**Source:** 01-01-M7
**Files:** All handlers that decode JSON

Handlers decode JSON without checking `Content-Type: application/json` header.

**Recommendation:** Validate Content-Type matches expected format.

---

### VALID-08: UpdateMemberStatus Allows Empty Updates
**Severity:** Medium
**Source:** 01-01-M8
**File:** `backend/internal/api/handlers/household.go:133-136`

Handler rejects empty updates correctly, but allows updates where both fields set to same values as current state (unnecessary database writes).

**Recommendation:** Service layer should check current values and no-op if unchanged.

---

### VALID-09: No Validation of Recipe Filter Query Parameters
**Severity:** Medium
**Source:** 01-01-M10
**File:** `backend/internal/api/handlers/recipes.go:21-24`

Name and tag filters have no length limits. Could be 10,000 characters, causing performance issues or SQL injection if storage doesn't sanitize.

**Recommendation:** Add length limits (name max 100 chars, tag max 50 chars).

---

### VALID-10: No Validation on Recipe Name Length
**Severity:** Medium
**Source:** 01-02-M2
**File:** `backend/internal/services/recipe_service.go:29-30`

Recipe name has no maximum length. User can submit 10,000-character names causing database bloat and UI rendering issues.

**Recommendation:** Add max length validation (200 characters reasonable for recipe name).

---

### VALID-11: No Validation on Recipe Servings Upper Bound
**Severity:** Medium
**Source:** 01-02-M3
**File:** `backend/internal/services/recipe_service.go:32-34`

Only checks positive value. User can create recipe with servings = 1,000,000, causing shopping list calculation overflow.

**Recommendation:** Add upper bound validation (max 100 servings).

---

### VALID-12: No Validation on Ingredients Array Size
**Severity:** Medium
**Source:** 01-02-M4
**File:** `backend/internal/services/recipe_service.go:35-37`

Ingredients array has no size limit. Recipe with 10,000 ingredients causes database bloat and performance degradation.

**Recommendation:** Add max array size (50 ingredients reasonable limit).

---

### VALID-13: Menu Service Doesn't Validate Days Count
**Severity:** Medium
**Source:** 01-02-M7
**File:** `backend/internal/services/menu_service.go:36-39`

Days count has no maximum. User can request 10,000-day menu generating 10,000 database records.

**Recommendation:** Add upper bound (max 31 days).

---

### VALID-14: Menu Service Doesn't Validate Servings Count
**Severity:** Medium
**Source:** 01-02-M8
**File:** `backend/internal/services/menu_service.go:40-43`

Same issue as VALID-11. User can request 1,000,000 servings causing calculation overflow.

**Recommendation:** Add upper bound (max 100 servings).

---

### VALID-15: Login and Register Don't Trim Email/Password Whitespace
**Severity:** Low
**Source:** 01-01-L4
**Files:** `backend/internal/api/handlers/auth.go:18-24,40-46`

Email and password not trimmed. Accidental whitespace causes login failures and duplicate accounts.

**Recommendation:** Trim both fields (email definitely, password debatable).

---

### VALID-16: GetAll Recipes Returns Empty Array vs Null Inconsistency
**Severity:** Medium
**Source:** 01-01-M9, 01-02-L6
**Files:** `recipes.go:19-35`, `recipe_storage.go:63-65`

Response could be `{"recipes":null}` or `{"recipes":[]}` depending on storage behavior. Frontend must handle both cases.

**Recommendation:** Normalize to always return empty array, never null.

---

## 4. Error Handling

### ERROR-01: Error Handling via String Matching
**Severity:** High
**Source:** 01-01-M1, 02-01-H4
**Files:** `household.go:82-89,140-146,169-178`, multiple handlers

Error routing based on `err.Error()` string comparison. Changing error message in service breaks HTTP status code mapping without compiler safety. Pattern appears in 15+ switch statements across codebase.

**Recommendation:** Define sentinel errors in domain package, use `errors.Is()` for type-safe error checking.

---

### ERROR-02: No Structured Error Response Helper
**Severity:** Medium
**Source:** 02-01-M7
**Files:** All handler files

Every handler writes JSON errors manually via string concatenation. Pattern duplicated 40+ times, causes JSON injection (VALID-01), and errors never logged.

**Recommendation:** Create `WriteError()` helper with proper JSON encoding and logging.

---

### ERROR-03: Errors Lost in Service-to-Handler Boundary
**Severity:** Medium
**Source:** 02-01-M8
**Files:** All handlers

When services return errors, handlers map to HTTP status and return generic message to client but never log actual error. Production failures invisible.

**Recommendation:** Log all errors before returning to client.

---

### ERROR-04: No Error Wrapping Context in Service Layer
**Severity:** Medium
**Source:** 02-01-M9
**Files:** All service files

Service methods return errors without adding context. Error "database locked" doesn't tell which operation failed. No breadcrumb trail from handler → service → storage.

**Recommendation:** Wrap errors with context using `fmt.Errorf("context: %w", err)`.

---

### ERROR-05: Recipe Create Returns 400 for All Errors
**Severity:** Low
**Source:** 01-01-L5
**File:** `backend/internal/api/handlers/recipes.go:70-73`

All service errors return 400 Bad Request. Internal errors (database failures) should be 500.

**Recommendation:** Use error type checking to map validation errors to 400, internal errors to 500.

---

### ERROR-06: No Distinction Between Business and Infrastructure Errors
**Severity:** Low
**Source:** 02-01-L2
**Category:** Error Classification

All errors are plain `error` type. Can't distinguish business rule violations (4xx) from infrastructure failures (5xx). Clients can't determine retry strategy.

**Recommendation:** Define BusinessError and InfrastructureError types for programmatic distinction.

---

## 5. Architecture & Code Quality

### ARCH-01: Services Depend on Concrete Storage Types
**Severity:** High
**Source:** 02-01-H1
**Files:** All 6 service files

No storage interfaces exist. Services coupled to concrete SQLite implementations. Blocks unit testing (requires real DB), prevents future database migration (big-bang change required), and eliminates deployment flexibility.

**Recommendation:** Define storage interfaces, services depend on interfaces, SQLite implements interfaces.

---

### ARCH-02: GetUserID and GetHouseholdID Have Inconsistent Signatures
**Severity:** High
**Source:** 01-01-H8
**File:** `backend/pkg/middleware/auth.go:48-58`

`GetUserID` takes `*http.Request`, `GetHouseholdID` takes `context.Context`. Inconsistent API causes confusion and mixed usage patterns in handlers.

**Recommendation:** Standardize both to accept `*http.Request`.

---

### ARCH-03: No Dependency Injection Container
**Severity:** High
**Source:** 02-01-H3
**File:** `backend/internal/api/router.go:12-81`

All dependencies manually wired in router. As codebase grows (currently 17+ constructor calls), becomes maintenance burden. No tooling to detect circular dependencies.

**Recommendation:** Extract to `wireDependencies()` helper function for current size. Revisit with DI library if grows beyond 20 services.

---

### ARCH-04: Handlers Access Context Inconsistently
**Severity:** Medium
**Source:** 02-01-M1
**Files:** `auth.go:49-58`, `household.go:20,43`

Mixed patterns for extracting context values. Some handlers pass `r`, others pass `r.Context()`. Cognitive load for developers.

**Recommendation:** Standardize both helper functions to take `*http.Request`.

---

### ARCH-05: Domain Package Contains Request/Response Types
**Severity:** Medium
**Source:** 02-01-M2
**Files:** `domain/auth.go`, `domain/household.go`

Domain package mixes pure entities with API-specific DTOs. Storage layer shouldn't need HTTP request/response types. API versioning forces domain changes.

**Recommendation:** Move request/response types to `internal/api/types/`, keep domain pure.

---

### ARCH-06: Router Does Too Much
**Severity:** Medium
**Source:** 02-01-M3
**File:** `backend/internal/api/router.go:12-81`

Single 70-line function handles dependency creation, service wiring, handler creation, and route definitions. Violates Single Responsibility Principle.

**Recommendation:** Extract dependency wiring to separate function (covered by ARCH-03 fix).

---

### ARCH-07: Services Can't Be Tested in Isolation
**Severity:** Medium
**Source:** 02-01-M4
**Files:** All service test files

Because services depend on concrete storage (ARCH-01), unit tests require real SQLite database setup. Only 1 of 6 services has tests due to setup complexity.

**Recommendation:** Blocked by ARCH-01 fix. After interfaces exist, create mocks for fast unit tests.

---

### ARCH-08: No Transaction Helper Abstraction
**Severity:** Medium
**Source:** 02-01-M5
**Files:** `household_service.go:85-135`

Transaction handling pattern (10+ lines of setup/teardown) duplicated across service methods. Error-prone, easy to forget `defer tx.Rollback()`.

**Recommendation:** Create `WithTransaction(fn func(tx) error)` helper to eliminate boilerplate.

---

### ARCH-09: Storage Exposes *sql.DB Directly
**Severity:** Medium
**Source:** 02-01-M6
**File:** `household_storage.go`

Storage has `DB()` method returning `*sql.DB` directly. Services call `.DB().Begin()` for transactions. Abstraction violation leaks SQL implementation to service layer.

**Recommendation:** Storage should provide `BeginTx()` method or accept optional transaction in methods.

---

### ARCH-10: pkg/ Package Contains Non-Reusable Code
**Severity:** Medium
**Source:** 02-01-M10
**Files:** `pkg/middleware/`, `pkg/utils/`

Go convention: `pkg/` should contain reusable libraries. Current code is application-specific (domain Claims struct, hardcoded CORS origins).

**Recommendation:** Move app-specific code to `internal/`, keep only truly generic utilities in `pkg/`.

---

### ARCH-11: Domain Package Has No Sub-Packages
**Severity:** Medium
**Source:** 02-01-M11
**File:** `backend/internal/domain/`

All domain types in single flat package. No clear bounded contexts. Manageable at current size (7 files) but becomes dumping ground as features grow.

**Recommendation:** Defer until natural split becomes clear (15+ domain files). Document for future reference.

---

### ARCH-12: No API Versioning Strategy
**Severity:** Medium
**Source:** 02-01-M12
**File:** `backend/internal/api/router.go`

All routes at root level with no versioning. Can't evolve API without breaking clients. Mobile apps can't be force-upgraded.

**Recommendation:** Document strategy (path-based `/api/v1/` recommended), implement when first breaking change needed.

---

### ARCH-13: Duplicate Coordinate Parsing Logic
**Severity:** Medium
**Source:** 01-01-M6
**Files:** `backend/internal/api/handlers/offers.go:36-50,88-102,129-143`

Coordinate parsing duplicated across three handlers. Bug fixes must be applied in multiple places.

**Recommendation:** Extract to `parseLocationParams(r)` helper function.

---

### ARCH-14: No Domain Services
**Severity:** Low
**Source:** 02-01-L1
**Files:** All service files

All "services" are infrastructure orchestration. Pure domain logic (menu generation algorithm, shopping aggregation) mixed with storage coordination. Harder to test and reuse.

**Recommendation:** Consider extracting domain services in `domain/services/` for pure business logic.

---

### ARCH-15: Handler Functions Have No Shared Pattern
**Severity:** Low
**Source:** 02-01-L3
**Files:** All handler files

Handlers follow similar but inconsistent patterns. Some extract auth early, others don't. No enforced template.

**Recommendation:** Document standard handler pattern in CONVENTIONS.md.

---

## 6. Performance

### PERF-01: N+1 Query Pattern in Shopping List Generation
**Severity:** High
**Source:** 01-02-M5, 02-02-H1
**Files:** `shopping_service.go:92-100`

Fetches each recipe individually in loop. 7-day menu makes 7 separate SELECT queries (30-100ms total latency). Can batch-fetch with single `IN` clause query.

**Recommendation:** Add `RecipeStorage.GetByIDs()` method for batch fetching (80% query reduction).

---

### PERF-02: Sequential Catalog Fetching
**Severity:** High
**Source:** 02-02-H3
**Files:** `tjek_service.go:324-355,398-429`

Offers fetched from 20+ catalogs sequentially. 20 catalogs × 300ms = 6 seconds per request. Can parallelize with goroutines for 5x speedup (1.2 seconds).

**Recommendation:** Use goroutines with semaphore (max 5 concurrent) to parallelize external API calls.

---

### PERF-03: No Caching for Tjek API Responses
**Severity:** High
**Source:** 02-02-H4
**Files:** Entire `tjek_service.go`

Every request triggers fresh API calls. Catalogs change weekly at most. With 1-hour TTL cache: first request 6s, subsequent <10ms. 600x speedup for cached requests. Also reduces external API load (good citizen behavior).

**Recommendation:** Add in-memory cache with TTL for catalog list, offers, and stores.

---

### PERF-04: SQLite WAL Mode Not Enabled
**Severity:** High
**Source:** 02-02-H5
**Files:** `db.go:11-32`

SQLite defaults to DELETE journal mode blocking all readers during writes. WAL mode allows concurrent reads during writes. Eliminates "database is locked" errors under concurrent load.

**Recommendation:** Execute `PRAGMA journal_mode=WAL` after opening database. Also set `synchronous=NORMAL` and `busy_timeout=5000`.

---

### PERF-05: Database Connection Pool Not Configured
**Severity:** High
**Source:** 01-02-L4, 02-02-H6
**Files:** `main.go:25-28`, `db.go:18-25`

Connection pool uses defaults (unlimited open, 2 idle). For SQLite + WAL, should limit total connections and keep idle connections warm.

**Recommendation:** Configure `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5min)`.

---

### PERF-06: GetAll Recipes Loads All Rows Without Limit
**Severity:** Medium
**Source:** 02-02-M1
**Files:** `recipe_storage.go:17-68`, `menu_service.go:25-29`

Loads all recipes into memory. At 500 recipes × 2KB avg = 1MB per request. Under concurrent load, causes memory churn.

**Recommendation:** Add pagination (default 50 per page) and load only summaries for menu generation.

---

### PERF-07: Recipe Tag Filter Uses String LIKE on JSON
**Severity:** Medium
**Source:** 02-02-M2
**File:** `recipe_storage.go:28-31`

Tag filtering uses `LIKE` on JSON text column forcing full table scan. At 500+ recipes, takes 10-20ms per query with no index usage possible.

**Recommendation:** Use SQLite JSON1 extension with `json_each()` or normalize tags to separate table.

---

### PERF-08: Tjek API Timeout Set to 30 Seconds
**Severity:** Medium
**Source:** 02-02-M3
**File:** `tjek_service.go:38-40`

30-second timeout too long. If Tjek API is slow, requests hang exhausting goroutines. Better to fail fast and retry.

**Recommendation:** Reduce to 5-10 seconds.

---

### PERF-09: No Exponential Backoff or Retry Logic
**Severity:** Medium
**Source:** 02-02-M4
**Files:** `tjek_service.go:466-481`

Tjek API calls have no retry mechanism. Transient failures cause complete request failures instead of automatic recovery.

**Recommendation:** Add exponential backoff retry (3 attempts max).

---

### PERF-10: Bubble Sort Used Instead of sort.Slice
**Severity:** Medium
**Source:** 01-02-M14, 02-02-M5
**Files:** `tjek_service.go:286-292,432-438`

Manual O(n²) bubble sort. With 100 offers: 4,950 comparisons vs 664 (7.5x slower, ~3ms difference). Go's `sort.Slice` already imported.

**Recommendation:** Replace with `sort.Slice` (O(n log n)).

---

### PERF-11: Menu Generation Algorithm Allows Recipe Duplication
**Severity:** Medium
**Source:** 02-02-M6
**File:** `menu_service.go:67-69`

Random selection without replacement. With 10 recipes and 5-day menu, ~40% chance of duplicate. Poor UX, not performance issue.

**Recommendation:** Use shuffle algorithm for variety (see DATA-10, same issue).

---

### PERF-12: No Database Query Timeout Context
**Severity:** Medium
**Source:** 02-02-M7
**Files:** All storage methods

Queries don't use context with timeout. Long-running queries can block indefinitely.

**Recommendation:** Use `db.QueryContext()` with 5-second timeout.

---

### PERF-13: Unbounded Response Size from Tjek API
**Severity:** Medium
**Source:** 02-02-M8
**Files:** `tjek_service.go:356-360`

Fetches all offers from 20+ catalogs (potentially 2,000 offers = 1MB) before limiting to 100. Discards 90%.

**Recommendation:** Early termination after collecting 100 offers, or use heap-based top-N selection.

---

### PERF-14: Shopping List Aggregation Creates Large Maps
**Severity:** Medium
**Source:** 02-02-M9
**Files:** `shopping_service.go:90-146`

Builds intermediate maps (~10KB per request). Under load (100 concurrent) = 1MB memory. Moderate GC pressure.

**Recommendation:** Pre-allocate slices with capacity hints or reuse buffers with sync.Pool.

---

### PERF-15: Sequential Tjek API Calls
**Severity:** High
**Source:** 01-02-M15, 02-02-H3
**Files:** `tjek_service.go:324-355,398-429`

Duplicate of PERF-02. Consolidated under PERF-02.

---

### PERF-16: Shopping Item ID Uses MD5
**Severity:** Medium
**Source:** 01-02-M6
**File:** `shopping_service.go:166-169`

MD5 used for generating shopping item IDs. Not a security issue (IDs aren't secrets), but MD5 is cryptographically broken. Code review red flag.

**Recommendation:** Use `hash/fnv` (faster, non-crypto) or UUIDs.

---

### PERF-17: Ingredient Categorization Uses Linear Lookup
**Severity:** Low
**Source:** 02-02-L1
**File:** `shopping_service.go:158-164`

Analysis shows this is O(1) map lookup, not linear. No performance issue. Correctly implemented.

**Recommendation:** None needed (incorrectly flagged, actually optimal).

---

## 7. Deployment & Observability

### DEPLOY-01: fly.toml Has Conflicting Memory Settings
**Severity:** High
**Source:** 02-02-H7
**File:** `fly.toml:19-23`

Specifies both `memory = '1gb'` and `memory_mb = 256` in same block. Unclear which Fly.io uses. Could cause OOM if 256MB or waste resources if 1GB.

**Recommendation:** Remove conflict, use only `memory_mb = 512` (reasonable for Go app).

---

### DEPLOY-02: Migration File Paths Are Relative
**Severity:** Medium
**Source:** 01-02-L1
**Files:** `db.go:52-58`

Migrations read via relative paths. Server must be run from `backend/` directory. Deployment fragility and tests require `os.Chdir()` workarounds.

**Recommendation:** Use `embed.FS` to embed migrations in binary.

---

### DEPLOY-03: Database Path Not Validated
**Severity:** Medium
**Source:** 02-01-M13
**Files:** `main.go:19-22`

`DATABASE_PATH` used as-is without validation. Invalid path could create database in wrong location, lost on container restart.

**Recommendation:** Validate path is absolute and parent directory exists/writable.

---

### DEPLOY-04: Migrations Run on Every Startup
**Severity:** Medium
**Source:** 02-01-M14
**Files:** `db.go:60-71`

Queries for each migration version on every startup (6 SELECT queries). Adds 10-50ms unnecessary latency, especially problematic with Fly.io scale-to-zero.

**Recommendation:** Fetch all applied versions in single query, check set membership.

---

### DEPLOY-05: No Graceful Shutdown Handling
**Severity:** Medium
**Source:** 01-02-L3, 02-01-L4
**Files:** `main.go:35`

Server uses `log.Fatal()` immediately exiting on signal. In-flight requests fail, SQLite writes may not flush. Poor UX during deploys.

**Recommendation:** Implement graceful shutdown with 10-second timeout for in-flight requests.

---

### DEPLOY-06: Health Endpoint Reveals Server Operational / No Depth
**Severity:** Low
**Source:** 01-01-L2, 02-01-L5
**Files:** `health.go:5-8`

Public `/health` endpoint confirms server running (minor reconnaissance info) and doesn't check database connectivity. Health check passes even if DB unreachable.

**Recommendation:** Add database ping to health check, return 503 if unhealthy.

---

### DEPLOY-07: Health Check Endpoint Missing
**Severity:** Medium
**Source:** 02-02-M11
**Category:** Deployment

No `/health` endpoint for Fly.io health checks. Platform uses TCP only. May route traffic to unhealthy instances (DB inaccessible, migrations failed).

**Recommendation:** Add health endpoint with DB ping, update fly.toml with HTTP check.

---

### DEPLOY-08: No Structured Logging Framework
**Severity:** Medium
**Source:** 02-01-M15
**Files:** Entire codebase

Only uses `log.Printf()` and `log.Fatal()`. No structured logging, levels, or context. Can't filter by level, add request ID, or parse logs in production.

**Recommendation:** Adopt `slog` from Go 1.21+ standard library for structured JSON logging.

---

### DEPLOY-09: No Request ID Tracing
**Severity:** Medium
**Source:** 02-01-M16
**Files:** No request ID middleware exists

No correlation between requests and log entries. Can't trace single request through handler → service → storage. Debugging production issues impossible with concurrent requests.

**Recommendation:** Add request ID middleware generating UUID per request, include in all logs.

---

### DEPLOY-10: Health Endpoint Returns Plaintext-ish JSON
**Severity:** Low
**Source:** 01-01-L1
**File:** `health.go:5-8`

Doesn't set `Content-Type: application/json` header. Minor inconsistency with other endpoints.

**Recommendation:** Add `Content-Type` header for consistency.

---

### DEPLOY-11: GetMemberStatuses Returns Empty Array
**Severity:** Low
**Source:** 01-01-L3
**File:** `household.go:104-111`

Returns empty array if household has no members. Every household should have at least one (creator). Empty array indicates data corruption.

**Recommendation:** Log warning if members array empty, consider returning error.

---

### DEPLOY-12: Cold Starts with min_machines_running=0
**Severity:** Medium
**Source:** 02-02-M10
**File:** `fly.toml:16`

App scales to zero when idle. First request after scale-down has 2-5 second cold start delay. Trade-off between cost savings and UX.

**Recommendation:** Decision based on usage: keep 0 for personal use, set to 1 ($1.94/month) for public beta.

---

### DEPLOY-13: No Monitoring/Metrics Collection
**Severity:** Low
**Source:** 02-01-L7
**Category:** Observability

No metrics for request counts, latencies, error rates, or query performance. Can't detect performance regressions or capacity plan.

**Recommendation:** Add Prometheus metrics middleware with `/metrics` endpoint.

---

## CONCERNS.md Verification

### Verified Issues from Original CONCERNS.md

**Security Concerns:**
- ✅ **AUTH-01** - Empty JWT secret (CONCERNS.md line 81-83)
- ✅ **AUTH-02** - Shopping list auth missing (line 67-68)
- ✅ **VALID-01** - Error message injection (line 8-11)
- ✅ **VALID-02** - Request body limits (line 104-107)
- ✅ **AUTH-05** - Rate limiting (line 98-101)
- ✅ **AUTH-06** - CORS localhost only (line 85-89)
- ✅ **AUTH-07** - Email validation (line 110-113)

**Data Integrity:**
- ✅ **DATA-01** - Registration not transactional (line 144-149)
- ✅ **DATA-04** - Foreign key enforcement (implied by cascade concerns)

**Code Quality:**
- ✅ **ERROR-01** - Error string matching (line 14-17)
- ✅ **ARCH-13** - Duplicate coordinate parsing (line 19-23)
- ✅ **VALID-03** - Manual path splitting (line 38-42)
- ✅ **PERF-10** - Bubble sort (line 25-29)

**Performance:**
- ✅ **PERF-01** - N+1 queries (line 117-121)
- ✅ **PERF-02/15** - Sequential fetching (line 123-127)
- ✅ **PERF-03** - No caching (line 129-133)
- ✅ **PERF-04** - WAL mode (line 161-163)
- ✅ **DEPLOY-01** - fly.toml memory (line 70-74)
- ✅ **DEPLOY-12** - Cold starts (line 166-170)

**Deployment:**
- ✅ **DEPLOY-02** - Migration paths (line 137-142)

### Net-New Discoveries Not in CONCERNS.md

**Critical (4 new):**
- AUTH-03: JWT algorithm not pinned
- AUTH-04: No IDOR protection
- DATA-02: Zero recipes silent failure

**High-Severity (14 new):**
- AUTH-08: Login timing attack
- AUTH-09: Context key collision
- AUTH-10: Password max length
- AUTH-11: Package-level JWT state
- AUTH-12: Env var validation
- DATA-03: JSON injection in tag filter
- DATA-05: SQL string concatenation
- DATA-06: Menu day PK collision
- DATA-09: Missing date index
- ARCH-01: No storage interfaces
- ARCH-02: Inconsistent signatures
- ARCH-03: No DI container
- DEPLOY-01: Memory conflict
- PERF-05: Connection pool

**Medium/Low (30+ new findings across all areas)**

### CONCERNS.md Items NOT Confirmed

All items from CONCERNS.md were verified during the reviews. No false positives identified. Several items were found to be more severe than initially suspected (e.g., foreign key enforcement was "mentioned" in CONCERNS.md but is actually HIGH severity).

---

**End of Comprehensive Findings Report**
**Total Findings:** 96 unique issues requiring remediation
**Next Steps:** Phase 3 plan prioritization and remediation scheduling
