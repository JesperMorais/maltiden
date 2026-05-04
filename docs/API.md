# Måltiden API-kontrakt

Base URL: `http://localhost:8080` (dev), `https://api.maltiden.se` (prod)

## Recent Changes (PR #177)

**New Endpoints:**
- `POST /auth/forgot-password` — Request a password reset email (public, rate-limited ~3/hour per IP)
- `POST /auth/reset-password` — Consume a reset token and set a new password (public, rate-limited ~3/hour per IP)

**New Frontend Routes:**
- `/forgot-password` — ForgotPasswordView
- `/reset-password` — ResetPasswordView (receives `?token=` from email link)

**Security:**
- Forgot-password always returns `200 { "ok": true }` regardless of whether the email is registered, preventing account enumeration.
- On successful reset, `token_version` is incremented — all existing JWTs for that user are immediately invalidated.

**New environment variables (backend):**
- `RESEND_API_KEY` — If set, reset emails are sent via Resend; otherwise the link is logged to stdout.
- `EMAIL_FROM` — Sender address (default: `Måltiden <no-reply@maltiden.app>`).

**No breaking changes** — purely additive.

---

## Recent Changes (PR #163)

**New Shopping List Endpoints:**
- `POST /shopping-list/items` — Add a custom item to the shopping list (auth required)
- `DELETE /shopping-list/items/:id` — Remove a custom item from the shopping list (auth required)

**Updated Response Shape:**
- `ShoppingItem` (returned by `GET /shopping-list` and `POST /shopping-list/items`) now includes `isCustom: boolean` — `true` for user-added items, `false` for recipe-generated items.

**No breaking changes** — purely additive.

---

## Recent Changes (PR #99)

**Auth client improvements (frontend only — no backend changes):**

- **Default request timeout reduced:** 10,000 ms → **5,000 ms** for all Axios requests (parse/parse-and-save endpoints keep their own 60 s override; offers endpoints keep 10–30 s).
- **401 redirect logic hardened to prevent loops:**
  - `/auth/*` endpoints (`/auth/login`, `/auth/register`) are now fully excluded from 401 redirect handling — a wrong-password 401 no longer clears the token or redirects.
  - Redirect is also suppressed when the user is already on `/register` (previously only `/login` was excluded).
  - A **10-second grace period** is enforced after a successful login/register (`markAuthSuccess()`). If a 401 arrives within that window the token is kept and no redirect happens, preventing a race condition where a fast in-flight request (e.g. `GET /households/me`) 401s right after JWT issuance.
- **`skipAuthRedirect` per-request option:** Individual Axios requests can set `{ skipAuthRedirect: true }` in their config to suppress the redirect entirely (used for background probes).

**No endpoint contract changes** — request/response shapes, HTTP methods, and URL paths are unchanged.

---

## Recent Changes (PR #85)

**New Endpoints:**
- `POST /feedback` — Submit user feedback (auth required, rate-limited to 5/hour per user)

**Behavioral Changes:**
- `GET /recipes` — Now uses optional auth (`OptionalAuth`). Authenticated requests are scoped to return only the household's own recipes plus global (seed) recipes. Unauthenticated requests return all recipes without scoping.

**Response shape changes:**
- `MenuResponseDay` (returned by `POST /menus/generate`, `PUT /menus/current`, `GET /menus/current`) now includes optional `recipeName` and `emoji` fields in each day object.
- Full `Recipe` object (returned by `GET /recipes/{id}`) now includes optional `householdId` field (omitted for global/seed recipes). The summary list response from `GET /recipes` does **not** include `householdId`.

**Rate limiting additions:**
- `POST /households/join` — 3 req/sec, burst 5 (brute-force protection on invite codes)
- `POST /recipes/parse` — 2 req/sec, burst 5 (AI API credit protection)
- `POST /recipes/parse-and-save` — 2 req/sec, burst 5 (AI API credit protection)

