<script setup lang="ts">
import { ref } from 'vue'
import RecipeListPanel from '@/components/recipes/RecipeListPanel.vue'
import AddRecipePanel from '@/components/recipes/AddRecipePanel.vue'
import FadeContent from '@/components/vue-bits/FadeContent.vue'
import BackLink from '@/components/common/BackLink.vue'

type Tab = 'list' | 'add'
const activeTab = ref<Tab>('list')

const recipeListRef = ref<InstanceType<typeof RecipeListPanel> | null>(null)

function switchToAdd() {
  activeTab.value = 'add'
}

function switchToList() {
  activeTab.value = 'list'
  recipeListRef.value?.refresh()
}
</script>

<template>
  <div class="recipes-view">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <BackLink :to="{ name: 'dashboard' }" label="Dashboard" />
        <h1 class="title">Recept</h1>
        <p class="description">
          Hantera dina recept — bläddra, sök eller lägg till nya.
        </p>

        <!-- Segmented tab control -->
        <div class="tab-control">
          <button
            class="tab-button"
            :class="{ active: activeTab === 'list' }"
            @click="switchToList"
          >
            Mina recept
          </button>
          <button
            class="tab-button"
            :class="{ active: activeTab === 'add' }"
            @click="switchToAdd"
          >
            Lägg till
          </button>
        </div>
      </div>
    </header>

    <!-- Main content -->
    <main class="content">
      <div class="content-container">
        <FadeContent :duration="500" :blur="true">
          <RecipeListPanel
            v-show="activeTab === 'list'"
            ref="recipeListRef"
            @navigate-to-add="switchToAdd"
          />
        </FadeContent>
        <FadeContent v-if="activeTab === 'add'" :duration="500" :blur="true">
          <AddRecipePanel
            @navigate-to-list="switchToList"
          />
        </FadeContent>
      </div>
    </main>
  </div>
</template>

<style scoped>
.recipes-view {
  min-height: 100vh;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

/* Header */
.header {
  padding: 2rem 2rem 0;
  background: linear-gradient(
    180deg,
    var(--bg-card) 0%,
    var(--bg-primary) 100%
  );
  border-bottom: 1px solid var(--border-color);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
}

.header-content :deep(.back-link) {
  margin-bottom: 0.5rem;
  margin-left: -1rem;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
  line-height: 1.2;
}

.description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0 0 1.5rem;
  line-height: 1.6;
}

/* Tab control */
.tab-control {
  display: inline-flex;
  background: var(--bg-secondary);
  border-radius: 14px;
  padding: 4px;
  gap: 4px;
  margin-bottom: -1px;
}

.tab-button {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  padding: 0.6rem 1.5rem;
  min-height: 44px;
  border: none;
  border-radius: 11px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.25s ease;
}

.tab-button:hover:not(.active) {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.5);
}

.tab-button.active {
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* Content */
.content {
  flex: 1;
  padding: 2rem;
  position: relative;
}

.content-container {
  max-width: 1200px;
  margin: 0 auto;
  position: relative;
}

/* Responsive */
@media (max-width: 768px) {
  .header {
    padding: 1.5rem 1rem 0;
  }

  .content {
    padding: 1.5rem 1rem;
  }

  .title {
    font-size: 1.75rem;
  }

  .description {
    font-size: 0.95rem;
  }
}
</style>
