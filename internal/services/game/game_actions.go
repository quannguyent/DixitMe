package game

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"
	"dixitme/internal/services/bot"

	"github.com/google/uuid"
)

// GameService defines the core game management operations
type GameService interface {
	// Game lifecycle
	CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error)
	JoinGame(roomCode string, playerID uuid.UUID, playerName string) (*GameState, error)
	AddBot(roomCode string, botLevel string) (*GameState, error)
	RemovePlayer(roomCode string, playerID uuid.UUID) (*GameState, error)
	DeleteGame(roomCode string, playerID uuid.UUID) error
	LeaveGame(roomCode string, playerID uuid.UUID) (*GameState, error)
	StartGame(roomCode string, playerID uuid.UUID) error
	GetGame(roomCode string) *GameState
	GetActiveGamesCount() int
}

// CreateGame creates a new game with the given room code
func (m *Manager) CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if room code already exists
	if _, exists := m.games[roomCode]; exists {
		return nil, fmt.Errorf("room code already exists")
	}

	// Create new game state
	gameID := uuid.New()
	now := time.Now()

	// Initialize deck with all available cards (1-84 for standard Dixit)
	deck := make([]int, 84)
	for i := 0; i < 84; i++ {
		deck[i] = i + 1
	}
	// Shuffle the deck
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	game := &GameState{
		ID:           gameID,
		RoomCode:     roomCode,
		Players:      make(map[uuid.UUID]*Player),
		Status:       models.GameStatusWaiting,
		RoundNumber:  0,
		MaxRounds:    999, // Will be determined by 30 points or empty deck
		Deck:         deck,
		UsedCards:    make([]int, 0),
		CreatedAt:    now,
		LastActivity: now,
	}

	// Add creator as first player
	creator := &Player{
		ID:          creatorID,
		Name:        creatorName,
		Score:       0,
		Position:    1,
		Hand:        make([]int, 0),
		IsConnected: true,
		IsActive:    true,
	}

	game.Players[creatorID] = creator

	// Store in memory
	m.games[roomCode] = game

	// Persist to database
	if err := m.PersistGame(context.Background(), game); err != nil {
		delete(m.games, roomCode)
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "games_room_code_key") {
			return nil, fmt.Errorf("room code '%s' is already taken, please try a different one", roomCode)
		}
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Persist creator player to database
	dbCreator := &models.Player{
		ID:       creatorID,
		Name:     creatorName,
		Type:     models.PlayerTypeHuman,
		AuthType: models.AuthTypeGuest,
	}
	if err := m.PersistPlayer(context.Background(), dbCreator); err != nil {
		// If player already exists, it's not an error (e.g., rejoining with same ID)
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			delete(m.games, roomCode)
			return nil, fmt.Errorf("failed to persist creator player: %w", err)
		}
	}

	// Persist creator as game player
	if err := m.PersistGamePlayer(context.Background(), game.ID, creator); err != nil {
		delete(m.games, roomCode)
		return nil, fmt.Errorf("failed to persist creator game player: %w", err)
	}

	// Store in Redis for scaling
	if err := m.StoreGameInRedis(context.Background(), game); err != nil {
		logger.Error("Failed to store game in Redis", "error", err, "room_code", roomCode)
	}

	return game, nil
}

// JoinGame adds a player to an existing game
func (m *Manager) JoinGame(roomCode string, playerID uuid.UUID, playerName string) (*GameState, error) {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Update activity
	game.LastActivity = time.Now()

	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("game already started")
	}

	if len(game.Players) >= 6 {
		return nil, fmt.Errorf("game is full")
	}

	// Check if player already exists
	if _, exists := game.Players[playerID]; exists {
		return nil, fmt.Errorf("player already in game")
	}

	// Create new player
	player := &Player{
		ID:          playerID,
		Name:        playerName,
		Score:       0,
		Position:    len(game.Players) + 1,
		Hand:        make([]int, 0),
		IsConnected: true,
		IsActive:    true,
	}

	game.Players[playerID] = player

	// Persist player
	if err := m.PersistGamePlayer(context.Background(), game.ID, player); err != nil {
		delete(game.Players, playerID)
		return nil, fmt.Errorf("failed to persist player: %w", err)
	}

	// Update Redis
	if err := m.StoreGameInRedis(context.Background(), game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast player joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Broadcast updated game state to refresh player count
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("%s joined the game", playerName))

	return game, nil
}

