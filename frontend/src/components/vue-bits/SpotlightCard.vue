<template>
  <div
    ref="divRef"
    @mousemove="handleMouseMove"
    @focus="handleFocus"
    @blur="handleBlur"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
    class="spotlight-card"
    :class="className"
  >
    <div
      class="spotlight-overlay"
      :style="{
        opacity,
        background: `radial-gradient(circle at ${position.x}px ${position.y}px, ${spotlightColor}, transparent 80%)`,
      }"
    />

    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, useTemplateRef } from 'vue'

interface SpotlightCardProps {
  className?: string
  spotlightColor?: string
}

const { className = '', spotlightColor = 'rgba(255, 107, 91, 0.15)' } =
  defineProps<SpotlightCardProps>()

const divRef = useTemplateRef<HTMLDivElement>('divRef')
const isFocused = ref(false)
const position = ref({ x: 0, y: 0 })
const opacity = ref(0)

const handleMouseMove = (e: MouseEvent) => {
  if (!divRef.value || isFocused.value) return
  const rect = divRef.value.getBoundingClientRect()
  position.value = { x: e.clientX - rect.left, y: e.clientY - rect.top }
}

const handleFocus = () => {
  isFocused.value = true
  opacity.value = 0.6
}

const handleBlur = () => {
  isFocused.value = false
  opacity.value = 0
}

const handleMouseEnter = () => {
  opacity.value = 0.6
}

const handleMouseLeave = () => {
  opacity.value = 0
}
</script>

<style scoped>
.spotlight-card {
  position: relative;
  overflow: hidden;
}

.spotlight-overlay {
  pointer-events: none;
  position: absolute;
  inset: 0;
  opacity: 0;
  transition: opacity 500ms ease-in-out;
  border-radius: inherit;
}
</style>
