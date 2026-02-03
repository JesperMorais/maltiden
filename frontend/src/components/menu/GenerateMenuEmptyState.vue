<script setup lang="ts">
const emit = defineEmits<{
  generate: []
}>()
</script>

<template>
  <div class="empty-state">
    <!-- Emoji illustration -->
    <div class="illustration">
      <span class="emoji calendar">🗓️</span>
      <span class="emoji plate">🍽️</span>
    </div>

    <!-- Content -->
    <div class="content">
      <h2 class="title">Skapa din veckomeny</h2>
      <p class="description">
        Generera en meny för 5 dagar med slumpmässiga recept. Du kan låsa dagar du vill behålla och
        generera nya för resten.
      </p>

      <!-- Features list -->
      <ul class="features">
        <li class="feature">
          <span class="feature-icon">📅</span>
          <span class="feature-text">5 måltider (Måndag-Fredag)</span>
        </li>
        <li class="feature">
          <span class="feature-icon">🎲</span>
          <span class="feature-text">Slumpmässiga recept</span>
        </li>
        <li class="feature">
          <span class="feature-icon">🔒</span>
          <span class="feature-text">Lås dagar du vill behålla</span>
        </li>
      </ul>

      <!-- Generate button -->
      <button class="generate-button" @click="emit('generate')">
        <span class="button-text">Generera meny</span>
        <span class="button-icon">✨</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  max-width: 600px;
  margin: 0 auto;
}

/* Illustration */
.illustration {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.emoji {
  font-size: 4rem;
  animation: bounce 2s ease-in-out infinite;
}

.emoji.calendar {
  animation-delay: 0s;
}

.emoji.plate {
  animation-delay: 0.2s;
}

@keyframes bounce {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-15px);
  }
}

/* Content */
.content {
  width: 100%;
}

.title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2rem;
  color: var(--text-primary);
  margin: 0 0 1rem 0;
  line-height: 1.2;
}

.description {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0 0 2rem 0;
  line-height: 1.6;
}

/* Features list */
.features {
  list-style: none;
  padding: 0;
  margin: 0 0 2.5rem 0;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.feature {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1rem;
  color: var(--text-primary);
}

.feature-icon {
  font-size: 1.5rem;
  line-height: 1;
}

.feature-text {
  line-height: 1;
}

/* Generate button */
.generate-button {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 2.5rem;
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-dark) 100%);
  color: white;
  border: none;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.1rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 4px 12px rgba(255, 107, 91, 0.3);
}

.generate-button:hover {
  transform: translateY(-2px) scale(1.05);
  box-shadow: 0 8px 20px rgba(255, 107, 91, 0.4);
}

.generate-button:active {
  transform: translateY(0) scale(0.98);
}

.button-text {
  line-height: 1;
}

.button-icon {
  font-size: 1.3rem;
  line-height: 1;
}

/* Responsive */
@media (max-width: 768px) {
  .empty-state {
    padding: 3rem 1.5rem;
  }

  .illustration .emoji {
    font-size: 3rem;
  }

  .title {
    font-size: 1.75rem;
  }

  .description {
    font-size: 1rem;
  }

  .feature {
    font-size: 0.95rem;
  }

  .generate-button {
    padding: 0.875rem 2rem;
    font-size: 1rem;
  }
}
</style>
