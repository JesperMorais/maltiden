<script setup lang="ts">
interface Props {
  width?: string
  height?: string
  radius?: string
  animate?: boolean
}

withDefaults(defineProps<Props>(), {
  width: '100%',
  height: '1rem',
  radius: '12px',
  animate: true
})
</script>

<template>
  <div
    class="skeleton-block"
    :class="{ animated: animate }"
    :style="{ width, height, borderRadius: radius }"
    aria-hidden="true"
  />
</template>

<style scoped>
.skeleton-block {
  background: var(--skeleton-base);
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.skeleton-block.animated::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    var(--shimmer-color) 50%,
    transparent 100%
  );
  transform: translateX(-100%);
  animation: skeleton-shimmer 2s ease-in-out infinite;
  will-change: transform;
}

@keyframes skeleton-shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}
</style>
