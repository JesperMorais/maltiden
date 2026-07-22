package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type redirectTransport struct {
	base string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = "http"
	req2.URL.Host = strings.TrimPrefix(t.base, "http://")
	return http.DefaultTransport.RoundTrip(req2)
}

func newTestSender(server *httptest.Server, apiKey, from string) *ResendSender {
	s := NewResendSender(apiKey, from)
	s.Client = &http.Client{
		Transport: &redirectTransport{base: server.URL},
	}
	return s
}

func TestNewResendSender_Smoke(t *testing.T) {
	s := NewResendSender("key-abc", "no-reply@example.com")
	if s == nil {
		t.Fatal("expected non-nil sender")
	}
	if s.APIKey != "key-abc" {
		t.Errorf("APIKey: got %q, want %q", s.APIKey, "key-abc")
	}
	if s.From != "no-reply@example.com" {
		t.Errorf("From: got %q, want %q", s.From, "no-reply@example.com")
	}
	if s.Client == nil {
		t.Fatal("expected non-nil http.Client")
	}
}

func TestResendSender_Send_HappyPath(t *testing.T) {
	var gotAuth string
	var gotBody resendPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"email_01"}`))
	}))
	defer srv.Close()

	s := newTestSender(srv, "re_secret", "noreply@example.com")
	if err := s.Send("user@example.com", "Hello", "World"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer re_secret" {
		t.Errorf("Authorization: got %q, want %q", gotAuth, "Bearer re_secret")
	}
	if gotBody.From != "noreply@example.com" {
		t.Errorf("From: got %q, want %q", gotBody.From, "noreply@example.com")
	}
	if len(gotBody.To) != 1 || gotBody.To[0] != "user@example.com" {
		t.Errorf("To: got %v, want [user@example.com]", gotBody.To)
	}
	if gotBody.Subject != "Hello" {
		t.Errorf("Subject: got %q, want %q", gotBody.Subject, "Hello")
	}
	if gotBody.Text != "World" {
		t.Errorf("Text: got %q, want %q", gotBody.Text, "World")
	}
}

func TestResendSender_Send_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	s := newTestSender(srv, "bad-key", "noreply@example.com")
	err := s.Send("user@example.com", "subj", "body")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention status 401, got: %v", err)
	}
}

func TestResendSender_Send_TransportError(t *testing.T) {
	s := NewResendSender("key", "from@example.com")
	s.Client = &http.Client{
		Transport: &redirectTransport{base: "http://127.0.0.1:1"},
	}
	err := s.Send("to@example.com", "subj", "body")
	if err == nil {
		t.Fatal("expected transport error")
	}
}
