<template>
  <div class="stepper-wrapper" v-bind="$attrs">
    <div class="stepper-card" :class="stepCircleContainerClassName">
      <!-- Step indicators -->
      <div
        class="stepper-indicators"
        :class="stepContainerClassName"
        :style="{ marginBottom: isCompleted ? '0' : '2rem' }"
      >
        <template v-for="(_, index) in stepsArray" :key="index + 1">
          <div
            @click="() => handleStepClick(index + 1)"
            class="step-circle"
            :class="{ 'step-locked': isCompleted && lockOnComplete }"
            :style="getStepIndicatorStyle(index + 1)"
          >
            <svg
              v-if="getStepStatus(index + 1) === 'complete'"
              class="step-check"
              fill="none"
              stroke="currentColor"
              :stroke-width="2"
              viewBox="0 0 24 24"
            >
              <Motion
                as="path"
                d="M5 13l4 4L19 7"
                stroke-linecap="round"
                stroke-linejoin="round"
                :initial="{ pathLength: 0, opacity: 0 }"
                :animate="
                  getStepStatus(index + 1) === 'complete'
                    ? { pathLength: 1, opacity: 1 }
                    : { pathLength: 0, opacity: 0 }
                "
              />
            </svg>
            <div v-else-if="getStepStatus(index + 1) === 'active'" class="step-active-dot" />
            <span v-else class="step-number">{{ index + 1 }}</span>
          </div>

          <div v-if="index < totalSteps - 1" class="step-connector">
            <Motion
              as="div"
              class="step-connector-fill"
              :initial="{ width: 0, backgroundColor: 'var(--border-color)' }"
              :animate="
                currentStep > index + 1
                  ? { width: '100%', backgroundColor: 'var(--accent)' }
                  : { width: 0, backgroundColor: 'var(--border-color)' }
              "
              :transition="{ type: 'spring', stiffness: 100, damping: 15, duration: 0.4 }"
            />
          </div>
        </template>
      </div>

      <!-- Content area -->
      <Motion
        as="div"
        class="stepper-content"
        :class="contentClassName"
        :style="{
          position: 'relative',
          overflow: 'hidden',
          marginBottom: isCompleted ? '0' : '2rem',
        }"
        :animate="{ height: isCompleted ? 0 : `${parentHeight + 1}px` }"
        :transition="{ type: 'spring', stiffness: 200, damping: 25, duration: 0.4 }"
      >
        <AnimatePresence :initial="false" mode="sync" :custom="direction">
          <Motion
            v-if="!isCompleted"
            ref="containerRef"
            as="div"
            :key="currentStep"
            :initial="getStepContentInitial()"
            :animate="{ x: '0%', opacity: 1 }"
            :exit="getStepContentExit()"
            :transition="{ type: 'tween', stiffness: 300, damping: 30, duration: 0.4 }"
            :style="{ position: 'absolute', left: 0, right: 0, top: 0 }"
          >
            <div ref="contentRef" v-if="slots.default && slots.default()[currentStep - 1]">
              <component :is="slots.default()[currentStep - 1]!" />
            </div>
          </Motion>
        </AnimatePresence>
      </Motion>

      <!-- Footer buttons -->
      <div v-if="!isCompleted" class="stepper-footer" :class="footerClassName">
        <div class="stepper-buttons" :class="{ 'justify-end': currentStep === 1 }">
          <button v-if="currentStep !== 1" @click="handleBack" class="stepper-back-btn">
            {{ backButtonText }}
          </button>
          <button
            @click="isLastStep ? handleComplete() : handleNext()"
            class="stepper-next-btn"
          >
            {{ isLastStep ? completeButtonText : nextButtonText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  useSlots,
  watch,
  onMounted,
  nextTick,
  useTemplateRef,
} from 'vue'
import { Motion, AnimatePresence } from 'motion-v'

interface StepperProps {
  initialStep?: number
  onStepChange?: (step: number) => void
  onFinalStepCompleted?: () => void
  stepCircleContainerClassName?: string
  stepContainerClassName?: string
  contentClassName?: string
  footerClassName?: string
  backButtonText?: string
  nextButtonText?: string
  completeButtonText?: string
  disableStepIndicators?: boolean
  lockOnComplete?: boolean
}

const props = withDefaults(defineProps<StepperProps>(), {
  initialStep: 1,
  onStepChange: () => {},
  onFinalStepCompleted: () => {},
  stepCircleContainerClassName: '',
  stepContainerClassName: '',
  contentClassName: '',
  footerClassName: '',
  backButtonText: 'Tillbaka',
  nextButtonText: 'Fortsätt',
  completeButtonText: 'Klar',
  disableStepIndicators: false,
  lockOnComplete: true,
})

const slots = useSlots()
const currentStep = ref(props.initialStep)
const direction = ref(1)
const isCompleted = ref(false)
const parentHeight = ref(0)
const containerRef = useTemplateRef<HTMLDivElement>('containerRef')
const contentRef = useTemplateRef<HTMLDivElement>('contentRef')

const stepsArray = computed(() => slots.default?.() || [])
const totalSteps = computed(() => stepsArray.value.length)
const isLastStep = computed(() => currentStep.value === totalSteps.value)

const getStepStatus = (step: number) => {
  if (isCompleted.value || currentStep.value > step) return 'complete'
  if (currentStep.value === step) return 'active'
  return 'inactive'
}

const getStepIndicatorStyle = (step: number) => {
  const status = getStepStatus(step)
  switch (status) {
    case 'active':
    case 'complete':
      return { backgroundColor: 'var(--accent)', color: '#fff' }
    default:
      return { backgroundColor: 'var(--bg-secondary)', color: 'var(--text-secondary)' }
  }
}

const getStepContentInitial = () => ({
  x: direction.value >= 0 ? '-100%' : '100%',
  opacity: 0,
})

const getStepContentExit = () => ({
  x: direction.value >= 0 ? '50%' : '-50%',
  opacity: 0,
})

const handleStepClick = (step: number) => {
  if (isCompleted.value && props.lockOnComplete) return
  if (!props.disableStepIndicators) {
    direction.value = step > currentStep.value ? 1 : -1
    updateStep(step)
  }
}

const measureHeight = () => {
  nextTick(() => {
    if (contentRef.value) {
      const height = contentRef.value.offsetHeight
      if (height > 0 && height !== parentHeight.value) {
        parentHeight.value = height
      }
    }
  })
}

const updateStep = (newStep: number) => {
  if (newStep >= 1 && newStep <= totalSteps.value) {
    currentStep.value = newStep
  }
}

const handleBack = () => {
  direction.value = -1
  updateStep(currentStep.value - 1)
}

const handleNext = () => {
  direction.value = 1
  updateStep(currentStep.value + 1)
}

const handleComplete = () => {
  isCompleted.value = true
  props.onFinalStepCompleted?.()
}

watch(currentStep, (newStep, oldStep) => {
  props.onStepChange?.(newStep)
  if (newStep !== oldStep && !isCompleted.value) {
    nextTick(measureHeight)
  } else if (!props.lockOnComplete && isCompleted.value) {
    isCompleted.value = false
    nextTick(measureHeight)
  }
})

onMounted(() => {
  if (props.initialStep !== 1) {
    currentStep.value = props.initialStep
  }
  measureHeight()
})
</script>

<style scoped>
.stepper-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
}

