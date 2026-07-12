package claude

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSendMessage_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL, http.DefaultTransport)
	resp, err := c.SendMessage(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})

	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
}

func TestSendMessage_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	c := newTestClient(server.URL, http.DefaultTransport)
	resp, err := c.SendMessage(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})

	if err == nil {
		t.Fatal("expected error for malformed JSON response")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
}

func TestSendMessage_TransportError(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})

	c := newTestClient("http://127.0.0.1:0", rt)
	resp, err := c.SendMessage(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})

	if err == nil {
		t.Fatal("expected error for transport failure")
	}
	if resp != nil {
		t.Errorf("expected nil response, got %+v", resp)
	}
}
