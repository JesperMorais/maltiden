# Måltiden

Svensk receptapp för veckoplanering och smarta inköpslistor.

## Team
- **David** – Lead backend (Go)
- **Jesper** – Lead frontend (Vue)
- **Philip** – Lead tester

## Tech Stack & Versioner

- **Backend:** Go 1.24 (monolith, lager-separation)
- **Frontend:** Vue 3, Vite, TypeScript strict, Node >=22.12.0
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
| Node | v22.22.0 LTS | >=22.12.0 (package.json) |

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
1. Generera veckomenyer (5 dagar default)
2. Flexibla dagar (skippa, fler personer, matlådor)
3. Smart inköpslista med kategorier + offline
4. Familje-inbjudan med kod (7 dagar giltighet)

## Iteration 2+
- Preferensinlärning (gilla/ogilla)
- Näringsbalans-varningar
- Skafferi/inventory
- Kylskåpsscan (Claude Vision)

## Datakällor
- **Recept:** Egen databas (svenska klassiker) – Philip matar in via Claude
- **Näringsvärden:** Livsmedelsverkets API (CC BY 4.0)
- **Spoonacular:** Endast realtidssökning, får EJ lagra data

## Auth-flöde
1. Registrera med Google/Apple/Email
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

## Kommunikation
- Svenska i all kommunikation
- Var kritisk och ärlig
- Konkret kod > teoretiska förklaringar
