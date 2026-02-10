# Coding Conventions

**Analysis Date:** 2026-02-06

## Language Split

This is a **Go + Vue 3/TypeScript** full-stack monorepo. Conventions differ by language. Both sides follow standard community patterns with minimal custom configuration.

---

## Naming Patterns

### Go (Backend)

**Files:**
- Use `snake_case.go` for all Go source files
- Name files by the domain entity or concern they handle: `auth_service.go`, `recipe_storage.go`, `household_service_test.go`
- Test files use the `_test.go` suffix, co-located with the code under test
- Examples: `backend/internal/services/auth_service.go`, `backend/internal/storage/sqlite/user_storage.go`

**Packages:**
- Short, lowercase, single-word package names: `domain`, `handlers`, `services`, `sqlite`, `middleware`, `utils`
- No underscores or hyphens in package names
- Package path mirrors directory: `backend/internal/storage/sqlite` = package `sqlite`

**Structs:**
- PascalCase for exported types: `AuthService`, `RecipeStorage`, `HouseholdHandler`
- Follow the pattern `{Entity}{Layer}` for layer-specific types: `UserStorage`, `RecipeService`, `AuthHandler`
- Domain types are plain entity names: `User`, `Recipe`, `Household`, `Ingredient`
- Request/Response types use `{Action}{Entity}Request`/`{Action}{Entity}Response`: `CreateRecipeRequest`, `AuthResponse`, `JoinHouseholdResponse`
- Internal (unexported) types use camelCase: `catalogResponse`, `hotspotResponse` (in `backend/internal/services/tjek_service.go`)

**Functions:**
- PascalCase for exported: `NewAuthService`, `GetByEmail`, `GenerateToken`
- camelCase for unexported: `generateInviteCode`, `parseTjekTime`, `isGroceryStore`
- Constructor pattern: `New{Type}(deps) *{Type}` -- always returns a pointer
- Method receivers: single lowercase letter `s` for services/storage, `h` for handlers: `func (s *AuthService) Register(...)`, `func (h *AuthHandler) Login(...)`

**Variables:**
- camelCase: `userID`, `householdID`, `authService`
- Short names for loop variables and small scopes: `r` for recipe, `m` for member, `err` for error
- Context keys use typed constants: `const UserIDKey contextKey = "user_id"` (in `backend/pkg/middleware/auth.go`)

**Entity IDs:**
- Prefixed with entity abbreviation: `usr_`, `hh_`, `hm_`, `inv_`, `rec_`
- Generated via `"{prefix}" + uuid.New().String()`
- See: `backend/internal/services/auth_service.go` lines 46-47

### TypeScript/Vue (Frontend)

**Files:**
- Vue components: `PascalCase.vue` -- `BaseButton.vue`, `DashboardHeader.vue`, `WeeklyMenuGrid.vue`
- TypeScript files: `camelCase.ts` -- `client.ts`, `token.ts`, `usePrefetch.ts`
- API service files: `{entity}.api.ts` -- `auth.api.ts`, `menu.api.ts`, `shopping.api.ts`
- Type files: `{entity}.types.ts` -- `dashboard.types.ts`, `offers.types.ts`, `error.types.ts`
- Mock files: `{entity}.mock.ts` -- `dashboard.mock.ts`, `auth.mock.ts`
- Store files: `{entity}.ts` in `stores/` -- `user.ts`, `dashboard.ts`, `menuGenerator.ts`
- Composable files: `use{Feature}.ts` -- `usePrefetch.ts`
- Directive files: `v{Name}.ts` -- `vPrefetch.ts`

**Components:**
- Multi-word PascalCase names: `BaseButton`, `DashboardHeader`, `MenuDayCard`
- Prefix reusable components with `Base`: `BaseButton.vue`, `BaseCard.vue`
- Group by feature in subdirectories: `components/dashboard/`, `components/menu/`, `components/landing/`, `components/common/`

**Functions:**
- camelCase: `fetchDashboard`, `handleLogout`, `toggleDayLock`
- Event handlers prefixed with `handle`: `handleDayClick`, `handleMealClick`, `handleSettings`
- Store actions are plain verbs: `login`, `register`, `clearError`, `fetchDashboard`
- Composables prefixed with `use`: `usePrefetch`, `useUserStore`, `useDashboardStore`

**Types/Interfaces:**
- PascalCase interfaces: `LoginRequest`, `AuthResponse`, `MenuDay`, `DraftMenuDay`
- Type aliases with PascalCase: `UserRole`, `ApiErrorCode`
- Union type literals for enums: `type UserRole = 'owner' | 'member' | 'guest'`

**Variables:**
- camelCase: `isLoading`, `currentUser`, `dashboardData`
- Boolean refs prefixed with `is`/`has`/`wants`: `isAuthenticated`, `hasTodaysMeal`, `isGenerating`
- Computed getters mirror the property: `const todaysMeal = computed(() => ...)`

---

## Code Style

