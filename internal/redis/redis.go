package redis

import (
	"context"

	"dixitme/internal/logger"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Initialize(redisURL string) {
	log := logger.GetLogger()

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Error("Failed to parse Redis URL", "error", err, "url", redisURL)
		panic(err)
	}

	Client = redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	_, err = Client.Ping(ctx).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		panic(err)
	}

	log.Info("Redis connection established")
}

func GetClient() *redis.Client {
	return Client
}
