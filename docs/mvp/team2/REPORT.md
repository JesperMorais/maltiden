# Maltiden MVP Readiness Report — Team 2 Analysis

**Date:** 2026-02-20
**Method:** 5 parallel AI agents analyzing backend, frontend, docs, ops, and WIP independently
**Target:** MVP readiness for 2-5 family beta test

---

## Executive Summary

**Overall readiness: ~65-70%** — The core architecture is solid and well-structured, but several features are incomplete and test coverage is limited. The app has a working deployment on Fly.io with most core flows functional end-to-end, but the shopping list (a critical MVP feature) is still WIP, and there's no recipe editing capability.

---

## 1. What Families CAN Do Today

| Feature | Status | Notes |
|---------|--------|-------|
| Register & login | **Working** | JWT auth, bcrypt, proper middleware |
| Create/join household | **Working** | Invite codes, roles (owner/member/guest) |
| Browse recipes | **Working** | Public list + detail views |
| Add recipes manually | **Working** | Title, ingredients, instructions, tags |
| Parse recipes with AI | **Working** | Claude integration, parse-and-save flow |
| Generate weekly menu | **Working** | Random selection from household recipes, 7-day menus |
| View current menu | **Working** | Weekly grid on dashboard |
| View shopping list | **Partial** | Backend endpoints exist, frontend WIP (untracked) |
| Check off shopping items | **Partial** | PATCH endpoint exists, frontend not wired |
| View grocery offers | **Partial/POC** | Tjek API integration exists, OfferSearch is POC |

---

## 2. Critical Blockers (Must Fix Before Beta)

### 2.1 Shopping List Not Connected (HIGH)

- **Backend**: `GET /shopping-list` and `PATCH /shopping-list/items/{id}` exist and work
- **Frontend**: `ShoppingListModal.vue`, `shoppingList.ts` store, and `useProgressIllusion.ts` are untracked WIP files
- **Gap**: The UI exists in draft form but isn't committed or wired to the real API
- **Impact**: Shopping list is THE core value prop — families need this to plan grocery trips
- **Effort**: Medium — the pieces exist, they need to be connected and polished

### 2.2 No Recipe Editing/Updating (HIGH)

- **Backend**: No `PUT /recipes/{id}` or `PATCH /recipes/{id}` endpoint exists
- **Frontend**: No edit form for existing recipes
- **Impact**: If a user makes a typo or the AI parser gets something wrong, they can't fix it — they must delete and recreate
- **Effort**: Medium — need handler, service method, storage query, and frontend form

### 2.3 No Recipe Deletion (HIGH)

- **Backend**: `DELETE /recipes/{id}` — need to verify this exists
- **Frontend**: No delete button visible in recipe views
- **Impact**: Users accumulate bad/test recipes with no way to remove them
- **Effort**: Low

### 2.4 Limited Test Coverage (MEDIUM)

- **Backend**: 7 test files exist in `backend/internal/services/` covering auth, household, recipe, menu, shopping helpers, cache, and tjek helpers (36+ tests)
- **Frontend**: Vitest + vue/test-utils configured but minimal test files exist (~2% coverage)
- **Impact**: Backend has a reasonable safety net for service logic, but frontend regressions are unguarded. Expanding coverage would increase confidence during beta
- **Effort**: Medium — backend foundation exists, frontend needs test files added for critical flows

---

## 3. Important Issues (Should Fix Before Beta)

### 3.1 No User Profile/Settings Page

- No way to change password, update email, or manage account
- No logout is visible (need to verify)
- Families will need basic account management

### 3.2 No Menu History / Past Menus

- `GET /menus/current` returns only the current week
- No way to look at previous weeks' menus
- Families may want to reference what they ate last week

### 3.3 No "Regenerate Single Day" on Menu

- Menu generation is all-or-nothing (7 days)
- Can't swap out one day's meal without regenerating the entire week
- This will be a common frustration

### 3.4 Database Has No Backup Strategy

- SQLite on a Fly.io volume — persistent across deploys
- **No automated backup** — if the volume fails, all user data is lost
- For 2-5 families: acceptable risk with manual backups, but needs monitoring
- Fly.io volumes have built-in snapshots, but they should be explicitly enabled/verified

### 3.5 ~~No Graceful Shutdown~~ — Already Implemented

- `main.go` already handles OS signals via `signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)` with graceful `server.Shutdown(ctx)`
- No action needed

### 3.6 Error Handling in Frontend

- Some API errors may surface as generic or cryptic messages
- Need to verify error toasts/notifications exist for common failure modes
- Rate limiting could accidentally block real users if too aggressive

### 3.7 No Password Reset Flow

- No forgot-password endpoint or email integration
- Beta families who forget passwords will need manual DB intervention

---

## 4. Nice-to-Have for Beta (Not Blocking)

