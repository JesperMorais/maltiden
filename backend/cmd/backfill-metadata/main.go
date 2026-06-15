// Command backfill-metadata is a standalone, human-run tool that enriches
// existing recipes with metadata (recipe-level: mainProtein/dietClass/batchable/
// cookMinutes; per-ingredient: canonicalName/gramsEquiv/isPantryStaple/
// isPerishable) using the Claude Message Batches API with Haiku 4.5.
//
// It is NEVER wired into server startup or CI. It requires ANTHROPIC_API_KEY
// and an explicit --confirm flag, prints an estimated cost before submitting,
// and is idempotent (re-running only touches recipes still lacking metadata).
//
//	go run ./cmd/backfill-metadata --confirm
//	go run ./cmd/backfill-metadata --confirm --limit 10
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/claude"
	"maltiden/pkg/gemini"
)

// approxCostPerRecipeUSD is a rough Haiku 4.5 Batch cost estimate per recipe,
// used only to print a heads-up before submitting. Not load-bearing.
const approxCostPerRecipeUSD = 0.0014

// approxGeminiCostPerRecipeUSD is the equivalent rough estimate for Gemini 2.5
// Flash-Lite ($0.10/$0.40 per MTok), ~5x cheaper than Haiku.
const approxGeminiCostPerRecipeUSD = 0.0003

// pollInterval is how long to wait between batch status polls.
const pollInterval = 45 * time.Second

// metadataSchema is the structured-output schema the batch path constrains
// Claude's response to. It mirrors the metadata we merge back into a recipe.
var metadataSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"mainProtein": map[string]interface{}{"type": "string"},
		"dietClass": map[string]interface{}{
			"type": "string",
			"enum": []string{"omnivore", "vegetarian", "vegan", "pescetarian"},
		},
		"batchable":   map[string]interface{}{"type": "boolean"},
		"cookMinutes": map[string]interface{}{"type": "integer"},
		"ingredients": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":           map[string]interface{}{"type": "string"},
					"canonicalName":  map[string]interface{}{"type": "string"},
					"gramsEquiv":     map[string]interface{}{"type": "number"},
					"isPantryStaple": map[string]interface{}{"type": "boolean"},
					"isPerishable":   map[string]interface{}{"type": "boolean"},
				},
				"required":             []string{"name", "canonicalName", "gramsEquiv", "isPantryStaple", "isPerishable"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []string{"mainProtein", "dietClass", "batchable", "cookMinutes", "ingredients"},
	"additionalProperties": false,
}

const backfillSystemPrompt = `You enrich Swedish recipes with metadata for Måltiden, a meal planning app.

Given a recipe (name, servings, ingredients, instructions), emit JSON metadata:
- "mainProtein": dominant protein in Swedish (e.g. "kyckling", "nötkött", "lax", "linser"); "" if none.
- "dietClass": one of "omnivore", "vegetarian", "vegan", "pescetarian".
- "batchable": true if the dish reheats/freezes well for batch cooking.
- "cookMinutes": total active cooking time in minutes (0 if unknown).
- "ingredients": one entry per recipe ingredient, matched by "name", with:
    - "canonicalName": normalized Swedish lemma ("riven ost" → "ost", "kycklingfilé" → "kyckling"), lowercase singular.
    - "gramsEquiv": best-effort grams for the listed amount (0 if unknown).
    - "isPantryStaple": true for shelf-stable staples (salt, olja, mjöl, ...).
    - "isPerishable": true for fresh items that spoil quickly (kött, fisk, mejeri, färska grönsaker).`

// recipeMetadata is the structured payload Claude returns per recipe.
type recipeMetadata struct {
	MainProtein string                   `json:"mainProtein"`
	DietClass   string                   `json:"dietClass"`
	Batchable   bool                     `json:"batchable"`
	CookMinutes int                      `json:"cookMinutes"`
	Ingredients []ingredientMetadata     `json:"ingredients"`
}

type ingredientMetadata struct {
	Name           string  `json:"name"`
	CanonicalName  string  `json:"canonicalName"`
	GramsEquiv     float64 `json:"gramsEquiv"`
	IsPantryStaple bool    `json:"isPantryStaple"`
	IsPerishable   bool    `json:"isPerishable"`
}

