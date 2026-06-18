// Command match-livsmedel is a standalone, human-run tool that matches a
// recipe's canonical ingredient names to Livsmedelsverket food-composition rows
// (livsmedelsnummer) OFFLINE, then computes and persists per-serving macros on
// the recipes table. The model ONLY classifies (picks a livsmedelsnummer from a
// provided shortlist); all arithmetic is pure Go (domain.ComputeRecipeMacros).
//
// It is NEVER wired into server startup or CI. It requires ANTHROPIC_API_KEY
// (claude, Haiku Batch) or GEMINI_API_KEY (gemini, Flash-Lite), an explicit
// --confirm flag, prints an estimated cost before submitting, and is idempotent
// (re-running only touches recipes still missing a match).
//
//	go run ./cmd/match-livsmedel --confirm
//	go run ./cmd/match-livsmedel --confirm --limit 10 --provider gemini
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
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
// Flash-Lite, ~5x cheaper than Haiku.
const approxGeminiCostPerRecipeUSD = 0.0003

// pollInterval is how long to wait between batch status polls.
const pollInterval = 45 * time.Second

// maxCandidatesPerName caps how many livsmedel options are offered per canonical
// name. The model picks the single best one (or 0) FROM these options.
const maxCandidatesPerName = 8

const matchSystemPrompt = `You match Swedish recipe ingredients to Livsmedelsverket food-composition entries for Måltiden, a meal planning app.

For each canonical ingredient name you are given a shortlist of candidate food entries (each with a "livsmedelsnummer" and "namn"). Pick the SINGLE best matching livsmedelsnummer FROM THE PROVIDED OPTIONS for that ingredient, or 0 if none of the options is a reasonable match.

Prefer plain, raw/basic forms over prepared dishes (e.g. "kyckling" → raw chicken meat, not a chicken casserole). Do NOT invent livsmedelsnummer values that are not in the options. Output one entry per canonical name.`

// matchSchema constrains the model output: one {canonicalName, livsmedelsnummer}
// per canonical name.
var matchSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"matches": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"canonicalName":    map[string]interface{}{"type": "string"},
					"livsmedelsnummer": map[string]interface{}{"type": "integer"},
				},
				"required":             []string{"canonicalName", "livsmedelsnummer"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []string{"matches"},
	"additionalProperties": false,
}

// geminiMatchSchema mirrors matchSchema but omits `additionalProperties`, which
// Gemini's responseSchema (an OpenAPI 3.0 subset) does not support.
var geminiMatchSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"matches": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"canonicalName":    map[string]interface{}{"type": "string"},
					"livsmedelsnummer": map[string]interface{}{"type": "integer"},
				},
				"required": []string{"canonicalName", "livsmedelsnummer"},
			},
		},
	},
	"required": []string{"matches"},
}

// matchResult is the structured payload the model returns per recipe.
type matchResult struct {
	Matches []struct {
		CanonicalName    string `json:"canonicalName"`
		Livsmedelsnummer int    `json:"livsmedelsnummer"`
	} `json:"matches"`
}

func main() {
	confirm := flag.Bool("confirm", false, "required: actually call the API and spend money")
	limit := flag.Int("limit", 0, "optional: process at most N recipes (0 = all needing a match)")
	provider := flag.String("provider", "claude", "LLM provider: claude (Haiku Batch) | gemini (Flash-Lite)")
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

	recipes, err := selectRecipesNeedingMatch(db, *limit)
	if err != nil {
		log.Fatalf("select recipes: %v", err)
	}

	if len(recipes) == 0 {
		log.Println("No recipes need livsmedel matching — nothing to do.")
		return
	}

	costPer := approxCostPerRecipeUSD
	label := "Haiku 4.5 Batch"
	if *provider == "gemini" {
		costPer = approxGeminiCostPerRecipeUSD
		label = "Gemini 2.5 Flash-Lite"
	}
	log.Printf("%d recipes need matching. Estimated cost: ~$%.4f USD (%s).", len(recipes), float64(len(recipes))*costPer, label)

	if !*confirm {
		log.Println("Dry run — pass --confirm to call the API and spend money.")
		return
	}

	recipeStorage := sqlite.NewRecipeStorage(db)
	livsmedelStorage := sqlite.NewLivsmedelStorage(db)
	ctx := context.Background()

	if *provider == "gemini" {
		client, err := gemini.NewClient(gemini.FlashLiteModel)
		if err != nil {
			log.Fatalf("gemini client: %v", err)
		}
		if err := runGeminiMatch(ctx, client, recipeStorage, livsmedelStorage, recipes); err != nil {
			log.Fatalf("match: %v", err)
		}
		return
	}

	client, err := claude.NewClient()
	if err != nil {
		log.Fatalf("claude client: %v", err)
	}
	if err := runBatchMatch(ctx, client, recipeStorage, livsmedelStorage, recipes); err != nil {
		log.Fatalf("match: %v", err)
	}
}

