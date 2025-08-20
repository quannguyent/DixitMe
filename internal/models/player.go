package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlayerType represents the type of player
type PlayerType string

const (
	PlayerTypeHuman PlayerType = "human"
	PlayerTypeBot   PlayerType = "bot"
)

// Player represents a player in a game session (can be guest or registered user)
type Player struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    *uuid.UUID     `json:"user_id,omitempty" gorm:"type:uuid;index"` // NULL for guest players
	Name      string         `json:"name" gorm:"not null"`
	Type      PlayerType     `json:"type" gorm:"default:'human'"`
	AuthType  AuthType       `json:"auth_type" gorm:"default:'guest'"`
	BotLevel  string         `json:"bot_level,omitempty"`      // easy, medium, hard
	SessionID *uuid.UUID     `json:"-" gorm:"type:uuid;index"` // Links to session for guests
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
