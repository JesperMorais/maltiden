package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// BatchRequest is a single entry in a Message Batches submission. CustomID lets
// us correlate each result back to the recipe it was built for.
type BatchRequest struct {
	CustomID string  `json:"custom_id"`
	Params   Request `json:"params"`
}

// BatchResponse is the poll-status shape returned by create/get on a batch.
type BatchResponse struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	ProcessingStatus string            `json:"processing_status"` // in_progress | ended | canceling
	RequestCounts    BatchRequestCounts `json:"request_counts"`
	ResultsURL       string            `json:"results_url"`
}

// BatchRequestCounts breaks down how many requests are in each terminal state.
type BatchRequestCounts struct {
	Processing int `json:"processing"`
	Succeeded  int `json:"succeeded"`
	Errored    int `json:"errored"`
	Canceled   int `json:"canceled"`
	Expired    int `json:"expired"`
}

// BatchResultItem is one line of the JSONL results stream.
type BatchResultItem struct {
	CustomID string      `json:"custom_id"`
	Result   BatchResult `json:"result"`
}

// BatchResult is the per-request outcome. Type is one of succeeded|errored|expired.
// Message is populated only on success; Error only on an errored result.
type BatchResult struct {
	Type    string       `json:"type"`
	Message *Response    `json:"message,omitempty"`
	Error   *BatchError  `json:"error,omitempty"`
}

// BatchError carries the API error detail for an errored result line.
type BatchError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// CreateBatch submits a set of requests for asynchronous batch processing.
func (c *Client) CreateBatch(ctx context.Context, requests []BatchRequest) (*BatchResponse, error) {
	payload := struct {
		Requests []BatchRequest `json:"requests"`
	}{Requests: requests}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal batch request: %w", err)
	}

	respBody, err := c.do(ctx, "POST", BatchURL, body)
	if err != nil {
		return nil, err
	}

	var batch BatchResponse
	if err := json.Unmarshal(respBody, &batch); err != nil {
		return nil, fmt.Errorf("unmarshal batch response: %w", err)
	}
	return &batch, nil
}

// GetBatch fetches the current status of a batch by ID.
func (c *Client) GetBatch(ctx context.Context, id string) (*BatchResponse, error) {
	respBody, err := c.do(ctx, "GET", BatchURL+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	var batch BatchResponse
	if err := json.Unmarshal(respBody, &batch); err != nil {
		return nil, fmt.Errorf("unmarshal batch response: %w", err)
	}
	return &batch, nil
}

// GetBatchResults fetches a batch's results, which the API returns as JSONL
// (one JSON object per line), and parses them into BatchResultItem values.
func (c *Client) GetBatchResults(ctx context.Context, id string) ([]BatchResultItem, error) {
	respBody, err := c.do(ctx, "GET", BatchURL+"/"+id+"/results", nil)
	if err != nil {
		return nil, err
	}
	return parseBatchResults(respBody)
}

// parseBatchResults parses a JSONL results body line-by-line. Exposed for
// testing the wire-shape parsing without a live HTTP layer.
func parseBatchResults(body []byte) ([]BatchResultItem, error) {
	var items []BatchResultItem
	scanner := bufio.NewScanner(bytes.NewReader(body))
	// Result lines (with an embedded message) can be large; raise the cap.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item BatchResultItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("parse batch result line: %w", err)
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan batch results: %w", err)
	}
	return items, nil
}
