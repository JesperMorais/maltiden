# TODO - Måltiden

Uppgiftslista för Jesper & David.

---

## Arbetsuppdelning

| Spår | Ansvarig | Teknik |
|------|----------|--------|
| Backend | David | Go, SQLite |
| Frontend | Jesper | Vue 3, TypeScript |
| Test | Philip | Manuell QA |

---

## Fas 1: MVP Grund ✅

### Backend ✅
- [x] Go-projektstruktur med lager-arkitektur (handler → service → storage → domain)
- [x] SQLite + migrations (001–007)
- [x] HTTP-server med router och dependency injection
- [x] Auth-system: register, login, JWT middleware
- [x] Hushållshantering: skapa, bjud in, gå med, roller
- [x] Receptdatabas: CRUD, JSON-lagring av ingredienser/instruktioner
- [x] Menygeneration: slumpa recept, hantera skip-dagar
- [x] Inköpslista: aggregera från meny, toggle checked
- [x] Rate limiting och request ID middleware

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

## Fas 2: Integration & Polish ✅

- [x] Frontend kopplad till riktig backend
- [x] CORS-konfiguration
- [x] Receptparsning via Claude API (parse + parse-and-save)
- [x] Erbjudanden-integration (Tjek API: sök, rabatter, butiker)
- [x] Skeleton loading-komponenter
- [x] Optimistic UI för medlemsstatus-toggles
- [x] Simulerade progress bars för lång-körande operationer
- [x] Animationer med motion-v
- [x] Vue-bits: RotatingText, SpotlightCard, CountUp, etc.

## Fas 3: Deployment ✅

- [x] Fly.io deployment (region: arn)
- [x] Multi-stage Dockerfile (Node build → Go build → slim runtime)
- [x] Miljövariabler konfigurerade (JWT_SECRET, DATABASE_PATH, etc.)
- [x] CI/CD med GitHub Actions (backend + frontend pipelines)
- [x] Claude Code auto-review på PR:ar

---

## Nästa steg (Iteration 2+)

| Feature | Backend | Frontend | Prioritet |
|---------|---------|----------|-----------|
| Preferenser (gilla/ogilla) | Spara i DB, vikta generation | UI för gilla/ogilla | Hög |
| Skafferi/inventory | CRUD för pantry items | Skafferi-vy | Medel |
| Kylskåpsscan | Claude Vision integration | Kamera-vy | Låg |
| Näringsbalans | Livsmedelsverkets API | Visa näringsvärden | Låg |
| Google/Apple OAuth | OAuth-flöde i backend | OAuth-knappar | Medel |
| PWA/Offline-stöd | — | Service worker, offline sync | Medel |

---

**Senast uppdaterad:** 2026-02-16