func main() {
	confirm := flag.Bool("confirm", false, "required: actually call the API and spend money")
	limit := flag.Int("limit", 0, "optional: process at most N recipes (0 = all lacking metadata)")
	provider := flag.String("provider", "claude", "LLM provider: claude (Haiku Batch) | gemini (Flash-Lite, offline enrichment)")
	flag.Parse()

	if *provider != "claude" && *provider != "gemini" {
		log.Fatalf("unknown provider %q (want claude|gemini)", *provider)
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/maltiden.db"
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	recipes, err := selectRecipesNeedingMetadata(db, *limit)
	if err != nil {
		log.Fatalf("select recipes: %v", err)
	}

	if len(recipes) == 0 {
		log.Println("No recipes need metadata — nothing to do.")
		return
	}

	costPer := approxCostPerRecipeUSD
	label := "Haiku 4.5 Batch"
	if *provider == "gemini" {
		costPer = approxGeminiCostPerRecipeUSD
		label = "Gemini 2.5 Flash-Lite"
	}
	log.Printf("%d recipes need metadata. Estimated cost: ~$%.4f USD (%s).", len(recipes), float64(len(recipes))*costPer, label)

	if !*confirm {
		log.Println("Dry run — pass --confirm to call the API and spend money.")
		return
	}

	storage := sqlite.NewRecipeStorage(db)
	ctx := context.Background()

	if *provider == "gemini" {
		client, err := gemini.NewClient(gemini.FlashLiteModel)
		if err != nil {
			log.Fatalf("gemini client: %v", err)
		}
		if err := runGeminiBackfill(ctx, client, storage, recipes); err != nil {
			log.Fatalf("backfill: %v", err)
		}
		return
	}

	client, err := claude.NewClient()
	if err != nil {
		log.Fatalf("claude client: %v", err)
	}
	if err := runBackfill(ctx, client, storage, recipes); err != nil {
		log.Fatalf("backfill: %v", err)
	}
}

// runBackfill builds + submits the batch, polls to completion, and merges
// successful results back into storage. Errored/expired custom_ids are logged
// and left un-backfilled for a future re-run.
func runBackfill(ctx context.Context, client *claude.Client, storage *sqlite.RecipeStorage, recipes []*domain.Recipe) error {
	byID := make(map[string]*domain.Recipe, len(recipes))
	requests := make([]claude.BatchRequest, 0, len(recipes))
	for _, r := range recipes {
		byID[r.ID] = r
		requests = append(requests, buildBatchRequest(r))
	}

	batch, err := client.CreateBatch(ctx, requests)
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}
	log.Printf("Submitted batch %s (%d requests). Re-run with this ID printed above to resume.", batch.ID, len(requests))

	// Bound the poll loop so an unexpected/empty/stuck status can't spin
	// forever. Batches are documented to finish within 24h; pollInterval is
	// 45s, so maxPolls covers ~24h with margin.
	const maxPolls = 2000
	for polls := 0; batch.ProcessingStatus != "ended"; polls++ {
		if polls >= maxPolls {
			return fmt.Errorf("batch %s did not end after %d polls (last status=%q)", batch.ID, polls, batch.ProcessingStatus)
		}
		log.Printf("Batch %s status=%s processing=%d succeeded=%d errored=%d",
			batch.ID, batch.ProcessingStatus,
			batch.RequestCounts.Processing, batch.RequestCounts.Succeeded, batch.RequestCounts.Errored)
		time.Sleep(pollInterval)
		batch, err = client.GetBatch(ctx, batch.ID)
		if err != nil {
			return fmt.Errorf("poll batch: %w", err)
		}
	}

	results, err := client.GetBatchResults(ctx, batch.ID)
	if err != nil {
		return fmt.Errorf("get batch results: %w", err)
	}

	var updated, skipped int
	for _, item := range results {
		recipe, ok := byID[item.CustomID]
		if !ok {
			log.Printf("result for unknown custom_id %s — skipping", item.CustomID)
			continue
		}
		if item.Result.Type != "succeeded" || item.Result.Message == nil {
			log.Printf("custom_id %s result=%s — left un-backfilled", item.CustomID, item.Result.Type)
			skipped++
			continue
		}

		meta, err := extractMetadata(item.Result.Message)
		if err != nil {
			log.Printf("custom_id %s: %v — left un-backfilled", item.CustomID, err)
			skipped++
			continue
		}

		mergeMetadata(recipe, meta)
		if err := storage.Update(recipe); err != nil {
			log.Printf("custom_id %s: update failed: %v", item.CustomID, err)
			skipped++
			continue
		}
		updated++
	}

	log.Printf("Backfill complete: %d updated, %d skipped (errored/expired/invalid).", updated, skipped)
	return nil
}

