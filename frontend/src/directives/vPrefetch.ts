/**
 * v-prefetch Directive
 *
 * Add to any element to prefetch a route on hover.
 * Usage: <RouterLink v-prefetch="'dashboard'" to="/dashboard">
 * Or:    <button v-prefetch:path="'/about'">
 */

import type { Directive } from 'vue'
import { prefetchRoute, prefetchPath } from '@/composables/usePrefetch'

export const vPrefetch: Directive<HTMLElement, string> = {
  mounted(el, binding) {
    const handler = () => {
      if (binding.arg === 'path') {
        // v-prefetch:path="/about"
        prefetchPath(binding.value)
      } else {
        // v-prefetch="'routeName'"
        prefetchRoute(binding.value)
      }
    }

    // Prefetch on hover (mouse)
    el.addEventListener('mouseenter', handler, { once: true, passive: true })

    // Prefetch on touch start (mobile) - users often pause briefly before tapping
    el.addEventListener('touchstart', handler, { once: true, passive: true })

    // Prefetch on focus (keyboard navigation)
    el.addEventListener('focus', handler, { once: true, passive: true })
  }
}

export default vPrefetch
