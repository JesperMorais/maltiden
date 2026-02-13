# Optimistic UI & Progress Bar Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add optimistic UI for instant-feel user actions and simulated progress bars for long-running operations (issue #14 tasks 2 & 3).

**Architecture:** Two new composables (`useOptimistic`, `useProgressBar`) following the existing composable patterns in `src/composables/`. One new component (`ProgressBar.vue`). Integrated into HouseholdWidget (member status toggles) and GenerateMenuView/AddRecipePanel (progress bars).

**Tech Stack:** Vue 3 Composition API, TypeScript strict mode, Pinia, existing API layer

---

### Task 1: Create `useOptimistic` composable

**Files:**
- Create: `frontend/src/composables/useOptimistic.ts`

**Step 1: Create the composable**

```ts
import { type Ref } from 'vue'

interface UseOptimisticOptions {
  /** Called when the API call fails after optimistic update */
  onError?: (error: unknown) => void
}

/**
 * Composable for optimistic UI updates.
 *
 * Applies a local state change immediately, then fires an async API call
 * in the background. Rolls back the change if the API call fails.
 *
 * @example
 * const { execute, isPending } = useOptimistic<boolean>({
 *   onError: () => console.error('Failed to update')
 * })
 *
 * // In a click handler:
 * execute(
 *   isEatingRef,        // ref to mutate
 *   !isEatingRef.value, // new value
 *   (newVal) => api.updateStatus(id, { isEatingToday: newVal })
 * )
 */
export function useOptimistic<T>(options: UseOptimisticOptions = {}) {
  const { onError } = options

  let pendingCount = 0

  function execute(
    target: Ref<T>,
    newValue: T,
    apiCall: (value: T) => Promise<unknown>
  ): void {
    const previousValue = target.value
    target.value = newValue
    pendingCount++

    apiCall(newValue)
      .catch((error: unknown) => {
        target.value = previousValue
        onError?.(error)
      })
      .finally(() => {
        pendingCount--
      })
  }

  return { execute }
}
```

**Step 2: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: PASS (no errors from this file)

**Step 3: Commit**

```bash
git add frontend/src/composables/useOptimistic.ts
git commit -m "feat: add useOptimistic composable for instant UI feedback"
```

---

### Task 2: Wire optimistic UI into HouseholdWidget for member status toggles

**Files:**
- Modify: `frontend/src/components/dashboard/HouseholdWidget.vue`
- Modify: `frontend/src/stores/dashboard.ts` (to expose mutable member data)

**Step 1: Update dashboard store to support local member mutation**

In `frontend/src/stores/dashboard.ts`, add a `updateMemberLocally` action that updates a member's status in the local dashboardData. This lets HouseholdWidget do optimistic updates through the store.

**Step 2: Add toggle handlers to HouseholdWidget**

- Import `useOptimistic` and `updateMemberStatus` from household API
- Make member rows tappable to toggle `isEatingToday`
- Add a lunchbox toggle icon that toggles `wantsLunchBox`
- Use `useOptimistic` to update the dashboard store member data instantly, then call API in background
- On error, the ref rolls back automatically

**Step 3: Add visual feedback for tappable rows**

- Add cursor pointer and hover effect to member items
- Add a subtle tap animation on status change

**Step 4: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`

**Step 5: Commit**

```bash
git add frontend/src/stores/dashboard.ts frontend/src/components/dashboard/HouseholdWidget.vue
git commit -m "feat: add optimistic member status toggles in HouseholdWidget"
```

---

### Task 3: Create `useProgressBar` composable

**Files:**
- Create: `frontend/src/composables/useProgressBar.ts`

**Step 1: Create the composable**

```ts
import { ref, onUnmounted } from 'vue'

interface UseProgressBarOptions {
  /** Estimated total duration in ms. Controls the speed curve. Default: 10000 */
  duration?: number
}

/**
 * Composable for simulated progress bar animation.
 *
 * Creates a fake progress value that advances quickly at first,
 * then slows down as it approaches ~90%, never reaching 100%
 * until `finish()` is called.
 *
 * @example
 * const { progress, isActive, start, finish } = useProgressBar({ duration: 15000 })
 *
 * async function handleGenerate() {
 *   start()
 *   try {
 *     await store.generateMenu()
 *   } finally {
 *     finish()
 *   }
 * }
 */
