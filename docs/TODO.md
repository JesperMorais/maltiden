# TODO - Måltiden

Uppgiftslista för Jesper & David.

---

## Arbetsuppdelning

För att kunna jobba parallellt delar vi upp i **Backend** och **Frontend**.
Nyckeln är att definiera API-kontrakt tidigt så båda kan jobba oberoende.

| Spår | Ansvarig | Teknik |
|------|----------|--------|
| Backend | ? | Go, SQLite |
| Frontend | ? | Vue 3, PWA |

---

## Fas 0: Grundsetup (gör tillsammans/synka)

**Mål:** Gemensam grund så ni kan jobba separat efteråt.

- [ ] Bestäm vem som tar Backend vs Frontend
- [ ] Definiera API-kontrakt (se nedan)
- [ ] Sätt upp Git-branches: `main`, `dev`, `feature/*`
- [ ] Bestäm merge-strategi (PR till `dev`, sedan `dev` → `main`)
- [ ] Skapa `.env.example` för båda projekten

### API-kontrakt (definiera först!)

Dokumentera dessa endpoints innan ni börjar koda:

```
POST   /api/auth/register     # Registrera användare
POST   /api/auth/login        # Logga in, returnerar JWT
POST   /api/auth/refresh      # Förnya JWT

GET    /api/users/me          # Hämta inloggad användare
GET    /api/household         # Hämta hushåll + medlemmar
POST   /api/household/invite  # Generera inbjudningskod
POST   /api/household/join    # Gå med via kod

GET    /api/recipes           # Lista recept (med sökning/filter)
GET    /api/recipes/:id       # Hämta ett recept

POST   /api/menus/generate    # Generera veckomeny
GET    /api/menus/current     # Hämta aktiv veckomeny
PATCH  /api/menus/:id/days/:day  # Uppdatera dag (skippa, personer, etc.)

GET    /api/shopping-list     # Hämta inköpslista för aktiv meny
PATCH  /api/shopping-list/:id # Bocka av vara
```

---

## Fas 1: MVP Grund (parallellt arbete)

### Backend-spår

**User Story:** *Som frontend vill jag ha fungerande API:er så jag kan bygga UI.*

#### 1.1 Projektsetup
- [ ] Sätt upp Go-projektstruktur enligt arkitektur
- [ ] Konfigurera SQLite + migrations
- [ ] Skapa grundläggande HTTP-server med router
- [ ] Implementera error-hantering och logging
- [ ] Skapa mock-data för utveckling

#### 1.2 Auth-system
- [ ] Skapa `users`-tabell (id, email, password_hash, created_at)
- [ ] Implementera `POST /api/auth/register`
- [ ] Implementera `POST /api/auth/login` (returnerar JWT)
- [ ] Skapa auth-middleware som validerar JWT
- [ ] Implementera `GET /api/users/me`

#### 1.3 Hushållshantering
- [ ] Skapa `households`-tabell + `household_members`-tabell
- [ ] Auto-skapa hushåll vid registrering
- [ ] Implementera `POST /api/household/invite` (generera kod)
- [ ] Implementera `POST /api/household/join` (validera kod, max 7 dagar)
- [ ] Implementera `GET /api/household`

#### 1.4 Receptdatabas
- [ ] Skapa `recipes`-tabell + `ingredients`-tabell
- [ ] Implementera `GET /api/recipes` (med pagination, sökning)
- [ ] Implementera `GET /api/recipes/:id`
- [ ] Seed: Lägg in 10+ testrecept i migration

#### 1.5 Menygeneration
- [ ] Skapa `menus`-tabell + `menu_days`-tabell
- [ ] Implementera `POST /api/menus/generate` (slumpa 5 recept)
- [ ] Implementera `GET /api/menus/current`
- [ ] Implementera `PATCH /api/menus/:id/days/:day`

#### 1.6 Inköpslista
- [ ] Skapa `shopping_items`-tabell
- [ ] Implementera `GET /api/shopping-list` (aggregera från meny)
- [ ] Implementera `PATCH /api/shopping-list/:id` (toggle checked)
- [ ] Logik: Slå ihop dubbletter, kategorisera

