package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/canary/commcomms/internal/auth"
)

type Config struct {
	Port      string
	Host      string
	JWTSecret string
}

func RunServer(ctx context.Context, cfg *Config, ready chan<- struct{}) error {
	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	// Create router with middleware chain
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware chain: rate limiting -> auth (for protected routes)
	// Public routes get rate limiting only
	publicHandler := auth.RateLimitMiddleware(auth.GeneralRateLimiter, auth.GetClientIP)(mux)

	// Create a separate mux for protected routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey)
		if userID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"user_id":"` + userID.(string) + `"}`))
	})

	// Apply auth middleware to protected routes
	protectedHandler := auth.AuthMiddleware(jwtService)(protectedMux)

	// Main handler that routes to public or protected handlers
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Routes that require authentication
		if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/api/v1" {
			protectedHandler.ServeHTTP(w, r)
			return
		}
		// Public routes
		publicHandler.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:      mainHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handler
	go func() {
		<-ctx.Done()

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Println("Shutting down server...")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	close(ready)
	return srv.ListenAndServe()
}

func main() {
	// Load configuration from environment
	cfg := &Config{
		Port:      getEnv("PORT", "8080"),
		Host:      getEnv("HOST", "localhost"),
		JWTSecret: getEnv("JWT_SECRET", ""),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Create context that listens for shutdown signals
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	ready := make(chan struct{})
	if err := RunServer(ctx, cfg, ready); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
