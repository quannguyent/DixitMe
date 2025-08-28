package game

import (
	"encoding/json"
	"fmt"

	"dixitme/internal/logger"

	"github.com/gorilla/websocket"
)

// GameBroadcastService defines broadcasting operations
type GameBroadcastService interface {
	BroadcastToGame(gameState *GameState, messageType MessageType, message interface{})
}

// BroadcastToGame sends a message to all connected players in a game
func (m *Manager) BroadcastToGame(game *GameState, messageType MessageType, payload interface{}) {
	log := logger.GetLogger()

	message := GameMessage{
		Type:    messageType,
		Payload: payload,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal broadcast message", "error", err)
		return
	}

	log.Info("Broadcasting message to game",
		"room_code", game.RoomCode,
		"message_type", messageType,
		"player_count", len(game.Players))

	// Add debug logging for GameState messages
	if messageType == MessageTypeGameState {
		if gameStatePayload, ok := payload.(GameStatePayload); ok {
			playerNames := make([]string, 0, len(gameStatePayload.GameState.Players))
			for _, player := range gameStatePayload.GameState.Players {
				playerNames = append(playerNames, fmt.Sprintf("%s(%s)", player.Name, map[bool]string{true: "bot", false: "human"}[player.IsBot]))
			}
			log.Info("GameState payload details",
				"room_code", game.RoomCode,
				"players", playerNames)
		}
	}

	// Send to all connected players, including fallback to global connection registry
	sentCount := 0
	for playerID, player := range game.Players {
		// Skip bots - they don't have WebSocket connections
		if player.IsBot {
			log.Debug("Skipping bot player", "player_name", player.Name, "player_id", playerID)
			continue
		}

		var conn *websocket.Conn

		// Try player's stored connection first
		if player.Connection != nil && player.IsConnected {
			conn = player.Connection
			log.Debug("Using stored connection", "player_id", playerID, "player_name", player.Name)
		} else {
			// Fallback: Try to get connection from global registry
			if globalConn := GetPlayerConnection(playerID); globalConn != nil {
				conn = globalConn
				// Update the player's connection reference
				player.Connection = conn
				player.IsConnected = true
				log.Debug("Using global registry connection", "player_id", playerID, "player_name", player.Name)
			} else {
				log.Warn("No connection found for player", "player_id", playerID, "player_name", player.Name)
			}
		}

		// Send message if we have a connection
		if conn != nil {
			if err := conn.WriteMessage(websocket.TextMessage, messageData); err != nil {
				logger.Error("Failed to send message to player",
					"error", err,
					"player_id", playerID,
					"room_code", game.RoomCode)
				// Mark player as disconnected and clear connection
				player.IsConnected = false
				player.Connection = nil
			} else {
				sentCount++
				log.Debug("Message sent successfully", "player_id", playerID, "player_name", player.Name)
			}
		}
	}

	log.Info("Broadcast completed",
		"room_code", game.RoomCode,
		"message_type", messageType,
		"messages_sent", sentCount,
		"total_players", len(game.Players))
}
