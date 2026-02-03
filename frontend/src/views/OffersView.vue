<script setup lang="ts">
import OfferSearch from '@/components/poc/OfferSearch.vue'
</script>

<template>
  <main class="offers-view">
    <OfferSearch />
  </main>
</template>

<style scoped>
.offers-view {
  min-height: 100vh;
  background: var(--bg-secondary, #f5f5f5);
  padding-top: 2rem;
  padding-bottom: 4rem;
}
</style>
