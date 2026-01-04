package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

// RequestIDKey is the context key for request ID.
const RequestIDKey contextKey = "request_id"

// ErrorResponse represents an error response with request ID.
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId,omitempty"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// Add request ID to response headers
	if requestID := GetRequestID(r.Context()); requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response with request ID.
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")

	requestID := GetRequestID(r.Context())
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:     message,
		RequestID: requestID,
	})
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from load balancer, etc.)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
