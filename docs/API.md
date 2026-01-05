# Måltiden API-kontrakt

Base URL: `http://localhost:8080` (dev), `https://api.maltiden.se` (prod)

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

## Status

| Endpoint | Backend | Frontend |
|----------|---------|----------|
| GET /health | ✅ | ⬜ |
| POST /auth/register | ⬜ | ⬜ |
| POST /auth/login | ⬜ | ⬜ |
| GET /households/me | ⬜ | ⬜ |
| POST /households/invite | ⬜ | ⬜ |
| POST /households/join | ⬜ | ⬜ |
| GET /households/members/status | ⬜ | ⬜ |
| PATCH /households/members/:id/status | ⬜ | ⬜ |
| DELETE /households/members/:id | ⬜ | ⬜ |
| GET /recipes | ⬜ | ⬜ |
| GET /recipes/:id | ⬜ | ⬜ |
| POST /recipes | ⬜ | ⬜ |
| POST /menus/generate | ⬜ | ⬜ |
| GET /menus/current | ⬜ | ⬜ |
| GET /shopping-list | ⬜ | ⬜ |
| PATCH /shopping-list/items/:id | ⬜ | ⬜ |
