# MVP Launch Prep — Implementation Plan

> **For Claude:** Execute with agent teams. See team setup at bottom.

**Goal:** Get Måltiden to launch-ready state for 2-5 beta families on mobile. Merge polish branch, fix CI, update docs.

**Architecture:** No new features. Fix broken tests, merge branches, update launch checklist. All frontend work.

**Tech Stack:** Vue 3, TypeScript, Playwright, Git

---

## Current State

- Branch `feat/fe_design-review-fixes` has 10 commits of polish work (skeletons, animations, touch targets, icons, contrast fixes)
- All M1-M6 mobile polish items are implemented
- Playwright: 211 pass, 61 fail (all pre-existing), 28 did not run
- Frontend builds clean: type-check ✅, lint ✅, build ✅, vitest 49/49 ✅

## Pre-flight: Failure Categories

| Category | Count | Root Cause | Fix Strategy |
|----------|-------|------------|--------------|
| e2e tests (need real backend) | 8 | Run against mock mode, need backend | Skip in mock config OR mark as e2e-only |
| qa-feedback.spec.ts | 24 | FeedbackWidget not mounted in test context | Fix beforeEach — widget needs auth + dashboard route |
| recipes.spec.ts | 4 | `getByRole('button', { name: /Recept.*Hantera/ })` flaky | Fix selector — use MobileBottomNav or navigateTo |
| screenshots.spec.ts | 10 | Same recipe nav selector + recipe card timing | Same fix as recipes + increase timeouts |
| interactions.spec.ts | 5 | Various — shopping list, menu gen, onboarding | Fix selectors and timing |
| dashboard.spec.ts | 1 | Shopping list widget items count | Fix selector or expected value |
| accessibility.spec.ts | 1 | Add recipe tab | Fix selector |
| shopping-list.spec.ts | 1 | Navigation from dashboard widget | Fix selector |

---

### Task 1: Fix recipe navigation selector across all test files

**Files:**
- Modify: `tests/recipes.spec.ts:7`
- Modify: `tests/screenshots.spec.ts:77,87,104`
- Modify: `tests/helpers.ts` (add new helper)

**Step 1: Add navigateToRecipes helper**

In `tests/helpers.ts`, add:

```ts
/** Navigate to recipes page via client-side routing */
export async function navigateToRecipes(page: Page) {
  await navigateTo(page, '/recipes')
  await page.waitForTimeout(1500)
}
```

**Step 2: Update recipes.spec.ts beforeEach**

Replace line 7:
```ts
// Old: await page.getByRole('button', { name: /Recept.*Hantera/ }).click()
// New:
await navigateTo(page, '/recipes')
await page.waitForTimeout(1000)
```

Update import to include `navigateTo`.

**Step 3: Update screenshots.spec.ts recipe sections**

Replace all `page.getByRole('button', { name: /Recept.*Hantera/ }).click()` with:
```ts
await navigateTo(page, '/recipes')
await page.waitForTimeout(1500)
```

**Step 4: Run affected tests**

```bash
node node_modules/.bin/playwright test tests/recipes.spec.ts tests/screenshots.spec.ts --reporter=line
```

**Step 5: Commit**

```bash
git add tests/recipes.spec.ts tests/screenshots.spec.ts tests/helpers.ts
git commit -m "fix: use navigateTo for recipe page tests instead of flaky button click"
```

---

### Task 2: Fix qa-feedback.spec.ts — FeedbackWidget needs auth context

**Files:**
- Modify: `tests/qa/qa-feedback.spec.ts`

**Step 1: Read and understand the test file**

The FeedbackWidget renders on authenticated pages (dashboard). Tests need to:
1. Login first
2. Navigate to dashboard
3. Wait for FeedbackWidget to mount

**Step 2: Fix the beforeEach blocks**

Each `test.describe` that requires the widget needs a proper login + dashboard load. Check if `beforeEach` already calls `login(page)` — if so, the issue is likely that the widget renders after a delay or behind a condition. Fix timing:

```ts
test.beforeEach(async ({ page }) => {
  await login(page)
  // FeedbackWidget may have a render delay
  await page.waitForTimeout(2000)
})
```

**Step 3: Fix individual selectors**

Check if `button[aria-label="Ge feedback"]` matches the actual rendered button. The FeedbackWidget uses a trigger button — verify the aria-label matches.

