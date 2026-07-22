# Day Card Popover with Matlåda Counter

## Date: 2026-03-02

## Problem

Clicking a day card in the weekly menu either did nothing useful or opened a recipe modal that fails in mock mode. Users need:
1. Quick access to view a recipe
2. Per-day matlåda (lunchbox) count that adjusts portions and shopping list

## Key Insight

The backend already scales shopping list ingredients based on each menu day's `servings` field via `GET /shopping-list?menuId=X`. The `PUT /menus/current` endpoint can update servings per day. So matlåda = extra portions on a day, zero backend changes needed.

## Data Flow

```
User sets 2 matlådor on a day with 4 base portions
→ totalServings = 4 + 2 = 6
→ PUT /menus/current updates day.servings = 6
→ GET /shopping-list recalculates: scale = 6/4 = 1.5×
→ Shopping list shows 50% more ingredients for that day
```

## Components

### DayActionPopover.vue (new)
- Floating popover anchored to selected day card
- Shows: meal emoji + name, "Se recept →" button, matlåda +/- counter
- Counter range: 0 to membersEatingToday.length
- Shows effective total: "X portioner (base + N matlådor)"
- Positioned below card with arrow, click-outside dismisses

### WeeklyMenuGrid.vue (modify)
- Remove inline expand actions (revert to simple card + selection)
- Render DayActionPopover positioned relative to selected card
- Pass day data, max lunchboxes, current count

### dashboard.ts store (modify)
- Change dayLunchBox from Record<string, boolean> to Record<string, number>
- getDayLunchBoxCount(date) / setDayLunchBoxCount(date, count)
- updateDayServings(date, newServings) — calls PUT /menus/current
- membersEatingToday.length as max

### DashboardView.vue (minor)
- Rewire events, init lunchbox state on mount

## What Stays the Same
- HouseholdWidget per-member matlåda toggle — unchanged
- Backend — zero changes needed
- Shopping list view — already shows scaled amounts
