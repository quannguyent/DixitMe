package http

import (
	"fmt"

	"dixitme/internal/data"
	"dixitme/internal/data/minio"
	"dixitme/internal/data/postgres"
	"dixitme/internal/game/core"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var store *data.Store
var manager *game.Manager

// InitStore injects the shared data store for handler package use.
func InitStore(s *data.Store) {
	if s == nil || s.DB == nil {
		panic(fmt.Errorf("handlers.InitStore: store is nil"))
	}
	store = s
}

// InitManager injects the game manager for handler package use.
func InitManager(m *game.Manager) {
	if m == nil {
		panic(fmt.Errorf("handlers.InitManager: manager is nil"))
	}
	manager = m
}

func getManager() (*game.Manager, error) {
	if manager == nil {
		return nil, fmt.Errorf("handlers.getManager: manager is nil")
	}
	return manager, nil
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
