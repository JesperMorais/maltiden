# Auth System & HTTP Handlers Security Review - Findings

**Review Date:** 2026-02-09
**Files Reviewed:** 12 (5 auth system + 7 handlers)
**Reviewer:** Code Review Agent

## Summary

| Severity | Count | Category Focus |
|----------|-------|----------------|
| **Critical** | 4 | Auth bypass, IDOR, secret handling |
| **High** | 8 | Input validation, injection risks, auth gaps |
| **Medium** | 13 | Error handling, code quality, consistency |
| **Low** | 5 | Information disclosure, minor issues |
| **Total** | 30 | |

---

## Critical Severity Issues

### C1. Empty JWT Secret Allows Token Forgery
**File:** `backend/pkg/utils/jwt.go:11`
**Severity:** Critical
**Category:** Authentication Bypass

**Issue:**
```go
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
```

If `JWT_SECRET` environment variable is not set or is empty, `jwtSecret` becomes an empty byte slice. This means all tokens are signed with an empty key, allowing any attacker to forge valid JWT tokens.

**Impact:**
- Complete authentication bypass
- Attacker can impersonate any user by crafting a token with arbitrary `userID` and `householdID`
- Full access to all protected endpoints

**Why It Matters:**
This is a complete security failure. Any deployment without `JWT_SECRET` set is fully compromised.

**Verified from CONCERNS.md:** Yes, already documented (line 81-83)

**Recommendation:**
- Add startup validation in `main.go` that fatals if `JWT_SECRET` is empty or < 32 characters
- Consider passing secret as dependency rather than package-level variable
- Add unit test that validates token generation/validation with non-empty secret

---

### C2. Shopping List Endpoints Lack Authentication
**Files:** `backend/internal/api/router.go:76-77`
**Severity:** Critical
**Category:** Authorization Bypass

**Issue:**
```go
mux.HandleFunc("GET /shopping-list", shoppingHandler.GetShoppingList)
mux.HandleFunc("PATCH /shopping-list/items/{id}", shoppingHandler.UpdateItem)
```

Both shopping list endpoints use `HandleFunc` instead of `Handle` with `middleware.RequireAuth`. Any unauthenticated user can:
- View any household's shopping list by guessing menuID
- Modify shopping item checked status

**Impact:**
- IDOR vulnerability: Access to other households' shopping data
- Data manipulation: Checking/unchecking items for other users
- Privacy violation: Shopping lists reveal dietary patterns

**Why It Matters:**
While menu IDs are UUIDs (hard to guess), this is security by obscurity. A leaked menuID (URL shared accidentally, logs, analytics) exposes the data.

**Verified from CONCERNS.md:** Yes, already documented (line 67-68)

**Recommendation:**
```go
mux.Handle("GET /shopping-list", middleware.RequireAuth(
    http.HandlerFunc(shoppingHandler.GetShoppingList),
))
mux.Handle("PATCH /shopping-list/items/{id}", middleware.RequireAuth(
    http.HandlerFunc(shoppingHandler.UpdateItem),
))
```
- Add householdID verification in handlers to ensure menuID belongs to authenticated user's household

---

### C3. JWT Algorithm Not Pinned - "none" Algorithm Attack
**File:** `backend/pkg/utils/jwt.go:34-37`
**Severity:** Critical
**Category:** Authentication Bypass

**Issue:**
```go
token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
    return jwtSecret, nil
})
```

The validation does not verify the signing algorithm. The `golang-jwt/jwt` library has protections against the "none" algorithm attack by default, but the code does not explicitly validate that the algorithm is HS256.

**Impact:**
- Potential for algorithm confusion attacks
- If library defaults change, vulnerability introduced silently
- No defense in depth

**Why It Matters:**
Best practice is to explicitly validate the expected algorithm. An attacker could potentially craft a token with a different algorithm.

**Recommendation:**
```go
token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
    // Validate algorithm is what we expect
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    if token.Method.Alg() != "HS256" {
        return nil, fmt.Errorf("unexpected algorithm: %v", token.Method.Alg())
    }
    return jwtSecret, nil
})
```