**Auth / security:**
- `RequireAuth` middleware now validates token version against the database on every request. Tokens invalidated server-side (e.g. after password change) are rejected with `401 unauthorized`.
- On `401` responses the Axios client now sets `sessionStorage.session_expired = 'true'` before redirecting to `/login`.

---

## Recent Changes (PR #82)

**New API Endpoints:**
- `PUT /recipes/{id}` - Update an existing recipe (auth required)
- `DELETE /recipes/{id}` - Delete a recipe (auth required)
- `PUT /menus/current` - Save exact recipe-day selections to the current active menu (auth required)

**Breaking / Additive Changes:**
- `emoji` field added to `CreateRecipeRequest`, `Recipe`, and `RecipeSummary` (optional, omitted when empty)
- Auth client now sets `sessionStorage.session_expired = 'true'` before redirecting to `/login` on 401

**Documentation fixes:**
- Corrected `POST /menus/generate` and `GET /menus/current` to show auth requirement
- Corrected `POST /recipes/parse` and `POST /recipes/parse-and-save` to show auth requirement
- Added missing auth requirement note to Shopping List section
- Added missing error cases: `household_not_found`, `code_required`, `already_member`, `no_fields_to_update` per endpoint and in error table
- Fixed `DELETE /households/members/:id` to document all three error cases (`cannot_remove`, `forbidden`, `not_found`)
- Added error cases (`invalid_days`, `invalid_servings`, `no_recipes_available`) to `POST /menus/generate`
- Fixed skip-day examples to include `"servings": 0`

## Changes in PR #35

**Recipe Navigation Consolidation:**
- Added new unified `/recipes` route with tabbed interface ("Mina recept" and "Lägg till")
- Added redirect from legacy `/recipes/parse` route to `/recipes`

**New API Endpoints:**
- `POST /recipes/parse` - Parse unstructured recipe text into structured data using AI
- `POST /recipes/parse-and-save` - Parse and immediately save recipe to database
- Both endpoints support 60-second timeout for AI processing
- Both endpoints accept optional `source` URL parameter

## Auth

### POST /auth/register
```json
// Request
{
  "email": "anna@example.com",
  "password": "minst8tecken",
  "name": "Anna"
}

// Response 201
{
  "token": "eyJhbG...",
  "user": {
    "id": "usr_abc123",
    "email": "anna@example.com",
    "name": "Anna",
    "householdId": "hh_xyz789"
  }
}

// Error 400
{ "error": "email_taken" }
```

### POST /auth/login
```json
// Request
{ "email": "anna@example.com", "password": "minst8tecken" }

// Response 200
{ "token": "eyJhbG...", "user": { ... } }

// Error 401
{ "error": "invalid_credentials" }
```

### POST /auth/forgot-password
Request a password reset email. **Always returns 200** regardless of whether the email
is registered, to avoid leaking account existence. Rate-limited to ~3 requests/hour
per IP.

When a matching account is found, the backend generates a 32-byte hex token (valid for
1 hour) and emails a reset link. By default the email is logged to stdout (operators
can grab the link from logs); if `RESEND_API_KEY` is set, the email is sent via Resend.
```json
// Request
{ "email": "anna@example.com" }

// Response 200 (always — even on unknown email or invalid format)
{ "ok": true }
```

### POST /auth/reset-password
Consume a reset token and set a new password. On success the user's `token_version`
is incremented, invalidating all existing JWTs for that user. Rate-limited to
~3 requests/hour per IP.
```json
// Request
{ "token": "abc123…", "newPassword": "Newpassword123" }

// Response 200
{ "ok": true }

// Error 400 — token does not exist
{ "error": "invalid_reset_token" }

// Error 400 — token has expired (>1h since creation)
{ "error": "expired_reset_token" }

// Error 400 — token has already been used
{ "error": "used_reset_token" }

// Error 400 — new password is shorter than 8 characters
{ "error": "weak_password" }
```

---

## Household

**Auth required:** `Authorization: Bearer <token>`