// geminiMetadataSchema mirrors metadataSchema but omits `additionalProperties`,
// which Gemini's responseSchema (an OpenAPI 3.0 subset) does not support.
var geminiMetadataSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"mainProtein": map[string]interface{}{"type": "string"},
		"dietClass": map[string]interface{}{
			"type": "string",
			"enum": []string{"omnivore", "vegetarian", "vegan", "pescetarian"},
		},
		"batchable":   map[string]interface{}{"type": "boolean"},
		"cookMinutes": map[string]interface{}{"type": "integer"},
		"ingredients": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":           map[string]interface{}{"type": "string"},
					"canonicalName":  map[string]interface{}{"type": "string"},
					"gramsEquiv":     map[string]interface{}{"type": "number"},
					"isPantryStaple": map[string]interface{}{"type": "boolean"},
					"isPerishable":   map[string]interface{}{"type": "boolean"},
				},
				"required": []string{"name", "canonicalName", "gramsEquiv", "isPantryStaple", "isPerishable"},
			},
		},
	},
	"required": []string{"mainProtein", "dietClass", "batchable", "cookMinutes", "ingredients"},
}

// runGeminiBackfill enriches recipes one synchronous call at a time via Gemini
// structured output. Reuses the same prompt, metadata struct, merge logic, and
// storage as the Claude path — only the LLM call differs. Failures on a single
// recipe are logged and skipped (left for a future re-run), never fatal.
func runGeminiBackfill(ctx context.Context, client *gemini.Client, storage *sqlite.RecipeStorage, recipes []*domain.Recipe) error {
	var updated, skipped int
	for _, r := range recipes {
		userContent := fmt.Sprintf(
			"Recipe: %s (serves %d)\n\nIngredients:\n%s\n\nInstructions:\n%s",
			r.Name, r.Servings, ingredientsText(r.Ingredients), instructionsText(r.Instructions),
		)

		raw, err := client.GenerateJSON(ctx, backfillSystemPrompt, userContent, geminiMetadataSchema)
		if err != nil {
			log.Printf("recipe %s: gemini call failed: %v — left un-backfilled", r.ID, err)
			skipped++
			continue
		}

		var meta recipeMetadata
		if err := json.Unmarshal(raw, &meta); err != nil {
			log.Printf("recipe %s: parse metadata: %v — left un-backfilled", r.ID, err)
			skipped++
			continue
		}

		mergeMetadata(r, &meta)
		if err := storage.Update(r); err != nil {
			log.Printf("recipe %s: update failed: %v", r.ID, err)
			skipped++
			continue
		}
		updated++
		if updated%25 == 0 {
			log.Printf("...%d/%d enriched", updated, len(recipes))
		}
	}
	log.Printf("Backfill complete: %d updated, %d skipped.", updated, skipped)
	return nil
}

// buildBatchRequest constructs a structured-output batch request for one recipe.
func buildBatchRequest(r *domain.Recipe) claude.BatchRequest {
	userContent := fmt.Sprintf(
		"Recipe: %s (serves %d)\n\nIngredients:\n%s\n\nInstructions:\n%s",
		r.Name, r.Servings, ingredientsText(r.Ingredients), instructionsText(r.Instructions),
	)
	return claude.BatchRequest{
		CustomID: r.ID,
		Params: claude.Request{
			Model:     claude.HaikuModel,
			MaxTokens: 2000,
			System:    backfillSystemPrompt,
			Messages: []claude.Message{
				{Role: "user", Content: userContent},
			},
			OutputConfig: &claude.OutputConfig{
				Format: &claude.OutputFormat{
					Type:   "json_schema",
					Schema: metadataSchema,
				},
			},
		},
	}
}

