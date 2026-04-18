package main

import (
	"context"
	"log"

	"github.com/cashflow/auth-service/internal/adapters/api"
	"github.com/cashflow/auth-service/internal/adapters/auth"
	"github.com/cashflow/auth-service/internal/adapters/handlers"
	"github.com/cashflow/auth-service/internal/adapters/repositories"
	"github.com/cashflow/auth-service/internal/config"
	"github.com/cashflow/auth-service/internal/core/services"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	// 1. Initialize Database
	ctx := context.Background()
	dbManager := repositories.NewDatabaseManager(ctx, cfg.DatabaseURL)
	defer dbManager.Close()

	// 2. Wire Hexagonal Architecture
	authAdapter := auth.NewSupabaseAuthAdapter(cfg.SupabaseURL, cfg.SupabaseKey)
	userRepo := repositories.NewUserRepository(dbManager.Pool)
	userSvc := services.NewUserService(userRepo, authAdapter)
	userHandler := handlers.NewUserHandler(userSvc)

	// 3. Initialize and Start Server
	server := api.NewServer(cfg.Port, userHandler, dbManager.Pool)
	
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
