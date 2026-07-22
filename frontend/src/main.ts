import { createApp } from 'vue'
import { createPinia } from 'pinia'
import * as Sentry from '@sentry/vue'

import App from './App.vue'
import router from './router'

// Global theme styles
import './styles/theme.css'

// Directives
import { vPrefetch } from './directives/vPrefetch'

const app = createApp(App)
const pinia = createPinia()

// Initialize Sentry in production when a DSN is configured. The DSN is
// build-time baked via Vite — set VITE_SENTRY_DSN_FE in CI before `npm run
// build`. In dev or without a DSN, Sentry stays silent.
if (import.meta.env.PROD && import.meta.env.VITE_SENTRY_DSN_FE) {
  Sentry.init({
    app,
    dsn: import.meta.env.VITE_SENTRY_DSN_FE,
    environment: import.meta.env.MODE,
    integrations: [Sentry.browserTracingIntegration({ router })],
    tracesSampleRate: 0.2,
    tracePropagationTargets: ['localhost', /^https:\/\/maltiden\.fly\.dev/],
    sendDefaultPii: false,
  })
}

app.use(pinia)
app.use(router)

// Register global directives
app.directive('prefetch', vPrefetch)

// Initialize auth before router to prevent redirect to /login on refresh
import { useUserStore } from './stores/user'
useUserStore(pinia).initFromToken()

// Initialize theme before mounting to prevent flash
import { useThemeStore } from './stores/theme'
const themeStore = useThemeStore(pinia)
themeStore.initTheme()

app.mount('#app')
