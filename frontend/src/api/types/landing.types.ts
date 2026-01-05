/**
 * Landing Page API Types
 *
 * This file defines the contract between frontend and backend
 * for the landing page data. Backend team should implement
 * endpoints that return data matching these interfaces.
 *
 * API Endpoint: GET /api/landing
 * Returns: LandingPageResponse
 */

// ============================================
// HERO SECTION TYPES
// ============================================

export interface HeroContent {
  /** Main headline text */
  title: string

  /** Supporting text under the headline */
  subtitle: string

  /** Text for the primary call-to-action button */
  ctaButtonText: string

  /** Route or URL for the CTA button */
  ctaButtonLink: string

  /** Optional URL for background image */
  backgroundImageUrl?: string
}

// ============================================
// FEATURES SECTION TYPES
// ============================================

export interface Feature {
  /** Unique identifier for the feature */
  id: string

  /** Icon identifier (e.g., "calendar", "shopping-cart", "users") */
  icon: string

  /** Feature title */
  title: string

  /** Brief description of the feature */
  description: string

  /** Display order (lower = first) */
  order: number
}

export interface FeaturesContent {
  /** Section heading */
  sectionTitle: string

  /** Array of features to display */
  features: Feature[]
}

// ============================================
// CTA SECTION TYPES
// ============================================

export interface CtaContent {
  /** CTA section headline */
  title: string

  /** Supporting description text */
  description: string

  /** Primary button configuration */
  primaryButton: {
    text: string
    link: string
  }

  /** Optional secondary button */
  secondaryButton?: {
    text: string
    link: string
  }
}

// ============================================
// COMBINED LANDING PAGE DATA
// ============================================

export interface LandingPageData {
  hero: HeroContent
  features: FeaturesContent
  cta: CtaContent

  /** Metadata for SEO/analytics */
  meta?: {
    pageTitle: string
    pageDescription: string
    lastUpdated: string
  }
}

// ============================================
// API RESPONSE WRAPPER
// ============================================

export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
}

export type LandingPageResponse = ApiResponse<LandingPageData>
