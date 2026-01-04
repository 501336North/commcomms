package api

import (
	"encoding/json"
	"net/http"
)

// DecodeJSON decodes JSON request body into the target struct.
// Returns false if decoding fails (caller should handle error response).
func DecodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}

// RequireContentType checks that the request has the expected content type.
func RequireContentType(w http.ResponseWriter, r *http.Request, contentType string) bool {
	if r.Header.Get("Content-Type") != contentType {
		WriteError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be "+contentType)
		return false
	}
	return true
}
