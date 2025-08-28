package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dixitme/internal/logger"
	"dixitme/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// GamePersistenceService defines database and cache persistence operations
type GamePersistenceService interface {
	// Database operations with context support
	PersistGame(ctx context.Context, game *GameState) error
	PersistPlayer(ctx context.Context, player *models.Player) error
	PersistGamePlayer(ctx context.Context, gameID uuid.UUID, player *Player) error
	UpdateGameStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error
	PersistRound(ctx context.Context, gameID uuid.UUID, round *Round) error
	UpdateRound(ctx context.Context, round *Round) error
	PersistCardSubmission(ctx context.Context, roundID, playerID uuid.UUID, cardID int) error
	PersistVote(ctx context.Context, roundID, playerID uuid.UUID, cardID int) error
	PersistGameCompletion(ctx context.Context, gameID, winnerID uuid.UUID) error
	PersistChatMessage(ctx context.Context, chatMessage *models.ChatMessage) error
	GetChatMessages(ctx context.Context, gameID uuid.UUID, phase string, limit int) ([]models.ChatMessage, error)

	// Redis operations with context support
	StoreGameInRedis(ctx context.Context, game *GameState) error
	LoadGameFromRedis(ctx context.Context, roomCode string) (*GameState, error)
	DeleteGameFromRedis(ctx context.Context, roomCode string) error
}

// Database persistence methods with dependency injection, context support, and transactions

func (m *Manager) PersistGame(ctx context.Context, game *GameState) error {
	log := logger.GetLogger()

	dbGame := &models.Game{
		ID:           game.ID,
		RoomCode:     game.RoomCode,
		Status:       game.Status,
		CurrentRound: game.RoundNumber,
		MaxRounds:    game.MaxRounds,
		CreatedAt:    game.CreatedAt,
	}

	if err := m.db.WithContext(ctx).Create(dbGame).Error; err != nil {
		log.Error("Failed to persist game",
			"game_id", game.ID,
			"room_code", game.RoomCode,
			"error", err)
		return fmt.Errorf("failed to persist game %s: %w", game.RoomCode, err)
	}

	log.Debug("Game persisted successfully",
		"game_id", game.ID,
		"room_code", game.RoomCode)
	return nil
}

func (m *Manager) PersistPlayer(ctx context.Context, player *models.Player) error {
	log := logger.GetLogger()

	// Use FirstOrCreate to handle existing players
	var existingPlayer models.Player
	result := m.db.WithContext(ctx).Where("id = ?", player.ID).FirstOrCreate(&existingPlayer, player)

	if result.Error != nil {
		log.Error("Failed to persist player",
			"player_id", player.ID,
			"player_name", player.Name,
			"error", result.Error)
		return fmt.Errorf("failed to persist player %s: %w", player.Name, result.Error)
	}

	if result.RowsAffected > 0 {
		log.Debug("Player created successfully",
			"player_id", player.ID,
			"player_name", player.Name)
	} else {
		log.Debug("Player already exists, using existing record",
			"player_id", player.ID,
			"player_name", existingPlayer.Name)
	}
	return nil
}

func (m *Manager) PersistGamePlayer(ctx context.Context, gameID uuid.UUID, player *Player) error {
	log := logger.GetLogger()

	dbGamePlayer := &models.GamePlayer{
		ID:       uuid.New(), // Generate a new UUID for the GamePlayer
		GameID:   gameID,
		PlayerID: player.ID,
		Score:    player.Score,
		Position: player.Position,
		IsActive: player.IsActive,
	}

	if err := m.db.WithContext(ctx).Create(dbGamePlayer).Error; err != nil {
		log.Error("Failed to persist game player",
			"game_id", gameID,
			"player_id", player.ID,
			"error", err)
		return fmt.Errorf("failed to persist game player: %w", err)
	}

	log.Debug("Game player persisted successfully",
		"game_id", gameID,
		"player_id", player.ID)
	return nil
}

func (m *Manager) UpdateGameStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error {
	log := logger.GetLogger()

	result := m.db.WithContext(ctx).Model(&models.Game{}).
		Where("id = ?", gameID).
		Update("status", status)

	if result.Error != nil {
		log.Error("Failed to update game status",
			"game_id", gameID,
			"status", status,
			"error", result.Error)
		return fmt.Errorf("failed to update game status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Warn("No game found to update status",
			"game_id", gameID,
			"status", status)
		return fmt.Errorf("no game found with ID %s", gameID)
	}

	log.Debug("Game status updated successfully",
		"game_id", gameID,
		"status", status)
	return nil
}

