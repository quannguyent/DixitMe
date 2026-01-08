package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"dixitme/internal/auth"
	"dixitme/internal/game/core"
	"dixitme/internal/platform/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should validate the origin
		return true
	},
}

const (
	redisChannelGameEvents = "dixitme:game:events"
)

// PubSubMessage represents a message sent over Redis for synchronization
type PubSubMessage struct {
	SourceID       string           `json:"source_id"`
	RoomCode       string           `json:"room_code"`
	TargetPlayerID *uuid.UUID       `json:"target_player_id,omitempty"` // nil means broadcast to all
	Message        game.GameMessage `json:"message"`
}

// Hub manages WebSocket connections and implements game.Broadcaster.
type Hub struct {
	manager     *game.Manager
	jwtService  *auth.JWTService
	redisClient *redis.Client
	instanceID  string
	conns       map[string]map[uuid.UUID]*websocket.Conn // roomCode -> playerID -> conn
	playerRooms map[uuid.UUID]string                     // playerID -> roomCode
	mu          sync.RWMutex
}

func NewHub(jwtService *auth.JWTService, redisClient *redis.Client) *Hub {
	h := &Hub{
		jwtService:  jwtService,
		redisClient: redisClient,
		instanceID:  uuid.New().String(),
		conns:       make(map[string]map[uuid.UUID]*websocket.Conn),
		playerRooms: make(map[uuid.UUID]string),
	}

	// Start listening for Redis messages
	go h.subscribeToRedis()

	return h
}

func (h *Hub) SetManager(m *game.Manager) {
	h.manager = m
}

// subscribeToRedis listens for messages from Redis and broadcasts them locally
func (h *Hub) subscribeToRedis() {
	if h.redisClient == nil {
		return
	}

	pubsub := h.redisClient.Subscribe(context.Background(), redisChannelGameEvents)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var pubSubMsg PubSubMessage
		if err := json.Unmarshal([]byte(msg.Payload), &pubSubMsg); err != nil {
			logger.Error("Failed to unmarshal Redis message", "error", err)
			continue
		}

		// Skip messages from self
		if pubSubMsg.SourceID == h.instanceID {
			continue
		}

		if pubSubMsg.TargetPlayerID != nil {
			h.sendToPlayerLocal(pubSubMsg.RoomCode, *pubSubMsg.TargetPlayerID, pubSubMsg.Message)
		} else {
			h.broadcastLocal(pubSubMsg.RoomCode, pubSubMsg.Message)
		}
	}
}

// Broadcast implements game.Broadcaster.
// It publishes the message to Redis to be picked up by all instances (including self via broadcastLocal if we didn't skip self, but here we publish and also broadcast locally for efficiency or let Redis handle it. To keep it simple and consistent, we can publish and let the subscription handle it, OR publish and broadcast locally and skip self in subscription.
// Standard pattern: Publish to Redis. Redis subscription handles dispatch.
// Optimization: Publish to Redis (for others) AND broadcast locally (for self), then skip self in subscription.
func (h *Hub) Broadcast(roomCode string, msg game.GameMessage) {
	// 1. Broadcast locally immediately
	h.broadcastLocal(roomCode, msg)

	// 2. Publish to Redis for other instances
	if h.redisClient != nil {
		pubSubMsg := PubSubMessage{
			SourceID: h.instanceID,
			RoomCode: roomCode,
			Message:  msg,
		}
		payload, err := json.Marshal(pubSubMsg)
		if err == nil {
			h.redisClient.Publish(context.Background(), redisChannelGameEvents, payload)
		} else {
			logger.Error("Failed to marshal PubSub message", "error", err)
		}
	}
}

// broadcastLocal sends a message to all local connections in a room
func (h *Hub) broadcastLocal(roomCode string, msg game.GameMessage) {
	h.mu.RLock()
	conns := h.conns[roomCode]
	h.mu.RUnlock()

	if len(conns) == 0 {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal WS message", "error", err, "room_code", roomCode)
		return
	}

	for playerID, conn := range conns {
		if conn == nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			logger.Error("Failed to send WS message", "error", err, "room_code", roomCode, "player_id", playerID)
		}
	}
}

