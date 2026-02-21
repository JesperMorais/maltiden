# Måltiden API-kontrakt

Base URL: `http://localhost:8080` (dev), `https://api.maltiden.se` (prod)

## Recent Changes (PR #82)

**New API Endpoints:**
- `PUT /recipes/{id}` - Update an existing recipe (auth required)
- `DELETE /recipes/{id}` - Delete a recipe (auth required)
- `PUT /menus/current` - Save exact recipe-day selections to the current active menu (auth required)

**Breaking / Additive Changes:**
- `emoji` field added to `CreateRecipeRequest`, `Recipe`, and `RecipeSummary` (optional, omitted when empty)
- Auth client now sets `sessionStorage.session_expired = 'true'` before redirecting to `/login` on 401

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
```json
// Request
{}

// Response 201
{
  "code": "ABC123",
  "expiresAt": "2025-01-20T12:00:00Z"
}
```

### POST /households/join
```json
// Request
{ "code": "ABC123" }

// Response 200
{ "householdId": "hh_xyz789" }

// Error 400
{ "error": "invalid_code" }
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
// Request (partial update)
{ "isEatingToday": false }
// or
{ "wantsLunchBox": true }
// or both
{ "isEatingToday": true, "wantsLunchBox": true }

// Response 200
{ "ok": true }
```

### DELETE /households/members/:id
```json
// Response 200
{ "ok": true }

// Error 403 (can't remove yourself or owner)
{ "error": "cannot_remove" }
```

---

## Recipes

### GET /recipes
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
      "emoji": "🍝",        // optional — omitted when empty
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
Delete a recipe. **Auth required.**
```json
// Response 204 — no body

// Error 404
{ "error": "not_found" }
```

### POST /recipes/parse
Parse unstructured recipe text into structured data using AI.
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
Parse recipe text and immediately save it to the database.
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
    { "date": "2025-01-20", "recipeId": "rec_001", "servings": 4 },
    { "date": "2025-01-21", "recipeId": "rec_002", "servings": 4 },
    { "date": "2025-01-22", "skip": true },
    { "date": "2025-01-23", "recipeId": "rec_003", "servings": 6 }
  ]
}
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
  "householdId": "hh_xyz789",
  "days": [
    { "date": "2026-02-24", "recipeId": "rec_001", "servings": 4 },
    { "date": "2026-02-25", "recipeId": "rec_002", "servings": 4 },
    { "date": "2026-02-26", "skip": true },
    { "date": "2026-02-27", "recipeId": "rec_003", "servings": 6 }
  ],
  "createdAt": "2026-02-21T08:00:00Z"
}

// Error 404
{ "error": "no_active_menu" }

// Error 400
{ "error": "invalid_days" }
```

### GET /menus/current
```json
// Response 200
{
  "id": "menu_001",
  "householdId": "hh_xyz789",
  "days": [
    { "date": "2026-02-24", "recipeId": "rec_001", "servings": 4 },
    { "date": "2026-02-25", "skip": true }
  ],
  "createdAt": "2026-02-21T08:00:00Z"
}

// Response 404 (ingen aktiv meny)
{ "error": "no_active_menu" }
```

---

## Shopping List

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
| `unauthorized` | 401 | Token saknas/ogiltig |
| `email_taken` | 400 | Email redan registrerad |
| `invalid_code` | 400 | Inbjudningskod ogiltig/utgången |
| `not_found` | 404 | Resursen finns inte |
| `no_active_menu` | 404 | Ingen aktiv meny |
| `invalid_days` | 400 | Ogiltigt dagformat i PUT /menus/current |
| `name_required` | 400 | Receptnamn saknas |
| `invalid_servings` | 400 | Ogiltigt antal portioner |
| `ingredients_required` | 400 | Ingredienser saknas |
| `instructions_required` | 400 | Instruktioner saknas |
| `name_too_long` | 400 | Receptnamnet är för långt |
| `too_many_ingredients` | 400 | För många ingredienser |
| `cannot_remove` | 403 | Kan inte ta bort sig själv eller ägaren |
| `forbidden` | 403 | Åtkomst nekad (resursen tillhör annat hushåll) |
| `menu_not_found` | 404 | Angivet menuId hittades inte |
| `no_recipes_available` | 400 | Inga recept att generera meny från |
| `invalid_input` | 400 | Ogiltig indata till recipe parser |
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
| GET /offers/search | ✅ | ✅ |
| GET /offers/discounts | ✅ | ✅ |
| GET /offers/stores | ✅ | ✅ |
