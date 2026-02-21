# V4 + V5 Execution Plan — Pre-Launch Tasks

**Created:** 2026-02-21
**Purpose:** Execute after context clear with agent team

---

## Overview

Two tasks remain before MVP soft launch:
- **V4:** Database backup strategy (ops)
- **V5:** Seed recipes (data + backend)

These are independent and can run in parallel.

---

## Team Structure

| Agent | Role | Task |
|-------|------|------|
| **Lead** | Coordinate, verify, update docs | Both |
| **Backend** | V4 (backup) + V5 backend (seed migration) | V4 + V5-BE |
| **Frontend** | N/A — no frontend work needed | — |

> Note: This can also be done with just Lead + Backend, or even a single agent.

---

## V4: Database Backup

**Goal:** Ensure SQLite data on Fly.io is backed up daily.

### Steps

1. **Check current Fly.io volume config:**
   ```bash
   fly volumes list -a maltiden
   fly volumes snapshots list <volume-id> -a maltiden
   ```

2. **Verify snapshots are enabled:**
   - Fly.io creates daily volume snapshots by default
   - Confirm this is active for the maltiden app volume
   - If not, enable via `fly volumes update`

3. **Document the backup strategy** in `docs/TODO.md`:
   - Volume snapshot frequency
   - How to restore from snapshot
   - Optional: manual backup command for extra safety:
     ```bash
     fly ssh sftp get /data/maltiden.db ./backups/maltiden-$(date +%Y%m%d).db -a maltiden
     ```

4. **Check off V4 in docs/TODO.md**

### Verification
- Run `fly volumes snapshots list` and confirm at least 1 snapshot exists
- Document the snapshot ID and date

---

## V5: Seed Recipes

**Goal:** New households start with 15-20 Swedish base recipes so they can immediately generate menus.

### Approach Options

**Option A: SQL Migration (recommended)**
- Create `backend/migrations/008_seed_recipes.sql`
- Insert 15-20 classic Swedish recipes as global recipes
- Runs automatically on deploy
- Recipes are shared across all households (existing behavior)

**Option B: Manual entry via app**
- Philip enters recipes via the recipe parser
- No code changes needed
- Slower but gives Philip ownership of recipe quality

**Recommended: Option A** — faster, reproducible, and can be extended later.

### Recipe List (15-20 Swedish classics)

1. Köttfärssås med pasta
2. Pannkakor
3. Pasta Carbonara
4. Kycklinggryta med ris
5. Tacos (fredagstacos)
6. Laxfilé med potatis och dillsås
7. Ärtsoppa med pannkakor
8. Falukorv med stuvade makaroner
9. Fiskpinnar med potatismos
10. Spaghetti Bolognese
11. Korvstroganoff med ris
12. Janssons frestelse
13. Pytt i panna
14. Vegetarisk pasta med pesto
15. Kyckling med currysås och ris
16. Stekt fläsk med löksås
17. Köttbullar med gräddsås, potatis och lingon
18. Ugnsbakad torsk med rotfrukter
19. Chili con carne
20. Tomatsoppa med grillad ost-macka

### Migration Format

Each recipe needs:
- `id`: `rec_` + UUID
- `name`: Swedish name
- `servings`: 4 (default)
- `emoji`: Relevant food emoji
- `tags`: JSON array (e.g., `["vardagsmat", "snabbt", "barn"]`)
- `ingredients`: JSON array of `{name, amount, unit}` objects
- `instructions`: JSON array of step strings
- `created_at`: Current timestamp

**Important:** Look at existing recipe Create storage code to understand the exact JSON column format for `ingredients`, `instructions`, and `tags`.

### Steps

1. **Read existing storage/migration patterns:**
   - `backend/migrations/007_*.sql` — latest migration
   - `backend/internal/storage/sqlite/recipe_storage.go` — Create method for JSON format

2. **Create migration file:**
   - `backend/migrations/008_seed_recipes.sql`
   - INSERT 15-20 recipes with proper JSON columns
   - Use realistic Swedish ingredient amounts and units (dl, msk, tsk, g, st)

3. **Test locally:**
   ```bash
   cd backend
   rm -f data/maltiden.db  # Fresh DB to test migration
   JWT_SECRET=dev-secret go run cmd/server/main.go  # Should apply all migrations
   curl http://localhost:8080/recipes | jq '.recipes | length'  # Should be 15-20
   curl http://localhost:8080/recipes | jq '.recipes[0]'  # Check format
   ```

4. **Verify a recipe loads correctly:**
   ```bash
   curl http://localhost:8080/recipes/<recipe-id> | jq
   ```
   Check that ingredients, instructions, and tags parse correctly.

5. **Check off V5 in docs/TODO.md**

### Verification
- `GET /recipes` returns 15-20 recipes
- Each recipe has valid ingredients with amounts and units
- Each recipe has at least 3 instructions
- Tags are meaningful (vardagsmat, fisk, vegetariskt, etc.)
- Menu generation works with seed recipes (test via UI or API)

---

## Post-Completion

After V4 and V5:
1. Update `docs/TODO.md` — check off V4, V5
2. Update `docs/PROJECT.md` — status to "MVP-redo för soft launch"
3. Update launch checklist — all items checked except Philip's QA
4. Commit and push (feature branch → PR → merge to dev)

---

## Quick Start Command

After context clear, run:
```
Read docs/mvp/V4_V5_PLAN.md and execute both V4 and V5.
Use a backend agent for the seed migration (V5) while lead handles V4 (Fly.io backup verification).
Commit and push after each task. Create PR to dev when done.
```
