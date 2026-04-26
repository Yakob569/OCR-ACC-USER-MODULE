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
	"github.com/cashflow/auth-service/internal/adapters/ocrclient"
	"github.com/cashflow/auth-service/internal/adapters/repositories"
	"github.com/cashflow/auth-service/internal/adapters/storage"
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

	dbManager := repositories.NewDatabaseManager(ctx, cfg.DatabaseURL, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	defer func() {
		log.Println("Closing database connection...")
		dbManager.Close()
	}()

	if dbManager.Pool == nil {
		log.Fatal("database connection is required to start the service")
	}

	// 2. Wire Hexagonal Architecture
	authAdapter := auth.NewJWTAuthAdapter(cfg.JWTSecret)
	userRepo := repositories.NewUserRepository(dbManager.Pool)
	groupRepo := repositories.NewReceiptGroupRepository(dbManager.Pool)
	imageRepo := repositories.NewReceiptImageRepository(dbManager.Pool)
	extractionRepo := repositories.NewOCRExtractionRepository(dbManager.Pool)
	jobRepo := repositories.NewOCRJobRepository(dbManager.Pool)
	objectStorageSvc, err := storage.NewObjectStorageService(cfg.MinIO)
	if err != nil {
		log.Fatal(err)
	}
	ocrEngineSvc := ocrclient.NewOCREngineService(cfg.OCREngine)
	userSvc := services.NewUserService(userRepo, authAdapter)
	groupSvc := services.NewReceiptGroupService(groupRepo)
	uploadSvc := services.NewReceiptUploadService(groupRepo, imageRepo, jobRepo, objectStorageSvc, cfg.OCRGroupMaxFiles, cfg.OCRMaxFileSizeMB)
	querySvc := services.NewReceiptQueryService(groupRepo, imageRepo, extractionRepo)
	ocrJobSvc := services.NewOCRJobService(jobRepo, imageRepo, extractionRepo, groupRepo, objectStorageSvc, ocrEngineSvc, cfg.OCREngine.MaxConcurrency)
	userHandler := handlers.NewUserHandler(userSvc)
	groupHandler := handlers.NewGroupHandler(groupSvc, uploadSvc, querySvc)

	// 3. Initialize Server
	server := api.NewServer(cfg.Port, userHandler, groupHandler, authAdapter, dbManager.Pool)

	// 4. Start Server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	go ocrJobSvc.StartWorkers(ctx)

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
