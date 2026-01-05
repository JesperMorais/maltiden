import type { LandingPageData } from '@/api/types/landing.types'

/**
 * Mock data for landing page development
 *
 * This mock data follows the exact structure expected from the API.
 * When the backend is ready, simply set VITE_USE_REAL_API=true.
 */
export const mockLandingData: LandingPageData = {
  hero: {
    title: 'Planera veckans måltider enkelt',
    subtitle:
      'Slipp stressen med middagsplanering. Måltiden genererar veckomenyer och inköpslistor automatiskt för hela hushållet.',
    ctaButtonText: 'Kom igång gratis',
    ctaButtonLink: '/register',
    backgroundImageUrl: undefined
  },

  features: {
    sectionTitle: 'Allt du behöver för enklare matvardag',
    features: [
      {
        id: 'auto-menu',
        icon: 'calendar-check',
        title: 'Automatisk veckomeny',
        description:
          'Få en personlig veckomeny genererad baserat på dina preferenser och vad du redan har hemma.',
        order: 1
      },
      {
        id: 'shopping-list',
        icon: 'shopping-cart',
        title: 'Smart inköpslista',
        description:
          'Ingredienser samlas automatiskt till en inköpslista. Fungerar offline i butiken.',
        order: 2
      },
      {
        id: 'household',
        icon: 'users',
        title: 'Hela hushållet',
        description:
          'Bjud in familjemedlemmar och planera måltider tillsammans. Justera portioner per dag.',
        order: 3
      },
      {
        id: 'preferences',
        icon: 'heart',
        title: 'Lär sig vad ni gillar',
        description: 'Gilla eller skippa recept så blir förslagen bättre och bättre över tid.',
        order: 4
      }
    ]
  },

  cta: {
    title: 'Vill du veta mer?',
    description: 'Läs mer om hur Måltiden kan förenkla din vardag och hjälpa hela familjen att planera måltider tillsammans.',
    primaryButton: {
      text: 'Om Måltiden',
      link: '/about'
    }
  },

  meta: {
    pageTitle: 'Måltiden - Enkel måltidsplanering för hela familjen',
    pageDescription:
      'Automatisk veckomeny och smart inköpslista. Planera måltider tillsammans med hela hushållet.',
    lastUpdated: new Date().toISOString()
  }
}