**Step 4: Run qa-feedback tests**

```bash
node node_modules/.bin/playwright test tests/qa/qa-feedback.spec.ts --reporter=line
```

**Step 5: Commit**

```bash
git add tests/qa/qa-feedback.spec.ts
git commit -m "fix: stabilize FeedbackWidget Playwright tests with proper auth and timing"
```

---

### Task 3: Fix interactions.spec.ts failures

**Files:**
- Modify: `tests/interactions.spec.ts`

**Step 1: Read the file and identify each failing test**

5 failures:
- Shopping list rapid check-off
- Recipe browse on mobile (375px) — 2 tests
- Menu generation
- Onboarding — 2 tests

**Step 2: Fix shopping list test**

Likely needs `navigateTo(page, '/shopping-list')` instead of widget click, or increased timeout.

**Step 3: Fix recipe browse mobile tests**

Same recipe navigation fix as Task 1 — use `navigateTo` instead of button click.

**Step 4: Fix menu generation test**

Check if generate button selector matches. The GenerateMenuEmptyState has a `.generate-button` class.

**Step 5: Fix onboarding tests**

Check selectors against actual rendered elements. Onboarding flow may have changed form structure.

**Step 6: Run and verify**

```bash
node node_modules/.bin/playwright test tests/interactions.spec.ts --reporter=line
```

**Step 7: Commit**

```bash
git add tests/interactions.spec.ts
git commit -m "fix: stabilize interaction Playwright tests with reliable selectors"
```

---

### Task 4: Fix remaining single-test failures

**Files:**
- Modify: `tests/dashboard.spec.ts`
- Modify: `tests/accessibility.spec.ts`
- Modify: `tests/shopping-list.spec.ts`

**Step 1: Fix dashboard shopping list widget test**

Read the test, check what "items count" it expects. Fix selector or expected value to match mock data.

**Step 2: Fix accessibility add-recipe tab test**

Check if the tab selector matches the current RecipesView tab structure.

**Step 3: Fix shopping-list navigation test**

Use `navigateTo` instead of widget click if that's the issue.

**Step 4: Run all three**

```bash
node node_modules/.bin/playwright test tests/dashboard.spec.ts tests/accessibility.spec.ts tests/shopping-list.spec.ts --reporter=line
```

**Step 5: Commit**

```bash
git add tests/dashboard.spec.ts tests/accessibility.spec.ts tests/shopping-list.spec.ts
git commit -m "fix: stabilize dashboard, accessibility, and shopping list tests"
```

---

### Task 5: Skip e2e tests in default config

**Files:**
- Modify: `playwright.config.ts`

**Step 1: Exclude e2e directory from default test run**

The e2e tests need a real backend. They should only run via `npm run test:e2e:real`. Update the default config to exclude them:

```ts
export default defineConfig({
  testDir: './tests',
  testIgnore: ['**/e2e/**'],
  // ... rest unchanged
})
```

**Step 2: Verify e2e tests are excluded**

```bash
node node_modules/.bin/playwright test --list 2>&1 | grep -c "e2e"
# Should be 0
```

**Step 3: Commit**

```bash
git add playwright.config.ts
git commit -m "fix: exclude e2e tests from default Playwright config (need real backend)"
```

---

### Task 6: Full test run — verify all pass

**Step 1: Run full suite**

```bash
node node_modules/.bin/playwright test --reporter=line > /tmp/pw-final.txt 2>&1
```

**Step 2: Check results**

Target: 0 failures (excluding e2e which are now skipped).

**Step 3: Fix any remaining failures**

Iterate until green.

**Step 4: Run frontend CI checks too**

```bash
cd frontend && npm run type-check && npm run lint && npm run build && npm run test
```

---

### Task 7: Update TODO.md — mark M1-M6 complete

**Files:**
- Modify: `docs/TODO.md`

**Step 1: Mark M1-M6 section as complete**

Change each M1-M6 item to show ✅ status. Update the launch checklist checkbox:
```
- [x] 🟠 M1–M6 mobilpolish klar
```

**Step 2: Update "Senast uppdaterad" date**

Change to `2026-03-10`.

**Step 3: Commit**

```bash
git add docs/TODO.md
git commit -m "docs: mark M1-M6 mobile polish complete in TODO"
```

---

### Task 8: Merge to dev