func (m *Manager) PersistRound(ctx context.Context, gameID uuid.UUID, round *Round) error {
	log := logger.GetLogger()

	// Use transaction for complex round persistence
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbRound := &models.GameRound{
			ID:            round.ID,
			GameID:        gameID,
			RoundNumber:   round.RoundNumber,
			StorytellerID: round.StorytellerID,
			Clue:          round.Clue,
			Status:        round.Status,
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(dbRound).Error; err != nil {
			log.Error("Failed to persist round in transaction",
				"round_id", round.ID,
				"game_id", gameID,
				"error", err)
			return fmt.Errorf("failed to persist round: %w", err)
		}

		log.Debug("Round persisted successfully in transaction",
			"round_id", round.ID,
			"game_id", gameID,
			"round_number", round.RoundNumber)
		return nil
	})
}

func (m *Manager) UpdateRound(ctx context.Context, round *Round) error {
	log := logger.GetLogger()

	updates := map[string]interface{}{
		"clue":   round.Clue,
		"status": round.Status,
	}

	result := m.db.WithContext(ctx).Model(&models.GameRound{}).
		Where("id = ?", round.ID).
		Updates(updates)

	if result.Error != nil {
		log.Error("Failed to update round",
			"round_id", round.ID,
			"error", result.Error)
		return fmt.Errorf("failed to update round: %w", result.Error)
	}

	log.Debug("Round updated successfully",
		"round_id", round.ID,
		"clue", round.Clue,
		"status", round.Status)
	return nil
}

func (m *Manager) PersistCardSubmission(ctx context.Context, roundID, playerID uuid.UUID, cardID int) error {
	log := logger.GetLogger()

	submission := &models.CardSubmission{
		ID:       uuid.New(),
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	if err := m.db.WithContext(ctx).Create(submission).Error; err != nil {
		log.Error("Failed to persist card submission",
			"round_id", roundID,
			"player_id", playerID,
			"card_id", cardID,
			"error", err)
		return fmt.Errorf("failed to persist card submission: %w", err)
	}

	log.Debug("Card submission persisted successfully",
		"round_id", roundID,
		"player_id", playerID,
		"card_id", cardID)
	return nil
}

func (m *Manager) PersistVote(ctx context.Context, roundID, playerID uuid.UUID, cardID int) error {
	log := logger.GetLogger()

	vote := &models.Vote{
		ID:       uuid.New(),
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	if err := m.db.WithContext(ctx).Create(vote).Error; err != nil {
		log.Error("Failed to persist vote",
			"round_id", roundID,
			"player_id", playerID,
			"card_id", cardID,
			"error", err)
		return fmt.Errorf("failed to persist vote: %w", err)
	}

	log.Debug("Vote persisted successfully",
		"round_id", roundID,
		"player_id", playerID,
		"card_id", cardID)
	return nil
}

func (m *Manager) PersistGameCompletion(ctx context.Context, gameID, winnerID uuid.UUID) error {
	log := logger.GetLogger()

	// Use transaction for game completion operations
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update game status
		if err := tx.Model(&models.Game{}).
			Where("id = ?", gameID).
			Updates(map[string]interface{}{
				"status":     models.GameStatusCompleted,
				"updated_at": time.Now(),
			}).Error; err != nil {
			log.Error("Failed to update game completion status in transaction",
				"game_id", gameID,
				"error", err)
			return fmt.Errorf("failed to update game completion status: %w", err)
		}

		// Create game history record
		gameHistory := &models.GameHistory{
			ID:       uuid.New(),
			GameID:   gameID,
			WinnerID: winnerID,
			// Additional fields would be populated from game state
			CreatedAt: time.Now(),
		}

		if err := tx.Create(gameHistory).Error; err != nil {
			log.Error("Failed to persist game history in transaction",
				"game_id", gameID,
				"winner_id", winnerID,
				"error", err)
			return fmt.Errorf("failed to persist game history: %w", err)
		}

		log.Info("Game completion persisted successfully",
			"game_id", gameID,
			"winner_id", winnerID)
		return nil
	})
}

func (m *Manager) PersistChatMessage(ctx context.Context, chatMessage *models.ChatMessage) error {
	log := logger.GetLogger()

	if err := m.db.WithContext(ctx).Create(chatMessage).Error; err != nil {
		log.Error("Failed to persist chat message",
			"message_id", chatMessage.ID,
			"game_id", chatMessage.GameID,
			"player_id", chatMessage.PlayerID,
			"error", err)
		return fmt.Errorf("failed to persist chat message: %w", err)
	}

	log.Debug("Chat message persisted successfully",
		"message_id", chatMessage.ID,
		"game_id", chatMessage.GameID,
		"player_id", chatMessage.PlayerID)
	return nil
}

