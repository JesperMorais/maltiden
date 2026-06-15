package claude

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc lets a test stub the HTTP transport without a live network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newStubClient(handler roundTripFunc) *Client {
	return &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Transport: handler},
		model:      DefaultModel,
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestCreateBatch_SendsRequestsAndParsesResponse(t *testing.T) {
	var captured struct {
		Requests []BatchRequest `json:"requests"`
	}
	client := newStubClient(func(r *http.Request) (*http.Response, error) {
		if r.Method != "POST" || r.URL.String() != BatchURL {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("missing/incorrect anthropic-version header")
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing x-api-key header")
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Fatalf("unmarshal captured request: %v", err)
		}
		return jsonResponse(200, `{"id":"batch_1","type":"message_batch","processing_status":"in_progress","request_counts":{"processing":1}}`), nil
	})

	resp, err := client.CreateBatch(context.Background(), []BatchRequest{
		{
			CustomID: "rec_1",
			Params: Request{
				Model:     HaikuModel,
				MaxTokens: 100,
				Messages:  []Message{{Role: "user", Content: "hi"}},
				OutputConfig: &OutputConfig{
					Format: &OutputFormat{Type: "json_schema", Schema: map[string]interface{}{"type": "object"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}
	if resp.ID != "batch_1" || resp.ProcessingStatus != "in_progress" {
		t.Errorf("unexpected batch response: %+v", resp)
	}
	if len(captured.Requests) != 1 || captured.Requests[0].CustomID != "rec_1" {
		t.Fatalf("request body not sent correctly: %+v", captured)
	}
	// Structured output must serialize under output_config.format.
	if captured.Requests[0].Params.OutputConfig == nil ||
		captured.Requests[0].Params.OutputConfig.Format.Type != "json_schema" {
		t.Errorf("output_config not serialized: %+v", captured.Requests[0].Params.OutputConfig)
	}
}

func TestGetBatch_ParsesStatus(t *testing.T) {
	client := newStubClient(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.String() != BatchURL+"/batch_1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		return jsonResponse(200, `{"id":"batch_1","processing_status":"ended","request_counts":{"succeeded":2,"errored":1}}`), nil
	})

	resp, err := client.GetBatch(context.Background(), "batch_1")
	if err != nil {
		t.Fatalf("GetBatch: %v", err)
	}
	if resp.ProcessingStatus != "ended" {
		t.Errorf("expected ended, got %q", resp.ProcessingStatus)
	}
	if resp.RequestCounts.Succeeded != 2 || resp.RequestCounts.Errored != 1 {
		t.Errorf("unexpected counts: %+v", resp.RequestCounts)
	}
}

func TestGetBatchResults_ParsesJSONL(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"custom_id":"rec_1","result":{"type":"succeeded","message":{"id":"msg_1","content":[{"type":"text","text":"{\"mainProtein\":\"kyckling\"}"}]}}}`,
		`{"custom_id":"rec_2","result":{"type":"errored","error":{"type":"invalid_request","message":"bad"}}}`,
		`{"custom_id":"rec_3","result":{"type":"expired"}}`,
		``, // trailing blank line should be skipped
	}, "\n")

	client := newStubClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != BatchURL+"/batch_1/results" {
			t.Errorf("unexpected results URL: %s", r.URL)
		}
		return jsonResponse(200, jsonl), nil
	})

	items, err := client.GetBatchResults(context.Background(), "batch_1")
	if err != nil {
		t.Fatalf("GetBatchResults: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 result items, got %d", len(items))
	}

	if items[0].CustomID != "rec_1" || items[0].Result.Type != "succeeded" || items[0].Result.Message == nil {
		t.Errorf("succeeded item malformed: %+v", items[0])
	}
	if items[0].Result.Message.Content[0].Text != `{"mainProtein":"kyckling"}` {
		t.Errorf("succeeded message text mismatch: %q", items[0].Result.Message.Content[0].Text)
	}
	if items[1].Result.Type != "errored" || items[1].Result.Error == nil || items[1].Result.Error.Type != "invalid_request" {
		t.Errorf("errored item malformed: %+v", items[1])
	}
	if items[2].Result.Type != "expired" {
		t.Errorf("expired item malformed: %+v", items[2])
	}
}

func TestParseBatchResults_DirectInvalidLine(t *testing.T) {
	_, err := parseBatchResults([]byte("not json\n"))
	if err == nil {
		t.Fatal("expected error on invalid JSONL line")
	}
}
