/**
 * Hover Prefetch System
 *
 * Preloads route chunks when user hovers over links.
 * Uses the ~200-400ms between hover and click to start loading,
 * making navigation feel instant.
 */

// Map of route names to their lazy import functions
const routeImports: Record<string, () => Promise<unknown>> = {
  landing: () => import('@/views/LandingView.vue'),
  register: () => import('@/views/OnboardingView.vue'),
  login: () => import('@/views/LoginView.vue'),
  dashboard: () => import('@/views/DashboardView.vue'),
  about: () => import('@/views/AboutView.vue')
}

// Track which routes have been prefetched to avoid duplicate requests
const prefetchedRoutes = new Set<string>()

/**
 * Prefetch a route's chunk by name
 */
export function prefetchRoute(routeName: string): void {
  // Skip if already prefetched or doesn't exist
  if (prefetchedRoutes.has(routeName) || !routeImports[routeName]) {
    return
  }

  // Mark as prefetched immediately to prevent duplicate calls
  prefetchedRoutes.add(routeName)

  // Trigger the dynamic import - browser will cache the chunk
  routeImports[routeName]().catch(() => {
    // If prefetch fails, remove from set so it can retry
    prefetchedRoutes.delete(routeName)
  })
}

/**
 * Prefetch a route by path
 */
export function prefetchPath(path: string): void {
  const routeMap: Record<string, string> = {
    '/': 'landing',
    '/register': 'register',
    '/login': 'login',
    '/dashboard': 'dashboard',
    '/about': 'about'
  }

  const routeName = routeMap[path]
  if (routeName) {
    prefetchRoute(routeName)
  }
}

/**
 * Check if a route has been prefetched
 */
export function isPrefetched(routeName: string): boolean {
  return prefetchedRoutes.has(routeName)
}

/**
 * Composable for prefetch functionality
 */
export function usePrefetch() {
  return {
    prefetchRoute,
    prefetchPath,
    isPrefetched
  }
}
