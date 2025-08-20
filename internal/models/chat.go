package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage represents a chat message in a game
type ChatMessage struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID      uuid.UUID `json:"game_id" gorm:"type:uuid;not null;index"`
	PlayerID    uuid.UUID `json:"player_id" gorm:"type:uuid;not null;index"`
	Message     string    `json:"message" gorm:"type:text;not null"`
	MessageType string    `json:"message_type" gorm:"type:varchar(20);default:'chat'"` // chat, system, emote
	Phase       string    `json:"phase" gorm:"type:varchar(20);not null"`              // lobby, storytelling, submitting, voting, scoring
	IsVisible   bool      `json:"is_visible" gorm:"default:true"`                      // For moderation
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Game   Game   `json:"game,omitempty" gorm:"foreignKey:GameID"`
	Player Player `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
}
