/**
 * Recipes API Mock Data
 */

import type { RecipeSummary, Recipe, CreateRecipeRequest, ParseRecipeRequest, ParseRecipeResponse } from '@/api/recipes.api'

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

export const mockRecipes: Recipe[] = [
  {
    id: 'rec_1',
    name: 'Pasta Carbonara',
    servings: 4,
    emoji: '🍝',
    tags: ['pasta', 'vardag', 'snabb'],
    ingredients: [
      { name: 'Spaghetti', amount: 400, unit: 'g' },
      { name: 'Bacon', amount: 200, unit: 'g' },
      { name: 'Ägg', amount: 4, unit: 'st' },
      { name: 'Parmesan', amount: 100, unit: 'g' },
      { name: 'Svartpeppar', amount: 1, unit: 'tsk' }
    ],
    instructions: [
      'Koka pastan enligt förpackningen',
      'Stek baconet krispigt',
      'Vispa ihop ägg och parmesan',
      'Blanda pasta med bacon',
      'Rör ner äggblandningen',
      'Servera med extra parmesan'
    ]
  },
  {
    id: 'rec_2',
    name: 'Kycklingwok',
    servings: 4,
    emoji: '🥘',
    tags: ['kyckling', 'asiatiskt', 'vardag'],
    ingredients: [
      { name: 'Kycklingfilé', amount: 500, unit: 'g' },
      { name: 'Wokgrönsaker', amount: 400, unit: 'g' },
      { name: 'Sojasås', amount: 3, unit: 'msk' },
      { name: 'Ris', amount: 4, unit: 'dl' }
    ],
    instructions: [
      'Skär kycklingen i strimlor',
      'Stek kycklingen i het wok',
      'Tillsätt grönsaker',
      'Krydda med sojasås',
      'Servera med ris'
    ]
  },
  {
    id: 'rec_3',
    name: 'Tacos',
    servings: 4,
    emoji: '🌮',
    tags: ['mexikanskt', 'fredagsmys', 'barn'],
    ingredients: [
      { name: 'Köttfärs', amount: 500, unit: 'g' },
      { name: 'Tacokrydda', amount: 1, unit: 'påse' },
      { name: 'Tacoskal', amount: 12, unit: 'st' },
      { name: 'Sallad', amount: 1, unit: 'st' },
      { name: 'Tomat', amount: 3, unit: 'st' },
      { name: 'Riven ost', amount: 200, unit: 'g' }
    ],
    instructions: [
      'Stek köttfärsen',
      'Tillsätt krydda och vatten',
      'Låt sjuda 5 min',
      'Skär grönsaker',
      'Servera med tillbehör'
    ]
  },
  {
    id: 'rec_4',
    name: 'Laxfilé med potatis',
    servings: 4,
    emoji: '🐟',
    tags: ['fisk', 'nyttigt', 'vardag'],
    ingredients: [
      { name: 'Laxfilé', amount: 600, unit: 'g' },
      { name: 'Potatis', amount: 800, unit: 'g' },
      { name: 'Citron', amount: 1, unit: 'st' },
      { name: 'Dill', amount: 1, unit: 'knippe' }
    ],
    instructions: [
      'Sätt ugnen på 200°C',
      'Koka potatisen',
      'Lägg laxen i ugnsform',
      'Salta, peppra och lägg på citron',
      'Grädda 15-20 min',
      'Servera med dill'
    ]
  },
  {
    id: 'rec_5',
    name: 'Köttfärssås',
    servings: 4,
    emoji: '🍖',
    tags: ['pasta', 'klassiker', 'barn'],
    ingredients: [
      { name: 'Köttfärs', amount: 400, unit: 'g' },
      { name: 'Krossade tomater', amount: 400, unit: 'g' },
      { name: 'Lök', amount: 1, unit: 'st' },
      { name: 'Vitlök', amount: 2, unit: 'klyftor' },
      { name: 'Pasta', amount: 400, unit: 'g' }
    ],
    instructions: [
      'Hacka lök och vitlök',
      'Bryn köttfärsen',
      'Tillsätt lök och vitlök',
      'Häll i krossade tomater',
      'Låt sjuda 20 min',
      'Servera med pasta'
    ]
  },
  {
    id: 'rec_6',
    name: 'Vegetarisk curry',
    servings: 4,
    emoji: '🥗',
    tags: ['vegetariskt', 'vardag'],
    ingredients: [
      { name: 'Kikärtor', amount: 400, unit: 'g' },
      { name: 'Kokosmjölk', amount: 400, unit: 'ml' },
      { name: 'Currypasta', amount: 2, unit: 'msk' },
      { name: 'Spenat', amount: 200, unit: 'g' },
      { name: 'Ris', amount: 4, unit: 'dl' }
    ],
    instructions: [
      'Fräs currypastan i olja',
      'Tillsätt kokosmjölk och kikärtor',
      'Låt sjuda 15 min',
      'Vänd ner spenaten',
      'Servera med ris'
    ]
  },
  {
    id: 'rec_7',
    name: 'Pizza',
    servings: 4,
    emoji: '🍕',
    tags: ['fredagsmys', 'barn'],
    ingredients: [
      { name: 'Pizzadeg', amount: 1, unit: 'st' },
      { name: 'Tomatsås', amount: 2, unit: 'dl' },
      { name: 'Mozzarella', amount: 200, unit: 'g' },
      { name: 'Skinka', amount: 150, unit: 'g' },
      { name: 'Champinjoner', amount: 100, unit: 'g' }
    ],
    instructions: [
      'Sätt ugnen på 250°C',
      'Kavla ut degen',
      'Bred på tomatsås',
      'Lägg på topping och ost',
      'Grädda 10-12 min'
    ]
  }
]

