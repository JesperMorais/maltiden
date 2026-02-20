# Service & Storage Layer Security Review - Findings

**Review Date:** 2026-02-09
**Files Reviewed:** 19 (6 services + 6 storage + 6 migrations + main.go)
**Reviewer:** Code Review Agent

## Summary

| Severity | Count | Category Focus |
|----------|-------|----------------|
| **Critical** | 2 | Transaction safety, data integrity |
| **High** | 5 | SQL injection risk, data loss, constraint gaps |
| **Medium** | 15 | Validation, error handling, cascade issues |
| **Low** | 6 | Edge cases, consistency |
| **Total** | 28 | |

---

## Critical Severity Issues

### C1. Registration Flow Not Transactional - Orphaned Data Risk
**File:** `backend/internal/services/auth_service.go:46-83`
**Severity:** Critical
**Category:** Data Integrity

**Issue:**
```go
// Create household
if err := s.householdStorage.Create(household); err != nil {
    return nil, err
}

// Create user
if err := s.userStorage.Create(user); err != nil {
    return nil, err  // ← Household already created!
}

// Add user as household owner
if err := s.householdStorage.AddMember(member); err != nil {
    return nil, err  // ← Household AND user already created!
}
```

Registration creates three database records (household, user, member) without a transaction. If any step fails after the first succeeds, orphaned records are left in the database.

**Impact:**
- Orphaned households with no owner if `userStorage.Create` fails
- Orphaned households AND users if `AddMember` fails
- Database pollution with incomplete registrations
- No way to clean up partial failures automatically

**Why It Matters:**
Every failed registration after the household is created leaves garbage in the database. If bcrypt hashing succeeds but the user insert fails (e.g., database connection dropped), the household exists forever with no owner.

**Verified from CONCERNS.md:** Yes, already documented (line 144-149)

**Recommendation:**
Wrap the entire registration flow in a database transaction:
```go
tx, err := s.householdStorage.DB().Begin()
if err != nil {
    return nil, err
}
defer tx.Rollback()

// Create household (using tx)
// Create user (using tx)
// Add member (using tx)

if err := tx.Commit(); err != nil {
    return nil, err
}
```

Requires adding Tx variants of `Create` methods to storage layers.

---

### C2. Menu Generation With Zero Recipes Returns Success But Creates Broken Menu
**File:** `backend/internal/services/menu_service.go:31-33`
**Severity:** Critical
**Category:** Business Logic Bug

**Issue:**
```go
if len(recipes) == 0 {
    return nil, nil  // ← Returns success with nil result
}
```

When the recipes table is empty (freshly deployed system before seed data), menu generation returns `(nil, nil)` which handlers treat as success. The handler writes a null response or crashes trying to access the response.

**Impact:**
- Silent failure - user thinks menu was generated but gets null/error
- Frontend crashes trying to access nil menu
- No error message explaining why menu generation failed
- New deployments (before seed) have broken menu generation

**Why It Matters:**
This is a data-dependent bug. In a production system where all recipes get deleted (admin mistake, migration bug), menu generation silently fails for all users with no error message.

**Recommendation:**
```go
if len(recipes) == 0 {
    return nil, errors.New("no_recipes_available")
}
```

Handler can then return proper 400 Bad Request with clear error message: "Cannot generate menu - no recipes in system."

---

## High Severity Issues

### H1. Recipe Tag Filter Vulnerable to JSON Injection
**File:** `backend/internal/storage/sqlite/recipe_storage.go:28-31`
**Severity:** High
**Category:** SQL Injection (limited scope)

**Issue:**
```go
if filter != nil && filter.Tag != "" {
    query += ` AND tags LIKE ?`
    args = append(args, "%\""+filter.Tag+"\"%")
}
```

Tag filter searches for `%"tag"%` in JSON string. If the tag contains `"`, it can match unintended tags or break the LIKE pattern.

**Example Attack:**
- Query with tag: `pasta"`
- SQL becomes: `LIKE '%"pasta"%'`
- Matches: `["pasta"]` ✓ (intended)
- Also matches: `["pasta", "other"]` ✓ (intended)
- But also matches: `["not-pasta"]` if JSON is malformed with `"not-pasta"` somewhere

More problematic: tag `%` matches ALL recipes (wildcard injection).

**Impact:**
- Tag filter bypass: user can see recipes from other tags
- Performance: `%` wildcard queries are slow on large tables
- Not a full SQL injection (parameterized query protects against that) but LIKE pattern injection

**Why It Matters:**
Recipe filtering is a core feature. Broken filtering shows wrong recipes in search results.

