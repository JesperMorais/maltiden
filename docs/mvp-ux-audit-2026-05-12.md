# MVP UX Audit — Loading & Error States
**Date:** 2026-05-12  
**Method:** Static read of every file under `frontend/src/views/`, cross-referenced with the stores, composables, and child components they import. For each view the audit asks: (a) which async calls are triggered, (b) what loading feedback exists, (c) what error feedback exists, and (d) what is missing or inconsistent.

---

## Per-View Analysis

### `DashboardView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `dashboardStore.fetchDashboard()` on mount; `dashboardStore.swapRecipeForDay()` on recipe swap; `removeMember()` direct API call |
| **Loading UX** | **Full-page skeleton** — `<DashboardSkeleton v-if="dashboardStore.isLoading" />`. Removes the entire content area and shows skeleton layout. |
| **Error UX** | **Full-page error state** — `<ErrorState>` with store's `error` string and a "Försök igen" retry button wired to `fetchDashboard(true)`. |
| **Inline mutation feedback** | Swap recipe: `toast.success` / `toast.error`. Remove member: optimistic update (snapshot + rollback) + `toast.success` / `toast.error`. |
| **Verdict** | **OK** — three-state (loading / error / data) pattern with skeleton, `ErrorState`, and toast feedback on mutations. |

---

### `AccountView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `dashboardStore.updateHouseholdName()` on household name save |
| **Loading UX** | No spinner shown during save. The save and cancel buttons are disabled via `:disabled="isSavingHouseholdName"` and the input gets `opacity: 0.6`. No visual "saving…" indicator. |
| **Error UX** | Inline `<p role="alert">` below the input with `householdNameError`. |
| **Non-wired sections** | Notifications toggles (`notificationsEnabled`, `mealReminders`, `shoppingReminders`) are local `ref()` values with a comment "not yet wired to backend". No async, no risk of missing feedback — but the feature is silently non-functional. |
| **Gaps** | **Needs loading spinner on save** — the button disable is correct but there is no spinner or "Sparar…" label to communicate progress. The gap is small (save is usually fast) but noticeable on slow connections. |
| **Verdict** | **Mostly OK, minor gap** — error feedback is present, but loading during save is invisible. |

---

### `RecipesView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | None directly in the view — delegates fully to `RecipeListPanel` (list + search) and `AddRecipePanel` (AI parse + save). |
| **Loading UX** | Delegated to child panels. |
| **Error UX** | Delegated to child panels. |
| **Child: `RecipeListPanel`** | `isLoading` + `useSkeleton(minDuration: 300)` → `<RecipesSkeleton>` layout. Error: `<ErrorState>` with retry button. **OK.** |
| **Child: `AddRecipePanel`** | `useProgressBar(duration: 20000)` drives animated `<ProgressBar>` during AI parse (Claude can take up to 60 s, so the simulated bar is appropriate). `isParsing`, `parseError` (inline), `isSaving` (button disabled). **OK.** |
| **Verdict** | **OK** — the view shell is a pass-through tab switcher; loading/error owned by panels which handle both well. |

---

### `GenerateMenuView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `store.generateInitialMenu()` and `store.regenerateUnlockedDays()` inside `handleInitialGenerate` / `handleRegenerate`; `store.saveDraftMenu(router)` on save |
| **Loading UX** | **Slot machine animation** serves as the primary visual during generation — cards "roll" while the API runs concurrently. `<GenerateMenuSkeleton>` shown via `useSkeleton` only for the gap between `store.isGenerating && !slotMachine.isAnimating.value` (a brief window). `MenuGeneratorActions` actions bar becomes disabled while `isLoading \|\| slotMachine.isAnimating.value`. |
| **Error UX** | `<ErrorState v-if="store.error">` with retry — but this is rendered *inside* the content container, not as a full-page replacement. When an error occurs after the menu grid appears, the `ErrorState` banner is appended **below** the grid (the condition `v-if` is not `v-else-if`). This means the grid and the error are simultaneously visible. |
| **Catch blocks** | Both `handleInitialGenerate` and `handleRegenerate` call `toast.error(...)` *and* `store.setError(...)`. The toast fires immediately; the `ErrorState` may appear below still-visible cards. |
| **Verdict** | **Gap: error/grid overlap** — `ErrorState` should be `v-else-if="store.error"` relative to the grid, not `v-if`. On first generation failure the guard `!showGrid` catches it, but on regeneration failure the grid remains shown alongside the error banner. |

