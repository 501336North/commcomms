package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var migrations = []struct {
	version int
	sql     string
}{
	{
		version: 1,
		sql: `
			CREATE TABLE IF NOT EXISTS users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email TEXT UNIQUE NOT NULL,
				handle TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				reputation INTEGER DEFAULT 0,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			);
			CREATE TABLE IF NOT EXISTS communities (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name TEXT UNIQUE NOT NULL,
				description TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW()
			);
		`,
	},
}

func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.version).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		// Wrap each migration in a transaction for atomicity
		if err := runMigrationInTransaction(ctx, pool, m.version, m.sql); err != nil {
			return err
		}
	}

	return nil
}

// runMigrationInTransaction executes a single migration within a transaction.
// If any part of the migration fails, the entire migration is rolled back.
func runMigrationInTransaction(ctx context.Context, pool *pgxpool.Pool, version int, sql string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