// AddBot adds a bot player to an existing game
func (m *Manager) AddBot(roomCode string, botLevel string) (*GameState, error) {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	game.Lock()
	defer game.Unlock()

	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("cannot add bot to game in progress")
	}

	if len(game.Players) >= 6 {
		return nil, fmt.Errorf("game is full")
	}

	// Create bot player
	botNames := bot.GetBotNames()
	botName := botNames[rand.Intn(len(botNames))]

	// Ensure unique bot name
	for {
		nameExists := false
		for _, player := range game.Players {
			if player.Name == botName {
				nameExists = true
				break
			}
		}
		if !nameExists {
			break
		}
		botName = botNames[rand.Intn(len(botNames))]
	}

	botID := uuid.New()

	// Create bot in bot manager
	botManager := bot.GetBotManager()
	botPlayer := botManager.CreateBot(botName, bot.BotDifficulty(botLevel))
	botPlayer.SetGameID(game.ID)

	// Create game player
	player := &Player{
		ID:          botID,
		Name:        botName,
		Score:       0,
		Position:    len(game.Players),
		Hand:        make([]int, 0),
		Connection:  nil,
		IsConnected: true, // Bots are always "connected"
		IsActive:    true,
		IsBot:       true,
		BotLevel:    botLevel,
	}

	game.Players[botID] = player

	// Persist bot player to database
	dbPlayer := &models.Player{
		ID:       botID,
		Name:     botName,
		Type:     models.PlayerTypeBot,
		BotLevel: botLevel,
	}

	if err := m.PersistPlayer(context.Background(), dbPlayer); err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot player: %w", err)
	}

	if err := m.PersistGamePlayer(context.Background(), game.ID, player); err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot game player: %w", err)
	}

	// Update Redis
	if err := m.StoreGameInRedis(context.Background(), game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast bot joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Broadcast updated game state to refresh player count
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("Bot %s (%s difficulty) joined the game", botName, botLevel))

	logger.Info("Bot added to game", "bot_id", botID, "bot_name", botName, "bot_level", botLevel, "room_code", roomCode)

	return game, nil
}

// RemovePlayer removes a player (including bots) from a game
func (m *Manager) RemovePlayer(roomCode string, playerID uuid.UUID) (*GameState, error) {
	log := logger.GetLogger()

	game := m.getGame(roomCode)
	if game == nil {
		return nil, fmt.Errorf("game not found")
	}

	game.Lock()
	defer game.Unlock()

	// Check if game is in waiting state
	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("cannot remove players from a game that has already started")
	}

	// Check if player exists
	player, exists := game.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player not found in game")
	}

	// Remove player from memory
	delete(game.Players, playerID)

	// Update activity timestamp
	game.LastActivity = time.Now()

	// Remove from database
	db := database.GetDB()
	if err := db.Where("game_id = ? AND player_id = ?", game.ID, playerID).Delete(&models.GamePlayer{}).Error; err != nil {
		// Rollback memory change
		game.Players[playerID] = player
		return nil, fmt.Errorf("failed to remove player from database: %w", err)
	}

	// Broadcast player left
	m.BroadcastToGame(game, MessageTypePlayerLeft, PlayerLeftPayload{PlayerID: playerID})

	// Broadcast updated game state to refresh player count
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})

	// Send system message
	playerType := "Player"
	if player.IsBot {
		playerType = "Bot"
	}
	m.SendSystemMessage(roomCode, fmt.Sprintf("%s %s left the game", playerType, player.Name))

	log.Info("Player removed from game", "player_id", playerID, "player_name", player.Name, "is_bot", player.IsBot, "room_code", roomCode)

	return game, nil
}

