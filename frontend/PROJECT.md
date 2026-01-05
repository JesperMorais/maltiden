# Måltiden

Svensk receptapp för veckoplanering och smarta inköpslistor.

## Team
- **David** – Lead backend (Go)
- **Jesper** – Lead frontend (Vue)
- **Philip** – Lead tester

## Tech Stack
- **Backend:** Go (monolith, lager-separation)
- **Frontend:** Vue 3 (PWA)
- **Databas:** SQLite
- **Hosting:** Fly.io (prod), lokal utveckling
- **Auth:** Google OAuth, Apple Sign In, Email/lösenord
- **AI:** Claude API (receptparsning, kylskåpsscan)

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
