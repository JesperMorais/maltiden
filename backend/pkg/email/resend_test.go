package email

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendSender_Send_ErrorStatusCodes(t *testing.T) {
	codes := []int{
		http.StatusBadRequest,          // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusUnprocessableEntity, // 422
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusServiceUnavailable,  // 503
	}

	for _, code := range codes {
		code := code
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, `{"message":"error"}`, code)
			}))
			defer srv.Close()

			s := newTestSender(srv, "key", "noreply@example.com")
			err := s.Send("user@example.com", "subj", "body")
			if err == nil {
				t.Fatalf("status %d: expected error, got nil", code)
			}
			want := fmt.Sprintf("%d", code)
			if !strings.Contains(err.Error(), want) {
				t.Errorf("status %d: error should mention %q, got: %v", code, want, err)
			}
		})
	}
}