### GET /households/me
```json
// Response 200
{
  "id": "hh_xyz789",
  "name": "Familjen Svensson",
  "inviteCode": "ABC123",
  "members": [
    { "id": "usr_abc123", "name": "Anna", "role": "owner" },
    { "id": "usr_def456", "name": "Erik", "role": "member" },
    { "id": "usr_ghi789", "name": "Lisa", "role": "guest" }
  ]
}

// Roles: "owner" (full access + can delete household)
//        "member" (full access)
//        "guest" (view only)

// Error 404
{ "error": "household_not_found" }
```

### PATCH /households/me
Rename the household. Requires `owner` or `member` role — guests receive 403.
```json
// Request
{ "name": "Familjen Johansson" }

// Response 200
{ "ok": true }

// Error 400 (name is empty)
{ "error": "household_name_required" }

// Error 400 (name exceeds 100 characters)
{ "error": "household_name_too_long" }

// Error 403 (caller is a guest)
{ "error": "forbidden" }

// Error 404 (household not found)
{ "error": "not_found" }
```

### POST /households/invite
Only members and owners may create invite codes. Guests receive 403.
```json
// Request
{}

// Response 201
{
  "code": "ABC123",
  "expiresAt": "2025-01-20T12:00:00Z"
}

// Error 403 (caller is a guest — only members and owners can invite)
{ "error": "forbidden" }
```

### POST /households/join
**Rate limited:** 3 req/sec, burst 5 (brute-force protection on invite codes). Returns `429 Too Many Requests` when exceeded.
```json
// Request
{ "code": "ABC123" }

// Response 200
{ "householdId": "hh_xyz789" }

// Error 400 (code field missing)
{ "error": "code_required" }

// Error 400 (code invalid or expired)
{ "error": "invalid_code" }

// Error 409 (user is already a member of a household)
{ "error": "already_member" }
```

### GET /households/members/status
```json
// Response 200
{
  "members": [
    {
      "id": "usr_abc123",
      "isEatingToday": true,
      "wantsLunchBox": false
    },
    {
      "id": "usr_def456",
      "isEatingToday": true,
      "wantsLunchBox": true
    }
  ]
}
```

### PATCH /households/members/:id/status
```json
// Request (partial update — at least one field required)
{ "isEatingToday": false }
// or
{ "wantsLunchBox": true }
// or both
{ "isEatingToday": true, "wantsLunchBox": true }

// Response 200
{ "ok": true }

// Error 400 (neither field provided)
{ "error": "no_fields_to_update" }

// Error 404 (member not found)
{ "error": "not_found" }
```

### DELETE /households/members/:id
```json
// Response 200
{ "ok": true }

// Error 403 (target is the owner, or caller is trying to remove themselves)
{ "error": "cannot_remove" }

// Error 403 (caller lacks permission to remove members)
{ "error": "forbidden" }

// Error 404 (member not found in household)
{ "error": "not_found" }
```

---

## Recipes

### GET /recipes
**Auth optional.** Provide `Authorization: Bearer <token>` to scope results to the household's own recipes plus global (seed) recipes. Unauthenticated requests return all recipes without household scoping.
```json
// Query parameters (all optional):
// ?name=köttfärs    - Filter by recipe name (partial match)
// ?tag=vardag       - Filter by tag

// Response 200
{
  "recipes": [
    {
      "id": "rec_001",
      "name": "Köttfärssås",
      "servings": 4,
      "emoji": "🍝",            // optional — omitted when empty
      "tags": ["vardag", "barn"]
    }
  ]
}
```

### GET /recipes/:id
```json
// Response 200
{
  "id": "rec_001",
  "name": "Köttfärssås",
  "servings": 4,
  "emoji": "🍝",            // optional — omitted when empty
  "householdId": "hh_xyz",  // optional — present for household-specific recipes
  "ingredients": [
    { "name": "Köttfärs", "amount": 400, "unit": "g" },
    { "name": "Krossade tomater", "amount": 400, "unit": "g" }
  ],
  "instructions": [
    "Bryn köttfärsen",
    "Tillsätt tomater",
    "Låt sjuda 20 min"
  ],
  "tags": ["vardag", "barn"],
  "createdAt": "2026-02-01T10:00:00Z"
}
```

