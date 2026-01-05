<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main class="about-page">
    <!-- Background decorations -->
    <div class="bg-decorations">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
      <div class="grain"></div>
    </div>

    <!-- Navigation -->
    <nav class="about-nav">
      <RouterLink to="/" class="back-link">
        <span class="back-arrow">←</span>
        <span>Tillbaka</span>
      </RouterLink>
    </nav>

    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content-simple">
        <div class="hero-badge">
          <span class="badge-icon">💝</span>
          <span>Vår historia</span>
        </div>
        <h1 class="hero-title">Om Måltiden</h1>
        <p class="hero-subtitle">
          Vi gör måltidsplanering till en fröjd för svenska familjer
        </p>
      </div>

      <!-- Decorative food elements -->
      <div class="hero-floats">
        <span class="float-item float-1">🥗</span>
        <span class="float-item float-2">🍲</span>
        <span class="float-item float-3">🥘</span>
      </div>
    </section>

    <!-- Mission Section -->
    <section class="mission-section">
      <div class="mission-container">
        <div class="mission-header">
          <span class="section-icon">✨</span>
          <h2>Vår Mission</h2>
        </div>

        <p class="mission-text">
          Måltiden är en smart måltidsplaneringsapp skapad för svenska familjer.
          Vi tror att bra mat börjar med bra planering – och att planering inte
          behöver vara tråkigt eller tidskrävande.
        </p>

        <div class="values-grid">
          <div class="value-card">
            <div class="value-icon">🍽️</div>
            <h3>Enklare vardag</h3>
            <p>Slipp frågan "vad ska vi äta?" – vi hjälper dig planera hela veckans måltider på några minuter.</p>
          </div>

          <div class="value-card">
            <div class="value-icon">🌱</div>
            <h3>Mindre matsvinn</h3>
            <p>Smarta inköpslistor och portion­sberäkning så att du köper precis det du behöver.</p>
          </div>

          <div class="value-card">
            <div class="value-icon">👨‍👩‍👧‍👦</div>
            <h3>Familjetid</h3>
            <p>Samla familjen kring middags­bordet med måltider alla älskar. Mindre stress, mer gemenskap.</p>
          </div>
        </div>

        <div class="mission-quote">
          <blockquote>
            "Vi vill att varje familj ska kunna njuta av god mat utan krångel"
          </blockquote>
          <div class="quote-decoration">
            <span>🍳</span>
            <span>❤️</span>
            <span>🥄</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Team Section -->
    <section class="team-section">
      <div class="team-container">
        <div class="team-header">
          <span class="section-icon">👥</span>
          <h2>Teamet bakom Måltiden</h2>
          <p>Tre passionerade utvecklare med en gemensam kärlek för god mat och smart teknik</p>
        </div>

        <div class="team-grid">
          <article class="team-card team-card-david">
            <div class="card-accent"></div>
            <div class="member-avatar">
              <img src="@/assets/images/david.jpeg" alt="David" class="avatar-image" />
              <div class="avatar-ring"></div>
            </div>
            <h3 class="member-name">David</h3>
            <p class="member-role">Backend-utvecklare</p>
            <p class="member-desc">
              Ser till att allt fungerar smidigt bakom kulisserna.
            </p>
          </article>

          <article class="team-card team-card-jesper">
            <div class="card-accent"></div>
            <div class="member-avatar">
              <img src="@/assets/images/jesper.jpeg" alt="Jesper" class="avatar-image" />
              <div class="avatar-ring"></div>
            </div>
            <h3 class="member-name">Jesper</h3>
            <p class="member-role">Frontend-utvecklare</p>
            <p class="member-desc">
              Skapar den användarvänliga upplevelsen du ser framför dig.
            </p>
          </article>

          <article class="team-card team-card-philip">
            <div class="card-accent"></div>
            <div class="member-avatar">
              <img src="@/assets/images/philip.jpeg" alt="Philip" class="avatar-image" />
              <div class="avatar-ring"></div>
            </div>
            <h3 class="member-name">Philip</h3>
            <p class="member-role">Testare</p>
            <p class="member-desc">
              Ser till att allt fungerar perfekt och felfritt.
            </p>
          </article>
        </div>

        <div class="team-footer">
          <div class="footer-emoji">🍝</div>
          <p>Tillsammans gör vi måltidsplanering roligare!</p>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="cta-section">
      <div class="cta-container">
        <h2>Redo att förenkla din matvardag?</h2>
        <p>Gå med tusentals svenska familjer som redan använder Måltiden</p>
        <RouterLink to="/register" class="cta-button">
          Kom igång gratis
          <span class="btn-arrow">→</span>
        </RouterLink>
      </div>
    </section>
  </main>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700;800&family=Fraunces:wght@700;800&display=swap');

