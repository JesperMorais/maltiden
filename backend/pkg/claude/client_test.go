package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// redirectTransport rewrites every request to the given base URL so the
// Client under test hits the httptest server instead of the real API.
type redirectTransport struct {
	base string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = "http"
	req2.URL.Host = strings.TrimPrefix(t.base, "http://")
	req2.URL.Path = ""
	return http.DefaultTransport.RoundTrip(req2)
}

func newTestClient(server *httptest.Server, apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Transport: &redirectTransport{base: server.URL},
		},
		model: DefaultModel,
	}
}

func TestNewClient_WithKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")
	c, err := NewClient()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_MissingKey(t *testing.T) {
	os.Unsetenv("ANTHROPIC_API_KEY")
	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is missing")
	}
}

func TestSendMessage_200(t *testing.T) {
	want := Response{
		ID: "msg_01",
		Content: []ContentBlock{
			{Type: "text", Text: "Hello"},
		},
		StopReason: "end_turn",
		Usage:      Usage{InputTokens: 10, OutputTokens: 5},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := newTestClient(srv, "test-key")
	resp, err := c.SendMessage(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != want.ID {
		t.Errorf("got ID %q, want %q", resp.ID, want.ID)
	}
	if len(resp.Content) != 1 || resp.Content[0].Text != "Hello" {
		t.Errorf("unexpected content: %v", resp.Content)
	}
}

func TestSendMessage_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"rate_limit"}`, http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := newTestClient(srv, "test-key")
	_, err := c.SendMessage(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestSendMessage_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := newTestClient(srv, "test-key")
	_, err := c.SendMessage(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for bad JSON response")
	}
}
