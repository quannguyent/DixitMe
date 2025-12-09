package http

import (
	"fmt"

	"dixitme/internal/data"
	"dixitme/internal/data/minio"
	"dixitme/internal/data/postgres"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var store *data.Store

// InitStore injects the shared data store for auth HTTP package use.
func InitStore(s *data.Store) {
	if s == nil || s.DB == nil {
		panic(fmt.Errorf("auth http InitStore: store is nil"))
	}
	store = s
}

func db() *gorm.DB {
	if store != nil && store.DB != nil {
		return store.DB
	}
	return postgres.GetDB()
}

func redisClient() *redis.Client {
	if store != nil && store.Redis != nil {
		return store.Redis
	}
	return nil
}

func minioClient() *minio.MinIOClient {
	if store != nil && store.Minio != nil {
		return store.Minio
	}
	return minio.GetClient()
}
