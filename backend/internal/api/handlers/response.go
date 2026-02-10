package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// maxBodySize is the default maximum request body size (1 MB).
const maxBodySize int64 = 1 << 20

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a JSON error response. Uses json.Marshal to prevent injection
// (fixes VALID-01 -- no more string concatenation with err.Error()).
func WriteError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code})
}

// DecodeJSON reads and decodes a JSON request body with size limit and Content-Type check.
// Returns false if decoding failed (error already written to w).
// Addresses VALID-02 (body size limits) and VALID-07 (Content-Type enforcement).
func DecodeJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, dst interface{}) bool {
	// Check Content-Type — only enforce when header is present (VALID-07)
	ct := r.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		WriteError(w, http.StatusUnsupportedMediaType, "content_type_must_be_json")
		return false
	}
	// Limit body size (VALID-02)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large")
		} else {
			WriteError(w, http.StatusBadRequest, "invalid_request")
		}
		return false
	}
	return true
}

// ValidateID checks that a path/query parameter ID is non-empty and has a minimum length
// consistent with the prefixed UUID format used in this codebase (e.g., "rec_abc123...").
// Returns false if validation failed (error already written to w). Addresses VALID-04.
func ValidateID(w http.ResponseWriter, id, paramName string) bool {
	if id == "" {
		WriteError(w, http.StatusBadRequest, paramName+"_required")
		return false
	}
	// IDs are prefix + UUID (e.g., "rec_<uuid>"), minimum length is prefix + underscore + UUID
	if len(id) < 4 {
		WriteError(w, http.StatusBadRequest, "invalid_"+paramName)
		return false
	}
	return true
}
