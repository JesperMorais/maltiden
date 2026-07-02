# UI Audit — Måltiden Frontend

*Audit date: 2026-07-02. Scope: `frontend/src` — all views, key components, `styles/theme.css`, `App.vue`, `index.html`. Audit only; no code changed.*

**Method:** Full read of the design-token layer and every view, plus the highest-traffic components (dashboard widgets, nav, buttons, modals). Every finding below was verified against the actual source, not inferred. Primary lens is the project's own target: **375–414px mobile, cooking in the kitchen / shopping in the store**, judged on "can a user finish the core action," not on decoration.

**Overall verdict:** The visual identity is genuinely good — warm cream + coral, Fraunces/Nunito pairing, cohesive dark mode, real empty/error/skeleton states, focus traps and ARIA on modals. This is far above template quality. The problems are (1) three concrete mobile/layout bugs, (2) a design-token system that exists but is ~95% unused, so consistency is maintained by hand, and (3) too many animated gimmicks competing on the dashboard, diluting an otherwise confident aesthetic.

---

## P1 — Bugs and blockers (fix first)

### 1. Mobile dashboard: HouseholdWidget jumps to the top, above "Idag"
`views/DashboardView.vue:302-318` — the ≤768px layout flattens `.main-area`/`.sidebar` with `display: contents` and re-orders children with `order: 1..4`. The comment says the intended order is *TodaysMeal → QuickActions → WeeklyMenu → ShoppingList*, but `HouseholdWidget` (third child of `.main-area`, `DashboardView.vue:153-162`) gets no `order` rule → defaults to `order: 0` → **renders first on every phone**, pushing today's meal below household chips. On the primary 375px target, the most important card is not first.

