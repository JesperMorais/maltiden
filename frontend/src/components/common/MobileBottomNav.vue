<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LayoutDashboard, BookOpen, ShoppingCart } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

const tabs = [
  { name: 'dashboard', label: 'Hem', icon: LayoutDashboard, to: '/dashboard' },
  { name: 'recipes', label: 'Recept', icon: BookOpen, to: '/recipes' },
  { name: 'shopping-list', label: 'Handla', icon: ShoppingCart, to: '/shopping-list' },
] as const

const activeTab = computed(() => {
  const name = route.name as string
  if (name === 'shopping-list') return 'shopping-list'
  if (name === 'recipes') return 'recipes'
  return 'dashboard'
})

function navigate(to: string) {
  router.push(to)
}
</script>

<template>
  <nav class="mobile-bottom-nav" aria-label="Mobilnavigation">
    <button
      v-for="tab in tabs"
      :key="tab.name"
      class="nav-tab"
      :class="{ active: activeTab === tab.name }"
      :aria-label="`Gå till ${tab.label}`"
      :aria-current="activeTab === tab.name ? 'page' : undefined"
      @click="navigate(tab.to)"
    >
      <component :is="tab.icon" :size="22" :stroke-width="2" />
      <span class="tab-label">{{ tab.label }}</span>
    </button>
  </nav>
</template>

<style scoped>
.mobile-bottom-nav {
  display: none;
}

@media (max-width: 768px) {
  .mobile-bottom-nav {
    display: flex;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: var(--bottom-nav-height);
    background: var(--bg-card);
    border-top: 1px solid var(--border-color);
    z-index: 1000;
    padding-bottom: env(safe-area-inset-bottom, 0px);
    box-shadow: var(--shadow-sm);
  }

  .nav-tab {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    min-height: 44px;
    -webkit-tap-highlight-color: transparent;
    transition: color 0.2s ease;
  }

  .nav-tab.active {
    color: var(--accent-text);
  }

  .tab-label {
    font-family: 'Nunito', sans-serif;
    font-size: 0.65rem;
    font-weight: 700;
    letter-spacing: 0.01em;
  }
}
</style>
