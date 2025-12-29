package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewPostgresPool_ValidConfig(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := Config{
		DatabaseURL: "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable",
	}

	// Act
	pool, err := NewPostgresPool(cfg)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, pool)

	err = pool.Ping(ctx)
	assert.NoError(t, err, "db.Ping() should return nil for valid config")

	// Cleanup
	pool.Close()
}

func TestNewPostgresPool_InvalidConfig(t *testing.T) {
	cfg := Config{
		DatabaseURL: "postgres://invalid:invalid@localhost:54321/nonexistent?sslmode=disable",
	}

	// Act
	pool, err := NewPostgresPool(cfg)

	// Assert
	require.Error(t, err, "NewPostgresPool should return error for invalid config")
	assert.Nil(t, pool, "pool should be nil when connection fails")
	assert.Contains(t, err.Error(), "connect", "error should indicate connection failure")
}

func TestPostgresPool_Close(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := Config{
		DatabaseURL: "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable",
	}

	pool, err := NewPostgresPool(cfg)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// Act
	pool.Close()

	// Assert - After close, Ping should fail
	err = pool.Ping(ctx)
	assert.Error(t, err, "Ping should fail after pool is closed")
}
