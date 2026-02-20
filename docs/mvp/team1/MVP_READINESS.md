# MVP Readiness Report

**Date:** 2026-02-20
**Target:** 2-5 test families

---

## Summary

Maltiden is feature-complete across all core workflows. What remains is UX polish, missing feedback mechanisms, and a few operational gaps — not fundamental feature work.

---

## Scorecard

| Area | Status | MVP-Ready? |
|------|--------|------------|
| Auth (register/login/JWT) | Done | Yes |
| Household management | Done | Yes |
| Recipe browsing & CRUD | Done | Yes |
| AI recipe parsing (Claude) | Done | Yes |
| Weekly menu generation | Done | Yes |
| Shopping list | Done | Yes |
| Grocery offers (Tjek API) | POC only | No (not integrated) |
| Landing page | Done | Yes |
| Onboarding flow | Done | Yes |
| Mobile responsive | Done | Yes |
| CI/CD pipeline | Done | Yes |
| Fly.io deployment | Done | Yes |
| Backend tests | Partial (36+ service tests) | Acceptable |
| Frontend tests | Partial (~2%) | Acceptable |
| Error tracking/monitoring | Missing | Risky |

---

## Core Flows (all working end-to-end)

1. **Registration & Login** — Email/password, bcrypt cost 12, JWT 7-day expiry, role-based access (owner/member/guest)
2. **Household** — Create on register, invite via 8-char codes (7-day expiry), join, member status tracking (isEatingToday, wantsLunchBox), removal with role guards
3. **Recipes** — Browse with search/tags, add manually, AI-parse from raw text via Claude, parse-and-save one-step flow, detail modal
4. **Menu Generation** — Generate 5-day menus, slot machine animation, lock/unlock days, regenerate unlocked only, skip days, custom servings
5. **Shopping List** — Auto-aggregated from menu, scaled by servings, categorized (Kott & Fisk, Mejeri, Frukt & Gront, Skafferi, Ovrigt), checkoff with optimistic updates, progress bar

---

## Gaps

### Critical (must fix before launch)

| # | Gap | Impact | Effort |
|---|-----|--------|--------|
| 1 | No toast/notification system | Users get no feedback on actions (save, generate, errors) | 1-2 days |
| 2 | No logout confirmation or session expiry UX | JWT expires silently after 7 days, user gets dumped to login | 0.5 day |
| 3 | No "forgot password" flow | Test family members locked out if they forget password | 1-2 days |
| 4 | Shopping list has no "clear all" or "new list" action | Stale items persist, confusing for weekly reset | 0.5 day |
| 5 | No error tracking (Sentry) | Won't know when things break for real users | 0.5 day |

### Important (not blocking)

| # | Gap | Impact | Effort |
|---|-----|--------|--------|
| 6 | Offers not integrated into shopping list | Tjek API works as POC but doesn't enhance the core flow | 2-3 days |
| 7 | No recipe editing | Once saved, recipes can't be modified | 1 day |
| 8 | No recipe deletion | Users can't remove bad recipes | 0.5 day |
| 9 | Menu "save" re-generates instead of saving current | TODO in code — save button calls POST /menus/generate again | 1 day |
| 10 | No invite sharing UX | Code displayed in modal but no copy-to-clipboard or share link | 0.5 day |
| 11 | Database backups | No automated backup strategy — volume deletion = data loss | 0.5 day |
| 12 | No onboarding tutorial | New users dropped into dashboard with no guidance | 1-2 days |

### Nice-to-have (post-MVP)

- Recipe preference learning (like/dislike)
- Nutritional info
- PWA/offline support
- OAuth (Google/Apple)
- Custom domain (maltiden.se)
- E2E tests
- Performance profiling (Lighthouse)

---

## Architecture Quality

**Backend: Production-grade**
- Clean layered architecture (handler -> service -> storage -> domain)
- Proper dependency injection
- Transaction safety on all multi-step operations
- Input validation at every layer
- Rate limiting on auth endpoints
- CORS, HTTPS, parameterized SQL

**Frontend: Production-grade**
- TypeScript strict mode with noUncheckedIndexedAccess
- Composition API throughout
- Pinia stores with proper loading/error states
- Typed API layer with mock fallbacks
- Responsive CSS with clamp() fluid typography

---

## Launch Plan

### Phase 0 — Ship Blockers (3-5 days)

1. Add toast notification system
2. Add Sentry error tracking (free tier)
3. Fix menu save to persist current selection (not re-generate)
4. Add copy-to-clipboard on invite codes
5. Add basic session expiry handling UX

### Phase 1 — Soft Launch

- Deploy to maltiden.fly.dev
- Invite 2-3 families manually
- Monitor Sentry + fly logs for first 48 hours

### Phase 2 — Feedback Sprint (weeks 2-3)

- Add recipe edit/delete based on user feedback
- Integrate offers into shopping list if users want it
- Add forgot-password if users report issues
- Polish based on real usage patterns

---

## Live Site

- **URL:** https://maltiden.fly.dev/
- **Health:** `{"status":"ok","db":"connected"}`
- **HTTPS:** Enforced
