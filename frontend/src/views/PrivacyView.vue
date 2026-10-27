<script setup lang="ts">
import { RouterLink } from 'vue-router'
import BackLink from '@/components/common/BackLink.vue'

// Keep this page in sync with what the backend actually stores
// (backend/migrations) and which third parties it calls. Update the date
// below whenever the content changes.
const lastUpdated = '2026-09-24'
const contactEmail = 'maltiden.app@gmail.com'
</script>

<template>
  <main class="privacy-page">
    <nav class="privacy-nav">
      <BackLink to="/" />
    </nav>

    <article class="privacy-content">
      <h1>Integritetspolicy</h1>
      <p class="updated">Senast uppdaterad {{ lastUpdated }}</p>

      <p>
        Här beskriver vi vilka uppgifter Måltiden sparar om dig, varför, vem vi delar dem med och
        hur du bestämmer över dem. Vi samlar bara in det som behövs för att appen ska fungera. Vi
        säljer inga uppgifter och använder ingen reklam eller spårning.
      </p>

      <h2>Vem ansvarar för dina uppgifter?</h2>
      <p>
        Måltiden drivs av teamet bakom appen (se <RouterLink to="/about">Om oss</RouterLink>). Har
        du frågor om dina uppgifter, mejla <a :href="`mailto:${contactEmail}`">{{ contactEmail }}</a
        >.
      </p>

      <h2>Vad vi sparar</h2>
      <ul>
        <li>
          <strong>Konto:</strong> e-postadress, förnamn och ditt lösenord i krypterad form (vi kan
          aldrig se lösenordet). Efternamnet du anger vid registrering används bara för att föreslå
          ett hushållsnamn och sparas inte separat.
        </li>
        <li>
          <strong>Hushåll:</strong> hushållets namn, vilka som är medlemmar och deras roll, när du
          gick med, om du äter hemma i dag eller vill ha matlåda, och inbjudningskoder.
        </li>
        <li>
          <strong>Mat och planering:</strong> recept som ert hushåll lägger till, veckomenyer,
          inköpslistan (bockade och egna varor) och menyinställningar som taggar ni vill undvika,
          ingredienser ni inte gillar och antal vegetariska dagar.
        </li>
        <li>
          <strong>Feedback:</strong> om du skickar feedback i appen sparar vi ditt betyg, din
          kommentar, vilken sida du var på, skärmbredd och webbläsarens användaragent.
        </li>
        <li><strong>Lösenordsåterställning:</strong> en engångslänk som gäller i en timme.</li>
      </ul>
      <p>
        Allt detta ligger i vår databas hos Fly.io i Stockholm. I din webbläsare sparar appen din
        inloggning (i upp till 7 dagar), valt tema och några planeringsval, så att du slipper göra
        om dem. Vi använder inga spårningscookies.
      </p>

      <h2>Varför vi sparar det</h2>
      <p>
        Uppgifterna behövs för att ge dig tjänsten du registrerat dig för: konto, delat hushåll,
        menyer och inköpslistor (avtal, artikel 6.1 b i GDPR). Felrapporter och tillfälliga
        IP-adresser för att stoppa missbruk behandlar vi för att hålla tjänsten säker och fungerande
        (berättigat intresse, artikel 6.1 f). Feedback är frivillig.
      </p>

      <h2>Vilka vi delar uppgifter med</h2>
      <p>Vi använder några underleverantörer. De får bara det som behövs för just deras uppgift:</p>
      <ul>
        <li><strong>Fly.io</strong>: driftar servern och databasen (Stockholm).</li>
        <li>
          <strong>Anthropic (Claude)</strong>: när du väljer att låta AI tolka ett recept skickas
          texten du klistrar in dit. Inga kontouppgifter skickas med.
        </li>
        <li>
          <strong>Google (Gemini)</strong>: när du skriver önskemål till veckomenyn eller låter AI
          ordna veckan skickas önskemålen och namnen på veckans recept dit. Inga kontouppgifter
          skickas med.
        </li>
        <li>
          <strong>Resend</strong>: skickar mejlet när du begär ett nytt lösenord, så det får din
          e-postadress och återställningslänken.
        </li>
        <li>
          <strong>Sentry</strong>: tar emot felrapporter när något går sönder, med teknisk
          information som webbläsare och vilken sida som kraschade. Vi har stängt av att
          personuppgifter skickas med automatiskt.
        </li>
        <li>
          <strong>Google Fonts</strong>: typsnitten laddas från Google, så din webbläsare kontaktar
          Googles servrar (som då ser din IP-adress).
        </li>
      </ul>
      <p>
        Några av dessa företag finns i USA. Överföringen sker med de skydd som GDPR kräver, till
        exempel EU:s standardavtalsklausuler eller EU–US Data Privacy Framework.
      </p>

      <h2>Hur länge vi sparar uppgifterna</h2>
      <p>
        Vi sparar dina uppgifter så länge du har ett konto. När du raderar kontot tas uppgifterna
        bort direkt ur databasen. De kan finnas kvar i säkerhetskopior av databasen, som raderas
        automatiskt efter högst fem dagar.
      </p>

      <h2>Dina rättigheter</h2>
      <ul>
        <li>
          <strong>Se och ta med dina uppgifter:</strong> under
          <RouterLink to="/konto">Mitt konto</RouterLink> kan du ladda ner allt vi har om dig och
          ditt hushåll som en fil.
        </li>
        <li>
          <strong>Radera ditt konto:</strong> också under Mitt konto. Är du ensam i hushållet
          raderas hushållets data också. Finns det andra medlemmar lämnar du hushållet och deras
          gemensamma data finns kvar.
        </li>
        <li><strong>Rätta eller invända:</strong> mejla oss så hjälper vi dig.</li>
        <li>
          <strong>Klaga:</strong> tycker du att vi hanterar dina uppgifter fel kan du vända dig till
          Integritetsskyddsmyndigheten (IMY), imy.se.
        </li>
      </ul>
    </article>
  </main>
</template>

<style scoped>
.privacy-page {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: 1.5rem 1rem 4rem;
}

.privacy-nav {
  max-width: 720px;
  margin: 0 auto 1rem;
}

.privacy-content {
  max-width: 720px;
  margin: 0 auto;
  font-family: 'Nunito', sans-serif;
  color: var(--text-primary);
  line-height: 1.6;
}

.privacy-content h1 {
  font-family: 'Fraunces', serif;
  font-weight: 800;
  font-size: 2.25rem;
  line-height: 1.2;
  margin: 0 0 0.25rem;
}

.privacy-content h2 {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.15rem;
  margin: 2rem 0 0.5rem;
}

.privacy-content p,
.privacy-content li {
  font-size: 1rem;
}

.privacy-content ul {
  padding-left: 1.25rem;
  margin: 0.5rem 0;
}

.privacy-content li {
  margin-bottom: 0.5rem;
}

.privacy-content a {
  color: var(--accent-text, var(--accent));
  overflow-wrap: anywhere;
}

.updated {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin: 0 0 1.5rem;
}
</style>
