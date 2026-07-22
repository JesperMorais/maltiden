<script setup lang="ts">
import { computed } from 'vue'
import { Motion } from 'motion-v'
import { Sparkles } from 'lucide-vue-next'

interface Props {
  /** Short Swedish explanation from the AI arranger; empty/absent renders nothing. */
  rationale?: string
}

const props = withDefaults(defineProps<Props>(), {
  rationale: '',
})

const text = computed(() => props.rationale.trim())
const hasRationale = computed(() => text.value.length > 0)
</script>

<template>
  <Motion
    v-if="hasRationale"
    tag="section"
    class="menu-rationale"
    aria-label="Förklaring av veckans meny"
    :initial="{ opacity: 0, y: 16 }"
    :animate="{ opacity: 1, y: 0 }"
    :transition="{ duration: 0.4, ease: 'easeOut' }"
  >
    <div class="rationale-header">
      <Sparkles :size="20" :stroke-width="2" class="header-icon" />
      <h2 class="rationale-title">Därför funkar veckan</h2>
    </div>
    <p class="rationale-text">{{ text }}</p>
  </Motion>
</template>

<style scoped>
.menu-rationale {
  margin-top: var(--space-lg);
  padding: var(--space-lg);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

.rationale-header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  margin-bottom: var(--space-sm);
}

.header-icon {
  color: var(--accent);
  flex-shrink: 0;
}

.rationale-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

.rationale-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0;
}

@media (max-width: 768px) {
  .menu-rationale {
    padding: var(--space-md);
  }

  .rationale-title {
    font-size: 1.1rem;
  }
}
</style>
