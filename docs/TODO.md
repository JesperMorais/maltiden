# TODO — Måltiden

Uppgiftslista för Jesper & David.

**Senast uppdaterad:** 2026-02-20

---

## Arbetsuppdelning

| Spår | Ansvarig | Teknik |
|------|----------|--------|
| Backend | David | Go, SQLite |
| Frontend | Jesper | Vue 3, TypeScript |
| Test | Philip | Manuell QA |

---

## Avklarat ✅

<details>
<summary>Fas 1: MVP Grund ✅</summary>

### Backend ✅
- [x] Go-projektstruktur med lager-arkitektur (handler → service → storage → domain)
- [x] SQLite + migrations (001–007)
- [x] HTTP-server med router och dependency injection
- [x] Auth-system: register, login, JWT middleware
- [x] Hushållshantering: skapa, bjud in, gå med, roller
- [x] Receptdatabas: skapa, lista, detalj, JSON-lagring av ingredienser/instruktioner
- [x] Menygeneration: slumpa recept, hantera skip-dagar, låsa dagar
- [x] Inköpslista: aggregera från meny, toggle checked
- [x] Rate limiting och request ID middleware
- [x] Graceful shutdown (SIGTERM/SIGINT med 10s timeout)

### Frontend ✅
- [x] Vue 3 + Vite + TypeScript strict setup
- [x] Vue Router med auth guards (requiresAuth, requiresMember)
- [x] Pinia stores (user, dashboard, menuGenerator)
- [x] API-klient med Axios, JWT interceptor, 401 auto-redirect
- [x] Designsystem: BaseButton, BaseCard, ProgressBar, etc.
- [x] Auth-vyer: login, registrering, onboarding
- [x] Dashboard med veckomeny, dagens måltid, inköpslista-widget
- [x] Menygeneration med slot machine-animation
- [x] Receptvyer med tabs (mina recept + lägg till)
- [x] Mock-läge för utveckling utan backend
- [x] SettingsModal med profil, utseende, utloggning

### Integration & Polish ✅
- [x] Frontend kopplad till riktig backend
- [x] CORS-konfiguration
- [x] Receptparsning via Claude API (parse + parse-and-save)
- [x] Erbjudanden-integration (Tjek API: sök, rabatter, butiker) — POC
- [x] Skeleton loading-komponenter
- [x] Optimistic UI för medlemsstatus-toggles
- [x] Simulerade progress bars för lång-körande operationer
- [x] Animationer med motion-v

### Deployment ✅
- [x] Fly.io deployment (region: arn)
- [x] Multi-stage Dockerfile (Node build → Go build → slim runtime)
- [x] Miljövariabler konfigurerade (JWT_SECRET, DATABASE_PATH, etc.)
- [x] CI/CD med GitHub Actions (backend + frontend pipelines)
- [x] Claude Code auto-review på PR:ar

</details>

---

## MVP — Redo för betafamiljer (2–5 familjer)

### Nuläge

Appen är **arkitekturmässigt solid** med alla kärnflöden på plats. Det som saknas för att bjuda in familjer är UX-polish, saknad feedback, och ett par kritiska luckiga funktioner.

**Vad familjer KAN göra idag:** registrera sig, logga in, skapa/gå med i hushåll, bläddra/lägga till recept, AI-parsa recept, generera veckomeny, se inköpslista-sammanfattning, logga ut.

**Vad som saknas:** redigera/ta bort recept, interaktiv inköpslista (markera varor), toast-feedback, spara meny (re-genererar idag), kopiera inbjudningskod.

---

### 🔴 Kritiskt — Måste fixas innan lansering

Utan dessa kan familjer inte använda appen meningsfullt.

#### K1. Toast/notifikationssystem
- **Problem:** Ingen visuell feedback vid sparning, generering, fel, eller lyckade handlingar
- **Nuläge:** `HouseholdWidget.vue` har `// TODO: Show toast notification` på 2 ställen. Inga andra filer hanterar detta
- **Scope:**
  - [ ] **FE:** Skapa `ToastNotification.vue` komponent (success/error/info)
  - [ ] **FE:** Skapa `useToast()` composable med queue-hantering
  - [ ] **FE:** Integrera i alla stores/views som har handlingar (recept, meny, hushåll)
