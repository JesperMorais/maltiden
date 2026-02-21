# Måltiden

Svensk receptapp för veckoplanering och smarta inköpslistor.

**Status:** MVP-redo — alla kritiska (K1–K5) och viktiga UX-uppgifter (V1–V3) klara. Kvar: databasbackup (V4), seed-recept (V5), QA-genomgång. Se `docs/TODO.md`.

## Team
- **David** – Lead backend (Go)
- **Jesper** – Lead frontend (Vue)
- **Philip** – Lead tester

## Tech Stack & Versioner

- **Backend:** Go 1.24 (monolith, lager-separation)
- **Frontend:** Vue 3, Vite, TypeScript strict, Node ^20.19.0 || >=22.12.0
- **Databas:** SQLite
- **Hosting:** Fly.io (prod), lokal utveckling
- **Auth:** Google OAuth, Apple Sign In, Email/lösenord
- **AI:** Claude API (receptparsning, kylskåpsscan)

### Versionskontroll — VIKTIGT

Gissa ALDRIG versioner av verktyg eller GitHub Actions. Verifiera ALLTID senaste version innan du skriver workflows, Dockerfiles eller liknande.

**GitHub Actions (kör dessa för att kolla senaste):**
```bash
gh api repos/actions/checkout/releases/latest --jq '.tag_name'
gh api repos/actions/setup-go/releases/latest --jq '.tag_name'
gh api repos/actions/setup-node/releases/latest --jq '.tag_name'
gh api repos/anthropics/claude-code-action/releases/latest --jq '.tag_name'
```

**Go & Node:**
```bash
curl -s https://go.dev/dl/?mode=json | head -5    # Senaste Go
head -3 backend/go.mod                              # Vad projektet använder
jq '.engines' frontend/package.json                 # Vad projektet kräver
```

**Senast verifierade (2026-02-04):**
| Verktyg | Senaste | Projektet använder |
|---------|---------|-------------------|
| `actions/checkout` | v6 | v6 |
| `actions/setup-go` | v6 | v6 |
| `actions/setup-node` | v6 | v6 |
| `anthropics/claude-code-action` | v1 | v1 |
| Go | 1.25.6 (1.24.12 stöds) | 1.24 (go.mod) |
| Node | v22.22.0 LTS | ^20.19.0 \|\| >=22.12.0 (package.json) |

## Arkitektur

```
cmd/server/main.go          # Entry point
internal/
├── api/handlers/           # HTTP-lim, ingen logik
├── services/               # Affärslogik, orkestrering
├── storage/sqlite/         # Databasaccess
└── domain/                 # Typer, inga dependencies
pkg/claude/                 # Claude API-klient
migrations/                 # SQL-migreringar
```

**Principer:**
- Handler → Service → Storage (enkelriktad)
- `domain/` är ren – inga DB/HTTP-imports
- Monolith men förberett för separation vid behov

## Git Workflow
- **Branches:**
  - `main` – Production (hostas av GitHub Pages, endast stabil kod)
  - `dev` – Utvecklings-main (default branch för features)
  - `feat/be_<feature>` – Backend features
  - `feat/fe_<feature>` – Frontend features
- **Merge-strategi:** Rebase only (inga merge commits)
- **Workflow:**
  1. Skapa feature branch från `dev`
  2. Utveckla och testa
  3. Rebasea mot `dev`: `git rebase dev`
  4. Merge till `dev`: `git checkout dev && git merge --ff-only feat/be_<feature>`
  5. När `dev` är stabil → merge till `main` för deploy

## Kärnfunktioner (MVP)

**Fungerar idag:**
1. Registrera/logga in/logga ut med JWT-auth och roller (owner/member/guest)
2. Skapa hushåll, bjud in med kod (kopiera till urklipp), gå med
3. Lägg till, redigera och ta bort recept manuellt eller via AI-parsning (Claude)
4. Generera veckomeny (5–7 dagar, skip-dagar, låsa dagar, anpassade portioner)
5. Spara meny med exakta recept-val (PUT /menus/current)
6. Interaktiv inköpslista med kategorier, avkryssning och progress
7. Toast-notifikationer (success/error/info/warning) med svenska meddelanden
8. Sessionshantering — tydligt meddelande vid JWT-utgång

**Kvar innan lansering (se `docs/TODO.md`):**
- Databasbackup (V4)
- Seed-recept (V5)
- QA-genomgång (Philip)

## Post-MVP
- Preferensinlärning (gilla/ogilla)
- Erbjudanden i inköpslista (Tjek API POC finns)
- Näringsbalans (Livsmedelsverkets API)
- Skafferi/inventory
- Kylskåpsscan (Claude Vision)
- PWA/offline-stöd
- Google/Apple OAuth

## Datakällor
- **Recept:** Egen databas (svenska klassiker) – Philip matar in via Claude
- **Näringsvärden:** Livsmedelsverkets API (CC BY 4.0) — ej implementerat ännu
- **Erbjudanden:** Tjek/etilbudsavis.dk API — POC klar, ej integrerad i inköpslista

## Auth-flöde
1. Registrera med email/lösenord (OAuth planerat post-MVP)
2. Skapa hushåll automatiskt
3. Generera inbjudningskod för familjemedlemmar
4. Kod gäller 7 dagar

## User Roles
- **owner** – Full access + kan ta bort hushållet
- **member** – Full access
- **guest** – Endast visning

## Besluttslogg
| Datum | Beslut | Motivering |
|-------|--------|------------|
| 2025-01 | Lager-arkitektur (handler→service→storage) | Balans mellan enkelhet och separation. Undviker överengineering men håller logik testbar. |
| 2025-01 | Egen auth (email först, Google senare) | Full kontroll, lärande, ingen vendor lock-in. Apple Sign In väntar tills iOS-app. |
| 2025-01 | Rolluppdelning David/Jesper | David lead backend, Jesper lead frontend. Parallell utveckling mot gemensamt API-kontrakt. |
| 2025-02 | Fly.io för hela stacken | En plattform, en deploy — Go serverar SPA. Undviker split-hosting och CORS-krångel i prod. |
| 2025-02 | Claude API för receptparsning | Fritext → strukturerad JSON. Graceful degradation — appen fungerar utan API-nyckel. |
| 2025-02 | Tjek API för erbjudanden | Gratis, ingen auth krävs. Svensk täckning via etilbudsavis.dk. |
| 2026-02 | Claude Code CI (auto-review + @claude) | Automatisk kodgranskning på PR:ar, interaktiv hjälp via kommentarer. |

## Kommunikation
- Svenska i all kommunikation
- Var kritisk och ärlig
- Konkret kod > teoretiska förklaringar