### POST /recipes
**Auth required.**
```json
// Request
{
  "name": "Köttfärssås",
  "servings": 4,
  "emoji": "🍝",            // optional
  "ingredients": [
    { "name": "Köttfärs", "amount": 400, "unit": "g" }
  ],
  "instructions": ["Bryn köttfärsen", "Tillsätt tomater"],
  "tags": ["vardag"]
}

// Response 201
{ "id": "rec_001" }

// Errors 400
{ "error": "name_required" }
{ "error": "invalid_servings" }
{ "error": "ingredients_required" }
{ "error": "instructions_required" }
{ "error": "name_too_long" }
{ "error": "too_many_ingredients" }
```

### PUT /recipes/{id}
Update an existing recipe. **Auth required.**
```json
// Request — same shape as POST /recipes
{
  "name": "Köttfärssås med lök",
  "servings": 4,
  "emoji": "🍝",            // optional
  "ingredients": [
    { "name": "Köttfärs", "amount": 500, "unit": "g" },
    { "name": "Lök", "amount": 1, "unit": "st" }
  ],
  "instructions": ["Hacka löken", "Bryn köttfärsen", "Tillsätt tomater"],
  "tags": ["vardag"]
}

// Response 200 — full recipe object
{
  "id": "rec_001",
  "name": "Köttfärssås med lök",
  "servings": 4,
  "emoji": "🍝",
  "ingredients": [...],
  "instructions": [...],
  "tags": ["vardag"],
  "createdAt": "2026-02-01T10:00:00Z"
}

// Error 404
{ "error": "not_found" }

// Errors 400
{ "error": "name_required" }
{ "error": "invalid_servings" }
{ "error": "ingredients_required" }
{ "error": "instructions_required" }
{ "error": "name_too_long" }
{ "error": "too_many_ingredients" }
```

### DELETE /recipes/{id}
Delete a recipe. **Auth required.** Only the owning household can delete a recipe; seed/global recipes cannot be deleted by anyone.
```json
// Response 204 — no body

// Error 403 (seed recipe, or recipe belongs to a different household)
{ "error": "forbidden" }

// Error 404
{ "error": "not_found" }
```

### POST /recipes/parse
Parse unstructured recipe text into structured data using AI. **Auth required.**
**Rate limited:** 2 req/sec, burst 5 (AI API credit protection).
```json
// Request
{
  "rawText": "Köttfärssås för 4 personer\n\nIngredienser:\n400g köttfärs\n400g krossade tomater\n...",
  "source": "https://optional-recipe-url.com"  // optional
}

// Response 200
{
  "recipe": {
    "name": "Köttfärssås",
    "servings": 4,
    "ingredients": [
      { "name": "Köttfärs", "amount": 400, "unit": "g" },
      { "name": "Krossade tomater", "amount": 400, "unit": "g" }
    ],
    "instructions": [
      "Bryn köttfärsen",
      "Tillsätt tomater",
      "Låt sjuda 20 min"
    ],
    "tags": ["vardag", "barn"],
    "emoji": "🍝"  // optional
  },
  "confidence": 0.95,
  "warnings": ["Could not parse exact cooking time"],  // optional
  "rawText": "Köttfärssås för 4 personer\n\n..."  // original input
}

// Timeout: 60 seconds
// Error 400
{ "error": "invalid_input", "message": "Text too short or empty" }
```

