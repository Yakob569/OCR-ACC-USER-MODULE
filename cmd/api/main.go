package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cashflow/auth-service/internal/adapters/api"
	"github.com/cashflow/auth-service/internal/adapters/auth"
	"github.com/cashflow/auth-service/internal/adapters/handlers"
	"github.com/cashflow/auth-service/internal/adapters/repositories"
	"github.com/cashflow/auth-service/internal/config"
	"github.com/cashflow/auth-service/internal/core/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 1. Initialize Database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbManager := repositories.NewDatabaseManager(ctx, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	defer func() {
		log.Println("Closing database connection...")
		dbManager.Close()
	}()

	// 2. Wire Hexagonal Architecture
	authAdapter := auth.NewSupabaseAuthAdapter(cfg.SupabaseProjectRef, cfg.SupabaseURL, cfg.SupabaseKey)
	userRepo := repositories.NewUserRepository(dbManager.Pool)
	userSvc := services.NewUserService(userRepo, authAdapter)
	userHandler := handlers.NewUserHandler(userSvc)

	// 3. Initialize Server
	server := api.NewServer(cfg.Port, userHandler, dbManager.Pool)
	
	// 4. Start Server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 5. Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutdown signal received")

	// 6. Graceful Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("✅ Service stopped gracefully")
}
