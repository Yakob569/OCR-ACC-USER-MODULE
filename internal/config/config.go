package config

import (
	"errors"
	"log"
	"os"

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
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
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
	}

	if cfg.DatabaseURL == "" && cfg.DBPass == "" {
		log.Println("Warning: Neither DATABASE_URL nor DB_PASS is set, using default values")
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "change-me-at-all-costs" {
		return nil, errors.New("JWT_SECRET must be set to a non-default value")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
