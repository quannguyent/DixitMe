package game

import (
	"github.com/google/uuid"
)

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
	MessageTypeGameDeleted    MessageType = "game_deleted"
	MessageTypeError          MessageType = "error"
	MessageTypeGameState      MessageType = "game_state"
	MessageTypeChatMessage    MessageType = "chat_message"
	MessageTypeChatHistory    MessageType = "chat_history"
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

type GameDeletedPayload struct {
	RoomCode string `json:"room_code"`
	Message  string `json:"message"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type GameStatePayload struct {
	GameState *GameState `json:"game_state"`
}
