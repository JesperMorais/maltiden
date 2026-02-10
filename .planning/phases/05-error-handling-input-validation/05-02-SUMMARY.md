---
phase: 05-error-handling-input-validation
plan: 02
subsystem: backend-input-validation
tags: [input-validation, body-limits, content-type, path-params, business-rules]
depends_on:
  requires: [05-01]
  provides: [input-validation-layer, body-size-limits, coordinate-validation, business-rule-bounds]
  affects: [06-xx, 07-xx]
tech-stack:
  added: []
  patterns: [decode-json-helper, validate-id-helper, location-param-parser]
key-files:
  created: []
  modified:
    - backend/internal/api/handlers/response.go
    - backend/internal/api/handlers/auth.go
    - backend/internal/api/handlers/household.go
    - backend/internal/api/handlers/recipes.go
    - backend/internal/api/handlers/menus.go
    - backend/internal/api/handlers/shopping.go
    - backend/internal/api/handlers/offers.go
    - backend/internal/domain/errors.go
    - backend/internal/services/menu_service.go
    - backend/internal/services/recipe_service.go
    - backend/internal/services/auth_service.go
decisions:
  - DecodeJSON helper combines MaxBytesReader + Content-Type check + JSON decode
  - ValidateID uses length check (not uuid.Parse) for prefixed IDs
  - Menu defaults preserved for 0 values (backwards compat), reject negative/large
  - Password whitespace NOT trimmed (intentional user input)
metrics:
  duration: 5min
  completed: 2026-02-09
---

# Phase 5 Plan 2: Input Validation Sweep Summary

DecodeJSON/ValidateID helpers with 1MB body limits, Content-Type enforcement, coordinate validation, and business rule bounds on all endpoints.

## One-liner

Closed 14 input validation findings with DecodeJSON (body limit + Content-Type), ValidateID, parseLocationParams, and service-layer bounds on names, servings, days, ingredients, and email format.

## What Was Done

### Task 1: Handler-level validation infrastructure (2dc4fb0)

**DecodeJSON helper (response.go):**
- Combines `http.MaxBytesReader` (1MB limit), Content-Type check, and JSON decoding
- Returns 415 for non-JSON Content-Type, 413 for oversized bodies, 400 for malformed JSON
- Replaced all 6 raw `json.NewDecoder(r.Body).Decode()` calls across auth, household, recipes, menus, shopping handlers

**ValidateID helper (response.go):**
- Validates non-empty and minimum length (4 chars) for prefixed UUID format (e.g., `hm_<uuid>`)
- Applied to: memberID, targetID, itemID path params; menuID query param

**PathValue migration (recipes.go):**
- Replaced `strings.Split(path, "/")` with `r.PathValue("id")` in GetByID (route already registered as `GET /recipes/{id}`)

**Offers coordinate validation (offers.go):**
- Extracted `parseLocationParams()` helper that returns error for unparseable lat/lng/radius
- Returns 400 "invalid_coordinates" instead of silently falling back to defaults
- Extracted `parseExcludeStores()` helper, eliminating duplicate parsing in SearchOffers and GetDiscounts

### Task 2: Service-level business rule validation (cfc7493)

**Menu service bounds (VALID-13, VALID-14):**
- Days: default 5 when 0 (backwards compat), reject outside 1-31
- Servings: default 4 when 0 (backwards compat), reject outside 1-100
- Validation moved before recipe fetch for early return

**Recipe service bounds (VALID-10, VALID-11, VALID-12):**
- Name length: max 200 characters
- Servings: max 100 (upper bound added to existing >0 check)
- Ingredients array: max 50 items

**Auth service (VALID-15):**
- Email and name trimmed with `strings.TrimSpace()` in Register
- Email trimmed in Login
- Basic email format validation (must contain "@" and ".")
- Password whitespace intentionally NOT trimmed

**Empty array normalization (VALID-16):**
- Already handled in recipe_storage.go GetAll() (lines 63-65: `if recipes == nil { recipes = []domain.RecipeSummary{} }`)

**New sentinel errors added to domain/errors.go:**
- `ErrInvalidDays` -- menu days out of range
- `ErrNameTooLong` -- recipe name exceeds 200 chars
- `ErrTooManyIngredients` -- recipe ingredients exceed 50
- `ErrInvalidEmail` -- email missing @ or .

## Findings Addressed

| Finding | Severity | Fix |
|---------|----------|-----|
| VALID-02: No body size limits | High | DecodeJSON with 1MB MaxBytesReader |
| VALID-03: Manual path splitting | Medium | r.PathValue("id") in recipes GetByID |
| VALID-04: No UUID validation | Medium | ValidateID helper for all path/query IDs |
| VALID-05: No GenerateMenuRequest validation | Medium | Days 1-31, servings 1-100 bounds |
| VALID-06: Offers ignore invalid coordinates | Medium | parseLocationParams returns error, handler returns 400 |
| VALID-07: No Content-Type validation | Medium | DecodeJSON checks Content-Type header |
| VALID-08: Empty updates allowed | Low | Already fixed in Phase 4 (verified) |
| VALID-09: No recipe filter validation | Low | Filters are query params, safe via parameterized SQL |
| VALID-10: No recipe name length | Medium | Max 200 chars in recipe service |
| VALID-11: No servings upper bound | Medium | Max 100 in recipe and menu services |
| VALID-12: No ingredients array size | Medium | Max 50 in recipe service |
| VALID-13: Menu days count | Medium | 1-31 range with 0-default |
| VALID-14: Menu servings count | Medium | 1-100 range with 0-default |
| VALID-15: No email trimming | Low | TrimSpace on email in Register and Login |
| VALID-16: Empty array vs null | Low | Already handled in recipe_storage.go |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Sentinel errors needed by Task 1 handlers**
- **Found during:** Task 1
- **Issue:** Handlers reference ErrInvalidDays, ErrNameTooLong, ErrTooManyIngredients, ErrInvalidEmail but these were planned for Task 2
- **Fix:** Added sentinel error definitions to domain/errors.go as part of Task 1 commit (required for compilation)
- **Impact:** Task 2 focuses only on service-layer logic, errors already defined

**2. [Rule 1 - Bug] Password trimming omitted intentionally**
- **Found during:** Task 2
- **Issue:** Plan suggested trimming password whitespace, but this changes the actual credential value
- **Fix:** Only email and name are trimmed; password whitespace preserved as intentional user input
- **Impact:** More secure behavior -- users with intentional spaces in passwords are not affected

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| DecodeJSON helper pattern | Consolidates 3 checks (size, content-type, decode) into one call; impossible to forget |
| ValidateID uses length not uuid.Parse | IDs use prefixed format (rec_, hm_, menu_); uuid.Parse would reject valid IDs |
| Menu 0-value defaults preserved | Backwards compatibility with frontend that may send 0 for "use defaults" |
| Password whitespace NOT trimmed | Trimming passwords changes credentials; could lock users out |
| Validation before business logic in menu service | Early return avoids unnecessary DB queries |

## Verification Results

- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `grep json.NewDecoder(r.Body)` in handlers returns only response.go (DecodeJSON itself)
- [x] `grep strings.Split(path` in handlers returns 0 matches
- [x] `grep MaxBytesReader` in response.go returns 1 match
- [x] All 4 new domain errors have handler mappings
- [x] 14 VALID-* findings addressed

## Next Phase Readiness

Phase 5 is complete (2/2 plans). Phase 6 (Performance & Database) can proceed:
- All error infrastructure is in place
- All input validation is in place
- Handler patterns are consistent and ready for any new endpoints
