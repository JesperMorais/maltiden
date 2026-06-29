# Smart menu generation (#248) — phase status

Status tracker for [issue #248](https://github.com/davdd1/maltiden/issues/248)
("Smart menu generation: ingredient economy, meal prep & dietary goals").

The issue's full phased plan is the source of truth for *scope*; this doc
records what has actually **landed on `dev`** (or is in an open PR against it)
versus what **remains**. Every "done" line below is grounded in a real commit
or PR and the file(s) it touched.

## Snapshot

| Phase | Theme | Status |
|-------|-------|--------|
| 0 | Data foundation (canonical names, recipe metadata, prefs table) | **Partial** — household preferences + disliked ingredients landed; canonical-ingredient normalization and recipe metadata (mainProtein/dietClass/batchable/cookMinutes columns, backfill) **not started** |
| 1 | Smart generator MVP (pure Go) | **Done** (selector core, overlap, variety + recency, shared-ingredients UX) |
| 2 | Meal prep / batch cooking | **Done** — 2-slot batch occupancy + API/UX types landed; per-week prep-läge toggle now wired in the generator UI |
| 3 | Nutrition & diet goals | **Partial** — recipe nutrition columns + weekly macro aggregation/display landed; soft weekly targets and the Livsmedelsverket import remain |
| 4 | Claude experience layer | **Implemented (PR #267, pending merge)** — wishes → constraints, AI weekday arrange + Swedish rationale, key-optional guardrail |

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

### Phase 2 (done) — batch cooking

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
- **Per-week prep-läge toggle (generator UI).** A switch on the generate view
  drives the `prepMode` request flag; batch/leftover markers map into the draft
  days, render as "Dubbel sats" / "Rester" badges on the day cards, and persist
  through save (`SaveMenuDay` now carries `prepMode`/`leftoverOf`). Markers are
  cleared on a partial (locked-day) regenerate, where week-level batch pairing
  would not cohere.
  - `prepMode` state + mapping in `frontend/src/stores/menuGenerator.ts`,
    toggle + styling in `frontend/src/views/GenerateMenuView.vue`,
    badges in `frontend/src/components/menu/MenuDayCard.vue`,
    `SaveMenuDay` fields in `frontend/src/api/menu.api.ts`,
    mock simulation in `frontend/src/mocks/menu.mock.ts`.

### Phase 3 (partial) — nutrition data layer + weekly aggregation

- **Per-serving recipe nutrition columns.** Four nullable REAL columns
  (`calories`, `protein_g`, `carbs_g`, `fat_g`) and a `Nutrition` domain
  struct with optional-pointer fields.
  - Commit `a87ef4f` feat(recipe): nutrition fields foundation (#263)
  - `backend/migrations/018_add_recipe_nutrition.sql`,
    `domain.Nutrition` in `backend/internal/domain/recipe.go`,
    read/write in `backend/internal/storage/sqlite/recipe_storage.go`.
- **Weekly macro aggregation + display.** Each menu day now carries the
  recipe's per-serving `Nutrition`; the menu response carries a weekly
  `MenuNutrition` total (`weeklyNutrition`) summing per-serving macros × that
  day's servings over cooked days. Leftover days are excluded so a batch dish
  is counted once (its full amount lives on the doubled cook-day); a `Partial`
  flag marks weeks where some recipe lacked data. Degrades gracefully — the
  total is omitted entirely when no recipe carries nutrition. Surfaced as a
  per-portion macro line on each day card and a "Näring för veckan" panel.
  - `weeklyNutrition` + per-day `Nutrition` in
    `backend/internal/services/menu_service.go`, `domain.MenuNutrition` /
    `MenuResponseDay.Nutrition` in `backend/internal/domain/menu.go`
    (covered by `TestWeeklyNutrition`); frontend types in
    `frontend/src/api/menu.api.ts`, panel in `GenerateMenuView.vue`,
    macro line in `MenuDayCard.vue`, store state in `menuGenerator.ts`.

> Note: the seed/existing recipes carry no nutrition data yet (no backfill),
> so the weekly total and per-day macros only appear once recipes are
> populated — the plumbing and UI are complete and exercisable in mock mode.

### Phase 4 (implemented, PR #267 pending merge) — Claude/Gemini experience layer

- **Free-text wishes → constraints** and **AI weekday arrangement + Swedish
  rationale**, on top of the deterministic Phase 0–3 generator. The algorithm
  still picks the week in pure Go; the AI layer only parses wishes into
  per-run constraints and permutes the chosen recipes across weekdays with a
  short rationale. Strict guardrail: the arranger output must be a permutation
  of exactly the offered slots; it never does quantity math. **Degrades
  gracefully without `GEMINI_API_KEY`** — the menu still generates, wishes are
  reported as ignored, no prose. Adapted to dev's preference model: a
  prep-mode wish routes to the request `PrepMode` flag; the protein-target
  wish is dropped (dev stores nutrition per-recipe, not as a preference target).
  - PR #267 (`feat/issue-248-smart-menu-experience`), commit `7399b4e`.

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

### Phase 3 — nutrition feature (aggregation/display landed; goals + data did not)

- **Soft weekly targets as penalty terms** in the greedy score (e.g.
  high-protein profile) — not implemented. Needs a household nutrition-target
  preference, which dev does not yet store.
- **Livsmedelsverket import + ingredient → `livsmedelsnummer` matching** —
  not started (no local food-composition table, no matcher). This is the
  source that would actually populate recipe nutrition at scale; until then
  the macro display only shows for hand-entered values.
- **Per-serving macros on recipe-detail / dashboard menu views** — the
  generate view now shows them, but the recipe-detail and dashboard menu
  surfaces still do not read the nutrition fields.

### Phase 4 — merge

- Land **PR #267** onto `dev` (CI green, mergeable). Update this doc's snapshot
  from "pending merge" to "Done" once merged, and delete the redundant local
  `phase4-onto-dev` branch (same single commit as the PR head).

## Migrations landed for #248

| Migration | Adds |
|-----------|------|
| `016_create_menu_preferences.sql` | `menu_preferences` table (excluded tags, defaults, vegetarian days) |
| `017_add_disliked_ingredients.sql` | `disliked_ingredients` column on `menu_preferences` |
| `018_add_recipe_nutrition.sql` | `calories` / `protein_g` / `carbs_g` / `fat_g` on `recipes` |
| `019_add_menu_day_batch_cooking.sql` | `prep_mode` / `leftover_of` on `menu_days` |
