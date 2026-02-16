# Måltiden API-kontrakt

Base URL: `http://localhost:8080` (dev), `https://api.maltiden.se` (prod)

## Recent Changes (PR #35)

**Recipe Navigation Consolidation:**
- Added new unified `/recipes` route with tabbed interface ("Mina recept" and "Lägg till")
- Added redirect from legacy `/recipes/parse` route to `/recipes`
- Route requires authentication and member access

**New API Endpoints:**
- `POST /recipes/parse` - Parse unstructured recipe text into structured data using AI
- `POST /recipes/parse-and-save` - Parse and immediately save recipe to database
- Both endpoints support 60-second timeout for AI processing
- Both endpoints accept optional `source` URL parameter

**Frontend Implementation:**
- Frontend API client implemented in `frontend/src/api/recipes.api.ts`
- Mock implementations available for development
- Integrated into new `RecipesView` component with tabs

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
// Response 200
{
  "recipes": [
    {
      "id": "rec_001",
      "name": "Köttfärssås",
      "servings": 4,
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
  "ingredients": [
    { "name": "Köttfärs", "amount": 400, "unit": "g" },
    { "name": "Krossade tomater", "amount": 400, "unit": "g" }
  ],
  "instructions": [
    "Bryn köttfärsen",
    "Tillsätt tomater",
    "Låt sjuda 20 min"
  ],
  "tags": ["vardag", "barn"]
}
```

### POST /recipes (admin/Philip)
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

### GET /menus/current
```json
// Response 200
{ "id": "menu_001", "days": [...] }

// Response 404 (ingen aktiv meny)
{ "error": "no_active_menu" }
```

---

## Shopping List

### GET /shopping-list
```json
// Query: ?menuId=menu_001

// Response 200
{
  "menuId": "menu_001",
  "categories": [
    {
      "name": "Kött & Fisk",
      "items": [
        { "name": "Köttfärs", "amount": 800, "unit": "g", "checked": false }
      ]
    },
    {
      "name": "Mejeri",
      "items": [
        { "name": "Grädde", "amount": 2, "unit": "dl", "checked": false }
      ]
    }
  ]
}
```

### PATCH /shopping-list/items/:id
```json
// Request
{ "checked": true }

// Response 200
{ "ok": true }
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
| POST /recipes/parse | ✅ | ✅ |
| POST /recipes/parse-and-save | ✅ | ✅ |
| POST /menus/generate | ✅ | ✅ |
| GET /menus/current | ✅ | ✅ |
| GET /shopping-list | ✅ | ✅ |
| PATCH /shopping-list/items/:id | ✅ | ✅ |
| GET /offers/search | ✅ | ✅ |
| GET /offers/discounts | ✅ | ✅ |
| GET /offers/stores | ✅ | ✅ |
