package game

import (
	"time"

	"github.com/google/uuid"
)

// Chat message structures
type ChatMessagePayload struct {
	ID          uuid.UUID `json:"id"`
	PlayerID    uuid.UUID `json:"player_id"`
	PlayerName  string    `json:"player_name"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"` // chat, system, emote
	Phase       string    `json:"phase"`
	Timestamp   time.Time `json:"timestamp"`
}

type ChatHistoryPayload struct {
	Messages []ChatMessagePayload `json:"messages"`
	Phase    string               `json:"phase"`
}

// Incoming message for sending chat
type SendChatMessage struct {
	RoomCode string `json:"room_code"`
	Message  string `json:"message"`
	Type     string `json:"type,omitempty"` // Optional: chat, emote
}
