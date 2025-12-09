package game

import (
	"context"

	"dixitme/internal/game/domain"

	"github.com/google/uuid"
)

// GameRepository defines the interface for game persistence
type GameRepository interface {
	CreateGame(game *GameState) error
	GetGame(roomCode string) (*GameState, error)
	UpdateGameStatus(gameID uuid.UUID, status domain.GameStatus) error
	AddPlayerToGame(gameID uuid.UUID, player *Player) error
	SaveRound(gameID uuid.UUID, round *Round) error
	UpdateRound(round *Round) error
	SaveCardSubmission(roundID, playerID uuid.UUID, cardID int) error
	SaveVote(roundID, playerID uuid.UUID, cardID int) error
	SaveGameCompletion(gameID, winnerID uuid.UUID) error
	SaveChatMessage(message *ChatMessage) error
}

// GameCache defines the interface for game caching (Redis)
type GameCache interface {
	SetGame(ctx context.Context, game *GameState) error
	GetGame(ctx context.Context, roomCode string) (*GameState, error)
}

// Broadcaster defines the interface for sending messages to players
// Currently `room.go` handles this directly via websocket.Conn in Player struct.
// We can abstract this later.
