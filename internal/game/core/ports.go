package game

import (
	"context"

	"github.com/google/uuid"
)

// GameRepository defines the interface for game persistence
type GameRepository interface {
	CreateGame(game *GameState) error
	LoadGameSnapshot(roomCode string) (*GameState, int, error)
	TrySaveGameSnapshot(game *GameState, version int) (bool, error)
	AddPlayerToGame(gameID uuid.UUID, player *Player) error
	SaveRound(gameID uuid.UUID, round *Round) error
	UpdateRound(round *Round) error
	SaveGameCompletion(gameID, winnerID uuid.UUID, finalScores map[uuid.UUID]int, usedCards []int) error
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
