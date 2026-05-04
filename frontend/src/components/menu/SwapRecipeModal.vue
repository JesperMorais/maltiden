<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Search, X, Shuffle, UtensilsCrossed, Check } from 'lucide-vue-next'
import { getRecipes } from '@/api/recipes.api'
import type { RecipeSummary } from '@/api/recipes.api'

interface Props {
  currentRecipeId: string
  dayDate: string
  tags?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  tags: () => [],
})

const emit = defineEmits<{
  close: []
  select: [recipeId: string]
}>()

const recipes = ref<RecipeSummary[]>([])
const isLoading = ref(false)
const loadError = ref<string | null>(null)
const query = ref('')

const filteredRecipes = computed(() => {
  const q = query.value.trim().toLowerCase()
  let pool = recipes.value

  // Tag filter (intersection if any tags provided)
  if (props.tags && props.tags.length > 0) {
    const tagSet = new Set(props.tags.map((t) => t.toLowerCase()))
    const tagFiltered = pool.filter((r) =>
      r.tags.some((t) => tagSet.has(t.toLowerCase())),
    )
    // Fall back to full pool if filter would empty the list
    if (tagFiltered.length > 0) pool = tagFiltered
  }

  if (!q) return pool

  return pool.filter((r) => {
    if (r.name.toLowerCase().includes(q)) return true
    if (r.tags.some((t) => t.toLowerCase().includes(q))) return true
    return false
  })
})

const swappablePool = computed(() =>
  filteredRecipes.value.filter((r) => r.id !== props.currentRecipeId),
)

function handleSelect(id: string) {
  if (id === props.currentRecipeId) {
    emit('close')
    return
  }
  emit('select', id)
}

