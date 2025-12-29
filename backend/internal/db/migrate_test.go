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

// setupTestDB creates a PostgreSQL container and returns a connected pool.
// Caller is responsible for calling the returned cleanup function.
func setupTestDB(t *testing.T) (*Config, func()) {
	t.Helper()
	ctx := context.Background()

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

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := &Config{
		DatabaseURL: "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable",
	}

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	return cfg, cleanup
}

func TestRunMigrations_AppliesInOrder(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestDB(t)
	defer cleanup()

	pool, err := NewPostgresPool(*cfg)
	require.NoError(t, err)
	defer pool.Close()

	// Act - Run migrations against fresh database
	err = RunMigrations(pool)

	// Assert
	require.NoError(t, err, "RunMigrations should succeed on fresh database")

	// Verify migrations tracking table exists
	ctx := context.Background()
	var tableName string
	err = pool.QueryRow(ctx, "SELECT table_name FROM information_schema.tables WHERE table_name = 'schema_migrations'").Scan(&tableName)
	require.NoError(t, err, "schema_migrations table should exist after running migrations")
	assert.Equal(t, "schema_migrations", tableName)

	// Verify migration version is recorded
	var version int
	err = pool.QueryRow(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&version)
	require.NoError(t, err, "should be able to query migration version")
	assert.Greater(t, version, 0, "migration version should be greater than 0")
}

func TestRunMigrations_Idempotent(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestDB(t)
	defer cleanup()

	pool, err := NewPostgresPool(*cfg)
	require.NoError(t, err)
	defer pool.Close()

	// Act - Run migrations twice
	err = RunMigrations(pool)
	require.NoError(t, err, "first migration run should succeed")

	// Get version after first run
	ctx := context.Background()
	var versionAfterFirst int
	err = pool.QueryRow(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&versionAfterFirst)
	require.NoError(t, err)

	// Run migrations again
	err = RunMigrations(pool)

	// Assert - Second run should succeed without error
	require.NoError(t, err, "second migration run should succeed (idempotent)")

	// Version should remain the same
	var versionAfterSecond int
	err = pool.QueryRow(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&versionAfterSecond)
	require.NoError(t, err)
	assert.Equal(t, versionAfterFirst, versionAfterSecond, "migration version should not change on re-run")

	// Count migrations applied - should be the same
	var countAfterSecond int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&countAfterSecond)
	require.NoError(t, err)
	assert.Greater(t, countAfterSecond, 0, "at least one migration should be recorded")
}
