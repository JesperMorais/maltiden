# Full Sweep Improvements — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all remaining CSS token misses, component duplication, accessibility gaps, mobile polish, and emoji-as-UI-icons across the frontend.

**Architecture:** Sequential wave approach — CSS tokens first (fixes foundations for dark mode), then component dedup (reduces files to touch), accessibility, mobile polish, emoji cleanup. One commit per wave.

**Tech Stack:** Vue 3, TypeScript, CSS variables, Lucide Vue icons

---

## Pre-flight: Items already fixed (skip these)

Confirmed by code read — do NOT re-fix:
- **M3**: `RecipeParseInput.vue:134-138` already has `@media (max-width: 480px) { .textarea { min-height: 180px; } }`
- **M4**: `TodaysMeal.vue:283-355` already has full mobile responsive with `min-height: 0`, `padding: 1rem`
- **M5**: `RecipeEditForm.vue:438-483` already has `@media (max-width: 480px)` wrapping
- **M6**: `DashboardHeader.vue:330-334` already has `min-height: 44px; min-width: 44px`
- **M6**: `BaseThemeToggle.vue:19-20` already has `width: 44px; height: 44px`
- **M2**: `FeaturesSection.vue:81-84` already uses `grid-template-columns: 1fr` (not `minmax(280px, 1fr)`)

---

### Task 1: Add missing CSS tokens to theme.css

**Files:**
- Modify: `frontend/src/styles/theme.css:36-39` (`:root` block) and `140-142` (`[data-theme="dark"]` block)

**Step 1: Add tokens to light theme**

In `theme.css`, after `--border-color-hover` (line 38), add:
```css
  --border-color-light: rgba(61, 44, 41, 0.05);
  --tab-hover-bg: rgba(255, 255, 255, 0.5);
  --tab-active-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
```

**Step 2: Add tokens to dark theme**

In `theme.css`, inside `[data-theme="dark"]`, after `--border-color-hover` (line 142), add:
```css
  --border-color-light: rgba(255, 255, 255, 0.05);
  --tab-hover-bg: rgba(255, 255, 255, 0.08);
  --tab-active-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
```

---

### Task 2: Replace raw rgba in RecipesView

**Files:**
- Modify: `frontend/src/views/RecipesView.vue:146,152`

**Step 1: Replace tab hover background**

Line 146: `background: rgba(255, 255, 255, 0.5);` → `background: var(--tab-hover-bg);`

**Step 2: Replace tab active shadow**

Line 152: `box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);` → `box-shadow: var(--tab-active-shadow);`

---

### Task 3: Fix undefined --border-color-light fallback in ShoppingListView

**Files:**
- Modify: `frontend/src/views/ShoppingListView.vue:294`

**Step 1: Remove rgba fallback**

Line 294: `border-bottom: 1px solid var(--border-color-light, rgba(0, 0, 0, 0.05));`
→ `border-bottom: 1px solid var(--border-color-light);`

**Step 2: Commit Wave 1**

```
git add frontend/src/styles/theme.css frontend/src/views/RecipesView.vue frontend/src/views/ShoppingListView.vue
git commit -m "fix: add missing CSS tokens and replace raw rgba values"
```

---

### Task 4: Replace custom back-link in ShoppingListView with BackLink component

**Files:**
- Modify: `frontend/src/views/ShoppingListView.vue:1-9` (script) and `52-55` (template) and `172-194` (styles)

**Step 1: Add BackLink import**

In `<script setup>`, add import: `import BackLink from '@/components/common/BackLink.vue'`

**Step 2: Replace template**

Replace lines 52-55:
```html
        <button class="back-link" @click="router.push({ name: 'dashboard' })">
          <span class="back-arrow">&larr;</span>
          <span>Dashboard</span>
        </button>
```
With:
```html
        <BackLink :to="{ name: 'dashboard' }" label="Dashboard" />
```

**Step 3: Remove unused CSS**

Delete `.back-link`, `.back-link:hover`, `.back-arrow` rules (lines 172-194). Remove `useRouter` import if no other usages remain (check — `router` is still used elsewhere in template for navigation actions, so keep it).

---

### Task 5: Replace custom back-link in AboutView with BackLink component

**Files:**
- Modify: `frontend/src/views/AboutView.vue:2,19-23` (script + template) and associated CSS

**Step 1: Add BackLink import**

Replace `import { RouterLink } from 'vue-router'` with:
```ts
import BackLink from '@/components/common/BackLink.vue'
```
(Keep `RouterLink` import only if used elsewhere in the file.)

**Step 2: Replace template**

Replace:
```html
      <RouterLink v-prefetch="'landing'" to="/" class="back-link">
        <span class="back-arrow">←</span>
        <span>Tillbaka</span>
      </RouterLink>
```
With:
```html
      <BackLink to="/" label="Tillbaka" />
```

