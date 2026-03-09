# Phase 4: Polish & Delight — Design Document

**Date:** 2026-03-09
**Branch:** feat/fe_design-review-fixes
**Goal:** Add skeleton loading states, empty state icons, route transitions, and micro-interactions.

---

## 1. Skeleton Loading States

### 1a. GenerateMenuView Skeleton

Add `GenerateMenuSkeleton.vue` in `components/skeleton/layouts/` mirroring the GenerateMenuView layout:
- Action bar skeleton (3 buttons as blocks)
- 5 day-card skeletons in a grid (matching MenuDayCard dimensions: emoji circle + title block + subtitle block)
- Reuse existing `SkeletonBlock` and `SkeletonCircle` primitives
- Wire into GenerateMenuView with `useSkeleton` composable (same pattern as ShoppingListView)

### 1b. RecipeDetailModal Skeleton

Replace the loading spinner inside RecipeDetailModal with inline skeleton blocks:
- Skeleton circle for emoji (80px)
- SkeletonBlock for title, tags, and section headings
- 4-5 SkeletonBlock rows for ingredients/instructions
- Show while `isLoading` is true, already gated by existing v-if

---

## 2. Empty State Lucide Icons + Entrance Animation

### 2a. Replace Emoji Icons

Replace emoji `icon` props on EmptyState usages with Lucide `#icon` slot:

| Location | Current | Replacement (Lucide) |
|----------|---------|---------------------|
| RecipeListPanel (no recipes) | `icon="📖"` | `BookOpen` via `#icon` slot |
| RecipeListPanel (no results) | `icon="🔍"` | `Search` via `#icon` slot |
| GenerateMenuEmptyState | Custom emoji-based | `CalendarDays` or keep custom component |
| ShoppingListWidget | Handled by redesign | Already uses Lucide |

### 2b. Entrance Animation

Add a subtle fade-up entrance to the `EmptyState` component using CSS:
```css
.empty-state {
  animation: empty-state-enter 0.4s ease-out;
}
@keyframes empty-state-enter {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
```

---

## 3. Route Page Transitions

### Implementation

Wrap `<RouterView>` in `App.vue` with Vue's `<Transition>` using dynamic transition name from route meta:

```vue
<RouterView v-slot="{ Component, route }">
  <Transition :name="route.meta.transition || 'page-fade'" mode="out-in">
    <component :is="Component" :key="route.path" />
  </Transition>
</RouterView>
```

### Transition Definitions (global CSS)

**page-fade** (default): opacity only, 200ms
```css
.page-fade-enter-active { transition: opacity 0.2s ease-out; }
.page-fade-leave-active { transition: opacity 0.15s ease-in; }
.page-fade-enter-from, .page-fade-leave-to { opacity: 0; }
```

**page-slide**: translateY + opacity, 250ms
```css
.page-slide-enter-active { transition: all 0.25s ease-out; }
.page-slide-leave-active { transition: all 0.15s ease-in; }
.page-slide-enter-from { opacity: 0; transform: translateY(16px); }
.page-slide-leave-to { opacity: 0; transform: translateY(-8px); }
```

Already respects `prefers-reduced-motion` via global rule in theme.css.

### Route Meta Assignments

| Route | Transition |
|-------|-----------|
| landing, login, register | page-fade |
| dashboard | page-fade |
| recipes, shopping-list, generate-menu, offers | page-slide |
| about | page-fade |

---

## 4. Micro-interactions

### 4a. Button Press Feedback

Add `active` state to BaseButton and accent buttons:
```css
button:active:not(:disabled) {
  transform: scale(0.97);
  transition: transform 0.1s ease;
}
```

### 4b. Card Hover Lift

Add a reusable pattern in theme.css:
```css
.card-interactive {
  transition: transform var(--duration-fast) ease, box-shadow var(--duration-fast) ease;
}
.card-interactive:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}
```

Apply to: RecipeCard, MenuDayCard (non-rolling state), TodaysMeal, QuickActions items.

### 4c. Shopping List Checkbox

Enhance the existing checkbox with a brief fill transition:
- Coral background-color transition on check (150ms)
- Checkmark scale-in (0 to 1, 150ms)
- Text strikethrough with color fade to --text-muted (200ms)

---

## 5. Typography Rhythm — SKIPPED

Existing Fraunces/Nunito pairing works well. Not worth the churn.

---

## Implementation Order

1. **Wave 5** — Skeletons (GenerateMenu + RecipeDetailModal)
2. **Wave 6** — Empty state icons + entrance animation
3. **Wave 7** — Route page transitions
4. **Wave 8** — Micro-interactions (button press, card hover, checkbox)
5. **Wave 9** — Verify: type-check, lint, build, test

One commit per wave.