### POST /recipes/parse-and-save
Parse recipe text and immediately save it to the database. **Auth required.**
**Rate limited:** 2 req/sec, burst 5 (AI API credit protection).
```json
// Request
{
  "rawText": "Köttfärssås för 4 personer\n\nIngredienser:\n400g köttfärs\n400g krossade tomater\n...",
  "source": "https://optional-recipe-url.com"  // optional
}

// Response 201
{
  "id": "rec_123",  // newly created recipe ID
  "recipe": {
    "name": "Köttfärssås",
    "servings": 4,
    "ingredients": [...],
    "instructions": [...],
    "tags": ["vardag", "barn"],
    "emoji": "🍝"
  },
  "confidence": 0.95,
  "warnings": [],
  "rawText": "Köttfärssås för 4 personer\n\n..."
}

// Timeout: 60 seconds
// Error 400
{ "error": "invalid_input" }
```

---

## Menu

**Auth required:** `Authorization: Bearer <token>`

### POST /menus/generate
```json
// Request
{
  "days": 5,
  "skipDays": ["2025-01-22"],
  "servings": 4,
  "extraPortions": { "2025-01-23": 2 }
}

// Response 201
{
  "id": "menu_001",
  "days": [
    { "date": "2025-01-20", "recipeId": "rec_001", "recipeName": "Köttfärssås", "emoji": "🍝", "servings": 4 },
    { "date": "2025-01-21", "recipeId": "rec_002", "recipeName": "Laxpasta", "emoji": "🐟", "servings": 4 },
    { "date": "2025-01-22", "skip": true, "servings": 0 },
    { "date": "2025-01-23", "recipeId": "rec_003", "recipeName": "Kycklinggryta", "emoji": "🍗", "servings": 6 }
  ]
}
// Note: recipeName and emoji are optional — omitted for skip days and when not set on the recipe.

// Error 400
{ "error": "invalid_days" }
{ "error": "invalid_servings" }
{ "error": "no_recipes_available" }
```

### PUT /menus/current
Save exact recipe-day selections to the current active menu. **Auth required.**

Use this to replace the generated menu's day assignments without regenerating from scratch.

```json
// Request
{
  "days": [
    { "date": "2026-02-24", "recipeId": "rec_001", "servings": 4 },
    { "date": "2026-02-25", "recipeId": "rec_002", "servings": 4 },
    { "date": "2026-02-26", "skip": true, "servings": 0 },
    { "date": "2026-02-27", "recipeId": "rec_003", "servings": 6 }
  ]
}

// Response 200 — updated menu
{
  "id": "menu_001",
  "days": [
    { "date": "2026-02-24", "recipeId": "rec_001", "recipeName": "Köttfärssås", "emoji": "🍝", "servings": 4 },
    { "date": "2026-02-25", "recipeId": "rec_002", "recipeName": "Laxpasta", "emoji": "🐟", "servings": 4 },
    { "date": "2026-02-26", "skip": true, "servings": 0 },
    { "date": "2026-02-27", "recipeId": "rec_003", "recipeName": "Kycklinggryta", "servings": 6 }
  ]
}
// Note: recipeName and emoji are optional — omitted for skip days and when not set on the recipe.

// Error 404
{ "error": "no_active_menu" }

// Error 400
{ "error": "invalid_days" }
```

### GET /menus/current
**Auth required.**
```json
// Response 200
{
  "id": "menu_001",
  "days": [
    { "date": "2026-02-24", "recipeId": "rec_001", "recipeName": "Köttfärssås", "emoji": "🍝", "servings": 4 },
    { "date": "2026-02-25", "skip": true, "servings": 0 }
  ]
}
// Note: recipeName and emoji are optional — omitted for skip days and when not set on the recipe.

// Response 404 (ingen aktiv meny)
{ "error": "no_active_menu" }
```

---

## Shopping List

**Auth required:** `Authorization: Bearer <token>`

