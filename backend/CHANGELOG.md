# Changelog

All notable changes to the Måltiden Go backend will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Smart menu generation (issue #248):
  - Per-household menu preferences store: `menu_preferences` table (excluded tags, default days/servings, vegetarian-days quota) with `GET`/`PUT /menus/preferences` endpoints.
  - Deterministic menu selector core replacing the old random shuffle: hard tag filtering, vegetarian-days quota, greedy slot filling (seedable RNG, fully unit-testable).
  - Greedy ingredient-overlap scoring — candidates rewarded for sharing non-staple ingredients with the week's picks (pantry staples excluded) to shrink the shopping list.
  - Variety and in-week recency penalties — repeating a cuisine/protein, or repeating one too soon within the week, costs score so weeks stay varied and well spaced.
  - Disliked-ingredients hard filter — recipes containing any household-disliked ingredient are dropped before scoring (`disliked_ingredients` column on `menu_preferences`).
  - Batch-cooking 2-slot occupancy — batchable cook-days double their servings and reuse the dish on the next day as a leftovers slot ("laga en gång, ät två gånger").
  - Recipe nutrition fields foundation — nullable per-serving `calories`, `protein_g`, `carbs_g`, `fat_g` columns and a `Nutrition` domain struct (data layer only; no aggregation yet).
  - Shared-ingredients UX support — the generate response surfaces the week's shared non-staple ingredients for display.
  - Migrations 016 (menu preferences), 017 (disliked ingredients), 018 (recipe nutrition), 019 (menu-day batch cooking).

## [0.1.0] - 2026-05-07

Initial backend release.

### Added

- JWT-based authentication with bcrypt password hashing (registration, login, 7-day HS256 tokens).
- Household management: creation on registration, invite codes, member roles (`owner`, `member`, `guest`).
- Recipe CRUD with ingredients and instructions stored as JSON columns.
- AI-powered recipe parser using the Claude API to convert unstructured text into structured recipes (graceful degradation when `ANTHROPIC_API_KEY` is unset).
- Weekly menu generation and retrieval/update of the current household menu.
- Shopping list derived from the active menu, with per-item state updates.
- Tjek API integration for fetching Swedish grocery store offers and discounts (no auth required).
- User feedback submission endpoint.
- SQLite storage with sequential SQL migrations auto-applied on startup.
- HTTP middleware for auth, CORS, request IDs, and rate limiting.
