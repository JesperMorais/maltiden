# Design Review Results: Måltiden — All Pages

**Review Date**: 2026-02-26  
**Routes Reviewed**: `/` · `/login` · `/register` (onboarding) · `/dashboard` · `/recipes` · `/menu/generate` · `/shopping-list` · `/about`  
**Focus Areas**: Visual Design · UX/Usability · Responsive/Mobile · Accessibility · Micro-interactions · Consistency · Performance

---

## Summary

Måltiden has an excellent visual foundation — warm palette, expressive typography (`Fraunces` + `Nunito`), and thoughtful animations. However, several **critical accessibility gaps** exist (disabled-button contrast, emoji as interactive elements without labels, missing `prefers-reduced-motion`), **consistency debt** accumulates across pages (duplicated theme-toggle SVG, three different "back" navigation patterns, retry buttons re-implementing BaseButton), and a **high CLS of 0.143** on the dashboard hurts perceived quality. Addressing these would bring the product to a production-ready standard.

---

## Issues

| # | Issue | Criticality | Category | Location |
|---|-------|-------------|----------|----------|
| 1 | Disabled "Logga in" button renders salmon-on-white at ~2.5:1 contrast (needs ≥ 4.5:1 per WCAG 2.1 AA) | 🔴 Critical | Accessibility | `src/views/LoginView.vue:127-135` · `src/components/common/BaseButton.vue:134-138` |
| 2 | Password-toggle buttons use emoji (👁️ / 🙈) as sole content with no `aria-label` — invisible to screen readers | 🔴 Critical | Accessibility | `src/views/OnboardingView.vue:398-404`, `551-556` |
| 3 | Dashboard CLS = 0.143 (exceeds 0.1 "good" threshold) — `RotatingText` likely causes layout shift on mount | 🔴 Critical | Performance | `src/views/DashboardView.vue:122-131` · `src/components/vue-bits/RotatingText.vue` |
| 4 | `RotatingText` and `FadeContent` animations have no `@media (prefers-reduced-motion: reduce)` guard — violates WCAG 2.3.3 | 🔴 Critical | Accessibility | `src/components/vue-bits/FadeContent.vue` · `src/components/vue-bits/RotatingText.vue` |
| 5 | No "Glömt lösenord?" (forgot password) link on login page — blocks locked-out users | 🟠 High | UX/Usability | `src/views/LoginView.vue:138-145` |
| 6 | Dashboard fully-loaded time of **25 574 ms** — timeout waiting for offline Go backend pollutes perceived performance even in mock mode | 🟠 High | Performance | `src/api/client.ts` |
| 7 | `sidebar::after` applies a fixed-position gradient overlay on mobile that permanently obscures page bottom content | 🟠 High | Responsive/Mobile | `src/views/DashboardView.vue:298-311` |
| 8 | Tab-button touch targets measure ~38px height — below 44 px minimum recommended for mobile | 🟠 High | Responsive/Mobile | `src/views/RecipesView.vue:155-165` |
| 9 | `retry-btn` in LandingView re-implements BaseButton styles instead of using the existing component | 🟠 High | Consistency | `src/views/LandingView.vue:131-148` |
| 10 | Login page FCP = 984 ms despite having no real data dependency — WavesBackground canvas animation starts before first paint | 🟠 High | Performance | `src/views/LoginView.vue:57-66` · `src/components/vue-bits/WavesBackground.vue` |
| 11 | Theme toggle SVG icon code is duplicated verbatim in `LoginView` and `OnboardingView` instead of a shared component | 🟡 Medium | Consistency | `src/views/LoginView.vue:87-89` · `src/views/OnboardingView.vue:203-207` |
| 12 | "Back" navigation uses three different patterns: `router.push()` (Recipes), `RouterLink` (Login), raw `<button>` (Recipes header) | 🟡 Medium | Consistency | `src/views/RecipesView.vue:30-33` · `src/views/LoginView.vue:78-81` · `src/views/OnboardingView.vue:197-199` |
| 13 | Dashboard sidebar `top` is hardcoded to `calc(70px + 2rem)` — breaks immediately if header height changes | 🟡 Medium | Consistency | `src/views/DashboardView.vue:271` |
| 14 | Emoji used as primary UI icons (🍽️, 🍳, 😅, 🏠, ⭐) mixed with Lucide SVG icons — no consistent icon system | 🟡 Medium | Consistency | Multiple views and components |
| 15 | `aria-describedby` on form inputs is conditionally set but the referenced element (`#login-error`) may not exist yet at mount, confusing AT | 🟡 Medium | Accessibility | `src/views/LoginView.vue:106-109`, `118-121` |
| 16 | Onboarding flow (6 steps, join/create branching) is 1 369-line single-file component — high maintenance risk and code splitting blocked | 🟡 Medium | UX/Usability | `src/views/OnboardingView.vue` (entire file) |
| 17 | `v-show` vs `v-if` inconsistency in RecipesView: list panel uses `v-show` (always in DOM), add panel uses `v-if` (re-created on switch) | 🟡 Medium | Performance | `src/views/RecipesView.vue:63-73` |
| 18 | BaseCard `::before` hover overlay stays in a "hovered" state on touch devices (no touch equivalent to `mouseleave`) | 🟡 Medium | Micro-interactions | `src/components/common/BaseCard.vue:32-48` |
| 19 | No empty-state illustration or CTA on the recipe list when zero recipes exist | 🟡 Medium | UX/Usability | `src/components/recipes/RecipeListPanel.vue` |
| 20 | Recipe grid can produce an orphaned single card at mid-width breakpoints due to CSS auto-fill columns | ⚪ Low | Responsive/Mobile | `src/components/recipes/RecipeListPanel.vue` |
| 21 | BaseButton spring easing (`cubic-bezier(0.34, 1.56, 0.64, 1)`) causes overshoot on `size-lg` which can feel unprofessional in form contexts | ⚪ Low | Micro-interactions | `src/components/common/BaseButton.vue:44` |
| 22 | Loading emoji (🍳) in LandingView loading state has no ARIA equivalent (`role="status"` or `aria-label`) | ⚪ Low | Accessibility | `src/views/LandingView.vue:17-22` |
| 23 | Dashboard greeting (`RotatingText`) adds ~2.5em of vertical space before main content with no option to dismiss | ⚪ Low | UX/Usability | `src/views/DashboardView.vue:228-243` |
| 24 | `createForm.householdName` uses `autocomplete="organization"` — functionally fine but semantically marginal for a household name | ⚪ Low | Accessibility | `src/views/OnboardingView.vue:587-596` |

