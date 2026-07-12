import { describe, it, expect, vi, afterEach } from 'vitest'
import { createApp, defineComponent, ref } from 'vue'
import { useClickOutside } from '../useClickOutside'

function setup(callback: () => void) {
  const elementRef = ref<HTMLElement | null>(null)
  const el = document.createElement('div')
  document.body.appendChild(el)
  elementRef.value = el

  const app = createApp(
    defineComponent({
      setup() {
        useClickOutside(elementRef, callback)
        return () => null
      },
    }),
  )
  app.mount(document.createElement('div'))

  return { app, el }
}

describe('useClickOutside', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('fires callback on mousedown outside the element', () => {
    const callback = vi.fn()
    const { el } = setup(callback)

    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))

    expect(callback).toHaveBeenCalledTimes(1)
    expect(el).toBeTruthy()
  })

  it('does not fire callback on mousedown inside the element', () => {
    const callback = vi.fn()
    const { el } = setup(callback)

    el.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))

    expect(callback).not.toHaveBeenCalled()
  })

  it('removes the listener on unmount', () => {
    const callback = vi.fn()
    const { app } = setup(callback)

    app.unmount()
    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))

    expect(callback).not.toHaveBeenCalled()
  })
})
