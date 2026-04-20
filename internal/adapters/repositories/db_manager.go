package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManager struct {
	Pool *pgxpool.Pool
}

func NewDatabaseManager(ctx context.Context, user, pass, host, port, dbname string) *DatabaseManager {
	// Use a connection string that explicitly disables SSL for local development
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
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
