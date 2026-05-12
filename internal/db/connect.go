package db

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	if os.Getenv("APP_ENV") == "production" ||
		os.Getenv("POOL_TUNED") == "true" ||
		os.Getenv("RAILWAY_ENVIRONMENT") == "production" {
		cfg.MaxConns = 30
		cfg.MinConns = 5
		cfg.MaxConnLifetime = time.Hour
		cfg.MaxConnIdleTime = 30 * time.Minute
		cfg.HealthCheckPeriod = time.Minute
	} else {
		cfg.MaxConns = 10
		cfg.MinConns = 0
		cfg.MaxConnIdleTime = time.Minute * 5
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
