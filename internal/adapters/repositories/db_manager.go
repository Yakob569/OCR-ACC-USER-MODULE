package repositories

import (
	"context"
	"log"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManager struct {
	Pool *pgxpool.Pool
}

func NewDatabaseManager(ctx context.Context, user, pass, host, port, dbname string) *DatabaseManager {
	// Parse a minimal valid DSN to get default config
	poolConfig, err := pgxpool.ParseConfig("postgres://localhost:5432")
	if err != nil {
		log.Fatalf("Unable to parse base config: %v", err)
	}

	// Manually set fields to avoid URL encoding issues with special characters in password
	poolConfig.ConnConfig.User = user
	poolConfig.ConnConfig.Password = pass
	poolConfig.ConnConfig.Host = host
	poolConfig.ConnConfig.Database = dbname

	p, err := strconv.ParseUint(port, 10, 16)
	if err == nil {
		poolConfig.ConnConfig.Port = uint16(p)
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