.about-page {
  --coral: #ff6b5b;
  --coral-dark: #e85a4a;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;
  --yellow-soft: #ffd93d;
  --orange-soft: #ffab5e;

  min-height: 100vh;
  background: var(--cream);
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
  opacity: 0.5;
}

.blob-1 {
  width: 600px;
  height: 600px;
  background: linear-gradient(135deg, var(--peach) 0%, var(--coral-light) 100%);
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

.blob-3 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, var(--coral-light) 0%, var(--peach) 100%);
  top: 60%;
  right: -100px;
  animation: float-slow 20s ease-in-out infinite;
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
  color: var(--text-muted);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 100px;
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(10px);
}

.back-link:hover {
  color: var(--coral);
  background: rgba(255, 107, 91, 0.15);
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
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(255, 107, 91, 0.1);
  color: var(--coral);
  padding: 0.5rem 1rem;
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
  animation: fade-in-up 0.8s ease-out 0.1s backwards;
}

.badge-icon {
  animation: pulse 2s ease-in-out infinite;
}

.hero-title {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(2.5rem, 6vw, 3.5rem);
  color: var(--text-dark);
  margin: 0 0 1rem;
  animation: fade-in-up 0.8s ease-out 0.2s backwards;
}

.hero-subtitle {
  font-family: 'Nunito', sans-serif;
  font-size: clamp(1.1rem, 2vw, 1.35rem);
  color: var(--text-muted);
  margin: 0;
  max-width: 500px;
  margin-inline: auto;
  line-height: 1.7;
  animation: fade-in-up 0.8s ease-out 0.3s backwards;
}

.hero-floats {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.float-item {
  position: absolute;
  font-size: 2.5rem;
  animation: float 6s ease-in-out infinite;
  filter: drop-shadow(0 4px 8px rgba(61, 44, 41, 0.2));
}

.float-1 { top: 10%; left: 5%; animation-delay: 0s; }
.float-2 { top: 20%; right: 8%; animation-delay: 1s; }
.float-3 { bottom: 15%; left: 10%; animation-delay: 2s; }

/* Mission Section */
.mission-section {
  position: relative;
  z-index: 1;
  padding: 6rem 2rem;
  background: var(--warm-white);
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

.section-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
  animation: bounce 2s ease-in-out infinite;
}

.mission-header h2,
.team-header h2 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: clamp(2rem, 4vw, 2.75rem);
  color: var(--text-dark);
  margin: 0;
}

.team-header p {
  font-family: 'Nunito', sans-serif;
  font-size: 1.1rem;
  color: var(--text-muted);
  margin: 1rem 0 0;
  max-width: 500px;
  margin-inline: auto;
}

.mission-text {
  font-family: 'Nunito', sans-serif;
  font-size: 1.2rem;
  line-height: 1.8;
  color: var(--text-muted);
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
  background: white;
  border-radius: 24px;
  padding: 2rem;
  text-align: center;
  box-shadow:
    0 10px 40px rgba(61, 44, 41, 0.06),
    0 0 0 1px rgba(255, 107, 91, 0.08);
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.value-card:hover {
  transform: translateY(-8px);
  box-shadow:
    0 20px 60px rgba(61, 44, 41, 0.1),
    0 0 0 1px rgba(255, 107, 91, 0.15);
}

.value-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.value-card h3 {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--text-dark);
  margin: 0 0 0.75rem;
}

.value-card p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.95rem;
  color: var(--text-muted);
  line-height: 1.6;
  margin: 0;
}

.mission-quote {
  text-align: center;
  padding: 3rem;
  background: linear-gradient(135deg, rgba(255, 107, 91, 0.08) 0%, rgba(255, 181, 153, 0.08) 100%);
  border-radius: 24px;
  position: relative;
}

.mission-quote blockquote {
  font-family: 'Fraunces', serif;
  font-weight: 700;
  font-size: clamp(1.25rem, 3vw, 1.5rem);
  color: var(--text-dark);
  font-style: italic;
  margin: 0 0 1.5rem;
  line-height: 1.5;
}

.quote-decoration {
  display: flex;
  justify-content: center;
  gap: 1rem;
  font-size: 1.5rem;
}

.quote-decoration span {
  animation: bounce 2s ease-in-out infinite;
}

.quote-decoration span:nth-child(2) { animation-delay: 0.2s; }
.quote-decoration span:nth-child(3) { animation-delay: 0.4s; }

