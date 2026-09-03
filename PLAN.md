# Måltiden Rework Plan

_Status: draft, 2026-09-03. Owner: Jesper. Branch: `dev-jesper`._

> **How to read this document.** Anything tagged **[UNVERIFIED]** is an
> assumption or recollection that was *not* checked against the code, the
> live APIs, or the law when this was written. Later agents must verify
> before building on it, and should update or remove the tag. Anything
> tagged **[DECISION NEEDED]** is a choice Jesper has not made yet.
> Everything in §3 without a tag was read directly from the repo on
> 2026-09-03 and can be trusted for that date.

## 1. The idea

Måltiden stops being an account-first meal planner and becomes a **local-first
grocery deals app** with an optional paid AI layer.

| Tier | What you get | Account? | Where data lives |
|------|--------------|----------|------------------|
| **Free** | Enter where you live → see this week's cheapest groceries at shops nearby, grouped by store and category. Save deals to a shopping list. Get a "where to go" plan for your shopping trip. | No | On the device (IndexedDB) |
| **Premium** | Everything in Free plus recipes, weekly menu generation, and an AI planner that answers "what can I cook cheaply this week?" using the real local offers as input. | Yes | Server (existing household model) |

The current recipe/menu feature set is not release-quality yet (Jesper's
judgement, not measured). It moves behind Premium where it can mature
without blocking the free launch.

## 2. Delivery form: installable PWA

Ship as a Progressive Web App, not a native app.

- Installable to the phone home screen from the browser, works offline for saved deals.
- Free tier needs **no server-side state**. The backend is an offers service only.
- One codebase (existing Vue 3 frontend). Design system, dark mode, and motion stay as-is.
- If App Store / Play presence is needed later, wrap the same PWA in Capacitor. Do not do this before the free tier has proven itself.
- **[UNVERIFIED]** iOS Safari PWA limits (storage eviction after inactivity, no push without install, geolocation prompt behaviour) were not checked. Verify before promising offline persistence on iPhone.

## 3. Where the codebase stands (2026-09-03, `origin/dev`)

Reusable as-is or nearly:

- `backend/internal/services/tjek_service.go` — fetches Tjek/etilbudsavis catalogs by lat/lng/radius, filters to Swedish grocery chains, in-memory TTL cache with retry.
- `GET /offers/search`, `/offers/discounts`, `/offers/stores` — already public (no auth), already take `lat`, `lng`, `radius`, `exclude`, `q`.
- `frontend/src/components/poc/OfferSearch.vue` + `api/offers.api.ts` + `mocks/offers.mock.ts` — proof of concept, not mounted on any route.
- Design system (`theme.css`, Nunito/Fraunces, Lucide, motion-v).
- Recipe, menu, parser, household, auth, password reset — all working, all move to Premium.

Missing entirely:

- PWA manifest, service worker, install prompt.
- Geolocation or postcode entry. No maps library.
- Store coordinates: `storeResponse` only keeps `street`, `city`, `dealer_id`.
- Offer categories and unit-price normalisation ("cheapest rice per kg").
- Any tier/subscription/payment code.
- Persistent offers storage. Everything is per-request proxying.

Coupling to undo:

- Every view except landing/login/about has `requiresAuth`. Landing funnels into signup.
- Shopping list requires a generated menu (`ShoppingService.GetShoppingList(menuID, householdID)`).
- Dashboard is the authenticated home. It becomes redundant.

Size for reference: backend ~9.5k LOC (+12k tests), frontend ~26k LOC, 21 migrations.

## 4. Data sourcing: Tjek vs own scrapers

**Decision: build around an `OfferSource` interface, start with Tjek, add own adapters per chain.**

Why not just Tjek: third-party terms, rate limits, uneven dealer coverage,
flyer hotspots carry weak structure. **[UNVERIFIED]**: that Matdax is absent
from Tjek (it is simply not in the current `isGroceryStore` include list),
that ICA offers vary per store on Tjek, and that Tjek has no unit-price
field. The existing `hotspotResponse` struct does parse `quantity.size` and
`quantity.unit`, so some unit-price derivation may already be possible.