func ingredientsText(ings []domain.Ingredient) string {
	out := ""
	for _, ing := range ings {
		out += fmt.Sprintf("- %g %s %s\n", ing.Amount, ing.Unit, ing.Name)
	}
	return out
}

func instructionsText(steps []string) string {
	out := ""
	for i, s := range steps {
		out += fmt.Sprintf("%d. %s\n", i+1, s)
	}
	return out
}

// extractMetadata pulls the JSON metadata out of a batch result message.
func extractMetadata(msg *claude.Response) (*recipeMetadata, error) {
	// Use the first text block rather than assuming Content[0] is text — the
	// response may carry other block types ahead of the JSON payload.
	var text string
	for _, block := range msg.Content {
		if block.Type == "text" {
			text = block.Text
			break
		}
	}
	if text == "" {
		return nil, fmt.Errorf("no text content in message")
	}
	var meta recipeMetadata
	if err := json.Unmarshal([]byte(text), &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &meta, nil
}

// mergeMetadata applies parsed metadata onto a recipe in place: the 4 recipe
// columns plus per-ingredient fields (matched by ingredient name). An invalid
// dietClass is dropped (treated as unknown).
func mergeMetadata(r *domain.Recipe, meta *recipeMetadata) {
	r.MainProtein = meta.MainProtein
	if domain.IsValidDietClass(meta.DietClass) {
		r.DietClass = meta.DietClass
	}
	r.Batchable = meta.Batchable
	r.CookMinutes = meta.CookMinutes

	// The model is asked for one entry per ingredient in order, so when the
	// counts match we trust positional alignment (robust to the model echoing a
	// reworded/normalized name). Otherwise fall back to case-insensitive name
	// matching — better than dropping all per-ingredient metadata silently.
	if len(meta.Ingredients) == len(r.Ingredients) {
		for i := range r.Ingredients {
			applyIngredientMetadata(&r.Ingredients[i], meta.Ingredients[i])
		}
		return
	}

	byName := make(map[string]ingredientMetadata, len(meta.Ingredients))
	for _, im := range meta.Ingredients {
		byName[normalizeName(im.Name)] = im
	}
	for i := range r.Ingredients {
		im, ok := byName[normalizeName(r.Ingredients[i].Name)]
		if !ok {
			continue
		}
		applyIngredientMetadata(&r.Ingredients[i], im)
	}
}

func applyIngredientMetadata(ing *domain.Ingredient, im ingredientMetadata) {
	ing.CanonicalName = im.CanonicalName
	ing.GramsEquiv = im.GramsEquiv
	ing.IsPantryStaple = im.IsPantryStaple
	ing.IsPerishable = im.IsPerishable
}

func normalizeName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// selectRecipesNeedingMetadata returns recipes that still lack recipe-level
// metadata (idempotent re-run filter). limit <= 0 means no limit.
func selectRecipesNeedingMetadata(db *sql.DB, limit int) ([]*domain.Recipe, error) {
	// diet_class is the completeness sentinel: a successfully enriched recipe
	// always gets one of the enum values, so an empty diet_class means "not yet
	// backfilled". We deliberately do NOT key off main_protein — it is
	// legitimately "" for meatless recipes (vegan/dessert), which would
	// otherwise be re-submitted (and re-billed) on every run.
	query := `SELECT id, name, servings, ingredients, instructions
	          FROM recipes
	          WHERE diet_class = ''
	          ORDER BY id`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []*domain.Recipe
	for rows.Next() {
		var (
			r                          domain.Recipe
			ingredientsJSON, stepsJSON string
		)
		if err := rows.Scan(&r.ID, &r.Name, &r.Servings, &ingredientsJSON, &stepsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(ingredientsJSON), &r.Ingredients); err != nil {
			r.Ingredients = []domain.Ingredient{}
		}
		if err := json.Unmarshal([]byte(stepsJSON), &r.Instructions); err != nil {
			r.Instructions = []string{}
		}
		recipes = append(recipes, &r)
	}
	return recipes, rows.Err()
}