### Go Formatting
- **Tool:** `gofmt` (standard Go formatter, implied by `go vet` in CI)
- **Indentation:** Tabs (Go standard)
- **No custom linter config:** No `.golangci.yml` present. CI runs `go vet` and `govulncheck` only
- **Line length:** No enforced limit; lines tend to be <120 chars

### TypeScript/Vue Formatting
- **Tool:** Prettier via `frontend/.prettierrc.json`
- **Key settings:**
  - `semi: false` -- no semicolons
  - `singleQuote: true` -- single quotes for strings
  - `printWidth: 100` -- 100 char line width
- **Indentation:** 2 spaces (Prettier default)
- **Linting:** ESLint 9 flat config via `frontend/eslint.config.ts`
  - Vue essential rules (`pluginVue.configs['flat/essential']`)
  - TypeScript recommended via `vueTsConfigs.recommended`
  - Custom rule: `@typescript-eslint/no-unused-vars` with `argsIgnorePattern: '^_'` (prefix unused args with `_`)
  - Prettier formatting skipped by ESLint (handled by Prettier separately)
- **Run commands:**
  - `npm run lint` -- ESLint with `--fix --cache`
  - `npm run format` -- Prettier write mode

### Vue Component Style
- Always use `<script setup lang="ts">` (Composition API with `<script setup>`)
- Order within SFC: `<script setup>` → `<template>` → `<style scoped>`
- All styles use `scoped` CSS (no CSS modules, no Tailwind)
- CSS uses CSS custom properties (design tokens) from `frontend/src/styles/theme.css`
- Example pattern from `frontend/src/components/common/BaseButton.vue`

---

## Import Organization

### Go
- **Standard library imports** first, blank line, then **internal packages**, blank line, then **third-party packages**
- Example from `backend/internal/services/auth_service.go`:
```go
import (
    "errors"
    "maltiden/internal/domain"
    "maltiden/internal/storage/sqlite"
    "maltiden/pkg/utils"
    "time"

    "github.com/google/uuid"
)
```
- Note: The Go standard grouping (stdlib + internal mixed, then third-party) is generally followed but stdlib and internal imports are sometimes in the same block

### TypeScript
- **Order observed:**
  1. Vue/framework imports (`vue`, `vue-router`, `pinia`)
  2. Internal module imports using `@/` alias
  3. Relative imports (sibling files)
- **Path alias:** `@/` maps to `frontend/src/` (configured in `frontend/tsconfig.app.json` and `frontend/vite.config.ts`)
- Use `import type` for type-only imports: `import type { MenuDay } from '@/api/types/dashboard.types'`
- Example from `frontend/src/stores/dashboard.ts`:
```typescript
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { DashboardData, Meal, MenuDay, Household, ShoppingListSummary } from '@/api/types/dashboard.types'
import { mockDashboardData } from '@/mocks/dashboard.mock'
import { useUserStore } from './user'
import apiClient from '@/api/client'
```

---

## Error Handling

### Go Patterns
- **Standard `if err != nil` pattern** throughout all backend code
- Return `(result, error)` tuples from services: `func (s *AuthService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error)`
- **Error strings** are short snake_case identifiers: `errors.New("invalid_code")`, `errors.New("already_member")`, `errors.New("not_found")`, `errors.New("cannot_remove")`
- **Handlers map error strings to HTTP status codes** via switch statements:
```go
switch err.Error() {
case "invalid_code":
    http.Error(w, `{"error":"invalid_code"}`, http.StatusBadRequest)
case "already_member":
    http.Error(w, `{"error":"already_member"}`, http.StatusConflict)
default:
    http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
}
```
- See: `backend/internal/api/handlers/household.go` lines 82-89
- **JSON error responses** are inline string literals: `` http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized) ``
- **`sql.ErrNoRows`** returns `(nil, nil)` instead of an error (not-found is not an error): see `backend/internal/storage/sqlite/user_storage.go` line 39
- **Wrapped errors** use `fmt.Errorf("context: %w", err)` for infrastructure errors: see `backend/internal/services/household_service.go` line 87

### TypeScript Patterns
- **try/catch/finally** in all async store actions with loading state management:
```typescript
isLoading.value = true
error.value = null
try {
    // ... async work
} catch (e) {
    error.value = e instanceof Error ? e.message : 'Fallback message'
} finally {
    isLoading.value = false
}
```
- See: `frontend/src/stores/user.ts` lines 79-100
- **Standardized error types** in `frontend/src/api/types/error.types.ts`:
  - `ApiError` interface with `status`, `code`, `message`
  - `parseAxiosError()` utility converts Axios errors to `ApiError`
  - `isApiError()` type guard
  - Error messages are in Swedish
- **API client interceptor** handles 401 globally (auto-redirect to login): `frontend/src/api/client.ts` lines 42-61

---

## Logging

### Go
- **Framework:** Standard library `log` package only
- **Pattern:** `log.Printf()` for informational, `log.Fatal()` for startup failures
- Minimal logging overall -- only in `backend/cmd/server/main.go`