export function useProgressBar(options: UseProgressBarOptions = {}) {
  const { duration = 10000 } = options

  const progress = ref(0)
  const isActive = ref(false)

  let animationFrame: number | null = null
  let startTime = 0
  let finishTimeout: ReturnType<typeof setTimeout> | null = null

  function easeProgress(elapsed: number): number {
    // Asymptotic curve: fast start, approaches 90% but never reaches it
    const t = elapsed / duration
    return 90 * (1 - Math.exp(-3 * t))
  }

  function tick() {
    if (!isActive.value) return
    const elapsed = Date.now() - startTime
    progress.value = Math.min(easeProgress(elapsed), 90)
    animationFrame = requestAnimationFrame(tick)
  }

  function start() {
    cleanup()
    progress.value = 0
    isActive.value = true
    startTime = Date.now()
    animationFrame = requestAnimationFrame(tick)
  }

  function finish() {
    if (animationFrame) cancelAnimationFrame(animationFrame)
    animationFrame = null
    progress.value = 100
    finishTimeout = setTimeout(() => {
      isActive.value = false
      progress.value = 0
    }, 400)
  }

  function cleanup() {
    if (animationFrame) cancelAnimationFrame(animationFrame)
    if (finishTimeout) clearTimeout(finishTimeout)
    animationFrame = null
    finishTimeout = null
  }

  onUnmounted(cleanup)

  return { progress, isActive, start, finish }
}
```

**Step 2: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`

**Step 3: Commit**

```bash
git add frontend/src/composables/useProgressBar.ts
git commit -m "feat: add useProgressBar composable for simulated progress"
```

---

### Task 4: Create `ProgressBar.vue` component

**Files:**
- Create: `frontend/src/components/common/ProgressBar.vue`

**Step 1: Create the component**

A thin animated bar matching existing design tokens. Props: `progress` (0-100), `active` (show/hide with transition).

Uses `--accent-gradient` for the fill, smooth CSS transition on width, and a subtle shine animation.

**Step 2: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`

**Step 3: Commit**

```bash
git add frontend/src/components/common/ProgressBar.vue
git commit -m "feat: add ProgressBar component with animated fill"
```

---

### Task 5: Wire progress bar into GenerateMenuView

**Files:**
- Modify: `frontend/src/views/GenerateMenuView.vue`

**Step 1: Replace spinner with progress bar during generation**

- Import `useProgressBar` and `ProgressBar`
- Create a progress bar instance with `duration: 8000` (menu generation is moderately fast)
- Wrap `generateInitialMenu` and `regenerateUnlockedDays` calls: `start()` before, `finish()` in finally
- In the loading overlay, add `<ProgressBar>` below the spinner emoji and loading text
- Keep the existing spinner emoji animation — add progress bar underneath it

**Step 2: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`

**Step 3: Commit**

```bash
git add frontend/src/views/GenerateMenuView.vue
git commit -m "feat: add progress bar to menu generation"
```

---

### Task 6: Wire progress bar into AddRecipePanel

**Files:**
- Modify: `frontend/src/components/recipes/AddRecipePanel.vue`

**Step 1: Add progress bar to recipe parsing loading overlay**

- Import `useProgressBar` and `ProgressBar`
- Create a progress bar instance with `duration: 20000` (Claude parsing is slower)
- Wrap `handleParse` call: `start()` before, `finish()` in finally
- Add `<ProgressBar>` to the existing loading overlay below the loading text

**Step 2: Verify types**

Run: `cd frontend && npx vue-tsc --noEmit`

**Step 3: Commit**

```bash
git add frontend/src/components/recipes/AddRecipePanel.vue
git commit -m "feat: add progress bar to recipe parsing"
```

---

### Task 7: Final verification

**Step 1: Full type check**

Run: `cd frontend && npx vue-tsc --noEmit`

**Step 2: Lint**

Run: `cd frontend && npm run lint`

**Step 3: Build**

Run: `cd frontend && npm run build`
