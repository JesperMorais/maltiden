import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

// Global theme styles
import './styles/theme.css'

// Directives
import { vPrefetch } from './directives/vPrefetch'

const app = createApp(App)
const pinia = createPinia()

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
