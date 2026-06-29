import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMenuGeneratorStore } from '../menuGenerator'
import type { GenerateMenuResponse, GenerateMenuRequest } from '@/api/menu.api'

vi.mock('@/api/client', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn() },
}))

vi.mock('@/api/menu.api', () => ({
  generateMenu: vi.fn(),
  saveMenu: vi.fn(),
}))

function buildWeek(start: Date, extra: Partial<GenerateMenuResponse> = {}): GenerateMenuResponse {
  const dates: string[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    dates.push(
      `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`,
    )
  }
  return {
    id: 'menu_test',
    days: dates.map((date, i) => ({
      date,
      recipeId: `rec_${i}`,
      recipeName: `Recept ${i}`,
      emoji: '🍽️',
      servings: 4,
    })),
    ...extra,
  }
}

describe('menuGenerator store — experience layer (Phase 4)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('sends wishes and arrange in the generateInitialMenu request body', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildWeek(start))

    const store = useMenuGeneratorStore()
    store.wishes = 'två vegetariska dagar'
    store.setArrange(true)
    store.initializeWeek(start)
    await store.generateInitialMenu()

    const req = (generateMenu as ReturnType<typeof vi.fn>).mock.calls[0]![0] as GenerateMenuRequest
    expect(req.wishes).toBe('två vegetariska dagar')
    expect(req.arrange).toBe(true)
  })

  it('reads rationale and wishesIgnored from the generate response', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(
      buildWeek(start, { rationale: 'En balanserad vecka.', wishesIgnored: true }),
    )

    const store = useMenuGeneratorStore()
    store.initializeWeek(start)
    await store.generateInitialMenu()

    expect(store.rationale).toBe('En balanserad vecka.')
    expect(store.wishesIgnored).toBe(true)
  })

  it('defaults rationale/wishesIgnored to empty/false when absent in the response', async () => {
    const { generateMenu } = await import('@/api/menu.api')
    const start = new Date(2026, 5, 17, 9, 0, 0)
    ;(generateMenu as ReturnType<typeof vi.fn>).mockResolvedValue(buildWeek(start))

    const store = useMenuGeneratorStore()
    store.initializeWeek(start)
    await store.generateInitialMenu()

    expect(store.rationale).toBe('')
    expect(store.wishesIgnored).toBe(false)
  })

  it('persists arrange to localStorage but does not persist wishes', () => {
    const store = useMenuGeneratorStore()
    store.setArrange(true)
    store.wishes = 'snabb vecka'
    expect(localStorage.getItem('maltiden_arrange')).toBe('true')
    expect(localStorage.getItem('maltiden_wishes')).toBeNull()
  })

  it('clears wishes on clearDraft but keeps the persisted arrange preference', () => {
    const store = useMenuGeneratorStore()
    store.setArrange(true)
    store.wishes = 'snabb vecka'

    store.clearDraft()
    expect(store.wishes).toBe('')
    expect(store.arrange).toBe(true)
  })
})
