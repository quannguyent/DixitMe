package game

import (
	"sync"
	"time"

	"dixitme/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// GameState represents the in-memory state of an active game
type GameState struct {
	ID           uuid.UUID             `json:"id"`
	RoomCode     string                `json:"room_code"`
	Players      map[uuid.UUID]*Player `json:"players"`
	CurrentRound *Round                `json:"current_round"`
	Status       models.GameStatus     `json:"status"`
	RoundNumber  int                   `json:"round_number"`
	MaxRounds    int                   `json:"max_rounds"`
	Deck         []int                 `json:"deck"`       // Remaining cards in deck
	UsedCards    []int                 `json:"used_cards"` // Cards that have been played
	CreatedAt    time.Time             `json:"created_at"`
	LastActivity time.Time             `json:"last_activity"`
	mu           sync.RWMutex          `json:"-"`
}

// Lock locks the game state for writing
func (gs *GameState) Lock() {
	gs.mu.Lock()
}

// Unlock unlocks the game state
func (gs *GameState) Unlock() {
	gs.mu.Unlock()
}

// UpdateActivity updates the last activity timestamp
func (gs *GameState) UpdateActivity() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.LastActivity = time.Now()
}

// IsInactive checks if the game has been inactive for the given duration
func (gs *GameState) IsInactive(duration time.Duration) bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return time.Since(gs.LastActivity) > duration
}

// Player represents an active player in the game
type Player struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Score       int             `json:"score"`
	Position    int             `json:"position"`
	Hand        []int           `json:"hand"` // Card IDs in player's hand
	Connection  *websocket.Conn `json:"-"`    // WebSocket connection
	IsConnected bool            `json:"is_connected"`
	IsActive    bool            `json:"is_active"`
	IsBot       bool            `json:"is_bot"`
	BotLevel    string          `json:"bot_level,omitempty"` // easy, medium, hard
}

// Round represents the current round state
type Round struct {
	ID              uuid.UUID                     `json:"id"`
	RoundNumber     int                           `json:"round_number"`
	StorytellerID   uuid.UUID                     `json:"storyteller_id"`
	Clue            string                        `json:"clue"`
	Status          models.RoundStatus            `json:"status"`
	StorytellerCard int                           `json:"storyteller_card,omitempty"`
	Submissions     map[uuid.UUID]*CardSubmission `json:"submissions"`
	Votes           map[uuid.UUID]*Vote           `json:"votes"`
	RevealedCards   []RevealedCard                `json:"revealed_cards,omitempty"`
	CreatedAt       time.Time                     `json:"created_at"`
}

// CardSubmission represents a submitted card for the current round
type CardSubmission struct {
	PlayerID uuid.UUID `json:"player_id"`
	CardID   int       `json:"card_id"`
}

// Vote represents a player's vote
type Vote struct {
	PlayerID uuid.UUID `json:"player_id"`
	CardID   int       `json:"card_id"`
}

// RevealedCard represents a card shown during voting phase
type RevealedCard struct {
	CardID    int       `json:"card_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	VoteCount int       `json:"vote_count"`
}