**Step 3: Remove CSS for custom back-link**

Delete the `.back-link` and `.back-arrow` scoped style rules.

**Step 4: Commit Wave 2**

```
git add frontend/src/views/ShoppingListView.vue frontend/src/views/AboutView.vue
git commit -m "fix: replace custom back-link markup with BackLink component"
```

---

### Task 6: Add aria-label to modal close buttons

**Files:**
- Modify: `frontend/src/components/dashboard/InviteModal.vue:68`
- Modify: `frontend/src/components/dashboard/SettingsModal.vue:59`
- Modify: `frontend/src/components/recipes/RecipeDetailModal.vue:301`

**Step 1: InviteModal close button**

Line 68: `<button class="close-btn" @click="emit('close')">`
→ `<button class="close-btn" aria-label="Stäng" @click="emit('close')">`

**Step 2: SettingsModal close button**

Line 59: `<button class="close-btn" @click="handleClose">`
→ `<button class="close-btn" aria-label="Stäng" @click="handleClose">`

**Step 3: RecipeDetailModal day-picker close button**

Line 301: `<button class="day-picker-close" @click="showDayPicker = false">&times;</button>`
→ `<button class="day-picker-close" aria-label="Stäng" @click="showDayPicker = false">&times;</button>`

---

### Task 7: Add aria-labels to DayActionPopover buttons

**Files:**
- Modify: `frontend/src/components/dashboard/DayActionPopover.vue:77-88,103-117`

**Step 1: Member toggle buttons**

Lines 77-88, each `<button>` for member avatars — add `aria-label`:
```html
        <button
          v-for="member in members"
          :key="member.id"
          class="member-avatar"
          :class="[
            member.role,
            { eating: eatingIds.has(member.id), 'not-eating': !eatingIds.has(member.id) },
          ]"
          :title="member.name"
          :aria-label="`${member.name}: ${eatingIds.has(member.id) ? 'äter, klicka för att ta bort' : 'äter inte, klicka för att lägga till'}`"
          @click="emit('toggle-member', member.id)"
        >
```

**Step 2: Counter buttons**

Line 103-108 (minus button):
```html
        <button
          class="counter-btn"
          :disabled="lunchBoxCount <= 0"
          aria-label="Minska antal matlådor"
          @click="emit('update-lunchbox', lunchBoxCount - 1)"
        >
```

Lines 111-117 (plus button):
```html
        <button
          class="counter-btn"
          :disabled="lunchBoxCount >= maxLunchBoxes"
          aria-label="Öka antal matlådor"
          @click="emit('update-lunchbox', lunchBoxCount + 1)"
        >
```

---

### Task 8: Make TodaysMeal keyboard-activatable

**Files:**
- Modify: `frontend/src/components/dashboard/TodaysMeal.vue:21`

**Step 1: Add role, tabindex, and keyboard handler**

Line 21: `<article class="todays-meal" @click="emit('click')">`
→ `<article class="todays-meal" role="button" tabindex="0" aria-label="Dagens måltid" @click="emit('click')" @keydown.enter="emit('click')" @keydown.space.prevent="emit('click')">`

---

### Task 9: Add ARIA tab roles to RecipesView

**Files:**
- Modify: `frontend/src/views/RecipesView.vue:35-50`

**Step 1: Add tablist and tab roles**

Replace the tab-control div and buttons:
```html
        <div class="tab-control" role="tablist">
          <button
            class="tab-button"
            :class="{ active: activeTab === 'list' }"
            role="tab"
            :aria-selected="activeTab === 'list'"
            :tabindex="activeTab === 'list' ? 0 : -1"
            @click="switchToList"
          >
            Mina recept
          </button>
          <button
            class="tab-button"
            :class="{ active: activeTab === 'add' }"
            role="tab"
            :aria-selected="activeTab === 'add'"
            :tabindex="activeTab === 'add' ? 0 : -1"
            @click="switchToAdd"
          >
            Lägg till
          </button>
        </div>
```

---

### Task 10: Add aria-hidden to recipe emoji in RecipeCard and MenuDayCard

**Files:**
- Modify: `frontend/src/components/recipes/RecipeCard.vue:23`
- Modify: `frontend/src/components/menu/MenuDayCard.vue:79-80`

**Step 1: RecipeCard**

Line 23: `<div class="recipe-emoji">{{ recipe.emoji || '🍽️' }}</div>`
→ `<div class="recipe-emoji" aria-hidden="true">{{ recipe.emoji || '🍽️' }}</div>`

**Step 2: MenuDayCard**

Line 79-80: `<div class="recipe-emoji" :class="{ 'landing-bounce': hasLanded }">`
→ `<div class="recipe-emoji" aria-hidden="true" :class="{ 'landing-bounce': hasLanded }">`