Why not just scrape: each chain (ICA, Coop, Axfood = Willys/Hemköp, Lidl,
Matdax) is a separate adapter that changes without notice. Needs
scheduling, storage and freshness handling we don't have yet. Legal
exposure lands on us instead of an intermediary.
**[UNVERIFIED]**: that the chains sit behind bot protection (Akamai or
similar) — recollection, not tested. That their offer pages are fed by
JSON endpoints with unit prices — plausible, not inspected. The legal
reading (EU database right, ToS enforceability in Sweden) is a layperson's
summary, not legal advice.

Plan:

1. Define `domain.Offer`, `domain.Store` and a Go interface
   `OfferSource { Fetch(ctx, area) ([]Offer, []Store, error) }`.
2. Wrap the existing Tjek service as the first `OfferSource`.
3. Add a scheduled fetch job that writes into an `offers` + `stores` table
   (SQLite, new migration). API reads from the table, not live from sources.
4. Add own adapters one chain at a time, starting with whichever chain Tjek
   covers worst near test postcodes. Merge by store; the UI never knows the source.
5. Drop Tjek when own adapters cover the chains that matter.

Open question: is the motivation avoiding Tjek's terms, or data quality? If
terms, Tjek exits earlier; if quality, Tjek stays as a fallback source.

## 5. Phases

### Phase 0 — Spike (about 1 week)

Goal: verify the data before designing the UI around it.

- [ ] Confirm Tjek returns store coordinates and opening hours; extend `storeResponse` if so. **[UNVERIFIED]** that Tjek's store objects include them at all; the current struct drops everything but street/city.
- [ ] Run `/offers/stores` for 4–5 real postcodes (Stockholm inner city, suburb, mid-size town, rural). Record which dealers actually appear.
- [ ] Check Matdax specifically. If absent on Tjek, inspect matdax.se for an adapter. **[UNVERIFIED]** whether Matdax publishes offers digitally in a scrapeable form at all.
- [ ] Read Tjek's terms: commercial use, caching, attribution.
- [ ] Inspect ICA, Coop, Willys, Hemköp, Lidl sites: what endpoint feeds their offers pages, is there bot protection, is there a unit price field. Also check whether any chain offers an official/partner API — **[UNVERIFIED]**, none were looked up.
- [ ] Decide `OfferSource` interface shape and the internal `Offer` model (must include unit price + category).

Exit: a short table of chains × source × data quality, and the decision for §4.

### Phase 1 — Free tier core: "Veckans fynd"

Frontend:

- [ ] New default route `/` → `DealsView` (replaces landing-into-signup).
- [ ] First-run location step: browser geolocation with postcode fallback. Store area in IndexedDB. Postcode → lat/lng needs a geocoder: **[DECISION NEEDED]** between a bundled Swedish postcode table, Nominatim/OSM, or a paid API; none evaluated **[UNVERIFIED]**.
- [ ] Deals view: best discounts this week, tabs/grouping by store and by category, unit-price shown. Search across offers (`/offers/search`).
- [ ] "Save" on any deal → local shopping list (IndexedDB). Check-off, clear, share as text. Library choice (`idb`, `localforage`, `dexie`, or raw IndexedDB) is **[DECISION NEEDED]**; `idb` is a suggestion only.
- [ ] PWA: manifest, icons, service worker, install prompt, offline shell + cached last result. `vite-plugin-pwa` is the suggested tool, **[UNVERIFIED]** that its current version supports Vite 7.
- [ ] Mock mode (`npm run dev:mock`) updated for the new flow.

Backend:

- [ ] `OfferSource` interface + Tjek adapter + offers/stores tables + fetch job (from §4).
- [ ] Category tagging and unit-price normalisation as a rules file with tests. Expect this to be maintained continuously.
- [ ] Cache/fetch keyed on a rounded geo-cell so neighbours share results. The ~1 km cell size is a guess **[UNVERIFIED]**; tune against how Tjek's radius parameter actually behaves.
- [ ] Keep endpoints public. Rate-limit by IP.

Exit: someone with no account installs the app, enters a postcode, sees real deals, saves a list, and it survives an app restart offline.

### Phase 2 — Where to go

