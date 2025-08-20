package websocket

import (
	"encoding/json"
)

// ConnectionMessage represents incoming WebSocket messages
type ConnectionMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Message types from client
const (
	ClientMessageJoinGame       = "join_game"
	ClientMessageCreateGame     = "create_game"
	ClientMessageStartGame      = "start_game"
	ClientMessageSubmitClue     = "submit_clue"
	ClientMessageSubmitCard     = "submit_card"
	ClientMessageSubmitVote     = "submit_vote"
	ClientMessageLeaveGame      = "leave_game"
	ClientMessageSendChat       = "send_chat"
	ClientMessageGetChatHistory = "get_chat_history"
)

// Payload structures for client messages
type JoinGamePayload struct {
	RoomCode   string `json:"room_code"`
	PlayerName string `json:"player_name"`
}

type CreateGamePayload struct {
	RoomCode   string `json:"room_code"`
	PlayerName string `json:"player_name"`
}

type StartGamePayload struct {
	RoomCode string `json:"room_code"`
}

type SubmitCluePayload struct {
	RoomCode string `json:"room_code"`
	Clue     string `json:"clue"`
	CardID   int    `json:"card_id"`
}

type SubmitCardPayload struct {
	RoomCode string `json:"room_code"`
	CardID   int    `json:"card_id"`
}

type SubmitVotePayload struct {
	RoomCode string `json:"room_code"`
	CardID   int    `json:"card_id"`
}

type LeaveGamePayload struct {
	RoomCode string `json:"room_code"`
}

type SendChatPayload struct {
	RoomCode    string `json:"room_code"`
	Message     string `json:"message"`
	MessageType string `json:"message_type,omitempty"` // chat, emote
}

type GetChatHistoryPayload struct {
	RoomCode string `json:"room_code"`
	Phase    string `json:"phase,omitempty"` // lobby, voting, all
	Limit    int    `json:"limit,omitempty"` // default 50
}