// LeaveGame removes the current player from the game (same as RemovePlayer but different context)
func (m *Manager) LeaveGame(roomCode string, playerID uuid.UUID) (*GameState, error) {
	// Reuse the RemovePlayer logic
	return m.RemovePlayer(roomCode, playerID)
}

// DeleteGame deletes an entire game (only allowed by the creator/manager)
func (m *Manager) DeleteGame(roomCode string, playerID uuid.UUID) error {
	log := logger.GetLogger()

	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	game.Lock()
	defer game.Unlock()

	// Check if game is in waiting state
	if game.Status != models.GameStatusWaiting {
		return fmt.Errorf("cannot delete a game that has already started")
	}

	// Check if the player is the creator (first player or has admin rights)
	// For simplicity, we'll allow any player to delete for now, but you could add creator validation

	// Remove from memory first
	m.mu.Lock()
	delete(m.games, roomCode)
	m.mu.Unlock()

	// Remove from database
	db := database.GetDB()
	if err := db.Where("room_code = ?", roomCode).Delete(&models.Game{}).Error; err != nil {
		// Restore to memory if database deletion failed
		m.mu.Lock()
		m.games[roomCode] = game
		m.mu.Unlock()
		return fmt.Errorf("failed to delete game from database: %w", err)
	}

	// Broadcast game deletion to all players
	m.BroadcastToGame(game, MessageTypeGameDeleted, GameDeletedPayload{RoomCode: roomCode})

	// Send system message
	m.SendSystemMessage(roomCode, "Game has been deleted by the lobby manager")

	log.Info("Game deleted", "room_code", roomCode, "deleted_by", playerID)

	return nil
}

// StartGame starts a game if conditions are met
func (m *Manager) StartGame(roomCode string, playerID uuid.UUID) error {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Update activity
	game.LastActivity = time.Now()

	// Check if player is in the game
	if _, exists := game.Players[playerID]; !exists {
		return fmt.Errorf("player not in game")
	}

	// Check if game can start (minimum 3 players)
	logger.Info("StartGame player count check",
		"room_code", roomCode,
		"player_count", len(game.Players),
		"requesting_player", playerID)

	if len(game.Players) < 3 {
		// Log player details for debugging
		playerNames := make([]string, 0, len(game.Players))
		for _, player := range game.Players {
			playerNames = append(playerNames, fmt.Sprintf("%s(%s)", player.Name, map[bool]string{true: "bot", false: "human"}[player.IsBot]))
		}
		logger.Error("Not enough players to start game",
			"room_code", roomCode,
			"player_count", len(game.Players),
			"players", playerNames)
		return fmt.Errorf("need at least 3 players to start")
	}

	if game.Status != models.GameStatusWaiting {
		return fmt.Errorf("game already started")
	}

	// Initialize game
	game.Status = models.GameStatusInProgress

	// Deal cards to players
	m.dealCards(game)

	// Start first round
	if err := m.startNewRound(game); err != nil {
		return fmt.Errorf("failed to start first round: %w", err)
	}

	// Update database
	if err := m.UpdateGameStatus(context.Background(), game.ID, models.GameStatusInProgress); err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}

	// Broadcast game started
	m.BroadcastToGame(game, MessageTypeGameStarted, GameStartedPayload{GameState: game})

	// Broadcast updated game state so frontend transitions correctly
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})

	// Send system message
	m.SendSystemMessage(roomCode, "Game started! Let the storytelling begin!")

	return nil
}

// Helper methods

func (m *Manager) GetGame(roomCode string) *GameState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.games[roomCode]
}

func (m *Manager) getGame(roomCode string) *GameState {
	return m.GetGame(roomCode)
}

func (m *Manager) GetActiveGamesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.games)
}
