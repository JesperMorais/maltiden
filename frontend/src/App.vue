<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import { RouterView } from 'vue-router'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const ToastNotification = defineAsyncComponent(() =>
  import('@/components/common/ToastNotification.vue')
)
const FeedbackWidget = defineAsyncComponent(() =>
  import('@/components/common/FeedbackWidget.vue')
)
const MobileBottomNav = defineAsyncComponent(() =>
  import('@/components/common/MobileBottomNav.vue')
)
</script>

<template>
  <a href="#main" class="skip-link">Hoppa till huvudinnehåll</a>
  <main id="main">
    <RouterView v-slot="{ Component, route }">
      <Transition :name="route.meta.transition as string || 'page-fade'" mode="out-in">
        <component :is="Component" :key="route.path" />
      </Transition>
    </RouterView>
  </main>
  <div v-if="userStore.isAuthenticated" class="bottom-nav-spacer" />
  <ToastNotification />
  <FeedbackWidget v-if="userStore.isAuthenticated" />
  <MobileBottomNav v-if="userStore.isAuthenticated" />
</template>

<style scoped>
.skip-link {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
  text-decoration: none;
}

.skip-link:focus {
  position: fixed;
  top: var(--space-sm);
  left: var(--space-sm);
  width: auto;
  height: auto;
  clip: auto;
  padding: var(--space-sm) var(--space-md);
  background: var(--color-accent);
  color: var(--color-bg);
  border-radius: var(--radius-sm);
  z-index: 9999;
  font-family: 'Nunito', system-ui, sans-serif;
  font-weight: 700;
}
</style>

<style>
/* Global reset and base styles */
*,
*::before,
*::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html {
  scroll-behavior: smooth;
}

body {
  min-height: 100vh;
  font-family: 'Nunito', system-ui, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Bottom nav spacer — prevents content from being hidden behind fixed nav */
.bottom-nav-spacer {
  display: none;
}

@media (max-width: 768px) {
  .bottom-nav-spacer {
    display: block;
    height: var(--bottom-nav-height);
  }
}

/* Page transitions */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* Slide transition for onboarding */
.page-slide-enter-active,
.page-slide-leave-active {
  transition: opacity 0.4s ease, transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-slide-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

.page-slide-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

/* Scale transition */
.page-scale-enter-active,
.page-scale-leave-active {
  transition: opacity 0.35s ease, transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.page-scale-enter-from {
  opacity: 0;
  transform: scale(0.95);
}

.page-scale-leave-to {
  opacity: 0;
  transform: scale(1.02);
}
</style>