---

### `ShoppingListView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `useShoppingList` composable (`fetchList`, `toggle`, `addCustomItem`, `removeCustomItem`); `dashboardStore.fetchDashboard()` as prerequisite if `!dashboardStore.dashboardData` |
| **Loading UX** | Custom inline skeleton (pulsing divs for category headings + item rows) gated by `useSkeleton(isLoading, {minDuration: 300})`. Does **not** use the shared skeleton component library — inline CSS animation instead. |
| **Secondary fetch gap** | `dashboardStore.fetchDashboard()` is called silently if data is absent. If it is loading, the view may resolve `menuId` as `null` and show the "Ingen aktiv meny" empty state prematurely. No loading indicator covers this secondary fetch. |
| **Error UX** | `<ErrorState v-else-if="error">` with a retry that calls `fetchList(menuId!)`. Correctly guards against the "no menu" and "empty list" empty states. |
| **Optimistic mutations** | `toggle()` and `removeCustomItem()` in `useShoppingList` — need to verify composable handles rollback, but the view does not show per-item spinners. Acceptable for a list; failure is silent if not handled in composable. |
| **Verdict** | **Minor gaps** — (a) skeleton is inline rather than shared component (inconsistency), (b) the secondary dashboard fetch is silent and can show a misleading "no menu" state before completing. |

---

### `LoginView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `userStore.login(email, password)` on submit |
| **Loading UX** | `BaseButton :loading="isLoading"` — the button shows a built-in spinner and is implicitly disabled. `canSubmit` computed prevents submission while in-flight. |
| **Error UX** | `<p v-show="error" role="alert">` — inline error block below the password field. Session-expiry toast (`toast.info`) on mount if `session_expired` is set in `sessionStorage`. |
| **Verdict** | **OK** — loading spinner on CTA + inline `role="alert"` error. Clear pattern. |

---

### `OnboardingView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | Delegated to `OnboardingJoinFlow` (join household) and `OnboardingCreateForm` (register + create household) |
| **Loading UX at view level** | None — the view is a step manager. |
| **Child: `OnboardingJoinFlow`** | `isValidatingCode`, `isSubmitting` refs used to disable buttons. `codeError` and `submitError` shown inline below inputs. No skeleton. No toast. |
| **Child: `OnboardingCreateForm`** | `isSubmitting` disables button. `createError` shown inline. Uses `userStore.register()` which sets `userStore.error`. No skeleton. No toast on failure — error is inline only. |
| **Gaps** | Neither child shows a spinner during submission — buttons are disabled but show no visual indication of in-progress state. `BaseButton :loading` prop exists and is used elsewhere (LoginView, ForgotPasswordView) but is not passed here. |
| **Verdict** | **Needs loading state on submit buttons** — `isSubmitting` is tracked but not passed to `BaseButton :loading`. Error feedback is inline, which is correct for forms. |

---

### `LandingView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `landingStore.fetchLandingData()` on mount |
| **Loading UX** | Custom spinner: `<Loader2>` icon with CSS spin animation + "Laddar..." text, wrapped in `role="status" aria-label="Laddar sidan"`. Full-page centred. |
| **Error UX** | Inline error state with `<AlertTriangle>` icon, error message from `landingStore.error`, and a "Försök igen" retry button. Full-page centred. |
| **Verdict** | **OK** — three-state guard, retry provided, accessible roles on loading indicator. Spinner rather than skeleton is acceptable for a marketing page that renders quickly or not at all. |

