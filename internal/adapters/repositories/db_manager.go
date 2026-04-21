package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		
		// Run migrations automatically
		runMigrations(ctx, pool)
	}

	return &DatabaseManager{
		Pool: pool,
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("Running database migrations...")

	// Look for migrations in current dir or /app/db/migrations
	migrationsDir := "db/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "/app/db/migrations"
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Printf("⚠️ Could not read migrations directory: %v", err)
		return
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, fileName := range sqlFiles {
		log.Printf("Executing migration: %s", fileName)
		filePath := filepath.Join(migrationsDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("⚠️ Could not read migration file %s: %v", fileName, err)
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Printf("⚠️ Could not begin transaction for %s: %v", fileName, err)
			continue
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			log.Printf("❌ Migration %s failed and was rolled back: %v", fileName, err)
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				log.Printf("❌ Could not commit migration %s: %v", fileName, err)
			} else {
				log.Printf("✅ Migration %s completed successfully", fileName)
			}
		}
	}
}

func (m *DatabaseManager) Close() {
	if m.Pool != nil {
		m.Pool.Close()
	}
}