export async function mockGetRecipes(): Promise<{ recipes: RecipeSummary[] }> {
  await delay(300)

  const summaries: RecipeSummary[] = mockRecipes.map(({ id, name, servings, tags, emoji }) => ({
    id,
    name,
    servings,
    tags,
    emoji
  }))

  return { recipes: summaries }
}

export async function mockGetRecipe(id: string): Promise<Recipe> {
  await delay(200)

  const recipe = mockRecipes.find(r => r.id === id)
  if (!recipe) {
    throw { response: { status: 404, data: { error: 'not_found' } } }
  }

  return { ...recipe }
}

export async function mockCreateRecipe(recipe: CreateRecipeRequest): Promise<{ id: string }> {
  await delay(500)

  const newId = 'rec_' + Date.now()
  mockRecipes.push({
    id: newId,
    name: recipe.name,
    servings: recipe.servings,
    emoji: recipe.emoji,
    tags: recipe.tags ?? [],
    ingredients: recipe.ingredients,
    instructions: recipe.instructions,
  })
  return { id: newId }
}

export async function mockUpdateRecipe(id: string, recipe: CreateRecipeRequest): Promise<Recipe> {
  await delay(400)

  const existing = mockRecipes.find((r) => r.id === id)
  if (!existing) {
    throw { response: { status: 404, data: { error: 'not_found' } } }
  }

  const updated: Recipe = { ...existing, ...recipe }
  const index = mockRecipes.indexOf(existing)
  mockRecipes[index] = updated
  return { ...updated }
}

export async function mockDeleteRecipe(id: string): Promise<void> {
  await delay(300)

  const index = mockRecipes.findIndex((r) => r.id === id)
  if (index === -1) {
    throw { response: { status: 404, data: { error: 'not_found' } } }
  }

  mockRecipes.splice(index, 1)
}

export async function mockParseRecipe(_req: ParseRecipeRequest): Promise<ParseRecipeResponse> {
  await delay(1500)

  return {
    recipe: {
      name: 'Pasta Carbonara',
      servings: 4,
      emoji: '🍝',
      ingredients: [
        { name: 'spaghetti', amount: 400, unit: 'g' },
        { name: 'guanciale eller bacon', amount: 200, unit: 'g' },
        { name: 'äggulor', amount: 4, unit: 'st' },
        { name: 'parmesan, riven', amount: 100, unit: 'g' },
        { name: 'svartpeppar', amount: 0, unit: 'efter smak' }
      ],
      instructions: [
        'Koka pastan enligt förpackningen',
        'Stek baconet knaprigt',
        'Vispa ihop äggulor och parmesan',
        'Blanda het pasta med bacon, ta från värmen',
        'Rör ner äggblandningen, salta och peppra'
      ],
      tags: ['pasta', 'italienskt', 'snabb']
    },
    confidence: 0.95,
    warnings: [],
    rawText: _req.rawText
  }
}

export async function mockParseAndSaveRecipe(req: ParseRecipeRequest): Promise<ParseRecipeResponse & { id: string }> {
  const parsed = await mockParseRecipe(req)
  return {
    ...parsed,
    id: 'rec_' + Date.now()
  }
}