---

### C4. No IDOR Protection in Menu/Shopping Operations
**Files:**
- `backend/internal/api/handlers/menus.go:33-36`
- `backend/internal/api/handlers/shopping.go:26-29`
**Severity:** Critical
**Category:** Insecure Direct Object Reference (IDOR)

**Issue:**
The `GetShoppingList` handler accepts a `menuId` query parameter but never verifies that the menu belongs to the authenticated user's household (because the endpoint isn't even authenticated - see C2).

Even if authentication is added (per C2), the handlers don't validate ownership:
```go
// shopping.go:26
list, err := h.shoppingService.GetShoppingList(menuID)
```

Similarly, menu generation doesn't validate that generated menus are scoped to the authenticated household beyond the initial `Generate` call.

**Impact:**
- Access to other households' menus and shopping lists
- Potential data manipulation across household boundaries

**Why It Matters:**
The system stores menu data per household but doesn't enforce access control at the handler layer. Adding auth (C2) is necessary but not sufficient.

**Recommendation:**
- Modify `GetShoppingList` to accept householdID from auth context
- Add a storage-layer check: `GetShoppingListForHousehold(menuID, householdID)`
- Return 403 Forbidden if menu doesn't belong to the household
- Add similar validation for all menu operations

---

## High Severity Issues

### H1. Error Message Injection in HTTP Responses
**Files:**
- `backend/internal/api/handlers/auth.go:30`
- `backend/internal/api/handlers/recipes.go:72`
- `backend/internal/api/handlers/offers.go:74,116,147`
**Severity:** High
**Category:** JSON Injection / Information Disclosure

**Issue:**
```go
// auth.go:30
http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)

// offers.go:74
http.Error(w, `{"error":"failed to fetch offers","details":"`+err.Error()+`"}`, http.StatusBadGateway)
```

Direct string concatenation of `err.Error()` into JSON response strings:
1. **JSON Injection:** If error message contains `"`, the JSON becomes malformed
2. **Information Disclosure:** Internal error details (SQL errors, file paths, stack traces) leak to clients

**Impact:**
- Broken JSON responses crash frontend
- Database schema exposed via SQL errors
- Internal file paths disclosed
- Stack traces may reveal code structure

**Why It Matters:**
Error messages are user-controlled in some cases (e.g., database constraint violations include user input). A quote in input can break responses.

**Verified from CONCERNS.md:** Yes, already documented (line 8-11)

**Recommendation:**
- Use `json.NewEncoder` for all error responses
- Create `writeError(w, code, message)` helper that properly serializes
- Never pass `err.Error()` to clients; log it server-side
- Use generic error messages for clients: "internal_error", "validation_failed"

---

### H2. No Request Body Size Limits
**Files:** All handlers that decode JSON (auth.go:21, household.go:75,128, recipes.go:65, menus.go:28, shopping.go:58, offers.go N/A)
**Severity:** High
**Category:** Denial of Service (DoS)

**Issue:**
```go
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
```

No `http.MaxBytesReader` wrapper. An attacker can send multi-GB request bodies to exhaust memory.

**Impact:**
- Memory exhaustion
- Server crash
- Denial of service for all users

**Why It Matters:**
A single malicious request with a 1GB JSON body can crash the server.

**Verified from CONCERNS.md:** Yes, already documented (line 104-107)

**Recommendation:**
```go
// Wrap body with size limit (e.g., 1MB)
r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    // Check if it was a size limit error
    if err.Error() == "http: request body too large" {
        http.Error(w, `{"error":"request_too_large"}`, http.StatusRequestEntityTooLarge)
        return
    }
    http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
    return
}
```

---

### H3. No Rate Limiting on Authentication Endpoints
**Files:** `backend/internal/api/router.go:40-41`, `backend/internal/api/handlers/auth.go`
**Severity:** High
**Category:** Brute Force / Credential Stuffing