### GET /shopping-list
```json
// Query: ?menuId=menu_001   — REQUIRED

// Response 200
{
  "menuId": "menu_001",
  "categories": [
    {
      "name": "Kött & Fisk",
      "items": [
        { "id": "item_001", "name": "Köttfärs", "amount": 800, "unit": "g", "checked": false, "isCustom": false }
      ]
    },
    {
      "name": "Mejeri",
      "items": [
        { "id": "item_002", "name": "Grädde", "amount": 2, "unit": "dl", "checked": false, "isCustom": false },
        { "id": "item_003", "name": "Parmesan", "amount": 1, "unit": "st", "checked": false, "isCustom": true }
      ]
    }
  ]
}
// isCustom: false — item generated from a recipe
// isCustom: true  — item added manually by the user

// Error 403 (menu belongs to a different household)
{ "error": "forbidden" }

// Error 404
{ "error": "menu_not_found" }
```

### PATCH /shopping-list/items/:id
```json
// Query: ?menuId=menu_001   — REQUIRED
// Body
{ "checked": true }

// Response 200
{ "ok": true }

// Error 403 (menu belongs to a different household)
{ "error": "forbidden" }

// Error 404
{ "error": "menu_not_found" }
```

### POST /shopping-list/items
Add a custom item to the shopping list. **Auth required.**
```json
// Query: ?menuId=menu_001   — REQUIRED

// Request
{
  "name": "Parmesan",    // required, max 200 chars
  "unit": "st",          // optional, max 20 chars (default: "st")
  "amount": 1            // optional, must be finite and ≤ 100000 (default: 1)
}

// Response 201
{
  "id": "citem_550e8400-e29b-41d4-a716-446655440000",
  "name": "Parmesan",
  "amount": 1,
  "unit": "st",
  "checked": false,
  "isCustom": true
}

// Error 400 — name field missing or empty
{ "error": "name_required" }

// Error 400 — name exceeds 200 characters
{ "error": "name_too_long" }

// Error 400 — unit exceeds 20 characters
{ "error": "unit_too_long" }

// Error 400 — amount is NaN or Infinity
{ "error": "invalid_amount" }

// Error 400 — amount exceeds 100000
{ "error": "amount_too_large" }

// Error 401 — missing or invalid auth token
{ "error": "unauthorized" }

// Error 403 — menu belongs to a different household (cross-tenant access)
{ "error": "forbidden" }

// Error 404 — menu does not exist
{ "error": "menu_not_found" }
```

### DELETE /shopping-list/items/:id
Remove a custom item from the shopping list. **Auth required.** Only custom items (IDs prefixed `citem_`) can be deleted via this endpoint; recipe-generated items are not deletable.
```json
// Response 200
{ "ok": true }

// Error 401 — missing or invalid auth token
{ "error": "unauthorized" }

// Error 404 — item not found, OR item belongs to a different household
// (cross-tenant probes return 404, not 403, to avoid leaking item existence)
{ "error": "not_found" }
```

---

## Feedback

**Auth required:** `Authorization: Bearer <token>`

### POST /feedback
Submit user feedback. Rate-limited to 5 submissions per user per hour.
```json
// Request
{
  "mood": "good",                     // required — "good" | "okay" | "bad"
  "categories": ["recipes", "menu"],  // optional — valid values: "recipes" | "menu" | "shopping" | "design" | "other"
  "comment": "Jättebra app!",         // optional — max 500 characters
  "page": "/dashboard",               // optional — current page path
  "viewportWidth": 1440,              // optional — screen width in pixels
  "userAgent": "Mozilla/5.0 ..."      // optional — browser user agent string
}

// Response 201
{ "id": "fb_abc123" }

// Error 400 — invalid mood value
{ "error": "invalid_mood" }

// Error 400 — comment exceeds 500 characters
{ "error": "comment_too_long" }

// Error 400 — unknown category in categories array
{ "error": "invalid_category" }

// Error 429 — more than 5 submissions in the last hour
{ "error": "feedback_rate_limited" }
```

---

## Offers (Grocery Deals)

**No auth required** — these are public endpoints.