**Change:** give the household wrapper an explicit `order: 5` (it's the least important widget on mobile), and add a comment that any new `.main-area`/`.sidebar` child must get an explicit order. Better long-term: drop the `display: contents` + `:deep` order hack entirely and let the mobile layout be its own flex column with explicit component order in the template (a `useMediaQuery` composable or a simple CSS-grid `grid-template-areas` swap is far less fragile).

### 2. Generate-menu save bar collides with the mobile bottom nav
`components/menu/MenuGeneratorActions.vue:67-73` — the actions bar is `position: sticky; bottom: 0; z-index: 10`. `components/common/MobileBottomNav.vue:49-62` is `position: fixed; bottom: 0; z-index: 1000` on every authenticated route. On mobile the fixed nav overlays the sticky bar — **"Spara meny" / "Generera om" sit behind or flush against the tab bar** until the user scrolls past the spacer. This is the single most important CTA in the app's core flow.

**Change:** on ≤768px set `bottom: calc(var(--bottom-nav-height) + env(safe-area-inset-bottom, 0px))` on `.menu-actions`, or hide `MobileBottomNav` on the `generate-menu` route via route meta (it's a focused flow — a bottom tab bar mid-wizard is noise anyway; hiding it is my recommendation).

### 3. Seven days forced into a 5-column grid
`stores/menuGenerator.ts:67-77` builds a full 7-day week (`for i < 7`), and the header says "Måndag – Söndag" — but `views/GenerateMenuView.vue:334-338` lays cards out as `repeat(5, 1fr)`. Result: a ragged 5+2 row on desktop, 3+3+1 at 1024px. It reads as broken, not designed.

**Change:** `repeat(7, 1fr)` on wide screens (cards are compact enough at 1200px container), or an intentional 4+3 split; at 1024px use `repeat(4, 1fr)`. If the product intent is actually a 5-day work-week menu, fix the store and subtitle instead — right now the code disagrees with itself (`orderedDays` doc comment even says "Mon-Fri").

### 4. Primary button text fails contrast
`components/common/BaseButton.vue:70-74` — white label on `linear-gradient(#ff6b5b → #e85a4a)` is ≈2.8:1. WCAG AA needs 4.5:1 (3:1 large). The theme already solves this — `--accent-text: #c4402e` exists precisely for this ("WCAG AA: 4.75:1", `theme.css:26`), and `TodaysMeal.vue:215-233` already uses `background: var(--accent-text)` with white text (≈5.9:1) for its "Se recept" pill. So the app currently ships **two different primary-CTA fills**, and the more common one fails AA.

**Change:** darken the button gradient to end at `--accent-text` territory (e.g. `#e85a4a → #c4402e`), or flip the label to `#3d2c29`-on-coral and test. Whichever you pick, make `TodaysMeal`'s pill and `BaseButton.variant-primary` the *same* surface — one primary CTA treatment app-wide.

### 5. Role badge "ÄGARE" is unreadable
`components/dashboard/DashboardHeader.vue:217-225` — `.role-badge.owner` puts white 0.7rem uppercase text on the orange `--role-owner-bg` gradient (≈1.9:1 against `#f6ad55`). The theme defines `--role-owner-text: #9a4217` with a documented AA ratio (`theme.css:111`) — it's just not used here.

**Change:** use the tinted-bg + dark-text pattern the tokens were designed for: `background: var(--role-owner-text-bg); color: var(--role-owner-text)`. Same for `.member`.

---

## P2 — Design system: tokens exist, nobody uses them

This is the highest-leverage structural change. Measured across all `.vue` files:

| Token family (defined in `theme.css:64-85`) | Usages in components |
|---|---|
| `--space-xs … --space-2xl` | **0** |
| `--duration-fast/normal/slow`, `--ease-*` | **0** (vs. 77 hand-written `transition: all 0.2s/0.3s`) |
| `--radius-sm … --radius-full` | 12 (vs. **201** hard-coded `border-radius: 14px/16px/18px/20px/24px/32px…`) |

Plus **33 raw hex values in 12 component files** (e.g. the category palette in `ShoppingListView.vue:63-71`, spark color `#ff6b5b` in `QuickActions.vue:85`, and colors inside `AccountView.vue`, `InviteModal.vue`, `RecipeParseInput.vue`, `OnboardingCreateForm.vue`…), directly violating the project's own "ALWAYS use `var(--token-name)`, never raw hex" rule. The shopping-list category colors also have **no dark-mode variants** — `#a67c4e` on `#3d322e` gets muddy.

**Changes, in order of value:**
1. **Radius:** the app uses ~6 real radii. Extend the scale (`--radius-xl: 20px`, `--radius-2xl: 24px`, `--radius-hero: 32px`), then mechanical find-replace. 201 call sites is why cards are 14/16/18/20px within one screen (`ShoppingListView` alone uses 16, 18, 14, 999).
2. **Category colors → theme:** move `categoryAccent` into `theme.css` as `--cat-meat`, `--cat-dairy`, … with dark variants, and read them in the component. This also fixes dark mode for free.
3. **Transitions:** replace `transition: all 0.3s ease` with scoped properties + tokens (`transition: border-color var(--duration-normal) var(--ease-out), transform …`). `transition: all` is also a perf smell on 30+ hover-transform cards.
4. **Font-family:** there are **215** `font-family: 'Nunito'…` declarations. Set `body { font-family: 'Nunito', … }` (already in `App.vue`) and `h1,h2,h3 { font-family: 'Fraunces', serif }` globally, then delete the per-class declarations — components should only override when deviating. This kills ~200 lines and makes a future font change a one-liner.
5. **Breakpoints:** 9 distinct values in use (768×27, 480×12, 1024×6, 640×3, plus one-offs 767, 600, 520, 500, 1280). `WeeklyMenuGrid` switches at 640 while its parent dashboard switches at 768 — between 640–768px the day cards are in cramped desktop mode inside a mobile column. Standardize on three (480 / 768 / 1024) and document them in `theme.css` as comments (CSS vars don't work in media queries; consider `@custom-media` via PostCSS if you want it enforced).

---

## P3 — Layout & hierarchy

- **Dashboard is over-animated and under-prioritized.** One screen runs: a JS `RotatingText` greeting cycling every 4.5s (`DashboardView.vue:124-132`), an infinite 8s shimmer sweep + two blurred blobs on TodaysMeal (`TodaysMeal.vue:136-154`), `SpotlightCard` mouse-tracking + `ClickSpark` particle bursts on QuickActions, and staggered `FadeContent` blurs on everything. Individually fine; together they compete and read as "AI demo." The design-system doc's own principle applies: spend boldness in one place. **Keep** the slot-machine generate animation (that's the app's signature moment, and it's in the right place), **keep** one entrance stagger, **cut** the rotating greeting (a static, personalized "God kväll, David — vad blir det till middag?" is warmer and calmer), the shimmer loop, and ClickSpark.
- **Every inner page hand-rolls the same header** (`Fraunces 2.5rem title + Nunito 1.1rem description + gradient bg + border-bottom`): `ShoppingListView:195-227`, `RecipesView:93-130`, `GenerateMenuView:276-318`, `AccountView`. Extract a `PageHeader` component (props: title, description, optional back-link slot, optional tabs slot). Four near-identical 30-line blocks today → guaranteed drift tomorrow (Generate already deviates: `3rem` padding, centered).
- **Landing loading state** (`LandingView.vue:42-47`): a full-screen spinner with "Laddar…" for what is essentially static content. Landing is the first paint a prospect sees — render hero copy instantly from local defaults and hydrate CMS data in place, or use the skeleton system you already built for the dashboard. A spinner on a marketing page is the worst of both worlds.
- **`GenerateMenuView` header cost on mobile:** `3rem/2rem` header padding + title + subtitle + 2-line description consumes ~40% of a 375px viewport before the first day card. Tighten to title + one line; the empty-state component already explains the flow.

## P3 — Typography

- **No type scale.** `AccountView.vue` alone uses ~20 distinct font sizes (0.75, 0.8, 0.85, 0.9, 0.95, 1.0, 1.05, 1.1, 1.5, 2.5rem…). Adjacent sizes like 0.9 vs 0.95 vs 1.0 carry no hierarchy information — they're noise. Define 7 steps as tokens (e.g. `--text-xs: 0.75rem` … `--text-3xl: 2.5rem`, roughly a 1.2 ratio) and snap everything to them. This is the single biggest "feels more designed" win available.
- **Fraunces is underused as a personality carrier.** It appears in page titles and `WeeklyMenuGrid`'s date numerals (`WeeklyMenuGrid.vue:392` — nice touch). Extend that idea to the app's key numbers: shopping-list progress count ("12 av 34"), portion counts, CountUp stats. Numbers in a food app are content; giving them the display face is a cheap signature.
- Body text `0.65rem` (`MobileBottomNav` labels) and `0.6rem` (`wip-pill`) are below the ~11px legibility floor at default zoom. Bump to 0.7rem minimum.

## P3 — States & content

- **"Tolka recept" WIP tile** (`QuickActions.vue:48-56,63`): shows a disabled action with a "WIP" pill and "Kommer snart" to end users, and `handleParseRecipe` in `DashboardView.vue:74-76` is dead code behind the `wip` guard. Don't advertise unavailable features in a 4-item quick-action list — remove the tile until it ships (the route still exists for direct nav). An empty promise costs trust; three working actions look better than four with one broken.
- **`handleRemoveMember` is a `console.log` TODO** (`DashboardView.vue:90-93`) — the UI exposes a remove-member affordance that silently does nothing. Hide the affordance until the API is wired; a button that no-ops is worse than no button.
- Empty states, error states with retry, and skeletons are consistently present and well-written (imperative, explains next action — "Generera en meny först så skapas din inköpslista automatiskt"). Genuinely good; keep this bar.

## P3 — Mobile & performance

- **Shopping list is the best mobile screen in the app** — tap-target tiles, category tints, count-per-category. Two nits: (1) unchecked tile borders are `--border-color` at 8% alpha — in a bright store, barely visible; use 2px `--border-color-hover` or a subtle fill instead. (2) `.tile:hover { transform: translateY(-2px) }` also fires on touch as a sticky hover; wrap in `@media (hover: hover)`.
- **QuickActions horizontal scroll row** (`QuickActions.vue:208-215`) has no scroll affordance — 4 items at ~72px fit in 375px, but if a 5th is added, overflow will be invisible. Add a fade-out mask or just keep it a 4-max grid.
- **Canvas waves + 2 blurred blobs + grain on Login/Landing** (`LoginView.vue:60-76`): the animated canvas runs continuously on mobile where blur filters are already expensive. Disable `WavesBackground` under 768px (blobs + grain alone carry the mood) — battery matters for a kitchen/store app.
- **Hero card-stack timer ignores reduced motion** (`HeroSection.vue:63-95`): the global CSS kills transition durations, so with `prefers-reduced-motion` the cards *snap* every 4s instead of not cycling. Guard the `setInterval` behind a `matchMedia('(prefers-reduced-motion: reduce)')` check. Same check belongs in `RotatingText` usage if you keep it.

## P3 — Accessibility (beyond the contrast items in P1)

Baseline is strong: focus traps on all three modals, `role="dialog"` + `aria-modal` + labelled titles, `role="alert"` on errors, `aria-live` toast container, global `:focus-visible` treatment, `env(safe-area-inset-bottom)` on the nav. Remaining issues:

- **Nested interactive controls:** `WeeklyMenuGrid.vue:113-175` — each `.day-card` is `role="button" tabindex="0"`, and when selected, `DayActionPopover` (with its own buttons) renders *inside* it. Buttons inside a button are invalid ARIA and make screen-reader/keyboard behavior unpredictable (Enter on a popover control also re-triggers the card handler). Move the popover out of the card node (teleport + anchor positioning), or make the card a plain container whose header is the toggle button.
- **Tabs without keyboard semantics:** `RecipesView.vue:35-58` uses `role="tablist"/tab/tabpanel` but no arrow-key navigation or `tabindex` roving as the ARIA tabs pattern requires. Either implement arrow keys or drop the tab roles and treat them as buttons (honestly fine for 2 options).
- **`ShoppingListView` progress bar** has no `role="progressbar"`/`aria-valuenow`; the text "12 av 34 varor" partly covers it, but tie them with `aria-describedby` or add the role.

---

## Suggested attack order

1. P1 #1–#3 (three small CSS/template fixes, all in the core flows) — one evening.
2. P1 #4–#5 contrast fixes + unify primary CTA — small, user-visible everywhere.
3. Token adoption sweep (radius → category colors → transitions → font-family purge) — mechanical, big consistency payoff.
4. Type scale + `PageHeader` extraction.
5. Dashboard de-gimmick pass + landing loading fix.
6. A11y cleanups (popover nesting, tabs, progressbar).
