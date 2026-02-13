<script setup lang="ts">
interface Props {
  loading: boolean
  transition?: boolean
}

withDefaults(defineProps<Props>(), {
  transition: true
})
</script>

<template>
  <Transition v-if="transition" name="skeleton-fade" mode="out-in">
    <div v-if="loading" key="skeleton" class="skeleton-state" aria-hidden="true" role="presentation">
      <slot name="skeleton" />
    </div>
    <div v-else key="content" class="content-state">
      <slot />
    </div>
  </Transition>
  <template v-else>
    <div v-if="loading" class="skeleton-state" aria-hidden="true" role="presentation">
      <slot name="skeleton" />
    </div>
    <div v-else class="content-state">
      <slot />
    </div>
  </template>
</template>

<style scoped>
.skeleton-fade-enter-active,
.skeleton-fade-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.skeleton-fade-enter-from {
  opacity: 0;
  transform: scale(0.98);
}

.skeleton-fade-leave-to {
  opacity: 0;
}
</style>
