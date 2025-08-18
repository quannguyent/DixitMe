package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"dixitme/internal/game"
	"dixitme/internal/models"

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

// ConnectionMessage represents incoming WebSocket messages
type ConnectionMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Message types from client
const (
	ClientMessageJoinGame   = "join_game"
	ClientMessageCreateGame = "create_game"
	ClientMessageStartGame  = "start_game"
	ClientMessageSubmitClue = "submit_clue"
	ClientMessageSubmitCard = "submit_card"
	ClientMessageSubmitVote = "submit_vote"
	ClientMessageLeaveGame  = "leave_game"
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

// HandleWebSocket handles WebSocket connections
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

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connection established for player %s", playerID)

	// Send initial connection confirmation
	welcomeMsg := game.GameMessage{
		Type: "connection_established",
		Payload: map[string]interface{}{
			"player_id": playerID,
		},
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Handle incoming messages
	for {
		var msg ConnectionMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if err := handleMessage(conn, playerID, msg); err != nil {
			log.Printf("Error handling message: %v", err)
			sendError(conn, err.Error())
		}
	}

	// Clean up on disconnect
	handleDisconnect(playerID)
}

func handleMessage(conn *websocket.Conn, playerID uuid.UUID, msg ConnectionMessage) error {
	manager := game.GetManager()

	switch msg.Type {
	case ClientMessageCreateGame:
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

	case ClientMessageJoinGame:
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

	case ClientMessageStartGame:
		var payload StartGamePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		return manager.StartGame(payload.RoomCode, playerID)

	case ClientMessageSubmitClue:
		var payload SubmitCluePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		return manager.SubmitClue(payload.RoomCode, playerID, payload.Clue, payload.CardID)

	case ClientMessageSubmitCard:
		var payload SubmitCardPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		return manager.SubmitCard(payload.RoomCode, playerID, payload.CardID)

	case ClientMessageSubmitVote:
		var payload SubmitVotePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		return manager.SubmitVote(payload.RoomCode, playerID, payload.CardID)

	case ClientMessageLeaveGame:
		var payload LeaveGamePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		return handleLeaveGame(playerID, payload.RoomCode)

	default:
		return sendError(conn, "Unknown message type: "+msg.Type)
	}
}

func handleLeaveGame(playerID uuid.UUID, roomCode string) error {
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

		// If game hasn't started, remove player completely
		if gameState.Status == models.GameStatusWaiting {
			delete(gameState.Players, playerID)
		}
	}

	return nil
}

func handleDisconnect(playerID uuid.UUID) {
	// Mark player as disconnected in all their games
	// This is a simplified implementation - in a real system you'd track which games a player is in
	log.Printf("Player %s disconnected", playerID)
}

func sendError(conn *websocket.Conn, message string) error {
	errorMsg := game.GameMessage{
		Type:    game.MessageTypeError,
		Payload: game.ErrorPayload{Message: message},
	}
	return conn.WriteJSON(errorMsg)
}
