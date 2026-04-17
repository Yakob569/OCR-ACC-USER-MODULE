package main

import (
	"context"
	"log"

	"github.com/cashflow/auth-service/internal/adapters/api"
	"github.com/cashflow/auth-service/internal/adapters/handlers"
	"github.com/cashflow/auth-service/internal/adapters/repositories"
	"github.com/cashflow/auth-service/internal/config"
	"github.com/cashflow/auth-service/internal/core/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	// 1. Initialize DB Pool
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("✅ Successfully connected to PostgreSQL")

	// 2. Wire Hexagonal Architecture
	userRepo := repositories.NewUserRepository(dbPool)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	// 3. Initialize and Start Server
	server := api.NewServer(cfg.Port, userHandler, dbPool)
	
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
