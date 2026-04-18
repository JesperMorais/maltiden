<script setup lang="ts">
import { computed } from 'vue'
import { Check, Circle } from 'lucide-vue-next'

interface Props {
  password: string
}

const props = defineProps<Props>()

const lengthOk = computed(() => props.password.length >= 8)
const hasUpper = computed(() => /[A-Z]/.test(props.password))
const hasLower = computed(() => /[a-z]/.test(props.password))
const hasDigit = computed(() => /[0-9]/.test(props.password))
const hasSpecial = computed(() => /[^A-Za-z0-9]/.test(props.password))

const typesMet = computed(() =>
  [hasUpper.value, hasLower.value, hasDigit.value, hasSpecial.value].filter(Boolean).length,
)

const requirements = computed(() => [
  { label: 'Minst 8 tecken', met: lengthOk.value },
  {
    label: `Minst 3 av 4 teckenslag (${typesMet.value}/4)`,
    met: typesMet.value >= 3,
  },
])

const typeChips = computed(() => [
  { label: 'Versal', met: hasUpper.value },
  { label: 'Gemen', met: hasLower.value },
  { label: 'Siffra', met: hasDigit.value },
  { label: 'Specialtecken', met: hasSpecial.value },
])
</script>

<template>
  <div class="pw-strength" role="list" aria-label="Lösenordskrav">
    <div
      v-for="req in requirements"
      :key="req.label"
      role="listitem"
      class="pw-req"
      :class="{ met: req.met }"
    >
      <Check v-if="req.met" :size="14" :stroke-width="3" class="pw-icon" />
      <Circle v-else :size="14" :stroke-width="2" class="pw-icon" />
      <span>{{ req.label }}</span>
    </div>

    <div class="pw-chips" aria-label="Teckenslag">
      <span
        v-for="chip in typeChips"
        :key="chip.label"
        class="pw-chip"
        :class="{ met: chip.met }"
      >
        {{ chip.label }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.pw-strength {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-top: 0.5rem;
  padding: 0.75rem 0.85rem;
  border-radius: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
}

.pw-req {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-family: 'Nunito', sans-serif;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-secondary);
  transition: color 0.2s ease;
}

.pw-req.met {
  color: var(--success);
}

.pw-icon {
  flex-shrink: 0;
}

.pw-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.25rem;
}

.pw-chip {
  font-family: 'Nunito', sans-serif;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  transition: all 0.2s ease;
}

.pw-chip.met {
  background: var(--success-bg);
  border-color: var(--success);
  color: var(--success);
}
</style>
