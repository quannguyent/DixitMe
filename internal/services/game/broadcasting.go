package game

import (
	"encoding/json"

	"dixitme/internal/logger"
)

// GameBroadcastService defines broadcasting operations
type GameBroadcastService interface {
	BroadcastToGame(gameState *GameState, messageType MessageType, message interface{})
}

// BroadcastToGame sends a message to all connected players in a game
func (m *Manager) BroadcastToGame(game *GameState, messageType MessageType, payload interface{}) {
	message := GameMessage{
		Type:    messageType,
		Payload: payload,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal broadcast message", "error", err)
		return
	}

	// Send to all connected players
	for _, player := range game.Players {
		if player.Connection != nil && player.IsConnected {
			if err := player.Connection.WriteMessage(1, messageData); err != nil {
				logger.Error("Failed to send message to player",
					"error", err,
					"player_id", player.ID,
					"room_code", game.RoomCode)
				// Mark player as disconnected
				player.IsConnected = false
			}
		}
	}
}
