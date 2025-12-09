package redis

import (
	"context"
	"encoding/json"
	"time"

	"dixitme/internal/game/core"

	"github.com/redis/go-redis/v9"
)

type GameCache struct {
	client *redis.Client
}

func NewGameCache(client *redis.Client) *GameCache {
	return &GameCache{client: client}
}

func (c *GameCache) SetGame(ctx context.Context, g *game.GameState) error {
	gameJSON, err := json.Marshal(g)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "game:"+g.RoomCode, gameJSON, time.Hour).Err()
}

func (c *GameCache) GetGame(ctx context.Context, roomCode string) (*game.GameState, error) {
	val, err := c.client.Get(ctx, "game:"+roomCode).Result()
	if err != nil {
		return nil, err
	}
	var g game.GameState
	if err := json.Unmarshal([]byte(val), &g); err != nil {
		return nil, err
	}
	return &g, nil
}
