// Package gemini is a minimal raw-HTTP client for Google's Gemini
// generateContent API. It is used ONLY by the offline backfill tool to enrich
// existing recipes — never in the menu-generation request path. The main app
// remains Claude-based (see pkg/claude); this exists so a one-off backfill can
// run on cheap Gemini credit.
package gemini

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
	// endpointTmpl is the generateContent REST endpoint; %s is the model.
	endpointTmpl = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

	// FlashLiteModel is the cheapest Gemini tier ($0.10/$0.40 per MTok).
	FlashLiteModel = "gemini-2.5-flash-lite"

	DefaultTimeout = 60 * time.Second
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewClient reads GEMINI_API_KEY (falling back to GOOGLE_API_KEY) from the
// environment. Returns an error when neither is set so the backfill can refuse
// to run rather than make unauthenticated calls.
func NewClient(model string) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY (or GOOGLE_API_KEY) environment variable not set")
	}
	if model == "" {
		model = FlashLiteModel
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		model:      model,
	}, nil
}

// --- request/response wire types ---

type part struct {
	Text string `json:"text"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type generationConfig struct {
	ResponseMIMEType string                 `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]interface{} `json:"responseSchema,omitempty"`
}

type generateRequest struct {
	SystemInstruction *content         `json:"systemInstruction,omitempty"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig,omitempty"`
}

type candidate struct {
	Content      content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

type generateResponse struct {
	Candidates []candidate `json:"candidates"`
}

// GenerateJSON sends a single prompt and returns the model's raw JSON text,
// constrained to schema via structured output. systemPrompt may be empty.
func (c *Client) GenerateJSON(ctx context.Context, systemPrompt, userContent string, schema map[string]interface{}) ([]byte, error) {
	req := generateRequest{
		Contents: []content{
			{Role: "user", Parts: []part{{Text: userContent}}},
		},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
		},
	}
	if systemPrompt != "" {
		req.SystemInstruction = &content{Parts: []part{{Text: systemPrompt}}}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	respBody, err := c.do(ctx, fmt.Sprintf(endpointTmpl, c.model), body)
	if err != nil {
		return nil, err
	}

	var resp generateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}
	for _, p := range resp.Candidates[0].Content.Parts {
		if p.Text != "" {
			return []byte(p.Text), nil
		}
	}
	return nil, fmt.Errorf("no text part in response (finishReason=%q)", resp.Candidates[0].FinishReason)
}

// do performs the authenticated POST and returns the raw body on a 2xx. It is
// the single HTTP seam; tests stub it via httpClient.Transport.
func (c *Client) do(ctx context.Context, url string, body []byte) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
