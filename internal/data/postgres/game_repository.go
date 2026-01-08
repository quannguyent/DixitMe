package postgres

import (
	"dixitme/internal/data/models"
	"dixitme/internal/game/core"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) CreateGame(g *game.GameState) error {
	payload, err := json.Marshal(g)
	if err != nil {
		return err
	}
	dbGame := &model.Game{
		ID:            g.ID,
		RoomCode:      g.RoomCode,
		Status:        g.Status,
		StateSnapshot: payload,
		Version:       1,
		CurrentRound:  g.RoundNumber,
		MaxRounds:     g.MaxRounds,
		CreatedAt:     g.CreatedAt,
	}
	return r.db.Create(dbGame).Error
}

func (r *GameRepository) LoadGameSnapshot(roomCode string) (*game.GameState, int, error) {
	var dbGame model.Game
	if err := r.db.Where("room_code = ?", roomCode).Take(&dbGame).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, fmt.Errorf("game not found")
		}
		return nil, 0, err
	}
	if len(dbGame.StateSnapshot) == 0 {
		return nil, 0, errors.New("game snapshot is empty")
	}

	var state game.GameState
	if err := json.Unmarshal(dbGame.StateSnapshot, &state); err != nil {
		return nil, 0, err
	}
	if state.ID == uuid.Nil {
		state.ID = dbGame.ID
	}
	if state.RoomCode == "" {
		state.RoomCode = dbGame.RoomCode
	}
	state.Version = dbGame.Version
	return &state, dbGame.Version, nil
}

func (r *GameRepository) TrySaveGameSnapshot(g *game.GameState, version int) (bool, error) {
	payload, err := json.Marshal(g)
	if err != nil {
		return false, err
	}
	result := r.db.Model(&model.Game{}).
		Where("id = ? AND version = ?", g.ID, version).
		Updates(map[string]interface{}{
			"state_snapshot": payload,
			"status":         g.Status,
			"current_round":  g.RoundNumber,
			"max_rounds":     g.MaxRounds,
			"version":        gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (r *GameRepository) AddPlayerToGame(gameID uuid.UUID, player *game.Player) error {
	// First ensure player exists
	dbPlayer := &model.Player{
		ID:   player.ID,
		Name: player.Name,
	}
	// Use FirstOrCreate to avoid duplicates if player exists
	if err := r.db.FirstOrCreate(dbPlayer, model.Player{ID: player.ID}).Error; err != nil {
		return err
	}

	// Then create game player relationship
	gamePlayer := &model.GamePlayer{
		GameID:   gameID,
		PlayerID: player.ID,
		Score:    player.Score,
		Position: player.Position,
		IsActive: player.IsActive,
	}
	return r.db.Create(gamePlayer).Error
}

func (r *GameRepository) SaveRound(gameID uuid.UUID, round *game.Round) error {
	dbRound := &model.GameRound{
		ID:            round.ID,
		GameID:        gameID,
		RoundNumber:   round.RoundNumber,
		StorytellerID: round.StorytellerID,
		Status:        round.Status,
		CreatedAt:     round.CreatedAt,
	}
	return r.db.Create(dbRound).Error
}

func (r *GameRepository) UpdateRound(round *game.Round) error {
	return r.db.Model(&model.GameRound{}).Where("id = ?", round.ID).Updates(map[string]interface{}{
		"clue":             round.Clue,
		"status":           round.Status,
		"storyteller_card": round.StorytellerCard,
	}).Error
}

func (r *GameRepository) SaveGameCompletion(gameID, winnerID uuid.UUID, finalScores map[uuid.UUID]int, usedCards []int) error {
	scoresPayload, err := json.Marshal(finalScores)
	if err != nil {
		return err
	}
	usedCardsPayload, err := json.Marshal(usedCards)
	if err != nil {
		return err
	}

	// Create game history
	history := &model.GameHistory{
		GameID:      gameID,
		WinnerID:    winnerID,
		FinalScores: scoresPayload,
		UsedCards:   usedCardsPayload,
		// Duration and TotalRounds would be calculated here
	}
	return r.db.Create(history).Error
}

func (r *GameRepository) SaveChatMessage(msg *game.ChatMessage) error {
	dbMsg := &model.ChatMessage{
		ID:          msg.ID,
		GameID:      msg.GameID,
		PlayerID:    msg.PlayerID,
		Message:     msg.Message,
		MessageType: msg.MessageType,
		Phase:       msg.Phase,
		IsVisible:   msg.IsVisible,
		CreatedAt:   msg.CreatedAt,
	}
	return r.db.Create(dbMsg).Error
}
