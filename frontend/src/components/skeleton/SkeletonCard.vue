<script setup lang="ts">
interface Props {
  padding?: 'none' | 'sm' | 'md' | 'lg'
  shadow?: boolean
  rounded?: boolean
}

withDefaults(defineProps<Props>(), {
  padding: 'md',
  shadow: true,
  rounded: true
})
</script>

<template>
  <div
    class="skeleton-card"
    :class="[`padding-${padding}`, { shadow, rounded }]"
    aria-hidden="true"
  >
    <slot />
  </div>
</template>

<style scoped>
.skeleton-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
}

/* Padding variants */
.padding-none { padding: 0; }
.padding-sm { padding: 1rem; }
.padding-md { padding: 1.5rem; }
.padding-lg { padding: 2.5rem; }

/* Shadow */
.shadow {
  box-shadow: var(--shadow-sm);
}

/* Rounded */
.rounded {
  border-radius: 24px;
}
</style>