**Issue:**
```go
mux.HandleFunc("POST /auth/register", authHandler.Register)
mux.HandleFunc("POST /auth/login", authHandler.Login)
```

No rate limiting middleware. Attackers can:
- Brute force passwords
- Enumerate valid email addresses (registration fails differently for existing emails)
- Perform credential stuffing attacks

**Impact:**
- Account compromise via brute force
- Email enumeration for phishing
- Resource exhaustion (bcrypt is CPU-intensive)

**Why It Matters:**
While bcrypt cost 12 provides ~250ms natural rate limiting per attempt, a distributed attack with 100 IPs makes 400 login attempts per second.

**Verified from CONCERNS.md:** Yes, already documented (line 98-101)

**Recommendation:**
- Add per-IP rate limiting middleware (e.g., 5 requests per minute for auth endpoints)
- Consider CAPTCHA after N failed attempts
- Implement account lockout after failed login attempts
- Add logging for failed auth attempts to detect attacks

---

### H4. CORS Restricted to Localhost Only
**File:** `backend/pkg/middleware/cors.go:11`
**Severity:** High
**Category:** Production Deployment Blocker

**Issue:**
```go
if origin == "http://localhost:5173" || origin == "http://localhost:4173" || origin == "http://127.0.0.1:5173" {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

Only localhost origins allowed. Production frontend will be blocked by CORS.

**Impact:**
- Production deployment non-functional if frontend and backend are on different domains
- API unusable from production frontend

**Why It Matters:**
This works in development but silently breaks in production. The backend will reject all requests from the production frontend.

**Verified from CONCERNS.md:** Yes, already documented (line 85-89)

**Recommendation:**
- Make allowed origins configurable via environment variable: `CORS_ORIGINS`
- Parse as comma-separated list
- Default to localhost for development
- In production: set to actual frontend domain(s)
```go
allowedOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
origin := r.Header.Get("Origin")
for _, allowed := range allowedOrigins {
    if origin == strings.TrimSpace(allowed) {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        break
    }
}
```

---

### H5. No Email Format Validation on Registration
**Files:** `backend/internal/api/handlers/auth.go:18-24`, service layer (auth_service.go)
**Severity:** High
**Category:** Input Validation

**Issue:**
Registration accepts any string as email without format validation:
```go
var req domain.RegisterRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
```

No check that `req.Email` is a valid email format. Service only checks password length (≥8).

**Impact:**
- Invalid emails in database (e.g., "not-an-email", "user@", "@domain.com")
- Password reset flow (when implemented) will fail
- User can't recover account
- Duplicate registrations with case-sensitive emails ("User@example.com" vs "user@example.com")

**Why It Matters:**
Once invalid email is in database, user account is permanently broken.

**Verified from CONCERNS.md:** Yes, already documented (line 110-113)

**Recommendation:**
- Normalize email: lowercase, trim whitespace
- Validate basic format: `[local]@[domain].[tld]` regex
- SQLite UNIQUE constraint already exists but is case-sensitive
- Consider email verification flow (send confirmation email)

---

### H6. Login Endpoint Leaks Errors Before bcrypt Comparison
**File:** `backend/internal/api/handlers/auth.go:49-52`
**Severity:** High
**Category:** Timing Attack (minor) / Email Enumeration

**Issue:**
```go
resp, err := h.authService.Login(req)
if err != nil {
    http.Error(w, `{"error":"invalid_credentials"}`, http.StatusUnauthorized)
    return
}
```

The handler correctly returns generic "invalid_credentials" message, but the service layer may return different errors. If the service distinguishes between "user not found" vs "wrong password", timing differences allow email enumeration.

**Impact:**
- Attacker can enumerate valid email addresses
- Timing side-channel reveals whether email exists

**Why It Matters:**
Valid email list is valuable for phishing and credential stuffing.

**Note:** Handler implementation is correct (generic message). Need to verify service layer doesn't leak via timing.

**Recommendation:**
- Ensure service always performs bcrypt comparison even for non-existent users (constant-time)
- Or: hash a dummy password when user not found to equalize timing
- Verify service returns identical error for both cases

---

### H7. Context Key Type Collision Risk
**File:** `backend/pkg/middleware/auth.go:10-13`
**Severity:** High
**Category:** Context Injection / Collision

**Issue:**
```go
type contextKey string