**Step 1: Rebase onto dev**

```bash
git fetch origin
git rebase origin/dev
```

**Step 2: Resolve any conflicts**

**Step 3: Merge into dev**

```bash
git checkout dev
git merge --ff-only feat/fe_design-review-fixes
git push origin dev
```

---

## Agent Team Setup

### Team: `mvp-launch-prep`

**3 agents working in parallel, then sequential merge:**

#### Wave 1 (parallel) — Test Fixes

| Agent | Name | Type | Tasks | Instructions |
|-------|------|------|-------|-------------|
| **A** | `test-fixer-nav` | general-purpose | Task 1, Task 5 | Fix recipe navigation selectors across all test files. Use `navigateTo(page, '/recipes')` instead of flaky button clicks. Also exclude e2e tests from default config. Commit each fix separately. |
| **B** | `test-fixer-widgets` | general-purpose | Task 2, Task 4 | Fix FeedbackWidget tests (auth + timing) and single-test failures (dashboard, accessibility, shopping-list). Read each test, check selectors against actual components, fix timing. Commit each fix separately. |
| **C** | `test-fixer-interactions` | general-purpose | Task 3 | Fix 5 failing interaction tests. Read each test carefully, compare selectors to actual component classes, use navigateTo for routing, fix timing issues. Commit when all pass. |

#### Wave 2 (sequential) — Verify & Merge

| Agent | Name | Type | Tasks | Instructions |
|-------|------|------|-------|-------------|
| **D** | `verifier` | general-purpose | Task 6, Task 7 | After Wave 1 completes: run full Playwright suite, fix any remaining failures. Run frontend CI (type-check, lint, build, test). Update TODO.md. |

#### Wave 3 (manual) — Merge

Task 8 is done by the orchestrator (you) after all agents complete.

### Agent Launch Commands

```
Wave 1 — Launch all 3 in parallel:

Agent A (test-fixer-nav):
  "Fix recipe navigation selectors in tests/recipes.spec.ts and tests/screenshots.spec.ts.
   Replace getByRole('button', { name: /Recept.*Hantera/ }).click() with navigateTo(page, '/recipes')
   from tests/helpers.ts. Also exclude tests/e2e/** from playwright.config.ts testIgnore.
   Read each file first, make changes, run the specific tests to verify, commit separately.
   Working directory: /home/david/Dev/personal/maltiden"

Agent B (test-fixer-widgets):
  "Fix Playwright test failures in tests/qa/qa-feedback.spec.ts (24 failures) and
   tests/dashboard.spec.ts, tests/accessibility.spec.ts, tests/shopping-list.spec.ts (1 each).
   Root cause: FeedbackWidget needs auth+dashboard loaded, other tests have stale selectors.
   Read each test file, read the referenced Vue components to check actual class names and text,
   fix selectors and timing. Run each test file after fixing to verify. Commit separately.
   Working directory: /home/david/Dev/personal/maltiden"

Agent C (test-fixer-interactions):
  "Fix 5 failing tests in tests/interactions.spec.ts. Failures: shopping list rapid check-off,
   recipe browse on mobile (375px) x2, menu generation, onboarding x2.
   Read the test file, read referenced Vue components for correct selectors,
   use navigateTo() for client-side routing instead of button clicks where appropriate.
   Run tests after each fix. Commit when all 5 pass.
   Working directory: /home/david/Dev/personal/maltiden"

Wave 2 — After Wave 1:

Agent D (verifier):
  "Run full Playwright suite: node node_modules/.bin/playwright test --reporter=line
   Target: 0 failures. Fix any remaining issues. Then run frontend CI:
   cd frontend && npm run type-check && npm run lint && npm run build && npm run test
   Finally update docs/TODO.md: mark M1-M6 as ✅, update date to 2026-03-10,
   check the launch checklist box for mobilpolish. Commit.
   Working directory: /home/david/Dev/personal/maltiden"
```

### Important Notes for All Agents

- **Never use `git add -u` or `git add .`** — always `git add <specific files>`
- **Never use `git add -A`** — explicitly name files
- **Read files before editing** — understand existing patterns
- **Run tests after changes** — don't commit blind
- **Use `node node_modules/.bin/playwright test <file>` to run Playwright** (not npx)
- **Mock mode on port 5174** — dev server should already be running
- **Commit messages:** English, lowercase, no AI references
