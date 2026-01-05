# CORS Middleware - Behövs för Frontend

Hej David!

## Problem
När frontend (Vue på `localhost:5173`) försöker anropa backend (Go på `localhost:8080`) blockeras requests av webbläsarens CORS-policy. Webbläsaren kräver att servern skickar specifika headers för att tillåta cross-origin requests.

## Lösning
Backend behöver en CORS middleware som lägger till rätt headers.

### Skapa fil: `pkg/middleware/cors.go`

```go
package middleware

import "net/http"

// CORS middleware for development
// Allows requests from frontend dev server
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow frontend origins
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://localhost:4173" || origin == "http://127.0.0.1:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

### Uppdatera: `internal/api/router.go`

Wrappa mux med CORS middleware i `NewRouter()`:

```go
func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// ... alla routes ...

	// Wrap with CORS middleware for frontend development
	return middleware.CORS(mux)
}
```

## Varför behövs detta?

1. **Preflight requests**: Webbläsaren skickar OPTIONS-request före POST/PUT för att kolla om servern tillåter det
2. **Authorization header**: Frontend skickar JWT token i `Authorization: Bearer <token>` header
3. **Content-Type**: Frontend skickar JSON med `Content-Type: application/json`

## För produktion

I produktion kan du antingen:
- Sätta rätt origin (t.ex. `https://maltiden.se`)
- Eller serva frontend och backend från samma domän (då behövs ingen CORS)

---

Jag hade lagt till detta men tar bort det så du kan implementera det själv på ditt sätt!

/Jesper (via Claude)
