package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser      string
	DBPass      string
	DBHost      string
	DBPort      string
	DBName      string
	Port        string
	SupabaseURL        string
	SupabaseKey        string
	SupabaseProjectRef string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	cfg := &Config{
		DBUser:             getEnv("DB_USER", ""),
		DBPass:             getEnv("DB_PASS", ""),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBName:             getEnv("DB_NAME", "postgres"),
		Port:               getEnv("PORT", "8080"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseKey:        getEnv("SUPABASE_KEY", ""),
		SupabaseProjectRef: getEnv("SUPABASE_PROJECT_REF", ""),
	}

	if cfg.DBPass == "" {
		return nil, fmt.Errorf("DB_PASS is a required environment variable")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
