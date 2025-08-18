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

// Player represents a player in the system
type Player struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"not null"`
	Type      PlayerType     `json:"type" gorm:"default:'human'"`
	BotLevel  string         `json:"bot_level,omitempty"` // easy, medium, hard
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

// Card represents a game card with tags for categorization and bot AI
type Card struct {
	ID          int       `json:"id" gorm:"primary_key"`
	ImageURL    string    `json:"image_url" gorm:"not null"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Extension   string    `json:"extension" gorm:"default:'.jpg'"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Tags []CardTag `json:"tags" gorm:"many2many:card_tag_relations;"`
}

// Tag represents a categorization tag that can be applied to cards
type Tag struct {
	ID          int       `json:"id" gorm:"primary_key"`
	Name        string    `json:"name" gorm:"unique;not null;index"`
	Slug        string    `json:"slug" gorm:"unique;not null;index"`
	Description string    `json:"description"`
	Color       string    `json:"color" gorm:"default:'#3B82F6'"` // Hex color for UI
	Weight      float64   `json:"weight" gorm:"default:1.0"`      // For weighted random selection
	Category    string    `json:"category"`                       // Group tags by category
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Cards []Card `json:"cards" gorm:"many2many:card_tag_relations;"`
}

// CardTag represents the many-to-many relationship between cards and tags
type CardTag struct {
	CardID int     `json:"card_id" gorm:"primaryKey"`
	TagID  int     `json:"tag_id" gorm:"primaryKey"`
	Weight float64 `json:"weight" gorm:"default:1.0"` // How strongly this tag applies to this card

	// Relationships
	Card Card `json:"card" gorm:"foreignKey:CardID"`
	Tag  Tag  `json:"tag" gorm:"foreignKey:TagID"`
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
