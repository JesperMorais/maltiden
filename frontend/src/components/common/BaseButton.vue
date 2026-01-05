<script setup lang="ts">
interface Props {
  variant?: 'primary' | 'secondary' | 'outline'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
}

withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false
})
</script>

<template>
  <button
    class="base-button"
    :class="[`variant-${variant}`, `size-${size}`, { loading, disabled }]"
    :disabled="disabled || loading"
  >
    <span class="button-content" :class="{ invisible: loading }">
      <slot />
    </span>
    <span v-if="loading" class="loader">
      <span class="loader-dot"></span>
      <span class="loader-dot"></span>
      <span class="loader-dot"></span>
    </span>
  </button>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@600;700;800&display=swap');

.base-button {
  --coral: #ff6b5b;
  --coral-dark: #e85a4a;
  --coral-light: #ff8a7d;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --peach: #ffb599;

  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  border: none;
  border-radius: 100px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5em;
  letter-spacing: 0.02em;
}

.base-button::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.base-button:hover:not(:disabled) {
  transform: translateY(-3px) scale(1.02);
}

.base-button:active:not(:disabled) {
  transform: translateY(-1px) scale(0.98);
}

/* Variant: Primary */
.variant-primary {
  background: linear-gradient(135deg, var(--coral) 0%, var(--coral-dark) 100%);
  color: var(--warm-white);
  box-shadow:
    0 4px 14px rgba(255, 107, 91, 0.35),
    0 2px 4px rgba(255, 107, 91, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.variant-primary::before {
  background: linear-gradient(135deg, var(--coral-light) 0%, var(--coral) 100%);
}

.variant-primary:hover:not(:disabled)::before {
  opacity: 1;
}

.variant-primary:hover:not(:disabled) {
  box-shadow:
    0 8px 24px rgba(255, 107, 91, 0.4),
    0 4px 8px rgba(255, 107, 91, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
}

/* Variant: Secondary */
.variant-secondary {
  background: linear-gradient(135deg, var(--cream) 0%, var(--warm-white) 100%);
  color: var(--text-dark);
  box-shadow:
    0 4px 12px rgba(61, 44, 41, 0.08),
    0 2px 4px rgba(61, 44, 41, 0.05),
    inset 0 0 0 2px rgba(255, 107, 91, 0.15);
}

.variant-secondary:hover:not(:disabled) {
  box-shadow:
    0 8px 20px rgba(61, 44, 41, 0.12),
    0 4px 8px rgba(61, 44, 41, 0.08),
    inset 0 0 0 2px var(--coral);
  color: var(--coral-dark);
}

/* Variant: Outline */
.variant-outline {
  background: transparent;
  color: var(--coral);
  box-shadow: inset 0 0 0 2.5px var(--coral);
}

.variant-outline:hover:not(:disabled) {
  background: var(--coral);
  color: var(--warm-white);
  box-shadow:
    inset 0 0 0 2.5px var(--coral),
    0 6px 16px rgba(255, 107, 91, 0.3);
}

/* Sizes */
.size-sm {
  padding: 0.6em 1.4em;
  font-size: 0.875rem;
}

.size-md {
  padding: 0.85em 2em;
  font-size: 1rem;
}

.size-lg {
  padding: 1em 2.5em;
  font-size: 1.125rem;
}

/* States */
.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
}

.loading {
  cursor: wait;
}

.button-content {
  display: flex;
  align-items: center;
  gap: 0.5em;
  position: relative;
  z-index: 1;
  transition: opacity 0.2s ease;
}

.button-content.invisible {
  opacity: 0;
}

.loader {
  position: absolute;
  display: flex;
  gap: 0.3em;
}

.loader-dot {
  width: 0.5em;
  height: 0.5em;
  border-radius: 50%;
  background: currentColor;
  animation: bounce 1.2s infinite ease-in-out;
}

.loader-dot:nth-child(1) { animation-delay: 0s; }
.loader-dot:nth-child(2) { animation-delay: 0.15s; }
.loader-dot:nth-child(3) { animation-delay: 0.3s; }

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1.2);
    opacity: 1;
  }
}
</style>
