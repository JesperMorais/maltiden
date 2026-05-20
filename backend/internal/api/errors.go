package api

import (
	"encoding/json"
	"net/http"
)

// WriteUnauthorized writes a 401 JSON response with {"error": code}.
func WriteUnauthorized(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": code})
}