.stepper-card {
  width: 100%;
  max-width: 28rem;
  padding: 2rem;
  border-radius: 2rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-md);
}

.stepper-indicators {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.step-circle {
  position: relative;
  outline: none;
  display: flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.step-locked {
  cursor: default;
}

.step-check {
  height: 1rem;
  width: 1rem;
  color: white;
  stroke: white;
}

.step-active-dot {
  height: 0.75rem;
  width: 0.75rem;
  border-radius: 50%;
  background: white;
}

.step-number {
  font-size: 0.875rem;
  font-family: 'Nunito', sans-serif;
}

.step-connector {
  position: relative;
  margin-left: 0.5rem;
  margin-right: 0.5rem;
  height: 2px;
  flex: 1;
  overflow: hidden;
  border-radius: 4px;
  background: var(--border-color);
}

.step-connector-fill {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
}

.stepper-footer {
  width: 100%;
}

.stepper-buttons {
  display: flex;
  width: 100%;
  justify-content: space-between;
}

.stepper-buttons.justify-end {
  justify-content: flex-end;
}

.stepper-back-btn {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  color: var(--text-secondary);
  background: transparent;
  cursor: pointer;
  transition: all 0.35s ease;
  border-radius: 8px;
  padding: 0.25rem 0.5rem;
  border: none;
}

.stepper-back-btn:hover {
  color: var(--text-primary);
}

.stepper-next-btn {
  border: none;
  background: var(--accent);
  transition: all 0.35s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 100px;
  color: white;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  padding: 0.5rem 1.25rem;
  cursor: pointer;
}

.stepper-next-btn:hover {
  background: var(--accent-dark, #e85a4a);
}

.stepper-next-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