### GET /offers/search
Search for grocery offers near a location (defaults to Haninge).
```json
// Query parameters:
// ?q=mjölk                                      - Search query (required)
// ?lat=59.168&lng=18.137&radius=10000          - Custom location (optional)
// ?exclude=Coop,ICA                            - Exclude stores (optional, comma-separated)

// Response 200
{
  "offers": [
    {
      "id": "offer_123",
      "heading": "Arla Mellanmjölk 1.5L",
      "description": "Ekologisk mellanmjölk",
      "price": 19.95,
      "prePrice": 23.95,           // optional — omitted when not discounted
      "currency": "SEK",
      "validFrom": "2026-02-10T00:00:00Z",
      "validTo": "2026-02-16T23:59:59Z",
      "storeName": "ICA Maxi",
      "storeLogo": "https://...",   // optional
      "storeAddress": "Handelsvägen 1",  // optional
      "storeCity": "Haninge",            // optional
      "imageUrl": "https://..."          // optional
    }
  ],
  "count": 1
}

// Timeout: 30 seconds
```

### GET /offers/discounts
Get top discounted offers sorted by discount percentage.
```json
// Query parameters:
// ?exclude=Coop,ICA    - Exclude stores (optional, comma-separated)

// Response 200
{
  "offers": [
    {
      "id": "offer_789",
      "heading": "Lax Filéer 500g",
      "price": 49.90,
      "prePrice": 89.90,
      "currency": "SEK",
      "validFrom": "2026-02-10T00:00:00Z",
      "validTo": "2026-02-16T23:59:59Z",
      "storeName": "Willys"
    }
  ],
  "count": 1
}

// Timeout: 30 seconds
```

### GET /offers/stores
Get list of available stores in the area.
```json
// Response 200
{
  "stores": ["ICA Maxi", "Coop", "Willys", "Hemköp", "City Gross"]
}

// Timeout: 10 seconds
```

---

## Error-format

Alla errors följer samma struktur:
```json
{
  "error": "error_code",
  "message": "Läsbar beskrivning (valfri)"
}
```

| Kod | HTTP | Betydelse |
|-----|------|-----------|
| `invalid_credentials` | 401 | Fel email/lösenord |
| `unauthorized` | 401 | Authorization-header saknas |
| `invalid_token_format` | 401 | Ogiltigt format på Authorization-headern (saknar "Bearer "-prefix) |
| `invalid_token` | 401 | JWT-token är ogiltig, utgången eller kan inte valideras |
| `token_revoked` | 401 | Token har återkallats (t.ex. efter lösenordsbyte eller att ha lämnat hushållet) |
| `invalid_reset_token` | 400 | Reset-token finns inte |
| `expired_reset_token` | 400 | Reset-token har gått ut (>1h sedan skapandet) |
| `used_reset_token` | 400 | Reset-token har redan använts |
| `weak_password` | 400 | Nytt lösenord är kortare än 8 tecken |
| `email_taken` | 400 | Email redan registrerad |
| `code_required` | 400 | Inbjudningskod saknas i requesten |
| `invalid_code` | 400 | Inbjudningskod ogiltig/utgången |
| `already_member` | 409 | Användaren är redan medlem i ett hushåll |
| `household_not_found` | 404 | Hushållet finns inte |
| `no_fields_to_update` | 400 | Minst ett fält krävs vid status-uppdatering |
| `not_found` | 404 | Resursen finns inte |
| `no_active_menu` | 404 | Ingen aktiv meny |
| `invalid_days` | 400 | Ogiltigt dagformat eller antal dagar |
| `name_required` | 400 | Receptnamn saknas |
| `invalid_servings` | 400 | Ogiltigt antal portioner |
| `ingredients_required` | 400 | Ingredienser saknas |
| `instructions_required` | 400 | Instruktioner saknas |
| `name_too_long` | 400 | Receptnamnet är för långt |
| `too_many_ingredients` | 400 | För många ingredienser |
| `cannot_remove` | 403 | Kan inte ta bort sig själv eller ägaren |
| `forbidden` | 403 | Åtkomst nekad (otillräckliga rättigheter eller fel hushåll) |
| `menu_not_found` | 404 | Angivet menuId hittades inte |
| `no_recipes_available` | 400 | Inga recept att generera meny från |
| `invalid_input` | 400 | Ogiltig indata till recipe parser |
| `invalid_mood` | 400 | Ogiltigt mood-värde för feedback (måste vara "good", "okay" eller "bad") |
| `comment_too_long` | 400 | Feedback-kommentar överstiger 500 tecken |
| `invalid_category` | 400 | Okänd feedback-kategori |
| `feedback_rate_limited` | 429 | Max 5 feedback-inlämningar per timme och användare |
| `service_unavailable` | 502 | Extern tjänst (Tjek API) svarade inte |
| `internal_error` | 500 | Oväntat serverfel |

