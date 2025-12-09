package redis

import (
	"context"

	"dixitme/internal/platform/logger"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func GetClient() *redis.Client {
	return Client
}

// New creates a redis client without touching package globals.
func New(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opt), nil
}

// Initialize creates a redis client, tests the connection, and sets the package global for legacy callers.
func Initialize(redisURL string) *redis.Client {
	log := logger.GetLogger()

	client, err := New(redisURL)
	if err != nil {
		log.Error("Failed to parse Redis URL", "error", err, "url", redisURL)
		panic(err)
	}

	// Test connection
	ctx := context.Background()
	if _, err = client.Ping(ctx).Result(); err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		panic(err)
	}

	Client = client
	log.Info("Redis connection established")
	return Client
}
