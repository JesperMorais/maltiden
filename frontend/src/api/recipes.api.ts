/**
 * Recipes API Service
 * Handles recipe CRUD operations
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import {
  mockGetRecipes,
  mockGetRecipe,
  mockCreateRecipe,
  mockUpdateRecipe,
  mockDeleteRecipe,
  mockParseRecipe,
  mockParseAndSaveRecipe,
} from '@/mocks/recipes.mock'

export interface Ingredient {
  name: string
  amount: number
  unit: string
  /** Swedish food database ("Livsmedelsdatabasen") item number, used for nutrition lookup. */
  livsmedelsnummer?: number
}

/** Per-serving nutrition values. Grams except `kcal`. */
export interface Macros {
  kcal: number
  protein: number
  carbs: number
  fat: number
}

export interface RecipeSummary {
  id: string
  name: string
  servings: number
  tags: string[]
  emoji?: string
  /** Per-serving nutrition, present only for enriched recipes. */
  macros?: Macros
}

export interface Recipe extends RecipeSummary {
  ingredients: Ingredient[]
  instructions: string[]
}

export interface CreateRecipeRequest {
  name: string
  servings: number
  ingredients: Ingredient[]
  instructions: string[]
  tags: string[]
  emoji?: string
}

export interface ParseRecipeRequest {
  rawText: string
  source?: string
}

export interface ParseRecipeResponse {
  recipe: CreateRecipeRequest & { emoji?: string }
  confidence: number
  warnings?: string[]
  rawText: string
}

/**
 * Get all recipes
 */
export async function getRecipes(): Promise<{ recipes: RecipeSummary[] }> {
  if (USE_MOCKS) {
    return mockGetRecipes()
  }

  const { data } = await apiClient.get<{ recipes: RecipeSummary[] }>('/recipes')
  return data
}

/**
 * Get a single recipe by ID
 */
export async function getRecipe(id: string): Promise<Recipe> {
  if (USE_MOCKS) {
    return mockGetRecipe(id)
  }

  const { data } = await apiClient.get<Recipe>(`/recipes/${id}`)
  return data
}

/**
 * Create a new recipe (admin only)
 */
export async function createRecipe(recipe: CreateRecipeRequest): Promise<{ id: string }> {
  if (USE_MOCKS) {
    return mockCreateRecipe(recipe)
  }

  const { data } = await apiClient.post<{ id: string }>('/recipes', recipe)
  return data
}

/**
 * Parse unstructured recipe text into structured data
 */
export async function parseRecipe(req: ParseRecipeRequest): Promise<ParseRecipeResponse> {
  if (USE_MOCKS) {
    return mockParseRecipe(req)
  }

  const { data } = await apiClient.post<ParseRecipeResponse>('/recipes/parse', req, { timeout: 60000 })
  return data
}

/**
 * Update an existing recipe
 */
export async function updateRecipe(id: string, recipe: CreateRecipeRequest): Promise<Recipe> {
  if (USE_MOCKS) {
    return mockUpdateRecipe(id, recipe)
  }

  const { data } = await apiClient.put<Recipe>(`/recipes/${id}`, recipe)
  return data
}

/**
 * Delete a recipe
 */
export async function deleteRecipe(id: string): Promise<void> {
  if (USE_MOCKS) {
    return mockDeleteRecipe(id)
  }

  await apiClient.delete(`/recipes/${id}`)
}

/**
 * Parse recipe text and save immediately
 */
export async function parseAndSaveRecipe(req: ParseRecipeRequest): Promise<ParseRecipeResponse & { id: string }> {
  if (USE_MOCKS) {
    return mockParseAndSaveRecipe(req)
  }

  const { data } = await apiClient.post<ParseRecipeResponse & { id: string }>('/recipes/parse-and-save', req, { timeout: 60000 })
  return data
}
