<script setup lang="ts">
import BaseCard from '@/components/common/BaseCard.vue'

interface Props {
  icon: string
  title: string
  description: string
}

defineProps<Props>()

// Map icon names to emojis for MVP (can be replaced with SVG icons later)
const iconMap: Record<string, string> = {
  'calendar-check': '📅',
  'shopping-cart': '🛒',
  'users': '👨‍👩‍👧‍👦',
  'heart': '❤️',
  'calendar': '📆',
  'list': '📝',
  'star': '⭐',
  'clock': '⏰'
}
</script>

<template>
  <BaseCard class="feature-card" padding="lg">
    <div class="icon-wrapper">
      <span class="icon">{{ iconMap[icon] || '✨' }}</span>
      <div class="icon-bg"></div>
    </div>
    <h3 class="feature-title">{{ title }}</h3>
    <p class="feature-description">{{ description }}</p>
    <div class="card-decoration"></div>
  </BaseCard>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700&family=Fraunces:wght@700&display=swap');

.feature-card {
  --coral: #ff6b5b;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;

  position: relative;
  text-align: center;
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  overflow: hidden;
}

.feature-card:hover {
  transform: translateY(-8px);
}

.icon-wrapper {
  position: relative;
  width: 80px;
  height: 80px;
  margin: 0 auto 1.5rem;
}

.icon {
  position: relative;
  z-index: 2;
  font-size: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  transition: transform 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.feature-card:hover .icon {
  transform: scale(1.15) rotate(5deg);
}

.icon-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, var(--peach) 0%, var(--coral-light) 100%);
  border-radius: 24px;
  transform: rotate(-6deg);
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  opacity: 0.2;
}

.feature-card:hover .icon-bg {
  transform: rotate(6deg) scale(1.1);
  opacity: 0.3;
}

.feature-title {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.35rem;
  color: var(--text-dark);
  margin: 0 0 0.75rem;
  line-height: 1.3;
}

.feature-description {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  color: var(--text-muted);
  line-height: 1.65;
  margin: 0;
}

.card-decoration {
  position: absolute;
  bottom: -30px;
  right: -30px;
  width: 100px;
  height: 100px;
  background: linear-gradient(135deg, var(--coral) 0%, var(--peach) 100%);
  border-radius: 50%;
  opacity: 0.05;
  transition: all 0.4s ease;
}

.feature-card:hover .card-decoration {
  transform: scale(1.5);
  opacity: 0.08;
}
</style>
