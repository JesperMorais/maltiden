package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newStubClient(rt roundTripFunc) *Client {
	return &Client{apiKey: "test-key", httpClient: &http.Client{Transport: rt}, model: FlashLiteModel}
}

func TestGenerateJSON_ParsesFirstTextPart(t *testing.T) {
	var gotURL, gotKey, gotBody string
	c := newStubClient(func(r *http.Request) (*http.Response, error) {
		gotURL = r.URL.String()
		gotKey = r.Header.Get("x-goog-api-key")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		resp := `{"candidates":[{"content":{"parts":[{"text":"{\"dietClass\":\"vegan\"}"}]},"finishReason":"STOP"}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(resp)), Header: make(http.Header)}, nil
	})

	out, err := c.GenerateJSON(context.Background(), "sys", "user", map[string]interface{}{"type": "object"})
	if err != nil {
		t.Fatalf("GenerateJSON: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(out, &parsed); err != nil || parsed["dietClass"] != "vegan" {
		t.Fatalf("unexpected output %q (%v)", out, err)
	}
	if !strings.Contains(gotURL, "gemini-2.5-flash-lite:generateContent") {
		t.Errorf("wrong endpoint: %s", gotURL)
	}
	if gotKey != "test-key" {
		t.Errorf("api key header not set, got %q", gotKey)
	}
	// systemInstruction + responseSchema must be present in the request body.
	if !strings.Contains(gotBody, "systemInstruction") || !strings.Contains(gotBody, "responseMimeType") {
		t.Errorf("request body missing expected fields: %s", gotBody)
	}
}

func TestGenerateJSON_APIError(t *testing.T) {
	c := newStubClient(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"error":"bad"}`)), Header: make(http.Header)}, nil
	})
	if _, err := c.GenerateJSON(context.Background(), "", "user", nil); err == nil {
		t.Fatal("expected error on 400")
	}
}

func TestGenerateJSON_NoCandidates(t *testing.T) {
	c := newStubClient(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"candidates":[]}`)), Header: make(http.Header)}, nil
	})
	if _, err := c.GenerateJSON(context.Background(), "", "user", nil); err == nil {
		t.Fatal("expected error on empty candidates")
	}
}