/* Team Section */
.team-section {
  position: relative;
  z-index: 1;
  padding: 6rem 2rem;
  background: linear-gradient(180deg, var(--warm-white) 0%, var(--cream) 100%);
}

.team-container {
  max-width: 1100px;
  margin: 0 auto;
}

.team-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin-bottom: 4rem;
}

.team-card {
  position: relative;
  background: white;
  border-radius: 28px;
  padding: 2.5rem 2rem;
  text-align: center;
  box-shadow:
    0 15px 50px rgba(61, 44, 41, 0.08),
    0 0 0 1px rgba(255, 107, 91, 0.1);
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  overflow: hidden;
}

.team-card:hover {
  transform: translateY(-10px) scale(1.02);
  box-shadow:
    0 25px 70px rgba(61, 44, 41, 0.12),
    0 0 0 2px rgba(255, 107, 91, 0.2);
}

.card-accent {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 6px;
  background: linear-gradient(90deg, var(--coral) 0%, var(--peach) 100%);
}

.team-card-david .card-accent {
  background: linear-gradient(90deg, #4a90d9 0%, #7ab8ff 100%);
}

.team-card-philip .card-accent {
  background: linear-gradient(90deg, #4ade80 0%, #86efac 100%);
}

.member-avatar {
  position: relative;
  width: 140px;
  height: 140px;
  margin: 0 auto 1.5rem;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
  box-shadow:
    0 8px 25px rgba(61, 44, 41, 0.15);
}

.avatar-emoji {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  font-size: 3rem;
  background: linear-gradient(135deg, var(--cream) 0%, white 100%);
  border-radius: 50%;
  box-shadow:
    0 8px 25px rgba(61, 44, 41, 0.1),
    inset 0 -3px 10px rgba(61, 44, 41, 0.05);
}

.avatar-ring {
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  border: 3px solid transparent;
  background: linear-gradient(135deg, var(--coral), var(--peach)) border-box;
  -webkit-mask:
    linear-gradient(#fff 0 0) padding-box,
    linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  animation: spin 8s linear infinite;
}

.team-card-david .avatar-ring {
  background: linear-gradient(135deg, #4a90d9, #7ab8ff) border-box;
}

.team-card-philip .avatar-ring {
  background: linear-gradient(135deg, #4ade80, #86efac) border-box;
}

.member-name {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 1.5rem;
  color: var(--text-dark);
  margin: 0 0 0.25rem;
}

.member-role {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--coral);
  margin: 0 0 1rem;
}

.team-card-david .member-role { color: #4a90d9; }
.team-card-philip .member-role { color: #22c55e; }

.member-desc {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-muted);
  line-height: 1.6;
  margin: 0;
}

.member-decoration {
  position: absolute;
  bottom: 1rem;
  right: 1.5rem;
  font-size: 1.5rem;
  opacity: 0.3;
  transition: all 0.3s ease;
}

.team-card:hover .member-decoration {
  opacity: 0.6;
  transform: scale(1.2) rotate(10deg);
}

.team-footer {
  text-align: center;
}

.footer-emoji {
  font-size: 3rem;
  margin-bottom: 1rem;
  animation: wiggle 3s ease-in-out infinite;
}

.team-footer p {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--text-muted);
  margin: 0;
}

/* CTA Section */
.cta-section {
  position: relative;
  z-index: 1;
  padding: 6rem 2rem;
  background: linear-gradient(135deg, var(--coral) 0%, var(--peach) 100%);
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
  color: white;
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
  color: var(--coral);
  background: white;
  padding: 1rem 2.5rem;
  border-radius: 100px;
  text-decoration: none;
  box-shadow: 0 10px 30px rgba(61, 44, 41, 0.2);
  transition: all 0.3s ease;
}

.cta-button:hover {
  transform: translateY(-3px);
  box-shadow: 0 15px 40px rgba(61, 44, 41, 0.25);
}

.btn-arrow {
  transition: transform 0.3s ease;
}

.cta-button:hover .btn-arrow {
  transform: translateX(4px);
}

/* Animations */
@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-15px) rotate(5deg); }
}

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

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes wiggle {
  0%, 100% { transform: rotate(-5deg); }
  50% { transform: rotate(5deg); }
}

/* Responsive */
@media (max-width: 768px) {
  .hero-section {
    padding: 2rem 1.5rem 4rem;
  }

  .float-item {
    display: none;
  }

  .mission-section,
  .team-section,
  .cta-section {
    padding: 4rem 1.5rem;
  }

  .team-grid {
    grid-template-columns: 1fr;
    max-width: 400px;
    margin-inline: auto;
  }
}
</style>