---

### Frontend-spår

**User Story:** *Som användare vill jag ha ett snyggt gränssnitt för att planera måltider.*

#### 1.1 Projektsetup
- [ ] Sätt upp Vue 3 + Vite + TypeScript
- [ ] Konfigurera PWA (vite-plugin-pwa)
- [ ] Sätt upp Vue Router
- [ ] Sätt upp Pinia för state management
- [ ] Skapa API-klient med axios/fetch
- [ ] **Skapa mock-API för lokal utveckling utan backend**

#### 1.2 Designsystem
- [ ] Bestäm färgpalett och typografi
- [ ] Skapa baskomponenter: Button, Input, Card, Modal
- [ ] Skapa layout-komponenter: Header, Navigation, Page
- [ ] Mobil-först responsiv design

#### 1.3 Auth-vyer
- [ ] Login-sida (`/login`)
- [ ] Registrerings-sida (`/register`)
- [ ] Auth-guard för skyddade routes
- [ ] Spara JWT i localStorage/cookie
- [ ] Visa inloggad användare i header

#### 1.4 Hushålls-vyer
- [ ] Hushålls-sida (`/household`)
- [ ] Visa medlemmar i hushållet
- [ ] "Bjud in"-knapp som visar inbjudningskod
- [ ] "Gå med"-formulär för att ange kod

#### 1.5 Recept-vyer
- [ ] Receptlista-sida (`/recipes`)
- [ ] Sökfält + kategorifilter
- [ ] Receptkort-komponent
- [ ] Receptdetalj-sida (`/recipes/:id`)

#### 1.6 Meny-vyer
- [ ] Veckomeny-sida (`/menu`) - huvudvy
- [ ] Visa 5-7 dagar med receptkort
- [ ] "Generera ny meny"-knapp
- [ ] Per dag: Skippa-toggle, antal personer, matlåda-checkbox
- [ ] Visuell feedback när dag ändras

#### 1.7 Inköpslista-vy
- [ ] Inköpslista-sida (`/shopping`)
- [ ] Gruppera efter kategori (mejeri, kött, etc.)
- [ ] Checkbox för att bocka av
- [ ] **Offline-stöd:** Spara lokalt, synka när online
- [ ] Visuell indikator om offline

---

## Fas 2: Integration & Test

**Mål:** Koppla ihop frontend och backend.

- [ ] Byt ut frontend mock-API mot riktig backend
- [ ] Testa alla flöden end-to-end
- [ ] Fixa CORS-konfiguration
- [ ] Philip: Testa alla user flows
- [ ] Fixa buggar som upptäcks

---

## Fas 3: Deployment

- [ ] Sätt upp Fly.io för backend
- [ ] Konfigurera miljövariabler (JWT_SECRET, etc.)
- [ ] Deploya frontend (Fly.io eller Vercel/Netlify)
- [ ] Konfigurera custom domain
- [ ] Sätt upp HTTPS
- [ ] Testa i produktion

---

## Iteration 2+ (efter MVP)

Dessa kan också delas upp parallellt:

| Feature | Backend | Frontend |
|---------|---------|----------|
| Preferenser (gilla/ogilla) | Spara i DB, vikta generation | UI för gilla/ogilla |
| Skafferi | CRUD för inventory | Skafferi-vy |
| Kylskåpsscan | Claude Vision integration | Kamera-vy |
| Näringsbalans | Livsmedelsverkets API | Visa näringsvärden |
| Google/Apple OAuth | OAuth-flöde i backend | OAuth-knappar |

---

## Synkpunkter

Förslag på när ni bör synka:

1. **Efter Fas 0** - API-kontrakt klara
2. **Mitt i Fas 1** - Snabb check att allt funkar
3. **Innan Fas 2** - Redo att integrera
4. **Efter Fas 2** - Redo för deploy

---

## Anteckningar & Beslut

| Datum | Beslut | Vem |
|-------|--------|-----|
| | | |

---

**Senast uppdaterad:** 2025-01-04
