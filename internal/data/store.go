package data

import (
	"fmt"

	"dixitme/internal/config"
	"dixitme/internal/data/minio"
	"dixitme/internal/data/postgres"
	redisstore "dixitme/internal/data/redis"
	"dixitme/internal/platform/logger"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Store bundles infra dependencies for pragmatic injection into features.
type Store struct {
	DB    *gorm.DB
	Redis *redis.Client
	Minio *minio.MinIOClient
}

// NewStore wires database, cache, and object storage using the provided config.
func NewStore(cfg *config.Config) (*Store, error) {
	log := logger.GetLogger()

	db := postgres.Initialize(cfg.DatabaseURL)
	redisClient := redisstore.Initialize(cfg.RedisURL)

	var minioClient *minio.MinIOClient
	if client, err := minio.Initialize(cfg.MinIO); err != nil {
		log.Warn("MinIO not available, continuing without object storage", "error", err)
	} else {
		minioClient = client
	}

	return &Store{
		DB:    db,
		Redis: redisClient,
		Minio: minioClient,
	}, nil
}

// MustStore is a helper for callers that want to panic on setup errors.
func MustStore(cfg *config.Config) *Store {
	store, err := NewStore(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to initialize store: %w", err))
	}
	return store
}
