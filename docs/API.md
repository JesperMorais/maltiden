# Måltiden API-kontrakt

Base URL: `http://localhost:8080` (dev), `https://api.maltiden.se` (prod)

## Recent Changes (PR #110)

**Bug Fixes:**
- Orphaned account handling: if `POST /households/join` fails after `POST /auth/register` succeeds, the frontend now detects the already-authenticated state and redirects to dashboard instead of showing an unrecoverable form error.
- Stale `currentMenuId`: `GET /menus/current` returning `null` now correctly resets the cached menu ID on the frontend, preventing shopping list fetches with an expired menu ID.

**Documentation Updates:**
- Added `/shopping-list` frontend route (was missing from route table)
- Corrected `POST /auth/register` error codes to match backend (`email_already_exists` at 409, added `weak_password` and `invalid_email`)
- Documented required `menuId` query parameter on shopping list endpoints
- Documented IDOR protection (403 `forbidden`) on shopping list endpoints
- Added `already_member` (409) and `code_required` (400) errors for `POST /households/join`
- Added `emoji` field to recipe and menu day response types
- Added `recipeName` field to menu day response type
- Expanded error code reference table

---

## Auth

### POST /auth/register
```json
// Request
{
  "email": "anna@example.com",
  "password": "minst8tecken",
  "name": "Anna",
  "householdName": "Familjen Svensson"  // optional — creates household with this name
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

// Error 409
{ "error": "email_already_exists" }

// Error 400
{ "error": "weak_password" }
{ "error": "invalid_email" }
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
```

### POST /households/invite
Guests cannot create invite codes — only owners and members can.
```json
// Request
{}

// Response 201
{
  "code": "ABC123",
  "expiresAt": "2025-01-20T12:00:00Z"
}

// Error 403
{ "error": "forbidden" }  // caller is a guest
```

### POST /households/join
```json
// Request
{ "code": "ABC123" }

// Response 200
{ "householdId": "hh_xyz789" }

// Error 400
{ "error": "code_required" }  // missing code field
{ "error": "invalid_code" }   // code not found or expired

// Error 409
{ "error": "already_member" }  // user is already in a household
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

### PATCH /households/members/{id}/status
```json
// Request (partial update — at least one field required)
{ "isEatingToday": false }
// or
{ "wantsLunchBox": true }
// or both
{ "isEatingToday": true, "wantsLunchBox": true }

// Response 200
{ "ok": true }

// Error 400
{ "error": "no_fields_to_update" }
```

### DELETE /households/members/{id}
```json
// Response 200
{ "ok": true }

// Error 403
{ "error": "cannot_remove" }  // can't remove yourself or owner
{ "error": "forbidden" }      // caller lacks permission
```

---

## Recipes

### GET /recipes
Public endpoint — no auth required.
```json
// Response 200
{
  "recipes": [
    {
      "id": "rec_001",
      "name": "Köttfärssås",
      "servings": 4,
      "tags": ["vardag", "barn"],
      "emoji": "🍝"  // optional
    }
  ]
}
```

### GET /recipes/{id}
Public endpoint — no auth required.
```json
// Response 200
{
  "id": "rec_001",
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
}
```

### POST /recipes
Auth required.
```json
// Request
{
  "name": "Köttfärssås",
  "servings": 4,
  "ingredients": [...],
  "instructions": [...],
  "tags": ["vardag"]
}

// Response 201
{ "id": "rec_001" }
```

### POST /recipes/parse
Parse unstructured recipe text into structured data using AI. Auth required.
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
Parse recipe text and immediately save it to the database. Auth required.
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
    { "date": "2025-01-21", "recipeId": "rec_002", "recipeName": "Pasta Carbonara", "emoji": "🍜", "servings": 4 },
    { "date": "2025-01-22", "skip": true, "servings": 0 },
    { "date": "2025-01-23", "recipeId": "rec_003", "recipeName": "Lax med ris", "emoji": "🐟", "servings": 6 }
  ]
}

// Error 400
{ "error": "invalid_days" }      // days must be 1–14
{ "error": "invalid_servings" }  // servings must be > 0
{ "error": "no_recipes_available" }
```

### GET /menus/current
```json
// Response 200
{
  "id": "menu_001",
  "days": [
    { "date": "2025-01-20", "recipeId": "rec_001", "recipeName": "Köttfärssås", "emoji": "🍝", "servings": 4 }
  ]
}

// Response 404 (no active menu)
{ "error": "no_active_menu" }
```

---

## Shopping List

**Auth required:** `Authorization: Bearer <token>`

