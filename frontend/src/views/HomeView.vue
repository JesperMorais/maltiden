<!-- frontend/src/views/HomeView.vue -->
 <script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { checkHealth } from '@/api/client';

  const backendStatus = ref<'checking' | 'online' | 'offline'>('checking')

  onMounted(async () => {
    backendStatus.value = (await checkHealth()) ? 'online' : 'offline'
  })
 </script>

<template>
  <div>
    <h1>Måltiden</h1>
    <p>Backend: {{ backendStatus }}</p>
  </div>
</template>
