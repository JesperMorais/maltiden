package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteUnauthorized(t *testing.T) {
	codes := []string{
		"unauthorized",
		"invalid_credentials",
		"invalid_token",
		"invalid_token_format",
		"token_revoked",
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteUnauthorized(w, code)

			res := w.Result()

			if res.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", res.StatusCode)
			}

			ct := res.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				t.Fatalf("expected Content-Type application/json, got %q", ct)
			}

			var body map[string]string
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}

			if got, ok := body["error"]; !ok || got != code {
				t.Fatalf("expected body[\"error\"] = %q, got %q", code, got)
			}

			if len(body) != 1 {
				t.Fatalf("expected exactly 1 key in body, got %d: %v", len(body), body)
			}
		})
	}
}
