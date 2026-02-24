<script setup lang="ts">
import { computed } from 'vue'
import type { ShoppingListSummary } from '@/api/types/dashboard.types'
import CountUp from '@/components/vue-bits/CountUp.vue'
import SpotlightCard from '@/components/vue-bits/SpotlightCard.vue'
import { ShoppingCart } from 'lucide-vue-next'

interface Props {
  shoppingList: ShoppingListSummary | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'view-list': []
}>()

const remainingItems = computed(() =>
  props.shoppingList ? props.shoppingList.totalItems - props.shoppingList.checkedItems : 0,
)
</script>

<template>
  <SpotlightCard spotlight-color="rgba(104, 211, 145, 0.15)" class-name="shopping-spotlight">
    <section class="shopping-widget" @click="emit('view-list')">
      <div class="widget-content">
        <div class="shopping-icon"><ShoppingCart :size="28" /></div>

        <div class="shopping-info" v-if="shoppingList">
          <h3 class="widget-title">Inköpslista</h3>
          <div class="shopping-stats">
            <span class="items-remaining">
              <CountUp :to="remainingItems" :duration="1.5" /> varor kvar
            </span>
            <span class="items-total">
              av <CountUp :to="shoppingList.totalItems" :duration="1.5" :delay="0.2" />
            </span>
          </div>

          <!-- Progress bar -->
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{
                width: `${(shoppingList.checkedItems / shoppingList.totalItems) * 100}%`,
              }"
            ></div>
          </div>
        </div>

        <div class="shopping-info" v-else>
          <h3 class="widget-title">Inköpslista</h3>
          <span class="empty-text">Ingen lista ännu</span>
        </div>

        <div class="view-arrow">→</div>
      </div>
    </section>
  </SpotlightCard>
</template>

<style scoped>
.shopping-widget {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 1.25rem;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.shopping-widget:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: var(--border-color-hover);
}

.widget-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.shopping-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--accent);
}

.shopping-info {
  flex: 1;
  min-width: 0;
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-primary);
  margin: 0 0 0.25rem;
}

.shopping-stats {
  display: flex;
  align-items: baseline;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}

.items-remaining {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--accent);
}

.items-total {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.progress-bar {
  width: 100%;
  height: 6px;
  background: var(--border-color);
  border-radius: 100px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--success) 0%, #68d391 100%);
  border-radius: 100px;
  transition: width 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.empty-text {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.view-arrow {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--accent);
  opacity: 0.5;
  transition: all 0.3s ease;
}

.shopping-widget:hover .view-arrow {
  opacity: 1;
  transform: translateX(4px);
}
</style>
