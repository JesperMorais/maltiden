<script setup lang="ts">
import { AnimatePresence, Motion } from 'motion-v'
import type { MotionProps } from 'motion-v'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

type TransitionType = NonNullable<MotionProps['transition']>
type InitialType = NonNullable<MotionProps['initial']>
type AnimateType = NonNullable<MotionProps['animate']>
type ExitType = NonNullable<MotionProps['exit']>

interface WordElement {
  characters: string[]
  needsSpace: boolean
}

interface RotatingTextProps {
  texts: string[]
  transition?: TransitionType
  initial?: InitialType
  animate?: AnimateType
  exit?: ExitType
  animatePresenceMode?: 'sync' | 'wait'
  animatePresenceInitial?: boolean
  rotationInterval?: number
  staggerDuration?: number
  staggerFrom?: 'first' | 'last' | 'center' | 'random' | number
  loop?: boolean
  auto?: boolean
  splitBy?: 'characters' | 'words' | 'lines'
  mainClassName?: string
  splitLevelClassName?: string
  elementLevelClassName?: string
}

const props = withDefaults(defineProps<RotatingTextProps>(), {
  transition: () =>
    ({
      type: 'spring',
      damping: 25,
      stiffness: 300,
    }) as TransitionType,
  initial: () => ({ y: '100%', opacity: 0 }) as InitialType,
  animate: () => ({ y: 0, opacity: 1 }) as AnimateType,
  exit: () => ({ y: '-120%', opacity: 0 }) as ExitType,
  animatePresenceMode: 'wait',
  animatePresenceInitial: false,
  rotationInterval: 2000,
  staggerDuration: 0,
  staggerFrom: 'first',
  loop: true,
  auto: true,
  splitBy: 'characters',
})

const currentTextIndex = ref(0)
let intervalId: ReturnType<typeof setInterval> | null = null
const prefersReducedMotion = ref(
  typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches,
)

const splitIntoCharacters = (text: string): string[] => {
  if (typeof Intl !== 'undefined' && 'Segmenter' in Intl) {
    const IntlWithSegmenter = Intl as typeof Intl & {
      Segmenter: new (
        locales?: string | string[],
        options?: { granularity: 'grapheme' | 'word' | 'sentence' },
      ) => {
        segment: (text: string) => Iterable<{ segment: string }>
      }
    }
    const segmenter = new IntlWithSegmenter.Segmenter('en', { granularity: 'grapheme' })
    return [...segmenter.segment(text)].map(({ segment }) => segment)
  }
  return [...text]
}

const elements = computed((): WordElement[] => {
  const currentText = props.texts[currentTextIndex.value]!
  switch (props.splitBy) {
    case 'characters': {
      const words = currentText.split(' ')
      return words.map((word, i) => ({
        characters: splitIntoCharacters(word),
        needsSpace: i !== words.length - 1,
      }))
    }
    case 'words': {
      const words = currentText.split(' ')
      return words.map((word, i) => ({
        characters: [word],
        needsSpace: i !== words.length - 1,
      }))
    }
    case 'lines': {
      const lines = currentText.split('\n')
      return lines.map((line, i) => ({
        characters: [line],
        needsSpace: i !== lines.length - 1,
      }))
    }
    default: {
      const parts = currentText.split(props.splitBy!)
      return parts.map((part, i) => ({
        characters: [part],
        needsSpace: i !== parts.length - 1,
      }))
    }
  }
})

const getStaggerDelay = (index: number, totalChars: number): number => {
  const { staggerDuration, staggerFrom } = props
  switch (staggerFrom) {
    case 'first':
      return index * staggerDuration
    case 'last':
      return (totalChars - 1 - index) * staggerDuration
    case 'center': {
      const center = Math.floor(totalChars / 2)
      return Math.abs(center - index) * staggerDuration
    }
    case 'random': {
      const randomIndex = Math.floor(Math.random() * totalChars)
      return Math.abs(randomIndex - index) * staggerDuration
    }
    default:
      return Math.abs((staggerFrom as number) - index) * staggerDuration
  }
}

const handleIndexChange = (newIndex: number): void => {
  currentTextIndex.value = newIndex
}

const next = (): void => {
  const isAtEnd = currentTextIndex.value === props.texts.length - 1
  const nextIndex = isAtEnd ? (props.loop ? 0 : currentTextIndex.value) : currentTextIndex.value + 1
  if (nextIndex !== currentTextIndex.value) {
    handleIndexChange(nextIndex)
  }
}

const cleanupInterval = (): void => {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
  }
}

const startInterval = (): void => {
  if (props.auto && !prefersReducedMotion.value) {
    intervalId = setInterval(next, props.rotationInterval)
  }
}

watch(
  () => [props.auto, props.rotationInterval] as const,
  () => {
    cleanupInterval()
    startInterval()
  },
)

onMounted(() => {
  startInterval()
})

onUnmounted(() => {
  cleanupInterval()
})
</script>

<template>
  <!-- Static fallback for prefers-reduced-motion -->
  <span
    v-if="prefersReducedMotion"
    class="rotating-text rotating-text-static"
    :class="mainClassName"
    v-bind="$attrs"
  >
    {{ texts[0] }}
  </span>

  <Motion
    v-else
    tag="span"
    class="rotating-text"
    :class="mainClassName"
    v-bind="$attrs"
    :transition="transition"
  >
    <span class="sr-only">
      {{ texts[currentTextIndex] }}
    </span>

    <AnimatePresence :mode="animatePresenceMode" :initial="animatePresenceInitial">
      <Motion
        :key="currentTextIndex"
        tag="span"
        :class="
          splitBy === 'lines'
            ? 'rotating-text-lines'
            : 'rotating-text-inner'
        "
        aria-hidden="true"
      >
        <span
          v-for="(wordObj, wordIndex) in elements"
          :key="wordIndex"
          class="rotating-text-word"
          :class="splitLevelClassName"
        >
          <Motion
            v-for="(char, charIndex) in wordObj.characters"
            :key="charIndex"
            tag="span"
            :initial="initial"
            :animate="animate"
            :exit="exit"
            :transition="{
              ...transition,
              delay: getStaggerDelay(
                elements
                  .slice(0, wordIndex)
                  .reduce((sum, word) => sum + word.characters.length, 0) + charIndex,
                elements.reduce((sum, word) => sum + word.characters.length, 0),
              ),
            }"
            class="rotating-text-char"
            :class="elementLevelClassName"
          >
            {{ char }}
          </Motion>
          <span v-if="wordObj.needsSpace" class="rotating-text-space">&nbsp;</span>
        </span>
      </Motion>
    </AnimatePresence>
  </Motion>
</template>

<style scoped>
.rotating-text {
  display: inline-flex;
  flex-wrap: wrap;
  white-space: pre-wrap;
  position: relative;
  overflow: hidden;
}

.rotating-text-inner {
  display: inline-flex;
  flex-wrap: wrap;
  white-space: pre-wrap;
  position: relative;
}

.rotating-text-lines {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.rotating-text-word {
  display: inline-flex;
}

.rotating-text-char {
  display: inline-block;
}

.rotating-text-space {
  white-space: pre;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

.rotating-text-static {
  display: inline-flex;
  white-space: pre-wrap;
}

@media (prefers-reduced-motion: reduce) {
  .rotating-text,
  .rotating-text-inner,
  .rotating-text-char {
    transition: none !important;
    animation: none !important;
  }
}
</style>
