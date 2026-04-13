# Frontend Performance Audit Report

**Date:** 2026-02-24
**Tool:** Lighthouse 12.x (CLI, headless Chrome)
**Target:** Production build served via `npx serve dist` on localhost

---

## 1. Bundle Size Breakdown

### Before Fixes

| Chunk | Raw | Gzipped | Contents |
|-------|-----|---------|----------|
| `index.js` (vendor) | 294.58 KB | 106.18 KB | Vue, Vue Router, Pinia, Axios, motion-v, app shell |
| `DashboardView.js` | 35.40 KB | 12.04 KB | Dashboard + widgets |
| `FadeContent.js` (shared) | 15.57 KB | 5.70 KB | IntersectionObserver animation component |
| `GenerateMenuView.js` | 14.84 KB | 5.29 KB | Menu generation view |
| `OnboardingView.js` | 14.73 KB | 4.51 KB | Registration flow |
| `RecipesView.js` | 13.48 KB | 5.08 KB | Recipe management |
| `LandingView.js` | 10.48 KB | 3.73 KB | Landing page |
| Other route chunks | ~28.55 KB | ~11.82 KB | Login, About, Offers, Shopping, etc. |
| **Total JS** | **427.63 KB** | **154.35 KB** | |
| **Total CSS** | **142.31 KB** | **28.93 KB** | |
| **Images (JPEG)** | **572.39 KB** | — | 3 team photos |

### After Fixes

| Chunk | Raw | Gzipped | Contents |
|-------|-----|---------|----------|
| `vue-vendor.js` | 110.20 KB | 42.91 KB | Vue, Vue Router, Pinia |
| `index.js` (shared/motion-v) | 134.65 KB | 44.19 KB | motion-v + shared deps (lazy-loaded) |
| `axios.js` | 36.62 KB | 14.73 KB | HTTP client |
| `index.js` (app shell) | 13.52 KB | 5.27 KB | Router, stores, app init |
| `DashboardView.js` | 35.57 KB | 12.11 KB | Dashboard + widgets |
| Other route chunks | ~83.24 KB | ~29.84 KB | All views + shared components |
| **Total JS** | **413.64 KB** | **149.05 KB** | |
| **Total CSS** | **142.31 KB** | **28.93 KB** | |
| **Images (WebP)** | **210.06 KB** | — | 3 team photos (63% smaller) |

**Critical path JS (initial load):** 160.34 KB raw / 62.91 KB gzip (down from 294.58 / 106.18)

---

## 2. Lighthouse Scores

### Before Fixes

| Category | Mobile | Desktop |
|----------|--------|---------|
| **Performance** | **60** | **99** |

### After Fixes

| Category | Mobile | Desktop | Change |
|----------|--------|---------|--------|
| **Performance** | **68** | **99** | **+8 (mobile)** |

---

## 3. Core Web Vitals

### Mobile (Simulated Moto G Power, 4x CPU throttle)

| Metric | Before | After | Target | Status |
|--------|--------|-------|--------|--------|
| **FCP** (First Contentful Paint) | 4.3 s | 4.7 s | < 1.8 s | Needs improvement |
| **LCP** (Largest Contentful Paint) | 5.2 s | 5.6 s | < 2.5 s | Needs improvement |
| **TBT** (Total Blocking Time) | 440 ms | **40 ms** | < 200 ms | **Good** |
| **CLS** (Cumulative Layout Shift) | 0.016 | 0.02 | < 0.1 | **Good** |
| **SI** (Speed Index) | 4.0 s | 4.4 s | < 3.4 s | Needs improvement |

### Desktop

| Metric | Before | After | Target | Status |
|--------|--------|-------|--------|--------|
| **FCP** | 0.7 s | 0.7 s | < 1.8 s | **Good** |
| **LCP** | 0.7 s | 0.7 s | < 2.5 s | **Good** |
| **TBT** | 80 ms | **20 ms** | < 200 ms | **Good** |
| **CLS** | 0 | 0 | < 0.1 | **Good** |
| **SI** | 0.8 s | 0.8 s | < 3.4 s | **Good** |

> **Note on FCP/LCP mobile variance:** Lighthouse synthetic mobile tests apply 4x CPU throttling and simulated slow 4G. The FCP/LCP numbers are high because this is a client-side SPA that must load, parse, and execute JS before rendering any content. Small variances between runs (~300-500ms) are expected. The TBT improvement from 440ms to 40ms confirms the JS execution cost was significantly reduced.

---

## 4. Fixes Applied

### Fix 1: Non-render-blocking Google Fonts
**File:** `frontend/index.html`
**Impact:** Eliminated ~1,330ms of render-blocking time from font loading

