package game

import (
	"fmt"
	"strings"
	"time"

	"dixitme/internal/logger"
	"dixitme/internal/models"

	"github.com/google/uuid"
)

// SendChatMessage handles sending chat messages in a game
func (m *Manager) SendChatMessage(roomCode string, playerID uuid.UUID, message string, messageType string) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	player, exists := game.Players[playerID]
	if !exists {
		return fmt.Errorf("player not in game")
	}

	// Validate message type
	if messageType == "" {
		messageType = "chat"
	}
	if messageType != "chat" && messageType != "emote" {
		return fmt.Errorf("invalid message type")
	}

	// Validate message content
	if len(strings.TrimSpace(message)) == 0 {
		return fmt.Errorf("message cannot be empty")
	}
	if len(message) > 500 { // Max message length
		return fmt.Errorf("message too long")
	}

	// Determine current phase
	currentPhase := "lobby"
	if game.Status == models.GameStatusInProgress && game.CurrentRound != nil {
		currentPhase = string(game.CurrentRound.Status)
	}

	// Only allow chat in lobby and voting phases
	if currentPhase != "lobby" && currentPhase != "voting" {
		return fmt.Errorf("chat not allowed in current phase")
	}

	// Create chat message
	chatMessage := models.ChatMessage{
		ID:          uuid.New(),
		GameID:      game.ID,
		PlayerID:    playerID,
		Message:     strings.TrimSpace(message),
		MessageType: messageType,
		Phase:       currentPhase,
		IsVisible:   true,
		CreatedAt:   time.Now(),
	}

	// Persist to database
	if err := m.persistChatMessage(&chatMessage); err != nil {
		return fmt.Errorf("failed to persist chat message: %w", err)
	}

	// Create payload
	payload := ChatMessagePayload{
		ID:          chatMessage.ID,
		PlayerID:    playerID,
		PlayerName:  player.Name,
		Message:     chatMessage.Message,
		MessageType: chatMessage.MessageType,
		Phase:       chatMessage.Phase,
		Timestamp:   chatMessage.CreatedAt,
	}

	// Broadcast to all players in the game
	m.BroadcastToGame(game, MessageTypeChatMessage, payload)

	return nil
}

// GetChatHistory retrieves chat messages for a game and phase
func (m *Manager) GetChatHistory(roomCode string, phase string, limit int) ([]ChatMessagePayload, error) {
	game := m.getGame(roomCode)
	if game == nil {
		return nil, fmt.Errorf("game not found")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Get messages from database
	messages, err := m.getChatMessages(game.ID, phase, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	// Convert to payload format
	payloads := make([]ChatMessagePayload, 0, len(messages))
	for _, msg := range messages {
		var playerName string
		if msg.PlayerID == uuid.Nil {
			playerName = "System"
		} else if player, exists := game.Players[msg.PlayerID]; exists {
			playerName = player.Name
		} else {
			playerName = "Unknown"
		}

		payloads = append(payloads, ChatMessagePayload{
			ID:          msg.ID,
			PlayerID:    msg.PlayerID,
			PlayerName:  playerName,
			Message:     msg.Message,
			MessageType: msg.MessageType,
			Phase:       msg.Phase,
			Timestamp:   msg.CreatedAt,
		})
	}

	return payloads, nil
}

// SendSystemMessage sends a system message (e.g., "Player joined", "Round started")
func (m *Manager) SendSystemMessage(roomCode string, message string) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	// Determine current phase
	currentPhase := "lobby"
	if game.Status == models.GameStatusInProgress && game.CurrentRound != nil {
		currentPhase = string(game.CurrentRound.Status)
	}

	// Create system message with a system player ID (using nil UUID)
	systemPlayerID := uuid.Nil
	chatMessage := models.ChatMessage{
		ID:          uuid.New(),
		GameID:      game.ID,
		PlayerID:    systemPlayerID,
		Message:     message,
		MessageType: "system",
		Phase:       currentPhase,
		IsVisible:   true,
		CreatedAt:   time.Now(),
	}

	// Persist to database
	if err := m.persistChatMessage(&chatMessage); err != nil {
		logger.Error("Failed to persist system message", "error", err)
		// Continue anyway - system messages are not critical
	}

	// Create payload
	payload := ChatMessagePayload{
		ID:          chatMessage.ID,
		PlayerID:    systemPlayerID,
		PlayerName:  "System",
		Message:     chatMessage.Message,
		MessageType: chatMessage.MessageType,
		Phase:       chatMessage.Phase,
		Timestamp:   chatMessage.CreatedAt,
	}

	// Broadcast to all players in the game
	m.BroadcastToGame(game, MessageTypeChatMessage, payload)

	return nil
}
