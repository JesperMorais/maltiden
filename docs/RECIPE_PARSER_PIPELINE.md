# Recipe Parser Pipeline - Technical Specification

This document outlines the technical implementation for an AI-powered recipe parsing pipeline that converts unstructured recipe text into structured data for the Måltiden database.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Phase 1: Dev Version](#phase-1-dev-version-internal-use)
4. [Phase 2: Customer-Facing Version](#phase-2-customer-facing-version)
5. [Security Considerations](#security-considerations)
6. [API Reference](#api-reference)
7. [Cost Estimation](#cost-estimation)

---

## Overview

### Problem Statement

Currently, adding recipes to Måltiden requires manual entry through the `POST /recipes` endpoint with fully structured data. This is time-consuming and error-prone.

### Solution

Use Claude API to parse unstructured recipe text (from websites, cookbooks, user input) into structured `Recipe` objects that match our domain model.

### Input Examples

```
Pasta Carbonara för 4 personer

Ingredienser:
- 400g spaghetti
- 200g guanciale eller bacon
- 4 äggulor
- 100g parmesan, riven
- Svartpeppar

Gör så här:
1. Koka pastan enligt förpackningen
2. Stek baconet knaprigt
3. Vispa ihop äggulor och parmesan
4. Blanda het pasta med bacon, ta från värmen
5. Rör ner äggblandningen, salta och peppra
```

### Output

```json
{
  "name": "Pasta Carbonara",
  "servings": 4,
  "emoji": "🍝",
  "tags": ["pasta", "italienskt", "snabb"],
  "ingredients": [
    { "name": "spaghetti", "amount": 400, "unit": "g" },
    { "name": "guanciale eller bacon", "amount": 200, "unit": "g" },
    { "name": "äggulor", "amount": 4, "unit": "st" },
    { "name": "parmesan, riven", "amount": 100, "unit": "g" },
    { "name": "svartpeppar", "amount": 0, "unit": "efter smak" }
  ],
  "instructions": [
    "Koka pastan enligt förpackningen",
    "Stek baconet knaprigt",
    "Vispa ihop äggulor och parmesan",
    "Blanda het pasta med bacon, ta från värmen",
    "Rör ner äggblandningen, salta och peppra"
  ]
}
```

### Smart-menu metadata (Phase 0)

The parser also emits additional fields consumed by the (future) smart-menu
generator. These are additive and ride alongside the fields above.

**Per-ingredient:**

| Field | Type | Description |
|-------|------|-------------|
| `canonicalName` | string | Base grocery-item name, Swedish, singular, no brand/prep words (e.g. `"gul lök"`) |
| `gramsEquiv` | number | Estimated total weight in grams for the parsed amount/unit |
| `isPantryStaple` | boolean | True for long-shelf-life basics (salt, socker, mjöl, olja, kryddor) |
| `isPerishable` | boolean | True for items that spoil within days (färskt kött, fisk, mejeri, färska grönsaker) |

**Per-recipe:**

| Field | Type | Description |
|-------|------|-------------|
| `mainProtein` | string | Dominant protein source in Swedish, or `""` if none |
| `dietClass` | string | One of `vegan`, `vegetarisk`, `pescetarian`, `allätare` |
| `batchable` | boolean | True if the dish keeps/reheats well for meal prep |
| `cookMinutes` | integer | Estimated total active + passive cooking time in minutes |

Ingredient fields live in the existing `ingredients` JSON column (no
migration needed). Recipe-level fields are persisted via migration `013`
(`main_protein`, `diet_class`, `batchable`, `cook_minutes` on `recipes`).
Backfilling existing recipes and using these fields in menu generation are
separate follow-up efforts — this phase only covers schema + parser output.

---

## Architecture

### System Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Recipe Parser Pipeline                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────┐    ┌──────────────┐    ┌─────────────┐    ┌────────────┐ │
│  │  Client  │───▶│   Handler    │───▶│   Parser    │───▶│   Claude   │ │
│  │          │    │  /parse      │    │   Service   │    │    API     │ │
│  └──────────┘    └──────────────┘    └─────────────┘    └────────────┘ │
│       │                 │                   │                  │        │
│       │                 ▼                   ▼                  │        │
│       │          ┌──────────────┐    ┌─────────────┐          │        │
│       │          │  Validation  │    │   Prompt    │          │        │
│       │          │  Middleware  │    │  Templates  │          │        │
│       │          └──────────────┘    └─────────────┘          │        │
│       │                                                        │        │
│       │    ┌───────────────────────────────────────────────────┘        │
│       │    │                                                             │
│       │    ▼                                                             │
│       │  ┌─────────────┐    ┌──────────────┐    ┌───────────────┐      │
│       │  │  Structured │───▶│   Preview    │───▶│    Recipe     │      │
│       │  │   Output    │    │   Response   │    │   Storage     │      │
│       │  └─────────────┘    └──────────────┘    └───────────────┘      │
│       │                            │                                    │
│       └────────────────────────────┘                                    │
│                    User confirms before save                            │
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Overview

| Component | Responsibility |
|-----------|----------------|
| **Handler** | HTTP endpoint, request validation, response formatting |
| **Parser Service** | Claude API integration, prompt management, response parsing |
| **Validation Middleware** | Input sanitization, rate limiting, auth checks |
| **Prompt Templates** | System prompts optimized for Swedish recipes |

---

## Phase 1: Dev Version (Internal Use)

The dev version is for internal team use only - simpler implementation without customer-facing safety measures.

### 1.1 New Files to Create

```
backend/
├── internal/
│   ├── api/handlers/
│   │   └── recipe_parser.go      # New handler
│   ├── services/
│   │   └── recipe_parser_service.go  # New service
│   └── domain/
│       └── recipe_parser.go      # New types (or extend recipe.go)
└── pkg/
    └── claude/
        └── client.go             # Claude API client wrapper
```

### 1.2 Domain Types

Add to `backend/internal/domain/recipe.go` or create new file:

```go
// ParseRecipeRequest - input for parsing
type ParseRecipeRequest struct {
    RawText string `json:"rawText"`
    Source  string `json:"source,omitempty"` // "manual", "url", "image"
}

// ParseRecipeResponse - parsed preview before save
type ParseRecipeResponse struct {
    Recipe     CreateRecipeRequest `json:"recipe"`
    Confidence float64             `json:"confidence"` // 0.0 - 1.0
    Warnings   []string            `json:"warnings,omitempty"`
    RawText    string              `json:"rawText"` // Echo back for reference
}
```

### 1.3 Claude API Client

Create `backend/pkg/claude/client.go`:

```go
package claude

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

const (
    BaseURL        = "https://api.anthropic.com/v1/messages"
    DefaultModel   = "claude-sonnet-4.5-20250929"
    DefaultTimeout = 30 * time.Second
)

type Client struct {
    apiKey     string
    httpClient *http.Client
    model      string
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Request struct {
    Model        string         `json:"model"`
    MaxTokens    int            `json:"max_tokens"`
    Messages     []Message      `json:"messages"`
    System       string         `json:"system,omitempty"`
    OutputConfig *OutputConfig  `json:"output_config,omitempty"`
}

type OutputConfig struct {
    Format *OutputFormat `json:"format,omitempty"`
}

type OutputFormat struct {
    Type   string      `json:"type"`
    Schema interface{} `json:"schema"`
}

type Response struct {
    ID           string        `json:"id"`
    Content      []ContentBlock `json:"content"`
    StopReason   string        `json:"stop_reason"`
    Usage        Usage         `json:"usage"`
}

type ContentBlock struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}

type Usage struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
}

func NewClient() (*Client, error) {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
    }
    return &Client{
        apiKey: apiKey,
        httpClient: &http.Client{
            Timeout: DefaultTimeout,
        },
        model: DefaultModel,
    }, nil
}

func (c *Client) SendMessage(ctx context.Context, req Request) (*Response, error) {
    if req.Model == "" {
        req.Model = c.model
    }

    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", BaseURL, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("x-api-key", c.apiKey)
    httpReq.Header.Set("anthropic-version", "2023-06-01")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
    }

    var response Response
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    return &response, nil
}
```

### 1.4 Recipe Parser Service

Create `backend/internal/services/recipe_parser_service.go`:

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "maltiden/internal/domain"
    "maltiden/pkg/claude"
    "time"
)

// JSON Schema for structured output
var recipeSchema = map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name":     map[string]interface{}{"type": "string"},
        "servings": map[string]interface{}{"type": "integer"},
        "emoji":    map[string]interface{}{"type": "string"},
        "tags": map[string]interface{}{
            "type":  "array",
            "items": map[string]interface{}{"type": "string"},
        },
        "ingredients": map[string]interface{}{
            "type": "array",
            "items": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "name":   map[string]interface{}{"type": "string"},
                    "amount": map[string]interface{}{"type": "number"},
                    "unit":   map[string]interface{}{"type": "string"},
                },
                "required":             []string{"name", "amount", "unit"},
                "additionalProperties": false,
            },
        },
        "instructions": map[string]interface{}{
            "type":  "array",
            "items": map[string]interface{}{"type": "string"},
        },
        "confidence": map[string]interface{}{"type": "number"},
        "warnings": map[string]interface{}{
            "type":  "array",
            "items": map[string]interface{}{"type": "string"},
        },
    },
    "required":             []string{"name", "servings", "ingredients", "instructions", "confidence"},
    "additionalProperties": false,
}