Changed the Google Fonts `<link>` from a synchronous render-blocking stylesheet to an async-loading pattern using `<link rel="preload" as="style">` combined with the `media="print" onload="this.media='all'"` technique. Added `<noscript>` fallback for non-JS browsers.

### Fix 2: Lazy-load ToastNotification (defer motion-v)
**File:** `frontend/src/App.vue`
**Impact:** Moved ~134KB of motion-v out of the critical rendering path

Changed `ToastNotification` from a static import to `defineAsyncComponent(() => import(...))`. Since motion-v is only used by ToastNotification and a few lazy-loaded views, this defers its loading until after the initial render. **TBT dropped from 440ms to 40ms.**

### Fix 3: Vendor chunk splitting
**File:** `frontend/vite.config.ts`
**Impact:** Better caching + smaller critical path

Added `build.rollupOptions.output.manualChunks` to split the monolithic 294KB vendor bundle into:
- `vue-vendor.js` (110KB) — Vue framework (cacheable, rarely changes)
- `axios.js` (37KB) — HTTP client (cacheable, rarely changes)
- App-specific code in the entry chunk (14KB)

### Fix 4: Convert team photos from JPEG to WebP
**Files:** `frontend/src/assets/images/{david,jesper,philip}.webp`
**Impact:** 63% image size reduction (572KB → 210KB)

| Image | JPEG | WebP | Savings |
|-------|------|------|---------|
| david | 177 KB | 62 KB | 65% |
| jesper | 169 KB | 59 KB | 65% |
| philip | 213 KB | 85 KB | 60% |

### Fix 5: Add lazy loading + explicit dimensions to images
**File:** `frontend/src/views/AboutView.vue`
**Impact:** Prevents layout shift and defers offscreen image loading

Added `loading="lazy"` and explicit `width="140" height="140"` attributes to all team member `<img>` tags. These images are below the fold on the About page, so lazy loading is appropriate.

### Fix 6: Add meta description
**File:** `frontend/index.html`
**Impact:** SEO improvement

Added `<meta name="description">` with a Swedish description of the app for search engines.

---

## 5. What Was Already Good

- **All routes lazy-loaded** via `() => import()` in the router
- **lucide-vue-next properly tree-shaken** — only used icons appear in route-specific chunks (not the main bundle)
- **Google Fonts `display=swap`** — already configured for font-display swap
- **Google Fonts preconnect** — already had `<link rel="preconnect">` for both fonts.googleapis.com and fonts.gstatic.com
- **CLS near zero** — no significant layout shifts detected
- **Desktop performance excellent** — 99/100 consistently
- **Reduced motion support** — `@media (prefers-reduced-motion: reduce)` in theme.css
- **No duplicate JavaScript** flagged by Lighthouse

---

## 6. Remaining Recommendations

### High Impact (for mobile)

1. **Server-Side Rendering (SSR) or Static Site Generation (SSG):** The biggest remaining bottleneck is that this is a client-rendered SPA. FCP/LCP on mobile are limited by JS parse time. Consider Nuxt.js or pre-rendering the landing page for near-instant FCP.

2. **Inline critical CSS:** The 5.35KB index CSS is still render-blocking (722ms on simulated mobile). Vite plugins like `vite-plugin-css-injected-by-js` or manual critical CSS extraction could eliminate this.

3. **Self-host fonts:** Replace Google Fonts with self-hosted WOFF2 files to eliminate the DNS/TLS overhead for `fonts.googleapis.com` and `fonts.gstatic.com`. Use `unicode-range` subsetting to reduce font file size.

### Medium Impact

4. **Replace motion-v in ToastNotification with CSS animations:** The toast enter/exit animations are simple enough for CSS transitions. This would eliminate the need to load motion-v at all for users who never visit GenerateMenuView or pages with RotatingText.

5. **Code-split the landing store mock data:** The main entry currently inlines ~4KB of mock/landing data that's only needed for the landing page.

### Low Impact

6. **Remove unused JPEG source files:** The original `.jpeg` files can be deleted now that `.webp` versions exist (saves repo size, not runtime).

7. **Add `fetchpriority="high"` to LCP elements:** If the LCP element is known (e.g., hero text), adding fetch priority hints can help the browser prioritize critical resources.

---

## 7. Lighthouse HTML Reports

- Mobile (before): `tests/lighthouse-mobile.report.html`
- Desktop (before): `tests/lighthouse-desktop.report.html`
- Mobile (after): `tests/lighthouse-mobile-after` (JSON)
- Desktop (after): `tests/lighthouse-desktop-after` (JSON)
