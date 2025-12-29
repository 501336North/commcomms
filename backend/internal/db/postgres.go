package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DatabaseURL string
}

func NewPostgresPool(cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