const systemPrompt = `You are a recipe parsing assistant for Måltiden, a Swedish meal planning app.

Your task is to extract structured recipe data from unstructured text input.

IMPORTANT RULES:
1. All output must be valid JSON matching the provided schema
2. Recipe names should be in Swedish if the input is Swedish
3. For "servings", extract the number of portions (default to 4 if not specified)
4. For "emoji", suggest ONE relevant food emoji that represents the dish
5. For "tags", suggest 2-5 relevant Swedish tags from: vardag, helg, snabb, vegetarisk, vegan, fisk, kyckling, kött, pasta, soppa, sallad, barn, fest, billigt, hälsosam
6. For "ingredients":
   - Parse amount as a number (use 0 for "efter smak" / "to taste")
   - Use standard Swedish units: g, kg, dl, l, msk, tsk, st, krm
   - Keep ingredient names in Swedish
7. For "instructions":
   - Keep as an array of strings, one step per item
   - Keep in Swedish
   - Clean up numbering (remove "1.", "2." etc.)
8. Set "confidence" (0.0-1.0) based on how well-structured the input was
9. Add "warnings" array for any issues (missing info, ambiguous amounts, etc.)

NEVER include anything except the JSON object in your response.`

type RecipeParserService struct {
    claudeClient *claude.Client
}

