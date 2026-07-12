import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createApp, defineComponent, ref, nextTick } from 'vue'
import { useFocusTrap } from '../useFocusTrap'

function setup(onEscape?: () => void) {
  const containerRef = ref<HTMLElement | null>(null)
  const isActive = ref(false)

  const container = document.createElement('div')
  const first = document.createElement('button')
  first.textContent = 'first'
  const second = document.createElement('button')
  second.textContent = 'second'
  container.appendChild(first)
  container.appendChild(second)
  document.body.appendChild(container)
  containerRef.value = container

  const app = createApp(
    defineComponent({
      setup() {
        useFocusTrap(containerRef, { isActive, onEscape })
        return () => null
      },
    }),
  )
  app.mount(document.createElement('div'))

  return { app, container, first, second, isActive }
}

describe('useFocusTrap', () => {
  let rafSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    rafSpy = vi.spyOn(window, 'requestAnimationFrame').mockImplementation((cb: FrameRequestCallback) => {
      cb(0)
      return 0
    })
  })

  afterEach(() => {
    rafSpy.mockRestore()
    document.body.innerHTML = ''
  })

  it('wraps focus from last to first element on Tab', async () => {
    const { second, first, isActive } = setup()
    isActive.value = true
    await nextTick()

    second.focus()
    expect(document.activeElement).toBe(second)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))

    expect(document.activeElement).toBe(first)
  })

  it('wraps focus from first to last element on Shift+Tab', async () => {
    const { second, first, isActive } = setup()
    isActive.value = true
    await nextTick()

    first.focus()
    expect(document.activeElement).toBe(first)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }))

    expect(document.activeElement).toBe(second)
  })

  it('restores previously focused element on deactivate', async () => {
    const outside = document.createElement('button')
    outside.textContent = 'outside'
    document.body.appendChild(outside)
    outside.focus()

    const { isActive } = setup()
    isActive.value = true
    await nextTick()

    isActive.value = false
    await nextTick()

    expect(document.activeElement).toBe(outside)
  })

  it('calls onEscape when Escape is pressed', async () => {
    const onEscape = vi.fn()
    const { isActive } = setup(onEscape)
    isActive.value = true
    await nextTick()

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))

    expect(onEscape).toHaveBeenCalledTimes(1)
  })
})