- [ ] Per-store shopping plan: saved list grouped by store, with cheapest-store suggestion for each item.
- [ ] Deep link to Google Maps / Apple Maps with stops in order. No in-app map yet. **[UNVERIFIED]**: Google Maps URLs support multi-stop via `waypoints`; whether Apple Maps URL scheme supports more than one intermediate stop was not checked. Route *optimisation* (order of stops) would need our own logic or a routing API — **[DECISION NEEDED]** whether that is in scope.
- [ ] Optional later: MapLibre + OSM tiles in-app if users ask.

Exit: a Sunday plan you can follow from the phone.

### Phase 3 — Accounts become optional

- [ ] Move login/register behind a "Premium" entry (account tab or upsell in deals view).
- [ ] Remove `requiresAuth` from free views; keep on recipes, menus, household, parser.
- [ ] Retire `DashboardView` as home. Keep household model unchanged for Premium.
- [ ] Onboarding flow becomes premium onboarding only.

### Phase 4 — Premium: AI planner

- [ ] Add `tier` claim to JWT; middleware `RequireTier("premium")`. **[DECISION NEEDED]**: tier per user or per household? Household is the current data model, so household is the simpler default, but it means one subscription covers all members.
- [ ] Bridge offers → recipes: feed the week's local offers into the Claude prompt for menu generation ("cook cheap this week").
- [ ] Recipe/menu/shopping list features finish maturing here before any paid launch.
- [ ] Billing: launch with invite code / beta flag. Add payments only once there is demand. Stripe is the suggestion; **[UNVERIFIED]** whether Swish or Klarna would convert better for a Swedish consumer audience, and no pricing has been decided **[DECISION NEEDED]**.

### Phase 5 — Cleanup

- [ ] Remove dead landing/onboarding components, `counter.ts` store, mounted POC.
- [ ] Update `CLAUDE.md` project structure, routes, and env vars.
- [ ] Rewrite README product framing.

## 6. What gets removed or parked

- Landing page that funnels into signup.
- Menu-dependent shopping list as the only shopping list. In Free, the saved-deals list *is* the shopping list. In Premium, menu ingredients feed into the same list.
- Dashboard as home screen.
- `requiresMember` gating on anything in the free tier.

## 7. Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Tjek ToS forbids commercial use / caching | Free tier source disappears | Phase 0 check; `OfferSource` interface makes swap cheap |
| Scraper adapters break (bot protection, redesigns) | Stale or missing chains | One adapter per chain, health check per source, Tjek as fallback |
| Uneven flyer coverage (ICA per-store, small chains absent) | Some postcodes look thin | Show "stores we cover here" honestly; postcode spike picks realistic launch areas |
| Unit-price normalisation from messy headings | Wrong "cheapest" claims | Rules file + tests; show raw offer text alongside computed unit price |
| Geolocation prompt on first visit hurts conversion | Users bounce | Postcode first, geolocation as an option |
| Legal (EU database right, chain ToS) | Takedown risk | Store only price/offer facts, attribute source, respect robots.txt |
| Premium never converts | AI cost with no revenue | Invite-code beta before payments; measure demand first |

## 8. Decisions log

- 2026-09-03: PWA over native. Free tier stores nothing server-side.
- 2026-09-03: `OfferSource` interface; Tjek first, own scrapers incrementally. Open: does Tjek stay as fallback?
- 2026-09-03: Payments deferred behind invite-code beta.

## 9. Open items for later agents

Collected from the tags above so nothing is lost:

1. Verify Tjek store objects: coordinates, opening hours, per-store vs per-dealer offers.
2. Verify Tjek ToS for commercial use and caching.
3. Verify Matdax presence on Tjek and on its own site.
4. Inspect each chain's offers endpoint: structure, unit price, bot protection, official API.
5. Confirm `vite-plugin-pwa` compatibility with Vite 7; pick IndexedDB library.
6. Check iOS Safari PWA storage/geolocation behaviour.
7. Check Apple Maps multi-stop URL support; decide if route optimisation is in scope.
8. Decide tier per user vs per household; decide payment provider and price.
9. Decide whether Tjek stays as a fallback source or exits once own adapters exist.
10. Legal review of scraping/database right before own adapters go to production.
11. Pick a postcode geocoding approach for the location fallback.