IDOR protection is enforced: the menu must belong to the caller's household.

### GET /shopping-list
```json
// Query: ?menuId=menu_001  (required)

// Response 200
{
  "menuId": "menu_001",
  "categories": [
    {
      "name": "Kött & Fisk",
      "items": [
        { "id": "item_001", "name": "Köttfärs", "amount": 800, "unit": "g", "checked": false }
      ]
    },
    {
      "name": "Mejeri",
      "items": [
        { "id": "item_002", "name": "Grädde", "amount": 2, "unit": "dl", "checked": false }
      ]
    }
  ]
}

// Error 400
{ "error": "invalid_input" }  // missing or malformed menuId

// Error 403
{ "error": "forbidden" }  // menu belongs to a different household

// Error 404
{ "error": "menu_not_found" }
```

### PATCH /shopping-list/items/{id}
```json
// Query: ?menuId=menu_001  (required)
// Request
{ "checked": true }

// Response 200
{ "ok": true }

// Error 400
{ "error": "invalid_input" }  // missing or malformed menuId

// Error 403
{ "error": "forbidden" }  // menu belongs to a different household

// Error 404
{ "error": "menu_not_found" }
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
      "product": "Arla Mellanmjölk 1.5L",
      "store": "ICA Maxi",
      "originalPrice": 23.95,
      "offerPrice": 19.95,
      "discount": 17,  // percentage
      "validFrom": "2026-02-10",
      "validTo": "2026-02-16",
      "catalogId": "cat_456",
      "catalogPages": [12, 13],
      "imageUrl": "https://..."
    }
  ],
  "query": "mjölk",
  "location": { "lat": 59.168, "lng": 18.137, "radius": 10000 }
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
      "product": "Lax Filéer 500g",
      "store": "Willys",
      "originalPrice": 89.90,
      "offerPrice": 49.90,
      "discount": 44,
      "validFrom": "2026-02-10",
      "validTo": "2026-02-16"
    }
  ]
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

## Error Format

All errors use the same structure:
```json
{
  "error": "error_code",
  "message": "Human-readable description (optional)"
}
```

| Code | HTTP | Description |
|------|------|-------------|
| `invalid_credentials` | 401 | Wrong email or password |
| `unauthorized` | 401 | Token missing or invalid |
| `email_already_exists` | 409 | Email already registered |
| `weak_password` | 400 | Password too short or too simple |
| `invalid_email` | 400 | Email format invalid |
| `code_required` | 400 | Invite code field missing |
| `invalid_code` | 400 | Invite code not found or expired |
| `already_member` | 409 | User is already in a household |
| `no_fields_to_update` | 400 | PATCH request body has no recognized fields |
| `cannot_remove` | 403 | Cannot remove yourself or the household owner |
| `forbidden` | 403 | Caller lacks permission for this action |
| `not_found` | 404 | Resource not found |
| `menu_not_found` | 404 | Menu ID not found |
| `no_active_menu` | 404 | No current active menu for household |
| `invalid_days` | 400 | days parameter out of valid range |
| `invalid_servings` | 400 | servings must be a positive integer |
| `no_recipes_available` | 400 | No recipes to generate a menu from |
| `invalid_input` | 400 | Request body or query param malformed |
| `internal_error` | 500 | Unexpected server error |

---

## Frontend Routes

The frontend uses Vue Router with the following routes:

| Route | Component | Auth Required | Member Access | Notes |
|-------|-----------|---------------|---------------|-------|
| `/` | LandingView | No | No | Public landing page |
| `/register` | OnboardingView | No | No | User registration |
| `/login` | LoginView | No | No | User login |
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
| GET /households/me | ✅ | ✅ |
| POST /households/invite | ✅ | ✅ |
| POST /households/join | ✅ | ✅ |
| GET /households/members/status | ✅ | ✅ |
| PATCH /households/members/{id}/status | ✅ | ✅ |
| DELETE /households/members/{id} | ✅ | ✅ |
| GET /recipes | ✅ | ✅ |
| GET /recipes/{id} | ✅ | ✅ |
| POST /recipes | ✅ | ✅ |
| POST /recipes/parse | ✅ | ✅ |
| POST /recipes/parse-and-save | ✅ | ✅ |
| POST /menus/generate | ✅ | ✅ |
| GET /menus/current | ✅ | ✅ |
| GET /shopping-list | ✅ | ✅ |
| PATCH /shopping-list/items/{id} | ✅ | ✅ |
| GET /offers/search | ✅ | ✅ |
| GET /offers/discounts | ✅ | ✅ |
| GET /offers/stores | ✅ | ✅ |
