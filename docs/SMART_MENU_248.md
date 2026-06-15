# Smart menu generation (#248) — phase status

Status tracker for [issue #248](https://github.com/davdd1/maltiden/issues/248)
("Smart menu generation: ingredient economy, meal prep & dietary goals").

The issue's full phased plan is the source of truth for *scope*; this doc
records what has actually **landed on `dev`** versus what **remains**. Every
"done" line below is grounded in a real commit and the file(s) it touched.

## Snapshot

| Phase | Theme | Status |
|-------|-------|--------|
| 0 | Data foundation (canonical names, recipe metadata, prefs table) | **Partial** — household preferences + disliked ingredients landed; canonical-ingredient normalization and recipe metadata (mainProtein/dietClass/batchable/cookMinutes columns, backfill) **not started** |
| 1 | Smart generator MVP (pure Go) | **Done** (selector core, overlap, variety + recency, shared-ingredients UX) |
| 2 | Meal prep / batch cooking | **Mostly done** — 2-slot batch occupancy + API/UX types landed; per-week prep-läge toggle in the generator UI still pending |
| 3 | Nutrition & diet goals | **Foundation only** — recipe nutrition columns landed; aggregation/balancing/targets (the actual feature) **not started** |
| 4 | Claude experience layer | **Not started** |

## Done

### Phase 0 (partial) — household preferences + disliked ingredients

- **Per-household menu preferences store.** New `menu_preferences` table
  (one row per household: `excluded_tags` JSON, `default_days`,
  `default_servings`, `vegetarian_days`) with `GET`/`PUT
  /menus/preferences` endpoints.
  - Commit `d1b3aee` feat(menu): preference store + selector core (#253)
  - `backend/migrations/016_create_menu_preferences.sql`,
    `backend/internal/storage/sqlite/menu_preferences_storage.go`,
    `backend/internal/api/handlers/menus.go` (`GetPreferences`/`UpdatePreferences`),
    `backend/internal/api/router.go` (`GET`/`PUT /menus/preferences`).
- **Disliked-ingredients hard filter.** `disliked_ingredients` JSON column
  added to `menu_preferences`; any recipe containing a disliked ingredient
  (name-normalized match) is filtered out before scoring.
  - Commit `4b59774` feat(menu): disliked-ingredients hard filter (#262)
  - `backend/migrations/017_add_disliked_ingredients.sql`,
    `filterByDislikedIngredients` in
    `backend/internal/services/menu_selector.go`.

> Note: this is the *household preferences* bullet of Phase 0. The rest of
> Phase 0 (canonical ingredient names, pantry/perishable flags, recipe
> metadata columns, Claude-driven backfill) has **not** landed — see
> Remaining.

### Phase 1 (done) — smart generator MVP

- **Selector core replacing the old `rand.Shuffle`.** Deterministic,
  I/O-free `menuSelector`: hard-filters excluded tags, honors a
  vegetarian-days quota, then greedily fills slots; unit-testable with a
  seeded RNG.
  - Commit `d1b3aee` feat(menu): preference store + selector core (#253)
  - `backend/internal/services/menu_selector.go` (`newMenuSelector`, `Next`),
    wired into `MenuService.Generate()` in
    `backend/internal/services/menu_service.go`.
- **Greedy ingredient-overlap scoring.** Each candidate is scored by how
  many of its non-staple ingredients are already on the week's shopping
  list; pantry staples (Kryddor / Såser & olja) are excluded so they don't
  dominate.
  - Commit `39f676a` feat(menu): greedy ingredient-overlap scoring (#258)
  - `overlapKeys`, `overlapScore`, `isPantryStaple` in
    `backend/internal/services/menu_selector.go`.
- **Variety + in-week recency penalties.** `score = overlapReward·overlap −
  varietyWeight·varietyPenalty − recencyWeight·recencyPenalty`. Variety
  penalizes repeating a cuisine/protein tag; recency penalizes repeating one
  too soon within the week (linear decay over `recencyDecaySl` slots).
  - Commit `b1b75d9` feat(menu): variety + recency penalties (#261)
  - `varietyTags`, `varietyPenalty`, `recencyPenalty`, `score` in
    `backend/internal/services/menu_selector.go`.
- **Shared-ingredients economy made visible (UX).** Backend computes the
  week's shared non-staple ingredients; frontend surfaces a "delade
  ingredienser" count + chips on the generate view.
  - Commit `a843ece` feat(menu): surface shared-ingredients UX (#260)
  - `computeSharedIngredients` + `domain.SharedIngredient`
    (`backend/internal/services/menu_selector.go`,
    `backend/internal/domain/menu.go`); frontend
    `frontend/src/views/GenerateMenuView.vue`,
    `frontend/src/stores/menuGenerator.ts`,
    `frontend/src/api/menu.api.ts`.

### Phase 2 (mostly done) — batch cooking

- **2-slot batch occupancy ("laga en gång, ät två gånger").** A batchable
  cook-day is doubled in servings (`PrepModeBatch`) and the next eligible
  day becomes a leftovers day (`LeftoverOf` = cook-day's date), front-loaded
  by the left-to-right scan. Deterministic post-pass that never changes
  *which* recipes were chosen.
  - Commit `fea0164` feat(menu): batch-cooking 2-slot occupancy (#264)
  - `applyBatchCooking`, `isBatchable`, `batchTag` in
    `backend/internal/services/menu_selector.go`;
    `domain.PrepModeBatch` / `LeftoverOf` in
    `backend/internal/domain/menu.go`;
    `backend/migrations/019_add_menu_day_batch_cooking.sql`
    (`prep_mode`, `leftover_of` on `menu_days`); `prepMode` request flag +
    `prepMode`/`leftoverOf` day fields in `frontend/src/api/menu.api.ts`.

### Phase 3 (foundation only) — nutrition data layer

- **Per-serving recipe nutrition columns.** Four nullable REAL columns
  (`calories`, `protein_g`, `carbs_g`, `fat_g`) and a `Nutrition` domain
  struct with optional-pointer fields. **Data layer only** — no aggregation
  or balancing logic.
  - Commit `a87ef4f` feat(recipe): nutrition fields foundation (#263)
  - `backend/migrations/018_add_recipe_nutrition.sql`,
    `domain.Nutrition` in `backend/internal/domain/recipe.go`,
    read/write in `backend/internal/storage/sqlite/recipe_storage.go`.

## Remaining

### Phase 0 — data foundation (the prerequisite, mostly outstanding)

- **Canonical ingredient-name normalization.** Overlap and disliked-filter
  matching currently key on raw lower-cased/trimmed names
  (`normalizeIngredientName`, `overlapKeys`), so "kycklinglårfilé" vs
  "kycklingfilé" vs "kyckling" do **not** merge. The issue calls
  normalization "THE prerequisite"; it is **not implemented**. No
  `canonicalName` field, no grams-equivalent, no `isPerishable` flag.
- **Recipe metadata columns.** No `mainProtein`, `dietClass`, `batchable`,
  or `cookMinutes` columns. The selector stands in for these with existing
  tags: `vegetariskt` for the veg quota, `batchcook` for batchability,
  cuisine/protein tags for variety. The explicit Phase 0 fields remain
  pending.
- **Claude parse-pipeline enrichment + one-off backfill** of existing
  recipes — not started.

### Phase 3 — nutrition feature (the data foundation landed; the feature did not)

- **Macro aggregation/balancing** — sum per-serving macros across the week,
  expose menu-level kcal/protein/carbs/fat. Nothing in
  `menu_selector.go` / `menu_service.go` reads the nutrition fields yet.
- **Soft weekly targets as penalty terms** in the greedy score (e.g.
  high-protein profile) — not implemented.
- **Livsmedelsverket import + ingredient → `livsmedelsnummer` matching** —
  not started (no local food-composition table, no matcher).
- **Per-serving macros on recipe + menu views (frontend)** — the frontend
  reads no nutrition fields today.

### Phase 2 — remaining slice

- **Per-week "prep-läge" toggle in the generator UI.** The backend accepts a
  `prepMode` request flag and the API types carry it, but the user-facing
  toggle control in `GenerateMenuView.vue` is not yet wired.

### Phase 4 — Claude experience layer

- Not started: algorithmic week → Claude weekday placement + Swedish
  rationale, free-text wishes → constraints, with the API-key-optional
  guardrail.

## Migrations landed for #248

| Migration | Adds |
|-----------|------|
| `016_create_menu_preferences.sql` | `menu_preferences` table (excluded tags, defaults, vegetarian days) |
| `017_add_disliked_ingredients.sql` | `disliked_ingredients` column on `menu_preferences` |
| `018_add_recipe_nutrition.sql` | `calories` / `protein_g` / `carbs_g` / `fat_g` on `recipes` |
| `019_add_menu_day_batch_cooking.sql` | `prep_mode` / `leftover_of` on `menu_days` |
