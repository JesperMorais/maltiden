<script setup lang="ts">
import { computed } from 'vue'
import SkeletonBlock from './SkeletonBlock.vue'

interface Props {
  lines?: number
  gap?: string
  lastLineWidth?: string
  lineHeight?: string
}

const props = withDefaults(defineProps<Props>(), {
  lines: 3,
  gap: '0.65rem',
  lastLineWidth: '60%',
  lineHeight: '0.85rem'
})

const lineWidths = computed(() => {
  return Array.from({ length: props.lines }, (_, i) => {
    if (i === props.lines - 1) return props.lastLineWidth
    // Vary widths between 85% and 100% for a natural look
    const percent = 85 + ((i * 7 + 3) % 16)
    return `${percent}%`
  })
})
</script>

<template>
  <div class="skeleton-text" :style="{ gap }" aria-hidden="true">
    <SkeletonBlock
      v-for="(width, i) in lineWidths"
      :key="i"
      :width="width"
      :height="lineHeight"
      radius="6px"
    />
  </div>
</template>

<style scoped>
.skeleton-text {
  display: flex;
  flex-direction: column;
}
</style>