- **Ansvarig:** Jesper
- **Effort:** 1–2 dagar

#### K2. Interaktiv inköpslista
- **Problem:** Dashboard visar bara en widget-sammanfattning (antal varor, progress). Inget fullständigt UI för att se/markera enskilda varor
- **Nuläge:**
  - Backend: `GET /shopping-list` + `PATCH /shopping-list/items/{id}` fungerar
  - Frontend: `useShoppingList.ts` composable finns med optimistic updates, `shopping.api.ts` finns med real/mock toggle
  - Saknas: Fullständig `ShoppingListPanel.vue` eller modal med varulista, kategorier, avkryssning
- **Scope:**
  - [ ] **FE:** Skapa `ShoppingListModal.vue` (eller panel) med kategoriserad varulista
  - [ ] **FE:** Wire:a composable till modal — visa varor per kategori, avkryssning, progress
  - [ ] **FE:** Koppla "visa lista"-knappen i `ShoppingListWidget.vue` till modalen
  - [ ] **FE:** Testa med riktig backend-data
- **Ansvarig:** Jesper
- **Effort:** 2–3 dagar

#### K3. Recept — redigera
- **Problem:** Recept kan inte ändras efter sparning. Om AI-parsern får fel eller man gör en typo → måste man radera och göra om
- **Nuläge:** Ingen `PUT/PATCH /recipes/{id}` endpoint. Ingen edit-form i frontend
- **Scope:**
  - [ ] **BE:** `UpdateRecipe()` i storage, service, handler
  - [ ] **BE:** `PUT /recipes/{id}` route i router
  - [ ] **BE:** Validering: bara hushållsmedlem kan redigera sina recept
  - [ ] **FE:** Återanvänd `RecipeEditForm.vue` i edit-läge
  - [ ] **FE:** Redigera-knapp i `RecipeDetailModal.vue`
  - [ ] **FE:** API-service: `updateRecipe(id, data)` i `recipes.api.ts`
- **Ansvarig:** David (BE) + Jesper (FE)
- **Effort:** 1–2 dagar

#### K4. Recept — ta bort
- **Problem:** Användare kan inte ta bort felaktiga eller testrecept
- **Nuläge:** Ingen `DELETE /recipes/{id}` endpoint. Ingen delete-knapp i UI
- **Scope:**
  - [ ] **BE:** `DeleteRecipe()` i storage, service, handler
  - [ ] **BE:** `DELETE /recipes/{id}` route
  - [ ] **BE:** Kontrollera att receptet inte är i aktiv meny (eller hantera)
  - [ ] **FE:** Ta-bort-knapp med bekräftelsedialog i `RecipeDetailModal.vue`
  - [ ] **FE:** API-service: `deleteRecipe(id)` i `recipes.api.ts`
- **Ansvarig:** David (BE) + Jesper (FE)
- **Effort:** 0.5–1 dag

#### K5. Meny — spara nuvarande val
- **Problem:** "Spara"-knappen anropar `POST /menus/generate` igen → genererar ny meny istället för att spara den visade
- **Nuläge:** Kommentar i `menuGenerator.ts:329`: "For MVP: Call generateMenu to create a new menu. In the future, this should call a PUT /menus/current endpoint"
- **Scope:**
  - [ ] **BE:** `PUT /menus/current` eller `PUT /menus/{id}` endpoint som sparar exakta recept-val
  - [ ] **BE:** Handler + service + storage för att uppdatera befintlig meny
  - [ ] **FE:** Ändra `saveMenu()` i store att anropa PUT istället för generate
  - [ ] **FE:** Bevara låsta dagar och använde val
- **Ansvarig:** David (BE) + Jesper (FE)
- **Effort:** 1–2 dagar

---

### 🟡 Viktigt — Bör fixas innan lansering

Appen fungerar utan dessa, men UX blir klart sämre.

#### V1. Kopiera inbjudningskod
- **Problem:** `handleShowInvite()` i `DashboardView.vue` gör bara `console.log('Show invite code')` med `// TODO: Show invite modal`
- **Scope:**
  - [ ] **FE:** Skapa invite-modal med kod-visning + kopiera-till-urklipp-knapp
  - [ ] **FE:** Eventuellt "Dela via..."-funktion (Web Share API)
