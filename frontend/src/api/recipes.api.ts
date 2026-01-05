/**
 * Recipes API Service
 * Handles recipe CRUD operations
 */

import apiClient from './client'
import { USE_MOCKS } from '@/mocks'
import { mockGetRecipes, mockGetRecipe, mockCreateRecipe } from '@/mocks/recipes.mock'

export interface Ingredient {
  name: string
  amount: number
  unit: string
}

export interface RecipeSummary {
  id: string
  name: string
  servings: number
  tags: string[]
  emoji?: string
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
