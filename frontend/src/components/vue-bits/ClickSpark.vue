<template>
  <div ref="containerRef" class="click-spark" @click="handleClick">
    <canvas ref="canvasRef" class="click-spark-canvas" />
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, useTemplateRef } from 'vue'

interface Spark {
  x: number
  y: number
  angle: number
  startTime: number
}

interface Props {
  sparkColor?: string
  sparkSize?: number
  sparkRadius?: number
  sparkCount?: number
  duration?: number
  easing?: 'linear' | 'ease-in' | 'ease-out' | 'ease-in-out'
  extraScale?: number
}

const props = withDefaults(defineProps<Props>(), {
  sparkColor: '#ff6b5b',
  sparkSize: 10,
  sparkRadius: 15,
  sparkCount: 8,
  duration: 400,
  easing: 'ease-out',
  extraScale: 1.0,
})

const containerRef = useTemplateRef<HTMLDivElement>('containerRef')
const canvasRef = useTemplateRef<HTMLCanvasElement>('canvasRef')
const sparks = ref<Spark[]>([])
let animationId: number | null = null

const easeFunc = computed(() => {
  return (t: number) => {
    switch (props.easing) {
      case 'linear':
        return t
      case 'ease-in':
        return t * t
      case 'ease-in-out':
        return t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t
      default:
        return t * (2 - t)
    }
  }
})

const handleClick = (e: MouseEvent) => {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top

  const now = performance.now()
  const newSparks: Spark[] = Array.from({ length: props.sparkCount }, (_, i) => ({
    x,
    y,
    angle: (2 * Math.PI * i) / props.sparkCount,
    startTime: now,
  }))

  sparks.value.push(...newSparks)

  // Start animation loop if not already running
  if (animationId === null) {
    animationId = requestAnimationFrame(draw)
  }
}

const draw = (timestamp: number) => {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!ctx || !canvas) {
    animationId = null
    return
  }

  ctx.clearRect(0, 0, canvas.width, canvas.height)

  sparks.value = sparks.value.filter((spark: Spark) => {
    const elapsed = timestamp - spark.startTime
    if (elapsed >= props.duration) {
      return false
    }

    const progress = elapsed / props.duration
    const eased = easeFunc.value(progress)

    const distance = eased * props.sparkRadius * props.extraScale
    const lineLength = props.sparkSize * (1 - eased)

    const x1 = spark.x + distance * Math.cos(spark.angle)
    const y1 = spark.y + distance * Math.sin(spark.angle)
    const x2 = spark.x + (distance + lineLength) * Math.cos(spark.angle)
    const y2 = spark.y + (distance + lineLength) * Math.sin(spark.angle)

    ctx.strokeStyle = props.sparkColor
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(x1, y1)
    ctx.lineTo(x2, y2)
    ctx.stroke()

    return true
  })

  // Only continue loop if there are active sparks
  if (sparks.value.length > 0) {
    animationId = requestAnimationFrame(draw)
  } else {
    animationId = null
  }
}

const resizeCanvas = () => {
  const canvas = canvasRef.value
  if (!canvas) return

  const parent = canvas.parentElement
  if (!parent) return

  const { width, height } = parent.getBoundingClientRect()
  if (canvas.width !== width || canvas.height !== height) {
    canvas.width = width
    canvas.height = height
  }
}

let resizeTimeout: ReturnType<typeof setTimeout>

const handleResize = () => {
  clearTimeout(resizeTimeout)
  resizeTimeout = setTimeout(resizeCanvas, 100)
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  const canvas = canvasRef.value
  if (!canvas) return

  const parent = canvas.parentElement
  if (!parent) return

  resizeObserver = new ResizeObserver(handleResize)
  resizeObserver.observe(parent)

  resizeCanvas()
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  clearTimeout(resizeTimeout)

  if (animationId !== null) {
    cancelAnimationFrame(animationId)
    animationId = null
  }
})
</script>

<style scoped>
.click-spark {
  position: relative;
  width: 100%;
  height: 100%;
}

.click-spark-canvas {
  position: absolute;
  inset: 0;
  pointer-events: none;
  width: 100%;
  height: 100%;
}
</style>
