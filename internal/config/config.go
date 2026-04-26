package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	DBUser             string
	DBPass             string
	DBHost             string
	DBPort             string
	DBName             string
	Port               string
	JWTSecret          string
	SupabaseURL        string
	SupabaseKey        string
	SupabaseProjectRef string
	MinIO              MinIOConfig
	OCREngine          OCREngineConfig
	OCRGroupMaxFiles   int
	OCRMaxFileSizeMB   int
}

type MinIOConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Endpoint        string
	CLAMAVURL       string
	UseSSL          bool
}

type OCREngineConfig struct {
	BaseURL        string
	TimeoutSeconds int
	MaxConcurrency int
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	ocrTimeoutSeconds, err := getEnvInt("OCR_ENGINE_TIMEOUT_SECONDS", 60)
	if err != nil {
		return nil, err
	}

	ocrMaxConcurrency, err := getEnvInt("OCR_ENGINE_MAX_CONCURRENCY", 4)
	if err != nil {
		return nil, err
	}

	ocrGroupMaxFiles, err := getEnvInt("OCR_GROUP_MAX_FILES", 30)
	if err != nil {
		return nil, err
	}

	ocrMaxFileSizeMB, err := getEnvInt("OCR_MAX_FILE_SIZE_MB", 10)
	if err != nil {
		return nil, err
	}

	minIOUseSSL, err := getEnvBool("MINIO_USE_SSL", true)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPass:             getEnv("DB_PASS", "postgres"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBName:             getEnv("DB_NAME", "postgres"),
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-at-all-costs"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseKey:        getEnv("SUPABASE_KEY", ""),
		SupabaseProjectRef: getEnv("SUPABASE_PROJECT_REF", ""),
		MinIO: MinIOConfig{
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY_ID", "DEg0h6SfVoMEYTxJCxbY"),
			SecretAccessKey: getEnv("MINIO_SECRET_ACCESS_KEY", "CYUQ1PrjTfxgG9kMU8oCoqgpOQ8SAeCcalOByFpv"),
			BucketName:      getEnv("MINIO_BUCKET_NAME", "dev"),
			Endpoint:        getEnv("MINIO_END_POINT", "minio-cli.addispay.et"),
			CLAMAVURL:       getEnv("MINIO_CLAMAV_URL", "ttps://filescan.addispay.et/api/v1/scan"),
			UseSSL:          minIOUseSSL,
		},
		OCREngine: OCREngineConfig{
			BaseURL:        getEnv("OCR_ENGINE_BASE_URL", "https://ocr-acc-module-3.onrender.com"),
			TimeoutSeconds: ocrTimeoutSeconds,
			MaxConcurrency: ocrMaxConcurrency,
		},
		OCRGroupMaxFiles: ocrGroupMaxFiles,
		OCRMaxFileSizeMB: ocrMaxFileSizeMB,
	}

	if cfg.DatabaseURL == "" && cfg.DBPass == "" {
		log.Println("Warning: Neither DATABASE_URL nor DB_PASS is set, using default values")
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "change-me-at-all-costs" {
		return nil, errors.New("JWT_SECRET must be set to a non-default value")
	}

	if cfg.OCREngine.TimeoutSeconds <= 0 {
		return nil, errors.New("OCR_ENGINE_TIMEOUT_SECONDS must be greater than zero")
	}

	if cfg.OCREngine.MaxConcurrency <= 0 {
		return nil, errors.New("OCR_ENGINE_MAX_CONCURRENCY must be greater than zero")
	}

	if cfg.OCRGroupMaxFiles <= 0 {
		return nil, errors.New("OCR_GROUP_MAX_FILES must be greater than zero")
	}

	if cfg.OCRMaxFileSizeMB <= 0 {
		return nil, errors.New("OCR_MAX_FILE_SIZE_MB must be greater than zero")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return value, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean: %w", key, err)
	}

	return value, nil
}
