package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Player represents a player in the system
type Player struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Game represents a game session
type Game struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
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

// GameStatus represents the current state of a game
type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusCompleted  GameStatus = "completed"
	GameStatusAbandoned  GameStatus = "abandoned"
)

// GamePlayer represents a player's participation in a specific game
type GamePlayer struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID   uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	PlayerID uuid.UUID `json:"player_id" gorm:"type:uuid;not null"`
	Score    int       `json:"score" gorm:"default:0"`
	Position int       `json:"position"` // Turn order
	IsActive bool      `json:"is_active" gorm:"default:true"`

	// Relationships
	Game   Game   `json:"game" gorm:"foreignKey:GameID"`
	Player Player `json:"player" gorm:"foreignKey:PlayerID"`
}

// GameRound represents a single round of the game
type GameRound struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID          uuid.UUID   `json:"game_id" gorm:"type:uuid;not null"`
	RoundNumber     int         `json:"round_number"`
	StorytellerID   uuid.UUID   `json:"storyteller_id" gorm:"type:uuid;not null"`
	Clue            string      `json:"clue"`
	Status          RoundStatus `json:"status" gorm:"default:'storytelling'"`
	StorytellerCard int         `json:"storyteller_card"` // Card ID chosen by storyteller
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	// Relationships
	Game        Game             `json:"game" gorm:"foreignKey:GameID"`
	Storyteller Player           `json:"storyteller" gorm:"foreignKey:StorytellerID"`
	Submissions []CardSubmission `json:"submissions" gorm:"foreignKey:RoundID"`
	Votes       []Vote           `json:"votes" gorm:"foreignKey:RoundID"`
}

// RoundStatus represents the current phase of a round
type RoundStatus string

const (
	RoundStatusStorytelling RoundStatus = "storytelling"
	RoundStatusSubmitting   RoundStatus = "submitting"
	RoundStatusVoting       RoundStatus = "voting"
	RoundStatusScoring      RoundStatus = "scoring"
	RoundStatusCompleted    RoundStatus = "completed"
)

// CardSubmission represents a card submitted by a player for a round
type CardSubmission struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RoundID  uuid.UUID `json:"round_id" gorm:"type:uuid;not null"`
	PlayerID uuid.UUID `json:"player_id" gorm:"type:uuid;not null"`
	CardID   int       `json:"card_id"` // ID of the card from the deck

	// Relationships
	Round  GameRound `json:"round" gorm:"foreignKey:RoundID"`
	Player Player    `json:"player" gorm:"foreignKey:PlayerID"`
}

// Vote represents a player's vote for which card they think belongs to the storyteller
type Vote struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RoundID  uuid.UUID `json:"round_id" gorm:"type:uuid;not null"`
	PlayerID uuid.UUID `json:"player_id" gorm:"type:uuid;not null"`
	CardID   int       `json:"card_id"` // ID of the card they voted for

	// Relationships
	Round  GameRound `json:"round" gorm:"foreignKey:RoundID"`
	Player Player    `json:"player" gorm:"foreignKey:PlayerID"`
}

// GameHistory stores completed games for statistics
type GameHistory struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID      uuid.UUID `json:"game_id" gorm:"type:uuid;not null"`
	WinnerID    uuid.UUID `json:"winner_id" gorm:"type:uuid"`
	TotalRounds int       `json:"total_rounds"`
	Duration    int       `json:"duration"` // Duration in minutes
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Game   Game   `json:"game" gorm:"foreignKey:GameID"`
	Winner Player `json:"winner" gorm:"foreignKey:WinnerID"`
}