const UserIDKey contextKey = "user_id"
const HouseholdIDKey contextKey = "household_id"
```

Context keys defined as typed string constants, which is correct. However, the actual string values are simple ("user_id", "household_id"). If any other middleware or library uses the same string, collision occurs.

**Impact:**
- Context value overwritten by other middleware
- Authentication context lost
- Potential privilege escalation if malicious middleware sets these keys

**Why It Matters:**
Context collisions are subtle and hard to debug. They can lead to auth bypasses.

**Recommendation:**
- Use unexported struct type for keys (idiomatic Go pattern):
```go
type contextKey struct{}

var userIDKey = contextKey{}
var householdIDKey = contextKey{}

// Usage:
ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
```
This makes collision impossible as the key is the type itself, not a string.

---

### H8. GetUserID and GetHouseholdID Have Inconsistent Signatures
**File:** `backend/pkg/middleware/auth.go:48-58`
**Severity:** High (maintainability)
**Category:** API Inconsistency

**Issue:**
```go
// Helper to get userID from context
func GetUserID(r *http.Request) string {
    userID, _ := r.Context().Value(UserIDKey).(string)
    return userID
}

// Helper to get householdID from context
func GetHouseholdID(ctx context.Context) string {
    householdID, _ := ctx.Value(HouseholdIDKey).(string)
    return householdID
}
```

`GetUserID` accepts `*http.Request` while `GetHouseholdID` accepts `context.Context`. This inconsistency:
- Makes API confusing
- Causes handlers to use different patterns (line 20 vs 43 in household.go)

**Impact:**
- Code inconsistency
- Bugs from mixing patterns
- Harder to maintain

**Why It Matters:**
Handlers use both patterns leading to confusion. Some extract context first, others pass request directly.

**Recommendation:**
- Standardize both to accept `*http.Request`:
```go
func GetUserID(r *http.Request) string { ... }
func GetHouseholdID(r *http.Request) string {
    householdID, _ := r.Context().Value(HouseholdIDKey).(string)
    return householdID
}
```

---

## Medium Severity Issues

### M1. Error Handling via String Matching (Fragile)
**Files:**
- `backend/internal/api/handlers/household.go:82-89,140-146,169-178`
**Severity:** Medium
**Category:** Maintainability / Fragility

**Issue:**
```go
switch err.Error() {
case "invalid_code":
    http.Error(w, `{"error":"invalid_code"}`, http.StatusBadRequest)
case "already_member":
    http.Error(w, `{"error":"already_member"}`, http.StatusConflict)
default:
    http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
}
```

Error routing based on string comparison of `err.Error()`. Changing error message in service breaks HTTP status code mapping.

**Impact:**
- No compiler safety
- Refactoring error messages silently breaks error handling
- All errors become 500 if string doesn't match

**Why It Matters:**
Service layer error message change (even a typo fix) breaks handler behavior without warning.

**Verified from CONCERNS.md:** Yes, already documented (line 14-17)

**Recommendation:**
- Define sentinel errors in domain package:
```go
var (
    ErrInvalidCode   = errors.New("invalid_code")
    ErrAlreadyMember = errors.New("already_member")
)
```
- Use `errors.Is()` in handlers:
```go
if errors.Is(err, domain.ErrInvalidCode) {
    http.Error(w, `{"error":"invalid_code"}`, http.StatusBadRequest)
    return
}
```

---

### M2. Recipes Use Manual Path Splitting Instead of PathValue
**Files:**
- `backend/internal/api/handlers/recipes.go:40-46`
- `backend/internal/api/handlers/shopping.go:43-49`
**Severity:** Medium
**Category:** Code Quality / Inconsistency

**Issue:**
```go
// recipes.go:40-46
path := r.URL.Path
parts := strings.Split(path, "/")
if len(parts) < 3 {
    http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
    return
}
id := parts[len(parts)-1]
```

Manual path parsing when Go 1.22+ provides `r.PathValue("id")`. Other handlers use PathValue correctly (household.go:121,161).

**Impact:**
- Inconsistent code patterns
- More error-prone (off-by-one errors)
- Harder to maintain

**Why It Matters:**
Manual parsing is fragile. Path like `/recipes/` (no ID) could crash with index out of bounds.

**Verified from CONCERNS.md:** Yes, already documented (line 38-42)

**Recommendation:**
```go
id := r.PathValue("id")
if id == "" {
    http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
    return
}
```

---

### M3. No Validation of Path Parameters
**Files:**
- `backend/internal/api/handlers/household.go:121-125,161-165`
- `backend/internal/api/handlers/recipes.go:46`
- `backend/internal/api/handlers/shopping.go:49`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
memberID := r.PathValue("id")
if memberID == "" {
    http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
    return
}
// Immediately pass to service with no format validation
```

