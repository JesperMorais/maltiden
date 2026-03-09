# Phase 4: Polish & Delight — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add skeleton loading states, empty state entrance animation, and remaining micro-interactions polish.

**Architecture:** Reuse existing skeleton primitives (SkeletonBlock, SkeletonCircle) and useSkeleton composable. CSS-only animations. No new dependencies.

**Tech Stack:** Vue 3, TypeScript, CSS variables, Lucide Vue icons

---

## Pre-flight: Already Done (skip these)

Confirmed by code read — do NOT re-implement:
- **Route transitions**: App.vue already has `<RouterView v-slot>` + `<Transition>` with page-fade, page-slide, page-scale CSS
- **Route meta**: All routes in router/index.ts already have `meta.transition` assigned
- **EmptyState Lucide icons**: RecipeListPanel already uses `#icon` slot with BookOpen/Search
- **ShoppingListView empty states**: Already uses Lucide ClipboardList/ShoppingCart via `#icon` slot
- **BaseButton active state**: Already has `:active` scale(0.98)
- **RecipeCard hover**: Already has translateY(-4px) + shadow-lg
- **MenuDayCard hover**: Already has translateY(-2px) + shadow-md
- **Shopping list checkbox**: Already has checkbox-pop animation, scale transitions, is-checked states
- **RecipeDetailModal loading**: Already uses Lucide Loader2 (replaced emoji in full-sweep plan)

---

### Task 1: Create GenerateMenuSkeleton layout

**Files:**
- Create: `frontend/src/components/skeleton/layouts/GenerateMenuSkeleton.vue`

**Step 1: Create the skeleton component**

Mirror the GenerateMenuView layout: 5-column grid of day-card skeletons matching MenuDayCard (20px radius, 280px min-height, emoji circle + title + subtitle).

```vue
<script setup lang="ts">
import SkeletonBlock from '../SkeletonBlock.vue'
import SkeletonCircle from '../SkeletonCircle.vue'
</script>

<template>
  <div class="generate-menu-skeleton" aria-hidden="true">
    <div class="menu-grid-skeleton">
      <div v-for="i in 5" :key="i" class="day-card-skeleton">
        <SkeletonBlock width="48px" height="12px" radius="6px" />
        <SkeletonCircle size="64px" />
        <SkeletonBlock width="80%" height="18px" radius="8px" />
        <SkeletonBlock width="60%" height="12px" radius="6px" />
      </div>
    </div>
    <!-- Action bar skeleton -->
    <div class="actions-skeleton">
      <SkeletonBlock width="90px" height="40px" radius="100px" />
      <SkeletonBlock width="140px" height="14px" radius="8px" />
      <div class="actions-right-skeleton">
        <SkeletonBlock width="130px" height="40px" radius="100px" />
        <SkeletonBlock width="110px" height="40px" radius="100px" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.generate-menu-skeleton {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.menu-grid-skeleton {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1.5rem;
}

.day-card-skeleton {
  background: var(--bg-card);
  border: 2px solid var(--border-color);
  border-radius: 20px;
  padding: 1.5rem;
  min-height: 280px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.actions-skeleton {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  box-shadow: var(--shadow-sm);
}

.actions-right-skeleton {
  display: flex;
  gap: 0.75rem;
}

@media (max-width: 1024px) {
  .menu-grid-skeleton {
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .menu-grid-skeleton {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }

  .day-card-skeleton {
    min-height: 220px;
    padding: 1rem;
  }

  .actions-skeleton {
    grid-template-columns: 1fr;
    justify-items: center;
    gap: 0.75rem;
  }

  .actions-right-skeleton {
    width: 100%;
    justify-content: center;
  }
}
</style>
```

**Step 2: Commit**

```bash
git add frontend/src/components/skeleton/layouts/GenerateMenuSkeleton.vue
git commit -m "feat: add GenerateMenuSkeleton layout component"
```

---

### Task 2: Wire GenerateMenuSkeleton into GenerateMenuView

**Files:**
- Modify: `frontend/src/views/GenerateMenuView.vue:1-12` (script imports) and `204-210` (template)

**Step 1: Add imports**