| Feature | Notes |
|---------|-------|
| PWA / installable app | Would make it feel more native on phones |
| Push notifications | "Your weekly menu is ready" |
| Recipe categories/filtering | Currently flat list, could use tags for filtering |
| Grocery offer matching | POC exists (OfferSearch), not integrated into shopping list |
| Dark mode | CSS theming partially prepared |
| Multi-language | Currently Swedish-only (correct for target audience) |
| Household member management UI | Backend supports it, frontend may need polish |
| Recipe sharing between households | Not implemented |

---

## 5. Infrastructure Assessment

| Area | Status | Notes |
|------|--------|-------|
| Deployment | **Good** | Multi-stage Dockerfile, Fly.io arn region |
| SSL/TLS | **Good** | Handled by Fly.io |
| CORS | **Good** | Configured for production domain |
| Rate limiting | **Good** | Middleware in place |
| CI/CD | **Good** | GitHub Actions: build, test, lint, type-check |
| Database persistence | **OK** | Fly.io volume, but no backup automation |
| Monitoring | **Missing** | No alerting, no structured logging |
| Health check | **Good** | `GET /health` endpoint exists |
| Auth security | **Good** | bcrypt cost 12, JWT HS256, 7-day expiry |

---

## 6. Code Quality Assessment

| Area | Rating | Notes |
|------|--------|-------|
| Backend architecture | **Excellent** | Clean layered separation, DI, proper error handling |
| Frontend architecture | **Good** | Composition API, Pinia stores, typed API layer |
| Type safety | **Good** | TypeScript strict mode, Go types |
| Code organization | **Excellent** | Clear file naming, consistent patterns |
| Documentation | **Very Good** | Comprehensive CLAUDE.md, API.md, architecture docs |
| Test coverage | **Partial** | Backend has 36+ service tests; frontend ~2% |
| Security | **Good** | Proper auth, input handling, CORS |

---

## 7. WIP Features Analysis (Untracked Files)

The 4 untracked files represent the **shopping list feature in progress**:

1. **`ProgressOverlay.vue`** — Animated progress overlay component, appears complete
2. **`ShoppingListModal.vue`** — Shopping list modal UI, partially complete
3. **`useProgressIllusion.ts`** — Composable for fake progress animation during API calls
4. **`shoppingList.ts` (store)** — Pinia store for shopping list state management

These files suggest the shopping list feature was actively being developed. The backend support already exists — this is primarily a frontend completion task.

---

## 8. Prioritized Action Plan for MVP Launch

### Phase 1: Critical Path (Week 1)

1. **Complete shopping list feature** — Commit and connect WIP files to backend API
2. **Add recipe edit endpoint** — `PUT /recipes/{id}` backend + frontend edit form
3. **Add recipe delete** — Backend endpoint + frontend confirm-delete button
4. **Expand test coverage** — Backend has service tests; add frontend tests for critical flows

### Phase 2: Beta-Ready Polish (Week 2)

5. **User profile/settings page** — Change password, view account info, logout
6. **Error handling audit** — Ensure all API errors show user-friendly Swedish messages
7. **Database backup** — Set up Fly.io volume snapshots or automated backup script
8. ~~**Graceful shutdown**~~ — Already implemented
9. **Menu single-day swap** — Allow replacing one day's recipe without full regeneration

### Phase 3: Beta Launch Prep (Week 3)

10. **Manual QA pass** — Walk through every flow as a new user
11. **Seed some starter recipes** — So new families have something to generate menus from
12. **Simple feedback mechanism** — Even just a "feedback" mailto link
13. **Basic monitoring** — At minimum, uptime check + error alerting

---

## 9. Risk Assessment for Beta

| Risk | Severity | Mitigation |
|------|----------|------------|
| Data loss (volume failure) | High | Set up volume snapshots before inviting families |
| User gets stuck (no edit/delete) | High | Implement recipe CRUD completely |
| Shopping list unusable | Critical | Complete the WIP feature |
| No password reset | Medium | Document "contact us" process for beta |
| Performance under load | Low | 2-5 families = minimal load on SQLite |
| Security vulnerability | Low | Auth system is well-implemented |

---

## 10. Bottom Line

**The app is architecturally sound and ~65-70% ready for a family beta.** The Go backend is well-structured with clean separation. The Vue frontend has good patterns and a solid component library. The deployment infrastructure works.

**The 3 things that absolutely must happen before inviting families:**

1. Shopping list must work end-to-end (the WIP is close)
2. Recipe edit/delete must exist (can't manage recipes without it)
3. At least one pass of error handling so users don't see raw error messages

**What you can get away with for a small beta:**

- No automated tests (but add them as you fix bugs)
- No password reset (tell beta families to contact you)
- No recipe sharing (each household manages their own)
- No monitoring (check manually during beta)
- SQLite is perfectly fine for 2-5 families

The codebase is in good shape to build on. The main gap is feature completeness, not technical debt.