Path parameters extracted but not validated:
- No UUID format check
- Empty string check only
- Could be SQL injection vector if storage layer isn't using parameterized queries

**Impact:**
- Invalid IDs passed to database
- Potential SQL injection if storage is vulnerable
- Unhelpful error messages (database error instead of "invalid ID format")

**Why It Matters:**
UUIDs have a specific format. Invalid formats should be rejected at handler layer.

**Recommendation:**
```go
memberID := r.PathValue("id")
if memberID == "" {
    http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
    return
}
// Validate UUID format
if _, err := uuid.Parse(memberID); err != nil {
    http.Error(w, `{"error":"invalid_id_format"}`, http.StatusBadRequest)
    return
}
```

---

### M4. No Validation of GenerateMenuRequest Fields
**File:** `backend/internal/api/handlers/menus.go:27-31`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
var req domain.GenerateMenuRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
    return
}
// No validation of req fields before passing to service
```

No validation of `GenerateMenuRequest` fields:
- Days count (could be 0, negative, or 10000)
- ServingSize (could be 0 or negative)
- RecipeIDs (if provided, could be empty array or invalid UUIDs)

**Impact:**
- Invalid menus generated
- Database pollution with bad data
- Potential infinite loops or crashes in service layer

**Why It Matters:**
Service layer should handle business logic, but basic input sanity checks belong in handler.

**Recommendation:**
```go
// After decode:
if req.Days <= 0 || req.Days > 30 {
    http.Error(w, `{"error":"invalid_days_count"}`, http.StatusBadRequest)
    return
}
if req.ServingSize <= 0 || req.ServingSize > 100 {
    http.Error(w, `{"error":"invalid_serving_size"}`, http.StatusBadRequest)
    return
}
```

---

### M5. Offers Handlers Silently Ignore Invalid Coordinate Parameters
**Files:** `backend/internal/api/handlers/offers.go:36-50,88-102,129-143`
**Severity:** Medium
**Category:** Input Validation / Silent Failure

**Issue:**
```go
if latStr := r.URL.Query().Get("lat"); latStr != "" {
    if parsed, err := strconv.ParseFloat(latStr, 64); err == nil {
        lat = parsed
    }
}
```

Invalid coordinate parameters are silently ignored. If user passes `lat=invalid`, it falls back to default without error.

**Impact:**
- User confusion: thinks they set coordinates but default used
- Silent failures make debugging hard
- API behavior unclear

**Why It Matters:**
Clients expect errors for invalid input. Silent fallback masks bugs.

**Recommendation:**
```go
if latStr := r.URL.Query().Get("lat"); latStr != "" {
    parsed, err := strconv.ParseFloat(latStr, 64)
    if err != nil {
        http.Error(w, `{"error":"invalid_latitude"}`, http.StatusBadRequest)
        return
    }
    lat = parsed
}
```

---

### M6. Duplicate Coordinate Parsing Logic
**Files:** `backend/internal/api/handlers/offers.go:36-50,88-102,129-143`
**Severity:** Medium
**Category:** Code Duplication

**Issue:**
Coordinate parsing (lat/lng/radius) duplicated across three handlers (SearchOffers, GetDiscounts, GetStores). Exclude-stores parsing duplicated across two handlers.

**Impact:**
- Bug fixes must be applied in multiple places
- Inconsistency risk
- Maintenance burden

**Why It Matters:**
DRY violation. If validation is added to one handler, must add to all.

**Verified from CONCERNS.md:** Yes, already documented (line 19-23)

**Recommendation:**
```go
func parseLocationParams(r *http.Request) (float64, float64, int, error) {
    // Single implementation
}

