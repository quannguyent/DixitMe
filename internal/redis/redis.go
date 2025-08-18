package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Initialize(redisURL string) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL:", err)
	}

	Client = redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	_, err = Client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Redis connection established")
}

func GetClient() *redis.Client {
	return Client
}
