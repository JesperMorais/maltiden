package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	validRecipe := func() CreateRecipeRequest {
		return CreateRecipeRequest{
			Name:     "Pasta Bolognese",
			Servings: 4,
			Emoji:    "🍝",
			Tags:     []string{"pasta", "kött"},
			Ingredients: []Ingredient{
				{Name: "pasta", Amount: 400, Unit: "g"},
			},
			Instructions: []string{"Koka pastan.", "Rör ihop såsen."},
		}
	}

	tests := []struct {
		name    string
		req     CreateRecipeRequest
		wantErr error
	}{
		{
			name:    "valid recipe passes",
			req:     validRecipe(),
			wantErr: nil,
		},

		// Name
		{
			name:    "name empty",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Name = ""; return r }(),
			wantErr: ErrNameRequired,
		},
		{
			name:    "name exactly 200 runes",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Name = strings.Repeat("å", 200); return r }(),
			wantErr: nil,
		},
		{
			name:    "name 201 runes",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Name = strings.Repeat("å", 201); return r }(),
			wantErr: ErrNameTooLong,
		},
		{
			name:    "name contains control char",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Name = "bad\x01name"; return r }(),
			wantErr: ErrContainsControlChar,
		},

		// Servings
		{
			name:    "servings 1",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Servings = 1; return r }(),
			wantErr: nil,
		},
		{
			name:    "servings 50",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Servings = 50; return r }(),
			wantErr: nil,
		},
		{
			name:    "servings 0",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Servings = 0; return r }(),
			wantErr: ErrInvalidServings,
		},
		{
			name:    "servings 51",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Servings = 51; return r }(),
			wantErr: ErrInvalidServings,
		},

		// Emoji
		{
			name:    "emoji exactly 8 runes",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Emoji = "🍝🍝🍝🍝🍝🍝🍝🍝"; return r }(),
			wantErr: nil,
		},
		{
			name: "emoji 9 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Emoji = "🍝🍝🍝🍝🍝🍝🍝🍝🍝"
				return r
			}(),
			wantErr: ErrEmojiTooLong,
		},

		// Tags
		{
			name: "exactly 12 tags",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Tags = make([]string, 12)
				for i := range r.Tags {
					r.Tags[i] = "tag"
				}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "13 tags",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Tags = make([]string, 13)
				for i := range r.Tags {
					r.Tags[i] = "tag"
				}
				return r
			}(),
			wantErr: ErrTooManyTags,
		},
		{
			name:    "tag exactly 50 runes",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Tags = []string{strings.Repeat("å", 50)}; return r }(),
			wantErr: nil,
		},
		{
			name:    "tag 51 runes",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Tags = []string{strings.Repeat("å", 51)}; return r }(),
			wantErr: ErrTagTooLong,
		},
		{
			name:    "tag contains control char",
			req:     func() CreateRecipeRequest { r := validRecipe(); r.Tags = []string{"bad\x01tag"}; return r }(),
			wantErr: ErrContainsControlChar,
		},

		// Ingredients
		{
			name: "exactly 40 ingredients",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = make([]Ingredient, 40)
				for i := range r.Ingredients {
					r.Ingredients[i] = Ingredient{Name: "salt", Amount: 1, Unit: "g"}
				}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "41 ingredients",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = make([]Ingredient, 41)
				for i := range r.Ingredients {
					r.Ingredients[i] = Ingredient{Name: "salt", Amount: 1, Unit: "g"}
				}
				return r
			}(),
			wantErr: ErrTooManyIngredients,
		},
		{
			name: "ingredient name exactly 80 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: strings.Repeat("å", 80), Amount: 1, Unit: "g"}}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "ingredient name 81 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: strings.Repeat("å", 81), Amount: 1, Unit: "g"}}
				return r
			}(),
			wantErr: ErrIngredientNameTooLong,
		},
		{
			name: "ingredient name control char",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "bad\x01", Amount: 1, Unit: "g"}}
				return r
			}(),
			wantErr: ErrContainsControlChar,
		},
		{
			name: "amount negative",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "salt", Amount: -1, Unit: "g"}}
				return r
			}(),
			wantErr: ErrInvalidAmount,
		},
		{
			name: "amount exactly 10000",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "salt", Amount: 10000, Unit: "g"}}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "amount 10001",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "salt", Amount: 10001, Unit: "g"}}
				return r
			}(),
			wantErr: ErrAmountTooLarge,
		},
		{
			name: "unit exactly 20 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "salt", Amount: 1, Unit: strings.Repeat("å", 20)}}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "unit 21 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "salt", Amount: 1, Unit: strings.Repeat("å", 21)}}
				return r
			}(),
			wantErr: ErrUnitTooLong,
		},
		{
			name: "unit control char",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Ingredients = []Ingredient{{Name: "salt", Amount: 1, Unit: "g\x01"}}
				return r
			}(),
			wantErr: ErrContainsControlChar,
		},

		// Instructions
		{
			name: "exactly 30 instructions",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Instructions = make([]string, 30)
				for i := range r.Instructions {
					r.Instructions[i] = "Steg."
				}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "31 instructions",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Instructions = make([]string, 31)
				for i := range r.Instructions {
					r.Instructions[i] = "Steg."
				}
				return r
			}(),
			wantErr: ErrTooManyInstructions,
		},
		{
			name: "instruction exactly 500 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Instructions = []string{strings.Repeat("å", 500)}
				return r
			}(),
			wantErr: nil,
		},
		{
			name: "instruction 501 runes",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Instructions = []string{strings.Repeat("å", 501)}
				return r
			}(),
			wantErr: ErrInstructionTooLong,
		},
		{
			name: "instruction control char",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				r.Instructions = []string{"bad\x01step"}
				return r
			}(),
			wantErr: ErrContainsControlChar,
		},

		// Total size
		{
			name: "total JSON over 32KB",
			req: func() CreateRecipeRequest {
				r := validRecipe()
				// 40 ingredients × 80-rune name + 30 instructions × 500 runes ≈ 3200 + 15000 bytes
				// pad name to tip over 32KB total
				r.Name = strings.Repeat("å", 200)
				r.Tags = make([]string, 12)
				for i := range r.Tags {
					r.Tags[i] = strings.Repeat("å", 50)
				}
				r.Ingredients = make([]Ingredient, 40)
				for i := range r.Ingredients {
					r.Ingredients[i] = Ingredient{Name: strings.Repeat("å", 80), Amount: 1, Unit: strings.Repeat("å", 20)}
				}
				// å is 2 UTF-8 bytes → 30×500 runes = 30000 bytes in JSON, pushes total >32KB
				r.Instructions = make([]string, 30)
				for i := range r.Instructions {
					r.Instructions[i] = strings.Repeat("å", 500)
				}
				return r
			}(),
			wantErr: ErrRecipeTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.Validate()
			if tt.wantErr == nil {
				if got != nil {
					t.Errorf("Validate() = %v, want nil", got)
				}
				return
			}
			if !errors.Is(got, tt.wantErr) {
				t.Errorf("Validate() = %v, want errors.Is(%v)", got, tt.wantErr)
			}
		})
	}
}