Add to the script imports (after the existing imports around line 11):
```ts
import GenerateMenuSkeleton from '@/components/skeleton/layouts/GenerateMenuSkeleton.vue'
import { useSkeleton } from '@/composables/useSkeleton'
```

**Step 2: Add skeleton state**

After line 31 (`const isLoading = computed(() => store.isLoading)`), add:
```ts
const { showSkeleton } = useSkeleton(
  computed(() => store.isGenerating && !slotMachine.isAnimating.value),
  { minDuration: 400 }
)
```

**Step 3: Add skeleton to template**

Before the `<GenerateMenuEmptyState>` line (around line 208), add the skeleton:
```html
        <!-- Skeleton loading state -->
        <GenerateMenuSkeleton v-if="showSkeleton" />

        <!-- Empty state -->
        <GenerateMenuEmptyState v-else-if="!showGrid" @generate="handleInitialGenerate" />
```

**Step 4: Commit**

```bash
git add frontend/src/views/GenerateMenuView.vue
git commit -m "feat: wire skeleton loading state into GenerateMenuView"
```

---

### Task 3: Add skeleton loading to RecipeDetailModal

**Files:**
- Modify: `frontend/src/components/recipes/RecipeDetailModal.vue:9,216-219`

**Step 1: Add skeleton imports**

Add after the existing Lucide imports (line 9):
```ts
import SkeletonBlock from '@/components/skeleton/SkeletonBlock.vue'
import SkeletonCircle from '@/components/skeleton/SkeletonCircle.vue'
```

**Step 2: Replace loading spinner with skeleton**

Replace lines 216-219 (the `<div v-if="isLoading" class="modal-loading">` block):

```html
        <div v-if="isLoading" class="modal-loading">
          <div class="modal-header">
            <SkeletonCircle size="64px" />
            <SkeletonBlock width="70%" height="24px" radius="12px" />
            <SkeletonBlock width="80px" height="14px" radius="8px" />
            <div class="skeleton-tags">
              <SkeletonBlock width="52px" height="22px" radius="100px" />
              <SkeletonBlock width="64px" height="22px" radius="100px" />
            </div>
          </div>
          <div class="modal-body">
            <SkeletonBlock width="110px" height="16px" radius="8px" />
            <div class="skeleton-list">
              <SkeletonBlock v-for="n in 5" :key="n" width="100%" height="14px" radius="6px" />
            </div>
            <SkeletonBlock width="110px" height="16px" radius="8px" />
            <div class="skeleton-list">
              <SkeletonBlock v-for="n in 3" :key="n" width="100%" height="14px" radius="6px" />
            </div>
          </div>
        </div>
```

**Step 3: Add skeleton CSS**

Add to the `<style scoped>` section (after the existing `.modal-loading` styles):

```css
.skeleton-tags {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
  margin-top: 0.5rem;
}

.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 1rem 0 1.5rem;
}
```

**Step 4: Remove old spinner CSS**

Remove `.spinner-emoji` styles if they exist (replaced by skeleton).

**Step 5: Commit**

```bash
git add frontend/src/components/recipes/RecipeDetailModal.vue
git commit -m "feat: replace RecipeDetailModal loading spinner with skeleton"
```

---

### Task 4: Add entrance animation to EmptyState component

**Files:**
- Modify: `frontend/src/components/common/EmptyState.vue:32-35` (style)

**Step 1: Add animation to empty-state class**

In the `.empty-state` CSS rule, add the animation:

```css
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  animation: empty-state-enter 0.4s ease-out;
}

@keyframes empty-state-enter {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
```

This is already covered by the global `prefers-reduced-motion` rule in theme.css (forces `animation-duration: 0.01ms`).

**Step 2: Commit**

```bash
git add frontend/src/components/common/EmptyState.vue
git commit -m "feat: add fade-up entrance animation to EmptyState"
```

---

### Task 5: Verify everything passes

**Step 1: Type check**

```bash
cd frontend && npm run type-check
```

**Step 2: Lint**

```bash
npm run lint
```

**Step 3: Build**

```bash
npm run build
```

**Step 4: Test**

```bash
npm run test
```

All must pass. Fix any issues before proceeding.

**Step 5: Commit any fixes**

Only if lint/type-check required changes:
```bash
git add -A && git commit -m "fix: resolve type-check and lint issues from Phase 4"
```
