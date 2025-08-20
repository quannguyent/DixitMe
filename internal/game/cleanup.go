package game

import (
	"time"

	"dixitme/internal/logger"
	"dixitme/internal/models"
)

// Game cleanup service

func (m *Manager) startCleanupService() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	logger.Info("Game cleanup service started", "interval", m.cleanupInterval)

	for {
		select {
		case <-ticker.C:
			m.cleanupInactiveGames()
		case <-m.stopCleanup:
			logger.Info("Game cleanup service stopped")
			return
		}
	}
}

func (m *Manager) cleanupInactiveGames() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string
	var emptyRooms []string
	var occupiedRooms []string

	for roomCode, game := range m.games {
		game.mu.RLock()

		// Count active/connected players
		activePlayerCount := 0
		for _, player := range game.Players {
			if player.IsConnected {
				activePlayerCount++
			}
		}

		emptyRoomTimeout := 10 * time.Minute    // Empty rooms: 10 minutes
		occupiedRoomTimeout := 30 * time.Minute // Rooms with players: 30 minutes

		var shouldRemove bool
		var reason string

		if activePlayerCount == 0 {
			// Empty room - use shorter timeout
			shouldRemove = game.IsInactive(emptyRoomTimeout)
			if shouldRemove {
				emptyRooms = append(emptyRooms, roomCode)
				reason = "Game closed - empty room (10 minutes)"
			}
		} else {
			// Room has players - use longer timeout
			shouldRemove = game.IsInactive(occupiedRoomTimeout)
			if shouldRemove {
				occupiedRooms = append(occupiedRooms, roomCode)
				reason = "Game closed due to inactivity (30 minutes)"
			}
		}

		if shouldRemove {
			toRemove = append(toRemove, roomCode)
			// Store reason for later use
			game.mu.RUnlock()

			// Notify all connected players that the game is being closed
			m.broadcastGameClosure(game, reason)

			// Mark game as abandoned in database
			m.markGameAsAbandoned(game)
		} else {
			game.mu.RUnlock()
		}
	}

	if len(toRemove) > 0 {
		logger.Info("Cleaning up inactive games",
			"total_count", len(toRemove),
			"empty_rooms", len(emptyRooms),
			"occupied_rooms", len(occupiedRooms),
			"empty_room_codes", emptyRooms,
			"occupied_room_codes", occupiedRooms)

		// Remove from memory
		for _, roomCode := range toRemove {
			delete(m.games, roomCode)
		}
	}
}

func (m *Manager) broadcastGameClosure(game *GameState, reason string) {
	// Send error message to notify players of game closure
	m.BroadcastToGame(game, MessageTypeError, ErrorPayload{
		Message: reason,
	})

	logger.Info("Broadcasted game closure",
		"room_code", game.RoomCode,
		"reason", reason,
		"player_count", len(game.Players))
}

func (m *Manager) markGameAsAbandoned(game *GameState) {
	if err := m.updateGameStatus(game.ID, models.GameStatusAbandoned); err != nil {
		logger.Error("Failed to mark game as abandoned",
			"error", err,
			"room_code", game.RoomCode,
			"game_id", game.ID)
	}
}

func (m *Manager) StopCleanupService() {
	select {
	case m.stopCleanup <- true:
	default:
		// Channel might be full or already closed
	}
}
