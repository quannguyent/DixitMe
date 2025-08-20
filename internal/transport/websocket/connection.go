package websocket

import (
	"net/http"

	"dixitme/internal/logger"
	"dixitme/internal/services/auth"
	"dixitme/internal/services/game"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should validate the origin
		return true
	},
}

// HandleWebSocket handles WebSocket connections (legacy, no auth)
func HandleWebSocket(c *gin.Context) {
	// Get player ID from query params or create new one
	playerIDStr := c.Query("player_id")
	var playerID uuid.UUID
	var err error

	if playerIDStr != "" {
		playerID, err = uuid.Parse(playerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
			return
		}
	} else {
		playerID = uuid.New()
	}

	handleWebSocketConnection(c, playerID, nil)
}

// HandleWebSocketWithAuth handles WebSocket connections with authentication support
func HandleWebSocketWithAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID, userInfo, err := extractPlayerInfo(c, jwtService)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
			return
		}

		handleWebSocketConnection(c, playerID, userInfo)
	}
}

// handleWebSocketConnection handles the actual WebSocket connection logic
func handleWebSocketConnection(c *gin.Context, playerID uuid.UUID, userInfo *auth.UserInfo) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	var playerName string
	var authType string
	if userInfo != nil {
		playerName = userInfo.Name
		authType = string(userInfo.AuthType)
		logger.Info("Authenticated WebSocket connection established",
			"player_id", playerID, "auth_type", authType, "name", playerName)
	} else {
		playerName = "Guest " + playerID.String()[:8]
		authType = "guest"
		logger.Info("Guest WebSocket connection established", "player_id", playerID)
	}

	// Send initial connection confirmation
	welcomeMsg := game.GameMessage{
		Type: "connection_established",
		Payload: map[string]interface{}{
			"player_id":     playerID,
			"player_name":   playerName,
			"auth_type":     authType,
			"authenticated": userInfo != nil,
		},
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		logger.Error("Failed to send welcome message", "error", err, "player_id", playerID)
		return
	}

	// Handle incoming messages
	for {
		var msg ConnectionMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket unexpected close", "error", err, "player_id", playerID)
			}
			break
		}

		if err := handleMessage(conn, playerID, msg); err != nil {
			logger.Error("Error handling WebSocket message", "error", err, "player_id", playerID, "message_type", msg.Type)
			sendError(conn, err.Error())
		}
	}

	// Clean up on disconnect
	handleDisconnect(playerID)
}

// handleDisconnect cleans up when a player disconnects
func handleDisconnect(playerID uuid.UUID) {
	// Mark player as disconnected in all their games
	// This is a simplified implementation - in a real system you'd track which games a player is in
	logger.Info("Player disconnected", "player_id", playerID)
}

// sendError sends an error message to the WebSocket client
func sendError(conn *websocket.Conn, message string) error {
	errorMsg := game.GameMessage{
		Type:    game.MessageTypeError,
		Payload: game.ErrorPayload{Message: message},
	}
	return conn.WriteJSON(errorMsg)
}