**Step 3: Commit Wave 3**

```
git add frontend/src/components/dashboard/InviteModal.vue frontend/src/components/dashboard/SettingsModal.vue frontend/src/components/recipes/RecipeDetailModal.vue frontend/src/components/dashboard/DayActionPopover.vue frontend/src/components/dashboard/TodaysMeal.vue frontend/src/views/RecipesView.vue frontend/src/components/recipes/RecipeCard.vue frontend/src/components/menu/MenuDayCard.vue
git commit -m "fix: add missing aria-labels, keyboard nav, and ARIA tab roles"
```

---

### Task 11: Fix M1 WCAG contrast — audit and apply --accent-text

**Files:**
- Audit all files using `var(--accent)` as text color on light backgrounds
- The main token `--accent-text: #c4402e` (4.75:1) is defined but may not be applied everywhere

**Step 1: Search for contrast issues**

Search for `color: var(--accent)` in components where it's used as text on cream/white backgrounds. The key places to check and fix:
- Any label/badge text that uses `--accent` instead of `--accent-text` on light bg
- Verify `HouseholdWidget.vue` invite button uses `--accent-text` or has sufficient contrast

This task requires a targeted audit — grep for `color: var(--accent)` and check each context.

**Step 2: Commit Wave 4**

```
git add [affected files]
git commit -m "fix: apply --accent-text for WCAG AA contrast on light backgrounds"
```

---

### Task 12: Replace emoji loader in LandingView with Lucide icons

**Files:**
- Modify: `frontend/src/views/LandingView.vue:1-2,41-57`

**Step 1: Add Lucide import**

Add to script: `import { Loader2, AlertTriangle } from 'lucide-vue-next'`

**Step 2: Replace loading emoji**

Lines 41-46, replace:
```html
    <div v-if="landingStore.isLoading" class="loading-state" role="status" aria-label="Laddar sidan">
      <div class="loader">
        <span class="loader-icon" aria-hidden="true">🍳</span>
        <p class="loader-text">Laddar...</p>
      </div>
    </div>
```
With:
```html
    <div v-if="landingStore.isLoading" class="loading-state" role="status" aria-label="Laddar sidan">
      <div class="loader">
        <Loader2 :size="48" :stroke-width="2" class="loader-icon" aria-hidden="true" />
        <p class="loader-text">Laddar...</p>
      </div>
    </div>
```

**Step 3: Replace error emoji**

Lines 49-57, replace:
```html
    <div v-else-if="landingStore.error" class="error-state">
      <div class="error-content">
        <span class="error-icon" aria-hidden="true">😅</span>
        <h2>Något gick fel</h2>
```
With:
```html
    <div v-else-if="landingStore.error" class="error-state">
      <div class="error-content">
        <AlertTriangle :size="48" :stroke-width="1.75" class="error-icon" aria-hidden="true" />
        <h2>Något gick fel</h2>
```

**Step 4: Update CSS**

Change `.loader-icon` from `font-size: 4rem` to:
```css
.loader-icon {
  color: var(--accent);
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
```

Change `.error-icon` from `font-size: 4rem` to:
```css
.error-icon {
  color: var(--warning);
  display: block;
  margin-bottom: 1rem;
}
```

Remove the old `@keyframes bounce`.

---

### Task 13: Replace emoji in RecipeDetailModal with Lucide icons

**Files:**
- Modify: `frontend/src/components/recipes/RecipeDetailModal.vue:4,215-217,225-226`

**Step 1: Add Lucide imports**

Add to existing imports: `import { Loader2, AlertTriangle } from 'lucide-vue-next'`

**Step 2: Replace loading spinner**

Lines 215-217:
```html
          <div class="spinner-emoji">🍳</div>
```
→
```html
          <Loader2 :size="48" :stroke-width="2" class="spinner-icon" aria-hidden="true" />
```

**Step 3: Replace warning icon in confirm delete**

Line 226:
```html
            <div class="confirm-icon">⚠️</div>
```
→
```html
            <div class="confirm-icon"><AlertTriangle :size="48" :stroke-width="1.75" /></div>
```

**Step 4: Update CSS**

Replace `.spinner-emoji` styles with:
```css
.spinner-icon {
  color: var(--accent);
  animation: spin 1.5s linear infinite;
}
```

Update `.confirm-icon`:
```css
.confirm-icon {
  color: var(--warning);
  margin-bottom: 1rem;
  display: flex;
  justify-content: center;
}
```

---

### Task 14: Replace trash emoji in RecipeEditForm with Lucide Trash2

**Files:**
- Modify: `frontend/src/components/recipe-parser/RecipeEditForm.vue:2,168,187`

**Step 1: Add Lucide import**