func parseExcludeStores(r *http.Request) []string {
    // Single implementation
}
```

---

### M7. No Content-Type Validation on JSON Endpoints
**Files:** All handlers that decode JSON
**Severity:** Medium
**Category:** Input Validation

**Issue:**
Handlers decode JSON without checking `Content-Type: application/json` header.

**Impact:**
- Accepts any content type
- Could decode form data as JSON (unexpected behavior)
- No clear API contract

**Why It Matters:**
REST best practice: validate Content-Type matches expected format.

**Recommendation:**
```go
if r.Header.Get("Content-Type") != "application/json" {
    http.Error(w, `{"error":"content_type_must_be_json"}`, http.StatusUnsupportedMediaType)
    return
}
```

---

### M8. UpdateMemberStatus Allows Empty Updates
**File:** `backend/internal/api/handlers/household.go:133-136`
**Severity:** Medium
**Category:** Input Validation (good but could be stricter)

**Issue:**
```go
if req.IsEatingToday == nil && req.WantsLunchBox == nil {
    http.Error(w, `{"error":"no_fields_to_update"}`, http.StatusBadRequest)
    return
}
```

Handler correctly rejects empty updates, but allows updates where both fields are set to the same values as current state (no actual change).

**Impact:**
- Unnecessary database writes
- Transaction log pollution

**Why It Matters:**
Minor optimization. Not a bug but could be improved.

**Recommendation:**
- Service layer should check current values and no-op if unchanged
- Or: return 204 No Content instead of 200 OK for no-op updates

---

### M9. GetAll Recipes Returns Empty Array vs Null Inconsistency
**File:** `backend/internal/api/handlers/recipes.go:19-35`
**Severity:** Medium
**Category:** API Consistency

**Issue:**
```go
recipes, err := h.recipeService.GetAll(filter)
if err != nil {
    http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
    return
}
response := domain.RecipesResponse{Recipes: recipes}
```

No check if `recipes` is nil vs empty array. Response could be `{"recipes":null}` or `{"recipes":[]}` depending on storage layer behavior.

**Impact:**
- Inconsistent API responses
- Frontend must handle both nil and empty array
- JSON serialization difference

**Why It Matters:**
API consistency improves client reliability.

**Recommendation:**
```go
if recipes == nil {
    recipes = []domain.Recipe{}
}
response := domain.RecipesResponse{Recipes: recipes}
```

---

### M10. No Validation of Recipe Filter Query Parameters
**File:** `backend/internal/api/handlers/recipes.go:21-24`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
filter := &domain.RecipeFilter{
    Name: r.URL.Query().Get("name"),
    Tag:  r.URL.Query().Get("tag"),
}
```

No validation or sanitization of filter parameters:
- Name could be 10000 characters
- Tag could contain SQL wildcards or special characters
- No length limits

**Impact:**
- Potential SQL injection if storage layer doesn't sanitize
- Performance: extremely long filter strings
- Resource exhaustion

**Why It Matters:**
Filter parameters directly used in database queries.

**Recommendation:**
```go
name := r.URL.Query().Get("name")
tag := r.URL.Query().Get("tag")

if len(name) > 100 {
    http.Error(w, `{"error":"name_filter_too_long"}`, http.StatusBadRequest)
    return
}
if len(tag) > 50 {
    http.Error(w, `{"error":"tag_filter_too_long"}`, http.StatusBadRequest)
    return
}
```

