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
	CreatedAt    time.Time             `json:"created_at"`
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

// GameMessage represents a WebSocket message
type GameMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	MessageTypePlayerJoined   MessageType = "player_joined"
	MessageTypePlayerLeft     MessageType = "player_left"
	MessageTypeGameStarted    MessageType = "game_started"
	MessageTypeRoundStarted   MessageType = "round_started"
	MessageTypeClueSubmitted  MessageType = "clue_submitted"
	MessageTypeCardSubmitted  MessageType = "card_submitted"
	MessageTypeVotingStarted  MessageType = "voting_started"
	MessageTypeVoteSubmitted  MessageType = "vote_submitted"
	MessageTypeRoundCompleted MessageType = "round_completed"
	MessageTypeGameCompleted  MessageType = "game_completed"
	MessageTypeError          MessageType = "error"
	MessageTypeGameState      MessageType = "game_state"
)

// WebSocket message payloads
type PlayerJoinedPayload struct {
	Player *Player `json:"player"`
}

type PlayerLeftPayload struct {
	PlayerID uuid.UUID `json:"player_id"`
}

type GameStartedPayload struct {
	GameState *GameState `json:"game_state"`
}

type RoundStartedPayload struct {
	Round *Round `json:"round"`
}

type ClueSubmittedPayload struct {
	Clue string `json:"clue"`
}

type CardSubmittedPayload struct {
	PlayerID uuid.UUID `json:"player_id"`
}

type VotingStartedPayload struct {
	RevealedCards []RevealedCard `json:"revealed_cards"`
}

type VoteSubmittedPayload struct {
	PlayerID uuid.UUID `json:"player_id"`
}

type RoundCompletedPayload struct {
	Scores        map[uuid.UUID]int `json:"scores"`
	RevealedCards []RevealedCard    `json:"revealed_cards"`
}

type GameCompletedPayload struct {
	FinalScores map[uuid.UUID]int `json:"final_scores"`
	Winner      uuid.UUID         `json:"winner"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type GameStatePayload struct {
	GameState *GameState `json:"game_state"`
}
