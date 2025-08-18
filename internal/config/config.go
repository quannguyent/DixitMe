package config

import (
	"log"
	"os"
	"strconv"

	"dixitme/internal/logger"
	"dixitme/internal/storage"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
	GinMode     string
	Logger      logger.Config
	MinIO       storage.MinIOConfig
	Auth        AuthConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't use the logger here since it's not initialized yet
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/dixitme?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:        getEnv("PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		Logger: logger.Config{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		MinIO: storage.MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("MINIO_BUCKET", "dixitme-cards"),
			UseSSL:          getBoolEnv("MINIO_USE_SSL", false),
			Region:          getEnv("MINIO_REGION", "us-east-1"),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
	}
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