func (m *Manager) GetChatMessages(ctx context.Context, gameID uuid.UUID, phase string, limit int) ([]models.ChatMessage, error) {
	log := logger.GetLogger()

	var messages []models.ChatMessage
	query := m.db.WithContext(ctx).Where("game_id = ?", gameID)

	if phase != "" {
		query = query.Where("phase = ?", phase)
	}

	if err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error; err != nil {
		log.Error("Failed to get chat messages",
			"game_id", gameID,
			"phase", phase,
			"limit", limit,
			"error", err)
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	log.Debug("Chat messages retrieved successfully",
		"game_id", gameID,
		"phase", phase,
		"count", len(messages))
	return messages, nil
}

// Redis operations with improved implementation and structured logging

func (m *Manager) StoreGameInRedis(ctx context.Context, game *GameState) error {
	log := logger.GetLogger()

	if m.redisClient == nil {
		log.Debug("Redis client not configured, skipping game storage")
		return nil // Redis not configured, not an error
	}

	// Create a comprehensive game representation for Redis
	gameData := map[string]interface{}{
		"id":           game.ID.String(),
		"room_code":    game.RoomCode,
		"status":       string(game.Status),
		"player_count": len(game.Players),
		"round_number": game.RoundNumber,
		"max_rounds":   game.MaxRounds,
		"created_at":   game.CreatedAt.Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
	}

	// Add player information
	var players []map[string]interface{}
	for _, player := range game.Players {
		players = append(players, map[string]interface{}{
			"id":       player.ID.String(),
			"name":     player.Name,
			"score":    player.Score,
			"is_bot":   player.IsBot,
			"position": player.Position,
		})
	}
	gameData["players"] = players

	data, err := json.Marshal(gameData)
	if err != nil {
		log.Error("Failed to marshal game data for Redis",
			"game_id", game.ID,
			"room_code", game.RoomCode,
			"error", err)
		return fmt.Errorf("failed to marshal game data: %w", err)
	}

	key := "game:" + game.RoomCode
	expiration := 24 * time.Hour

	if err := m.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		log.Error("Failed to store game in Redis",
			"game_id", game.ID,
			"room_code", game.RoomCode,
			"key", key,
			"error", err)
		return fmt.Errorf("failed to store game in Redis: %w", err)
	}

	log.Debug("Game stored in Redis successfully",
		"game_id", game.ID,
		"room_code", game.RoomCode,
		"key", key,
		"expiration", expiration)
	return nil
}

func (m *Manager) LoadGameFromRedis(ctx context.Context, roomCode string) (*GameState, error) {
	log := logger.GetLogger()

	if m.redisClient == nil {
		log.Debug("Redis client not configured, cannot load game")
		return nil, nil // Redis not configured, not an error
	}

	key := "game:" + roomCode
	data, err := m.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Debug("Game not found in Redis cache",
				"room_code", roomCode,
				"key", key)
			return nil, nil // Not found, not an error
		}
		log.Error("Failed to load game from Redis",
			"room_code", roomCode,
			"key", key,
			"error", err)
		return nil, fmt.Errorf("failed to load game from Redis: %w", err)
	}

	var gameData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &gameData); err != nil {
		log.Error("Failed to unmarshal game data from Redis",
			"room_code", roomCode,
			"key", key,
			"error", err)
		return nil, fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	// TODO: Implement full GameState reconstruction from Redis data
	// This is a placeholder implementation that documents the partial caching approach
	log.Warn("LoadGameFromRedis partially implemented - only basic validation performed",
		"room_code", roomCode,
		"data_keys", len(gameData))

	// For now, return nil to indicate we should load from database
	// In a full implementation, you would reconstruct the complete GameState here
	return nil, nil
}

func (m *Manager) DeleteGameFromRedis(ctx context.Context, roomCode string) error {
	log := logger.GetLogger()

	if m.redisClient == nil {
		log.Debug("Redis client not configured, cannot delete game")
		return nil // Redis not configured, not an error
	}

	key := "game:" + roomCode
	result := m.redisClient.Del(ctx, key)

	if err := result.Err(); err != nil {
		log.Error("Failed to delete game from Redis",
			"room_code", roomCode,
			"key", key,
			"error", err)
		return fmt.Errorf("failed to delete game from Redis: %w", err)
	}

	deletedCount := result.Val()
	log.Debug("Game deletion attempt completed",
		"room_code", roomCode,
		"key", key,
		"deleted_count", deletedCount)

	return nil
}