---

## Frontend Routes

The frontend uses Vue Router with the following routes:

| Route | Component | Auth Required | Member Access | Notes |
|-------|-----------|---------------|---------------|-------|
| `/` | LandingView | No | No | Public landing page |
| `/register` | OnboardingView | No | No | User registration |
| `/login` | LoginView | No | No | User login |
| `/forgot-password` | ForgotPasswordView | No | No | Request password reset email |
| `/reset-password` | ResetPasswordView | No | No | Consume reset token from email link |
| `/dashboard` | DashboardView | Yes | No | Main dashboard (guests can view) |
| `/menu/generate` | GenerateMenuView | Yes | Yes | Menu generator (members only) |
| `/recipes` | RecipesView | Yes | Yes | Unified recipes page with tabs |
| `/recipes/parse` | *(redirect to /recipes)* | Yes | Yes | Legacy route, redirects to recipes |
| `/shopping-list` | ShoppingListView | Yes | Yes | Shopping list (members only) |
| `/about` | AboutView | No | No | About page |
| `/offers-poc` | OffersView | No | No | Offers POC page |

**Authentication Guard:**
- Routes with `requiresAuth: true` redirect to `/login` if not authenticated
- Routes with `requiresMember: true` redirect to `/dashboard` if user is a guest
- Authenticated users trying to access `/login` are redirected to `/dashboard`

---

## Status

| Endpoint | Backend | Frontend |
|----------|---------|----------|
| GET /health | ✅ | ⬜ |
| POST /auth/register | ✅ | ✅ |
| POST /auth/login | ✅ | ✅ |
| POST /auth/forgot-password | ✅ | ✅ |
| POST /auth/reset-password | ✅ | ✅ |
| GET /households/me | ✅ | ✅ |
| PATCH /households/me | ✅ | ✅ |
| POST /households/invite | ✅ | ✅ |
| POST /households/join | ✅ | ✅ |
| GET /households/members/status | ✅ | ✅ |
| PATCH /households/members/:id/status | ✅ | ✅ |
| DELETE /households/members/:id | ✅ | ✅ |
| GET /recipes | ✅ | ✅ |
| GET /recipes/:id | ✅ | ✅ |
| POST /recipes | ✅ | ✅ |
| PUT /recipes/{id} | ✅ | ✅ |
| DELETE /recipes/{id} | ✅ | ✅ |
| POST /recipes/parse | ✅ | ✅ |
| POST /recipes/parse-and-save | ✅ | ✅ |
| POST /menus/generate | ✅ | ✅ |
| PUT /menus/current | ✅ | ✅ |
| GET /menus/current | ✅ | ✅ |
| GET /shopping-list | ✅ | ✅ |
| PATCH /shopping-list/items/:id | ✅ | ✅ |
| POST /shopping-list/items | ✅ | ✅ |
| DELETE /shopping-list/items/:id | ✅ | ✅ |
| GET /offers/search | ✅ | ✅ |
| GET /offers/discounts | ✅ | ✅ |
| GET /offers/stores | ✅ | ✅ |
| POST /feedback | ✅ | ✅ |
