<script setup lang="ts">
import type { ShoppingListSummary } from '@/api/types/dashboard.types'

interface Props {
  shoppingList: ShoppingListSummary | null
}

defineProps<Props>()

const emit = defineEmits<{
  'view-list': []
}>()
</script>

<template>
  <section class="shopping-widget" @click="emit('view-list')">
    <div class="widget-content">
      <div class="shopping-icon">🛒</div>

      <div class="shopping-info" v-if="shoppingList">
        <h3 class="widget-title">Inköpslista</h3>
        <div class="shopping-stats">
          <span class="items-remaining">
            {{ shoppingList.totalItems - shoppingList.checkedItems }} varor kvar
          </span>
          <span class="items-total">
            av {{ shoppingList.totalItems }}
          </span>
        </div>

        <!-- Progress bar -->
        <div class="progress-bar">
          <div
            class="progress-fill"
            :style="{
              width: `${(shoppingList.checkedItems / shoppingList.totalItems) * 100}%`
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
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@600;700;800&display=swap');

.shopping-widget {
  --coral: #ff6b5b;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;
  --green: #48bb78;

  background: linear-gradient(135deg, var(--warm-white) 0%, var(--cream) 100%);
  border-radius: 20px;
  padding: 1.25rem;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow:
    0 4px 20px rgba(61, 44, 41, 0.05),
    0 0 0 1px rgba(255, 107, 91, 0.06);
}

.shopping-widget:hover {
  transform: translateY(-2px);
  box-shadow:
    0 8px 30px rgba(61, 44, 41, 0.1),
    0 0 0 1px rgba(255, 107, 91, 0.15);
}

.widget-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.shopping-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.shopping-info {
  flex: 1;
  min-width: 0;
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-dark);
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
  color: var(--coral);
}

.items-total {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.progress-bar {
  width: 100%;
  height: 6px;
  background: rgba(61, 44, 41, 0.1);
  border-radius: 100px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--green) 0%, #68d391 100%);
  border-radius: 100px;
  transition: width 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.empty-text {
  font-family: 'Nunito', sans-serif;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.view-arrow {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--coral);
  opacity: 0.5;
  transition: all 0.3s ease;
}

.shopping-widget:hover .view-arrow {
  opacity: 1;
  transform: translateX(4px);
}
</style>
