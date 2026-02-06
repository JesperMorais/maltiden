<script setup lang="ts">
import SkeletonBlock from './SkeletonBlock.vue'

interface Props {
  size?: string
}

withDefaults(defineProps<Props>(), {
  size: '36px'
})
</script>

<template>
  <SkeletonBlock :width="size" :height="size" radius="50%" />
</template>