---

### `AboutView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | None — fully static content |
| **Loading UX** | N/A |
| **Error UX** | N/A |
| **Verdict** | **N/A** — static page, no async, no gaps. |

---

### `ForgotPasswordView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `forgotPassword(email)` API call on submit |
| **Loading UX** | `BaseButton :loading="isLoading"` — spinner on CTA. `canSubmit` computed includes `!isLoading.value` so double-submit is impossible. |
| **Error UX** | **Intentionally none** — errors are caught and silently swallowed. The view always transitions to a "Kolla din inkorg" success state regardless of outcome. This is correct security practice (prevents account-existence enumeration). Comment in code explains the reasoning. |
| **Verdict** | **OK by design** — deliberate error suppression for security, documented in code. |

---

### `ResetPasswordView.vue`

| Aspect | Detail |
|---|---|
| **Async sources** | `resetPassword(token, password)` API call on submit |
| **Loading UX** | `BaseButton :loading="isLoading"` — spinner on CTA. `canSubmit` guards `!isLoading.value`. |
| **Error UX** | Inline `<p v-show="error" role="alert">` with specific Swedish messages per error code (`invalid_reset_token`, `expired_reset_token`, `used_reset_token`, `weak_password`, default). Token presence validated on mount — shows error immediately if URL has no token. Real-time password mismatch hint shown inline. |
| **Verdict** | **OK** — comprehensive error mapping, inline feedback, loading spinner. Best-in-class among form views. |

---

## Summary Findings

### Patterns Observed

- **Three-state guards** (`isLoading → skeleton/spinner`, `error → ErrorState`, `data → content`) are used consistently in the full-page views: `DashboardView`, `LandingView`, `GenerateMenuView`.
- **`BaseButton :loading` prop** is the standard CTA loading pattern and used correctly in `LoginView`, `ForgotPasswordView`, `ResetPasswordView`. It is *missed* in `OnboardingJoinFlow` and `OnboardingCreateForm` despite `isSubmitting` being tracked.
- **Inline error with `role="alert"`** is the form-error standard and applied consistently across auth flows.
- **Toast notifications** are used for mutations (swap recipe, remove member, reset password success, generate menu failure) — not for page-load errors, which is appropriate.
- **Retry controls** are present on all `<ErrorState>` usages except the `GenerateMenuView` overlap issue.

### Top 3 Gaps

1. **`GenerateMenuView.vue` — error/grid co-visibility during regeneration.** The `<ErrorState>` is `v-if="store.error"` (unconditional) while the menu grid is also showing. On regeneration failure, both render simultaneously. Fix: make `v-else-if` or set `store.clearDraft()` on error so `showGrid` becomes false.

2. **`OnboardingView.vue` child forms — missing loading spinner on submit.** `OnboardingJoinFlow` and `OnboardingCreateForm` both track `isSubmitting` but never pass it to `BaseButton :loading`. Users get no visual confirmation the form is processing. Fix: add `:loading="isSubmitting"` to both submit buttons.

3. **`ShoppingListView.vue` — silent secondary dashboard fetch.** The `dashboardStore.fetchDashboard()` prerequisite call has no loading indicator of its own. On first visit with a cold store, `menuId` may resolve to `null` temporarily, causing a flash of "Ingen aktiv meny" before the real state arrives. Fix: gate the `EmptyState` for `!menuId` on `!dashboardStore.isLoading` as well, or initialise `menuId` eagerly.

### Minor Notes

- **`AccountView.vue`** — household name save has no visual "saving…" indicator (buttons disabled, no spinner). Low severity since the operation is fast.
- **`ShoppingListView.vue`** — uses inline skeleton CSS animation instead of the shared `SkeletonBlock`/`SkeletonItem` component library. Cosmetic inconsistency.
- **`LandingView.vue`** — uses a spinner (`Loader2`) rather than a skeleton. Acceptable for a marketing page (content is monolithic, not incremental), but differs from the skeleton approach used in app views.