---

## Criticality Legend

- 🔴 **Critical**: Breaks functionality or violates accessibility standards (WCAG 2.1 AA)
- 🟠 **High**: Significantly impacts user experience, design quality, or core flows
- 🟡 **Medium**: Noticeable issue that should be addressed in next sprint
- ⚪ **Low**: Nice-to-have / polish improvement

---

## Next Steps (Recommended Priority Order)

### Sprint 1 — Critical fixes
1. **Fix disabled button contrast** — Use `opacity: 0.4` and ensure underlying button colours still pass contrast at all opacity levels, or apply a dedicated disabled token.
2. **Add `aria-label` to password toggles** — Replace emoji with `aria-label="Visa lösenord"` / `aria-label="Dölj lösenord"`.
3. **Wrap all animations in `prefers-reduced-motion`** — Add `@media (prefers-reduced-motion: reduce)` to `FadeContent` and `RotatingText`.
4. **Reserve layout space for `RotatingText`** — Use a fixed `min-height` on `.dashboard-greeting` so it doesn't shift content as it loads.

### Sprint 2 — High-impact UX & Performance
5. **Add "Glömt lösenord?"** link to login.
6. **Increase API timeout or abort faster** — Set a 3–5 s timeout in `client.ts` so mock mode doesn't block for 25 s.
7. **Fix mobile fixed-gradient overlay** — Remove `position: fixed` on `sidebar::after`, use a local `sticky` gradient inside `.sidebar` instead.
8. **Increase tab-button height** to `min-height: 44px` on Recipes view.
9. **Extract `BaseThemeToggle` component** — Reuse in LoginView and OnboardingView.

### Sprint 3 — Consistency & Debt
10. **Replace `retry-btn` in LandingView with `<BaseButton>`.**
11. **Standardise back-navigation** — Create a `<BackLink>` component used across all views.
12. **Use a CSS custom property for header height** (`--header-height: 70px`) referenced everywhere instead of hard-coded `70px`.
13. **Extract Onboarding steps** into sub-components (`OnboardingCodeStep.vue`, `OnboardingMemberForm.vue`, etc.).
14. **Audit emoji vs. Lucide usage** — Choose one icon strategy and apply consistently.
