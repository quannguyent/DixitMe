package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusCompleted  GameStatus = "completed"
	GameStatusAbandoned  GameStatus = "abandoned"
)

// Game represents a game session
type Game struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	RoomCode     string         `json:"room_code" gorm:"unique;not null"`
	Status       GameStatus     `json:"status" gorm:"default:'waiting'"`
	CurrentRound int            `json:"current_round" gorm:"default:1"`
	MaxRounds    int            `json:"max_rounds" gorm:"default:6"` // 3 players * 2 rounds each
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Players []GamePlayer `json:"players" gorm:"foreignKey:GameID"`
	Rounds  []GameRound  `json:"rounds" gorm:"foreignKey:GameID"`
}

// GamePlayer represents a player's participation in a specific game
type GamePlayer struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	GameID   uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	PlayerID uuid.UUID `json:"player_id" gorm:"type:uuid;not null"`
	Score    int       `json:"score" gorm:"default:0"`
	Position int       `json:"position"` // Turn order
	IsActive bool      `json:"is_active" gorm:"default:true"`

	// Relationships
	Game   Game   `json:"game" gorm:"foreignKey:GameID"`
	Player Player `json:"player" gorm:"foreignKey:PlayerID"`
}

// GameHistory stores completed games for statistics
type GameHistory struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	GameID      uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	WinnerID    uuid.UUID `json:"winner_id" gorm:"type:uuid"`
	TotalRounds int       `json:"total_rounds"`
	Duration    int       `json:"duration"` // Duration in minutes
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Game   Game   `json:"game" gorm:"foreignKey:GameID"`
	Winner Player `json:"winner" gorm:"foreignKey:WinnerID"`
}
