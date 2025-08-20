package models

import (
	"time"

	"github.com/google/uuid"
)

// RoundStatus represents the current phase of a round
type RoundStatus string

const (
	RoundStatusStorytelling RoundStatus = "storytelling"
	RoundStatusSubmitting   RoundStatus = "submitting"
	RoundStatusVoting       RoundStatus = "voting"
	RoundStatusScoring      RoundStatus = "scoring"
	RoundStatusCompleted    RoundStatus = "completed"
)

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