---

### M11. CreateInvite Checks Role but Not Member Status
**File:** `backend/internal/api/handlers/household.go:49-54`
**Severity:** Medium
**Category:** Authorization Logic

**Issue:**
```go
role, err := h.householdService.GetMemberRole(householdID, userID)
if err != nil || role == "" || role == "guest" {
    http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
    return
}
```

Checks role but not if member has been removed or is inactive. A removed member's JWT (valid for 7 days) could still create invites.

**Impact:**
- Removed members can invite new users for up to 7 days after removal
- Stale tokens bypass removal

**Why It Matters:**
Token-based auth has no revocation. Membership changes don't invalidate tokens.

**Recommendation:**
- Check member status in database, not just role
- Return error if member record doesn't exist
- Consider implementing token revocation/blacklist
- Or: reduce token expiry to 1 day

---

### M12. Shopping List menuId Query Parameter Has No Household Verification
**File:** `backend/internal/api/handlers/shopping.go:20-24`
**Severity:** Medium (blocked by C2 - no auth)
**Category:** Authorization

**Issue:**
```go
menuID := r.URL.Query().Get("menuId")
if menuID == "" {
    http.Error(w, `{"error":"menu_id_required"}`, http.StatusBadRequest)
    return
}
```

Once authentication is added (C2), this handler still needs household verification. No check that menuID belongs to authenticated user's household.

**Impact:**
- IDOR vulnerability persists even after adding auth
- Access to other households' shopping lists

**Why It Matters:**
Adding `RequireAuth` wrapper (C2 fix) is insufficient. Handler must validate ownership.

**Recommendation:**
```go
householdID := middleware.GetHouseholdID(r.Context())
list, err := h.shoppingService.GetShoppingListForHousehold(menuID, householdID)
// Service layer verifies menu belongs to household
```

---

### M13. UpdateItem Accepts Any menuId for Path itemId
**File:** `backend/internal/api/handlers/shopping.go:51-55`
**Severity:** Medium
**Category:** Input Validation / Authorization

**Issue:**
```go
menuID := r.URL.Query().Get("menuId")
if menuID == "" {
    http.Error(w, `{"error":"menu_id_required"}`, http.StatusBadRequest)
    return
}
```

No validation that itemID (from path) actually belongs to the menu specified in query parameter. Could pass itemID from one menu and menuID from another.

**Impact:**
- Data inconsistency
- Checking items from wrong menus
- Confusion in shopping list state

**Why It Matters:**
API contract is unclear. Should itemID be globally unique or menu-scoped?

**Recommendation:**
- Service layer should validate itemID belongs to menuID
- Return 404 if item not found in specified menu
- Or: remove menuID query parameter and look up item's menu in service layer

---

## Low Severity Issues

### L1. Health Endpoint Returns Plaintext-ish JSON
**File:** `backend/internal/api/handlers/health.go:5-8`
**Severity:** Low
**Category:** Minor Inconsistency

**Issue:**
```go
func Health(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}
```

Doesn't set `Content-Type: application/json` header like other handlers.

**Impact:**
- Clients may misinterpret response
- Inconsistent with other endpoints

**Why It Matters:**
Minor polish issue. Not a bug but inconsistent API behavior.

**Recommendation:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"status":"ok"}`))
```

---

### L2. Health Endpoint Reveals Server is Operational
**File:** `backend/internal/api/handlers/health.go:5-8`
**Severity:** Low
**Category:** Information Disclosure (minor)

**Issue:**
Public `/health` endpoint confirms server is running. Could be used for reconnaissance.

**Impact:**
- Attackers know backend is online
- Target for DDoS
- Reveals Go backend (via HTTP headers)

**Why It Matters:**
Health checks are standard but provide reconnaissance info.

**Recommendation:**
- Consider requiring authentication for health endpoint
- Or: implement different endpoint for load balancer vs admin
- Current implementation is standard practice, so this is very low priority

---