func NewRecipeParserService() (*RecipeParserService, error) {
    client, err := claude.NewClient()
    if err != nil {
        return nil, fmt.Errorf("failed to create Claude client: %w", err)
    }
    return &RecipeParserService{
        claudeClient: client,
    }, nil
}

type parsedRecipeResponse struct {
    Name         string             `json:"name"`
    Servings     int                `json:"servings"`
    Emoji        string             `json:"emoji"`
    Tags         []string           `json:"tags"`
    Ingredients  []domain.Ingredient `json:"ingredients"`
    Instructions []string           `json:"instructions"`
    Confidence   float64            `json:"confidence"`
    Warnings     []string           `json:"warnings"`
}

func (s *RecipeParserService) ParseRecipe(rawText string) (*domain.ParseRecipeResponse, error) {
    if rawText == "" {
        return nil, fmt.Errorf("raw text is required")
    }

    // Limit input size to prevent abuse
    if len(rawText) > 10000 {
        return nil, fmt.Errorf("input text too long (max 10000 characters)")
    }

    req := claude.Request{
        MaxTokens: 2000,
        System:    systemPrompt,
        Messages: []claude.Message{
            {
                Role:    "user",
                Content: fmt.Sprintf("Parse this recipe:\n\n%s", rawText),
            },
        },
        OutputConfig: &claude.OutputConfig{
            Format: &claude.OutputFormat{
                Type:   "json_schema",
                Schema: recipeSchema,
            },
        },
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := s.claudeClient.SendMessage(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("claude API error: %w", err)
    }

    if len(resp.Content) == 0 || resp.Content[0].Type != "text" {
        return nil, fmt.Errorf("unexpected response format from Claude")
    }

    var parsed parsedRecipeResponse
    if err := json.Unmarshal([]byte(resp.Content[0].Text), &parsed); err != nil {
        return nil, fmt.Errorf("failed to parse Claude response: %w", err)
    }

    // Build response
    return &domain.ParseRecipeResponse{
        Recipe: domain.CreateRecipeRequest{
            Name:         parsed.Name,
            Servings:     parsed.Servings,
            Emoji:        parsed.Emoji,
            Tags:         parsed.Tags,
            Ingredients:  parsed.Ingredients,
            Instructions: parsed.Instructions,
        },
        Confidence: parsed.Confidence,
        Warnings:   parsed.Warnings,
        RawText:    rawText,
    }, nil
}
```

### 1.5 Handler

Create `backend/internal/api/handlers/recipe_parser.go`:

```go
package handlers

import (
    "encoding/json"
    "maltiden/internal/domain"
    "maltiden/internal/services"
    "net/http"
)

type RecipeParserHandler struct {
    parserService *services.RecipeParserService
    recipeService *services.RecipeService
}

func NewRecipeParserHandler(
    parserService *services.RecipeParserService,
    recipeService *services.RecipeService,
) *RecipeParserHandler {
    return &RecipeParserHandler{
        parserService: parserService,
        recipeService: recipeService,
    }
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
    Error string `json:"error"`
}

// POST /recipes/parse - Parse raw text into structured recipe
func (h *RecipeParserHandler) ParseRecipe(w http.ResponseWriter, r *http.Request) {
    var req domain.ParseRecipeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request"})
        return
    }

    // Validate input at handler level
    if req.RawText == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "rawText is required"})
        return
    }
    if len(req.RawText) > 10000 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "input too long (max 10000 characters)"})
        return
    }

    result, err := h.parserService.ParseRecipe(req.RawText)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse recipe"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(result); err != nil {
        // Log error (in production, use proper logging)
        // log.Printf("Error encoding response: %v", err)
    }
}

