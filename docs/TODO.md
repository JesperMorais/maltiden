# TODO — Måltiden

Uppgiftslista för Jesper & David.

**Senast uppdaterad:** 2026-03-10

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

Alla kritiska (🔴) och viktiga UX-uppgifter (🟡 V1–V5) är **klara**. Kvar innan lansering: mobilpolish (🟠 M1–M6) och Philips QA-genomgång.

**Målgrupp:** 2–5 betafamiljer. **Primär plattform: mobil.** Familjerna kommer främst använda appen i telefonen — vid matlagning, i mataffären, vid middagsplanering i soffan. Desktop är sekundärt.

**Vad familjer KAN göra idag:** registrera sig, logga in, skapa/gå med i hushåll, bläddra 20 svenska basrecept, lägga till/redigera/ta bort recept, AI-parsa recept, generera veckomeny, spara meny, se & markera inköpslista, kopiera inbjudningskod, logga ut — med toast-feedback på alla handlingar.

---

### 🔴 Kritiskt — Måste fixas innan lansering

Utan dessa kan familjer inte använda appen meningsfullt.

#### K1. Toast/notifikationssystem ✅
- **Klart:** `ToastNotification.vue` + `useToast()` composable med success/error/info/warning, motion-v animationer, queue (max 3), auto-dismiss. Integrerat i HouseholdWidget och alla nyckelflöden.

#### K2. Interaktiv inköpslista ✅
- **Klart:** `ShoppingListView.vue` kopplad till router, `useShoppingList.ts` composable ansluten till riktig API, widget-knapp navigerar till fullständig vy med kategorier och avkryssning.

#### K3. Recept — redigera ✅
- **Klart:** `PUT /recipes/{id}` endpoint (handler → service → storage). Frontend: redigera-knapp i `RecipeDetailModal.vue`, återanvänder `RecipeEditForm.vue`, `updateRecipe()` i API-service + mock.

#### K4. Recept — ta bort ✅
- **Klart:** `DELETE /recipes/{id}` endpoint (204 No Content). Frontend: ta-bort-knapp med bekräftelsedialog i `RecipeDetailModal.vue`, `deleteRecipe()` i API-service + mock.

#### K5. Meny — spara nuvarande val ✅
- **Klart:** `PUT /menus/current` endpoint (transaktionell uppdatering av menydagar). Frontend: `saveMenu()` i store anropar PUT med exakta recept-val istället för att re-generera.

---

### 🟡 Viktigt — Bör fixas innan lansering

Appen fungerar utan dessa, men UX blir klart sämre.

#### V1. Kopiera inbjudningskod ✅
- **Klart:** `InviteModal.vue` med kod-visning, kopiera-till-urklipp och "Kopierad!"-feedback. Kopplad till `handleShowInvite()` i dashboard.

#### V2. Sessionshantering UX ✅
- **Klart:** 401-interceptor sätter `session_expired`-flagga i sessionStorage. `LoginView.vue` visar toast "Din session har löpt ut. Logga in igen." vid redirect.

#### V3. Felhantering — audit ✅
- **Klart:** Svenska toast-meddelanden på alla nyckelflöden: recept (skapa/redigera/ta bort), menygeneration/sparning, inköpslista-toggle. Nätverksfel och serverfel hanteras.

#### V4. Databasbackup ✅
- **Klart:** Fly.io volume snapshots verifierade — dagliga automatiska snapshots aktiva (volym `vol_r635e3l1jj1x36nr`, region arn, 1 GB). 5 dagars retention med snapshots var ~24h. Manuell backup möjlig via `fly ssh sftp get /data/maltiden.db ./backups/maltiden-$(date +%Y%m%d).db -a maltiden`. Restore via `fly volumes restore <snapshot-id>`.

#### V5. Startrecept / seed-data ✅
- **Klart:** 20 svenska basrecept via SQL-migrations (`004_seed_recipes.sql` + `008_seed_more_recipes.sql`). Recept: Pasta Carbonara, Kycklingwok, Tacos, Laxfilé, Köttfärssås, Pannkakor, Kycklinggryta, Ärtsoppa, Falukorv, Fiskpinnar, Korvstroganoff, Janssons frestelse, Pytt i panna, Vegetarisk pasta med pesto, Stekt fläsk, Köttbullar, Ugnsbakad torsk, Chili con carne, Tomatsoppa med ostmacka, Kyckling med currysås. Taggar: vardag, barn, klassiker, husmanskost, fisk, vegetariskt, m.fl.

---

### 🟠 Mobilpolish — Måste fixas innan lansering

Appen är desktop-first idag men betafamiljerna använder primärt mobil. Dessa CSS-justeringar krävs för bra UX på 375px–414px (iPhone SE/12/13/14).

#### M1. WCAG AA kontrastproblem ✅
- **Klart:** `--accent-text` (`#c4402e`, 4.75:1) applicerat på alla accent-element. Hardcoded `#FF6B5B` ersatt med CSS-variabler.

#### M2. Landing page — grid-overflow på mobil ✅
- **Klart:** Features-grid, hero och CTA anpassade för 375px. `minmax(280px, 1fr)` → responsiv kolumnbredd, padding reducerad på mobil.

#### M3. Receptparsern — textarea för hög på mobil ✅
- **Klart:** `min-height` reducerad till 150px på <480px-skärmar.

#### M4. TodaysMeal-kort för högt på mobil ✅
- **Klart:** `min-height: 200px`, `padding: 1.5rem 1rem` på <480px. Bättre fold-position på mobil.

#### M5. Receptredigeringsformulär trångt på mobil ✅
- **Klart:** Ingrediensrader stackar vertikalt på <480px istället för fasta 90px-kolumner.

#### M6. Touch targets under 44px-minimum ✅
- **Klart:** Alla interaktiva element minst 44×44px. Emoji-ikoner ersatta med Lucide-komponenter.

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
- [x] Alla 🔴-uppgifter klara (K1–K5)
- [x] 🟡 V1 klar (kopiera inbjudningskod)
- [x] 🟡 V2 klar (sessionshantering UX)
- [x] 🟡 V3 klar (felhantering audit)
- [x] 🟡 V4 klar (databasbackup)
- [x] 🟡 V5 klar (20 seed-recept)
- [x] Minst 15 seed-recept finns (20 st)
- [x] Fly.io volume snapshots verifierade (dagliga, 5d retention)
- [x] 🟠 M1–M6 mobilpolish klar
- [ ] Testat på riktig telefon (iPhone SE 375px + Android ~390px)
- [ ] Philip gör en komplett QA-genomgång av alla flöden

### Soft launch
- Deploy till `maltiden.fly.dev`
- Bjud in 2–3 familjer manuellt
- Övervaka Fly.io-loggar första 48 timmarna
- Enkel feedback-kanal (Messenger-grupp / mail)

### Efter 2 veckor
- Samla feedback
- Prioritera 🟢-listan baserat på vad användarna faktiskt efterfrågar
- Fixa buggar som dyker upp
