<template>
  <div
    class="glare-hover"
    :class="className"
    :style="containerStyle"
    @mouseenter="animateIn"
    @mouseleave="animateOut"
  >
    <div ref="overlayRef" :style="overlayStyle" />
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import type { CSSProperties } from 'vue'

interface GlareHoverProps {
  glareColor?: string
  glareOpacity?: number
  glareAngle?: number
  glareSize?: number
  transitionDuration?: number
  playOnce?: boolean
  className?: string
}

const props = withDefaults(defineProps<GlareHoverProps>(), {
  glareColor: '#ffffff',
  glareOpacity: 0.4,
  glareAngle: -45,
  glareSize: 250,
  transitionDuration: 650,
  playOnce: false,
  className: '',
})

const overlayRef = useTemplateRef<HTMLDivElement>('overlayRef')

const rgba = computed(() => {
  const hex = props.glareColor.replace('#', '')
  let result = props.glareColor

  if (/^[\dA-Fa-f]{6}$/.test(hex)) {
    const r = parseInt(hex.slice(0, 2), 16)
    const g = parseInt(hex.slice(2, 4), 16)
    const b = parseInt(hex.slice(4, 6), 16)
    result = `rgba(${r}, ${g}, ${b}, ${props.glareOpacity})`
  } else if (/^[\dA-Fa-f]{3}$/.test(hex)) {
    const r = parseInt(hex[0]! + hex[0]!, 16)
    const g = parseInt(hex[1]! + hex[1]!, 16)
    const b = parseInt(hex[2]! + hex[2]!, 16)
    result = `rgba(${r}, ${g}, ${b}, ${props.glareOpacity})`
  }

  return result
})

const containerStyle = computed<CSSProperties>(() => ({
  position: 'relative',
  overflow: 'hidden',
}))

const overlayStyle = computed<CSSProperties>(() => ({
  position: 'absolute',
  inset: '0',
  background: `linear-gradient(${props.glareAngle}deg,
      hsla(0,0%,0%,0) 60%,
      ${rgba.value} 70%,
      hsla(0,0%,0%,0) 100%)`,
  backgroundSize: `${props.glareSize}% ${props.glareSize}%, 100% 100%`,
  backgroundRepeat: 'no-repeat',
  backgroundPosition: '-100% -100%, 0 0',
  pointerEvents: 'none',
  borderRadius: 'inherit',
}))

const animateIn = () => {
  const el = overlayRef.value
  if (!el) return
  el.style.transition = 'none'
  el.style.backgroundPosition = '-100% -100%, 0 0'
  void el.offsetHeight
  el.style.transition = `${props.transitionDuration}ms ease`
  el.style.backgroundPosition = '100% 100%, 0 0'
}

const animateOut = () => {
  const el = overlayRef.value
  if (!el) return
  if (props.playOnce) {
    el.style.transition = 'none'
    el.style.backgroundPosition = '-100% -100%, 0 0'
  } else {
    el.style.transition = `${props.transitionDuration}ms ease`
    el.style.backgroundPosition = '-100% -100%, 0 0'
  }
}
</script>

<style scoped>
.glare-hover {
  position: relative;
  overflow: hidden;
}
</style>