// POST /recipes/parse-and-save - Parse and immediately save
func (h *RecipeParserHandler) ParseAndSave(w http.ResponseWriter, r *http.Request) {
    var req domain.ParseRecipeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request"})
        return
    }

    // Validate input at handler level
    if req.RawText == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "rawText is required"})
        return
    }
    if len(req.RawText) > 10000 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "input too long (max 10000 characters)"})
        return
    }

    parsed, err := h.parserService.ParseRecipe(req.RawText)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse recipe"})
        return
    }

    // Save the parsed recipe
    created, err := h.recipeService.Create(parsed.Recipe)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to save recipe"})
        return
    }

    response := map[string]interface{}{
        "id":         created.ID,
        "recipe":     parsed.Recipe,
        "confidence": parsed.Confidence,
        "warnings":   parsed.Warnings,
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        // Log error (in production, use proper logging)
        // log.Printf("Error encoding response: %v", err)
    }
}
```

### 1.6 Router Integration

Add to `backend/internal/api/router.go`:

```go
// In SetupRoutes or similar function:

parserService, err := services.NewRecipeParserService()
if err != nil {
    log.Fatalf("Failed to initialize recipe parser service: %v", err)
}
parserHandler := handlers.NewRecipeParserHandler(parserService, recipeService)

// Protected routes (require auth)
mux.Handle("POST /recipes/parse", authMiddleware(http.HandlerFunc(parserHandler.ParseRecipe)))
mux.Handle("POST /recipes/parse-and-save", authMiddleware(http.HandlerFunc(parserHandler.ParseAndSave)))
```

### 1.7 Environment Variable

The Claude API client requires an API key to authenticate with Anthropic's services.

**Obtaining an API Key:**
1. Sign up at [console.anthropic.com](https://console.anthropic.com/)
2. Navigate to API Keys section
3. Create a new API key
4. Copy the key (starts with `sk-ant-api03-`)

**Setting up the environment variable:**

For local development (`.env` file):
```bash
ANTHROPIC_API_KEY=sk-ant-api03-xxxxx
```

For production deployment:
- Use your hosting provider's environment variable management
- Or use a secrets manager (AWS Secrets Manager, Google Secret Manager, etc.)
- Never commit the API key to version control

**Security notes:**
- Keep separate API keys for dev/staging/production environments
- Rotate keys periodically
- Never expose in frontend code or logs
- The application will fail to start if the key is not set

### 1.8 Dev Version Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/recipes/parse` | Yes | Parse text, return preview |
| POST | `/recipes/parse-and-save` | Yes | Parse and save in one step |