// SendToPlayer implements game.Broadcaster.
func (h *Hub) SendToPlayer(roomCode string, playerID uuid.UUID, msg game.GameMessage) {
	// 1. Send locally if connected
	h.sendToPlayerLocal(roomCode, playerID, msg)

	// 2. Publish to Redis for other instances (in case player is connected elsewhere?
	// Actually, a player is usually connected to one instance. But in a reconn scenario or if we don't know where they are...
	// If we know they are local, we don't strictly need to publish, but for simplicity and to handle edge cases (e.g. multiple tabs?), we can publish.
	// However, usually we can just check if local. If not found locally, publish?
	// For full statelessness, we should publish if we don't find them, or just always publish.
	// Let's always publish to be safe and consistent with the "stateless" goal,
	// but we can optimization: check if local first.
	// Since we did sendToPlayerLocal, if it succeeded, we might not need to publish?
	// But `sendToPlayerLocal` doesn't return success/fail easily here without changing sig.
	// Also, a user might have multiple connections (e.g. multiple tabs) potentially on different servers if sticky sessions aren't perfect?
	// Let's stick to: Send Local + Publish (Skip Self).

	if h.redisClient != nil {
		pubSubMsg := PubSubMessage{
			SourceID:       h.instanceID,
			RoomCode:       roomCode,
			TargetPlayerID: &playerID,
			Message:        msg,
		}
		payload, err := json.Marshal(pubSubMsg)
		if err == nil {
			h.redisClient.Publish(context.Background(), redisChannelGameEvents, payload)
		} else {
			logger.Error("Failed to marshal PubSub message", "error", err)
		}
	}
}

// sendToPlayerLocal sends a message to a specific local connection
func (h *Hub) sendToPlayerLocal(roomCode string, playerID uuid.UUID, msg game.GameMessage) {
	h.mu.RLock()
	// Check if the room exists
	if roomConns, ok := h.conns[roomCode]; ok {
		// Check if the player is in the room
		if conn, ok := roomConns[playerID]; ok && conn != nil {
			if err := conn.WriteJSON(msg); err != nil {
				logger.Error("Failed to send WS message to player", "error", err, "room_code", roomCode, "player_id", playerID)
			}
		}
	}
	h.mu.RUnlock()
}

// HandleWebSocket handles WebSocket connections without auth.
func (h *Hub) HandleWebSocket(c *gin.Context) {
	h.handleWebSocketConnection(c, h.extractPlayerID(c), nil)
}

// HandleWebSocketWithAuth handles WebSocket connections with JWT auth.
func (h *Hub) HandleWebSocketWithAuth(c *gin.Context) {
	playerID := h.extractPlayerID(c)
	var userInfo *auth.UserInfo

	token := extractTokenFromWebSocket(c)
	if token != "" && h.jwtService != nil {
		if info, err := h.jwtService.ExtractUserInfo(token); err == nil {
			userInfo = info
			playerID = info.SessionID
		}
	}

	h.handleWebSocketConnection(c, playerID, userInfo)
}

func (h *Hub) extractPlayerID(c *gin.Context) uuid.UUID {
	playerIDStr := c.Query("player_id")
	if playerIDStr != "" {
		if id, err := uuid.Parse(playerIDStr); err == nil {
			return id
		}
	}
	return uuid.New()
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
	ClientMessageSendChat   = "send_chat"
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

// handleWebSocketConnection upgrades and manages lifecycle.
func (h *Hub) handleWebSocketConnection(c *gin.Context, playerID uuid.UUID, userInfo *auth.UserInfo) {
	if h.manager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server not ready"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	playerName := fmt.Sprintf("Guest %s", playerID.String()[:8])
	authType := "guest"
	if userInfo != nil {
		playerName = userInfo.Name
		authType = string(userInfo.AuthType)
		logger.Info("Authenticated WebSocket connection established",
			"player_id", playerID, "auth_type", authType, "name", playerName)
	} else {
		logger.Info("Guest WebSocket connection established", "player_id", playerID)
	}

	h.registerConnection("", playerID, conn)
	defer h.unregisterConnection("", playerID)

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

	for {
		var msg ConnectionMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket unexpected close", "error", err, "player_id", playerID)
			}
			if roomCode, ok := h.getPlayerRoom(playerID); ok {
				h.clearPlayerRoom(playerID)
				h.manager.MarkPlayerDisconnected(roomCode, playerID)
			}
			break
		}

		if err := h.handleMessage(conn, playerID, msg); err != nil {
			logger.Error("Error handling WebSocket message", "error", err, "player_id", playerID, "message_type", msg.Type)
			sendError(conn, err.Error())
		}
	}
}

