package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

func TestSentrySmoke_PanicCapturedByTransport(t *testing.T) {
	rec := &sentry.MockTransport{}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:       "",
		Transport: rec,
	}); err != nil {
		t.Fatalf("sentry.Init: %v", err)
	}

	handler := sentryhttp.New(sentryhttp.Options{Repanic: false}).Handle(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("smoke")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/smoke", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	sentry.Flush(2 * time.Second)

	events := rec.Events()
	if len(events) == 0 {
		t.Fatal("expected at least one sentry event, got none")
	}

	ev := events[0]
	if len(ev.Exception) == 0 && ev.Message == "" {
		t.Errorf("event has neither Exception nor Message: %+v", ev)
	}
}
