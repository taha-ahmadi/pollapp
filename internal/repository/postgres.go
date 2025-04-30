package repository

import (
	"context"
	"fmt"
	"time"

	"pollapp/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	Pool *pgxpool.Pool
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.Postgres) (*PostgresDB, error) {
	connString := cfg.DSN()

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// ToDo: read from env
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{Pool: pool}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