func (h *Hub) registerConnection(roomCode string, playerID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if roomCode == "" {
		roomCode = "__lobby__"
	}
	if _, ok := h.conns[roomCode]; !ok {
		h.conns[roomCode] = make(map[uuid.UUID]*websocket.Conn)
	}
	h.conns[roomCode][playerID] = conn
}

func (h *Hub) unregisterConnection(roomCode string, playerID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if roomCode == "" {
		roomCode = "__lobby__"
	}
	if players, ok := h.conns[roomCode]; ok {
		delete(players, playerID)
		if len(players) == 0 {
			delete(h.conns, roomCode)
		}
	}
}

func extractTokenFromWebSocket(c *gin.Context) string {
	if token := c.Query("token"); token != "" {
		return token
	}
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
	}
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}
	return ""
}

func (h *Hub) handleMessage(conn *websocket.Conn, playerID uuid.UUID, msg ConnectionMessage) error {
	if h.manager == nil {
		return fmt.Errorf("server not ready")
	}
	manager := h.manager

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

		h.registerConnection(payload.RoomCode, playerID, conn)
		h.setPlayerRoom(playerID, payload.RoomCode)

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

		h.registerConnection(payload.RoomCode, playerID, conn)
		h.setPlayerRoom(playerID, payload.RoomCode)

		return conn.WriteJSON(game.GameMessage{
			Type:    game.MessageTypeGameState,
			Payload: game.GameStatePayload{GameState: gameState},
		})

	case ClientMessageStartGame:
		var payload StartGamePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := manager.StartGame(payload.RoomCode, playerID); err != nil {
			return err
		}

	case ClientMessageSubmitClue:
		var payload SubmitCluePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := manager.SubmitClue(payload.RoomCode, playerID, payload.Clue, payload.CardID); err != nil {
			return err
		}

	case ClientMessageSubmitCard:
		var payload SubmitCardPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := manager.SubmitCard(payload.RoomCode, playerID, payload.CardID); err != nil {
			return err
		}

	case ClientMessageSubmitVote:
		var payload SubmitVotePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := manager.SubmitVote(payload.RoomCode, playerID, payload.CardID); err != nil {
			return err
		}

	case ClientMessageLeaveGame:
		var payload LeaveGamePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := manager.LeaveGame(payload.RoomCode, playerID); err != nil {
			return err
		}
		h.unregisterConnection(payload.RoomCode, playerID)
		h.clearPlayerRoom(playerID)

	case ClientMessageSendChat:
		var payload SendChatPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := manager.SendChatMessage(payload.RoomCode, playerID, payload.Message, payload.MessageType); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

// sendError sends an error message to a client
func sendError(conn *websocket.Conn, errMsg string) {
	conn.WriteJSON(game.GameMessage{
		Type: game.MessageTypeError,
		Payload: game.ErrorPayload{
			Message: errMsg,
		},
	})
}

func (h *Hub) getPlayerRoom(playerID uuid.UUID) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	roomCode, ok := h.playerRooms[playerID]
	return roomCode, ok
}

func (h *Hub) setPlayerRoom(playerID uuid.UUID, roomCode string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.playerRooms[playerID] = roomCode
}

func (h *Hub) clearPlayerRoom(playerID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.playerRooms, playerID)
}