- **Ansvarig:** Jesper
- **Effort:** 0.5 dag

#### V2. Sessionshantering UX
- **Problem:** JWT löper ut efter 7 dagar utan förvarning. Användaren får 401 och slängs till login-sidan utan förklaring
- **Nuläge:** 401-interceptor i `client.ts` tar bort token och redirectar till `/login`
- **Scope:**
  - [ ] **FE:** Visa "Din session har löpt ut, logga in igen"-meddelande (kräver toast-system, K1)
  - [ ] **FE:** Eventuellt proaktiv kontroll av token-expiry vid app-start
- **Ansvarig:** Jesper
- **Effort:** 0.5 dag (efter K1)

#### V3. Felhantering — audit
- **Problem:** Vissa API-fel kan visa kryptiska meddelanden eller bara hamna i console.log
- **Scope:**
  - [ ] **FE:** Gå igenom alla API-anrop och se till att fel visas som svenska, användarvänliga toasts
  - [ ] **FE:** Särskilt: rate limiting-fel, nätverksfel, server 500
- **Ansvarig:** Jesper
- **Effort:** 1 dag (efter K1)

#### V4. Databasbackup
- **Problem:** SQLite på Fly.io-volym — om volymen försvinner, försvinner all data
- **Scope:**
  - [ ] **OPS:** Verifiera att Fly.io volume snapshots är aktiverade
  - [ ] **OPS:** Sätt upp daglig backup-rutin (kan vara simpelt `fly ssh sftp get`)
- **Ansvarig:** David
- **Effort:** 0.5 dag

#### V5. Startrecept / seed-data
- **Problem:** Nya familjer startar med 0 recept → kan inte generera menyer
- **Scope:**
  - [ ] **BE/DATA:** Skapa 15–20 svenska basisrecept (köttfärssås, pannkakor, pasta carbonara, etc.)
  - [ ] **BE:** Seed-script eller migration som lägger in dem som "globala" recept
  - [ ] Alternativt: Philip matar in via appen
- **Ansvarig:** Philip + David
- **Effort:** 1–2 dagar

---

### 🟢 Kan vänta — Post-MVP / baserat på feedback

| Feature | Beskrivning | Effort |
|---------|-------------|--------|
| Glömt lösenord | Reset-flöde via email. Kan hanteras manuellt under beta ("kontakta oss") | 1–2 dagar |
| Felspårning (Sentry) | Automatisk error tracking. Manuell loggkoll räcker för 2–5 familjer | 0.5 dag |
| Erbjudanden i inköpslista | Tjek API finns som POC, integrera i shopping-flödet | 2–3 dagar |
| Menyhistorik | Se tidigare veckors menyer | 1 dag |
| Byt enskild dag i meny | Byta recept på en dag utan att regenerera allt | 1 dag |
| Profilsida (separat vy) | Idag bara SettingsModal. Flytta till egen `/settings`-route | 0.5 dag |
| PWA / installbar app | Service worker, offline-stöd | 2–3 dagar |
| Preferensinlärning | Gilla/ogilla recept, vikta menygeneration | 2–3 dagar |
| Skafferi/inventory | CRUD för pantry items + vy | 2–3 dagar |
| Kylskåpsscan | Claude Vision integration | 3–5 dagar |
| Näringsbalans | Livsmedelsverkets API | 2–3 dagar |
| Google/Apple OAuth | Extra inloggningsmetoder | 2–3 dagar |

---

## Lansering

### Före lansering (checklista)
- [ ] Alla 🔴-uppgifter klara
- [ ] Minst 🟡 V1 + V4 klara
- [ ] Philip gör en komplett QA-genomgång av alla flöden
- [ ] Minst 15 seed-recept finns
- [ ] Fly.io volume snapshots verifierade

### Soft launch
- Deploy till `maltiden.fly.dev`
- Bjud in 2–3 familjer manuellt
- Övervaka Fly.io-loggar första 48 timmarna
- Enkel feedback-kanal (Messenger-grupp / mail)

### Efter 2 veckor
- Samla feedback
- Prioritera 🟢-listan baserat på vad användarna faktiskt efterfrågar
- Fixa buggar som dyker upp
