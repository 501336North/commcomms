// Package main contains tests for the server binary.
package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainServerStarts verifies that the server binary compiles and starts
// without panic on valid configuration.
//
// RED PHASE: This test MUST FAIL because main.go does not exist yet.
func TestMainServerStarts(t *testing.T) {
	// GIVEN - A minimal server configuration
	cfg := &Config{
		Port: "8080",
		Host: "localhost",
	}

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// WHEN - We start the server in a goroutine
	serverErr := make(chan error, 1)
	serverReady := make(chan struct{})

	go func() {
		if err := RunServer(ctx, cfg, serverReady); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for server to be ready or timeout
	select {
	case <-serverReady:
		// Server is ready
	case err := <-serverErr:
		t.Fatalf("Server failed to start: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Server did not become ready in time")
	}

	// THEN - Server should be reachable via HTTP
	resp, err := http.Get("http://localhost:8080/health")
	require.NoError(t, err, "Server should be reachable")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")

	// Trigger graceful shutdown
	cancel()

	// Wait for server to shut down gracefully
	select {
	case <-serverErr:
		// Server shut down (may or may not send error)
	case <-time.After(2 * time.Second):
		// Timeout waiting for shutdown - acceptable for this test
	}
}