---

## Phase 2: Customer-Facing Version

For production customer use, add these safety and operational features.

### 2.1 Security Measures

#### 2.1.1 Input Validation & Sanitization

```go
// Add to recipe_parser_service.go

import (
    "regexp"
    "strings"
    "unicode/utf8"
)

var (
    // Patterns that might indicate prompt injection attempts
    suspiciousPatterns = []*regexp.Regexp{
        regexp.MustCompile(`(?i)(ignore|disregard|forget).*(previous|above|instructions)`),
        regexp.MustCompile(`(?i)system\s*prompt`),
        regexp.MustCompile(`(?i)you\s+are\s+(now|a)`),
        regexp.MustCompile(`(?i)\[.*?(INST|SYSTEM).*?\]`),
    }

    // Maximum lengths
    MaxInputLength     = 10000
    MaxIngredientsCount = 50
    MaxInstructionsCount = 30
)

func (s *RecipeParserService) validateInput(rawText string) error {
    // Check length
    if len(rawText) > MaxInputLength {
        return fmt.Errorf("input too long (max %d characters)", MaxInputLength)
    }

    // Check for valid UTF-8
    if !utf8.ValidString(rawText) {
        return fmt.Errorf("invalid text encoding")
    }

    // Check for suspicious patterns (potential prompt injection)
    lowerText := strings.ToLower(rawText)
    for _, pattern := range suspiciousPatterns {
        if pattern.MatchString(lowerText) {
            // Log for security monitoring but don't reveal details to user
            // log.Printf("SECURITY: Potential prompt injection detected")
            return fmt.Errorf("invalid input format")
        }
    }

    return nil
}
```

#### 2.1.2 Output Validation

```go
// Validate the parsed output before returning to user
func (s *RecipeParserService) validateOutput(parsed *parsedRecipeResponse) error {
    // Validate required fields
    if parsed.Name == "" {
        return fmt.Errorf("parsed recipe missing name")
    }
    if parsed.Servings <= 0 || parsed.Servings > 100 {
        return fmt.Errorf("invalid servings count")
    }
    if len(parsed.Ingredients) == 0 {
        return fmt.Errorf("no ingredients parsed")
    }
    if len(parsed.Ingredients) > MaxIngredientsCount {
        return fmt.Errorf("too many ingredients")
    }
    if len(parsed.Instructions) == 0 {
        return fmt.Errorf("no instructions parsed")
    }
    if len(parsed.Instructions) > MaxInstructionsCount {
        return fmt.Errorf("too many instructions")
    }

    // Validate ingredient amounts (prevent negative/extreme values)
    for _, ing := range parsed.Ingredients {
        if ing.Amount < 0 || ing.Amount > 10000 {
            return fmt.Errorf("invalid ingredient amount: %s", ing.Name)
        }
    }

    return nil
}
```

### 2.2 Rate Limiting

#### 2.2.1 Rate Limiter Implementation

Create `backend/pkg/ratelimit/limiter.go`:

**Note for production:** The implementation below is suitable for development and small-scale deployments. For production at scale, consider using established libraries like `golang.org/x/time/rate` or distributed rate limiting with Redis to handle multiple server instances correctly.

```go
package ratelimit

import (
    "sync"
    "time"
)

type RateLimiter struct {
    mu       sync.Mutex
    requests map[string][]time.Time
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }

    // Cleanup old entries periodically
    go rl.cleanup()

    return rl
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    windowStart := now.Add(-rl.window)

    // Filter out old requests
    var recent []time.Time
    for _, t := range rl.requests[key] {
        if t.After(windowStart) {
            recent = append(recent, t)
        }
    }

    if len(recent) >= rl.limit {
        rl.requests[key] = recent
        return false
    }

    rl.requests[key] = append(recent, now)
    return true
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        windowStart := now.Add(-rl.window)

        for key, times := range rl.requests {
            var recent []time.Time
            for _, t := range times {
                if t.After(windowStart) {
                    recent = append(recent, t)
                }
            }
            if len(recent) == 0 {
                delete(rl.requests, key)
            } else {
                rl.requests[key] = recent
            }
        }
        rl.mu.Unlock()
    }
}
```

