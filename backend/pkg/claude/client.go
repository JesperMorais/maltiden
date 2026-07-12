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
	DefaultModel   = "claude-sonnet-4-5-20250929"
	DefaultTimeout = 30 * time.Second
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
	baseURL    string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
}

type Response struct {
	ID         string         `json:"id"`
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
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
		model:   DefaultModel,
		baseURL: BaseURL,
	}, nil
}

// newTestClient builds a Client bypassing NewClient's env-var requirement,
// for injecting a custom RoundTripper/base URL in tests.
func newTestClient(baseURL string, rt http.RoundTripper) *Client {
	return &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Transport: rt, Timeout: DefaultTimeout},
		model:      DefaultModel,
		baseURL:    baseURL,
	}
}

func (c *Client) SendMessage(ctx context.Context, req Request) (*Response, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
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
