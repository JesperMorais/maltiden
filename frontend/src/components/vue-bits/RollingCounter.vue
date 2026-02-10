<template>
  <div class="counter-container" :style="containerStyle">
    <div class="counter-digits" :style="counterStyles">
      <div
        v-for="place in places"
        :key="place"
        class="counter-digit-col"
        :style="digitStyles"
      >
        <Motion
          v-for="digit in 10"
          :key="digit - 1"
          tag="span"
          class="counter-digit"
          :animate="{ y: getDigitPosition(place, digit - 1) }"
        >
          {{ digit - 1 }}
        </Motion>
      </div>
    </div>
    <div class="counter-gradient-overlay">
      <div class="counter-gradient-top" :style="topGradientStyles" />
      <div class="counter-gradient-bottom" :style="bottomGradientStyles" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Motion } from 'motion-v'
import type { CSSProperties } from 'vue'

interface CounterProps {
  value: number
  fontSize?: number
  padding?: number
  places?: number[]
  gap?: number
  borderRadius?: number
  horizontalPadding?: number
  textColor?: string
  fontWeight?: string | number
  containerStyle?: CSSProperties
  gradientHeight?: number
  gradientFrom?: string
  gradientTo?: string
}

const props = withDefaults(defineProps<CounterProps>(), {
  fontSize: 18,
  padding: 0,
  places: () => [10, 1],
  gap: 2,
  borderRadius: 4,
  horizontalPadding: 2,
  textColor: 'var(--accent)',
  fontWeight: 'bold',
  containerStyle: () => ({}),
  gradientHeight: 0,
  gradientFrom: 'transparent',
  gradientTo: 'transparent',
})

const digitHeight = computed(() => props.fontSize + props.padding)

const counterStyles = computed<CSSProperties>(() => ({
  fontSize: `${props.fontSize}px`,
  gap: `${props.gap}px`,
  borderRadius: `${props.borderRadius}px`,
  paddingLeft: `${props.horizontalPadding}px`,
  paddingRight: `${props.horizontalPadding}px`,
  color: props.textColor,
  fontWeight: props.fontWeight,
}))

const digitStyles = computed<CSSProperties>(() => ({
  height: `${digitHeight.value}px`,
}))

const topGradientStyles = computed<CSSProperties>(() => ({
  height: `${props.gradientHeight}px`,
  background: `linear-gradient(to bottom, ${props.gradientFrom}, ${props.gradientTo})`,
}))

const bottomGradientStyles = computed<CSSProperties>(() => ({
  height: `${props.gradientHeight}px`,
  background: `linear-gradient(to top, ${props.gradientFrom}, ${props.gradientTo})`,
}))

const springValues = ref<Record<number, number>>({})

const initializeSpringValues = () => {
  props.places.forEach((place) => {
    springValues.value[place] = Math.floor(props.value / place)
  })
}

initializeSpringValues()

watch(
  () => props.value,
  (newValue, oldValue) => {
    if (newValue === oldValue) return
    props.places.forEach((place) => {
      const newRoundedValue = Math.floor(newValue / place)
      const oldRoundedValue = springValues.value[place]
      if (newRoundedValue !== oldRoundedValue) {
        springValues.value[place] = newRoundedValue
      }
    })
  },
  { immediate: true },
)

watch(
  () => digitHeight.value,
  () => {
    positionCache.clear()
  },
)

const positionCache = new Map<string, number>()

const getDigitPosition = (place: number, digit: number): number => {
  const springValue = springValues.value[place] || 0
  const cacheKey = `${place}-${digit}-${springValue}`

  if (positionCache.has(cacheKey)) {
    return positionCache.get(cacheKey)!
  }

  const placeValue = springValue % 10
  const offset = (10 + digit - placeValue) % 10
  let position = offset * digitHeight.value

  if (offset > 5) {
    position -= 10 * digitHeight.value
  }

  if (positionCache.size > 200) {
    const keysToDelete = [...positionCache.keys()].slice(0, positionCache.size - 100)
    for (const key of keysToDelete) {
      positionCache.delete(key)
    }
  }

  positionCache.set(cacheKey, position)
  return position
}
</script>

<style scoped>
.counter-container {
  position: relative;
  display: inline-block;
}

.counter-digits {
  display: flex;
  overflow: hidden;
}

.counter-digit-col {
  position: relative;
  width: 1ch;
  font-variant-numeric: tabular-nums;
}

.counter-digit {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.counter-gradient-overlay {
  pointer-events: none;
  position: absolute;
  inset: 0;
}

.counter-gradient-top {
  position: absolute;
  top: 0;
  width: 100%;
}

.counter-gradient-bottom {
  position: absolute;
  bottom: 0;
  width: 100%;
}
</style>