### L3. GetMemberStatuses Returns Empty Members Array if Household Has No Members
**File:** `backend/internal/api/handlers/household.go:104-111`
**Severity:** Low
**Category:** API Design

**Issue:**
No validation that household actually has members. Returns empty array if no members found.

**Impact:**
- Empty household edge case
- Client can't distinguish between "no members" vs "household not found"

**Why It Matters:**
Every household should have at least one member (creator). Empty members array indicates data inconsistency.

**Recommendation:**
- Add validation in service layer: if members array is empty, log warning
- Consider returning error if household has no members (data corruption)

---

### L4. Login and Register Don't Trim Email/Password Whitespace
**Files:** `backend/internal/api/handlers/auth.go:18-24,40-46`
**Severity:** Low
**Category:** UX / Input Sanitization

**Issue:**
Email and password fields not trimmed. User can register with ` user@example.com ` (leading/trailing spaces).

**Impact:**
- User confusion: "correct password doesn't work" (due to space)
- Duplicate accounts with whitespace variants

**Why It Matters:**
Common UX issue. Forms often include accidental whitespace.

**Recommendation:**
```go
req.Email = strings.TrimSpace(req.Email)
req.Password = strings.TrimSpace(req.Password)
```
Note: Some argue passwords should not be trimmed (spaces could be intentional). Email should definitely be trimmed.

---

### L5. Recipe Create Returns 400 for All Errors
**File:** `backend/internal/api/handlers/recipes.go:70-73`
**Severity:** Low
**Category:** HTTP Status Code Accuracy

**Issue:**
```go
resp, err := h.recipeService.Create(req)
if err != nil {
    http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
    return
}
```

All service errors return 400 Bad Request. Internal errors (database failures) should be 500.

**Impact:**
- Incorrect HTTP semantics
- Client can't distinguish validation errors from server errors

**Why It Matters:**
HTTP status codes have meaning. 400 = client error, 500 = server error.

**Recommendation:**
- Use error string matching (like household.go) or sentinel errors
- Map validation errors to 400
- Map internal errors to 500

---

## Summary of CONCERNS.md Verification

**Verified Issues (documented in CONCERNS.md):**
- Empty JWT secret (C1) ✓
- Shopping list missing auth (C2) ✓
- Error message injection (H1) ✓
- No request body size limits (H2) ✓
- No rate limiting (H3) ✓
- CORS localhost-only (H4) ✓
- No email validation (H5) ✓
- Error string matching fragility (M1) ✓
- Duplicate coordinate parsing (M6) ✓
- Manual path splitting vs PathValue (M2) ✓

**New Issues Found (not in CONCERNS.md):**
- JWT algorithm not pinned (C3)
- No IDOR protection in menu/shopping (C4)
- Context key collision risk (H7)
- Inconsistent GetUserID/GetHouseholdID signatures (H8)
- No path parameter validation (M3)
- No GenerateMenuRequest validation (M4)
- Offers handlers silent fallback (M5)
- No Content-Type validation (M7)
- Various API consistency issues (M8-M13, L1-L5)

---

## Recommendations Priority

**Immediate (before production):**
1. Fix C1 (JWT secret validation) - blocks all security
2. Fix C2 (shopping list auth) - critical vulnerability
3. Fix H1 (error message injection) - breaks responses
4. Fix H2 (request body limits) - prevents DoS
5. Fix H3 (rate limiting) - prevents brute force
6. Fix H4 (CORS config) - blocks production deployment

**High Priority (within sprint):**
7. Fix C3 (JWT algorithm pinning) - defense in depth
8. Fix C4 (IDOR protection) - after C2 is fixed
9. Fix H5 (email validation) - prevents bad data
10. Fix M1 (sentinel errors) - improves maintainability

**Medium Priority (next sprint):**
11-23. Address M2-M13 issues systematically

**Low Priority (polish):**
24-30. Address L1-L5 as time permits

---

**Review Complete:** 2026-02-09
**Next Steps:** Prioritize fixes, create implementation tasks, add tests for each fix
