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

func NewDatabaseManager(ctx context.Context, databaseURL, user, pass, host, port, dbname string) *DatabaseManager {
	var connStr string
	if databaseURL != "" {
		connStr = databaseURL
	} else {
		// Use a connection string that explicitly disables SSL for local development
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Printf("⚠️ Warning: Unable to connect to database: %v", err)
		return &DatabaseManager{Pool: nil}
	}

	if err := pool.Ping(ctx); err != nil {
		log.Printf("⚠️ Warning: Database ping failed: %v", err)
	} else {
		log.Println("✅ Successfully connected to PostgreSQL")
	}

	return &DatabaseManager{
		Pool: pool,
	}
}

func (m *DatabaseManager) Close() {
	if m.Pool != nil {
		m.Pool.Close()
	}
}
