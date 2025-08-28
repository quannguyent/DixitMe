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

// AreAllHumanPlayersAFK checks if all human players are AFK
func (gs *GameState) AreAllHumanPlayersAFK(afkDuration time.Duration) bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	humanPlayerCount := 0
	afkPlayerCount := 0

	for _, player := range gs.Players {
		if !player.IsBot {
			humanPlayerCount++
			if player.IsAFK(afkDuration) {
				afkPlayerCount++
			}
		}
	}

	// All human players are AFK if we have human players and all are AFK
	return humanPlayerCount > 0 && humanPlayerCount == afkPlayerCount
}

// GetActiveHumanPlayerCount returns the count of active (non-AFK) human players
func (gs *GameState) GetActiveHumanPlayerCount(afkDuration time.Duration) int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	activeCount := 0
	for _, player := range gs.Players {
		if !player.IsBot && !player.IsAFK(afkDuration) {
			activeCount++
		}
	}

	return activeCount
}

// Player represents an active player in the game
type Player struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Score         int             `json:"score"`
	Position      int             `json:"position"`
	Hand          []int           `json:"hand"` // Card IDs in player's hand
	Connection    *websocket.Conn `json:"-"`    // WebSocket connection
	IsConnected   bool            `json:"is_connected"`
	IsActive      bool            `json:"is_active"`
	IsBot         bool            `json:"is_bot"`
	BotLevel      string          `json:"bot_level,omitempty"`      // easy, medium, hard
	LastActivity  time.Time       `json:"last_activity"`            // Track when player was last active
	WasReplaced   bool            `json:"was_replaced"`             // Flag to indicate if this player was replaced by a bot
	ReplacementID *uuid.UUID      `json:"replacement_id,omitempty"` // ID of the bot that replaced this player
}

// UpdateActivity updates the player's last activity timestamp
func (p *Player) UpdateActivity() {
	p.LastActivity = time.Now()
}

// IsAFK checks if the player has been inactive for the given duration
// This includes disconnected players and players who left the game
func (p *Player) IsAFK(duration time.Duration) bool {
	if p.IsBot {
		return false // Bots are never AFK
	}

	// Player is AFK if:
	// 1. Not connected and inactive for the duration, OR
	// 2. Not active (left the game) and not replaced yet
	return (!p.IsConnected && time.Since(p.LastActivity) > duration) ||
		(!p.IsActive && !p.WasReplaced)
}

// IsDisconnected checks if the player is disconnected (not a bot and not connected)
func (p *Player) IsDisconnected() bool {
	return !p.IsBot && !p.IsConnected
}

// HasLeftGame checks if the player has left the game (inactive but not replaced)
func (p *Player) HasLeftGame() bool {
	return !p.IsBot && !p.IsActive && !p.WasReplaced
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