#### 2.2.2 Apply Rate Limiting

```go
// In handler or middleware

var parseRateLimiter = ratelimit.NewRateLimiter(
    10,              // 10 requests
    time.Minute,     // per minute
)

func (h *RecipeParserHandler) ParseRecipe(w http.ResponseWriter, r *http.Request) {
    // Get user ID from auth context
    userID := r.Context().Value("userID").(string)

    if !parseRateLimiter.Allow(userID) {
        w.Header().Set("Retry-After", "60")
        http.Error(w, `{"error": "rate_limit_exceeded"}`, http.StatusTooManyRequests)
        return
    }

    // ... rest of handler
}
```

### 2.3 Usage Tracking & Billing

#### 2.3.1 Usage Tracking Table

```sql
-- Add to migrations

CREATE TABLE IF NOT EXISTS api_usage (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    household_id TEXT,
    endpoint TEXT NOT NULL,
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    cost_usd REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_api_usage_user ON api_usage(user_id);
CREATE INDEX idx_api_usage_household ON api_usage(household_id);
CREATE INDEX idx_api_usage_created ON api_usage(created_at);
```

#### 2.3.2 Usage Tracking Service

```go
// Calculate and store usage
func (s *RecipeParserService) trackUsage(userID string, usage claude.Usage) {
    // Claude Sonnet 4.5 pricing (as of 2026)
    // Input: $3 per 1M tokens
    // Output: $15 per 1M tokens
    inputCost := float64(usage.InputTokens) * 3.0 / 1_000_000
    outputCost := float64(usage.OutputTokens) * 15.0 / 1_000_000
    totalCost := inputCost + outputCost

    // Store in database (implement storage)
    // s.usageStorage.Create(userID, "recipe_parse", usage, totalCost)
}
```

### 2.4 Quota Management

```go
// Check user's monthly quota before allowing parse
type QuotaConfig struct {
    FreeMonthlyParses    int // Free tier
    PremiumMonthlyParses int // Premium tier
}

var DefaultQuota = QuotaConfig{
    FreeMonthlyParses:    10,
    PremiumMonthlyParses: 100,
}

func (s *RecipeParserService) checkQuota(userID string) error {
    // Get user's usage this month
    // usage := s.usageStorage.GetMonthlyCount(userID, "recipe_parse")
    // quota := s.getUserQuota(userID)

    // if usage >= quota {
    //     return fmt.Errorf("monthly quota exceeded")
    // }
    return nil
}
```

### 2.5 Audit Logging

```go
// Log all parse attempts for security review
type AuditLog struct {
    Timestamp   time.Time
    UserID      string
    Action      string
    InputHash   string   // SHA256 of input (don't store raw for privacy)
    InputLength int
    Success     bool
    ErrorType   string
    IPAddress   string
}

func (s *RecipeParserService) audit(userID, action string, input string, err error) {
    // Implement audit logging to database or external service
}
```

### 2.6 Error Response Standardization

```go
// Consistent error responses for customer-facing API
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

var (
    ErrRateLimited    = APIError{"rate_limited", "Too many requests. Please wait before trying again."}
    ErrQuotaExceeded  = APIError{"quota_exceeded", "Monthly parsing limit reached. Upgrade for more."}
    ErrInvalidInput   = APIError{"invalid_input", "Could not understand the recipe format."}
    ErrParseFailed    = APIError{"parse_failed", "Failed to parse recipe. Try a different format."}
    ErrServiceError   = APIError{"service_error", "Service temporarily unavailable."}
)
```

---

## Security Considerations

### 5.1 OWASP Top 10 for LLMs

