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
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --coral-tint: rgba(255, 107, 91, 0.03);

  background: linear-gradient(
    165deg,
    var(--warm-white) 0%,
    var(--cream) 100%
  );
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
    rgba(255, 255, 255, 0.8) 0%,
    rgba(255, 248, 240, 0.4) 100%
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
  box-shadow:
    0 4px 20px rgba(61, 44, 41, 0.06),
    0 2px 8px rgba(61, 44, 41, 0.04),
    0 0 0 1px rgba(255, 107, 91, 0.05);
}

.shadow:hover {
  box-shadow:
    0 12px 40px rgba(61, 44, 41, 0.1),
    0 4px 12px rgba(61, 44, 41, 0.06),
    0 0 0 1px rgba(255, 107, 91, 0.1);
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
