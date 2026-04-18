package repositories

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManager struct {
	Pool *pgxpool.Pool
}

func NewDatabaseManager(ctx context.Context, dbURL string) *DatabaseManager {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL")

	return &DatabaseManager{
		Pool: pool,
	}
}

func (m *DatabaseManager) Close() {
	if m.Pool != nil {
		m.Pool.Close()
	}
}
