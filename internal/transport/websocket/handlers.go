package websocket

import (
	"encoding/json"
	"fmt"

	"dixitme/internal/models"
	"dixitme/internal/services/game"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// handleMessage routes incoming WebSocket messages to appropriate handlers
func handleMessage(conn *websocket.Conn, playerID uuid.UUID, msg ConnectionMessage) error {
	manager := game.GetManager()

	switch msg.Type {
	case ClientMessageCreateGame:
		return handleCreateGame(conn, playerID, msg, manager)
	case ClientMessageJoinGame:
		return handleJoinGame(conn, playerID, msg, manager)
	case ClientMessageStartGame:
		return handleStartGame(msg, manager, playerID)
	case ClientMessageSubmitClue:
		return handleSubmitClue(msg, manager, playerID)
	case ClientMessageSubmitCard:
		return handleSubmitCard(msg, manager, playerID)
	case ClientMessageSubmitVote:
		return handleSubmitVote(msg, manager, playerID)
	case ClientMessageLeaveGame:
		return handleLeaveGame(playerID, msg)
	case ClientMessageSendChat:
		return handleSendChat(msg, manager, playerID)
	case ClientMessageGetChatHistory:
		return handleGetChatHistory(conn, msg, manager)
	default:
		return sendError(conn, "Unknown message type: "+msg.Type)
	}
}

// handleCreateGame handles game creation requests
func handleCreateGame(conn *websocket.Conn, playerID uuid.UUID, msg ConnectionMessage, manager *game.Manager) error {
	var payload CreateGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	gameState, err := manager.CreateGame(payload.RoomCode, playerID, payload.PlayerName)
	if err != nil {
		return err
	}

	// Set the connection for this player
	if player, exists := gameState.Players[playerID]; exists {
		player.Connection = conn
	}

	// Send game state
	return conn.WriteJSON(game.GameMessage{
		Type:    game.MessageTypeGameState,
		Payload: game.GameStatePayload{GameState: gameState},
	})
}

// handleJoinGame handles game join requests
func handleJoinGame(conn *websocket.Conn, playerID uuid.UUID, msg ConnectionMessage, manager *game.Manager) error {
	var payload JoinGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	gameState, err := manager.JoinGame(payload.RoomCode, playerID, payload.PlayerName)
	if err != nil {
		return err
	}

	// Set the connection for this player
	if player, exists := gameState.Players[playerID]; exists {
		player.Connection = conn
	}

	// Send game state
	return conn.WriteJSON(game.GameMessage{
		Type:    game.MessageTypeGameState,
		Payload: game.GameStatePayload{GameState: gameState},
	})
}

// handleStartGame handles game start requests
func handleStartGame(msg ConnectionMessage, manager *game.Manager, playerID uuid.UUID) error {
	var payload StartGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	return manager.StartGame(payload.RoomCode, playerID)
}

// handleSubmitClue handles clue submission requests
func handleSubmitClue(msg ConnectionMessage, manager *game.Manager, playerID uuid.UUID) error {
	var payload SubmitCluePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	return manager.SubmitClue(payload.RoomCode, playerID, payload.Clue, payload.CardID)
}

// handleSubmitCard handles card submission requests
func handleSubmitCard(msg ConnectionMessage, manager *game.Manager, playerID uuid.UUID) error {
	var payload SubmitCardPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	return manager.SubmitCard(payload.RoomCode, playerID, payload.CardID)
}

// handleSubmitVote handles vote submission requests
func handleSubmitVote(msg ConnectionMessage, manager *game.Manager, playerID uuid.UUID) error {
	var payload SubmitVotePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	return manager.SubmitVote(payload.RoomCode, playerID, payload.CardID)
}

// handleLeaveGame handles leave game requests
func handleLeaveGame(playerID uuid.UUID, msg ConnectionMessage) error {
	var payload LeaveGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	return handlePlayerLeaveGame(playerID, payload.RoomCode)
}

// handleSendChat handles chat message requests
func handleSendChat(msg ConnectionMessage, manager *game.Manager, playerID uuid.UUID) error {
	var payload SendChatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	return manager.SendChatMessage(payload.RoomCode, playerID, payload.Message, payload.MessageType)
}

// handleGetChatHistory handles chat history requests
func handleGetChatHistory(conn *websocket.Conn, msg ConnectionMessage, manager *game.Manager) error {
	var payload GetChatHistoryPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	limit := payload.Limit
	if limit <= 0 {
		limit = 50
	}

	messages, err := manager.GetChatHistory(payload.RoomCode, payload.Phase, limit)
	if err != nil {
		return err
	}

	// Send chat history back to requesting client
	return conn.WriteJSON(game.GameMessage{
		Type: game.MessageTypeChatHistory,
		Payload: game.ChatHistoryPayload{
			Messages: messages,
			Phase:    payload.Phase,
		},
	})
}

// handlePlayerLeaveGame handles the logic when a player leaves a game
func handlePlayerLeaveGame(playerID uuid.UUID, roomCode string) error {
	manager := game.GetManager()
	gameState := manager.GetGame(roomCode)

	if gameState == nil {
		return nil // Game doesn't exist, nothing to do
	}

	gameState.Lock()
	defer gameState.Unlock()

	// Remove player from game
	if player, exists := gameState.Players[playerID]; exists {
		player.IsActive = false
		player.IsConnected = false
		player.Connection = nil

		// Broadcast player left
		manager.BroadcastToGame(gameState, game.MessageTypePlayerLeft, game.PlayerLeftPayload{
			PlayerID: playerID,
		})

		// Send system message
		manager.SendSystemMessage(roomCode, fmt.Sprintf("%s left the game", player.Name))

		// If game hasn't started, remove player completely
		if gameState.Status == models.GameStatusWaiting {
			delete(gameState.Players, playerID)
		}
	}

	return nil
}
