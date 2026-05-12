package domain

import (
	"errors"
	"strings"
	"testing"
)

func happyRecipe() CreateRecipeRequest {
	return CreateRecipeRequest{
		Name:     "Köttbullar med lingonsylt",
		Servings: 4,
		Emoji:    "🍖",
		Tags:     []string{"kött", "klassiker"},
		Ingredients: []Ingredient{
			{Name: "Nötfärs", Amount: 500, Unit: "g"},
			{Name: "Lök", Amount: 1, Unit: "st"},
			{Name: "Ägg", Amount: 2, Unit: "st"},
		},
		Instructions: []string{
			"Blanda färsen med hackad lök och ägg.",
			"Rulla till bollar och stek i smör tills gyllenbruna.",
			"Servera med lingonsylt och potatismos.",
		},
	}
}

func TestCreateRecipeRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*CreateRecipeRequest)
		wantErr error
	}{
		{
			name:    "happy path Swedish recipe",
			mutate:  func(_ *CreateRecipeRequest) {},
			wantErr: nil,
		},
		{
			name:    "name required",
			mutate:  func(r *CreateRecipeRequest) { r.Name = "" },
			wantErr: ErrNameRequired,
		},
		{
			name:    "name exactly 200 runes is ok",
			mutate:  func(r *CreateRecipeRequest) { r.Name = strings.Repeat("å", 200) },
			wantErr: nil,
		},
		{
			name:    "name 201 runes is too long",
			mutate:  func(r *CreateRecipeRequest) { r.Name = strings.Repeat("å", 201) },
			wantErr: ErrNameTooLong,
		},
		{
			name:    "name with control char refused",
			mutate:  func(r *CreateRecipeRequest) { r.Name = "Recept\x01namn" },
			wantErr: ErrContainsControlChar,
		},
		{
			name:    "servings 1 is ok",
			mutate:  func(r *CreateRecipeRequest) { r.Servings = 1 },
			wantErr: nil,
		},
		{
			name:    "servings 50 is ok",
			mutate:  func(r *CreateRecipeRequest) { r.Servings = 50 },
			wantErr: nil,
		},
		{
			name:    "servings 0 invalid",
			mutate:  func(r *CreateRecipeRequest) { r.Servings = 0 },
			wantErr: ErrInvalidServings,
		},
		{
			name:    "servings 51 invalid",
			mutate:  func(r *CreateRecipeRequest) { r.Servings = 51 },
			wantErr: ErrInvalidServings,
		},
		{
			name:    "emoji 8 runes is ok",
			mutate:  func(r *CreateRecipeRequest) { r.Emoji = "abcdefgh" },
			wantErr: nil,
		},
		{
			name:    "emoji 9 runes too long",
			mutate:  func(r *CreateRecipeRequest) { r.Emoji = "abcdefghi" },
			wantErr: ErrEmojiTooLong,
		},
		{
			name:    "empty emoji is ok",
			mutate:  func(r *CreateRecipeRequest) { r.Emoji = "" },
			wantErr: nil,
		},
		{
			name:    "12 tags ok",
			mutate:  func(r *CreateRecipeRequest) { r.Tags = make([]string, 12) },
			wantErr: nil,
		},
		{
			name:    "13 tags too many",
			mutate:  func(r *CreateRecipeRequest) { r.Tags = make([]string, 13) },
			wantErr: ErrTooManyTags,
		},
		{
			name:    "tag exactly 50 runes ok",
			mutate:  func(r *CreateRecipeRequest) { r.Tags = []string{strings.Repeat("ö", 50)} },
			wantErr: nil,
		},
		{
			name:    "tag 51 runes too long",
			mutate:  func(r *CreateRecipeRequest) { r.Tags = []string{strings.Repeat("ö", 51)} },
			wantErr: ErrTagTooLong,
		},
		{
			name:    "tag with control char refused",
			mutate:  func(r *CreateRecipeRequest) { r.Tags = []string{"ok", "bad\x02tag"} },
			wantErr: ErrContainsControlChar,
		},
		{
			name:    "ingredients required",
			mutate:  func(r *CreateRecipeRequest) { r.Ingredients = nil },
			wantErr: ErrIngredientsRequired,
		},
		{
			name: "40 ingredients ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = make([]Ingredient, 40)
				for i := range r.Ingredients {
					r.Ingredients[i] = Ingredient{Name: "x", Amount: 1, Unit: "g"}
				}
			},
			wantErr: nil,
		},
		{
			name: "41 ingredients too many",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = make([]Ingredient, 41)
				for i := range r.Ingredients {
					r.Ingredients[i] = Ingredient{Name: "x", Amount: 1, Unit: "g"}
				}
			},
			wantErr: ErrTooManyIngredients,
		},
		{
			name: "ingredient name required",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "", Amount: 1, Unit: "g"}}
			},
			wantErr: ErrIngredientNameRequired,
		},
		{
			name: "ingredient name exactly 80 runes ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: strings.Repeat("ä", 80), Amount: 1, Unit: "g"}}
			},
			wantErr: nil,
		},
		{
			name: "ingredient name 81 runes too long",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: strings.Repeat("ä", 81), Amount: 1, Unit: "g"}}
			},
			wantErr: ErrIngredientNameTooLong,
		},
		{
			name: "ingredient name control char refused",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "mjöl\x03", Amount: 1, Unit: "g"}}
			},
			wantErr: ErrContainsControlChar,
		},
		{
			name: "amount zero ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "salt", Amount: 0, Unit: "g"}}
			},
			wantErr: nil,
		},
		{
			name: "amount negative invalid",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "salt", Amount: -1, Unit: "g"}}
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "amount 10000 ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "vatten", Amount: 10000, Unit: "ml"}}
			},
			wantErr: nil,
		},
		{
			name: "amount over 10000 too large",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "vatten", Amount: 10001, Unit: "ml"}}
			},
			wantErr: ErrAmountTooLarge,
		},
		{
			name: "unit exactly 20 runes ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "x", Amount: 1, Unit: strings.Repeat("g", 20)}}
			},
			wantErr: nil,
		},
		{
			name: "unit 21 runes too long",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "x", Amount: 1, Unit: strings.Repeat("g", 21)}}
			},
			wantErr: ErrUnitTooLong,
		},
		{
			name: "unit control char refused",
			mutate: func(r *CreateRecipeRequest) {
				r.Ingredients = []Ingredient{{Name: "x", Amount: 1, Unit: "g\x04"}}
			},
			wantErr: ErrContainsControlChar,
		},
		{
			name:    "instructions required",
			mutate:  func(r *CreateRecipeRequest) { r.Instructions = nil },
			wantErr: ErrInstructionsRequired,
		},
		{
			name: "30 instructions ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Instructions = make([]string, 30)
				for i := range r.Instructions {
					r.Instructions[i] = "steg"
				}
			},
			wantErr: nil,
		},
		{
			name: "31 instructions too many",
			mutate: func(r *CreateRecipeRequest) {
				r.Instructions = make([]string, 31)
				for i := range r.Instructions {
					r.Instructions[i] = "steg"
				}
			},
			wantErr: ErrTooManyInstructions,
		},
		{
			name: "instruction exactly 500 runes ok",
			mutate: func(r *CreateRecipeRequest) {
				r.Instructions = []string{strings.Repeat("å", 500)}
			},
			wantErr: nil,
		},
		{
			name: "instruction 501 runes too long",
			mutate: func(r *CreateRecipeRequest) {
				r.Instructions = []string{strings.Repeat("å", 501)}
			},
			wantErr: ErrInstructionTooLong,
		},
		{
			name: "instruction control char refused",
			mutate: func(r *CreateRecipeRequest) {
				r.Instructions = []string{"Koka upp vatten.\x05Salta."}
			},
			wantErr: ErrContainsControlChar,
		},
		{
			// 30 instructions × 500 'å' runes × 2 UTF-8 bytes = 30 000 bytes of text alone,
			// pushing the marshalled payload well over the 32 768-byte limit.
			name: "JSON payload too large",
			mutate: func(r *CreateRecipeRequest) {
				step := strings.Repeat("å", 500)
				r.Instructions = make([]string, 30)
				for i := range r.Instructions {
					r.Instructions[i] = step
				}
				r.Name = strings.Repeat("å", 200)
				r.Tags = make([]string, 12)
				for i := range r.Tags {
					r.Tags[i] = strings.Repeat("ö", 50)
				}
				r.Ingredients = make([]Ingredient, 40)
				for i := range r.Ingredients {
					r.Ingredients[i] = Ingredient{
						Name:   strings.Repeat("ä", 80),
						Amount: 1,
						Unit:   strings.Repeat("å", 20),
					}
				}
			},
			wantErr: ErrRecipePayloadTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := happyRecipe()
			tt.mutate(&req)
			err := req.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected errors.Is(%v), got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateRecipeRequestValidateDelegates(t *testing.T) {
	req := UpdateRecipeRequest{
		Name:         "Pannkaka",
		Servings:     2,
		Ingredients:  []Ingredient{{Name: "Mjöl", Amount: 200, Unit: "g"}},
		Instructions: []string{"Blanda alla ingredienser."},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid UpdateRecipeRequest, got %v", err)
	}

	req.Name = ""
	if err := req.Validate(); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
}
