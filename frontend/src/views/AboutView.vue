<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { UtensilsCrossed, Sprout, UsersRound } from 'lucide-vue-next'
import SpotlightCard from '@/components/vue-bits/SpotlightCard.vue'
import FadeContent from '@/components/vue-bits/FadeContent.vue'
</script>

<template>
  <main class="about-page">
    <!-- Background decorations -->
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="grain"></div>
    </div>

    <!-- Navigation -->
    <nav class="about-nav">
      <RouterLink v-prefetch="'landing'" to="/" class="back-link">
        <span class="back-arrow">←</span>
        <span>Tillbaka</span>
      </RouterLink>
    </nav>

    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content-simple">
        <div class="hero-badge">Vår historia</div>
        <h1 class="hero-title">Om Måltiden</h1>
        <p class="hero-subtitle">
          Vi gör måltidsplanering till en fröjd för svenska familjer
        </p>
      </div>
    </section>

    <!-- Mission Section -->
    <section class="mission-section">
      <div class="mission-container">
        <div class="mission-header">
          <h2>Vår Mission</h2>
        </div>

        <p class="mission-text">
          Måltiden är en smart måltidsplaneringsapp skapad för svenska familjer.
          Vi tror att bra mat börjar med bra planering – och att planering inte
          behöver vara tråkigt eller tidskrävande.
        </p>

        <div class="values-grid">
          <div class="value-card">
            <div class="value-icon">
              <UtensilsCrossed :size="28" :stroke-width="1.75" />
            </div>
            <h3>Enklare vardag</h3>
            <p>Slipp frågan "vad ska vi äta?" – vi hjälper dig planera hela veckans måltider på några minuter.</p>
          </div>

          <div class="value-card">
            <div class="value-icon">
              <Sprout :size="28" :stroke-width="1.75" />
            </div>
            <h3>Mindre matsvinn</h3>
            <p>Smarta inköpslistor och portion­sberäkning så att du köper precis det du behöver.</p>
          </div>

          <div class="value-card">
            <div class="value-icon">
              <UsersRound :size="28" :stroke-width="1.75" />
            </div>
            <h3>Familjetid</h3>
            <p>Samla familjen kring middags­bordet med måltider alla älskar. Mindre stress, mer gemenskap.</p>
          </div>
        </div>

        <div class="mission-quote">
          <blockquote>
            "Vi vill att varje familj ska kunna njuta av god mat utan krångel"
          </blockquote>
        </div>
      </div>
    </section>

    <!-- Team Section -->
    <section class="team-section">
      <div class="team-container">
        <div class="team-header">
          <h2>Teamet bakom Måltiden</h2>
        </div>

        <div class="team-grid">
          <FadeContent :blur="true" :duration="800" :delay="100">
            <SpotlightCard
              spotlight-color="rgba(74, 144, 217, 0.15)"
              border-radius="24px"
              class-name="team-card"
            >
              <div class="team-card-inner">
                <div class="member-initial member-initial-david">D</div>
                <h3 class="member-name">David</h3>
                <span class="member-role">Utvecklare</span>
              </div>
            </SpotlightCard>
          </FadeContent>

          <FadeContent :blur="true" :duration="800" :delay="250">
            <SpotlightCard
              spotlight-color="rgba(255, 107, 91, 0.15)"
              border-radius="24px"
              class-name="team-card"
            >
              <div class="team-card-inner">
                <div class="member-initial member-initial-jesper">J</div>
                <h3 class="member-name">Jesper</h3>
                <span class="member-role">Utvecklare</span>
              </div>
            </SpotlightCard>
          </FadeContent>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="cta-section">
      <div class="cta-container">
        <h2>Redo att förenkla din matvardag?</h2>
        <p>Gör måltidsplaneringen enklare för hela familjen</p>
        <RouterLink to="/register" class="cta-button">
          Kom igång gratis
          <span class="btn-arrow">→</span>
        </RouterLink>
      </div>
    </section>
  </main>
</template>

<style scoped>
.about-page {
  min-height: 100vh;
  background: var(--bg-secondary);
  position: relative;
  overflow-x: hidden;
}

/* Background decorations */
.bg-decorations {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.3;
}

.blob-1 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, var(--peach) 0%, var(--accent-light) 100%);
  top: -200px;
  right: -200px;
  animation: float-slow 25s ease-in-out infinite;
}

.blob-2 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, var(--yellow-soft) 0%, var(--orange-soft) 100%);
  bottom: 20%;
  left: -200px;
  animation: float-slow 30s ease-in-out infinite reverse;
}

.grain {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 400 400' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
  opacity: 0.03;
}

/* Navigation */
.about-nav {
  position: relative;
  z-index: 10;
  padding: 1.5rem 2rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 100px;
  transition: all 0.3s ease;
  background: var(--bg-card);
  backdrop-filter: blur(10px);
}

.back-link:hover {
  color: var(--accent);
  background: var(--accent-bg);
}

.back-arrow {
  transition: transform 0.3s ease;
}

.back-link:hover .back-arrow {
  transform: translateX(-4px);
}