### TypeScript
- **Framework:** `console` (browser built-in)
- **Patterns:**
  - `console.error('Failed to generate menu:', e)` for caught errors in stores
  - `console.warn('Using mock dashboard data:', e)` for fallback behavior
  - `console.log('Day clicked:', day)` for TODO placeholder handlers
  - `console.error('Network error - API may be unavailable')` in API client

---

## Comments

### Go Comments
- **GoDoc-style** on exported functions: `// CreateInvite generates an 8-character invite code valid for 7 days.`
- Short inline comments for non-obvious logic: `// no 0/O/1/I to avoid confusion`
- Section comments separate concerns: `// Public routes`, `// Protected routes`
- TODO format: `//TODO: Better error handling` (see `backend/internal/api/handlers/auth.go`)

### TypeScript Comments
- **JSDoc `/** */` blocks** on all exported functions and modules:
```typescript
/**
 * Login with email and password
 */
export async function login(email: string, password: string): Promise<AuthResponse> {
```
- **Module-level doc comments** at top of files:
```typescript
/**
 * Auth API Service
 * Handles login, register, and authentication endpoints
 */
```
- **Section dividers** in type files using `// ============================================`
- **TSDoc `/** */`** on interface properties in type definition files:
```typescript
export interface MenuDay {
  /** ISO date string (YYYY-MM-DD) */
  date: string
  /** Whether this is today */
  isToday: boolean
}
```
- **TODO format:** `// TODO: Description` (see `frontend/src/views/DashboardView.vue`)
- **Inline comments** for store organization: `// State`, `// Computed`, `// Actions`

---

## Function Design

### Go
- **Small, focused functions** -- most are 10-30 lines
- **Single responsibility** per function: handlers parse request + call service + write response
- **Consistent handler pattern:**
  1. Extract auth from context
  2. Parse request body
  3. Call service
  4. Handle error (switch on error string)
  5. Write JSON response
- **No validation library** -- manual validation in services with early returns

### TypeScript
- **Store actions** follow loading/error/try-catch pattern consistently
- **API service functions** are standalone exported async functions (not class methods)
- **Mock-first pattern** in API services: every function checks `USE_MOCKS` before making HTTP calls
- **Computed getters** are favored over methods for derived state in stores

---

## Module Design

### Go
- **No interfaces defined** -- services depend on concrete storage types directly
- **Constructor injection** via `New{Type}(deps)` functions
- **Router wires everything** in `backend/internal/api/router.go`: storage → service → handler → mux
- **Clean layer separation:** domain types → storage → services → handlers

### TypeScript
- **Named exports** preferred: `export async function login(...)`, `export const useUserStore = defineStore(...)`
- **Default exports** only for the Axios client instance: `export default apiClient` in `frontend/src/api/client.ts`
- **Barrel file** for mocks: `frontend/src/mocks/index.ts` re-exports all mock data and the `USE_MOCKS` flag
- **No barrel files** for components, stores, or API modules -- import directly from specific files
- **API types** separated from API service files: types in `frontend/src/api/types/`, services in `frontend/src/api/`

---

## Vue Component Patterns

**Props pattern** -- Use `interface Props` + `withDefaults(defineProps<Props>(), { ... })`:
```typescript
interface Props {
  variant?: 'primary' | 'secondary' | 'outline'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
}

withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false
})
```
See: `frontend/src/components/common/BaseButton.vue`

**Emits pattern** -- Use typed `defineEmits<{ eventName: [payload] }>()`:
```typescript
const emit = defineEmits<{
  'day-click': [day: MenuDay]
}>()
```
See: `frontend/src/components/dashboard/WeeklyMenuGrid.vue`

**Store usage pattern** -- Composition function stores (`defineStore('name', () => { ... })`), not Options API stores. Always organize return as `State → Computed → Actions`:
```typescript
return {
  // State
  currentUser,
  isLoading,
  // Computed
  userName,
  isOwner,
  // Actions
  login,
  logout
}
```
See: `frontend/src/stores/user.ts`

---

## CSS Conventions

- **Scoped styles only** -- all `<style scoped>` blocks
- **CSS custom properties** (design tokens) from `frontend/src/styles/theme.css`
- **BEM-lite naming** -- flat class names with hyphens: `.dashboard-header`, `.menu-title`, `.day-card`
- **Responsive breakpoints:** `768px` (mobile), `1024px` (tablet)
- **Animation patterns:** CSS transitions with cubic-bezier easing, `Transition` component for enter/leave
- **No utility classes or CSS framework** -- all custom CSS

---

## Mock System

- **Flag:** `USE_MOCKS` in `frontend/src/mocks/index.ts` -- enabled in dev mode unless `VITE_USE_REAL_API=true`
- **Pattern:** Every API service function checks `USE_MOCKS` at the top and returns mock data directly
- **Mock files:** Co-located in `frontend/src/mocks/` with `{entity}.mock.ts` naming
- **Dev command:** `npm run dev:mock` runs Vite in mock mode

---

*Convention analysis: 2026-02-06*
*Update when patterns change*