Based on [OWASP LLM Top 10](https://cheatsheetseries.owasp.org/cheatsheets/LLM_Prompt_Injection_Prevention_Cheat_Sheet.html):

| Risk | Mitigation |
|------|------------|
| **Prompt Injection** | Input validation, pattern detection, system prompt isolation |
| **Data Leakage** | Don't include sensitive data in prompts, validate outputs |
| **Denial of Service** | Rate limiting, input size limits, timeouts |
| **Excessive Agency** | Recipe parser has no tools/actions, output-only |

### 5.2 Defense in Depth

```
┌─────────────────────────────────────────────────────────────────┐
│                     Security Layers                              │
├─────────────────────────────────────────────────────────────────┤
│  Layer 1: Authentication                                         │
│  └── JWT validation, session management                          │
├─────────────────────────────────────────────────────────────────┤
│  Layer 2: Rate Limiting                                          │
│  └── Per-user, per-IP, global limits                            │
├─────────────────────────────────────────────────────────────────┤
│  Layer 3: Input Validation                                       │
│  └── Size limits, encoding checks, pattern detection            │
├─────────────────────────────────────────────────────────────────┤
│  Layer 4: Prompt Isolation                                       │
│  └── System prompt separate from user input                     │
├─────────────────────────────────────────────────────────────────┤
│  Layer 5: Output Validation                                      │
│  └── Schema enforcement, value bounds checking                  │
├─────────────────────────────────────────────────────────────────┤
│  Layer 6: Audit & Monitoring                                     │
│  └── Log all requests, alert on anomalies                       │
└─────────────────────────────────────────────────────────────────┘
```

### 5.3 API Key Security

- Store `ANTHROPIC_API_KEY` in environment variables or secret manager
- Never expose in frontend code or logs
- Rotate keys periodically
- Use separate keys for dev/staging/production

---

## API Reference

### POST /recipes/parse

Parse raw recipe text into structured format.

**Request:**
```json
{
  "rawText": "Pasta Carbonara för 4 personer\n\nIngredienser:\n- 400g spaghetti...",
  "source": "manual"
}
```

**Response (200):**
```json
{
  "recipe": {
    "name": "Pasta Carbonara",
    "servings": 4,
    "emoji": "🍝",
    "tags": ["pasta", "italienskt", "snabb"],
    "ingredients": [
      { "name": "spaghetti", "amount": 400, "unit": "g" }
    ],
    "instructions": [
      "Koka pastan enligt förpackningen"
    ]
  },
  "confidence": 0.95,
  "warnings": [],
  "rawText": "..."
}
```

**Errors:**
| Code | Description |
|------|-------------|
| 400 | Invalid request / input validation failed |
| 401 | Not authenticated |
| 429 | Rate limit exceeded |
| 500 | Parse failed / service error |

---

## Cost Estimation

### Claude API Pricing (Sonnet 4.5)

| Type | Cost |
|------|------|
| Input | $3 / 1M tokens |
| Output | $15 / 1M tokens |

### Typical Recipe Parse

| Component | Tokens | Cost |
|-----------|--------|------|
| System prompt | ~500 | $0.0015 |
| User input (recipe) | ~300 | $0.0009 |
| Output (structured) | ~400 | $0.006 |
| **Total per parse** | ~1200 | **~$0.008** |

### Monthly Cost Estimates

| Users | Parses/user/month | Total parses | Est. cost |
|-------|-------------------|--------------|-----------|
| 100 | 5 | 500 | $4 |
| 1,000 | 5 | 5,000 | $40 |
| 10,000 | 5 | 50,000 | $400 |

---

## Implementation Checklist

### Dev Version (Phase 1)
- [ ] Create Claude API client (`pkg/claude/client.go`)
- [ ] Create parser service (`services/recipe_parser_service.go`)
- [ ] Add domain types (`domain/recipe.go`)
- [ ] Create handler (`handlers/recipe_parser.go`)
- [ ] Add routes to router
- [ ] Add `ANTHROPIC_API_KEY` to environment
- [ ] Test with sample recipes

### Production Version (Phase 2)
- [ ] Add input validation with pattern detection
- [ ] Implement output validation
- [ ] Add rate limiting middleware
- [ ] Create usage tracking table
- [ ] Implement quota management
- [ ] Add audit logging
- [ ] Set up monitoring/alerting
- [ ] Add frontend UI for parsing
- [ ] Documentation for users

---

## References

- [Claude API Structured Outputs](https://platform.claude.com/docs/en/build-with-claude/structured-outputs)
- [OWASP LLM Prompt Injection Prevention](https://cheatsheetseries.owasp.org/cheatsheets/LLM_Prompt_Injection_Prevention_Cheat_Sheet.html)
- [LLM Security Risks 2026](https://sombrainc.com/blog/llm-security-risks-2026)
