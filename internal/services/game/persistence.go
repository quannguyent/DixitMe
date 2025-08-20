package game

import (
	"context"
	"encoding/json"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/models"
	"dixitme/internal/redis"

	"github.com/google/uuid"
)

// Database persistence methods

func (m *Manager) persistGame(game *GameState) error {
	db := database.GetDB()

	dbGame := &models.Game{
		ID:           game.ID,
		RoomCode:     game.RoomCode,
		Status:       game.Status,
		CurrentRound: game.RoundNumber,
		MaxRounds:    game.MaxRounds,
		CreatedAt:    game.CreatedAt,
	}

	return db.Create(dbGame).Error
}

func (m *Manager) persistPlayer(player *models.Player) error {
	db := database.GetDB()
	return db.Create(player).Error
}

func (m *Manager) persistGamePlayer(gameID uuid.UUID, player *Player) error {
	db := database.GetDB()

	gamePlayer := &models.GamePlayer{
		GameID:   gameID,
		PlayerID: player.ID,
		Position: player.Position,
		Score:    player.Score,
		IsActive: player.IsActive,
	}

	return db.Create(gamePlayer).Error
}

func (m *Manager) updateGameStatus(gameID uuid.UUID, status models.GameStatus) error {
	db := database.GetDB()
	return db.Model(&models.Game{}).Where("id = ?", gameID).Update("status", status).Error
}

func (m *Manager) persistRound(gameID uuid.UUID, round *Round) error {
	db := database.GetDB()

	dbRound := &models.GameRound{
		ID:            round.ID,
		GameID:        gameID,
		RoundNumber:   round.RoundNumber,
		StorytellerID: round.StorytellerID,
		Clue:          round.Clue,
		Status:        round.Status,
		CreatedAt:     round.CreatedAt,
	}

	return db.Create(dbRound).Error
}

func (m *Manager) updateRound(round *Round) error {
	db := database.GetDB()

	updates := map[string]interface{}{
		"clue":   round.Clue,
		"status": round.Status,
	}

	return db.Model(&models.GameRound{}).Where("id = ?", round.ID).Updates(updates).Error
}

func (m *Manager) persistCardSubmission(roundID, playerID uuid.UUID, cardID int) error {
	db := database.GetDB()

	submission := &models.CardSubmission{
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	return db.Create(submission).Error
}

func (m *Manager) persistVote(roundID, playerID uuid.UUID, cardID int) error {
	db := database.GetDB()

	vote := &models.Vote{
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	return db.Create(vote).Error
}

func (m *Manager) persistGameCompletion(gameID, winnerID uuid.UUID) error {
	db := database.GetDB()

	history := &models.GameHistory{
		GameID:   gameID,
		WinnerID: winnerID,
	}

	return db.Create(history).Error
}

// Chat persistence

func (m *Manager) persistChatMessage(chatMessage *models.ChatMessage) error {
	db := database.GetDB()
	return db.Create(chatMessage).Error
}

func (m *Manager) getChatMessages(gameID uuid.UUID, phase string, limit int) ([]models.ChatMessage, error) {
	db := database.GetDB()

	var messages []models.ChatMessage
	query := db.Where("game_id = ? AND is_visible = true", gameID)

	if phase != "all" && phase != "" {
		query = query.Where("phase = ?", phase)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse order to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Redis operations

func (m *Manager) storeGameInRedis(game *GameState) error {
	redisClient := redis.GetClient()
	if redisClient == nil {
		return nil // Redis not configured
	}

	// Create a simplified version for Redis storage
	gameData := map[string]interface{}{
		"id":           game.ID.String(),
		"room_code":    game.RoomCode,
		"status":       string(game.Status),
		"player_count": len(game.Players),
		"round_number": game.RoundNumber,
		"created_at":   game.CreatedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(gameData)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := "game:" + game.RoomCode

	// Store with expiration (24 hours)
	return redisClient.Set(ctx, key, data, 24*time.Hour).Err()
}
