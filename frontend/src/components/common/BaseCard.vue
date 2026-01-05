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
  <div class="base-card" :class="[`padding-${padding}`, { shadow, rounded }]">
    <slot />
  </div>
</template>

<style scoped>
.base-card {
  background: linear-gradient(
    165deg,
    var(--bg-primary) 0%,
    var(--bg-secondary) 100%
  );
  border: 1px solid var(--border-color);
  position: relative;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.base-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: linear-gradient(
    165deg,
    var(--bg-card) 0%,
    var(--bg-secondary) 100%
  );
  opacity: 0;
  transition: opacity 0.3s ease;
}

.base-card:hover::before {
  opacity: 1;
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

.shadow:hover {
  box-shadow: var(--shadow-md);
}

/* Rounded */
.rounded {
  border-radius: 24px;
}

/* Ensure content is above the pseudo-element */
.base-card > :deep(*) {
  position: relative;
  z-index: 1;
}
</style>
