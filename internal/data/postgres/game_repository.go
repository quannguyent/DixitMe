package postgres

import (
	"dixitme/internal/data/models"
	"dixitme/internal/game/core"
	"dixitme/internal/game/domain"
	"errors"
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
	dbGame := &model.Game{
		ID:           g.ID,
		RoomCode:     g.RoomCode,
		Status:       g.Status,
		CurrentRound: g.RoundNumber,
		MaxRounds:    g.MaxRounds,
		CreatedAt:    g.CreatedAt,
	}
	return r.db.Create(dbGame).Error
}

func (r *GameRepository) GetGame(roomCode string) (*game.GameState, error) {
	// This is a complex mapping from DB back to Domain.
	// For now, Room logic often relies on in-memory state.
	// If the server restarts, we need to load from DB.
	// This implementation is a placeholder for full reconstruction.
	return nil, errors.New("loading game from DB not fully implemented")
}

func (r *GameRepository) UpdateGameStatus(gameID uuid.UUID, status domain.GameStatus) error {
	return r.db.Model(&model.Game{}).Where("id = ?", gameID).Update("status", status).Error
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

func (r *GameRepository) SaveCardSubmission(roundID, playerID uuid.UUID, cardID int) error {
	submission := &model.CardSubmission{
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}
	return r.db.Create(submission).Error
}

func (r *GameRepository) SaveVote(roundID, playerID uuid.UUID, cardID int) error {
	vote := &model.Vote{
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}
	return r.db.Create(vote).Error
}

func (r *GameRepository) SaveGameCompletion(gameID, winnerID uuid.UUID) error {
	// Update game status
	if err := r.db.Model(&model.Game{}).Where("id = ?", gameID).Update("status", domain.GameStatusCompleted).Error; err != nil {
		return err
	}

	// Create game history
	history := &model.GameHistory{
		GameID:   gameID,
		WinnerID: winnerID,
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