Add: `import { Trash2 } from 'lucide-vue-next'`

**Step 2: Replace ingredient remove button**

Line 168: `<button class="remove-btn" @click="removeIngredient(i)" title="Ta bort">🗑️</button>`
→ `<button class="remove-btn" @click="removeIngredient(i)" aria-label="Ta bort ingrediens"><Trash2 :size="16" /></button>`

**Step 3: Replace instruction remove button**

Line 187: `<button class="remove-btn" @click="removeInstruction(i)" title="Ta bort">🗑️</button>`
→ `<button class="remove-btn" @click="removeInstruction(i)" aria-label="Ta bort steg"><Trash2 :size="16" /></button>`

---

### Task 15: Replace emoji strings in RecipeListPanel with Lucide icon slots

**Files:**
- Modify: `frontend/src/components/recipes/RecipeListPanel.vue:2,118,158-165,168-173`

**Step 1: Add Lucide imports**

Add: `import { AlertTriangle, Search, BookOpen } from 'lucide-vue-next'`

**Step 2: Replace ErrorState emoji**

Line 118: `<ErrorState v-if="error" icon="⚠️" :description="error" @retry="fetchRecipes" />`
→
```html
    <ErrorState v-if="error" :description="error" @retry="fetchRecipes">
      <template #icon><AlertTriangle :size="48" :stroke-width="1.75" color="var(--warning)" /></template>
    </ErrorState>
```

Wait — ErrorState doesn't have an `#icon` slot. It uses an `icon` prop rendered as text. We need to check if we should modify ErrorState itself.

Actually, looking at the ErrorState component:
```html
<span class="error-state-icon" aria-hidden="true">{{ icon ?? '😅' }}</span>
```

It just renders a string. We should add slot support like EmptyState has. Let me adjust the plan.

**Step 2a: Add icon slot to ErrorState component**

Modify `frontend/src/components/common/ErrorState.vue:25`:
```html
    <div class="error-state-icon">
      <slot name="icon">
        <span v-if="icon" aria-hidden="true">{{ icon }}</span>
        <AlertTriangle v-else :size="48" :stroke-width="1.75" />
      </slot>
    </div>
```

Add Lucide import to ErrorState: `import { AlertTriangle } from 'lucide-vue-next'`

This also replaces the default 😅 emoji with a proper Lucide icon as the fallback.

**Step 2b: Update RecipeListPanel ErrorState**

Line 118: `<ErrorState v-if="error" icon="⚠️" :description="error" @retry="fetchRecipes" />`
→ `<ErrorState v-if="error" :description="error" @retry="fetchRecipes" />`

(Remove `icon="⚠️"` — the default AlertTriangle is now correct.)

**Step 3: Replace EmptyState emoji props**

Lines 158-165 (no recipes):
```html
      <EmptyState
        v-else-if="!recipes.length"
        icon="📖"
        title="Inga recept ännu"
```
→
```html
      <EmptyState
        v-else-if="!recipes.length"
        title="Inga recept ännu"
        description="Börja med att lägga till ditt första recept."
        action-label="Lägg till recept"
        @action="emit('navigate-to-add')"
      >
        <template #icon><BookOpen :size="48" :stroke-width="1.75" color="var(--text-muted)" /></template>
      </EmptyState>
```

Lines 168-173 (no search results):
```html
      <EmptyState
        v-else
        icon="🔍"
        title="Inga träffar"
        description="Försök med andra sökord eller ta bort filter."
      />
```
→
```html
      <EmptyState
        v-else
        title="Inga träffar"
        description="Försök med andra sökord eller ta bort filter."
      >
        <template #icon><Search :size="48" :stroke-width="1.75" color="var(--text-muted)" /></template>
      </EmptyState>
```

**Step 4: Also update RecipeDetailModal ErrorState**

Line 221: `<ErrorState v-else-if="error" icon="⚠️" title="Kunde inte ladda receptet" :show-retry="false" />`
→ `<ErrorState v-else-if="error" title="Kunde inte ladda receptet" :show-retry="false" />`

**Step 5: Commit Wave 5**

```
git add frontend/src/views/LandingView.vue frontend/src/components/recipes/RecipeDetailModal.vue frontend/src/components/recipe-parser/RecipeEditForm.vue frontend/src/components/recipes/RecipeListPanel.vue frontend/src/components/common/ErrorState.vue
git commit -m "fix: replace remaining UI emoji with Lucide icons"
```

---

### Task 16: Verify and run type-check + lint

**Step 1: Run type-check**

```bash
cd frontend && npm run type-check
```

**Step 2: Run lint**

```bash
npm run lint
```

**Step 3: Fix any issues**

Fix type errors or lint issues from the changes.

**Step 4: Run build**

```bash
npm run build
```

Ensure production build succeeds.