**Recommendation:**
1. Escape special LIKE characters (`%`, `_`) in the tag before building the pattern
2. Or better: Use JSON1 extension for proper JSON querying:
```go
query += ` AND EXISTS (SELECT 1 FROM json_each(tags) WHERE value = ?)`
args = append(args, filter.Tag)
```

---

### H2. No Foreign Key Enforcement in SQLite
**File:** `backend/internal/storage/sqlite/db.go` (missing line)
**Severity:** High
**Category:** Data Integrity

**Issue:**
SQLite requires `PRAGMA foreign_keys = ON` to enforce foreign key constraints. The database initialization doesn't enable this pragma.

**Impact:**
- All `FOREIGN KEY` declarations in migrations are ignored
- Orphaned menu_days when menus are deleted (ON DELETE CASCADE doesn't work)
- Orphaned shopping_items when menus are deleted
- Orphaned household_members when users are deleted
- Database integrity entirely depends on application logic

**Why It Matters:**
The schema defines foreign keys with `ON DELETE CASCADE`, but SQLite ignores them by default. If a menu is deleted, menu_days remain in the database forever, accumulating over time.

**Recommendation:**
In `db.go`, immediately after opening the database:
```go
db, err := sql.Open("sqlite3", path)
if err != nil {
    return nil, err
}

// Enable foreign key constraints (required for SQLite)
if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
    return nil, err
}
```

This is REQUIRED for data integrity.

---

### H3. UpdateMemberStatus Builds SQL Query via String Concatenation
**File:** `backend/internal/storage/sqlite/household_storage.go:205`
**Severity:** High
**Category:** SQL Injection Risk (mitigated by limited inputs)

**Issue:**
```go
query := "UPDATE household_members SET " + strings.Join(updates, ", ") +
         " WHERE household_id = ? AND user_id = ?"
_, err := s.db.Exec(query, args...)
```

SQL query string is built via concatenation. The `updates` slice contains strings like `"is_eating_today = ?"` which are NOT user-controlled, but this pattern is fragile.

**Impact:**
- If this pattern is copied to other code with user-controlled field names, SQL injection occurs
- Code review difficulty: string concatenation SQL is a red flag pattern
- Maintenance risk: future developer might add user-controlled fields

**Why It Matters:**
While currently safe (field names are hardcoded in lines 184, 192), the pattern itself is dangerous and sets a bad precedent.

**Recommendation:**
Use explicit query variants instead:
```go
if isEatingToday != nil && wantsLunchBox != nil {
    query = "UPDATE household_members SET is_eating_today = ?, wants_lunch_box = ? WHERE household_id = ? AND user_id = ?"
    _, err = s.db.Exec(query, boolToInt(*isEatingToday), boolToInt(*wantsLunchBox), householdID, userID)
} else if isEatingToday != nil {
    query = "UPDATE household_members SET is_eating_today = ? WHERE household_id = ? AND user_id = ?"
    _, err = s.db.Exec(query, boolToInt(*isEatingToday), householdID, userID)
} else {
    // wantsLunchBox only
}
```

---

### H4. Menu Day Primary Key Collision Risk
**File:** `backend/internal/storage/sqlite/menu_storage.go:47`
**Severity:** High
**Category:** Data Corruption

**Issue:**
```go
_, err = tx.Exec(
    `INSERT INTO menu_days (id, menu_id, date, recipe_id, servings, skip)
     VALUES (?, ?, ?, ?, ?, ?)`,
    menu.ID+"_"+day.Date, menu.ID, day.Date, recipeID, day.Servings, skip,
)
```

Menu day ID is generated by concatenating `menuID + "_" + date`. If a menu is created for the same household on the same day (user regenerates menu), this can collide.

**Example:**
1. User generates menu → creates `menu_abc_2026-02-09`
2. User regenerates menu → creates new menu `menu_xyz` with day `menu_xyz_2026-02-09`
3. No collision YET (different menu IDs)
4. BUT: If menu generation logic changes to reuse menu IDs, collision occurs

**Impact:**
- Menu regeneration could fail with "UNIQUE constraint failed"
- Data loss if INSERT OR REPLACE is used later
- Fragile ID generation pattern

**Why It Matters:**
Primary key generation should be guaranteed unique. Relying on menuID uniqueness is fragile.

**Recommendation:**
Use UUID for menu day IDs:
```go
dayID := "md_" + uuid.New().String()
_, err = tx.Exec(
    `INSERT INTO menu_days (id, menu_id, date, recipe_id, servings, skip)
     VALUES (?, ?, ?, ?, ?, ?)`,
    dayID, menu.ID, day.Date, recipeID, day.Servings, skip,
)
```

---

### H5. Password Validation Only Checks Minimum Length
**File:** `backend/internal/services/auth_service.go:27-29`
**Severity:** High
**Category:** Security - Weak Validation

**Issue:**
```go
if len(req.Password) < 8 {
    return nil, errors.New("password must be at least 8 characters")
}
```

Password validation only checks minimum length. No maximum length check.

**Impact:**
- bcrypt has a 72-byte limit. Passwords longer than 72 bytes are silently truncated by bcrypt
- User sets password `"mypassword" + ("a" * 1000)` → bcrypt hashes first 72 bytes
- User tries to log in with full password → bcrypt hashes first 72 bytes → LOGIN SUCCEEDS
- BUT user might not remember they added trailing characters
- Worse: if user copies password with extra whitespace, bcrypt truncates differently

**Why It Matters:**
bcrypt's 72-byte limit is a security gotcha. Accepting unlimited password lengths leads to confusing behavior.

**Verified from CONCERNS.md:** Mentioned as bcrypt 72-byte limit (01-01-FINDINGS.md, line 110)

**Recommendation:**
```go
if len(req.Password) < 8 {
    return nil, errors.New("password_too_short")
}
if len(req.Password) > 72 {
    return nil, errors.New("password_too_long")
}
```

Also consider trimming whitespace (see L4 in previous findings).

---

## Medium Severity Issues

### M1. Login Timing Attack Possible
**File:** `backend/internal/services/auth_service.go:99-110`
**Severity:** Medium
**Category:** Security - Side Channel

**Issue:**
```go
user, err := s.userStorage.GetByEmail(req.Email)
if err != nil {
    return nil, err
}
if user == nil {
    return nil, errors.New("invalid credentials")  // ← Fast path
}

// Validate password
if !utils.CheckPassword(req.Password, user.PasswordHash) {
    return nil, errors.New("invalid credentials")  // ← Slow path (bcrypt)
}
```

When user doesn't exist, the function returns immediately. When user exists but password is wrong, bcrypt takes ~250ms to verify. This timing difference allows email enumeration.

**Attack:**
1. Try login with email `test@example.com` + wrong password
2. Measure response time
3. If < 50ms → email doesn't exist
4. If > 200ms → email exists, wrong password

**Impact:**
- Attacker can enumerate all registered emails
- Email list valuable for phishing attacks
- Combined with credential stuffing, increases attack surface

**Why It Matters:**
Best practice is constant-time login. Handler correctly returns same error message, but timing still leaks information.

**Verified from CONCERNS.md:** Partially mentioned (login timing - 01-01-FINDINGS.md H6)

**Recommendation:**
Always run bcrypt comparison, even for non-existent users:
```go
user, err := s.userStorage.GetByEmail(req.Email)
if err != nil {
    return nil, err
}

// Use dummy hash if user doesn't exist
passwordHash := "$2a$12$dummy.hash.that.fails.check"
if user != nil {
    passwordHash = user.PasswordHash
}

// Always check password (constant time)
if !utils.CheckPassword(req.Password, passwordHash) || user == nil {
    return nil, errors.New("invalid credentials")
}
```

---

### M2. No Validation on Recipe Name Length
**File:** `backend/internal/services/recipe_service.go:29-30`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
if req.Name == "" {
    return nil, errors.New("name_required")
}
// No max length check
```

Recipe name has no maximum length validation. User can submit 10,000-character recipe name.

**Impact:**
- Database pollution with huge text fields
- Frontend rendering issues (recipe name overflows UI)
- API response bloat
- Potential DoS via extremely long names

**Why It Matters:**
All text inputs should have reasonable max length limits. 200 characters is more than enough for a recipe name.

**Recommendation:**
```go
if req.Name == "" {
    return nil, errors.New("name_required")
}
if len(req.Name) > 200 {
    return nil, errors.New("name_too_long")
}
```

Apply similar limits to all string fields (emoji, tags, ingredient names, instructions).

---

### M3. No Validation on Recipe Servings Upper Bound
**File:** `backend/internal/services/recipe_service.go:32-34`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
if req.Servings <= 0 {
    return nil, errors.New("invalid_servings")
}
// No upper bound
```

Servings validation only checks positive value. User can create recipe with servings = 1,000,000.

**Impact:**
- Shopping list amount calculations overflow: `amount * (servings / recipeServings)`
- Example: recipe for 4 servings needs 500g meat. User generates menu with 1M servings → 125,000kg meat in shopping list
- Frontend crashes displaying huge numbers
- Floating point overflow in calculations

**Why It Matters:**
Unbounded numeric inputs are dangerous for calculations. 100 servings is already unreasonably high.

**Recommendation:**
```go
if req.Servings <= 0 || req.Servings > 100 {
    return nil, errors.New("invalid_servings")
}
```

---

### M4. No Validation on Ingredients Array Size
**File:** `backend/internal/services/recipe_service.go:35-37`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
if len(req.Ingredients) == 0 {
    return nil, errors.New("ingredients_required")
}
// No max length check
```

Ingredients array has no size limit. User can submit recipe with 10,000 ingredients.

**Impact:**
- Database JSON field bloat
- Shopping list generation becomes extremely slow (N+1 pattern scales with ingredient count)
- API response size explodes
- Frontend performance degrades

**Why It Matters:**
A recipe with 100+ ingredients is unrealistic. Most recipes have 5-15 ingredients.

**Recommendation:**
```go
if len(req.Ingredients) == 0 {
    return nil, errors.New("ingredients_required")
}
if len(req.Ingredients) > 50 {
    return nil, errors.New("too_many_ingredients")
}
```

Also validate each ingredient's name/unit/amount fields for length and range.

---

### M5. Shopping Service N+1 Query Pattern Confirmed
**File:** `backend/internal/services/shopping_service.go:92-100`
**Severity:** Medium
**Category:** Performance

**Issue:**
```go
for _, day := range menu.Days {
    if day.Skip || day.RecipeID == "" {
        continue
    }

    recipe, err := s.recipeStorage.GetByID(day.RecipeID)  // ← N queries
    if err != nil || recipe == nil {
        continue
    }
    // ...
}
```

Shopping list generation fetches each recipe individually in a loop. For a 7-day menu, this makes 7 separate SELECT queries.

**Impact:**
- 7 round trips to SQLite for a weekly menu
- 30 round trips for a monthly menu
- Scales linearly with menu length
- Unnecessary database load

**Why It Matters:**
While SQLite is fast for single queries, this is inefficient. CONCERNS.md already identified this (line 117-121).

**Verified from CONCERNS.md:** Yes, already documented (line 117-121)

**Recommendation:**
Collect all recipe IDs first, then fetch in one query:
```go
var recipeIDs []string
for _, day := range menu.Days {
    if !day.Skip && day.RecipeID != "" {
        recipeIDs = append(recipeIDs, day.RecipeID)
    }
}

// Batch fetch
recipes, err := s.recipeStorage.GetByIDs(recipeIDs)  // New method
```

Requires adding `GetByIDs` method to recipe storage with `WHERE id IN (?, ?, ?)` query.

---

### M6. Shopping Item ID Generation Uses MD5 (Not a Security Issue)
**File:** `backend/internal/services/shopping_service.go:166-169`
**Severity:** Medium (Low-Medium)
**Category:** Code Quality

**Issue:**
```go
func generateItemID(menuID, name, unit string) string {
    hash := md5.Sum([]byte(menuID + "_" + strings.ToLower(name) + "_" + unit))
    return fmt.Sprintf("item_%x", hash[:8])
}
```

MD5 is used for generating shopping item IDs. While MD5 is cryptographically broken, this is NOT a security issue (IDs are not secrets).

**Impact:**
- MD5 collision risk (though extremely low for this use case)
- Code review red flag (reviewers see `crypto/md5` and worry)
- Better alternatives exist

**Why It Matters:**
Using broken crypto functions, even for non-security purposes, is a code smell. Modern Go has better options.

**Recommendation:**
Use a hash function from `hash/fnv` (faster, non-crypto):
```go
import "hash/fnv"

func generateItemID(menuID, name, unit string) string {
    h := fnv.New64a()
    h.Write([]byte(menuID + "_" + strings.ToLower(name) + "_" + unit))
    return fmt.Sprintf("item_%x", h.Sum(nil)[:8])
}
```

Or just use UUIDs: `"item_" + uuid.New().String()`

---

### M7. Menu Service Doesn't Validate Days Count
**File:** `backend/internal/services/menu_service.go:36-39`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
days := req.Days
if days <= 0 {
    days = 5  // Default
}
// No upper bound check
```

Days count has no maximum. User can request a 10,000-day menu.

**Impact:**
- Generates 10,000 menu_day records in database
- Menu response JSON becomes multi-megabyte
- Frontend crashes trying to render 10,000 days
- Database bloat

**Why It Matters:**
A 365-day menu is already unrealistic. Most users want 5-7 days.

**Recommendation:**
```go
days := req.Days
if days <= 0 {
    days = 5
}
if days > 31 {
    return nil, errors.New("max_31_days")
}
```

---

### M8. Menu Service Doesn't Validate Servings Count
**File:** `backend/internal/services/menu_service.go:40-43`
**Severity:** Medium
**Category:** Input Validation

**Issue:**
```go
servings := req.Servings
if servings <= 0 {
    servings = 4
}
// No upper bound
```

Same issue as recipe servings (M3). User can request 1,000,000 servings per day.

**Impact:**
- Shopping list calculations overflow
- Massive ingredient amounts in shopping list
- Frontend rendering issues

**Recommendation:**
```go
servings := req.Servings
if servings <= 0 {
    servings = 4
}
if servings > 100 {
    return nil, errors.New("invalid_servings")
}
```

---

### M9. Menu Generation Random Recipe Selection Can Duplicate
**File:** `backend/internal/services/menu_service.go:68`
**Severity:** Medium
**Category:** Business Logic - Poor UX

**Issue:**
```go
recipe := recipes[rand.Intn(len(recipes))]
day.RecipeID = recipe.ID
```

Random recipe selection doesn't prevent duplicates. User can get the same recipe multiple days in a row.

**Impact:**
- Poor user experience: "Why do I have tacos 3 days in a row?"
- Menu feels less useful if it lacks variety
- Users have to manually regenerate multiple times

**Why It Matters:**
While not a bug per se, this makes the feature less useful. Users expect variety in their weekly menu.

**Recommendation:**
Track used recipes and avoid immediate repeats:
```go
usedRecipes := make(map[string]bool)
for i := 0; i < days; i++ {
    // Filter out recently used recipes
    availableRecipes := filterAvailable(recipes, usedRecipes)
    if len(availableRecipes) == 0 {
        // If all recipes used, reset
        usedRecipes = make(map[string]bool)
        availableRecipes = recipes
    }

    recipe := availableRecipes[rand.Intn(len(availableRecipes))]
    usedRecipes[recipe.ID] = true
    day.RecipeID = recipe.ID
}
```

Or use shuffle algorithm to guarantee no duplicates within menu length.

---

### M10. Recipe Deletion Leaves Orphaned Menu Days
**File:** `backend/internal/storage/sqlite/recipe_storage.go` (missing deletion method)
**Severity:** Medium
**Category:** Data Integrity / Missing Feature

**Issue:**
There is no `Delete` method in `RecipeStorage`. When deletion is added later, it will need to handle menu_days that reference the recipe.

**Impact:**
- Future deletion feature will break menu_days (foreign key violation if FK enabled)
- Or orphaned menu_days with null recipe_id (if FK not enabled - see H2)
- Shopping list generation breaks when recipe is missing

**Why It Matters:**
This will be a problem when recipe management is added. Need to decide cascade behavior NOW.

**Recommendation:**
Options for deletion behavior:
1. **Prevent deletion** if recipe is used in any active menu: `SELECT COUNT(*) FROM menu_days WHERE recipe_id = ?`
2. **Cascade delete** menu_days: Change FK to `ON DELETE CASCADE`
3. **Set null** in menu_days: Change FK to `ON DELETE SET NULL`, handle nulls in shopping service

Best option: (1) - prevent deletion of recipes in use. Add `DeleteRecipe` method that checks for usage first.

---

### M11. Invite Code Entropy is Sufficient But Not Documented
**File:** `backend/internal/services/household_service.go:206-214`
**Severity:** Medium (Low-Medium)
**Category:** Security - Good Implementation, Lacking Documentation

**Issue:**
```go
const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I to avoid confusion
code := make([]byte, 8)
```

Invite code generation uses 32-character charset (5 bits per character), 8 characters = 40 bits entropy. This is good, but entropy is not documented.

**Impact:**
- Future developers might reduce charset or length without realizing security implications
- No comment explaining why 8 characters is sufficient

**Why It Matters:**
Invite codes need sufficient entropy to prevent brute force guessing. 40 bits = 1 trillion combinations, which is adequate for 7-day expiry.

**Recommendation:**
Add comment documenting entropy:
```go
// generateInviteCode creates a random 8-character alphanumeric code.
// Entropy: 32-char charset ^ 8 chars = 2^40 ~= 1 trillion combinations
// With 7-day expiry, brute force guessing is impractical.
func generateInviteCode() (string, error) {
    const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 32 chars, no ambiguous 0/O/1/I
    code := make([]byte, 8)
    // ...
}
```

---

### M12. JoinHousehold Transaction Doesn't Lock Invite Code
**File:** `backend/internal/services/household_service.go:85-132`
**Severity:** Medium
**Category:** Race Condition

**Issue:**
```go
// Validate invite code before starting the transaction
invite, err := s.householdStorage.GetInviteByCode(req.Code)
// ... validation checks ...

// Begin transaction
tx, err := s.householdStorage.DB().Begin()
// ... add member, mark invite used ...
```

Invite code is read BEFORE transaction starts. Two users can read the same unused invite simultaneously, both pass validation, then both try to use it within transactions.

**Impact:**
- Race condition: two users could join using same invite code
- One transaction wins, other gets unique constraint violation on `household_members(household_id, user_id)`
- Not a security issue (both users get added to same household) but confusing UX

**Why It Matters:**
Invite codes should be single-use. Race condition allows double-use.

**Recommendation:**
Move invite validation inside transaction or use SELECT FOR UPDATE:
```go
tx, err := s.householdStorage.DB().Begin()
if err != nil {
    return nil, err
}
defer tx.Rollback()

// Read invite inside transaction with row lock
invite, err := s.householdStorage.GetInviteByCodeTx(tx, req.Code)
// ... rest of transaction
```

Requires `GetInviteByCodeTx` method that uses transaction.

---

### M13. Menu Day Date Format Not Validated
**File:** `backend/internal/services/menu_service.go:56`
**Severity:** Medium
**Category:** Data Quality

**Issue:**
```go
date := today.AddDate(0, 0, i).Format("2006-01-02")
```

Dates are generated correctly by the service, but if the API ever accepts user-provided dates, there's no validation that they're in `YYYY-MM-DD` format.

**Impact:**
- Future API changes could allow malformed dates
- Database contains unparseable date strings
- Shopping list date sorting breaks

**Why It Matters:**
Date format validation should exist at the API boundary to prevent future bugs.

**Recommendation:**
Add validation in handlers that accept dates (currently none, but future APIs might):
```go
func validateDate(dateStr string) error {
    _, err := time.Parse("2006-01-02", dateStr)
    return err
}
```

---

### M14. Bubble Sort Used Instead of sort.Slice (Confirmed)
**File:** `backend/internal/services/tjek_service.go:286-292, 432-438`
**Severity:** Medium (Low)
**Category:** Performance / Code Quality

**Issue:**
```go
// GetAvailableStores - lines 286-292
for i := 0; i < len(stores)-1; i++ {
    for j := i + 1; j < len(stores); j++ {
        if stores[j] < stores[i] {
            stores[i], stores[j] = stores[j], stores[i]
        }
    }
}

// GetTopDiscounts - lines 432-438
for i := 0; i < len(discountedOffers)-1; i++ {
    for j := i + 1; j < len(discountedOffers); j++ {
        if discountedOffers[j].discount > discountedOffers[i].discount {
            discountedOffers[i], discountedOffers[j] = discountedOffers[j], discountedOffers[i]
        }
    }
}
```

Manual O(n²) bubble sort used when `sort.Slice` is already imported in the same package (shopping_service.go:138).

**Impact:**
- Slow for large result sets (100+ stores, 100+ offers)
- Inconsistent codebase (sort.Slice used elsewhere)
- Code review red flag

**Why It Matters:**
CONCERNS.md already identified this (line 25-29). Using standard library is faster and more readable.

**Verified from CONCERNS.md:** Yes, already documented (line 25-29)

**Recommendation:**
```go
// GetAvailableStores
sort.Strings(stores)

// GetTopDiscounts
sort.Slice(discountedOffers, func(i, j int) bool {
    return discountedOffers[i].discount > discountedOffers[j].discount
})
```

---

### M15. Tjek Service Sequential Catalog Fetching (Confirmed)
**File:** `backend/internal/services/tjek_service.go:324-355, 398-429`
**Severity:** Medium
**Category:** Performance

**Issue:**
```go
for _, catalog := range catalogs {
    // ...
    offers, err := s.getCatalogOffers(catalog, store)  // ← Sequential HTTP calls
    if err != nil {
        continue
    }
    // ...
}
```

Offers are fetched from each catalog sequentially. With 20 grocery store catalogs, this makes 20 sequential HTTP requests.

**Impact:**
- Slow API response: 20 catalogs × 500ms per request = 10 seconds
- Poor user experience
- Unnecessary latency

**Why It Matters:**
CONCERNS.md already identified this (line 123-127). These requests are independent and can run in parallel.

**Verified from CONCERNS.md:** Yes, already documented (line 123-127)

**Recommendation:**
Use goroutines with WaitGroup or errgroup:
```go
var wg sync.WaitGroup
offersChan := make(chan []domain.TjekOffer, len(catalogs))

for _, catalog := range catalogs {
    wg.Add(1)
    go func(cat catalogResponse) {
        defer wg.Done()
        offers, err := s.getCatalogOffers(cat, storesByDealer[cat.DealerID])
        if err == nil {
            offersChan <- offers
        }
    }(catalog)
}

wg.Wait()
close(offersChan)

for offers := range offersChan {
    allOffers = append(allOffers, offers...)
}
```

Add rate limiting (max 5 concurrent requests) to avoid overwhelming Tjek API.

---

## Low Severity Issues

### L1. Migration File Paths Are Relative to Working Directory
**File:** `backend/internal/storage/sqlite/db.go:52-58`
**Severity:** Low
**Category:** Fragility

**Issue:**
```go
migrations := []struct {
    version int
    file    string
}{
    {1, "migrations/001_create_users.sql"},
    // ...
}

sqlBytes, err := os.ReadFile(m.file)  // ← Relative path
```

Migration files are read via relative paths. Server must be run from `backend/` directory.

**Impact:**
- Running server from wrong directory fails silently (migrations not found)
- Tests require `os.Chdir()` workarounds (see CONCERNS.md line 139-142)
- Deployment fragility

**Why It Matters:**
CONCERNS.md already identified this (line 137-142). Tests work around it with directory changes.

**Verified from CONCERNS.md:** Yes, already documented (line 137-142)

**Recommendation:**
Use `embed.FS` to embed migrations in binary:
```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Read from embedded FS
data, err := migrationsFS.ReadFile("migrations/001_create_users.sql")
```

This makes migrations part of the compiled binary, eliminating path issues.

---

### L2. No WAL Mode for SQLite Concurrency
**File:** `backend/internal/storage/sqlite/db.go` (missing line)
**Severity:** Low
**Category:** Performance

**Issue:**
SQLite defaults to DELETE journal mode, which blocks readers during writes. WAL mode allows concurrent reads.

**Impact:**
- Readers block during writes
- Lower concurrency for read-heavy workloads
- Not critical for low-traffic family app

**Why It Matters:**
CONCERNS.md mentions WAL mode (line 161-163). WAL mode is a simple optimization for SQLite.

**Verified from CONCERNS.md:** Mentioned (line 161-163)

**Recommendation:**
Enable WAL mode after opening database:
```go
if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
    return nil, err
}
```

---

### L3. No Graceful Shutdown in main.go
**File:** `backend/cmd/server/main.go:35`
**Severity:** Low
**Category:** Reliability

**Issue:**
```go
log.Fatal(http.ListenAndServe(":"+port, router))
```

Server has no graceful shutdown handling. In-flight requests are killed when server stops.

**Impact:**
- Mid-request database writes can be interrupted
- SQLite transactions may leave database in inconsistent state
- Docker container stop has 10-second SIGKILL timeout

**Why It Matters:**
Production servers should handle SIGTERM gracefully to finish in-flight requests.

**Recommendation:**
```go
srv := &http.Server{
    Addr:    ":" + port,
    Handler: router,
}

// Graceful shutdown on SIGTERM/SIGINT
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}()

log.Fatal(srv.ListenAndServe())
```

---

### L4. Database Connection Not Configured for Performance
**File:** `backend/internal/storage/sqlite/db.go:18-21`
**Severity:** Low
**Category:** Performance

**Issue:**
```go
db, err := sql.Open("sqlite3", path)
if err != nil {
    return nil, err
}
```

SQLite connection opened with default settings. No connection pool configuration.

**Impact:**
- Default max open connections might be too low for concurrent requests
- No connection lifetime settings
- Not critical for SQLite (single file) but Go's sql.DB manages a pool

**Recommendation:**
Configure connection pool after opening:
```go
db.SetMaxOpenConns(25)       // Limit concurrent connections
db.SetMaxIdleConns(5)        // Keep idle connections
db.SetConnMaxLifetime(5 * time.Minute)
```

For SQLite, `MaxOpenConns` should be low (1-25) since SQLite is single-writer.

---

### L5. Shopping Storage Uses Ingredient Name as Part of Primary Key
**File:** `backend/internal/storage/sqlite/shopping_storage.go:40-43`
**Severity:** Low
**Category:** Schema Design

**Issue:**
```go
INSERT INTO shopping_items (id, menu_id, ingredient_name, checked)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET checked = ?
```

The `ingredient_name` column is stored but `id` is the primary key (generated hash). This creates redundancy.

**Impact:**
- Data duplication: ingredient name stored twice (in hash and in column)
- Potential inconsistency if hash generation changes
- Not a bug, just suboptimal schema

**Why It Matters:**
The `ingredient_name` column is never queried directly (always accessed via `id` hash). It serves only as metadata.

**Recommendation:**
Remove `ingredient_name` column from schema (migration 005) or use it as part of composite key:
```sql
CREATE TABLE shopping_items (
    menu_id TEXT NOT NULL,
    ingredient_name TEXT NOT NULL,
    checked INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (menu_id, ingredient_name),
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);
```

This is cleaner and avoids hash generation entirely.

---

### L6. Recipe Tag Filter Returns Empty Array Instead of Nil Inconsistency
**File:** `backend/internal/storage/sqlite/recipe_storage.go:63-65`
**Severity:** Low
**Category:** API Consistency

**Issue:**
```go
if recipes == nil {
    recipes = []domain.RecipeSummary{}
}
```

Query result is explicitly converted from nil to empty array. This is good, but other storage methods (menu, household) don't do this consistently.

**Impact:**
- API inconsistency: some endpoints return `null`, others return `[]`
- Frontend must handle both cases

**Why It Matters:**
Consistent API responses are easier for clients. Previous findings (01-01 M9) noted this.

**Recommendation:**
Apply this pattern to all storage methods that return slices:
- MenuStorage.GetCurrentByHousehold (line 102-104) already does this ✓
- HouseholdStorage.GetMemberStatuses (line 149-151) already does this ✓
- Consistent!

Actually, this is already consistent across the codebase. Mark as "non-issue" but document the pattern.

---

## Migration Schema Review

### Migration 001 (users table): ✓ Good
- Indexes on email (lookup) and household_id (joins)
- UNIQUE constraint on email
- No issues found

### Migration 002 (households & members): ✓ Good with FK concern
- Foreign keys defined with CASCADE
- UNIQUE(household_id, user_id) prevents duplicate memberships
- CHECK constraint on role is correct
- **Issue:** Foreign keys not enforced (see H2)

### Migration 003 (recipes): ✓ Good
- Index on name (for filtering)
- JSON columns for arrays (appropriate for SQLite)
- No issues found

### Migration 004 (seed recipes): ✓ Good
- INSERT OR IGNORE prevents duplicates on re-run
- Valid JSON in ingredients/instructions
- No issues found

### Migration 005 (menus & shopping): ⚠ FK cascade missing verification
- Foreign keys defined with ON DELETE CASCADE
- **Issue:** Depends on PRAGMA foreign_keys = ON (see H2)
- shopping_items schema (see L5 - redundant ingredient_name column)

### Migration 006 (invites & member status): ✓ Good
- Invite code UNIQUE constraint correct
- ON DELETE SET NULL for used_by (correct - preserve invite history)
- ALTER TABLE for member status columns (correct pattern)
- No issues found

---

## Summary of CONCERNS.md Verification

**Verified Issues (documented in CONCERNS.md):**
- Registration not transactional (C1) ✓
- N+1 query pattern in shopping service (M5) ✓
- Bubble sort in tjek service (M14) ✓
- Sequential catalog fetching (M15) ✓
- Migration relative paths (L1) ✓
- WAL mode not enabled (L2) ✓

**New Issues Found (not in CONCERNS.md):**
- Menu generation with zero recipes (C2)
- Recipe tag filter LIKE injection (H1)
- Foreign keys not enforced (H2)
- SQL string concatenation in UpdateMemberStatus (H3)
- Menu day PK collision risk (H4)
- Password max length validation (H5)
- Login timing attack (M1)
- All input validation gaps (M2-M4, M7-M8)
- Menu random selection duplicates (M9)
- Recipe deletion cascade handling (M10)
- Invite code entropy documentation (M11)
- JoinHousehold race condition (M12)
- Date format validation (M13)

---

## Recommendations Priority

**Immediate (Critical - before any production use):**
1. C1: Wrap registration in transaction - prevents data corruption
2. C2: Return error on zero recipes - prevents silent failures
3. H2: Enable PRAGMA foreign_keys = ON - required for data integrity

**High Priority (Before feature development continues):**
4. H1: Fix recipe tag filter injection
5. H3: Refactor UpdateMemberStatus string concatenation
6. H4: Use UUIDs for menu day IDs
7. H5: Add password max length validation

**Medium Priority (Within sprint):**
8. M1: Fix login timing attack
9. M2-M4, M7-M8: Add all input validation (length limits, bounds)
10. M5: Batch fetch recipes in shopping service
11. M9: Prevent duplicate recipes in menu generation
12. M10: Decide recipe deletion cascade behavior
13. M14-M15: Fix Tjek service performance (parallel fetching, sort.Slice)

**Low Priority (Polish):**
14. M11: Document invite code entropy
15. M12: Fix invite code race condition
16. L1-L5: Address low-severity issues as time permits

---

**Review Complete:** 2026-02-09
**Next Steps:**
1. Fix critical issues (C1, C2, H2) immediately
2. Create tasks for high-priority validation gaps
3. Plan performance improvements (M5, M15) for next sprint
4. Address low-priority items during code cleanup sessions