// distinctCanonicals returns the recipe's distinct, non-empty canonical names
// that are still unmatched (Livsmedelsnummer == 0), in first-seen order.
func distinctCanonicals(r *domain.Recipe) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, ing := range r.Ingredients {
		c := strings.ToLower(strings.TrimSpace(ing.CanonicalName))
		if c == "" || ing.Livsmedelsnummer != 0 {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

// buildPrompt builds the user prompt for a recipe: each unmatched canonical name
// with its shortlist of candidate livsmedel options. Returns the prompt text and
// the offered options keyed BY CANONICAL NAME (canonical → set of candidate
// numbers offered for THAT name), so the guardrail can reject a number the model
// assigns to a name it was never offered for — not just numbers invented outside
// the recipe entirely. When no canonical has any candidate, ok is false.
func buildPrompt(r *domain.Recipe, ls *sqlite.LivsmedelStorage) (prompt string, offered map[string]map[int]bool, ok bool) {
	offered = make(map[string]map[int]bool)
	var b strings.Builder
	fmt.Fprintf(&b, "Recipe: %s\n\nMatch each canonical ingredient name to the best livsmedelsnummer from its options (0 if none fit):\n\n", r.Name)
	any := false
	for _, c := range distinctCanonicals(r) {
		cands, err := ls.Search(c, maxCandidatesPerName)
		if err != nil {
			log.Printf("recipe %s: search %q failed: %v", r.ID, c, err)
			continue
		}
		fmt.Fprintf(&b, "- %q options:\n", c)
		set := make(map[int]bool, len(cands))
		offered[c] = set
		if len(cands) == 0 {
			b.WriteString("    (no candidates — answer 0)\n")
			any = true
			continue
		}
		for _, lv := range cands {
			set[lv.Livsmedelsnummer] = true
			fmt.Fprintf(&b, "    {livsmedelsnummer: %d, namn: %q}\n", lv.Livsmedelsnummer, lv.Namn)
		}
		any = true
	}
	return b.String(), offered, any
}

// applyMatch validates the model's matches against the offered options and the
// livsmedel table, sets Ingredient.Livsmedelsnummer for matching canonical
// names, computes macros, and persists. Returns true when the recipe was
// updated. Single-recipe failures are logged and return false, never fatal.
func applyMatch(r *domain.Recipe, res *matchResult, offered map[string]map[int]bool, ls *sqlite.LivsmedelStorage, storage *sqlite.RecipeStorage) bool {
	// Collect proposed numbers, dropping anything the model did not pick from the
	// shortlist offered FOR THAT canonical name (it must pick from its own
	// options, never invent or borrow another ingredient's number).
	proposed := make(map[string]int)
	numbers := make([]int, 0, len(res.Matches))
	for _, m := range res.Matches {
		if m.Livsmedelsnummer == 0 {
			continue
		}
		c := strings.ToLower(strings.TrimSpace(m.CanonicalName))
		if set := offered[c]; !set[m.Livsmedelsnummer] {
			log.Printf("recipe %s: dropping %d for %q — not in the options offered for that name", r.ID, m.Livsmedelsnummer, m.CanonicalName)
			continue
		}
		proposed[c] = m.Livsmedelsnummer
		numbers = append(numbers, m.Livsmedelsnummer)
	}
	if len(numbers) == 0 {
		return false
	}

	// GUARDRAIL: confirm the proposed numbers actually exist in the table; drop
	// any that don't (treat as unmatched).
	table, err := ls.GetByNumbers(dedupeInts(numbers))
	if err != nil {
		log.Printf("recipe %s: validate numbers failed: %v", r.ID, err)
		return false
	}

	changed := false
	for i := range r.Ingredients {
		ing := &r.Ingredients[i]
		if ing.Livsmedelsnummer != 0 {
			continue
		}
		c := strings.ToLower(strings.TrimSpace(ing.CanonicalName))
		num, ok := proposed[c]
		if !ok {
			continue
		}
		if _, exists := table[num]; !exists {
			continue // unknown number, drop
		}
		ing.Livsmedelsnummer = num
		changed = true
	}
	if !changed {
		return false
	}

	r.Macros = domain.ComputeRecipeMacros(*r, table)
	if err := storage.Update(r); err != nil {
		log.Printf("recipe %s: update failed: %v", r.ID, err)
		return false
	}
	return true
}

func dedupeInts(in []int) []int {
	seen := make(map[int]struct{}, len(in))
	out := in[:0:0]
	for _, n := range in {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// runGeminiMatch matches recipes one synchronous call at a time via Gemini.
func runGeminiMatch(ctx context.Context, client *gemini.Client, storage *sqlite.RecipeStorage, ls *sqlite.LivsmedelStorage, recipes []*domain.Recipe) error {
	var updated, skipped int
	for _, r := range recipes {
		prompt, offered, ok := buildPrompt(r, ls)
		if !ok {
			skipped++
			continue
		}
		raw, err := client.GenerateJSON(ctx, matchSystemPrompt, prompt, geminiMatchSchema)
		if err != nil {
			log.Printf("recipe %s: gemini call failed: %v — left unmatched", r.ID, err)
			skipped++
			continue
		}
		var res matchResult
		if err := json.Unmarshal(raw, &res); err != nil {
			log.Printf("recipe %s: parse matches: %v — left unmatched", r.ID, err)
			skipped++
			continue
		}
		if applyMatch(r, &res, offered, ls, storage) {
			updated++
			if updated%25 == 0 {
				log.Printf("...%d/%d matched", updated, len(recipes))
			}
		} else {
			skipped++
		}
	}
	log.Printf("Match complete: %d updated, %d skipped.", updated, skipped)
	return nil
}

// runBatchMatch builds + submits a Haiku batch, polls to completion, and applies
// validated matches. Errored/expired/invalid custom_ids are logged and left
// unmatched for a future re-run.
func runBatchMatch(ctx context.Context, client *claude.Client, storage *sqlite.RecipeStorage, ls *sqlite.LivsmedelStorage, recipes []*domain.Recipe) error {
	byID := make(map[string]*domain.Recipe, len(recipes))
	offeredByID := make(map[string]map[string]map[int]bool, len(recipes))
	requests := make([]claude.BatchRequest, 0, len(recipes))
	for _, r := range recipes {
		prompt, offered, ok := buildPrompt(r, ls)
		if !ok {
			continue
		}
		byID[r.ID] = r
		offeredByID[r.ID] = offered
		requests = append(requests, claude.BatchRequest{
			CustomID: r.ID,
			Params: claude.Request{
				Model:     claude.HaikuModel,
				MaxTokens: 2000,
				System:    matchSystemPrompt,
				Messages:  []claude.Message{{Role: "user", Content: prompt}},
				OutputConfig: &claude.OutputConfig{
					Format: &claude.OutputFormat{
						Type:   "json_schema",
						Schema: matchSchema,
					},
				},
			},
		})
	}
	if len(requests) == 0 {
		log.Println("No recipes had any canonical names to match — nothing to submit.")
		return nil
	}

	batch, err := client.CreateBatch(ctx, requests)
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}
	log.Printf("Submitted batch %s (%d requests).", batch.ID, len(requests))

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
			log.Printf("custom_id %s result=%s — left unmatched", item.CustomID, item.Result.Type)
			skipped++
			continue
		}
		res, err := extractMatch(item.Result.Message)
		if err != nil {
			log.Printf("custom_id %s: %v — left unmatched", item.CustomID, err)
			skipped++
			continue
		}
		if applyMatch(recipe, res, offeredByID[item.CustomID], ls, storage) {
			updated++
		} else {
			skipped++
		}
	}

	log.Printf("Match complete: %d updated, %d skipped (errored/expired/invalid/no-match).", updated, skipped)
	return nil
}

// extractMatch pulls the JSON match payload out of a batch result message.
func extractMatch(msg *claude.Response) (*matchResult, error) {
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
	var res matchResult
	if err := json.Unmarshal([]byte(text), &res); err != nil {
		return nil, fmt.Errorf("parse matches: %w", err)
	}
	return &res, nil
}

// selectRecipesNeedingMatch returns Phase-0-enriched recipes (diet_class != '')
// that still have at least one ingredient with a non-empty canonicalName and an
// unset livsmedelsnummer. The candidate set is loaded by the DB filter; the
// "needs a match" decision is made in Go over the ingredients JSON, so the
// re-run filter is idempotent (fully matched recipes are skipped). limit <= 0
// means no limit.
func selectRecipesNeedingMatch(db *sql.DB, limit int) ([]*domain.Recipe, error) {
	query := `SELECT id, name, servings, ingredients, instructions
	          FROM recipes
	          WHERE diet_class != ''
	          ORDER BY id`

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
		if !needsMatch(&r) {
			continue
		}
		recipes = append(recipes, &r)
		if limit > 0 && len(recipes) >= limit {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Stable order for reproducible runs.
	sort.Slice(recipes, func(i, j int) bool { return recipes[i].ID < recipes[j].ID })
	return recipes, nil
}

// needsMatch reports whether a recipe has any ingredient with a canonical name
// but no livsmedelsnummer yet (idempotent re-run filter).
func needsMatch(r *domain.Recipe) bool {
	for _, ing := range r.Ingredients {
		if strings.TrimSpace(ing.CanonicalName) != "" && ing.Livsmedelsnummer == 0 {
			return true
		}
	}
	return false
}
