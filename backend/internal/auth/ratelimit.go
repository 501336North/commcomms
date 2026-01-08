package auth

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter provides token bucket rate limiting per key (typically IP address).
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*tokenBucket
	rate     int           // tokens per interval
	interval time.Duration // refill interval
	capacity int           // max tokens
}

type tokenBucket struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a rate limiter with specified rate (requests per interval).
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		interval: interval,
		capacity: rate * 2, // Allow burst up to 2x rate
	}

	// Cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(bucket.lastCheck) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given key should be allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[key]

	if !exists {
		rl.buckets[key] = &tokenBucket{
			tokens:    rl.capacity - 1, // Use one token immediately
			lastCheck: now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastCheck)
	tokensToAdd := int(elapsed / rl.interval) * rl.rate
	bucket.tokens += tokensToAdd
	if bucket.tokens > rl.capacity {
		bucket.tokens = rl.capacity
	}
	bucket.lastCheck = now

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates HTTP middleware that applies rate limiting.
// keyFunc extracts the rate limit key from the request (typically client IP).
func RateLimitMiddleware(limiter *RateLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !limiter.Allow(key) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP extracts the client IP from the request.
// Checks X-Forwarded-For and X-Real-IP headers for proxied requests.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take only the first IP (original client), trim spaces
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// Common rate limiters for different endpoints
var (
	// LoginRateLimiter: 10 attempts per 15 minutes per IP
	LoginRateLimiter = NewRateLimiter(10, 15*time.Minute)

	// RegisterRateLimiter: 5 attempts per hour per IP
	RegisterRateLimiter = NewRateLimiter(5, time.Hour)

	// GeneralRateLimiter: 100 requests per minute per IP
	GeneralRateLimiter = NewRateLimiter(100, time.Minute)

	// MessageRateLimiter: 30 messages per minute per user
	MessageRateLimiter = NewRateLimiter(30, time.Minute)
)