/* Hero Section */
.hero-section {
  position: relative;
  z-index: 1;
  padding: 4rem 2rem 6rem;
  max-width: 800px;
  margin: 0 auto;
}

.hero-content-simple {
  text-align: center;
  animation: fade-in-up 0.8s ease-out;
}

.hero-badge {
  display: inline-block;
  background: var(--accent-bg);
  color: var(--accent);
  padding: 0.5rem 1.25rem;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
  animation: fade-in-up 0.8s ease-out 0.1s backwards;
}

.hero-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(2.5rem, 6vw, 3.5rem);
  color: var(--text-primary);
  margin: 0 0 1rem;
  animation: fade-in-up 0.8s ease-out 0.2s backwards;
}

.hero-subtitle {
  font-family: 'Nunito', sans-serif;
  font-size: clamp(1.1rem, 2vw, 1.35rem);
  color: var(--text-secondary);
  margin: 0;
  max-width: 500px;
  margin-inline: auto;
  line-height: 1.7;
  animation: fade-in-up 0.8s ease-out 0.3s backwards;
}

/* Mission Section */
.mission-section {
  position: relative;
  z-index: 1;
  padding: 6rem 2rem;
  background: var(--bg-primary);
}

.mission-container {
  max-width: 1000px;
  margin: 0 auto;
}

.mission-header,
.team-header {
  text-align: center;
  margin-bottom: 3rem;
}

.mission-header h2,
.team-header h2 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(2rem, 4vw, 2.75rem);
  color: var(--text-primary);
  margin: 0;
}

.mission-text {
  font-family: 'Nunito', sans-serif;
  font-size: 1.2rem;
  line-height: 1.8;
  color: var(--text-secondary);
  text-align: center;
  max-width: 700px;
  margin: 0 auto 4rem;
}

.values-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
  margin-bottom: 4rem;
}

.value-card {
  background: var(--bg-card);
  border-radius: 24px;
  padding: 2rem;
  text-align: center;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.value-card:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-md);
  border-color: var(--accent);
}

.value-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: var(--accent-bg);
  color: var(--accent);
  margin: 0 auto 1rem;
}

.value-card h3 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0 0 0.75rem;
}

.value-card p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
}

.mission-quote {
  text-align: center;
  padding: 3rem;
  background: var(--accent-bg);
  border-radius: 24px;
  position: relative;
}

.mission-quote blockquote {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: clamp(1.25rem, 3vw, 1.5rem);
  color: var(--text-primary);
  font-style: italic;
  margin: 0;
  line-height: 1.5;
}

/* Team Section */
.team-section {
  position: relative;
  z-index: 1;
  padding: 6rem 2rem;
  background: linear-gradient(180deg, var(--bg-primary) 0%, var(--bg-secondary) 100%);
}

.team-container {
  max-width: 640px;
  margin: 0 auto;
}

.team-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 2rem;
}

.team-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.team-card:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-lg);
  border-color: var(--accent);
}

.team-card-inner {
  padding: 3rem 2rem;
  text-align: center;
}

.member-initial {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 96px;
  height: 96px;
  border-radius: 50%;
  margin: 0 auto 1.5rem;
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.25rem;
  color: var(--text-on-accent);
  letter-spacing: -0.02em;
}

.member-initial-david {
  background: linear-gradient(135deg, #4a90d9 0%, #7ab8ff 100%);
  box-shadow: 0 8px 24px rgba(74, 144, 217, 0.3);
}

.member-initial-jesper {
  background: linear-gradient(135deg, var(--accent) 0%, var(--peach) 100%);
  box-shadow: 0 8px 24px rgba(255, 107, 91, 0.3);
}

.member-name {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.5rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.member-role {
  display: inline-block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-secondary);
}

/* CTA Section */
.cta-section {
  position: relative;
  z-index: 1;
  padding: 6rem 2rem;
  background: linear-gradient(135deg, var(--accent) 0%, var(--peach) 100%);
  text-align: center;
}

.cta-container {
  max-width: 600px;
  margin: 0 auto;
}

.cta-section h2 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  color: var(--text-on-accent);
  margin: 0 0 1rem;
}

.cta-section p {
  font-family: 'Nunito', sans-serif;
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.9);
  margin: 0 0 2rem;
}

.cta-button {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.1rem;
  color: var(--accent);
  background: var(--bg-card);
  padding: 1rem 2.5rem;
  border-radius: 100px;
  text-decoration: none;
  box-shadow: var(--shadow-md);
  transition: all 0.3s ease;
}

.cta-button:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg);
}

.btn-arrow {
  transition: transform 0.3s ease;
}

.cta-button:hover .btn-arrow {
  transform: translateX(4px);
}

/* Animations */
@keyframes float-slow {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, 30px); }
}

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .hero-section {
    padding: 2rem 1.5rem 4rem;
  }

  .mission-section,
  .team-section,
  .cta-section {
    padding: 4rem 1.5rem;
  }

  .team-grid {
    grid-template-columns: 1fr;
    max-width: 360px;
    margin-inline: auto;
  }
}
</style>