function handleRandom() {
  const pool = swappablePool.value
  if (pool.length === 0) return
  const idx = Math.floor(Math.random() * pool.length)
  const pick = pool[idx]
  if (pick) emit('select', pick.id)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

function onBackdropClick() {
  emit('close')
}

onMounted(async () => {
  document.addEventListener('keydown', onKeydown)
  isLoading.value = true
  try {
    const { recipes: list } = await getRecipes()
    recipes.value = list
  } catch (e) {
    loadError.value = 'Kunde inte ladda recept. Försök igen.'
    console.error('Failed to load recipes for swap:', e)
  } finally {
    isLoading.value = false
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div
    class="swap-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="swap-modal-title"
    @click.self="onBackdropClick"
  >
    <div class="swap-modal" @click.stop>
      <header class="swap-header">
        <div class="swap-title-wrap">
          <h2 id="swap-modal-title" class="swap-title">Byt recept</h2>
          <p class="swap-subtitle">Välj ett annat recept för dagen</p>
        </div>
        <button
          type="button"
          class="close-btn"
          aria-label="Stäng"
          @click="emit('close')"
        >
          <X :size="22" />
        </button>
      </header>

      <div class="search-wrap">
        <Search :size="16" class="search-icon" aria-hidden="true" />
        <input
          v-model="query"
          type="search"
          class="search-input"
          placeholder="Sök recept eller tagg…"
          aria-label="Sök recept"
        />
      </div>

      <div class="recipe-list" role="list">
        <p v-if="isLoading" class="status-msg">Laddar recept…</p>
        <p v-else-if="loadError" class="status-msg error">{{ loadError }}</p>
        <p v-else-if="filteredRecipes.length === 0" class="status-msg">
          Inga recept matchar din sökning.
        </p>

        <button
          v-for="recipe in filteredRecipes"
          v-else
          :key="recipe.id"
          type="button"
          role="listitem"
          class="recipe-row"
          :class="{ current: recipe.id === currentRecipeId }"
          :aria-current="recipe.id === currentRecipeId ? 'true' : undefined"
          @click="handleSelect(recipe.id)"
        >
          <span v-if="recipe.emoji" class="row-emoji" aria-hidden="true">{{ recipe.emoji }}</span>
          <UtensilsCrossed v-else :size="22" class="row-icon" aria-hidden="true" />
          <span class="row-body">
            <span class="row-name">{{ recipe.name }}</span>
            <span v-if="recipe.tags.length > 0" class="row-tags">
              {{ recipe.tags.slice(0, 3).join(' · ') }}
            </span>
          </span>
          <Check
            v-if="recipe.id === currentRecipeId"
            :size="18"
            class="row-current"
            aria-label="Nuvarande recept"
          />
        </button>
      </div>

      <footer class="swap-footer">
        <button
          type="button"
          class="random-btn"
          :disabled="swappablePool.length === 0 || isLoading"
          @click="handleRandom"
        >
          <Shuffle :size="16" />
          Slumpa nytt
        </button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.swap-backdrop {
  position: fixed;
  inset: 0;
  background: rgb(0 0 0 / 0.45);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  z-index: 100;
  animation: fadeIn var(--duration-fast, 150ms) var(--ease-default, ease-out);
}

.swap-modal {
  background: var(--bg-primary);
  border-top-left-radius: var(--radius-xl, 20px);
  border-top-right-radius: var(--radius-xl, 20px);
  width: 100%;
  max-width: 480px;
  max-height: 92vh;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
  animation: slideUp var(--duration-normal, 220ms) var(--ease-default, ease-out);
}

.swap-header {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem 1rem 0.5rem;
}

.swap-title-wrap {
  flex: 1;
  min-width: 0;
}

.swap-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.35rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

.swap-subtitle {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0.15rem 0 0;
}

.close-btn {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 50%;
  color: var(--text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.close-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.search-wrap {
  position: relative;
  padding: 0.5rem 1rem 0.75rem;
}

.search-icon {
  position: absolute;
  top: 50%;
  left: 1.75rem;
  transform: translateY(-50%);
  color: var(--text-muted);
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 44px;
  padding: 0 0.85rem 0 2.4rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-primary);
  background: var(--bg-card);
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md, 10px);
  outline: none;
  transition: border-color 0.2s ease;
}

.search-input:focus {
  border-color: var(--accent);
}

.recipe-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  -webkit-overflow-scrolling: touch;
}

.status-msg {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-muted);
  text-align: center;
  padding: 1.5rem 0.5rem;
  margin: 0;
}

.status-msg.error {
  color: var(--accent, #ff6b5b);
}

.recipe-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  min-height: 56px;
  padding: 0.65rem 0.85rem;
  background: var(--bg-card);
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md, 10px);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.recipe-row:hover {
  border-color: var(--accent);
  transform: translateY(-1px);
}

.recipe-row:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px var(--accent-focus-ring);
}

.recipe-row.current {
  border-color: var(--accent);
  background: var(--bg-hover);
}

.row-emoji {
  font-size: 1.5rem;
  line-height: 1;
  flex-shrink: 0;
}

.row-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.row-body {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1;
  min-width: 0;
}

.row-name {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.row-tags {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.7rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.row-current {
  color: var(--accent);
  flex-shrink: 0;
}

.swap-footer {
  padding: 0.75rem 1rem 1rem;
  border-top: 1px solid var(--border-color);
  background: var(--bg-primary);
}

.random-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  min-height: 48px;
  padding: 0 1rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text-on-accent, #fff);
  background: var(--accent);
  border: none;
  border-radius: var(--radius-md, 10px);
  cursor: pointer;
  transition: filter 0.2s ease, transform 0.15s ease;
}

.random-btn:hover:not(:disabled) {
  filter: brightness(1.05);
  transform: translateY(-1px);
}

.random-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { transform: translateY(24px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

/* Desktop: centered card */
@media (min-width: 640px) {
  .swap-backdrop {
    align-items: center;
  }

  .swap-modal {
    border-radius: var(--radius-xl, 20px);
    max-height: 80vh;
  }
}

@media (prefers-reduced-motion: reduce) {
  .swap-backdrop,
  .swap-modal,
  .recipe-row,
  .random-btn {
    animation: none !important;
    transition: none !important;
  }
}
</style>
