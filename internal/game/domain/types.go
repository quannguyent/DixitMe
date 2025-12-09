package domain

import "time"

// PlayerType represents the type of player
type PlayerType string

const (
	PlayerTypeHuman PlayerType = "human"
	PlayerTypeBot   PlayerType = "bot"
)

// AuthType represents the authentication method
type AuthType string

const (
	AuthTypeGuest    AuthType = "guest"
	AuthTypePassword AuthType = "password"
	AuthTypeGoogle   AuthType = "google"
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusCompleted  GameStatus = "completed"
	GameStatusAbandoned  GameStatus = "abandoned"
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

// Card represents a game card (Domain)
type Card struct {
	ID          int       `json:"id"`
	ImageURL    string    `json:"image_url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Extension   string    `json:"extension"`
	IsActive    bool      `json:"is_active"`
	Tags        []CardTag `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Tag represents a categorization tag (Domain)
type Tag struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Category    string  `json:"category"`
	Color       string  `json:"color"`
	Weight      float64 `json:"weight"`
}

// CardTag represents the relationship between cards and tags (Domain)
type CardTag struct {
	CardID int     `json:"card_id"`
	TagID  int     `json:"tag_id"`
	Weight float64 `json:"weight"`
	Tag    Tag     `json:"tag"`
}
